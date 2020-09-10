package mssql

import (
	"github.com/System-Glitch/goyave/v3/database"
	"gorm.io/driver/sqlserver"
)

func init() {
	database.RegisterDialect("mssql", "sqlserver://{username}:{password}@{host}:{port}?database={name}&{options}", sqlserver.Open)
}
