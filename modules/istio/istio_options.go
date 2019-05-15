package istio

// Options are common options necessary to specify for all Istio calls.
type Options struct {
	ContextName string
	ConfigPath  string
	Namespace   string
}

// NewOptions returns an Options configured based on the provided parameters.
func NewOptions(contextName, configPath string) *Options {
	return &Options{
		ContextName: contextName,
		ConfigPath:  configPath,
		Namespace:   "default",
	}
}
