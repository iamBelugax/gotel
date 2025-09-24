package gotel

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Provider is the central struct that encapsulates all OpenTelemetry SDK components.
type Provider struct {
	config        *config
	tracer        trace.Tracer
	traceExporter sdktrace.SpanExporter
	traceProvider *sdktrace.TracerProvider
}

// NewProvider initializes and configures the OpenTelemetry SDK based on the provided configuration.
// It sets up the resource, exporters and providers for tracing, metrics and logging.
func NewProvider(ctx context.Context, serviceName string, opts ...Option) (*Provider, error) {
	conf := DefaultConfig(serviceName, opts...)
	p := &Provider{config: conf}

	// Create resource describing the service.
	resource, err := p.createResource(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource : %v", err)
	}

	// Initialize tracing (exporters, processors, providers).
	if err := p.initTracing(ctx, resource); err != nil {
		return nil, err
	}

	return p, nil
}

// Tracer returns the configured tracer.
func (p *Provider) Tracer() trace.Tracer {
	return p.tracer
}

// createResource builds an OTEL resource from service metadata and custom attributes.
func (p *Provider) createResource(ctx context.Context) (*resource.Resource, error) {
	// Standard semantic attributes (service name, version, environment).
	attributes := []resource.Option{
		resource.WithAttributes(
			semconv.ServiceName(p.config.Service.Name),
			semconv.ServiceVersion(p.config.Service.Version),
			semconv.DeploymentEnvironment(p.config.Service.Environment),
		),
	}

	// Append custom resource attributes, if provided.
	if len(p.config.ResourceAttrs) > 0 {
		customAttrs := make([]attribute.KeyValue, 0, len(p.config.ResourceAttrs))
		for key, value := range p.config.ResourceAttrs {
			customAttrs = append(customAttrs, attribute.String(key, fmt.Sprintf("%v", value)))
		}
		attributes = append(attributes, resource.WithAttributes(customAttrs...))
	}

	return resource.New(ctx, attributes...)
}

// initTracing configures the tracer provider, exporter, sampler and processor.
func (p *Provider) initTracing(ctx context.Context, res *resource.Resource) error {
	var (
		err      error
		exporter sdktrace.SpanExporter
	)

	if p.config.Debug {
		exporter, err = stdouttrace.New(stdouttrace.WithPrettyPrint())
	} else {
		opts := []otlptracegrpc.Option{
			otlptracegrpc.WithEndpoint(p.config.Exporter.Endpoint),
			otlptracegrpc.WithTimeout(p.config.Exporter.ExportTimeout),
		}

		if p.config.Security.Insecure {
			opts = append(opts, otlptracegrpc.WithDialOption(
				grpc.WithTransportCredentials(insecure.NewCredentials()),
			))
		} else {
			opts = append(opts, otlptracegrpc.WithDialOption(
				grpc.WithTransportCredentials(p.config.Security.TLSCredentials),
			))
		}

		if len(p.config.Exporter.Headers) > 0 {
			opts = append(opts, otlptracegrpc.WithHeaders(p.config.Exporter.Headers))
		}

		exporter, err = otlptracegrpc.New(ctx, opts...)
	}

	if err != nil {
		return fmt.Errorf("failed to create trace exporter: %w", err)
	}

	sampler := sdktrace.TraceIDRatioBased(p.config.Tracing.SamplingRatio)
	processor := sdktrace.NewBatchSpanProcessor(
		exporter,
		sdktrace.WithBatchTimeout(p.config.Exporter.BatchTimeout),
	)
	p.traceProvider = sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sampler),
		sdktrace.WithSpanProcessor(processor),
	)

	p.traceExporter = exporter
	otel.SetTracerProvider(p.traceProvider)

	p.tracer = p.traceProvider.Tracer(p.config.Service.Name)
	return nil
}

// Shutdown gracefully shuts down all telemetry exporters.
func (p *Provider) Shutdown(ctx context.Context) error {
	if err := p.traceProvider.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown trace provider: %w", err)
	}
	return nil
}
