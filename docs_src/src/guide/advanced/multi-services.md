# Multi-services

Sometimes you need to run several services in the same executable. For example if you are hosting a websocket server on top of your web API. Goyave can be run in a goroutine and stopped on-demand.

All functions are features below require the `goyave` package to be imported.

``` go
import "github.com/System-Glitch/goyave"
```

## Startup hooks

Startup hooks are function executed in a goroutine after the server finished initializing. This is especially useful when you want to start other services or execute specific commands while being sure the server is up and running, ready to respond to incoming requests. Startup hooks must be registered **before** the `goyave.Start()` call.

#### goyave.RegisterStartupHook

Register a startup hook to execute some code once the server is ready and running.

| Parameters    | Return                 |
| ------------- | ---------------------- |
| `hook func()` | `void`                 |

**Example:**
``` go
goyave.RegsiterStartupHook(func() {
    fmt.Println("Server ready.")
})
```

#### goyave.ClearStartupHooks

Clear all registered startup hooks. Useful when you are writing tests or developing a service able to restart your server multiple times.

| Parameters | Return |
|------------|--------|
|            | `void` |

**Example:**
``` go
goyave.ClearStartupHooks()
```

## Start the server

#### goyave.Start

Starts the server. This functions needs a route registrer function as a parameter. Learn more in the [routing](../basics/routing) section.  
The configuration is not reloaded if you call `Start` multiple times. Only the first startup loads the config files and they are kept in memory until the program exits.
This operation is **blocking**. 

| Parameters                            | Return |
|---------------------------------------|--------|
| `routeRegistrer func(*goyave.Router)` | `void` |

**Examples:**
``` go
goyave.Start(routes.Register)
```

**Running the server in the background:**

You can start the server in a goroutine. However, if you do this and the main goroutine terminates, the server will not shutdown gracefully and the program will exit right away. Be sure to call `goyave.Stop()` to stop the server gracefully before exiting. Learn more in the next section.
``` go
go goyave.Start(routes.Register)
//...
goyave.Stop()
```

## Stop the server

When the running process receives a `SIGINT` or a `SIGTERM` signal, for example when you press `CTRL+C` to interrupt the program, the server will shutdown gracefully, so you don't have to handle that yourself.

However, if you start the server in a goroutine, you have the responsability to shutdown properly. If you exit the program manually or if the main goroutine terminates, ensure that `goyave.Stop()` is called. If the program exits because of an interruption signal, the server will shutdown gracefully.

#### goyave.Stop

Stop the server gracefully without interrupting any active connections. Make sure the program doesn't exit and waits instead for Stop to return.

Stop does not attempt to close nor wait for hijacked connections such as WebSockets. The caller of `Stop` should separately notify such long-lived connections of shutdown and wait for them to close, if desired.

| Parameters | Return |
|------------|--------|
|            | `void` |

**Examples:**
``` go
goyave.Stop()
```

## Server status

The `goyave.IsReady()` function lets you know if the server is running or not.

This function should not be used to wait for the server to be ready. Use a startup hook instead.

#### goyave.IsReady

Returns true if the server is ready to receive and serve incoming requests.

| Parameters | Return |
|------------|--------|
|            | `bool` |

**Example:**
``` go
if goyave.IsReady() {
    fmt.Println("Server is ready")
}
```
