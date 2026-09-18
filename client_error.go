package goyave

import (
	"fmt"
	"net/http"
)

// ClientError represents an error caused by the client at the origin of the request.
//
// These errors should not be wrapped using the `errwrap` package since they are not
// server errors and only need to result in a 4xx response.
// Wrapping them doesn't break the functionality, it just wastes processing time used for
// stackframe collection.
//
// This interface provides a way for services to return standardized client errors
// without leaking into the presentation / HTTP layer.
type ClientError interface {
	error

	// Message returns a human-readable, safe to expose string (often generated using the `lang` package).
	// This message will be written in the response.
	//
	// If an empty string is returned, only the response status is set in the response and no body is written directly.
	// This results in the execution of the status handler corresponding to the [ClientError.Code].
	Message() string

	// Code returns the HTTP status code that needs to be written to the response.
	// The returned code should only be in the 400 range.
	Code() int
}

type clientError struct {
	message string
	code    int
}

func (e clientError) Error() string {
	message := e.message
	if message == "" {
		message = http.StatusText(e.Code())
	}
	return fmt.Sprintf("client error %d: %s", e.Code(), message)
}

func (e clientError) Message() string {
	return e.message
}

func (e clientError) Code() int {
	return e.code
}

func (e clientError) Is(err error) bool {
	switch clientErr := err.(type) {
	case clientError, *clientError:
		return true
	case ClientError:
		return clientErr.Code() == e.code
	}
	return false
}

func (e clientError) As(target any) bool {
	var clientErr ClientError
	switch t := target.(type) {
	case **ConflictError:
		*t = Conflict(e.message)
		clientErr = *t
	case **NotAcceptableError:
		*t = NotAcceptable(e.message)
		clientErr = *t
	case **NotFoundError:
		*t = NotFound(e.message)
		clientErr = *t
	case **UnprocessableEntityError:
		*t = UnprocessableEntity(e.message)
		clientErr = *t
	case **LockedError:
		*t = Locked(e.message)
		clientErr = *t
	case **ForbiddenError:
		*t = Forbidden(e.message)
		clientErr = *t
	case **UnauthorizedError:
		*t = Unauthorized(e.message)
		clientErr = *t
	case **BadRequestError:
		*t = BadRequest(e.message)
		clientErr = *t
	// No need to handle this case here, errors.As already does since
	// the type is the exact same.
	// case **clientError:
	// 	*t = &clientError{
	// 		message: e.message,
	// 		code:    e.code,
	// 	}
	// 	clientErr = *t
	default:
		return false
	}

	return clientErr.Code() == e.code
}

// NewClientError returns a new generic client error.
// The returned error implements [errors.Is] and [errors.As]
// and will match the specific client error types based on their code.
//
// Prefer using the specific client error constructors.
func NewClientError(code int, message string) ClientError {
	return &clientError{
		message: message,
		code:    code,
	}
}

//----------------------

type ConflictError struct {
	message string
}

func (e ConflictError) Error() string {
	message := e.message
	if message == "" {
		message = http.StatusText(e.Code())
	}
	return fmt.Sprintf("client error %d: %s", e.Code(), message)
}

func (e ConflictError) Message() string {
	return e.message
}

func (ConflictError) Code() int {
	return http.StatusConflict
}

func Conflict(message string) *ConflictError {
	return &ConflictError{message: message}
}

//----------------------

type NotAcceptableError struct {
	message string
}

func (e NotAcceptableError) Error() string {
	message := e.message
	if message == "" {
		message = http.StatusText(e.Code())
	}
	return fmt.Sprintf("client error %d: %s", e.Code(), message)
}

func (e NotAcceptableError) Message() string {
	return e.message
}

func (NotAcceptableError) Code() int {
	return http.StatusNotAcceptable
}

func NotAcceptable(message string) *NotAcceptableError {
	return &NotAcceptableError{message: message}
}

//----------------------

type NotFoundError struct {
	message string
}

func (e NotFoundError) Error() string {
	message := e.message
	if message == "" {
		message = http.StatusText(e.Code())
	}
	return fmt.Sprintf("client error %d: %s", e.Code(), message)
}

func (e NotFoundError) Message() string {
	return e.message
}

func (NotFoundError) Code() int {
	return http.StatusNotFound
}

func NotFound(message string) *NotFoundError {
	return &NotFoundError{message: message}
}

//----------------------

type UnprocessableEntityError struct {
	message string
}

func (e UnprocessableEntityError) Error() string {
	message := e.message
	if message == "" {
		message = http.StatusText(e.Code())
	}
	return fmt.Sprintf("client error %d: %s", e.Code(), message)
}

func (e UnprocessableEntityError) Message() string {
	return e.message
}

func (UnprocessableEntityError) Code() int {
	return http.StatusUnprocessableEntity
}

func UnprocessableEntity(message string) *UnprocessableEntityError {
	return &UnprocessableEntityError{message: message}
}

//----------------------

type LockedError struct {
	message string
}

func (e LockedError) Error() string {
	message := e.message
	if message == "" {
		message = http.StatusText(e.Code())
	}
	return fmt.Sprintf("client error %d: %s", e.Code(), message)
}

func (e LockedError) Message() string {
	return e.message
}

func (LockedError) Code() int {
	return http.StatusLocked
}

func Locked(message string) *LockedError {
	return &LockedError{message: message}
}

//----------------------

type ForbiddenError struct {
	message string
}

func (e ForbiddenError) Error() string {
	message := e.message
	if message == "" {
		message = http.StatusText(e.Code())
	}
	return fmt.Sprintf("client error %d: %s", e.Code(), message)
}

func (e ForbiddenError) Message() string {
	return e.message
}

func (ForbiddenError) Code() int {
	return http.StatusForbidden
}

func Forbidden(message string) *ForbiddenError {
	return &ForbiddenError{message: message}
}

//----------------------

type UnauthorizedError struct {
	message string
}

func (e UnauthorizedError) Error() string {
	message := e.message
	if message == "" {
		message = http.StatusText(e.Code())
	}
	return fmt.Sprintf("client error %d: %s", e.Code(), message)
}

func (e UnauthorizedError) Message() string {
	return e.message
}

func (UnauthorizedError) Code() int {
	return http.StatusUnauthorized
}

func Unauthorized(message string) *UnauthorizedError {
	return &UnauthorizedError{message: message}
}

//----------------------

type BadRequestError struct {
	message string
}

func (e BadRequestError) Error() string {
	message := e.message
	if message == "" {
		message = http.StatusText(e.Code())
	}
	return fmt.Sprintf("client error %d: %s", e.Code(), message)
}

func (e BadRequestError) Message() string {
	return e.message
}

func (BadRequestError) Code() int {
	return http.StatusBadRequest
}

func BadRequest(message string) *BadRequestError {
	return &BadRequestError{message: message}
}

// TODO the list is not exhaustive but should largely cover the most common cases.
