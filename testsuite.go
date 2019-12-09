package goyave

import (
	"context"
	"crypto/tls"
	"net/http"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/System-Glitch/goyave/v2/config"
	"github.com/System-Glitch/goyave/v2/lang"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	testify "github.com/stretchr/testify/suite"
)

// ITestSuite is an extension of testify's Suite for
// Goyave-specific testing.
type ITestSuite interface {
	RunServer(func(*Router), func())
	Timeout() time.Duration
	SetTimeout(time.Duration)
	T() *testing.T
	SetT(*testing.T)
}

// TestSuite is an extension of testify's Suite for
// Goyave-specific testing.
type TestSuite struct {
	suite.Suite
	timeout     time.Duration // Timeout for functional tests
	httpClient  *http.Client
	previousEnv string
	mu          sync.Mutex
}

// Timeout get the timeout for test failure when using RunServer.
func (s *TestSuite) Timeout() time.Duration {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.timeout
}

// SetTimeout set the timeout for test failure when using RunServer.
func (s *TestSuite) SetTimeout(timeout time.Duration) {
	s.mu.Lock()
	s.timeout = timeout
	s.mu.Unlock()
}

// RunServer start the application and run the given functional test procedure.
// The test fails if the suite's timeout is exceeded.
func (s *TestSuite) RunServer(routeRegistrer func(*Router), procedure func()) {
	c := make(chan bool, 1)
	c2 := make(chan bool, 1)
	ctx, cancel := context.WithTimeout(context.Background(), s.Timeout())
	defer cancel()

	RegisterStartupHook(func() {
		procedure()
		Stop()
		ClearStartupHooks()
		c <- true
	})

	go func() {
		Start(routeRegistrer)
		c2 <- true
	}()

	select {
	case <-ctx.Done():
		s.Fail("Timeout exceeded in goyave.TestSuite.RunServer")
	case <-c:
	}
	<-c2
}

// getHTTPClient get suite's http client or create it if it doesn't exist yet.
// The HTTP client is created with a timeout, disabled redirect and disabled TLS cert checking.
func (s *TestSuite) getHTTPClient() *http.Client {
	config := &tls.Config{
		InsecureSkipVerify: true,
	}

	return &http.Client{
		Timeout:   s.Timeout(),
		Transport: &http.Transport{TLSClientConfig: config},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}

// RunTest run a test suite with prior initialization of a test environment.
// The GOYAVE_ENV environment variable is automatically set to "test" and restored
// to its original value at the end of the test run.
func RunTest(t *testing.T, suite ITestSuite) bool {
	if suite.Timeout() == 0 {
		suite.SetTimeout(5 * time.Second)
	}
	oldEnv := os.Getenv("GOYAVE_ENV")
	os.Setenv("GOYAVE_ENV", "test")
	defer os.Setenv("GOYAVE_ENV", oldEnv)
	if err := config.Load(); err != nil {
		return assert.Fail(t, "Failed to load config", err)
	}
	lang.LoadDefault()
	lang.LoadAllAvailableLanguages()

	testify.Run(t, suite)
	return !t.Failed()
}
