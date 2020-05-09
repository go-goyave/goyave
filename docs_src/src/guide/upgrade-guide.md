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
