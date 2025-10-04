package gotel

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

// HTTPMiddleware provides OpenTelemetry instrumentation for HTTP handlers.
type HTTPMiddleware struct {
	serviceName string
	tracer      *Tracer
	metrics     *CommonMetrics
}

// NewHTTPMiddleware creates and returns a new HTTPMiddleware instance.
func NewHTTPMiddleware(serviceName string, tracer *Tracer, metrics *CommonMetrics) *HTTPMiddleware {
	return &HTTPMiddleware{
		tracer:      tracer,
		metrics:     metrics,
		serviceName: serviceName,
	}
}

func (m *HTTPMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		spanName := fmt.Sprintf("%s %s", r.Method, r.URL.Path)
		ctx := otel.GetTextMapPropagator().Extract(r.Context(), propagation.HeaderCarrier(r.Header))

		ctx, span := m.tracer.StartSpan(ctx, spanName,
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(
				semconv.URLPathKey.String(r.URL.String()),
				semconv.HTTPRequestMethodKey.String(r.Method),
			),
		)
		defer span.End()

		attrs := []attribute.KeyValue{
			attribute.String("method", r.Method),
			attribute.String("route", r.URL.Path),
		}

		m.metrics.HTTPActiveRequests.Add(ctx, 1, metric.WithAttributes(attrs...))
		defer m.metrics.HTTPActiveRequests.Add(ctx, -1, metric.WithAttributes(attrs...))

		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		start := time.Now()
		next.ServeHTTP(wrapped, r.WithContext(ctx))
		duration := time.Since(start)

		span.WithAttributes(semconv.HTTPResponseStatusCodeKey.Int(wrapped.statusCode))
		if wrapped.statusCode >= http.StatusBadRequest {
			span.WithStatus(codes.Error, "HTTP Request Failed")
		}

		statusAttrs := append(attrs, attribute.String("status_code", strconv.Itoa(wrapped.statusCode)))
		m.metrics.HTTPRequestsTotal.Add(ctx, 1, metric.WithAttributes(statusAttrs...))
		m.metrics.HTTPRequestDuration.Record(ctx, duration.Seconds(), metric.WithAttributes(attrs...))
	})
}

type responseWriter struct {
	statusCode int
	http.ResponseWriter
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}
