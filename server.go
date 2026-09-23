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
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	"errors"

	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"goyave.dev/goyave/v5/config"
	"goyave.dev/goyave/v5/internal/otel"
	"goyave.dev/goyave/v5/lang"
	"goyave.dev/goyave/v5/slog"
	"goyave.dev/goyave/v5/util/errwrap"
	"goyave.dev/goyave/v5/util/fsutil"
	"goyave.dev/goyave/v5/util/fsutil/osfs"
)

// serverKey is a context key used to store the server instance into its base context.
type serverKey struct{}

// OpenTelemetryOptions for enabling and tweaking the native OpenTelemetry integration.
type OpenTelemetryOptions struct {
	// TracerProvider if given, enables tracing using OpenTelemetry.
	TracerProvider trace.TracerProvider

	// MeterProvider if given, enables metric reporting using OpenTelemetry.
	MeterProvider metric.MeterProvider

	// TracePropagators if given, enables span propagation across this process's boundaries.
	TracePropagators []propagation.TextMapPropagator
}

// Options represent server creation options.
type Options struct {
	// Logger used by the server.
	// If no logger is provided in the options, a new [slog.Logger] outputting
	// to [os.Stderr] is created. The handler used depends on the [config.App.Debug] value.
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

	// Context optionally defines the server's root context.
	//
	// This context is enriched with the server's logger and the server instance.
	// The server can thus be retrieved using [ServerFromContext].
	//
	// If no given, defaults to [context.Background].
	Context context.Context

	// BaseContext optionally defines a function that returns the base context
	// for the server. It will be used as base context for all incoming requests.
	//
	// The `parent` parameter is the server's root context. See [Options.Context].
	//
	// The provided `net.Listener` is the specific Listener that's
	// about to start accepting requests.
	//
	// If not given, the default is the server's context.
	//
	// If the context is canceled, the server won't shut down automatically, you are
	// responsible of calling `server.Stop()` if you want this to happen. Otherwise the
	// server will continue serving requests, at the risk of generating "context canceled" errors.
	//
	// It is not recommended to return a context that is canceled when the server shutdown is requested
	// because that would prevent finishing to handle ongoing requests gracefully.
	// If you want to control shutdown this way, this function should use [context.WithoutCancel].
	BaseContext func(parent context.Context, ln net.Listener) context.Context

	// ConnContext optionally specifies a function that modifies
	// the context used for a new connection `c`. The provided context
	// is derived from the base context and has the server instance value, which can
	// be retrieved using `goyave.ServerFromContext(ctx)`.
	ConnContext func(ctx context.Context, c net.Conn) context.Context

	OpenTelemetry OpenTelemetryOptions

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
	Lang   *lang.Languages

	router *Router

	// logger the logger for default output
	// Writes to stderr by default.
	logger *slog.Logger

	meters *otel.HTTPServerMeters

	ctx context.Context

	tracer trace.Tracer

	host         string
	baseURL      string
	proxyBaseURL string

	stopChannel chan struct{}
	sigChannel  chan os.Signal

	baseContext   func(context.Context, net.Listener) context.Context
	listenConfig  *net.ListenConfig
	startupHooks  []func(*Server)
	shutdownHooks []func(*Server)

	port int

	state atomic.Uint32 // 0 -> created, 1 -> preparing, 2 -> ready, 3 -> stopped

	debug bool
}

// New create a new [*Server] using the given options.
func New(cfg *config.Base, opts Options) (*Server, error) {
	ctx := context.Background()
	if opts.Context != nil {
		ctx = opts.Context
	}

	slogger := opts.Logger
	if slogger == nil {
		slogger = slog.New(slog.NewHandler(cfg.App.Debug, os.Stderr)).WithContext(ctx)
	}
	ctx = slog.Context(ctx, slogger)

	langFS := opts.LangFS
	if langFS == nil {
		langFS = &osfs.FS{}
	}

	languages := lang.New()
	languages.Default = cfg.App.DefaultLanguage
	if err := languages.LoadAllAvailableLanguages(langFS); err != nil {
		return nil, errwrap.New(err)
	}

	var tracer trace.Tracer
	if opts.OpenTelemetry.TracerProvider != nil {
		tracer = otel.Tracer(opts.OpenTelemetry.TracerProvider)
	}

	var meters *otel.HTTPServerMeters
	if opts.OpenTelemetry.MeterProvider != nil {
		var err error
		meters, err = otel.NewHTTPServerMeter(otel.Meter(opts.OpenTelemetry.MeterProvider))
		if err != nil {
			return nil, errwrap.New(err)
		}
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
		baseContext:   opts.BaseContext,
		listenConfig:  opts.ListenConfig,
		config:        &cfg.Server,
		debug:         cfg.App.Debug,
		Lang:          languages,
		stopChannel:   make(chan struct{}, 1),
		startupHooks:  []func(*Server){},
		shutdownHooks: []func(*Server){},
		host:          host,
		port:          port,
		logger:        slogger,
		tracer:        tracer,
		meters:        meters,
	}
	server.ctx = context.WithValue(ctx, serverKey{}, server)
	server.server.BaseContext = server.internalBaseContext
	server.refreshURLs()
	server.server.ErrorLog = log.New(&errLogWriter{server: server}, "", 0)

	server.router = NewRouter(server)
	server.server.Handler = server.router
	return server, nil
}

func (s *Server) internalBaseContext(ln net.Listener) context.Context {
	ctx := s.ctx
	if s.baseContext != nil {
		ctx = s.baseContext(s.ctx, ln)
		if ctx == nil {
			panic("server options BaseContext returned a nil context")
		}
	}
	return ctx
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

// Host returns the hostname and port the server is running on.
func (s *Server) Host() string {
	return net.JoinHostPort(s.host, strconv.Itoa(s.port))
}

// Port returns the port the server is running on.
func (s *Server) Port() int {
	return s.port
}

// Context returns the server's root context, enriched with the server's logger and the server instance.
func (s *Server) Context() context.Context {
	return s.ctx
}

// Logger returns the server's logger.
func (s *Server) Logger() *slog.Logger {
	return s.logger
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

// Router returns the root router.
func (s *Server) Router() *Router {
	return s.router
}

// Start the server. This operation is blocking and returns when the server is closed.
func (s *Server) Start() error {
	swapped := s.state.CompareAndSwap(0, 1)
	if !swapped {
		return errwrap.New("server was already started")
	}

	s.router.ClearRegexCache()

	defer func() {
		s.state.Store(3)
		// Notify the shutdown is complete so Stop() can return
		s.stopChannel <- struct{}{}
		close(s.stopChannel)
	}()

	var ln net.Listener
	var err error
	if s.listenConfig != nil {
		ln, err = s.listenConfig.Listen(s.ctx, "tcp", s.server.Addr)
	} else {
		ln, err = net.Listen("tcp", s.server.Addr)
	}
	if err != nil {
		return errwrap.New(err)
	}

	select {
	case <-s.ctx.Done():
		return errwrap.New([]any{"cannot start the server, context is canceled", context.Canceled})
	default:
	}

	s.port = ln.Addr().(*net.TCPAddr).Port
	s.refreshURLs()
	defer func() {
		for _, hook := range s.shutdownHooks {
			hook(s)
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
	if err := s.server.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
		s.state.Store(3)
		return errwrap.New(err)
	}
	return nil
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
		s.logger.Error(errwrap.NewSkip(err, 3))
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
	w.server.logger.Error(fmt.Errorf("%s", p))
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
