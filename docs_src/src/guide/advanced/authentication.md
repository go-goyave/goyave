---
meta:
  - name: "og:title"
    content: "Authentication - Goyave"
  - name: "twitter:title"
    content: "Authentication - Goyave"
  - name: "title"
    content: "Authentication - Goyave"
---

# Authentication <Badge text="Since v2.5.0"/>

[[toc]]

## Introduction

Goyave provides a convenient and expandable way of handling authentication in your application. Authentication can be enabled when registering your routes:

``` go
import "github.com/System-Glitch/goyave/v3/auth"

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
	Email    string `gorm:"type:char(100);unique_index" auth:"username"`
	Name     string `gorm:"type:char(100)"`
	Password string `gorm:"type:char(60)" auth:"password"`
}
```

::: warning
- The username should be **unique**.
- Passwords should be **hashed** before being stored in the database.

Built-in Goyave Authenticators use [`bcrypt`](https://pkg.go.dev/golang.org/x/crypto/bcrypt) to check if a password matches the user request.
:::

When a user is successfully authenticated on a protected route, its information is available in the controller handler, through the request `User` field.

``` go
func Hello(response *goyave.Response, request *goyave.Request) {
	user := request.User.(*model.User)
	response.String(http.StatusOK, "Hello " + user.Name)
}
```

::: tip
Remember that Goyave is primarily focused on APIs. It doesn't use session nor cookies in its core features, making requests **stateless**.

If you want to implement cookie or session-based authentication, be sure to protect your application from [CSRF attacks](https://en.wikipedia.org/wiki/Cross-site_request_forgery).
:::

### Basic Auth

[Basic authentication](https://en.wikipedia.org/wiki/Basic_access_authentication) is an authentication method using the `Authorization` header and a simple username and password combination with the following format: `username:password`, encoded in base64. There are two built-in Authenticators for Basic auth.

#### Database provider

This Authenticator fetches the user information from the database, using the field tags explained earlier.

To apply this protection to your routes, add the following middleware:

``` go
authenticator := auth.Middleware(&model.User{}, &auth.BasicAuthenticator{})
router.Middleware(authenticator)
```

You can then try requesting a protected route:
```
$ curl -u username:password http://localhost:8080/hello
Hello Jérémy
```

#### Config provider

This Authenticator fetches the user information from the config. This method is good for quick proof-of-concepts, as it requires minimum setup, but shouldn't be used in real-world applications.

- The `auth.basic.username` config entry defines the username that must be matched.
- The `auth.basic.password` config entry defines the password that must be matched.

To apply this protection to your routes, start by adding the following content to your configuration:

```json
{
  ...
  "auth": {
    "basic": {
      "username": "admin",
      "password": "admin"
    }
  }
}
```

Then, add the following middleware:

``` go
router.Middleware(auth.ConfigBasicAuth())
```

The model used for this Authenticator is `auth.BasicUser`:
``` go
type BasicUser struct {
	Name string
}
```

You can then try requesting a protected route:
```
$ curl -u username:password http://localhost:8080/hello
```

#### auth.ConfigBasicAuth

Create a new authenticator middleware for config-based Basic authentication. On auth success, the request user is set to a `auth.BasicUser`.
The user is authenticated if the `auth.basic.username` and `auth.basic.password` config entries match the request's Authorization header.

| Parameters | Return              |
|------------|---------------------|
|            | `goyave.Middleware` |

### JSON Web Token (JWT)

JWT, or [JSON Web Token](https://en.wikipedia.org/wiki/JSON_Web_Token), is an open standard of authentication that defines a compact and self-contained way for securely transmitting information between parties as a JSON object. This information can be verified and trusted because it is digitally signed. JWTs can be signed using a secret (with the HMAC algorithm) or a public/private key pair using RSA or ECDSA. Goyave uses HMAC-SHA256 in its implementation.

JWT Authentication comes with two configuration entries:

- `auth.jwt.expiry`: the number of seconds a token is valid for. Defaults to `300` (5 minutes).
- `auth.jwt.secret`: the secret used for the HMAC signature. This entry **doesn't have a default value**, you need to define it yourself. Use a key that is **at least 256 bits long**.

To apply JWT protection to your routes, start by adding the following content to your configuration:

```json
{
  ...
  "auth": {
    "jwt": {
      "expiry": 300,
      "secret": "jwt-secret"
    }
  }
}
```

Then, add the following middleware:

``` go
authenticator := auth.Middleware(&model.User{}, &auth.JWTAuthenticator{})
router.Middleware(authenticator)
```

To request a protected route, you will need to add the following header:
```
Authorization: Bearer <YOUR_TOKEN>
```

---

This Authenticator comes with a built-in login controller for password grant, using the field tags explained earlier. You can register the `/auth/login` route using the helper function `auth.JWTRoutes(router)`.

#### auth.JWTRoutes

Create a `/auth` route group and registers the `POST /auth/login` validated route. Returns the new route group.

Validation rules are as follows:
- `username`: required string
- `password`: required string

The given model is used for username and password retrieval and for instantiating an authenticated request's user.

Ensure that the given router **is not** protected by JWT authentication, otherwise your users wouldn't be able to log in.

| Parameters              | Return           |
|-------------------------|------------------|
| `router *goyave.Router` | `*goyave.Router` |
| `model interface{}`     |                  |

**Example:**
``` go
func Register(router *goyave.Router) {
	auth.JWTRoutes(router, &model.User{})
}
```

#### auth.NewJWTController

If you want or need ot register the routes yourself, you can instantiate a new JWTController using `auth.NewJWTController()`.

This function creates a new `JWTController` that will be using the given model for login and token generation.

A `JWTController` contains one handler called `Login`.

| Parameters          | Return                |
|---------------------|-----------------------|
| `model interface{}` | `*auth.JWTController` |

**Example:**
``` go
jwtRouter := router.Subrouter("/auth")
jwtRouter.Route("POST", "/login", auth.NewJWTController(&model.User{}).Login).Validate(validation.RuleSet{
	"username": {"required", "string"},
	"password": {"required", "string"},
})
```

::: tip
By default, the controller will use the "username" and "password" fields from incoming requests for the authentication process. This can be changed by modifying the controller's  `UsernameField` and `PasswordField` structure fields:
```go
jwtController := auth.NewJWTController(&model.User{})
jwtController.UsernameField = "email"
```
:::

#### auth.GenerateToken

You may need to generate a token yourself outside of the login route. This function generates a new JWT.

The token is created using the HMAC SHA256 method and signed using the `auth.jwt.secret` config entry.  
The token is set to expire in the amount of seconds defined by the `auth.jwt.expiry` config entry.

The generated token will contain the following claims:
- `userid`: has the value of the `id` parameter
- `nbf`: "Not before", the current timestamp is used
- `exp`: "Expriy", the current timestamp plus the `auth.jwt.expiry` config entry.

| Parameters       | Return   |
|------------------|----------|
| `id interface{}` | `string` |
|                  | `error`  |

**Example:**
``` go
token, err := auth.GenerateToken(user.ID)
if err != nil {
	panic(err)
}
fmt.Println(token)
```

### Writing custom Authenticator

The Goyave authentication system is expandable, meaning that you can implement more authentication methods by creating a new `Authenticator`.

The typical `Authenticator` is an empty struct implementing the `Authenticator` interface:
``` go
type MyAuthenticator struct{}

// Ensure you're correctly implementing Authenticator.
var _ auth.Authenticator = (*MyAuthenticator)(nil) // implements Authenticator
```

The next step is to implement the `Authenticate` method. Its purpose is explained at the start of this guide.

In this example, we are going to authenticate the user using a simple token stored in the database.
``` go
func (a *MyAuthenticator) Authenticate(request *goyave.Request, user interface{}) error {
	token, ok := request.BearerToken()

	if !ok {
		return fmt.Errorf(lang.Get(request.Lang, "auth.no-credentials-provided"))
	}

	// Find the struct field tagged with `auth:"token"`
	columns := auth.FindColumns(user, "token")

	// Find the user in the database using its token
	result := database.Conn().Where(columns[0].Name+" = ?", token).First(user)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			// User not found, return "These credentials don't match our records."
			return fmt.Errorf(lang.Get(request.Lang, "auth.invalid-credentials"))
		}
		// Database error
		panic(result.Error)
	}

	// Authentication successful
	return nil
}
```

#### auth.FindColumns

Find columns in the given struct. A field matches if it has a "auth" tag with the given value.
Returns a slice of found fields, ordered as the input `fields` slice.

Promoted fields are matched as well.

If the nth field is not found, the nth value of the returned slice will be `nil`.

| Parameters          | Return           |
|---------------------|------------------|
| `strct interface{}` | `[]*auth.Column` |
| `fields ...string`  |                  |

**Example**:

Given the following struct and `username`, `notatag`, `password`:

``` go
type TestUser struct {
	gorm.Model
	Name     string `gorm:"type:varchar(100)"`
	Password string `gorm:"type:varchar(100)" auth:"password"`
	Email    string `gorm:"type:varchar(100);unique_index" auth:"username"`
}
```

``` go
fields := auth.FindColumns(user, "username", "notatag", "password")
```

The result will be the `Email` field, `nil` and the `Password` field.

::: tip
The `Column` struct is defined as follows:
``` go
type Column struct {
	Name  string
	Field *reflect.StructField
}
```
:::

## Permissions

<p style="text-align: center">
    <img :src="$withBase('/undraw_in_progress_ql66.svg')" height="150" alt="In progress">
</p>

::: warning
This feature is not implemented yet and is coming in a future release.

[Watch](https://github.com/System-Glitch/goyave) the github repository to stay updated.
:::
