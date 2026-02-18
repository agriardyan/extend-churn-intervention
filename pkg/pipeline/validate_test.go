package pipeline

import (
	"context"
	"strings"
	"testing"

	"github.com/AccelByte/extend-churn-intervention/pkg/action"
	"github.com/AccelByte/extend-churn-intervention/pkg/rule"
	"github.com/AccelByte/extend-churn-intervention/pkg/signal"
)

// mockRule for testing
type mockRule struct {
	id          string
	signalTypes []string
	enabled     bool
}

func (m *mockRule) ID() string            { return m.id }
func (m *mockRule) Name() string          { return "Mock Rule" }
func (m *mockRule) SignalTypes() []string { return m.signalTypes }
func (m *mockRule) Evaluate(ctx context.Context, sig signal.Signal) (bool, *rule.Trigger, error) {
	return false, nil, nil
}
func (m *mockRule) Config() rule.RuleConfig {
	return rule.RuleConfig{
		ID:      m.id,
		Type:    "mock",
		Enabled: m.enabled,
	}
}

// mockAction for testing
type mockAction struct {
	id      string
	enabled bool
}

func (m *mockAction) ID() string   { return m.id }
func (m *mockAction) Name() string { return "Mock Action" }
func (m *mockAction) Execute(ctx context.Context, trigger *rule.Trigger, playerCtx *signal.PlayerContext) error {
	return nil
}
func (m *mockAction) Rollback(ctx context.Context, trigger *rule.Trigger, playerCtx *signal.PlayerContext) error {
	return nil
}
func (m *mockAction) Config() action.ActionConfig {
	return action.ActionConfig{
		ID:      m.id,
		Type:    "mock",
		Enabled: m.enabled,
	}
}

func TestValidateWiring_AllRegistered(t *testing.T) {
	// Setup registries
	ruleRegistry := rule.NewRegistry()
	actionRegistry := action.NewRegistry()

	// Register rules
	rule1 := &mockRule{id: "rage-quit-rule", enabled: true}
	rule2 := &mockRule{id: "session-decline-rule", enabled: true}
	ruleRegistry.Register(rule1)
	ruleRegistry.Register(rule2)

	// Register actions
	action1 := &mockAction{id: "dispatch-comeback-challenge", enabled: true}
	action2 := &mockAction{id: "grant-item", enabled: true}
	actionRegistry.Register(action1)
	actionRegistry.Register(action2)

	// Create config
	config := &Config{
		Rules: []RuleConfig{
			{ID: "rage-quit-rule", Type: "rage_quit", Enabled: true, Actions: []string{"dispatch-comeback-challenge"}},
			{ID: "session-decline-rule", Type: "session_decline", Enabled: true, Actions: []string{"grant-item"}},
		},
		Actions: []ActionConfig{
			{ID: "dispatch-comeback-challenge", Type: "dispatch_comeback_challenge", Enabled: true},
			{ID: "grant-item", Type: "grant_item", Enabled: true},
		},
	}

	// Validate - should pass
	err := ValidateWiring(ruleRegistry, actionRegistry, config)
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

func TestValidateWiring_MissingRule(t *testing.T) {
	ruleRegistry := rule.NewRegistry()
	actionRegistry := action.NewRegistry()

	// Register only one rule, but config has two
	rule1 := &mockRule{id: "rage-quit-rule", enabled: true}
	ruleRegistry.Register(rule1)

	// Register actions
	action1 := &mockAction{id: "dispatch-comeback-challenge", enabled: true}
	actionRegistry.Register(action1)

	config := &Config{
		Rules: []RuleConfig{
			{ID: "rage-quit-rule", Type: "rage_quit", Enabled: true, Actions: []string{"dispatch-comeback-challenge"}},
			{ID: "missing-rule", Type: "missing_type", Enabled: true, Actions: []string{"dispatch-comeback-challenge"}},
		},
		Actions: []ActionConfig{
			{ID: "dispatch-comeback-challenge", Type: "dispatch_comeback_challenge", Enabled: true},
		},
	}

	err := ValidateWiring(ruleRegistry, actionRegistry, config)
	if err == nil {
		t.Error("expected error for missing rule, got nil")
	}

	if !strings.Contains(err.Error(), "missing-rule") {
		t.Errorf("expected error to mention 'missing-rule', got: %v", err)
	}
	if !strings.Contains(err.Error(), "not registered") {
		t.Errorf("expected error to mention 'not registered', got: %v", err)
	}
}

func TestValidateWiring_MissingAction(t *testing.T) {
	ruleRegistry := rule.NewRegistry()
	actionRegistry := action.NewRegistry()

	// Register rule
	rule1 := &mockRule{id: "rage-quit-rule", enabled: true}
	ruleRegistry.Register(rule1)

	// Register only one action, but config has two
	action1 := &mockAction{id: "dispatch-comeback-challenge", enabled: true}
	actionRegistry.Register(action1)

	config := &Config{
		Rules: []RuleConfig{
			{ID: "rage-quit-rule", Type: "rage_quit", Enabled: true, Actions: []string{"dispatch-comeback-challenge"}},
		},
		Actions: []ActionConfig{
			{ID: "dispatch-comeback-challenge", Type: "dispatch_comeback_challenge", Enabled: true},
			{ID: "missing-action", Type: "missing_type", Enabled: true},
		},
	}

	err := ValidateWiring(ruleRegistry, actionRegistry, config)
	if err == nil {
		t.Error("expected error for missing action, got nil")
	}

	if !strings.Contains(err.Error(), "missing-action") {
		t.Errorf("expected error to mention 'missing-action', got: %v", err)
	}
	if !strings.Contains(err.Error(), "not registered") {
		t.Errorf("expected error to mention 'not registered', got: %v", err)
	}
}

func TestValidateWiring_DisabledRuleNotValidated(t *testing.T) {
	ruleRegistry := rule.NewRegistry()
	actionRegistry := action.NewRegistry()

	// Register only one rule
	rule1 := &mockRule{id: "rage-quit-rule", enabled: true}
	ruleRegistry.Register(rule1)

	action1 := &mockAction{id: "dispatch-comeback-challenge", enabled: true}
	actionRegistry.Register(action1)

	config := &Config{
		Rules: []RuleConfig{
			{ID: "rage-quit-rule", Type: "rage_quit", Enabled: true, Actions: []string{"dispatch-comeback-challenge"}},
			// This rule is disabled, so it should not be validated
			{ID: "disabled-rule", Type: "some_type", Enabled: false, Actions: []string{"dispatch-comeback-challenge"}},
		},
		Actions: []ActionConfig{
			{ID: "dispatch-comeback-challenge", Type: "dispatch_comeback_challenge", Enabled: true},
		},
	}

	// Should pass because disabled rule is not validated
	err := ValidateWiring(ruleRegistry, actionRegistry, config)
	if err != nil {
		t.Errorf("expected no error for disabled rule, got: %v", err)
	}
}

func TestValidateWiring_DisabledActionNotValidated(t *testing.T) {
	ruleRegistry := rule.NewRegistry()
	actionRegistry := action.NewRegistry()

	rule1 := &mockRule{id: "rage-quit-rule", enabled: true}
	ruleRegistry.Register(rule1)

	// Register only one action
	action1 := &mockAction{id: "dispatch-comeback-challenge", enabled: true}
	actionRegistry.Register(action1)

	config := &Config{
		Rules: []RuleConfig{
			{ID: "rage-quit-rule", Type: "rage_quit", Enabled: true, Actions: []string{"dispatch-comeback-challenge"}},
		},
		Actions: []ActionConfig{
			{ID: "dispatch-comeback-challenge", Type: "dispatch_comeback_challenge", Enabled: true},
			// This action is disabled, so it should not be validated
			{ID: "disabled-action", Type: "some_type", Enabled: false},
		},
	}

	// Should pass because disabled action is not validated
	err := ValidateWiring(ruleRegistry, actionRegistry, config)
	if err != nil {
		t.Errorf("expected no error for disabled action, got: %v", err)
	}
}

func TestValidateWiring_MultipleErrors(t *testing.T) {
	ruleRegistry := rule.NewRegistry()
	actionRegistry := action.NewRegistry()

	// Don't register anything

	config := &Config{
		Rules: []RuleConfig{
			{ID: "missing-rule-1", Type: "type1", Enabled: true, Actions: []string{"missing-action-1"}},
			{ID: "missing-rule-2", Type: "type2", Enabled: true, Actions: []string{"missing-action-2"}},
		},
		Actions: []ActionConfig{
			{ID: "missing-action-1", Type: "action_type1", Enabled: true},
			{ID: "missing-action-2", Type: "action_type2", Enabled: true},
		},
	}

	err := ValidateWiring(ruleRegistry, actionRegistry, config)
	if err == nil {
		t.Error("expected error for missing rules and actions, got nil")
	}

	errMsg := err.Error()
	// Should contain all missing items
	if !strings.Contains(errMsg, "missing-rule-1") {
		t.Errorf("expected error to mention 'missing-rule-1', got: %v", err)
	}
	if !strings.Contains(errMsg, "missing-rule-2") {
		t.Errorf("expected error to mention 'missing-rule-2', got: %v", err)
	}
	if !strings.Contains(errMsg, "missing-action-1") {
		t.Errorf("expected error to mention 'missing-action-1', got: %v", err)
	}
	if !strings.Contains(errMsg, "missing-action-2") {
		t.Errorf("expected error to mention 'missing-action-2', got: %v", err)
	}
}

func TestValidateWiring_EmptyConfig(t *testing.T) {
	ruleRegistry := rule.NewRegistry()
	actionRegistry := action.NewRegistry()

	config := &Config{
		Rules:   []RuleConfig{},
		Actions: []ActionConfig{},
	}

	// Empty config should validate successfully
	err := ValidateWiring(ruleRegistry, actionRegistry, config)
	if err != nil {
		t.Errorf("expected no error for empty config, got: %v", err)
	}
}
