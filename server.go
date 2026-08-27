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
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	stderrors "errors"

	"github.com/samber/lo"
	"gorm.io/gorm"
	"goyave.dev/goyave/v5/config"
	"goyave.dev/goyave/v5/database"
	"goyave.dev/goyave/v5/lang"
	"goyave.dev/goyave/v5/slog"
	"goyave.dev/goyave/v5/util/errors"
	"goyave.dev/goyave/v5/util/fsutil"
	"goyave.dev/goyave/v5/util/fsutil/osfs"
	"goyave.dev/goyave/v5/validation"
)

// serverKey is a context key used to store the server instance into its base context.
type serverKey struct{}

// Options represent server creation options.
type Options struct {

	// Config used by the server and propagated to all its components.
	// If no configuration is provided, automatically load
	// the default configuration using the default configuration source.
	// TODO only needs App (debug and defaultLanguage) and Server config...
	Config *config.Base

	// Logger used by the server and propagated to all its components.
	// If no logger is provided in the options, uses the default logger.
	Logger *slog.Logger

	// LangFS the file system from which the language files
	// will be loaded. This file system is expected to contain
	// a `resources/lang` directory.
	// If not provided, uses `osfs.FS` as a default.
	LangFS fsutil.FS

	// HTTP2 configures HTTP/2 connections.
	HTTP2 *http.HTTP2Config

	// ListenConfig optionally specifies the configuration for the network listener.
	// If not provided, the default net.ListenConfig is used.
	// This can be useful for customizing keep-alives and other network-level settings
	// for optimal performance in large traffic scenarios.
	ListenConfig *net.ListenConfig

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

	// MaxHeaderValueCount controls the maximum number of header
	// values that the server is willing to parse from a request.
	// If zero, http.DefaultMaxHeaderValueCount is used.
	// Note that comma-separated values in a single header line are
	// counted once, while values sent as multiple header lines are
	// counted multiple times.
	MaxHeaderValueCount int

	// DisableClientPriority specifies whether client-specified priority, as
	// specified in RFC 9218, should be respected or not.
	//
	// This field only takes effect if using HTTP/2, and if no custom write
	// scheduler is defined for the HTTP/2 server. Otherwise, this field is a
	// no-op.
	//
	// If set to true, requests will be served in a round-robin manner, without
	// prioritization.
	DisableClientPriority bool
}

// Server the central component of a Goyave application.
type Server struct {
	server *http.Server
	config *config.Server
	debug  bool // TODO test this is setup on New
	Lang   *lang.Languages

	router *Router
	db     *gorm.DB

	services map[string]Service

	// Logger the logger for default output
	// Writes to stderr by default.
	// FIXME changing the logger doesn't change the one in the base context
	// Should we really store the logger here and leave it accessible? we can set it using options
	Logger *slog.Logger

	host         string
	baseURL      string
	proxyBaseURL string

	stopChannel chan struct{}
	sigChannel  chan os.Signal

	ctx           context.Context
	baseContext   func(net.Listener) context.Context
	listenConfig  *net.ListenConfig
	startupHooks  []func(*Server)
	shutdownHooks []func(*Server)

	port int

	state atomic.Uint32 // 0 -> created, 1 -> preparing, 2 -> ready, 3 -> stopped
}

// New create a new `Server` using the given options.
func New(opts Options) (*Server, error) {
	cfg := opts.Config

	if opts.Config == nil {
		var validationErrors *validation.Errors
		var err error
		cfg, validationErrors, err = config.Load[config.Base](context.Background(), config.Default()) // TODO detach config loading from server
		if err != nil {
			return nil, errors.New(err)
		}
		if validationErrors != nil {
			err = errors.New("configuration validation errors")
			slog.Default().Error(err, "errors", validationErrors) // TODO readability won't be great...
			return nil, err
		}
	}

	slogger := opts.Logger
	if slogger == nil {
		slogger = slog.New(slog.NewHandler(cfg.App.Debug, os.Stderr))
	}

	langFS := opts.LangFS
	if langFS == nil {
		langFS = &osfs.FS{}
	}

	languages := lang.New()
	languages.Default = cfg.App.DefaultLanguage
	if err := languages.LoadAllAvailableLanguages(langFS); err != nil {
		return nil, err
	}

	host := cfg.Server.Host
	port := cfg.Server.Port

	server := &Server{
		server: &http.Server{
			Addr:                  net.JoinHostPort(host, strconv.Itoa(port)),
			WriteTimeout:          time.Duration(cfg.Server.WriteTimeoutMs) * time.Millisecond,
			ReadTimeout:           time.Duration(cfg.Server.ReadTimeoutMs) * time.Millisecond,
			ReadHeaderTimeout:     time.Duration(cfg.Server.ReadHeaderTimeoutMs) * time.Millisecond,
			IdleTimeout:           time.Duration(cfg.Server.IdleTimeoutMs) * time.Millisecond,
			ConnState:             opts.ConnState,
			ConnContext:           opts.ConnContext,
			MaxHeaderBytes:        opts.MaxHeaderBytes,
			MaxHeaderValueCount:   opts.MaxHeaderValueCount,
			HTTP2:                 opts.HTTP2,
			DisableClientPriority: opts.DisableClientPriority,
		},
		ctx:           context.Background(),
		baseContext:   opts.BaseContext,
		listenConfig:  opts.ListenConfig,
		config:        &cfg.Server,
		debug:         cfg.App.Debug,
		services:      make(map[string]Service),
		Lang:          languages,
		stopChannel:   make(chan struct{}, 1),
		startupHooks:  []func(*Server){},
		shutdownHooks: []func(*Server){},
		host:          host,
		port:          port,
		Logger:        slogger,
	}
	server.server.BaseContext = server.internalBaseContext
	server.refreshURLs()
	server.server.ErrorLog = log.New(&errLogWriter{server: server}, "", 0)

	// TODO database connections could be created outside of New? they are only passed to repositories and have nothing to do with the server itself
	if len(cfg.Database) > 0 {
		db, err := database.New(&cfg.Database[0], lo.Ternary(cfg.App.Debug, func() *slog.Logger { return server.Logger }, nil))
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

func (s *Server) isIPv6(host string) bool {
	return strings.IndexByte(host, ':') >= 0
}

func (s *Server) getAddress() string {
	shouldShowPort := s.port != 80
	host := s.config.Domain
	if len(host) == 0 {
		host = s.config.Host
		switch host {
		case "0.0.0.0":
			host = "127.0.0.1"
		case "::":
			host = "::1"
		}
	}

	if shouldShowPort {
		return "http://" + net.JoinHostPort(host, strconv.Itoa(s.port))
	}

	if s.isIPv6(host) {
		host = "[" + host + "]"
	}

	return "http://" + host
}

func (s *Server) getProxyAddress() string {
	if len(s.config.Proxy.Host) == 0 {
		return s.getAddress()
	}

	var shouldShowPort bool
	proto := s.config.Proxy.Protocol
	port := s.config.Proxy.Port
	if proto == "https" {
		shouldShowPort = port != 443
	} else {
		shouldShowPort = port != 80
	}
	host := s.config.Proxy.Host
	if shouldShowPort {
		host = net.JoinHostPort(host, strconv.Itoa(s.port))
	} else if s.isIPv6(host) {
		host = "[" + host + "]"
	}

	return proto + "://" + host + s.config.Proxy.Base
}

func (s *Server) refreshURLs() {
	s.baseURL = s.getAddress()
	s.proxyBaseURL = s.getProxyAddress()
}

// Service returns the service identified by the given name.
// Panics if no service could be found with the given name.
// TODO re-assess the service dependency container (does it belong here, in the Server?)
// TODO create a simple struct "ServiceContainer" or something that would be passed to the route registration function?
// The only "challenge" remaining is figuring out a clean way to access the services when creating controllers
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

// RegisterService on this server using its name (returned by `Service.Name()`).
// A service's name should be unique.
// `Service.Init(server)` is called on the given service upon registration.
// TODO remove service container
func (s *Server) RegisterService(service Service) {
	s.services[service.Name()] = service
}

// Host returns the hostname and port the server is running on.
func (s *Server) Host() string {
	return net.JoinHostPort(s.host, strconv.Itoa(s.port))
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

func (s *Server) HasDB() bool { // TODO Detach DB from server
	return s.db != nil
}

// DB returns the root database instance. Panics if no
// database connection is set up.
func (s *Server) DB() *gorm.DB { // TODO Detach DB from server
	if s.db == nil {
		panic(errors.NewSkip("no database connection", 3))
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
func (s *Server) Transaction(opts ...*sql.TxOptions) func() { // TODO Detach DB from server
	if s.db == nil {
		panic(errors.NewSkip("no database connection", 3))
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
func (s *Server) ReplaceDB(dialector gorm.Dialector) error { // TODO Detach DB from server
	if err := s.CloseDB(); err != nil {
		return err
	}

	db, err := database.NewFromDialector(nil, func() *slog.Logger { return s.Logger }, dialector)
	if err != nil {
		return err
	}

	s.db = db
	return nil
}

// CloseDB close the database connection if there is one.
// Does nothing and returns `nil` if there is no connection.
func (s *Server) CloseDB() error { // TODO Detach DB from server
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

	var ln net.Listener
	var err error
	if s.listenConfig != nil {
		ln, err = s.listenConfig.Listen(context.Background(), "tcp", s.server.Addr)
	} else {
		ln, err = net.Listen("tcp", s.server.Addr)
	}
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
	// Add the server and logger to the context
	// TODO document slogger added to base context
	loggerCtx := slog.Context(baseCtx, s.Logger)
	// TODO add debug too? but context starts to be overloaded...
	s.ctx = context.WithValue(loggerCtx, serverKey{}, s)

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
	if s.router.regexCache == nil {
		panic(errors.NewSkip("router's regex cache has already been cleared, did you call RegisterRoutes twice?", 3))
	}
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
