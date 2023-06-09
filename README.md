<p align="center">
    <img src="https://raw.githubusercontent.com/go-goyave/goyave/master/resources/img/logo/goyave_text.png#gh-light-mode-only" alt="Goyave Logo" width="550"/>
    <img src="https://raw.githubusercontent.com/go-goyave/goyave/master/resources/img/logo/goyave_text_dark.png#gh-dark-mode-only" alt="Goyave Logo" width="550"/>
</p>

<p align="center">
    <a href="https://github.com/go-goyave/goyave/releases"><img src="https://img.shields.io/github/v/release/go-goyave/goyave?include_prereleases" alt="Version"/></a>
    <a href="https://github.com/go-goyave/goyave/actions"><img src="https://github.com/go-goyave/goyave/workflows/CI/badge.svg?branch=develop" alt="Build Status"/></a>
    <a href="https://coveralls.io/github/go-goyave/goyave?branch=master"><img src="https://coveralls.io/repos/github/go-goyave/goyave/badge.svg" alt="Coverage Status"/></a>
    <a href="https://goreportcard.com/report/goyave.dev/goyave"><img src="https://goreportcard.com/badge/goyave.dev/goyave" alt="Go Report"/></a>
</p>

<p align="center">
    <a href="https://github.com/go-goyave/goyave/blob/master/LICENSE"><img src="https://img.shields.io/dub/l/vibe-d.svg" alt="License"/></a>
    <a href="https://github.com/avelino/awesome-go"><img src="https://awesome.re/mentioned-badge.svg" alt="Awesome"/></a>
    <a href="https://discord.gg/mfemDMc"><img src="https://img.shields.io/discord/744264895209537617?logo=discord" alt="Discord"/></a>
</p>

<h2 align="center">An Elegant Golang Web Framework</h2>

Goyave is a progressive and accessible web application framework focused on REST APIs, aimed at making backend development easy and enjoyable. It has a philosophy of cleanliness and conciseness to make programs more elegant, easier to maintain and more focused. Goyave is an **opinionated** framework, helping your applications keeping a uniform architecture and limit redundancy. With Goyave, expect a full package with minimum setup.

- **Clean Code**: Goyave has an expressive, elegant syntax, a robust structure and conventions. Minimalist calls and reduced redundancy are among the Goyave's core principles.
- **Fast Development**: Develop faster and concentrate on the business logic of your application thanks to the many helpers and built-in functions.
- **Powerful functionalities**: Goyave is accessible, yet powerful. The framework includes routing, request parsing, validation, localization, testing, authentication, and more!
- **Reliability**: Error reporting is made easy thanks to advanced error handling and panic recovery. The framework is deeply tested.

## Table of contents

- [Learning Goyave](#learning-goyave)
- [Getting started](#getting-started)
- [Features tour](#features-tour)
- [Contributing](#contributing)
- [License](#license)

## Learning Goyave

The Goyave framework has an extensive documentation covering in-depth subjects and teaching you how to run a project using Goyave from setup to deployment.

- [Read the documentation](https://goyave.dev/guide/installation.html)
- [pkg.go.dev](https://pkg.go.dev/goyave.dev/goyave/v4)
- [Example project](https://github.com/go-goyave/goyave-blog-example)

## Getting started

### Requirements

- Go 1.18+

### Install using the template project

You can bootstrap your project using the [Goyave template project](https://github.com/go-goyave/goyave-template). This project has a complete directory structure already set up for you.

#### Linux / MacOS

```
$ curl https://goyave.dev/install.sh | bash -s github.com/username/projectname
```

#### Windows (Powershell)

```
> & ([scriptblock]::Create((curl "https://goyave.dev/install.ps1").Content)) -moduleName github.com/username/projectname
```

---

Run `go run .` in your project's directory to start the server, then try to request the `hello` route.
```
$ curl http://localhost:8080/hello
Hi!
```

There is also an `echo` route, with basic validation of the request body.
```
$ curl -H "Content-Type: application/json" -X POST -d '{"text":"abc 123"}' http://localhost:8080/echo
abc 123
```

## Features tour

This section's goal is to give a **brief** look at the main features of the framework. It doesn't describe everything the framework has to offer, so don't consider this documentation. If you want a complete reference and documentation, head to [pkg.go.dev](https://pkg.go.dev/goyave.dev/goyave/v4) and the [official documentation](https://goyave.dev/guide/).

- [Hello world from scratch](#hello-world-from-scratch)
- [Configuration](#configuration)
- [Routing](#routing)
- [Controller](#controller)
- [Middleware](#middleware)
- [Validation](#validation)
- [Database](#database)
- [Localization](#localization)
- [Testing](#testing)
- [Status handlers](#status-handlers)
- [CORS](#cors)
- [Authentication](#authentication)
- [Rate limiting](#rate-limiting)
- [Websocket](#websocket)

### Hello world from scratch

The example below shows a basic `Hello world` application using Goyave.

``` go
import "goyave.dev/goyave/v4"

func registerRoutes(router *goyave.Router) {
    router.Get("/hello", func(response *goyave.Response, request *goyave.Request) {
        response.String(http.StatusOK, "Hello world!")
    })
}

func main() {
    if err := goyave.Start(registerRoutes); err != nil {
        os.Exit(err.(*goyave.Error).ExitCode)
    }
}
```

### Configuration

To configure your application, use the `config.json` file at your project's root. If you are using the template project, copy `config.example.json` to `config.json`. The following code is an example of configuration for a local development environment:

```json
{
    "app": {
        "name": "goyave_template",
        "environment": "localhost",
        "debug": true,
        "defaultLanguage": "en-US"
    },
    "server": {
        "host": "127.0.0.1",
        "maintenance": false,
        "protocol": "http",
        "domain": "",
        "port": 8080,
        "httpsPort": 8081,
        "timeout": 10,
        "maxUploadSize": 10
    },
    "database": {
        "connection": "mysql",
        "host": "127.0.0.1",
        "port": 3306,
        "name": "goyave",
        "username": "root",
        "password": "root",
        "options": "charset=utf8mb4&collation=utf8mb4_general_ci&parseTime=true&loc=Local",
        "maxOpenConnections": 20,
        "maxIdleConnections": 20,
        "maxLifetime": 300,
        "autoMigrate": false
    }
}
```

If this config file misses some config entries, the default values will be used. 

All entries are **validated**. That means that the application will not start if you provided an invalid value in your config (for example if the specified port is not a number). That also means that a goroutine trying to change a config entry with the incorrect type will panic.  
Entries can be registered with a default value, their type and authorized values from any package. 

**Getting a value:**
```go
config.GetString("app.name") // "goyave"
config.GetBool("app.debug") // true
config.GetInt("server.port") // 80
config.Has("app.name") // true
```

**Setting a value:**
```go
config.Set("app.name", "my awesome app")
```

**Using environment variables:**
```json
{
  "database": {
    "host": "${DB_HOST}"
  }
}
```

**Learn more about configuration in the [documentation](https://goyave.dev/guide/configuration.html).**

### Routing

Routing is an essential part of any Goyave application. Routes definition is the action of associating a URI, sometimes having parameters, with a handler which will process the request and respond to it. Separating and naming routes clearly is important to make your API or website clear and expressive.

Routes are defined in **routes registrer functions**. The main route registrer is passed to `goyave.Start()` and is executed automatically with a newly created root-level **router**.

``` go
func Register(router *goyave.Router) {
    // Register your routes here

    // With closure, not recommended
    router.Get("/hello", func(response *goyave.Response, r *goyave.Request) {
        response.String(http.StatusOK, "Hi!")
    })

    router.Get("/hello", myHandlerFunction)
    router.Post("/user", user.Register).Validate(user.RegisterRequest)
    router.Route("PUT|PATCH", "/user", user.Update).Validate(user.UpdateRequest)
    router.Route("POST", "/product", product.Store).Validate(product.StoreRequest).Middleware(middleware.Trim)
}
```

URIs can have parameters, defined using the format `{name}` or `{name:pattern}`. If a regular expression pattern is not defined, the matched variable will be anything until the next slash. 

**Example:**
``` go
router.Get("/product/{key}", product.Show)
router.Get("/product/{id:[0-9]+}", product.ShowById)
router.Get("/category/{category}/{id:[0-9]+}", category.Show)
```

Route parameters can be retrieved as a `map[string]string` in handlers using the request's `Params` attribute.
``` go
func myHandlerFunction(response *goyave.Response, request *goyave.Request) {
    category := request.Params["category"]
    id, _ := strconv.Atoi(request.Params["id"])
    //...
}
```

**Learn more about routing in the [documentation](https://goyave.dev/guide/basics/routing.html).**

### Controller

Controllers are files containing a collection of Handlers related to a specific feature. Each feature should have its own package. For example, if you have a controller handling user registration, user profiles, etc, you should create a `http/controller/user` package. Creating a package for each feature has the advantage of cleaning up route definitions a lot and helps keeping a clean structure for your project.

A `Handler` is a `func(*goyave.Response, *goyave.Request)`. The first parameter lets you write a response, and the second contains all the information extracted from the raw incoming request.

Handlers receive a `goyave.Response` and a `goyave.Request` as parameters.  
`goyave.Request` can give you a lot of information about the incoming request, such as its headers, cookies, or body. Learn more [here](https://goyave.dev/guide/basics/requests.html).  
`goyave.Response` implements `http.ResponseWriter` and is used to write a response. If you didn't write anything before the request lifecycle ends, `204 No Content` is automatically written. Learn everything about reponses [here](https://goyave.dev/guide/basics/responses.html).

Let's take a very simple CRUD as an example for a controller definition:
**http/controller/product/product.go**:
``` go
func Index(response *goyave.Response, request *goyave.Request) {
    products := []model.Product{}
    result := database.Conn().Find(&products)
    if response.HandleDatabaseError(result) {
        response.JSON(http.StatusOK, products)
    }
}

func Show(response *goyave.Response, request *goyave.Request) {
    product := model.Product{}
    result := database.Conn().First(&product, request.Params["id"])
    if response.HandleDatabaseError(result) {
        response.JSON(http.StatusOK, product)
    }
}

func Store(response *goyave.Response, request *goyave.Request) {
    product := model.Product{
        Name:  request.String("name"),
        Price: request.Numeric("price"),
    }
    if err := database.Conn().Create(&product).Error; err != nil {
        response.Error(err)
    } else {
        response.JSON(http.StatusCreated, map[string]uint{"id": product.ID})
    }
}

func Update(response *goyave.Response, request *goyave.Request) {
    product := model.Product{}
    db := database.Conn()
    result := db.Select("id").First(&product, request.Params["id"])
    if response.HandleDatabaseError(result) {
        if err := db.Model(&product).Update("name", request.String("name")).Error; err != nil {
            response.Error(err)
        }
    }
}

func Destroy(response *goyave.Response, request *goyave.Request) {
    product := model.Product{}
    db := database.Conn()
    result := db.Select("id").First(&product, request.Params["id"])
    if response.HandleDatabaseError(result) {
        if err := db.Delete(&product).Error; err != nil {
            response.Error(err)
        }
    }
}
```

**Learn more about controllers in the [documentation](https://goyave.dev/guide/basics/controllers.html).**

### Middleware

Middleware are handlers executed before the controller handler. They are a convenient way to filter, intercept or alter HTTP requests entering your application. For example, middleware can be used to authenticate users. If the user is not authenticated, a message is sent to the user even before the controller handler is reached. However, if the user is authenticated, the middleware will pass to the next handler. Middleware can also be used to sanitize user inputs, by trimming strings for example, to log all requests into a log file, to automatically add headers to all your responses, etc.

``` go
func MyCustomMiddleware(next goyave.Handler) goyave.Handler {
    return func(response *goyave.Response, request *goyave.Request) {
        // Do something
        next(response, request) // Pass to the next handler
    }
}
```

To assign a middleware to a router, use the `router.Middleware()` function. Many middleware can be assigned at once. The assignment order is important as middleware will be **executed in order**.

``` go
router.Middleware(middleware.MyCustomMiddleware)
```

**Learn more about middleware in the [documentation](https://goyave.dev/guide/basics/middleware.html).**


### Validation

Goyave provides a powerful, yet easy way to validate all incoming data, no matter its type or its format, thanks to a large number of validation rules.

Incoming requests are validated using **rules set**, which associate rules with each expected field in the request.

Validation rules can **alter the raw data**. That means that when you validate a field to be number, if the validation passes, you are ensured that the data you'll be using in your controller handler is a `float64`. Or if you're validating an IP, you get a `net.IP` object.

Validation is automatic. You just have to define a rules set and assign it to a route. When the validation doesn't pass, the request is stopped and the validation errors messages are sent as a response.

Rule sets are defined in the same package as the controller, typically in a separate file named `request.go`. Rule sets are named after the name of the controller handler they will be used with, and end with `Request`. For example, a rule set for the `Store` handler will be named `StoreRequest`. If a rule set can be used for multiple handlers, consider using a name suited for all of them. The rules for a store operation are often the same for update operations, so instead of duplicating the set, create one unique set called `UpsertRequest`.

**Example:** (`http/controller/product/request.go`)
``` go
var (
    StoreRequest = validation.RuleSet{
        "name":  validation.List{"required", "string", "between:3,50"},
        "price": validation.List{"required", "numeric", "min:0.01"},
        "image": validation.List{"nullable", "file", "image", "max:2048", "count:1"},
    }

    // ...
)
```

Once your rules sets are defined, you need to assign them to your routes using the `Validate()` method.

``` go
router.Post("/product", product.Store).Validate(product.StoreRequest)
```


**Learn more about validation in the [documentation](https://goyave.dev/guide/basics/validation.html).**

### Database

Most web applications use a database. In this section, we are going to see how Goyave applications can query a database, using the awesome [Gorm ORM](https://gorm.io/).

Database connections are managed by the framework and are long-lived. When the server shuts down, the database connections are closed automatically. So you don't have to worry about creating, closing or refreshing database connections in your application.

Very few code is required to get started with databases. There are some [configuration](https://goyave.dev/guide/configuration.html#configuration-reference) options that you need to change though:

- `database.connection`
- `database.host`
- `database.port`
- `database.name`
- `database.username`
- `database.password`
- `database.options`
- `database.maxOpenConnection`
- `database.maxIdleConnection`
- `database.maxLifetime`

``` go
user := model.User{}
db := database.Conn()
db.First(&user)

fmt.Println(user)
```

Models are usually just normal Golang structs, basic Go types, or pointers of them. `sql.Scanner` and `driver.Valuer` interfaces are also supported.

```go
func init() {
    database.RegisterModel(&User{})
}

type User struct {
    gorm.Model
    Name         string
    Age          sql.NullInt64
    Birthday     *time.Time
    Email        string  `gorm:"type:varchar(100);uniqueIndex"`
    Role         string  `gorm:"size:255"` // set field size to 255
    MemberNumber *string `gorm:"unique;not null"` // set member number to unique and not null
    Num          int     `gorm:"autoIncrement"` // set num to auto incrementable
    Address      string  `gorm:"index:addr"` // create index with name `addr` for address
    IgnoreMe     int     `gorm:"-"` // ignore this field
}
```

**Learn more about using databases in the [documentation](https://goyave.dev/guide/basics/database.html).**

### Localization

The Goyave framework provides a convenient way to support multiple languages within your application. Out of the box, Goyave only provides the `en-US` language.

Language files are stored in the `resources/lang` directory.

```
.
└── resources
    └── lang
        └── en-US (language name)
            ├── fields.json (optional)
            ├── locale.json (optional)
            └── rules.json (optional)
```

The `fields.json` file contains the field names translations and their rule-specific messages. Translating field names helps making more expressive messages instead of showing the technical field name to the user. Rule-specific messages let you override a validation rule message for a specific field.

**Example:**
``` json
{
    "email": {
        "name": "email address",
        "rules": {
            "required": "You must provide an :field."
        }
    }
}
```

The `locale.json` file contains all language lines that are not related to validation. This is the place where you should write the language lines for your user interface or for the messages returned by your controllers.

**Example:**
``` json
{
    "product.created": "The product have been created with success.",
    "product.deleted": "The product have been deleted with success."
}
```

The `rules.json` file contains the validation rules messages. These messages can have **[placeholders](https://goyave.dev/guide/basics/validation.html#placeholders)**, which will be automatically replaced by the validator with dynamic values. If you write custom validation rules, their messages shall be written in this file.

**Example:**

``` json
{
    "integer": "The :field must be an integer.",
    "starts_with": "The :field must start with one of the following values: :values.",
    "same": "The :field and the :other must match."
}
```

When an incoming request enters your application, the core language middleware checks if the `Accept-Language` header is set, and set the `goyave.Request`'s `Lang` attribute accordingly. Localization is handled automatically by the validator.

``` go
func ControllerHandler(response *goyave.Response, request *goyave.Request) {
    response.String(http.StatusOK, lang.Get(request.Lang, "my-custom-message"))
}
```

**Learn more about localization in the [documentation](https://goyave.dev/guide/advanced/localization.html).**

### Testing

Goyave provides an API to ease the unit and functional testing of your application. This API is an extension of [testify](https://github.com/stretchr/testify). `goyave.TestSuite` inherits from testify's `suite.Suite`, and sets up the environment for you. That means:

- `GOYAVE_ENV` environment variable is set to `test` and restored to its original value when the suite is done.
- All tests are run using your project's root as working directory. This directory is determined by the presence of a `go.mod` file.
- Config and language files are loaded before the tests start. As the environment is set to `test`, you **need** a `config.test.json` in the root directory of your project.

This setup is done by the function `goyave.RunTest`, so you shouldn't run your test suites using testify's `suite.Run()` function.

The following example is a **functional** test and would be located in the `test` package.

``` go
import (
    "github.com/username/projectname/http/route"
    "goyave.dev/goyave/v4"
)

type CustomTestSuite struct {
    goyave.TestSuite
}

func (suite *CustomTestSuite) TestHello() {
    suite.RunServer(route.Register, func() {
        resp, err := suite.Get("/hello", nil)
        suite.Nil(err)
        suite.NotNil(resp)
        if resp != nil {
            defer resp.Body.Close()
            suite.Equal(200, resp.StatusCode)
            suite.Equal("Hi!", string(suite.GetBody(resp)))
        }
    })
}

func TestCustomSuite(t *testing.T) {
    goyave.RunTest(t, new(CustomTestSuite))
}
```

When writing functional tests, you can retrieve the response body  easily using `suite.GetBody(response)`.

``` go 
resp, err := suite.Get("/get", nil)
suite.Nil(err)
if err == nil {
    defer resp.Body.Close()
    suite.Equal("response content", string(suite.GetBody(resp)))
}
```

**URL-encoded requests:**

``` go
headers := map[string]string{"Content-Type": "application/x-www-form-urlencoded; param=value"}
resp, err := suite.Post("/product", headers, strings.NewReader("field=value"))
suite.Nil(err)
if err == nil {
    defer resp.Body.Close()
    suite.Equal("response content", string(suite.GetBody(resp)))
}
```

**JSON requests:**

``` go
headers := map[string]string{"Content-Type": "application/json"}
body, _ := json.Marshal(map[string]interface{}{"name": "Pizza", "price": 12.5})
resp, err := suite.Post("/product", headers, bytes.NewReader(body))
suite.Nil(err)
if err == nil {
    defer resp.Body.Close()
    suite.Equal("response content", string(suite.GetBody(resp)))
}
```

**Testing JSON response:**

``` go
suite.RunServer(route.Register, func() {
    resp, err := suite.Get("/product", nil)
    suite.Nil(err)
    if err == nil {
        defer resp.Body.Close()
        json := map[string]interface{}{}
        err := suite.GetJSONBody(resp, &json)
        suite.Nil(err)
        if err == nil { // You should always check parsing error before continuing.
            suite.Equal("value", json["field"])
            suite.Equal(float64(42), json["number"])
        }
    }
})
```

The testing API has many more features such as record generators, factories, database helpers, a middleware tester, support for multipart and file uploads...

**Learn more about testing in the [documentation](https://goyave.dev/guide/advanced/testing.html).**

### Status handlers

Status handlers are regular handlers executed during the finalization step of the request's lifecycle if the response body is empty but a status code has been set. Status handler are mainly used to implement a custom behavior for user or server errors (400 and 500 status codes).

The following file `http/controller/status/status.go` is an example of custom 404 error handling:
``` go
package status

import "goyave.dev/goyave/v4"

func NotFound(response *goyave.Response, request *goyave.Request) {
    if err := response.RenderHTML(response.GetStatus(), "errors/404.html", nil); err != nil {
        response.Error(err)
    }
}
```

Status handlers are registered in the **router**.

``` go
// Use "status.NotFound" for empty responses having status 404 or 405.
router.StatusHandler(status.NotFound, 404)
```

**Learn more about status handlers in the [documentation](https://goyave.dev/guide/advanced/status-handlers.html).**


### CORS

Goyave provides a built-in CORS module. CORS options are set on **routers**. If the passed options are not `nil`, the CORS core middleware is automatically added.

``` go
router.CORS(cors.Default())
```

CORS options should be defined **before middleware and route definition**. All of this router's sub-routers **inherit** CORS options by default. If you want to remove the options from a sub-router, or use different ones, simply create another `cors.Options` object and assign it.

`cors.Default()` can be used as a starting point for custom configuration.

``` go
options := cors.Default()
options.AllowedOrigins = []string{"https://google.com", "https://images.google.com"}
router.CORS(options)
```

**Learn more about CORS in the [documentation](https://goyave.dev/guide/advanced/cors.html).**

### Authentication

Goyave provides a convenient and expandable way of handling authentication in your application. Authentication can be enabled when registering your routes:

``` go
import "goyave.dev/goyave/v4/auth"

//...

authenticator := auth.Middleware(&model.User{}, &auth.BasicAuthenticator{})
router.Middleware(authenticator)
```

Authentication is handled by a simple middleware calling an **Authenticator**. This middleware also needs a model, which will be used to fetch user information on a successful login.

Authenticators use their model's struct fields tags to know which field to use for username and password. To make your model compatible with authentication, you must add the `auth:"username"` and `auth:"password"` tags:

``` go
type User struct {
    gorm.Model
    Email    string `gorm:"type:char(100);uniqueIndex" auth:"username"`
    Name     string `gorm:"type:char(100)"`
    Password string `gorm:"type:char(60)" auth:"password"`
}
```

When a user is successfully authenticated on a protected route, its information is available in the controller handler, through the request `User` field.

``` go
func Hello(response *goyave.Response, request *goyave.Request) {
    user := request.User.(*model.User)
    response.String(http.StatusOK, "Hello " + user.Name)
}
```

**Learn more about authentication in the [documentation](https://goyave.dev/guide/advanced/authentication.html).**

### Rate limiting

Rate limiting is a crucial part of public API development. If you want to protect your data from being crawled, protect yourself from DDOS attacks, or provide different tiers of access to your API, you can do it using Goyave's built-in rate limiting middleware.

This middleware uses either a client's IP or an authenticated client's ID (or any other way of identifying a client you may need) and maps a quota, a quota duration and a request count to it. If a client exceeds the request quota in the given quota duration, this middleware will block and return `HTTP 429 Too Many Requests`.

The middleware will always add the following headers to the response:
- `RateLimit-Limit`: containing the requests quota in the time window
- `RateLimit-Remaining`: containing the remaining requests quota in the current window
- `RateLimit-Reset`: containing the time remaining in the current window, specified in seconds

```go
import "goyave.dev/goyave/v4/middleware/ratelimiter"

ratelimiterMiddleware := ratelimiter.New(func(request *goyave.Request) ratelimiter.Config {
    return ratelimiter.Config {
        RequestQuota:  100,
        QuotaDuration: time.Minute,
        // 100 requests per minute allowed
        // Client IP will be used as identifier
    }
})

router.Middleware(ratelimiterMiddleware)
```

**Learn more about rate limiting in the [documentation](https://goyave.dev/guide/advanced/rate-limiting.html).**

### Websocket

Goyave is using [`gorilla/websocket`](https://github.com/gorilla/websocket) and adds a layer of abstraction to it to make it easier to use. You don't have to write the connection upgrading logic nor the close handshake. Just like regular HTTP handlers, websocket handlers benefit from reliable error handling and panic recovery.

Here is an example of an "echo" feature:
```go
upgrader := websocket.Upgrader{}
router.Get("/websocket", upgrader.Handler(func(c *websocket.Conn, request *goyave.Request) error {
    for {
        mt, message, err := c.ReadMessage()
        if err != nil {
            return err
        }
        goyave.Logger.Printf("recv: %s", message)
        err = c.WriteMessage(mt, message)
        if err != nil {
            return fmt.Errorf("write: %w", err)
        }
    }
}))
```

**Learn more about websockets in the [documentation](https://goyave.dev/guide/advanced/websocket.html).**

## Contributing

Thank you for considering contributing to the Goyave framework! You can find the contribution guide in the [documentation](https://goyave.dev/guide/contribution-guide.html).

I have many ideas for the future of Goyave. I would be infinitely grateful to whoever want to support me and let me continue working on Goyave and making it better and better.

You can support me on Github Sponsor.

<a href="https://github.com/sponsors/System-Glitch">❤ Sponsor me!</a>

I'm very grateful to my patrons, sponsors and donators:

- Ben Hyrman
- Massimiliano Bertinetti
- ethicnology
- Yariya
- sebastien-d-me
- Nebojša Koturović

### Contributors

A big "Thank you" to the Goyave contributors:

- [Kuinox](https://github.com/Kuinox) (Powershell install script)
- [Alexandre GV.](https://github.com/alexandregv) (Install script MacOS compatibility)
- [jRimbault](https://github.com/jRimbault) (CI and code analysis)
- [Guillermo Galvan](https://github.com/gmgalvan) (Request extra data)
- [Albert Shirima](https://github.com/agbaraka) (Rate limiting and various contributions)
- [Łukasz Sowa](https://github.com/Morishiri) (Custom claims in JWT)
- [Zao SOULA](https://github.com/zaosoula) (Custom `gorm.Config{}` in config file)
- [Ajtene Kurtaliqi](https://github.com/akurtaliqi) (Documentation landing page)
- [Louis Laurent](https://github.com/ulphidius) ([`gyv`](https://github.com/go-goyave/gyv) productivity CLI)
- [Clement3](https://github.com/Clement3) (`search` feature on [`goyave.dev/filter`](https://github.com/go-goyave/filter))
- [Darkweak](https://github.com/darkweak) (`HTTP cache, RFC compliant` middleware based on [Souin HTTP cache system](https://github.com/darkweak/souin))
- [Jason C Keller](https://github.com/imuni4fun) (Testify interface compatibility)

## Used by

<p align="center">
    <a href="https://adagio.io" target="_blank" rel="nofollow">
        <img src=".github/usedby/adagio.webp" alt="Adagio.io"/>
    </a>
</p>

<p align="center">
    Do you want to be featured here? <a href="https://github.com/go-goyave/goyave/issues/new?template=used_by.md" target="_blank" rel="nofollow">Open an issue</a>.
</p>

## License

The Goyave framework is MIT Licensed. Copyright © 2019 Jérémy LAMBERT (SystemGlitch) 
