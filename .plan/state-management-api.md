# State Management API Reference

## Redis Operations

### GetChurnState
```go
func GetChurnState(ctx context.Context, client *redis.Client, userID string) (*ChurnState, error)
```
Retrieves player's churn state from Redis. Returns a new default state if player doesn't exist.

**Returns:**
- New player: Default state with zero values
- Existing player: Unmarshaled state from Redis

**Example:**
```go
state, err := state.GetChurnState(ctx, redisClient, "user-123")
if err != nil {
    log.Errorf("failed to get state: %v", err)
}
```

### UpdateChurnState
```go
func UpdateChurnState(ctx context.Context, client *redis.Client, userID string, state *ChurnState) error
```
Saves player's churn state to Redis with 30-day TTL.

**Example:**
```go
state.Sessions.ThisWeek++
err := state.UpdateChurnState(ctx, redisClient, "user-123", state)
```

### DeleteChurnState
```go
func DeleteChurnState(ctx context.Context, client *redis.Client, userID string) error
```
Removes player's churn state from Redis.

## Session Management

### CheckWeeklyReset
```go
func CheckWeeklyReset(state *ChurnState, now time.Time) bool
```
Checks if weekly reset should occur (7+ days since last reset). Performs reset if needed.

**Side Effects:**
- Moves `ThisWeek` → `LastWeek`
- Resets `ThisWeek` to 0
- Updates `LastReset` timestamp
- Cancels active challenges

**Returns:** `true` if reset occurred, `false` otherwise

**Example:**
```go
if state.CheckWeeklyReset(state, time.Now()) {
    log.Info("weekly reset performed")
}
```

## Churn Detection

### IsChurning
```go
func IsChurning(state *ChurnState, now time.Time) bool
```
Determines if player is exhibiting churn behavior.

**Churn Criteria:**
- Was active last week (`LastWeek > 0`)
- No activity this week (`ThisWeek == 0`)
- At least 7 days since last reset

**Example:**
```go
if state.IsChurning(state, time.Now()) {
    log.Info("player is churning")
}
```

### ShouldTriggerIntervention
```go
func ShouldTriggerIntervention(state *ChurnState, now time.Time) bool
```
Determines if intervention should be triggered.

**Conditions:**
- Player is churning
- No active challenge
- Not in cooldown period

**Example:**
```go
if state.ShouldTriggerIntervention(state, time.Now()) {
    // Trigger intervention
    state.CreateChallenge(state, 3, currentWins, expiresAt, "churn_detected")
}
```

## Intervention Management

### CanTriggerIntervention
```go
func CanTriggerIntervention(state *ChurnState, now time.Time) bool
```
Checks if intervention can be triggered based on cooldown.

**Returns:** `true` if allowed, `false` if in cooldown

### SetInterventionCooldown
```go
func SetInterventionCooldown(state *ChurnState, now time.Time, cooldownDuration time.Duration)
```
Sets cooldown period for next intervention.

**Side Effects:**
- Updates `LastTimestamp`
- Sets `CooldownUntil`
- Increments `TotalTriggered`

**Example:**
```go
state.SetInterventionCooldown(state, time.Now(), 48*time.Hour)
```

## Challenge Management

### CreateChallenge
```go
func CreateChallenge(state *ChurnState, winsNeeded int, currentWins int, expiresAt time.Time, reason string)
```
Creates a new comeback challenge for the player.

**Parameters:**
- `winsNeeded`: Number of wins required to complete challenge
- `currentWins`: Player's current total win count
- `expiresAt`: Challenge expiration time
- `reason`: Reason for triggering (e.g., "churn_detected")

**Example:**
```go
expiresAt := time.Now().Add(7 * 24 * time.Hour)
state.CreateChallenge(state, 3, playerWins, expiresAt, "churn_detected")
```

### UpdateChallengeProgress
```go
func UpdateChallengeProgress(state *ChurnState, newWinCount int, now time.Time) bool
```
Updates challenge progress with new win count.

**Returns:** `true` if challenge completed, `false` otherwise

**Side Effects:**
- Updates `WinsCurrent` (wins since challenge start)
- Deactivates challenge if completed or expired

**Example:**
```go
if state.UpdateChallengeProgress(state, newTotalWins, time.Now()) {
    log.Info("challenge completed!")
    // Grant rewards
}
```

## Data Models

### ChurnState
```go
type ChurnState struct {
    Sessions     SessionState      `json:"sessions"`
    Challenge    ChallengeState    `json:"challenge"`
    Intervention InterventionState `json:"intervention"`
}
```

### SessionState
```go
type SessionState struct {
    ThisWeek  int       `json:"thisWeek"`   // Sessions this week
    LastWeek  int       `json:"lastWeek"`   // Sessions last week
    LastReset time.Time `json:"lastReset"`  // Last weekly reset time
}
```

### ChallengeState
```go
type ChallengeState struct {
    Active        bool      `json:"active"`        // Challenge active?
    WinsNeeded    int       `json:"winsNeeded"`    // Wins required
    WinsCurrent   int       `json:"winsCurrent"`   // Current progress
    WinsAtStart   int       `json:"winsAtStart"`   // Wins when started
    ExpiresAt     time.Time `json:"expiresAt"`     // Expiration time
    TriggerReason string    `json:"triggerReason"` // Why triggered
}
```

### InterventionState
```go
type InterventionState struct {
    LastTimestamp  time.Time `json:"lastTimestamp"`  // Last intervention time
    CooldownUntil  time.Time `json:"cooldownUntil"`  // Cooldown end time
    TotalTriggered int       `json:"totalTriggered"` // Total interventions
}
```

## Constants

```go
const (
    DefaultTTL = 30 * 24 * time.Hour  // 30 days
    KeyPrefix  = "anti-churn:"         // Redis key prefix
)
```

## Complete Example Flow

```go
// Get player state
state, err := state.GetChurnState(ctx, redisClient, userID)
if err != nil {
    return err
}

// Check for weekly reset
state.CheckWeeklyReset(state, time.Now())

// Record session activity
state.Sessions.ThisWeek++

// Check if intervention should be triggered
if state.ShouldTriggerIntervention(state, time.Now()) {
    // Create challenge
    expiresAt := time.Now().Add(7 * 24 * time.Hour)
    state.CreateChallenge(state, 3, currentWins, expiresAt, "churn_detected")
    
    // Set cooldown
    state.SetInterventionCooldown(state, time.Now(), 48*time.Hour)
}

// Update challenge progress if active
if state.Challenge.Active {
    if state.UpdateChallengeProgress(state, newWinCount, time.Now()) {
        // Challenge completed - grant rewards in Phase 4
        log.Info("challenge completed!")
    }
}

// Save state
err = state.UpdateChurnState(ctx, redisClient, userID, state)
if err != nil {
    return err
}
```

## Testing

### Run Unit Tests
```bash
# All tests
go test ./pkg/state/...

# With coverage
go test ./pkg/state/... -cover

# Verbose output
go test -v ./pkg/state/...
```

### Run Integration Tests
```bash
# Requires Redis running on localhost:6379
go run -tags=integration test_redis_integration.go
```

## Error Handling

All functions return errors that should be checked:

```go
// Redis operations
state, err := state.GetChurnState(ctx, client, userID)
if err != nil {
    // Handle Redis connection or unmarshal errors
    log.Errorf("failed to get state: %v", err)
    return err
}

// Business logic functions don't return errors
// They use logrus for debug/info logging
resetOccurred := state.CheckWeeklyReset(state, time.Now())
```

## Logging Levels

- **Debug:** State operations, churn detection details
- **Info:** Challenge events, intervention triggers
- **Warning:** Invalid operations (e.g., updating inactive challenge)
- **Error:** Redis failures, unmarshal errors
