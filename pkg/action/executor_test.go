package action

import (
	"context"
	"testing"

	"github.com/AccelByte/extends-anti-churn/pkg/rule"
	"github.com/AccelByte/extends-anti-churn/pkg/signal"
	"github.com/AccelByte/extends-anti-churn/pkg/state"
)

// testAction is a simple action for testing
type testAction struct {
	id             string
	name           string
	config         ActionConfig
	executeFunc    func(ctx context.Context, trigger *rule.Trigger, playerCtx *signal.PlayerContext) error
	rollbackFunc   func(ctx context.Context, trigger *rule.Trigger, playerCtx *signal.PlayerContext) error
	executeCalled  bool
	rollbackCalled bool
}

func (a *testAction) ID() string           { return a.id }
func (a *testAction) Name() string         { return a.name }
func (a *testAction) Config() ActionConfig { return a.config }

func (a *testAction) Execute(ctx context.Context, trigger *rule.Trigger, playerCtx *signal.PlayerContext) error {
	a.executeCalled = true
	if a.executeFunc != nil {
		return a.executeFunc(ctx, trigger, playerCtx)
	}
	return nil
}

func (a *testAction) Rollback(ctx context.Context, trigger *rule.Trigger, playerCtx *signal.PlayerContext) error {
	a.rollbackCalled = true
	if a.rollbackFunc != nil {
		return a.rollbackFunc(ctx, trigger, playerCtx)
	}
	return nil
}

func TestNewExecutor(t *testing.T) {
	registry := NewRegistry()
	executor := NewExecutor(registry)

	if executor == nil {
		t.Fatal("Expected non-nil executor")
	}

	if executor.GetRegistry() != registry {
		t.Error("Expected executor to use provided registry")
	}
}

func TestExecutor_Execute_Success(t *testing.T) {
	registry := NewRegistry()
	executor := NewExecutor(registry)

	action := &testAction{
		id:     "test_action",
		name:   "Test Action",
		config: ActionConfig{ID: "test_action", Enabled: true},
	}
	registry.Register(action)

	trigger := rule.NewTrigger("test_rule", "test-user", "test reason", 10)
	playerCtx := &signal.PlayerContext{
		UserID: "test-user",
		State:  &state.ChurnState{},
	}

	result, err := executor.Execute(context.Background(), "test_action", trigger, playerCtx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !result.Success {
		t.Error("Expected successful result")
	}

	if !action.executeCalled {
		t.Error("Expected action Execute to be called")
	}
}

func TestExecutor_Execute_ActionNotFound(t *testing.T) {
	registry := NewRegistry()
	executor := NewExecutor(registry)

	trigger := rule.NewTrigger("test_rule", "test-user", "test reason", 10)
	playerCtx := &signal.PlayerContext{
		UserID: "test-user",
		State:  &state.ChurnState{},
	}

	result, err := executor.Execute(context.Background(), "nonexistent_action", trigger, playerCtx)
	if err == nil {
		t.Error("Expected error for nonexistent action")
	}

	if result != nil {
		t.Error("Expected nil result for nonexistent action")
	}
}

func TestExecutor_Execute_ActionError(t *testing.T) {
	registry := NewRegistry()
	executor := NewExecutor(registry)

	expectedError := &testError{msg: "action failed"}
	action := &testAction{
		id:     "failing_action",
		name:   "Failing Action",
		config: ActionConfig{ID: "failing_action", Enabled: true},
		executeFunc: func(ctx context.Context, trigger *rule.Trigger, playerCtx *signal.PlayerContext) error {
			return expectedError
		},
	}
	registry.Register(action)

	trigger := rule.NewTrigger("test_rule", "test-user", "test reason", 10)
	playerCtx := &signal.PlayerContext{
		UserID: "test-user",
		State:  &state.ChurnState{},
	}

	result, err := executor.Execute(context.Background(), "failing_action", trigger, playerCtx)
	if err == nil {
		t.Error("Expected error from failing action")
	}

	if result.Success {
		t.Error("Expected unsuccessful result")
	}

	if result.Error != expectedError {
		t.Errorf("Expected error %v, got %v", expectedError, result.Error)
	}
}

func TestExecutor_ExecuteMultiple_Success(t *testing.T) {
	registry := NewRegistry()
	executor := NewExecutor(registry)

	action1 := &testAction{
		id:     "action1",
		name:   "Action 1",
		config: ActionConfig{ID: "action1", Enabled: true},
	}
	action2 := &testAction{
		id:     "action2",
		name:   "Action 2",
		config: ActionConfig{ID: "action2", Enabled: true},
	}

	registry.Register(action1)
	registry.Register(action2)

	trigger := rule.NewTrigger("test_rule", "test-user", "test reason", 10)
	playerCtx := &signal.PlayerContext{
		UserID: "test-user",
		State:  &state.ChurnState{},
	}

	results, err := executor.ExecuteMultiple(context.Background(), []string{"action1", "action2"}, trigger, playerCtx, false)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("Expected 2 results, got %d", len(results))
	}

	if !action1.executeCalled || !action2.executeCalled {
		t.Error("Expected both actions to be executed")
	}

	for _, result := range results {
		if !result.Success {
			t.Error("Expected all results to be successful")
		}
	}
}

func TestExecutor_ExecuteMultiple_WithRollback(t *testing.T) {
	registry := NewRegistry()
	executor := NewExecutor(registry)

	action1 := &testAction{
		id:     "action1",
		name:   "Action 1",
		config: ActionConfig{ID: "action1", Enabled: true},
	}
	action2 := &testAction{
		id:     "action2",
		name:   "Action 2 (Fails)",
		config: ActionConfig{ID: "action2", Enabled: true},
		executeFunc: func(ctx context.Context, trigger *rule.Trigger, playerCtx *signal.PlayerContext) error {
			return &testError{msg: "action2 failed"}
		},
	}

	registry.Register(action1)
	registry.Register(action2)

	trigger := rule.NewTrigger("test_rule", "test-user", "test reason", 10)
	playerCtx := &signal.PlayerContext{
		UserID: "test-user",
		State:  &state.ChurnState{},
	}

	results, err := executor.ExecuteMultiple(context.Background(), []string{"action1", "action2"}, trigger, playerCtx, true)
	if err == nil {
		t.Error("Expected error from failing action")
	}

	if len(results) != 2 {
		t.Fatalf("Expected 2 results (including failed action), got %d", len(results))
	}

	if !action1.executeCalled {
		t.Error("Expected action1 to be executed")
	}

	if !action1.rollbackCalled {
		t.Error("Expected action1 to be rolled back")
	}

	if action2.rollbackCalled {
		t.Error("Did not expect action2 to be rolled back (it failed)")
	}
}

func TestExecutor_ExecuteMultiple_NoRollback(t *testing.T) {
	registry := NewRegistry()
	executor := NewExecutor(registry)

	action1 := &testAction{
		id:     "action1",
		name:   "Action 1",
		config: ActionConfig{ID: "action1", Enabled: true},
	}
	action2 := &testAction{
		id:     "action2",
		name:   "Action 2 (Fails)",
		config: ActionConfig{ID: "action2", Enabled: true},
		executeFunc: func(ctx context.Context, trigger *rule.Trigger, playerCtx *signal.PlayerContext) error {
			return &testError{msg: "action2 failed"}
		},
	}

	registry.Register(action1)
	registry.Register(action2)

	trigger := rule.NewTrigger("test_rule", "test-user", "test reason", 10)
	playerCtx := &signal.PlayerContext{
		UserID: "test-user",
		State:  &state.ChurnState{},
	}

	results, err := executor.ExecuteMultiple(context.Background(), []string{"action1", "action2"}, trigger, playerCtx, false)
	if err == nil {
		t.Error("Expected error from failing action")
	}

	if len(results) != 2 {
		t.Fatalf("Expected 2 results (including failed action), got %d", len(results))
	}

	if !action1.executeCalled {
		t.Error("Expected action1 to be executed")
	}

	if action1.rollbackCalled {
		t.Error("Did not expect rollback when rollbackOnError is false")
	}
}

type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}
