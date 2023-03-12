package goyave

import (
	"log"

	"gorm.io/gorm"
	"goyave.dev/goyave/v4/config"
	"goyave.dev/goyave/v4/lang"
)

type Composable interface {
	Init(*Server)
	Server() *Server
	DB() *gorm.DB
	Config() *config.Config
	Lang() *lang.Languages
	Logger() *log.Logger
	ErrLogger() *log.Logger
	AccessLogger() *log.Logger
	// TODO plugins? ability to access an instance-scoped service/system from this
	// Plugin(name string) (any, ok)
	// RequirePlugin(name string) any (or panics)
}

type Registrer interface {
	Composable
	RegisterRoutes(*RouterV5)
}

type Component struct {
	server *Server
}

var _ Composable = (*Component)(nil)

func (c *Component) Init(server *Server) {
	c.server = server
}

func (c *Component) Server() *Server {
	return c.server
}

func (c *Component) Logger() *log.Logger {
	return c.server.Logger
}

func (c *Component) ErrLogger() *log.Logger {
	return c.server.ErrLogger
}

func (c *Component) AccessLogger() *log.Logger {
	return c.server.AccessLogger
}

func (c *Component) DB() *gorm.DB {
	return c.server.DB().Session(&gorm.Session{NewDB: true})
}

func (c *Component) Config() *config.Config {
	return c.server.Config()
}

func (c *Component) Lang() *lang.Languages {
	return c.server.Lang
}
