package otel

import (
	"context"
	"net/http"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	otelsemconv "go.opentelemetry.io/otel/semconv/v1.43.0"
	"go.opentelemetry.io/otel/semconv/v1.43.0/httpconv"
	"go.opentelemetry.io/otel/trace"
	"goyave.dev/goyave/v5/internal/otel/semconv"
)

const LoggerName = "goyave.dev/goyave/v6"

const (
	OpenTelemetryTracerName = "goyave.dev/goyave/v6"
	OpenTelemetryMeterName  = "goyave.dev/goyave/v6"

	EventNameWriteHeader = "goyave.response.write_header"

	Version = "0.1.0"
)

// Tracer returns an OpenTelemetry tracer configured for Goyave.
func Tracer(provider trace.TracerProvider) trace.Tracer {
	// Custom options not allowed, only the Goyave instrumentation needs to be in control of them.
	return provider.Tracer(OpenTelemetryTracerName, trace.WithInstrumentationVersion(Version))
}

type SpanData struct {
	StartTime   time.Time
	Request     *http.Request
	Propagators propagation.TextMapPropagator
	Route       string
}

// StartSpan adds an OpenTelemetry span to the trace with the given name.
func StartSpan(ctx context.Context, tracer trace.Tracer, spanData SpanData) context.Context {
	if spanData.Propagators != nil {
		ctx = spanData.Propagators.Extract(ctx, propagation.HeaderCarrier(spanData.Request.Header))
	}

	// Custom options not allowed, only the Goyave instrumentation needs to be in control of them.
	// Custom attributes can be added to the span later if developers chose to do so.
	ctx, _ = tracer.Start(ctx, semconv.SpanName(spanData.Request.Method, spanData.Route),
		trace.WithSpanKind(trace.SpanKindServer),
		trace.WithAttributes(semconv.SpanAttrs(spanData.Request, spanData.Route)...),
		trace.WithTimestamp(spanData.StartTime),
	)
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

type ServerMetricData struct {
	Error           error
	Request         *http.Request
	Route           string
	RequestDuration time.Duration
	RequestSize     int64
	ResponseSize    int64
	ResponseStatus  int
}

func (m *HTTPServerMeters) RecordMetrics(ctx context.Context, data ServerMetricData) {
	attrs := semconv.MetricAttrs(data.Request, data.Route, data.ResponseStatus, data.Error)
	measureOpt := metric.WithAttributeSet(attribute.NewSet(attrs...))

	m.requestBodySizeHistogram.Inst().Record(ctx, data.RequestSize, measureOpt)
	m.responseBodySizeHistogram.Inst().Record(ctx, data.ResponseSize, measureOpt)
	m.requestDurationHistogram.Inst().Record(ctx, float64(data.RequestDuration)/float64(time.Second), measureOpt)
}

// TODO docs testutil.NewServer (mock providers)
// TODO docs propagation and baggage
