# Architecture Concepts

## Introduction

Understanding your development tools and knowing what happens in the background is crucial. Mastering your tools and environment incredibily decreases the risk of errors, the ease of debugging and helps making your code work in harmony with the framework. The goal of this section is to give you an overview of the general functioning and design of the framework, to make you more comfortable and confident using it.

## Lifecycle

### Server

The very first step of the server lifecycle is the **server setup**, taking place when you call `goyave.Start(route.Register)` in your application's main function.

Goyave starts by loading the [configuration](./configuration) file from the core of the framework. The application's configuration file is then loaded, overriding the default values.

The second step of the initialization takes a very similar approach to load the [language](./advanced/localization) files. The `en-US` language is available by default inside the framework and is used as the default language. When it's loaded, the framework will look for custom language files inside the working directory and will override the `en-US` language entries if needed.

Then, if enabled, the automatic migrations are run, thus creating the [database](./basics/database) connection pool. If the automatic migrations are not enabled, no connection to the database will be established until the application requires one.

That is only now that [routes](./basics/routing) are registered using the route registrer provided to the `Start()` function. That means that at this registrer has already access to all the configuration and language features, which can be handy if you want to generate different routes based on the languages your application supports.

Finally, the framework starts listening for incoming HTTP requests and serves them. The server also listens for interruption signals so it can finish serving ongoing requests before shutting down gracefully. In the next section, we will get into more details about the lifecycle of each request.

### Requests

When an incoming request is received, it's first passed through the Gorilla Mux router.

## Directory structure

## Database
