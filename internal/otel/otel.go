package otel

import (
	"context"
	"net/http"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	otelsemconv "go.opentelemetry.io/otel/semconv/v1.43.0"
	"go.opentelemetry.io/otel/semconv/v1.43.0/httpconv"
	"go.opentelemetry.io/otel/trace"
	"goyave.dev/goyave/v5/internal/otel/semconv"
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
func Tracer(provider trace.TracerProvider) trace.Tracer {
	// Custom options not allowed, only the Goyave instrumentation needs to be in control of them.
	return provider.Tracer(OpenTelemetryTracerName, trace.WithInstrumentationVersion(Version))
}

// StartSpan adds an OpenTelemetry span to the trace with the given name.
// TODO span filters and custom attributes
func StartSpan(ctx context.Context, tracer trace.Tracer, request *http.Request) context.Context {
	ctx, _ = tracer.Start(ctx, SpanNameServe, // TODO span name Method+Route
		trace.WithSpanKind(trace.SpanKindServer),
		trace.WithAttributes(semconv.SpanAttrs(request)...),
		// Custom options not allowed, only the Goyave instrumentation needs to be in control of them.
		// Custom attributes can be added to the span later if developers chose to do so.
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
func Meter(provider metric.MeterProvider) metric.Meter {
	// Custom options not allowed, only the Goyave instrumentation needs to be in control of them.
	return provider.Meter(OpenTelemetryMeterName, metric.WithInstrumentationVersion(Version))
}

type HTTPServerMeters struct {
	requestBodySizeHistogram  httpconv.ServerRequestBodySize
	responseBodySizeHistogram httpconv.ServerResponseBodySize
	requestDurationHistogram  httpconv.ServerRequestDuration
}

// NewHTTPServerMeter returns a structure holding the histograms for all HTTP server metrics.
func NewHTTPServerMeter(meter metric.Meter) (*HTTPServerMeters, error) {
	requestBodySizeHistogram, err := httpconv.NewServerRequestBodySize(meter) // FIXME should be defined in the parse middleware
	if err != nil {
		return nil, err
	}
	responseBodySizeHistogram, err := httpconv.NewServerResponseBodySize(meter)
	if err != nil {
		return nil, err
	}

	requestDurationHistogram, err := httpconv.NewServerRequestDuration(
		meter,
		metric.WithExplicitBucketBoundaries(
			0.005, 0.01, 0.025, 0.05, 0.075, 0.1,
			0.25, 0.5, 0.75, 1, 2.5, 5, 7.5, 10,
		),
	)
	if err != nil {
		return nil, err
	}

	// TODO when stable, add support for httpconv.ClientActiveRequests

	return &HTTPServerMeters{
		requestBodySizeHistogram:  requestBodySizeHistogram,
		responseBodySizeHistogram: responseBodySizeHistogram,
		requestDurationHistogram:  requestDurationHistogram,
	}, nil
}

func (m *HTTPServerMeters) RecordMetrics(ctx context.Context, data ServerMetricData) { // TODO pass necessary request (request, response)
	// attributes := n.MetricAttributes(md.ServerName, md.Req, md.StatusCode, md.Route, md.AdditionalAttributes)
	// o := metric.WithAttributeSet(attribute.NewSet(attributes...))

	// m.requestBodySizeHistogram.Inst().Record(ctx, requestSize, o...)
	// m.responseBodySizeHistogram.Inst().Record(ctx, responseSize, o...)
	// m.requestDurationHistogram.Inst().Record(ctx, float64(requestDuration)/float64(time.Second), o...)
}

type ServerMetricData struct {
	Request    *http.Request
	Route      string
	ServerAddr string
	ServerPort int
}

// TODO docs testutil.NewServer
