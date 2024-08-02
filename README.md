# gomaxprocs-injector

Inject CFS-aware `GOMAXPROCS` environment variable into every pod.

```
GOMAXPROCS=max(1, floor(CPU limit of the container))
```

Note that, this project is a workaround for
https://github.com/golang/go/issues/33803. If golang addresses the issue
internally, this project would be no longer needed.

Although there is already an
[automaxprocs](https://github.com/uber-go/automaxprocs) package that
automatically sets `GOMAXPROCS`, many applications still cannot use it for various
reasons. `gomaxprocs-injector` can complement these applications by injecting
optimized `GOMAXPROCS` environment variable into their pods, improving their
performance.

## Getting Started

### Install cert-manager
gomaxprocs-injector uses [cert-manager](https://cert-manager.io/docs/) for
certificate management of Admission Webhook. Make sure you have already
installed cert-manager before you start.

- [Install cert-manager on kubernetes](https://cert-manager.io/docs/installation/)

### Deploy gomaxprocs-injector
```sh
VERSION=latest envsubst < gomaxprocs-injector.yaml | kubectl apply -f -

# Wait for gomaxprocs-injector to be rollout
kubectl rollout status -n gomaxprocs-injector deployment gomaxprocs-injector
```

### Clean up
```sh
VERSION=latest envsubst < gomaxprocs-injector.yaml | kubectl delete -f -
```

## Disabling injection

Injection can be disabled for a pod by adding `gomaxprocs-injector/inject:
disabled` annotation.
