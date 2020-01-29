# Authentication

<p><Badge text="Since v2.5.0"/></p>

[[toc]]

## Introduction

Goyave provides a convenient and expandable way of handling authentication in your application. Authentication can be enabled when registering your routes:

``` go
import "github.com/System-Glitch/goyave/v2/auth"

//...

authenticator := auth.Middleware(&model.User{}, &auth.BasicAuthenticator{})
router.Middleware(authenticator)
```

Authentication is handled by a simple middleware calling an **Authenticator**. This middleware also needs a model, which will be used to fetch user information on a successful login.

#### auth.Middleware

Middleware create a new authenticator middleware to authenticate the given model using the given authenticator.

| Parameters                    | Return              |
|-------------------------------|---------------------|
| `model interface{}`           | `goyave.Middleware` |
| `authenticator Authenticator` |                     |

**Example:**
``` go
authenticator := auth.Middleware(&model.User{}, &auth.BasicAuthenticator{})
router.Middleware(authenticator)
```

## Authenticators

This section will go into more details about Authenticators and explain the built-in ones. You will also learn how to implement an authenticator yourself.

**`Authenticator`** is a functional interface with a single method accepting a request and a model pointer as parameters.

``` go
Authenticate(request *goyave.Request, user interface{}) error
```

The goal of this function is to check user credentials, most of the time from the request's **headers**. If they are correct and the user can be authenticated, the `user` parameter is updated with the user's information. User information is most of the time fetched from the database.

On the other hand, if the user cannot be authenticated, the `Authenticate` method must return an `error` containing a localized message. For example, the error could be that the token lifetime is expired, thus "Your authentication token is expired." will be returned.

Authenticators use their model's struct fields tags to know which field to use for username and password. To make your model compatible with authentication, you must add the `auth:"username"` and `auth:"password"` tags:

``` go
type User struct {
	gorm.Model
	Email    string `gorm:"type:varchar(100);unique_index" auth:"username"`
	Name     string `gorm:"type:varchar(100)"`
	Password string `gorm:"type:varchar(60)" auth:"password"`
}
```

::: warning
- The username should be **unique**.
- Passwords should be **hashed** before being stored in the database.

Built-in Goyave Authenticators use [`bcrypt`](https://godoc.org/golang.org/x/crypto/bcrypt) to check if a password matches the user request.
:::

When a user is successfully authenticated on a protected route, its information is available in the controller handler, through, the request `User` field.

``` go
user := request.User.(*model.User)
response.String(http.StatusOK, "Hello " + user.Name)
```

::: tip
Remember that Goyave is primarily focused on APIs. It doesn't use session nor cookies in its core features, making requests **stateless**.
:::

### Basic Auth

[Basic authentication](https://en.wikipedia.org/wiki/Basic_access_authentication) is an authentication method using the `Authorization` header and a simple username and password combination. There are two built-in Authenticators for Basic auth.

#### Database provider

This Authenticator fetches the user information from the database, using the field tags explained earlier.

#### Config provider

This Authenticator fetches the user information from the config. This method is good for quick proof-of-concepts, as it requires minimum setup, but shouldn't be used in real-world applications.

- The `authUsername` config entry defines the username that must be matched.
- The `authPassword` config entry defines the password that must be matched.

To apply this protection to your routes, add the following middleware:

``` go
router.Middleware(auth.ConfigBasicAuth())
```

The model used for this Authenticator is `auth.BasicUser`:
``` go
type BasicUser struct {
	Name string
}
```

#### auth.ConfigBasicAuth

Create a new authenticator middleware for config-based Basic authentication. On auth success, the request user is set to a `auth.BasicUser`.
The user is authenticated if the `authUsername` and `authPassword` config entries match the request's Authorization header.

| Parameters | Return              |
|------------|---------------------|
|            | `goyave.Middleware` |

### JSON Web Token (JWT)

### Writing custom Authenticator

## Permissions

<p style="text-align: center">
    <img :src="$withBase('/undraw_in_progress_ql66.svg')" height="150" alt="In progress">
</p>

::: warning
This feature is not implemented yet and is coming in a future release.

[Watch](https://github.com/System-Glitch/goyave) the github repository to stay updated.
:::
