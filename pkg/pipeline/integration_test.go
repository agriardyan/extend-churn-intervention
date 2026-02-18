package pipeline

import (
	"context"
	"log/slog"
	"testing"

	"github.com/AccelByte/extends-anti-churn/pkg/action"
	asyncapi_oauth "github.com/AccelByte/extends-anti-churn/pkg/pb/accelbyte-asyncapi/iam/oauth/v1"
	asyncapi_social "github.com/AccelByte/extends-anti-churn/pkg/pb/accelbyte-asyncapi/social/statistic/v1"
	"github.com/AccelByte/extends-anti-churn/pkg/rule"
	"github.com/AccelByte/extends-anti-churn/pkg/service"
	"github.com/AccelByte/extends-anti-churn/pkg/signal"
	signalBuiltin "github.com/AccelByte/extends-anti-churn/pkg/signal/builtin"
	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
)

// setupTestProcessor creates a processor with builtin event processors registered
func setupTestProcessor(stateStore service.StateStore, loginSessionTracker service.LoginSessionTracker) *signal.Processor {
	processor := signal.NewProcessor(stateStore, "test-namespace")

	// Register builtin event processors
	signalBuiltin.RegisterEventProcessors(
		processor.GetEventProcessorRegistry(),
		processor.GetStateStore(),
		processor.GetNamespace(),
		&signalBuiltin.EventProcessorDependencies{
			LoginTrackingStore: loginSessionTracker,
		},
	)

	return processor
}

// TestIntegration_PipelineWiring tests that all pipeline components work together
func TestIntegration_PipelineWiring(t *testing.T) {
	// Create miniredis instance
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("Failed to create miniredis: %v", err)
	}
	defer mr.Close()

	// Create Redis client
	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	defer client.Close()

	// Create state store
	store := service.NewRedisChurnStateStore(client, service.RedisChurnStateStoreConfig{})
	loginSessionTracker := service.NewRedisLoginSessionTrackingStore(client, service.RedisLoginSessionTrackingStoreConfig{})

	// Create signal processor
	processor := setupTestProcessor(store, loginSessionTracker)

	// Create rule registry
	registry := rule.NewRegistry()

	// Create rule engine
	ruleEngine := rule.NewEngine(registry)

	// Create action registry
	actionRegistry := action.NewRegistry()

	// Create action executor
	executor := action.NewExecutor(actionRegistry)

	// Create rule-to-action mappings (empty for now)
	ruleActions := map[string][]string{}

	// Create logger (use nil for default logger)
	var logger *slog.Logger = nil

	// Create pipeline manager
	manager := NewManager(processor, ruleEngine, executor, ruleActions, logger)

	ctx := context.Background()

	// Test 1: Process OAuth event
	t.Run("ProcessOAuthEvent", func(t *testing.T) {
		oauthEvent := &asyncapi_oauth.OauthTokenGenerated{
			UserId: "test-user-123",
		}

		err := manager.ProcessOAuthEvent(ctx, oauthEvent)
		if err != nil {
			t.Fatalf("Failed to process OAuth event: %v", err)
		}
	})

	// Test 2: Process Stat event
	t.Run("ProcessStatEvent", func(t *testing.T) {
		statEvent := &asyncapi_social.StatItemUpdated{
			UserId: "test-user-456",
			Payload: &asyncapi_social.StatItem{
				StatCode:    "rse-rage-quit",
				LatestValue: 3.0,
			},
		}

		err := manager.ProcessStatEvent(ctx, statEvent)
		if err != nil {
			t.Fatalf("Failed to process stat event: %v", err)
		}
	})

	// Test 3: Process nil OAuth event (should return error)
	t.Run("ProcessNilOAuthEvent", func(t *testing.T) {
		err := manager.ProcessOAuthEvent(ctx, nil)
		if err == nil {
			t.Error("Expected error for nil OAuth event, got nil")
		}
	})

	// Test 4: Process nil Stat event (should return error)
	t.Run("ProcessNilStatEvent", func(t *testing.T) {
		err := manager.ProcessStatEvent(ctx, nil)
		if err == nil {
			t.Error("Expected error for nil stat event, got nil")
		}
	})
}
