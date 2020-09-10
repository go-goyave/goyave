---
meta:
  - name: "og:title"
    content: "CORS - Goyave"
  - name: "twitter:title"
    content: "CORS - Goyave"
  - name: "title"
    content: "CORS - Goyave"
---

# CORS <Badge text="Since v2.3.0"/>

[[toc]]

## Introduction

CORS, or "[Cross-Origin Resource Sharing](https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS)" is a mechanism that uses additional HTTP headers to tell browsers to give a web application running at one origin, **access to selected resources from a different origin**. A web application executes a cross-origin HTTP request when it requests a resource that has a different origin (domain, protocol, or port) from its own. Enabling CORS is done by adding a set of specific headers allowing the browser and server to communicate about which requests, methods and headers are or are not allowed. CORS support also comes with **pre-flight** `OPTIONS` requests support.

Most of the time, the API is using another domain as the clients. For security reasons, browsers restrict cross-origin HTTP requests initiated from scripts. That's why you should configure CORS for your API.

## Enabling CORS

All functions below require the `cors` package to be imported.

``` go
import "github.com/System-Glitch/goyave/v3/cors"
```

CORS options are set on **routers**. If the passed options are not `nil`, the CORS core middleware is automatically added.

``` go
router.CORS(cors.Default())
```

CORS options should be defined **before middleware and route definition**. All of this router's sub-routers **inherit** CORS options by default. If you want to remove the options from a sub-router, or use different ones, simply create another `cors.Options` object and assign it.

``` go
router.CORS(cors.Default())

subrouter := router.Subrouter("/products")
subrouter.CORS(nil) // Remove CORS options

options := cors.Default()
options.AllowCredentials = true
subrouter.CORS(options) // Different CORS options
```

::: tip
All routes defined in a router having CORS options will match the `OPTIONS` HTTP method to allow **pre-flight** requests, even if it's not explicitly told in the route definition.
:::

## Options

`cors.Default()` can be used as a starting point for custom configuration.

``` go
options := cors.Default()
options.AllowedOrigins = []string{"https://google.com", "https://images.google.com"}
router.CORS(options)
```

Find the options reference below:

### AllowOrigins

A list of origins a cross-domain request can be executed from. If the first value in the slice is `*` or if the slice is empty, all origins will be allowed.

**Type:** `[]string`  
**Default:** `["*"]`

### AllowedMethods

A list of methods the client is allowed to use with cross-domain requests.

**Type:** `[]string`  
**Default:** `["HEAD", "GET", "POST", "PUT", "PATCH", "DELETE"]`

### AllowedHeaders

A list of non simple headers the client is allowed to use with cross-domain requests. If the first value in the slice is `*`, all headers will be allowed. If the slice is empty, the request's headers will be reflected.

**Type:** `[]string`  
**Default:** `["Origin", "Accept", "Content-Type", "X-Requested-With", "Authorization"]`

### ExposedHeaders

Indicates which headers are safe to expose to the API of a CORS API specification.

**Type:** `[]string`  
**Default:** `[]`

### MaxAge

Indicates how long the results of a preflight request can be cached.

**Type:** `time.Duration`  
**Default:** `12 hours (43200 seconds)`

### AllowCredentials

Indicates whether the request can include user credentials like cookies, HTTP authentication or client side SSL certificates.

**Type:** `bool`  
**Default:** `false`

### OptionsPassthrough

Instructs **pre-flight** to let other potential next handlers to process the `OPTIONS` method. Turn this on if your application handles `OPTIONS`.

**Type:** `bool`  
**Default:** `false`
