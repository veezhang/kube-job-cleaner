package controller

type Options struct {
	DryRun bool
}

type JobOptions struct {
	Namespace     string
	ResyncPeriod  int
	CheckInterval int
}
