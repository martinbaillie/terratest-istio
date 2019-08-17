package istio

import (
	"fmt"
	"os"
	"testing"

	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/stretchr/testify/require"

	"istio.io/istio/istioctl/pkg/kubernetes"
	"istio.io/istio/istioctl/pkg/util/configdump"
	"istio.io/istio/istioctl/pkg/writer/compare"
	"istio.io/istio/pilot/pkg/model"

	adminv2 "github.com/envoyproxy/go-control-plane/envoy/admin/v2alpha"
	apiv2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	wconfigdump "istio.io/istio/istioctl/pkg/writer/envoy/configdump"
)

// GetBootstrapConfigDumpForPod queries the pod's Envoy sidecar for currently
// configured bootstrap. If anything goes wrong, the function will
// automatically fail the test.
// NOTE:
// https://www.envoyproxy.io/docs/envoy/latest/api-v2/admin/v2alpha/config_dump.proto#admin-v2alpha-bootstrapconfigdump
func GetBootstrapConfigDumpForPod(
	t *testing.T, o *Options, pod string) *adminv2.BootstrapConfigDump {
	d, err := GetBootstrapConfigDumpForPodE(t, o, pod)
	require.NoError(t, err)
	return d
}

// GetBootstrapConfigDumpForPodE queries the pod's Envoy sidecar for currently
// configured bootstrap.
// NOTE:
// https://www.envoyproxy.io/docs/envoy/latest/api-v2/admin/v2alpha/config_dump.proto#admin-v2alpha-bootstrapconfigdump
func GetBootstrapConfigDumpForPodE(
	t *testing.T, o *Options, pod string) (*adminv2.BootstrapConfigDump, error) {
	cd, err := configDumpForPod(t, o, pod)
	if err != nil {
		return nil, err
	}

	return cd.GetBootstrapConfigDump()
}

// GetClustersConfigDumpForPod queries the pod's Envoy sidecar for currently
// configured clusters. If anything goes wrong, the function will
// automatically fail the test.
// NOTE: https://www.envoyproxy.io/docs/envoy/latest/api-v2/admin/v2alpha/config_dump.proto#admin-v2alpha-clustersconfigdump
func GetClustersConfigDumpForPod(t *testing.T, o *Options, pod string) *adminv2.ClustersConfigDump {
	d, err := GetClustersConfigDumpForPodE(t, o, pod)
	require.NoError(t, err)
	return d
}

// GetClustersConfigDumpForPodE queries the pod's Envoy sidecar for currently
// configured clusters.
// NOTE:
// https://www.envoyproxy.io/docs/envoy/latest/api-v2/admin/v2alpha/config_dump.proto#admin-v2alpha-clustersconfigdump
func GetClustersConfigDumpForPodE(
	t *testing.T, o *Options, pod string) (*adminv2.ClustersConfigDump, error) {
	cd, err := configDumpForPod(t, o, pod)
	if err != nil {
		return nil, err
	}

	return cd.GetClusterConfigDump()
}

// GetListenersConfigDumpForPod queries the pod's Envoy sidecar for currently
// configured listeners. If anything goes wrong, the function will
// automatically fail the test.
// NOTE: https://www.envoyproxy.io/docs/envoy/latest/api-v2/admin/v2alpha/config_dump.proto#admin-v2alpha-listenersconfigdump
func GetListenersConfigDumpForPod(
	t *testing.T, o *Options, pod string) *adminv2.ListenersConfigDump {
	d, err := GetListenersConfigDumpForPodE(t, o, pod)
	require.NoError(t, err)
	return d
}

// GetListenersConfigDumpForPodE queries the pod's Envoy sidecar for currently
// configured listeners.
// NOTE:
// https://www.envoyproxy.io/docs/envoy/latest/api-v2/admin/v2alpha/config_dump.proto#admin-v2alpha-listenersconfigdump
func GetListenersConfigDumpForPodE(
	t *testing.T, o *Options, pod string) (*adminv2.ListenersConfigDump, error) {
	cd, err := configDumpForPod(t, o, pod)
	if err != nil {
		return nil, err
	}

	return cd.GetListenerConfigDump()
}

// GetRoutesConfigDumpForPod queries the pod's Envoy sidecar for currently
// configured routes. If anything goes wrong, the function will automatically
// fail the test.
// NOTE:
// https://www.envoyproxy.io/docs/envoy/latest/api-v2/admin/v2alpha/config_dump.proto#admin-v2alpha-routesconfigdump
func GetRoutesConfigDumpForPod(t *testing.T, o *Options, pod string) *adminv2.RoutesConfigDump {
	d, err := GetRoutesConfigDumpForPodE(t, o, pod)
	require.NoError(t, err)
	return d
}

// GetRoutesConfigDumpForPodE queries the pod's Envoy sidecar for currently
// configured routes.
// NOTE:
// https://www.envoyproxy.io/docs/envoy/latest/api-v2/admin/v2alpha/config_dump.proto#admin-v2alpha-routesconfigdump
func GetRoutesConfigDumpForPodE(
	t *testing.T, o *Options, pod string) (*adminv2.RoutesConfigDump, error) {
	cd, err := configDumpForPod(t, o, pod)
	if err != nil {
		return nil, err
	}

	return cd.GetRouteConfigDump()
}

// IsClustersConfigClusteredTo returns true if the Envoy config has a cluster
// that matches the specific fqdn, subset, direction and port combination. The
// zero value is treated like a wildcard.
func IsClustersConfigClusteredTo(
	config *adminv2.ClustersConfigDump, fqdn, subset, direction string, port int) bool {
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
	config *adminv2.ListenersConfigDump, listenerType, listenerAddr string, listenerPort int) bool {
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
// virtual host name, rather than route name, is used for comparison.
// NOTE:
// https://www.envoyproxy.io/docs/envoy/latest/api-v2/api/v2/route/route.proto#envoy-api-msg-route-virtualhost
func IsRoutesConfigRoutingTo(config *adminv2.RoutesConfigDump, host string, port int) bool {
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
func routeConfigVHostNameMatches(routeConfig *apiv2.RouteConfiguration, vhostName string) bool {
	if routeConfig != nil {
		for _, vhost := range routeConfig.VirtualHosts {
			if vhost.Name == vhostName {
				return true
			}
		}
	}
	return false
}

// ArePilotsSyncedToPod diffs the Istio Pilot(s) view of an Envoy proxy's
// configuration with the actual configuration exhibited by the proxy. The
// function returns true if everything is in sync, false if not, and uses
// stdout to display differences. If anything goes wrong during comparison
// it will automatically fail the test.
func ArePilotsSyncedToPod(t *testing.T, o *Options, pod string) bool {
	d, err := ArePilotsSyncedToPodE(t, o, pod)
	require.NoError(t, err)
	return d
}

// ArePilotsSyncedToPodE diffs the Istio Pilot(s) view of an Envoy proxy's
// configuration with the actual configuration exhibited by the proxy. The
// function returns true if everything is in sync, false if not, and uses
// stdout to display differences.
func ArePilotsSyncedToPodE(t *testing.T, o *Options, pod string) (bool, error) {
	proxyBytes, err := configBytesFromProxy(t, o, pod)
	if err != nil {
		return false, err
	}

	pilotBytes, err := configBytesFromPilots(t, o, pod)
	if err != nil {
		return false, err
	}

	c, err := compare.NewComparator(os.Stdout, pilotBytes, proxyBytes)
	if err != nil {
		return false, err
	}

	err = c.Diff()
	return err == nil, err
}

// configDumpForPod gets the unmarshaled JSON configuration for a pod.
func configDumpForPod(t *testing.T, o *Options, pod string) (*configdump.Wrapper, error) {
	t.Helper()

	b, err := configBytesFromProxy(t, o, pod)
	if err != nil {
		return nil, err
	}

	cd := &configdump.Wrapper{}
	if err := cd.UnmarshalJSON(b); err != nil {
		return nil, err
	}
	return cd, nil
}

// configBytesFromProxy dumps the Istio Envoy proxy configuration from a pod.
func configBytesFromProxy(t *testing.T, o *Options, pod string) ([]byte, error) {
	t.Helper()

	if o == nil {
		o = NewOptions("", "")
	}

	kubeClient, err := kubernetes.NewClient(o.ConfigPath, o.ContextName)
	if err != nil {
		return nil, err
	}

	logger.Logf(t, "Gathering proxy config from Envoy sidecar of pod: %s", pod)
	return kubeClient.EnvoyDo(pod, o.Namespace, "GET", "config_dump", nil)
}

// configBytesFromPilots dumps the Istio Envoy proxy configuration for
// a pod from the perspective of the Istio Pilots.
func configBytesFromPilots(t *testing.T, o *Options, pod string) (
	map[string][]byte, error) {
	t.Helper()

	if o == nil {
		o = NewOptions("", "")
	}

	kubeClient, err := kubernetes.NewClient(o.ConfigPath, o.ContextName)
	if err != nil {
		return nil, err
	}

	logger.Logf(t, "Gathering proxy config from Pilot(s) for pod: %s", pod)
	return kubeClient.AllPilotsDiscoveryDo(
		o.IstioNamespace,
		"GET",
		fmt.Sprintf("/debug/config_dump?proxyID=%s.%s", pod, o.Namespace),
		nil)
}
