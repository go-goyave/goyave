package goyave

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync/atomic"
	"syscall"
	"time"

	stderrors "errors"

	"gorm.io/gorm"
	"goyave.dev/goyave/v5/config"
	"goyave.dev/goyave/v5/database"
	"goyave.dev/goyave/v5/lang"
	"goyave.dev/goyave/v5/slog"
	"goyave.dev/goyave/v5/util/errors"
	"goyave.dev/goyave/v5/util/fsutil"
	"goyave.dev/goyave/v5/util/fsutil/osfs"
)

// serverKey is a context key used to store the server instance into its base context.
type serverKey struct{}

// Options represent server creation options.
type Options struct {

	// Config used by the server and propagated to all its components.
	// If no configuration is provided, automatically load
	// the default configuration using `config.Load()`.
	Config *config.Config

	// Logger used by the server and propagated to all its components.
	// If no logger is provided in the options, uses the default logger.
	Logger *slog.Logger

	// LangFS the file system from which the language files
	// will be loaded. This file system is expected to contain
	// a `resources/lang` directory.
	// If not provided, uses `osfs.FS` as a default.
	LangFS fsutil.FS

	// ConnState specifies an optional callback function that is
	// called when a client connection changes state. See the
	// `http.ConnState` type and associated constants for details.
	ConnState func(net.Conn, http.ConnState)

	// Context optionnally defines a function that returns the base context
	// for the server. It will be used as base context for all incoming requests.
	//
	// The provided `net.Listener` is the specific Listener that's
	// about to start accepting requests.
	//
	// If not given, the default is `context.Background()`.
	//
	// The context returned then has a the server instance added to it as a value.
	// The server can thus be retrieved using `goyave.ServerFromContext(ctx)`.
	//
	// If the context is canceled, the server won't shut down automatically, you are
	// responsible of calling `server.Stop()` if you want this to happen. Otherwise the
	// server will continue serving requests, at the risk of generating "context canceled" errors.
	BaseContext func(net.Listener) context.Context

	// ConnContext optionally specifies a function that modifies
	// the context used for a new connection `c`. The provided context
	// is derived from the base context and has the server instance value, which can
	// be retrieved using `goyave.ServerFromContext(ctx)`.
	ConnContext func(ctx context.Context, c net.Conn) context.Context

	// MaxHeaderBytes controls the maximum number of bytes the
	// server will read parsing the request header's keys and
	// values, including the request line. It does not limit the
	// size of the request body.
	// If zero, http.DefaultMaxHeaderBytes is used.
	MaxHeaderBytes int
}

// Server the central component of a Goyave application.
type Server struct {
	server *http.Server
	config *config.Config
	Lang   *lang.Languages

	router *Router
	db     *gorm.DB

	services map[string]Service

	// Logger the logger for default output
	// Writes to stderr by default.
	Logger *slog.Logger

	host         string
	baseURL      string
	proxyBaseURL string

	stopChannel chan struct{}
	sigChannel  chan os.Signal

	ctx           context.Context
	baseContext   func(net.Listener) context.Context
	startupHooks  []func(*Server)
	shutdownHooks []func(*Server)

	port int

	state atomic.Uint32 // 0 -> created, 1 -> preparing, 2 -> ready, 3 -> stopped
}

// New create a new `Server` using the given options.
func New(opts Options) (*Server, error) {
	cfg := opts.Config

	if opts.Config == nil {
		var err error
		cfg, err = config.Load()
		if err != nil {
			return nil, errors.New(err)
		}
	}

	slogger := opts.Logger
	if slogger == nil {
		slogger = slog.New(slog.NewHandler(cfg.GetBool("app.debug"), os.Stderr))
	}

	langFS := opts.LangFS
	if langFS == nil {
		langFS = &osfs.FS{}
	}

	languages := lang.New()
	languages.Default = cfg.GetString("app.defaultLanguage")
	if err := languages.LoadAllAvailableLanguages(langFS); err != nil {
		return nil, err
	}

	port := cfg.GetInt("server.port")
	host := cfg.GetString("server.host") + ":" + strconv.Itoa(port)

	server := &Server{
		server: &http.Server{
			Addr:              host,
			WriteTimeout:      time.Duration(cfg.GetInt("server.writeTimeout")) * time.Second,
			ReadTimeout:       time.Duration(cfg.GetInt("server.readTimeout")) * time.Second,
			ReadHeaderTimeout: time.Duration(cfg.GetInt("server.readHeaderTimeout")) * time.Second,
			IdleTimeout:       time.Duration(cfg.GetInt("server.idleTimeout")) * time.Second,
			ConnState:         opts.ConnState,
			ConnContext:       opts.ConnContext,
			MaxHeaderBytes:    opts.MaxHeaderBytes,
		},
		ctx:           context.Background(),
		baseContext:   opts.BaseContext,
		config:        cfg,
		services:      make(map[string]Service),
		Lang:          languages,
		stopChannel:   make(chan struct{}, 1),
		startupHooks:  []func(*Server){},
		shutdownHooks: []func(*Server){},
		host:          cfg.GetString("server.host"),
		port:          port,
		Logger:        slogger,
	}
	server.server.BaseContext = server.internalBaseContext
	server.refreshURLs()
	server.server.ErrorLog = log.New(&errLogWriter{server: server}, "", 0)

	if cfg.GetString("database.connection") != "none" {
		db, err := database.New(cfg, func() *slog.Logger { return server.Logger })
		if err != nil {
			return nil, errors.New(err)
		}
		server.db = db
	}

	server.router = NewRouter(server)
	server.server.Handler = server.router
	return server, nil
}

func (s *Server) internalBaseContext(_ net.Listener) context.Context {
	return s.ctx
}

func (s *Server) getAddress(cfg *config.Config) string {
	shouldShowPort := s.port != 80
	host := cfg.GetString("server.domain")
	if len(host) == 0 {
		host = cfg.GetString("server.host")
		if host == "0.0.0.0" {
			host = "127.0.0.1"
		}
	}

	if shouldShowPort {
		host += ":" + strconv.Itoa(s.port)
	}

	return "http://" + host
}

func (s *Server) getProxyAddress(cfg *config.Config) string {
	if !cfg.Has("server.proxy.host") {
		return s.getAddress(cfg)
	}

	var shouldShowPort bool
	proto := cfg.GetString("server.proxy.protocol")
	port := cfg.GetInt("server.proxy.port")
	if proto == "https" {
		shouldShowPort = port != 443
	} else {
		shouldShowPort = port != 80
	}
	host := cfg.GetString("server.proxy.host")
	if shouldShowPort {
		host += ":" + strconv.Itoa(port)
	}

	return proto + "://" + host + cfg.GetString("server.proxy.base")
}

func (s *Server) refreshURLs() {
	s.baseURL = s.getAddress(s.config)
	s.proxyBaseURL = s.getProxyAddress(s.config)
}

// Service returns the service identified by the given name.
// Panics if no service could be found with the given name.
func (s *Server) Service(name string) Service {
	if s, ok := s.services[name]; ok {
		return s
	}
	panic(errors.Errorf("service %q does not exist", name))
}

// LookupService search for a service by its name. If the service
// identified by the given name exists, it is returned with the `true` boolean.
// Otherwise returns `nil` and `false`.
func (s *Server) LookupService(name string) (Service, bool) {
	service, ok := s.services[name]
	return service, ok
}

// RegisterService on thise server using its name (returned by `Service.Name()`).
// A service's name should be unique.
// `Service.Init(server)` is called on the given service upon registration.
func (s *Server) RegisterService(service Service) {
	s.services[service.Name()] = service
}

// Host returns the hostname and port the server is running on.
func (s *Server) Host() string {
	return s.host + ":" + strconv.Itoa(s.port)
}

// Port returns the port the server is running on.
func (s *Server) Port() int {
	return s.port
}

// BaseURL returns the base URL of your application.
// If "server.domain" is set in the config, uses it instead
// of an IP address.
func (s *Server) BaseURL() string {
	return s.baseURL
}

// ProxyBaseURL returns the base URL of your application based on the "server.proxy" configuration.
// This is useful when you want to generate an URL when your application is served behind a reverse proxy.
// If "server.proxy.host" configuration is not set, returns the same value as "BaseURL()".
func (s *Server) ProxyBaseURL() string {
	return s.proxyBaseURL
}

// IsReady returns true if the server has finished initializing and
// is ready to serve incoming requests.
// This operation is concurrently safe.
func (s *Server) IsReady() bool {
	return s.state.Load() == 2
}

// RegisterStartupHook to execute some code once the server is ready and running.
// All startup hooks are executed in a single goroutine and in order of registration.
func (s *Server) RegisterStartupHook(hook func(*Server)) {
	s.startupHooks = append(s.startupHooks, hook)
}

// ClearStartupHooks removes all startup hooks.
func (s *Server) ClearStartupHooks() {
	s.startupHooks = []func(*Server){}
}

// RegisterShutdownHook to execute some code after the server stopped.
// Shutdown hooks are executed before `Start()` returns and are NOT executed
// in a goroutine, meaning that the shutdown process can be blocked by your
// shutdown hooks. It is your responsibility to implement a timeout mechanism
// inside your hook if necessary.
func (s *Server) RegisterShutdownHook(hook func(*Server)) {
	s.shutdownHooks = append(s.shutdownHooks, hook)
}

// ClearShutdownHooks removes all shutdown hooks.
func (s *Server) ClearShutdownHooks() {
	s.shutdownHooks = []func(*Server){}
}

// Config returns the server's config.
func (s *Server) Config() *config.Config {
	return s.config
}

// DB returns the root database instance. Panics if no
// database connection is set up.
func (s *Server) DB() *gorm.DB {
	if s.db == nil {
		panic(errors.NewSkip("No database connection. Database is set to \"none\" in the config", 3))
	}
	return s.db
}

// Transaction makes it so all DB requests are run inside a transaction.
//
// Returns the rollback function. When you are done, call this function to
// complete the transaction and roll it back. This will also restore the original
// DB so it can be used again out of the transaction.
//
// This is used for tests. This operation is not concurrently safe.
func (s *Server) Transaction(opts ...*sql.TxOptions) func() {
	if s.db == nil {
		panic(errors.NewSkip("No database connection. Database is set to \"none\" in the config", 3))
	}
	ogDB := s.db
	s.db = s.db.Begin(opts...)
	return func() {
		err := s.db.Rollback().Error
		s.db = ogDB
		if err != nil {
			panic(errors.New(err))
		}
	}
}

// ReplaceDB manually replace the automatic DB connection.
// If a connection already exists, closes it before discarding it.
// This can be used to create a mock DB in tests. Using this function
// is not recommended outside of tests. Prefer using a custom dialect.
// This operation is not concurrently safe.
func (s *Server) ReplaceDB(dialector gorm.Dialector) error {
	if err := s.CloseDB(); err != nil {
		return err
	}

	db, err := database.NewFromDialector(s.config, func() *slog.Logger { return s.Logger }, dialector)
	if err != nil {
		return err
	}

	s.db = db
	return nil
}

// CloseDB close the database connection if there is one.
// Does nothing and returns `nil` if there is no connection.
func (s *Server) CloseDB() error {
	if s.db == nil {
		return nil
	}
	db, err := s.db.DB()
	if err != nil {
		if stderrors.Is(err, gorm.ErrInvalidDB) {
			return nil
		}
		return errors.New(err)
	}
	return errors.New(db.Close())
}

// Router returns the root router.
func (s *Server) Router() *Router {
	return s.router
}

// Start the server. This operation is blocking and returns when the server is closed.
func (s *Server) Start() error {
	swapped := s.state.CompareAndSwap(0, 1)
	if !swapped {
		return errors.New("server was already started")
	}

	defer func() {
		s.state.Store(3)
		// Notify the shutdown is complete so Stop() can return
		s.stopChannel <- struct{}{}
		close(s.stopChannel)
	}()

	ln, err := net.Listen("tcp", s.server.Addr)
	if err != nil {
		return errors.New(err)
	}
	baseCtx := context.Background()
	if s.baseContext != nil {
		baseCtx = s.baseContext(ln)
		if baseCtx == nil {
			panic("server options BaseContext returned a nil context")
		}
	}
	s.ctx = context.WithValue(baseCtx, serverKey{}, s)

	select {
	case <-s.ctx.Done():
		return errors.New("cannot start the server, context is canceled")
	default:
	}

	s.port = ln.Addr().(*net.TCPAddr).Port
	s.refreshURLs()
	defer func() {
		for _, hook := range s.shutdownHooks {
			hook(s)
		}
		if err := s.CloseDB(); err != nil {
			s.Logger.Error(err)
		}
	}()

	s.state.Store(2)

	go func(s *Server) {
		if s.IsReady() {
			// We check if the server is ready to prevent startup hook execution
			// if `Serve` returned an error before the goroutine started
			for _, hook := range s.startupHooks {
				hook(s)
			}
		}
	}(s)
	if err := s.server.Serve(ln); err != nil && !stderrors.Is(err, http.ErrServerClosed) {
		s.state.Store(3)
		return errors.New(err)
	}
	return nil
}

// RegisterRoutes runs the given `routeRegistrer` function with this Server and its router.
// The router's regex cache is cleared after the `routeRegistrer` function returns.
// This method should only be called once.
func (s *Server) RegisterRoutes(routeRegistrer func(*Server, *Router)) {
	routeRegistrer(s, s.router)
	s.router.ClearRegexCache()
}

// Stop gracefully shuts down the server without interrupting any
// active connections.
//
// `Stop()` does not attempt to close nor wait for hijacked
// connections such as WebSockets. The caller of `Stop` should
// separately notify such long-lived connections of shutdown and wait
// for them to close, if desired. This can be done using shutdown hooks.
//
// If registered, the OS signal channel is closed.
//
// Make sure the program doesn't exit before `Stop()` returns.
//
// After being stopped, a `Server` is not meant to be re-used.
//
// This function can be called from any goroutine and is concurrently safe.
// Calling this function several times is safe. Calls after the first one are no-op.
func (s *Server) Stop() {
	state := s.state.Swap(3)
	if state == 0 || state == 3 {
		// Start has not been called or Stop has already been called, do nothing
		return
	}
	if s.sigChannel != nil {
		signal.Stop(s.sigChannel)
		close(s.sigChannel)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := s.server.Shutdown(ctx)
	if err != nil {
		s.Logger.Error(errors.NewSkip(err, 3))
	}

	<-s.stopChannel // Wait for stop channel before returning
}

// RegisterSignalHook creates a channel listening on SIGINT and SIGTERM. When receiving such
// signal, the server is stopped automatically and the listener on these signals is removed.
func (s *Server) RegisterSignalHook() {
	// Sometimes users may not want to have a sigChannel setup
	// also we don't want it in tests
	// users will have to manually call this function if they want the shutdown on signal feature

	s.sigChannel = make(chan os.Signal, 64)
	signal.Notify(s.sigChannel, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		_, ok := <-s.sigChannel
		if ok {
			s.Stop()
		}
	}()
}

// errLogWriter is a proxy io.Writer that pipes into the server logger.
// This is used so the error logger (type `*log.Logger`) of the underlying
// std HTTP server write to the same logger as the rest of the application.
type errLogWriter struct {
	server *Server
}

func (w errLogWriter) Write(p []byte) (n int, err error) {
	w.server.Logger.Error(fmt.Errorf("%s", p))
	return len(p), nil
}

// ServerFromContext returns the `*goyave.Server` stored in the given context or `nil`.
// This is safe to call using any context retrieved from incoming HTTP requests as this value
// is automatically injected when the server is created.
func ServerFromContext(ctx context.Context) *Server {
	s, ok := ctx.Value(serverKey{}).(*Server)
	if !ok {
		return nil
	}
	return s
}
