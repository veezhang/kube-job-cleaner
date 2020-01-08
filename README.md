# kube-job-cleaner

Automatically delete completed Jobs.

In some cases, you cannot customize the kubernetes's component arguments. So it's impossible to support the job's `ttlSecondsAfterFinished` feature by add `--feature-gates=TTLAfterFinished=true` to `controller`, `apiserver`, `scheduler` arguments list.

And so it's come.

You just need to add a annotations(`kube-job-cleaner/ttlSecondsAfterFinished`) in the job. For example,

```yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: pi
  annotations:
    kube-job-cleaner/ttlSecondsAfterFinished: "30"
spec:
  template:
    spec:
      containers:
      - name: pi
        image: perl
        command: ["perl",  "-Mbignum=bpi", "-wle", "print bpi(2000)"]
      restartPolicy: Never
```

## Deploy & Example

This is a example in kubernets cluster. Also you can deploy out of cluster.

```shell
# Create Namespace
kubectl create -f https://raw.githubusercontent.com/veezhang/kube-job-cleaner/master/deploy/namespace.yaml
# Create RBAC
kubectl create -f https://raw.githubusercontent.com/veezhang/kube-job-cleaner/master/deploy/rbac.yaml
# Create deployment
kubectl create -f https://raw.githubusercontent.com/veezhang/kube-job-cleaner/master/deploy/deployment.yaml
# Get the log
kubectl -n kube-job-cleaner logs -f -l app=kube-job-cleaner
# Open another terminal to create a job
kubectl create -f https://k8s.io/examples/controllers/job.yaml
```

## Clearn

```shell
# Create deployment
kubectl delete -f https://raw.githubusercontent.com/veezhang/kube-job-cleaner/master/deploy/deployment.yaml
# Create RBAC
kubectl delete -f https://raw.githubusercontent.com/veezhang/kube-job-cleaner/master/deploy/rbac.yaml
# Create Namespace
kubectl delete -f https://raw.githubusercontent.com/veezhang/kube-job-cleaner/master/deploy/namespace.yaml
```

## Command Usage

```shell
Usage of /usr/local/bin/kube-job-cleaner:
  -add_dir_header
        If true, adds the file directory to the header
  -alsologtostderr
        log to standard error as well as files
  -dry-run
        If true, only print the log.
  -job-check-interval int
        The job check interval secends. (default 60)
  -job-namespace string
        If non-empty, only watch this namespace; otherwise all namespaces.
  -job-resync-period int
        The job informer resync period secends. (default 60)
  -kubeconfig string
        Paths to a kubeconfig. Only required if out-of-cluster.
  -log_backtrace_at value
        when logging hits line file:N, emit a stack trace
  -log_dir string
        If non-empty, write log files in this directory
  -log_file string
        If non-empty, use this log file
  -log_file_max_size uint
        Defines the maximum size a log file can grow to. Unit is megabytes. If the value is 0, the maximum file size is unlimited. (default 1800)
  -logtostderr
        log to standard error instead of files (default true)
  -master --kubeconfig
        (Deprecated: switch to --kubeconfig) The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.
  -skip_headers
        If true, avoid header prefixes in the log messages
  -skip_log_headers
        If true, avoid headers when opening log files
  -stderrthreshold value
        logs at or above this threshold go to stderr (default 2)
  -v value
        number for the log level verbosity
  -vmodule value
        comma-separated list of pattern=N settings for file-filtered logging
```
