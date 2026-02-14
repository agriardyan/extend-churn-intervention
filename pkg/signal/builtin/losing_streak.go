package builtin

import (
	"context"
	"fmt"
	"time"

	statistic "github.com/AccelByte/extends-anti-churn/pkg/pb/accelbyte-asyncapi/social/statistic/v1"
	"github.com/AccelByte/extends-anti-churn/pkg/service"
	"github.com/AccelByte/extends-anti-churn/pkg/signal"
)

// Signal type constants for built-in signals
const (
	TypeLosingStreak = "losing_streak"
)

// LosingStreakEventProcessor processes "rse-current-losing-streak" stat events into LosingStreakSignal.
type LosingStreakEventProcessor struct {
	stateStore service.StateStore
	namespace  string
}

// NewLosingStreakEventProcessor creates a new losing streak event processor.
func NewLosingStreakEventProcessor(stateStore service.StateStore, namespace string) *LosingStreakEventProcessor {
	return &LosingStreakEventProcessor{
		stateStore: stateStore,
		namespace:  namespace,
	}
}

func (p *LosingStreakEventProcessor) EventType() string {
	return "rse-current-losing-streak"
}

func (p *LosingStreakEventProcessor) Process(ctx context.Context, event interface{}) (signal.Signal, error) {
	statEvent, ok := event.(*statistic.StatItemUpdated)
	if !ok {
		return nil, fmt.Errorf("expected *statistic.StatItemUpdated, got %T", event)
	}

	userID := statEvent.GetUserId()
	value := statEvent.GetPayload().GetLatestValue()

	// Load player state
	churnState, err := p.stateStore.GetChurnState(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to load churn state for user %s: %w", userID, err)
	}

	playerCtx := signal.BuildPlayerContext(userID, p.namespace, churnState)

	return NewLosingStreakSignal(userID, time.Now(), int(value), playerCtx), nil
}

// LosingStreakSignal represents a player losing a match.
type LosingStreakSignal struct {
	signalType    string
	userID        string
	timestamp     time.Time
	metadata      map[string]interface{}
	context       *signal.PlayerContext
	CurrentStreak int
}

// NewLosingStreakSignal creates a new losing streak signal.
func NewLosingStreakSignal(userID string, timestamp time.Time, currentStreak int, context *signal.PlayerContext) *LosingStreakSignal {
	metadata := map[string]interface{}{
		"current_streak": currentStreak,
		"stat_code":      "rse-current-losing-streak",
	}
	return &LosingStreakSignal{
		signalType:    TypeLosingStreak,
		userID:        userID,
		timestamp:     timestamp,
		metadata:      metadata,
		context:       context,
		CurrentStreak: currentStreak,
	}
}

// Type implements Signal interface.
func (s *LosingStreakSignal) Type() string {
	return s.signalType
}

// UserID implements Signal interface.
func (s *LosingStreakSignal) UserID() string {
	return s.userID
}

// Timestamp implements Signal interface.
func (s *LosingStreakSignal) Timestamp() time.Time {
	return s.timestamp
}

// Metadata implements Signal interface.
func (s *LosingStreakSignal) Metadata() map[string]interface{} {
	return s.metadata
}

// Context implements Signal interface.
func (s *LosingStreakSignal) Context() *signal.PlayerContext {
	return s.context
}
