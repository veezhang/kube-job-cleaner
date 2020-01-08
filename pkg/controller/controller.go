package controller

import (
	"k8s.io/client-go/kubernetes"
)

type Controller struct {
	kclient *kubernetes.Clientset
	opts    Options
}

func NewController(kclient *kubernetes.Clientset, opts Options) *Controller {
	return &Controller{
		kclient: kclient,
		opts:    opts,
	}
}
