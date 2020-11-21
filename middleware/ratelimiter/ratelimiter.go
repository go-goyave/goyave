package ratelimiter

import (
	"net/http"
	"strings"
	"time"

	"github.com/System-Glitch/goyave/v3"
)

// LimiterConfig for setting configuration for the limiter middleware
type LimiterConfig struct {
	// Maximum number of requests in a client can send
	RequestQuota int

	// Duration or time taken until the quota expires and renews
	QuotaDuration time.Duration

	// Unique identifier for requestors. Can be userID or IP
	// Defaults to Remote Address if it is empty
	ClientID interface{}
}

// LimiterConfigFunc acts as a factory for LimiterConfig structs
type LimiterConfigFunc func(request *goyave.Request) LimiterConfig

// New initializes new a rate limiter middleware
func New(configFn LimiterConfigFunc) goyave.Middleware {

	lstore := newLimiterStore()

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

			l := lstore.get(key)

			if l == nil {
				l = newLimiter(config.RequestQuota, config.QuotaDuration)
				lstore.set(key, l)
			}

			if ok := l.validateAndUpdate(response); !ok {
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
	addrSlice := strings.Split(remoteAddr, ":")

	if len(addrSlice) == 0 {
		return remoteAddr
	}

	return addrSlice[0]
}
