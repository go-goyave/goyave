---
meta:
  - name: "og:title"
    content: "Rate Limiting - Goyave"
  - name: "twitter:title"
    content: "Rate Limiting - Goyave"
  - name: "title"
    content: "Rate Limiting - Goyave"
---

# Rate Limiting <Badge text="Since v3.5.0"/>

[[toc]]

## Introduction

Rate limiting is a crucial part of public API development. If you want to protect your data from being crawled, protect yourself from DDOS attacks, or provide different tiers of access to your API, you can do it using Goyave's built-in rate limiting middleware.

This middleware uses either a client's IP or an authenticated client's ID (or any other way of identifying a client you may need) and maps a quota, a quota duration and a request count to it. If a client exceeds the request quota in the given quota duration, this middleware will block and return `HTTP 429 Too Many Requests`.

The middleware will always add the following headers to the response:
- `RateLimit-Limit`: containing the requests quota in the time window
- `RateLimit-Remaining`: containing the remaining requests quota in the current window
- `RateLimit-Reset`: containing the time remaining in the current window, specified in seconds

::: warning 
This implementation is based on this [IETF **draft**](https://tools.ietf.org/id/draft-polli-ratelimit-headers-04.html). Being a **draft**, this is **not** yet a standard.
:::

## Usage

The rate middleware initializer takes a function as parameter. This function will be executed **for each** request and returns a limiter configuration.

```go
import "github.com/System-Glitch/goyave/v3/middleware/ratelimiter"

ratelimiterMiddleware := ratelimiter.New(func(request *goyave.Request) ratelimiter.Config {
    return ratelimiter.Config {
        RequestQuota:  100,
        QuotaDuration: time.Minute,
        // 100 requests per minute allowed
        // Client IP will be used as identifier
    }
})

router.Middleware(ratelimiterMiddleware)
```

If you want to identify authenticated users, set the `ClientID` field in the `ratelimiter.Config`. Additionally, you can set a different quota and quota duration based on the user.
```go
ratelimiterMiddleware := ratelimiter.New(func(request *goyave.Request) ratelimiter.Config {
    user := request.User.(*model.User)
    return ratelimiter.Config {
        ClientID: user.ID,
        RequestQuota:  user.RequestQuota,
        QuotaDuration: time.Minute,
    }
})
```

::: tip Note
If either `RequestQuota` or `QuotaDuration` is equal to 0 in the returned `ratelimiter.Config`, the middleware is skipped and immediately passes to the next handler.
:::

## Reference

#### ratelimiter.New

Create a new rate limiter middleware.

| Parameters                        | Return              |
|-----------------------------------|---------------------|
| `configFn ratelimiter.ConfigFunc` | `goyave.Middleware` |

:::tip Note
`ratelimiter.ConfigFunc` is an alias for `func(request *goyave.Request) ratelimiter.Config`
:::

#### ratelimiter.Config

```go
type Config struct {
    // Maximum number of requests in a client can send
    RequestQuota int

	// Duration or time taken until the quota expires and renews
    QuotaDuration time.Duration

    // Unique identifier for requestors. Can be userID or IP
    // Defaults to Remote Address if it is empty
    ClientID interface{}
}
```