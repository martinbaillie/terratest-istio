# terratest-istio

[![CircleCI](https://circleci.com/gh/martinbaillie/terratest-istio/tree/master.svg?style=shield)](https://circleci.com/gh/martinbaillie/terratest-istio/tree/master)
[![GoDoc](https://godoc.org/martinbaillie/terratest-istio?status.svg)](https://godoc.org/github.com/martinbaillie/terratest-istio/modules/istio)
[![Codecov](https://codecov.io/gh/martinbaillie/terratest-istio/branch/master/graph/badge.svg)](https://codecov.io/gh/martinbaillie/terratest-istio)
[![Go Report Card](https://goreportcard.com/badge/github.com/martinbaillie/terratest-istio)](https://goreportcard.com/report/github.com/martinbaillie/terratest-istio)

| Supports     | Istio 1.2.x   | Istio 1.1.x   | Istio 1.0.x   | 
| -------------|:-------------:|:-------------:|:-------------:|
| **Go 1.12**  |:white_check_mark:|:white_check_mark:|:construction:|
| **Go 1.11**  |:white_check_mark:|:white_check_mark:|:construction:|

## About

An Istio module for the [Terratest](https://github.com/gruntwork-io/terratest) infrastructure testing library.

### Features
##### Envoy sidecar config
- Remote Istio Envoy sidecar configuration dumps (Bootstrap/Clusters/Listeners/Routes).
- Marshaling into Envoy [Go control plane](https://github.com/envoyproxy/go-control-plane) objects.
- Helper functions for checking key configuration attributes of sidecar Clusters/Listeners/Routes.
- Confirm Istio Pilot(s) configuration is in-sync with Istio Envoy sidecars.

##### Authn status
_Coming soon_

Check what authentication policies and destination rules Pilot uses to config a proxy instance, and check if TLS settings are compatible between them.

> NOTE: `CONFLICT` status does not always mean an inconsistency between
> destination rules and policies. Services without sidecars will show as
> `CONFLICT`ed as well e.g. `svc/kubernetes`, `svc/istio-policy.`

##### RBAC
_Coming soon_

Check the TLS/JWT/RBAC settings based on Envoy config

## Examples

### Envoy sidecars

Combine with [Terratest's Kubernetes
module](https://godoc.org/github.com/gruntwork-io/terratest/modules/k8s) to test
expected Istio configuration of Envoy sidecars in pods.

##### Clusters
```go
clusters, err := GetClustersConfigDumpForPodE(t, nil, "details-v1-abcdef")
require.Nil(t, err)
assert.True(t, IsClustersConfigClusteredTo(clusters, "reviews.default.svc.cluster.local",
            "http", "inbound", 9080))
```

##### Listeners
```go
listeners, err := GetListenersConfigDumpForPodE(t, nil, "details-v1-abcdef")
require.Nil(t, err)
assert.True(t, IsListenersConfigListeningOn(listeners, "HTTP", "0.0.0.0", 9080))
```

##### Routes
```go
routes, err := GetRoutesConfigDumpForPodE(t, nil, "details-v1-abcdef")
require.Nil(t, err)
assert.True(t, IsRoutesConfigRoutingTo(routes, "reviews.default.svc.cluster.local", 9080))
```

##### Bootstrap
```go
bootstrap, err := GetBootstrapConfigDumpForPodE(t, nil, "details-v1-abcdef")
require.Nil(t, err)
// Check `bootstrap` for expected Envoy bootstrap configuration
// ...
```

##### Pilot sync
```go
assert.True(t, ArePilotsSyncedToPod(t, nil, "details-v1-abcdef")
```
