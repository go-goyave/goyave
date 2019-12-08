# Changelog

[[toc]]

## v2.1.0

- `filesystem.GetMIMEType` now detects `css`, `js`, `json` and `jsonld` files based on their extension.
- Added maintenance mode.
    - Can be [toggled at runtime](./advanced/multi-services.html#maintenance-mode).
    - The server can be started in maintenance mode using the `maintenance` config option. (Defaults to `false`)
- Added [advanced array validation](./basics/validation.html#validating-arrays), with support for n-dimensional arrays.<Badge text="BETA" type="warn"/>
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
- The server now shuts down when it encounters an error.
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
