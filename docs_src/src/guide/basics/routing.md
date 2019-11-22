# Routing

[[toc]]

## Introduction

Routing is an essential part of any Goyave application. Routes definition is the action of associating a URI, sometimes having parameters, with a handler which will process the request and respond to it. Separating and naming routes clearly is important to make your API or website clear and expressive.

All features below require the `goyave` package to be imported.

``` go
import "github.com/System-Glitch/goyave"
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

| Parameters                           | Return |
|--------------------------------------|--------|
| `methods string`                     | `void` |
| `uri string`                         |        |
| `handler goyave.Handler`             |        |
| `validationRules validation.RuleSet` |        |

**Examples:**
``` go
router.Route("GET", "/hello", myHandlerFunction, nil)
router.Route("POST", "/user", user.Register, userrequest.Register)
router.Route("PUT|PATCH", "/user", user.Update, userrequest.Update)
```

::: tip
- `goyave.Handler` is an alias for `func(*goyave.Response, *goyave.Request)`.
- Learn more about validation and rules sets [here](./validation).
:::

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

## Validation

You can assign a validation rules set to each route. Learn more in the dedicated [section](./validation). You should always validate incoming requests.

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

Grouping routes makes it easier to define multiple routes having the same prefix and/or middlewares.

Let's take a simple scenario where we want to implement a user CRUD. All our routes will start with `/user`, so we are going to create a sub-router for it:
``` go
userRouter := router.Subrouter("/user")
```

In our application, user profiles are public: anyone can see the user profiles without being authenticated. However, only authenticated users can modify their information and delete their account. We don't want to add some redundancy and apply the authentication middleware for each route needing it, so we are going to create another sub-router.
```go
userRouter.Route("GET", "/{username}", user.Show, nil)
userRouter.Route("POST", "/", user.Register, userrequest.Register)

authUserRouter := userRouter.Subrouter("/") // Don't add a prefix
authUserRouter.Middleware(authenticationMiddleware)
authUserRouter.Route("PUT", "/", user.Update, userrequest.Update)
authUserRouter.Route("DELETE", "/", user.Delete, nil)
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

Middleware are handlers executed before the controller handler. Learn more in the dedicated [section](./middleware).

Middleware are applied to a router or a sub-router **before the routes definition**. Therefore, all routes in that router and its sub-routers will execute them before executing their associated handler.

To assign a middleware to a router, use the `router.Middleware()` function. Many middleware can be assigned at once. The assignment order is important as middleware will be **executed in order**.

#### Router.Middleware

Middleware apply one or more middleware(s) to the route group.

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

Static serve a directory and its subdirectories of static resources.
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
