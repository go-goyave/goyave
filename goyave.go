package goyave

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/System-Glitch/goyave/config"
	"github.com/System-Glitch/goyave/database"
	"github.com/System-Glitch/goyave/lang"
)

// Start starts the web server.
// The routeRegistrer parameter is a function aimed at registering all your routes and middlewares.
//  import "github.com/System-Glitch/goyave"
//  import "routes"
//  func main() {
// 	    goyave.start(routes.RegisterRoutes)
//  }
func Start(routeRegistrer func(*Router)) {
	err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
		return
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
	httpsAddress := getAddress("https")
	timeout := time.Duration(config.Get("timeout").(float64)) * time.Second
	server := &http.Server{
		Addr:         getAddress("http"),
		WriteTimeout: timeout,
		ReadTimeout:  timeout,
		IdleTimeout:  timeout * 2,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "https://"+httpsAddress+r.RequestURI, http.StatusMovedPermanently)
		}),
	}
	go func() {
		if err := server.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()

	registerShutdownHook(func(ctx context.Context) {
		server.Shutdown(ctx)
	})
}

func startServer(router *Router) {
	timeout := time.Duration(config.Get("timeout").(float64)) * time.Second
	protocol := config.GetString("protocol")
	server := &http.Server{
		Addr:         getAddress(protocol),
		WriteTimeout: timeout,
		ReadTimeout:  timeout,
		IdleTimeout:  timeout * 2,
		Handler:      router.muxRouter,
	}
	go func() {
		if protocol == "https" {
			go startTLSRedirectServer()
			if err := server.ListenAndServeTLS(config.GetString("tlsCert"), config.GetString("tlsKey")); err != nil {
				log.Fatal(err)
			}
		} else {
			if err := server.ListenAndServe(); err != nil {
				log.Fatal(err)
			}
		}
	}()

	registerShutdownHook(func(ctx context.Context) {
		server.Shutdown(ctx)
		database.Close()
	})
}

func registerShutdownHook(hook func(context.Context)) {
	sigChannel := make(chan os.Signal, 1)
	signal.Notify(sigChannel, os.Interrupt)

	<-sigChannel // Block until SIGINT received

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	hook(ctx)
}
