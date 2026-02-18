package service

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
)

const (
	loginSessionTrackingStoreDefaultTTL = 28 * 24 * time.Hour // 4 weeks retention
	loginSessionTrackingStoreKeyPrefix  = "session_tracking:"
)

// SessionTrackingData tracks session counts per week for login session tracking.
// Uses yearWeek format (YYYYWW) as keys to track login counts for each week.
// Retains data for 4 weeks to detect both quick and slow churn patterns.
// Example: {"202610": 5, "202609": 3} means 5 logins in week 10 of 2026, 3 in week 9.
type SessionTrackingData struct {
	LoginCount map[string]int `json:"loginCount"` // Key: yearWeek (e.g., "202610"), Value: login count
}

type RedisLoginSessionTrackingStore struct {
	client *redis.Client
	cfg    RedisLoginSessionTrackingStoreConfig
}

type RedisLoginSessionTrackingStoreConfig struct{}

func NewRedisLoginSessionTrackingStore(client *redis.Client, cfg RedisLoginSessionTrackingStoreConfig) *RedisLoginSessionTrackingStore {
	return &RedisLoginSessionTrackingStore{
		client: client,
		cfg:    cfg,
	}
}

func makeLoginSessionTrackingStoreKey(userID string) string {
	return fmt.Sprintf("%s%s", loginSessionTrackingStoreKeyPrefix, userID)
}

// getYearWeek returns the year-week string in format "YYYYWW" (e.g., "202610" for week 10 of 2026)
func getYearWeek(t time.Time) string {
	year, week := t.ISOWeek()
	return fmt.Sprintf("%04d%02d", year, week)
}

func (r *RedisLoginSessionTrackingStore) IncrementSessionCount(ctx context.Context, userID string) error {
	key := makeLoginSessionTrackingStoreKey(userID)
	yearWeek := getYearWeek(time.Now())

	// Atomic increment using HINCRBY
	err := r.client.HIncrBy(ctx, key, yearWeek, 1).Err()
	if err != nil {
		return fmt.Errorf("failed to increment session count: %w", err)
	}

	// Cleanup old weeks (older than 4 weeks from current)
	allWeeks, err := r.client.HKeys(ctx, key).Result()
	if err == nil && len(allWeeks) > 0 {
		// Calculate the threshold (4 weeks ago)
		fourWeeksAgo := getYearWeek(time.Now().Add(-4 * 7 * 24 * time.Hour))

		var toDelete []string
		for _, week := range allWeeks {
			if week < fourWeeksAgo {
				toDelete = append(toDelete, week)
			}
		}

		if len(toDelete) > 0 {
			r.client.HDel(ctx, key, toDelete...)
		}
	}

	// Set TTL on the entire hash
	r.client.Expire(ctx, key, loginSessionTrackingStoreDefaultTTL)

	return nil
}

// GetSessionData retrieves session tracking data for a user from Redis.
// Returns new tracking data with empty map if none exists.
func (r *RedisLoginSessionTrackingStore) GetSessionData(ctx context.Context, userID string) (*SessionTrackingData, error) {
	key := makeLoginSessionTrackingStoreKey(userID)

	// Get all fields from hash using HGETALL
	data, err := r.client.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get session data: %w", err)
	}

	// Convert string values to int
	loginCount := make(map[string]int)
	for week, countStr := range data {
		count, err := strconv.Atoi(countStr)
		if err != nil {
			// Skip invalid entries
			continue
		}
		loginCount[week] = count
	}

	return &SessionTrackingData{
		LoginCount: loginCount,
	}, nil
}

// SaveSessionData saves session tracking data for a user to Redis.
// Uses HSET to store the map as a hash.
func (r *RedisLoginSessionTrackingStore) SaveSessionData(ctx context.Context, userID string, data *SessionTrackingData) error {
	key := makeLoginSessionTrackingStoreKey(userID)

	// Delete existing hash first
	r.client.Del(ctx, key)

	// Set all fields in the hash
	if len(data.LoginCount) > 0 {
		// Convert map to []interface{} for HSET
		fields := make([]interface{}, 0, len(data.LoginCount)*2)
		for week, count := range data.LoginCount {
			fields = append(fields, week, count)
		}

		if err := r.client.HSet(ctx, key, fields...).Err(); err != nil {
			return fmt.Errorf("failed to set session data: %w", err)
		}

		// Set TTL
		r.client.Expire(ctx, key, loginSessionTrackingStoreDefaultTTL)
	}

	return nil
}
