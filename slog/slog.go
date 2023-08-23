package slog

import (
	"context"
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
	// TODO ability to chose the output for each log level?
}

// New creates a new Logger with the given non-nil Handler and a nil context.
func New(h slog.Handler) *Logger {
	return &Logger{slog.New(h)}
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

func (l *Logger) DebugWithSource(ctx context.Context, source uintptr, msg string, args ...any) {
	l.log(ctx, slog.LevelDebug, source, msg, args...)
}

func (l *Logger) InfoWithSource(ctx context.Context, source uintptr, msg string, args ...any) {
	l.log(ctx, slog.LevelInfo, source, msg, args...)
}

func (l *Logger) WarnWithSource(ctx context.Context, source uintptr, msg string, args ...any) {
	l.log(ctx, slog.LevelWarn, source, msg, args...)
}

func (l *Logger) Error(err error, args ...any) {
	l.logError(nil, 0, err, args...)
}

func (l *Logger) ErrorCtx(ctx context.Context, err error, args ...any) {
	l.logError(ctx, 0, err, args...)
}

func (l *Logger) ErrorWithSource(ctx context.Context, source uintptr, err error, args ...any) {
	l.logError(ctx, source, err, args...)
}

func (l *Logger) logError(ctx context.Context, source uintptr, err error, args ...any) {
	r := l.makeRecord(slog.LevelError, err.Error(), source, args...)

	if ctx == nil {
		ctx = context.Background()
	}

	switch e := err.(type) {
	case *errors.Error:
		l.handleError(ctx, e, r)
	case unwrapper:
		for _, e := range e.Unwrap() {
			l.handleReason(ctx, e, r)
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
	record.AddAttrs(trace)
	if err.Len() == 0 {
		_ = l.Handler().Handle(ctx, record)
		return
	}

	for _, r := range err.Unwrap() {
		l.handleReason(ctx, r, record)
	}
}

func (l *Logger) handleReason(ctx context.Context, reason error, record slog.Record) {
	clone := record.Clone()
	clone.Message = reason.Error()
	switch e := reason.(type) {
	case *errors.Error:
		l.handleError(ctx, e, clone)
	case errors.Reason:
		if _, isDevMode := l.Handler().(*DevModeHandler); !isDevMode {
			clone.AddAttrs(slog.Any("reason", e.Value()))
		}
		_ = l.Handler().Handle(ctx, clone)
	default:
		_ = l.Handler().Handle(ctx, clone)
	}
}
