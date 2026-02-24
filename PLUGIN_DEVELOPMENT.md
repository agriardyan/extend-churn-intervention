# Plugin Development Guide

This guide explains how to extend the churn intervention system with custom plugins. The system uses a plugin-based architecture where **Signals**, **Rules**, and **Actions** are all pluggable components that you add without modifying the core framework.

## Table of Contents

- [Design Principles](#design-principles)
- [Architecture Overview](#architecture-overview)
- [Core Concepts](#core-concepts)
  - [Signals](#signals)
  - [Rules](#rules)
  - [Actions](#actions)
  - [Signal → Rule → Action Contract](#signal--rule--action-contract)
- [Adding a New Stat Listener](#adding-a-new-stat-listener)
- [Adding a New Event Type Handler](#adding-a-new-event-type-handler)
- [Adding a New Rule](#adding-a-new-rule)
- [Adding a New Action](#adding-a-new-action)
- [Configuration](#configuration)
- [Testing Your Plugin](#testing-your-plugin)
- [Best Practices](#best-practices)
- [Plugin Checklist](#plugin-checklist)
- [Examples in the Codebase](#examples-in-the-codebase)

---

## Design Principles

1. **Pluggability** — All domain logic lives in `builtin` packages; core packages define only interfaces. You never modify core code to add a plugin.
2. **Explicitness over Abstraction** — Code is self-contained and clear rather than DRY. Prefer explicit wiring over clever helpers.
3. **Dependency Inversion** — Core depends on abstractions, implementations depend on core. Services are injected via `Dependencies` structs.
4. **Registry Pattern** — All extension points use thread-safe registries. Register once at startup; lookup at runtime.
5. **Event-Driven** — The system reacts to player events in real-time via gRPC calls from Kafka Connect.

**Technology stack:** Go 1.23 · AccelByte AGS (gRPC AsyncAPI) · Redis · YAML config · miniredis for testing

### Framework Philosophy

This framework is **non-opinionated** about what plugins do internally. Rules may call external services. Actions may write to external systems, including game stats (e.g., writing a stat to trigger a challenge in Extend Challenge Service is a valid pattern). You decide what's appropriate for your use case.

**The only constraint is: avoid circular event loops.**

```
❌ Rule listens to event X → Action writes X → triggers event X again → infinite loop
✅ Rule listens to event X → Action writes to a different system
```

---

## Architecture Overview

```
┌──────────────────────────────────────────────────────────────┐
│             AccelByte Platform (via Kafka Connect)           │
│                 (OAuth Events, Stat Updates)                 │
└─────────────────────────┬────────────────────────────────────┘
                          │ gRPC AsyncAPI
                          ▼
┌──────────────────────────────────────────────────────────────┐
│                      Event Handlers                          │
│                (OAuth Handler, Stat Handler)                 │
└─────────────────────────┬────────────────────────────────────┘
                          │ Raw Events
                          ▼
┌──────────────────────────────────────────────────────────────┐
│                     Signal Processor                         │
│  ┌─────────────────┐          ┌──────────────────────────┐   │
│  │ EventProcessors │ ──────►  │  Player Context Loader   │   │
│  │ (per event type)│          │       (Redis)            │   │
│  └─────────────────┘          └──────────────────────────┘   │
└─────────────────────────┬────────────────────────────────────┘
                          │ Enriched Signals
                          ▼
┌──────────────────────────────────────────────────────────────┐
│                     Pipeline Manager                         │
└─────────────────────────┬────────────────────────────────────┘
                          │ Signals
                          ▼
┌──────────────────────────────────────────────────────────────┐
│                       Rule Engine                            │
│  Rule Registry  ──►  Evaluation  ──►   Trigger Match         │
│  (rage_quit,          (signal            (priority sort)     │
│   losing_streak,       type filter)                          │
│   session_decline)                                           │
└─────────────────────────┬────────────────────────────────────┘
                          │ Triggers
                          ▼
┌──────────────────────────────────────────────────────────────┐
│                     Action Executor                          │
│  Action Registry  ──►  Execute  ──►  State Update /          │
│  (challenge,                         External API            │
│   grant_item, ...)                                           │
└──────────────────────────────────────────────────────────────┘
```

**Data flow:**
1. **Event Ingestion** — AccelByte sends OAuth login and stat update events via gRPC
2. **Signal Processing** — Events are normalized into domain signals enriched with player context (churn state, cooldowns)
3. **Rule Evaluation** — Signals are evaluated against registered rules to detect churn patterns
4. **Action Execution** — Triggered rules dispatch associated actions (challenges, item grants, notifications)

**Package layout:**
```
pkg/
├── signal/              # Signal processing (events → signals)
│   ├── builtin/         # Built-in event processors and signals
│   └── processor.go     # Core processor with EventProcessor registry
├── rule/                # Rule evaluation (signals → triggers)
│   ├── builtin/         # Built-in rules (rage_quit, losing_streak, session_decline)
│   └── engine.go        # Rule evaluation engine
├── action/              # Action execution (triggers → side effects)
│   ├── builtin/         # Built-in actions (comeback_challenge, grant_item, send_email)
│   └── executor.go      # Action execution coordinator with rollback
├── pipeline/            # Orchestration layer + startup validation
├── service/             # Service interfaces and state models
└── handler/             # gRPC event handlers (OAuth, stat updates)
```

---

## Core Concepts

Understanding what belongs in each component will prevent common mistakes.

### Signals

Signals are **normalized, enriched, read-only domain events**. They represent a player behavior snapshot at a specific moment in time.

**A signal carries:**
- Event-specific data (e.g., current streak length, quit count, login timestamp)
- Player context (user ID, churn state, cooldowns, active interventions)
- Timestamp of when the event occurred

**What belongs in a signal:**
- ✅ Event-specific data fields
- ✅ Player context loaded from Redis (cheap, always needed)
- ✅ Metadata for rules to make decisions

**What does NOT belong in a signal:**
- ❌ Business logic or conditional decisions
- ❌ State mutations or side effects
- ❌ Rule evaluation logic

**When to create a new signal:** When you have a new event type (new stat code, new gRPC event) and need to carry its data through the pipeline.

---

### Rules

Rules are **pattern detectors** that evaluate signals and decide whether to trigger an intervention.

**What belongs in a rule:**
- ✅ Conditional threshold checks (e.g., streak > 5)
- ✅ Pattern matching against signal data
- ✅ Cooldown guards (don't re-trigger during active interventions)
- ✅ Priority assignment for trigger ordering
- ✅ External service calls for data enrichment, when needed lazily after a threshold check passes (see below)

**Prefer to avoid in a rule:**
- ⚠️ Unconditional external API calls on every signal — prefer lazy loading after a quick check
- ⚠️ State mutations — those belong in actions where rollback support exists

**When to create a new rule:** When you want to detect a new churn pattern, or apply different thresholds to an existing pattern.

#### Where Does Data Enrichment Happen?

This is the most important architectural decision when building rules. Choose based on cost and usage frequency:

**Option 1: Enrich early in the Signal Processor** (for cheap, widely-needed data)
```
Event → Signal Processor → fetch common context → Enriched Signal → all rules
```
Use for:
- ✅ Player state from Redis (cheap, all rules need it)
- ✅ Active intervention history (cheap, multiple rules check it)

The `PlayerContext` loaded by the signal processor covers this base enrichment.

**Option 2: Lazy load inside the rule** (for expensive, conditionally-needed data)
```
Event → Signal → Rule → quick check passes → fetch expensive data → decision
```
Use for:
- ⚠️ External API calls (slow, network latency)
- ⚠️ Data only needed after a quick preliminary check passes
- ⚠️ Data needed by only one specific rule

**Example of two-stage rule evaluation:**
```go
func (r *ClanActivityRule) Evaluate(ctx context.Context, sig signal.Signal) (bool, *rule.Trigger, error) {
    // Stage 1: Quick check with pre-loaded context (no I/O)
    if sig.Context().State.Cooldown.IsOnCooldown() {
        return false, nil, nil
    }

    // Stage 2: Only fetch expensive data if quick check passes
    clanActivity, err := r.clanService.GetClanActivity(ctx, sig.UserID())
    if err != nil {
        return false, nil, err
    }

    if clanActivity.ActiveMembers > 5 {
        return true, rule.NewTrigger(r.ID(), sig.UserID(), "inactive player, active clan", r.config.Priority), nil
    }

    return false, nil, nil
}
```

**Trade-off summary:**

| Approach | Pros | Cons | When to Use |
|----------|------|------|-------------|
| **Enrich Early** | Rules stay stateless. Consistent data across rules. Easy to test. | Every signal pays the cost even if unused. | Cheap data. Multiple rules use it. |
| **Lazy Load** | Only pay the cost when needed. Short-circuit before expensive calls. | Rule needs a service dependency. Harder to test. | Expensive calls. Single rule uses it. Conditional on threshold. |

---

### Actions

Actions are the **stateful executors** — they're where side effects happen.

**What belongs in an action:**
- ✅ External API calls (grant items, create challenges, send notifications)
- ✅ Writes to external systems, including game stats when an integration requires it (e.g., a stat write to trigger an Extend Challenge)
- ✅ Error handling and retry logic

**What does NOT belong in an action:**
- ❌ Rule evaluation logic or pattern detection
- ❌ Conditional business logic that decides *whether* to act (that's the rule's job)
- ❌ Writes that feed back into the same event the rule is listening to (circular loop)

**When to create a new action:** When you need a new type of intervention or integration with a new external system.

**Key practices:**
- Design for idempotency when possible — safe to retry means safe to recover from failures
- Implement `Rollback()` for reversible operations; return `action.ErrRollbackNotSupported` for irreversible ones
- Keep actions focused: one action, one side effect

---

### Signal → Rule → Action Contract

```
┌────────────┐         ┌────────────┐         ┌────────────┐
│   SIGNAL   │────────>│    RULE    │────────>│   ACTION   │
│            │         │            │         │            │
│ "What      │         │ "Should we │         │ "What do   │
│  happened?"│         │  react?"   │         │  we do?"   │
│            │         │            │         │            │
│ Normalized │         │ Pattern    │         │ Side       │
│ Event +    │         │ Detection  │         │ Effects    │
│ Context    │         │            │         │            │
└────────────┘         └────────────┘         └────────────┘

   Read-Only            Decision Maker         Stateful
   Immutable            Produces Triggers      Executor
   Enriched                                    Changes World
```

This separation gives you:
- **Testability** — each component tested in isolation with no external dependencies
- **Reusability** — same signal triggers multiple rules; same action used by multiple rules
- **Composability** — mix and match rules and actions in `config/pipeline.yaml` without code changes
- **Maintainability** — changes isolated to one component

---

## Adding a New Stat Listener

**Use case:** You want to listen to a new stat code like `rse-player-level` or `rse-daily-quests-completed`.

### Step 1: Create the Event Processor

Create a new file in `pkg/signal/builtin/`:

```go
// pkg/signal/builtin/player_level_event_processor.go
package builtin

import (
    "context"
    "fmt"
    "time"

    asyncapi_social "github.com/AccelByte/extend-churn-intervention/pkg/pb/accelbyte-asyncapi/social/statistic/v1"
    "github.com/AccelByte/extend-churn-intervention/pkg/service"
    "github.com/AccelByte/extend-churn-intervention/pkg/signal"
)

type PlayerLevelEventProcessor struct {
    stateStore service.StateStore
    namespace  string
}

func NewPlayerLevelEventProcessor(stateStore service.StateStore, namespace string) *PlayerLevelEventProcessor {
    return &PlayerLevelEventProcessor{stateStore: stateStore, namespace: namespace}
}

// EventType returns the stat code this processor handles.
func (p *PlayerLevelEventProcessor) EventType() string {
    return "rse-player-level"  // The stat code to listen for
}

// Process converts the stat update event into a signal.
func (p *PlayerLevelEventProcessor) Process(ctx context.Context, event interface{}) (signal.Signal, error) {
    statEvent, ok := event.(*asyncapi_social.StatItemUpdated)
    if !ok {
        return nil, fmt.Errorf("invalid event type for player level processor")
    }

    userID := statEvent.GetUserId()

    playerState, err := p.stateStore.GetChurnState(ctx, userID)
    if err != nil {
        return nil, err
    }

    return &PlayerLevelSignal{
        userID:        userID,
        timestamp:     time.Now(),
        level:         int(statEvent.GetPayload().GetValue()),
        playerContext: &signal.PlayerContext{
            UserID:    userID,
            Namespace: p.namespace,
            State:     playerState,
        },
    }, nil
}
```

### Step 2: Create the Signal Type

```go
// pkg/signal/builtin/player_level_signal.go
package builtin

import (
    "time"
    "github.com/AccelByte/extend-churn-intervention/pkg/signal"
)

const TypePlayerLevel = "player_level"

type PlayerLevelSignal struct {
    userID        string
    timestamp     time.Time
    level         int
    playerContext *signal.PlayerContext
}

func (s *PlayerLevelSignal) Type() string {
    return TypePlayerLevel
}

func (s *PlayerLevelSignal) UserID() string {
    return s.userID
}

func (s *PlayerLevelSignal) Timestamp() time.Time {
    return s.timestamp
}

func (s *PlayerLevelSignal) Context() *signal.PlayerContext {
    return s.playerContext
}

func (s *PlayerLevelSignal) Level() int {
    return s.level
}
```

### Step 3: Register the Event Processor

Add registration in `pkg/signal/builtin/event_processors.go`:

```go
func RegisterEventProcessors(
    registry *signal.EventProcessorRegistry,
    stateStore service.StateStore,
    namespace string,
    deps *EventProcessorDependencies,
) {
    // Existing registrations...
    registry.Register(NewPlayerLevelEventProcessor(stateStore, namespace))
}
```

### Step 4: Create a Rule to Use It

See [Adding a New Rule](#adding-a-new-rule) below.

**That's it!** The stat update handler will automatically route `rse-player-level` stat updates to your processor.

---

## Adding a New Event Type Handler

**Use case:** You want to handle a completely new event type from Kafka Connect (e.g., party events, purchase events).

### Step 1: Define the Protobuf

Ensure you have the protobuf definitions in `pkg/pb/` for your event type.

### Step 2: Create the Event Processor

```go
// pkg/signal/builtin/party_event_processor.go
package builtin

type PartyEventProcessor struct {
    stateStore service.StateStore
    namespace  string
}

func (p *PartyEventProcessor) EventType() string {
    return "party_disbanded"  // Custom event type string
}

func (p *PartyEventProcessor) Process(ctx context.Context, event interface{}) (signal.Signal, error) {
    partyEvent, ok := event.(*pb_party.PartyDisbanded)
    if !ok {
        return nil, fmt.Errorf("invalid event type")
    }
    // Extract data and create signal...
}
```

### Step 3: Register the Processor

```go
// In pkg/signal/builtin/event_processors.go
registry.Register(NewPartyEventProcessor(stateStore, namespace))
```

### Step 4: Create the gRPC Handler

```go
// pkg/handler/party_handler.go
package handler

type PartyEventHandler struct {
    pb_party.UnimplementedPartyEventServiceServer
    pipelineManager *pipeline.Manager
}

func (h *PartyEventHandler) OnPartyDisbanded(
    ctx context.Context,
    event *pb_party.PartyDisbanded,
) (*pb_party.PartyResponse, error) {
    err := h.pipelineManager.ProcessEvent(ctx, "party_disbanded", event)
    if err != nil {
        return nil, err
    }
    return &pb_party.PartyResponse{Success: true}, nil
}
```

### Step 5: Register the Handler in main.go

```go
// In main.go, in the "DEVELOPER: Register your gRPC event handlers here" section:
partyHandler := handler.NewPartyEventHandler(pipelineManager)
pb_party.RegisterPartyEventServiceServer(s, partyHandler)
```

---

## Adding a New Rule

**Use case:** You want to detect a new churn pattern (e.g., "player stuck on a level for 2 weeks").

### Step 1: Create the Rule

```go
// pkg/rule/builtin/stuck_player.go
package builtin

import (
    "context"
    "github.com/AccelByte/extend-churn-intervention/pkg/rule"
    "github.com/AccelByte/extend-churn-intervention/pkg/signal"
    signalBuiltin "github.com/AccelByte/extend-churn-intervention/pkg/signal/builtin"
    "github.com/sirupsen/logrus"
)

const StuckPlayerRuleID = "stuck_player"

type StuckPlayerRule struct {
    config         rule.RuleConfig
    levelThreshold int
    daysThreshold  int
}

func NewStuckPlayerRule(config rule.RuleConfig) *StuckPlayerRule {
    logrus.Infof("creating stuck player rule with levelThreshold=%d",
        config.GetParameterInt("level_threshold", 5))
    return &StuckPlayerRule{
        config:         config,
        levelThreshold: config.GetParameterInt("level_threshold", 5),
        daysThreshold:  config.GetParameterInt("days_threshold", 14),
    }
}

func (r *StuckPlayerRule) ID() string {
    return r.config.ID
}

func (r *StuckPlayerRule) Name() string {
    return "Stuck Player Detection"
}

func (r *StuckPlayerRule) SignalTypes() []string {
    return []string{signalBuiltin.TypePlayerLevel}
}

func (r *StuckPlayerRule) Config() rule.RuleConfig {
    return r.config
}

func (r *StuckPlayerRule) Evaluate(ctx context.Context, sig signal.Signal) (bool, *rule.Trigger, error) {
    levelSignal, ok := sig.(*signalBuiltin.PlayerLevelSignal)
    if !ok {
        return false, nil, nil
    }

    // Check if player is stuck at low level
    if levelSignal.Level() <= r.levelThreshold {
        // Check if they've been stuck for long enough
        // (You'd implement time tracking logic here)

        trigger := rule.NewTrigger(
            r.ID(),
            sig.UserID(),
            "Player stuck at low level",
            r.config.Priority,
        )
        trigger.Metadata["level"] = levelSignal.Level()
        trigger.Metadata["threshold"] = r.levelThreshold

        return true, trigger, nil
    }

    return false, nil, nil
}
```

### Step 2: Register the Rule Type

```go
// In pkg/rule/builtin/init.go
func RegisterRules(deps *Dependencies) {
    // Existing rules...
    rule.RegisterRuleType(StuckPlayerRuleID, func(config rule.RuleConfig) (rule.Rule, error) {
        return NewStuckPlayerRule(config), nil
    })
}
```

If your rule needs a service dependency (like `LoginSessionTracker`), add it to the `Dependencies` struct and pass it through:

```go
type Dependencies struct {
    LoginSessionTracker service.LoginSessionTracker
    MyClanService       service.ClanService  // Add here
}

func RegisterRules(deps *Dependencies) {
    rule.RegisterRuleType(MyRuleID, func(config rule.RuleConfig) (rule.Rule, error) {
        return NewMyRule(config, deps.MyClanService), nil
    })
}
```

### Step 3: Configure in pipeline.yaml

```yaml
rules:
  - id: stuck-player-detection
    type: stuck_player
    enabled: true
    actions: [send-tutorial-hint]
    parameters:
      level_threshold: 5
      days_threshold: 14
```

---

## Adding a New Action

**Use case:** You want to execute a custom intervention (e.g., send a push notification).

### Step 1: Create the Action

```go
// pkg/action/builtin/send_push_notification.go
package builtin

import (
    "context"
    "github.com/AccelByte/extend-churn-intervention/pkg/action"
    "github.com/AccelByte/extend-churn-intervention/pkg/rule"
    "github.com/AccelByte/extend-churn-intervention/pkg/signal"
    "github.com/sirupsen/logrus"
)

const SendPushNotificationActionID = "send_push_notification"

type SendPushNotificationAction struct {
    config  action.ActionConfig
    message string
}

func NewSendPushNotificationAction(config action.ActionConfig) *SendPushNotificationAction {
    return &SendPushNotificationAction{
        config:  config,
        message: config.GetParameterString("message", "We miss you! Come back and claim your reward!"),
    }
}

func (a *SendPushNotificationAction) ID() string {
    return a.config.ID
}

func (a *SendPushNotificationAction) Name() string {
    return "Send Push Notification"
}

func (a *SendPushNotificationAction) Config() action.ActionConfig {
    return a.config
}

func (a *SendPushNotificationAction) Execute(
    ctx context.Context,
    trigger *rule.Trigger,
    playerCtx *signal.PlayerContext,
) error {
    // TODO: integrate with your push notification service
    logrus.Infof("[NO-OP] would send push notification to user %s: %s", trigger.UserID, a.message)
    return nil
}

func (a *SendPushNotificationAction) Rollback(
    ctx context.Context,
    trigger *rule.Trigger,
    playerCtx *signal.PlayerContext,
) error {
    return action.ErrRollbackNotSupported  // Push notifications can't be unsent
}
```

### Step 2: Register the Action Type

```go
// In pkg/action/builtin/init.go
func RegisterActions(deps *Dependencies) {
    // Existing actions...
    action.RegisterActionType(SendPushNotificationActionID, func(config action.ActionConfig) (action.Action, error) {
        return NewSendPushNotificationAction(config), nil
    })
}
```

If your action needs an external dependency, add it to the `Dependencies` struct:

```go
type Dependencies struct {
    StateStore         service.StateStore
    EntitlementGranter service.EntitlementGranter
    UserStatUpdater    service.UserStatisticUpdater
    PushService        PushNotificationService  // Add here
    Namespace          string
}
```

### Step 3: Configure in pipeline.yaml

```yaml
actions:
  - id: welcome-back-push
    type: send_push_notification
    enabled: true
    parameters:
      message: "We've prepared special rewards for you!"
```

### Step 4: Link to a Rule

```yaml
rules:
  - id: session-decline
    type: session_decline
    enabled: true
    actions: [grant-item, welcome-back-push]
```

---

## Configuration

### pipeline.yaml Structure

```yaml
# Rules define churn detection logic
rules:
  - id: unique-rule-id           # Must be unique across all rules
    type: rule_type              # Matches registered type ID (e.g., "stuck_player")
    enabled: true                # Set false to disable without removing
    actions: [action-id-1, ...]  # Action IDs to execute when triggered
    parameters:                  # Rule-specific parameters
      threshold: 5

# Actions define interventions
actions:
  - id: unique-action-id         # Must be unique; referenced by rules
    type: action_type            # Matches registered type ID
    enabled: true
    parameters:
      key: value
```

### Environment Variables in Config

Use `${ENV_VAR:default_value}` syntax for secrets and deployment-specific values:

```yaml
parameters:
  item_id: ${REWARD_ITEM_ID:COMEBACK_REWARD}
  api_key: ${PUSH_NOTIFICATION_KEY:test-key}
```

### Startup Validation

The pipeline validates all wiring at startup. If a rule references an action ID that isn't registered, or a rule type that doesn't exist, the service will **refuse to start** with a clear error. This prevents silent runtime failures from config typos.

---

## Testing Your Plugin

### Unit Tests

Structure tests as table-driven to cover multiple scenarios cleanly:

```go
// pkg/rule/builtin/stuck_player_test.go
func TestStuckPlayerRule_Evaluate(t *testing.T) {
    tests := []struct {
        name          string
        level         int
        expectTrigger bool
    }{
        {name: "triggers when below threshold", level: 3, expectTrigger: true},
        {name: "triggers at threshold", level: 5, expectTrigger: true},
        {name: "no trigger above threshold", level: 6, expectTrigger: false},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            r := NewStuckPlayerRule(rule.RuleConfig{
                ID:         "test-stuck",
                Type:       StuckPlayerRuleID,
                Enabled:    true,
                Parameters: map[string]interface{}{"level_threshold": 5},
            })

            sig := &PlayerLevelSignal{
                userID:        "test-user",
                timestamp:     time.Now(),
                level:         tt.level,
                playerContext: &signal.PlayerContext{
                    State: &service.ChurnState{},
                },
            }

            matched, trigger, err := r.Evaluate(context.Background(), sig)
            require.NoError(t, err)
            assert.Equal(t, tt.expectTrigger, matched)
            if tt.expectTrigger {
                assert.NotNil(t, trigger)
            }
        })
    }
}
```

### Testing with miniredis

For rules or actions that use Redis-backed services, use `miniredis`:

```go
func setupTestTracker(t *testing.T) service.LoginSessionTracker {
    mr, err := miniredis.Run()
    require.NoError(t, err)
    t.Cleanup(mr.Close)

    client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
    return service.NewRedisLoginSessionTrackingStore(client, service.RedisLoginSessionTrackingStoreConfig{})
}
```

### Manual Testing

1. **Seed Redis test data:**
   ```bash
   # Simulate a player who had activity last week but none this week
   redis-cli HSET "session_tracking:test-user" "202607" "5"
   ```

2. **Send a test event** using `grpcurl` or BloomRPC targeting port 6565.

3. **Check logs:**
   Check the app logs.

---

## Best Practices

### 1. Follow Scope Guidelines

Before implementing, verify your plugin fits the system scope:
- ✅ **DO**: Detect churn patterns and execute interventions
- ❌ **DON'T**: Update game state owned by other systems (game server, challenge system)

See `README.md` for detailed scope boundaries.

### 2. Use Dependency Injection

```go
// Good: dependencies injected, easy to mock in tests
type MyRule struct {
    config      rule.RuleConfig
    someService service.SomeService
}

// Bad: hard-coded dependencies, untestable
type MyRule struct {
    // makes direct Redis/API calls internally
}
```

### 3. Guard Against Nil Context

The player context and state may be nil for new players:

```go
playerCtx := sig.Context()
if playerCtx == nil || playerCtx.State == nil {
    return false, nil, nil
}
```

### 4. Handle Errors Gracefully

```go
func (a *MyAction) Execute(ctx context.Context, trigger *rule.Trigger, _ *signal.PlayerContext) error {
    if err := a.doSomething(ctx, trigger.UserID); err != nil {
        logrus.Errorf("my action failed for user %s: %v", trigger.UserID, err)
        return fmt.Errorf("my action: %w", err)
    }
    logrus.Infof("my action completed for user %s", trigger.UserID)
    return nil
}
```

### 5. Log Appropriately

```go
logrus.Infof("rule triggered for user %s: reason=%s", userID, reason)   // Important events
logrus.Debugf("evaluating condition: value=%d threshold=%d", v, t)       // Debug detail
logrus.Errorf("failed to process user %s: %v", userID, err)             // Failures
logrus.Warnf("skipping user %s: already on cooldown", userID)           // Non-critical
```

### 6. Document Intent

Add a comment at the top of your rule/action explaining:
- What pattern you're detecting (or what intervention you're executing)
- Why the default thresholds are set to what they are
- Any assumptions or known limitations

---

## Plugin Checklist

- [ ] Implementation file(s) created in `pkg/*/builtin/`
- [ ] Type registered in `init.go`
- [ ] Configuration added to `config/pipeline.yaml`
- [ ] Unit tests written (table-driven, covering happy path + edge cases)
- [ ] `make test` passes
- [ ] `make build` succeeds

---

## Examples in the Codebase

### Stat Listeners (Event Processors)
- `pkg/signal/builtin/rage_quit_event_processor.go` — Listens to `rse-rage-quit`
- `pkg/signal/builtin/losing_streak_event_processor.go` — Listens to `rse-current-losing-streak`
- `pkg/signal/builtin/oauth_event_processor.go` — Handles OAuth login events + increments weekly session count

### Rules
- `pkg/rule/builtin/rage_quit.go` — Simple threshold check on a stat value
- `pkg/rule/builtin/losing_streak.go` — Threshold check with cooldown guard
- `pkg/rule/builtin/session_decline.go` — Map-based weekly tracking with lazy service load

### Actions
- `pkg/action/builtin/dispatch_comeback_challenge.go` — External state write with rollback
- `pkg/action/builtin/grant_item.go` — External API call via AccelByte SDK
- `pkg/action/builtin/send-email.go` — No-op placeholder showing the interface

### Handlers
- `pkg/handler/oauth_event_handler.go` — OAuth login handler
- `pkg/handler/stat_event_handler.go` — Stat update handler (routes by stat code)

---

## Need Help?

- Check `README.md` for architectural guidelines and scope boundaries
- Review existing implementations in `pkg/*/builtin/` (e.g. `pkg/signal/builtin`, `pkg/rule/builtin`, `pkg/action/builtin`)
- Run `make test` to verify your changes
- Use the `/add-plugin` Claude Code skill to generate plugin scaffolding interactively
