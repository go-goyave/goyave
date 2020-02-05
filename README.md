<p align="center">
    <img src="resources/img/logo/goyave_text.png" alt="Goyave Logo" width="550"/>
</p>

<p align="center">
    <a href="https://github.com/System-Glitch/goyave/actions"><img src="https://github.com/System-Glitch/goyave/workflows/Test/badge.svg" alt="Build Status"/></a>
    <a href="https://github.com/System-Glitch/goyave/releases"><img src="https://img.shields.io/github/v/release/System-Glitch/goyave?include_prereleases" alt="Version"/></a>
    <a href="https://goreportcard.com/report/github.com/System-Glitch/goyave"><img src="https://goreportcard.com/badge/github.com/System-Glitch/goyave" alt="Go Report"/></a>
    <a href="https://coveralls.io/github/System-Glitch/goyave?branch=master"><img src="https://coveralls.io/repos/github/System-Glitch/goyave/badge.svg" alt="Coverage Status"/></a>
    <a href="https://github.com/System-Glitch/goyave/blob/master/LICENSE"><img src="https://img.shields.io/dub/l/vibe-d.svg" alt="License"/></a>
</p>

<h2 align="center">An Elegant Golang Web Framework</h2>

Goyave is a progressive and accessible web application framework, aimed at making development easy and enjoyable. It has a philosophy of cleanliness and conciseness to make programs more elegant, easier to maintain and more focused.

<table>
    <tr>
        <td valign="top">
            <h3>Clean Code</h3>
            <p>Goyave has an expressive, elegant syntax, a robust structure and conventions. Minimalist calls and reduced redundancy are among the Goyave's core principles.</p>
        </td>
        <td valign="top">
            <h3>Fast Development</h3>
            <p>Develop faster and concentrate on the business logic of your application thanks to the many helpers and built-in functions.</p>
        </td>
        <td valign="top">
            <h3>Powerful functionalities</h3>
            <p>Goyave is accessible, yet powerful. The framework includes routing, request parsing, validation, localization, testing, authentication, and more!</p>
        </td>
    </tr>
</table>

Most golang frameworks for web development don't have a strong directory structure nor conventions to make applications have a uniform architecture and limit redundancy. This makes it difficult to work with them on different projects. In companies, having a well-defined and documented architecture helps new developers integrate projects faster, and reduces the time needed for maintaining them. For open source projects, it helps newcomers understanding the project and makes it easier to contribute.

## Table of contents

- [Leaning Goyave](#learning-goyave)
- [Getting started](#getting-started)
- [Features tour](#features-tour)
- [Contributing](#contributing)
- [License](#license)

## Learning Goyave

The Goyave framework has an extensive documentation covering in-depth subjects and teaching you how to run a project using Goyave from setup to deployment.

<a href="https://system-glitch.github.io/goyave/guide/installation"><h3 align="center">Read the documentation</h3></a>

<a href="https://godoc.org/github.com/System-Glitch/goyave"><h3 align="center">GoDoc</h3></a>

## Getting started

### Requirements

- Go 1.13+
- Go modules

### Install using the template project

You can bootstrap your project using the [Goyave template project](https://github.com/System-Glitch/goyave-template). This project has a complete directory structure already set up for you.

#### Linux / MacOS

```
$ curl https://raw.githubusercontent.com/System-Glitch/goyave/master/install.sh | bash -s my-project
```

#### Windows (Powershell)

```
> & ([scriptblock]::Create((curl "https://raw.githubusercontent.com/System-Glitch/goyave/master/install.ps1").Content)) -projectName my-project
```

---

Run `go run my-project` in your project's directory to start the server, then try to request the `hello` route.
```
$ curl http://localhost:8080/hello
Hi!
```

There is also an `echo` route, with basic validation of query parameters.
```
$ curl http://localhost:8080/echo?text=abc%20123
abc 123
```

## Features tour

This section's goal is to give a **brief** look at the main features of the framework. Don't consider this documentation. If you want a complete reference and documentation, head to [GoDoc](https://godoc.org/github.com/System-Glitch/goyave) and the [official documentation](https://system-glitch.github.io/goyave/guide/).

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

### Hello world from scratch

The example below shows a basic `Hello world` application using Goyave.

``` go
import "github.com/System-Glitch/goyave/v2"

func registerRoutes(router *goyave.Router) {
    router.Route("GET", "/hello", func(response *goyave.Response, request *goyave.Request) {
        response.String(http.StatusOK, "Hello world!")
    }, nil)
}

func main() {
    goyave.Start(registerRoutes)
}
```

### Configuration

To configure your application, use the `config.json` file at your project's root. If you are using the template project, copy `config.example.json` to `config.json`. The following code is an example of configuration for a local development environment:

```json
{
    "appName": "goyave_template",
    "environment": "localhost",
    "host": "127.0.0.1",
    "port": 8080,
    "httpsPort": 8081,
    "protocol": "http",
    "debug": true,
    "timeout": 10,
    "maxUploadSize": 10,
    "defaultLanguage": "en-US",
    "dbConnection": "mysql",
    "dbHost": "127.0.0.1",
    "dbPort": 3306,
    "dbName": "goyave",
    "dbUsername": "root",
    "dbPassword": "root",
    "dbOptions": "charset=utf8&parseTime=true&loc=Local",
    "dbMaxOpenConnections": 20,
    "dbMaxIdleConnections": 20,
    "dbMaxLifetime": 300,
    "dbAutoMigrate": false
}
```

If this config file misses some config entries, the default values will be used. All values from the framework's core are **validated**. That means that the application will not start if you provided an invalid value in your config (For example if the specified port is not a number).

**Getting a value:**
```go
config.GetString("appName") // "goyave"
config.GetBool("debug") // true
config.Has("appName") // true
```

**Setting a value:**
```go
config.Set("appName", "my awesome app")
```

**Learn more about configuration in the [documentation](https://system-glitch.github.io/goyave/guide/configuration.html).**

### Routing

Routing is an essential part of any Goyave application. Routes definition is the action of associating a URI, sometimes having parameters, with a handler which will process the request and respond to it. Separating and naming routes clearly is important to make your API or website clear and expressive.

Routes are defined in **routes registrer functions**. The main route registrer is passed to `goyave.Start()` and is executed automatically with a newly created root-level **router**.

``` go
func Register(router *goyave.Router) {
    // Register your routes here

    // With closure, not recommended
    router.Route("GET", "/hello", func(response *goyave.Response, r *goyave.Request) {
        response.String(http.StatusOK, "Hi!")
    }, nil)

    router.Route("GET", "/hello", myHandlerFunction, nil)
    router.Route("POST", "/user", user.Register, userrequest.Register)
    router.Route("PUT|PATCH", "/user", user.Update, userrequest.Update)
}
```

**Method signature:**

| Parameters                           | Return |
| ------------------------------------ | ------ |
| `methods string`                     | `void` |
| `uri string`                         |        |
| `handler goyave.Handler`             |        |
| `validationRules validation.RuleSet` |        |

URIs can have parameters, defined using the format `{name}` or `{name:pattern}`. If a regular expression pattern is not defined, the matched variable will be anything until the next slash. 

**Example:**
``` go
router.Route("GET", "/products/{key}", product.Show, nil)
router.Route("GET", "/products/{id:[0-9]+}", product.ShowById, nil)
router.Route("GET", "/categories/{category}/{id:[0-9]+}", category.Show, nil)
```

Route parameters can be retrieved as a `map[string]string` in handlers using the request's `Params` attribute.
``` go
func myHandlerFunction(response *goyave.Response, request *goyave.Request) {
    category := request.Params["category"]
    id, _ := strconv.Atoi(request.Params["id"])
    //...
}
```

**Learn more about routing in the [documentation](https://system-glitch.github.io/goyave/guide/basics/routing.html).**

### Controller

Controllers are files containing a collection of Handlers related to a specific feature. Each feature should have its own package. For example, if you have a controller handling user registration, user profiles, etc, you should create a `http/controller/user` package. Creating a package for each feature has the advantage of cleaning up route definitions a lot and helps keeping a clean structure for your project.

A `Handler` is a `func(*goyave.Response, *goyave.Request)`. The first parameter lets you write a response, and the second contains all the information extracted from the raw incoming request.

Handlers receive a `goyave.Response` and a `goyave.Request` as parameters.  
`goyave.Request` can give you a lot of information about the incoming request, such as its headers, cookies, or body. Learn more [here](https://system-glitch.github.io/goyave/guide/basics/requests.html).  
`goyave.Response` implements `http.ResponseWriter` and is used to write a response. If you didn't write anything before the request lifecycle ends, `204 No Content` is automatically written. Learn everything about reponses [here](https://system-glitch.github.io/goyave/guide/basics/responses.html).

Let's take a very simple CRUD as an example for a controller definition:
**http/controllers/product/product.go**:
``` go
func Index(response *goyave.Response, request *goyave.Request) {
    products := []model.Product{}
    result := database.GetConnection().Find(&products)
    if response.HandleDatabaseError(result) {
        response.JSON(http.StatusOK, products)
    }
}

func Show(response *goyave.Response, request *goyave.Request) {
    product := model.Product{}
    id, _ := strconv.ParseUint(request.Params["id"], 10, 64)
    result := database.GetConnection().First(&product, id)
    if response.HandleDatabaseError(result) {
        response.JSON(http.StatusOK, product)
    }
}

func Store(response *goyave.Response, request *goyave.Request) {
    product := model.Product{
        Name:  request.String("name"),
        Price: request.Numeric("price"),
    }
    if err := database.GetConnection().Create(&product).Error; err != nil {
        response.Error(err)
    } else {
        response.JSON(http.StatusCreated, map[string]uint{"id": product.ID})
    }
}

func Update(response *goyave.Response, request *goyave.Request) {
    id, _ := strconv.ParseUint(request.Params["id"], 10, 64)
    product := model.Product{}
    db := database.GetConnection()
    result := db.Select("id").First(&product, id)
    if response.HandleDatabaseError(result) {
        if err := db.Model(&product).Update("name", request.String("name")).Error; err != nil {
            response.Error(err)
        }
    }
}

func Destroy(response *goyave.Response, request *goyave.Request) {
    id, _ := strconv.ParseUint(request.Params["id"], 10, 64)
    product := model.Product{}
    db := database.GetConnection()
    result := db.Select("id").First(&product, id)
    if response.HandleDatabaseError(result) {
        if err := db.Delete(&product).Error; err != nil {
            response.Error(err)
        }
    }
}
```

**Learn more about controllers in the [documentation](https://system-glitch.github.io/goyave/guide/basics/controllers.html).**

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

**Learn more about middleware in the [documentation](https://system-glitch.github.io/goyave/guide/basics/middleware.html).**


### Validation

Goyave provides a powerful, yet easy way to validate all incoming data, no matter its type or its format, thanks to a large number of validation rules.

Incoming requests are validated using **rules set**, which associate rules with each expected field in the request.

Validation rules can **alter the raw data**. That means that when you validate a field to be number, if the validation passes, you are ensured that the data you'll be using in your controller handler is a `float64`. Or if you're validating an IP, you get a `net.IP` object.

Validation is automatic. You just have to define a rules set and assign it to a route. When the validation doesn't pass, the request is stopped and the validation errors messages are sent as a response.

The `http/request` directory contains the requests validation rules sets. You should have one package per feature, regrouping all requests handled by the same controller. The package should be named `<feature_name>request`.

**Example:** (`http/request/productrequest/product.go`)
``` go
var (
    Store validation.RuleSet = validation.RuleSet{
        "name":  {"required", "string", "between:3,50"},
        "price": {"required", "numeric", "min:0.01"},
        "image": {"nullable", "file", "image", "max:2048", "count:1"},
    }
    
    // ...
)
```

Once your rules sets are defined, you need to assign them to your routes. The rule set for a route is the last parameter of the route definition.

``` go
router.Route("POST", "/product", product.Store, productrequest.Store)
```


**Learn more about validation in the [documentation](https://system-glitch.github.io/goyave/guide/basics/validation.html).**

### Database

Most web applications use a database. In this section, we are going to see how Goyave applications can query a database, using the awesome [Gorm ORM](https://gorm.io/).

Database connections are managed by the framework and are long-lived. When the server shuts down, the database connections are closed automatically. So you don't have to worry about creating, closing or refreshing database connections in your application.

Very few code is required to get started with databases. There are some [configuration](https://system-glitch.github.io/goyave/guide/configuration.html#configuration-reference) options that you need to change though:

- `dbConnection`
- `dbHost`
- `dbPort`
- `dbName`
- `dbUsername`
- `dbPassword`
- `dbOptions`
- `dbMaxOpenConnection`
- `dbMaxIdleConnection`
- `dbMaxLifetime`

``` go
user := model.User{}
db := database.GetConnection()
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
    Email        string  `gorm:"type:varchar(100);unique_index"`
    Role         string  `gorm:"size:255"` // set field size to 255
    MemberNumber *string `gorm:"unique;not null"` // set member number to unique and not null
    Num          int     `gorm:"AUTO_INCREMENT"` // set num to auto incrementable
    Address      string  `gorm:"index:addr"` // create index with name `addr` for address
    IgnoreMe     int     `gorm:"-"` // ignore this field
}
```

**Learn more about using databases in the [documentation](https://system-glitch.github.io/goyave/guide/basics/database.html).**

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

The `rules.json` file contains the validation rules messages. These messages can have **[placeholders](https://system-glitch.github.io/goyave/guide/basics/validation.html#placeholders)**, which will be automatically replaced by the validator with dynamic values. If you write custom validation rules, their messages shall be written in this file.

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

**Learn more about localization in the [documentation](https://system-glitch.github.io/goyave/guide/advanced/localization.html).**

### Testing

Goyave provides an API to ease the unit and functional testing of your application. This API is an extension of [testify](https://github.com/stretchr/testify). `goyave.TestSuite` inherits from testify's `suite.Suite`, and sets up the environment for you. That means:

- `GOYAVE_ENV` environment variable is set to `test` and restored to its original value when the suite is done.
- All tests are run using your project's root as working directory. This directory is determined by the presence of a `go.mod` file.
- Config and language files are loaded before the tests start. As the environment is set to `test`, you **need** a `config.test.json` in the root directory of your project.

This setup is done by the function `goyave.RunTest`, so you shouldn't run your test suites using testify's `suite.Run()` function.

The following example is a **functional** test and would be located in the `test` package.

``` go
import (
    "my-project/http/route"
    "github.com/System-Glitch/goyave/v2"
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
    suite.Equal("response content", string(suite.GetBody(resp)))
}
```

**URL-encoded requests:**

``` go
headers := map[string]string{"Content-Type": "application/x-www-form-urlencoded; param=value"}
resp, err := suite.Post("/product", headers, strings.NewReader("field=value"))
suite.Nil(err)
if err == nil {
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
    suite.Equal("response content", string(suite.GetBody(resp)))
}
```

**Testing JSON respones:**

``` go
suite.RunServer(route.Register, func() {
    resp, err := suite.Get("/product", nil)
    suite.Nil(err)
    if err == nil {
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

**Learn more about testing in the [documentation](https://system-glitch.github.io/goyave/guide/advanced/testing.html).**

### Status handlers

Status handlers are regular handlers executed during the finalization step of the request's lifecycle if the response body is empty but a status code has been set. Status handler are mainly used to implement a custom behavior for user or server errors (400 and 500 status codes).

The following file `http/controller/status/status.go` is an example of custom 404 error handling:
``` go
package status

import "github.com/System-Glitch/goyave/v2"

func NotFound(response *goyave.Response, request *goyave.Request) {
    response.RenderHTML(response.GetStatus(), "errors/404.html", nil)
}
```

Status handlers are registered in the **router**.

``` go
// Use "status.NotFound" for empty responses having status 404 or 405.
router.StatusHandler(status.NotFound, 404)
```

**Learn more about status handlers in the [documentation](https://system-glitch.github.io/goyave/guide/advanced/status-handlers.html).**


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

**Learn more about CORS in the [documentation](https://system-glitch.github.io/goyave/guide/advanced/cors.html).**

## Contributing

Thank you for considering contributing to the Goyave framework! You can find the contribution guide in the [documentation](https://system-glitch.github.io/goyave/guide/contribution-guide.html).

I have many ideas for the future of Goyave. I would be infinitely grateful to whoever want to support me and let me continue working on Goyave and making it better and better.

You can support also me on Patreon:

<a href="https://www.patreon.com/bePatron?u=25997573">
    <img src="https://c5.patreon.com/external/logo/become_a_patron_button@2x.png" width="160">
</a>

### Contributors

A big "Thank you" to the Goyave contributors:

- [Kuinox](https://github.com/Kuinox) (Powershell install script)
- [Alexandre GV.](https://github.com/alexandregv) (Install script MacOS compatibility)

## License

The Goyave framework is MIT Licensed. Copyright © 2019 Jérémy LAMBERT (SystemGlitch) 
