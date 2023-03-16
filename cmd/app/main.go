package main

import (
	"github.com/ogamor69/affise/internal/app"
	"github.com/ogamor69/affise/internal/handler"
	"log"
	"net"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handler.Handler)
	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	// Create a listener with a limit of 100 simultaneous connections
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("Failed to create listener: %v", err)
	}
	limitedListener := app.NewLimitedListener(listener, 100)

	// Chan for errors that occur when the server is running
	httpErrChan := make(chan error, 1)
	// Chan for graceful shutdown
	httpShutdownChan := make(chan struct{})

	// Go routine for graceful shutdown
	go func() {
		if err := server.Serve(limitedListener); err != nil && err != http.ErrServerClosed {
			httpErrChan <- err
		}
	}()

	// Function for graceful shutdown
	app.PerformGracefulShutdown(server, httpErrChan, httpShutdownChan)
}
