package slog

import (
	"context"
	"fmt"
	"io"
	"reflect"
	"runtime"
	"time"

	"log/slog"

	"goyave.dev/goyave/v5/util/errors"
)

type unwrapper interface {
	Unwrap() []error
}

// Logger an extension of standard `*slog.Logger` overriding the `Error()` and `ErrorCtx()`
// functions so they take an error as parameter and handle `*errors.Error` gracefully.
type Logger struct {
	*slog.Logger
}

// New creates a new Logger with the given non-nil Handler and a nil context.
func New(h slog.Handler) *Logger {
	return &Logger{Logger: slog.New(h)}
}

// With returns a new Logger that includes the given arguments, converted to
// Attrs as in [Logger.Log].
// The Attrs will be added to each output from the Logger.
// The new Logger shares the old Logger's context.
// The new Logger's handler is the result of calling WithAttrs on the receiver's
// handler.
func (l *Logger) With(args ...any) *Logger {
	return &Logger{Logger: l.Logger.With(args...)}
}

// DebugWithSource logs at `LevelDebug`. The given source will be used instead of the automatically collecting it from the caller.
func (l *Logger) DebugWithSource(ctx context.Context, source uintptr, msg string, args ...any) {
	l.log(ctx, slog.LevelDebug, source, msg, args...)
}

// InfoWithSource logs at `LevelInfo`. The given source will be used instead of the automatically collecting it from the caller.
func (l *Logger) InfoWithSource(ctx context.Context, source uintptr, msg string, args ...any) {
	l.log(ctx, slog.LevelInfo, source, msg, args...)
}

// WarnWithSource logs at `LevelWarn`. The given source will be used instead of the automatically collecting it from the caller.
func (l *Logger) WarnWithSource(ctx context.Context, source uintptr, msg string, args ...any) {
	l.log(ctx, slog.LevelWarn, source, msg, args...)
}

// Error logs the given error at `LevelError`.
func (l *Logger) Error(err error, args ...any) {
	l.logError(context.Background(), 0, err, args...)
}

// ErrorCtx logs the given error at `LevelError` with the given context.
func (l *Logger) ErrorCtx(ctx context.Context, err error, args ...any) {
	l.logError(ctx, 0, err, args...)
}

// ErrorWithSource logs at `LevelError`. The given source will be used instead of the automatically collecting it from the caller.
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
	case *errors.Error:
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
	r := l.makeRecord(level, msg, source, args...)

	if ctx == nil {
		ctx = context.Background()
	}

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

func (l *Logger) handleError(ctx context.Context, err *errors.Error, record slog.Record) {
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
	case *errors.Error:
		l.handleError(ctx, e, clone)
	case errors.Reason:
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

// StructValue recursively convert a structure, structure pointer or map to a `slog.GroupValue`.
// If the given value implements `slog.LogValuer`, this value is returned instead.
// Returns AnyValue if the type is not supported.
func StructValue(v any) slog.Value {
	return structValue(reflect.Indirect(reflect.ValueOf(v)))
}

func structValue(v reflect.Value) slog.Value {
	if valuer, ok := v.Interface().(slog.LogValuer); ok {
		return valuer.LogValue()
	}
	var attrs []slog.Attr
	switch v.Kind() {
	case reflect.Struct:
		t := v.Type()
		numField := t.NumField()
		attrs = make([]slog.Attr, 0, numField)
		for i := 0; i < numField; i++ {
			fieldType := t.Field(i)
			fieldValue := v.Field(i)
			if !fieldType.IsExported() {
				continue
			}
			attrs = append(attrs, slog.Any(fieldType.Name, structValue(fieldValue)))
		}
	case reflect.Map:
		iter := v.MapRange()
		for iter.Next() {
			key := iter.Key()
			value := iter.Value()

			attrs = append(attrs, slog.Any(fmt.Sprintf("%s", key.Interface()), structValue(value)))
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
