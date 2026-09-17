package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefault(t *testing.T) {
	t.Run("Base", func(t *testing.T) {
		var _ Section = &Base{}
		var _ DefaultValuer[Base] = &Base{}
		assert.NotEmpty(t, (&Base{}).RuleSet())
		assert.Equal(t, Base{
			App:    App{}.Default(),
			Server: Server{}.Default(),
		}, (&Base{}).Default())
	})

	t.Run("App", func(t *testing.T) {
		var _ Section = &App{}
		var _ DefaultValuer[App] = &App{}
		assert.NotEmpty(t, (&App{}).RuleSet())
		assert.Equal(t, App{
			Name:            "goyave",
			Environment:     "localhost",
			DefaultLanguage: "en-US",
			Debug:           true,
		}, (&App{}).Default())
	})

	t.Run("Server", func(t *testing.T) {
		var _ Section = &Server{}
		var _ DefaultValuer[Server] = &Server{}
		assert.NotEmpty(t, (&App{}).RuleSet())
		assert.Equal(t, Server{
			Host:                    "::1",
			Domain:                  "",
			Port:                    8080,
			WriteTimeoutMs:          10000, // 10s
			ReadTimeoutMs:           10000, // 10s
			ReadHeaderTimeoutMs:     10000, // 10s
			IdleTimeoutMs:           20000, // 20s
			WebsocketCloseTimeoutMs: 10000, // 10s
			MaxUploadSize:           10,    // 10MiB
			Proxy:                   Proxy{}.Default(),
		}, (&Server{}).Default())
	})

	t.Run("Proxy", func(t *testing.T) {
		var _ Section = &Proxy{}
		var _ DefaultValuer[Proxy] = &Proxy{}
		assert.NotEmpty(t, (&Proxy{}).RuleSet())
		assert.Equal(t, Proxy{
			Protocol: "http",
			Host:     "",
			Port:     80,
			Base:     "",
		}, (&Proxy{}).Default())
	})
}
