package goyave

// Service is the bridge between the REST layer of your application and
// the domain. It is responsible of the business logic.
// Services usually bundle a repository in which the database logic is implemented.
//
// Services receive data that is expected to be validated and correctly formatted as
// a DTO (Data Transfer Object) structure. They in turn return DTOs or errors so
// controllers can format a clean HTTP response.
type Service interface {
	// Init is a function called automatically when the service
	// is registered on a server.
	Init(*Server)

	// Name returns the unique name identifier for the service.
	// The returned value should be a constant to make it easier
	// to retrieve the service.
	Name() string
}

// TODO Document why Service, why repository, etc
