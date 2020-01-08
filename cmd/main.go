package main

import (
	"flag"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/veezhang/kube-job-cleaner/pkg/controller"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"

	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

func main() {
	klog.InitFlags(nil)

	opts := controller.Options{}
	jobOpts := controller.JobOptions{}

	flag.BoolVar(&opts.DryRun, "dry-run", false, "If true, only print the log.")
	flag.StringVar(&jobOpts.Namespace, "job-namespace", "", "If non-empty, only watch this namespace; otherwise all namespaces.")
	flag.IntVar(&jobOpts.ResyncPeriod, "job-resync-period", 60, "The job informer resync period secends.")
	flag.IntVar(&jobOpts.CheckInterval, "job-check-interval", 60, "The job check interval secends.")

	flag.Parse()

	sigs := make(chan os.Signal, 1)
	stop := make(chan struct{})

	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	cfg, err := config.GetConfig()
	if err != nil {
		klog.Errorf("Get kubernetes config falied: %s.", err)
		os.Exit(1)
	}

	kclient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		klog.Errorf("Create kubernetes client falied: %s.", err)
		os.Exit(1)
	}

	c := controller.NewController(kclient, opts)
	jobController := controller.NewJobController(c, jobOpts)

	if opts.DryRun {
		klog.Info("Performing dry run.")
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)

	go func() {
		jobController.Run(stop)
		klog.Info("jobController down.")
		wg.Done()
	}()

	<-sigs

	klog.Info("Shutting down.")

	close(stop)
	wg.Wait()

	klog.Info("Shut down.")
}
