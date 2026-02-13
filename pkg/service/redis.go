package service

import (
	"context"

	"github.com/AccelByte/extends-anti-churn/pkg/state"
	"github.com/go-redis/redis/v8"
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

// GetChurnState retrieves player state from Redis.
func (r *RedisStateStore) GetChurnState(ctx context.Context, userID string) (*state.ChurnState, error) {
	return state.GetChurnState(ctx, r.client, userID)
}

// UpdateChurnState updates player state in Redis.
func (r *RedisStateStore) UpdateChurnState(ctx context.Context, userID string, churnState *state.ChurnState) error {
	return state.UpdateChurnState(ctx, r.client, userID, churnState)
}
