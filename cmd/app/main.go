package main

import (
	"github.com/ogamor69/affise/internal/app"
	"github.com/ogamor69/affise/internal/handler"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handler.Handler)
	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	//Chan for errors that occur when the server is running
	httpErrChan := make(chan error, 1)
	// Chan for graceful shutdown
	httpShutdownChan := make(chan struct{})

	//Go routine for graceful shutdown
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			httpErrChan <- err
		}
	}()

	// Function for graceful shutdown
	app.PerformGracefulShutdown(server, httpErrChan, httpShutdownChan)
}
