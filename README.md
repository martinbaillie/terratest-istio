# terratest-istio

[![CircleCI](https://circleci.com/gh/martinbaillie/terratest-istio/tree/master.svg?style=shield)](https://circleci.com/gh/martinbaillie/terratest-istio/tree/master)
[![GoDoc](https://godoc.org/martinbaillie/terratest-istio?status.svg)](https://godoc.org/github.com/martinbaillie/terratest-istio/modules/istio)
[![Codecov](https://codecov.io/gh/martinbaillie/terratest-istio/branch/master/graph/badge.svg)](https://codecov.io/gh/martinbaillie/terratest-istio)
[![Go Report Card](https://goreportcard.com/badge/github.com/martinbaillie/terratest-istio)](https://goreportcard.com/report/github.com/martinbaillie/terratest-istio)

## About

An Istio module for the [Terratest](https://github.com/gruntwork-io/terratest) infrastructure testing library.

### Features
##### Envoy sidecar config
- Remote Istio Envoy sidecar configuration dumps (Bootstrap/Clusters/Listeners/Routes).
- Marshaling into Envoy [Go control plane](https://github.com/envoyproxy/go-control-plane) objects.
- Helper functions for checking key configuration attributes of sidecar Clusters/Listeners/Routes.

##### RBAC
_Coming soon_

##### Pilot sync status
_Coming soon_

### Tested against
- Go 1.12 [Istio 1.1.x]
- Go 1.11 [Istio 1.1.x]

## Examples

### Envoy sidecars
##### Clusters
```go
clusters, err := GetClustersConfigDumpForPodE(t, nil, "details")
require.Nil(t, err)
assert.True(t, IsClustersConfigClusteredTo(clusters, "reviews.default.svc.cluster.local",
            "http", "inbound", 9080))
```

##### Listeners
```go
listeners, err := GetListenersConfigDumpForPodE(t, nil, "details")
require.Nil(t, err)
assert.True(t, IsListenersConfigListeningOn(listeners, "HTTP", "0.0.0.0", 9080))
```

##### Routes
```go
routes, err := GetRoutesConfigDumpForPodE(t, nil, "details")
require.Nil(t, err)
assert.True(t, IsRoutesConfigRoutingTo(routes, "reviews.default.svc.cluster.local", 9080))
```

##### Bootstrap
```go
bootstrap, err := GetBootstrapConfigDumpForPodE(t, nil, "details")
require.Nil(t, err)
// Check `bootstrap` for expected Envoy bootstrap configuration
// ...
```
