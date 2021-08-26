package log

import (
	"fmt"
	"net"
	"strconv"
	"time"

	"goyave.dev/goyave/v4"
)

const (
	// Format is the fmt format code for common logs
	Format string = "%s %s %s [%s] \"%s %s %s\" %d %d"

	// TimestampFormat is the time format used in common logs
	TimestampFormat string = "02/Jan/2006:15:04:05 -0700"
)

// CommonLogFormatter build a log entry using the Common Log Format.
func CommonLogFormatter(now time.Time, response *goyave.Response, request *goyave.Request, length int) string {
	req := request.Request()
	url := request.URI()

	username := "-"
	if url.User != nil {
		if name := url.User.Username(); name != "" {
			username = name
		}
	}

	host, _, err := net.SplitHostPort(req.RemoteAddr)

	if err != nil {
		host = req.RemoteAddr
	}

	uri := req.RequestURI

	// Requests using the CONNECT method over HTTP/2.0 must use
	// the authority field (aka r.Host) to identify the target.
	// Refer: https://httpwg.github.io/specs/rfc7540.html#CONNECT
	if req.ProtoMajor == 2 && req.Method == "CONNECT" {
		uri = req.Host
	}
	if uri == "" {
		uri = url.RequestURI()
	}

	return fmt.Sprintf(Format,
		host,
		"-",
		username,
		now.Format(TimestampFormat),
		req.Method,
		strconv.QuoteToASCII(uri),
		req.Proto,
		response.GetStatus(),
		length,
	)
}

// CombinedLogFormatter build a log entry using the Combined Log Format.
func CombinedLogFormatter(now time.Time, response *goyave.Response, request *goyave.Request, length int) string {
	return fmt.Sprintf("%s \"%s\" \"%s\"", CommonLogFormatter(now, response, request, length), request.Referrer(), request.UserAgent())
}
