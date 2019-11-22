package goyave

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/System-Glitch/goyave/config"
	"github.com/System-Glitch/goyave/database"
	"github.com/System-Glitch/goyave/lang"
)

var server *http.Server
var redirectServer *http.Server
var sigChannel chan os.Signal

var startupHooks []func()
var ready bool = false
var mutex = &sync.Mutex{}

// IsReady returns true if the server has finished initializing and
// is ready to serve incoming requests.
func IsReady() bool {
	mutex.Lock()
	defer mutex.Unlock()
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
// The routeRegistrer parameter is a function aimed at registering all your routes and middlewares.
//  import (
//      "github.com/System-Glitch/goyave"
//      "routes"
//  )
//
//  func main() {
// 	    goyave.start(routes.Register)
//  }
func Start(routeRegistrer func(*Router)) {
	mutex.Lock()
	if !config.IsLoaded() {
		if err := config.Load(); err != nil {
			return
		}
	}

	lang.LoadDefault()
	lang.LoadAllAvailableLanguages()

	if config.GetBool("dbAutoMigrate") && config.GetString("dbConnection") != "none" {
		database.Migrate()
	}

	router := newRouter()
	routeRegistrer(router)
	mutex.Unlock()
	startServer(router)
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
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	mutex.Lock()
	stop(ctx)
	sigChannel <- syscall.SIGINT // Clear shutdown hook
	mutex.Unlock()
}

func stop(ctx context.Context) error {
	var err error
	if server != nil {
		err = server.Shutdown(ctx)
		database.Close()
		server = nil
		ready = false
		if redirectServer != nil {
			redirectServer.Shutdown(ctx)
			redirectServer = nil
		}
	}
	return err
}

// TODO add public shutdown hooks

func getAddress(protocol string) string {
	var port string
	if protocol == "https" {
		port = "httpsPort"
	} else {
		port = "port"
	}
	return config.GetString("host") + ":" + strconv.FormatInt(int64(config.Get(port).(float64)), 10)
}

func startTLSRedirectServer() {
	mutex.Lock()
	httpsAddress := getAddress("https")
	timeout := time.Duration(config.Get("timeout").(float64)) * time.Second
	redirectServer = &http.Server{
		Addr:         getAddress("http"),
		WriteTimeout: timeout,
		ReadTimeout:  timeout,
		IdleTimeout:  timeout * 2,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "https://"+httpsAddress+r.RequestURI, http.StatusPermanentRedirect)
		}),
	}

	go func() {
		mutex.Lock()
		ok := server != nil
		mutex.Unlock()
		if ok {
			if err := redirectServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				fmt.Println("The TLS redirect server encountered an error:")
				fmt.Println(err)
			}
		}
	}()

	mutex.Unlock()
}

func startServer(router *Router) {
	timeout := time.Duration(config.Get("timeout").(float64)) * time.Second
	protocol := config.GetString("protocol")
	server = &http.Server{
		Addr:         getAddress(protocol),
		WriteTimeout: timeout,
		ReadTimeout:  timeout,
		IdleTimeout:  timeout * 2,
		Handler:      router.muxRouter,
	}

	go func() {
		if protocol == "https" {
			go startTLSRedirectServer()
			runStartupHooks()
			if err := server.ListenAndServeTLS(config.GetString("tlsCert"), config.GetString("tlsKey")); err != nil && err != http.ErrServerClosed {
				fmt.Println(err)
				Stop()
				return
			}
		} else {
			runStartupHooks()
			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				fmt.Println(err)
				Stop()
				return
			}
		}
	}()

	registerShutdownHook(func(ctx context.Context) {
		mutex.Lock()
		stop(ctx)
		mutex.Unlock()
	})
}

func runStartupHooks() {
	go func() {
		time.Sleep(100 * time.Millisecond) // TODO improve startup hooks
		mutex.Lock()
		if server != nil {
			ready = true
			mutex.Unlock()
			for _, hook := range startupHooks {
				hook()
			}
			return
		}
		mutex.Unlock()
	}()
}

func registerShutdownHook(hook func(context.Context)) {
	mutex.Lock()
	sigChannel = make(chan os.Signal, 1)
	signal.Notify(sigChannel, syscall.SIGINT, syscall.SIGTERM)
	mutex.Unlock()

	<-sigChannel // Block until SIGINT or SIGTERM received

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	hook(ctx)
}
