---
meta:
  - name: "og:title"
    content: "Upgrade guide - Goyave"
  - name: "twitter:title"
    content: "Upgrade guide - Goyave"
  - name: "title"
    content: "Upgrade guide - Goyave"
---

# Upgrade Guide

Although Goyave is developed with backwards compatibility, breaking changes can happen, especially in the project's early days. This guide will help you to upgrade your applications using older versions of the framework. Bear in mind that if you are several versions behind, you will have to follow the instructions for each in-between versions.

[[toc]]

## v2.x.x to v3.0.0

### Routing changes

Routing has been improved by changing how validation and route-specific middleware are registered. The signature of the router functions have been simplified by removing the validation and middleware parameters from `Route()`, `Get()`, `Post()`, etc. This is now done through two new chainable methods on the `Route`:

```go
router.Post("/echo", hello.Echo, hellorequest.Echo)

// Becomes
router.Post("/echo", hello.Echo).Validate(hello.EchoRequest)
```

```go
router.Post("/echo", hello.Echo, nil, middleware.Trim, middleware.Gzip())

// Becomes
router.Post("/echo", hello.Echo).Middleware(middleware.Trim, middleware.Gzip())
```

```go
router.Post("/echo", hello.Echo, hellorequest.Echo, middleware.Trim)

// Becomes
router.Post("/echo", hello.Echo).Validate(hello.EchoRequest).Middleware(middleware.Trim)
```

### Convention changes

This release brought changes to the conventions. Although your applications can still work with the old ones, it's recommended to make the change.

- Move `validation.go` and `placeholders.go` to a new `http/validation` package. Don't forget to change the `package` instruction in these files.
- In `kernel.go`, import your `http/validation` package instead of `http/request`.
- Validation rule sets are now located in a `request.go` file in the same package as the controller. So if you had `http/request/productrequest/product.go`, take the content of that file and move it to `http/controller/product/request.go`. Rule sets are now named after the name of the controller handler they will be used with, and end with `Request`. For example, a rule set for the `Store` handler will be named `StoreRequest`. If a rule set can be used for multiple handlers, consider using a name suited for all of them. The rules for a store operation are often the same for update operations, so instead of duplicating the set, create one unique set called `UpsertRequest`. You will likely just have to add `Request` at the end of the name of your sets.
- Update your route definition by changing the rule sets you use.
```go
router.Post("/echo", hello.Echo, hellorequest.Echo)

// Becomes
router.Post("/echo", hello.Echo).Validate(hello.EchoRequest)
```

### Validation changes

Although the validation changes are internally huge, there is only a tiny amount of code to change to update your application. You will have to update all your handlers accessing the `request.Rules` field. This field is no longer a `validation.RuleSet` and has been changed to `*validation.Rules`, which will be easier to use, as the rules are already parsed. Refer to the [alternative validation syntax](./basics/validation.html#alternative-syntax) documentation for more details about this new structure.

- The following rules now pass if the validated data type is not supported: `greater_than`, `greater_than_equal`, `lower_than`, `lower_than_equal`, `size`.

### Minor changes

- Recovery middleware now correctly handles panics with a `nil` value. You may have to update your custom status handler for the HTTP `500` error code.
- Log `Formatter` now receive the length of the response (in bytes) instead of the full body.
  - `log.Formatter` is now `func(now time.Time, response *goyave.Response, request *goyave.Request, length int) string`.
  - If you were just using `len(body)`, just replace it with `length`.
  - If you were using the content of the body in your logger, you will have to implement a [chained writer](./basics/responses.html#chained-writers).

## v1.0.0 to v2.0.0

This first update comes with refactoring and package renaming to better fit the Go conventions.

- `goyave.Request.URL()` has been renamed to `goyave.Request.URI()`.
    - `goyave.Request.URL()` is now a data accessor for URL fields.
- The `helpers` package has been renamed to `helper`.
    - The `filesystem` package thus has a different path: `github.com/System-Glitch/goyave/v2/helper/filesystem`.

::: tip
Because this version contains breaking changes. Goyave had to move to `v2.0.0`. You need to change the path of your imports to upgrade.

Change `github.com/System-Glitch/goyave` to `github.com/System-Glitch/goyave/v2`.
:::
