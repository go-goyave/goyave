package ratelimiter

import (
	"net/http"
	"strings"
	"time"

	"goyave.dev/goyave/v3"
)

// Config for setting configuration for the limiter middleware
type Config struct {
	// Unique identifier for requestors. Can be userID or IP
	// Defaults to Remote Address if it is empty
	ClientID interface{}

	// Duration or time taken until the quota expires and renews
	QuotaDuration time.Duration

	// Maximum number of requests in a client can send
	RequestQuota int
}

// ConfigFunc acts as a factory for Config structs
type ConfigFunc func(request *goyave.Request) Config

// New initializes new a rate limiter middleware
func New(configFn ConfigFunc) goyave.Middleware {
	lstore := newLimiterStore()
	return newWithStore(configFn, &lstore)
}

func newWithStore(configFn ConfigFunc, lstore *limiterStore) goyave.Middleware {

	return func(next goyave.Handler) goyave.Handler {

		return func(response *goyave.Response, request *goyave.Request) {

			config := configFn(request)

			if config.RequestQuota == 0 || config.QuotaDuration == 0 {
				next(response, request)
				return
			}

			if config.ClientID == nil {
				config.ClientID = defaultClientID(request)
			}

			key := config.ClientID

			l := lstore.get(key, config)

			if !l.validateAndUpdate(response) {
				response.Status(http.StatusTooManyRequests)
				return
			}

			next(response, request)
		}
	}
}

func defaultClientID(request *goyave.Request) string {
	remoteAddr := request.RemoteAddress()

	// strip off port number
	e := strings.Index(remoteAddr, ":")

	if e == -1 {
		return remoteAddr
	}

	return remoteAddr[0:e]
}
