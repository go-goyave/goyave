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

**Model**: A structure reflecting a database table structure. An instance of a model is a single database record.

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

#### 1. Recovery

The **recovery** middleware is executed. This middleware ensures that any unrecovered panic is handled. Instead of never returning a response in case of a panic, the server will then return an HTTP 500 Error. If debugging is enabled if the configuration, the response will contain the error message and the stacktrace will be printed in the console. It's important to keep this behavior in mind when handling errors in your handlers.

#### 2. Parsing

The request is **parsed** by a second middleware. This middleware will automatically detect the request's body format based on the headers and attempt to parse it. If the request can't be parsed, the request's data is simply set to `nil`. This middleware supports JSON requests.

#### 3. Language

The `Accept-Language` header is checked. If it's there, its value is parsed and the request's language attribute is set accordingly so localization is easy in the following handlers. If the header is missing, invalid, or asks for an unsupported language, the framework falls back to the default language defined in the configuration. Learn more [here](./advanced/localization).

#### 4. Application middlewares

Application middlewares are executed. These middlewares are implemented and defined by the application developer. Note that some application middlewares are already available in the framework. Learn more in the [middlewares](./basics/middlewares) section. At this moment, the request is not validated yet, so application middlewares can be used for authentication or automatic string trimming for example. Bear in mind that manipulating unvalidated data can be dangerous, especially in form-data where the data types are not converted by the validator yet.

#### 5. Validation

The data is validated last. The validation middleware immediately passes if no rules have been defined for the current route, else, it check if the data parsing was successful. An automatic response is sent if that is not the case. The data is passed through the validator, which converts the data types and validates it. The request is stopped if the validation is not successful, and the validation errors are sent as a response. Be careful when working with unvalidated requests (which you should never do!) because if the request's parsing fails, `request.Data` will be `nil`.

#### 6. Controller handler

Finally, if the request has not been stopped by a middleware, the controller handler is executed.
If the controller handler didn't write anything as a response, an empty response with the HTTP status code 204 "No Content" is automatically sent, so you don't have to do it yourself.

## Directory structure

Goyave follows the principle of "**Convention is better than configuration**". That means that the framework will attempt to automatically get the resources it needs from predefined directories.
The typical and recommended directory structure for Goyave applications is as follows:

:::vue
.
├── database
│   └── models
│       └── *...*
├── http
│   ├── controllers
│   │   └── *...*
│   ├── middlewares
│   │   └── *...*
│   ├── requests
│   │   ├── placeholders.go (*optional*)
│   │   ├── validation.go (*optional*)
│   │   └── *...*
│   └── routes
│       └── routes.go
│
├── resources
│   ├── lang
│   │   └── en-US (*language name*)
│   │       ├── fields.json (*optional*)
│   │       ├── locale.json (*optional*)
│   │       └── rules.json (*optional*)
│   └── img (*optional*)
│       └── *...*
│ 
├── .gitignore
├── config.json
├── go.mod
└── kernel.go
:::

### Database directory

The `database` directory stores the models. Each model should have its own file in the `models` package. This directory can also contain database-related code such as repositories, if you want to use this pattern.

### HTTP directory

The `http` directory contains all the HTTP-related code. This is where most of your code will be written.

#### HTTP controllers

The `http/controllers` directory contains the controller packages. Each feature should have its own package. For example, if you have a controller handling user registration, user profiles, etc, you should create a `http/controllers/user` package. Creating a package for each feature has the advantage of cleaning up route definitions a lot and helps keeping a clean structure for your project. Learn more [here](./basics/controllers).

#### HTTP middlewares

The `http/middlewares` directory contains the application middlewares. Each middleware should have its own file. Learn more [here](./basics/middlewares).
    
#### HTTP requests

The `http/requests` directory contains the requests validation rule sets. You should have one file per feature, regrouping all requests handled by the same controller. You can also create one package per feature, just like controllers, if you so desire.

This directory can also contain a `placeholders.go` file, which will define validation rule messages placeholders. Learn more [here](./basics/validation#placeholders).

This directory can also contain a `validation.go` file, which will define custom validation rules. Learn more [here](./basics/validation).

#### HTTP Routes

The `http/routes` directory contains the routes définitions. By default, all routes are registered in the `routes.go` file, but for bigger projects, split the route definitions into multiple files.

### Resources directory

The `resources` directory is meant to store static resources such as images, HTML documents and language files. This directory shouldn't be used as a storage for dynamic content such as user profile pictures.

#### Language resources directory

The `resources/lang` directory contains your application's supported languages and translations. Each language has its own directory and should be named by an [ISO 639-1](https://en.wikipedia.org/wiki/List_of_ISO_639-1_codes) language code. You can also append a variant to your languages: `en-US`, `en-UK`, `fr-FR`, `fr-CA`, ... **Case is important.**

Each language directory contains three files. Each file is **optional**.
- `fields.json`: field names translations and field-specific rule messages.
- `locale.json`: all other language lines.
- `rules.json`: validation rules messages.

Learn more about localization [here](./advanced/localization).

## Database

Database connections are managed by the framework and are long-lived. When the server shuts down, the database connections are closed automatically. So you don't have to worry about creating, closing or refreshing database connections in your application.

If automatic migrations are enabled, all registered models at the time of startup will be auto-migrated. They must be registered before the server starts, ideally from an `init()` function next to each model definition.

Learn more in the [database](./basics/database) section.
