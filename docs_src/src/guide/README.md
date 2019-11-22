# Introduction

Welcome to the Goyave documentation!  


This documentation is both a guide and a reference for Goyave application building. You will find instructions covering the basics as well as more advanced topics, from project setup to deployment. But first, let's talk about the framework itself.

Goyave is a framework aiming at **cleanliness**, **speed** and **power**. Goyave applications stay clean and concise thanks to minimalist function calls and route handlers. The framework gives you all the tools to create an easily readable and maintainable web application, which let you concentrate on the business logic. Although Goyave handles many things for you, such as headers or marshaling, this characteristic doesn't compromise on your freedom of code.

Most golang frameworks for web development don't have a strong directory structure nor conventions to make applications have a uniform architecture and limit redundancy. This makes it difficult to work with them on different projects. In companies, having a well-defined and documented architecture helps new developers integrate projects faster, and reduces the time needed for maintaining them. For open source projets, it helps newcomers understanding the project and makes it easier to contribute.

::: tip Note
Please feel free to sudgest changes, ask for more details, report grammar errors, or notice of uncovered scenarios by [creating an issue](https://github.com/System-Glitch/goyave/issues/new/choose) with the proposal template.
:::

## Roadmap

::: img-row <img :src="$withBase('/undraw_to_do_list_a49b.svg')" height="150" alt="Roadmap"/>
### Next release

- Integrated testing functions
- Maintenance mode (always return HTTP 503 when enabled)
- Native handlers
:::

### Ideas for future releases

- Direct support for authentication
- Plugins
- Improved Gorm integration
- CLI utility to help creating controllers, middlewares, etc
- Email helpers
- Server shutdown hooks (to gracefully close websocket connections for example)
- Queues/Scheduler system
- Logging
- Custom HTTP error handlers (for custom 404/500/... messages)
- Named routes, CORS
- Placeholders in language lines and pluralization
- Array validation
- And more!
