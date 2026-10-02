package database

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"testing"
	"time"

	"log/slog"

	"github.com/stretchr/testify/assert"
	"goyave.dev/goyave/v6/slogx"
)

func TestLogger(t *testing.T) {
	t.Run("New", func(t *testing.T) {
		l := NewLogger()
		assert.Equal(t, 200*time.Millisecond, l.SlowThreshold)
	})

	t.Run("LogMode", func(t *testing.T) {
		l := NewLogger()
		l2 := l.LogMode(0).(*Logger)
		assert.Equal(t, l.SlowThreshold, l2.SlowThreshold)
	})

	t.Run("Info", func(t *testing.T) {
		buf := bytes.NewBufferString("")
		slogger := slogx.New(slogx.NewHandler(false, buf))
		ctx := slogx.Context(t.Context(), slogger)
		l := NewLogger()

		l.Info(ctx, "message %d", 1)

		assert.Regexp(t, `{"time":"\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{1,9}((\+\d{2}:\d{2})|Z)?","level":"INFO","source":{"function":".+","file":".+","line":\d+},"msg":"message 1"}\n`, buf.String())
	})

	t.Run("Warn", func(t *testing.T) {
		buf := bytes.NewBuffer(make([]byte, 0, 1024))
		slogger := slogx.New(slogx.NewHandler(false, buf))
		ctx := slogx.Context(t.Context(), slogger)
		l := NewLogger()

		l.Warn(ctx, "message %d", 1)

		assert.Regexp(t, `{"time":"\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{1,9}((\+\d{2}:\d{2})|Z)?","level":"WARN","source":{"function":".+","file":".+","line":\d+},"msg":"message 1"}\n`, buf.String())
	})

	t.Run("Error", func(t *testing.T) {
		buf := bytes.NewBuffer(make([]byte, 0, 1024))
		slogger := slogx.New(slogx.NewHandler(false, buf))
		ctx := slogx.Context(t.Context(), slogger)
		l := NewLogger()

		l.Error(ctx, "message %d", 1)

		assert.Regexp(t, `{"time":"\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{1,9}((\+\d{2}:\d{2})|Z)?","level":"ERROR","source":{"function":".+","file":".+","line":\d+},"msg":"message 1"}\n`, buf.String())
	})

	t.Run("Trace", func(t *testing.T) {
		cases := []struct {
			want          *regexp.Regexp
			err           error
			begin         time.Time
			desc          string
			sql           string
			slowThreshold time.Duration
			rowsAffected  int64
			level         slog.Level
			wantEmpty     bool
		}{
			{
				desc:          "debug",
				begin:         time.Now(),
				sql:           "SELECT * FROM some_table",
				rowsAffected:  4,
				err:           nil,
				level:         slog.LevelDebug,
				slowThreshold: -1,
				want: regexp.MustCompile(
					fmt.Sprintf(`{"time":"\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{1,9}((\+\d{2}:\d{2})|Z)?","level":"DEBUG","msg":"%s"}\n`,
						fmt.Sprintf(`%s\[\d+\.\d+ms\] %s\[rows:4\]%s SELECT \* FROM some_table`, regexp.QuoteMeta(strings.ReplaceAll(slogx.Yellow, "\033", `\u001b`)), regexp.QuoteMeta(strings.ReplaceAll(slogx.Blue, "\033", `\u001b`)), regexp.QuoteMeta(strings.ReplaceAll(slogx.Reset, "\033", `\u001b`))),
					),
				),
			},
			{
				desc:          "debug_slow",
				begin:         time.Now().Add(-time.Second),
				sql:           "SELECT * FROM some_table",
				rowsAffected:  4,
				err:           nil,
				level:         slog.LevelDebug,
				slowThreshold: time.Millisecond * 200,
				want: regexp.MustCompile(
					fmt.Sprintf(`{"time":"\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{1,9}((\+\d{2}:\d{2})|Z)?","level":"WARN","msg":"%s"}\n`,
						fmt.Sprintf(`SLOW SQL >= 200ms\\n%s\[\d+\.\d+ms\] %s\[rows:4\]%s SELECT \* FROM some_table`, regexp.QuoteMeta(strings.ReplaceAll(slogx.Reset+slogx.Red, "\033", `\u001b`)), regexp.QuoteMeta(strings.ReplaceAll(slogx.Blue, "\033", `\u001b`)), regexp.QuoteMeta(strings.ReplaceAll(slogx.Reset, "\033", `\u001b`))),
					),
				),
			},
			{
				desc:          "debug_slow_disabled_threshold",
				begin:         time.Now().Add(-time.Second),
				sql:           "SELECT * FROM some_table",
				rowsAffected:  4,
				err:           nil,
				level:         slog.LevelDebug,
				slowThreshold: 0,
				want: regexp.MustCompile(
					fmt.Sprintf(`{"time":"\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{1,9}((\+\d{2}:\d{2})|Z)?","level":"DEBUG","msg":"%s"}\n`,
						fmt.Sprintf(`%s\[\d+\.\d+ms\] %s\[rows:4\]%s SELECT \* FROM some_table`, regexp.QuoteMeta(strings.ReplaceAll(slogx.Yellow, "\033", `\u001b`)), regexp.QuoteMeta(strings.ReplaceAll(slogx.Blue, "\033", `\u001b`)), regexp.QuoteMeta(strings.ReplaceAll(slogx.Reset, "\033", `\u001b`))),
					),
				),
			},
			{
				desc:          "warn_slow",
				begin:         time.Now().Add(-time.Second),
				sql:           "SELECT * FROM some_table",
				rowsAffected:  4,
				err:           nil,
				level:         slog.LevelWarn,
				slowThreshold: time.Millisecond * 200,
				want: regexp.MustCompile(
					fmt.Sprintf(`{"time":"\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{1,9}((\+\d{2}:\d{2})|Z)?","level":"WARN","msg":"%s"}\n`,
						fmt.Sprintf(`SLOW SQL >= 200ms\\n%s\[\d+\.\d+ms\] %s\[rows:4\]%s SELECT \* FROM some_table`, regexp.QuoteMeta(strings.ReplaceAll(slogx.Reset+slogx.Red, "\033", `\u001b`)), regexp.QuoteMeta(strings.ReplaceAll(slogx.Blue, "\033", `\u001b`)), regexp.QuoteMeta(strings.ReplaceAll(slogx.Reset, "\033", `\u001b`))),
					),
				),
			},
			{
				desc:          "error_slow",
				begin:         time.Now().Add(-time.Second),
				sql:           "SELECT * FROM some_table",
				rowsAffected:  4,
				err:           nil,
				level:         slog.LevelError,
				slowThreshold: -1,
				wantEmpty:     true,
			},
			{
				desc:          "error",
				begin:         time.Now().Add(-time.Second),
				sql:           "SELECT * FROM some_table",
				rowsAffected:  0,
				err:           fmt.Errorf("no such table: some_table"),
				level:         slog.LevelDebug,
				slowThreshold: -1,
				want: regexp.MustCompile(
					fmt.Sprintf(`{"time":"\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{1,9}((\+\d{2}:\d{2})|Z)?","level":"ERROR","msg":"%s"}\n`,
						fmt.Sprintf(`no such table: some_table\\n%s\[\d+\.\d+ms\] %s\[rows:0\]%s SELECT \* FROM some_table`, regexp.QuoteMeta(strings.ReplaceAll(slogx.Reset+slogx.Yellow, "\033", `\u001b`)), regexp.QuoteMeta(strings.ReplaceAll(slogx.Blue, "\033", `\u001b`)), regexp.QuoteMeta(strings.ReplaceAll(slogx.Reset, "\033", `\u001b`))),
					),
				),
			},
		}

		for _, c := range cases {
			t.Run(c.desc, func(t *testing.T) {
				buf := bytes.NewBuffer(make([]byte, 0, 1024))
				slogger := slogx.New(slog.NewJSONHandler(buf, &slog.HandlerOptions{Level: c.level}))
				ctx := slogx.Context(t.Context(), slogger)
				l := NewLogger()

				if c.slowThreshold > -1 {
					l.SlowThreshold = c.slowThreshold
				}

				l.Trace(ctx, c.begin, func() (sql string, rowsAffected int64) { return c.sql, c.rowsAffected }, c.err)

				if c.wantEmpty {
					assert.Empty(t, buf.String())
				} else {
					assert.Regexp(t, c.want, buf.String())
				}
			})
		}
	})
}

func TestDiscardLogger(t *testing.T) {
	t.Run("LogMode", func(t *testing.T) {
		l := NewDiscardLogger()
		assert.Same(t, l, l.LogMode(0))
	})

	t.Run("Info", func(t *testing.T) {
		l := NewDiscardLogger()
		buf := bytes.NewBufferString("")
		slogger := slogx.New(slogx.NewHandler(false, buf))
		ctx := slogx.Context(t.Context(), slogger)
		l.Info(ctx, "message")
		assert.Empty(t, buf.String())
	})

	t.Run("Warn", func(t *testing.T) {
		l := NewDiscardLogger()
		buf := bytes.NewBufferString("")
		slogger := slogx.New(slogx.NewHandler(false, buf))
		ctx := slogx.Context(t.Context(), slogger)
		l.Warn(ctx, "message")
		assert.Empty(t, buf.String())
	})

	t.Run("Error", func(t *testing.T) {
		l := NewDiscardLogger()
		buf := bytes.NewBufferString("")
		slogger := slogx.New(slogx.NewHandler(false, buf))
		ctx := slogx.Context(t.Context(), slogger)
		l.Error(ctx, "message")
		assert.Empty(t, buf.String())
	})
}
