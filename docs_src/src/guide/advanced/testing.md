---
meta:
  - name: "og:title"
    content: "Testing - Goyave"
  - name: "twitter:title"
    content: "Testing - Goyave"
  - name: "title"
    content: "Testing - Goyave"
---

# Testing <Badge text="Since v2.2.0"/>

[[toc]]

## Introduction

Goyave provides an API to ease the unit and functional testing of your application. This API is an extension of [testify](https://github.com/stretchr/testify). `goyave.TestSuite` inherits from testify's `suite.Suite`, and sets up the environment for you. That means:

- `GOYAVE_ENV` environment variable is set to `test` and restored to its original value when the suite is done.
- All tests are run using your project's root as working directory. This directory is determined by the presence of a `go.mod` file.
- Config and language files are loaded before the tests start. As the environment is set to `test`, you **need** a `config.test.json` in the root directory of your project.

This setup is done by the function `goyave.RunTest`, so you shouldn't run your test suites using testify's `suite.Run()` function.

The following example is a **functional** test and would be located in the `test` package.

``` go
import (
    "github.com/username/projectname/http/route"
    "github.com/System-Glitch/goyave/v3"
)

type CustomTestSuite struct {
    goyave.TestSuite
}

func (suite *CustomTestSuite) TestHello() {
    suite.RunServer(route.Register, func() {
        resp, err := suite.Get("/hello", nil)
        suite.Nil(err)
        suite.NotNil(resp)
        if resp != nil {
            defer resp.Body.Close()
            suite.Equal(200, resp.StatusCode)
            suite.Equal("Hi!", string(suite.GetBody(resp)))
        }
    })
}

func TestCustomSuite(t *testing.T) {
    goyave.RunTest(t, new(CustomTestSuite))
}
```

We will explain in more details what this test does in the following sections, but in short, this test runs the server, registers all your application routes and executes the second parameter as a server startup hook. The test requests the `/hello` route with the method `GET` and checks the content of the response. The server automatically shuts down after the hook is executed and before `RunServer` returns. See the available assertions in the [testify's documentation](https://pkg.go.dev/github.com/stretchr/testify).

This test is a **functional** test. Therefore, it requires route registration and should be located in the `test` package.

### Running the tests

Goyave tests can be run like regular tests, using the `go test` command. **It is recommended to run tests using `go test ./...` to run all tests, including subpackages**.

::: warning
Because tests using `goyave.TestSuite` are using the global config, are changing environment variables and working directory and often bind a port, they are **not run in parallel** to avoid conflicts. You don't have to use `-p 1` in your test command, test suites execution is locked by a mutex.
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

The response body can be retrieved easily using [`suite.GetBody(response)`](#suite-getbody).

``` go 
resp, err := suite.Get("/get", nil)
suite.Nil(err)
if err == nil {
    defer resp.Body.Close()
    suite.Equal("response content", string(suite.GetBody(resp)))
}
```

#### URL-encoded requests

``` go
headers := map[string]string{"Content-Type": "application/x-www-form-urlencoded; param=value"}
resp, err := suite.Post("/product", headers, strings.NewReader("field=value"))
suite.Nil(err)
if err == nil {
    defer resp.Body.Close()
    suite.Equal("response content", string(suite.GetBody(resp)))
}
```

#### JSON requests

``` go
headers := map[string]string{"Content-Type": "application/json"}
body, _ := json.Marshal(map[string]interface{}{"name": "Pizza", "price": 12.5})
resp, err := suite.Post("/product", headers, bytes.NewReader(body))
suite.Nil(err)
if err == nil {
    defer resp.Body.Close()
    suite.Equal("response content", string(suite.GetBody(resp)))
}
```

:::tip
If you need to test another method, you can use the [`suite.Request()`](#testsuite-request) function.
:::

### Timeout

`goyave.TestSuite` has a default timeout value of **5 seconds**. This timeout is used for the `RunServer` function as well as for the request functions(`Get`, `Post`, etc.). If the timeout expires, the test fails. This prevents your test from freezing if something goes wrong.

The timeout can be modified as needed using `suite.SetTimeout()`:
``` go
suite.SetTimeout(10 * time.Second)
```

### Testing JSON reponses

It is very likely that you will need to check the content of a JSON response when testing your application. Instead of unmarshaling the JSON yourself, Goyave provides the [`suite.GetJSONBody`](#suite-getjsonbody) function. This function decodes the raw body of the request. If the data cannot be decoded, or is invalid JSON, the test fails and the function returns `nil`.

``` go
suite.RunServer(route.Register, func() {
    resp, err := suite.Get("/product", nil)
    suite.Nil(err)
    if err == nil {
        defer resp.Body.Close()
        json := map[string]interface{}{}
        err := suite.GetJSONBody(resp, &json)
        suite.Nil(err)
        if err == nil { // You should always check parsing error before continuing.
            suite.Equal("value", json["field"])
            suite.Equal(float64(42), json["number"])
        }
    }
})
```

### Multipart and file upload

You may need to test requests requiring file uploads. The best way to do this is using Go's `multipart.Writer`. Goyave provides functions to help you create a Multipart form.

``` go
suite.RunServer(route.Register, func() {
    const path = "profile.png"
    body := &bytes.Buffer{}
    writer := multipart.NewWriter(body)
    suite.WriteField(writer, "email", "johndoe@example.org")
    suite.WriteFile(writer, path, "profile_picture", filepath.Base(path))
    if err := writer.Close(); err != nil {
        panic(err)
    }

    // Don't forget to set the "Content-Type" header!
    headers := map[string]string{"Content-Type": writer.FormDataContentType()}

    resp, err := suite.Post("/register", headers, body)
    suite.Nil(err)
    if err == nil {
        defer resp.Body.Close()
        suite.Equal("Welcome!", string(suite.GetBody(resp)))
    }
})
```

::: tip
You can write a multi-file upload by calling `suite.WriteFile` successively using the same field name.
:::

## Testing middleware

You can unit-test middleware using the [`suite.Middleware`](#suite-middleware) function. This function passes a `*goyave.Request` to your middlware and returns the `*http.Response`. This function also takes a test procedure function as a parameter. This function will simulate a **controller handler**, so you can test if the middleware alters the request.

``` go
rawRequest := httptest.NewRequest("GET", "/test-route", nil)
rawRequest.Header.Set("Content-Type", "application/json")
request := suite.CreateTestRequest(rawRequest)
request.Data = map[string]interface{}{"text": "  \n  test  \t"}

result := suite.Middleware(middleware.Trim, request, func(response *Response, request *Request) {
    suite.Equal("application/json", request.Header().Get("Content-Type"))
    suite.Equal("test", request.String("text"))
})

suite.Equal(200, result.StatusCode)
```

If you want to test a blocking middleware, flag the test as failed in the test procedure. Indeed, the procedure shouldn't be executed if your middleware doesn't pass to the next handler.

``` go
request := suite.CreateTestRequest(nil)
suite.Middleware(middleware.Auth, request, func(response *Response, request *Request) {
    suite.Fail("Auth middleware passed")
})
```

## TestSuite reference

::: table
[RunServer](#testsuite-runserver)
[Timeout](#testsuite-timeout)
[SetTimeout](#testsuite-settimeout)
[Middleware](#testsuite-middleware)
[Get](#testsuite-get)
[Post](#testsuite-post)
[Put](#testsuite-put)
[Patch](#testsuite-patch)
[Delete](#testsuite-delete)
[Request](#testsuite-request)
[GetBody](#testsuite-getbody)
[GetJSONBody](#testsuite-getjsonbody)
[CreateTestFiles](#testsuite-createtestfiles)
[CreateTestRequest](#testsuite-createtestrequest)
[CreateTestResponse](#testsuite-createtestresponse)
[CreateTestResponseWithRequest](#testsuite-createtestresponsewithrequest)
[WriteFile](#testsuite-writefile)
[WriteField](#testsuite-writefield)
[ClearDatabase](#testsuite-cleardatabase)
[ClearDatabaseTables](#testsuite-cleardatabasetables)
[RunTest](#goyave-runtest)
:::

#### TestSuite.RunServer

RunServer start the application and run the given functional test procedure.

This function is the equivalent of `goyave.Start()`.  
The test fails if the suite's timeout is exceeded.  
The server automatically shuts down when the function ends.  
This function is synchronized, that means that the server is properly stopped when the function returns.


| Parameters                            | Return |
|---------------------------------------|--------|
| `routeRegistrer func(*goyave.Router)` | `void` |
| `procedure func()`                    |        |

#### TestSuite.Timeout

Get the timeout for test failure when using RunServer or requests.

| Parameters | Return          |
|------------|-----------------|
|            | `time.Duration` |

#### TestSuite.SetTimeout

Set the timeout for test failure when using RunServer or requests.

| Parameters      | Return |
|-----------------|--------|
| `time.Duration` |        |


#### TestSuite.Middleware

Executes the given middleware and returns the HTTP response. Core middleware (recovery, parsing and language) is not executed.

| Parameters                     | Return           |
|--------------------------------|------------------|
| `middleware goyave.Middleware` | `*http.Response` |
| `request *goyave.Request`      |                  |
| `procedure goyave.Handler`     |                  |

#### TestSuite.Get

Execute a GET request on the given route. Headers are optional.

| Parameters                  | Return           |
|-----------------------------|------------------|
| `route string`              | `*http.Response` |
| `headers map[string]string` | `error`          |

#### TestSuite.Post

Execute a POST request on the given route. Headers and body are optional.

| Parameters                  | Return           |
|-----------------------------|------------------|
| `route string`              | `*http.Response` |
| `headers map[string]string` | `error`          |
| `body io.Reader`            |                  |

#### TestSuite.Put

Execute a PUT request on the given route. Headers and body are optional.

| Parameters                  | Return           |
|-----------------------------|------------------|
| `route string`              | `*http.Response` |
| `headers map[string]string` | `error`          |
| `body io.Reader`            |                  |

#### TestSuite.Patch

Execute a PATCH request on the given route. Headers and body are optional.

| Parameters                  | Return           |
|-----------------------------|------------------|
| `route string`              | `*http.Response` |
| `headers map[string]string` | `error`          |
| `body io.Reader`            |                  |

#### TestSuite.Delete

Execute a DELETE request on the given route. Headers and body are optional.

| Parameters                  | Return           |
|-----------------------------|------------------|
| `route string`              | `*http.Response` |
| `headers map[string]string` | `error`          |
| `body io.Reader`            |                  |

#### TestSuite.Request

Execute a request on the given route. Headers and body are optional.

| Parameters                  | Return           |
|-----------------------------|------------------|
| `method string`             | `*http.Response` |
| `route string`              | `error`          |
| `headers map[string]string` |                  |
| `body io.Reader`            |                  |

#### TestSuite.GetBody

Read the whole body of a response. If read failed, test fails and return empty byte slice.

| Parameters                | Return   |
|---------------------------|----------|
| `response *http.Response` | `[]byte` |

#### TestSuite.GetJSONBody

Read the whole body of a response and decode it as JSON. If read or decode failed, test fails. The `data` parameter should be a pointer.

| Parameters                | Return  |
|---------------------------|---------|
| `response *http.Response` | `error` |
| `data interface{}`        |         |

#### TestSuite.CreateTestFiles

Create a slice of `filesystem.File` from the given paths. Files are passed to a temporary http request and parsed as Multipart form, to reproduce the way files are obtained in real scenarios.

| Parameters        | Return              |
|-------------------|---------------------|
| `paths ...string` | `[]filesystem.File` |

#### TestSuite.CreateTestRequest

Create a `*goyave.Request` from the given raw request. This function is aimed at making it easier to unit test Requests.

If passed request is `nil`, a default `GET` request to `/` is used.

| Parameters                 | Return            |
|----------------------------|-------------------|
| `rawRequest *http.Request` | `*goyave.Request` |

**Example:**
``` go
rawRequest := httptest.NewRequest("GET", "/test-route", nil)
rawRequest.Header.Set("Content-Type", "application/json")
request := suite.CreateTestRequest(rawRequest)
request.Lang = "en-US"
request.Data = map[string]interface{}{"field": "value"}
```

#### TestSuite.CreateTestResponse

Create an empty response with the given response writer. This function is aimed at making it easier to unit test Responses.

| Parameters                     | Return             |
|--------------------------------|--------------------|
| `recorder http.ResponseWriter` | `*goyave.Response` |

**Example:**
``` go
writer := httptest.NewRecorder()
response := suite.CreateTestResponse(writer)
response.Status(http.StatusNoContent)
result := writer.Result()
fmt.Println(result.StatusCode) // 204
```

#### TestSuite.CreateTestResponseWithRequest

Create an empty response with the given response writer and HTTP request. This function is aimed at making it easier to unit test Responses needing the raw request's information, such as redirects.

| Parameters                     | Return             |
|--------------------------------|--------------------|
| `recorder http.ResponseWriter` | `*goyave.Response` |
| `rawRequest *http.Request`     |                    |

**Example:**
``` go
writer := httptest.NewRecorder()
rawRequest := httptest.NewRequest("POST", "/test-route", strings.NewReader("body"))
response := suite.CreateTestResponseWithRequest(writer, rawRequest)
response.Status(http.StatusNoContent)
result := writer.Result()
fmt.Println(result.StatusCode) // 204
```

#### TestSuite.WriteFile

Write a file to the given writer. This function is handy for file upload testing. The test fails if an error occurred.

| Parameters                | Return |
|---------------------------|--------|
| `write *multipart.Writer` | `void` |
| `path string`             |        |
| `fieldName string`        |        |
| `fileName string`         |        |

#### TestSuite.WriteField

Create and write a new multipart form field. The test fails if the field couldn't be written.

| Parameters                | Return |
|---------------------------|--------|
| `write *multipart.Writer` | `void` |
| `fieldName string`        |        |
| `value string`            |        |

#### TestSuite.ClearDatabase

Delete all records in all tables. This function only clears the tables of registered models.

| Parameters | Return |
|------------|--------|
|            | `void` |

#### TestSuite.ClearDatabaseTables

Drop all tables. This function only clears the tables of registered models.

| Parameters | Return |
|------------|--------|
|            | `void` |

#### goyave.RunTest

Run a test suite with prior initialization of a test environment. The `GOYAVE_ENV` environment variable is automatically set to "test" and restored to its original value at the end of the test run.

All tests are run using your project's root as working directory. This directory is determined by the presence of a `go.mod` file.

The function returns true if the test passed.

| Parameters         | Return |
|--------------------|--------|
| `t *testing.T`     | `bool` |
| `suite ITestSuite` |        |

::: tip
`ITestSuite` is the interface `TestSuite` is implementing.
:::

## Database testing

You may need to test features interacting with your database. Goyave provides a handy way to generate and save records in your database: **factories**.

**All registered models records are automatically deleted from the database when each test suite completes.**

It is a good practice to use a separate database dedicated for testing, named `myapp_test` for example. Don't forget to change the database information in your `config.test.json` file.

All functions below require the `database`package to be imported.

``` go
import "github.com/System-Glitch/goyave/v3/database"
```

::: tip
You may want to use a clean database for each of your tests. You can clear your database before each test using [`suite.SetupTest()`](https://pkg.go.dev/github.com/stretchr/testify/suite?tab=doc#SetupTestSuite).

``` go
func (suite *CustomTestSuite) SetupTest() {
    suite.ClearDatabase()
}
```
:::

### Generators

Factories need a **generator function**. These functions generate a single random record. You can use the faking library of your choice, but in this example we are going to use [`github.com/bxcodec/faker`](https://github.com/bxcodec/faker).

```go
import "github.com/bxcodec/faker/v3"

func UserGenerator() interface{} {
    user := &User{}
    user.Name = faker.Name()

    faker.SetGenerateUniqueValues(true)
    user.Email = faker.Email()
    faker.SetGenerateUniqueValues(false)
    return user
}
```

::: tip
- `database.Generator` is an alias for `func() interface{}`.
- Generator functions should be declared in the same file as the model it is generating.
:::

Generators can also create associated records. Associated records should be generated using their respective generators. In the following example, we are generating users for an application allowing users to write blog posts.

``` go
func UserGenerator() interface{} {
    user := &User{}
    // ... Generate users fields ...

    // Generate between 0 and 10 blog posts
    rand.Seed(time.Now().UnixNano())
    user.Posts = database.NewFactory(PostGenerator).Generate(rand.Intn(10)).([]*model.Post)

    return user
}
```

### Using factories

You can create a factory from any `database.Generator`.

``` go
factory := database.NewFactory(model.UserGenerator)

// Generate 5 random users
records := factory.Generate(5).([]*model.User)

// Generate and insert 5 random users into the database
insertedRecords := factory.Save(5).([]*model.User)
```

Note that generated records will not have an ID if they are not inserted into the database.

Associated records created by the generator will also be inserted on `factory.Save`.

#### Overrides

It is possible to override some of the generated data if needed, for example if you need to test the behavior of a function with a specific value. All generated structures will be merged with the override.

``` go
override := &model.User{
    Name: "Jérémy",
}
records := factory.Override(override).Generate(10).([]*model.User)
// All generated records will have the same name: "Jérémy"
```

::: warning
Overrides must be of the **same type** as the generated record.
:::

#### Factory reference

::: table
[NewFactory](#database-newfactory)
[Override](#factory-override)
:::

#### database.NewFactory

Create a new Factory. The given generator function will be used to generate records.

| Parameters                     | Return             |
|--------------------------------|--------------------|
| `generator database.Generator` | `database.Factory` |

#### Factory.Override

Set an override model for generated records. Values present in the override model will replace the ones in the generated records. This function expects a struct **pointer** as parameter. This function returns the same instance of `Factory` so this method can be chained.

| Parameters             | Return             |
|------------------------|--------------------|
| `override interface{}` | `database.Factory` |

#### Factory.Generate

Generate a number of records using the given factory.

Returns a slice of the actual type of the generated records, meaning you can type-assert safely.

| Parameters  | Return        |
|-------------|---------------|
| `count int` | `interface{}` |

#### Factory.Save

Save generate a number of records using the given factory, insert them in the database and return the inserted records.

The returned slice is a slice of the actual type of the generated records, meaning you can type-assert safely.

| Parameters  | Return        |
|-------------|---------------|
| `count int` | `interface{}` |


### Seeders

Seeders are functions which create a number of random records in the database in order to create a full and realistic test environment. Seeders can also generate records for your models' relations.

Seeders are written in the `database/seeder` package. Each seeder should have its own file. 

``` go
package seeder

import (
    "github.com/username/projectname/database/model"
    "github.com/System-Glitch/goyave/v3/database"
)

func User() {
    database.NewFactory(model.UserGenerator).Save(10)
}
```
