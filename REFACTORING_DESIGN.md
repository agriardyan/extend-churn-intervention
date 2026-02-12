# Refactoring Design: Rule Engine & Action Framework

**Date:** February 11, 2026  
**Version:** 1.0  
**Status:** Proposal

---

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [Current Architecture Analysis](#current-architecture-analysis)
3. [Proposed Architecture](#proposed-architecture)
4. [Core Components](#core-components)
5. [Interface Definitions](#interface-definitions)
6. [Implementation Examples](#implementation-examples)
7. [Configuration System](#configuration-system)
8. [Migration Strategy](#migration-strategy)
9. [Benefits & Trade-offs](#benefits--trade-offs)
10. [Open Source Considerations](#open-source-considerations)

---

## Executive Summary

### Goals

- **Separation of Concerns**: Decouple signal detection (rules) from response execution (actions)
- **Extensibility**: Enable users to easily add custom rules and actions without modifying core code
- **Configuration-Driven**: Support YAML/JSON configuration for rule and action pipelines
- **Open Source Ready**: Clean API boundaries, comprehensive documentation, plugin examples

### Key Changes

```
BEFORE:                          AFTER:
┌─────────────────┐             ┌──────────────────────────────────┐
│ Event Handler   │             │       Event Handler              │
│ - Detection     │             └──────────────────────────────────┘
│ - Action        │                          │
│ - State         │                          ▼
│ (All mixed)     │             ┌──────────────────────────────────┐
└─────────────────┘             │      Signal Processor            │
                                │  (Normalizes events to signals)  │
                                └──────────────────────────────────┘
                                             │
                                             ▼
                                ┌──────────────────────────────────┐
                                │        Rule Engine               │
                                │  ┌──────────────────────────┐   │
                                │  │ Rule 1: Rage Quit        │   │
                                │  │ Rule 2: Losing Streak    │   │
                                │  │ Rule 3: Session Decline  │   │
                                │  │ Rule N: Custom...        │   │
                                │  └──────────────────────────┘   │
                                └──────────────────────────────────┘
                                             │
                                             ▼ (Triggers)
                                ┌──────────────────────────────────┐
                                │      Action Executor             │
                                │  ┌──────────────────────────┐   │
                                │  │ Action 1: Create Challenge│  │
                                │  │ Action 2: Grant Item     │   │
                                │  │ Action 3: Send Email     │   │
                                │  │ Action N: Custom...      │   │
                                │  └──────────────────────────┘   │
                                └──────────────────────────────────┘
```

---

## Current Architecture Analysis

### What Works Well

✅ **Clear Event Flow**: gRPC handlers receive well-defined protobuf events  
✅ **State Management**: Redis-based player state with TTL  
✅ **Separation of Handlers**: OAuth and Statistic handlers are separate  
✅ **Testability**: Good unit test coverage with mocked dependencies

### Pain Points

❌ **Tight Coupling**: Detection logic mixed with action execution in handlers  
❌ **Hard-Coded Rules**: Thresholds (3 rage quits, 5 losses) are constants  
❌ **Limited Extensibility**: Adding new rules requires editing handler code  
❌ **Action Rigidity**: Only one action type (create challenge + grant item)  
❌ **Configuration**: No way to adjust rules without recompilation

### Example of Current Coupling

```go
// pkg/handler/statistic.go (current)
func (h *StatisticHandler) handleRageQuit(ctx context.Context, userID string, value float64) error {
    state := h.store.GetState(userID)
    
    // RULE LOGIC (mixed with handler)
    if value >= 3 {
        // ACTION LOGIC (mixed with rule)
        state.CreateChallenge(...)
        h.grantEntitlement(...)
    }
    
    h.store.SaveState(userID, state)
}
```

**Problem**: Cannot change rule logic, action behavior, or threshold without editing code.

---

## Proposed Architecture

### High-Level Design

```
┌─────────────────────────────────────────────────────────────────┐
│                        Event Layer                               │
│  (gRPC Handlers - OAuth, Statistic, Platform Events)            │
└──────────────────────┬──────────────────────────────────────────┘
                       │ Raw Events
                       ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Signal Processor                              │
│  • Normalizes events into domain signals                        │
│  • Enriches with player context (state, history)                │
│  • Example: "RageQuitSignal", "WinSignal", "LoginSignal"       │
└──────────────────────┬──────────────────────────────────────────┘
                       │ Signals
                       ▼
┌─────────────────────────────────────────────────────────────────┐
│                      Rule Engine                                 │
│  ┌────────────────────────────────────────────────────────┐    │
│  │  Rule Registry (Plugin Architecture)                   │    │
│  │  • RageQuitRule: threshold=3, cooldown=48h            │    │
│  │  • LosingStreakRule: threshold=5, window=7d           │    │
│  │  • SessionDeclineRule: minSessions=1, period=7d       │    │
│  │  • CustomRule: user-defined...                        │    │
│  └────────────────────────────────────────────────────────┘    │
│                                                                  │
│  Rule Evaluation:                                               │
│  1. Match signal type to registered rules                       │
│  2. Check rule conditions (thresholds, cooldowns, state)       │
│  3. Emit triggers for matched rules                            │
└──────────────────────┬──────────────────────────────────────────┘
                       │ Triggers
                       ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Action Executor                               │
│  ┌────────────────────────────────────────────────────────┐    │
│  │  Action Registry (Plugin Architecture)                 │    │
│  │  • CreateChallengeAction: wins=3, duration=7d         │    │
│  │  • GrantItemAction: itemID, quantity                  │    │
│  │  • SendNotificationAction: template, channels         │    │
│  │  • WebhookAction: url, payload                        │    │
│  │  • CustomAction: user-defined...                      │    │
│  └────────────────────────────────────────────────────────┘    │
│                                                                  │
│  Action Execution:                                              │
│  1. Match trigger to configured action pipeline                 │
│  2. Execute actions sequentially or in parallel                 │
│  3. Handle rollback on failures (optional)                     │
│  4. Update state and emit metrics                              │
└─────────────────────────────────────────────────────────────────┘
```

### Directory Structure

```
pkg/
├── common/              # Utilities (logging, tracing, utils)
├── handler/             # gRPC event handlers (thin layer)
│   ├── oauth.go         # Receives OAuth events → emits signals
│   └── statistic.go     # Receives stat events → emits signals
│
├── signal/              # NEW: Signal definitions and processor
│   ├── signal.go        # Signal interface and common types
│   ├── processor.go     # Event → Signal conversion
│   └── types.go         # LoginSignal, WinSignal, RageQuitSignal, etc.
│
├── rule/                # NEW: Rule engine
│   ├── engine.go        # Rule evaluation orchestrator
│   ├── registry.go      # Plugin registry for rules
│   ├── rule.go          # Rule interface
│   │
│   ├── builtin/         # Built-in rules (examples)
│   │   ├── rage_quit.go
│   │   ├── losing_streak.go
│   │   └── session_decline.go
│   │
│   └── config.go        # Rule configuration structs
│
├── action/              # NEW: Action executor
│   ├── executor.go      # Action orchestrator
│   ├── registry.go      # Plugin registry for actions
│   ├── action.go        # Action interface
│   │
│   ├── builtin/         # Built-in actions (examples)
│   │   ├── create_challenge.go
│   │   ├── grant_item.go
│   │   ├── send_notification.go
│   │   └── webhook.go
│   │
│   └── config.go        # Action configuration structs
│
├── state/               # Player state management (refactored)
│   ├── models.go        # State data structures
│   ├── store.go         # Storage interface (Redis impl)
│   └── context.go       # PlayerContext (state + metadata)
│
├── pipeline/            # NEW: Rule → Action pipeline
│   ├── pipeline.go      # Connects rules to actions
│   └── config.go        # Pipeline configuration (YAML/JSON)
│
└── pb/                  # Generated protobuf code
```

---

## Core Components

### 1. Signal Processor

**Purpose**: Normalize heterogeneous events into domain signals with consistent schema.

**Responsibilities**:
- Receive raw events from gRPC handlers
- Enrich with player context (state, history, metadata)
- Emit structured signals to rule engine
- Handle signal validation and error cases

**Example Signals**:
- `LoginSignal`: Player logged in (from OAuth token event)
- `LogoutSignal`: Player logged out
- `WinSignal`: Player won a match
- `LossSignal`: Player lost a match
- `RageQuitSignal`: Player rage quit
- `StatUpdateSignal`: Generic stat update

### 2. Rule Engine

**Purpose**: Evaluate signals against configured rules and emit triggers.

**Responsibilities**:
- Maintain registry of available rules
- Match incoming signals to applicable rules
- Evaluate rule conditions (thresholds, state checks, cooldowns)
- Emit triggers when rules match
- Support rule prioritization and chaining
- Track rule evaluation metrics

**Rule Types**:
- **Threshold Rules**: Fire when metric exceeds value (e.g., 3+ rage quits)
- **Pattern Rules**: Detect sequences (e.g., 5 consecutive losses)
- **Time-based Rules**: Evaluate over time windows (e.g., session decline)
- **Composite Rules**: Combine multiple conditions with AND/OR logic

### 3. Action Executor

**Purpose**: Execute configured actions in response to triggers.

**Responsibilities**:
- Maintain registry of available actions
- Execute action pipelines (sequential or parallel)
- Handle action failures and retries
- Support rollback mechanisms (optional)
- Update player state after actions
- Emit action metrics and logs

**Action Types**:
- **State Actions**: Modify player state (create challenge, set cooldown)
- **API Actions**: Call external services (grant item, send notification)
- **Webhook Actions**: POST to external URLs
- **Composite Actions**: Execute multiple actions as one

---

## Interface Definitions

### Signal Interface

```go
// pkg/signal/signal.go

// Signal represents a normalized domain event with player context
type Signal interface {
    // Type returns the signal type identifier (e.g., "login", "rage_quit")
    Type() string
    
    // UserID returns the player identifier
    UserID() string
    
    // Timestamp returns when the signal occurred
    Timestamp() time.Time
    
    // Metadata returns additional signal-specific data
    Metadata() map[string]interface{}
    
    // Context returns enriched player context (state, history)
    Context() *PlayerContext
}

// PlayerContext wraps player state with additional metadata
type PlayerContext struct {
    UserID      string
    State       *state.ChurnState
    Namespace   string
    SessionInfo map[string]interface{}
}

// Example concrete signal
type RageQuitSignal struct {
    userID       string
    timestamp    time.Time
    quitCount    int
    matchContext map[string]interface{}
    context      *PlayerContext
}

func (s *RageQuitSignal) Type() string { return "rage_quit" }
func (s *RageQuitSignal) UserID() string { return s.userID }
func (s *RageQuitSignal) Timestamp() time.Time { return s.timestamp }
func (s *RageQuitSignal) Metadata() map[string]interface{} {
    return map[string]interface{}{
        "quit_count": s.quitCount,
        "match_context": s.matchContext,
    }
}
func (s *RageQuitSignal) Context() *PlayerContext { return s.context }
```

### Rule Interface

```go
// pkg/rule/rule.go

// Rule evaluates signals and emits triggers when conditions are met
type Rule interface {
    // ID returns unique rule identifier
    ID() string
    
    // Name returns human-readable rule name
    Name() string
    
    // SignalTypes returns which signal types this rule handles
    SignalTypes() []string
    
    // Evaluate checks if the signal matches rule conditions
    // Returns true and trigger data if rule matches, false otherwise
    Evaluate(ctx context.Context, signal signal.Signal) (bool, *Trigger, error)
    
    // Config returns the rule's configuration
    Config() RuleConfig
}

// Trigger represents a rule match that should execute actions
type Trigger struct {
    RuleID      string
    UserID      string
    Timestamp   time.Time
    Reason      string                 // Human-readable reason
    Metadata    map[string]interface{} // Rule-specific data for actions
    Priority    int                     // For action ordering
}

// RuleConfig is the base configuration for all rules
type RuleConfig struct {
    ID          string            `yaml:"id" json:"id"`
    Name        string            `yaml:"name" json:"name"`
    Enabled     bool              `yaml:"enabled" json:"enabled"`
    Priority    int               `yaml:"priority" json:"priority"`
    Cooldown    *CooldownConfig   `yaml:"cooldown,omitempty" json:"cooldown,omitempty"`
    Conditions  map[string]interface{} `yaml:"conditions" json:"conditions"`
}

// CooldownConfig defines rate limiting for rule triggers
type CooldownConfig struct {
    Duration time.Duration `yaml:"duration" json:"duration"`
    Scope    string        `yaml:"scope" json:"scope"` // "global" or "per_user"
}

// Example concrete rule
type RageQuitRule struct {
    config    RuleConfig
    threshold int
    cooldown  time.Duration
}

func (r *RageQuitRule) ID() string { return r.config.ID }
func (r *RageQuitRule) Name() string { return r.config.Name }
func (r *RageQuitRule) SignalTypes() []string { return []string{"rage_quit"} }

func (r *RageQuitRule) Evaluate(ctx context.Context, sig signal.Signal) (bool, *Trigger, error) {
    // Type assert to get specific signal
    rqSignal, ok := sig.(*signal.RageQuitSignal)
    if !ok {
        return false, nil, fmt.Errorf("expected RageQuitSignal")
    }
    
    // Check cooldown
    if r.isOnCooldown(sig.Context().State) {
        return false, nil, nil
    }
    
    // Check threshold
    quitCount := rqSignal.Metadata()["quit_count"].(int)
    if quitCount < r.threshold {
        return false, nil, nil
    }
    
    // Rule matched - create trigger
    trigger := &Trigger{
        RuleID:    r.ID(),
        UserID:    sig.UserID(),
        Timestamp: sig.Timestamp(),
        Reason:    fmt.Sprintf("Player rage quit %d times (threshold: %d)", quitCount, r.threshold),
        Metadata: map[string]interface{}{
            "quit_count": quitCount,
            "threshold":  r.threshold,
        },
        Priority: r.config.Priority,
    }
    
    return true, trigger, nil
}
```

### Action Interface

```go
// pkg/action/action.go

// Action performs operations in response to triggers
type Action interface {
    // ID returns unique action identifier
    ID() string
    
    // Name returns human-readable action name
    Name() string
    
    // Execute performs the action
    Execute(ctx context.Context, trigger *rule.Trigger, playerCtx *signal.PlayerContext) error
    
    // Rollback undoes the action (optional, can return ErrNotSupported)
    Rollback(ctx context.Context, trigger *rule.Trigger, playerCtx *signal.PlayerContext) error
    
    // Config returns the action's configuration
    Config() ActionConfig
}

// ActionConfig is the base configuration for all actions
type ActionConfig struct {
    ID         string                 `yaml:"id" json:"id"`
    Name       string                 `yaml:"name" json:"name"`
    Type       string                 `yaml:"type" json:"type"` // "create_challenge", "grant_item", etc.
    Enabled    bool                   `yaml:"enabled" json:"enabled"`
    Async      bool                   `yaml:"async" json:"async"`
    Retry      *RetryConfig           `yaml:"retry,omitempty" json:"retry,omitempty"`
    Parameters map[string]interface{} `yaml:"parameters" json:"parameters"`
}

// RetryConfig defines retry behavior for failed actions
type RetryConfig struct {
    MaxAttempts int           `yaml:"max_attempts" json:"max_attempts"`
    Delay       time.Duration `yaml:"delay" json:"delay"`
    Backoff     string        `yaml:"backoff" json:"backoff"` // "linear", "exponential"
}

// Example concrete action
type CreateChallengeAction struct {
    config       ActionConfig
    winsNeeded   int
    durationDays int
}

func (a *CreateChallengeAction) ID() string { return a.config.ID }
func (a *CreateChallengeAction) Name() string { return a.config.Name }

func (a *CreateChallengeAction) Execute(ctx context.Context, trigger *rule.Trigger, playerCtx *signal.PlayerContext) error {
    // Fetch current wins from state
    currentWins := playerCtx.State.GetTotalWins()
    
    // Create challenge in player state
    challenge := state.Challenge{
        Active:       true,
        WinsNeeded:   a.winsNeeded,
        WinsCurrent:  0,
        WinsAtStart:  currentWins,
        ExpiresAt:    time.Now().Add(time.Duration(a.durationDays) * 24 * time.Hour),
        TriggerReason: trigger.Reason,
    }
    
    playerCtx.State.Challenge = challenge
    
    // State will be saved by pipeline
    return nil
}

func (a *CreateChallengeAction) Rollback(ctx context.Context, trigger *rule.Trigger, playerCtx *signal.PlayerContext) error {
    playerCtx.State.Challenge.Active = false
    return nil
}
```

### Registry Pattern

```go
// pkg/rule/registry.go

// Registry manages available rules
type Registry struct {
    rules map[string]Rule
    mu    sync.RWMutex
}

func NewRegistry() *Registry {
    return &Registry{
        rules: make(map[string]Rule),
    }
}

// Register adds a rule to the registry
func (r *Registry) Register(rule Rule) error {
    r.mu.Lock()
    defer r.mu.Unlock()
    
    if _, exists := r.rules[rule.ID()]; exists {
        return fmt.Errorf("rule %s already registered", rule.ID())
    }
    
    r.rules[rule.ID()] = rule
    return nil
}

// GetBySignalType returns all rules that handle a signal type
func (r *Registry) GetBySignalType(signalType string) []Rule {
    r.mu.RLock()
    defer r.mu.RUnlock()
    
    var matching []Rule
    for _, rule := range r.rules {
        for _, st := range rule.SignalTypes() {
            if st == signalType {
                matching = append(matching, rule)
                break
            }
        }
    }
    
    return matching
}

// Similar pattern for pkg/action/registry.go
```

### Pipeline Configuration

```go
// pkg/pipeline/pipeline.go

// Pipeline connects rules to actions
type Pipeline struct {
    Name        string
    Rules       []string              // Rule IDs to evaluate
    Actions     map[string][]string   // Rule ID → Action IDs
    Concurrent  bool                  // Execute actions in parallel
}

// PipelineManager orchestrates signal → rule → action flow
type PipelineManager struct {
    signalProcessor *signal.Processor
    ruleEngine      *rule.Engine
    actionExecutor  *action.Executor
    pipelines       map[string]*Pipeline
}

func (pm *PipelineManager) ProcessSignal(ctx context.Context, sig signal.Signal) error {
    // 1. Evaluate rules
    triggers, err := pm.ruleEngine.Evaluate(ctx, sig)
    if err != nil {
        return fmt.Errorf("rule evaluation failed: %w", err)
    }
    
    // 2. For each trigger, execute configured actions
    for _, trigger := range triggers {
        pipeline := pm.getPipelineForRule(trigger.RuleID)
        if pipeline == nil {
            continue
        }
        
        actions := pipeline.Actions[trigger.RuleID]
        if err := pm.actionExecutor.Execute(ctx, trigger, sig.Context(), actions, pipeline.Concurrent); err != nil {
            return fmt.Errorf("action execution failed: %w", err)
        }
    }
    
    return nil
}
```

---

## Implementation Examples

### Example 1: Built-in Rage Quit Rule

```go
// pkg/rule/builtin/rage_quit.go

package builtin

import (
    "context"
    "fmt"
    "time"
    
    "extends-anti-churn/pkg/rule"
    "extends-anti-churn/pkg/signal"
)

type RageQuitRule struct {
    config    rule.RuleConfig
    threshold int
    cooldown  time.Duration
}

func NewRageQuitRule(config rule.RuleConfig) (*RageQuitRule, error) {
    threshold, ok := config.Conditions["threshold"].(int)
    if !ok {
        return nil, fmt.Errorf("threshold must be an integer")
    }
    
    var cooldown time.Duration
    if config.Cooldown != nil {
        cooldown = config.Cooldown.Duration
    }
    
    return &RageQuitRule{
        config:    config,
        threshold: threshold,
        cooldown:  cooldown,
    }, nil
}

func (r *RageQuitRule) ID() string { return r.config.ID }
func (r *RageQuitRule) Name() string { return r.config.Name }
func (r *RageQuitRule) SignalTypes() []string { return []string{"rage_quit"} }
func (r *RageQuitRule) Config() rule.RuleConfig { return r.config }

func (r *RageQuitRule) Evaluate(ctx context.Context, sig signal.Signal) (bool, *rule.Trigger, error) {
    if !r.config.Enabled {
        return false, nil, nil
    }
    
    // Type assertion
    metadata := sig.Metadata()
    quitCount, ok := metadata["quit_count"].(int)
    if !ok {
        return false, nil, fmt.Errorf("quit_count not found in signal metadata")
    }
    
    // Check cooldown
    playerState := sig.Context().State
    if r.cooldown > 0 && time.Now().Before(playerState.Intervention.CooldownUntil) {
        return false, nil, nil // Still on cooldown
    }
    
    // Check threshold
    if quitCount < r.threshold {
        return false, nil, nil
    }
    
    // Rule matched!
    trigger := &rule.Trigger{
        RuleID:    r.ID(),
        UserID:    sig.UserID(),
        Timestamp: sig.Timestamp(),
        Reason:    fmt.Sprintf("Player rage quit %d times (threshold: %d)", quitCount, r.threshold),
        Metadata: map[string]interface{}{
            "quit_count": quitCount,
            "threshold":  r.threshold,
        },
        Priority: r.config.Priority,
    }
    
    return true, trigger, nil
}
```

### Example 2: Custom User-Defined Rule

```go
// examples/custom_rules/weekend_warrior.go

package custom_rules

import (
    "context"
    "time"
    
    "extends-anti-churn/pkg/rule"
    "extends-anti-churn/pkg/signal"
)

// WeekendWarriorRule triggers when player only plays on weekends
// Encourages weekday engagement
type WeekendWarriorRule struct {
    config          rule.RuleConfig
    minWeekendLogins int
    maxWeekdayLogins int
}

func (r *WeekendWarriorRule) Evaluate(ctx context.Context, sig signal.Signal) (bool, *rule.Trigger, error) {
    if sig.Type() != "login" {
        return false, nil, nil
    }
    
    state := sig.Context().State
    weekendLogins := r.countWeekendLogins(state.Sessions)
    weekdayLogins := r.countWeekdayLogins(state.Sessions)
    
    if weekendLogins >= r.minWeekendLogins && weekdayLogins <= r.maxWeekdayLogins {
        trigger := &rule.Trigger{
            RuleID:    r.ID(),
            UserID:    sig.UserID(),
            Timestamp: sig.Timestamp(),
            Reason:    "Player only active on weekends - encourage weekday play",
            Metadata: map[string]interface{}{
                "weekend_logins": weekendLogins,
                "weekday_logins": weekdayLogins,
            },
        }
        return true, trigger, nil
    }
    
    return false, nil, nil
}

// Helper methods...
func (r *WeekendWarriorRule) countWeekendLogins(sessions SessionState) int { /* ... */ }
func (r *WeekendWarriorRule) countWeekdayLogins(sessions SessionState) int { /* ... */ }
```

### Example 3: Built-in Grant Item Action

```go
// pkg/action/builtin/grant_item.go

package builtin

import (
    "context"
    "fmt"
    
    "extends-anti-churn/pkg/action"
    "extends-anti-churn/pkg/rule"
    "extends-anti-churn/pkg/signal"
    
    "github.com/AccelByte/accelbyte-go-sdk/platform-sdk/pkg/platformclient/fulfillment"
)

type GrantItemAction struct {
    config              action.ActionConfig
    fulfillmentService  *fulfillment.Fulfillment
    itemID              string
    quantity            int
    namespace           string
}

func NewGrantItemAction(config action.ActionConfig, fulfillmentSvc *fulfillment.Fulfillment, namespace string) (*GrantItemAction, error) {
    itemID, ok := config.Parameters["item_id"].(string)
    if !ok {
        return nil, fmt.Errorf("item_id parameter required")
    }
    
    quantity := 1
    if q, ok := config.Parameters["quantity"].(int); ok {
        quantity = q
    }
    
    return &GrantItemAction{
        config:             config,
        fulfillmentService: fulfillmentSvc,
        itemID:             itemID,
        quantity:           quantity,
        namespace:          namespace,
    }, nil
}

func (a *GrantItemAction) ID() string { return a.config.ID }
func (a *GrantItemAction) Name() string { return a.config.Name }
func (a *GrantItemAction) Config() action.ActionConfig { return a.config }

func (a *GrantItemAction) Execute(ctx context.Context, trigger *rule.Trigger, playerCtx *signal.PlayerContext) error {
    request := &fulfillment.FulfillItemParams{
        Namespace: a.namespace,
        UserID:    trigger.UserID,
        Body: &platformclientmodels.FulfillmentRequest{
            ItemID:   &a.itemID,
            Quantity: int32(a.quantity),
            Source:   "REWARD", // or "ACHIEVEMENT", "PROMOTION"
        },
    }
    
    _, err := a.fulfillmentService.FulfillItemShort(request)
    if err != nil {
        return fmt.Errorf("failed to grant item %s to user %s: %w", a.itemID, trigger.UserID, err)
    }
    
    // Log success
    log.Infof("Granted %dx %s to user %s (trigger: %s)", a.quantity, a.itemID, trigger.UserID, trigger.Reason)
    
    return nil
}

func (a *GrantItemAction) Rollback(ctx context.Context, trigger *rule.Trigger, playerCtx *signal.PlayerContext) error {
    // Items cannot be easily revoked in AccelByte
    return action.ErrRollbackNotSupported
}
```

### Example 4: Custom Webhook Action

```go
// examples/custom_actions/webhook.go

package custom_actions

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "time"
    
    "extends-anti-churn/pkg/action"
    "extends-anti-churn/pkg/rule"
    "extends-anti-churn/pkg/signal"
)

// WebhookAction sends HTTP POST to external endpoint
type WebhookAction struct {
    config     action.ActionConfig
    url        string
    headers    map[string]string
    httpClient *http.Client
}

func NewWebhookAction(config action.ActionConfig) (*WebhookAction, error) {
    url, ok := config.Parameters["url"].(string)
    if !ok {
        return nil, fmt.Errorf("url parameter required")
    }
    
    headers := make(map[string]string)
    if h, ok := config.Parameters["headers"].(map[string]string); ok {
        headers = h
    }
    
    return &WebhookAction{
        config:  config,
        url:     url,
        headers: headers,
        httpClient: &http.Client{
            Timeout: 10 * time.Second,
        },
    }, nil
}

func (a *WebhookAction) Execute(ctx context.Context, trigger *rule.Trigger, playerCtx *signal.PlayerContext) error {
    // Build payload
    payload := map[string]interface{}{
        "event":     "churn_detected",
        "user_id":   trigger.UserID,
        "rule_id":   trigger.RuleID,
        "reason":    trigger.Reason,
        "timestamp": trigger.Timestamp.Format(time.RFC3339),
        "metadata":  trigger.Metadata,
    }
    
    jsonData, err := json.Marshal(payload)
    if err != nil {
        return fmt.Errorf("failed to marshal webhook payload: %w", err)
    }
    
    // Create request
    req, err := http.NewRequestWithContext(ctx, "POST", a.url, bytes.NewBuffer(jsonData))
    if err != nil {
        return fmt.Errorf("failed to create webhook request: %w", err)
    }
    
    req.Header.Set("Content-Type", "application/json")
    for k, v := range a.headers {
        req.Header.Set(k, v)
    }
    
    // Send request
    resp, err := a.httpClient.Do(req)
    if err != nil {
        return fmt.Errorf("webhook request failed: %w", err)
    }
    defer resp.Body.Close()
    
    if resp.StatusCode >= 400 {
        return fmt.Errorf("webhook returned error status: %d", resp.StatusCode)
    }
    
    return nil
}

func (a *WebhookAction) Rollback(ctx context.Context, trigger *rule.Trigger, playerCtx *signal.PlayerContext) error {
    return action.ErrRollbackNotSupported
}
```

---

## Configuration System

### YAML Configuration Example

**Design Decision**: This configuration uses inline `actions` in `rules` field for simplicity. Each rule explicitly declares which actions it triggers.

```yaml
# config/pipelines.yaml

version: "1.0"

# Rule definitions with inline action mappings
rules:
  - id: "rage_quit_detector"
    name: "Rage Quit Detection"
    type: "builtin.rage_quit"
    enabled: true
    actions: ["create_comeback_challenge", "grant_speed_booster", "webhook_analytics"]  # Actions triggered
    priority: 10
    cooldown:
      duration: "48h"
      scope: "per_user"
    conditions:
      threshold: 3
      
  - id: "losing_streak_detector"
    name: "Losing Streak Detection"
    type: "builtin.losing_streak"
    enabled: true
    actions: ["create_comeback_challenge", "grant_speed_booster", "webhook_analytics"]
    priority: 15
    cooldown:
      duration: "48h"
      scope: "per_user"
    conditions:
      threshold: 5
      consecutive: true
      
  - id: "session_decline_detector"
    name: "Session Frequency Decline"
    type: "builtin.session_decline"
    enabled: true
    actions: ["create_comeback_challenge", "send_comeback_email"]  # Different actions
    priority: 20
    conditions:
      min_last_week: 3
      max_this_week: 0
      window_days: 7

# Action definitions
actions:
  - id: "create_comeback_challenge"
    name: "Create Comeback Challenge"
    type: "builtin.create_challenge"
    enabled: true
    async: false
    parameters:
      wins_needed: 3
      duration_days: 7
      
  - id: "grant_speed_booster"
    name: "Grant Speed Booster Reward"
    type: "builtin.grant_item"
    enabled: true
    async: false
    retry:
      max_attempts: 3
      delay: "5s"
      backoff: "exponential"
    parameters:
      item_id: "${SPEED_BOOSTER_ITEM_ID}"  # From environment
      quantity: 1
      
  - id: "send_comeback_email"
    name: "Send Comeback Email"
    type: "builtin.send_notification"
    enabled: false  # Disabled by default
    async: true
    parameters:
      template: "comeback_challenge_created"
      channels: ["email", "push"]
      
  - id: "webhook_analytics"
    name: "Send Churn Event to Analytics"
    type: "custom.webhook"
    enabled: true
    async: true
    parameters:
      url: "https://analytics.example.com/churn-events"
      headers:
        Authorization: "Bearer ${ANALYTICS_API_KEY}"
```

**Benefits of Inline Actions**:
- ✅ Simpler structure (no separate pipelines section)
- ✅ Clear mapping at rule level
- ✅ Easier to understand and maintain
- ✅ Less configuration indirection
      - "session_decline_detector"
    actions:
      rage_quit_detector:
        - "create_comeback_challenge"
        - "grant_speed_booster"
        - "webhook_analytics"
      losing_streak_detector:
        - "create_comeback_challenge"
        - "grant_speed_booster"
        - "webhook_analytics"
      session_decline_detector:
        - "create_comeback_challenge"
        - "send_comeback_email"  # Different action for this rule
        - "webhook_analytics"
    concurrent: false  # Execute actions sequentially
```

### Loading Configuration

```go
// pkg/pipeline/config.go

package pipeline

import (
    "fmt"
    "os"
    
    "gopkg.in/yaml.v3"
    
    "extends-anti-churn/pkg/action"
    "extends-anti-churn/pkg/rule"
)

type Config struct {
    Version   string                `yaml:"version"`
    Rules     []rule.RuleConfig     `yaml:"rules"`
    Actions   []action.ActionConfig `yaml:"actions"`
}

type RuleConfig struct {
    ID         string                 `yaml:"id"`
    Type       string                 `yaml:"type"`
    Enabled    bool                   `yaml:"enabled"`
    Actions    []string               `yaml:"actions"`      // Action IDs to execute when rule triggers
    Priority   int                    `yaml:"priority"`
    Cooldown   *CooldownConfig        `yaml:"cooldown,omitempty"`
    Conditions map[string]interface{} `yaml:"conditions,omitempty"`
}

func LoadConfig(path string) (*Config, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("failed to read config file: %w", err)
    }
    
    // Expand environment variables
    expandedData := os.ExpandEnv(string(data))
    
    var config Config
    if err := yaml.Unmarshal([]byte(expandedData), &config); err != nil {
        return nil, fmt.Errorf("failed to parse config: %w", err)
    }
    
    // Validate configuration
    if err := config.Validate(); err != nil {
        return nil, fmt.Errorf("invalid config: %w", err)
    }
    
    return &config, nil
}

// Validate ensures config is valid
func (c *Config) Validate() error {
    // Build action ID set for validation
    actionIDs := make(map[string]bool)
    for _, action := range c.Actions {
        actionIDs[action.ID] = true
    }
    
    // Validate each rule's action references
    for _, rule := range c.Rules {
        for _, actionID := range rule.Actions {
            if !actionIDs[actionID] {
                return fmt.Errorf("rule %s references unknown action: %s", rule.ID, actionID)
            }
        }
    }
    
    return nil
}
    }
    
    return &config, nil
}

// Initialize creates rule/action instances from config
func (c *Config) Initialize(ruleRegistry *rule.Registry, actionRegistry *action.Registry, deps Dependencies) error {
    // Register rules
    for _, ruleConfig := range c.Rules {
        rule, err := createRule(ruleConfig, deps)
        if err != nil {
            return fmt.Errorf("failed to create rule %s: %w", ruleConfig.ID, err)
        }
        if err := ruleRegistry.Register(rule); err != nil {
            return err
        }
    }
    
    // Register actions
    for _, actionConfig := range c.Actions {
        action, err := createAction(actionConfig, deps)
        if err != nil {
            return fmt.Errorf("failed to create action %s: %w", actionConfig.ID, err)
        }
        if err := actionRegistry.Register(action); err != nil {
            return err
        }
    }
    
    return nil
}

// Factory functions
func createRule(config rule.RuleConfig, deps Dependencies) (rule.Rule, error) {
    switch config.Type {
    case "builtin.rage_quit":
        return builtin.NewRageQuitRule(config)
    case "builtin.losing_streak":
        return builtin.NewLosingStreakRule(config)
    case "builtin.session_decline":
        return builtin.NewSessionDeclineRule(config)
    default:
        return nil, fmt.Errorf("unknown rule type: %s", config.Type)
    }
}

func createAction(config action.ActionConfig, deps Dependencies) (action.Action, error) {
    switch config.Type {
    case "builtin.create_challenge":
        return builtin.NewCreateChallengeAction(config)
    case "builtin.grant_item":
        return builtin.NewGrantItemAction(config, deps.FulfillmentService, deps.Namespace)
    case "builtin.send_notification":
        return builtin.NewSendNotificationAction(config, deps.NotificationService)
    case "custom.webhook":
        return custom.NewWebhookAction(config)
    default:
        return nil, fmt.Errorf("unknown action type: %s", config.Type)
    }
}
```

---

## Migration Strategy

### Phase 1: Add New Components (Non-Breaking)

**Goal**: Introduce new architecture alongside existing code.

1. Create new packages: `signal/`, `rule/`, `action/`, `pipeline/`
2. Implement interfaces and base components
3. Port existing logic to built-in rules/actions
4. Add comprehensive unit tests for new components
5. Keep existing handlers functional

**Timeline**: 1-2 weeks

### Phase 2: Refactor Handlers (Parallel Operation)

**Goal**: Make handlers delegate to new system while maintaining compatibility.

```go
// pkg/handler/statistic.go (refactored)

func (h *StatisticHandler) OnMessage(ctx context.Context, msg *statistic.StatItemUpdated) error {
    // NEW: Convert event to signal
    sig, err := h.signalProcessor.ProcessStatEvent(ctx, msg)
    if err != nil {
        return err
    }
    
    // NEW: Process through pipeline
    if err := h.pipelineManager.ProcessSignal(ctx, sig); err != nil {
        log.Errorf("Pipeline processing failed: %v", err)
        // Fall back to old logic? Or fail?
        return err
    }
    
    // State is saved by action executor
    return nil
}
```

**Timeline**: 1 week

### Phase 3: Configuration & Documentation

**Goal**: Enable configuration-driven operation and document plugin system.

1. Create default `pipelines.yaml` configuration
2. Document plugin development guide
3. Create example custom rules/actions
4. Add configuration validation
5. Update README and ONBOARDING docs

**Timeline**: 1 week

### Phase 4: Cleanup & Release

**Goal**: Remove old code, finalize open source release.

1. Remove old detection logic from handlers
2. Deprecate unused functions
3. Add migration guide for existing deployments
4. Tag v2.0.0 release
5. Publish to GitHub with comprehensive docs

**Timeline**: 1 week

**Total Timeline**: 4-5 weeks for complete migration

---

## Benefits & Trade-offs

### Benefits

✅ **Modularity**: Rules and actions are independent, testable units  
✅ **Extensibility**: Users can add custom rules/actions without forking  
✅ **Configuration**: Adjust thresholds and actions via YAML without recompilation  
✅ **Testability**: Each component can be unit tested in isolation  
✅ **Reusability**: Actions can be reused across multiple rules  
✅ **Open Source Friendly**: Clear extension points and plugin examples  
✅ **Maintainability**: Separation of concerns makes code easier to understand  
✅ **Flexibility**: Different games can use different rule/action combinations  

### Trade-offs

⚠️ **Complexity**: More abstractions and interfaces to understand  
⚠️ **Performance**: Small overhead from signal processing and rule evaluation  
⚠️ **Migration Effort**: Requires refactoring existing code (4-5 weeks)  
⚠️ **Configuration Complexity**: YAML config can become large for many rules  
⚠️ **Learning Curve**: Contributors need to understand new architecture  

### Mitigation Strategies

- **Documentation**: Comprehensive guides and examples
- **Performance**: Benchmark critical paths, optimize hot code
- **Migration**: Phased approach with backward compatibility
- **Configuration**: Provide sensible defaults, validation, and examples
- **Learning**: Tutorial-style docs, video walkthroughs, plugin templates

---

## Open Source Considerations

### Repository Structure

```
extends-anti-churn/
├── README.md                    # Overview, quick start
├── ONBOARDING.md                # Detailed developer guide
├── LICENSE                      # Apache 2.0 or MIT
├── CONTRIBUTING.md              # Contribution guidelines
├── CODE_OF_CONDUCT.md           # Community standards
│
├── docs/                        # Documentation
│   ├── architecture.md          # This design doc
│   ├── plugin-development.md    # How to create custom rules/actions
│   ├── configuration.md         # Config reference
│   ├── api-reference.md         # Interface documentation
│   └── examples/                # Tutorial examples
│
├── examples/                    # Example implementations
│   ├── basic-setup/             # Minimal configuration
│   ├── custom-rule/             # Custom rule example
│   ├── custom-action/           # Custom action example
│   └── advanced-pipeline/       # Complex multi-rule setup
│
├── config/                      # Configuration files
│   ├── pipelines.yaml           # Default pipeline config
│   └── pipelines.example.yaml   # Template for users
│
├── pkg/                         # Core library (importable)
│   ├── signal/
│   ├── rule/
│   ├── action/
│   ├── pipeline/
│   └── state/
│
├── cmd/                         # Executables
│   ├── server/                  # Main gRPC server
│   └── cli/                     # Optional CLI tool
│
└── test/                        # Integration tests
    ├── fixtures/
    └── integration/
```

### License Recommendations

**Option 1: Apache 2.0**
- More business-friendly
- Explicit patent grant
- Allows commercial use
- Used by most AccelByte projects

**Option 2: MIT**
- Simpler, more permissive
- Less legal overhead
- Easier for small teams

**Recommendation**: Apache 2.0 (aligns with AccelByte ecosystem)

### Plugin Discovery

Allow users to register external plugins:

```go
// main.go

func main() {
    // Load core config
    config := pipeline.LoadConfig("config/pipelines.yaml")
    
    // Load plugin registry (optional)
    if pluginDir := os.Getenv("PLUGIN_DIR"); pluginDir != "" {
        plugins := loadPlugins(pluginDir)
        for _, p := range plugins {
            ruleRegistry.Register(p.Rules...)
            actionRegistry.Register(p.Actions...)
        }
    }
    
    // Initialize system
    // ...
}
```

### Documentation Priorities

1. **Quick Start**: Get running in 5 minutes with default config
2. **Plugin Tutorial**: Create custom rule step-by-step
3. **API Reference**: Interface documentation with examples
4. **Configuration Guide**: All YAML options explained
5. **Architecture Overview**: This design doc (simplified)
6. **Use Cases**: Real-world scenarios and solutions

### Community Features

- **GitHub Discussions**: For Q&A and feature requests
- **Example Plugins**: Showcase community contributions
- **Plugin Marketplace**: (Future) Directory of third-party plugins
- **CI/CD Templates**: GitHub Actions for testing plugins
- **Docker Compose**: Easy local development setup

---

## Next Steps

### Immediate Actions

1. **Review & Approve Design**: Gather feedback on this proposal
2. **Prototype Core Interfaces**: Implement signal, rule, action interfaces
3. **Port One Rule**: Convert rage quit detector to new architecture
4. **Validate Approach**: Ensure design meets extensibility goals

### Validation Questions

- [ ] Does the interface design feel intuitive for plugin developers?
- [ ] Is the configuration system flexible enough for diverse use cases?
- [ ] Are we over-engineering? Could we simplify further?
- [ ] What's missing from the plugin API?
- [ ] How do we handle versioning of plugins vs core?

### Feedback Requested

Please review this design and provide feedback on:

1. **Interface Design**: Are the interfaces clean and extensible?
2. **Configuration Approach**: Is YAML the right choice? Too complex?
3. **Migration Strategy**: Is the phased approach reasonable?
4. **Open Source Readiness**: What else is needed for OS release?
5. **Performance**: Any concerns about the overhead?
6. **Alternative Approaches**: Are there better patterns we should consider?

---

**Document Status**: 📝 Draft - Awaiting Review  
**Author**: GitHub Copilot (AI Assistant)  
**Review Requested**: February 11, 2026  
**Target Implementation**: Q1 2026
