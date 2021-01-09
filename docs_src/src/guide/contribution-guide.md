---
meta:
  - name: "og:title"
    content: "Contribution guide - Goyave"
  - name: "twitter:title"
    content: "Contribution guide - Goyave"
  - name: "title"
    content: "Contribution guide - Goyave"
---

# Contributing to Goyave

Thank you very much for your time contributing to the Goyave framework!

## Workflow

New features and changes are first discussed in issues. This is where we settle on a design and discuss about it. The issue is then added to the [Github project](https://github.com/System-Glitch/goyave/projects) so everyone can know what is being worked on. If a pull request follows, the discussion is moved to it.

- When reporting an [issue](https://github.com/System-Glitch/goyave/issues/new/choose), please use one of the available issue templates.
- For pull requests, please use the pull request template and select the `develop` branch as target branch.
    - Ensure that the submitted code works, is documented, respects the [Golang coding style](https://golang.org/doc/effective_go.html) and is covered by tests. All new pull requests will be automatically tested.
    - The project is linted using [golangci-lint](https://github.com/golangci/golangci-lint) and the configuration defined in `.golangci.yml`.
    - Update the documentation if needed, but don't build it.
    - Please use the latest stable version of the Go programming language. All versions from 1.13 to the latest are tested in the Github Actions workflow.
    - You can run tests locally using the `run_test.sh` script. It will setup a database container for you and shut it down when the tests are finished.

**Where to start?**

If you would like to contribute but you are not sure on what you could work, the [issues section](https://github.com/System-Glitch/goyave/issues) is a good place to start. Check the issues with the "contributions welcome" tag. Of course, other ideas are also very much welcome!

## Design philosophy

The goal of the project is to provide an opinionated framework for REST API building. Using it must be as simple as possible and with as few code as possible, but still flexible and hackable. We don't want to gate too many things so developers have the freedom to build whatever they want, even edge cases. However, the goal is to make the most common cases a breeze, so they can focus on what their app really does: the business logic. For example, it is extremely common to have some sort of permissions on resources in REST APIs, so instead of forcing developers to spend time working on this, which can be complicated, Goyave would provide a comprehensive set of tools to achieve this.

## Financial support

I have many ideas for the future of Goyave. I would be infinitely grateful to whoever want to support me and let me continue working on Goyave and making it better and better.

You can support me on Github Sponsor or Patreon.

<iframe src="https://github.com/sponsors/System-Glitch/button" title="Sponsor System-Glitch" height="35" width="116" style="border: 0;"></iframe>

<a href="https://www.patreon.com/bePatron?u=25997573" data-patreon-widget-type="become-patron-button">Become a Patron!</a><script async src="https://c6.patreon.com/becomePatronButton.bundle.js"></script>

I'm very grateful to my patrons and donators:

- Ben Hyrman
- Massimiliano Bertinetti