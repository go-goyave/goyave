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
- Config is now protected with a `sync.Mutex` to avoid data races in multi-threaded environments.
- Better use of mutexes to protect the server instance and avoid data races in multi-threaded environments.
- Config can now be reloaded manually.
- Added the [`Trim`](./basics/middleware#trim) middleware.
