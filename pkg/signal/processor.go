package signal

import (
	"context"
	"fmt"
	"time"

	oauth "github.com/AccelByte/extends-anti-churn/pkg/pb/accelbyte-asyncapi/iam/oauth/v1"
	statistic "github.com/AccelByte/extends-anti-churn/pkg/pb/accelbyte-asyncapi/social/statistic/v1"
	"github.com/AccelByte/extends-anti-churn/pkg/state"

	"github.com/go-redis/redis/v8"
)

// StateStore defines the interface for accessing player state.
// This allows for easier testing and different storage implementations.
type StateStore interface {
	GetChurnState(ctx context.Context, userID string) (*state.ChurnState, error)
	UpdateChurnState(ctx context.Context, userID string, state *state.ChurnState) error
}

// RedisStateStore implements StateStore using Redis.
type RedisStateStore struct {
	client *redis.Client
}

// NewRedisStateStore creates a new Redis-backed state store.
func NewRedisStateStore(client *redis.Client) *RedisStateStore {
	return &RedisStateStore{
		client: client,
	}
}

// GetChurnState retrieves player state from Redis.
func (r *RedisStateStore) GetChurnState(ctx context.Context, userID string) (*state.ChurnState, error) {
	return state.GetChurnState(ctx, r.client, userID)
}

// UpdateChurnState updates player state in Redis.
func (r *RedisStateStore) UpdateChurnState(ctx context.Context, userID string, churnState *state.ChurnState) error {
	return state.UpdateChurnState(ctx, r.client, userID, churnState)
}

// Processor converts raw events into domain signals with enriched context.
type Processor struct {
	stateStore             StateStore
	mapperRegistry         *MapperRegistry
	eventProcessorRegistry *EventProcessorRegistry
	namespace              string
}

// NewProcessor creates a new signal processor.
func NewProcessor(stateStore StateStore, namespace string) *Processor {
	return &Processor{
		stateStore:             stateStore,
		mapperRegistry:         NewMapperRegistry(),
		eventProcessorRegistry: NewEventProcessorRegistry(),
		namespace:              namespace,
	}
}

// GetMapperRegistry returns the mapper registry for this processor.
// This allows registering custom signal mappers.
func (p *Processor) GetMapperRegistry() *MapperRegistry {
	return p.mapperRegistry
}

// GetEventProcessorRegistry returns the event processor registry.
// This allows registering custom event processors.
func (p *Processor) GetEventProcessorRegistry() *EventProcessorRegistry {
	return p.eventProcessorRegistry
}

// LoadPlayerContext implements ContextLoader interface.
// This allows event processors to load player context.
func (p *Processor) LoadPlayerContext(ctx context.Context, userID string) (*PlayerContext, error) {
	return p.loadPlayerContext(ctx, userID)
}

// loadPlayerContext loads the player's state and wraps it in a PlayerContext.
func (p *Processor) loadPlayerContext(ctx context.Context, userID string) (*PlayerContext, error) {
	// Load player state from store
	churnState, err := p.stateStore.GetChurnState(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get churn state: %w", err)
	}

	// Build player context
	playerContext := &PlayerContext{
		UserID:      userID,
		State:       churnState,
		Namespace:   p.namespace,
		SessionInfo: make(map[string]interface{}),
	}

	// Add session metadata
	playerContext.SessionInfo["sessions_this_week"] = churnState.Sessions.ThisWeek
	playerContext.SessionInfo["sessions_last_week"] = churnState.Sessions.LastWeek
	playerContext.SessionInfo["challenge_active"] = churnState.Challenge.Active
	playerContext.SessionInfo["on_cooldown"] = time.Now().Before(churnState.Intervention.CooldownUntil)

	return playerContext, nil
}

// GetStateStore returns the state store used by this processor.
// This is useful for testing and direct state access.
func (p *Processor) GetStateStore() StateStore {
	return p.stateStore
}

// ProcessEvent processes any event type using registered event processors.
// This is the generic entry point for all event processing.
func (p *Processor) ProcessEvent(ctx context.Context, eventType string, event interface{}) (Signal, error) {
	processor := p.eventProcessorRegistry.Get(eventType)
	if processor == nil {
		return nil, fmt.Errorf("no event processor registered for event type '%s'", eventType)
	}

	return processor.Process(ctx, event, p)
}

// ProcessOAuthEvent processes OAuth token events (convenience wrapper for backward compatibility).
func (p *Processor) ProcessOAuthEvent(ctx context.Context, event *oauth.OauthTokenGenerated) (Signal, error) {
	return p.ProcessEvent(ctx, "oauth_token_generated", event)
}

// ProcessStatEvent processes statistic update events (convenience wrapper for backward compatibility).
func (p *Processor) ProcessStatEvent(ctx context.Context, event *statistic.StatItemUpdated) (Signal, error) {
	return p.ProcessEvent(ctx, "stat_item_updated", event)
}
