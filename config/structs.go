package config

import v "goyave.dev/goyave/v5/validation"

// Base the base configuration for built-in features.
// Can be embedded into a custom config structure.
type Base struct {
	App    App
	Server Server
}

func (s Base) RuleSet() v.RuleSet {
	return v.RuleSet{
		{Path: v.CurrentElement, Rules: v.List{v.Required(), v.WithMessage(v.Object(), "config.root-object-validation")}},
		{Path: "App", Rules: s.App.RuleSet()},
		{Path: "Server", Rules: s.Server.RuleSet()},
	}
}

func (s Base) Default() Base {
	return Base{
		App:    s.App.Default(),
		Server: s.Server.Default(),
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
		{Path: "Debug", Rules: v.List{v.Required(), v.Bool()}},
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
	// TODO move this to the websocket config section?
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
