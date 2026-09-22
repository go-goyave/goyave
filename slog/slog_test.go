package slog

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"regexp"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel/attribute"
	"goyave.dev/goyave/v5/util/errwrap"
)

type testValuerError struct{}

func (testValuerError) Error() string {
	return "test error"
}

func (testValuerError) LogValue() slog.Value {
	return slog.StringValue("test value")
}

type proxyHandler struct {
	*DevModeHandler
	ctxKey any
}

func (p proxyHandler) Handle(ctx context.Context, r slog.Record) error {
	v := ctx.Value(p.ctxKey)
	if v != nil {
		r.AddAttrs(slog.Any("CONTEXT_VALUE", v))
	}
	return p.DevModeHandler.Handle(ctx, r)
}

func TestLogger(t *testing.T) {
	t.Run("New", func(t *testing.T) {
		handler := NewDevModeHandler(bytes.NewBuffer(make([]byte, 0, 10)), nil)
		l := New(handler)
		assert.Equal(t, &Logger{Logger: slog.New(handler)}, l)
	})

	t.Run("With", func(t *testing.T) {
		l := New(NewDevModeHandler(bytes.NewBuffer(make([]byte, 0, 10)), nil))
		l2 := l.With(slog.String("attr_1", "val1"))

		handler := NewDevModeHandler(bytes.NewBuffer(make([]byte, 0, 10)), nil)
		expected := &Logger{Logger: slog.New(handler.WithAttrs([]slog.Attr{slog.String("attr_1", "val1")}))}

		assert.Equal(t, expected, l2)
	})

	t.Run("WithGroup", func(t *testing.T) {
		l := New(NewDevModeHandler(bytes.NewBuffer(make([]byte, 0, 10)), nil))
		l2 := l.WithGroup("group_name")

		handler := NewDevModeHandler(bytes.NewBuffer(make([]byte, 0, 10)), nil)
		expected := &Logger{Logger: slog.New(handler.WithGroup("group_name"))}

		assert.Equal(t, expected, l2)
	})

	t.Run("WithContext", func(t *testing.T) {
		l := New(NewDevModeHandler(bytes.NewBuffer(make([]byte, 0, 10)), nil))
		l2 := l.WithContext(t.Context())

		assert.NotSame(t, l, l2)
		assert.Same(t, t.Context(), l2.ctx)
	})

	t.Run("Log_with_context", func(t *testing.T) {
		type ctxKey struct{}
		ctx := context.WithValue(t.Context(), ctxKey{}, "test-value")
		buf := bytes.NewBuffer(make([]byte, 0, 1024))
		handler := &proxyHandler{
			DevModeHandler: NewDevModeHandler(buf, &DevModeHandlerOptions{Level: slog.LevelDebug}),
			ctxKey:         ctxKey{},
		}
		l := New(handler).WithContext(ctx)

		_, file, _, ok := runtime.Caller(0)
		if !assert.True(t, ok) {
			return
		}
		expectedSource := fmt.Sprintf("%s:\\d{1,3}", regexp.QuoteMeta(file))

		cases := []struct {
			f    func(msg string, args ...any)
			want *regexp.Regexp
			desc string
		}{
			{desc: "Debug", f: l.Debug, want: regexp.MustCompile(fmt.Sprintf(`^\n%s DEBUG %s \d{4}/\d{2}/\d{2} \d{2}:\d{2}:\d{2}\.\d{1,6}%s \(%s\)%s\nmessage%s\n%sattr: %sval\n%sCONTEXT_VALUE: %stest-value\n$`, regexp.QuoteMeta(BGCyan+WhiteBold), regexp.QuoteMeta(Reset), regexp.QuoteMeta(Gray), expectedSource, regexp.QuoteMeta(Reset), regexp.QuoteMeta(Reset), regexp.QuoteMeta(WhiteBold), regexp.QuoteMeta(Reset), regexp.QuoteMeta(WhiteBold), regexp.QuoteMeta(Reset)))},
			{desc: "Info", f: l.Info, want: regexp.MustCompile(fmt.Sprintf(`^\n%s INFO %s \d{4}/\d{2}/\d{2} \d{2}:\d{2}:\d{2}\.\d{1,6}%s \(%s\)%s\nmessage%s\n%sattr: %sval\n%sCONTEXT_VALUE: %stest-value\n$`, regexp.QuoteMeta(BGGray+WhiteBold), regexp.QuoteMeta(Reset), regexp.QuoteMeta(Gray), expectedSource, regexp.QuoteMeta(Reset), regexp.QuoteMeta(Reset), regexp.QuoteMeta(WhiteBold), regexp.QuoteMeta(Reset), regexp.QuoteMeta(WhiteBold), regexp.QuoteMeta(Reset)))},
			{desc: "Warn", f: l.Warn, want: regexp.MustCompile(fmt.Sprintf(`^\n%s WARN %s \d{4}/\d{2}/\d{2} \d{2}:\d{2}:\d{2}\.\d{1,6}%s \(%s\)%s\n%smessage%s\n%sattr: %sval\n%sCONTEXT_VALUE: %stest-value\n$`, regexp.QuoteMeta(BGYellow+GrayBold), regexp.QuoteMeta(Reset), regexp.QuoteMeta(Gray), expectedSource, regexp.QuoteMeta(Reset), regexp.QuoteMeta(Yellow), regexp.QuoteMeta(Reset), regexp.QuoteMeta(WhiteBold), regexp.QuoteMeta(Reset), regexp.QuoteMeta(WhiteBold), regexp.QuoteMeta(Reset)))},
		}

		for _, c := range cases {
			t.Run(c.desc, func(t *testing.T) {
				c.f("message", slog.String("attr", "val"))

				assert.Regexp(t, c.want, buf.String())
				buf.Reset() // Subtests cannot be run in parallel here!
			})
		}
	})

	t.Run("Log_with_source", func(t *testing.T) {
		pc, file, line, ok := runtime.Caller(0)
		if !assert.True(t, ok) {
			return
		}
		expectedSource := regexp.QuoteMeta(fmt.Sprintf("%s:%d", file, line))

		buf := bytes.NewBuffer(make([]byte, 0, 1024))
		l := New(NewDevModeHandler(buf, &DevModeHandlerOptions{Level: slog.LevelDebug}))

		cases := []struct {
			f    func(ctx context.Context, source uintptr, msg string, args ...any)
			want *regexp.Regexp
			desc string
		}{
			{desc: "DebugWithSource", f: l.DebugWithSource, want: regexp.MustCompile(fmt.Sprintf(`^\n%s DEBUG %s \d{4}/\d{2}/\d{2} \d{2}:\d{2}:\d{2}\.\d{1,6}%s \(%s\)%s\nmessage%s\n%sattr: %sval\n$`, regexp.QuoteMeta(BGCyan+WhiteBold), regexp.QuoteMeta(Reset), regexp.QuoteMeta(Gray), expectedSource, regexp.QuoteMeta(Reset), regexp.QuoteMeta(Reset), regexp.QuoteMeta(WhiteBold), regexp.QuoteMeta(Reset)))},
			{desc: "InfoWithSource", f: l.InfoWithSource, want: regexp.MustCompile(fmt.Sprintf(`^\n%s INFO %s \d{4}/\d{2}/\d{2} \d{2}:\d{2}:\d{2}\.\d{1,6}%s \(%s\)%s\nmessage%s\n%sattr: %sval\n$`, regexp.QuoteMeta(BGGray+WhiteBold), regexp.QuoteMeta(Reset), regexp.QuoteMeta(Gray), expectedSource, regexp.QuoteMeta(Reset), regexp.QuoteMeta(Reset), regexp.QuoteMeta(WhiteBold), regexp.QuoteMeta(Reset)))},
			{desc: "WarnWithSource", f: l.WarnWithSource, want: regexp.MustCompile(fmt.Sprintf(`^\n%s WARN %s \d{4}/\d{2}/\d{2} \d{2}:\d{2}:\d{2}\.\d{1,6}%s \(%s\)%s\n%smessage%s\n%sattr: %sval\n$`, regexp.QuoteMeta(BGYellow+GrayBold), regexp.QuoteMeta(Reset), regexp.QuoteMeta(Gray), expectedSource, regexp.QuoteMeta(Reset), regexp.QuoteMeta(Yellow), regexp.QuoteMeta(Reset), regexp.QuoteMeta(WhiteBold), regexp.QuoteMeta(Reset)))},
		}

		for _, c := range cases {
			t.Run(c.desc, func(t *testing.T) {
				c.f(nil, pc, "message", slog.String("attr", "val"))

				assert.Regexp(t, c.want, buf.String())
				buf.Reset() // Subtests cannot be run in parallel here!
			})
		}
	})

	t.Run("Log_Error", func(t *testing.T) {
		type ctxKey struct{}
		ctx := context.WithValue(t.Context(), ctxKey{}, "test-value")

		buf := bytes.NewBuffer(make([]byte, 0, 1024))
		handler := &proxyHandler{
			DevModeHandler: NewDevModeHandler(buf, &DevModeHandlerOptions{Level: slog.LevelDebug}),
			ctxKey:         ctxKey{},
		}
		l := New(handler).WithContext(t.Context())

		pc, file, line, ok := runtime.Caller(0)
		if !assert.True(t, ok) {
			return
		}
		expectedSource := fmt.Sprintf("%s:\\d{1,3}", regexp.QuoteMeta(file))

		cases := []struct {
			f    func()
			want *regexp.Regexp
			desc string
		}{
			{
				desc: "Error",
				f:    func() { l.Error(fmt.Errorf("err message"), slog.String("attr", "val")) },
				want: regexp.MustCompile(fmt.Sprintf(`^\n%s ERROR %s \d{4}/\d{2}/\d{2} \d{2}:\d{2}:\d{2}\.\d{1,6}%s \(%s\)%s\n%serr message%s\n%sattr: %sval\n$`, regexp.QuoteMeta(BGRed+WhiteBold), regexp.QuoteMeta(Reset), regexp.QuoteMeta(Gray), expectedSource, regexp.QuoteMeta(Reset), regexp.QuoteMeta(Red), regexp.QuoteMeta(Reset), regexp.QuoteMeta(WhiteBold), regexp.QuoteMeta(Reset))),
			},
			{
				desc: "Error_WithContext",
				f:    func() { l.WithContext(ctx).Error(fmt.Errorf("err message"), slog.String("attr", "val")) },
				want: regexp.MustCompile(fmt.Sprintf(`^\n%s ERROR %s \d{4}/\d{2}/\d{2} \d{2}:\d{2}:\d{2}\.\d{1,6}%s \(%s\)%s\n%serr message%s\n%sattr: %sval\n%sCONTEXT_VALUE: %stest-value\n$`, regexp.QuoteMeta(BGRed+WhiteBold), regexp.QuoteMeta(Reset), regexp.QuoteMeta(Gray), expectedSource, regexp.QuoteMeta(Reset), regexp.QuoteMeta(Red), regexp.QuoteMeta(Reset), regexp.QuoteMeta(WhiteBold), regexp.QuoteMeta(Reset), regexp.QuoteMeta(WhiteBold), regexp.QuoteMeta(Reset))),
			},
			{
				desc: "nil Error",
				f:    func() { l.Error(errwrap.New(nil), slog.String("attr", "val")) },
				want: regexp.MustCompile(fmt.Sprintf(`^\n%s ERROR %s \d{4}/\d{2}/\d{2} \d{2}:\d{2}:\d{2}\.\d{1,6}%s \(%s\)%s\n%s<nil>%s\n%sattr: %sval\n$`, regexp.QuoteMeta(BGRed+WhiteBold), regexp.QuoteMeta(Reset), regexp.QuoteMeta(Gray), expectedSource, regexp.QuoteMeta(Reset), regexp.QuoteMeta(Red), regexp.QuoteMeta(Reset), regexp.QuoteMeta(WhiteBold), regexp.QuoteMeta(Reset))),
			},
			{
				desc: "ErrorContext",
				f:    func() { l.ErrorContext(ctx, fmt.Errorf("err message"), slog.String("attr", "val")) },
				want: regexp.MustCompile(fmt.Sprintf(`^\n%s ERROR %s \d{4}/\d{2}/\d{2} \d{2}:\d{2}:\d{2}\.\d{1,6}%s \(%s\)%s\n%serr message%s\n%sattr: %sval\n%sCONTEXT_VALUE: %stest-value\n$`, regexp.QuoteMeta(BGRed+WhiteBold), regexp.QuoteMeta(Reset), regexp.QuoteMeta(Gray), expectedSource, regexp.QuoteMeta(Reset), regexp.QuoteMeta(Red), regexp.QuoteMeta(Reset), regexp.QuoteMeta(WhiteBold), regexp.QuoteMeta(Reset), regexp.QuoteMeta(WhiteBold), regexp.QuoteMeta(Reset))),
			},
			{
				desc: "ErrorWithSource",
				f: func() {
					// Ignore "do not pass a nil Context" so we know passing a nil context doesn't crash
					l.ErrorWithSource(nil, pc, fmt.Errorf("err message"), slog.String("attr", "val")) //nolint:staticcheck
				},
				want: regexp.MustCompile(fmt.Sprintf(`^\n%s ERROR %s \d{4}/\d{2}/\d{2} \d{2}:\d{2}:\d{2}\.\d{1,6}%s \(%s\)%s\n%serr message%s\n%sattr: %sval\n$`, regexp.QuoteMeta(BGRed+WhiteBold), regexp.QuoteMeta(Reset), regexp.QuoteMeta(Gray), regexp.QuoteMeta(fmt.Sprintf("%s:%d", file, line)), regexp.QuoteMeta(Reset), regexp.QuoteMeta(Red), regexp.QuoteMeta(Reset), regexp.QuoteMeta(WhiteBold), regexp.QuoteMeta(Reset))),
			},
			{
				desc: "errors.Error",
				f:    func() { l.Error(errwrap.New("err message"), slog.String("attr", "val")) },
				want: regexp.MustCompile(fmt.Sprintf(`^\n%s ERROR %s \d{4}/\d{2}/\d{2} \d{2}:\d{2}:\d{2}\.\d{1,6}%s \(%s\)%s\n%serr message%s\n%sattr: %sval\n%strace: \n%s(.|\n)+\n$`, regexp.QuoteMeta(BGRed+WhiteBold), regexp.QuoteMeta(Reset), regexp.QuoteMeta(Gray), expectedSource, regexp.QuoteMeta(Reset), regexp.QuoteMeta(Red), regexp.QuoteMeta(Reset), regexp.QuoteMeta(WhiteBold), regexp.QuoteMeta(Reset), regexp.QuoteMeta(WhiteBold), regexp.QuoteMeta(Reset))),
			},
			{
				desc: "errors.Error_empty",
				f:    func() { l.Error(&errwrap.Error{}, slog.String("attr", "val")) },
				want: regexp.MustCompile(fmt.Sprintf(`^\n%s ERROR %s \d{4}/\d{2}/\d{2} \d{2}:\d{2}:\d{2}\.\d{1,6}%s \(%s\)%s\n%sgoyave.dev/goyave/util/errwrap\.Error: the Error doesn't wrap any reason \(empty reasons slice\)%s\n%sattr: %sval\n%strace: %s\n$`, regexp.QuoteMeta(BGRed+WhiteBold), regexp.QuoteMeta(Reset), regexp.QuoteMeta(Gray), expectedSource, regexp.QuoteMeta(Reset), regexp.QuoteMeta(Red), regexp.QuoteMeta(Reset), regexp.QuoteMeta(WhiteBold), regexp.QuoteMeta(Reset), regexp.QuoteMeta(WhiteBold), regexp.QuoteMeta(Reset))),
			},
			{
				desc: "errors.Error_multiple", // We expect three separated messages to be printed
				f: func() {
					l.Error(errwrap.New([]any{fmt.Errorf("err message"), errwrap.New("nested error"), "reason"}), slog.String("attr", "val"))
				},
				want: regexp.MustCompile(fmt.Sprintf(
					`^\n%s ERROR %s \d{4}/\d{2}/\d{2} \d{2}:\d{2}:\d{2}\.\d{1,6}%s \(%s\)%s\n%serr message%s\n%sattr: %sval\n%strace: \n%s(.|\n)+\n\n%s ERROR %s \d{4}/\d{2}/\d{2} \d{2}:\d{2}:\d{2}\.\d{1,6}%s \(%s\)%s\n%snested error%s\n%sattr: %sval\n%strace: \n%s(.|\n)+\n\n%s ERROR %s \d{4}/\d{2}/\d{2} \d{2}:\d{2}:\d{2}\.\d{1,6}%s \(%s\)%s\n%sreason%s\n%sattr: %sval\n%strace: \n%s(.|\n)+\n$`,
					regexp.QuoteMeta(BGRed+WhiteBold), regexp.QuoteMeta(Reset), regexp.QuoteMeta(Gray), expectedSource, regexp.QuoteMeta(Reset), regexp.QuoteMeta(Red), regexp.QuoteMeta(Reset), regexp.QuoteMeta(WhiteBold), regexp.QuoteMeta(Reset), regexp.QuoteMeta(WhiteBold), regexp.QuoteMeta(Reset),
					regexp.QuoteMeta(BGRed+WhiteBold), regexp.QuoteMeta(Reset), regexp.QuoteMeta(Gray), expectedSource, regexp.QuoteMeta(Reset), regexp.QuoteMeta(Red), regexp.QuoteMeta(Reset), regexp.QuoteMeta(WhiteBold), regexp.QuoteMeta(Reset), regexp.QuoteMeta(WhiteBold), regexp.QuoteMeta(Reset),
					regexp.QuoteMeta(BGRed+WhiteBold), regexp.QuoteMeta(Reset), regexp.QuoteMeta(Gray), expectedSource, regexp.QuoteMeta(Reset), regexp.QuoteMeta(Red), regexp.QuoteMeta(Reset), regexp.QuoteMeta(WhiteBold), regexp.QuoteMeta(Reset), regexp.QuoteMeta(WhiteBold), regexp.QuoteMeta(Reset),
				)),
			},
			{
				desc: "Valuer",
				f:    func() { l.Error(errwrap.New(testValuerError{}), slog.String("attr", "val")) },
				want: regexp.MustCompile(fmt.Sprintf(`^\n%s ERROR %s \d{4}/\d{2}/\d{2} \d{2}:\d{2}:\d{2}\.\d{1,6}%s \(%s\)%s\n%stest error%s\n%sattr: %sval\n%strace: \n%s(.|\n)+\n%sreason: %stest value\n`, regexp.QuoteMeta(BGRed+WhiteBold), regexp.QuoteMeta(Reset), regexp.QuoteMeta(Gray), expectedSource, regexp.QuoteMeta(Reset), regexp.QuoteMeta(Red), regexp.QuoteMeta(Reset), regexp.QuoteMeta(WhiteBold), regexp.QuoteMeta(Reset), regexp.QuoteMeta(WhiteBold), regexp.QuoteMeta(Reset), regexp.QuoteMeta(WhiteBold), regexp.QuoteMeta(Reset))),
			},
			{
				desc: "fmt_multi_error", // We expect three separated messages to be printed
				f: func() {
					l.Error(fmt.Errorf("err message: %w, %w", fmt.Errorf("nested error"), fmt.Errorf("reason")), slog.String("attr", "val"))
				},
				want: regexp.MustCompile(fmt.Sprintf(
					`^\n%s ERROR %s \d{4}/\d{2}/\d{2} \d{2}:\d{2}:\d{2}\.\d{1,6}%s \(%s\)%s\n%serr message: nested error, reason%s\n%sattr: %sval\n\n%s ERROR %s \d{4}/\d{2}/\d{2} \d{2}:\d{2}:\d{2}\.\d{1,6}%s \(%s\)%s\n%snested error%s\n%sattr: %sval\n\n%s ERROR %s \d{4}/\d{2}/\d{2} \d{2}:\d{2}:\d{2}\.\d{1,6}%s \(%s\)%s\n%sreason%s\n%sattr: %sval\n$`,
					regexp.QuoteMeta(BGRed+WhiteBold), regexp.QuoteMeta(Reset), regexp.QuoteMeta(Gray), expectedSource, regexp.QuoteMeta(Reset), regexp.QuoteMeta(Red), regexp.QuoteMeta(Reset), regexp.QuoteMeta(WhiteBold), regexp.QuoteMeta(Reset),
					regexp.QuoteMeta(BGRed+WhiteBold), regexp.QuoteMeta(Reset), regexp.QuoteMeta(Gray), expectedSource, regexp.QuoteMeta(Reset), regexp.QuoteMeta(Red), regexp.QuoteMeta(Reset), regexp.QuoteMeta(WhiteBold), regexp.QuoteMeta(Reset),
					regexp.QuoteMeta(BGRed+WhiteBold), regexp.QuoteMeta(Reset), regexp.QuoteMeta(Gray), expectedSource, regexp.QuoteMeta(Reset), regexp.QuoteMeta(Red), regexp.QuoteMeta(Reset), regexp.QuoteMeta(WhiteBold), regexp.QuoteMeta(Reset),
				)),
			},
			// reason appended to attrs if not in dev mode
		}

		for _, c := range cases {
			t.Run(c.desc, func(t *testing.T) {
				c.f()
				assert.Regexp(t, c.want, buf.String())
				buf.Reset() // Subtests cannot be run in parallel here!
			})
		}
	})

	t.Run("Log_Error_reason_in_prod", func(t *testing.T) {
		// When dev mode is disabled, a "reason" attr should be added to the record when printing an error

		pc, file, line, ok := runtime.Caller(0)
		if !assert.True(t, ok) {
			return
		}

		buf := bytes.NewBuffer(make([]byte, 0, 1024))
		l := New(slog.NewJSONHandler(buf, &slog.HandlerOptions{AddSource: true}))

		err := errwrap.New("reason")
		r := regexp.MustCompile(fmt.Sprintf(`{"time":"\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{1,9}((\+\d{2}:\d{2})|Z)?","level":"ERROR","source":{"function":".+","file":"%s","line":%d},"msg":"reason","trace":".+","reason":"reason"}\n`, regexp.QuoteMeta(file), line))

		l.ErrorWithSource(context.Background(), pc, err)
		assert.Regexp(t, r, buf.String())
	})

	t.Run("DiscardLogger", func(t *testing.T) {
		logger := DiscardLogger()

		assert.NotPanics(t, func() {
			logger.Info("This log should be discarded")
		})
	})

	t.Run("Default", func(t *testing.T) {
		assert.NotNil(t, Default())
		logger := New(NewHandler(false, os.Stderr))
		SetDefault(logger)
		assert.Same(t, logger, Default())

		t.Setenv("ENV", "PROD")
		assert.True(t, isProduction())
		t.Setenv("ENV", "staging")
		assert.False(t, isProduction())
	})

	t.Run("Context", func(t *testing.T) {
		assert.Same(t, defaultLogger, FromContext(t.Context()))

		l := &Logger{}
		ctx := Context(t.Context(), l)
		l2 := FromContext(ctx)
		assert.NotSame(t, l, l2)
		assert.Equal(t, t.Context(), l2.ctx)
	})

	t.Run("Options", func(t *testing.T) {
		opts := []Option{
			WithContext(t.Context()),
			WithOpenTelemetry(true),
			WithOpenTelemetryOptions(otelslog.WithAttributes(attribute.Bool("test", true))),
		}
		logger := New(NewDevModeHandler(io.Discard, nil), opts...)
		assert.Equal(t, t.Context(), logger.ctx)
		assert.IsType(t, &slog.MultiHandler{}, logger.Handler()) // Cannot really check the rest unfortunately
	})
}
