package otel

import (
	"context"
	"errors"
	"net/http"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	semconv "go.opentelemetry.io/otel/semconv/v1.43.0"
	"go.opentelemetry.io/otel/trace"
	"goyave.dev/goyave/v5/util/errwrap"
)

const LoggerName = "goyave.dev/goyave/v6"

const (
	OpenTelemetryTracerName = "goyave.dev/goyave/v6"
	OpenTelemetryMeterName  = "goyave.dev/goyave/v6"

	SpanNameServe = "goyave.serve"

	EventNameWriteHeader = "goyave.response.write_header"

	Version = "0.1.0"
)

// Tracer returns an OpenTelemetry tracer configured for Goyave.
func Tracer(provider trace.TracerProvider, opts ...trace.TracerOption) trace.Tracer {
	return provider.Tracer(OpenTelemetryTracerName, append([]trace.TracerOption{trace.WithInstrumentationVersion(Version)}, opts...)...)
}

// StartSpan adds an OpenTelemetry span to the trace with the given name.
func StartSpan(ctx context.Context, tracer trace.Tracer, request *http.Request) context.Context {
	ctx, _ = tracer.Start(ctx, SpanNameServe,
		trace.WithSpanKind(trace.SpanKindServer),
		trace.WithAttributes(
			semconv.HTTPRequestMethodKey.String(request.Method),
		),
	)
	// TODO parent span retrieved from request headers?
	return ctx
}

// SpanError record an error event in the current OpenTelemetry span.
func SpanError(ctx context.Context, err error) {
	if err == nil {
		return
	}
	span := trace.SpanFromContext(ctx)

	span.SetStatus(codes.Error, err.Error())

	opts := []trace.EventOption{}
	wrapped, ok := errors.AsType[*errwrap.Error](err)
	if ok {
		opts = append(opts, trace.WithAttributes(semconv.ExceptionStacktrace(wrapped.StackFrames().String())))
	}

	span.RecordError(err, opts...)
}

// EndSpan ends an OpenTelemetry span with the given error.
func EndSpan(ctx context.Context, status int) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(semconv.HTTPResponseStatusCode(status))
	span.End()
}

// AddAttr adds OpenTelemetry attributes to the current span.
func AddAttr(ctx context.Context, attrs ...attribute.KeyValue) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attrs...)
}

// Meter returns an OpenTelemetry meter configured for Goyave.
func Meter(provider metric.MeterProvider, opts ...metric.MeterOption) metric.Meter {
	return provider.Meter(OpenTelemetryMeterName, append([]metric.MeterOption{metric.WithInstrumentationVersion(Version)}, opts...)...)
}

// TODO support metrics
// - requestBodySizeHistogram (parse middleware)
// - responseBodySizeHistogram (Response)
// - requestDurationHistogram (Router)
// - optional support for active_requests

// TODO docs testutil.NewServer
