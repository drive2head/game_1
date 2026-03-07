package handler

import (
	"context"
	"log"
	"net/http"
	"strings"

	"nhooyr.io/websocket"

	"game_1/game"
	"game_1/hub"
)

type WSHandler struct {
	store *game.LobbyStore
	hub   *hub.Hub
}

func NewWSHandler(store *game.LobbyStore, h *hub.Hub) *WSHandler {
	return &WSHandler{store: store, hub: h}
}

func (h *WSHandler) HandleWS(w http.ResponseWriter, r *http.Request) {
	code := extractWSLobbyCode(r.URL.Path)
	if code == "" {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	playerID := r.URL.Query().Get("playerId")
	if playerID == "" {
		http.Error(w, "missing playerId", http.StatusBadRequest)
		return
	}

	lobby := h.store.Get(code)
	if lobby == nil {
		http.Error(w, "lobby not found", http.StatusNotFound)
		return
	}

	player := lobby.FindPlayerByID(playerID)
	if player == nil {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true,
	})
	if err != nil {
		log.Printf("[ws] ошибка upgrade: %v", err)
		return
	}

	ctx := context.Background()

	client := hub.NewClient(conn, playerID, code, h.hub)
	h.hub.Register(client)

	go client.WritePump(ctx)
	client.ReadPump(ctx)
}

func extractWSLobbyCode(path string) string {
	// /api/lobbies/{code}/ws → code
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) >= 4 && parts[0] == "api" && parts[1] == "lobbies" && parts[3] == "ws" {
		return parts[2]
	}
	return ""
}
