# Gotel: OpenTelemetry Toolkit for Go

A simple OpenTelemetry library for Go applications that provides distributed
tracing, metrics collection and structured logging with minimal setup
complexity.

## Understanding OpenTelemetry

OpenTelemetry is an observability framework that helps you understand what's
happening inside your applications when they're running in production. Think of
it as your application's health monitoring system that provides three key types
of information about your software.

### The Three Pillars of Observability

Modern observability relies on three fundamental signals that work together to
give you complete visibility into your applications:

**Distributed Tracing** acts like a GPS tracker for requests as they flow
through your system. When a user clicks a button on your website, that action
might trigger dozens of operations across multiple services, databases, and
external APIs. Tracing creates a timeline that shows you exactly what happened,
how long each step took, and where problems occurred. Each "span" in a trace
represents a single operation, and spans can be nested to show parent-child
relationships between operations.

**Metrics** provide the vital signs of your application. These are numerical
measurements that tell you how your system is performing over time. Common
metrics include request counts, response times, error rates, CPU usage, and
memory consumption. Unlike traces that show individual requests, metrics give
you aggregate data that helps you spot trends and understand system health at
scale.

**Structured Logging** captures detailed information about specific events in
your application. While traditional logs are often unstructured text messages,
structured logs use a consistent format that makes them searchable and
correlatable with traces and metrics. When something goes wrong, logs provide
the context you need to understand exactly what happened.

## OpenTelemetry Architecture

OpenTelemetry consists of several interconnected components that work together
to provide comprehensive observability capabilities.

### The OpenTelemetry API

The API defines the interfaces that developers use to instrument their
applications. This includes creating spans for tracing, recording metrics, and
emitting log records. The API is designed to be minimal and stable, providing a
consistent interface regardless of which SDK implementation you use.

The API follows semantic conventions that standardize how common operations
should be instrumented. For example, HTTP server spans should include specific
attributes like request method, URL path, and response status code. These
conventions ensure consistency across different applications and make it easier
to build generic monitoring dashboards.

### The OpenTelemetry SDK

The SDK provides the actual implementation of telemetry collection and
processing. It handles tasks like sampling decisions, batch processing, and data
export. The SDK is configurable, allowing you to adjust behavior for different
environments and requirements.

Key SDK components include samplers that decide which traces to collect,
processors that transform and batch telemetry data, and exporters that send data
to monitoring backends. This modular design allows fine-grained control over
performance and functionality.

### Instrumentation Libraries

OpenTelemetry provides automatic instrumentation for popular frameworks and
libraries. These instrumentation libraries inject observability code into
existing codebases without requiring manual changes. For Go applications, this
includes instrumentation for HTTP servers and clients, database drivers, and
popular frameworks.

Automatic instrumentation significantly reduces the effort required to achieve
basic observability. You can get meaningful traces and metrics from your
application with minimal code changes, then add custom instrumentation for
business-specific operations.

### The OpenTelemetry Collector

The Collector is a separate process that receives, processes, and exports
telemetry data. It can perform tasks like data transformation, filtering, and
routing to multiple backends. The Collector also provides a buffer between your
application and monitoring systems, improving reliability and performance.

While not strictly required, the Collector becomes essential in production
environments where you need advanced data processing capabilities or want to
avoid tight coupling between applications and monitoring backends.

## Configuration

Gotel provides extensive configuration options through functional options. The
configuration system ensures type safety while providing flexibility for
different deployment environments.

### Service Identification

Every OpenTelemetry service requires identification through name, version, and
environment. These attributes appear in all telemetry data and help organize
observability data in monitoring systems:

```go
provider, err := gotel.NewProvider(
  ctx,
  gotel.WithServiceInfo("user-service", "2.1.0", "production"),
)
```

### Exporter Configuration

Configure the OTLP endpoint and behavior for sending telemetry data:

```go
provider, err := gotel.NewProvider(ctx,
  gotel.WithServiceInfo("my-service", "1.0.0", "production"),
  gotel.WithEndpoint("otlp-collector.monitoring.svc.cluster.local:4317"),
  gotel.WithBatchTimeout(10 * time.Second),
  gotel.WithExportTimeout(30 * time.Second),
  gotel.WithHeaders(map[string]string{
    "x-api-key": "your-api-key",
  }),
)
```

### Security Configuration

Configure TLS and authentication for secure telemetry transmission:

```go
provider, err := gotel.NewProvider(ctx,
  gotel.WithServiceInfo("secure-service", "1.0.0", "production"),
  gotel.WithEndpoint("secure-collector.example.com:4317"),
  gotel.WithTLSCredentials(TLSCredentials),
)
```

### Sampling Configuration

Control which traces get collected to manage performance and costs:

```go
provider, err := gotel.NewProvider(ctx,
  gotel.WithServiceInfo("high-traffic-service", "1.0.0", "production"),
  gotel.WithSamplingRatio(0.1), // Collect 10% of traces
)
```

### Resource Attributes

Add custom attributes that describe your service and environment:

```go
provider, err := gotel.NewProvider(ctx,
  gotel.WithServiceInfo("my-service", "1.0.0", "production"),
  gotel.WithResourceAttrs(map[string]any{
    "deployment.environment": "kubernetes",
    "k8s.cluster.name":      "prod-cluster-east",
    "k8s.namespace.name":    "applications",
  }),
)
```

## Features in Detail

### Custom Tracing

Beyond automatic instrumentation, you can add custom tracing for business
operations:

```go
func ProcessOrder(ctx context.Context, tracer *gotel.Tracer, orderID string) error {
  return tracer.WithSpan(ctx, "process_order", func(ctx context.Context, span *gotel.Span) error {
    span.WithAttributes(
      attribute.String("order.id", orderID),
      attribute.String("order.status", "processing"),
    )

    // Your business logic here
    if err := validateOrder(ctx, orderID); err != nil {
      span.AddEvent("validation_failed", attribute.String("error", err.Error()))
      return err
    }

    span.AddEvent("order_validated")
    return fulfillOrder(ctx, orderID)
  })
}
```

### Custom Metrics

Create application-specific metrics for business insights:

```go
// Create custom metrics
orderCounter, err := registry.Counter("orders_total", "Total number of orders processed")

orderValue, err := registry.Histogram(
  "order_value_dollars",
  "Distribution of order values",
  metric.WithExplicitBucketBoundaries(10, 25, 50, 100, 250, 500, 1000),
)

// Use in business logic
func ProcessOrder(ctx context.Context, order Order) {
  orderCounter.Add(
    ctx, 1,
    metric.WithAttributes(
      attribute.String("status", order.Status),
      attribute.String("region", order.ShippingRegion),
    ),
  )

  orderValue.Record(
    ctx, order.TotalValue,
    metric.WithAttributes(
      attribute.String("currency", order.Currency),
    ),
  )
}
```

### Database Instrumentation

Add comprehensive database observability:

```go
// Wrap your database connection
db, err := sql.Open("postgres", connectionString)
if err != nil {
  return err
}

dbTracer := gotel.NewDBTracer(tracer, metrics, "user_db", "postgresql")
tracedDB := gotel.NewTracedDB(db, dbTracer)

// All operations are automatically instrumented
rows, err := tracedDB.QueryContext(
  ctx,
  "SELECT id, name, email FROM users WHERE active = $1", true,
)

// Transactions are also instrumented
tx, err := tracedDB.BeginTx(ctx, nil)
if err != nil {
  return err
}
defer tx.Rollback()

result, err := tx.ExecContext(
  ctx, "UPDATE users SET last_login = $1 WHERE id = $2", time.Now(), userID,
)
if err != nil {
  return err
}

return tx.Commit()
```

### Structured Logging

Use context-aware structured logging:

```go
logger := provider.Logger()

func HandleRequest(ctx context.Context, request Request) {
  // Logs automatically include trace context
  logger.Info(ctx, "Processing request",
    zap.String("user_id", request.UserID),
    zap.String("action", request.Action),
  )

  if err := processRequest(ctx, request); err != nil {
    logger.Error(ctx, "Request processing failed",
      zap.Error(err),
      zap.String("user_id", request.UserID),
    )
    return
  }

  logger.Info(ctx, "Request completed successfully")
}
```
