package signal

import (
	"context"
	"fmt"
	"testing"
	"time"

	oauth "github.com/AccelByte/extends-anti-churn/pkg/pb/accelbyte-asyncapi/iam/oauth/v1"
	statistic "github.com/AccelByte/extends-anti-churn/pkg/pb/accelbyte-asyncapi/social/statistic/v1"
	"github.com/AccelByte/extends-anti-churn/pkg/state"
)

// mockStateStore is a simple in-memory state store for testing
type mockStateStore struct {
	states map[string]*state.ChurnState
}

func newMockStateStore() *mockStateStore {
	return &mockStateStore{
		states: make(map[string]*state.ChurnState),
	}
}

func (m *mockStateStore) GetChurnState(ctx context.Context, userID string) (*state.ChurnState, error) {
	if s, ok := m.states[userID]; ok {
		return s, nil
	}
	// Return new state if not found
	return &state.ChurnState{
		Sessions: state.SessionState{
			ThisWeek:  0,
			LastWeek:  0,
			LastReset: time.Now(),
		},
		Challenge: state.ChallengeState{
			Active: false,
		},
		Intervention: state.InterventionState{},
	}, nil
}

func (m *mockStateStore) UpdateChurnState(ctx context.Context, userID string, churnState *state.ChurnState) error {
	m.states[userID] = churnState
	return nil
}

// setupTestProcessor creates a processor with builtin mappers registered
func setupTestProcessor() *Processor {
	store := newMockStateStore()
	processor := NewProcessor(store, "test-namespace")

	// Register test mappers inline
	processor.GetMapperRegistry().Register(&testRageQuitMapper{})
	processor.GetMapperRegistry().Register(&testMatchWinMapper{})
	processor.GetMapperRegistry().Register(&testLosingStreakMapper{})

	return processor
}

// Test mapper implementations - use generic signals to avoid circular dependencies
type testRageQuitMapper struct{}

func (m *testRageQuitMapper) StatCode() string {
	return "rse-rage-quit"
}

func (m *testRageQuitMapper) MapToSignal(userID string, timestamp time.Time, value float64, context *PlayerContext) Signal {
	metadata := map[string]interface{}{
		"quit_count": int(value),
		"stat_code":  "rse-rage-quit",
	}
	return &BaseSignal{
		signalType: "rage_quit",
		userID:     userID,
		timestamp:  timestamp,
		metadata:   metadata,
		context:    context,
	}
}

type testMatchWinMapper struct{}

func (m *testMatchWinMapper) StatCode() string {
	return "rse-match-wins"
}

func (m *testMatchWinMapper) MapToSignal(userID string, timestamp time.Time, value float64, context *PlayerContext) Signal {
	metadata := map[string]interface{}{
		"total_wins": int(value),
		"stat_code":  "rse-match-wins",
	}
	return &BaseSignal{
		signalType: "match_win",
		userID:     userID,
		timestamp:  timestamp,
		metadata:   metadata,
		context:    context,
	}
}

type testLosingStreakMapper struct{}

func (m *testLosingStreakMapper) StatCode() string {
	return "rse-current-losing-streak"
}

func (m *testLosingStreakMapper) MapToSignal(userID string, timestamp time.Time, value float64, context *PlayerContext) Signal {
	metadata := map[string]interface{}{
		"current_streak": int(value),
		"stat_code":      "rse-current-losing-streak",
	}
	return &BaseSignal{
		signalType: "losing_streak",
		userID:     userID,
		timestamp:  timestamp,
		metadata:   metadata,
		context:    context,
	}
}

func TestNewProcessor(t *testing.T) {
	store := newMockStateStore()
	processor := NewProcessor(store, "test-namespace")

	if processor == nil {
		t.Fatal("Expected non-nil processor")
	}

	if processor.namespace != "test-namespace" {
		t.Errorf("Expected namespace 'test-namespace', got '%s'", processor.namespace)
	}

	if processor.GetStateStore() != store {
		t.Error("Expected processor to use provided state store")
	}
}

func TestProcessor_ProcessOAuthEvent(t *testing.T) {
	store := newMockStateStore()
	processor := NewProcessor(store, "test-namespace")

	// Create event using the proper structure
	event := &oauth.OauthTokenGenerated{}
	// Set the UserId field directly on the event (it's embedded)
	event.UserId = "test-user-123"

	signal, err := processor.ProcessOAuthEvent(context.Background(), event)
	if err != nil {
		t.Fatalf("Failed to process OAuth event: %v", err)
	}

	if signal == nil {
		t.Fatal("Expected non-nil signal")
	}

	if signal.Type() != "login" {
		t.Errorf("Expected signal type 'login', got '%s'", signal.Type())
	}

	if signal.UserID() != "test-user-123" {
		t.Errorf("Expected user ID 'test-user-123', got '%s'", signal.UserID())
	}

	if signal.Context() == nil {
		t.Fatal("Expected non-nil player context")
	}

	if signal.Context().Namespace != "test-namespace" {
		t.Errorf("Expected namespace 'test-namespace', got '%s'", signal.Context().Namespace)
	}
}

func TestProcessor_ProcessOAuthEvent_NilEvent(t *testing.T) {
	store := newMockStateStore()
	processor := NewProcessor(store, "test-namespace")

	_, err := processor.ProcessOAuthEvent(context.Background(), nil)
	if err == nil {
		t.Error("Expected error for nil event")
	}
}

func TestProcessor_ProcessOAuthEvent_EmptyUserID(t *testing.T) {
	store := newMockStateStore()
	processor := NewProcessor(store, "test-namespace")

	event := &oauth.OauthTokenGenerated{}
	event.UserId = ""

	_, err := processor.ProcessOAuthEvent(context.Background(), event)
	if err == nil {
		t.Error("Expected error for empty user ID")
	}
}

func TestProcessor_ProcessStatEvent_RageQuit(t *testing.T) {
	processor := setupTestProcessor()

	event := &statistic.StatItemUpdated{
		UserId: "test-user-123",
		Payload: &statistic.StatItem{
			StatCode:    "rse-rage-quit",
			LatestValue: 5.0,
		},
	}

	signal, err := processor.ProcessStatEvent(context.Background(), event)
	if err != nil {
		t.Fatalf("Failed to process stat event: %v", err)
	}

	if signal.Type() != "rage_quit" {
		t.Errorf("Expected signal type 'rage_quit', got '%s'", signal.Type())
	}

	if signal.Metadata()["quit_count"] != 5 {
		t.Errorf("Expected quit_count=5 in metadata")
	}
}

func TestProcessor_ProcessStatEvent_MatchWins(t *testing.T) {
	processor := setupTestProcessor()

	event := &statistic.StatItemUpdated{
		UserId: "test-user-456",
		Payload: &statistic.StatItem{
			StatCode:    "rse-match-wins",
			LatestValue: 42.0,
		},
	}

	signal, err := processor.ProcessStatEvent(context.Background(), event)
	if err != nil {
		t.Fatalf("Failed to process stat event: %v", err)
	}

	if signal.Type() != "match_win" {
		t.Errorf("Expected signal type 'match_win', got '%s'", signal.Type())
	}

	if signal.Metadata()["total_wins"] != 42 {
		t.Errorf("Expected total_wins=42 in metadata")
	}
}

func TestProcessor_ProcessStatEvent_LosingStreak(t *testing.T) {
	processor := setupTestProcessor()

	event := &statistic.StatItemUpdated{
		UserId: "test-user-789",
		Payload: &statistic.StatItem{
			StatCode:    "rse-current-losing-streak",
			LatestValue: 7.0,
		},
	}

	signal, err := processor.ProcessStatEvent(context.Background(), event)
	if err != nil {
		t.Fatalf("Failed to process stat event: %v", err)
	}

	if signal.Type() != "losing_streak" {
		t.Errorf("Expected signal type 'losing_streak', got '%s'", signal.Type())
	}

	if signal.Metadata()["current_streak"] != 7 {
		t.Errorf("Expected current_streak=7 in metadata")
	}
}

func TestProcessor_ProcessStatEvent_UnknownStatCode(t *testing.T) {
	store := newMockStateStore()
	processor := NewProcessor(store, "test-namespace")

	event := &statistic.StatItemUpdated{
		UserId: "test-user-999",
		Payload: &statistic.StatItem{
			StatCode:    "unknown-stat-code",
			LatestValue: 99.5,
		},
	}

	signal, err := processor.ProcessStatEvent(context.Background(), event)
	if err != nil {
		t.Fatalf("Failed to process stat event: %v", err)
	}

	if signal.Type() != TypeStatUpdate {
		t.Errorf("Expected signal type '%s', got '%s'", TypeStatUpdate, signal.Type())
	}

	statSignal, ok := signal.(*StatUpdateSignal)
	if !ok {
		t.Fatal("Expected StatUpdateSignal")
	}

	if statSignal.StatCode != "unknown-stat-code" {
		t.Errorf("Expected stat code 'unknown-stat-code', got '%s'", statSignal.StatCode)
	}

	if statSignal.Value != 99.5 {
		t.Errorf("Expected value 99.5, got %f", statSignal.Value)
	}
}

func TestProcessor_ProcessStatEvent_NilEvent(t *testing.T) {
	store := newMockStateStore()
	processor := NewProcessor(store, "test-namespace")

	_, err := processor.ProcessStatEvent(context.Background(), nil)
	if err == nil {
		t.Error("Expected error for nil event")
	}
}

func TestProcessor_ProcessStatEvent_EmptyUserID(t *testing.T) {
	store := newMockStateStore()
	processor := NewProcessor(store, "test-namespace")

	event := &statistic.StatItemUpdated{
		Payload: &statistic.StatItem{
			UserId:      "",
			StatCode:    "rse-rage-quit",
			LatestValue: 3.0,
		},
	}

	_, err := processor.ProcessStatEvent(context.Background(), event)
	if err == nil {
		t.Error("Expected error for empty user ID")
	}
}

func TestProcessor_ProcessStatEvent_EmptyStatCode(t *testing.T) {
	store := newMockStateStore()
	processor := NewProcessor(store, "test-namespace")

	event := &statistic.StatItemUpdated{
		Payload: &statistic.StatItem{
			UserId:      "test-user-123",
			StatCode:    "",
			LatestValue: 3.0,
		},
	}

	_, err := processor.ProcessStatEvent(context.Background(), event)
	if err == nil {
		t.Error("Expected error for empty stat code")
	}
}

func TestProcessor_LoadPlayerContext(t *testing.T) {
	store := newMockStateStore()

	// Set up existing state
	existingState := &state.ChurnState{
		Sessions: state.SessionState{
			ThisWeek:  5,
			LastWeek:  10,
			LastReset: time.Now().Add(-24 * time.Hour),
		},
		Challenge: state.ChallengeState{
			Active:      true,
			WinsNeeded:  3,
			WinsCurrent: 1,
		},
		Intervention: state.InterventionState{
			CooldownUntil: time.Now().Add(24 * time.Hour),
		},
	}
	store.states["test-user"] = existingState

	processor := NewProcessor(store, "test-namespace")

	ctx, err := processor.loadPlayerContext(context.Background(), "test-user")
	if err != nil {
		t.Fatalf("Failed to load player context: %v", err)
	}

	if ctx.UserID != "test-user" {
		t.Errorf("Expected user ID 'test-user', got '%s'", ctx.UserID)
	}

	if ctx.Namespace != "test-namespace" {
		t.Errorf("Expected namespace 'test-namespace', got '%s'", ctx.Namespace)
	}

	if ctx.State.Sessions.ThisWeek != 5 {
		t.Errorf("Expected sessions this week 5, got %d", ctx.State.Sessions.ThisWeek)
	}

	if ctx.State.Sessions.LastWeek != 10 {
		t.Errorf("Expected sessions last week 10, got %d", ctx.State.Sessions.LastWeek)
	}

	// Check session info metadata
	if ctx.SessionInfo["sessions_this_week"] != 5 {
		t.Errorf("Expected session info this week 5, got %v", ctx.SessionInfo["sessions_this_week"])
	}

	if ctx.SessionInfo["challenge_active"] != true {
		t.Errorf("Expected challenge_active true, got %v", ctx.SessionInfo["challenge_active"])
	}

	if ctx.SessionInfo["on_cooldown"] != true {
		t.Errorf("Expected on_cooldown true, got %v", ctx.SessionInfo["on_cooldown"])
	}
}

// Test error handling when state store fails
type errorStateStore struct{}

func (e *errorStateStore) GetChurnState(ctx context.Context, userID string) (*state.ChurnState, error) {
	return nil, fmt.Errorf("mock state store error")
}

func (e *errorStateStore) UpdateChurnState(ctx context.Context, userID string, churnState *state.ChurnState) error {
	return fmt.Errorf("mock state store error")
}

func TestProcessor_ProcessOAuthEvent_StateStoreError(t *testing.T) {
	store := &errorStateStore{}
	processor := NewProcessor(store, "test-namespace")

	event := &oauth.OauthTokenGenerated{}
	event.UserId = "test-user"

	_, err := processor.ProcessOAuthEvent(context.Background(), event)
	if err == nil {
		t.Error("Expected error when state store fails")
	}
}

func TestProcessor_ProcessStatEvent_StateStoreError(t *testing.T) {
	store := &errorStateStore{}
	processor := NewProcessor(store, "test-namespace")

	event := &statistic.StatItemUpdated{
		Payload: &statistic.StatItem{
			UserId:      "test-user",
			StatCode:    "rse-rage-quit",
			LatestValue: 3.0,
		},
	}

	_, err := processor.ProcessStatEvent(context.Background(), event)
	if err == nil {
		t.Error("Expected error when state store fails")
	}
}
