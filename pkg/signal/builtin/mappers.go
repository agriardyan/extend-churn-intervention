package builtin

import (
	"time"

	"github.com/AccelByte/extends-anti-churn/pkg/signal"
)

// RageQuitMapper maps "rse-rage-quit" stat events to RageQuitSignal.
type RageQuitMapper struct{}

func (m *RageQuitMapper) StatCode() string {
	return "rse-rage-quit"
}

func (m *RageQuitMapper) MapToSignal(userID string, timestamp time.Time, value float64, context *signal.PlayerContext) signal.Signal {
	return NewRageQuitSignal(userID, timestamp, int(value), context)
}

// MatchWinMapper maps "rse-match-wins" stat events to WinSignal.
type MatchWinMapper struct{}

func (m *MatchWinMapper) StatCode() string {
	return "rse-match-wins"
}

func (m *MatchWinMapper) MapToSignal(userID string, timestamp time.Time, value float64, context *signal.PlayerContext) signal.Signal {
	return NewWinSignal(userID, timestamp, int(value), context)
}

// LosingStreakMapper maps "rse-current-losing-streak" stat events to LossSignal.
type LosingStreakMapper struct{}

func (m *LosingStreakMapper) StatCode() string {
	return "rse-current-losing-streak"
}

func (m *LosingStreakMapper) MapToSignal(userID string, timestamp time.Time, value float64, context *signal.PlayerContext) signal.Signal {
	return NewLossSignal(userID, timestamp, int(value), context)
}

// RegisterBuiltinMappers registers all built-in signal mappers with the registry.
func RegisterBuiltinMappers(registry *signal.MapperRegistry) {
	registry.Register(&RageQuitMapper{})
	registry.Register(&MatchWinMapper{})
	registry.Register(&LosingStreakMapper{})
}
