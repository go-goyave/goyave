# Architecture Concepts

## Introduction

Understanding your development tools and knowing what happens in the background is crucial. Mastering your tools and environment incredibily decreases the risk of errors, the ease of debugging and helps making your code work in harmony with the framework. The goal of this section is to give you an overview of the general functioning and design of the framework, to make you more comfortable and confident using it.

## Terminology

This section will briefly explain the technical words used in the following sections.

**Lifecycle**: An execution from start to finish, with intermediary steps.

**Framework core**: Features and behaviors executed internally and that are invisible to the application developer.

**Handler**: A function receiving incoming requests and a response writer. Multiple handlers can be executed for the same request.

**Router**: The root-level handler responsible for the execution of the correct controller handler.

**Route**: Also called "endpoint", an URL definition linked to a controller handler. If a request matches the definition, the router will execute the associated controller handler.

**Controller**: A source file implementing the business logic linked to a specific resource and associated routes.

**Middleware**: A handler executed before controller handlers. Middlewares can intercept the request, modify its data, and send a response before the controller handler is reached.

**Application**: A program using the Goyave framework as a library.

## Lifecycle

### Server

The very first step of the server lifecycle is the **server setup**, taking place when you call `goyave.Start(route.Register)` in your application's main function.

Goyave starts by loading the [configuration](./configuration) file from the core of the framework. The application's configuration file is then loaded, overriding the default values.

The second step of the initialization takes a very similar approach to load the [language](./advanced/localization) files. The `en-US` language is available by default inside the framework and is used as the default language. When it's loaded, the framework will look for custom language files inside the working directory and will override the `en-US` language entries if needed.

Then, if enabled, the automatic migrations are run, thus creating the [database](./basics/database) connection pool. If the automatic migrations are not enabled, no connection to the database will be established until the application requires one.

That is only now that [routes](./basics/routing) are registered using the route registrer provided to the `Start()` function. That means that at this registrer has already access to all the configuration and language features, which can be handy if you want to generate different routes based on the languages your application supports.

Finally, the framework starts listening for incoming HTTP requests and serves them. The server also listens for interruption signals so it can finish serving ongoing requests before shutting down gracefully. In the next section, we will get into more details about the lifecycle of each request.

### Requests

When an incoming request is received, it's first passed through the [Gorilla Mux](https://github.com/gorilla/mux) router so your server knows which handler to execute when a user requests a specific URL. Then, the framework's internal handler creates a `goyave.Request` object and a `goyave.Response` object from the raw request. These two objects are fundamental features of the framework as you are going to use them to retrieve the requests' data and write your responses.

Before executing the handler, the middlewares are executed. The framework features a few core middlewares, which are executed **first** and for all routes and all requests.

1. The **recovery** middleware is executed. This middleware ensures that any un-recovered panic is handled. Instead of never returning a response in case of a panic, the server will then return an HTTP 500 Error. If debugging is enabled if the configuration, the response will contain the error message and the stacktrace will be printed in the console. It's important to keep this behavior in mind when handling errors in your handlers.
2. The request is **parsed** by a second middleware. This middleware will automatically detect the request's body format based on the headers and attempt to parse it. If the request can't be parsed, the request's data is simply set to `nil`. This middleware supports JSON requests.
3. The `Accept-Language` header is checked. If it's there, its value is parsed and the request's language attribute is set accordingly so localization is easy in the following handlers. If the header is missing, invalid, or asks for an unsupported language, the framework falls back to the default language defined in the configuration. Learn more [here](./advanced/localization).
4. Application middlewares are executed. These middlewares are implemented and defined by the application developer. Note that some application middlewares are already available in the framework. Learn more in the [middlewares](./basics/middlewares) section. At this moment, the request is not validated yet, so application middlewares can be used for authentication or automatic string trimming for example. Bear in mind that manipulating unvalidated data can be dangerous, especially in form-data where the data types are not converted by the validator yet.
5. The data is validated last. The validation middleware immediately passes if no rules have been defined for the current route, else, it check if the data parsing was successful. An automatic response is sent if that is not the case. The data is passed through the validator, which converts the data types and validates it. The request is stopped if the validation is not successful, and the validation errors are sent as a response. Be careful when working with unvalidated requests (which you should never do!) because if the request's parsing fails, `request.Data` will be `nil`. 
6. Finally, if the request has not been stopped by a middleware, the controller handler is executed.
7. If the controller handler didn't write anything as a response, an empty response with the HTTP status code 204 "No Content" is automatically sent, so you don't have to do it yourself.

## Directory structure

## Database
