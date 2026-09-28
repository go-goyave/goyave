package sqlite

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.43.0"
	"gorm.io/driver/sqlite"
	"goyave.dev/goyave/v5/database"
)

func TestDialect(t *testing.T) {
	dialect := &Dialect{}
	mockDB := &sql.DB{}
	dialector := dialect.Open(mockDB)
	require.NotNil(t, dialector)
	implDialector, ok := dialector.(*sqlite.Dialector)
	require.True(t, ok)
	assert.Same(t, mockDB, implDialector.Conn)

	cfg := database.DSNConfig{
		DatabaseName: "test_db",
		Options:      "mode=memory",
		// Below options should be ignored, not relevant for SQLite
		Host:     "127.0.0.1",
		Port:     5431,
		Username: "johndoe",
		Password: "secret",
	}
	assert.Equal(t, "file:test_db?mode=memory", dialect.DSN(cfg))

	assert.Equal(t, []attribute.KeyValue{semconv.DBSystemNameSQLite}, dialect.Attributes())
}
