package gotel

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// Span is a wrapper around the OpenTelemetry span.
type Span struct {
	sp     trace.Span
	tracer trace.Tracer
}

// NewSpan creates and starts a new span using the provided tracer and name.
func NewSpan(
	ctx context.Context, tracer trace.Tracer, name string, options ...trace.SpanStartOption,
) (context.Context, *Span) {
	ctx, span := tracer.Start(ctx, name, options...)
	return ctx, &Span{tracer: tracer, sp: span}
}

// WithAttributes adds one or more attributes to the span.
func (s *Span) WithAttributes(attrs ...attribute.KeyValue) *Span {
	s.sp.SetAttributes(attrs...)
	return s
}

// WithStatus sets the status of the span explicitly.
func (s *Span) WithStatus(code codes.Code, description string) *Span {
	s.sp.SetStatus(code, description)
	return s
}

// AddEvent adds an event to the span with optional attributes.
func (s *Span) AddEvent(name string, attrs ...attribute.KeyValue) *Span {
	s.sp.AddEvent(name, trace.WithAttributes(attrs...))
	return s
}

// WithError records the given error on the span and sets the span status to error.
func (s *Span) WithError(err error) *Span {
	if err != nil {
		s.sp.RecordError(err)
		s.sp.SetStatus(codes.Error, err.Error())
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
	s.sp.End()
}

// End completes the Span.
func (s *Span) End() {
	s.sp.End()
}

// Context returns the SpanContext of the underlying span.
func (s *Span) Context() trace.SpanContext {
	return s.sp.SpanContext()
}

// IsRecording reports whether the span is actively recording events.
func (s *Span) IsRecording() bool {
	return s.sp.IsRecording()
}

// ChildSpan starts a new child span from the current one.
func (s *Span) ChildSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, *Span) {
	ctx, span := s.tracer.Start(ctx, name, opts...)
	return ctx, &Span{tracer: s.tracer, sp: span}
}

// Tracer is a wrapper around the OpenTelemetry tracer.
type Tracer struct {
	tracer trace.Tracer
}

// NewTracer creates a new Tracer with an associated service name.
func NewTracer(tracer trace.Tracer) *Tracer {
	return &Tracer{tracer: tracer}
}

// StartSpan starts a new span with the given name and options.
func (t *Tracer) StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, *Span) {
	return NewSpan(ctx, t.tracer, name, opts...)
}

// WithSpan executes the given function within a new span.
func (t *Tracer) WithSpan(
	ctx context.Context, name string, fn func(context.Context, *Span) error, opts ...trace.SpanStartOption,
) error {
	ctx, span := t.StartSpan(ctx, name, opts...)
	defer span.sp.End()

	if err := fn(ctx, span); err != nil {
		span.WithError(err)
		return err
	}

	span.WithStatus(codes.Ok, "success")
	return nil
}
