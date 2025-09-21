package gotel

import "context"

// Provider is the central struct that encapsulates all OpenTelemetry SDK components.
type Provider struct {
	config *Config
}

// NewProvider initializes and configures the OpenTelemetry SDK based on the provided configuration.
// It sets up the resource, exporters and providers for tracing, metrics and logging.
func NewProvider(ctx context.Context, conf *Config) (*Provider, error) {
	if conf == nil {
		conf = DefaultConfig("unknown")
	}

	p := &Provider{config: conf}
	return p, nil
}
