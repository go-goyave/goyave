# Installation

This guide will walk you through the installation process. The rest of the guide assumes you are using the template project, as it is the recommended option.

## Requirements

- Go 1.13+
- Go modules

## Template project

You can bootstrap your project using the **[Goyave template project](https://github.com/System-Glitch/goyave_template)**.

First, download the template and unzip it:
```
$ curl -LOk https://github.com/System-Glitch/goyave_template/archive/master.zip && unzip master.zip && rm master.zip
```

Rename `goyave_template-master` to the name of our project (`my-project` in our example), then `cd` into it and init a git repository.
```
$ mv goyave_template-master my-project
$ cd my-project
$ git init
```

Copy `config.example.json` into a new file `config.json`. Update the configuration to make it fit your needs. See the [configuration](./configuration) section for more details.

::: tip
It is a good practice to ignore the actual config to prevent it being added to the version control system. Each developer may have different settings for their environment.
:::

Finally, you'll have to replace all references to `goyave_template` with the name of your project. You can use the following command:
```
$ find ./ -type f \( -iname \*.go -o -iname \*.mod \) -exec sed -i "s/goyave_template/my-project/g" {} \;
```

Run `go run my-project` to start the server, then try to request the `hello` route.
```
$ curl http://localhost:8080/hello
Hi!
```

::: warning
The template project setup will be more streamlined in the future to make it easier to setup projects.
:::

## From scratch

::: warning
Installing your project from scratch is not recommended as you will likely not use the same directory structure as the template project. Respecting the standard [directory structure](./architecture-concepts#directory-structure) is important and helps keeping a consistent environment across the Goyave applications.
:::

If you prefer to setup your project from scratch, for example if you don't plan on using some of the framework's features or if you want to use a different directory structure, you can!

In a terminal, run:
```
$ mkdir myproject && cd myproject
$ go mod init my-project
$ go get -u github.com/System-Glitch/goyave
```

Now that your project directory is set up and the dependencies are installed, let's start with the program entry point, `kernel.go`:
``` go
package main

import (
    "my-project/http/routes"
    "github.com/System-Glitch/goyave"
)

func main() {
    goyave.Start(routes.Register)
}
```

::: tip
`goyave.Start()` is blocking. You can run it in a goroutine if you want to process other things in the background. See the [multi-services](./advanced/multi-services) section for more details.
:::

Now we need to create the package in which we will register our routes. Create a new package `http/routes`:
```
$ mkdir http
$ mkdir http/routes
```

Create `http/routes/routes.go`:
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
Your routes definitions should be separated from the handler functions. Handlers should be defined in a `http/controllers` directory.
:::

Run your server and request your route:
```
$ go run main

# In another terminal:
$ curl http://localhost:8080/hello
Hi!
```

You should also create a config file for your application. Learn more [here](./configuration).
