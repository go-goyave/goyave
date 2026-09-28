package slog

import (
	"context"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel/log"
)

type options struct {
	otelProvider log.LoggerProvider
	ctx          context.Context
	otelOptions  []otelslog.Option
}

// Option defines a logger setting.
type Option func(o *options)

// WithContext sets a default context for the logger for log operations
// where the context isn't specified (such as [Logger.Info], [Logger.Warn], [Logger.Error], ...)
// By default, uses [context.Background].
func WithContext(ctx context.Context) Option {
	return func(o *options) {
		o.ctx = ctx
	}
}

// WithOpenTelemetryProvider enables OpenTelemetry logs.
// When provided, the [slog.Handler] is wrapped into a [*slog.MultiHandler] and
// an [*otelslog.Handler] is added. To configure the otelslog handler, use [WithOpenTelemetryOptions].
func WithOpenTelemetryProvider(provider log.LoggerProvider) Option {
	return func(o *options) {
		o.otelProvider = provider
	}
}

// WithOpenTelemetryOptions sets the [*otelslog.Handler] options.
// Does nothing if OpenTelemetry logs are disabled. See [WithOpenTelemetry].
func WithOpenTelemetryOptions(opts ...otelslog.Option) Option {
	return func(o *options) {
		o.otelOptions = opts
	}
}
