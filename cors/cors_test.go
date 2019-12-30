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
	assert.Equal(t, headers.Get("Access-Control-Max-Age"), "43200")

	options.MaxAge = 5 * time.Second
	options.configureMaxAge(headers)
	assert.Equal(t, headers.Get("Access-Control-Max-Age"), "5")

	options.MaxAge = 6*time.Second + 500*time.Millisecond
	options.configureMaxAge(headers)
	assert.Equal(t, headers.Get("Access-Control-Max-Age"), "6")
}

func TestConfigureAllowedHeaders(t *testing.T) {
	options := Default()
	headers := http.Header{}
	requestHeaders := http.Header{}

	options.configureAllowedHeaders(headers, requestHeaders)
	assert.Equal(t, headers.Get("Access-Control-Allow-Headers"), "Origin, Accept, Content-Type, X-Requested-With, Authorization")

	options.AllowedHeaders = []string{}
	requestHeaders.Set("Access-Control-Request-Headers", "Accept, Origin")

	options.configureAllowedHeaders(headers, requestHeaders)
	assert.Equal(t, headers.Get("Access-Control-Allow-Headers"), "Accept, Origin")
	assert.Equal(t, headers.Get("Vary"), "Access-Control-Request-Headers")
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

}

func TestPreflight(t *testing.T) {

}
