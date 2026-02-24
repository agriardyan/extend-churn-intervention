# Churn Intervention Event Handler

A plugin-based event handler for the AccelByte Extends platform that detects player churn signals and triggers automated interventions.

## Why is This?

**Player churn poses a significant threat to game success, particularly for newly launched titles.** When players become frustrated (e.g. from losing streaks, feeling stuck, or losing momentum) they may quit and never return. First impressions matter: early-stage player retention typically determines long-term success, with players who disengage in their first sessions being substantially less likely to return.

**Traditional analytics pipelines may respond too slowly for effective intervention.** Batch-processed analytics systems often operate on daily or longer cycles, creating delays between when a player experiences frustration and when that signal becomes actionable. This latency can mean the difference between successfully re-engaging a player and losing them permanently. Especially in moments of acute frustration like consecutive losses or other gameplay obstacles.

**This system enables near real-time detection and intervention.** By processing player events as they occur (logins, match results, stat updates), the system can identify potential churn signals within seconds and trigger automated interventions (e.g. granting rewards, creating comeback challenges, or offering personalized incentives), all while players are still engaged with the game, increasing the opportunity for successful re-engagement.

## What is This?

This service listens to game events (OAuth logins, stat updates) via Kafka Connect, analyzes player behavior patterns, and triggers interventions (challenges, rewards) to re-engage at-risk players. In this codebase, we provide three built-in churn detection rules (losing streaks, rage quits, session decline) and several intervention actions (dispatching comeback challenges, granting items, sending notifications). The system is designed to be easily extensible with new rules and actions.

**Example flows:**
- Player loses 5 matches in a row → `losing_streak` signal → "Comeback Challenge" (configured in Challenge Service, e.g. win 3 matches in 7 days)
- Player shows behavior of rage quit → `rage_quit` signal → "Comeback Challenge"
- Player's weekly logins drop to 0 (was active last week) → `session_decline` signal → Grant reward item + send email notification

## Use Cases

### You Should Use This If:

- **You need real-time churn intervention** — Your game requires immediate responses to player frustration (seconds/minutes, not hours/days)
- **You have event streams available** — Your game emits player events (logins, match results, stat updates) via AccelByte AGS
- **You want automated retention** — You want to trigger interventions (challenges, rewards, notifications) without manual oversight
- **You're on AccelByte Gaming Services** — This service is built specifically for the AccelByte Extends platform

**Common scenarios:**
- Competitive games where losing streaks cause immediate disengagement
- New game launches where first-session retention is critical
- Live ops teams wanting to A/B test different intervention strategies
- Games with comeback mechanics that need automatic activation

### What Can This Tool Actually Do?

**Detection capabilities:**
- Track login patterns and detect when players stop coming back
- Count consecutive match losses in real-time
- Identify rage quit behavior (quit immediately after losing)
- Monitor any stat-based behavioral patterns you configure

**Intervention capabilities (out-of-the-box):**
- **Create time-limited challenges** — "Win 3 matches in the next 7 days to earn rewards" (requires [fork of extend-challenge-service](https://github.com/agriardyan/extend-challenge-service))
- **Grant in-game items/currency** — Automatically give players entitlements via AccelByte Platform
- **Send notifications** — Email integration stub included (extend with your email provider)
- **Track intervention history** — Built-in cooldown system prevents spamming the same player

**Concrete example:**
```
Player "Sarah123" loses 5 matches in a row
  ↓
System detects losing_streak signal within 2 seconds
  ↓
Checks: Has Sarah received a comeback challenge in the last 7 days? (No)
  ↓
Creates challenge: "Win 3 matches in 7 days → Get 500 gems"
  ↓
Sarah sees the challenge in-game (via extend-challenge-service)
  ↓
Sarah wins 3 matches → Challenge auto-completes → 500 gems granted
```

**Another example:**
```
Player "Mike456" was active 2 weeks ago but hasn't logged in this week
  ↓
System detects session_decline signal
  ↓
Grants comeback reward: 1000 gold coins (immediate)
  ↓
Sends email: "We miss you! Here's 1000 gold to welcome you back"
```

## Architecture

### Pipeline Flow

```
Game Events → Signal Detection → Rule Evaluation → Action Execution
     ↓              ↓                    ↓                 ↓
  gRPC         EventProcessor         Rule              Action
  Handler      (normalize)         (should we?)       (intervene!)
```

### Key Concepts

- **Signal**: Normalized behavioral event (e.g., `login`, `rage_quit`, `losing_streak`)
- **Rule**: Logic to detect churn risk (e.g., "5+ consecutive losses")
- **Action**: Intervention to execute (e.g., create challenge, grant reward)
- **Pipeline**: Orchestrates the full flow with configurable rule-to-action mappings in `config/pipeline.yaml`

## 🎯 Design Philosophy

This framework is intentionally **non-opinionated** about what you do inside your plugins. Rules may call external services (for lazy data enrichment), and actions may write to external systems — including updating game stats when an integration requires it (e.g., triggering a challenge in Extend Challenge Service via a stat write).

**The one pattern to avoid is circular event loops:**

```
❌ Listen to event X → action updates X → triggers event X again → infinite loop
```

Everything else is a judgement call. The guidance below reflects sensible defaults for most churn intervention use cases, not hard restrictions.

### What This System Is Designed For

**Detection** — Identify churn risk signals, for example:
- Session decline patterns (weekly login count drops)
- Losing streaks and rage quits
- Other behavioral indicators of player disengagement

**Intervention** — Execute retention strategies, for example:
- Create time-limited comeback challenges
- Grant reward items via AccelByte entitlements
- Trigger notifications
- Record intervention history with cooldown management

### Sensible Default: Think in Terms of Ownership

A useful starting point when designing a plugin is to ask who **owns** a given piece of data. The table below reflects the default ownership model for a typical AGS game setup. Components you own, you can freely read and write. Components owned by other systems, you should generally only listen to — though writing is sometimes necessary (see the circular dependency rule below).

| Component | Owner | Default Role |
|-----------|-------|--------------|
| Game stats (`rse-match-wins`, `rse-current-losing-streak`, `rse-rage-quit`) | Game Server | **Listen** — react to stat events |
| Challenge progress tracking | Challenge System | **Listen** — react to completion events |
| Player sessions, login/logout | IAM Service | **Listen** — react to OAuth events |
| Churn detection logic | **Churn Intervention** | **Own** — implement rules |
| Intervention execution | **Churn Intervention** | **Own** — create challenges, grant rewards |
| Intervention history & cooldowns | **Churn Intervention** | **Own** — track what we did |
| Weekly login counts | **Churn Intervention** | **Own** — `session_tracking:*` Redis keys |

This table is a design aid, not an enforcement. Deviating is fine when you have a good reason (e.g., writing a stat specifically to trigger an Extend Challenge flow). The key question is always: *does writing this data create a circular event loop?*

### The One Rule: Avoid Circular Dependencies

The only firm guideline is to not create loops where your action feeds back into the same event your rule is listening to:

```yaml
# ❌ CIRCULAR: listens to rse-match-wins → action updates rse-match-wins
#    This triggers itself endlessly.
- id: record-wins
  type: match_win          # listens to rse-match-wins stat
  actions: [update-match-wins-stat]

# ✅ FINE: listens to rse-match-wins → dispatches a challenge
#    Writing to the challenge system does not re-emit rse-match-wins.
- id: losing-streak
  type: losing_streak
  actions: [dispatch-comeback-challenge]

# ✅ FINE: action writes a stat to trigger an Extend Challenge
#    Acceptable when the written stat is different from the listened stat.
- id: session-decline
  type: session_decline
  actions: [grant-item, send-email-notification-after-granting-item]
```

## Dependencies

### AccelByte Gaming Services (AGS)

This service runs on top of [AccelByte Gaming Services (AGS)](https://accelbyte.io). The following AGS features must be enabled in your namespace:

| AGS Feature | Used For |
|-------------|---------|
| **IAM** | OAuth event streaming — detecting player logins |
| **Statistics** | Stat update events — detecting losing streaks, rage quits, match wins |
| **Platform / Entitlements** | Granting reward items via the `grant-item` action |

### Extend Apps Required for Comeback Challenges

The `dispatch-comeback-challenge` action requires two additional Extend apps to be deployed in your AccelByte Admin Portal. Without them, the action will record an intervention but players won't be able to receive, see or complete a challenge.

| Extend App | Repository | Purpose |
|------------|-----------|---------|
| **Extend Challenge Service** | [Fork of extend-challenge-service](https://github.com/agriardyan/extend-challenge-service) | Stores challenge definitions, assign challenges, and player progress. Provides REST/gRPC APIs so players can query active challenges and claim rewards upon completion. |
| **Extend Challenge Event Handler** | [extend-challenge-event-handler](https://github.com/AccelByte/extend-challenge-event-handler) | Listens to real-time stat update events from AGS and automatically advances player progress toward challenge goals. Marks goals complete when targets are reached. |

**End-to-end flow with all three services:**

```
Churn Intervention          Extend Challenge Service    Extend Challenge Event Handler
        │                           │                               │
        │  1. detect churn signal   │                               │
        │  2. dispatch challenge ──►│                               │
        │     (create challenge)    │                               │
        │                           │                  3. player earns wins
        │                           │◄─── (stat events) ────────────┤
        │                           │     (updates progress)        │
        │                           │     (completes goal)          │
        │                           │                               │
        │                  4. player claims reward                   │
        │                     (entitlement granted via AGS)         │
```

## Getting Started

### Prerequisites

- Go 1.23+
- Docker (for Redis and proto generation)
- AccelByte namespace with IAM, Statistics, and Platform features enabled
- (For comeback challenges) [extend-challenge-service](https://github.com/AccelByte/extend-challenge-service) and [extend-challenge-event-handler](https://github.com/AccelByte/extend-challenge-event-handler) deployed as Extend apps

### Quick Start

```bash
# 1. Install dependencies
go mod download

# 2. Copy environment template
cp .env.template .env

# 3. Configure your AccelByte credentials in .env
# AB_BASE_URL, AB_CLIENT_ID, AB_CLIENT_SECRET, AB_NAMESPACE

# 4. Build
make build

# 5. Run tests
make test

# 6. Run locally (requires Redis)
docker run -d -p 6379:6379 redis:7-alpine
go run main.go
```

### Configuration

Edit `config/pipeline.yaml` to configure detection rules and interventions:

```yaml
rules:
  - id: losing-streak
    type: losing_streak
    enabled: true
    actions: [dispatch-comeback-challenge]
    parameters:
      threshold: 5

actions:
  - id: dispatch-comeback-challenge
    type: dispatch_comeback_challenge
    enabled: true
    parameters:
      wins_needed: 3
      duration_days: 7
      cooldown_hours: 168
```

Supports `${ENV_VAR:default}` substitution in parameter values.

## Built-in Rules

| Rule ID | Type | Signal | Description |
|---------|------|--------|-------------|
| `rage-quit` | `rage_quit` | `rage_quit` | Triggers when quit count reaches threshold (default: 3) |
| `losing-streak` | `losing_streak` | `losing_streak` | Triggers when consecutive losses reach threshold (default: 5) |
| `session-decline` | `session_decline` | `login` | Triggers when player had activity in a past week but none this week (4-week detection window) |

## Built-in Actions

| Action ID | Type | Description |
|-----------|------|-------------|
| `dispatch-comeback-challenge` | `dispatch_comeback_challenge` | Creates a time-limited comeback challenge (win N matches in X days). **Requires** [Fork of extend-challenge-service](https://github.com/agriardyan/extend-challenge-service) and [extend-challenge-event-handler](https://github.com/AccelByte/extend-challenge-event-handler) to be deployed. |
| `grant-item` | `grant_item` | Grants an item/entitlement via AccelByte platform (configurable via `REWARD_ITEM_ID` env var) |
| `send-email-notification-after-granting-item` | `send_email_notification_after_granting_item` | No-op stub — extend to send real email notifications |

## Extending the System

See **📄 [PLUGIN DEVELOPMENT](PLUGIN_DEVELOPMENT.md)** for a detailed developer guide with complete code templates.

For AI-assisted plugin creation, use the **`/add-plugin`** Claude Code skill which interactively generates all required files.

### Quick Example: Adding a New Rule

1. **Create the rule** in `pkg/rule/builtin/my_rule.go`:

```go
const MyRuleID = "my_rule"

type MyRule struct {
    config    rule.RuleConfig
    threshold int
}

func (r *MyRule) Evaluate(ctx context.Context, sig signal.Signal) (bool, *rule.Trigger, error) {
    if sig.Type() != signalBuiltin.TypeRageQuit {
        return false, nil, nil
    }
    // ... detection logic ...
    return true, rule.NewTrigger(r.ID(), sig.UserID(), "reason", r.config.Priority), nil
}
```

2. **Register it** in `pkg/rule/builtin/init.go`:

```go
func RegisterRules(deps *Dependencies) {
    rule.RegisterRuleType(MyRuleID, func(config rule.RuleConfig) (rule.Rule, error) {
        return NewMyRule(config), nil
    })
}
```

3. **Configure it** in `config/pipeline.yaml`:

```yaml
rules:
  - id: my-detection
    type: my_rule
    enabled: true
    actions: [grant-item]
    parameters:
      threshold: 3
```

## Project Structure

```
.
├── config/
│   └── pipeline.yaml              # Rules and actions configuration
├── internal/
│   ├── app/                       # Application setup and run logic
│   ├── bootstrap/                 # Service initialization (actions, rules, signals, pipeline)
│   ├── config/                    # Configuration loading and management
│   └── server/                    # gRPC server, metrics, and telemetry setup
├── pkg/
│   ├── action/                    # Intervention execution framework
│   │   ├── action.go              # Core Action interface
│   │   ├── executor.go            # Action execution logic with cooldown management
│   │   ├── factory.go             # Action factory for creating instances from config
│   │   ├── registry.go            # Action type registration
│   │   └── builtin/               # Built-in actions: grant_item, dispatch_comeback_challenge, send_email
│   ├── common/                    # Logging, env helpers, OpenTelemetry
│   ├── handler/                   # gRPC event handlers (OAuth, stat updates)
│   ├── pb/                        # Generated protobuf code for AccelByte events
│   ├── pipeline/                  # Pipeline orchestration and startup validation
│   ├── proto/                     # Protobuf definitions for AccelByte events
│   ├── rule/                      # Churn detection rule framework
│   │   ├── rule.go                # Core Rule interface
│   │   ├── engine.go              # Rule evaluation engine
│   │   ├── factory.go             # Rule factory for creating instances from config
│   │   ├── registry.go            # Rule type registration
│   │   └── builtin/               # Built-in rules: rage_quit, losing_streak, session_decline
│   ├── service/                   # Service abstractions and state models
│   │   ├── churn_state.go         # ChurnState, InterventionRecord, CooldownState
│   │   ├── login_session_tracker.go  # Weekly session tracking (Redis Hash)
│   │   ├── interfaces.go          # StateStore, LoginSessionTracker, EntitlementGranter
│   │   ├── platform.go            # AccelByte platform integration
│   │   └── models.go              # Data models and types
│   └── signal/                    # Event normalization framework
│       ├── signal.go              # Core Signal interface
│       ├── processor.go           # Signal processing logic
│       ├── event_processor.go     # EventProcessor interface for event-to-signal conversion
│       └── builtin/               # Built-in event processors and signals: OAuth, rage_quit, losing_streak
├── .claude/
│   └── skills/
│       └── add-plugin/            # /add-plugin Claude Code skill
├── main.go                        # Application entry point
├── CLAUDE.md                      # AI developer guide
├── PLUGIN_DEVELOPMENT.md          # Human developer guide for adding plugins
└── README.md                      # This file
```

## Testing

```bash
# Run all tests
make test

# Run specific package tests
go test -v ./pkg/rule/...
go test -v ./pkg/signal/...
go test -v ./pkg/action/...

# Run with coverage
go test -v -cover ./...

# Run integration tests
go test -v ./pkg/pipeline/...
```

## Deployment

```bash
# Build Docker image
make build

# Docker image: extend-churn-intervention:latest
```

The service requires the following environment variables (see `.env.template`):
- `AB_BASE_URL`, `AB_CLIENT_ID`, `AB_CLIENT_SECRET`, `AB_NAMESPACE` — AccelByte credentials
- `REDIS_HOST`, `REDIS_PORT`, `REDIS_PASSWORD` — Redis connection
- `REWARD_ITEM_ID` — Item ID to grant. See Store's Item at AccelByte AGS Admin Portal to find out the item ID.

## Monitoring

Prometheus metrics are available on port 8080:

```
http://localhost:8080/metrics
```

The gRPC server listens on port 6565.

## Contributing

When adding new rules or actions:

1. ✅ Ask: "Is this churn DETECTION or INTERVENTION?"
2. ✅ Check: "Am I trying to maintain state that another system owns?"
3. ✅ Verify: "Does this fit the read-react pattern?"
4. ✅ Read: `PLUGIN_DEVELOPMENT.md` for the full guide
5. ✅ Test: Write comprehensive unit tests
6. ✅ Configure: Update `config/pipeline.yaml`
