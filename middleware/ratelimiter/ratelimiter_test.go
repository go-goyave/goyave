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
	numberOfRequests := 7
	quotaDuration := 5 * time.Second

	ratelimiterMiddleware := New(func(request *goyave.Request) LimiterConfig {
		return LimiterConfig{
			RequestQuota:  requestQuota,
			QuotaDuration: quotaDuration,
		}
	})

	request := suite.CreateTestRequest(nil)

	var result *http.Response

	for i := 0; i < numberOfRequests; i++ {
		result = suite.Middleware(
			ratelimiterMiddleware,
			request,
			func(response *goyave.Response, request *goyave.Request) {},
		)
	}

	suite.Equal(
		fmt.Sprintf("%v, %v;w=%v", requestQuota, requestQuota, quotaDuration.Seconds()),
		result.Header.Get("RateLimit-Limit"),
	)

	suite.Equal(
		fmt.Sprintf("%v", requestQuota-numberOfRequests),
		result.Header.Get("RateLimit-Remaining"),
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

func (suite *RateLimiterMiddlewareTestSuite) TestLimiterConfigDefaults() {
	request := suite.CreateTestRequest(nil)

	rateLimiterMiddleware := New(func(request *goyave.Request) LimiterConfig {
		return LimiterConfig{
			RequestQuota:  1,
			QuotaDuration: 2 * time.Second,
		}
	})

	// Check if responseHandler works

	var result *http.Response

	for i := 0; i < 2; i++ {
		result = suite.Middleware(
			rateLimiterMiddleware,
			request,
			func(response *goyave.Response, request *goyave.Request) {},
		)
	}

	defer result.Body.Close()

	bj := map[string]map[string]interface{}{}

	err := suite.GetJSONBody(result, &bj)

	suite.Nil(err)

	suite.Equal(float64(http.StatusTooManyRequests), bj["errors"]["status"])
	suite.Equal("Too many requests", bj["errors"]["title"])
}
