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
	return NewLosingStreakSignal(userID, timestamp, int(value), context)
}

// LosingStreakSignal represents a player losing a match.
type LosingStreakSignal struct {
	signal.BaseSignal
	CurrentStreak int
}

// NewLosingStreakSignal creates a new losing streak signal.
func NewLosingStreakSignal(userID string, timestamp time.Time, currentStreak int, context *signal.PlayerContext) *LosingStreakSignal {
	metadata := map[string]interface{}{
		"current_streak": currentStreak,
		"stat_code":      "rse-current-losing-streak",
	}
	return &LosingStreakSignal{
		BaseSignal:    signal.NewBaseSignal(TypeLosingStreak, userID, timestamp, metadata, context),
		CurrentStreak: currentStreak,
	}
}
