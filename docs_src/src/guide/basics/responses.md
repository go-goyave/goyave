---
meta:
  - name: "og:title"
    content: "Responses - Goyave"
  - name: "twitter:title"
    content: "Responses - Goyave"
  - name: "title"
    content: "Responses - Goyave"
---

# Responses

[[toc]]

## Introduction

Handlers receive a `goyave.Response` and a `goyave.Request` as parameters.

`goyave.Response` implements `http.ResponseWriter`. This object brings a number of convenient methods to write HTTP responses.

If you didn't write anything before the request lifecycle ends, `204 No Content` is automatically written.

## Reference

All functions below require the `goyave` package to be imported.

``` go
import "github.com/System-Glitch/goyave/v3"
```

**List of response methods**:
::: table
[GetStatus](#response-getstatus)
[GetError](#response-geterror)
[GetStacktrace](#response-getstacktrace)
[IsEmpty](#response-isempty)
[IsHeaderWritten](#response-isheaderwritten)
[Header](#response-header)
[Status](#response-status)
[JSON](#response-json)
[String](#response-string)
[Write](#response-write)
[File](#response-file)
[Download](#response-download)
[Error](#response-error)
[Cookie](#response-cookie)
[Redirect](#response-redirect)
[TemporaryRedirect](#response-temporaryredirect)
[Render](#response-render)
[RenderHTML](#response-renderhtml)
[HandleDatabaseError](#response-handledatabaseerror)
:::

#### Response.GetStatus

Returns the response code for this request or `0` if not yet set.

| Parameters | Return |
|------------|--------|
|            | `int`  |

**Example:**
``` go
fmt.Println(response.GetStatus()) // 200
```

#### Response.GetError

Returns the value which caused a panic in the request's handling, or `nil`. The response error is also set when [`Error()`](#response-error) is called.

This method is mainly used in [status handlers](../advanced/status-handlers.html).

| Parameters | Return        |
|------------|---------------|
|            | `interface{}` |

**Example:**
``` go
fmt.Println(response.GetError()) // "panic: something wrong happened"
```

#### Response.GetStacktrace

Return the stacktrace of when the error occurred, or an empty string. The stacktrace is captured by the recovery middleware.

| Parameters | Return   |
|------------|----------|
|            | `string` |

**Example:**
``` go
fmt.Println(response.GetStacktrace()) // "goroutine 1 [running]:
									  //  main.main()
									  //	/tmp/sandbox930868764/prog.go:8 +0x39"
```

#### Response.IsEmpty

Return true if nothing has been written to the response body yet.

| Parameters | Return |
|------------|--------|
|            | `bool` |

**Example:**
``` go
fmt.Println(response.IsEmpty()) // true
```

#### Response.IsHeaderWritten

return true if the response header has been written. Once the response header is written, you cannot change the response status and headers anymore.

| Parameters | Return |
|------------|--------|
|            | `bool` |

**Example:**
``` go
fmt.Println(response.IsHeaderWritten()) // false
```

#### Response.Header

Returns the Header map that will be sent.

| Parameters | Return        |
|------------|---------------|
|            | `http.Header` |

**Example:**
``` go
header := response.Header()
header.Set("Content-Type", "application/json")
```

#### Response.Status

Set the response status code. Calling this method a second time will have no effect.

| Parameters   | Return |
|--------------|--------|
| `status int` | `void` |

**Example:**
``` go
response.Status(http.StatusOK)
```

#### Response.JSON

Write JSON data as a response. This method automatically sets the `Content-Type` header.

| Parameters         | Return  |
|--------------------|---------|
| `responseCode int` | `error` |
| `data interface{}` |         |

**Example:**
``` go
response.JSON(http.StatusOK, map[string]interface{}{
    "name": "John Doe",
    "tags": []string{"tag1", "tag2"},
})
```

#### Response.String

Write a string as a response.

| Parameters         | Return  |
|--------------------|---------|
| `responseCode int` | `error` |
| `message string`   |         |

**Example:**
``` go
response.String(http.StatusOK, "Hello there!")
```

#### Response.Write

Write the data as a response. Can be used to write in-memory files. This method can be called successively.

Returns the number of bytes written.

| Parameters    | Return  |
|---------------|---------|
| `data []byte` | `int`   |
|               | `error` |

**Example:**
``` go
response.Write([]byte("Hello there!"))
```

#### Response.File

Write a file as an inline element.

Automatically detects the file MIME type and sets the "Content-Type" header accordingly. If the file doesn't exist, respond with status `404 Not Found`. The given path can be relative or absolute.

If you want the file to be sent as a download ("Content-Disposition: attachment"), use the `Download` function instead.

| Parameters    | Return  |
|---------------|---------|
| `file string` | `error` |

**Example:**
``` go
response.File("/path/to/file")
```

#### Response.Download

Write a file as an attachment element.

Automatically detects the file MIME type and sets the "Content-Type" header accordingly. If the file doesn't exist, respond with status `404 Not Found`. The given path can be relative or absolute.

The `fileName` parameter defines the name the client will see. In other words, it sets the header "Content-Disposition" to "attachment; filename="${fileName}""

If you want the file to be sent as an inline element ("Content-Disposition: inline"), use the `File` function instead.

| Parameters        | Return  |
|-------------------|---------|
| `file string`     | `error` |
| `fileName string` |         |

**Example:**
``` go
response.Download("/path/to/file", "awesome.txt")
```

#### Response.Error

Print the error in the console and return it with an error code `500`.

If debugging is enabled in the config, the error is also written in the response using the JSON format, and the stacktrace is printed in the console. If debugging is not enabled, only the stauts code is set, which means you can still write to the response, or use your error [status handler](../advanced/status-handlers.html).

| Parameters        | Return  |
|-------------------|---------|
| `err interface{}` | `error` |

**Example:**
``` go
v, err := strconv.Atoi("-42")
response.Error(err)
```

#### Response.Cookie

Add a Set-Cookie header to the response. The provided cookie must have a valid Name. Invalid cookies may be silently dropped.

| Parameters             | Return |
|------------------------|--------|
| `cookie *http.Cookie*` | `void` |

**Example:**
``` go
cookie := &http.Cookie{
    Name:  "cookie-name",
    Value: "value",
}
response.Cookie(cookie)
```

::: warning
Protect yourself from [CSRF attacks](https://en.wikipedia.org/wiki/Cross-site_request_forgery) when using cookies!
:::

#### Response.Redirect

Send a permanent redirect response. (HTTP 308)

| Parameters   | Return |
|--------------|--------|
| `url string` | `void` |

**Example:**
``` go
response.Redirect("/login")
```

#### Response.TemporaryRedirect

Send a temporary redirect response. (HTTP 307)

| Parameters   | Return |
|--------------|--------|
| `url string` | `void` |

**Example:**
``` go
response.TemporaryRedirect("/maintenance")
```

#### Response.Render

Render a text template with the given data. This method uses the [Go's template API](https://golang.org/pkg/text/template/).

The template path is relative to the `resources/template` directory.

| Parameters            | Return  |
|-----------------------|---------|
| `responseCode int`    | `error` |
| `templatePath string` |         |
| `data interface{}`    |         |

**Example:**
``` go
type Inventory struct {
	Material string
	Count    uint
}

sweaters := Inventory{"wool", 17}

// data can also be a map[string]interface{}
// Here, "resources/template/template.txt" will be used.
if err := response.Render(http.StatusOK, "template.txt", sweaters); err != nil {
	response.Error(err)
}
```

#### Response.RenderHTML

Render an HTML template with the given data. This method uses the [Go's template API](https://golang.org/pkg/html/template/).

The template path is relative to the `resources/template` directory.

| Parameters            | Return  |
|-----------------------|---------|
| `responseCode int`    | `error` |
| `templatePath string` |         |
| `data interface{}`    |         |

**Example:**
``` go
type Inventory struct {
	Material string
	Count    uint
}

sweaters := Inventory{"wool", 17}

// data can also be a map[string]interface{}
// Here, "resources/template/inventory.html" will be used.
if err := response.RenderHTML(http.StatusOK, "inventory.html", sweaters); err != nil {
	response.Error(err)
}
```

#### Response.HandleDatabaseError

Takes a database query result and checks if any error has occurred.

Automatically writes HTTP status code 404 Not Found if the error is a "Not found" error. Calls `Response.Error()` if there is another type of error.

Returns `true` if there is no error.

| Parameters    | Return |
|---------------|--------|
| `db *gorm.DB` | `bool` |

**Example:**
``` go
product := model.Product{}
result := database.Conn().First(&product, id)
if response.HandleDatabaseError(result) {
    response.JSON(http.StatusOK, product)
}
```

## Chained writers

<p><Badge text="Since v2.7.0"/></p>

It is possible to replace the `io.Writer` used by the `Response` object. This allows for more flexibility when manipulating the data you send to the client. It makes it easier to compress your response, write it to logs, etc. You can chain as many writers as you want.

Note that at the time your writer's `Write()` method is called, the request header **is already written**, therefore, changing headers or status doesn't have any effect. If you want to alter the headers, do so in a `PreWrite(b []byte)` function (from the `goyave.PreWriter` interface).

The writer replacement is most often done in a middleware. If your writer implements `io.Closer`, it will be automatically closed at the end of the request's lifecycle.

The following example is a simple implementation of a middleware logging everything sent by the server to the client.
``` go
import (
	"io"
	"log"

	"github.com/System-Glitch/goyave/v3"
)

type LogWriter struct {
	writer   io.Writer
	response *goyave.Response
	body     []byte
}

func (w *LogWriter) PreWrite(b []byte) {
	// All chained writers should implement goyave.PreWriter
	// to allow the modification of headers and status before
	// they are written.
	if pr, ok := w.writer.(goyave.PreWriter); ok {
		pr.PreWrite(b)
	}
}

func (w *LogWriter) Write(b []byte) (int, error) {
	w.body = append(w.body, b...)
	return w.writer.Write(b)
}

func (w *LogWriter) Close() error {
    // The chained writer MUST be closed if it's closeable.
	// Therefore, all chained writers should implement io.Closer.

	goyave.Logger.Println("RESPONSE", w.response.GetStatus(), string(w.body))

	if wr, ok := w.writer.(io.Closer); ok {
		return wr.Close()
	}
	return nil
}

func LogMiddleware(next goyave.Handler) goyave.Handler {
	return func(response *goyave.Response, request *goyave.Request) {
		logWriter := &LogWriter{
			writer:   response.Writer(),
			response: response,
		}
		response.SetWriter(logWriter)

		next(response, request)
	}
}
```

#### Response.Writer

Return the current writer used to write the response.

Note that the returned writer is not necessarily a `http.ResponseWriter`, as it can be replaced using `SetWriter`.

| Parameters | Return      |
|------------|-------------|
|            | `io.Writer` |

#### Response.SetWriter

Set the writer used to write the response.

This can be used to chain writers, for example to enable gzip compression, or for logging. The original `http.ResponseWriter` is always kept.

| Parameters         | Return |
|--------------------|--------|
| `writer io.Writer` |        |
