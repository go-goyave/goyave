# Installation

This guide will walk you through the installation process. The rest of the guide assumes you are using the template project, as it is the recommended option.

## Requirements

- Go 1.13+
- Go modules

## Template project

::: warning
The template project has not been implemented yet.
:::

## From scratch

::: warning
Installing your project from scratch is not recommended as you will likely not use the same directory structure as the template project. Respecting the standard directory structure is important and helps keeping a consistent environment across the Goyave applications.
:::

If you prefer to setup your project from scratch, for example if you don't plan on using some of the framework's features or if you want to use a different directory structure, you can!

In a terminal, run:
```
$ mkdir myproject && cd myproject
$ go mod init my-project
$ go get -u github.com/System-Glitch/goyave
```

Now that your project directory is set up and the dependencies are installed, let's start with the program entry point, `main.go`:
``` go
package main

import (
    "routes"
    "github.com/System-Glitch/goyave"
)

func main() {
    goyave.Start(routes.Register)
}
```

::: tip
`goyave.Start()` is blocking. You can run it in a goroutine if you want to process other things in the background.
:::

Now we need to create the package in which we will register our routes. Create a new package `routes`:
```
$ mkdir routes
```

Create `routes/routes.go`:
``` go
package routes

import "github.com/System-Glitch/goyave"

// Register all the routes
func Register(router *goyave.Router) {
	router.Route("GET", "/hello", hello, nil)
}

// Handler function for the "/hello" route
func hello(response *goyave.Response, request *goyave.Request) {
	response.String(http.StatusOK, "Hi!")
}
```

Here we registered a very simple route displaying "Hi!". Learn more about routing [here](./basics/routing).

::: tip
Your routes definitions should be separated from the handler functions. Handlers should be defined in a `controllers` directory.
:::

Run your server and request your route:
```
$ go run main

# In another terminal:
$ curl http://localhost:8080/hello
Hi!
```

You should also create a config file for your application. Learn more [here](./configuration).
