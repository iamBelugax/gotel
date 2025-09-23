package gotel

import (
	"maps"
	"time"

	"google.golang.org/grpc/credentials"
)

// Config holds the configuration settings for the OpenTelemetry provider.
// It defines how telemetry is exported, sampled, secured, and annotated.
type Config struct {
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

// DefaultConfig returns a Config struct pre populated with sensible default
// values for a development environment.
func DefaultConfig(serviceName string) *Config {
	return &Config{
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
}

// WithServiceInfo configures the core service identification attributes.
func (c *Config) WithServiceInfo(name, version, environment string) *Config {
	c.Service.Name = name
	c.Service.Version = version
	c.Service.Environment = environment
	return c
}

// WithEndpoint sets the OTLP exporter's target gRPC endpoint.
func (c *Config) WithEndpoint(endpoint string) *Config {
	c.Exporter.Endpoint = endpoint
	return c
}

// WithHeader adds additional header to OTLP requests.
func (c *Config) WithHeader(key string, val string) *Config {
	c.Exporter.Headers[key] = val
	return c
}

// WithHeaders adds additional headers to OTLP requests.
func (c *Config) WithHeaders(headers map[string]string) *Config {
	maps.Copy(c.Exporter.Headers, headers)
	return c
}

// WithExportTimeout sets the maximum allowed duration for an OTLP export operation.
func (c *Config) WithExportTimeout(timeout time.Duration) *Config {
	c.Exporter.ExportTimeout = timeout
	return c
}

// WithBatchTimeout sets the maximum wait time before an OTLP batch is sent.
func (c *Config) WithBatchTimeout(timeout time.Duration) *Config {
	c.Exporter.BatchTimeout = timeout
	return c
}

// WithSamplingRatio sets the fraction of traces to sample.
// Values are clamped between 0.0 (never sample) and 1.0 (always sample).
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

// WithResourceAttr adds or updates a single resource attribute (key-value pair).
func (c *Config) WithResourceAttr(key string, value any) *Config {
	c.ResourceAttrs[key] = value
	return c
}

// WithResourceAttrs bulk-adds or updates multiple resource attributes.
func (c *Config) WithResourceAttrs(attrs map[string]any) *Config {
	maps.Copy(c.ResourceAttrs, attrs)
	return c
}

// WithDebug enables or disables debug mode.
// When enabled, telemetry is also printed to stdout using OTLP stdout exporters.
func (c *Config) WithDebug(debug bool) *Config {
	c.Debug = debug
	return c
}

// WithInsecure enables or disables insecure mode (skips TLS verification).
func (c *Config) WithInsecure(insecure bool) *Config {
	c.Security.Insecure = insecure
	c.Security.TLSCredentials = nil
	return c
}

// WithTLSCredentials sets explicit gRPC transport credentials for secure telemetry transmission.
func (c *Config) WithTLSCredentials(creds credentials.TransportCredentials) *Config {
	c.Security.Insecure = false
	c.Security.TLSCredentials = creds
	return c
}
