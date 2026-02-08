package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"
)

type readinessResponse struct {
	Ready bool `json:"ready"`
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(readinessResponse{Ready: true})
	})

	server := &http.Server{
		Addr:              ":" + serverPort(),
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("backend listening on %s", server.Addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}

func serverPort() string {
	if port := os.Getenv("PORT"); port != "" {
		return port
	}
	return "8080"
}
