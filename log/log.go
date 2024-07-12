package log

import (
	"io"
	"log/slog"

	"github.com/samber/lo"
	"goyave.dev/goyave/v5"
)

// Context contains all information needed for a `Formatter`.
type Context struct {
	goyave.Component
	Request *goyave.Request
	Status  int
	Length  int
}

// Formatter is a function that builds a log entry.
// As logs are written at the end of the request's lifecycle, all the
// data is available to formatters at the time they are called, and all
// modifications will have no effect.
//
// The first returned value is the message, usually formatted using a standard
// like Common Log Format or Combined Log Format.
// The second returned value is a slice of structured logging attributes.
// These attributes are ignored in dev mode (`app.debug = true`) to avoid clutter.
type Formatter func(ctx *Context) (message string, attributes []slog.Attr)

// Writer chained writer keeping response body in memory.
// Used for loggin in common format.
type Writer struct {
	goyave.CommonWriter
	formatter Formatter
	request   *goyave.Request
	response  *goyave.Response
	length    int
}

var _ io.Closer = (*Writer)(nil)
var _ goyave.PreWriter = (*Writer)(nil)

// NewWriter create a new log writer.
// The given Request and Response will be used and passed to the given
// formatter.
func NewWriter(server *goyave.Server, response *goyave.Response, request *goyave.Request, formatter Formatter) *Writer {
	writer := &Writer{
		CommonWriter: goyave.NewCommonWriter(response.Writer()),
		request:      request,
		response:     response,
		formatter:    formatter,
	}
	writer.Init(server)
	return writer
}

// Write writes the data as a response and keeps its length in memory
// for later logging.
func (w *Writer) Write(b []byte) (int, error) {
	w.length += len(b)
	return w.CommonWriter.Write(b)
}

// Close the writer and its child ResponseWriter, flushing response
// output to the logs.
func (w *Writer) Close() error {
	ctx := &Context{
		Component: w.Component,
		Request:   w.request,
		Status:    w.response.GetStatus(),
		Length:    w.length,
	}
	message, attrs := w.formatter(ctx)

	if w.Config().GetBool("app.debug") {
		// In dev mode, we omit the details to avoid clutter. The message itself is enough.
		w.Logger().Info(message)
	} else {
		w.Logger().Info(message, lo.Map(attrs, func(a slog.Attr, _ int) any { return a })...)
	}

	return w.CommonWriter.Close()
}

// AccessMiddleware captures response data and outputs it to the logger at the
// INFO level. The message and attributes logged are defined by the `Formatter`.
type AccessMiddleware struct {
	goyave.Component
	Formatter Formatter
}

// Handle adds the access logging chained writer to the response.
func (m *AccessMiddleware) Handle(next goyave.Handler) goyave.Handler {
	return func(response *goyave.Response, request *goyave.Request) {
		logWriter := NewWriter(m.Server(), response, request, m.Formatter)
		response.SetWriter(logWriter)

		next(response, request)
	}
}

// CommonLogMiddleware captures response data and outputs it to the default logger
// using the common log format.
func CommonLogMiddleware() goyave.Middleware {
	return &AccessMiddleware{Formatter: CommonLogFormatter}
}

// CombinedLogMiddleware captures response data and outputs it to the default logger
// using the combined log format.
func CombinedLogMiddleware() goyave.Middleware {
	return &AccessMiddleware{Formatter: CombinedLogFormatter}
}
