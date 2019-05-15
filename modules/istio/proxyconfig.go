package istio

import (
	"fmt"
	"testing"

	"github.com/gruntwork-io/terratest/modules/logger"

	"github.com/stretchr/testify/require"

	"istio.io/istio/istioctl/pkg/kubernetes"
	"istio.io/istio/istioctl/pkg/util/configdump"
	"istio.io/istio/pilot/pkg/model"

	adminapi "github.com/envoyproxy/go-control-plane/envoy/admin/v2alpha"
	api "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	wconfigdump "istio.io/istio/istioctl/pkg/writer/envoy/configdump"
)

// GetBootstrapConfigDumpForPod queries the pod's Envoy sidecar for currently
// configured bootstrap. If anything goes wrong, the function will
// automatically fail the test.
// NOTE: https://www.envoyproxy.io/docs/envoy/latest/api-v2/admin/v2alpha/config_dump.proto#admin-v2alpha-bootstrapconfigdump
func GetBootstrapConfigDumpForPod(
	t *testing.T,
	options *Options,
	pod string,
) *adminapi.BootstrapConfigDump {
	d, err := GetBootstrapConfigDumpForPodE(t, options, pod)
	require.NoError(t, err)
	return d
}

// GetBootstrapConfigDumpForPodE queries the pod's Envoy sidecar for currently
// configured bootstrap.
// NOTE: https://www.envoyproxy.io/docs/envoy/latest/api-v2/admin/v2alpha/config_dump.proto#admin-v2alpha-bootstrapconfigdump
func GetBootstrapConfigDumpForPodE(
	t *testing.T,
	options *Options,
	pod string,
) (
	*adminapi.BootstrapConfigDump, error) {
	cw, err := configDumpForPod(t, options, pod)
	if err != nil {
		return nil, err
	}

	return cw.GetBootstrapConfigDump()
}

// GetClustersConfigDumpForPod queries the pod's Envoy sidecar for currently
// configured clusters. If anything goes wrong, the function will
// automatically fail the test.
// NOTE: https://www.envoyproxy.io/docs/envoy/latest/api-v2/admin/v2alpha/config_dump.proto#admin-v2alpha-clustersconfigdump
func GetClustersConfigDumpForPod(
	t *testing.T,
	options *Options,
	pod string,
) *adminapi.ClustersConfigDump {
	d, err := GetClustersConfigDumpForPodE(t, options, pod)
	require.NoError(t, err)
	return d
}

// GetClustersConfigDumpForPodE queries the pod's Envoy sidecar for currently
// configured clusters.
// NOTE: https://www.envoyproxy.io/docs/envoy/latest/api-v2/admin/v2alpha/config_dump.proto#admin-v2alpha-clustersconfigdump
func GetClustersConfigDumpForPodE(
	t *testing.T,
	options *Options,
	pod string,
) (
	*adminapi.ClustersConfigDump, error) {
	cw, err := configDumpForPod(t, options, pod)
	if err != nil {
		return nil, err
	}

	return cw.GetClusterConfigDump()
}

// GetListenersConfigDumpForPod queries the pod's Envoy sidecar for currently
// configured listeners. If anything goes wrong, the function will
// automatically fail the test.
// NOTE: https://www.envoyproxy.io/docs/envoy/latest/api-v2/admin/v2alpha/config_dump.proto#admin-v2alpha-listenersconfigdump
func GetListenersConfigDumpForPod(
	t *testing.T,
	options *Options,
	pod string,
) *adminapi.ListenersConfigDump {
	d, err := GetListenersConfigDumpForPodE(t, options, pod)
	require.NoError(t, err)
	return d
}

// GetListenersConfigDumpForPodE queries the pod's Envoy sidecar for currently
// configured listeners.
// NOTE: https://www.envoyproxy.io/docs/envoy/latest/api-v2/admin/v2alpha/config_dump.proto#admin-v2alpha-listenersconfigdump
func GetListenersConfigDumpForPodE(
	t *testing.T,
	options *Options,
	pod string,
) (
	*adminapi.ListenersConfigDump, error) {
	cw, err := configDumpForPod(t, options, pod)
	if err != nil {
		return nil, err
	}

	return cw.GetListenerConfigDump()
}

// GetRoutesConfigDumpForPod queries the pod's Envoy sidecar for currently
// configured routes. If anything goes wrong, the function will automatically
// fail the test.
// NOTE: https://www.envoyproxy.io/docs/envoy/latest/api-v2/admin/v2alpha/config_dump.proto#admin-v2alpha-routesconfigdump
func GetRoutesConfigDumpForPod(
	t *testing.T,
	options *Options,
	pod string,
) *adminapi.RoutesConfigDump {
	d, err := GetRoutesConfigDumpForPodE(t, options, pod)
	require.NoError(t, err)
	return d
}

// GetRoutesConfigDumpForPodE queries the pod's Envoy sidecar for currently
// configured routes.
// NOTE: https://www.envoyproxy.io/docs/envoy/latest/api-v2/admin/v2alpha/config_dump.proto#admin-v2alpha-routesconfigdump
func GetRoutesConfigDumpForPodE(
	t *testing.T,
	options *Options,
	pod string,
) (
	*adminapi.RoutesConfigDump, error) {
	cw, err := configDumpForPod(t, options, pod)
	if err != nil {
		return nil, err
	}

	return cw.GetRouteConfigDump()
}

// IsClustersConfigClusteredTo returns true if the Envoy config has a cluster
// that matches the specific fqdn, subset, direction and port combination. The
// zero value is treated like a wildcard.
func IsClustersConfigClusteredTo(
	config *adminapi.ClustersConfigDump,
	fqdn, subset, direction string,
	port int,
) bool {
	filter := &wconfigdump.ClusterFilter{
		FQDN:      model.Hostname(fqdn),
		Subset:    subset,
		Direction: model.TrafficDirection(direction),
		Port:      port,
	}

	for _, cluster := range config.StaticClusters {
		if filter.Verify(cluster.Cluster) {
			return true
		}
	}

	for _, cluster := range config.DynamicActiveClusters {
		if filter.Verify(cluster.Cluster) {
			return true
		}
	}

	return false
}

// IsListenersConfigListeningOn returns true if the Envoy config is listening
// on the specific protocol type, address and port combination. The zero value
// is treated like a wildcard.
func IsListenersConfigListeningOn(
	config *adminapi.ListenersConfigDump,
	listenerType, listenerAddr string,
	listenerPort int,
) bool {
	filter := &wconfigdump.ListenerFilter{
		Type:    listenerType,
		Address: listenerAddr,
		Port:    uint32(listenerPort),
	}

	for _, listener := range config.StaticListeners {
		if filter.Verify(listener.Listener) {
			return true
		}
	}

	for _, listener := range config.DynamicActiveListeners {
		if filter.Verify(listener.Listener) {
			return true
		}
	}

	return false
}

// IsRoutesConfigRoutingTo returns true if the Envoy config has a route
// that involves the specific host and port combination. The more identifiable
// virtual host name, rather than route name, is used for comparison. See:
// https://www.envoyproxy.io/docs/envoy/latest/api-v2/api/v2/route/route.proto#envoy-api-msg-route-virtualhost
func IsRoutesConfigRoutingTo(
	config *adminapi.RoutesConfigDump,
	host string,
	port int,
) bool {
	var vhostName string
	if port > 0 {
		vhostName = fmt.Sprintf("%s:%d", host, port)
	} else {
		vhostName = host
	}

	for _, route := range config.StaticRouteConfigs {
		if routeConfigVHostNameMatches(route.RouteConfig, vhostName) {
			return true
		}
	}

	for _, route := range config.DynamicRouteConfigs {
		if routeConfigVHostNameMatches(route.RouteConfig, vhostName) {
			return true
		}
	}

	return false
}

// routeConfigVHostNameMatches returns true when the name parameter is found in
// the route config.
func routeConfigVHostNameMatches(routeConfig *api.RouteConfiguration, vhostName string) bool {
	if routeConfig != nil {
		for _, vhost := range routeConfig.VirtualHosts {
			if vhost.Name == vhostName {
				return true
			}
		}
	}
	return false
}

// configDumpForPod dumps the Istio Envoy proxy configuration for a pod.
func configDumpForPod(t *testing.T, options *Options, pod string) (*configdump.Wrapper, error) {
	t.Helper()

	if options == nil {
		options = NewOptions("", "")
	}

	kubeClient, err := kubernetes.NewClient(options.ConfigPath, options.ContextName)
	if err != nil {
		return nil, err
	}

	logger.Logf(t, "Gathering proxy config dump from Envoy sidecar of pod: %s", pod)
	b, err := kubeClient.EnvoyDo(pod, options.Namespace, "GET", "config_dump", nil)
	if err != nil {
		return nil, err
	}

	cw := &configdump.Wrapper{}
	if err := cw.UnmarshalJSON(b); err != nil {
		return nil, err
	}
	return cw, nil
}
