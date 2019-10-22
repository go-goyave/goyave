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
)

// Start starts the web server.
// The routeRegistrer parameter is a function aimed at registering all your routes and middlewares.
//  import "github.com/System-Glitch/goyave"
//  import "routes"
//  func main() {
// 	    goyave.start(routes.RegisterRoutes)
//  }
func Start(routeRegistrer func(*Router)) {
	// TODO implement start
	err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
		return
	}

	router := newRouter()
	routeRegistrer(router)
	startServer(router)
}

func startServer(router *Router) {
	timeout := time.Duration(config.Get("timeout").(float64)) * time.Second
	server := &http.Server{
		Addr:         config.Get("host").(string) + ":" + strconv.FormatInt(int64(config.Get("port").(float64)), 10),
		WriteTimeout: timeout,
		ReadTimeout:  timeout,
		IdleTimeout:  timeout * 2,
		Handler:      router.muxRouter,
	}
	go func() {
		if err := server.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()

	// Process signals channel
	sigChannel := make(chan os.Signal, 1)

	// Graceful shutdown via SIGINT
	signal.Notify(sigChannel, os.Interrupt)

	<-sigChannel // Block until SIGINT received

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	server.Shutdown(ctx)
}

func loadDB() {
	// TODO implement loadDB
}
