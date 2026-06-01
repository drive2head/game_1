package game

import "sync"

type LobbyStore struct {
	mu      sync.RWMutex
	lobbies map[string]*Lobby
}

func NewLobbyStore() *LobbyStore {
	return &LobbyStore{
		lobbies: make(map[string]*Lobby),
	}
}

func (s *LobbyStore) Create(lobby *Lobby) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lobbies[lobby.Code] = lobby
}

func (s *LobbyStore) Get(code string) *Lobby {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.lobbies[code]
}

func (s *LobbyStore) Delete(code string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.lobbies, code)
}

func (s *LobbyStore) Exists(code string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.lobbies[code]
	return ok
}
