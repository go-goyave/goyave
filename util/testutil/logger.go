package testutil

// LogWriter implementation of `io.Writer` redirecting the logs to `testing.T.Log()`
type LogWriter struct {
	T interface {
		Log(args ...any)
	}
}

func (w LogWriter) Write(b []byte) (int, error) {
	w.T.Log(string(b))
	return len(b), nil
}
