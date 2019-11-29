# Changelog

[[toc]]

## v2.0.0

- Documentation and README improvements.
- In the configuration:
    - The default value of `dbConnection` has been changed to `none`.
    - The default value of `dbAutoMigrate` has been changed to `false`.
- Added [request data accessors](./basics/requests#accessors).
- Some refactoring and package renaming have been done to better respect the Go conventions.
    - The `helpers` package have been renamed to `helper`
- The server now shuts down when it encounters an error.
- New [`validation.GetFieldType`](./basics/validation#validation-getfieldtype) function.
- Config and Lang are now protected with a `sync.RWMutex` to avoid data races in multi-threaded environments.
- Greatly improve concurrency.
- Config can now be reloaded manually.
- Added the [`Trim`](./basics/middleware#trim) middleware.
- `goyave.Response` now implements `http.ResponseWriter`.
    - All writing functions can now return an error.
- Added the [`NativeHandler`](./basics/routing#native-handlers) compatibility layer.
- Fixed a bug preventing the static resources handler to find `index.html` if a directory with a depth of one was requested without a trailing slash.
