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
	signal.BaseSignal
	TotalWins int
}

// NewWinSignal creates a new win signal.
func NewWinSignal(userID string, timestamp time.Time, totalWins int, context *signal.PlayerContext) *WinSignal {
	metadata := map[string]interface{}{
		"total_wins": totalWins,
		"stat_code":  "rse-match-wins",
	}
	return &WinSignal{
		BaseSignal: signal.NewBaseSignal(TypeMatchWin, userID, timestamp, metadata, context),
		TotalWins:  totalWins,
	}
}
