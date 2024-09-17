package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"log/slog"
)

func main() {
	// Define the port flag
	port := flag.String("port", "8080", "Port number to listen on")
	flag.Parse()

	if *port == "" {
		slog.Error("Port number is required")
		return
	}

	http.HandleFunc("/hi", func(w http.ResponseWriter, r *http.Request) {
		slog.Info("received your message!")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Hello, World!")
	})

	slog.Info("Starting server on port", "port", *port)
	if err := http.ListenAndServe(":"+*port, nil); err != nil {
		slog.Error("Error starting server", "error", err.Error())
		os.Exit(1)
	}
}
