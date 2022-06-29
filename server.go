package goyave

import (
	"context"
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

	"gorm.io/gorm"
	"goyave.dev/goyave/v4/config"
	"goyave.dev/goyave/v4/database"
	"goyave.dev/goyave/v4/lang"
)

// Server the central component of a Goyave application.
type Server struct {
	server *http.Server
	config *config.Config
	Lang   *lang.Languages

	router *RouterV5
	db     *gorm.DB

	// TODO use a logging library?

	// Logger the logger for default output
	// Writes to stdout by default.
	Logger *log.Logger

	// AccessLogger the logger for access. This logger
	// is used by the logging middleware.
	// Writes to stdout by default.
	AccessLogger *log.Logger

	// ErrLogger the logger in which errors and stacktraces are written.
	// Writes to stderr by default.
	ErrLogger *log.Logger

	baseURL      string
	proxyBaseURL string

	stopChannel chan struct{}
	sigChannel  chan os.Signal

	startupHooks  []func(*Server)
	shutdownHooks []func(*Server)

	state uint32 // 0 -> created, 1 -> preparing, 2 -> ready, 3 -> stopped
}

// New create a new `Server` using automatically loaded configuration.
// See `config.Load()` for more details.
func New() (*Server, error) {
	cfg, err := config.LoadV5()
	if err != nil {
		return nil, &Error{err, ExitInvalidConfig}
	}
	return NewWithConfig(cfg)
}

// NewWithConfig create a new `Server` using the provided configuration.
func NewWithConfig(cfg *config.Config) (*Server, error) {
	timeout := time.Duration(cfg.GetInt("server.timeout")) * time.Second

	var db *gorm.DB
	var err error
	if cfg.GetString("database.connection") != "none" {
		db, err = database.New(cfg)
		if err != nil {
			return nil, &Error{err, ExitDatabaseError}
		}
	}

	errLogger := log.New(os.Stderr, "", log.LstdFlags)

	languages := lang.New()
	languages.Default = cfg.GetString("app.defaultLanguage")
	if err := languages.LoadAllAvailableLanguages(); err != nil {
		return nil, &Error{err, ExitLanguageError}
	}

	return &Server{
		server: &http.Server{
			Addr:         cfg.GetString("server.host") + ":" + strconv.Itoa(cfg.GetInt("server.port")),
			WriteTimeout: timeout,
			ReadTimeout:  timeout,
			IdleTimeout:  timeout * 2,
			Handler:      router,
			ErrorLog:     errLogger, // TODO what if it is replaced in the goyave.Server struct?
		},
		config:        cfg,
		db:            db,
		Lang:          languages,
		stopChannel:   make(chan struct{}, 1),
		startupHooks:  []func(*Server){},
		shutdownHooks: []func(*Server){},
		baseURL:       getAddressV5(cfg),
		proxyBaseURL:  getProxyAddress(cfg),
		Logger:        log.New(os.Stdout, "", log.LstdFlags),
		AccessLogger:  log.New(os.Stdout, "", 0),
		ErrLogger:     errLogger,
	}, nil
}

func getAddressV5(cfg *config.Config) string {
	port := cfg.GetInt("server.port")
	shouldShowPort := port != 80
	host := cfg.GetString("server.domain")
	if len(host) == 0 {
		host = cfg.GetString("server.host")
		if host == "0.0.0.0" {
			host = "127.0.0.1"
		}
	}

	if shouldShowPort {
		host += ":" + strconv.Itoa(port)
	}

	return "http://" + host
}

func getProxyAddress(cfg *config.Config) string {
	if !cfg.Has("server.proxy.host") {
		return getAddressV5(cfg)
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
	state := atomic.LoadUint32(&s.state)
	return state == 2
}

// RegisterStartupHook to execute some code once the server is ready and running.
// Startup hooks are executed in their own goroutine.
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

// DB returns a new gorm session.
func (s *Server) DB() *gorm.DB {
	if s.db == nil {
		s.ErrLogger.Panicf("No database connection. Database is set to \"none\" in the config")
	}
	return s.db.Session(&gorm.Session{NewDB: true})
}

// ReplaceDB manually replace the automatic DB connection.
// If a connection already exists, closes it before discarding it.
// This can be used to create a mock DB in tests. Using this function
// is not recommended outside of tests. Prefer using a custom dialect.
// This operation is not concurrently safe.
func (s *Server) ReplaceDB(dialector gorm.Dialector) error {
	if err := s.closeDB(); err != nil {
		return err
	}

	db, err := database.NewFromDialector(s.config, dialector)
	if err != nil {
		return err
	}

	s.db = db
	return nil
}

func (s *Server) closeDB() error {
	if s.db == nil {
		return nil
	}
	db, err := s.db.DB()
	if err != nil {
		return err
	}
	return db.Close()
}

// Start the server.
//
// The routeRegistrer parameter is a function aimed at registering all your routes and middleware.
//
// Auto-migrations are run first if they are enabled in the configuration.
//
// Errors returned can be safely type-asserted to `*goyave.Error`.
func (s *Server) Start(routeRegistrer func(*Server, *RouterV5)) error { // Give the routeRegister in New ? (would be before automigrations :/ )
	state := atomic.LoadUint32(&s.state)
	if state == 1 || state == 2 {
		return &Error{
			err:      fmt.Errorf("Server is already running"),
			ExitCode: ExitStateError,
		}
	} else if state == 3 {
		return &Error{
			err:      fmt.Errorf("Cannot restart a stopped server"),
			ExitCode: ExitStateError,
		}
	}
	atomic.StoreUint32(&s.state, 1)

	defer func() {
		atomic.StoreUint32(&s.state, 3)
		// Notify the shutdown is complete so Stop() can return
		s.stopChannel <- struct{}{}
		close(s.stopChannel)
	}()

	if err := s.prepare(routeRegistrer); err != nil {
		return err
	}

	ln, err := net.Listen("tcp", s.server.Addr)
	if err != nil {
		return &Error{err, ExitNetworkError}
	}
	defer func() {
		for _, hook := range s.shutdownHooks {
			hook(s)
		}
		if err := s.closeDB(); err != nil {
			s.ErrLogger.Println(err)
		}
	}()

	atomic.StoreUint32(&s.state, 2)

	for _, hook := range s.startupHooks {
		go func(f func(*Server)) {
			if s.IsReady() {
				f(s)
			}
		}(hook)
	}
	if err := s.server.Serve(ln); err != nil && err != http.ErrServerClosed {
		atomic.StoreUint32(&s.state, 3)
		return &Error{err, ExitHTTPError}
	}
	return nil
}

func (s *Server) prepare(routeRegistrer func(*Server, *RouterV5)) error { // TODO rename routeRegistrer to "initServer"?
	s.router = NewRouterV5(s)
	routeRegistrer(s, s.router)
	s.router.ClearRegexCache()
	s.server.Handler = s.router

	if s.config.GetBool("database.autoMigrate") && s.db != nil {
		if err := database.MigrateV5(s.DB()); err != nil {
			return &Error{err, ExitDatabaseError}
		}
	}
	return nil
}

// Stop gracefully shuts down the server without interrupting any
// active connections.
//
// `Stop()` does not attempt to close nor wait for hijacked
// connections such as WebSockets. The caller of Stop should
// separately notify such long-lived connections of shutdown and wait
// for them to close, if desired. This can be done using shutdown hooks.
//
// Make sure the program doesn't exit before `Stop()` returns.
//
// After being stopped, a `Server` is not meant to be re-used.
//
// This function can be called from any goroutine and is concurrently safe.
func (s *Server) Stop() {
	if s.sigChannel != nil {
		signal.Stop(s.sigChannel)
	}
	state := atomic.LoadUint32(&s.state)
	atomic.StoreUint32(&s.state, 3)
	if state == 0 {
		// Start has not been called, do nothing
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := s.server.Shutdown(ctx)
	if err != nil {
		s.ErrLogger.Println(err)
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
		select {
		case _, ok := <-s.sigChannel:
			if ok {
				s.Stop()
			}
		}
	}()
}
