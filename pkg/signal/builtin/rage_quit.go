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
	signalType   string
	userID       string
	timestamp    time.Time
	metadata     map[string]interface{}
	context      *signal.PlayerContext
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
		signalType:   TypeRageQuit,
		userID:       userID,
		timestamp:    timestamp,
		metadata:     metadata,
		context:      context,
		QuitCount:    quitCount,
		MatchContext: make(map[string]interface{}),
	}
}

// Type implements Signal interface.
func (s *RageQuitSignal) Type() string {
	return s.signalType
}

// UserID implements Signal interface.
func (s *RageQuitSignal) UserID() string {
	return s.userID
}

// Timestamp implements Signal interface.
func (s *RageQuitSignal) Timestamp() time.Time {
	return s.timestamp
}

// Metadata implements Signal interface.
func (s *RageQuitSignal) Metadata() map[string]interface{} {
	return s.metadata
}

// Context implements Signal interface.
func (s *RageQuitSignal) Context() *signal.PlayerContext {
	return s.context
}
