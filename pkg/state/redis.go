// Copyright (c) 2025 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package state

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/AccelByte/extends-anti-churn/pkg/common"

	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

const (
	// DefaultTTL is the default TTL for player state in Redis (30 days)
	DefaultTTL = 30 * 24 * time.Hour
	// KeyPrefix is the prefix for all churn state keys
	KeyPrefix = "churn_intervention:user_state:"
)

// InitRedisClient initializes and returns a Redis client with retry logic
func InitRedisClient(ctx context.Context) (*redis.Client, error) {
	redisHost := common.GetEnv("REDIS_HOST", "localhost")
	redisPort := common.GetEnv("REDIS_PORT", "6379")
	redisPassword := common.GetEnv("REDIS_PASSWORD", "")
	maxRetries := common.GetEnvInt("REDIS_MAX_RETRIES", 5)
	retryDelayMs := common.GetEnvInt("REDIS_RETRY_DELAY_MS", 1000)

	client := redis.NewClient(&redis.Options{
		Addr:         redisHost + ":" + redisPort,
		Password:     redisPassword,
		DB:           0, // use default DB
		MaxRetries:   3,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	// Retry connection with exponential backoff
	for i := 0; i < maxRetries; i++ {
		_, err := client.Ping(ctx).Result()
		if err == nil {
			logrus.Infof("connected to Redis at %s:%s (attempt %d/%d)",
				redisHost, redisPort, i+1, maxRetries)
			return client, nil
		}

		if i < maxRetries-1 {
			delay := time.Duration(retryDelayMs*(i+1)) * time.Millisecond
			logrus.Warnf("Redis connection failed (attempt %d/%d): %v, retrying in %v...",
				i+1, maxRetries, err, delay)
			time.Sleep(delay)
		}
	}

	return nil, fmt.Errorf("failed to connect to Redis at %s:%s after %d attempts", redisHost, redisPort, maxRetries)
}

// makeKey creates a Redis key for a player
func makeKey(userID string) string {
	return fmt.Sprintf("%s%s", KeyPrefix, userID)
}

// GetChurnState retrieves the churn state for a player from Redis
func GetChurnState(ctx context.Context, client *redis.Client, userID string) (*ChurnState, error) {
	key := makeKey(userID)

	data, err := client.Get(ctx, key).Result()
	if err == redis.Nil {
		// Player doesn't exist, return new state
		logrus.Infof("no existing state for user %s, returning new state", userID)
		return &ChurnState{
			Sessions: SessionState{
				ThisWeek:  0,
				LastWeek:  0,
				LastReset: time.Now(),
			},
			Challenge: ChallengeState{
				Active: false,
			},
			Intervention: InterventionState{
				TotalTriggered: 0,
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
func UpdateChurnState(ctx context.Context, client *redis.Client, userID string, state *ChurnState) error {
	key := makeKey(userID)

	data, err := json.Marshal(state)
	if err != nil {
		logrus.Errorf("failed to marshal state for user %s: %v", userID, err)
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	if err := client.Set(ctx, key, data, DefaultTTL).Err(); err != nil {
		logrus.Errorf("failed to set state for user %s: %v", userID, err)
		return fmt.Errorf("failed to set state: %w", err)
	}

	logrus.Infof("updated state for user %s with TTL %v", userID, DefaultTTL)
	return nil
}

// DeleteChurnState deletes the churn state for a player from Redis
func DeleteChurnState(ctx context.Context, client *redis.Client, userID string) error {
	key := makeKey(userID)

	if err := client.Del(ctx, key).Err(); err != nil {
		logrus.Errorf("failed to delete state for user %s: %v", userID, err)
		return fmt.Errorf("failed to delete state: %w", err)
	}

	logrus.Infof("deleted state for user %s", userID)
	return nil
}
