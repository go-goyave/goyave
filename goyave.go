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
	"sync"
	"syscall"
	"time"

	"github.com/System-Glitch/goyave/v2/config"
	"github.com/System-Glitch/goyave/v2/database"
	"github.com/System-Glitch/goyave/v2/lang"
)

var (
	server             *http.Server
	redirectServer     *http.Server
	router             *Router
	maintenanceHandler http.Handler
	sigChannel         chan os.Signal
	stopChannel        chan bool
	hookChannel        chan bool

	startupHooks       []func()
	ready              bool = false
	maintenanceEnabled bool = false
	mutex                   = &sync.RWMutex{}
	once               sync.Once
)

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

// Start starts the web server.
// The routeRegistrer parameter is a function aimed at registering all your routes and middleware.
//  import (
//      "github.com/System-Glitch/goyave/v2"
//      "my-project/route"
//  )
//
//  func main() {
// 	    goyave.start(route.Register)
//  }
//
// Panic if the server is already running.
func Start(routeRegistrer func(*Router)) {
	if IsReady() {
		log.Panicf("Server is already running.")
	}

	mutex.Lock()
	if !config.IsLoaded() {
		if err := config.Load(); err != nil {
			fmt.Println(err)
			return
		}
	}

	lang.LoadDefault()
	lang.LoadAllAvailableLanguages()

	if config.GetBool("dbAutoMigrate") && config.GetString("dbConnection") != "none" {
		database.Migrate()
	}

	router = newRouter()
	routeRegistrer(router)
	startServer(router)
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
		hookChannel <- true // Clear shutdown hook
		<-hookChannel
		hookChannel = nil
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
			<-stopChannel
			redirectServer = nil
			stopChannel = nil
		}
	}
	return err
}

func getHost(protocol string) string {
	var port string
	if protocol == "https" {
		port = "httpsPort"
	} else {
		port = "port"
	}
	return config.GetString("host") + ":" + strconv.FormatInt(int64(config.Get(port).(float64)), 10)
}

func getAddress(protocol string) string {
	var shouldShowPort bool
	var port string
	if protocol == "https" {
		p := int64(config.Get("httpsPort").(float64))
		port = strconv.FormatInt(p, 10)
		shouldShowPort = p != 443
	} else {
		p := int64(config.Get("port").(float64))
		port = strconv.FormatInt(p, 10)
		shouldShowPort = p != 80
	}
	host := config.GetString("domain")
	if len(host) == 0 {
		host = config.GetString("host")
	}

	if shouldShowPort {
		host += ":" + port
	}

	return protocol + "://" + host
}

func startTLSRedirectServer() {
	httpsAddress := getAddress("https")
	timeout := time.Duration(config.Get("timeout").(float64)) * time.Second
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
		fmt.Printf("The TLS redirect server encountered an error: %s\n", err.Error())
		redirectServer = nil
		return
	}

	stopChannel = make(chan bool, 1)

	ok := ready
	r := redirectServer

	go func() {
		if ok && r != nil {
			if err := r.Serve(ln); err != nil && err != http.ErrServerClosed {
				fmt.Printf("The TLS redirect server encountered an error: %s\n", err.Error())
				mutex.Lock()
				redirectServer = nil
				stopChannel = nil
				ln.Close()
				mutex.Unlock()
				return
			}
		}
		ln.Close()
		stopChannel <- true
	}()
}

func startServer(router *Router) {
	timeout := time.Duration(config.Get("timeout").(float64)) * time.Second
	protocol := config.GetString("protocol")
	server = &http.Server{
		Addr:         getHost(protocol),
		WriteTimeout: timeout,
		ReadTimeout:  timeout,
		IdleTimeout:  timeout * 2,
		Handler:      router,
	}

	if config.GetBool("maintenance") {
		server.Handler = getMaintenanceHandler()
		maintenanceEnabled = true
	}

	ln, err := net.Listen("tcp", server.Addr)
	if err != nil {
		fmt.Println(err)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		stop(ctx)
		mutex.Unlock()
		return
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
		if err := s.ServeTLS(ln, config.GetString("tlsCert"), config.GetString("tlsKey")); err != nil && err != http.ErrServerClosed {
			fmt.Println(err)
			Stop()
		}
	} else {

		s := server
		mutex.Unlock()
		runStartupHooks()
		if err := s.Serve(ln); err != nil && err != http.ErrServerClosed {
			fmt.Println(err)
			Stop()
		}
	}
}

func runStartupHooks() {
	for _, hook := range startupHooks {
		go hook()
	}
}

func registerShutdownHook(hook func(context.Context) error) {
	hookChannel = make(chan bool)
	sigChannel = make(chan os.Signal)
	signal.Notify(sigChannel, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		hookChannel <- true
		select {
		case <-hookChannel:
			hookChannel <- true
		case <-sigChannel: // Block until SIGINT or SIGTERM received
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			mutex.Lock()
			close(hookChannel)
			hookChannel = nil
			sigChannel = nil
			hook(ctx)
			mutex.Unlock()
		}
	}()
}
