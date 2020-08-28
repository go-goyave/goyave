---
meta:
  - name: "og:title"
    content: "Middleware - Goyave"
  - name: "twitter:title"
    content: "Middleware - Goyave"
  - name: "title"
    content: "Middleware - Goyave"
---

# Middleware

[[toc]]

## Introduction

Middleware are handlers executed before the controller handler. They are a convenient way to filter, intercept or alter HTTP requests entering your application. For example, middleware can be used to authenticate users. If the user is not authenticated, a message is sent to the user even before the controller handler is reached. However, if the user is authenticated, the middleware will pass to the next handler. Middleware can also be used to sanitize user inputs, by trimming strings for example, to log all requests into a log file, to automatically add headers to all your responses, etc.

Writing middleware is as easy as writing standard handlers. In fact, middleware are handlers, but they have an additional responsibility: when they are done, the may or may not pass to the next handler, which is either another middleware or a controller handler.

## Writing middleware

Each middleware is written in its own file inside the `http/middleware` directory. A `Middleware` is a function returning a `Handler`.

::: tip
`goyave.Middleware` is an alias for `func(goyave.Handler) goyave.Handler`
:::

**Example:**
``` go
func MyCustomMiddleware(next goyave.Handler) goyave.Handler {
	return func(response *goyave.Response, request *goyave.Request) {
        // Do something
        next(response, request) // Pass to the next handler
    }
}
```

When implementing middleware, keep in mind that the request **has not been validated yet**! Manipulating unvalidated data can be dangerous, especially in form-data where the data types are not converted by the validator yet. In middleware, you should always check if the request has been parsed correctly before trying to access its data:
``` go
if request.Data != nil {
    // Parsing OK
}
```

If you want your middleware to stop the request and respond immediately before reaching the controller handler, simply don't call the `next` handler. In the following example, consider that you developed a custom authentication system:
``` go
func CustomAuthentication(next goyave.Handler) goyave.Handler {
	return func(response *goyave.Response, request *goyave.Request) {
        if !auth.Check(request) {
            response.Status(http.StatusUnauthorized)
            return
        }

        next(response, request)
    }
}
```

::: tip
When a middleware stops a request, following middleware are **not** executed neither.
:::

## Applying Middleware

When your middleware is ready, you will need to apply it to a router or a specific route. When applying a middleware to a router, all routes and subrouters will execute this middleware when matched.

```go
router.Middleware(middleware.MyCustomMiddleware)

router.Get("/products", product.Index).Middleware(middleware.MyCustomMiddleware)
```

## Built-in middleware

Built-in middleware is located in the `middleware` package.
``` go
import "github.com/System-Glitch/goyave/v3/middleware"
```

### DisallowNonValidatedFields

DisallowNonValidatedFields validates that all fields in the request are validated by the rule set. The middleware stops the request and sends "422 Unprocessable Entity" and an error message if the user has sent non-validated field(s). Fields ending with `_confirmation` are ignored.

If the body parsing failed, this middleware immediately passes to the next handler. **This middleware shall only be used with requests having a rule set defined.**

The returned error message can be customized using the entry `disallow-non-validated-fields` in the `locale.json` language file.

```go
router.Middleware(middleware.DisallowNonValidatedFields)
```

### Trim

<p><Badge text="Since v2.0.0"/></p>

Trim removes all leading and trailing white space from string fields.

For example, `" \t  trimmed\n  \t"` will be transformed to `"trimmed"`.

```go
router.Middleware(middleware.Trim)
```

### Gzip

<p><Badge text="Since v2.7.0"/></p>

Gzip compresses HTTP responses with default compression level for clients that support it via the `Accept-Encoding` header.

```go
router.Middleware(middleware.Gzip())
```

The compression level can be specified using `GzipLevel(level)`. The compression level should be `gzip.DefaultCompression`, `gzip.NoCompression`, or any integer value between `gzip.BestSpeed` and `gzip.BestCompression` inclusive. `gzip.HuffmanOnly` is also available.

``` go
router.Middleware(middleware.GzipLevel(gzip.BestCompression))
```
