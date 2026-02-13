package builtin

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/AccelByte/extends-anti-churn/pkg/rule"
	"github.com/AccelByte/extends-anti-churn/pkg/service/mock"
	"github.com/AccelByte/extends-anti-churn/pkg/signal"
)

// testSignal is a simple signal implementation for testing
type testClanSignal struct {
	signalType string
	userID     string
	timestamp  time.Time
	metadata   map[string]interface{}
	context    *signal.PlayerContext
}

func (s *testClanSignal) Type() string                     { return s.signalType }
func (s *testClanSignal) UserID() string                   { return s.userID }
func (s *testClanSignal) Timestamp() time.Time             { return s.timestamp }
func (s *testClanSignal) Metadata() map[string]interface{} { return s.metadata }
func (s *testClanSignal) Context() *signal.PlayerContext   { return s.context }

func TestClanActivityRule_NoPlayerContext(t *testing.T) {
	// Setup
	mockService := mock.NewClanService()
	deps := rule.NewRuleDependencies().WithClanService(mockService)

	config := rule.RuleConfig{
		ID:   "clan_activity_1",
		Type: ClanActivityRuleID,
		Parameters: map[string]interface{}{
			"min_active_members": 3,
		},
	}

	r := NewClanActivityRule(config, deps)

	// Signal without player context
	sig := &testClanSignal{
		signalType: "oauth_token_generated",
		userID:     "user123",
		timestamp:  time.Now(),
		metadata:   map[string]interface{}{},
		context:    nil, // No context
	}

	// Execute
	triggered, trigger, err := r.Evaluate(context.Background(), sig)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if triggered {
		t.Error("Expected rule not to trigger without player context")
	}
	if trigger != nil {
		t.Error("Expected nil trigger")
	}
}

func TestClanActivityRule_PlayerNotInClan(t *testing.T) {
	// Setup
	mockService := mock.NewClanService()
	deps := rule.NewRuleDependencies().WithClanService(mockService)

	config := rule.RuleConfig{
		ID:   "clan_activity_1",
		Type: ClanActivityRuleID,
		Parameters: map[string]interface{}{
			"min_active_members": 3,
		},
	}

	r := NewClanActivityRule(config, deps)

	// Signal with player context but no clan
	sig := &testClanSignal{
		signalType: "oauth_token_generated",
		userID:     "user123",
		timestamp:  time.Now(),
		metadata:   map[string]interface{}{},
		context: &signal.PlayerContext{
			UserID: "user123",
			SessionInfo: map[string]interface{}{
				// No clan_id
				"level": 10,
			},
		},
	}

	// Execute
	triggered, trigger, err := r.Evaluate(context.Background(), sig)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if triggered {
		t.Error("Expected rule not to trigger for player not in clan")
	}
	if trigger != nil {
		t.Error("Expected nil trigger")
	}
}

func TestClanActivityRule_ServiceUnavailable(t *testing.T) {
	// Setup - no service provided
	deps := rule.NewRuleDependencies() // No clan service

	config := rule.RuleConfig{
		ID:   "clan_activity_1",
		Type: ClanActivityRuleID,
		Parameters: map[string]interface{}{
			"min_active_members": 3,
		},
	}

	r := NewClanActivityRule(config, deps)

	// Signal with clan ID
	sig := &testClanSignal{
		signalType: "oauth_token_generated",
		userID:     "user123",
		timestamp:  time.Now(),
		metadata:   map[string]interface{}{},
		context: &signal.PlayerContext{
			UserID: "user123",
			SessionInfo: map[string]interface{}{
				"clan_id": "clan-abc",
			},
		},
	}

	// Execute
	triggered, trigger, err := r.Evaluate(context.Background(), sig)

	// Assert
	if err == nil {
		t.Error("Expected error when service is unavailable")
	}
	if triggered {
		t.Error("Expected rule not to trigger when service unavailable")
	}
	if trigger != nil {
		t.Error("Expected nil trigger")
	}
}

func TestClanActivityRule_ServiceError(t *testing.T) {
	// Setup
	mockService := mock.NewClanService().WithError(errors.New("external API timeout"))
	deps := rule.NewRuleDependencies().WithClanService(mockService)

	config := rule.RuleConfig{
		ID:   "clan_activity_1",
		Type: ClanActivityRuleID,
		Parameters: map[string]interface{}{
			"min_active_members": 3,
		},
	}

	r := NewClanActivityRule(config, deps)

	// Signal with clan ID
	sig := &testClanSignal{
		signalType: "oauth_token_generated",
		userID:     "user123",
		timestamp:  time.Now(),
		metadata:   map[string]interface{}{},
		context: &signal.PlayerContext{
			UserID: "user123",
			SessionInfo: map[string]interface{}{
				"clan_id": "clan-abc",
			},
		},
	}

	// Execute
	triggered, trigger, err := r.Evaluate(context.Background(), sig)

	// Assert
	if err == nil {
		t.Error("Expected error to be propagated from service")
	}
	if triggered {
		t.Error("Expected rule not to trigger when service errors")
	}
	if trigger != nil {
		t.Error("Expected nil trigger")
	}
}

func TestClanActivityRule_LowActivity_Triggered(t *testing.T) {
	// Setup
	mockService := mock.NewClanService().WithActivity(&rule.ClanActivity{
		ClanID:                 "clan-abc",
		ActiveMembersLast7Days: 2, // Below threshold of 3
		TotalMembers:           10,
		RecentEvents:           []rule.ClanEvent{},
	})
	deps := rule.NewRuleDependencies().WithClanService(mockService)

	config := rule.RuleConfig{
		ID:   "clan_activity_1",
		Type: ClanActivityRuleID,
		Parameters: map[string]interface{}{
			"min_active_members": 3,
		},
	}

	r := NewClanActivityRule(config, deps)

	// Signal with clan ID
	sig := &testClanSignal{
		signalType: "match_win",
		userID:     "user123",
		timestamp:  time.Now(),
		metadata:   map[string]interface{}{},
		context: &signal.PlayerContext{
			UserID: "user123",
			SessionInfo: map[string]interface{}{
				"clan_id": "clan-abc",
			},
		},
	}

	// Execute
	triggered, trigger, err := r.Evaluate(context.Background(), sig)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if !triggered {
		t.Error("Expected rule to trigger for low clan activity")
	}
	if trigger == nil {
		t.Fatal("Expected trigger to be returned")
	}
	if trigger.Metadata == nil {
		t.Fatal("Expected trigger metadata to be returned")
	}

	// Verify metadata contents
	metadata := trigger.Metadata
	if clanID, ok := metadata["clan_id"].(string); !ok || clanID != "clan-abc" {
		t.Errorf("Expected clan_id='clan-abc' in metadata, got: %v", metadata["clan_id"])
	}
	if activeMembers, ok := metadata["active_members"].(int); !ok || activeMembers != 2 {
		t.Errorf("Expected active_members=2 in metadata, got: %v", metadata["active_members"])
	}
	if totalMembers, ok := metadata["total_members"].(int); !ok || totalMembers != 10 {
		t.Errorf("Expected total_members=10 in metadata, got: %v", metadata["total_members"])
	}
	if threshold, ok := metadata["threshold"].(int); !ok || threshold != 3 {
		t.Errorf("Expected threshold=3 in metadata, got: %v", metadata["threshold"])
	}
	if activityRate, ok := metadata["activity_rate"].(float64); !ok || activityRate != 0.2 {
		t.Errorf("Expected activity_rate=0.2 in metadata, got: %v", metadata["activity_rate"])
	}
}

func TestClanActivityRule_HealthyActivity_NotTriggered(t *testing.T) {
	// Setup
	mockService := mock.NewClanService().WithActivity(&rule.ClanActivity{
		ClanID:                 "clan-xyz",
		ActiveMembersLast7Days: 5, // Above threshold of 3
		TotalMembers:           10,
		RecentEvents:           []rule.ClanEvent{},
	})
	deps := rule.NewRuleDependencies().WithClanService(mockService)

	config := rule.RuleConfig{
		ID:   "clan_activity_1",
		Type: ClanActivityRuleID,
		Parameters: map[string]interface{}{
			"min_active_members": 3,
		},
	}

	r := NewClanActivityRule(config, deps)

	// Signal with clan ID
	sig := &testClanSignal{
		signalType: "oauth_token_generated",
		userID:     "user456",
		timestamp:  time.Now(),
		metadata:   map[string]interface{}{},
		context: &signal.PlayerContext{
			UserID: "user456",
			SessionInfo: map[string]interface{}{
				"clan_id": "clan-xyz",
			},
		},
	}

	// Execute
	triggered, trigger, err := r.Evaluate(context.Background(), sig)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if triggered {
		t.Error("Expected rule not to trigger for healthy clan activity")
	}
	if trigger != nil {
		t.Error("Expected nil trigger when not triggered")
	}
}

func TestClanActivityRule_ExactThreshold_NotTriggered(t *testing.T) {
	// Setup
	mockService := mock.NewClanService().WithActivity(&rule.ClanActivity{
		ClanID:                 "clan-xyz",
		ActiveMembersLast7Days: 3, // Exactly at threshold
		TotalMembers:           10,
	})
	deps := rule.NewRuleDependencies().WithClanService(mockService)

	config := rule.RuleConfig{
		ID:   "clan_activity_1",
		Type: ClanActivityRuleID,
		Parameters: map[string]interface{}{
			"min_active_members": 3,
		},
	}

	r := NewClanActivityRule(config, deps)

	// Signal with clan ID
	sig := &testClanSignal{
		signalType: "oauth_token_generated",
		userID:     "user789",
		timestamp:  time.Now(),
		context: &signal.PlayerContext{
			UserID: "user789",
			SessionInfo: map[string]interface{}{
				"clan_id": "clan-xyz",
			},
		},
	}

	// Execute
	triggered, _, err := r.Evaluate(context.Background(), sig)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if triggered {
		t.Error("Expected rule not to trigger when exactly at threshold")
	}
}

func TestClanActivityRule_Metadata(t *testing.T) {
	config := rule.RuleConfig{
		ID:   "clan_activity_1",
		Type: ClanActivityRuleID,
		Parameters: map[string]interface{}{
			"min_active_members": 5,
		},
	}

	r := NewClanActivityRule(config, nil)

	if r.ID() != "clan_activity_1" {
		t.Errorf("Expected ID 'clan_activity_1', got '%s'", r.ID())
	}

	if r.Name() != "Clan Activity Detection" {
		t.Errorf("Expected name 'Clan Activity Detection', got '%s'", r.Name())
	}

	expectedDesc := "Detects players in clans with fewer than 5 active members in the last 7 days"
	if r.Description() != expectedDesc {
		t.Errorf("Expected description '%s', got '%s'", expectedDesc, r.Description())
	}

	signalTypes := r.SignalTypes()
	if len(signalTypes) != 2 {
		t.Errorf("Expected 2 signal types, got %d", len(signalTypes))
	}
}

func TestClanActivityRule_Validate(t *testing.T) {
	tests := []struct {
		name        string
		minMembers  int
		expectError bool
	}{
		{"Valid threshold", 3, false},
		{"Valid threshold 1", 1, false},
		{"Invalid threshold 0", 0, true},
		{"Invalid threshold negative", -1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := rule.RuleConfig{
				ID:   "test",
				Type: ClanActivityRuleID,
				Parameters: map[string]interface{}{
					"min_active_members": tt.minMembers,
				},
			}

			r := NewClanActivityRule(config, nil)
			err := r.Validate()

			if tt.expectError && err == nil {
				t.Error("Expected validation error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no validation error, got: %v", err)
			}
		})
	}
}
