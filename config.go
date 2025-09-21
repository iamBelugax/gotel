package gotel

import (
	"maps"
	"time"
)

// Config holds the configuration settings for the OpenTelemetry provider.
type Config struct {
	Service       *ServiceInfo   // Contains metadata about the application.
	OTLP          *OTLPConfig    // Contains configuration for the OTLP exporter.
	Tracing       *TracingConfig // Contains configuration for distributed tracing.
	ResourceAttrs map[string]any // Additional key-value pairs describing the resource emitting telemetry.
	Debug         bool           // Enables stdout exporters for all signals printing telemetry to the console.
}

// OTLPConfig defines settings for the OpenTelemetry Protocol (OTLP) exporter.
type OTLPConfig struct {
	Endpoint      string            // Endpoint is the target gRPC OTLP collector endpoint.
	Headers       map[string]string // Headers specifies additional request headers sent with OTLP requests.
	ExportTimeout time.Duration     // ExportTimeout is the maximum allowed duration for an export operation.
	BatchTimeout  time.Duration     // BatchTimeout is the maximum time the exporter waits before sending a batch.
}

// TracingConfig defines sampling and tracing-related settings.
type TracingConfig struct {
	SamplingRatio float64 // SamplingRatio determines the fraction of traces to record (1.0 = always, 0.0 = never).
}

// ServiceInfo holds identifying information about the instrumented service.
type ServiceInfo struct {
	Name        string // Name is the required unique identifier for the service.
	Version     string // Version specifies the semantic version of the service.
	Environment string // Environment categorizes the deployment.
}

// DefaultConfig returns a Config struct pre populated with sensible default
// values for development environment.
func DefaultConfig(serviceName string) *Config {
	return &Config{
		Debug: false,
		Service: &ServiceInfo{
			Name:        serviceName,
			Version:     "1.0.0",
			Environment: "development",
		},
		OTLP: &OTLPConfig{
			Endpoint:      "localhost:4317",
			BatchTimeout:  5 * time.Second,
			ExportTimeout: 30 * time.Second,
			Headers:       make(map[string]string),
		},
		Tracing: &TracingConfig{
			SamplingRatio: 1.0,
		},
		ResourceAttrs: make(map[string]any),
	}
}

// WithServiceInfo configures the core service identification attributes.
func (c *Config) WithServiceInfo(name, version, environment string) *Config {
	c.Service.Name = name
	c.Service.Version = version
	c.Service.Environment = environment
	return c
}

// WithOTLPEndpoint configures the OTLP exporter's target endpoint.
func (c *Config) WithOTLPEndpoint(endpoint string) *Config {
	c.OTLP.Endpoint = endpoint
	return c
}

// WithOTLPHeaders configures additional headers for OTLP requests.
func (c *Config) WithOTLPHeaders(headers map[string]string) *Config {
	maps.Copy(c.OTLP.Headers, headers)
	return c
}

// WithExportTimeout sets the maximum allowed duration for an OTLP export operation.
func (c *Config) WithExportTimeout(timeout time.Duration) *Config {
	c.OTLP.ExportTimeout = timeout
	return c
}

// WithBatchTimeout sets the maximum wait time before an OTLP batch is sent.
func (c *Config) WithBatchTimeout(timeout time.Duration) *Config {
	c.OTLP.BatchTimeout = timeout
	return c
}

// WithSamplingRatio sets the fraction of traces to sample (1.0 = always, 0.0 = never).
func (c *Config) WithSamplingRatio(ratio float64) *Config {
	if ratio < 0.0 {
		ratio = 0.0
	}
	if ratio > 1.0 {
		ratio = 1.0
	}
	c.Tracing.SamplingRatio = ratio
	return c
}

// WithResourceAttr adds or updates a resource attribute (key-value pair).
func (c *Config) WithResourceAttr(key string, value any) *Config {
	c.ResourceAttrs[key] = value
	return c
}

// WithResourceAttrs bulk-sets resource attributes.
func (c *Config) WithResourceAttrs(attrs map[string]any) *Config {
	maps.Copy(c.ResourceAttrs, attrs)
	return c
}

// WithDebug enables or disables debug mode (stdout exporters).
func (c *Config) WithDebug(debug bool) *Config {
	c.Debug = debug
	return c
}
