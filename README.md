# Anti-Churn Event Handler

A plugin-based event handler for the AccelByte Extends platform that detects player churn signals and triggers automated interventions.

## What is This?

This service listens to game events (OAuth logins, stat updates) via Kafka Connect, analyzes player behavior patterns, and triggers interventions (challenges, rewards) to re-engage at-risk players.

**Example flow:**
1. Player loses 5 matches in a row → `losing_streak` signal detected
2. `losing-streak` rule triggers → Creates "Comeback Challenge" (win 3 matches in 7 days)
3. Player completes challenge → `challenge-completion` rule triggers → Grants reward item

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
- **Pipeline**: Orchestrates the full flow with configurable rule-to-action mappings

## 🎯 Architectural Boundaries

### ✅ What This System SHOULD Do

**Detection** - Identify churn risk signals:
- Session decline patterns
- Losing streaks and rage quits
- Challenge completion/failure
- Other behavioral indicators of player disengagement

**Intervention** - Execute retention strategies:
- Create time-limited challenges
- Grant rewards/items
- Trigger notifications (future)
- Record intervention history

### ❌ What This System SHOULD NOT Do

**❌ DO NOT use this as a general game state manager**

This system is **read-only** for game state. It REACTS to events, it does NOT maintain primary game state.

**Examples of WRONG usage:**

```yaml
# ❌ WRONG: Updating game stats
- id: record-wins
  actions: [update-match-wins-stat]  # Game server owns this!

# ❌ WRONG: Managing challenge progress
- id: track-challenge
  actions: [increment-challenge-progress]  # Challenge system owns this!

# ❌ WRONG: Resetting player stats
- id: reset-streak
  actions: [reset-losing-streak-stat]  # Game server owns this!
```

**Examples of CORRECT usage:**

```yaml
# ✅ CORRECT: Detecting churn and intervening
- id: losing-streak
  type: losing_streak
  actions: [comeback-challenge]  # Create intervention

# ✅ CORRECT: Rewarding completion
- id: challenge-completion
  type: challenge_completion
  actions: [grant-item]  # Reward player
```

### Ownership Boundaries

| Component | Owner | Anti-Churn Role |
|-----------|-------|----------------|
| Game stats (`rse-match-wins`, `rse-current-losing-streak`) | Game Server | **Read Only** - Listen to events |
| Challenge progress tracking | Challenge System | **Read Only** - Listen to completion events |
| Player sessions, login/logout | IAM Service | **Read Only** - Listen to OAuth events |
| Churn detection logic | **Anti-Churn** | **Owns** - Implement rules |
| Intervention execution | **Anti-Churn** | **Owns** - Create challenges, grant rewards |
| Intervention history | **Anti-Churn** | **Owns** - Track what we did |

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
  - id: my-churn-rule
    type: my_rule_type
    enabled: true
    actions: [my-action]
    parameters:
      threshold: 5

actions:
  - id: my-action
    type: my_action_type
    enabled: true
    parameters:
      duration_days: 7
```

## Extending the System

### Adding a New Churn Detection Rule

**Example**: Detect players who haven't logged in for 7 days

1. **Create the rule** in `pkg/rule/builtin/`:

```go
type InactivityRule struct {
    config rule.RuleConfig
}

func (r *InactivityRule) Evaluate(ctx context.Context, sig signal.Signal) (bool, *rule.Trigger, error) {
    // Only evaluate login signals
    if sig.Type() != "login" {
        return false, nil, nil
    }

    // Check if player hasn't logged in for 7+ days
    lastLogin := sig.Context().State.Sessions.LastLogin
    if time.Since(lastLogin) > 7*24*time.Hour {
        return true, &rule.Trigger{
            RuleID: r.config.ID,
            UserID: sig.UserID(),
            Reason: "Player inactive for 7+ days",
        }, nil
    }

    return false, nil, nil
}
```

2. **Register it** in `pkg/rule/builtin/init.go`:

```go
func RegisterRules() {
    rule.RegisterRuleType("inactivity", func(config rule.RuleConfig) (rule.Rule, error) {
        return NewInactivityRule(config), nil
    })
}
```

3. **Configure it** in `config/pipeline.yaml`:

```yaml
rules:
  - id: player-inactivity
    type: inactivity
    enabled: true
    actions: [comeback-challenge]
```

### Adding a New Intervention Action

**Example**: Send a push notification

1. **Create the action** in `pkg/action/builtin/`:

```go
type PushNotificationAction struct {
    config action.ActionConfig
    notificationService NotificationService
}

func (a *PushNotificationAction) Execute(ctx context.Context, trigger *rule.Trigger, playerCtx *signal.PlayerContext) error {
    message := "We miss you! Come back and claim your reward!"
    return a.notificationService.Send(ctx, trigger.UserID, message)
}
```

2. **Register and configure** (similar to rules)

### Adding a New Signal Type

See `CLAUDE.md` for detailed instructions on adding new event processors.

## Project Structure

```
.
├── cmd/                    # Entry points
├── config/                 # Configuration files
│   └── pipeline.yaml      # Rules and actions configuration
├── pkg/
│   ├── action/            # Intervention execution
│   │   └── builtin/       # Built-in actions
│   ├── handler/           # gRPC event handlers
│   ├── pipeline/          # Pipeline orchestration
│   ├── rule/              # Churn detection rules
│   │   └── builtin/       # Built-in rules
│   ├── signal/            # Event normalization
│   │   └── builtin/       # Built-in signals
│   ├── state/             # Player state management
│   └── service/           # External service clients
├── main.go                # Application entry point
├── CLAUDE.md              # Developer guide for AI assistance
└── README.md              # This file
```

## Testing

```bash
# Run all tests
make test

# Run specific package tests
go test -v ./pkg/rule/...

# Run with coverage
go test -v -cover ./...

# Run integration tests
go test -v ./pkg/pipeline/...
```

## Built-in Rules

- **rage-quit**: Detects players who quit after consecutive losses
- **losing-streak**: Detects extended losing streaks
- **session-decline**: Detects significant drop in play sessions
- **challenge-completion**: Grants rewards when challenges are completed

## Built-in Actions

- **comeback-challenge**: Creates a time-limited challenge (win N matches in X days)
- **grant-item**: Grants an item/entitlement to the player

## Common Pitfalls

### ❌ Don't Mix Concerns

```yaml
# WRONG: Anti-churn updating game stats
- id: update-stats
  type: stat_updater
  actions: [increment-wins]  # Game server should do this!
```

### ❌ Don't Create Circular Dependencies

```yaml
# WRONG: Listening to stat event and updating same stat
rules:
  - id: record-wins
    # Listens to rse-match-wins event
    actions: [update-match-wins]  # Updates rse-match-wins (circular!)
```

### ❌ Don't Replace Existing Systems

```yaml
# WRONG: Challenge progress tracking should be in challenge system
- id: track-challenge-progress
  type: challenge_tracker
  actions: [increment-progress]  # Not our responsibility!
```

### ✅ Do Focus on Detection + Intervention

```yaml
# CORRECT: Detect churn risk and trigger intervention
- id: at-risk-player
  type: losing_streak
  actions: [comeback-challenge, send-notification]  # Our job!
```

## Deployment

See `Dockerfile` and deployment documentation for production setup.

## Monitoring

The service exposes Prometheus metrics on port 8080:

```
http://localhost:8080/metrics
```

## Contributing

When adding new rules or actions:

1. ✅ Ask: "Is this churn DETECTION or INTERVENTION?"
2. ✅ Check: "Am I trying to maintain state that another system owns?"
3. ✅ Verify: "Does this fit the read-react pattern?"
4. ✅ Test: Write comprehensive unit tests
5. ✅ Document: Update this README and CLAUDE.md

## License

[Add your license here]

## Support

For issues, questions, or contributions, please [add contact info or issue tracker link].
