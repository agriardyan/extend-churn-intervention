package signal

import (
	"context"
	"fmt"
	"time"

	oauth "github.com/AccelByte/extend-churn-intervention/pkg/pb/accelbyte-asyncapi/iam/oauth/v1"
	statistic "github.com/AccelByte/extend-churn-intervention/pkg/pb/accelbyte-asyncapi/social/statistic/v1"
	"github.com/AccelByte/extend-churn-intervention/pkg/service"
)

// Processor converts raw events into domain signals with enriched context.
type Processor struct {
	stateStore             service.StateStore
	eventProcessorRegistry *EventProcessorRegistry
	namespace              string
}

// NewProcessor creates a new signal processor.
func NewProcessor(stateStore service.StateStore, namespace string) *Processor {
	return &Processor{
		stateStore:             stateStore,
		eventProcessorRegistry: NewEventProcessorRegistry(),
		namespace:              namespace,
	}
}

// GetEventProcessorRegistry returns the event processor registry.
// This allows registering custom event processors.
func (p *Processor) GetEventProcessorRegistry() *EventProcessorRegistry {
	return p.eventProcessorRegistry
}

// GetStateStore returns the state store used by this processor.
// This is useful for passing to event processors that need state access.
func (p *Processor) GetStateStore() service.StateStore {
	return p.stateStore
}

// GetNamespace returns the namespace for this processor.
// This is useful for passing to event processors that need namespace.
func (p *Processor) GetNamespace() string {
	return p.namespace
}

// ProcessEvent processes any event type using registered event processors.
// This is the generic entry point for all event processing.
func (p *Processor) ProcessEvent(ctx context.Context, eventType string, event interface{}) (Signal, error) {
	processor := p.eventProcessorRegistry.Get(eventType)
	if processor == nil {
		return nil, fmt.Errorf("no event processor registered for event type '%s'", eventType)
	}

	return processor.Process(ctx, event)
}

// ProcessOAuthEvent processes OAuth token events (convenience wrapper).
func (p *Processor) ProcessOAuthEvent(ctx context.Context, event *oauth.OauthTokenGenerated) (Signal, error) {
	return p.ProcessEvent(ctx, "oauth_token_generated", event)
}

// ProcessStatEvent processes statistic update events.
// Routes to stat-code-specific event processors if registered,
// otherwise falls back to a generic StatUpdateSignal.
func (p *Processor) ProcessStatEvent(ctx context.Context, event *statistic.StatItemUpdated) (Signal, error) {
	if event == nil {
		return nil, fmt.Errorf("stat event is nil")
	}

	payload := event.GetPayload()
	if payload == nil {
		return nil, fmt.Errorf("stat event payload is nil")
	}

	userID := event.GetUserId()
	statCode := payload.GetStatCode()

	if userID == "" {
		return nil, fmt.Errorf("user ID is empty in stat event")
	}
	if statCode == "" {
		return nil, fmt.Errorf("stat code is empty in stat event")
	}

	// Route to stat-code-specific processor if registered
	processor := p.eventProcessorRegistry.Get(statCode)
	if processor != nil {
		return processor.Process(ctx, event)
	}

	// Fallback: load context and create generic stat update signal
	churnState, err := p.stateStore.GetChurnState(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to load churn state for user %s: %w", userID, err)
	}

	playerCtx := BuildPlayerContext(userID, p.namespace, churnState)
	return NewStatUpdateSignal(userID, time.Now(), statCode, payload.GetLatestValue(), playerCtx), nil
}
