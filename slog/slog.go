package slog

import (
	"context"
	"fmt"
	"io"
	"os"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"time"

	"log/slog"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"goyave.dev/goyave/v5/internal/otel"
	"goyave.dev/goyave/v5/util/errwrap"
)

type unwrapper interface {
	Unwrap() []error
}

// Logger an extension of standard [*slog.Logger] overriding the [slog.Logger.Error] and [slog.Logger.ErrorContext]
// functions so they take an error as parameter and handle [*errwrap.Error] gracefully.
type Logger struct {
	*slog.Logger
	ctx context.Context
}

// New creates a new Logger with the given non-nil Handler and a nil context.
func New(h slog.Handler, opts ...Option) *Logger {
	options := &options{}
	for _, o := range opts {
		o(options)
	}

	handler := h
	if options.enableOpenTelemetry {
		handler = slog.NewMultiHandler(h, otelslog.NewHandler(otel.LoggerName, options.otelOptions...))
	}

	return &Logger{
		Logger: slog.New(handler),
		ctx:    options.ctx,
	}
}

// WithContext returns a new [*Logger] with the given context attached.
// This context will be used by default for logging operations that don't specify a
// context (such as [Logger.Info], [Logger.Warn], [Logger.Error], ...)
func (l *Logger) WithContext(ctx context.Context) *Logger {
	return &Logger{
		Logger: l.Logger,
		ctx:    ctx,
	}
}

// With returns a new [Logger] that includes the given arguments, converted to
// [slog.Attr] as in [Logger.Log].
// The [slog.Attr] will be added to each output from the [Logger].
// The new [Logger] shares the old Logger's context.
// The new [Logger]'s handler is the result of calling [slog.Logger.With] on the receiver's
// handler.
func (l *Logger) With(args ...any) *Logger {
	return &Logger{Logger: l.Logger.With(args...), ctx: l.ctx}
}

// WithGroup returns a new [Logger] that starts a group, if name is non-empty.
// The keys of all attributes added to the [Logger] will be qualified by the given
// name. (How that qualification happens depends on the [slog.Handler.WithGroup]
// method of the [Logger]'s [slog.Handler].)
//
// If name is empty, [Logger.WithGroup] returns the receiver.
func (l *Logger) WithGroup(name string) *Logger {
	return &Logger{Logger: l.Logger.WithGroup(name), ctx: l.ctx}
}

// Debug logs at [slog.LevelDebug].
func (l *Logger) Debug(msg string, args ...any) {
	l.log(l.ctx, slog.LevelDebug, 0, msg, args...)
}

// Info logs at [slog.LevelInfo].
func (l *Logger) Info(msg string, args ...any) {
	l.log(l.ctx, slog.LevelInfo, 0, msg, args...)
}

// Warn logs at [slog.LevelWarn].
func (l *Logger) Warn(msg string, args ...any) {
	l.log(l.ctx, slog.LevelWarn, 0, msg, args...)
}

// DebugWithSource logs at [slog.LevelDebug]. The given source will be used instead of the automatically collecting it from the caller.
func (l *Logger) DebugWithSource(ctx context.Context, source uintptr, msg string, args ...any) {
	l.log(ctx, slog.LevelDebug, source, msg, args...)
}

// InfoWithSource logs at [slog.LevelInfo]. The given source will be used instead of the automatically collecting it from the caller.
func (l *Logger) InfoWithSource(ctx context.Context, source uintptr, msg string, args ...any) {
	l.log(ctx, slog.LevelInfo, source, msg, args...)
}

// WarnWithSource logs at [slog.LevelWarn]. The given source will be used instead of the automatically collecting it from the caller.
func (l *Logger) WarnWithSource(ctx context.Context, source uintptr, msg string, args ...any) {
	l.log(ctx, slog.LevelWarn, source, msg, args...)
}

// Error logs the given error at [slog.LevelError].
func (l *Logger) Error(err error, args ...any) {
	l.logError(l.ctx, 0, err, args...)
}

// ErrorContext logs the given error at [slog.LevelError] with the given context.
func (l *Logger) ErrorContext(ctx context.Context, err error, args ...any) {
	l.logError(ctx, 0, err, args...)
}

// ErrorWithSource logs at [slog.LevelError]. The given source will be used instead of the automatically collecting it from the caller.
func (l *Logger) ErrorWithSource(ctx context.Context, source uintptr, err error, args ...any) {
	l.logError(ctx, source, err, args...)
}

func (l *Logger) logError(ctx context.Context, source uintptr, err error, args ...any) {
	msg := "<nil>"
	if err != nil {
		msg = err.Error()
	}
	r := l.makeRecord(slog.LevelError, msg, source, args...)

	if ctx == nil {
		ctx = context.Background()
	}

	switch e := err.(type) {
	case *errwrap.Error:
		l.handleError(ctx, e, r)
	case unwrapper:
		l.handleReason(ctx, err, nil, r)
		for _, e := range e.Unwrap() {
			l.handleReason(ctx, e, nil, r)
		}
	default:
		_ = l.Handler().Handle(ctx, r)
	}
}

func (l *Logger) log(ctx context.Context, level slog.Level, source uintptr, msg string, args ...any) {
	if ctx == nil {
		ctx = context.Background()
	}

	handler := l.Handler()
	if !handler.Enabled(ctx, level) {
		return
	}
	r := l.makeRecord(level, msg, source, args...)

	_ = l.Handler().Handle(ctx, r)
}

func (l *Logger) makeRecord(level slog.Level, msg string, pc uintptr, args ...any) slog.Record {
	if pc == 0 {
		var pcs [1]uintptr
		runtime.Callers(4, pcs[:])
		pc = pcs[0]
	}
	r := slog.NewRecord(time.Now(), level, msg, pc)
	r.Add(args...)
	return r
}

func (l *Logger) handleError(ctx context.Context, err *errwrap.Error, record slog.Record) {
	trace := slog.String("trace", err.StackFrames().String())
	if err.Len() == 0 {
		record.AddAttrs(trace)
		_ = l.Handler().Handle(ctx, record)
		return
	}

	for _, r := range err.Unwrap() {
		l.handleReason(ctx, r, &trace, record)
	}
}

func (l *Logger) handleReason(ctx context.Context, reason error, trace *slog.Attr, record slog.Record) {
	clone := record.Clone()
	if reason == nil {
		clone.Message = "<nil>"
	} else {
		clone.Message = reason.Error()
	}
	switch e := reason.(type) {
	case *errwrap.Error:
		l.handleError(ctx, e, clone)
	case errwrap.Reason:
		if trace != nil {
			clone.AddAttrs(*trace)
		}
		if _, isDevMode := l.Handler().(*DevModeHandler); !isDevMode {
			clone.AddAttrs(slog.Any("reason", e.Value()))
		}
		_ = l.Handler().Handle(ctx, clone)
	default:
		if trace != nil {
			clone.AddAttrs(*trace)
		}
		if slogValuer, ok := reason.(slog.LogValuer); ok {
			clone.AddAttrs(slog.Any("reason", slogValuer.LogValue()))
		}
		_ = l.Handler().Handle(ctx, clone)
	}
}

// StructValue recursively convert a structure, structure pointer or map to a [slog.GroupValue].
// If the given value implements [slog.LogValuer], this value is returned instead.
// Returns AnyValue if the type is not supported.
func StructValue(v any) slog.Value {
	seen := map[uintptr]struct{}{}
	value := reflect.ValueOf(v)
	if value.Kind() == reflect.Pointer && value.IsValid() {
		ptr := value.Pointer()
		seen[ptr] = struct{}{}
	}
	return structValue(reflect.Indirect(value), seen)
}

func structValue(v reflect.Value, seen map[uintptr]struct{}) slog.Value {
	if !v.IsValid() {
		return slog.StringValue("<nil>")
	}

	if v.Kind() == reflect.Pointer {
		ptr := v.Pointer()
		if ptr == 0 {
			return slog.StringValue("<nil>")
		}
		if _, ok := seen[ptr]; ok {
			return slog.StringValue("<already_seen>")
		}
		seen[ptr] = struct{}{}

		v = v.Elem()
	}

	if valuer, ok := reflect.TypeAssert[slog.LogValuer](v); ok {
		return valuer.LogValue()
	}
	var attrs []slog.Attr
	switch v.Kind() {
	case reflect.Struct:
		t := v.Type()
		numField := t.NumField()
		attrs = make([]slog.Attr, 0, numField)
		for i := range numField {
			fieldType := t.Field(i)
			fieldValue := v.Field(i)
			if !fieldType.IsExported() {
				continue
			}
			attrs = append(attrs, slog.Any(fieldType.Name, structValue(fieldValue, seen)))
		}
	case reflect.Map:
		iter := v.MapRange()
		for iter.Next() {
			key := iter.Key()
			value := iter.Value()

			attrs = append(attrs, slog.Any(fmt.Sprintf("%v", key.Interface()), structValue(value, seen)))
		}
	case reflect.Slice, reflect.Array:
		len := v.Len()
		for i := range len {
			attrs = append(attrs, slog.Any(strconv.Itoa(i), structValue(v.Index(i), seen)))
		}
	default:
		return slog.AnyValue(v.Interface())
	}
	return slog.GroupValue(attrs...)
}

// DiscardLogger returns a new Logger that discards all logs.
func DiscardLogger() *Logger {
	return &Logger{Logger: slog.New(slog.NewTextHandler(io.Discard, nil))}
}

var defaultLogger = New(NewHandler(!isProduction(), os.Stderr))

func isProduction() bool {
	env := strings.ToLower(os.Getenv("ENV"))
	return env == "prod" || env == "production"
}

// Default returns the default global logger.
// This logger uses the JSON handler and outputs to [os.Stderr].
func Default() *Logger {
	return defaultLogger
}

// SetDefault replace the default logger.
//
// This operation is not concurrently safe.
func SetDefault(logger *Logger) {
	defaultLogger = logger
}

// loggerCtxKey the key used to store the logger in the context.
type loggerCtxKey struct{}

// Context inject the given logger as a context value. The logger
// can be retrieved from the returned context using [FromContext].
//
// The given context is attached to the logger using [Logger.WithContext].
func Context(ctx context.Context, logger *Logger) context.Context {
	return context.WithValue(ctx, loggerCtxKey{}, logger.WithContext(ctx))
}

// FromContext return the logger stored in the context. If there
// is no logger in the context, returns the default logger instead.
//
// The default logger uses the JSON handler and outputs to [os.Stderr].
func FromContext(ctx context.Context) *Logger {
	if u, ok := ctx.Value(loggerCtxKey{}).(*Logger); ok {
		return u
	}
	return defaultLogger
}
