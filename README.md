# dynowatch

A DYNamic Object WATCHer for Kubernetes

## What Is It?

Dynowatch is an event listener for Kubernetes, and can be configured to watch for changes to any
Kubernetes object type - including custom resources!
When a change is detected, Dynowatch emits a [CloudEvent](https://cloudevents.io/) to its
configured CloudEvent receiver.

What can you do with these events?
You decide!
Anything that is capable of consuming a CloudEvent can consume events emitted by Dynowatch.

## Try It!

Before you begin:

- Ensure you have access to a Kubernetes cluster.
- You have the Go SDK version 1.20 or higher installed.

Next, clone this repository and cd into it:

```sh
$ git clone https://github.com/kubearchive/dynowatch.git
$ cd dynowatch
```

Now run the watcher - by default it is configured to watch `Job` objects:

```sh
$ make run
```

Want to see some events? In a separate shell, create a simple CronJob:

```sh
$ kubectl apply -f - <<EOF
apiVersion: batch/v1
kind: CronJob
metadata:
  name: hello
spec:
  schedule: "* * * * *"
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: hello
            image: busybox:1.28
            imagePullPolicy: IfNotPresent
            command:
            - /bin/sh
            - -c
            - date; echo Hello from the Kubernetes cluster
          restartPolicy: OnFailure
EOF
```

...and watch the events start to be delivered!

```sh
2023-12-13T18:28:11-05:00	INFO	Delivered event	{"controller": "jobs", "controllerGroup": "batch", "controllerKind": "Job", "Job": {"name":"hello-28375167","namespace":"default"}, "namespace": "default", "name": "hello-28375167", "reconcileID": "000a5f3b-8fef-4ee6-82a5-9edab0fd6093"}
2023-12-13T18:28:11-05:00	INFO	Delivered event	{"controller": "jobs", "controllerGroup": "batch", "controllerKind": "Job", "Job": {"name":"hello-28375168","namespace":"default"}, "namespace": "default", "name": "hello-28375168", "reconcileID": "4ad228b5-71ee-4044-a5f7-73bf1f6515e0"}
2023-12-13T18:28:11-05:00	INFO	Delivered event	{"controller": "jobs", "controllerGroup": "batch", "controllerKind": "Job", "Job": {"name":"hello-28375166","namespace":"default"}, "namespace": "default", "name": "hello-28375166", "reconcileID": "a7057df1-a2fa-4d32-8fea-fd257755ee35"}
```

The [manager](/docs/manager.md) reference contains more information on how to configure Dynowatch.

## Prior Art

This project was inspired by the following projects:

- [Tekton Results](https://github.com/tektoncd/results). This consists of a _watcher_ and
  _apiserver_ that are tightly coupled. The watcher has some controller code that is type-agnostic,
  but is challenging to extend.
- [Kubewatch](https://github.com/robusta-dev/kubewatch). This is capable of emitting CloudEvents
  for core Kubernetes objects. It does not have support for custom resources yet (see
  [issue #18](https://github.com/robusta-dev/kubewatch/issues/18)).
