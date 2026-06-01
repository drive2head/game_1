package hub

import (
	"encoding/json"
	"log"
	"sync"

	"game_1/game"
)

type Hub struct {
	mu      sync.RWMutex
	clients map[string]*Client // playerId → *Client
	store   *game.LobbyStore
}

func NewHub(store *game.LobbyStore) *Hub {
	return &Hub{
		clients: make(map[string]*Client),
		store:   store,
	}
}

func (h *Hub) Register(c *Client) {
	h.mu.Lock()
	if old, ok := h.clients[c.playerID]; ok {
		old.Close()
	}
	h.clients[c.playerID] = c
	h.mu.Unlock()

	lobby := h.store.Get(c.lobbyCode)
	if lobby == nil {
		return
	}

	wasOffline := false
	p := lobby.FindPlayerByID(c.playerID)
	if p != nil && !p.Online {
		wasOffline = true
	}

	lobby.SetPlayerOnline(c.playerID, true)

	if wasOffline {
		h.BroadcastToLobby(c.lobbyCode, NewOutgoing("player_reconnected", map[string]string{
			"playerId": c.playerID,
		}))
	}

	h.BroadcastLobbyUpdated(c.lobbyCode)
}

func (h *Hub) Unregister(c *Client) {
	h.mu.Lock()
	if current, ok := h.clients[c.playerID]; ok && current == c {
		delete(h.clients, c.playerID)
	}
	h.mu.Unlock()

	lobby := h.store.Get(c.lobbyCode)
	if lobby == nil {
		return
	}

	lobby.SetPlayerOnline(c.playerID, false)

	h.BroadcastToLobby(c.lobbyCode, NewOutgoing("player_disconnected", map[string]string{
		"playerId": c.playerID,
	}))
	h.BroadcastLobbyUpdated(c.lobbyCode)
}

func (h *Hub) HandleMessage(c *Client, msg IncomingMessage) {
	switch msg.Type {
	case "leave_lobby":
		h.handleLeaveLobby(c)
	case "player_ready":
		h.handlePlayerReady(c, msg.Payload)
	case "start_game":
		h.handleStartGame(c)
	default:
		h.sendError(c, "unknown_command", "неизвестная команда: "+msg.Type)
	}
}

func (h *Hub) handleLeaveLobby(c *Client) {
	lobby := h.store.Get(c.lobbyCode)
	if lobby == nil {
		return
	}

	empty, err := lobby.RemovePlayer(c.playerID)
	if err != nil {
		h.sendError(c, err.Error(), "невозможно покинуть лобби")
		return
	}

	h.mu.Lock()
	delete(h.clients, c.playerID)
	h.mu.Unlock()
	c.Close()

	if empty {
		h.store.Delete(c.lobbyCode)
		return
	}

	h.BroadcastToLobby(c.lobbyCode, NewOutgoing("player_left", map[string]string{
		"playerId": c.playerID,
	}))
	h.BroadcastLobbyUpdated(c.lobbyCode)
}

func (h *Hub) handlePlayerReady(c *Client, raw json.RawMessage) {
	var payload ReadyPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		h.sendError(c, "invalid_payload", "невалидный payload")
		return
	}

	lobby := h.store.Get(c.lobbyCode)
	if lobby == nil {
		return
	}

	if err := lobby.SetPlayerReady(c.playerID, payload.Ready); err != nil {
		h.sendError(c, err.Error(), "невозможно изменить готовность")
		return
	}

	h.BroadcastLobbyUpdated(c.lobbyCode)
}

func (h *Hub) handleStartGame(c *Client) {
	lobby := h.store.Get(c.lobbyCode)
	if lobby == nil {
		return
	}

	if !lobby.IsAdmin(c.playerID) {
		h.sendError(c, "not_admin", "только администратор может запустить игру")
		return
	}

	session, err := lobby.StartGame()
	if err != nil {
		h.sendError(c, err.Error(), "невозможно запустить игру")
		return
	}

	h.BroadcastToLobby(c.lobbyCode, NewOutgoing("game_started", map[string]interface{}{
		"sessionId":        session.ID,
		"turnOrder":        session.TurnOrder,
		"phase":            session.Phase,
		"currentTurnIndex": session.CurrentTurnIndex,
	}))

	snap := lobby.Snapshot()
	for _, p := range snap.Players {
		role := "regular"
		var location interface{} = session.Location
		if p.ID == session.SpyID {
			role = "spy"
			location = nil
		}
		h.SendToPlayer(p.ID, NewOutgoing("role_assigned", map[string]interface{}{
			"role":     role,
			"location": location,
		}))
	}

	h.BroadcastLobbyUpdated(c.lobbyCode)
}

func (h *Hub) BroadcastToLobby(code string, msg OutgoingMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("[hub] ошибка маршалинга broadcast: %v", err)
		return
	}

	lobby := h.store.Get(code)
	if lobby == nil {
		return
	}

	snap := lobby.Snapshot()
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, p := range snap.Players {
		if client, ok := h.clients[p.ID]; ok {
			client.Send(data)
		}
	}
}

func (h *Hub) BroadcastLobbyUpdated(code string) {
	lobby := h.store.Get(code)
	if lobby == nil {
		return
	}
	snap := lobby.Snapshot()
	h.BroadcastToLobby(code, NewOutgoing("lobby_updated", map[string]interface{}{
		"lobby": snap,
	}))
}

func (h *Hub) SendToPlayer(playerID string, msg OutgoingMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("[hub] ошибка маршалинга personal: %v", err)
		return
	}

	h.mu.RLock()
	client, ok := h.clients[playerID]
	h.mu.RUnlock()

	if ok {
		client.Send(data)
	}
}

func (h *Hub) sendError(c *Client, errCode, message string) {
	data, err := json.Marshal(NewOutgoing("error", ErrorPayload{
		Error:   errCode,
		Message: message,
	}))
	if err != nil {
		return
	}
	c.Send(data)
}
