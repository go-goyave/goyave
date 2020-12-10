package ratelimiter

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/System-Glitch/goyave/v3"
)

type RateLimiterMiddlewareTestSuite struct {
	goyave.TestSuite
}

func TestRateLimiterMiddlewareTestSuite(t *testing.T) {
	goyave.RunTest(t, new(RateLimiterMiddlewareTestSuite))
}

func (suite *RateLimiterMiddlewareTestSuite) TestLimiterResponseHeaders() {

	requestQuota := 10
	quotaDuration := 5 * time.Second
	clientID := "client_1234"

	// To be used for testing RateLimit-Reset
	var secondsToQuotaReset time.Duration = 5555

	limiterConfig := Config{
		RequestQuota:  requestQuota,
		QuotaDuration: quotaDuration,
		ClientID:      clientID,
	}

	// stub a limiter with pre-defined resetAt.
	limiter := &limiter{
		config:   limiterConfig,
		counter:  0,
		resetsAt: time.Now().Add(secondsToQuotaReset * time.Second),
	}

	// stub a new store and set the pre-defined limiter
	store := newLimiterStore()
	store.set(clientID, limiter)

	ratelimiterMiddleware := newWithStore(func(request *goyave.Request) Config {
		return limiterConfig
	}, &store)

	request := suite.CreateTestRequest(nil)

	result := suite.Middleware(
		ratelimiterMiddleware,
		request,
		func(response *goyave.Response, request *goyave.Request) {},
	)

	suite.Equal(
		fmt.Sprintf("%v, %v;w=%v", requestQuota, requestQuota, quotaDuration.Seconds()),
		result.Header.Get("RateLimit-Limit"),
	)

	suite.Equal(
		fmt.Sprintf("%v", requestQuota-1),
		result.Header.Get("RateLimit-Remaining"),
	)

	suite.Equal(
		fmt.Sprintf("%v", int64(secondsToQuotaReset)),
		result.Header.Get("RateLimit-Reset"),
	)

	limiter.counter = 10
	request = suite.CreateTestRequest(nil)
	result = suite.Middleware(
		ratelimiterMiddleware,
		request,
		func(response *goyave.Response, request *goyave.Request) {},
	)
	suite.Equal(
		"0",
		result.Header.Get("RateLimit-Remaining"),
	)
}

func (suite *RateLimiterMiddlewareTestSuite) TestClientExceedsTheAllowedQuota() {
	const quota = 2
	ratelimiterMiddleware := New(func(request *goyave.Request) Config {
		return Config{
			RequestQuota:  quota,
			QuotaDuration: 10 * time.Minute,
		}
	})

	var result *http.Response

	for i := 0; i < quota+1; i++ {
		request := suite.CreateTestRequest(nil)
		result = suite.Middleware(
			ratelimiterMiddleware,
			request,
			func(response *goyave.Response, request *goyave.Request) {
				if i+1 > quota {
					suite.Fail("Handler executed, should be blocking when rate limit exceeded")
				}
			},
		)
	}

	suite.Equal(http.StatusTooManyRequests, result.StatusCode)
}

func (suite *RateLimiterMiddlewareTestSuite) TestRequestQuotaResetsAfterQuotaDurationExpires() {
	const quota = 5
	ratelimiterMiddleware := New(func(request *goyave.Request) Config {
		return Config{
			RequestQuota:  quota,
			QuotaDuration: time.Second,
		}
	})

	request := suite.CreateTestRequest(nil)

	var result *http.Response

	for i := 0; i < quota+1; i++ {
		result = suite.Middleware(
			ratelimiterMiddleware,
			request,
			func(response *goyave.Response, request *goyave.Request) {
				if i+1 > quota {
					suite.Fail("Handler executed, should be blocking when rate limit exceeded")
				}
			},
		)
	}

	suite.Equal(http.StatusTooManyRequests, result.StatusCode)

	time.Sleep(time.Second)

	result = suite.Middleware(
		ratelimiterMiddleware,
		request,
		func(response *goyave.Response, request *goyave.Request) {},
	)

	suite.Equal(http.StatusNoContent, result.StatusCode)
}

func (suite *RateLimiterMiddlewareTestSuite) TestLimiterQuotaIsZero() {
	// This middleware should be skipped if the quota or the duration is equal to zero
	ratelimiterMiddleware := New(func(request *goyave.Request) Config {
		return Config{
			RequestQuota:  0,
			QuotaDuration: 2 * time.Second,
		}
	})

	request := suite.CreateTestRequest(nil)
	for i := 0; i < 5; i++ {
		result := suite.Middleware(
			ratelimiterMiddleware,
			request,
			func(response *goyave.Response, request *goyave.Request) {},
		)
		suite.Equal(http.StatusNoContent, result.StatusCode)
	}

	const quota = 5
	ratelimiterMiddleware = New(func(request *goyave.Request) Config {
		return Config{
			RequestQuota:  quota,
			QuotaDuration: 0,
		}
	})
	for i := 0; i < quota+1; i++ {
		result := suite.Middleware(
			ratelimiterMiddleware,
			request,
			func(response *goyave.Response, request *goyave.Request) {},
		)
		suite.Equal(http.StatusNoContent, result.StatusCode)
	}
}

func (suite *RateLimiterMiddlewareTestSuite) TestDefaultClientID() {
	request := suite.CreateTestRequest(nil)
	request.Request().RemoteAddr = "127.0.0.1"

	suite.Equal("127.0.0.1", defaultClientID(request))

	request.Request().RemoteAddr = "127.0.0.1:1111"
	suite.Equal("127.0.0.1", defaultClientID(request))
}
