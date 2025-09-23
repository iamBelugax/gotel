package gotel

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// Span is a wrapper around the OpenTelemetry span that provides convenient methods
// for setting attributes, errors, statuses and events.
type Span struct {
	trace.Span
	tracer trace.Tracer
}

// NewSpan creates and starts a new span using the provided tracer and name.
func NewSpan(
	ctx context.Context, tracer trace.Tracer, name string, options ...trace.SpanStartOption,
) (context.Context, *Span) {
	ctx, span := tracer.Start(ctx, name, options...)
	return ctx, &Span{tracer: tracer, Span: span}
}

// WithAttributes adds one or more attributes to the span.
func (s *Span) WithAttributes(attrs ...attribute.KeyValue) *Span {
	s.Span.SetAttributes(attrs...)
	return s
}

// WithError records the given error on the span and sets the span status to error.
func (s *Span) WithError(err error) *Span {
	if err != nil {
		s.Span.RecordError(err)
		s.Span.SetStatus(codes.Error, err.Error())
	}
	return s
}

// EndWithError records the given error (if any) and then ends the span.
func (s *Span) EndWithError(err error) {
	if err != nil {
		s.WithError(err)
	} else {
		s.WithStatus(codes.Ok, "success")
	}
	s.End()
}

// WithStatus sets the status of the span explicitly.
func (s *Span) WithStatus(code codes.Code, description string) *Span {
	s.Span.SetStatus(code, description)
	return s
}

// AddEvent adds an event to the span with optional attributes.
func (s *Span) AddEvent(name string, attrs ...attribute.KeyValue) *Span {
	s.Span.AddEvent(name, trace.WithAttributes(attrs...))
	return s
}
