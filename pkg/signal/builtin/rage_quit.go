package builtin

import (
	"time"

	"github.com/AccelByte/extends-anti-churn/pkg/signal"
)

// Signal type constants for built-in signals
const (
	TypeRageQuit = "rage_quit"
)

// RageQuitMapper maps "rse-rage-quit" stat events to RageQuitSignal.
type RageQuitMapper struct{}

func (m *RageQuitMapper) StatCode() string {
	return "rse-rage-quit"
}

func (m *RageQuitMapper) MapToSignal(userID string, timestamp time.Time, value float64, context *signal.PlayerContext) signal.Signal {
	return NewRageQuitSignal(userID, timestamp, int(value), context)
}

// RageQuitSignal represents a player rage quitting.
type RageQuitSignal struct {
	signal.BaseSignal
	QuitCount    int
	MatchContext map[string]interface{}
}

// NewRageQuitSignal creates a new rage quit signal.
func NewRageQuitSignal(userID string, timestamp time.Time, quitCount int, context *signal.PlayerContext) *RageQuitSignal {
	metadata := map[string]interface{}{
		"quit_count": quitCount,
		"stat_code":  "rse-rage-quit",
	}
	return &RageQuitSignal{
		BaseSignal:   signal.NewBaseSignal(TypeRageQuit, userID, timestamp, metadata, context),
		QuitCount:    quitCount,
		MatchContext: make(map[string]interface{}),
	}
}
