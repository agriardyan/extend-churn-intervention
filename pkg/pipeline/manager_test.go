package pipeline_test

import (
	"context"
	"errors"
	"testing"

	"github.com/AccelByte/extends-anti-churn/pkg/action"
	asyncapi_iam "github.com/AccelByte/extends-anti-churn/pkg/pb/accelbyte-asyncapi/iam/oauth/v1"
	asyncapi_social "github.com/AccelByte/extends-anti-churn/pkg/pb/accelbyte-asyncapi/social/statistic/v1"
	"github.com/AccelByte/extends-anti-churn/pkg/pipeline"
	"github.com/AccelByte/extends-anti-churn/pkg/rule"
	"github.com/AccelByte/extends-anti-churn/pkg/service"
	"github.com/AccelByte/extends-anti-churn/pkg/signal"
	signalBuiltin "github.com/AccelByte/extends-anti-churn/pkg/signal/builtin"
	"github.com/AccelByte/extends-anti-churn/pkg/state"
)

// mockRule for testing
type mockRule struct {
	id          string
	shouldMatch bool
	actionIDs   []string
}

func (m *mockRule) ID() string {
	return m.id
}

func (m *mockRule) Name() string {
	return "Mock Rule"
}

func (m *mockRule) SignalTypes() []string {
	return []string{"rage_quit", "login"}
}

func (m *mockRule) Evaluate(ctx context.Context, sig signal.Signal) (bool, *rule.Trigger, error) {
	if !m.shouldMatch {
		return false, nil, nil
	}

	trigger := &rule.Trigger{
		RuleID:    m.id,
		UserID:    sig.UserID(),
		Timestamp: sig.Timestamp(),
		Reason:    "mock rule triggered",
		Metadata: map[string]any{
			"actions": m.actionIDs,
		},
		Priority: 1,
	}
	return true, trigger, nil
}

func (m *mockRule) Config() rule.RuleConfig {
	return rule.RuleConfig{
		ID:      m.id,
		Type:    "mock",
		Enabled: true,
	}
}

// mockAction for testing
type mockAction struct {
	id         string
	shouldFail bool
}

func (m *mockAction) ID() string {
	return m.id
}

func (m *mockAction) Name() string {
	return "Mock Action"
}

func (m *mockAction) Execute(ctx context.Context, trigger *rule.Trigger, playerCtx *signal.PlayerContext) error {
	if m.shouldFail {
		return errors.New("mock action failed")
	}
	return nil
}

func (m *mockAction) Rollback(ctx context.Context, trigger *rule.Trigger, playerCtx *signal.PlayerContext) error {
	return nil
}

func (m *mockAction) Config() action.ActionConfig {
	return action.ActionConfig{
		ID:      m.id,
		Type:    "mock",
		Enabled: true,
	}
}

// mockStateStore for testing
type mockStateStore struct {
	state *state.ChurnState
}

func (m *mockStateStore) GetChurnState(ctx context.Context, userID string) (*state.ChurnState, error) {
	if m.state != nil {
		return m.state, nil
	}
	return &state.ChurnState{}, nil
}

func (m *mockStateStore) UpdateChurnState(ctx context.Context, userID string, state *state.ChurnState) error {
	m.state = state
	return nil
}

// setupTestProcessor creates a processor with builtin event processors registered
func setupTestProcessor(stateStore service.StateStore) *signal.Processor {
	processor := signal.NewProcessor(stateStore, "test")

	// Register builtin event processors
	signalBuiltin.RegisterEventProcessors(
		processor.GetEventProcessorRegistry(),
		processor.GetStateStore(),
		processor.GetNamespace(),
	)

	return processor
}

func TestNewManager(t *testing.T) {
	stateStore := &mockStateStore{}
	processor := setupTestProcessor(stateStore)

	ruleRegistry := rule.NewRegistry()
	engine := rule.NewEngine(ruleRegistry)

	actionRegistry := action.NewRegistry()
	executor := action.NewExecutor(actionRegistry)

	manager := pipeline.NewManager(processor, engine, executor, nil, nil)

	if manager == nil {
		t.Fatal("expected manager to be created")
	}
}

func TestProcessOAuthEvent_NoSignal(t *testing.T) {
	ctx := context.Background()
	stateStore := &mockStateStore{}
	processor := setupTestProcessor(stateStore)

	ruleRegistry := rule.NewRegistry()
	engine := rule.NewEngine(ruleRegistry)

	actionRegistry := action.NewRegistry()
	executor := action.NewExecutor(actionRegistry)

	manager := pipeline.NewManager(processor, engine, executor, nil, nil)

	// Event that generates a login signal
	event := &asyncapi_iam.OauthTokenGenerated{
		UserId:    "test-user",
		Namespace: "test",
	}

	// Should succeed even though no signal is generated
	err := manager.ProcessOAuthEvent(ctx, event)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestProcessOAuthEvent_WithRuleTrigger(t *testing.T) {
	ctx := context.Background()

	// Setup state store with session data to trigger session decline
	stateStore := &mockStateStore{
		state: &state.ChurnState{
			Sessions: state.SessionState{
				ThisWeek: 10,
				LastWeek: 20, // 50% decline
			},
		},
	}
	processor := setupTestProcessor(stateStore)

	mockRule := &mockRule{
		id:          "test-rule",
		shouldMatch: true,
		actionIDs:   []string{"test-action"},
	}

	ruleRegistry := rule.NewRegistry()
	ruleRegistry.Register(mockRule)
	engine := rule.NewEngine(ruleRegistry)

	// Setup action
	mockAction := &mockAction{
		id:         "test-action",
		shouldFail: false,
	}

	actionRegistry := action.NewRegistry()
	actionRegistry.Register(mockAction)
	executor := action.NewExecutor(actionRegistry)

	manager := pipeline.NewManager(processor, engine, executor, nil, nil)

	event := &asyncapi_iam.OauthTokenGenerated{
		UserId:    "test-user",
		Namespace: "test",
	}

	err := manager.ProcessOAuthEvent(ctx, event)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestProcessStatEvent_RageQuit(t *testing.T) {
	ctx := context.Background()

	// Setup state store - processor will track consecutive losses
	stateStore := &mockStateStore{
		state: &state.ChurnState{},
	}
	processor := setupTestProcessor(stateStore)

	// Setup rule that will match rage quit
	mockRule := &mockRule{
		id:          "rage-quit-rule",
		shouldMatch: true,
		actionIDs:   []string{"comeback-challenge"},
	}

	ruleRegistry := rule.NewRegistry()
	ruleRegistry.Register(mockRule)
	engine := rule.NewEngine(ruleRegistry)

	// Setup action
	mockAction := &mockAction{
		id:         "comeback-challenge",
		shouldFail: false,
	}

	actionRegistry := action.NewRegistry()
	actionRegistry.Register(mockAction)
	executor := action.NewExecutor(actionRegistry)

	manager := pipeline.NewManager(processor, engine, executor, nil, nil)

	event := &asyncapi_social.StatItemUpdated{
		UserId:    "test-user",
		Namespace: "test",
		Payload: &asyncapi_social.StatItem{
			StatCode: "match-loss",
			UserId:   "test-user",
		},
	}

	err := manager.ProcessStatEvent(ctx, event)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestProcessStatEvent_NoRuleTrigger(t *testing.T) {
	ctx := context.Background()

	stateStore := &mockStateStore{}
	processor := setupTestProcessor(stateStore)

	// Setup rule that won't match
	mockRule := &mockRule{
		id:          "no-match-rule",
		shouldMatch: false,
	}

	ruleRegistry := rule.NewRegistry()
	ruleRegistry.Register(mockRule)
	engine := rule.NewEngine(ruleRegistry)

	actionRegistry := action.NewRegistry()
	executor := action.NewExecutor(actionRegistry)

	manager := pipeline.NewManager(processor, engine, executor, nil, nil)

	event := &asyncapi_social.StatItemUpdated{
		UserId:    "test-user",
		Namespace: "test",
		Payload: &asyncapi_social.StatItem{
			StatCode: "match-win",
			UserId:   "test-user",
		},
	}

	err := manager.ProcessStatEvent(ctx, event)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestProcessStatEvent_MultipleActions(t *testing.T) {
	ctx := context.Background()

	stateStore := &mockStateStore{
		state: &state.ChurnState{},
	}
	processor := setupTestProcessor(stateStore)

	// Setup rule with multiple actions
	mockRule := &mockRule{
		id:          "multi-action-rule",
		shouldMatch: true,
		actionIDs:   []string{"action1", "action2"},
	}

	ruleRegistry := rule.NewRegistry()
	ruleRegistry.Register(mockRule)
	engine := rule.NewEngine(ruleRegistry)

	// Setup multiple actions
	action1 := &mockAction{id: "action1", shouldFail: false}
	action2 := &mockAction{id: "action2", shouldFail: false}

	actionRegistry := action.NewRegistry()
	actionRegistry.Register(action1)
	actionRegistry.Register(action2)
	executor := action.NewExecutor(actionRegistry)

	manager := pipeline.NewManager(processor, engine, executor, nil, nil)

	event := &asyncapi_social.StatItemUpdated{
		UserId:    "test-user",
		Namespace: "test",
		Payload: &asyncapi_social.StatItem{
			StatCode: "match-loss",
			UserId:   "test-user",
		},
	}

	err := manager.ProcessStatEvent(ctx, event)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestProcessStatEvent_ActionFailure(t *testing.T) {
	ctx := context.Background()

	stateStore := &mockStateStore{
		state: &state.ChurnState{},
	}
	processor := setupTestProcessor(stateStore)

	mockRule := &mockRule{
		id:          "failing-action-rule",
		shouldMatch: true,
		actionIDs:   []string{"failing-action"},
	}

	ruleRegistry := rule.NewRegistry()
	ruleRegistry.Register(mockRule)
	engine := rule.NewEngine(ruleRegistry)

	// Setup action that will fail
	mockAction := &mockAction{
		id:         "failing-action",
		shouldFail: true,
	}

	actionRegistry := action.NewRegistry()
	actionRegistry.Register(mockAction)
	executor := action.NewExecutor(actionRegistry)

	manager := pipeline.NewManager(processor, engine, executor, nil, nil)

	event := &asyncapi_social.StatItemUpdated{
		UserId:    "test-user",
		Namespace: "test",
		Payload: &asyncapi_social.StatItem{
			StatCode: "match-loss",
			UserId:   "test-user",
		},
	}

	// Pipeline should not fail even if action fails
	err := manager.ProcessStatEvent(ctx, event)
	if err != nil {
		t.Fatalf("expected pipeline to continue despite action failure, got: %v", err)
	}
}

func TestGetStats(t *testing.T) {
	stateStore := &mockStateStore{}
	processor := setupTestProcessor(stateStore)

	ruleRegistry := rule.NewRegistry()
	engine := rule.NewEngine(ruleRegistry)

	actionRegistry := action.NewRegistry()
	executor := action.NewExecutor(actionRegistry)

	manager := pipeline.NewManager(processor, engine, executor, nil, nil)

	stats := manager.GetStats()

	// Just verify the structure exists
	if stats.ProcessorStats.TotalEventsProcessed < 0 {
		t.Error("invalid processor stats")
	}
}
