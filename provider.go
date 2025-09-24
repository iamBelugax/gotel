package gotel

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutlog"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/metric"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Provider is the central struct that encapsulates all OpenTelemetry SDK components.
type Provider struct {
	config *config

	traceProvider  *sdktrace.TracerProvider
	metricProvider *sdkmetric.MeterProvider
	logProvider    *sdklog.LoggerProvider

	traceExporter  sdktrace.SpanExporter
	metricExporter sdkmetric.Exporter
	logExporter    sdklog.Exporter

	tracer trace.Tracer
	meter  metric.Meter
	logger *ZapLogger
}

// NewProvider initializes and configures the OpenTelemetry SDK based on the provided configuration.
// It sets up the resource, exporters and providers for tracing, metrics and logging.
func NewProvider(ctx context.Context, opts ...Option) (*Provider, error) {
	conf := DefaultConfig(opts...)
	p := &Provider{config: conf}

	resource, err := p.createResource(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource : %v", err)
	}

	if err := p.initTracing(ctx, resource); err != nil {
		return nil, err
	}

	if err := p.initMetrics(ctx, resource); err != nil {
		return nil, err
	}

	if err := p.initLogging(ctx, resource); err != nil {
		return nil, err
	}

	return p, nil
}

// Tracer returns the configured tracer.
func (p *Provider) Tracer() trace.Tracer {
	return p.tracer
}

// Meter returns the configured meter.
func (p *Provider) Meter() metric.Meter {
	return p.meter
}

// Logger returns the configured Zap logger with OpenTelemetry integration.
func (p *Provider) Logger() *ZapLogger {
	return p.logger
}

// Shutdown gracefully shuts down all telemetry exporters.
func (p *Provider) Shutdown(ctx context.Context) error {
	var errs []error

	if err := p.traceProvider.Shutdown(ctx); err != nil {
		errs = append(errs, fmt.Errorf("failed to shutdown trace provider: %w", err))
	}

	if err := p.metricProvider.Shutdown(ctx); err != nil {
		errs = append(errs, fmt.Errorf("failed to shutdown metric provider: %w", err))
	}

	if err := p.logProvider.Shutdown(ctx); err != nil {
		errs = append(errs, fmt.Errorf("failed to shutdown log provider: %w", err))
	}

	if len(errs) > 0 {
		return fmt.Errorf("shutdown failed with errors: %v", errs)
	}
	return nil
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

// initMetrics sets up the meter provider and exporter.
func (p *Provider) initMetrics(ctx context.Context, res *resource.Resource) error {
	var (
		err      error
		exporter sdkmetric.Exporter
	)

	if p.config.Debug {
		exporter, err = stdoutmetric.New(stdoutmetric.WithPrettyPrint())
	} else {
		opts := []otlpmetricgrpc.Option{
			otlpmetricgrpc.WithEndpoint(p.config.Exporter.Endpoint),
			otlpmetricgrpc.WithTimeout(p.config.Exporter.ExportTimeout),
		}

		if p.config.Security.Insecure {
			opts = append(opts, otlpmetricgrpc.WithDialOption(
				grpc.WithTransportCredentials(insecure.NewCredentials()),
			))
		} else {
			opts = append(opts, otlpmetricgrpc.WithDialOption(
				grpc.WithTransportCredentials(p.config.Security.TLSCredentials),
			))
		}

		if len(p.config.Exporter.Headers) > 0 {
			opts = append(opts, otlpmetricgrpc.WithHeaders(p.config.Exporter.Headers))
		}

		exporter, err = otlpmetricgrpc.New(ctx, opts...)
	}

	if err != nil {
		return fmt.Errorf("failed to create metric exporter: %w", err)
	}

	p.metricExporter = exporter
	p.metricProvider = sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(
			sdkmetric.NewPeriodicReader(exporter, sdkmetric.WithInterval(15*time.Second)),
		),
		sdkmetric.WithResource(res),
	)
	otel.SetMeterProvider(p.metricProvider)
	p.meter = p.metricProvider.Meter(p.config.Service.Name)
	return nil
}

// initLogging sets up the logger provider, exporter, and integrates with Zap.
func (p *Provider) initLogging(ctx context.Context, res *resource.Resource) error {
	var (
		err      error
		exporter sdklog.Exporter
	)

	if p.config.Debug {
		exporter, err = stdoutlog.New(stdoutlog.WithPrettyPrint())
	} else {
		opts := []otlploggrpc.Option{
			otlploggrpc.WithEndpoint(p.config.Exporter.Endpoint),
			otlploggrpc.WithTimeout(p.config.Exporter.ExportTimeout),
		}

		if p.config.Security.Insecure {
			opts = append(opts, otlploggrpc.WithDialOption(
				grpc.WithTransportCredentials(insecure.NewCredentials()),
			))
		} else {
			opts = append(opts, otlploggrpc.WithDialOption(
				grpc.WithTransportCredentials(p.config.Security.TLSCredentials),
			))
		}

		if len(p.config.Exporter.Headers) > 0 {
			opts = append(opts, otlploggrpc.WithHeaders(p.config.Exporter.Headers))
		}

		exporter, err = otlploggrpc.New(ctx, opts...)
	}

	if err != nil {
		return fmt.Errorf("failed to create log exporter: %w", err)
	}

	p.logExporter = exporter
	p.logProvider = sdklog.NewLoggerProvider(
		sdklog.WithProcessor(sdklog.NewBatchProcessor(exporter)),
		sdklog.WithResource(res),
	)
	global.SetLoggerProvider(p.logProvider)

	zapLogger, err := newZapLogger(
		p.config.Service.Name,
		p.config.Service.Version,
		p.config.Logging.Level,
		p.config.Debug,
	)
	if err != nil {
		return fmt.Errorf("failed to create zap logger: %w", err)
	}

	p.logger = zapLogger
	return nil
}
