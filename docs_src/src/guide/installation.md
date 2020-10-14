---
meta:
  - name: "og:title"
    content: "Installation - Goyave"
  - name: "twitter:title"
    content: "Installation - Goyave"
  - name: "title"
    content: "Installation - Goyave"
---

# Installation

This guide will walk you through the installation process. The rest of the guide assumes you are using the template project, as it is the recommended option.

## Requirements

- Go 1.13+
- Go modules

## Template project

You can bootstrap your project using the **[Goyave template project](https://github.com/System-Glitch/goyave-template)**. This project has a complete directory structure already set up for you.

#### Linux / MacOS

```
$ curl https://raw.githubusercontent.com/System-Glitch/goyave/master/install.sh | bash -s github.com/username/projectname
```

#### Windows (Powershell)

```
> & ([scriptblock]::Create((curl "https://raw.githubusercontent.com/System-Glitch/goyave/master/install.ps1").Content)) -moduleName github.com/username/projectname
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

## From scratch

::: warning
Installing your project from scratch is not recommended as you will likely not use the same directory structure as the template project. Respecting the standard [directory structure](./architecture-concepts.html#directory-structure) is important and helps keeping a consistent environment across the Goyave applications.
:::

If you prefer to setup your project from scratch, for example if you don't plan on using some of the framework's features or if you want to use a different directory structure, you can!

In a terminal, run:
```
$ mkdir myproject && cd myproject
$ go mod init github.com/username/projectname
$ go get -u github.com/System-Glitch/goyave/v3
```

Now that your project directory is set up and the dependencies are installed, let's start with the program entry point, `kernel.go`:
``` go
package main

import (
    "github.com/username/projectname/http/route"
    "github.com/System-Glitch/goyave/v3"
)

func main() {
    if err := goyave.Start(route.Register); err != nil {
      os.Exit(err.(*goyave.Error).ExitCode)
    }
}
```

::: tip
`goyave.Start()` is blocking. You can run it in a goroutine if you want to process other things in the background. See the [multi-services](./advanced/multi-services.html) section for more details.
:::

Now we need to create the package in which we will register our routes. Create a new package `http/route`:
```
$ mkdir http
$ mkdir http/route
```

Create `http/route/route.go`:
``` go
package routes

import "github.com/System-Glitch/goyave/v3"

// Register all the routes
func Register(router *goyave.Router) {
	router.Get("GET", "/hello", hello)
}

// Handler function for the "/hello" route
func hello(response *goyave.Response, request *goyave.Request) {
	response.String(http.StatusOK, "Hi!")
}
```

Here we registered a very simple route displaying "Hi!". Learn more about routing [here](./basics/routing.html).

::: tip
Your routes definitions should be separated from the handler functions. Handlers should be defined in a `http/controller` directory.
:::

Run your server and request your route:
```
$ go run .

# In another terminal:
$ curl http://localhost:8080/hello
Hi!
```

You should also create a config file for your application. Learn more [here](./configuration.html).

It is a good practice to ignore the actual config to prevent it being added to the version control system. Each developer may have different settings for their environment. To do so, add `config.json` to your `.gitignore`.
