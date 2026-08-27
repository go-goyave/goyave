package config

import v "goyave.dev/goyave/v5/validation"

// Base the base configuration for built-in features.
// Can be embedded into a custom config structure.
type Base struct {
	App      App
	Database []DatabaseConnection
	Server   Server
}

func (s Base) RuleSet() v.RuleSet {
	return v.RuleSet{
		{Path: v.CurrentElement, Rules: v.List{v.Required(), v.WithMessage(v.Object(), "config.root-object-validation")}},
		{Path: "App", Rules: s.App.RuleSet()},
		{Path: "Server", Rules: s.Server.RuleSet()},
		{Path: "Database", Rules: v.List{v.Required(), v.Array()}},
		{Path: "Database[]", Rules: DatabaseConnection{}.RuleSet()},
	}
}

func (s Base) Default() Base {
	return Base{
		App:      s.App.Default(),
		Server:   s.Server.Default(),
		Database: []DatabaseConnection{},
	}
}

// App the general application details.
type App struct {
	Name        string // Not used by the framework but is a very common entry
	Environment string // Not used by the framework but is a very common entry

	// DefaultLanguage the name of the language to use by default
	// for logs and for translated responses.
	// Defaults to "en-US".
	DefaultLanguage string

	// Debug when true
	//  - the logger will use a human-readable log formatter
	//    for structured logs instead of JSON.
	//  - error details will be sent in the HTTP responses.
	// This setting should be set to false in production.
	Debug bool
}

func (s App) RuleSet() v.RuleSet {
	return v.RuleSet{
		{Path: v.CurrentElement, Rules: v.List{v.Required(), v.Object()}},
		{Path: "Name", Rules: v.List{v.Required(), v.String()}},
		{Path: "Environment", Rules: v.List{v.Required(), v.String()}},
		{Path: "DefaultLanguage", Rules: v.List{v.Required(), v.String()}},
	}
}

func (s App) Default() App {
	return App{
		Name:            "goyave",
		Environment:     "localhost",
		DefaultLanguage: "en-US",
		Debug:           true,
	}
}

// Server the HTTP server's configuration.
type Server struct {
	// Host on which the server listens.
	// Use "0.0.0.0" or "::" to listen on any address.
	// Defaults to "::1" (localhost).
	Host string
	// Domain is used for URL generation. Leave empty to use
	// IP instead when generating URLs pointing to the server.
	Domain string
	Proxy  Proxy
	// MaxUploadSize maximum size of the request, in MiB. Used
	// by the parse middleware.
	MaxUploadSize float64
	// Port number. If set to 0, an available port number is automatically chosen.
	// It can be retrieved with `server.Port()`. The chosen port is not reflected in
	// the configuration.
	Port int
	// WriteTimeoutMs corresponds to `http.Server.WriteTimeout` (in milliseconds).
	WriteTimeoutMs int
	// ReadTimeoutMs corresponds to `http.Server.ReadTimeout` (in milliseconds).
	ReadTimeoutMs int
	// ReadHeaderTimeoutMs corresponds to `http.Server.ReadHeaderTimeout` (in milliseconds).
	ReadHeaderTimeoutMs int
	// IdleTimeoutMs corresponds to `http.Server.IdleTimeout` (in milliseconds).
	IdleTimeoutMs int
	// WebsocketCloseTimeoutMs represents the maximum time allowed for the websocket
	// close handshake (in milliseconds).
	WebsocketCloseTimeoutMs int
}

func (s Server) RuleSet() v.RuleSet {
	return v.RuleSet{
		{Path: v.CurrentElement, Rules: v.List{v.Required(), v.Object()}},
		{Path: "Host", Rules: v.List{v.Required(), v.String()}},
		{Path: "Domain", Rules: v.List{v.String()}},
		{Path: "Port", Rules: v.List{v.Required(), v.Int(), v.Between(0, 65535)}},
		{Path: "WriteTimeoutMs", Rules: v.List{v.Required(), v.Int(), v.Min(0)}},
		{Path: "ReadTimeoutMs", Rules: v.List{v.Required(), v.Int(), v.Min(0)}},
		{Path: "ReadHeaderTimeoutMs", Rules: v.List{v.Required(), v.Int(), v.Min(0)}},
		{Path: "IdleTimeoutMs", Rules: v.List{v.Required(), v.Int(), v.Min(0)}},
		{Path: "WebsocketCloseTimeoutMs", Rules: v.List{v.Required(), v.Int(), v.Min(0)}},
		{Path: "MaxUploadSize", Rules: v.List{v.Required(), v.Float64(), v.Min(0)}},
		{Path: "Proxy", Rules: s.Proxy.RuleSet()},
	}
}

func (s Server) Default() Server {
	return Server{
		Host:                    "::1",
		Domain:                  "",
		Port:                    8080,
		WriteTimeoutMs:          10000, // 10s
		ReadTimeoutMs:           10000, // 10s
		ReadHeaderTimeoutMs:     10000, // 10s
		IdleTimeoutMs:           20000, // 20s
		WebsocketCloseTimeoutMs: 10000, // 10s
		MaxUploadSize:           10,    // 10MiB
		Proxy:                   s.Proxy.Default(),
	}
}

// Proxy configuration for URL generation when your application is running behind a reverse proxy
// such as apache or nginx. These entries don't have any impact on networking and are not required.
type Proxy struct {
	// Protocol http or https
	Protocol string
	// Host public host or domain of your proxy.
	Host string
	// Base the base path (usually starts with `/`)
	Base string
	Port int
}

func (Proxy) RuleSet() v.RuleSet {
	return v.RuleSet{
		{Path: v.CurrentElement, Rules: v.List{v.Object()}},
		{Path: "Protocol", Rules: v.List{v.String(), v.In([]string{"http", "https"})}},
		{Path: "Host", Rules: v.List{v.String()}},
		{Path: "Port", Rules: v.List{v.Int(), v.Between(0, 65535)}},
		{Path: "Base", Rules: v.List{v.String()}},
	}
}

func (Proxy) Default() Proxy {
	return Proxy{
		Protocol: "http",
		Host:     "",
		Port:     80,
		Base:     "",
	}
}

// DatabaseConnection configuration for a single database connection.
type DatabaseConnection struct {
	// Dialect the name of the SQL dialect (e.g.: "postgres")
	Dialect      string
	Host         string
	DatabaseName string
	Username     string
	Password     string
	// Options passed to the DSN when creating the connection.
	Options string

	GORM GORM

	Port int
	// MaxOpenConnections the maximum number of open connections to the database.
	// If equal to 0, there is no limit on the number of connections.
	// Recommended default value is 20.
	MaxOpenConnections int
	// MaxIdleConnections the maximum number of connections in the idle connection pool.
	// If equal to 0, no idle connections are retained.
	// Recommended default value is 20.
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
}

func (s DatabaseConnection) RuleSet() v.RuleSet {
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
		{Path: "GORM", Rules: s.GORM.RuleSet()},
	}
}

func (s DatabaseConnection) Default() DatabaseConnection {
	return DatabaseConnection{
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
		GORM:                       s.GORM.Default(),
	}
}

// GORM the settings for GORM, matching `gorm.Config`.
// Note that these settings can be manually modified from code after the connection is open.
type GORM struct {
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

func (GORM) RuleSet() v.RuleSet {
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

func (GORM) Default() GORM {
	return GORM{
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
