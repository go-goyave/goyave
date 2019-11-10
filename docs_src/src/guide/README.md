# Introduction

Welcome to the Goyave documentation!  


This documentation is both a guide and a reference for Goyave application building. You will find instructions covering the basics as well as more advanced topics, from project setup to deployment. But first, let's talk about the framework itself.

Goyave is a framework aiming at **cleanliness**, **speed** and **power**. Goyave applications stay clean and concise thanks to minimalist function calls and route handlers. The framework gives you all the tools to create an easily readable and maintainable web application, which let you concentrate on the business logic. Although Goyave handles many things for you, such as headers or marshaling, this characteristic doesn't compromise on your freedom of code. The framework benefits from the speed of a compiled language and uses the awesome Gorilla Mux router in the background.

::: warning
The documentation is not complete yet and still being written. It is highly subject to change in the near future.
:::

::: tip Note
Please feel free to sudgest changes, ask for more details, report grammar errors, or notice of uncovered scenarios by [creating an issue](https://github.com/System-Glitch/goyave/issues/new/choose) with the proposal template.
:::

## Roadmap

<p style="text-align: center">
    <img src="/undraw_to_do_list_a49b.svg" height="150" alt="Roadmap">
</p>

### Next release

- Integrated testing functions
- Maintenance mode (always return HTTP 503 when enabled)

### Ideas for future releases

- Direct support for authentication
- Plugins
- CLI utility to help creating controllers, middlewares, etc
- Email helpers
- Server shutdown hooks (to gracefully close websocket connections for example)
- Improve threading
- And more!
