package log

import (
	"io"
	"time"

	"goyave.dev/goyave/v3"
)

// Formatter is a function that builds a log entry.
// As logs are written at the end of the request's lifecycle, all the
// data is available to formatters at the time they are called, and all
// modifications will have no effect.
type Formatter func(now time.Time, response *goyave.Response, request *goyave.Request, length int) string

// Writer chained writer keeping response body in memory.
// Used for loggin in common format.
type Writer struct {
	formatter Formatter
	writer    io.Writer
	now       time.Time
	request   *goyave.Request
	response  *goyave.Response
	length    int
}

var _ io.Closer = (*Writer)(nil)
var _ goyave.PreWriter = (*Writer)(nil)

// NewWriter create a new LogWriter.
// The given Request and Response will be used and passed to the given
// formatter.
func NewWriter(response *goyave.Response, request *goyave.Request, formatter Formatter) *Writer {
	return &Writer{
		now:       time.Now(),
		request:   request,
		writer:    response.Writer(),
		response:  response,
		formatter: formatter,
	}
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
	goyave.AccessLogger.Println(w.formatter(w.now, w.response, w.request, w.length))

	if wr, ok := w.writer.(io.Closer); ok {
		return wr.Close()
	}
	return nil
}

// Middleware captures response data and outputs it to the default logger
// using the given formatter.
func Middleware(formatter Formatter) goyave.Middleware {
	return func(next goyave.Handler) goyave.Handler {
		return func(response *goyave.Response, request *goyave.Request) {
			logWriter := NewWriter(response, request, formatter)
			response.SetWriter(logWriter)

			next(response, request)
		}
	}
}

// CommonLogMiddleware captures response data and outputs it to the default logger
// using the common log format.
func CommonLogMiddleware() goyave.Middleware {
	return Middleware(CommonLogFormatter)
}

// CombinedLogMiddleware captures response data and outputs it to the default logger
// using the combined log format.
func CombinedLogMiddleware() goyave.Middleware {
	return Middleware(CombinedLogFormatter)
}
