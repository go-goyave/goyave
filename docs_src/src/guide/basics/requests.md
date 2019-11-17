# Requests

[[toc]]

## Introduction

Handlers receive a `goyave.Response` and a `goyave.Request` as parameters. This section is a technical reference of the `Request` object. This object can give you a lot of information about the incoming request, such as its headers, cookies, or body.

All functions below require the `goyave` package to be imported.

``` go
import "github.com/System-Glitch/goyave"
```

## Methods

#### Request.Method

The HTTP method (GET, POST, PUT, etc.).

| Parameters | Return   |
|------------|----------|
|            | `string` |

**Example:**
``` go
fmt.Println(request.Method()) // GET
```

#### Request.Protocol

The protocol used by this request, "HTTP/1.1" for example.

| Parameters | Return   |
|------------|----------|
|            | `string` |

**Example:**
``` go
fmt.Println(request.Protocol()) // "HTTP/1.1"
```

#### Request.URL

URL specifies the URI being requested. Use this if you absolutely need the raw query params, url, etc. Otherwise use the provided methods and fields of the `goyave.Request`.

| Parameters | Return     |
|------------|------------|
|            | `*url.URL` |

**Example:**
``` go
fmt.Println(request.URL().Path) // "/foo/bar"
```

#### Request.Header

Header contains the request header fields either received by the server or to be sent by the client. Header names are case-insensitive.

If the raw request has the following header lines,
```
Host: example.com
accept-encoding: gzip, deflate
Accept-Language: en-us
fOO: Bar
foo: two
```
then the header map will look like this:
```go
Header = map[string][]string{
    "Accept-Encoding": {"gzip, deflate"},
    "Accept-Language": {"en-us"},
    "Foo": {"Bar", "two"},
}
```

| Parameters | Return        |
|------------|---------------|
|            | `http.Header` |

**Example:**
``` go
fmt.Println(request.Header().Get("Content-Type")) // "application/json"
```

#### Request.ContentLength

ContentLength records the length (in bytes) of the associated content. The value -1 indicates that the length is unknown.

| Parameters | Return  |
|------------|---------|
|            | `int64` |

**Example:**
``` go
fmt.Println(request.ContentLength()) // 758
```

#### Request.RemoteAddress

RemoteAddress allows to record the network address that sent the request, usually for logging.

| Parameters | Return   |
|------------|----------|
|            | `string` |

**Example:**
``` go
fmt.Println(request.RemoteAddress()) // 192.168.0.10:1234
```

#### Request.Cookies

Cookies returns the HTTP cookies sent with the request.

| Parameters    | Return           |
|---------------|------------------|
| `name string` | `[]*http.Cookie` |

**Example:**
``` go
cookie := request.Cookies("cookie-name")
fmt.Println(cookie[0].value)
```

#### Request.Referrer

Referrer returns the referring URL, if sent in the request.

| Parameters | Return   |
|------------|----------|
|            | `string` |

**Example:**
``` go
fmt.Println(request.Referrer()) // "https://github.com/System-Glitch/goyave/"
```

#### Request.UserAgent

UserAgent returns the client's User-Agent, if sent in the request.

| Parameters | Return   |
|------------|----------|
|            | `string` |

**Example:**
``` go
fmt.Println(request.UserAgent()) // "Mozilla/5.0 ..."
```

## Attributes

#### Request.Rules

The validation rule set associated with this request. See the [validation](./validation) section for more information.

#### Request.Data

A `map[string]interface{}` containing the request's data. The key is the name of the field. This map contains the data from the request's body **and** the URL query string values. The request's body parameters takes precedence over the URL query string values.

For the given JSON request:
```json
{
    "name": "John Doe",
    "tags": ["tag1", "tag2"]
}
```
``` go
fmt.Println(request.Data["name"]) // "John Doe"
fmt.Println(request.Data["tags"]) // "[tag1 tag2]" ([]string)
```

You would obtain the same result for the following url-encoded request:
```
name=John%20Doe&tags[]=tag1&tags[]=tag2
```

Because the `Data` attribute can hold any type of data, you will likely need type assertion. If you validated your request, checking for type assertion errors is not necessary.
``` go
tags, _ := request.Data["tags"].([]string)
```

#### Request.Params

`Params` is a `map[string]string` of the route parameters.

For the given route definition and request:
```
/categories/{category}/{product_id}
```
```
/categories/3/5
```
``` go
fmt.Println(request.Params["category"]) // "3"
fmt.Println(request.Params["product_id"]) // "5"
```

#### Request.Lang

`Lang` indicates the language desired by the user. This attribute is automatically set by a core middleware, based on the `Accept-Language` header, if present. If the desired language is not available, the default language is used.

Learn more in the [localization](../advanced/localization) section.

``` go
fmt.Println(request.Lang) // "en-US"
fmt.Println(lang.Get(request.Lang, "validation.rules.required")) // "The :field is required."
```
