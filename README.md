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

## Learning Goyave

The Goyave framework has an extensive documentation covering in-depth subjects and teaching you how to run a project using Goyave from setup to deployment.

<a href="https://system-glitch.github.io/goyave/guide/installation"><h3 align="center">Read the documentation</h3></a>

<a href="https://godoc.org/github.com/System-Glitch/goyave"><h3 align="center">GoDoc</h3></a>

## Getting Started

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
|--------------------------------------|--------|
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

### Validation

### Database

### Localization

### Testing

### CORS

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
