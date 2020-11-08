package ratelimiter

import (
	"net/http"
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
	ClientID string

	// Handles response when rate limit exceeds
	ResponseHandler goyave.Handler
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

			if config.ResponseHandler == nil {
				config.ResponseHandler = defaultResponseHandler
			}

			if config.ClientID == "" {
				config.ClientID = defaultClientID(request)
			}

			key := config.ClientID

			l := lstore.get(key)

			if l == nil {
				l = newLimiter(config.RequestQuota, config.QuotaDuration)
				lstore.set(key, l)
			}

			if time.Since(l.updatedAt) < l.quotaDuration {
				if l.counter >= l.requestQuota {
					response.Status(http.StatusTooManyRequests)
					config.ResponseHandler(response, request)
					return
				}

			} else {
				lstore.reset(key)
			}

			lstore.increment(key)

			next(response, request)
		}
	}
}

func defaultResponseHandler(response *goyave.Response, request *goyave.Request) {
	response.JSON(http.StatusTooManyRequests, map[string]interface{}{
		"errors": map[string]interface{}{
			"status": 429,
			"title":  "Too many requests",
		},
	})
}

func defaultClientID(request *goyave.Request) string {
	return request.RemoteAddress()
}
