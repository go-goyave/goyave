---
meta:
  - name: "og:title"
    content: "Routing - Goyave"
  - name: "twitter:title"
    content: "Routing - Goyave"
  - name: "title"
    content: "Routing - Goyave"
---

# Routing

[[toc]]

## Introduction

Routing is an essential part of any Goyave application. Routes definition is the action of associating a URI, sometimes having parameters, with a handler which will process the request and respond to it. Separating and naming routes clearly is important to make your API or website clear and expressive.

All features below require the `goyave` package to be imported.

``` go
import "github.com/System-Glitch/goyave/v3"
```

Routes are defined in **routes registrer functions**. The main route registrer is passed to `goyave.Start()` and is executed automatically with a newly created root-level **router**.

``` go
func Register(router *goyave.Router) {
    // Register your routes here
}
```

## Basic routing

Although it's not recommended, routes can be defined using **closures**. This is a very simple way of defining routes that can be used for scaffolding or quick testing.

``` go
router.Route("GET", "/hello", func(response *goyave.Response, r *goyave.Request) {
    response.String(http.StatusOK, "Hi!")
})
```

#### Router.Route

Register a new route.  
Multiple methods can be passed using a pipe-separated string.

If the route matches the `GET` method, the `HEAD` method is automatically added to the matcher if it's missing.

If the router has CORS options set, the `OPTIONS` method is automatically added to the matcher if it's missing, so it allows pre-flight requests.

Returns the generated route.

| Parameters               | Return          |
|--------------------------|-----------------|
| `methods string`         | `*goyave.Route` |
| `uri string`             |                 |
| `handler goyave.Handler` |                 |

**Examples:**
``` go
router.Route("GET", "/hello", myHandlerFunction)
router.Route("POST", "/user", user.Register)
router.Route("PUT|PATCH", "/user", user.Update)
router.Route("POST", "/product", product.Store)
```

::: tip
`goyave.Handler` is an alias for `func(*goyave.Response, *goyave.Request)`.
:::

You can also register routes by using the `Get`, `Post`, `Put`, `Patch`, `Delete` and `Options` methods:
``` go
router.Get("/hello", myHandlerFunction)
router.Post("/user", user.Register)
router.Put("/product/{id:[0-9]+}", product.Update)
router.Patch("/product/{id:[0-9]+}", product.Update)
router.Delete("/product/{id:[0-9]+}", product.Destroy)
router.Options("/options", myHandlerFunction)
```

| Parameters               | Return          |
|--------------------------|-----------------|
| `uri string`             | `*goyave.Route` |
| `handler goyave.Handler` |                 |

## Route reference

<p><Badge text="Since v2.6.0"/></p>

::: table
[Name](#route-name)
[GetName](#route-getname)
[BuildURL](#route-buildurl)
[GetURI](#route-geturi)
[GetFullURI](#route-getfulluri)
[GetMethods](#route-getmethods)
[Validate](#route-validate)
[Middleware](#route-middleware)
::: 

#### Route.Name

Set the name of this route.

Panics if a route with the same name already exists.

Returns itself.

| Parameters    | Return          |
|---------------|-----------------|
| `name string` | `*goyave.Route` |

**Examples:**
``` go
router.Route("GET", "/product/{id:[0-9]+}", myHandlerFunction).Name("product.show")
```

#### Route.GetName

Get the name of this route.

| Parameters | Return   |
|------------|----------|
|            | `string` |

**Examples:**
``` go
fmt.Println(route.GetName()) // "product.create"
```

#### Route.BuildURL

Build a full URL pointing to this route.

Panics if the amount of parameters doesn't match the amount of actual parameters for this route.

| Parameters             | Return   |
|------------------------|----------|
| `parameters ...string` | `string` |

**Examples:**
``` go
fmt.Println(route.BuildURL("42")) // "http://localhost:8080/product/42"
```

#### Route.GetURI

Get the URI of this route.

The returned URI is relative to the parent router of this route, it is NOT the full path to this route.

Note that this URI may contain route parameters in their définition format. Use the request's URI if you want to see the URI as it was requested by the client.

| Parameters | Return   |
|------------|----------|
|            | `string` |

**Examples:**
``` go
fmt.Println(route.GetURI()) // "/{id:[0-9]+}"
```

#### Route.GetFullURI

Get the full URI of this route.

Note that this URI may contain route parameters in their définition format. Use the request's URI if you want to see the URI as it was requested by the client.

| Parameters | Return   |
|------------|----------|
|            | `string` |

**Examples:**
``` go
fmt.Println(route.GetFullURI()) // "/product/{id:[0-9]+}"
```

#### Route.GetMethods

Returns the methods the route matches against.

| Parameters | Return     |
|------------|------------|
|            | `[]string` |

**Examples:**
``` go
fmt.Println(route.GetMethods()) // [GET OPTIONS]
```

#### Route.Validate

<p><Badge text="Since v3.0.0"/></p>

Validate adds validation rules to this route. If the user-submitted data doesn't pass validation, the user will receive an error and messages explaining what is wrong.

Returns itself.

| Parameters                         | Return          |
|------------------------------------|-----------------|
| `validationRules validation.Ruler` | `*goyave.Route` |

**Examples:**
``` go
router.Post("/user", user.Register).Validate(user.RegisterRequest)
```

::: tip
Learn more about validation and rules sets [here](./validation.html).
:::

#### Route.Middleware

<p><Badge text="Since v3.0.0"/></p>

Register middleware for this route only.

Returns itself.

| Parameters                        | Return          |
|-----------------------------------|-----------------|
| `middleware ...goyave.Middleware` | `*goyave.Route` |

**Examples:**
``` go
router.Put("/product", product.Update).Middleware(middleware.Admin)
```

::: tip
Learn more about middleware [here](./middleware.html).
:::

## Route parameters

URIs can have parameters, defined using the format `{name}` or `{name:pattern}`. If a regular expression pattern is not defined, the matched variable will be anything until the next slash. 

**Example:**
``` go
router.Get("/product/{key}", product.Show)
router.Get("/product/{id:[0-9]+}", product.ShowById)
router.Get("/category/{category}/{id:[0-9]+}", category.Show)
```

Regex groups can be used inside patterns, as long as they are non-capturing (`(?:re)`). For example:
``` go
router.Get("/category/{category}/{sort:(?:asc|desc|new)}", category.ShowSorted)
```

Route parameters can be retrieved as a `map[string]string` in handlers using the request's `Params` attribute.
``` go
func myHandlerFunction(response *goyave.Response, request *goyave.Request) {
    category := request.Params["category"]
    id, _ := strconv.Atoi(request.Params["id"])
    //...
}
```

## Handling HEAD

The `HEAD` HTTP method requests the headers that would be returned if the request's URL was instead requested with the HTTP `GET` method. The `HEAD` method is automatically handled for routes matching the `GET` method. When a route is matched with the `HEAD` method, it is executed as usual but the response body is discarded. That means that database queries and other operations **are** executed.

You may want to add a route definition exclusively for the `HEAD` method to prevent expensive operations to be executed. Register it **before** the corresponding `GET` route so it will be matched first. Keep in mind the returned headers **should** be the same as the ones returned by the `GET` handler!

```go
router.Route("HEAD", "/test", func(response *Response, r *Request) {
    response.Header().Set("Content-Type", "application/json; charset=utf-8")
    response.Status(http.StatusOK)
})
router.Get("/test", func(response *Response, r *Request) {
    response.JSON(http.StatusOK, map[string]string{"message": "hello world"})
})
```

## Named routes

<p><Badge text="Since v2.6.0"/></p>

It is possible to give a name to your routes to make it easier to retrieve them later and build dynamic URLs.

``` go
router.Route("GET", "/product/{id:[0-9]+}", myHandlerFunction).Name("product.show")
```

The route can now be retrieved from any router or from the global helper:

``` go
route := router.GetRoute("product.show")
// or
route := goyave.GetRoute("product.show")

fmt.Println(route.BuildURL("42")) // "http://localhost:8080/product/42"
```

#### goyave.GetRoute

Get a named route. Returns nil if the route doesn't exist.

| Parameters    | Return          |
|---------------|-----------------|
| `name string` | `*goyave.Route` |

## Validation

You can assign a validation rules set to each route. Learn more in the dedicated [section](./validation.html). You should always validate incoming requests.

``` go
router.Route("POST", "/product", product.Store).Validate(validation.RuleSet{
	"Name":  {"required", "string", "min:4"},
	"Price": {"required", "numeric"},
})
```

::: tip
It's not recommended to define rules set directly in the route definition. You should define rules sets in your controller package.
:::

## Middleware

Middleware are handlers executed before the controller handler. Learn more in the dedicated [section](./middleware.html).

Middleware are applied to a router or a sub-router **before the routes definition**. Therefore, all routes in that router and its sub-routers will execute them before executing their associated handler.

To assign a middleware to a router, use the `router.Middleware()` function. Many middleware can be assigned at once. The assignment order is important as middleware will be **executed in order**.

#### Router.Middleware

Middleware apply one or more middleware to the route group.

| Parameters                        | Return |
|-----------------------------------|--------|
| `middleware ...goyave.Middleware` | `void` |

**Example:**
``` go
router.Middleware(middleware.DisallowNonValidatedFields)
```

---

Middleware can also be applied to specific routes. You can add as many as you want.

**Example:**
``` go
router.Route("POST", "/product", product.Store).Validate(product.StoreRequest).Middleware(middleware.Trim)
```

## Groups and sub-routers

Grouping routes makes it easier to define multiple routes having the same prefix and/or middleware.

Let's take a simple scenario where we want to implement a user CRUD. All our routes will start with `/user`, so we are going to create a sub-router for it:
``` go
userRouter := router.Subrouter("/user")
```

In our application, user profiles are public: anyone can see the user profiles without being authenticated. However, only authenticated users can modify their information and delete their account. We don't want to add some redundancy and apply the authentication middleware for each route needing it, so we are going to create another sub-router. Sub-routers having an empty prefix are called **route groups**.
```go
userRouter.Get("/{username}", user.Show)
userRouter.Post("", user.Register).Validate(user.RegisterRequest)

authUserRouter := userRouter.Subrouter("") // Don't add a prefix
authUserRouter.Middleware(authenticationMiddleware)
authUserRouter.Put("/{id}", user.Update).Validate(user.UpdateRequest)
authUserRouter.Delete("/{id}", user.Delete)
```

To improve your routes definition readability, you should create a new route registrer for each feature. In our example, our definitions would look like this:
``` go
func registerUserRoutes(router *goyave.Router) {
    //...
}

// Register is the main route registrer.
func Register(router *goyave.Router) {
    registerUserRoutes(router)
    registerProductRoutes(router)
    //...
}
```

Sub-routers are checked before routes, meaning that they have priority over the latter. If you have a router sharing a prefix with a higher-level level route, **it will never match** because the sub-router will match first.
``` go
subrouter := router.Subrouter("/product")
subrouter.Get("/{id:[0-9]+}", handler)

router.Get("/product/{id:[0-9]+}", handler) // This route will never match
router.Get("/product/category", handler)    // This one neither
```

## Serve static resources

The Goyave router provides a way to serve a directory of static resources, including its sub-directories.

Let's say you have the following directory structure:

:::vue
.
└── static
     ├── js
     │   └── index.js
     ├── img
     │   ├── favicon.ico
     │   └── logo.png
     ├── css
     │   └── styles.css
     └── index.html
:::

If you want to serve the `static` directory, register the following route:

``` go
router.Static("/", "static", false)
```

If a user requests `http://yourdomain.com/js/index.js`, the corresponding file will be sent as a response.

If no file is given (`http://yourdomain.com/`), or if the request URI is a directory (`http://yourdomain.com/img`), Goyave will look for a `index.html` file and send it if it exists. An error 404 Not Found is otherwise returned.

::: tip
This method is especially useful to serve Single Page Applications from your API. (Angular, Vue.js, React applications)
:::

#### Router.Static

Static serve a directory and its sub-directories of static resources.
Set the `download` parameter to true if you want the files to be sent as an attachment instead of an inline element.

The `directory` parameter can be a relative or an absolute path.

| Parameters                        | Return |
|-----------------------------------|--------|
| `uri string`                      | `void` |
| `directory string`                |        |
| `download bool`                   |        |
| `middleware ...goyave.Middleware` |        |

**Example:**
``` go
router.Static("/public", "/path/to/static/dir", false)
```

## Native handlers

<p><Badge text="Since v2.0.0"/></p>

#### goyave.NativeHandler

NativeHandler is an adapter function for `http.Handler`. With this adapter, you can plug non-Goyave handlers to your application.

Just remember that the body contains the raw data, which haven't been validated nor converted. This means that **native handlers are not guaranteed to work** and cannot modify the request data. Request properties, such as headers, can still be modified.

The actual response writer passed to the native handler is a `goyave.Response`.

::: warning
This feature is a compatibility layer with the rest of the Golang web ecosystem. Prefer using Goyave handlers if possible.
:::


| Parameters             | Return           |
|------------------------|------------------|
| `handler http.Handler` | `goyave.Handler` |

**Example:**
``` go
httpHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("Hello world"))
})
router.Route("GET", "/user", goyave.NativeHandler(httpHandler))
```

#### goyave.NativeMiddleware

`NativeMiddleware` is an adapter function for standard library middleware.

Native middleware work like native handlers. See [`NativeHandler`](#goyave-nativehandler) for more details.

| Parameters                               | Return              |
|------------------------------------------|---------------------|
| `middleware goyave.NativeMiddlewareFunc` | `goyave.Middelware` |

::: tip
`goyave.NativeMiddlewareFunc` is defined as follows:

``` go
// NativeMiddlewareFunc is a function which receives an http.Handler and returns another http.Handler.
type NativeMiddlewareFunc func(http.Handler) http.Handler
```
:::

**Example:**
``` go
middleware := goyave.NativeMiddleware(func(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Hello world"))
        next.ServeHTTP(w, r) // Don't call "next" if your middleware is blocking.
    })
})
router.Middleware(middleware)
```


