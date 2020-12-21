---
meta:
  - name: "og:title"
    content: "Requests - Goyave"
  - name: "twitter:title"
    content: "Requests - Goyave"
  - name: "title"
    content: "Requests - Goyave"
---

# Requests

[[toc]]

## Introduction

Handlers receive a `goyave.Response` and a `goyave.Request` as parameters. This section is a technical reference of the `Request` object. This object can give you a lot of information about the incoming request, such as its headers, cookies, or body.

All functions below require the `goyave` package to be imported.

``` go
import "github.com/System-Glitch/goyave/v3"
```

## Methods

::: table
[Request](#request-request)
[Method](#request-method)
[Protocol](#request-protocol)
[URI](#request-uri)
[Header](#request-header)
[ContentLength](#request-contentlength)
[RemoteAddress](#request-remoteaddress)
[Cookies](#request-cookies)
[Referrer](#request-referrer)
[UserAgent](#request-useragent)
[CORSOptions](#request-corsoptions)
[ToStruct](#request-tostruct)
[BasicAuth](#request-basicauth)
[BearerToken](#request-bearertoken)
:::

#### Request.Request

Return the raw http request. Prefer using the "goyave.Request" accessors.

| Parameters | Return          |
|------------|-----------------|
|            | `*http.Request` |

#### Request.Method

The HTTP method (GET, POST, PUT, etc.).

| Parameters | Return   |
|------------|----------|
|            | `string` |

**Example:**
``` go
fmt.Println(request.Method()) // "GET"
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

#### Request.URI

URI specifies the URI being requested. Use this if you absolutely need the raw query params, url, etc. Otherwise use the provided methods and fields of the `goyave.Request`.

| Parameters | Return     |
|------------|------------|
|            | `*url.URL` |

**Example:**
``` go
fmt.Println(request.URI().Path) // "/foo/bar"
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

::: warning
Protect yourself from [CSRF attacks](https://en.wikipedia.org/wiki/Cross-site_request_forgery) when using cookies!
:::

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

#### Request.CORSOptions

<p><Badge text="Since v2.3.0"/></p>

Returns the CORS options applied to this request, or `nil`. Learn more about CORS [here](../advanced/cors.html).

The returned object is a copy of the options applied to the router. Therefore, altering the returned object will not alter the router's options.

| Parameters | Return          |
|------------|-----------------|
|            | `*cors.Options` |

**Example:**
``` go
fmt.Println(request.CORSOptions().AllowedMethods) // "[HEAD GET POST PUT PATCH DELETE]"
```

#### Request.ToStruct

ToStruct map the request data to a struct. The given "dst" parameter should be a struct pointer.

| Parameters        | Return  |
|-------------------|---------|
| `dst interface{}` | `error` |

**Example:**
``` go
type UserInsertRequest struct {
  Username string
  Email    string
}

userInsertRequest := UserInsertRequest{}

if err := request.ToStruct(&userInsertRequest); err != nil {
  panic(err)
}

fmt.Println(userInsertRequest) // {johndoe johndoe@example.org}
```

#### Request.BasicAuth

Returns the username and password provided in the request's `Authorization` header, if the request uses HTTP Basic Authentication.

| Parameters | Return            |
|------------|-------------------|
|            | `username string` |
|            | `password string` |
|            | `ok bool`         |

**Example:**
``` go
username, password, ok := request.BasicAuth()
fmt.Println(username) // "admin"
fmt.Println(password) // "secret"
fmt.Println(ok) // true
```

#### Request.BearerToken

<p><Badge text="Since v2.5.0"/></p>

Extract the auth token from the "Authorization" header. Only takes tokens of type "Bearer".

Returns empty string if no token found or the header is invalid.

| Parameters | Return         |
|------------|----------------|
|            | `token string` |
|            | `ok bool`      |

**Example:**
``` go
token, ok := request.BearerToken()
fmt.Println(token) // "ey..."
fmt.Println(ok) // true
```

### Accessors

<p><Badge text="Since v2.0.0"/></p>

Accessors are helper functions to retrieve request data without having to write the type assertion. This is helpful to make your controllers cleaner. You shouldn't use these accessors in middleware because they assume the data has been converted by the validation already. Data can still be accessed through the `Data` attribute. There is currently no accessor for slices.

::: table
[Has](#request-has)
[String](#request-string)
[Numeric](#request-numeric)
[Integer](#request-integer)
[Bool](#request-bool)
[File](#request-file)
[Timezone](#request-timezone)
[IP](#request-ip)
[URL](#request-url)
[UUID](#request-uuid)
[Date](#request-date)
[Object](#request-object)
:::

#### Request.Has

Check if the given field exists in the request data.

| Parameters     | Return |
|----------------|--------|
| `field string` | `bool` |

**Example:**
``` go
fmt.Println(request.Has("name")) // true
```

#### Request.String

Get a string field from the request data. Panics if the field is not a string or doesn't exist.

| Parameters     | Return   |
|----------------|----------|
| `field string` | `string` |

**Example:**
``` go
fmt.Println(request.String("name")) // "JohnDoe"
```

#### Request.Numeric

Get a numeric field from the request data. Panics if the field is not numeric or doesn't exist.

| Parameters     | Return    |
|----------------|-----------|
| `field string` | `float64` |

**Example:**
``` go
fmt.Println(request.Numeric("price")) // 42.3
```

#### Request.Integer

Get an integer field from the request data. Panics if the field is not an integer or doesn't exist.

| Parameters     | Return |
|----------------|--------|
| `field string` | `int`  |

**Example:**
``` go
fmt.Println(request.Integer("age")) // 32
```

#### Request.Bool

Get a bool field from the request data. Panics if the field is not a bool or doesn't exist.

| Parameters     | Return |
|----------------|--------|
| `field string` | `bool` |

**Example:**
``` go
fmt.Println(request.Bool("EULA")) // true
```

#### Request.File

Get a file field from the request data. Panics if the field is not a file or doesn't exist.

| Parameters     | Return              |
|----------------|---------------------|
| `field string` | `[]filesystem.File` |

**Example:**
``` go
for _, f := range request.File("images") {
    fmt.Println(f.Header.Filename)
}
```

#### Request.Timezone

Get a timezone field from the request data. Panics if the field is not a timezone or doesn't exist.

| Parameters     | Return           |
|----------------|------------------|
| `field string` | `*time.Location` |

**Example:**
``` go
fmt.Println(request.Timezone("tz").String()) // "America/New_York"
```

#### Request.IP

Get an IP field from the request data. Panics if the field is not an IP or doesn't exist.

| Parameters     | Return   |
|----------------|----------|
| `field string` | `net.IP` |

**Example:**
``` go
fmt.Println(request.IP("host").String()) // "127.0.0.1"
```

#### Request.URL

Get a URL field from the request data. Panics if the field is not a URL or doesn't exist.

| Parameters     | Return     |
|----------------|------------|
| `field string` | `*url.URL` |

**Example:**
``` go
fmt.Println(request.URL("link").String()) // "https://google.com"
```

#### Request.UUID

Get a UUID field from the request data. Panics if the field is not a UUID or doesn't exist.

| Parameters     | Return      |
|----------------|-------------|
| `field string` | `uuid.UUID` |

**Example:**
``` go
fmt.Println(request.UUID("id").String()) // "3bbcee75-cecc-5b56-8031-b6641c1ed1f1"
```

#### Request.Date

Get a date field from the request data. Panics if the field is not a date or doesn't exist.

| Parameters     | Return      |
|----------------|-------------|
| `field string` | `time.Time` |

**Example:**
``` go
fmt.Println(request.Date("birthdate").String()) // "2019-11-21 00:00:00 +0000 UTC"
```

#### Request.Object

Get an object field from the request data. Panics if the field is not an object or doesn't exist.

| Parameters     | Return                   |
|----------------|--------------------------|
| `field string` | `map[string]interface{}` |

**Example:**
``` go
fmt.Println(request.Object("object")) // map[hello:world]
```

## Attributes

::: table
[Rules](#request-rules)
[Data](#request-data)
[Params](#request-params)
[Lang](#request-lang)
[Extra](#request-extra)
:::

#### Request.Rules

The validation rule set associated with this request. See the [validation](./validation.html) section for more information.

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
name=John%20Doe&tags=tag1&tags=tag2
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

Learn more in the [localization](../advanced/localization.html) section.

``` go
fmt.Println(request.Lang) // "en-US"
fmt.Println(lang.Get(request.Lang, "validation.rules.required")) // "The :field is required."
```

#### Request.Extra

<p><Badge text="Since v3.3.0"/></p>

`Extra` is a `map[string]interface{}` meant to contain extra data, which is not part of the request body (`request.Data`). This allows middleware to process some data and pass it to other handlers.

**Example:**
``` go
func MyCustomMiddleware(next goyave.Handler) goyave.Handler {
	return func(response *goyave.Response, request *goyave.Request) {
        request.Extra["custom"] = "extra data"
        next(response, request) // Pass to the next handler with the extra data
    }
}
```

#### Request.User

<p><Badge text="Since v2.5.0"/></p>

`User` is an `interface{}` containing the authenticated user if the route is protected, `nil` otherwise.

Learn more in the [authentication](../advanced/authentication.html) section.

**Example:**
``` go
fmt.Println(request.User.(*model.User).Name) // "John Doe"
```
