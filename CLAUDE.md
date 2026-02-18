# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Go-based churn intervention event handler for the AccelByte Extends platform. Detects player churn signals from game events and triggers automated interventions. Uses a plugin-based architecture with a Signal → Rule → Action pipeline.

**Module**: `github.com/AccelByte/extend-churn-intervention`

## 🎯 Critical: System Scope and Boundaries

**IMPORTANT**: This is a churn DETECTION and INTERVENTION system, NOT a general game state manager.

### What This System Does

✅ **Read-React Pattern**: Listen to events → Detect churn signals → Execute interventions

```
Game Events (read-only) → Detect Pattern → Intervene
```

**Valid operations:**
- Listen to game stats events (`rse-match-wins`, `rse-current-losing-streak`, `rse-rage-quit`)
- Detect behavioral patterns (losing streaks, session decline, rage quits)
- Create interventions (challenges, rewards)
- Track intervention history (what we did, when, cooldowns)
- Evaluate challenge completions (to grant rewards)

### What This System Does NOT Do

❌ **DO NOT implement features that:**
- Update game statistics owned by other systems
- Maintain primary game state (stats, progress, player data)
- Create circular event loops (listen to stat → update same stat)
- Replace functionality of existing systems (game server, challenge system)

**Examples of INVALID implementations:**

```go
// ❌ WRONG: Updating game stats
type RecordWinsAction struct {}
func (a *RecordWinsAction) Execute() error {
    return updateStatAPI("rse-match-wins", newValue)  // Game server owns this!
}

// ❌ WRONG: Circular dependency
type RecordWinsRule struct {}
// Listens to rse-match-wins event
// Executes action that updates rse-match-wins
// This creates an infinite loop!
```

**Examples of VALID implementations:**

```go
// ✅ CORRECT: Detecting churn and creating intervention
type LosingStreakRule struct {}
func (r *LosingStreakRule) Evaluate(sig signal.Signal) (bool, *Trigger, error) {
    // Detect if player is on losing streak
    // Return trigger to create comeback challenge
}

// ✅ CORRECT: Executing intervention
type ComebackChallengeAction struct {}
func (a *ComebackChallengeAction) Execute() error {
    // Create challenge in our state store
    // Challenge system will track actual progress
    return createChallenge()
}
```

### Ownership Table

| Component | Owner | Our Role |
|-----------|-------|----------|
| Game stats (`rse-*`) | Game Server | **Read Only** - Listen to events |
| Challenge progress | Challenge System | **Read Only** - Listen to completion |
| OAuth/Sessions | IAM Service | **Read Only** - Listen to login events |
| Churn detection rules | **Anti-Churn** | **Own** - Implement detection logic |
| Intervention actions | **Anti-Churn** | **Own** - Create challenges, grant rewards |
| Intervention history | **Anti-Churn** | **Own** - Track cooldowns, history |
| Session login counts | **Anti-Churn** | **Own** - Track weekly login counts in `session_tracking:*` Redis keys |

**Golden Rule**: If another system already owns it, we LISTEN to it, we don't UPDATE it.

### Implementation Checklist

Before implementing a new rule or action, ask:

1. ✅ **Is this churn DETECTION?** (analyzing signals to identify at-risk players)
2. ✅ **Is this INTERVENTION?** (creating challenges, granting rewards to re-engage)
3. ❌ **Am I updating state owned by another system?** (game stats, challenge progress)
4. ❌ **Does this create a circular dependency?** (listening to event → updating same event source)

If you answered YES to questions 3 or 4, **STOP** and reconsider the approach.

## Common Commands

```bash
make build        # Build Docker image (includes proto generation)
make test         # Run all tests: go test -v -cover ./...
make proto        # Generate protobuf files (requires Docker)
make clean        # Clean generated files and Docker images

# Run a single test
go test -v -run TestName ./pkg/signal/...

# Run tests for a specific package
go test -v -cover ./pkg/rule/...
```

## Architecture

### Pipeline Flow

```
gRPC Events → Signal Processor → Rule Engine → Action Executor
```

1. **Event Handlers** (`pkg/handler/`) receive gRPC events from Kafka Connect (OAuth login/logout, stat updates)
2. **Signal Processor** (`pkg/signal/`) normalizes events into signals via EventProcessors and enriches them with player context
3. **Rule Engine** (`pkg/rule/`) evaluates signals against configured rules and produces triggers
4. **Action Executor** (`pkg/action/`) executes triggered actions (e.g., create challenge, grant item) with rollback support
5. **Pipeline Manager** (`pkg/pipeline/`) orchestrates the full flow, maps rules to actions, and validates wiring at startup

### Plugin/Registry Pattern

Each domain layer (signal, rule, action) uses the same extensibility pattern:
- **Interface** defined in the package root (e.g., `pkg/rule/rule.go`)
- **Registry** provides thread-safe registration and lookup (`pkg/rule/registry.go`)
- **Factory** creates instances from YAML config (`pkg/rule/factory.go`)
- **Built-in implementations** live in `builtin/` subdirectory (e.g., `pkg/rule/builtin/`)
- **Registration** happens via `Register*()` functions called from `main.go`

To add a new rule/action/signal: implement the interface, create a factory, register it, and add config to `config/pipeline.yaml`.

### Extending the System

**BEFORE implementing any extension, verify it fits our scope (see "System Scope and Boundaries" above).**

**Adding a new stat code signal (e.g., `rse-other-stats`):**

1. **Verify**: This stat is owned by the game server, we're only listening to it (read-only)

2. Create an EventProcessor in `pkg/signal/builtin/`:
   ```go
   type OtherStatsEventProcessor struct {
       stateStore service.StateStore
       namespace  string
   }

   func NewOtherStatsEventProcessor(stateStore service.StateStore, namespace string) *OtherStatsEventProcessor {
       return &OtherStatsEventProcessor{stateStore: stateStore, namespace: namespace}
   }

   func (p *OtherStatsEventProcessor) EventType() string {
       return "rse-other-stats"  // The stat code to handle
   }

   func (p *OtherStatsEventProcessor) Process(ctx context.Context, event interface{}) (signal.Signal, error) {
       // Extract data from event
       // Load player context from our state store
       // Return signal (read-only operation!)
   }
   ```

3. Register it in `pkg/signal/builtin/event_processors.go`:
   ```go
   func RegisterEventProcessors(
       registry *signal.EventProcessorRegistry,
       stateStore service.StateStore,
       namespace string,
       deps *EventProcessorDependencies,
   ) {
       // ... existing processors ...
       registry.Register(NewOtherStatsEventProcessor(stateStore, namespace))
   }
   ```

4. The signal processor will automatically route stat events with that stat code to your processor.

**Adding a new rule that needs a service dependency:**

Rules that need service dependencies (e.g., `LoginSessionTracker`) receive them through the `Dependencies` struct in `pkg/rule/builtin/init.go`:

```go
// In pkg/rule/builtin/init.go
type Dependencies struct {
    LoginSessionTracker service.LoginSessionTracker
    // Add new dependencies here
}

func RegisterRules(deps *Dependencies) {
    rule.RegisterRuleType(MyRuleID, func(config rule.RuleConfig) (rule.Rule, error) {
        return NewMyRule(config, deps.LoginSessionTracker), nil
    })
}
```

### Signal Processing

The signal processor (`pkg/signal/processor.go`) uses a unified EventProcessor pattern:
- OAuth events are routed to `OAuthEventProcessor` (event type: `"oauth_token_generated"`)
  - Also calls `LoginSessionTracker.IncrementSessionCount()` to track weekly login counts
- Stat events are routed by stat code to per-stat-code processors
- Unknown stat codes fall back to generic `StatUpdateSignal`
- All processors implement the same `EventProcessor` interface

**Built-in signals:**
- `login` — OAuth token generated (player login), also increments weekly session count
- `rage_quit` — Player quit count from `rse-rage-quit` stat
- `losing_streak` — Consecutive losses from `rse-current-losing-streak` stat
- `match_win` — Total wins from `rse-match-wins` stat

### Session Tracking (Weekly Map)

The `LoginSessionTracker` service tracks weekly login counts using a Redis Hash with `YYYYWW` keys:

```go
type SessionTrackingData struct {
    LoginCount map[string]int // Key: "YYYYWW" (e.g., "202610"), Value: login count
}
```

- **Storage**: Redis Hash at key `session_tracking:{userID}`
- **Operations**: `HINCRBY` for atomic increment, `HGETALL` for retrieval, `HDEL` for cleanup
- **Retention**: 4 weeks (28 days TTL); entries older than 4 weeks are auto-deleted
- **Churn detection**: `SessionDeclineRule` uses this map to detect players who had activity in any recent week but have none in the current week — handles multi-week absences correctly

### Startup Validation

The pipeline validates its wiring at startup (`pkg/pipeline/validate.go`):
- Checks that all enabled rules in config are registered
- Checks that all enabled actions in config are registered
- Fails fast with clear error messages if configuration is incorrect
- Prevents silent runtime failures from typos or missing registrations

### Key Packages

- `pkg/service/` — External service abstractions and state models:
  - `churn_state.go` — `ChurnState`, `InterventionRecord`, `CooldownState` models
  - `login_session_tracker.go` — `SessionTrackingData` and `RedisLoginSessionTrackingStore`
  - `interfaces.go` — `StateStore`, `LoginSessionTracker`, `EntitlementGranter` interfaces
- `pkg/common/` — Logging, env helpers, OpenTelemetry setup
- `pkg/pb/` — Generated protobuf code (do not edit manually)
- `pkg/pipeline/` — Pipeline orchestration, configuration, and validation

### Configuration

- `config/pipeline.yaml` — Defines rules, actions, and their mappings. Supports `${ENV_VAR:default}` syntax in action parameters.
- `.env` / `.env.template` — Environment variables for local development. Required vars include `AB_BASE_URL`, `AB_CLIENT_ID`, `AB_CLIENT_SECRET`, `AB_NAMESPACE`, and Redis connection settings.

### Ports

- **6565** — gRPC server (receives events)
- **8080** — Prometheus metrics (`/metrics`)

## Key Dependencies

- `github.com/AccelByte/accelbyte-go-sdk` — Platform SDK for IAM auth and entitlements
- `github.com/go-redis/redis/v8` — State persistence
- `github.com/alicebob/miniredis/v2` — Redis mocking in tests
- `gopkg.in/yaml.v3` — Pipeline config parsing

## Developer Workflow

When adding new functionality, look for the section markers in `main.go`:

```go
// DEVELOPER: Register your event processors here.
signalBuiltin.RegisterEventProcessors(processor.GetEventProcessorRegistry(), stateStore, namespace, &signalBuiltin.EventProcessorDependencies{
    LoginTrackingStore: loginSessionTracker,
})

// DEVELOPER: Register your rule types here.
ruleBuiltin.RegisterRules(&ruleBuiltin.Dependencies{
    LoginSessionTracker: loginSessionTracker,
})

// DEVELOPER: Register your action types here.
actionBuiltin.RegisterActions(deps)

// DEVELOPER: Register your gRPC event handlers here.
pb_iam.RegisterOauthTokenOauthTokenGeneratedServiceServer(s, oauthListener)
```

These markers identify the four extension points for the pluggable architecture.

For a detailed developer walkthrough with code templates, see **`PLUGIN_DEVELOPMENT.md`**.

You can also use the `/add-plugin` Claude Code skill to interactively generate new plugin code.

## Testing

Each package has comprehensive tests:
- `pkg/signal/` — Signal processing and EventProcessor tests
- `pkg/rule/` — Rule evaluation and trigger generation tests
- `pkg/action/` — Action execution and rollback tests
- `pkg/pipeline/` — Integration tests, validation tests, manager tests
- `pkg/handler/` — Handler tests with mock pipeline

Run tests with `make test` or `go test ./...`

## When User Requests Are Out of Scope

If a user asks to implement features that violate the architectural boundaries:

1. **Explain the concern**: Point out the ownership conflict or circular dependency
2. **Reference this document**: Share the "System Scope and Boundaries" section
3. **Suggest alternatives**:
   - "This stat update should happen in the game server"
   - "Challenge progress tracking belongs in a challenge system"
   - "We can LISTEN to that event, but not UPDATE that data"
4. **Offer valid alternatives**: Show how to achieve the goal within proper boundaries

**Example response:**
> "I see you want to update `rse-match-wins` when a player wins. However, this violates our architectural boundary - the game server owns game statistics, and we only listen to them (read-only).
>
> If you need to track wins for churn intervention purposes, we should:
> 1. Listen to `rse-match-wins` events from the game server
> 2. Detect if wins indicate recovering engagement (no longer at-risk)
> 3. Store this in our intervention history
>
> The game server should remain the source of truth for `rse-match-wins`."

This maintains architectural integrity while helping users achieve their goals correctly.
