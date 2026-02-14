package builtin

import (
	"context"
	"fmt"
	"time"

	statistic "github.com/AccelByte/extends-anti-churn/pkg/pb/accelbyte-asyncapi/social/statistic/v1"
	"github.com/AccelByte/extends-anti-churn/pkg/signal"
)

// Signal type constants for built-in signals
const (
	TypeRageQuit = "rage_quit"
)

// RageQuitEventProcessor processes "rse-rage-quit" stat events into RageQuitSignal.
type RageQuitEventProcessor struct{}

func (p *RageQuitEventProcessor) EventType() string {
	return "rse-rage-quit"
}

func (p *RageQuitEventProcessor) Process(ctx context.Context, event interface{}, loader signal.PlayerContextLoader) (signal.Signal, error) {
	statEvent, ok := event.(*statistic.StatItemUpdated)
	if !ok {
		return nil, fmt.Errorf("expected *statistic.StatItemUpdated, got %T", event)
	}

	userID := statEvent.GetUserId()
	value := statEvent.GetPayload().GetLatestValue()

	playerCtx, err := loader.Load(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to load player context: %w", err)
	}

	return NewRageQuitSignal(userID, time.Now(), int(value), playerCtx), nil
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
