// Copyright (c) 2025 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package state

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

// HealthChecker provides Redis health check functionality
type HealthChecker struct {
	client *redis.Client
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(client *redis.Client) *HealthChecker {
	return &HealthChecker{client: client}
}

// Check performs a Redis health check
func (h *HealthChecker) Check(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	_, err := h.client.Ping(ctx).Result()
	if err != nil {
		logrus.Errorf("Redis health check failed: %v", err)
		return err
	}

	logrus.Debugf("Redis health check passed")
	return nil
}

// IsHealthy returns true if Redis is accessible
func (h *HealthChecker) IsHealthy(ctx context.Context) bool {
	return h.Check(ctx) == nil
}
