package builtin

import (
	"context"
	"fmt"
	"time"

	statistic "github.com/AccelByte/extends-anti-churn/pkg/pb/accelbyte-asyncapi/social/statistic/v1"
	"github.com/AccelByte/extends-anti-churn/pkg/signal"
	"github.com/sirupsen/logrus"
)

// StatEventProcessor processes statistic update events into appropriate signals.
type StatEventProcessor struct {
	mapperRegistry *signal.MapperRegistry
}

// NewStatEventProcessor creates a new stat event processor with a mapper registry.
func NewStatEventProcessor(mapperRegistry *signal.MapperRegistry) *StatEventProcessor {
	return &StatEventProcessor{
		mapperRegistry: mapperRegistry,
	}
}

func (p *StatEventProcessor) EventType() string {
	return "stat_item_updated"
}

func (p *StatEventProcessor) Process(ctx context.Context, event interface{}, contextLoader signal.ContextLoader) (signal.Signal, error) {
	statEvent, ok := event.(*statistic.StatItemUpdated)
	if !ok {
		return nil, fmt.Errorf("expected *statistic.StatItemUpdated, got %T", event)
	}

	if statEvent == nil {
		return nil, fmt.Errorf("stat event is nil")
	}

	payload := statEvent.GetPayload()
	if payload == nil {
		return nil, fmt.Errorf("stat event payload is nil")
	}

	userID := statEvent.GetUserId()
	statCode := payload.GetStatCode()
	value := payload.GetLatestValue()

	if userID == "" {
		return nil, fmt.Errorf("user ID is empty in stat event")
	}
	if statCode == "" {
		return nil, fmt.Errorf("stat code is empty in stat event")
	}

	// Load player context
	playerCtx, err := contextLoader.LoadPlayerContext(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to load player context for user %s: %w", userID, err)
	}

	timestamp := time.Now()

	// Try to find a registered mapper for this stat code
	mapper := p.mapperRegistry.Get(statCode)
	if mapper != nil {
		sig := mapper.MapToSignal(userID, timestamp, value, playerCtx)
		logrus.Debugf("processed stat event for user %s into %s (code=%s)", userID, sig.Type(), statCode)
		return sig, nil
	}

	// Fallback: create generic stat update signal for unknown stat codes
	sig := signal.NewStatUpdateSignal(userID, timestamp, statCode, value, playerCtx)
	logrus.Debugf("processed stat event for user %s into StatUpdateSignal (code=%s, value=%f)", userID, statCode, value)
	return sig, nil
}
