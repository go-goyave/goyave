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

### Convention changes

This release brought changes to the conventions. Although your applications can still work with the old ones, it's recommended to make the change.

- Move `validation.go` and `placeholders.go` to a new `http/validation` package. Don't forget to change the `package` instruction in these files.
- In `kernel.go`, import your `http/validation` package instead of `http/request`.
- Validation rule sets are now located in a `request.go` file in the same package as the controller. So if you had `http/request/productrequest/product.go`, take the content of that file and move it to `http/controller/product/request.go`. Rule sets are now named after the name of the controller handler they will be used with, and end with `Request`. For example, a rule set for the `Store` handler will be named `StoreRequest`. If a rule set can be used for multiple handlers, consider using a name suited for all of them. The rules for a store operation are often the same for update operations, so instead of duplicating the set, create one unique set called `UpsertRequest`. You will likely just have to add `Request` at the end of the name of your sets.
- Update your route definition by changing the rule sets you use.
```go
router.Post("/echo", hello.Echo, hellorequest.Echo)

// Becomes
router.Post("/echo", hello.Echo, hello.EchoRequest)
```

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
