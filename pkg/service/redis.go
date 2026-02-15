package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

const (
	// DefaultTTL is the default TTL for player state in Redis (30 days)
	DefaultTTL = 30 * 24 * time.Hour
	// KeyPrefix is the prefix for all churn state keys
	KeyPrefix = "churn_intervention:user_state:"
)

// RedisStateStore implements StateStore using Redis.
type RedisStateStore struct {
	client *redis.Client
	cfg    RedisStateStoreConfig
}

type RedisStateStoreConfig struct{}

// NewRedisStateStore creates a new Redis-backed state store.
func NewRedisStateStore(
	client *redis.Client,
	cfg RedisStateStoreConfig,
) *RedisStateStore {
	return &RedisStateStore{
		client: client,
		cfg:    cfg,
	}
}

// makeKey creates a Redis key for a player
func makeKey(userID string) string {
	return fmt.Sprintf("%s%s", KeyPrefix, userID)
}

// GetChurnState retrieves the churn state for a player from Redis
func (r *RedisStateStore) GetChurnState(ctx context.Context, userID string) (*ChurnState, error) {
	key := makeKey(userID)

	data, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		// Player doesn't exist, return new state
		logrus.Infof("no existing state for user %s, returning new state", userID)
		return &ChurnState{
			Sessions: SessionState{
				ThisWeek:  0,
				LastWeek:  0,
				LastReset: time.Now(),
			},
			SignalHistory:       []ChurnSignal{},
			InterventionHistory: []InterventionRecord{},
			Cooldown: CooldownState{
				LastInterventionAt: time.Time{},
				CooldownUntil:      time.Time{},
				InterventionCounts: make(map[string]int),
				LastSignalAt:       make(map[string]time.Time),
			},
		}, nil
	}
	if err != nil {
		logrus.Errorf("failed to get state for user %s: %v", userID, err)
		return nil, fmt.Errorf("failed to get state: %w", err)
	}

	var state ChurnState
	if err := json.Unmarshal([]byte(data), &state); err != nil {
		logrus.Errorf("failed to unmarshal state for user %s: %v", userID, err)
		return nil, fmt.Errorf("failed to unmarshal state: %w", err)
	}

	logrus.Infof("retrieved state for user %s", userID)
	return &state, nil
}

// UpdateChurnState updates the churn state for a player in Redis
func (r *RedisStateStore) UpdateChurnState(ctx context.Context, userID string, state *ChurnState) error {
	key := makeKey(userID)

	data, err := json.Marshal(state)
	if err != nil {
		logrus.Errorf("failed to marshal state for user %s: %v", userID, err)
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	if err := r.client.Set(ctx, key, data, DefaultTTL).Err(); err != nil {
		logrus.Errorf("failed to set state for user %s: %v", userID, err)
		return fmt.Errorf("failed to set state: %w", err)
	}

	logrus.Infof("updated state for user %s with TTL %v", userID, DefaultTTL)
	return nil
}

// DeleteChurnState deletes the churn state for a player from Redis
func (r *RedisStateStore) DeleteChurnState(ctx context.Context, userID string) error {
	key := makeKey(userID)

	if err := r.client.Del(ctx, key).Err(); err != nil {
		logrus.Errorf("failed to delete state for user %s: %v", userID, err)
		return fmt.Errorf("failed to delete state: %w", err)
	}

	logrus.Infof("deleted state for user %s", userID)
	return nil
}
