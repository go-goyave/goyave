package clickhouse

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.43.0"
	"gorm.io/driver/clickhouse"
	"goyave.dev/goyave/v6/database"
)

func TestDialect(t *testing.T) {
	dialect := &Dialect{}
	mockDB := &sql.DB{}
	dialector := dialect.Open(mockDB)
	require.NotNil(t, dialector)
	implDialector, ok := dialector.(*clickhouse.Dialector)
	require.True(t, ok)
	assert.Same(t, mockDB, implDialector.Conn)

	cfg := database.DSNConfig{
		Host:         "127.0.0.1",
		Port:         8123,
		DatabaseName: "test_db",
		Username:     "johndoe",
		Password:     "secret",
		Options:      "ssl_mode=disable",
	}
	assert.Equal(t, "clickhouse://johndoe:secret@127.0.0.1:8123/test_db?ssl_mode=disable", dialect.DSN(cfg))

	assert.Equal(t, []attribute.KeyValue{semconv.DBSystemNameClickHouse}, dialect.Attributes())
}
