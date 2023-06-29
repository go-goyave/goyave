package mssql

import (
	"gorm.io/driver/sqlserver"
	"goyave.dev/goyave/v5/database"
)

func init() {
	database.RegisterDialect("mssql", "sqlserver://{username}:{password}@{host}:{port}?database={name}&{options}", sqlserver.Open)
}
