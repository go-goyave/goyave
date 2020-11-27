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

	limiterConfig := LimiterConfig{
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

	ratelimiterMiddleware := newWithStore(func(request *goyave.Request) LimiterConfig {
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

}

func (suite *RateLimiterMiddlewareTestSuite) TestWhenClientExceedsTheAllowedQuota() {

	ratelimiterMiddleware := New(func(request *goyave.Request) LimiterConfig {
		return LimiterConfig{
			RequestQuota:  2,
			QuotaDuration: 10 * time.Minute,
		}
	})

	var result *http.Response

	for i := 0; i < 3; i++ {
		request := suite.CreateTestRequest(nil)
		result = suite.Middleware(
			ratelimiterMiddleware,
			request,
			func(response *goyave.Response, request *goyave.Request) {},
		)
	}

	suite.Equal(http.StatusTooManyRequests, result.StatusCode)
}

func (suite *RateLimiterMiddlewareTestSuite) TestRequestQuotaResetsAfterQuotaDurationExpires() {

	ratelimiterMiddleware := New(func(request *goyave.Request) LimiterConfig {
		return LimiterConfig{
			RequestQuota:  5,
			QuotaDuration: 2 * time.Second,
		}
	})

	request := suite.CreateTestRequest(nil)

	var result *http.Response

	for i := 0; i < 6; i++ {
		result = suite.Middleware(
			ratelimiterMiddleware,
			request,
			func(response *goyave.Response, request *goyave.Request) {},
		)
	}

	suite.Equal(http.StatusTooManyRequests, result.StatusCode)

	time.Sleep(2 * time.Second)

	result = suite.Middleware(
		ratelimiterMiddleware,
		request,
		func(response *goyave.Response, request *goyave.Request) {},
	)

	suite.Equal(http.StatusNoContent, result.StatusCode)
}

func (suite *RateLimiterMiddlewareTestSuite) TestDefaultClientID() {
	request := suite.CreateTestRequest(nil)
	request.Request().RemoteAddr = "127.0.0.1"

	suite.Equal("127.0.0.1", defaultClientID(request))

	request.Request().RemoteAddr = "127.0.0.1:1111"
	suite.Equal("127.0.0.1", defaultClientID(request))
}
