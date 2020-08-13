# Contributing to Goyave

Thank you very much for your time contributing to the Goyave framework!

- When reporting an [issue](https://github.com/System-Glitch/goyave/issues/new/choose), please use one of the available issue templates.
- For pull requests, please use the pull request template and select the `develop` branch as target branch.
    - Ensure that the submitted code works, is documented, respects the [Golang coding style](https://golang.org/doc/effective_go.html) and is covered by tests. All new pull requests will be automatically tested.
    - The project is linted using [golangci-lint](https://github.com/golangci/golangci-lint) and the configuration defined in `.golangci.yml`.
    - Update the documentation if needed, but don't build it.
    - Please use the latest stable version of the Go programming language. All versions from 1.13 to the latest are tested in the Github Actions workflow.
    - You can run tests locally using the `run_test.sh` script. It will setup a database container for you and shut it down when the tests are finished.
