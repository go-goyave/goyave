# Routing

[[toc]]

## Introduction

Routing is an essential part of any Goyave application. Routes definition is the action of associating a URI, sometimes having parameters, with a handler which will process the request and respond to it. Separating and naming routes clearly is important to make your API or website clear and expressive.

All features below require the `goyave` package to be imported.

``` go
import "github.com/System-Glitch/goyave/v2"
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
}, nil)
```

#### Router.Route

Register a new route.  
Multiple methods can be passed using a pipe-separated string.

If the router has CORS options set, the `OPTIONS` method is automatically added to the matcher if it's missing, so it allows pre-flight requests.

Returns the generated route.

| Parameters                           | Return          |
|--------------------------------------|-----------------|
| `methods string`                     | `*goyave.Route` |
| `uri string`                         |                 |
| `handler goyave.Handler`             |                 |
| `validationRules validation.RuleSet` |                 |

**Examples:**
``` go
router.Route("GET", "/hello", myHandlerFunction, nil)
router.Route("POST", "/user", user.Register, userrequest.Register)
router.Route("PUT|PATCH", "/user", user.Update, userrequest.Update)
```

::: tip
- `goyave.Handler` is an alias for `func(*goyave.Response, *goyave.Request)`.
- Learn more about validation and rules sets [here](./validation.html).
:::

## Route reference

<p><Badge text="Since v2.6.0"/></p>

::: table
[Name](#route-name)
[GetName](#route-getname)
[BuildURL](#route-buildurl)
[GetURI](#route-geturi)
[GetMethods](#route-getmethods)
::: 

#### Route.Name

Set the name of this route.

Panics if a route with the same name already exists.

| Parameters    | Return |
|---------------|--------|
| `name string` | `void` |

**Examples:**
``` go
router.Route("GET", "/product/{id:[0-9]+}", myHandlerFunction, nil).Name("product.show")
```

#### Route.GetName

Get the name of this route.

| Parameters | Return   |
|------------|----------|
|            | `string` |

**Examples:**
``` go
fmt.Println(route.GetName()) // "product-create"
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

Note that this URI may contain route parameters in their définition format. Use the request's URI if you want to see the URI as it was requested by the client.

| Parameters | Return   |
|------------|----------|
|            | `string` |

**Examples:**
``` go
fmt.Println(route.GetURI()) // "/product/{id:[0-9]+}"
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

## Route parameters

URIs can have parameters, defined using the format `{name}` or `{name:pattern}`. If a regular expression pattern is not defined, the matched variable will be anything until the next slash. 

**Example:**
``` go
router.Route("GET", "/products/{key}", product.Show, nil)
router.Route("GET", "/products/{id:[0-9]+}", product.ShowById, nil)
router.Route("GET", "/categories/{category}/{id:[0-9]+}", category.Show, nil)
```

Regex groups can be used inside patterns, as long as they are non-capturing (`(?:re)`). For example:
``` go
router.Route("GET", "/categories/{category}/{sort:(?:asc|desc|new)}", category.ShowSorted, nil)
```

Route parameters can be retrieved as a `map[string]string` in handlers using the request's `Params` attribute.
``` go
func myHandlerFunction(response *goyave.Response, request *goyave.Request) {
    category := request.Params["category"]
    id, _ := strconv.Atoi(request.Params["id"])
    //...
}
```

## Named routes

It is possible to give a name to your routes to make it easier to retrieve them later and build dynamic URLs.

``` go
router.Route("GET", "/product/{id:[0-9]+}", myHandlerFunction, nil).Name("product.show")
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
router.Route("POST", "/products", product.Create, validation.RuleSet{
	"Name":  []string{"required", "string", "min:4"},
	"Price": []string{"required", "numeric"},
})
```

::: tip
It's not recommended to define rules set directly in the route definition. You should define rules sets in the `http/requests` directory and have one file per feature, regrouping all requests handled by the same controller. You can also create one package per feature, just like controllers, if you so desire.
:::

If you don't want your route to be validated, or if validation is not necessary, just pass `nil` as the last parameter.
``` go
router.Route("GET", "/products/{id}", product.Show, nil)
```

## Groups and sub-routers

Grouping routes makes it easier to define multiple routes having the same prefix and/or middleware.

Let's take a simple scenario where we want to implement a user CRUD. All our routes will start with `/user`, so we are going to create a sub-router for it:
``` go
userRouter := router.Subrouter("/user")
```

In our application, user profiles are public: anyone can see the user profiles without being authenticated. However, only authenticated users can modify their information and delete their account. We don't want to add some redundancy and apply the authentication middleware for each route needing it, so we are going to create another sub-router.
```go
userRouter.Route("GET", "/{username}", user.Show, nil)
userRouter.Route("POST", "", user.Register, userrequest.Register)

authUserRouter := userRouter.Subrouter("") // Don't add a prefix
authUserRouter.Middleware(authenticationMiddleware)
authUserRouter.Route("PUT", "/{id}", user.Update, userrequest.Update)
authUserRouter.Route("DELETE", "/{id}", user.Delete, nil)
```

To improve your routes definition readability, you should create a new route registrer for each feature. In our example, our definitions would look like this:
``` go
func registerUserRoutes(router *goyave.Router) {
    //...
}

// Register is the main route registrer.
func Register(router *goyave.Router) {
    registerUserRoutes(router)
    registerProductsRoutes(router)
    //...
}
```

## Middleware

Middleware are handlers executed before the controller handler. Learn more in the dedicated [section](./middleware.html).

Middleware are applied to a router or a sub-router **before the routes definition**. Therefore, all routes in that router and its sub-routers will execute them before executing their associated handler.

To assign a middleware to a router, use the `router.Middleware()` function. Many middleware can be assigned at once. The assignment order is important as middleware will be **executed in order**.

#### Router.Middleware

Middleware apply one or more middleware to the route group.

| Parameters                 | Return |
|----------------------------|--------|
| `middleware ...Middleware` | `void` |

**Example:**
``` go
router.Middleware(middleware.DisallowNonValidatedFields)
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

| Parameters         | Return |
|--------------------|--------|
| `uri string`       | `void` |
| `directory string` |        |
| `download bool`    |        |

**Example:**
``` go
router.Static("/public", "/path/to/static/dir", false)
```

## Native handlers

<p><Badge text="Since v2.0.0"/></p>

#### goyave.NativeHandler

NativeHandler is an adapter function for `http.Handler`. With this adapter, you can plug non-Goyave handlers to your application.

If the request is a JSON request, the native handler will not be able to read the body, as it has already been parsed by the framework and is stored in the `goyave.Request` object. However, form data can be accessed as usual. Just remember that it contains the raw data, which haven't been validated nor converted. This means that **native handlers are not guaranteed to work**.

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
router.Route("GET", "/user", goyave.NativeHandler(httpHandler), nil)
```

#### goyave.NativeMiddleware

::: warning
**Deprecated**: Goyave doesn't use gorilla/mux anymore. This function will be removed in a future major release.
::: 

NativeMiddleware is an adapter function `mux.MiddlewareFunc`. With this adapter, you can plug [Gorilla Mux middleware](https://github.com/gorilla/mux#middleware) to your application.

Native middleware work like native handlers. See [`NativeHandler`](#goyave-nativehandler) for more details.

| Parameters                      | Return              |
|---------------------------------|---------------------|
| `middleware mux.MiddlewareFunc` | `goyave.Middelware` |

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


