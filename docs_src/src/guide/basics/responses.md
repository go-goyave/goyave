# Responses

Handlers receive a `goyave.Response` and a `goyave.Request` as parameters. This section is a technical reference of the `Response` object.

`goyave.Response` implements `http.ResponseWriter`. This object brings a number of convenient methods to write HTTP responses.

If you didn't write anything before the request lifecycle ends, `204 No Content` is automatically written.

All functions below require the `goyave` package to be imported.

``` go
import "github.com/System-Glitch/goyave"
```

**List of response methods**:
::: table
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
[CreateTestResponse](#response-createtestresponse)
:::

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

Write the given status code.

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

Print the error in the console and return it with an error code 500. If debugging is enabled in the config, the error is also written in the response using the JSON format, and the stacktrace is printed in the console.

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

#### Response.CreateTestResponse

Create an empty response with the given response writer. This function is aimed at making it easier to unit test Responses.

| Parameters                     | Return |
|--------------------------------|--------|
| `recorder http.ResponseWriter` | `void` |

**Example:**
``` go
writer := httptest.NewRecorder()
response := goyave.CreateTestResponse(writer)
response.Status(http.StatusNoContent)
result := writer.Result()
fmt.Println(result.StatusCode) // 204
```
