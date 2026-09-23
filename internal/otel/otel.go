package otel

import (
	"context"
	"net/http"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	otelsemconv "go.opentelemetry.io/otel/semconv/v1.43.0"
	"go.opentelemetry.io/otel/trace"
	"goyave.dev/goyave/v5/internal/otel/semconv"
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
// TODO span filters and custom attributes
func StartSpan(ctx context.Context, tracer trace.Tracer, request *http.Request) context.Context {
	ctx, _ = tracer.Start(ctx, SpanNameServe, // TODO span name Method+Route
		trace.WithSpanKind(trace.SpanKindServer),
		trace.WithAttributes(semconv.SpanAttrs(request)...),
	)
	// TODO parent span retrieved from request headers? -> Propagator
	return ctx
}

// EndSpan ends an OpenTelemetry span with the given error and http status code.
func EndSpan(ctx context.Context, err error, status int) {
	span := trace.SpanFromContext(ctx)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
	}
	span.SetAttributes(otelsemconv.HTTPResponseStatusCode(status))
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

// TODO docs testutil.NewServer
