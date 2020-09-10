---
meta:
  - name: "og:title"
    content: "Status Handlers - Goyave"
  - name: "twitter:title"
    content: "Status Handlers - Goyave"
  - name: "title"
    content: "Status Handlers - Goyave"
---

# Status Handlers <Badge text="Since v2.4.0"/>

[[toc]]

## Introduction

Status handlers are regular handlers executed during the **finalization** step of the [request's lifecycle](../architecture-concepts.html#requests) if the response body is empty but a status code has been set. Status handler are mainly used to implement a custom behavior for user or server errors (400 and 500 status codes).

Goyave comes with a default error status handler. When a panic occurs or the [`Response.Error()`](../basics/responses.html#response-error) method is called, the request's status is set to `500 Internal Server Error` and the request error is set. The latter can be accessed using [`Response.GetError()`](../basics/responses.html#response-geterror). The error is printed in the console. If debugging is enabled in the config, the error is also written in the response using the JSON format, and the stacktrace is printed in the console. If debugging is not enabled, the following is returned:

``` json
{
    "error": "Internal Server Error"
}
```

The status handler covering all the other errors in the `400` and `500` status codes ranges has a similar behavior but doesn't print anything to the console. For example, if the user requests a route that doesn't exist, the following is returned:

``` json
{
    "error": "Not Found"
}
```

## Writing status handlers

As said earlier, status handlers are regular handlers. The only difference is that they are executed at the very end of the request's lifecycle. Ideally, create a new controller for your status handlers.

`http/controller/status/status.go`:
``` go
package status

import "github.com/System-Glitch/goyave/v3"

func NotFound(response *goyave.Response, request *goyave.Request) {
    if err := response.RenderHTML(response.GetStatus(), "errors/404.html", nil); err != nil {
        response.Error(err)
    }
}
```

::: warning
Avoid panicking in status handlers, as they are **not protected** by the recovery middleware!
:::

### Expanding default status handlers

You can expand default status handlers by calling them in your custom status handler. This is especially useful if you want to use error tracking software without altering the default behavior.

```go
func Panic(response *goyave.Response, request *goyave.Request) {
    tracker.ReportError(response.GetError())
    goyave.PanicStatusHandler(response, request)
}
```

## Registering status handlers

Status handlers are registered in the **router**.

#### Router.StatusHandler

Set a handler for responses with an empty body. The handler will be automatically executed if the request's life-cycle reaches its end and nothing has been written in the response body.

Multiple status codes can be given. The handler will be executed if one of them matches.

Status handlers are **inherited** as a copy in sub-routers. Modifying a child's status handler will not modify its parent's. That means that you can define different status handlers for certain route groupes if you so desire.

| Parameters                  | Return |
|-----------------------------|--------|
| `handler Handler`           | `void` |
| `status int`                |        |
| `additionalStatuses ...int` |        |

**Example:**
``` go
func errorStatusHandler(response *Response, request *Request) {
	message := map[string]string{
		"error": http.StatusText(response.GetStatus()),
	}
	response.JSON(response.GetStatus(), message)
}

// Use "errorStatusHandler" for empty responses having status 404 or 405.
router.StatusHandler(errorStatusHandler, 404, 405)
```
