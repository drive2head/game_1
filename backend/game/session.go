package game

import (
	"crypto/rand"
	"fmt"
	"math/big"

	"github.com/google/uuid"
)

type SessionPhase string

const (
	PhaseQuestion     SessionPhase = "question"
	PhaseVoteProposal SessionPhase = "vote_proposal"
	PhaseVoting       SessionPhase = "voting"
	PhaseSpyGuess     SessionPhase = "spy_guess"
	PhaseFinished     SessionPhase = "finished"
)

type Session struct {
	ID               string       `json:"id"`
	Location         string       `json:"location"`
	SpyID            string       `json:"spyId"`
	TurnOrder        []string     `json:"turnOrder"`
	CurrentTurnIndex int          `json:"currentTurnIndex"`
	Phase            SessionPhase `json:"phase"`
}

func NewSession(playerIDs []string, locations []string) (*Session, error) {
	location, err := pickRandom(locations)
	if err != nil {
		return nil, fmt.Errorf("pick location: %w", err)
	}

	spyID, err := pickRandom(playerIDs)
	if err != nil {
		return nil, fmt.Errorf("pick spy: %w", err)
	}

	turnOrder, err := shuffleCopy(playerIDs)
	if err != nil {
		return nil, fmt.Errorf("shuffle turn order: %w", err)
	}

	return &Session{
		ID:               uuid.New().String(),
		Location:         location,
		SpyID:            spyID,
		TurnOrder:        turnOrder,
		CurrentTurnIndex: 0,
		Phase:            PhaseQuestion,
	}, nil
}

func pickRandom[T any](items []T) (T, error) {
	var zero T
	if len(items) == 0 {
		return zero, fmt.Errorf("empty slice")
	}
	n, err := rand.Int(rand.Reader, big.NewInt(int64(len(items))))
	if err != nil {
		return zero, err
	}
	return items[n.Int64()], nil
}

func shuffleCopy(items []string) ([]string, error) {
	result := make([]string, len(items))
	copy(result, items)

	for i := len(result) - 1; i > 0; i-- {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		if err != nil {
			return nil, err
		}
		j := n.Int64()
		result[i], result[j] = result[j], result[i]
	}
	return result, nil
}
