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
	// churnStateStoreDefaultTTL is the default TTL for player state in Redis (30 days)
	churnStateStoreDefaultTTL = 30 * 24 * time.Hour
	// churnStateStoreKeyPrefix is the prefix for all churn state keys
	churnStateStoreKeyPrefix = "churn_intervention:user_state:"
)

// RedisChurnStateStore implements StateStore using Redis.
type RedisChurnStateStore struct {
	client *redis.Client
	cfg    RedisChurnStateStoreConfig
}

type RedisChurnStateStoreConfig struct{}

// NewRedisChurnStateStore creates a new Redis-backed state store.
func NewRedisChurnStateStore(
	client *redis.Client,
	cfg RedisChurnStateStoreConfig,
) *RedisChurnStateStore {
	return &RedisChurnStateStore{
		client: client,
		cfg:    cfg,
	}
}

// makeChurnStateStoreKey creates a Redis key for a player
func makeChurnStateStoreKey(userID string) string {
	return fmt.Sprintf("%s%s", churnStateStoreKeyPrefix, userID)
}

// GetChurnState retrieves the churn state for a player from Redis
func (r *RedisChurnStateStore) GetChurnState(ctx context.Context, userID string) (*ChurnState, error) {
	key := makeChurnStateStoreKey(userID)

	data, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		// Player doesn't exist, return new state
		logrus.Infof("no existing state for user %s, returning new state", userID)
		return &ChurnState{
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
func (r *RedisChurnStateStore) UpdateChurnState(ctx context.Context, userID string, state *ChurnState) error {
	key := makeChurnStateStoreKey(userID)

	data, err := json.Marshal(state)
	if err != nil {
		logrus.Errorf("failed to marshal state for user %s: %v", userID, err)
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	if err := r.client.Set(ctx, key, data, churnStateStoreDefaultTTL).Err(); err != nil {
		logrus.Errorf("failed to set state for user %s: %v", userID, err)
		return fmt.Errorf("failed to set state: %w", err)
	}

	logrus.Infof("updated state for user %s with TTL %v", userID, churnStateStoreDefaultTTL)
	return nil
}

// DeleteChurnState deletes the churn state for a player from Redis
func (r *RedisChurnStateStore) DeleteChurnState(ctx context.Context, userID string) error {
	key := makeChurnStateStoreKey(userID)

	if err := r.client.Del(ctx, key).Err(); err != nil {
		logrus.Errorf("failed to delete state for user %s: %v", userID, err)
		return fmt.Errorf("failed to delete state: %w", err)
	}

	logrus.Infof("deleted state for user %s", userID)
	return nil
}
