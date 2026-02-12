package rule

import (
	"context"
	"testing"

	"github.com/AccelByte/extends-anti-churn/pkg/signal"
)

// mockRule is a simple rule implementation for testing
type mockRule struct {
	id          string
	name        string
	signalTypes []string
	config      RuleConfig
}

func (m *mockRule) ID() string            { return m.id }
func (m *mockRule) Name() string          { return m.name }
func (m *mockRule) SignalTypes() []string { return m.signalTypes }
func (m *mockRule) Config() RuleConfig    { return m.config }
func (m *mockRule) Evaluate(ctx context.Context, sig signal.Signal) (bool, *Trigger, error) {
	return false, nil, nil
}

func TestNewRegistry(t *testing.T) {
	registry := NewRegistry()

	if registry == nil {
		t.Fatal("Expected non-nil registry")
	}

	if registry.Count() != 0 {
		t.Errorf("Expected empty registry, got count %d", registry.Count())
	}
}

func TestRegistry_Register(t *testing.T) {
	registry := NewRegistry()
	rule := &mockRule{
		id:     "test_rule",
		name:   "Test Rule",
		config: RuleConfig{ID: "test_rule", Enabled: true},
	}

	err := registry.Register(rule)
	if err != nil {
		t.Fatalf("Failed to register rule: %v", err)
	}

	if registry.Count() != 1 {
		t.Errorf("Expected count 1, got %d", registry.Count())
	}

	// Try to register same rule again
	err = registry.Register(rule)
	if err == nil {
		t.Error("Expected error when registering duplicate rule")
	}
}

func TestRegistry_Get(t *testing.T) {
	registry := NewRegistry()
	rule := &mockRule{
		id:     "test_rule",
		name:   "Test Rule",
		config: RuleConfig{ID: "test_rule", Enabled: true},
	}

	registry.Register(rule)

	retrieved := registry.Get("test_rule")
	if retrieved == nil {
		t.Fatal("Expected to retrieve rule")
	}

	if retrieved.ID() != "test_rule" {
		t.Errorf("Expected rule ID 'test_rule', got '%s'", retrieved.ID())
	}

	// Try to get non-existent rule
	notFound := registry.Get("non_existent")
	if notFound != nil {
		t.Error("Expected nil for non-existent rule")
	}
}

func TestRegistry_Unregister(t *testing.T) {
	registry := NewRegistry()
	rule := &mockRule{
		id:     "test_rule",
		name:   "Test Rule",
		config: RuleConfig{ID: "test_rule", Enabled: true},
	}

	registry.Register(rule)

	err := registry.Unregister("test_rule")
	if err != nil {
		t.Fatalf("Failed to unregister rule: %v", err)
	}

	if registry.Count() != 0 {
		t.Errorf("Expected count 0 after unregister, got %d", registry.Count())
	}

	// Try to unregister non-existent rule
	err = registry.Unregister("non_existent")
	if err == nil {
		t.Error("Expected error when unregistering non-existent rule")
	}
}

func TestRegistry_GetBySignalType(t *testing.T) {
	registry := NewRegistry()

	// Register rules with different signal types
	rule1 := &mockRule{
		id:          "login_rule",
		signalTypes: []string{"login"},
		config:      RuleConfig{ID: "login_rule", Enabled: true},
	}
	rule2 := &mockRule{
		id:          "rage_quit_rule",
		signalTypes: []string{"rage_quit"},
		config:      RuleConfig{ID: "rage_quit_rule", Enabled: true},
	}
	rule3 := &mockRule{
		id:          "multi_rule",
		signalTypes: []string{"login", "logout"},
		config:      RuleConfig{ID: "multi_rule", Enabled: true},
	}
	rule4 := &mockRule{
		id:          "all_rule",
		signalTypes: []string{}, // Empty means all types
		config:      RuleConfig{ID: "all_rule", Enabled: true},
	}
	rule5 := &mockRule{
		id:          "disabled_rule",
		signalTypes: []string{"login"},
		config:      RuleConfig{ID: "disabled_rule", Enabled: false},
	}

	registry.Register(rule1)
	registry.Register(rule2)
	registry.Register(rule3)
	registry.Register(rule4)
	registry.Register(rule5)

	// Test getting rules for "login"
	loginRules := registry.GetBySignalType("login")
	if len(loginRules) != 3 {
		t.Errorf("Expected 3 rules for 'login', got %d", len(loginRules))
	}

	// Test getting rules for "rage_quit"
	rageQuitRules := registry.GetBySignalType("rage_quit")
	if len(rageQuitRules) != 2 {
		t.Errorf("Expected 2 rules for 'rage_quit', got %d", len(rageQuitRules))
	}

	// Test getting rules for non-existent signal type
	noneRules := registry.GetBySignalType("non_existent")
	if len(noneRules) != 1 { // Only the all_rule should match
		t.Errorf("Expected 1 rule for 'non_existent', got %d", len(noneRules))
	}
}

func TestRegistry_GetAll(t *testing.T) {
	registry := NewRegistry()

	rule1 := &mockRule{id: "rule1", config: RuleConfig{ID: "rule1", Enabled: true}}
	rule2 := &mockRule{id: "rule2", config: RuleConfig{ID: "rule2", Enabled: true}}
	rule3 := &mockRule{id: "rule3", config: RuleConfig{ID: "rule3", Enabled: false}}

	registry.Register(rule1)
	registry.Register(rule2)
	registry.Register(rule3)

	allRules := registry.GetAll()
	if len(allRules) != 3 {
		t.Errorf("Expected 3 rules, got %d", len(allRules))
	}
}

func TestRuleConfig_GetConditionHelpers(t *testing.T) {
	config := RuleConfig{
		Conditions: map[string]interface{}{
			"int_value":    42,
			"float_value":  3.14,
			"string_value": "test",
			"bool_value":   true,
		},
	}

	// Test GetConditionInt
	if val := config.GetConditionInt("int_value", 0); val != 42 {
		t.Errorf("Expected int 42, got %d", val)
	}
	if val := config.GetConditionInt("missing", 99); val != 99 {
		t.Errorf("Expected default 99, got %d", val)
	}

	// Test GetConditionFloat
	if val := config.GetConditionFloat("float_value", 0.0); val != 3.14 {
		t.Errorf("Expected float 3.14, got %f", val)
	}
	if val := config.GetConditionFloat("missing", 9.99); val != 9.99 {
		t.Errorf("Expected default 9.99, got %f", val)
	}

	// Test GetConditionString
	if val := config.GetConditionString("string_value", ""); val != "test" {
		t.Errorf("Expected string 'test', got '%s'", val)
	}
	if val := config.GetConditionString("missing", "default"); val != "default" {
		t.Errorf("Expected default 'default', got '%s'", val)
	}

	// Test GetConditionBool
	if val := config.GetConditionBool("bool_value", false); val != true {
		t.Errorf("Expected bool true, got %v", val)
	}
	if val := config.GetConditionBool("missing", false); val != false {
		t.Errorf("Expected default false, got %v", val)
	}
}
