package clickhouse

import (
	"gorm.io/driver/clickhouse"
	"goyave.dev/goyave/v5/database"
)

func init() {
	database.RegisterDialect("clickhouse", "clickhouse://{username}:{password}@{host}:{port}/{name}?{options}", clickhouse.Open)
}
