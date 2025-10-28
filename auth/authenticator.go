package auth

import (
	"context"
	"fmt"
	"net/http"

	"goyave.dev/goyave/v5"
)

// MetaAuth the authentication middleware will only authenticate the user
// if this meta is present in the matched route or any of its parent and is equal to `true`.
const MetaAuth = "goyave.require-auth"

const defaultRealm = "Authorization required"

// Authenticator is an object in charge of authenticating a client.
//
// The generic type should be a DTO and not be a pointer. The `request.User`
// will use this type on successful authentication.
type Authenticator[T any] interface {
	goyave.Composable

	// Authenticate fetch the user corresponding to the credentials
	// found in the given request and returns it.
	// If no user can be authenticated, returns the error detailing why the
	// authentication failed. The error message is already localized.
	//
	// The error returned doesn't need to be wrapped as it will only
	// be used for the message returned in the response.
	//
	// If an unexpected error happens (e.g.: database error), this
	// method should panic instead of returning an error.
	Authenticate(request *goyave.Request) (*T, error)
}

// SchemeAuthenticator let Authenticators indicate their scheme for use in the WWW-Authenticate
// header returned when authentication fails.
// The header is not returned if the authenticator doesn't implement this interface.
type SchemeAuthenticator interface {
	// Scheme returns the authenticator's scheme.
	// The returned value should be one of the values in the IANA's HTTP Authentication Schemes list.
	// https://www.iana.org/assignments/http-authschemes/http-authschemes.xhtml
	Scheme() string
}

// UserService is the dependency of authenticators used to retrieve a user by its "username".
//
// A username is actually any identifier (an ID, a email, a name, etc). It is the responsibility
// of the service implementation to check the type of the "username" and either convert it or
// return an error simulating a non-existing record (`gorm.ErrRecordNotFound`).
//
// If the record could not be found, the error returned should be of type `gorm.ErrRecordNotFound`.
type UserService[T any] interface {
	FindByUsername(ctx context.Context, username any) (*T, error)
}

// Unauthorizer can be implemented by Authenticators to define custom behavior
// when authentication fails.
type Unauthorizer interface {
	OnUnauthorized(response *goyave.Response, request *goyave.Request, err error)
}

// Handler a middleware that automatically sets the request's `User` if the
// authenticator succeeds.
//
// Supports the `auth.Unauthorizer` interface.
//
// The T parameter represents the user DTO and should not be a pointer.
type Handler[T any] struct {
	Authenticator[T]

	// Realm describes the protected area. It is returned in the
	// `WWW-Authenticate` header on authentication failure if the
	// Authenticator implements `SchemeAuthenticator`.
	// If empty, "Authorization Required" will be used by default.
	Realm string
}

// Handle on success, set the request's `User` to the user returned by the authenticator
// and inject it in the request's `context.Context`. The user can be retrieved from the
// context using `UserFromContext`.
//
// Blocks if the authentication is not successful.
// If the authenticator implements `SchemeAuthenticator`, add the `WWW-Authenticate` header
// to the response.
// If the authenticator implements `Unauthorizer`, `OnUnauthorized` is called,
// otherwise returns a default `401 Unauthorized` error.
// If the matched route doesn't contain the `MetaAuth` or if it's not equal to `true`,
// the middleware is skipped.
func (m *Handler[T]) Handle(next goyave.Handler) goyave.Handler {
	return func(response *goyave.Response, request *goyave.Request) {
		if requireAuth, ok := request.Route.LookupMeta(MetaAuth); !ok || requireAuth != true {
			next(response, request)
			return
		}

		user, err := m.Authenticate(request)
		if err != nil {
			if authenticateHeader := m.getAuthenticateHeader(); authenticateHeader != "" {
				response.Header().Set("WWW-Authenticate", authenticateHeader)
			}
			if unauthorizer, ok := m.Authenticator.(Unauthorizer); ok {
				unauthorizer.OnUnauthorized(response, request, err)
				return
			}
			response.JSON(http.StatusUnauthorized, map[string]string{"error": err.Error()})
			return
		}
		request.User = user
		request.WithContext(ContextWithUser(request.Context(), user))
		next(response, request)
	}
}

func (m *Handler[T]) getAuthenticateHeader() string {
	sa, ok := m.Authenticator.(SchemeAuthenticator)
	if !ok {
		return ""
	}
	return fmt.Sprintf(`%s realm="%s", charset="UTF-8"`, sa.Scheme(), m.Realm)
}

// Middleware returns an authentication middleware which will use the given
// authenticator and set the request's `User` according to the generic type `T`, which
// should be a DTO.
//
// This middleware should be used as a global middleware, and all routes (ou routers)
// that require authentication should have the meta `MetaAuth` set to `true`.
// If the matched route or any of its parent doesn't have this meta or if it's not equal to
// `true`, the authentication is skipped.
func Middleware[T any](authenticator Authenticator[T]) *Handler[T] {
	return MiddlewareWithRealm(authenticator, defaultRealm)
}

// MiddlewareWithRealm is the same as `Middleware` but with a custom realm description.
// The realm describes the protected area and is returned in the `WWW-Authenticate` header
// when the authentication fails.
// Note that the `WWW-Authenticate` header is NOT added to the response if the authenticator
// doesn't implement `SchemeAuthenticator`.
func MiddlewareWithRealm[T any](authenticator Authenticator[T], realm string) *Handler[T] {
	return &Handler[T]{
		Authenticator: authenticator,
		Realm:         realm,
	}
}

// userCtxKey the key used to store the authenticated user in the context.
type userCtxKey struct{}

// ContextWithUser inject the given user as a context value. The user
// can be retrieved from the returned context using `UserFromContext`.
func ContextWithUser(ctx context.Context, user any) context.Context {
	return context.WithValue(ctx, userCtxKey{}, user)
}

// UserFromContext return the authenticated user stored in the given context or nil.
// The type T should be identical to the type used in your Authenticator. If the user
// stored in the context isn't of type `*T`, nil is returned.
//
//	user := UserFromContext[dto.InternalUser](ctx)
func UserFromContext[T any](ctx context.Context) *T {
	if u, ok := ctx.Value(userCtxKey{}).(*T); ok {
		return u
	}
	return nil
}
