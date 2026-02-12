package signal

import (
	"context"
	"fmt"
	"time"

	oauth "github.com/AccelByte/extends-anti-churn/pkg/pb/accelbyte-asyncapi/iam/oauth/v1"
	statistic "github.com/AccelByte/extends-anti-churn/pkg/pb/accelbyte-asyncapi/social/statistic/v1"
	"github.com/AccelByte/extends-anti-churn/pkg/state"

	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
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
	stateStore     StateStore
	mapperRegistry *MapperRegistry
	namespace      string
}

// NewProcessor creates a new signal processor.
func NewProcessor(stateStore StateStore, namespace string) *Processor {
	return &Processor{
		stateStore:     stateStore,
		mapperRegistry: NewMapperRegistry(),
		namespace:      namespace,
	}
}

// GetMapperRegistry returns the mapper registry for this processor.
// This allows registering custom signal mappers.
func (p *Processor) GetMapperRegistry() *MapperRegistry {
	return p.mapperRegistry
}

// ProcessOAuthEvent converts an OAuth token event into a LoginSignal.
func (p *Processor) ProcessOAuthEvent(ctx context.Context, event *oauth.OauthTokenGenerated) (Signal, error) {
	if event == nil {
		return nil, fmt.Errorf("oauth event is nil")
	}

	userID := event.GetUserId()
	if userID == "" {
		return nil, fmt.Errorf("user ID is empty in oauth event")
	}

	// Load player context
	playerCtx, err := p.loadPlayerContext(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to load player context for user %s: %w", userID, err)
	}

	// Create a generic login signal using StatUpdateSignal as fallback
	// The actual LoginSignal should be created by registering a custom mapper
	// that processes OAuth events, but for now we return a simple signal
	metadata := map[string]interface{}{
		"event": "oauth_token_generated",
	}
	signal := &BaseSignal{
		signalType: "login",
		userID:     userID,
		timestamp:  time.Now(),
		metadata:   metadata,
		context:    playerCtx,
	}

	logrus.Debugf("processed OAuth event for user %s into login signal", userID)
	return signal, nil
}

// ProcessStatEvent converts a statistic update event into appropriate signals.
// Uses registered SignalMappers for extensible stat-to-signal conversion.
func (p *Processor) ProcessStatEvent(ctx context.Context, event *statistic.StatItemUpdated) (Signal, error) {
	if event == nil {
		return nil, fmt.Errorf("stat event is nil")
	}

	payload := event.GetPayload()
	if payload == nil {
		return nil, fmt.Errorf("stat event payload is nil")
	}

	userID := event.GetUserId() // User ID is in the event, not the payload
	statCode := payload.GetStatCode()
	value := payload.GetLatestValue()

	if userID == "" {
		return nil, fmt.Errorf("user ID is empty in stat event")
	}
	if statCode == "" {
		return nil, fmt.Errorf("stat code is empty in stat event")
	}

	// Load player context
	playerCtx, err := p.loadPlayerContext(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to load player context for user %s: %w", userID, err)
	}

	timestamp := time.Now()

	// Try to find a registered mapper for this stat code
	mapper := p.mapperRegistry.Get(statCode)
	if mapper != nil {
		signal := mapper.MapToSignal(userID, timestamp, value, playerCtx)
		logrus.Debugf("processed stat event for user %s into %s (code=%s)", userID, signal.Type(), statCode)
		return signal, nil
	}

	// Fallback: create generic stat update signal for unknown stat codes
	signal := NewStatUpdateSignal(userID, timestamp, statCode, value, playerCtx)
	logrus.Debugf("processed stat event for user %s into StatUpdateSignal (code=%s, value=%f)", userID, statCode, value)
	return signal, nil
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
