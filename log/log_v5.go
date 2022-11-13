package log

import (
	"io"
	"time"

	"goyave.dev/goyave/v4"
)

// LogContext contains all information needed for a `Formatter`.
type LogContext struct {
	goyave.Component
	Now      time.Time
	Response *goyave.ResponseV5
	Request  *goyave.RequestV5
	Length   int
}

// Formatter is a function that builds a log entry.
// As logs are written at the end of the request's lifecycle, all the
// data is available to formatters at the time they are called, and all
// modifications will have no effect.
type FormatterV5 func(ctx *LogContext) string

// Writer chained writer keeping response body in memory.
// Used for loggin in common format.
type WriterV5 struct {
	goyave.Component
	formatter FormatterV5
	writer    io.Writer
	now       time.Time
	request   *goyave.RequestV5
	response  *goyave.ResponseV5
	length    int
}

var _ io.Closer = (*WriterV5)(nil)
var _ goyave.PreWriter = (*WriterV5)(nil)

// NewWriter create a new log writer.
// The given Request and Response will be used and passed to the given
// formatter.
func NewWriterV5(component goyave.Component, response *goyave.ResponseV5, request *goyave.RequestV5, formatter FormatterV5) *WriterV5 {
	return &WriterV5{
		Component: component,
		now:       time.Now(),
		request:   request,
		writer:    response.Writer(),
		response:  response,
		formatter: formatter,
	}
}

// PreWrite calls PreWrite on the
// child writer if it implements PreWriter.
func (w *WriterV5) PreWrite(b []byte) {
	if pr, ok := w.writer.(goyave.PreWriter); ok {
		pr.PreWrite(b)
	}
}

// Write writes the data as a response and keeps its length in memory
// for later logging.
func (w *WriterV5) Write(b []byte) (int, error) {
	w.length += len(b)
	return w.writer.Write(b)
}

// Close the writer and its child ResponseWriter, flushing response
// output to the logs.
func (w *WriterV5) Close() error {
	ctx := &LogContext{
		Component: w.Component,
		Now:       w.now,
		Response:  w.response,
		Request:   w.request,
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
type MiddlewareV5 struct {
	goyave.Component
	Formatter FormatterV5
}

func (m *MiddlewareV5) Handle(next goyave.HandlerV5) goyave.HandlerV5 {
	return func(response *goyave.ResponseV5, request *goyave.RequestV5) {
		logWriter := NewWriterV5(m.Component, response, request, m.Formatter)
		response.SetWriter(logWriter)

		next(response, request)
	}
}

// CommonLogMiddleware captures response data and outputs it to the default logger
// using the common log format.
func CommonLogMiddlewareV5() goyave.MiddlewareV5 {
	return &MiddlewareV5{Formatter: CommonLogFormatterV5}
}

// CombinedLogMiddleware captures response data and outputs it to the default logger
// using the combined log format.
func CombinedLogMiddlewareV5() goyave.MiddlewareV5 {
	return &MiddlewareV5{Formatter: CombinedLogFormatterV5}
}
