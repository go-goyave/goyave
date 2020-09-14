---
meta:
  - name: "og:title"
    content: "Logging - Goyave"
  - name: "twitter:title"
    content: "Logging - Goyave"
  - name: "title"
    content: "Logging - Goyave"
---

# Logging <Badge text="Since v2.8.0"/>

[[toc]]

## Introduction

Logging is an important part of all applications. The framework provides a standard and flexible way to log accesses, errors and regular information. This standard logging feature should be preferred in all applications and modules developed using Goyave for consistency across the framework's environment.

## Custom loggers

The framework provides three [Go standard loggers](https://golang.org/pkg/log/):
- `goyave.Logger`: the logger for regular and miscellaneous information. Outputs to `os.Stdout` with `log.LstdFlags` by default.
- `goyave.AccessLogger`: the logger used by the logging middleware (see below). Outputs to `os.Stdout` with no flags by default.
- `goyave.ErrLogger`: the logger for errors and stacktraces. Outputs to `os.Stderr` with `log.LstdFlags` by default.

All these loggers can be modified or replaced entirely. Modifications to standard loggers should be done **before** calling `goyave.Start()`.

**Example:**
```go
func main() {
    // Replace de default logger
    goyave.Logger = log.New(os.Stdout, "myapp", log.Ldate | log.Ltime | log.Lshortfile)

    goyave.Logger.Println("Starting...")
    goyave.RegisterStartupHook(func() {
        goyave.Logger.Println("Started.")
    })
    if err := goyave.Start(registerRoutes); err != nil {
        os.Exit(err.(*goyave.Error).ExitCode)
    }
}
```

## Common and Combined access logs

To enable logging of accesses using the [Common Log Format](https://en.wikipedia.org/wiki/Common_Log_Format), simply register the `CommonLogMiddleware`. Alternatively, you can use `CombinedLogMiddleware` for Combined Log Format.

``` go
import (
    "github.com/System-Glitch/goyave/v3/log"
    "github.com/System-Glitch/goyave/v3"
)

func registerRoutes(router *goyave.Router) {
    // Common log format
    router.Middleware(log.CommonLogMiddleware())

    // Combined log format
    router.Middleware(log.CombinedLogMiddleware())
}
```

Each request the server receives will be logged using the `goyave.AccessLogger` logger.

#### log.CommonLogMiddleware

CommonLogMiddleware captures response data and outputs it to the default logger using the common log format.

| Parameters | Return              |
|------------|---------------------|
|            | `goyave.Middleware` |

#### log.CombinedLogMiddleware

CombinedLogMiddleware captures response data and outputs it to the default logger using the combined log format.

| Parameters | Return              |
|------------|---------------------|
|            | `goyave.Middleware` |

### Custom format

It is possible to implement custom formatters for access logs. A `Formatter` is a function with the following signature:

``` go
func(now time.Time, response *goyave.Response, request *goyave.Request, length int) string
```

- `now` is the time at which the server received the request
- `length` the length of the response body

**Example:**
``` go
func CustomFormatter(now time.Time, response *goyave.Response, request *goyave.Request, length int) string {
  return fmt.Sprintf("%s %s %s %s %d %d",
    now.Format(TimestampFormat),
    host,
    req.Method,
    strconv.QuoteToASCII(uri),
    response.GetStatus(),
    length,
  )
}
```

#### log.Middleware

Middleware captures response data and outputs it to the default logger using the given formatter.

| Parameters            | Return              |
|-----------------------|---------------------|
| `formatter Formatter` | `goyave.Middleware` |

**Example:**
``` go
router.Middleware(log.Middleware(CustomFormatter))
```