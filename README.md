<p align="center">
    <img src="resources/img/logo/goyave_text.png" alt="Goyave Logo" width="550"/>
</p>

<p align="center">
    <a href="https://travis-ci.com/System-Glitch/goyave"><img src="https://api.travis-ci.com/System-Glitch/goyave.svg" alt="Build Status"/></a>
    <a href="https://github.com/System-Glitch/goyave/releases"><img src="https://img.shields.io/github/v/release/System-Glitch/goyave?include_prereleases" alt="Version"/></a>
    <a href="https://goreportcard.com/report/github.com/System-Glitch/goyave"><img src="https://goreportcard.com/badge/github.com/System-Glitch/goyave" alt="Go Report"/></a>
    <a href="https://coveralls.io/github/System-Glitch/goyave?branch=master"><img src="https://coveralls.io/repos/github/System-Glitch/goyave/badge.svg" alt="Coverage Status"/></a>
    <a href="https://github.com/System-Glitch/goyave/blob/master/LICENSE"><img src="https://img.shields.io/dub/l/vibe-d.svg" alt="License"/></a>
</p>

<h2 align="center">An Elegant Golang Web Framework</h2>

Goyave is a progressive and accessible web application framework, aimed at making development easy and enjoyable. It has a philosophy of cleanliness and conciseness to make programs more elegant, easier to maintain and more focused.

<table>
    <tr>
        <td valign="top">
            <h3>Clean Code</h3>
            <p>Goyave has an expressive, elegant syntax, a robust structure and conventions. Minimalist calls and reduced redundancy are among the Goyave's core principles.</p>
        </td>
        <td valign="top">
            <h3>Fast Development</h3>
            <p>Develop faster and concentrate on the business logic of your application thanks to the many helpers and built-in functions.</p>
        </td>
        <td valign="top">
            <h3>Powerful functionalities</h3>
            <p>Goyave is accessible, yet powerful. The frameworks includes request parsing, validation, localization, and more!</p>
        </td>
    </tr>
</table>

## Getting Started

The example belows shows a basic `Hello world` application using Goyave.

``` go
import "github.com/System-Glitch/goyave"

func Register(router *goyave.Router) {
	router.Route("GET", "/hello", func(response *goyave.Response, request *goyave.Request) {
	    response.String(http.StatusOK, "Hello world!")
    }, nil)
}

func main() {
	goyave.Start(registerRoutes)
}
```

### Scaffolding

You can bootstrap your project using the [Goyave template project](https://github.com/System-Glitch/goyave-template). This project has a complete directory structure already set up for you. Follow the **[installation instructions](https://system-glitch.github.io/goyave/guide/installation#template-project)** and start building awesome applications!

## Learning Goyave

The Goyave framework has an extensive documentation covering in-depth subjects and teaching you how to run a project using Goyave from setup to deployment.

<a href="https://system-glitch.github.io/goyave/guide/installation"><h3 align="center">Read the documentation</h3></a>

## Requirements

- Go 1.13+
- Go modules

## Contributing

Thank you for considering contributing to the Goyave framework! You can find the contribution guide in the [documentation](https://system-glitch.github.io/goyave/guide/contribution-guide.html).

I have many ideas for the future of Goyave. I would be infinitely grateful to whoever want to support me and let me continue working on Goyave and making it better and better.

You can support also me on Patreon:

<a href="https://www.patreon.com/bePatron?u=25997573">
	<img src="https://c5.patreon.com/external/logo/become_a_patron_button@2x.png" width="160">
</a>

## License

The Goyave framework is MIT Licensed. Copyright © 2019 Jérémy LAMBERT (SystemGlitch) 
