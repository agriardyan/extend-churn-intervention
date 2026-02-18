# Plugin Development Guide

This guide explains how to extend the churn intervention system with custom plugins. The system uses a plugin-based architecture where **Signals**, **Rules**, and **Actions** are all pluggable components.

## Table of Contents

- [Architecture Overview](#architecture-overview)
- [Adding a New Stat Listener](#adding-a-new-stat-listener)
- [Adding a New Event Type Handler](#adding-a-new-event-type-handler)
- [Adding a New Rule](#adding-a-new-rule)
- [Adding a New Action](#adding-a-new-action)
- [Configuration](#configuration)
- [Testing Your Plugin](#testing-your-plugin)

---

## Architecture Overview

### Pipeline Flow

```
gRPC Events → Signal Processor → Rule Engine → Action Executor
     ↓              ↓                  ↓              ↓
  Handler     EventProcessor          Rule         Action
```

### Plugin Extension Points

1. **Event Processors** - Process events into signals (e.g., stat updates, OAuth events)
2. **Rules** - Evaluate signals to detect churn patterns
3. **Actions** - Execute interventions when rules trigger
4. **Handlers** - Receive gRPC events from Kafka Connect

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
    "github.com/AccelByte/extends-anti-churn/pkg/signal"
    "github.com/AccelByte/extends-anti-churn/pkg/service"
    asyncapi_social "github.com/AccelByte/extends-anti-churn/pkg/pb/accelbyte-asyncapi/social/statistic/v1"
)

type PlayerLevelEventProcessor struct {
    stateStore service.StateStore
    namespace  string
}

func NewPlayerLevelEventProcessor(stateStore service.StateStore, namespace string) *PlayerLevelEventProcessor {
    return &PlayerLevelEventProcessor{
        stateStore: stateStore,
        namespace:  namespace,
    }
}

// EventType returns the stat code this processor handles
func (p *PlayerLevelEventProcessor) EventType() string {
    return "rse-player-level"  // The stat code to listen for
}

// Process converts the stat update event into a signal
func (p *PlayerLevelEventProcessor) Process(ctx context.Context, event interface{}) (signal.Signal, error) {
    statEvent, ok := event.(*asyncapi_social.StatItemUpdated)
    if !ok {
        return nil, fmt.Errorf("invalid event type for player level processor")
    }

    userID := statEvent.GetUserId()

    // Load player context
    playerState, err := p.stateStore.GetChurnState(ctx, userID)
    if err != nil {
        return nil, err
    }

    playerContext := &signal.PlayerContext{
        UserID:    userID,
        Namespace: p.namespace,
        State:     playerState,
    }

    // Extract the player level from the stat
    level := int(statEvent.GetPayload().GetValue())

    // Create a signal with the level data
    return &PlayerLevelSignal{
        userID:        userID,
        timestamp:     time.Now(),
        level:         level,
        playerContext: playerContext,
    }, nil
}
```

### Step 2: Create the Signal Type

```go
// pkg/signal/builtin/player_level_signal.go
package builtin

import (
    "time"
    "github.com/AccelByte/extends-anti-churn/pkg/signal"
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
    deps *EventProcessorDependencies,
    namespace string,
) {
    // Existing registrations...

    // Register your new processor
    registry.Register(NewPlayerLevelEventProcessor(stateStore, namespace))
}
```

### Step 4: Create a Rule to Use It

See [Adding a New Rule](#adding-a-new-rule) below.

**That's it!** The stat update handler will automatically route `rse-player-level` stat updates to your processor.

---

## Adding a New Event Type Handler

**Use case:** You want to handle a completely new event type from Kafka Connect (e.g., party events, clan events, purchase events).

### Step 1: Define the Protobuf

First, ensure you have the protobuf definitions in `pkg/pb/` for your event type.

### Step 2: Create the Event Processor

```go
// pkg/signal/builtin/party_event_processor.go
package builtin

type PartyEventProcessor struct {
    stateStore service.StateStore
    namespace  string
}

func NewPartyEventProcessor(stateStore service.StateStore, namespace string) *PartyEventProcessor {
    return &PartyEventProcessor{
        stateStore: stateStore,
        namespace:  namespace,
    }
}

func (p *PartyEventProcessor) EventType() string {
    return "party_disbanded"  // Custom event type
}

func (p *PartyEventProcessor) Process(ctx context.Context, event interface{}) (signal.Signal, error) {
    partyEvent, ok := event.(*pb_party.PartyDisbanded)
    if !ok {
        return nil, fmt.Errorf("invalid event type")
    }

    // Extract data and create signal
    // ... implementation ...
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

import (
    "context"
    pb_party "github.com/AccelByte/extends-anti-churn/pkg/pb/party/v1"
    "github.com/AccelByte/extends-anti-churn/pkg/pipeline"
)

type PartyEventHandler struct {
    pb_party.UnimplementedPartyEventServiceServer
    pipelineManager *pipeline.Manager
}

func NewPartyEventHandler(pm *pipeline.Manager) *PartyEventHandler {
    return &PartyEventHandler{
        pipelineManager: pm,
    }
}

func (h *PartyEventHandler) OnPartyDisbanded(
    ctx context.Context,
    event *pb_party.PartyDisbanded,
) (*pb_party.PartyResponse, error) {
    // Process the event through the pipeline
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

**Done!** Your new event type will flow through the pipeline.

---

## Adding a New Rule

**Use case:** You want to detect a new churn pattern (e.g., "player stuck on level 5 for 2 weeks").

### Step 1: Create the Rule

```go
// pkg/rule/builtin/stuck_player.go
package builtin

import (
    "context"
    "github.com/AccelByte/extends-anti-churn/pkg/rule"
    "github.com/AccelByte/extends-anti-churn/pkg/signal"
    signalBuiltin "github.com/AccelByte/extends-anti-churn/pkg/signal/builtin"
)

const StuckPlayerRuleID = "stuck_player"

type StuckPlayerRule struct {
    config         rule.RuleConfig
    levelThreshold int
    daysThreshold  int
}

func NewStuckPlayerRule(config rule.RuleConfig) *StuckPlayerRule {
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

**Use case:** You want to execute a custom intervention (e.g., send push notification, create special quest).

### Step 1: Create the Action

```go
// pkg/action/builtin/send_push_notification.go
package builtin

import (
    "context"
    "github.com/AccelByte/extends-anti-churn/pkg/action"
    "github.com/AccelByte/extends-anti-churn/pkg/rule"
    "github.com/AccelByte/extends-anti-churn/pkg/signal"
)

const SendPushNotificationActionID = "send_push_notification"

type SendPushNotificationAction struct {
    config  action.ActionConfig
    message string
}

func NewSendPushNotificationAction(config action.ActionConfig) *SendPushNotificationAction {
    return &SendPushNotificationAction{
        config:  config,
        message: config.GetParameterString("message", "We miss you!"),
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
    // TODO: Integrate with push notification service
    logrus.Infof("Sending push notification to user %s: %s", trigger.UserID, a.message)
    return nil
}

func (a *SendPushNotificationAction) Rollback(
    ctx context.Context,
    trigger *rule.Trigger,
    playerCtx *signal.PlayerContext,
) error {
    // Push notifications cannot be rolled back
    return action.ErrRollbackNotSupported
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
    actions: [grant-item, welcome-back-push]  # Add your action
```

---

## Configuration

### pipeline.yaml Structure

```yaml
# Rules define churn detection logic
rules:
  - id: unique-rule-id           # Must be unique
    type: rule_type              # Matches registered rule type
    enabled: true                # Can disable without removing
    actions: [action-id-1, ...]  # Actions to trigger
    parameters:                  # Rule-specific parameters
      key: value

# Actions define interventions
actions:
  - id: unique-action-id         # Must be unique
    type: action_type            # Matches registered action type
    enabled: true                # Can disable without removing
    parameters:                  # Action-specific parameters
      key: value
```

### Environment Variables in Config

Use `${ENV_VAR:default_value}` syntax:

```yaml
parameters:
  item_id: ${REWARD_ITEM_ID:DEFAULT_ITEM}
  api_key: ${PUSH_NOTIFICATION_KEY:test-key}
```

---

## Testing Your Plugin

### Unit Tests

```go
// pkg/rule/builtin/stuck_player_test.go
package builtin

import (
    "context"
    "testing"
    "github.com/AccelByte/extends-anti-churn/pkg/rule"
)

func TestStuckPlayerRule_Evaluate(t *testing.T) {
    config := rule.RuleConfig{
        ID:       "test-stuck-player",
        Type:     StuckPlayerRuleID,
        Enabled:  true,
        Parameters: map[string]interface{}{
            "level_threshold": 5,
        },
    }

    rule := NewStuckPlayerRule(config)

    // Create test signal
    signal := &PlayerLevelSignal{
        userID:    "test-user",
        timestamp: time.Now(),
        level:     3,  // Below threshold
    }

    matched, trigger, err := rule.Evaluate(context.Background(), signal)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }

    if !matched {
        t.Error("expected rule to trigger")
    }

    if trigger == nil {
        t.Fatal("expected trigger to be created")
    }
}
```

### Integration Tests

```go
// Test the full pipeline
func TestStuckPlayerPipeline(t *testing.T) {
    // Setup pipeline with your rule and action
    // Send test event
    // Verify action was executed
}
```

### Manual Testing

1. **Add test data to Redis:**
   ```bash
   redis-cli HSET "session_tracking:test-user" "202608" "0"
   ```

2. **Send test event via gRPC:**
   Use tools like `grpcurl` or BloomRPC

3. **Check logs:**
   ```bash
   tail -f logs/anti-churn.log | grep "stuck_player"
   ```

---

## Best Practices

### 1. **Follow the Scope Guidelines**

✅ **DO**: Detect churn patterns and execute interventions
❌ **DON'T**: Update game state owned by other systems

See `CLAUDE.md` for detailed scope guidelines.

### 2. **Use Dependency Injection**

```go
// Good: Testable and flexible
type MyRule struct {
    service SomeService
}

// Bad: Hard to test
type MyRule struct {
    // Direct API calls inside
}
```

### 3. **Handle Errors Gracefully**

```go
func (a *MyAction) Execute(...) error {
    if err := a.doSomething(); err != nil {
        logrus.Errorf("failed to do something: %v", err)
        return fmt.Errorf("action failed: %w", err)
    }
    return nil
}
```

### 4. **Log Appropriately**

```go
logrus.Infof("rule triggered for user %s", userID)  // Important events
logrus.Debugf("checking condition: %v", value)      // Debug info
logrus.Errorf("failed to process: %v", err)         // Errors
```

### 5. **Document Your Plugin**

Add comments explaining:
- What pattern you're detecting
- Why the thresholds are set
- What the intervention does
- Any assumptions or limitations

---

## Plugin Checklist

When adding a new plugin, verify:

- [ ] Implementation files created
- [ ] Type registered in `init.go`
- [ ] Configuration added to `pipeline.yaml`
- [ ] Unit tests written
- [ ] Integration test added (if applicable)
- [ ] Documentation updated
- [ ] `make test` passes
- [ ] `make build` succeeds

---

## Examples in the Codebase

### Stat Listeners
- `pkg/signal/builtin/rage_quit_event_processor.go` - Listens to `rse-rage-quit`
- `pkg/signal/builtin/losing_streak_event_processor.go` - Listens to `rse-current-losing-streak`
- `pkg/signal/builtin/match_win_event_processor.go` - Listens to `rse-match-wins`

### Rules
- `pkg/rule/builtin/rage_quit.go` - Detects rage quit behavior
- `pkg/rule/builtin/losing_streak.go` - Detects losing streaks
- `pkg/rule/builtin/session_decline.go` - Detects session frequency decline

### Actions
- `pkg/action/builtin/dispatch_comeback_challenge.go` - Creates comeback challenges
- `pkg/action/builtin/grant_item.go` - Grants items via Platform API
- `pkg/action/builtin/send-email.go` - Email notification (no-op placeholder)

### Handlers
- `pkg/handler/oauth_event_handler.go` - Handles OAuth login events
- `pkg/handler/stat_event_handler.go` - Handles stat update events

---

## Need Help?

- Check `CLAUDE.md` for architectural guidelines
- Review existing implementations in `pkg/*/builtin/`
- Run `make test` to verify your changes
- Check logs for debugging: `tail -f logs/anti-churn.log`

---

**Happy Plugin Development! 🚀**
