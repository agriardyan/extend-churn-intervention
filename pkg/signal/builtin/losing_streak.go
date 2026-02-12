package builtin

import (
	"time"

	"github.com/AccelByte/extends-anti-churn/pkg/signal"
)

// Signal type constants for built-in signals
const (
	TypeLosingStreak = "losing_streak"
)

// LosingStreakMapper maps "rse-current-losing-streak" stat events to LossSignal.
type LosingStreakMapper struct{}

func (m *LosingStreakMapper) StatCode() string {
	return "rse-current-losing-streak"
}

func (m *LosingStreakMapper) MapToSignal(userID string, timestamp time.Time, value float64, context *signal.PlayerContext) signal.Signal {
	return NewLossSignal(userID, timestamp, int(value), context)
}

// LossSignal represents a player losing a match.
type LossSignal struct {
	signal.BaseSignal
	CurrentStreak int
}

// NewLossSignal creates a new loss signal.
func NewLossSignal(userID string, timestamp time.Time, currentStreak int, context *signal.PlayerContext) *LossSignal {
	metadata := map[string]interface{}{
		"current_streak": currentStreak,
		"stat_code":      "rse-current-losing-streak",
	}
	return &LossSignal{
		BaseSignal:    signal.NewBaseSignal(TypeLosingStreak, userID, timestamp, metadata, context),
		CurrentStreak: currentStreak,
	}
}
