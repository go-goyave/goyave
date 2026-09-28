package bigquery

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"gorm.io/driver/bigquery"
	"goyave.dev/goyave/v5/database"
)

func TestDialect(t *testing.T) {
	dialect := &Dialect{}
	mockDB := &sql.DB{}
	dialector := dialect.Open(mockDB)
	require.NotNil(t, dialector)
	implDialector, ok := dialector.(*bigquery.Dialector)
	require.True(t, ok)
	assert.Same(t, mockDB, implDialector.Conn)

	cfg := database.DSNConfig{
		DatabaseName: "projectID/eu-west1/dataset",
		Options:      "ssl_mode=disable",
		// Below options should be disabled
		Host:     "127.0.0.1",
		Port:     5431,
		Username: "johndoe",
		Password: "secret",
	}
	assert.Equal(t, "bigquery://projectID/eu-west1/dataset?ssl_mode=disable", dialect.DSN(cfg))

	assert.Equal(t, []attribute.KeyValue{dbSystemNameBigquery}, dialect.Attributes())
}
