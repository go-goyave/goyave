package log

import (
	"io"

	"goyave.dev/goyave/v4"
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
type Formatter func(ctx *Context) string

// Writer chained writer keeping response body in memory.
// Used for loggin in common format.
type Writer struct {
	goyave.Component
	formatter Formatter
	writer    io.Writer
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
		request:   request,
		writer:    response.Writer(),
		response:  response,
		formatter: formatter,
	}
	writer.Init(server)
	return writer
}

// PreWrite calls PreWrite on the
// child writer if it implements PreWriter.
func (w *Writer) PreWrite(b []byte) {
	if pr, ok := w.writer.(goyave.PreWriter); ok {
		pr.PreWrite(b)
	}
}

// Write writes the data as a response and keeps its length in memory
// for later logging.
func (w *Writer) Write(b []byte) (int, error) {
	w.length += len(b)
	return w.writer.Write(b)
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
	w.AccessLogger().Println(w.formatter(ctx))

	if wr, ok := w.writer.(io.Closer); ok {
		return wr.Close()
	}
	return nil
}

// Middleware captures response data and outputs it to the default logger
// using the given formatter.
type Middleware struct {
	goyave.Component
	Formatter Formatter
}

// Handle adds the log chained writer to the response.
func (m *Middleware) Handle(next goyave.Handler) goyave.Handler {
	return func(response *goyave.Response, request *goyave.Request) {
		logWriter := NewWriter(m.Server(), response, request, m.Formatter)
		response.SetWriter(logWriter)

		next(response, request)
	}
}

// CommonLogMiddleware captures response data and outputs it to the default logger
// using the common log format.
func CommonLogMiddleware() goyave.Middleware {
	return &Middleware{Formatter: CommonLogFormatter}
}

// CombinedLogMiddleware captures response data and outputs it to the default logger
// using the combined log format.
func CombinedLogMiddleware() goyave.Middleware {
	return &Middleware{Formatter: CombinedLogFormatter}
}
