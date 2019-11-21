# Upgrade Guide

Although Goyave is developed with backwards compatibility, breaking changes can happen, especially in the project's early days. This guide will help you to upgrade your applications using older versions of the framework. Bear in mind that if you are several versions behind, you will have to follow the instructions for each in-between versions.

[[toc]]

## v1.0.0 to v1.1.0

- `goyave.Request.URL()` has been renamed to `goyave.Request.URI()`.
    - `goyave.Request.URL()` is now a data accessor for URL fields.
