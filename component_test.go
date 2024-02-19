package goyave

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"goyave.dev/goyave/v5/config"

	_ "goyave.dev/goyave/v5/database/dialect/sqlite"
)

func TestComponent(t *testing.T) {
	cfg := config.LoadDefault()
	cfg.Set("database.connection", "sqlite3")
	cfg.Set("database.name", "test_component.db")
	cfg.Set("database.options", "mode=memory")
	server, err := New(Options{Config: cfg})
	require.NoError(t, err)
	defer func() {
		assert.NoError(t, server.CloseDB())
	}()
	service := &DummyService{}
	server.RegisterService(service)

	c := &Component{}
	c.Init(server)

	s, ok := c.LookupService("dummy")
	assert.Equal(t, service, s)
	assert.True(t, ok)

	s = c.Service("dummy")
	assert.Equal(t, service, s)

	s, ok = c.LookupService("not_a_service")
	assert.Nil(t, s)
	assert.False(t, ok)

	assert.Panics(t, func() {
		server.Service("not_a_service")
	})

	assert.Equal(t, server.Logger, c.Logger())
	assert.Equal(t, server.config, c.Config())
	assert.Equal(t, server.Lang, c.Lang())
	assert.Equal(t, server.db, c.DB())
}
