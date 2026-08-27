package log

import (
	"io"

	stdslog "log/slog"

	"github.com/samber/lo"
	"goyave.dev/goyave/v5"
	"goyave.dev/goyave/v5/slog"
	"goyave.dev/goyave/v5/util/errors"
)

// Context contains all information needed for a `Formatter`.
type Context struct {
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
type Formatter func(ctx *Context) (message string, attributes []stdslog.Attr)

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
func NewWriter(response *goyave.Response, request *goyave.Request, formatter Formatter) *Writer {
	writer := &Writer{
		CommonWriter: goyave.NewCommonWriter(response.Writer()),
		request:      request,
		response:     response,
		formatter:    formatter,
	}
	return writer
}

// Write writes the data as a response and keeps its length in memory
// for later logging.
func (w *Writer) Write(b []byte) (int, error) {
	w.length += len(b)
	n, err := w.CommonWriter.Write(b)
	return n, errors.New(err)
}

// Close the writer and its child ResponseWriter, flushing response
// output to the logs.
func (w *Writer) Close() error {
	ctx := &Context{
		Request: w.request,
		Status:  w.response.GetStatus(),
		Length:  w.length,
	}
	message, attrs := w.formatter(ctx)

	// TODO Previously we only printed the message in dev mode to avoid clutter.
	// Passing around the debug config entry isn't as easy now, so let's print them anyway.
	slog.FromContext(w.request.Context()).Info(message, lo.Map(attrs, func(a stdslog.Attr, _ int) any { return a })...)

	return errors.New(w.CommonWriter.Close())
}

// AccessMiddleware captures response data and outputs it to the logger at the
// INFO level. The message and attributes logged are defined by the `Formatter`.
type AccessMiddleware struct {
	Formatter Formatter
}

// Handle adds the access logging chained writer to the response.
func (m *AccessMiddleware) Handle(next goyave.Handler) goyave.Handler {
	return func(response *goyave.Response, request *goyave.Request) {
		logWriter := NewWriter(response, request, m.Formatter)
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
