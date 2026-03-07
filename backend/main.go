package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"game_1/game"
	"game_1/handler"
	"game_1/hub"
)

type readinessResponse struct {
	Ready bool `json:"ready"`
}

func main() {
	store := game.NewLobbyStore()
	wsHub := hub.NewHub(store)
	lobbyH := handler.NewLobbyHandler(store)
	wsH := handler.NewWSHandler(store, wsHub)

	mux := http.NewServeMux()

	mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(readinessResponse{Ready: true})
	})

	mux.HandleFunc("/api/lobbies", lobbyH.CreateLobby)

	mux.HandleFunc("/api/lobbies/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.Trim(r.URL.Path, "/")
		parts := strings.Split(path, "/")

		if len(parts) == 4 && parts[3] == "join" {
			lobbyH.JoinLobby(w, r)
			return
		}
		if len(parts) == 4 && parts[3] == "ws" {
			wsH.HandleWS(w, r)
			return
		}

		http.NotFound(w, r)
	})

	server := &http.Server{
		Addr:              ":" + serverPort(),
		Handler:           corsMiddleware(mux),
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

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
