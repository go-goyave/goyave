package goyave

import (
	"log"

	"gorm.io/gorm"
	"goyave.dev/goyave/v4/config"
	"goyave.dev/goyave/v4/lang"
)

type IController interface {
	setServer(*Server)
	Server() *Server
	DB() *gorm.DB
	Config() *config.Config
	Lang() *lang.Languages
}

type Registrer interface {
	IController
	RegisterRoutes(*RouterV5)
}

type Controller struct {
	server *Server
}

var _ IController = (*Controller)(nil)

func (c *Controller) setServer(server *Server) {
	c.server = server
}

func (c *Controller) Server() *Server {
	return c.server
}

func (c *Controller) Logger() *log.Logger {
	return c.server.Logger
}

func (c *Controller) ErrLogger() *log.Logger {
	return c.server.ErrLogger
}

func (c *Controller) AccessLogger() *log.Logger {
	return c.server.AccessLogger
}

func (c *Controller) DB() *gorm.DB {
	return c.server.DB()
}

func (c *Controller) Config() *config.Config {
	return c.server.Config()
}

func (c *Controller) Lang() *lang.Languages {
	return c.server.Lang
}
