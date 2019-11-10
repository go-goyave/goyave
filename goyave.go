package goyave

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"time"

	"github.com/System-Glitch/goyave/config"
	"github.com/System-Glitch/goyave/database"
	"github.com/System-Glitch/goyave/lang"
)

var server *http.Server = nil
var redirectServer *http.Server = nil
var sigChannel chan os.Signal

var startupHooks []func()
var ready bool = false
var mutex = &sync.Mutex{}
var serverMutex = &sync.Mutex{}

// IsReady returns true if the server has finished initializing and
// is ready to serve incoming requests.
func IsReady() bool {
	mutex.Lock()
	r := ready
	mutex.Unlock()
	return r
}

// RegisterStartupHook to execute some code once the server is ready and running.
func RegisterStartupHook(hook func()) {
	startupHooks = append(startupHooks, hook)
}

// ClearStartupHooks removes all startup hooks.
func ClearStartupHooks() {
	startupHooks = []func(){}
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
	err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	lang.LoadDefault()
	lang.LoadAllAvailableLanguages()

	if config.GetBool("dbAutoMigrate") && config.GetString("dbConnection") != "none" {
		database.Migrate()
	}

	router := newRouter()
	routeRegistrer(router)
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
	mutex.Lock()
	sigChannel <- os.Interrupt
	mutex.Unlock()
}

func stop(ctx context.Context) error {
	mutex.Lock()
	err := server.Shutdown(ctx)
	database.Close()
	serverMutex.Lock()
	server = nil
	ready = false
	if redirectServer != nil {
		redirectServer.Shutdown(ctx)
		redirectServer = nil
	}
	serverMutex.Unlock()
	mutex.Unlock()
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
	mutex.Unlock()

	go func() {
		if err := redirectServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Println(err)
		}
	}()
}

func startServer(router *Router) {
	timeout := time.Duration(config.Get("timeout").(float64)) * time.Second
	protocol := config.GetString("protocol")
	serverMutex.Lock()
	server = &http.Server{
		Addr:         getAddress(protocol),
		WriteTimeout: timeout,
		ReadTimeout:  timeout,
		IdleTimeout:  timeout * 2,
		Handler:      router.muxRouter,
	}
	serverMutex.Unlock()

	go func() {
		serverMutex.Lock()
		if protocol == "https" {
			go startTLSRedirectServer()
			runStartupHooks()
			if err := server.ListenAndServeTLS(config.GetString("tlsCert"), config.GetString("tlsKey")); err != nil && err != http.ErrServerClosed {
				fmt.Println(err)
			}
		} else {
			runStartupHooks()
			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				fmt.Println(err)
			}
		}
		serverMutex.Unlock()
	}()

	registerShutdownHook(func(ctx context.Context) {
		stop(ctx)
	})
}

func runStartupHooks() {
	go func() {
		time.Sleep(100 * time.Millisecond)
		mutex.Lock()
		ready = true
		mutex.Unlock()
		for _, hook := range startupHooks {
			hook()
		}
	}()
}

func registerShutdownHook(hook func(context.Context)) {
	mutex.Lock()
	sigChannel = make(chan os.Signal, 1)
	signal.Notify(sigChannel, os.Interrupt)
	mutex.Unlock()

	<-sigChannel // Block until SIGINT received

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	hook(ctx)
}
