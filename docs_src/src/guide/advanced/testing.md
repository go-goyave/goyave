# Testing <Badge text="Since v2.2.0"/>

[[toc]]

## Introduction

Goyave provides an API to ease the unit and functional testing of your application. This API is an extension of [testify](https://github.com/stretchr/testify). `goyave.TestSuite` inherits from testify's `suite.Suite`, and sets up the environment for you. That means:

- `GOYAVE_ENV` environment variable is set to `test` and restored to its original value when the suite is done.
- All tests are run using your project's root as a working directory. This directory is determined by the presence of a `go.mod` file.
- Config and language files are loaded before the tests start. As the environment is set to `test`, you **need** a `config.test.json` in the root directory of your project.

This setup is done by the function `goyave.RunTest`, so you shouldn't run your test suites using testify's `suite.Run()` function.

``` go
import (
    "my-project/http/route"
    "github.com/System-Glitch/goyave/v2"
)

type CustomTestSuite struct {
	goyave.TestSuite
}

func (suite *CustomTestSuite) TestBasicTest() {
    suite.RunServer(route.Register, func() {
		resp, err := suite.Get("/hello", nil)
		suite.Nil(err)
		suite.NotNil(resp)
		if resp != nil {
			suite.Equal(200, resp.StatusCode)
			suite.Equal("Hi!", string(suite.GetBody(resp)))
		}
	})
}

func TestCustomSuite(t *testing.T) {
	RunTest(t, new(CustomTestSuite))
}
```

We will explain in more details what this test does in the following sections, but in short, this test runs the server, registers all your application routes and executes the second parameter as a server startup hook. The test requests the `/hello` route with the method `GET` and checks the content of the response. The server automatically shuts down after the hook is executed and before `RunServer` returns.

::: warning
You cannot run Goyave test suites in parallel.
:::

## HTTP Tests

As shown in the example in the introduction, you can easily run a test server and send requests to it using the `suite.RunServer()` function. 

This function takes two parameters.
- The first is a route registrer function. You should always use your main route registrer to avoid unexpected problems with inherited middleware and route groups.
- The second parameter is a startup hook for the server, in which you will be running your test procedure.

This function is the equivalent of `goyave.Start`, but doesn't require a `goyave.Stop()` call: the server stops automatically when the startup hook is finished. All startup hooks are then cleared. The function returns when the server is properly shut down.

If you registered startup hooks in your main function, these won't be executed. If you need them for your tests, you need to register them before calling `suite.RunServer()`.

### Request and response

There are helper functions for the following HTTP methods:
- GET
- POST
- PUT
- PATCH
- DELETE

| Parameters                  | Return           |
|-----------------------------|------------------|
| `route string`              | `*http.Response` |
| `headers map[string]string` |                  |
| `body io.Reader`            |                  |

*Note*: The `Get` function doesn't have a third parameter as GET requests shouldn't have a body. The headers and body are optional, and can be `nil`.

The response body can be retrieved easily using `suite.GetBody(response)`.

``` go 
resp, err := suite.Get("/get", nil)
suite.Nil(err)
if err == nil {
    suite.Equal("response content", string(suite.GetBody(resp)))
}
```

``` go
headers := map[string]string{"Content-Type": "application/x-www-form-urlencoded; param=value"}
resp, err = suite.Post("/post", headers, strings.NewReader("field=value"))
suite.Nil(err)
if err == nil {
    suite.Equal("post", string(suite.GetBody(resp)))
}
```

If you need to test another method, you can use the [`suite.Request()`](#testsuite-request) function.

### Timeout

`goyave.TestSuite` has a default timeout value of **5 seconds**. This timeout is used for the `RunServer` function as well as for the request functions(`Get`, `Post`, etc.). If the timeout expires, the test fails. This prevents your test from freezing if something goes wrong.

The timeout can be modified as needed using `suite.SetTimeout()`:
``` go
suite.SetTimeout(10 * time.Second)
```

### Testing JSON reponses

### File upload

## Testing middleware

## TestSuite reference

## Database testing

<p style="text-align: center">
    <img :src="$withBase('/undraw_in_progress_ql66.svg')" height="150" alt="In progress">
</p>

::: warning
This feature is not implemented yet and is coming in a future release.

[Watch](https://github.com/System-Glitch/goyave) the github repository to stay updated.
:::

### Factories

### Seeders
