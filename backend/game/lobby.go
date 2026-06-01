package game

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"sync"

	"github.com/google/uuid"
)

var (
	ErrLobbyFull       = errors.New("lobby_full")
	ErrGameInProgress  = errors.New("game_in_progress")
	ErrNicknameTaken   = errors.New("nickname_taken")
	ErrInvalidNickname = errors.New("invalid_nickname")
	ErrNotAdmin        = errors.New("not_admin")
	ErrNotEnoughPlayers = errors.New("not_enough_players")
	ErrPlayersNotReady = errors.New("players_not_ready")
	ErrWrongPhase      = errors.New("wrong_phase")
	ErrPlayerNotFound  = errors.New("player_not_found")
)

type LobbyState string

const (
	LobbyStateWaiting LobbyState = "waiting"
	LobbyStateInGame  LobbyState = "in_game"
)

type Player struct {
	ID       string `json:"id"`
	Nickname string `json:"nickname"`
	IsAdmin  bool   `json:"isAdmin"`
	Ready    bool   `json:"ready"`
	Online   bool   `json:"online"`
}

type LobbySettings struct {
	MinPlayers         int      `json:"minPlayers"`
	MaxPlayers         int      `json:"maxPlayers"`
	Locations          []string `json:"locations"`
	VoteThreshold      float64  `json:"voteThreshold"`
	VotingTimerSeconds int      `json:"votingTimerSeconds"`
}

type Lobby struct {
	mu             sync.RWMutex
	Code           string         `json:"code"`
	Players        []Player       `json:"players"`
	Settings       LobbySettings  `json:"settings"`
	State          LobbyState     `json:"state"`
	CurrentSession *Session       `json:"currentSession"`
}

func DefaultSettings() LobbySettings {
	return LobbySettings{
		MinPlayers:         3,
		MaxPlayers:         8,
		Locations:          DefaultLocations(),
		VoteThreshold:      0.5,
		VotingTimerSeconds: 30,
	}
}

func NewLobby(nickname string) (*Lobby, *Player, error) {
	nickname = strings.TrimSpace(nickname)
	if err := validateNickname(nickname); err != nil {
		return nil, nil, err
	}

	code, err := generateInviteCode()
	if err != nil {
		return nil, nil, fmt.Errorf("generate invite code: %w", err)
	}

	player := Player{
		ID:       uuid.New().String(),
		Nickname: nickname,
		IsAdmin:  true,
		Ready:    false,
		Online:   false,
	}

	lobby := &Lobby{
		Code:     code,
		Players:  []Player{player},
		Settings: DefaultSettings(),
		State:    LobbyStateWaiting,
	}

	return lobby, &player, nil
}

func (l *Lobby) AddPlayer(nickname string) (*Player, error) {
	nickname = strings.TrimSpace(nickname)
	if err := validateNickname(nickname); err != nil {
		return nil, err
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	if l.State == LobbyStateInGame {
		for i, p := range l.Players {
			if p.Nickname == nickname && !p.Online {
				l.Players[i].Online = true
				return &l.Players[i], nil
			}
		}
		return nil, ErrGameInProgress
	}

	if len(l.Players) >= l.Settings.MaxPlayers {
		return nil, ErrLobbyFull
	}

	for _, p := range l.Players {
		if p.Nickname == nickname {
			return nil, ErrNicknameTaken
		}
	}

	player := Player{
		ID:       uuid.New().String(),
		Nickname: nickname,
		IsAdmin:  false,
		Ready:    false,
		Online:   false,
	}
	l.Players = append(l.Players, player)
	return &player, nil
}

func (l *Lobby) RemovePlayer(playerID string) (empty bool, err error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.State != LobbyStateWaiting {
		return false, ErrWrongPhase
	}

	idx := -1
	for i, p := range l.Players {
		if p.ID == playerID {
			idx = i
			break
		}
	}
	if idx == -1 {
		return false, ErrPlayerNotFound
	}

	wasAdmin := l.Players[idx].IsAdmin
	l.Players = append(l.Players[:idx], l.Players[idx+1:]...)

	if len(l.Players) == 0 {
		return true, nil
	}

	if wasAdmin {
		l.Players[0].IsAdmin = true
	}

	return false, nil
}

func (l *Lobby) SetPlayerOnline(playerID string, online bool) {
	l.mu.Lock()
	defer l.mu.Unlock()

	for i, p := range l.Players {
		if p.ID == playerID {
			l.Players[i].Online = online
			if !online && l.State == LobbyStateWaiting {
				l.Players[i].Ready = false
			}
			break
		}
	}
}

func (l *Lobby) SetPlayerReady(playerID string, ready bool) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.State != LobbyStateWaiting {
		return ErrWrongPhase
	}

	for i, p := range l.Players {
		if p.ID == playerID {
			l.Players[i].Ready = ready
			return nil
		}
	}
	return ErrPlayerNotFound
}

func (l *Lobby) StartGame() (*Session, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.State != LobbyStateWaiting {
		return nil, ErrWrongPhase
	}

	if len(l.Players) < l.Settings.MinPlayers {
		return nil, ErrNotEnoughPlayers
	}

	for _, p := range l.Players {
		if !p.Ready {
			return nil, ErrPlayersNotReady
		}
	}

	playerIDs := make([]string, len(l.Players))
	for i, p := range l.Players {
		playerIDs[i] = p.ID
	}

	session, err := NewSession(playerIDs, l.Settings.Locations)
	if err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}

	l.CurrentSession = session
	l.State = LobbyStateInGame

	for i := range l.Players {
		l.Players[i].Ready = false
	}

	return session, nil
}

func (l *Lobby) FindPlayerByID(playerID string) *Player {
	l.mu.RLock()
	defer l.mu.RUnlock()

	for i, p := range l.Players {
		if p.ID == playerID {
			return &l.Players[i]
		}
	}
	return nil
}

func (l *Lobby) IsAdmin(playerID string) bool {
	l.mu.RLock()
	defer l.mu.RUnlock()

	for _, p := range l.Players {
		if p.ID == playerID {
			return p.IsAdmin
		}
	}
	return false
}

type LobbyView struct {
	Code           string        `json:"code"`
	Players        []Player      `json:"players"`
	Settings       LobbySettings `json:"settings"`
	State          LobbyState    `json:"state"`
	CurrentSession *Session      `json:"currentSession"`
}

func (l *Lobby) Snapshot() LobbyView {
	l.mu.RLock()
	defer l.mu.RUnlock()

	players := make([]Player, len(l.Players))
	copy(players, l.Players)

	return LobbyView{
		Code:           l.Code,
		Players:        players,
		Settings:       l.Settings,
		State:          l.State,
		CurrentSession: l.CurrentSession,
	}
}

func validateNickname(nickname string) error {
	if len(nickname) == 0 || len(nickname) > 20 {
		return ErrInvalidNickname
	}
	return nil
}

const inviteCodeLength = 6
const inviteCodeChars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func generateInviteCode() (string, error) {
	code := make([]byte, inviteCodeLength)
	for i := range code {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(inviteCodeChars))))
		if err != nil {
			return "", err
		}
		code[i] = inviteCodeChars[n.Int64()]
	}
	return string(code), nil
}
