# Churn Intervention Event Handler

A plugin-based event handler for the AccelByte Extends platform that detects player churn signals and triggers automated interventions.

## Why is This?

**Player churn poses a significant threat to game success, particularly for newly launched titles.** When players become frustrated (e.g. from losing streaks, feeling stuck, or losing momentum) they may quit and never return. First impressions matter: early-stage player retention typically determines long-term success, with players who disengage in their first sessions being substantially less likely to return.

**Traditional analytics pipelines may respond too slowly for effective intervention.** Batch-processed analytics systems often operate on daily or longer cycles, creating delays between when a player experiences frustration and when that signal becomes actionable. This latency can mean the difference between successfully re-engaging a player and losing them permanently. Especially in moments of acute frustration like consecutive losses or other gameplay obstacles.

**This system enables near real-time detection and intervention.** By processing player events as they occur (logins, match results, stat updates), the system can identify potential churn signals within seconds and trigger automated interventions (e.g. granting rewards, creating comeback challenges, or offering personalized incentives), all while players are still engaged with the game, increasing the opportunity for successful re-engagement.

## What is This?

This service listens to game events (OAuth logins, stat updates) via Kafka Connect, analyzes player behavior patterns, and triggers interventions (challenges, rewards) to re-engage at-risk players.

**Example flows:**
- Player loses 5 matches in a row → `losing_streak` signal → "Comeback Challenge" (configured in Challenge Service, e.g. win 3 matches in 7 days)
- Player shows behavior of rage quit → `rage_quit` signal → "Comeback Challenge"
- Player's weekly logins drop to 0 (was active last week) → `session_decline` signal → Grant reward item + email notification

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

## 🎯 Architectural Boundaries

### ✅ What This System SHOULD Do

**Detection** — Identify churn risk signals:
- Session decline patterns (weekly login count drops)
- Losing streaks and rage quits
- Other behavioral indicators of player disengagement

**Intervention** — Execute retention strategies:
- Create time-limited comeback challenges
- Grant reward items via AccelByte entitlements
- Trigger email notifications
- Record intervention history with cooldown management

### ❌ What This System SHOULD NOT Do

This system is supposed to be **read-only** for game state. It REACTS to events, it does store its own historical data and churning state but NOT maintain primary game state.

An exception about the read-only may be made, e.g. when using extends-challenge-service, because extends-challenge-service needs stat update to trigger the challenge.

```yaml
# ❌ WRONG: Updating game stats (game server owns this)
- id: record-wins
  actions: [update-match-wins-stat]

# ❌ WRONG: Managing challenge progress (challenge system owns this)
- id: track-challenge
  actions: [increment-challenge-progress]
```

```yaml
# ✅ CORRECT: Detect churn and intervene
- id: losing-streak
  type: losing_streak
  actions: [dispatch-comeback-challenge]

# ✅ CORRECT: Session decline → reward + notify
- id: session-decline
  type: session_decline
  actions: [grant-item, send-email-notification-after-granting-item]
```

### Ownership Boundaries

| Component | Owner | Churn Intervention Role |
|-----------|-------|----------------|
| Game stats (`rse-match-wins`, `rse-current-losing-streak`, `rse-rage-quit`) | Game Server | **Read Only** - Listen to events |
| Challenge progress tracking | Challenge System | **Read Only** - Listen to completion events |
| Player sessions, login/logout | IAM Service | **Read Only** - Listen to OAuth events |
| Churn detection logic | **Churn Intervention** | **Owns** - Implement rules |
| Intervention execution | **Churn Intervention** | **Owns** - Create challenges, grant rewards |
| Intervention history & cooldowns | **Churn Intervention** | **Owns** - Track what we did |
| Weekly login counts | **Churn Intervention** | **Owns** - `session_tracking:*` Redis keys |

**Golden Rule**: If another system already owns it, we LISTEN to it, we don't UPDATE it.

## Getting Started

### Prerequisites

- Go 1.23+
- Docker (for Redis and proto generation)
- AccelByte account with namespace configured

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

### Session Decline Detection

The `session_decline` rule uses a map-based weekly tracking approach:

- On every login, a weekly count is incremented in Redis (`session_tracking:{userID}` Hash key)
- Weeks are keyed as `YYYYWW` (ISO week format, e.g., `"202610"`)
- Data is retained for 4 weeks — handles gradual churn and multi-week absences
- A player is considered churning if they had activity in any tracked week but have none in the current week

## Built-in Actions

| Action ID | Type | Description |
|-----------|------|-------------|
| `dispatch-comeback-challenge` | `dispatch_comeback_challenge` | Creates a time-limited challenge (win N matches in X days) |
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
├── pkg/
│   ├── action/                    # Intervention execution framework
│   │   └── builtin/               # grant_item, dispatch_comeback_challenge, send_email
│   ├── handler/                   # gRPC event handlers (OAuth, stat updates)
│   ├── pipeline/                  # Pipeline orchestration and startup validation
│   ├── rule/                      # Churn detection rule framework
│   │   └── builtin/               # rage_quit, losing_streak, session_decline
│   ├── service/                   # Service abstractions and state models
│   │   ├── churn_state.go         # ChurnState, InterventionRecord, CooldownState
│   │   ├── login_session_tracker.go  # Weekly session tracking (Redis Hash)
│   │   └── interfaces.go          # StateStore, LoginSessionTracker, EntitlementGranter
│   ├── signal/                    # Event normalization framework
│   │   └── builtin/               # OAuth, rage_quit, losing_streak event processors
│   └── common/                    # Logging, env helpers, OpenTelemetry
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
- `REWARD_ITEM_ID` — Item ID to grant (default: `COMEBACK_REWARD`)

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
