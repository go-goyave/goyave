package slog

import (
	"bytes"
	"context"
	"io"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"log/slog"
)

// Colors
const (
	Reset       = "\033[0m"
	Red         = "\033[31m"
	Green       = "\033[32m"
	Yellow      = "\033[33m"
	Blue        = "\033[34m"
	Magenta     = "\033[35m"
	Cyan        = "\033[36m"
	CyanBold    = "\033[36;1m"
	White       = "\033[37m"
	Gray        = "\033[90m"
	WhiteBold   = "\033[37;1m"
	BlueBold    = "\033[34;1m"
	MagentaBold = "\033[35;1m"
	RedBold     = "\033[31;1m"
	YellowBold  = "\033[33;1m"
)

var (
	// Indent the string used to indent attribute groups in the dev mode handler.
	Indent = "  "
)

// DevModeHandlerOptions options for the dev mode handler.
type DevModeHandlerOptions struct {
	// Level reports the minimum record level that will be logged.
	// The handler discards records with lower levels.
	// If Level is nil, the handler assumes `LevelInfo`.
	// The handler calls `Level.Level()` for each record processed;
	// to adjust the minimum level dynamically, use a `slog.LevelVar`.
	Level slog.Leveler
}

// DevModeHandler is a `slog.Handler` that writes Records to an io.Writer.
// The records are formatted to be easily readable by humans.
// This handler is meant for development use only as it doesn't provide optimal
// performance and its output is not machine-readable.
type DevModeHandler struct {
	opts   *DevModeHandlerOptions
	w      io.Writer
	attrs  []slog.Attr
	groups []string
	mu     sync.Mutex
}

// NewHandler creates a new `slog.Handler` with default options.
// If `devMode` is true, a `*DevModeHandler` is returned, else a `*slog.JSONHandler`.
func NewHandler(devMode bool, w io.Writer) slog.Handler {
	if devMode {
		return NewDevModeHandler(w, &DevModeHandlerOptions{Level: slog.LevelDebug})
	}
	return slog.NewJSONHandler(w, &slog.HandlerOptions{Level: slog.LevelInfo, AddSource: true})
}

// NewDevModeHandler creates a new `DevModeHandler` that writes to w, using the given options.
// If `opts` is `nil`, the default options are used.
func NewDevModeHandler(w io.Writer, opts *DevModeHandlerOptions) *DevModeHandler {
	if opts == nil {
		opts = &DevModeHandlerOptions{}
	}
	return &DevModeHandler{w: w, opts: opts}
}

// Handle formats its argument `Record` in an output easily readable by humans.
// The output contains multiple lines:
//   - The first one contains the log level, the time and the source
//   - The second one contains the message
//   - The next lines contain the attributes and groups, if any
//
// Each call to `Handle` results in a single serialized call to `io.Writer.Write()`.
func (h *DevModeHandler) Handle(_ context.Context, r slog.Record) error {

	buf := bytes.NewBuffer(make([]byte, 0, 1024))

	buf.WriteByte('\n')                  // TODO use line separator (lipgloss)
	buf.WriteByte('[')                   // TODO use lipgloss background and color instead of brackets
	buf.WriteString(levelColor(r.Level)) // Change color depending on level
	buf.WriteString(r.Level.String())
	buf.WriteString(Reset)
	buf.WriteString("] ")

	buf.WriteString(r.Time.Format("2006/01/02 15:04:05.999999"))
	fs := runtime.CallersFrames([]uintptr{r.PC})
	f, _ := fs.Next()
	buf.WriteString(Gray)
	buf.WriteString(" (")
	buf.WriteString(f.File)
	buf.WriteByte(':')
	buf.WriteString(strconv.Itoa(f.Line))
	buf.WriteString(")")
	buf.WriteString(Reset)
	buf.WriteByte('\n')
	buf.WriteString(messageColor(r.Level))
	buf.WriteString(r.Message)
	buf.WriteString(Reset)
	buf.WriteByte('\n')

	indent := 0
	for _, group := range h.groups {
		indentString := strings.Repeat(Indent, indent)
		buf.WriteString(indentString)
		buf.WriteString(WhiteBold)
		buf.WriteString(group)
		buf.WriteString(":\n")
		indent++
	}
	for _, attr := range h.attrs {
		printAttr(attr, buf, indent)
	}
	r.Attrs(func(a slog.Attr) bool {
		printAttr(a, buf, indent)
		return true
	})

	h.mu.Lock()
	defer h.mu.Unlock()
	_, err := h.w.Write(buf.Bytes())
	return err
}

func levelColor(level slog.Level) string {
	switch level {
	case slog.LevelDebug:
		return CyanBold
	case slog.LevelInfo:
		return WhiteBold
	case slog.LevelWarn:
		return YellowBold
	case slog.LevelError:
		return RedBold
	}
	return WhiteBold
}

func messageColor(level slog.Level) string {
	switch level {
	case slog.LevelDebug, slog.LevelInfo:
		return ""
	case slog.LevelWarn:
		return Yellow
	case slog.LevelError:
		return Red
	}
	return ""
}

// Enabled reports whether the handler handles records at the given level.
// The handler ignores records whose level is lower.
func (h *DevModeHandler) Enabled(_ context.Context, level slog.Level) bool {
	minLevel := slog.LevelInfo
	if h.opts.Level != nil {
		minLevel = h.opts.Level.Level()
	}
	return level >= minLevel
}

// WithAttrs returns a new `DevModeHandler` whose attributes consists
// of h's attributes followed by attrs.
func (h *DevModeHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &DevModeHandler{
		opts:  h.opts,
		w:     h.w,
		attrs: append(h.attrs, attrs...),
	}
}

// WithGroup returns a new `DevModeHandler` whose attributes are wrapped
// into a named group. All the handler's attributes will be printed indented
// into the added group.
func (h *DevModeHandler) WithGroup(name string) slog.Handler {
	return &DevModeHandler{
		opts:   h.opts,
		w:      h.w,
		attrs:  append(make([]slog.Attr, 0, len(h.attrs)), h.attrs...),
		groups: append(h.groups, name),
	}
}

func printAttr(attr slog.Attr, buf *bytes.Buffer, indent int) {

	indentString := strings.Repeat(Indent, indent)
	buf.WriteString(indentString)
	buf.WriteString(WhiteBold)
	buf.WriteString(attr.Key)
	buf.WriteString(": ")

	if attr.Value.Kind() == slog.KindAny {
		// This may be a struct or map, convert it if needed
		attr.Value = StructValue(attr.Value.Any())
	}

	if attr.Value.Kind() == slog.KindGroup {
		buf.WriteByte('\n')
		printGroup(attr.Value.Group(), buf, indent+1)
	} else {
		val := attr.Value.String()
		if strings.Contains(val, "\n") {
			// Break line if the message is multi-line (such as stacktrace)
			// Otherwise print it next to attr name so the log is more compact
			buf.WriteByte('\n')
			buf.WriteString(indentString)
		}
		buf.WriteString(Reset)
		buf.WriteString(val)
		buf.WriteByte('\n')
	}
}

func printGroup(group []slog.Attr, buf *bytes.Buffer, indent int) {
	for _, attr := range group {
		printAttr(attr, buf, indent)
	}
}
