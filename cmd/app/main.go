package main

import (
	"github.com/ogamor69/affise/internal/app"
	"github.com/ogamor69/affise/internal/handler"
	"log"
	"net"
	"net/http"
	"time"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handler.Handler)
	server := &http.Server{
		Handler: mux,
	}

	// Create a listener with a limit of 100 simultaneous connections
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("Failed to create listener: %v", err)
	}
	limitedListener := app.NewLimitedListener(app.LimitedListenerConfig{
		Listener:      listener,
		Limit:         100,
		AcceptTimeout: 5 * time.Second,
		ErrorCallback: func(err error) {
			log.Printf("Error: %v", err)
		},
	})

	// Start the server
	go func() {
		if err := server.Serve(limitedListener); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	// Channels for errors that occur when the server is running and for graceful shutdown
	httpErrChan := make(chan error, 1)
	httpShutdownChan := make(chan struct{})

	// Call PerformGracefulShutdown function
	app.PerformGracefulShutdown(server, httpErrChan, httpShutdownChan)

	// Wait for all connections to finish
	limitedListener.Wait()

	log.Println("Server has been gracefully shut down")
}
