package slog

import (
	"context"

	"go.opentelemetry.io/contrib/bridges/otelslog"
)

type options struct {
	ctx                 context.Context
	otelOptions         []otelslog.Option
	enableOpenTelemetry bool
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

// WithOpenTelemetry enables or disables OpenTelemetry logs.
// When enabled, the [slog.Handler] is wrapped into a [*slog.MultiHandler] and
// an [*otelslog.Handler] is added. To configure the otelslog handler, use [WithOpenTelemetryOptions].
func WithOpenTelemetry(enabled bool) Option {
	return func(o *options) {
		o.enableOpenTelemetry = enabled
	}
}

// WithOpenTelemetryOptions sets the [*otelslog.Handler] options.
// Does nothing if OpenTelemetry logs are disabled. See [WithOpenTelemetry].
func WithOpenTelemetryOptions(opts ...otelslog.Option) Option {
	return func(o *options) {
		o.otelOptions = opts
	}
}
