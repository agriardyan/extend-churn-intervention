package handler

import (
	"context"

	"github.com/AccelByte/extends-anti-churn/pkg/action"
	"github.com/AccelByte/extends-anti-churn/pkg/pipeline"
	"github.com/AccelByte/extends-anti-churn/pkg/rule"
	"github.com/AccelByte/extends-anti-churn/pkg/service"
	"github.com/AccelByte/extends-anti-churn/pkg/signal"
	signalBuiltin "github.com/AccelByte/extends-anti-churn/pkg/signal/builtin"
	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
)

// setupTestPipeline creates a complete test pipeline with Redis backend
func setupTestPipeline(namespace string, mr *miniredis.Miniredis) *pipeline.Manager {
	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	// Create signal processor with Redis state store
	stateStore := service.NewRedisChurnStateStore(client, service.RedisChurnStateStoreConfig{})
	loginSessionTrackingStore := service.NewRedisLoginSessionTrackingStore(client, service.RedisLoginSessionTrackingStoreConfig{})
	processor := signal.NewProcessor(stateStore, namespace)

	// Register builtin event processors
	signalBuiltin.RegisterEventProcessors(
		processor.GetEventProcessorRegistry(),
		processor.GetStateStore(),
		processor.GetNamespace(),
		&signalBuiltin.EventProcessorDependencies{
			LoginTrackingStore: loginSessionTrackingStore,
		},
	)

	// Create rule registry and engine
	ruleRegistry := rule.NewRegistry()
	engine := rule.NewEngine(ruleRegistry)

	// Create action registry and executor
	actionRegistry := action.NewRegistry()
	executor := action.NewExecutor(actionRegistry)

	// Create pipeline manager with empty rule-to-actions mapping
	ruleActions := make(map[string][]string)
	return pipeline.NewManager(processor, engine, executor, ruleActions, nil)
}

// getRedisClient returns a Redis client connected to the miniredis instance
func getRedisClient(mr *miniredis.Miniredis) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
}

// getStateFromRedis retrieves churn state from Redis
func getStateFromRedis(ctx context.Context, client *redis.Client, userID string) (*service.ChurnState, error) {
	stateStore := service.NewRedisChurnStateStore(client, service.RedisChurnStateStoreConfig{})
	return stateStore.GetChurnState(ctx, userID)
}
