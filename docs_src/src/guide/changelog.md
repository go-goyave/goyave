---
meta:
  - name: "og:title"
    content: "Changelog - Goyave"
  - name: "twitter:title"
    content: "Changelog - Goyave"
  - name: "title"
    content: "Changelog - Goyave"
---

# Changelog

[[toc]]

## v2.10.x

### v2.10.1

- Changed the behavior of `response.File()` and `response.Download()` to respond with a status 404 if the given file doesn't exist instead of panicking.
- Improved error handling:
    - `log.Panicf` is not used anymore to print panics, removing possible duplicate logs.
    - Added error checks during automatic migrations.
    - `goyave.Start()` now exits the program with the following error codes:
        - `2`: Panic (server already running, error when loading language files, etc)
        - `3`: Configuration is invalid
        - `4`: An error occurred when opening network listener
        - `5`: An error occurred in the HTTP server

This change will require a slightly longer `main` function but offers better flexibility for error handling and multi-services.

``` go
if err := goyave.Start(route.Register); err != nil {
	os.Exit(err.(*goyave.Error).ExitCode)
}
```

- Fixed a bug in `TestSuite`: HTTP client was re-created everytime `getHTTPClient()` was called.
- Fixed testing documentation examples that didn't close http response body.
- Documentation meta improvements.
- Protect JSON requests with `maxUploadSize`. 
- The server will now automatically return `413 Payload Too Large` if the request's size exceeds the `maxUploadSize` defined in configuration.
- The request parsing middleware doesn't drain the body anymore, improving native handler compatibility.
- Set a default status handler for all 400 errors.
- Fixed a bug preventing query parameters to be parsed when the request had the `Content-Type: application/json` header.
- Added a dark theme for the documentation. It can be toggled by clicking the moon icon next to the search bar.

### v2.10.0

- Added router `Get`, `Post`, `Put`, `Patch`, `Delete` and `Options` methods to register routes directly without having to specify a method string.
- Added [placeholder](./advanced/localization.html#placeholders) support in regular language lines.

## v2.9.0

- Added [hidden fields](./basics/database.html#hidden-fields).
- Entirely removed Gorilla mux. This change is not breaking: native middleware still work the same.

## v2.8.0

- Added a built-in logging system.
    - Added a middleware for access logs using the Common Log Format or Combined Log Format. This allows custom formatters too.
    - Added three standard loggers: `goyave.Logger`, `goyave.AccessLogger` and `goyave.ErrLogger`
- Fixed bug: the gzip middleware now closes underlying writer on close.

## v2.7.x

### v2.7.1

- Changed MIME type of `js` and `mjs` files to `text/javascript`. This is in accordance with an [IETF draft](https://datatracker.ietf.org/doc/draft-ietf-dispatch-javascript-mjs/) that treats application/javascript as obsolete.
- Improved error handling: stacktrace wasn't relevant on unexpected panic since it was retrieved from the route's request handler therefore not including the real source of the panic. Stacktrace retrieval has been moved to the recovery middleware to fix this.

### v2.7.0

- Added `Request.Request()` accessor to get the raw `*http.Request`.
- Fixed a bug allowing non-core middleware applied to the root router to be executed when the "Not Found" or "Method Not Allowed" routes were matched.
- Fixed a bug making route groups (sub-routers with empty prefix) conflict with their parent router when two routes having the same path but different methods are registered in both routers.
- Added [chained writers](./basics/responses.html#chained-writers).
- Added [gzip compression middleware](./basics/middleware.html#gzip).
- Added the ability to register route-specific middleware in `Router.Static()`.

## v2.6.0

- Custom router implementation. Goyave is not using gorilla/mux anymore. The new router is twice as fast and uses about 3 times less memory.
- Now redirects to configured protocol if request scheme doesn't match.
- Added [named routes](./basics/routing.html#named-routes).
- Added `Route.GetFullURI()` and `Route.BuildURL()` for dynamic URL generation.
- Added `helper.IndexOfStr()` and `helper.ContainsStr()` for better performance when using string slices.
- Moved from GoDoc to [pkg.go.dev](https://pkg.go.dev/github.com/System-Glitch/goyave/v2).
- Print errors to stderr.

## v2.5.0

- Added an [authentication system](./advanced/authentication.html).
- Various optimizations.
- Various documentation improvements.
- Added `dbMaxLifetime` configuration entry.
- Moved from Travis CI to Github Actions.
- Fixed a bug making duplicate log entries on error.
- Fixed a bug preventing language lines containing a dot to be retrieved.
- Fixed `TestSuite.GetJSONBody()` not working with structs and slices.
- Added `TestSuite.ClearDatabaseTables()`.
- Added `Config.Has()` and `Config.Register()` to check for the existence of a config entry and to allow custom config entries valdiation.
- Added `Request.BearerToken()`.
- Added `Response.HandleDatabaseError()` for easier database error handling and shorter controller handlers. 

## v2.4.x

### v2.4.3

- Improved string validation by taking grapheme clusters into consideration when calculating length.
- `lang.LoadDefault` now correctly creates a fresh language map and clones the default `en-US` language. This avoids the default language entries to be overridden permanently.  

### v2.4.2

- Don't override `Content-Type` header when sending a file if already set.
- Fixed a bug with validation message placeholder `:values`, which was mistakenly using the `:value` placeholder.

### v2.4.1

- Bundle default config and language in executable to avoid needing to deploy `$GOROOT/pkg/mod/github.com/!system-!glitch/goyave/` with the application.

### v2.4.0

- Added [template rendring](./basics/responses.html#response-render).
- Fixed PostgreSQL options not working.
- `TestSuite.Middleware()` now has a more realistic behavior: the finalization step of the request life-cycle is now also executed. This may require your tests to be updated if those check the status code in the response.
- Added [status handlers](./advanced/status-handlers.html).

## v2.3.0

- Added [CORS options](./advanced/cors.html).

## v2.2.x

### v2.2.1

- Added `domain` config entry. This entry is used for url generation, especially for the TLS redirect.
- Don't show port in TLS redirect response if ports are standard (80 for HTTP, 443 for HTTPS).

### v2.2.0

- Added [testing API](./advanced/testing.html).
- Fixed links in documentation.
- Fixed `models` package in template project. (Changed to `model`)
- Added [`database.ClearRegisteredModels`](./basics/database.html#database-clearregisteredmodels)

## v2.1.0

- `filesystem.GetMIMEType` now detects `css`, `js`, `json` and `jsonld` files based on their extension.
- Added maintenance mode.
    - Can be [toggled at runtime](./advanced/multi-services.html#maintenance-mode).
    - The server can be started in maintenance mode using the `maintenance` config option. (Defaults to `false`)
- Added [advanced array validation](./basics/validation.html#validating-arrays), with support for n-dimensional arrays.
- Malformed request messages can now be localized. (`malformed-request` and `malformed-json` entries in `locale.json`)
- Modified the validator to allow [manual validation](./basics/validation.html#manual-validation).

## v2.0.0

- Documentation and README improvements.
- In the configuration:
    - The default value of `dbConnection` has been changed to `none`.
    - The default value of `dbAutoMigrate` has been changed to `false`.
- Added [request data accessors](./basics/requests.html#accessors).
- Some refactoring and package renaming have been done to better respect the Go conventions.
    - The `helpers` package have been renamed to `helper`
- The server now shuts down when it encounters an error during startup.
- New [`validation.GetFieldType`](./basics/validation.html#validation-getfieldtype) function.
- Config and Lang are now protected with a `sync.RWMutex` to avoid data races in multi-threaded environments.
- Greatly improve concurrency.
- Config can now be reloaded manually.
- Added the [`Trim`](./basics/middleware.html#trim) middleware.
- `goyave.Response` now implements `http.ResponseWriter`.
    - All writing functions can now return an error.
- Added the [`NativeHandler`](./basics/routing.html#native-handlers) compatibility layer.
- Fixed a bug preventing the static resources handler to find `index.html` if a directory with a depth of one was requested without a trailing slash.
- Now panics when calling `Start()` while the server is already running.
