package builtin

import (
	"time"

	"github.com/AccelByte/extends-anti-churn/pkg/signal"
)

const (
	TypeMatchWin = "match_win"
)

// MatchWinMapper maps "rse-match-wins" stat events to WinSignal.
type MatchWinMapper struct{}

func (m *MatchWinMapper) StatCode() string {
	return "rse-match-wins"
}

func (m *MatchWinMapper) MapToSignal(userID string, timestamp time.Time, value float64, context *signal.PlayerContext) signal.Signal {
	return NewWinSignal(userID, timestamp, int(value), context)
}

// WinSignal represents a player winning a match.
type WinSignal struct {
	signalType string
	userID     string
	timestamp  time.Time
	metadata   map[string]interface{}
	context    *signal.PlayerContext
	TotalWins  int
}

// NewWinSignal creates a new win signal.
func NewWinSignal(userID string, timestamp time.Time, totalWins int, context *signal.PlayerContext) *WinSignal {
	metadata := map[string]interface{}{
		"total_wins": totalWins,
		"stat_code":  "rse-match-wins",
	}
	return &WinSignal{
		signalType: TypeMatchWin,
		userID:     userID,
		timestamp:  timestamp,
		metadata:   metadata,
		context:    context,
		TotalWins:  totalWins,
	}
}

// Type implements Signal interface.
func (s *WinSignal) Type() string {
	return s.signalType
}

// UserID implements Signal interface.
func (s *WinSignal) UserID() string {
	return s.userID
}

// Timestamp implements Signal interface.
func (s *WinSignal) Timestamp() time.Time {
	return s.timestamp
}

// Metadata implements Signal interface.
func (s *WinSignal) Metadata() map[string]interface{} {
	return s.metadata
}

// Context implements Signal interface.
func (s *WinSignal) Context() *signal.PlayerContext {
	return s.context
}
