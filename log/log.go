package log

import (
	"io"
	"log"
	"time"

	"github.com/System-Glitch/goyave/v2"
)

// Formatter is a function that builds a log entry.
// As logs are written at the end of the request's lifecycle, all the
// data is available to formatters at the time they are called, and all
// modifications will have no effect.
type Formatter func(now time.Time, response *goyave.Response, request *goyave.Request, body []byte) string

// CommonLogWriter chained writer keeping response body in memory.
// Used for loggin in common format.
type CommonLogWriter struct {
	now       time.Time
	request   *goyave.Request
	writer    io.Writer
	response  *goyave.Response
	body      []byte
	formatter Formatter
}

var _ io.Closer = (*CommonLogWriter)(nil)

// NewCommonLogWriter create a new CommonLogWriter.
// The given Request and Response will be used and passed to the given
// formatter.
func NewCommonLogWriter(response *goyave.Response, request *goyave.Request, formatter Formatter) *CommonLogWriter {
	return &CommonLogWriter{
		now:       time.Now(),
		request:   request,
		writer:    response.Writer(),
		response:  response,
		formatter: formatter,
	}
}

// Write writes the data as a response and keeps it in memory
// for later logging.
func (w *CommonLogWriter) Write(b []byte) (int, error) {
	w.body = append(w.body, b...)
	return w.writer.Write(b)
}

// Close the writer and its child ResponseWriter, flushing response
// output to the logs.
func (w *CommonLogWriter) Close() error {
	// TODO use default logger
	log.Println(w.formatter(w.now, w.response, w.request, w.body))

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
			logWriter := NewCommonLogWriter(response, request, formatter)
			response.SetWriter(logWriter)

			next(response, request)
		}
	}
}

// CommonMiddleware captures response data and outputs it to the default logger
// using the common log format.
func CommonMiddleware() goyave.Middleware {
	return Middleware(CommonLogFormatter)
}

// CombinedMiddleware captures response data and outputs it to the default logger
// using the combined log format.
func CombinedMiddleware() goyave.Middleware {
	return Middleware(CombinedLogFormatter)
}
