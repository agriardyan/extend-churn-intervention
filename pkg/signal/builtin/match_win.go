package builtin

import (
	"context"
	"fmt"
	"time"

	statistic "github.com/AccelByte/extends-anti-churn/pkg/pb/accelbyte-asyncapi/social/statistic/v1"
	"github.com/AccelByte/extends-anti-churn/pkg/signal"
)

const (
	TypeMatchWin = "match_win"
)

// MatchWinEventProcessor processes "rse-match-wins" stat events into WinSignal.
type MatchWinEventProcessor struct{}

func (p *MatchWinEventProcessor) EventType() string {
	return "rse-match-wins"
}

func (p *MatchWinEventProcessor) Process(ctx context.Context, event interface{}, loader signal.PlayerContextLoader) (signal.Signal, error) {
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

	return NewWinSignal(userID, time.Now(), int(value), playerCtx), nil
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
