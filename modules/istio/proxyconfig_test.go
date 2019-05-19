package istio

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Integration test pre-requisites:
// - Local Kubernetes (KinD, minikube, K3s etc.).
// - Istio BookInfo sample deployed to default namespace.

func TestGetBootstrapConfigDumpForPod(t *testing.T) {
	t.Parallel()

	bootstrap, err := GetBootstrapConfigDumpForPodE(t, nil, "nonexistent")
	assert.Nil(t, bootstrap)
	assert.NotNil(t, err)

	bootstrap = GetBootstrapConfigDumpForPod(t, nil, getBookInfoDetailsPod(t).Name)
	assert.NotNil(t, bootstrap)
	assert.NotNil(t, bootstrap.Bootstrap)
	assert.NotNil(t, bootstrap.Bootstrap.Node)
}

func TestGetClustersConfigDumpForPod(t *testing.T) {
	t.Parallel()

	clusters, err := GetClustersConfigDumpForPodE(t, nil, "nonexistent")
	assert.Nil(t, clusters)
	assert.NotNil(t, err)

	clusters = GetClustersConfigDumpForPod(t, nil, getBookInfoDetailsPod(t).Name)
	assert.NotNil(t, clusters)
	assert.True(t, len(clusters.DynamicActiveClusters)+len(clusters.StaticClusters) > 0)
}

func TestGetListenersConfigDumpForPod(t *testing.T) {
	t.Parallel()

	listeners, err := GetListenersConfigDumpForPodE(t, nil, "nonexistent")
	assert.Nil(t, listeners)
	assert.NotNil(t, err)

	listeners = GetListenersConfigDumpForPod(t, nil, getBookInfoDetailsPod(t).Name)
	assert.NotNil(t, listeners)
	assert.True(t, len(listeners.DynamicActiveListeners)+len(listeners.StaticListeners) > 0)
}

func TestGetRoutesConfigDumpForPod(t *testing.T) {
	t.Parallel()

	routes, err := GetRoutesConfigDumpForPodE(t, nil, "nonexistent")
	assert.Nil(t, routes)
	assert.NotNil(t, err)

	routes = GetRoutesConfigDumpForPod(t, nil, getBookInfoDetailsPod(t).Name)
	assert.NotNil(t, routes)
	assert.True(t, len(routes.DynamicRouteConfigs)+len(routes.StaticRouteConfigs) > 0)
}

func TestIsClustersConfigClusteredTo(t *testing.T) {
	t.Parallel()

	clusters, err := GetClustersConfigDumpForPodE(t, nil, getBookInfoDetailsPod(t).Name)
	require.Nil(t, err)

	// Istio BookInfo 'details-v1' deployment should result in pods with an
	// Envoy sidecar that has a:
	// - static cluster inbound on http@details.default.svc.cluster.local:9080
	// - dynamic cluster outbound to *@reviews.default.svc.cluster.local:9080
	// - dynamic cluster outbound to *@istio-policy.istio-system.svc.cluster.local:15014

	var tests = []struct {
		assert    bool
		FQDN      string
		subset    string
		direction string
		port      int
	}{
		{true, "details.default.svc.cluster.local", "http", "inbound", 9080},
		{true, "details.default.svc.cluster.local", "", "inbound", 9080},
		{true, "details.default.svc.cluster.local", "http", "", 9080},
		{true, "details.default.svc.cluster.local", "http", "inbound", 0},
		{true, "details.default.svc.cluster.local", "", "", 0},
		{true, "", "", "", 9080},
		{true, "", "", "", 0},
		{false, "", "", "", 444},
		{false, "nonexistent.default.svc.cluster.local", "", "", 0},
		{true, "reviews.default.svc.cluster.local", "", "outbound", 9080},
		{true, "reviews.default.svc.cluster.local", "", "", 9080},
		{true, "reviews.default.svc.cluster.local", "", "outbound", 0},
		{false, "nonexistent.default.svc.cluster.local", "", "outbound", 9080},
		{true, "istio-policy.istio-system.svc.cluster.local", "", "outbound", 15014},
		{true, "istio-policy.istio-system.svc.cluster.local", "", "", 15014},
		{true, "istio-policy.istio-system.svc.cluster.local", "", "outbound", 0},
	}
	for _, test := range tests {
		t.Run(
			fmt.Sprintf("%s:%s@%s:%s",
				wildcardForZeroValue(test.direction),
				wildcardForZeroValue(test.subset),
				wildcardForZeroValue(test.FQDN),
				wildcardForZeroValue(strconv.Itoa(test.port)),
			),
			func(t *testing.T) {
				if test.assert {
					assert.True(t, IsClustersConfigClusteredTo(
						clusters, test.FQDN, test.subset, test.direction, test.port))
				} else {
					assert.False(t, IsClustersConfigClusteredTo(
						clusters, test.FQDN, test.subset, test.direction, test.port))
				}
			},
		)
	}
}

// TestIsListenersConfigListeningOn ...
func TestIsListenersConfigListeningOn(t *testing.T) {
	t.Parallel()

	listeners, err := GetListenersConfigDumpForPodE(t, nil, getBookInfoDetailsPod(t).Name)
	require.Nil(t, err)

	// Istio BookInfo 'details-v1' deployment should result in pods with an
	// Envoy sidecar listening on: HTTP@0.0.0.0:9080
	var tests = []struct {
		assert       bool
		listenerType string
		listenerAddr string
		listenerPort int
	}{
		{true, "HTTP", "", 0},
		{true, "", "0.0.0.0", 0},
		{true, "", "", 9080},
		{true, "HTTP", "0.0.0.0", 9080},
		{false, "TCP", "0.0.0.0", 9080},
		{false, "UDP", "0.0.0.0", 9080},
		{false, "HTTP", "127.0.0.1", 9080},
		{false, "HTTP", "0.0.0.0", 1234},
	}
	for _, test := range tests {
		t.Run(
			fmt.Sprintf("%s@%s:%s",
				wildcardForZeroValue(test.listenerType),
				wildcardForZeroValue(test.listenerAddr),
				wildcardForZeroValue(strconv.Itoa(test.listenerPort)),
			),
			func(t *testing.T) {
				if test.assert {
					assert.True(t, IsListenersConfigListeningOn(
						listeners, test.listenerType, test.listenerAddr, test.listenerPort))
				} else {
					assert.False(t, IsListenersConfigListeningOn(
						listeners, test.listenerType, test.listenerAddr, test.listenerPort))
				}
			},
		)
	}
}

func TestIsRoutesConfigRoutingTo(t *testing.T) {
	t.Parallel()

	routes, err := GetRoutesConfigDumpForPodE(t, nil, getBookInfoDetailsPod(t).Name)
	require.Nil(t, err)

	// Istio BookInfo 'details-v1' deployment should result in pods with an
	// Envoy sidecar routing to:
	// - reviews.default.svc.cluster.local:9080
	// - backend
	var tests = []struct {
		assert bool
		host   string
		port   int
	}{
		{true, "reviews.default.svc.cluster.local", 9080},
		{false, "reviews.default.svc.cluster.local", 0},
		{false, "nonexistent.default.svc.cluster.local", 0},
		{false, "nonexistent.default.svc.cluster.local", 9080},
		{true, "backend", 0},
	}
	for _, test := range tests {
		t.Run(
			fmt.Sprintf("%s:%s",
				wildcardForZeroValue(test.host),
				wildcardForZeroValue(strconv.Itoa(test.port)),
			),
			func(t *testing.T) {
				if test.assert {
					assert.True(t, IsRoutesConfigRoutingTo(routes, test.host, test.port))
				} else {
					assert.False(t, IsRoutesConfigRoutingTo(routes, test.host, test.port))
				}
			},
		)
	}
}

func wildcardForZeroValue(in string) string {
	if in == "" || in == "0" {
		in = "*"
	}
	return in
}

func getBookInfoDetailsPod(t *testing.T) *corev1.Pod {
	t.Helper()
	o := k8s.NewKubectlOptions("", "")
	detailsPods := k8s.ListPods(t, o, metav1.ListOptions{LabelSelector: "app=details"})
	require.Equal(t, len(detailsPods), 1)
	return &detailsPods[0]
}
