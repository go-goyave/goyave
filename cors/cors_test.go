package cors

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConfigureMaxAge(t *testing.T) {
	options := Default()
	headers := http.Header{}

	options.configureMaxAge(headers)
	assert.Equal(t, "43200", headers.Get("Access-Control-Max-Age"))

	options.MaxAge = 5 * time.Second
	options.configureMaxAge(headers)
	assert.Equal(t, "5", headers.Get("Access-Control-Max-Age"))

	options.MaxAge = 6*time.Second + 500*time.Millisecond
	options.configureMaxAge(headers)
	assert.Equal(t, "6", headers.Get("Access-Control-Max-Age"))
}

func TestConfigureAllowedHeaders(t *testing.T) {
	options := Default()
	headers := http.Header{}
	requestHeaders := http.Header{}

	options.configureAllowedHeaders(headers, requestHeaders)
	assert.Equal(t, "Origin, Accept, Content-Type, X-Requested-With, Authorization", headers.Get("Access-Control-Allow-Headers"))

	options.AllowedHeaders = []string{}
	requestHeaders.Set("Access-Control-Request-Headers", "Accept, Origin")

	options.configureAllowedHeaders(headers, requestHeaders)
	assert.Equal(t, "Accept, Origin", headers.Get("Access-Control-Allow-Headers"))
	assert.Equal(t, "Access-Control-Request-Headers", headers.Get("Vary"))
}

func TestConfigureAllowedMethods(t *testing.T) {
	options := Default()
	options.AllowedMethods = []string{http.MethodGet, http.MethodPost}
	headers := http.Header{}

	options.configureAllowedMethods(headers)
	assert.Equal(t, "GET, POST", headers.Get("Access-Control-Allow-Methods"))
}

func TestConfigureExposedHeaders(t *testing.T) {
	options := Default()
	headers := http.Header{}

	options.configureExposedHeaders(headers)
	assert.Empty(t, headers.Get("Access-Control-Expose-Headers"))

	options.ExposedHeaders = []string{"Content-Type", "Accept"}
	options.configureExposedHeaders(headers)
	assert.Equal(t, "Content-Type, Accept", headers.Get("Access-Control-Expose-Headers"))
}

func TestConfigureCredentials(t *testing.T) {
	options := Default()
	headers := http.Header{}

	options.configureCredentials(headers)
	assert.Empty(t, headers.Get("Access-Control-Allow-Credentials"))

	options.AllowCredentials = true
	options.configureCredentials(headers)
	assert.Equal(t, "true", headers.Get("Access-Control-Allow-Credentials"))
}

func TestConfigureOrigin(t *testing.T) {
	options := Default()
	headers := http.Header{}
	requestHeaders := http.Header{}

	options.configureOrigin(headers, requestHeaders)
	assert.Equal(t, "*", headers.Get("Access-Control-Allow-Origin"))

	headers = http.Header{}
	requestHeaders = http.Header{"Origin": {"https://google.com"}}
	options.AllowedOrigins = []string{"https://google.com", "https://images.google.com"}

	options.configureOrigin(headers, requestHeaders)
	assert.Equal(t, "https://google.com", headers.Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "Origin", headers.Get("Vary"))

	headers = http.Header{}
	requestHeaders = http.Header{"Origin": {"https://systemglitch.me"}}
	options.configureOrigin(headers, requestHeaders)
	assert.Empty(t, headers.Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "Origin", headers.Get("Vary"))
}

func TestConfigureCommon(t *testing.T) {
	options := Default()
	options.AllowCredentials = true
	options.AllowedOrigins = []string{"https://google.com", "https://images.google.com"}
	options.ExposedHeaders = []string{"Accept", "Content-Type"}
	headers := http.Header{}
	requestHeaders := http.Header{"Origin": {"https://images.google.com"}}

	options.ConfigureCommon(headers, requestHeaders)
	assert.Equal(t, "https://images.google.com", headers.Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "Origin", headers.Get("Vary"))
	assert.Equal(t, "true", headers.Get("Access-Control-Allow-Credentials"))
	assert.Equal(t, "Accept, Content-Type", headers.Get("Access-Control-Expose-Headers"))
}

func TestPreflight(t *testing.T) {
	options := Default()
	options.AllowedMethods = []string{http.MethodGet, http.MethodPut}
	options.MaxAge = 42 * time.Second
	headers := http.Header{}
	requestHeaders := http.Header{}

	options.HandlePreflight(headers, requestHeaders)
	assert.Equal(t, "GET, PUT", headers.Get("Access-Control-Allow-Methods"))
	assert.Equal(t, "Origin, Accept, Content-Type, X-Requested-With, Authorization", headers.Get("Access-Control-Allow-Headers"))
	assert.Equal(t, "42", headers.Get("Access-Control-Max-Age"))
}
