package goyave

const (
	// ExitInvalidConfig the exit code returned when the config
	// validation doesn't pass.
	ExitInvalidConfig = 3

	// ExitNetworkError the exit code returned when an error
	// occurs when opening the network listener
	ExitNetworkError = 4

	// ExitHTTPError the exit code returned when an error
	// occurs in the HTTP server (port already in use for example)
	ExitHTTPError = 5

	// ExitDatabaseError the exit code returned when the
	// connection to the database could not be established.
	ExitDatabaseError = 6

	// ExitStateError the exit code returned when
	// server.Start is called on an already running server
	// or a stopped server
	ExitStateError = 7

	// ExitLanguageError the exit code returned when
	// the language files could not be loaded
	ExitLanguageError = 8
)

// Error wrapper for errors directely related to the server itself.
// Contains an exit code and the original error.
type Error struct {
	err      error
	ExitCode int
}

func (e *Error) Error() string {
	return e.err.Error()
}

func (e *Error) Unwrap() error {
	return e.err
}
