# OpenTelemetry Toolkit for Go

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

### Why OpenTelemetry Matters

OpenTelemetry is a Cloud Native Computing Foundation project that emerged from
merging two earlier projects: OpenTracing (focused on distributed tracing) and
OpenCensus (focused on metrics and tracing). This unification created a
comprehensive observability framework that's now supported by virtually every
major monitoring and APM vendor.

The framework consists of several key components that work together seamlessly.
The OpenTelemetry API defines how developers interact with observability
features in their code. The SDK provides the actual implementation of data
collection and processing. Instrumentation libraries automatically add
observability to popular frameworks and libraries without requiring code
changes. Finally, the OpenTelemetry Collector receives, processes, and exports
telemetry data to various backends.
