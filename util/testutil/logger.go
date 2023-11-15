package testutil

// LogWriter implementation of `io.Writer` redirecting the logs to `testing.T.Log()`
type LogWriter struct {
	t interface {
		Log(args ...any)
	}
}

func (w LogWriter) Write(b []byte) (int, error) {
	w.t.Log(string(b))
	return len(b), nil
}
