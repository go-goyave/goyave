package log

import (
	"fmt"
	"log/slog"
	"net"
	"strconv"

	"github.com/samber/lo"
)

const (
	// Format is the fmt format code for common logs
	Format string = "%s %s %s [%s] \"%s %s %s\" %d %d"

	// TimestampFormat is the time format used in common logs
	TimestampFormat string = "02/Jan/2006:15:04:05 -0700"
)

// CommonLogFormatter build a log entry using the Common Log Format.
func CommonLogFormatter(ctx *Context) (string, []slog.Attr) {
	req := ctx.Request.Request()
	url := ctx.Request.URL()

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

	message := fmt.Sprintf(Format,
		host,
		"-",
		username,
		ctx.Request.Now.Format(TimestampFormat),
		req.Method,
		escapeToASCII(uri),
		req.Proto,
		ctx.Status,
		ctx.Length,
	)

	details := slog.Group("details",
		slog.String("host", host),
		slog.String("username", username),
		slog.Time("time", ctx.Request.Now),
		slog.String("method", req.Method),
		slog.String("uri", uri),
		slog.String("proto", req.Proto),
		slog.Int("status", ctx.Status),
		slog.Int("length", ctx.Length),
	)

	return message, []slog.Attr{details}
}

func escapeToASCII(s string) string {
	quoted := strconv.QuoteToASCII(s)
	return quoted[1 : len(quoted)-1]
}

// CombinedLogFormatter build a log entry using the Combined Log Format.
func CombinedLogFormatter(ctx *Context) (string, []slog.Attr) {
	message, attrs := CommonLogFormatter(ctx)

	message = fmt.Sprintf("%s \"%s\" \"%s\"", message, ctx.Request.Referrer(), ctx.Request.UserAgent())

	details := lo.Map(attrs[0].Value.Group(), func(a slog.Attr, _ int) any { return a })
	details = append(details, slog.String("referrer", ctx.Request.Referrer()), slog.String("userAgent", ctx.Request.UserAgent()))
	attrs = []slog.Attr{slog.Group("details", details...)}

	return message, attrs
}
