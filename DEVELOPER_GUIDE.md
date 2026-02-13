# Developer Guide: Extend Churn Intervention (Event Handler)

## 1. Introduction

Welcome to the Extend Churn Intervention (Event Handler) developer guide. This system is designed to detect player churn signals and automatically trigger interventions to re-engage players before they leave the game.

The codebase is built with a **pluggable architecture** where all core components (signals, rules, actions) can be extended without modifying the core system. This guide will help you understand the architecture and how to extend it with your own custom implementations.

### Key Design Principles

1. **Pluggability** - All domain logic lives in `builtin` packages; core packages define only interfaces
2. **Explicitness over Abstraction** - Code is self-contained and clear rather than DRY
3. **Dependency Inversion** - Core depends on abstractions, implementations depend on core
4. **Registry Pattern** - All extensibility points use thread-safe registries
5. **Event-Driven** - System reacts to player events in real-time via gRPC call from Kafka Connect

### Technology Stack

- **Language**: Go 1.21+
- **Event Source**: AccelByte Platform (gRPC AsyncAPI)
- **State Storage**: Redis
- **Configuration**: YAML
- **Testing**: Go standard testing + miniredis

---

## 2. High-Level Architecture Overview

### System Components

```
┌─────────────────────────────────────────────────────────────────┐
│               AccelByte Platform (via Kafka Connect)            │
│                   (OAuth Events, Stat Updates)                  │
└────────────────────────┬────────────────────────────────────────┘
                         │ gRPC AsyncAPI
                         ▼
┌─────────────────────────────────────────────────────────────────┐
│                         Event Handlers                          │
│                   (OAuth Handler, Stat Handler)                 │
└────────────────────────┬────────────────────────────────────────┘
                         │ Raw Events
                         ▼
┌─────────────────────────────────────────────────────────────────┐
│                      Signal Processor                            │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────────┐ │
│  │   Event      │  │   Signal     │  │   Player Context     │ │
│  │  Processors  │→ │   Mappers    │→ │   Loader (Redis)     │ │
│  └──────────────┘  └──────────────┘  └──────────────────────┘ │
└────────────────────────┬────────────────────────────────────────┘
                         │ Enriched Signals
                         ▼
┌─────────────────────────────────────────────────────────────────┐
│                      Pipeline Manager                            │
└────────────────────────┬────────────────────────────────────────┘
                         │ Signals
                         ▼
┌─────────────────────────────────────────────────────────────────┐
│                        Rule Engine                               │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────────┐ │
│  │ Rule Registry│→ │  Evaluation  │→ │   Trigger Match      │ │
│  │ (Rage Quit,  │  │              │  │   (Priority Sort)    │ │
│  │  Losing      │  │              │  │                      │ │
│  │  Streak...)  │  │              │  │                      │ │
│  └──────────────┘  └──────────────┘  └──────────────────────┘ │
└────────────────────────┬────────────────────────────────────────┘
                         │ Triggers
                         ▼
┌─────────────────────────────────────────────────────────────────┐
│                       Action Executor                            │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────────┐ │
│  │    Action    │→ │   Execute    │→ │   State Update       │ │
│  │   Registry   │  │              │  │   (Challenge, Item)  │ │
│  │ (Challenge,  │  │              │  │                      │ │
│  │  Grant Item) │  │              │  │                      │ │
│  └──────────────┘  └──────────────┘  └──────────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
```

### Data Flow

1. **Event Ingestion**: AccelByte sends events (OAuth login, stat updates) via Kafka Connect gRPC
2. **Signal Processing**: Events are normalized into domain signals with player context
3. **Rule Evaluation**: Signals are evaluated against registered rules to detect churn patterns
4. **Action Execution**: Triggered rules execute associated actions (create challenges, grant items)

### Package Structure

```
pkg/
├── signal/              # Signal processing (events → signals)
│   ├── builtin/         # Built-in signals, mappers, event processors
│   ├── processor.go     # Core processor (registry-based routing)
│   └── signal.go        # Signal interface
│
├── rule/                # Rule evaluation (signals → triggers)
│   ├── builtin/         # Built-in rules (rage quit, losing streak, etc.)
│   ├── engine.go        # Rule evaluation engine
│   ├── factory.go       # Rule factory with type registry
│   └── rule.go          # Rule interface
│
├── action/              # Action execution (triggers → side effects)
│   ├── builtin/         # Built-in actions (challenges, item grants)
│   ├── executor.go      # Action execution coordinator
│   ├── factory.go       # Action factory with type registry
│   └── action.go        # Action interface
│
├── pipeline/            # Orchestration layer
│   └── manager.go       # Coordinates signal → rule → action flow
│
├── state/               # Player state management
│   ├── logic.go         # Business logic (churn detection, challenges)
│   └── redis.go         # Redis persistence
│
└── handler/             # gRPC event handlers
    ├── oauth.go         # OAuth token events
    └── statistic.go     # Stat update events
```

---

## 3. Core Concepts: Signals, Rules, and Actions

### Signals

**What Signals Are:**
- Normalized, enriched domain events
- Represent player behavior or state changes
- Carry both event data and player context (state, history)
- Immutable snapshots of a specific moment in time

**What Belongs in a Signal:**
- ✅ Event-specific data (e.g., total wins, quit count, login timestamp)
- ✅ Player context (user ID, churn state, session history)
- ✅ Metadata for rules to make decisions
- ✅ Timestamp of when the event occurred

**What Does NOT Belong in a Signal:**
- ❌ Business logic or decision-making
- ❌ State mutations or side effects
- ❌ Multiple event aggregations
- ❌ Rule evaluation logic
- ❌ Action execution

**Signal Responsibilities:**
1. Implement the `Signal` interface (Type, UserID, Timestamp, Metadata, Context)
2. Provide read-only access to event data
3. Enrich raw events with player context

**Example Signal Types:**
- `OauthTokenGeneratedSignal` - Player logged in
- `WinSignal` - Player won a match
- `LosingStreakSignal` - Player has consecutive losses
- `RageQuitSignal` - Player quit after losing

**When to Create a New Signal:**
- You have a new event type from AccelByte or another source
- You need to detect a new pattern of player behavior
- Existing signals don't capture the data you need

---

### Rules

**What Rules Are:**
- Pattern detectors that evaluate signals
- Stateless evaluators (all state comes from signal context)
- Decision makers that produce triggers when conditions are met
- Can have configurable parameters (thresholds, cooldowns)

**What Belongs in a Rule:**
- ✅ Conditional logic (if/else, comparisons)
- ✅ Threshold checks (e.g., "losing streak > 5")
- ✅ Pattern matching (e.g., "session decline > 50%")
- ✅ Cooldown management (don't trigger too frequently to avoid abuse)
- ✅ Priority assignment for trigger ordering

**What Does NOT Belong in a Rule:**
- ❌ Side effects (state changes, API calls, database writes)
- ❌ Action execution
- ❌ Complex business workflows
- ❌ Direct AccelByte API calls
- ❌ Heavyweight computations that should be pre-calculated

**Rule Responsibilities:**
1. Implement the `Rule` interface (ID, Name, SignalTypes, Evaluate, Config)
2. Declare which signal types it listens to
3. Evaluate signals and return match status + optional trigger
4. Be deterministic (same signal → same result)
5. Be stateless (all context from signal)

**Example Rules:**
- `RageQuitRule` - Detects when player quits after consecutive losses
- `LosingStreakRule` - Triggers on extended losing streaks
- `SessionDeclineRule` - Detects drop in play sessions week-over-week
- `ChallengeCompletionRule` - Detects when player completes a challenge

**When to Create a New Rule:**
- You need to detect a new churn pattern
- You want to trigger interventions based on new conditions
- You need different thresholds for the same pattern

**Rule Best Practices:**
- Keep rules focused and single-purpose
- Use parameters for configurability (don't hardcode thresholds)
- Document the pattern you're detecting
- Return descriptive trigger reasons
- Consider rule priority for action ordering

**Where Does Data Enrichment Happen?**

This depends on the **cost and usage pattern** of the enrichment:

**1. Lightweight, Frequently-Used Data → Signal Processor**

✅ **Enrich Early (Signal Processor):**
```
Event → Signal Processor → Fetch common context → Enriched Signal → Rules
```

Use this for:
- ✅ Player state from Redis (cheap, always needed)
- ✅ Challenge state (cheap, used by multiple rules)
- ✅ Session history (already in Redis, multiple rules check it)

The `PlayerContext` loaded by `LoadPlayerContext()` contains this base enrichment that most rules need.

**2. Expensive, Conditionally-Used Data → Lazy Loading in Rules**

⚠️ **Defer Enrichment (Lazy Loading):**
```
Event → Signal → Rule → Quick check → Threshold met? → Fetch expensive data → Decision
```

Use this for:
- ⚠️ External API calls (slow, network latency)
- ⚠️ Complex database queries (expensive)
- ⚠️ Data only needed after initial threshold check
- ⚠️ Data used by only ONE specific rule

**Example - Two-Stage Rule Evaluation:**
```go
func (r *ClanActivityRule) Evaluate(ctx context.Context, sig signal.Signal) (bool, *Trigger, error) {
    // Stage 1: Quick check using pre-loaded context
    if sig.Context().State.Sessions.ThisWeek > 10 {
        return false, nil, nil  // Player is active, no need to check clan
    }
    
    // Stage 2: Only fetch expensive data if threshold met
    // Player is inactive, now check if their clan is active (might re-engage them)
    clanActivity, err := r.clanService.GetClanActivity(ctx, sig.UserID())
    if err != nil {
        return false, nil, err
    }
    
    if clanActivity.ActiveMembers > 5 {
        return true, NewTrigger(r.id, sig.UserID(), "Inactive player, active clan"), nil
    }
    
    return false, nil, nil
}
```

**Performance Trade-offs:**

| Approach | Pros | Cons | When to Use |
|----------|------|------|-------------|
| **Enrich Early** | Rules stay stateless.<br>Consistent data across rules.<br>Easier to test. | Every signal pays enrichment cost.<br>Wasted work if rule doesn't trigger. | Cheap data.<br>Multiple rules use it.<br>Always needed. |
| **Lazy Load** | Only pay cost when needed.<br>Can short-circuit before expensive calls | Rules become stateful.<br>Harder to test.<br>Multiple rules may duplicate calls. | Expensive operations<br>Conditional on thresholds.<br>Rule-specific data. |

**Best Practice - Hybrid Approach:**
1. Load **common, cheap context** in Signal Processor (player state, challenges)
2. Load **expensive, conditional data** lazily in rules after threshold checks
3. **Cache** expensive data if multiple rules need it (pass via trigger metadata)

**Where Actions Fit:**
Actions ALWAYS make external calls - that's their job:
- Rules decide "should we act?" (quick evaluation, mostly from pre-loaded context)
- Actions execute "what do we do?" (can be slow, fetch/update external systems)

---

### Actions

**What Actions Are:**
- Side effects executed when rules trigger
- Stateful operations that change the world
- Can interact with external systems (AccelByte, database)
- Should be idempotent when possible

**What Belongs in an Action:**
- ✅ State mutations (create challenge, update cooldown)
- ✅ External API calls (grant items, send notifications)
- ✅ Database writes
- ✅ Complex workflows and orchestration
- ✅ Error handling and retries

**What Does NOT Belong in an Action:**
- ❌ Rule evaluation logic
- ❌ Pattern detection
- ❌ Conditional business logic (should be in rules)
- ❌ Signal creation or processing

**Action Responsibilities:**
1. Implement the `Action` interface (ID, Type, Execute)
2. Perform the side effect (create challenge, grant item, etc.)
3. Handle errors gracefully
4. Return execution result
5. Be idempotent if possible (safe to retry)

**Example Actions:**
- `ComebackChallengeAction` - Creates a time-limited challenge for the player
- `GrantItemAction` - Rewards player with an item via AccelByte API
- `SendNotificationAction` (future) - Sends push notification
- `AdjustMatchmakingAction` (future) - Tweaks matchmaking parameters

**When to Create a New Action:**
- You need a new type of intervention
- You want to integrate with a new external system
- You need complex multi-step workflows

**Action Best Practices:**
- Keep actions focused (single responsibility)
- Handle failures gracefully (log errors, return meaningful results)
- Consider idempotency (can it run twice safely?)
- Use retries for transient failures
- Document required dependencies (API credentials, etc.)

---

### The Signal-Rule-Action Contract

```
┌────────────┐         ┌────────────┐         ┌────────────┐
│   SIGNAL   │────────▶│    RULE    │────────▶│   ACTION   │
│            │         │            │         │            │
│ "What      │         │ "Should we │         │ "What do   │
│  happened?"│         │  react?"   │         │  we do?"   │
│            │         │            │         │            │
│ Normalized │         │ Pattern    │         │ Side       │
│ Event +    │         │ Detection  │         │ Effects    │
│ Context    │         │            │         │            │
└────────────┘         └────────────┘         └────────────┘

   Read-Only            Stateless              Stateful
   Immutable            Decision Maker         Executor
   Enriched             Produces Triggers      Changes World
```

**Key Separation of Concerns:**

1. **Signals**: "Here's what happened and the context around it"
2. **Rules**: "Based on this signal, should we do something?"
3. **Actions**: "Now actually do something about it"

This separation enables:
- ✅ Testability (each component tested in isolation)
- ✅ Reusability (same signal triggers multiple rules, same action used by multiple rules)
- ✅ Composability (mix and match rules and actions in config)
- ✅ Maintainability (changes isolated to one component)

---

## 4. Examples

This section provides concrete examples of how to extend the system with custom signals, rules, and actions.

### Example 1: Creating a Custom Signal
*[Placeholder - will be added]*

### Example 2: Creating a Custom Rule
*[Placeholder - will be added]*

### Example 3: Creating a Custom Action
*[Placeholder - will be added]*

### Example 4: Integrating with Pipeline Configuration
*[Placeholder - will be added]*

### Example 5: Testing Your Custom Components
*[Placeholder - will be added]*

---

## Next Steps

- Read the [Architecture Refactoring Documentation](EVENTPROCESSOR_REFACTORING.md) for details on the pluggable architecture
- Check [ONBOARDING.md](ONBOARDING.md) for environment setup
- Review built-in implementations in `pkg/*/builtin/` packages
- See [config/pipeline.yaml](config/pipeline.yaml) for configuration examples

## Contributing

When adding new components:
1. Follow the interface contracts strictly
2. Add comprehensive tests
3. Update configuration examples
4. Document your component's purpose and parameters
5. Use the explicit pattern (no abstraction helpers like BaseSignal)
6. Register your component in the appropriate registry

## Support

For questions or issues:
- Check existing implementations in `builtin` packages
- Review test files for usage examples
- Consult the architecture documentation
- Reach out to the team for guidance
