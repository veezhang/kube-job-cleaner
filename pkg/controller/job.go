package controller

import (
	"reflect"
	"strconv"
	"sync"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog"
)

const (
	JobTTLSecondsAfterFinishedAnnotationsKey string = "kube-job-cleaner/ttlSecondsAfterFinished"
)

type JobController struct {
	controller *Controller
	informer   cache.SharedIndexInformer
	opts       JobOptions
}

func NewJobController(c *Controller, opts JobOptions) *JobController {
	jobController := &JobController{
		controller: c,
		opts:       opts,
	}

	informer := cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
				return c.kclient.BatchV1().Jobs(jobController.opts.Namespace).List(options)
			},
			WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
				return c.kclient.BatchV1().Jobs(jobController.opts.Namespace).Watch(options)
			},
		},
		&batchv1.Job{},
		time.Second*time.Duration(jobController.opts.ResyncPeriod),
		cache.Indexers{},
	)

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			jobController.Handle(obj)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			if !reflect.DeepEqual(oldObj, newObj) {
				jobController.Handle(newObj)
			}
		},
		DeleteFunc: func(obj interface{}) {
			jobController.Handle(obj)
		},
	})

	jobController.informer = informer

	return jobController
}

func (c *JobController) Run(stopCh <-chan struct{}) {
	klog.Info("Listening for job changes.")

	wg := &sync.WaitGroup{}
	wg.Add(2)

	go func() {
		klog.Info("Start job informer run.")
		c.informer.Run(stopCh)
		wg.Done()
		klog.Info("Stop job informer run.")
	}()

	go func() {
		c.checkInterval(stopCh)
		wg.Done()
	}()

	<-stopCh
	wg.Wait()
}

func (c *JobController) checkInterval(stopCh <-chan struct{}) {
	klog.Info("Start job check interval.")
	interval := time.Second * time.Duration(c.opts.CheckInterval)
	timer := time.NewTimer(interval)
	for {
		select {
		case <-timer.C:
			for _, obj := range c.informer.GetStore().List() {
				c.Handle(obj)
			}
			timer.Reset(interval)
		case <-stopCh:
			timer.Stop()
			klog.Info("Stop job check interval.")
			return
		}
	}
}

func (c *JobController) Handle(obj interface{}) {
	jobObj, ok := obj.(*batchv1.Job)
	if !ok {
		klog.Warning("Not Job.")
		return
	}

	completionTime := jobObj.Status.CompletionTime

	if completionTime.IsZero() {
		return
	}

	// Get the ttlSecondsAfterFinished
	annotations := jobObj.GetAnnotations()
	ttlSecondsAfterFinishedStr, ok := annotations[JobTTLSecondsAfterFinishedAnnotationsKey]
	if !ok {
		return
	}
	ttlSecondsAfterFinished, err := strconv.ParseInt(ttlSecondsAfterFinishedStr, 10, 64)
	if err != nil {
		klog.Warningf("annotations[%s] is Not integer number : %s.", JobTTLSecondsAfterFinishedAnnotationsKey, err)
		return
	}

	now := time.Now()
	completedSeconds := now.Sub(completionTime.Time).Seconds()

	klog.V(3).Infof("Handle Job : %s, CompletionTime = %s, Completed = %f, TTLSecondsAfterFinished = %d.",
		jobObj.GetName(),
		jobObj.Status.CompletionTime,
		completedSeconds,
		ttlSecondsAfterFinished,
	)

	if completedSeconds < float64(ttlSecondsAfterFinished) {
		return
	}

	jobname := jobObj.GetName()

	if c.controller.opts.DryRun {
		klog.V(2).Infof("Job '%s' will be deleted(dry-run).", jobname)
		return
	}

	klog.V(2).Infof("Job '%s' will be deleted.", jobname)

	deletePolicy := metav1.DeletePropagationBackground
	jobDeleteOptions := metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}
	if err := c.controller.kclient.BatchV1().Jobs(jobObj.Namespace).Delete(jobname, &jobDeleteOptions); err != nil {
		if !apierrs.IsNotFound(err) {
			klog.Warningf("Delete job falied : .", err)
		}
	}
}
