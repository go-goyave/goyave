package slog

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

type testLogValuer int

func (t testLogValuer) LogValue() slog.Value {
	return slog.StringValue(fmt.Sprintf("value_%d", t))
}

func TestNewHandler(t *testing.T) {
	cases := []struct {
		want    any
		w       io.Writer
		devMode bool
	}{
		{
			devMode: true,
			w:       bytes.NewBuffer(make([]byte, 0, 10)),
			want:    &DevModeHandler{w: bytes.NewBuffer(make([]byte, 0, 10)), mu: &sync.Mutex{}, opts: &DevModeHandlerOptions{Level: slog.LevelDebug}},
		},
		{
			devMode: false,
			w:       bytes.NewBuffer(make([]byte, 0, 10)),
			want:    slog.NewJSONHandler(bytes.NewBuffer(make([]byte, 0, 10)), &slog.HandlerOptions{Level: slog.LevelInfo, AddSource: true}),
		},
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("devMode_%v", c.devMode), func(t *testing.T) {
			assert.Equal(t, c.want, NewHandler(c.devMode, c.w))
		})
	}
}

func TestDevModeHandler(t *testing.T) {
	t.Run("New", func(t *testing.T) {
		cases := []struct {
			opts *DevModeHandlerOptions
			want *DevModeHandler
			desc string
		}{
			{
				desc: "nil_options",
				opts: nil,
				want: &DevModeHandler{w: bytes.NewBuffer(make([]byte, 0, 10)), mu: &sync.Mutex{}, opts: &DevModeHandlerOptions{}},
			},
			{
				desc: "log_level_options",
				opts: &DevModeHandlerOptions{Level: slog.LevelError},
				want: &DevModeHandler{w: bytes.NewBuffer(make([]byte, 0, 10)), mu: &sync.Mutex{}, opts: &DevModeHandlerOptions{Level: slog.LevelError}},
			},
		}

		for _, c := range cases {
			t.Run(c.desc, func(t *testing.T) {
				buf := bytes.NewBuffer(make([]byte, 0, 10))
				handler := NewDevModeHandler(buf, c.opts)
				assert.Equal(t, c.want, handler)
			})
		}
	})

	t.Run("Enabled", func(t *testing.T) {
		cases := []struct {
			opts  *DevModeHandlerOptions
			level slog.Level
			want  bool
		}{
			{
				opts:  &DevModeHandlerOptions{}, // Default level
				level: slog.LevelDebug,
				want:  false,
			},
			{
				opts:  &DevModeHandlerOptions{}, // Default level
				level: slog.LevelInfo,
				want:  true,
			},
			{
				opts:  &DevModeHandlerOptions{Level: slog.LevelDebug},
				level: slog.LevelDebug,
				want:  true,
			},
			{
				opts:  &DevModeHandlerOptions{Level: slog.LevelDebug},
				level: slog.LevelError,
				want:  true,
			},
			{
				opts:  &DevModeHandlerOptions{Level: slog.LevelWarn},
				level: slog.LevelError,
				want:  true,
			},
			{
				opts:  &DevModeHandlerOptions{Level: slog.LevelWarn},
				level: slog.LevelWarn,
				want:  true,
			},
			{
				opts:  &DevModeHandlerOptions{Level: slog.LevelWarn},
				level: slog.LevelInfo,
				want:  false,
			},
		}

		for _, c := range cases {
			t.Run(fmt.Sprintf("%s_%s", c.opts.Level, c.level), func(t *testing.T) {
				buf := bytes.NewBuffer(make([]byte, 0, 10))
				handler := NewDevModeHandler(buf, c.opts)
				assert.Equal(t, c.want, handler.Enabled(context.Background(), c.level))
			})
		}
	})

	t.Run("WithAttrs", func(t *testing.T) {
		buf := bytes.NewBuffer(make([]byte, 0, 10))

		cases := []struct {
			desc       string
			want       *DevModeHandler
			baseGroups []string
			attrs      []slog.Attr
			baseAttrs  []slog.Attr
		}{
			{
				desc: "copy_empty",
				want: &DevModeHandler{opts: &DevModeHandlerOptions{}, w: buf, mu: &sync.Mutex{}, attrs: []slog.Attr{}},
			},
			{
				desc:       "copy",
				baseAttrs:  []slog.Attr{slog.String("a", "a"), slog.Int("b", 2)},
				baseGroups: []string{"group"},
				want:       &DevModeHandler{opts: &DevModeHandlerOptions{}, w: buf, mu: &sync.Mutex{}, attrs: []slog.Attr{slog.String("a", "a"), slog.Int("b", 2)}, groups: []string{"group"}},
			},
			{
				desc:  "append_empty",
				attrs: []slog.Attr{slog.String("a", "a"), slog.Int("b", 2)},
				want:  &DevModeHandler{opts: &DevModeHandlerOptions{}, w: buf, mu: &sync.Mutex{}, attrs: []slog.Attr{slog.String("a", "a"), slog.Int("b", 2)}},
			},
			{
				desc:      "append",
				baseAttrs: []slog.Attr{slog.String("a", "a"), slog.Int("b", 2)},
				attrs:     []slog.Attr{slog.String("c", "c"), slog.Int("d", 3)},
				want:      &DevModeHandler{opts: &DevModeHandlerOptions{}, w: buf, mu: &sync.Mutex{}, attrs: []slog.Attr{slog.String("a", "a"), slog.Int("b", 2), slog.String("c", "c"), slog.Int("d", 3)}},
			},
			{
				desc:       "append_with_group",
				baseGroups: []string{"group"},
				baseAttrs:  []slog.Attr{slog.String("a", "a"), slog.Int("b", 2)},
				attrs:      []slog.Attr{slog.String("c", "c"), slog.Int("d", 3)},
				want:       &DevModeHandler{opts: &DevModeHandlerOptions{}, w: buf, mu: &sync.Mutex{}, attrs: []slog.Attr{slog.String("a", "a"), slog.Int("b", 2), slog.String("c", "c"), slog.Int("d", 3)}, groups: []string{"group"}},
			},
		}

		for _, c := range cases {
			t.Run(c.desc, func(t *testing.T) {
				handler := NewDevModeHandler(buf, &DevModeHandlerOptions{})
				handler.attrs = c.baseAttrs
				handler.groups = c.baseGroups

				assert.Equal(t, c.want, handler.WithAttrs(c.attrs))
			})
		}
	})

	t.Run("WithGroup", func(t *testing.T) {
		buf := bytes.NewBuffer(make([]byte, 0, 10))

		cases := []struct {
			desc       string
			group      string
			want       *DevModeHandler
			baseGroups []string
			baseAttrs  []slog.Attr
		}{
			{
				desc:      "copy",
				baseAttrs: []slog.Attr{slog.String("a", "a"), slog.Int("b", 2)},
				group:     "group",
				want:      &DevModeHandler{opts: &DevModeHandlerOptions{}, w: buf, mu: &sync.Mutex{}, attrs: []slog.Attr{slog.String("a", "a"), slog.Int("b", 2)}, groups: []string{"group"}},
			},
			{
				desc:       "append",
				baseGroups: []string{"base"},
				group:      "group",
				want:       &DevModeHandler{opts: &DevModeHandlerOptions{}, w: buf, mu: &sync.Mutex{}, attrs: []slog.Attr{}, groups: []string{"base", "group"}},
			},
			{
				desc:       "append_with_attrs",
				baseGroups: []string{"base"},
				baseAttrs:  []slog.Attr{slog.String("a", "a"), slog.Int("b", 2)},
				group:      "group",
				want:       &DevModeHandler{opts: &DevModeHandlerOptions{}, w: buf, mu: &sync.Mutex{}, attrs: []slog.Attr{slog.String("a", "a"), slog.Int("b", 2)}, groups: []string{"base", "group"}},
			},
		}

		for _, c := range cases {
			t.Run(c.desc, func(t *testing.T) {
				handler := NewDevModeHandler(buf, &DevModeHandlerOptions{})
				handler.attrs = c.baseAttrs
				handler.groups = c.baseGroups

				assert.Equal(t, c.want, handler.WithGroup(c.group))
			})
		}
	})
}

func TestDevModeHandlerFormat(t *testing.T) {
	t.Run("levelColor", func(t *testing.T) {
		cases := []struct {
			want  string
			level slog.Level
		}{
			{level: slog.LevelDebug, want: BGCyan + WhiteBold},
			{level: -5, want: BGCyan + WhiteBold},
			{level: -1, want: BGCyan + WhiteBold},
			{level: slog.LevelInfo, want: BGGray + WhiteBold},
			{level: 1, want: BGGray + WhiteBold},
			{level: 3, want: BGGray + WhiteBold},
			{level: slog.LevelWarn, want: BGYellow + GrayBold},
			{level: 5, want: BGYellow + GrayBold},
			{level: 7, want: BGYellow + GrayBold},
			{level: slog.LevelError, want: BGRed + WhiteBold},
			{level: 9, want: BGRed + WhiteBold},
		}

		for _, c := range cases {
			t.Run(c.level.String(), func(t *testing.T) {
				assert.Equal(t, c.want, levelColor(c.level))
			})
		}
	})

	t.Run("messageColor", func(t *testing.T) {
		cases := []struct {
			want  string
			level slog.Level
		}{
			{level: slog.LevelDebug, want: ""},
			{level: -5, want: ""},
			{level: -1, want: ""},
			{level: slog.LevelInfo, want: ""},
			{level: 1, want: ""},
			{level: 3, want: ""},
			{level: slog.LevelWarn, want: Yellow},
			{level: 5, want: Yellow},
			{level: 7, want: Yellow},
			{level: slog.LevelError, want: Red},
			{level: 9, want: Red},
		}

		for _, c := range cases {
			t.Run(c.level.String(), func(t *testing.T) {
				assert.Equal(t, c.want, messageColor(c.level))
			})
		}
	})

	t.Run("Handle", func(t *testing.T) {
		time := lo.Must(time.Parse(time.RFC3339Nano, "2023-04-09T15:04:05.123456789Z"))

		pc, file, line, ok := runtime.Caller(1)
		if !assert.True(t, ok) {
			return
		}

		expectedSource := fmt.Sprintf("%s:%d", file, line)

		cases := []struct {
			r    func() slog.Record                      // Create the record
			h    func(h *DevModeHandler) *DevModeHandler // Alters the handler, if provided (can be nil)
			want string
			desc string
		}{
			{
				desc: "level_debug",
				r:    func() slog.Record { return slog.NewRecord(time, slog.LevelDebug, "message", pc) },
				want: fmt.Sprintf("\n%s DEBUG %s 2023/04/09 15:04:05.123456%s (%s)%s\nmessage%s\n", BGCyan+WhiteBold, Reset, Gray, expectedSource, Reset, Reset),
			},
			{
				desc: "level_info",
				r:    func() slog.Record { return slog.NewRecord(time, slog.LevelInfo, "message", pc) },
				want: fmt.Sprintf("\n%s INFO %s 2023/04/09 15:04:05.123456%s (%s)%s\nmessage%s\n", BGGray+WhiteBold, Reset, Gray, expectedSource, Reset, Reset),
			},
			{
				desc: "level_warn",
				r:    func() slog.Record { return slog.NewRecord(time, slog.LevelWarn, "message", pc) },
				want: fmt.Sprintf("\n%s WARN %s 2023/04/09 15:04:05.123456%s (%s)%s\n%smessage%s\n", BGYellow+GrayBold, Reset, Gray, expectedSource, Reset, Yellow, Reset),
			},
			{
				desc: "level_error",
				r:    func() slog.Record { return slog.NewRecord(time, slog.LevelError, "message", pc) },
				want: fmt.Sprintf("\n%s ERROR %s 2023/04/09 15:04:05.123456%s (%s)%s\n%smessage%s\n", BGRed+WhiteBold, Reset, Gray, expectedSource, Reset, Red, Reset),
			},
			{
				desc: "one_attr",
				r: func() slog.Record {
					r := slog.NewRecord(time, slog.LevelError, "message", pc)
					r.AddAttrs(slog.String("attr_name", "attr_value"))
					return r
				},
				want: fmt.Sprintf("\n%s ERROR %s 2023/04/09 15:04:05.123456%s (%s)%s\n%smessage%s\n%sattr_name: %sattr_value\n", BGRed+WhiteBold, Reset, Gray, expectedSource, Reset, Red, Reset, WhiteBold, Reset),
			},
			{
				desc: "one_attr_with_line_breaks",
				r: func() slog.Record {
					r := slog.NewRecord(time, slog.LevelError, "message", pc)
					r.AddAttrs(slog.String("attr_name", "attr_value\non several\nlines"))
					return r
				},
				want: fmt.Sprintf("\n%s ERROR %s 2023/04/09 15:04:05.123456%s (%s)%s\n%smessage%s\n%sattr_name: \n%sattr_value\non several\nlines\n", BGRed+WhiteBold, Reset, Gray, expectedSource, Reset, Red, Reset, WhiteBold, Reset),
			},
			{
				desc: "two_attrs",
				r: func() slog.Record {
					r := slog.NewRecord(time, slog.LevelError, "message", pc)
					r.AddAttrs(slog.String("attr_name", "attr_value"), slog.Int("number", 123))
					return r
				},
				want: fmt.Sprintf("\n%s ERROR %s 2023/04/09 15:04:05.123456%s (%s)%s\n%smessage%s\n%sattr_name: %sattr_value\n%snumber: %s123\n", BGRed+WhiteBold, Reset, Gray, expectedSource, Reset, Red, Reset, WhiteBold, Reset, WhiteBold, Reset),
			},
			{
				desc: "group_and_subgroup",
				r: func() slog.Record {
					r := slog.NewRecord(time, slog.LevelError, "message", pc)
					r.AddAttrs(slog.Group("group_name", slog.String("attr_1", "val1"), slog.Int("attr_2", 123), slog.Group("subgroup", slog.String("attr_3", "val3"))))
					return r
				},
				want: fmt.Sprintf("\n%s ERROR %s 2023/04/09 15:04:05.123456%s (%s)%s\n%smessage%s\n%sgroup_name: \n  %sattr_1: %sval1\n  %sattr_2: %s123\n  %ssubgroup: \n    %sattr_3: %sval3\n", BGRed+WhiteBold, Reset, Gray, expectedSource, Reset, Red, Reset, WhiteBold, WhiteBold, Reset, WhiteBold, Reset, WhiteBold, WhiteBold, Reset),
			},
			{
				desc: "attr_and_group",
				r: func() slog.Record {
					r := slog.NewRecord(time, slog.LevelError, "message", pc)
					r.AddAttrs(slog.Group("group_name", slog.String("attr_1", "val1"), slog.Int("attr_2", 123)), slog.String("root_attr", "root_value"))
					return r
				},
				want: fmt.Sprintf("\n%s ERROR %s 2023/04/09 15:04:05.123456%s (%s)%s\n%smessage%s\n%sgroup_name: \n  %sattr_1: %sval1\n  %sattr_2: %s123\n%sroot_attr: %sroot_value\n", BGRed+WhiteBold, Reset, Gray, expectedSource, Reset, Red, Reset, WhiteBold, WhiteBold, Reset, WhiteBold, Reset, WhiteBold, Reset),
			},
			{
				desc: "with_one_handler_attr",
				r: func() slog.Record {
					r := slog.NewRecord(time, slog.LevelError, "message", pc)
					r.AddAttrs(slog.String("attr_name", "attr_value"))
					return r
				},
				h: func(h *DevModeHandler) *DevModeHandler {
					return h.WithAttrs([]slog.Attr{slog.String("handler_attr", "handler_value")}).(*DevModeHandler)
				},
				want: fmt.Sprintf("\n%s ERROR %s 2023/04/09 15:04:05.123456%s (%s)%s\n%smessage%s\n%shandler_attr: %shandler_value\n%sattr_name: %sattr_value\n", BGRed+WhiteBold, Reset, Gray, expectedSource, Reset, Red, Reset, WhiteBold, Reset, WhiteBold, Reset),
			},
			{
				desc: "with_one_handler_attr_and_group",
				r: func() slog.Record {
					r := slog.NewRecord(time, slog.LevelError, "message", pc)
					r.AddAttrs(slog.String("attr_name", "attr_value"))
					return r
				},
				h: func(h *DevModeHandler) *DevModeHandler {
					return h.WithAttrs([]slog.Attr{slog.String("handler_attr", "handler_value")}).WithGroup("group").(*DevModeHandler)
				},
				want: fmt.Sprintf("\n%s ERROR %s 2023/04/09 15:04:05.123456%s (%s)%s\n%smessage%s\n%sgroup:\n  %shandler_attr: %shandler_value\n  %sattr_name: %sattr_value\n", BGRed+WhiteBold, Reset, Gray, expectedSource, Reset, Red, Reset, WhiteBold, WhiteBold, Reset, WhiteBold, Reset),
			},
			{
				desc: "generic_struct_conversion",
				r: func() slog.Record {
					type strct struct {
						Subgroup struct {
							Attr3 string
						}
						Attr1      string
						unexported string // Should not be visible
						Attr2      int
					}

					r := slog.NewRecord(time, slog.LevelError, "message", pc)
					s := &strct{
						Attr1:      "val1",
						Attr2:      123,
						unexported: "hidden",
						Subgroup:   struct{ Attr3 string }{Attr3: "val3"},
					}
					r.AddAttrs(slog.Any("struct", s))
					return r
				},
				want: fmt.Sprintf("\n%s ERROR %s 2023/04/09 15:04:05.123456%s (%s)%s\n%smessage%s\n%sstruct: \n  %sSubgroup: \n    %sAttr3: %sval3\n  %sAttr1: %sval1\n  %sAttr2: %s123\n", BGRed+WhiteBold, Reset, Gray, expectedSource, Reset, Red, Reset, WhiteBold, WhiteBold, WhiteBold, Reset, WhiteBold, Reset, WhiteBold, Reset),
			},
			{
				desc: "map_conversion",
				r: func() slog.Record {
					r := slog.NewRecord(time, slog.LevelError, "message", pc)
					m := map[string]any{
						"Subgroup": map[string]any{ // We leave only one key because map iteration order is not guaranteed, making the test randomly fail
							"Attr3": "val3",
						},
					}
					r.AddAttrs(slog.Any("map", m))
					return r
				},
				want: fmt.Sprintf("\n%s ERROR %s 2023/04/09 15:04:05.123456%s (%s)%s\n%smessage%s\n%smap: \n  %sSubgroup: \n    %sAttr3: %sval3\n", BGRed+WhiteBold, Reset, Gray, expectedSource, Reset, Red, Reset, WhiteBold, WhiteBold, WhiteBold, Reset),
			},
			{
				desc: "logvaluer",
				r: func() slog.Record {
					r := slog.NewRecord(time, slog.LevelError, "message", pc)
					r.AddAttrs(slog.Any("logvalue", StructValue(testLogValuer(123))))
					return r
				},
				want: fmt.Sprintf("\n%s ERROR %s 2023/04/09 15:04:05.123456%s (%s)%s\n%smessage%s\n%slogvalue: %svalue_123\n", BGRed+WhiteBold, Reset, Gray, expectedSource, Reset, Red, Reset, WhiteBold, Reset),
			},
		}

		for _, c := range cases {
			t.Run(c.desc, func(t *testing.T) {
				buf := bytes.NewBuffer(make([]byte, 0, 1024))
				handler := NewDevModeHandler(buf, &DevModeHandlerOptions{Level: slog.LevelDebug})

				if c.h != nil {
					handler = c.h(handler)
				}
				assert.NoError(t, handler.Handle(context.Background(), c.r()))

				assert.Equal(t, c.want, buf.String())
			})
		}
	})
}
