package gotel

import (
	"maps"
	"time"

	"google.golang.org/grpc/credentials"
)

// config holds the configuration settings for the OpenTelemetry provider.
// It defines how telemetry is exported, sampled, secured, and annotated.
type config struct {
	Service       *ServiceInfo    // Contains metadata about the service.
	Exporter      *ExporterConfig // Configuration for the OTLP exporter.
	Tracing       *TracingConfig  // Configuration for distributed tracing.
	Security      *SecurityConfig // Security related settings.
	ResourceAttrs map[string]any  // Additional key-value pairs describing the resource emitting telemetry.
	Debug         bool            // Enables stdout exporters for all signals, printing telemetry to the console.
}

// ExporterConfig defines settings for the OpenTelemetry Protocol (OTLP) exporter.
type ExporterConfig struct {
	Endpoint      string            // Target gRPC OTLP collector endpoint.
	Headers       map[string]string // Additional request headers sent with OTLP requests.
	ExportTimeout time.Duration     // Maximum allowed duration for an export operation.
	BatchTimeout  time.Duration     // Maximum wait time before the exporter sends a batch.
}

// TracingConfig defines sampling and tracing-related settings.
type TracingConfig struct {
	SamplingRatio float64 // Fraction of traces to record (1.0 = always, 0.0 = never).
}

// ServiceInfo holds identifying information about the instrumented service.
type ServiceInfo struct {
	Name        string // Unique identifier for the service.
	Version     string // Semantic version of the service.
	Environment string // Deployment environment (e.g. "development", "staging", "production").
}

// SecurityConfig configures authentication and encryption for telemetry transmission.
type SecurityConfig struct {
	Insecure       bool                             // If true, skips TLS verification.
	TLSCredentials credentials.TransportCredentials // Transport credentials for secure gRPC connections to collectors.
}

// An Option is a function that modifies a Config struct.
type Option func(*config)

// DefaultConfig returns a Config struct pre populated with sensible default
// values for a development environment, then applies the provided options.
func DefaultConfig(serviceName string, opts ...Option) *config {
	conf := &config{
		Debug: false,
		Service: &ServiceInfo{
			Name:        serviceName,
			Version:     "1.0.0",
			Environment: "development",
		},
		ResourceAttrs: make(map[string]any),
		Security:      &SecurityConfig{Insecure: true},
		Tracing:       &TracingConfig{SamplingRatio: 1.0},
		Exporter: &ExporterConfig{
			Endpoint:      "localhost:4317",
			BatchTimeout:  5 * time.Second,
			ExportTimeout: 30 * time.Second,
			Headers:       make(map[string]string),
		},
	}

	for _, opt := range opts {
		opt(conf)
	}

	return conf
}

// WithServiceInfo configures the core service identification attributes.
func WithServiceInfo(name, version, environment string) Option {
	return func(c *config) {
		c.Service.Name = name
		c.Service.Version = version
		c.Service.Environment = environment
	}
}

// WithEndpoint sets the OTLP exporter's target gRPC endpoint.
func WithEndpoint(endpoint string) Option {
	return func(c *config) {
		c.Exporter.Endpoint = endpoint
	}
}

// WithHeader adds additional header to OTLP requests.
func WithHeader(key string, val string) Option {
	return func(c *config) {
		c.Exporter.Headers[key] = val
	}
}

// WithHeaders adds additional headers to OTLP requests.
func WithHeaders(headers map[string]string) Option {
	return func(c *config) {
		maps.Copy(c.Exporter.Headers, headers)
	}
}

// WithExportTimeout sets the maximum allowed duration for an OTLP export operation.
func WithExportTimeout(timeout time.Duration) Option {
	return func(c *config) {
		c.Exporter.ExportTimeout = timeout
	}
}

// WithBatchTimeout sets the maximum wait time before an OTLP batch is sent.
func WithBatchTimeout(timeout time.Duration) Option {
	return func(c *config) {
		c.Exporter.BatchTimeout = timeout
	}
}

// WithSamplingRatio sets the fraction of traces to sample.
// Values are clamped between 0.0 (never sample) and 1.0 (always sample).
func WithSamplingRatio(ratio float64) Option {
	return func(c *config) {
		if ratio < 0.0 {
			ratio = 0.0
		}
		if ratio > 1.0 {
			ratio = 1.0
		}
		c.Tracing.SamplingRatio = ratio
	}
}

// WithResourceAttr adds or updates a single resource attribute (key-value pair).
func WithResourceAttr(key string, value any) Option {
	return func(c *config) {
		c.ResourceAttrs[key] = value
	}
}

// WithResourceAttrs bulk-adds or updates multiple resource attributes.
func WithResourceAttrs(attrs map[string]any) Option {
	return func(c *config) {
		maps.Copy(c.ResourceAttrs, attrs)
	}
}

// WithDebug enables or disables debug mode.
// When enabled, telemetry is also printed to stdout using OTLP stdout exporters.
func WithDebug(debug bool) Option {
	return func(c *config) {
		c.Debug = debug
	}
}

// WithInsecure enables or disables insecure mode (skips TLS verification).
func WithInsecure(insecure bool) Option {
	return func(c *config) {
		c.Security.Insecure = insecure
		c.Security.TLSCredentials = nil
	}
}

// WithTLSCredentials sets explicit gRPC transport credentials for secure telemetry transmission.
func WithTLSCredentials(creds credentials.TransportCredentials) Option {
	return func(c *config) {
		c.Security.Insecure = false
		c.Security.TLSCredentials = creds
	}
}
