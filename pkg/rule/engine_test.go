package rule

import (
	"context"
	"testing"
	"time"

	"github.com/AccelByte/extends-anti-churn/pkg/signal"
	signalBuiltin "github.com/AccelByte/extends-anti-churn/pkg/signal/builtin"
	"github.com/AccelByte/extends-anti-churn/pkg/state"
)

// testRule is a rule that always matches for testing
type testRule struct {
	id          string
	name        string
	signalTypes []string
	config      RuleConfig
	shouldMatch bool
	shouldError bool
}

func (r *testRule) ID() string            { return r.id }
func (r *testRule) Name() string          { return r.name }
func (r *testRule) SignalTypes() []string { return r.signalTypes }
func (r *testRule) Config() RuleConfig    { return r.config }

func (r *testRule) Evaluate(ctx context.Context, sig signal.Signal) (bool, *Trigger, error) {
	if r.shouldError {
		return false, nil, &testError{msg: "test error"}
	}

	if !r.shouldMatch {
		return false, nil, nil
	}

	trigger := NewTrigger(r.id, sig.UserID(), "test trigger", r.config.Priority)
	trigger.Metadata["test"] = true

	return true, trigger, nil
}

type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}

func TestNewEngine(t *testing.T) {
	registry := NewRegistry()
	engine := NewEngine(registry)

	if engine == nil {
		t.Fatal("Expected non-nil engine")
	}

	if engine.GetRegistry() != registry {
		t.Error("Expected engine to use provided registry")
	}
}

func TestEngine_Evaluate_NoRules(t *testing.T) {
	registry := NewRegistry()
	engine := NewEngine(registry)

	playerCtx := &signal.PlayerContext{
		UserID: "test-user",
		State:  &state.ChurnState{},
	}
	sig := signalBuiltin.NewOauthTokenGeneratedSignal("test-user", time.Now(), playerCtx)

	triggers, err := engine.Evaluate(context.Background(), sig)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(triggers) != 0 {
		t.Errorf("Expected 0 triggers, got %d", len(triggers))
	}
}

func TestEngine_Evaluate_NilSignal(t *testing.T) {
	registry := NewRegistry()
	engine := NewEngine(registry)

	triggers, err := engine.Evaluate(context.Background(), nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if triggers != nil {
		t.Error("Expected nil triggers for nil signal")
	}
}

func TestEngine_Evaluate_SingleMatchingRule(t *testing.T) {
	registry := NewRegistry()
	rule := &testRule{
		id:          "test_rule",
		name:        "Test Rule",
		signalTypes: []string{"oauth_token_generated"},
		config:      RuleConfig{ID: "test_rule", Enabled: true, Priority: 10},
		shouldMatch: true,
	}
	registry.Register(rule)

	engine := NewEngine(registry)

	playerCtx := &signal.PlayerContext{
		UserID: "test-user",
		State:  &state.ChurnState{},
	}
	sig := signalBuiltin.NewOauthTokenGeneratedSignal("test-user", time.Now(), playerCtx)

	triggers, err := engine.Evaluate(context.Background(), sig)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(triggers) != 1 {
		t.Fatalf("Expected 1 trigger, got %d", len(triggers))
	}

	if triggers[0].RuleID != "test_rule" {
		t.Errorf("Expected rule ID 'test_rule', got '%s'", triggers[0].RuleID)
	}

	if triggers[0].UserID != "test-user" {
		t.Errorf("Expected user ID 'test-user', got '%s'", triggers[0].UserID)
	}
}

func TestEngine_Evaluate_MultipleMatchingRules(t *testing.T) {
	registry := NewRegistry()

	rule1 := &testRule{
		id:          "rule1",
		name:        "Rule 1",
		signalTypes: []string{"oauth_token_generated"},
		config:      RuleConfig{ID: "rule1", Enabled: true, Priority: 10},
		shouldMatch: true,
	}
	rule2 := &testRule{
		id:          "rule2",
		name:        "Rule 2",
		signalTypes: []string{"oauth_token_generated"},
		config:      RuleConfig{ID: "rule2", Enabled: true, Priority: 20},
		shouldMatch: true,
	}
	rule3 := &testRule{
		id:          "rule3",
		name:        "Rule 3",
		signalTypes: []string{"oauth_token_generated"},
		config:      RuleConfig{ID: "rule3", Enabled: true, Priority: 5},
		shouldMatch: true,
	}

	registry.Register(rule1)
	registry.Register(rule2)
	registry.Register(rule3)

	engine := NewEngine(registry)

	playerCtx := &signal.PlayerContext{
		UserID: "test-user",
		State:  &state.ChurnState{},
	}
	sig := signalBuiltin.NewOauthTokenGeneratedSignal("test-user", time.Now(), playerCtx)

	triggers, err := engine.Evaluate(context.Background(), sig)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(triggers) != 3 {
		t.Fatalf("Expected 3 triggers, got %d", len(triggers))
	}

	// Check that triggers are sorted by priority (highest first)
	if triggers[0].Priority != 20 {
		t.Errorf("Expected first trigger priority 20, got %d", triggers[0].Priority)
	}
	if triggers[1].Priority != 10 {
		t.Errorf("Expected second trigger priority 10, got %d", triggers[1].Priority)
	}
	if triggers[2].Priority != 5 {
		t.Errorf("Expected third trigger priority 5, got %d", triggers[2].Priority)
	}
}

func TestEngine_Evaluate_NoMatchingRules(t *testing.T) {
	registry := NewRegistry()

	rule := &testRule{
		id:          "test_rule",
		name:        "Test Rule",
		signalTypes: []string{"oauth_token_generated"},
		config:      RuleConfig{ID: "test_rule", Enabled: true, Priority: 10},
		shouldMatch: false, // Won't match
	}
	registry.Register(rule)

	engine := NewEngine(registry)

	playerCtx := &signal.PlayerContext{
		UserID: "test-user",
		State:  &state.ChurnState{},
	}
	sig := signalBuiltin.NewOauthTokenGeneratedSignal("test-user", time.Now(), playerCtx)

	triggers, err := engine.Evaluate(context.Background(), sig)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(triggers) != 0 {
		t.Errorf("Expected 0 triggers, got %d", len(triggers))
	}
}

func TestEngine_Evaluate_RuleError(t *testing.T) {
	registry := NewRegistry()

	rule := &testRule{
		id:          "error_rule",
		name:        "Error Rule",
		signalTypes: []string{"oauth_token_generated"},
		config:      RuleConfig{ID: "error_rule", Enabled: true, Priority: 10},
		shouldError: true,
	}
	registry.Register(rule)

	engine := NewEngine(registry)

	playerCtx := &signal.PlayerContext{
		UserID: "test-user",
		State:  &state.ChurnState{},
	}
	sig := signalBuiltin.NewOauthTokenGeneratedSignal("test-user", time.Now(), playerCtx)

	triggers, err := engine.Evaluate(context.Background(), sig)
	// Engine should not return error, but log it and continue
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Rule error should not produce triggers
	if len(triggers) != 0 {
		t.Errorf("Expected 0 triggers when rule errors, got %d", len(triggers))
	}
}

func TestEngine_Evaluate_MixedResults(t *testing.T) {
	registry := NewRegistry()

	matchingRule := &testRule{
		id:          "matching_rule",
		name:        "Matching Rule",
		signalTypes: []string{"oauth_token_generated"},
		config:      RuleConfig{ID: "matching_rule", Enabled: true, Priority: 10},
		shouldMatch: true,
	}
	nonMatchingRule := &testRule{
		id:          "non_matching_rule",
		name:        "Non Matching Rule",
		signalTypes: []string{"oauth_token_generated"},
		config:      RuleConfig{ID: "non_matching_rule", Enabled: true, Priority: 5},
		shouldMatch: false,
	}
	errorRule := &testRule{
		id:          "error_rule",
		name:        "Error Rule",
		signalTypes: []string{"oauth_token_generated"},
		config:      RuleConfig{ID: "error_rule", Enabled: true, Priority: 15},
		shouldError: true,
	}

	registry.Register(matchingRule)
	registry.Register(nonMatchingRule)
	registry.Register(errorRule)

	engine := NewEngine(registry)

	playerCtx := &signal.PlayerContext{
		UserID: "test-user",
		State:  &state.ChurnState{},
	}
	sig := signalBuiltin.NewOauthTokenGeneratedSignal("test-user", time.Now(), playerCtx)

	triggers, err := engine.Evaluate(context.Background(), sig)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Only matching rule should produce trigger
	if len(triggers) != 1 {
		t.Fatalf("Expected 1 trigger, got %d", len(triggers))
	}

	if triggers[0].RuleID != "matching_rule" {
		t.Errorf("Expected matching_rule to trigger, got '%s'", triggers[0].RuleID)
	}
}

func TestEngine_EvaluateMultiple(t *testing.T) {
	registry := NewRegistry()

	rule := &testRule{
		id:          "test_rule",
		name:        "Test Rule",
		signalTypes: []string{"oauth_token_generated", "match_win"},
		config:      RuleConfig{ID: "test_rule", Enabled: true, Priority: 10},
		shouldMatch: true,
	}
	registry.Register(rule)

	engine := NewEngine(registry)

	playerCtx := &signal.PlayerContext{
		UserID: "test-user",
		State:  &state.ChurnState{},
	}

	signals := []signal.Signal{
		signalBuiltin.NewOauthTokenGeneratedSignal("test-user", time.Now(), playerCtx),
		signalBuiltin.NewWinSignal("test-user", time.Now(), 5, playerCtx),
	}

	triggers, err := engine.EvaluateMultiple(context.Background(), signals)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(triggers) != 2 {
		t.Errorf("Expected 2 triggers, got %d", len(triggers))
	}
}

func TestEngine_EvaluateMultiple_Empty(t *testing.T) {
	registry := NewRegistry()
	engine := NewEngine(registry)

	triggers, err := engine.EvaluateMultiple(context.Background(), []signal.Signal{})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(triggers) != 0 {
		t.Errorf("Expected 0 triggers, got %d", len(triggers))
	}
}
