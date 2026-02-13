package rule_test

import (
	"testing"

	"github.com/AccelByte/extends-anti-churn/pkg/rule"
	ruleExamples "github.com/AccelByte/extends-anti-churn/pkg/rule/examples"
)

func init() {
	// Register builtin rules for all tests
	// Pass nil dependencies - tests will mock as needed
	ruleExamples.RegisterRules(nil)
}

const (
	// Rule type constants for testing
	rageQuitRuleID       = "rage_quit"
	losingStreakRuleID   = "losing_streak"
	sessionDeclineRuleID = "session_decline"
)

func TestCreateRule_RageQuit(t *testing.T) {
	config := rule.RuleConfig{
		ID:       "rage_quit_1",
		Type:     rageQuitRuleID,
		Enabled:  true,
		Priority: 10,
		Parameters: map[string]interface{}{
			"threshold": 3,
		},
	}

	r, err := rule.CreateRule(config)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if r == nil {
		t.Fatal("Expected non-nil rule")
	}

	if r.ID() != config.ID {
		t.Errorf("Expected rule ID '%s', got '%s'", config.ID, r.ID())
	}

	if r.Name() != "Rage Quit Detection" {
		t.Errorf("Expected rule name 'Rage Quit Detection', got '%s'", r.Name())
	}
}

func TestCreateRule_LosingStreak(t *testing.T) {
	config := rule.RuleConfig{
		ID:       "losing_streak_1",
		Type:     losingStreakRuleID,
		Enabled:  true,
		Priority: 10,
		Parameters: map[string]interface{}{
			"threshold": 5,
		},
	}

	r, err := rule.CreateRule(config)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if r == nil {
		t.Fatal("Expected non-nil rule")
	}

	if r.ID() != config.ID {
		t.Errorf("Expected rule ID '%s', got '%s'", config.ID, r.ID())
	}

	if r.Name() != "Losing Streak Detection" {
		t.Errorf("Expected rule name 'Losing Streak Detection', got '%s'", r.Name())
	}
}

func TestCreateRule_SessionDecline(t *testing.T) {
	config := rule.RuleConfig{
		ID:       "session_decline_1",
		Type:     sessionDeclineRuleID,
		Enabled:  true,
		Priority: 10,
	}

	r, err := rule.CreateRule(config)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if r == nil {
		t.Fatal("Expected non-nil rule")
	}

	if r.ID() != config.ID {
		t.Errorf("Expected rule ID '%s', got '%s'", config.ID, r.ID())
	}

	if r.Name() != "Session Decline Detection" {
		t.Errorf("Expected rule name 'Session Decline Detection', got '%s'", r.Name())
	}
}

func TestCreateRule_Disabled(t *testing.T) {
	config := rule.RuleConfig{
		ID:       "disabled_rule",
		Type:     rageQuitRuleID,
		Enabled:  false,
		Priority: 10,
	}

	r, err := rule.CreateRule(config)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if r != nil {
		t.Error("Expected nil rule for disabled config")
	}
}

func TestCreateRule_UnknownType(t *testing.T) {
	config := rule.RuleConfig{
		ID:       "unknown_rule",
		Type:     "unknown_type",
		Enabled:  true,
		Priority: 10,
	}

	r, err := rule.CreateRule(config)
	if err == nil {
		t.Error("Expected error for unknown rule type")
	}

	if r != nil {
		t.Error("Expected nil rule for unknown type")
	}
}

func TestCreateRules_Multiple(t *testing.T) {
	configs := []rule.RuleConfig{
		{
			ID:       "rage_quit_1",
			Type:     rageQuitRuleID,
			Enabled:  true,
			Priority: 10,
		},
		{
			ID:       "losing_streak_1",
			Type:     losingStreakRuleID,
			Enabled:  true,
			Priority: 10,
		},
		{
			ID:       "session_decline_1",
			Type:     sessionDeclineRuleID,
			Enabled:  true,
			Priority: 10,
		},
	}

	rules, errors := rule.CreateRules(configs)

	if len(errors) != 0 {
		t.Errorf("Expected no errors, got %d", len(errors))
	}

	if len(rules) != 3 {
		t.Fatalf("Expected 3 rules, got %d", len(rules))
	}
}

func TestCreateRules_WithErrors(t *testing.T) {
	configs := []rule.RuleConfig{
		{
			ID:       "valid_rule",
			Type:     rageQuitRuleID,
			Enabled:  true,
			Priority: 10,
		},
		{
			ID:       "invalid_rule",
			Type:     "unknown_type",
			Enabled:  true,
			Priority: 10,
		},
		{
			ID:       "disabled_rule",
			Type:     losingStreakRuleID,
			Enabled:  false,
			Priority: 10,
		},
	}

	rules, errors := rule.CreateRules(configs)

	if len(errors) != 1 {
		t.Errorf("Expected 1 error, got %d", len(errors))
	}

	// Should have 1 valid rule (disabled rules return nil, not error)
	if len(rules) != 1 {
		t.Errorf("Expected 1 rule, got %d", len(rules))
	}
}

func TestRegisterRules(t *testing.T) {
	registry := rule.NewRegistry()

	configs := []rule.RuleConfig{
		{
			ID:       "rage_quit_1",
			Type:     rageQuitRuleID,
			Enabled:  true,
			Priority: 10,
		},
		{
			ID:       "losing_streak_1",
			Type:     losingStreakRuleID,
			Enabled:  true,
			Priority: 10,
		},
	}

	err := rule.RegisterRules(registry, configs)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify rules are registered
	allRules := registry.GetAll()
	if len(allRules) != 2 {
		t.Errorf("Expected 2 registered rules, got %d", len(allRules))
	}
}

func TestRegisterRules_DuplicateID(t *testing.T) {
	registry := rule.NewRegistry()

	configs := []rule.RuleConfig{
		{
			ID:       "same_id",
			Type:     rageQuitRuleID,
			Enabled:  true,
			Priority: 10,
		},
		{
			ID:       "same_id", // Duplicate ID
			Type:     losingStreakRuleID,
			Enabled:  true,
			Priority: 10,
		},
	}

	err := rule.RegisterRules(registry, configs)
	if err == nil {
		t.Error("Expected error for duplicate rule ID")
	}
}
