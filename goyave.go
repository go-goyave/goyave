package goyave

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"goyave.dev/goyave/v3/config"
	"goyave.dev/goyave/v3/database"
	"goyave.dev/goyave/v3/lang"
)

var (
	server             *http.Server
	redirectServer     *http.Server
	router             *Router
	maintenanceHandler http.Handler
	sigChannel         chan os.Signal
	tlsStopChannel     chan struct{} = make(chan struct{}, 1)
	stopChannel        chan struct{} = make(chan struct{}, 1)
	hookChannel        chan struct{} = make(chan struct{}, 1)

	// Critical config entries (cached for better performance)
	protocol        string
	maxPayloadSize  int64
	defaultLanguage string

	startupHooks       []func()
	shutdownHooks      []func()
	ready              bool = false
	maintenanceEnabled bool = false
	mutex                   = &sync.RWMutex{}
	once               sync.Once

	// Logger the logger for default output
	// Writes to stdout by default.
	Logger *log.Logger = log.New(os.Stdout, "", log.LstdFlags)

	// AccessLogger the logger for access. This logger
	// is used by the logging middleware.
	// Writes to stdout by default.
	AccessLogger *log.Logger = log.New(os.Stdout, "", 0)

	// ErrLogger the logger in which errors and stacktraces are written.
	// Writes to stderr by default.
	ErrLogger *log.Logger = log.New(os.Stderr, "", log.LstdFlags)
)

const (
	// ExitInvalidConfig the exit code returned when the config
	// validation doesn't pass.
	ExitInvalidConfig = 3

	// ExitNetworkError the exit code returned when an error
	// occurs when opening the network listener
	ExitNetworkError = 4

	// ExitHTTPError the exit code returned when an error
	// occurs in the HTTP server (port already in use for example)
	ExitHTTPError = 5
)

// Error wrapper for errors directely related to the server itself.
// Contains an exit code and the original error.
type Error struct {
	ExitCode int
	Err      error
}

func (e *Error) Error() string {
	return e.Err.Error()
}

// IsReady returns true if the server has finished initializing and
// is ready to serve incoming requests.
func IsReady() bool {
	mutex.RLock()
	defer mutex.RUnlock()
	return ready
}

// RegisterStartupHook to execute some code once the server is ready and running.
func RegisterStartupHook(hook func()) {
	mutex.Lock()
	startupHooks = append(startupHooks, hook)
	mutex.Unlock()
}

// ClearStartupHooks removes all startup hooks.
func ClearStartupHooks() {
	mutex.Lock()
	startupHooks = []func(){}
	mutex.Unlock()
}

// RegisterShutdownHook to execute some code after the server stopped.
// Shutdown hooks are executed before goyave.Start() returns.
func RegisterShutdownHook(hook func()) {
	mutex.Lock()
	shutdownHooks = append(shutdownHooks, hook)
	mutex.Unlock()
}

// ClearShutdownHooks removes all shutdown hooks.
func ClearShutdownHooks() {
	mutex.Lock()
	shutdownHooks = []func(){}
	mutex.Unlock()
}

// Start starts the web server.
// The routeRegistrer parameter is a function aimed at registering all your routes and middleware.
//  import (
//      "goyave.dev/goyave/v3"
//      "github.com/username/projectname/route"
//  )
//
//  func main() {
//      if err := goyave.Start(route.Register); err != nil {
//          os.Exit(err.(*goyave.Error).ExitCode)
//      }
//  }
//
// Errors returned can be safely type-asserted to "*goyave.Error".
// Panics if the server is already running.
func Start(routeRegistrer func(*Router)) error {
	if IsReady() {
		ErrLogger.Panicf("Server is already running.")
	}

	mutex.Lock()
	if !config.IsLoaded() {
		if err := config.Load(); err != nil {
			ErrLogger.Println(err)
			mutex.Unlock()
			return &Error{ExitInvalidConfig, err}
		}
	}

	// Performance improvements by loading critical config entries beforehand
	cacheCriticalConfig()

	lang.LoadDefault()
	lang.LoadAllAvailableLanguages()

	if config.GetBool("database.autoMigrate") && config.GetString("database.connection") != "none" {
		database.Migrate()
	}

	router = newRouter()
	routeRegistrer(router)
	regexCache = nil // Clear regex cache
	return startServer(router)
}

func cacheCriticalConfig() {
	maxPayloadSize = int64(config.GetFloat("server.maxUploadSize") * 1024 * 1024)
	defaultLanguage = config.GetString("app.defaultLanguage")
	protocol = config.GetString("server.protocol")
}

// EnableMaintenance replace the main server handler with the "Service Unavailable" handler.
func EnableMaintenance() {
	mutex.Lock()
	server.Handler = getMaintenanceHandler()
	maintenanceEnabled = true
	mutex.Unlock()
}

// DisableMaintenance replace the main server handler with the original router.
func DisableMaintenance() {
	mutex.Lock()
	server.Handler = router
	maintenanceEnabled = false
	mutex.Unlock()
}

// IsMaintenanceEnabled return true if the server is currently in maintenance mode.
func IsMaintenanceEnabled() bool {
	mutex.RLock()
	defer mutex.RUnlock()
	return maintenanceEnabled
}

// GetRoute get a named route.
// Returns nil if the route doesn't exist.
func GetRoute(name string) *Route {
	mutex.Lock()
	defer mutex.Unlock()
	return router.namedRoutes[name]
}

func getMaintenanceHandler() http.Handler {
	once.Do(func() {
		maintenanceHandler = http.HandlerFunc(func(resp http.ResponseWriter, request *http.Request) {
			resp.WriteHeader(http.StatusServiceUnavailable)
		})
	})
	return maintenanceHandler
}

// Stop gracefully shuts down the server without interrupting any
// active connections.
//
// Make sure the program doesn't exit and waits instead for Stop to return.
//
// Stop does not attempt to close nor wait for hijacked
// connections such as WebSockets. The caller of Stop should
// separately notify such long-lived connections of shutdown and wait
// for them to close, if desired.
func Stop() {
	mutex.Lock()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	stop(ctx)
	if sigChannel != nil {
		hookChannel <- struct{}{} // Clear shutdown hook
		<-hookChannel
		sigChannel = nil
	}
	mutex.Unlock()
}

func stop(ctx context.Context) error {
	var err error
	if server != nil {
		err = server.Shutdown(ctx)
		database.Close()
		server = nil
		router = nil
		ready = false
		maintenanceEnabled = false
		if redirectServer != nil {
			redirectServer.Shutdown(ctx)
			<-tlsStopChannel
			redirectServer = nil
		}

		for _, hook := range shutdownHooks {
			hook()
		}
		stopChannel <- struct{}{}
	}
	return err
}

func getHost(protocol string) string {
	var port string
	if protocol == "https" {
		port = "server.httpsPort"
	} else {
		port = "server.port"
	}
	return config.GetString("server.host") + ":" + strconv.Itoa(config.GetInt(port))
}

func getAddress(protocol string) string {
	var shouldShowPort bool
	var port string
	if protocol == "https" {
		p := config.GetInt("server.httpsPort")
		port = strconv.Itoa(p)
		shouldShowPort = p != 443
	} else {
		p := config.GetInt("server.port")
		port = strconv.Itoa(p)
		shouldShowPort = p != 80
	}
	host := config.GetString("server.domain")
	if len(host) == 0 {
		host = config.GetString("server.host")
		if host == "0.0.0.0" {
			host = "127.0.0.1"
		}
	}

	if shouldShowPort {
		host += ":" + port
	}

	return protocol + "://" + host
}

// BaseURL returns the base URL of your application.
func BaseURL() string {
	return getAddress(config.GetString("server.protocol"))
}

func startTLSRedirectServer() {
	httpsAddress := getAddress("https")
	timeout := time.Duration(config.GetInt("server.timeout")) * time.Second
	redirectServer = &http.Server{
		Addr:         getHost("http"),
		WriteTimeout: timeout,
		ReadTimeout:  timeout,
		IdleTimeout:  timeout * 2,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			address := httpsAddress + r.URL.Path
			query := r.URL.Query()
			if len(query) != 0 {
				address += "?" + query.Encode()
			}
			http.Redirect(w, r, address, http.StatusPermanentRedirect)
		}),
	}

	ln, err := net.Listen("tcp", redirectServer.Addr)
	if err != nil {
		ErrLogger.Printf("The TLS redirect server encountered an error: %s\n", err.Error())
		redirectServer = nil
		return
	}

	ok := ready
	r := redirectServer

	go func() {
		if ok && r != nil {
			if err := r.Serve(ln); err != nil && err != http.ErrServerClosed {
				ErrLogger.Printf("The TLS redirect server encountered an error: %s\n", err.Error())
				mutex.Lock()
				redirectServer = nil
				ln.Close()
				mutex.Unlock()
				return
			}
		}
		ln.Close()
		tlsStopChannel <- struct{}{}
	}()
}

func startServer(router *Router) error {
	defer func() {
		<-stopChannel // Wait for stop() to finish before returning
	}()
	timeout := time.Duration(config.GetInt("server.timeout")) * time.Second
	server = &http.Server{
		Addr:         getHost(protocol),
		WriteTimeout: timeout,
		ReadTimeout:  timeout,
		IdleTimeout:  timeout * 2,
		Handler:      router,
	}

	if config.GetBool("server.maintenance") {
		server.Handler = getMaintenanceHandler()
		maintenanceEnabled = true
	}

	ln, err := net.Listen("tcp", server.Addr)
	if err != nil {
		ErrLogger.Println(err)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		stop(ctx)
		mutex.Unlock()
		return &Error{ExitNetworkError, err}
	}
	defer ln.Close()
	registerShutdownHook(stop)
	<-hookChannel

	ready = true
	if protocol == "https" {
		startTLSRedirectServer()

		s := server
		mutex.Unlock()
		runStartupHooks()
		if err := s.ServeTLS(ln, config.GetString("server.tls.cert"), config.GetString("server.tls.key")); err != nil && err != http.ErrServerClosed {
			ErrLogger.Println(err)
			Stop()
			return &Error{ExitHTTPError, err}
		}
	} else {

		s := server
		mutex.Unlock()
		runStartupHooks()
		if err := s.Serve(ln); err != nil && err != http.ErrServerClosed {
			ErrLogger.Println(err)
			Stop()
			return &Error{ExitHTTPError, err}
		}
	}

	return nil
}

func runStartupHooks() {
	for _, hook := range startupHooks {
		go hook()
	}
}

func registerShutdownHook(hook func(context.Context) error) {
	sigChannel = make(chan os.Signal, 64)
	signal.Notify(sigChannel, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		hookChannel <- struct{}{}
		select {
		case <-hookChannel:
			hookChannel <- struct{}{}
		case <-sigChannel: // Block until SIGINT or SIGTERM received
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			mutex.Lock()
			sigChannel = nil
			hook(ctx)
			mutex.Unlock()
		}
	}()
}

// TODO refactor server sartup (use context)
