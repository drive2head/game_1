package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"game_1/game"
)

type LobbyHandler struct {
	store *game.LobbyStore
}

func NewLobbyHandler(store *game.LobbyStore) *LobbyHandler {
	return &LobbyHandler{store: store}
}

type nicknameRequest struct {
	Nickname string `json:"nickname"`
}

type lobbyResponse struct {
	Lobby    game.LobbyView `json:"lobby"`
	PlayerID string         `json:"playerId"`
}

func (h *LobbyHandler) CreateLobby(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req nicknameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorBody("bad_request"))
		return
	}

	lobby, player, err := game.NewLobby(req.Nickname)
	if err != nil {
		writeGameError(w, err)
		return
	}

	h.store.Create(lobby)

	writeJSON(w, http.StatusCreated, lobbyResponse{
		Lobby:    lobby.Snapshot(),
		PlayerID: player.ID,
	})
}

func (h *LobbyHandler) JoinLobby(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	code := extractLobbyCode(r.URL.Path)
	if code == "" {
		writeJSON(w, http.StatusBadRequest, errorBody("bad_request"))
		return
	}

	lobby := h.store.Get(code)
	if lobby == nil {
		writeJSON(w, http.StatusNotFound, errorBody("lobby_not_found"))
		return
	}

	var req nicknameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorBody("bad_request"))
		return
	}

	player, err := lobby.AddPlayer(req.Nickname)
	if err != nil {
		writeGameError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, lobbyResponse{
		Lobby:    lobby.Snapshot(),
		PlayerID: player.ID,
	})
}

func extractLobbyCode(path string) string {
	// /api/lobbies/{code}/join → code
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) >= 4 && parts[0] == "api" && parts[1] == "lobbies" {
		return parts[2]
	}
	return ""
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func errorBody(code string) map[string]string {
	return map[string]string{"error": code}
}

func writeGameError(w http.ResponseWriter, err error) {
	switch err {
	case game.ErrInvalidNickname:
		writeJSON(w, http.StatusBadRequest, errorBody("invalid_nickname"))
	case game.ErrLobbyFull:
		writeJSON(w, http.StatusConflict, errorBody("lobby_full"))
	case game.ErrGameInProgress:
		writeJSON(w, http.StatusConflict, errorBody("game_in_progress"))
	case game.ErrNicknameTaken:
		writeJSON(w, http.StatusConflict, errorBody("nickname_taken"))
	default:
		writeJSON(w, http.StatusInternalServerError, errorBody("internal_error"))
	}
}
