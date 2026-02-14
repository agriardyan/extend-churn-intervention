package signal

import (
	"context"
	"fmt"
	"testing"
	"time"

	oauth "github.com/AccelByte/extends-anti-churn/pkg/pb/accelbyte-asyncapi/iam/oauth/v1"
	statistic "github.com/AccelByte/extends-anti-churn/pkg/pb/accelbyte-asyncapi/social/statistic/v1"
	"github.com/AccelByte/extends-anti-churn/pkg/service"
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

// setupTestProcessor creates a processor with test event processors registered
func setupTestProcessor(stores ...service.StateStore) *Processor {
	var store service.StateStore
	if len(stores) > 0 {
		store = stores[0]
	} else {
		store = newMockStateStore()
	}
	processor := NewProcessor(store, "test-namespace")

	// Register test event processors with dependencies
	processor.GetEventProcessorRegistry().Register(&testOAuthEventProcessor{stateStore: store, namespace: "test-namespace"})
	processor.GetEventProcessorRegistry().Register(&testRageQuitEventProcessor{stateStore: store, namespace: "test-namespace"})
	processor.GetEventProcessorRegistry().Register(&testMatchWinEventProcessor{stateStore: store, namespace: "test-namespace"})
	processor.GetEventProcessorRegistry().Register(&testLosingStreakEventProcessor{stateStore: store, namespace: "test-namespace"})

	return processor
}

// testSignal is a simple signal implementation for testing
type testSignal struct {
	signalType string
	userID     string
	timestamp  time.Time
	metadata   map[string]interface{}
	context    *PlayerContext
}

func (s *testSignal) Type() string                     { return s.signalType }
func (s *testSignal) UserID() string                   { return s.userID }
func (s *testSignal) Timestamp() time.Time             { return s.timestamp }
func (s *testSignal) Metadata() map[string]interface{} { return s.metadata }
func (s *testSignal) Context() *PlayerContext          { return s.context }

// Test event processor implementations
type testOAuthEventProcessor struct {
	stateStore service.StateStore
	namespace  string
}

func (p *testOAuthEventProcessor) EventType() string {
	return "oauth_token_generated"
}

func (p *testOAuthEventProcessor) Process(ctx context.Context, event interface{}) (Signal, error) {
	oauthEvent, ok := event.(*oauth.OauthTokenGenerated)
	if !ok {
		return nil, fmt.Errorf("expected *oauth.OauthTokenGenerated, got %T", event)
	}

	userID := oauthEvent.GetUserId()
	if userID == "" {
		return nil, fmt.Errorf("user ID is empty")
	}

	churnState, err := p.stateStore.GetChurnState(ctx, userID)
	if err != nil {
		return nil, err
	}

	playerCtx := BuildPlayerContext(userID, p.namespace, churnState)

	metadata := map[string]interface{}{
		"event": "oauth_token_generated",
	}
	return &testSignal{
		signalType: "login",
		userID:     userID,
		timestamp:  time.Now(),
		metadata:   metadata,
		context:    playerCtx,
	}, nil
}

// testRageQuitEventProcessor processes "rse-rage-quit" stat events for testing.
type testRageQuitEventProcessor struct {
	stateStore service.StateStore
	namespace  string
}

func (p *testRageQuitEventProcessor) EventType() string {
	return "rse-rage-quit"
}

func (p *testRageQuitEventProcessor) Process(ctx context.Context, event interface{}) (Signal, error) {
	statEvent, ok := event.(*statistic.StatItemUpdated)
	if !ok {
		return nil, fmt.Errorf("expected *statistic.StatItemUpdated, got %T", event)
	}

	userID := statEvent.GetUserId()
	if userID == "" {
		return nil, fmt.Errorf("user ID is empty")
	}

	churnState, err := p.stateStore.GetChurnState(ctx, userID)
	if err != nil {
		return nil, err
	}

	playerCtx := BuildPlayerContext(userID, p.namespace, churnState)

	value := statEvent.GetPayload().GetLatestValue()
	metadata := map[string]interface{}{
		"quit_count": int(value),
		"stat_code":  "rse-rage-quit",
	}
	return &testSignal{
		signalType: "rage_quit",
		userID:     userID,
		timestamp:  time.Now(),
		metadata:   metadata,
		context:    playerCtx,
	}, nil
}

// testMatchWinEventProcessor processes "rse-match-wins" stat events for testing.
type testMatchWinEventProcessor struct {
	stateStore service.StateStore
	namespace  string
}

func (p *testMatchWinEventProcessor) EventType() string {
	return "rse-match-wins"
}

func (p *testMatchWinEventProcessor) Process(ctx context.Context, event interface{}) (Signal, error) {
	statEvent, ok := event.(*statistic.StatItemUpdated)
	if !ok {
		return nil, fmt.Errorf("expected *statistic.StatItemUpdated, got %T", event)
	}

	userID := statEvent.GetUserId()
	if userID == "" {
		return nil, fmt.Errorf("user ID is empty")
	}

	churnState, err := p.stateStore.GetChurnState(ctx, userID)
	if err != nil {
		return nil, err
	}

	playerCtx := BuildPlayerContext(userID, p.namespace, churnState)

	value := statEvent.GetPayload().GetLatestValue()
	metadata := map[string]interface{}{
		"total_wins": int(value),
		"stat_code":  "rse-match-wins",
	}
	return &testSignal{
		signalType: "match_win",
		userID:     userID,
		timestamp:  time.Now(),
		metadata:   metadata,
		context:    playerCtx,
	}, nil
}

// testLosingStreakEventProcessor processes "rse-current-losing-streak" stat events for testing.
type testLosingStreakEventProcessor struct {
	stateStore service.StateStore
	namespace  string
}

func (p *testLosingStreakEventProcessor) EventType() string {
	return "rse-current-losing-streak"
}

func (p *testLosingStreakEventProcessor) Process(ctx context.Context, event interface{}) (Signal, error) {
	statEvent, ok := event.(*statistic.StatItemUpdated)
	if !ok {
		return nil, fmt.Errorf("expected *statistic.StatItemUpdated, got %T", event)
	}

	userID := statEvent.GetUserId()
	if userID == "" {
		return nil, fmt.Errorf("user ID is empty")
	}

	churnState, err := p.stateStore.GetChurnState(ctx, userID)
	if err != nil {
		return nil, err
	}

	playerCtx := BuildPlayerContext(userID, p.namespace, churnState)

	value := statEvent.GetPayload().GetLatestValue()
	metadata := map[string]interface{}{
		"current_streak": int(value),
		"stat_code":      "rse-current-losing-streak",
	}
	return &testSignal{
		signalType: "losing_streak",
		userID:     userID,
		timestamp:  time.Now(),
		metadata:   metadata,
		context:    playerCtx,
	}, nil
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
	processor := setupTestProcessor(store)

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
	processor := setupTestProcessor(store)

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
	processor := setupTestProcessor(store)

	event := &oauth.OauthTokenGenerated{}
	event.UserId = "test-user"

	_, err := processor.ProcessOAuthEvent(context.Background(), event)
	if err == nil {
		t.Error("Expected error when state store fails")
	}
}

func TestProcessor_ProcessStatEvent_StateStoreError(t *testing.T) {
	store := &errorStateStore{}
	processor := setupTestProcessor(store)

	event := &statistic.StatItemUpdated{
		UserId: "test-user",
		Payload: &statistic.StatItem{
			StatCode:    "rse-rage-quit",
			LatestValue: 3.0,
		},
	}

	_, err := processor.ProcessStatEvent(context.Background(), event)
	if err == nil {
		t.Error("Expected error when state store fails")
	}
}
