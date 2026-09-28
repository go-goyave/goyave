package database

import (
	"github.com/XSAM/otelsql"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	v "goyave.dev/goyave/v5/validation"
)

type DSNConfig struct {
	Host         string
	DatabaseName string
	Username     string
	Password     string

	// Options additional options passed to the DSN when creating the connection.
	Options string

	Port int
}

// Config configuration for a single database connection.
type Config struct {

	// Dialect the name of the SQL dialect (e.g.: "postgres")
	Dialect string

	DSNConfig

	GORM GORMConfig // TODO support config for other ORMs (create ORM adapters: https://github.com/go-goyave/goyave/discussions/285)

	// MaxOpenConnections the maximum number of open connections to the database.
	// If equal to 0, there is no limit on the number of connections.
	// Recommended default value is 20.
	MaxOpenConnections int
	// MaxIdleConnections the maximum number of connections in the idle connection pool.
	// If equal to 0, no idle connections are retained.
	// Recommended default value is 20.
	//
	// In tests using the database (mocked or not), it is recommended to set this value
	// to at least one; otherwise the test connection would be closed.
	MaxIdleConnections int
	// MaxLifetime the maximum time (in seconds) a connection may be reused.
	// Expired connections may be closed lazily before reuse.
	// If equal to 0, connections are not closed due to a connection's age.
	// Recommended default value is 300s.
	MaxLifetime int
	// MaxIdleTime the maximum time (in seconds) a connection may be idle.
	// Expired connections may be closed lazily before reuse.
	// If equal to 0, connections are not closed due to a connection's idle time.
	// Recommended default value is 0.
	MaxIdleTime int
	// DefaultReadQueryTimeoutMs the maximum execution time for read queries (in milliseconds).
	// Recommended default value is 20000ms.
	DefaultReadQueryTimeoutMs int
	// DefaultWriteQueryTimeoutMs the maximum execution time for write queries (in milliseconds).
	// Recommended default value is 40000ms.
	DefaultWriteQueryTimeoutMs int

	// Debug true to log all queries.
	Debug bool
}

func (s Config) RuleSet() v.RuleSet {
	return v.RuleSet{
		{Path: v.CurrentElement, Rules: v.List{v.Object()}},
		{Path: "Dialect", Rules: v.List{v.Required(), v.String()}},
		{Path: "Host", Rules: v.List{v.Required(), v.String()}},
		{Path: "Port", Rules: v.List{v.Required(), v.Int(), v.Between(0, 65535)}},
		{Path: "DatabaseName", Rules: v.List{v.Required(), v.String()}},
		{Path: "Username", Rules: v.List{v.String()}}, // Username and password not required because some drivers don't need it (e.g.: sqlite)
		{Path: "Password", Rules: v.List{v.String()}},
		{Path: "Options", Rules: v.List{v.String()}},
		{Path: "MaxOpenConnections", Rules: v.List{v.Required(), v.Int(), v.Min(0)}},
		{Path: "MaxIdleConnections", Rules: v.List{v.Required(), v.Int(), v.Min(0)}},
		{Path: "MaxLifetime", Rules: v.List{v.Required(), v.Int(), v.Min(0)}},
		{Path: "MaxIdleTime", Rules: v.List{v.Required(), v.Int(), v.Min(0)}},
		{Path: "DefaultReadQueryTimeoutMs", Rules: v.List{v.Required(), v.Int(), v.Min(0)}},
		{Path: "DefaultWriteQueryTimeoutMs", Rules: v.List{v.Required(), v.Int(), v.Min(0)}},
		{Path: "Debug", Rules: v.List{v.Required(), v.Bool()}},
		{Path: "GORM", Rules: s.GORM.RuleSet()},
	}
}

func (s Config) Default() Config {
	return Config{
		Dialect:                    "",
		Host:                       "",
		DatabaseName:               "",
		Username:                   "",
		Password:                   "",
		Options:                    "",
		Port:                       0,
		MaxOpenConnections:         20,
		MaxIdleConnections:         20,
		MaxLifetime:                300,
		MaxIdleTime:                0,
		DefaultReadQueryTimeoutMs:  20000, // 20s
		DefaultWriteQueryTimeoutMs: 40000, // 40s
		Debug:                      true,
		GORM:                       s.GORM.Default(),
	}
}

// GORMConfig the settings for GORMConfig, matching `gorm.Config`.
// Note that these settings can be manually modified from code after the connection is open.
type GORMConfig struct {
	// PrepareStmtMaxSize the max size for the prepared statement cache
	PrepareStmtMaxSize int
	// PrepareStmtTTL the prepared statement cache TTL (in seconds)
	PrepareStmtTTL                           int
	CreateBatchSize                          int
	PrepareStmt                              bool
	DryRun                                   bool
	SkipDefaultTransaction                   bool
	DisableNestedTransaction                 bool
	AllowGlobalUpdate                        bool
	FullSaveAssociations                     bool
	QueryFields                              bool
	PropagateUnscoped                        bool
	TranslateError                           bool
	DisableAutomaticPing                     bool
	DisableForeignKeyConstraintWhenMigrating bool
	IgnoreRelationshipsWhenMigrating         bool
}

func (GORMConfig) RuleSet() v.RuleSet {
	return v.RuleSet{
		{Path: v.CurrentElement, Rules: v.List{v.Object()}},
		{Path: "PrepareStmtMaxSize", Rules: v.List{v.Int(), v.Min(0)}},
		{Path: "PrepareStmtTTL", Rules: v.List{v.Int(), v.Min(0)}},
		{Path: "CreateBatchSize", Rules: v.List{v.Int(), v.Min(0)}},
		{Path: "PrepareStmt", Rules: v.List{v.Bool()}},
		{Path: "DryRun", Rules: v.List{v.Bool()}},
		{Path: "SkipDefaultTransaction", Rules: v.List{v.Bool()}},
		{Path: "DisableNestedTransaction", Rules: v.List{v.Bool()}},
		{Path: "AllowGlobalUpdate", Rules: v.List{v.Bool()}},
		{Path: "FullSaveAssociations", Rules: v.List{v.Bool()}},
		{Path: "QueryFields", Rules: v.List{v.Bool()}},
		{Path: "PropagateUnscoped", Rules: v.List{v.Bool()}},
		{Path: "TranslateError", Rules: v.List{v.Bool()}},
		{Path: "DisableAutomaticPing", Rules: v.List{v.Bool()}},
		{Path: "DisableForeignKeyConstraintWhenMigrating", Rules: v.List{v.Bool()}},
		{Path: "IgnoreRelationshipsWhenMigrating", Rules: v.List{v.Bool()}},
	}
}

func (GORMConfig) Default() GORMConfig {
	return GORMConfig{
		PrepareStmtMaxSize:                       0,
		PrepareStmtTTL:                           0,
		CreateBatchSize:                          1000,
		PrepareStmt:                              true,
		DryRun:                                   false,
		SkipDefaultTransaction:                   false,
		DisableNestedTransaction:                 false,
		AllowGlobalUpdate:                        false,
		FullSaveAssociations:                     false,
		QueryFields:                              false,
		PropagateUnscoped:                        false,
		TranslateError:                           false,
		DisableAutomaticPing:                     false,
		DisableForeignKeyConstraintWhenMigrating: false,
		IgnoreRelationshipsWhenMigrating:         false,
	}
}

type options struct {
	otelMeterProvider metric.MeterProvider
	otelTraceProvier  trace.TracerProvider
	otelOptions       []otelsql.Option
}

func (o options) getOTelOptions(dialect Dialect) []otelsql.Option {
	if dialect == nil {
		return []otelsql.Option{}
	}
	return append([]otelsql.Option{
		otelsql.WithAttributes(dialect.Attributes()...),
		otelsql.WithMeterProvider(o.otelMeterProvider),
		otelsql.WithTracerProvider(o.otelTraceProvier),
	}, o.otelOptions...)
}

// Option defines a dynamic DB setting that cannot be set through
// serializable configuration ([Config]).
type Option func(o *options)

// WithMeterProvider enables DB stats metrics reporting with OpenTelemetry.
func WithMeterProvider(p metric.MeterProvider) Option {
	return func(o *options) {
		o.otelMeterProvider = p
	}
}

// WithTraceProvider enables SQL tracing with OpenTelemetry.
func WithTraceProvider(p trace.TracerProvider) Option {
	return func(o *options) {
		o.otelTraceProvier = p
	}
}

// WithOpenTelemetryOptions sets additional options for the OpenTelemetry SQL instrumentation.
func WithOpenTelemetryOptions(opts ...otelsql.Option) Option {
	return func(o *options) {
		o.otelOptions = opts
	}
}
