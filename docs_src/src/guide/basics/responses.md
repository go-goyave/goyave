# Responses

Handlers receive a `goyave.Response` and a `goyave.Request` as parameters. This section is a technical reference of the `Response` object.

`goyave.Response` implements `http.ResponseWriter`. This object brings a number of convenient methods to write HTTP responses.

If you didn't write anything before the request lifecycle ends, `204 No Content` is automatically written.

All functions below require the `goyave` package to be imported.

``` go
import "github.com/System-Glitch/goyave/v2"
```

**List of response methods**:
::: table
[GetStatus](#response-getstatus)
[GetError](#response-geterror)
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

Write the given status code. Calling this method a second time will have no effect.

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

Automatically detects the file MIME type and sets the "Content-Type" header accordingly. It is advised to call `filesystem.FileExists()` before sending a file to avoid a panic and return a 404 error. The given path can be relative or absolute.

If you want the file to be sent as a download ("Content-Disposition: attachment"), use the "Download" function instead.

| Parameters    | Return  |
|---------------|---------|
| `file string` | `error` |

**Example:**
``` go
response.File("/path/to/file")
```

#### Response.Download

Write a file as an attachment element.

Automatically detects the file MIME type and sets the "Content-Type" header accordingly. It is advised to call `filesystem.FileExists()` before sending a file to avoid a panic and return a 404 error if the file doesn't exist. The given path can be relative or absolute.

If you want the file to be sent as a download ("Content-Disposition: attachment"), use the "Download" function instead.

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
response.Render(http.StatusOK, "template.txt", sweaters)
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
response.RenderHTML(http.StatusOK, "inventory.html", sweaters)
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
result := database.GetConnection().First(&product, id)
if response.HandleDatabaseError(result) {
    response.JSON(http.StatusOK, product)
}
```
