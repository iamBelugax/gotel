package gotel

import (
	"maps"
	"time"

	"google.golang.org/grpc/credentials"
)

// config holds the configuration settings for the OpenTelemetry provider.
type config struct {
	// Service contains identifying information about the running service,
	Service *ServiceInfo

	// Exporter specifies configuration for the OTLP exporter, including
	// endpoint, headers, timeouts and batch settings.
	Exporter *ExporterConfig

	// Tracing configures distributed tracing, including sampling ratio and
	// trace propagation behavior.
	Tracing *TracingConfig

	// Logging configures application logging, including log level and
	// integration with OpenTelemetry logging.
	Logging *LoggingConfig

	// Security holds TLS or insecure transport credentials used to connect
	// to an OTLP collector.
	Security *SecurityConfig

	// ResourceAttrs contains additional custom key-value attributes describing
	// the resource emitting telemetry.
	ResourceAttrs map[string]any

	// Debug, when true, enables stdout exporters for tracing, metrics, and logs.
	Debug bool
}

type ExporterConfig struct {
	Endpoint      string
	Headers       map[string]string
	ExportTimeout time.Duration
	BatchTimeout  time.Duration
}

type TracingConfig struct {
	SamplingRatio float64 // (1.0 = always, 0.0 = never).
}

type LoggingConfig struct {
	Level string
}

type ServiceInfo struct {
	Name        string
	Version     string
	Environment string
}

type SecurityConfig struct {
	Insecure       bool
	TLSCredentials credentials.TransportCredentials
}

type Option func(*config)

// DefaultConfig returns a Config struct pre populated with sensible default
// values for a development environment.
func DefaultConfig(opts ...Option) *config {
	conf := &config{
		Debug: false,
		Service: &ServiceInfo{
			Name:        "unknown",
			Version:     "1.0.0",
			Environment: "development",
		},
		ResourceAttrs: make(map[string]any),
		Security:      &SecurityConfig{Insecure: true},
		Tracing:       &TracingConfig{SamplingRatio: 1.0},
		Logging:       &LoggingConfig{Level: "debug"},
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

// WithLogLevel sets the minimum logging level.
func WithLogLevel(level string) Option {
	return func(c *config) {
		c.Logging.Level = level
	}
}
