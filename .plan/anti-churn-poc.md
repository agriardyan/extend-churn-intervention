# Anti-Churn Reward System
## Proof of Concept (PoC)

**Version:** 1.0  
**Purpose:** Validate feasibility and effectiveness of automated churn detection and intervention  
**Scope:** Single FPS game, simplified implementation  
**Duration:** 6 weeks

---

## Executive Summary

### Objective

Prove that an automated anti-churn system using AccelByte Extend can:
1. **Detect** at-risk players accurately using behavioral signals
2. **Intervene** automatically with targeted rewards
3. **Function correctly** end-to-end with AccelByte services

**Note:** This PoC validates that the system **works correctly**. Actual retention improvement testing will be a separate milestone with real players in a production game.

### Why PoC First?

Before investing in a full pluggable platform, validate technical feasibility:
- ✅ Can we accurately detect at-risk players?
- ✅ Does the integration with AccelByte services work smoothly?
- ✅ Is AccelByte Extend suitable for this use case?
- ✅ Can we build this efficiently with Go?
- ✅ Does the end-to-end flow work correctly?

**Retention improvement testing comes later** with real players in a production game.

### PoC Approach: Custom CLI Game

To have complete control over testing conditions, we'll build a simple CLI single-player game:

**"Ready, Set, Enter!"** - A terminal-based game where the player races against a bot to press Enter 100 times first.

**Why Build a Custom Game?**
- ✅ **Complete control** over game mechanics and testing scenarios
- ✅ **Super fast development** - 1 week for CLI vs 4+ weeks for web game
- ✅ **Perfect test bed** - Generate exact churn scenarios with bots
- ✅ **Easy automation** - Spawn 100s of bot players for data generation
- ✅ **Clear signals** - Simple win/loss mechanics, obvious rage quits
- ✅ **Rapid iteration** - Change mechanics in minutes, test immediately
- ✅ **No networking complexity** - Single-player means zero network issues
- ✅ **Single binary distribution** - No installation needed (Go binary)

**Single-Player vs Bot Design:**
- No matchmaking delays - instant matches!
- No session management complexity
- Bot with dynamic, natural speed (realistic challenge)
- Single balanced difficulty (~45-55% win rate)
- Perfect for testing all churn signals
- **Go binary = just download and run** (no dependencies!)

**See:** [Ready, Set, Enter! - Game Design Document](ready-set-enter-game-design.md)

### Scope Limitations (Intentional)

**What's INCLUDED:**
- Custom CLI game: "Ready, Set, Enter!" (single-player vs bot)
- Bot with balanced difficulty and natural speed variance
- 3 churn signals: Rage quits + Session frequency decline + Losing streak
- 1 reward type: Conditional rewards ("Win 3 matches, get reward")
- Automated intervention system (no manual triggers)
- Bot system for generating test data
- Full Go tech stack (game + Extend)
- Redis for Extend state management
- 6-7 weeks timeline (includes game development)

**What's EXCLUDED (For Later):**
- A/B testing (this PoC validates functionality, not retention improvement)
- Real player testing (use bots for PoC)
- Multiplayer (not needed for PoC - adds weeks of complexity)
- Multi-game support (not needed for PoC)
- Pluggable architecture (adds complexity)
- Multiple reward types (keep it simple)
- Advanced player segmentation (just "all players")
- Machine learning (rules-based only)
- Production-grade monitoring (basic dashboards only)

### Expected Outcome

**Success Criteria:**
- ✅ System correctly detects 80%+ of simulated churn scenarios
- ✅ Challenges are created and tracked accurately
- ✅ Rewards are granted correctly upon challenge completion
- ✅ Zero critical bugs during testing
- ✅ Extend Event Handler performs reliably (99%+ uptime)
- ✅ Clear decision: Ready for real player testing? Yes/No

**Timeline:** 6-7 weeks
- Week 1: Build CLI game (single-player vs bot)
- Week 2-3: Build anti-churn Extend app (Go)
- Week 4: Integration testing
- Week 5-6: Generate test data with bots
- Week 7: Validation and decision

**Next Milestone (Separate):** Retention improvement testing with real players in production game

---

## Table of Contents

1. [Game Selection](#game-selection)
2. [Simplified Architecture](#simplified-architecture)
3. [Churn Signals (2 Only)](#churn-signals)
4. [Reward Strategy](#reward-strategy)
5. [Implementation Plan](#implementation-plan)
6. [Testing Approach](#testing-approach)
7. [Success Metrics](#success-metrics)
8. [Go/No-Go Decision](#go-no-go-decision)

---

## Game Selection

### Custom CLI Game: "Ready, Set, Enter!"

**Game Type:** Terminal-based single-player vs bot  
**Complete Design:** See [Ready, Set, Enter! - Game Design Document](ready-set-enter-game-design.md)

#### Why Build a Custom Game?

Instead of using an existing game, we're building a simple CLI single-player game specifically for this PoC:

**Advantages:**

1. **Complete Control Over Testing**
   - Design exact churn scenarios we want to test
   - Generate specific player behaviors with bots
   - Modify game mechanics instantly for testing

2. **Super Fast Development (1 week)**
   - No UI/graphics needed (terminal only)
   - No multiplayer networking (single-player vs bot)
   - Simple mechanics (press Enter to win)
   - Focus on AccelByte integration, not game complexity

3. **Perfect Churn Signals**
   - Clear win/loss outcomes
   - Easy rage quit detection (quit mid-match while losing to bot)
   - Simple session tracking (terminal launches)
   - Obvious performance decline (losing streaks against bot)

4. **Easy Data Generation**
   - Spawn 100s of bot players
   - Simulate exact behaviors (rage quitters, declining sessions, losing streaks)
   - Generate weeks of data in hours
   - Full control over test conditions

5. **Zero Dependencies**
   - No need to negotiate access to existing game
   - No existing player base to worry about
   - No networking issues or latency
   - Can test aggressively without business risk
   - Complete data access

#### Game Overview

**Core Gameplay:**
```
Player launches CLI
↓
Clicks "Start Match" (instant!)
↓
Race to press Enter 100 times first
↓
Match duration: 30-60 seconds
↓
Winner determined, stats updated
↓
Repeat or quit
```

**What Makes It Perfect for Testing:**

| Aspect | How It Helps Testing |
|--------|---------------------|
| **Fast Matches** | 30-60 seconds = quick feedback loop |
| **Clear Winners** | No ambiguity in performance tracking |
| **Simple Mechanics** | Easy to generate with bots |
| **Terminal-Based** | Easy to automate testing |
| **Single-Player** | No network issues, instant matches |
| **Dynamic Bot** | Natural speed creates realistic challenge |

#### Bot Opponent Design

**Single Balanced Difficulty:**

- **Bot Speed:** ~8.3 presses/sec (balanced challenge)
- **Expected Win Rate:** 45-55% for average players
- **Natural Behavior:** Variance + bursts + gradual fatigue

**Natural Speed Variance:**
- Random variance (±30ms)
- Occasional bursts of speed (human focus moments)
- Gradual fatigue (slows down slightly over match)
- Creates realistic, competitive matches

**Why This Works:**
- Creates enough losses to trigger churn signals
- Not so hard that players give up immediately
- Natural variance makes matches feel unpredictable
- Simple to implement and test

#### Development Timeline

**Week 1: Build the Game**
- Days 1-2: Setup + AccelByte authentication + core mechanics
- Days 3-4: Bot implementation with natural speed
- Days 5-6: Statistics integration + reward system
- Day 7: Testing + polish

**Total: 7 working days (1 week)**

**Deliverables:**
- Playable CLI game vs bot (single balanced difficulty)
- Full AccelByte integration
- Bot system for testing
- Ready for anti-churn system integration
- **Single binary** for Windows/Mac/Linux (no installation!)

**Technology:**
- Go with Bubble Tea (terminal UI)
- AccelByte Go SDK
- Single binary distribution (~10-15MB)
- Cross-compile from one machine

---

## Simplified Architecture

### No Pluggable Design (PoC Only)

```
┌──────────────────────────────────────────────────────────┐
│         READY, SET, ENTER! (CLI SINGLE-PLAYER)           │
│                                                          │
│  Terminal UI → Keyboard Input → Match Logic → Bot AI    │
│                       ↓                                   │
│              AccelByte Go SDK                            │
└────────────────────────┬─────────────────────────────────┘
                         │
                         ↓
          ┌──────────────────────────────┐
          │  AccelByte Gaming Services   │
          ├──────────────────────────────┤
          │ • IAM (Login)                │
          │ • Statistics Service ← EVENTS│
          │ • Entitlement (Boosters)     │
          └──────────────┬───────────────┘
                         │
                         │ Events Published
                         ↓
          ┌──────────────────────────────┐
          │  ANTI-CHURN SYSTEM           │
          │  (AccelByte Extend - Go)     │
          │                              │
          │  Listens to:                 │
          │  • statItemUpdated           │
          │  • oauthTokenGenerated       │
          │  • statisticCycleReset       │
          │                              │
          │  Hard-coded Detection:       │
          │  1. Rage Quit Detector       │
          │     IF quit while losing     │
          │     → rage_quit_count++      │
          │                              │
          │  2. Session Tracker          │
          │     IF logins decline 50%    │
          │     → at_risk                │
          │                              │
          │  3. Losing Streak Tracker    │
          │     IF losses >= 5 in row    │
          │     → at_risk                │
          │                              │
          │  4. Direct Intervention      │
          │     IF any signal triggers   │
          │     → Grant via Entitlement  │
          │     → Track in Redis         │
          └──────────────┬───────────────┘
                         │
                 ┌───────┴────────┐
                 │                │
                 ↓                ↓
    ┌────────────────┐  ┌─────────────────┐
    │ AGS Entitlement│  │ Redis (Managed) │
    │                │  │                 │
    │ Grant boosters │  │ Extend State:   │
    │ conditionally  │  │ • Sessions      │
    │                │  │ • Challenges    │
    └────────────────┘  │ • Cooldowns     │
                        └─────────────────┘
```

### Key Simplifications

1. **No Configuration Files:** All logic hard-coded in Extend app
2. **No Pluggable Signals:** Just 3 detectors built into the code
3. **Managed Redis:** AccelByte provides Redis for Extend state (zero infrastructure setup)
4. **No Complex Scoring:** Simple IF/THEN rules
5. **Single Reward Type:** Only conditional rewards (win 3 matches)
6. **Minimal Error Handling:** Focus on happy path
7. **Single-Player Game:** No matchmaking, session, or networking
8. **Bot Opponent:** Instant matches, zero network dependency
9. **No A/B Testing:** Focus on functional correctness, not retention improvement

**Why This Is OK for PoC:**
- Proves the system works functionally (detection + intervention + rewards)
- Easier to debug (less abstraction, simpler scope)
- Still demonstrates core integration (Extend + AGS + Redis)
- Faster to build and iterate (6-7 weeks total)
- Can measure detection accuracy and system reliability
- Complete control over test environment (no dependencies on external games)
- No networking complexity to debug (single-player CLI)
- Retention improvement testing comes later with real game and real players

---

## Churn Signals

### Signal 1: Rage Quit Detection

**Definition:** Player quits mid-match (disconnects or forfeits) while losing badly

**Why This Signal?**
- Strong predictor of frustration-driven churn
- Easy to detect in competitive games
- Clear in Ready, Set, Enter!: player behind by 30+ presses and quits

**Implementation:**
```
Listen to:
  - GameSessionEndedEvent
  - statItemUpdated for match progress

Logic:
  1. During match:
     - Track both players' press counts in real-time
     - Calculate deficit: opponent_count - player_count
  
  2. When player disconnects mid-match:
     - Check if match was still in progress (neither reached 100)
     - Check if player was losing: deficit >= 30 presses
     
     IF match_not_finished
        AND player_deficit >= 30
     THEN:
        rage_quit_count_this_week++
     
  3. Weekly check (in Extend Event Handler):
     IF rse-rage-quit >= 3  (reads from AGS Statistics)
     THEN:
        Trigger intervention
```

**Example Scenario:**
```
Match Progress:
  Player:   [████████░░░░░░░░░░] 42/100
  Opponent: [████████████████░░] 78/100
  
Player presses ESC (forfeit) or closes terminal
  → Detected as rage quit (deficit = 36 presses)
  → Game increments: rse-rage-quit
  → AGS Weekly Cycle tracks count (auto-resets Monday)
```

**Data Storage (AGS Statistics):**
- `rse-rage-quit` - Base stat code, added to Weekly Statistic Cycle (auto-resets Monday)

**Note:** The stat code name is `rse-rage-quit`. When added to a Weekly Statistic Cycle in AGS Admin Portal, it tracks "rage quits this week" with auto-reset.

**Why AGS Statistics (not Redis)?**
- Admin dashboard visibility - stakeholders can see "rage quits this week" in real-time
- Built-in monitoring and alerts
- Essential for PoC demonstration (show that detection is working)
- Statistic Cycles handle weekly reset automatically

**Threshold:** 3 rage quits in a 7-day period = at-risk

---

### Signal 2: Session Frequency Decline

**Definition:** Player's login frequency (terminal launches) drops by 50%+ compared to previous week

**Why This Signal?**
- Universal churn indicator across all game types
- Easy to calculate from login events
- Reliable early warning signal

**Implementation:**
```
Listen to:
  - oauthTokenGenerated (login event when CLI starts)
  - statisticCycleReset (when weekly cycle resets)

Logic:
  1. On each login:
     - Game increments stat: rse-session-login
     - AGS Statistic Cycle tracks count
  
  2. When Statistic Cycle resets (Monday 00:00):
     - Extend Event Handler listens to statisticCycleReset event
     - Extend captures: valueBeforeReset from event
     - Extend stores in Redis: sessions.lastWeek = valueBeforeReset
     - AGS automatically resets rse-session-login to 0
  
  3. Daily check (if player has >= 7 days history):
     - Extend reads from Redis: sessions.lastWeek
     - Extend reads from AGS: current value of rse-session-login
     - Calculate: decline_ratio = current / lastWeek
     
     IF decline_ratio < 0.5
     THEN:
        Trigger intervention
```

**Example Scenario:**
```
Week 1 (baseline):
  Mon: Launch CLI → Game increments rse-session-login (1)
  Tue: Launch CLI → Game increments rse-session-login (2)
  Wed: Launch CLI → Game increments rse-session-login (3)
  Thu: Launch CLI → Game increments rse-session-login (4)
  Fri: Launch CLI → Game increments rse-session-login (5)
  Sat: Launch CLI → Game increments rse-session-login (6)
  Total: rse-session-login = 6
  
  Monday 00:00: AGS Statistic Cycle resets
    → Extend captures: valueBeforeReset = 6
    → Extend stores in Redis: sessions.lastWeek = 6
    → AGS resets: rse-session-login = 0

Week 2 (declining):
  Mon: Launch CLI → Game increments rse-session-login (1)
  Tue: Launch CLI → Game increments rse-session-login (2)
  Wed: [no launch]
  Thu: [no launch]
  Fri: [no launch]
  Sat: [no launch]
  Total: rse-session-login = 2
  
  Extend checks: 2 / 6 = 33% (67% decline) → Trigger intervention
```

**Data Storage Strategy (Hybrid Approach):**

**Why Hybrid Storage?**

We use BOTH AGS Statistics and Redis for different purposes:

**AGS Statistics = Visibility & Monitoring**
- Admin dashboard can see real-time churn metrics
- Stakeholders can monitor system without custom tooling
- Built-in charts, trends, and alerts
- Essential for PoC demonstration and debugging

**Redis = Flexible Backend Logic**
- Extend-computed state (previous week, challenges, cooldowns)
- Complex JSON structures not possible in Statistics
- Fast read/write for backend logic
- Invisible to admins (that's fine - it's internal state)

**If we put everything in Redis:** Admins lose visibility into churn signals. They'd need custom dashboards just to see "how many users rage quit this week" - adds complexity to PoC.

**If we put everything in AGS Statistics:** We'd need workarounds for derived state (previous week) and lose flexibility for complex challenge tracking.

**Hybrid = Best of both worlds** for this PoC.

---

**AGS Statistics (Churn Aggregations - for Admin Visibility):**
- `rse-session-login` - Base stat code, incremented by game on login, added to Weekly Statistic Cycle (auto-resets Monday)
- `rse-rage-quit` - Base stat code, incremented by game on rage quit, added to Weekly Statistic Cycle (auto-resets Monday)

**Note:** These are the base stat code names. The Statistic Cycle provides the weekly aggregation and auto-reset behavior.

**Redis (Extend-Managed State - for Backend Logic):**
- Session tracking: `thisWeek`, `lastWeek`, `lastReset` (derived from AGS events)
- Challenge state: `active`, `winsNeeded`, `winsCurrent`, `expiresAt`
- Intervention cooldowns: `lastTimestamp`, `cooldownUntil`

**How Session Tracking Works with Redis:**

```go
// Extend Event Handler (Go) - on login event
func onLogin(ctx context.Context, event LoginEvent) error {
    // Read state from Redis
    var state ChurnState
    err := redis.HGet(ctx, fmt.Sprintf("churn_state:%s", event.UserID), "state").Scan(&state)
    if err == redis.Nil {
        state = getDefaultState()
    }
    
    // Check if weekly reset needed
    if shouldResetWeekly(state.Sessions.LastReset) {
        state.Sessions.LastWeek = state.Sessions.ThisWeek
        state.Sessions.ThisWeek = 0
        state.Sessions.LastReset = getLastMonday()
    }
    
    // Increment current week
    state.Sessions.ThisWeek++
    
    // Save back to Redis
    stateJSON, _ := json.Marshal(state)
    return redis.HSet(ctx, fmt.Sprintf("churn_state:%s", event.UserID), "state", stateJSON).Err()
}

type ChurnState struct {
    Sessions SessionState `json:"sessions"`
    Challenge ChallengeState `json:"challenge"`
    Intervention InterventionState `json:"intervention"`
}

type SessionState struct {
    ThisWeek   int       `json:"thisWeek"`
    LastWeek   int       `json:"lastWeek"`
    LastReset  time.Time `json:"lastReset"`
}
```

**Benefits of Redis:**
- ✅ No Statistic Cycles workaround needed
- ✅ Natural weekly reset logic in code
- ✅ Previous week data preserved automatically
- ✅ Flexible JSON structure
- ✅ Managed by AccelByte (zero infrastructure setup)

**Threshold:** 50% decline (e.g., 6 logins last week → 3 this week)

---

### Signal 3: Losing Streak (Bonus Detection)

**Definition:** Player loses 5+ matches in a row

**Why This Signal?**
- Indicates declining performance or skill mismatch
- Often precedes rage quits
- Easy to track in competitive games

**Implementation:**
```
Listen to:
  - statItemUpdated for match_wins and match_losses

Logic:
  1. After each match:
     IF player_won:
        current_losing_streak = 0  # Reset streak
     ELSE:
        current_losing_streak++
  
  2. Check threshold:
     IF current_losing_streak >= 5
     THEN:
        Trigger intervention
```

**Example Scenario:**
```
Match History:
  Match 1: LOSS (45 vs 100 presses) → Streak: 1
  Match 2: LOSS (67 vs 100 presses) → Streak: 2
  Match 3: LOSS (52 vs 100 presses) → Streak: 3
  Match 4: LOSS (71 vs 100 presses) → Streak: 4
  Match 5: LOSS (63 vs 100 presses) → Streak: 5 ⚠️ TRIGGER
```

**Data Storage (AGS Statistics):**
- `rse-current-losing-streak` - Reset on win
- `rse-match-wins` - Total wins (also used for Daily Cycle)
- `rse-match-losses` - Total losses
- `rse-match-played` - Total matches (added to Daily Statistic Cycle for daily tracking)

**Threshold:** 5 consecutive losses = at-risk

---

### Combined Logic (Simple)

**Trigger intervention IF ANY of:**
- Rage quit count >= 3 this week **OR**
- Session frequency decline > 50% **OR**
- Losing streak >= 5 consecutive losses

**Cooldown Rules:**
- Max 1 intervention per player per week
- Min 48 hours between interventions
- Check cooldown before triggering any signal

**Implementation (Go):**
```go
func checkForIntervention(ctx context.Context, userID string) (bool, error) {
    // Read state from Redis
    var state ChurnState
    err := redis.HGet(ctx, fmt.Sprintf("churn_state:%s", userID), "state").Scan(&state)
    if err != nil && err != redis.Nil {
        return false, err
    }
    
    // Check cooldown first (48 hours minimum)
    if time.Now().Before(state.Intervention.CooldownUntil) {
        return false, nil // Too soon
    }
    
    // Check weekly limit
    if state.Intervention.CountThisWeek >= 1 {
        return false, nil // Already got one this week
    }
    
    // Check signals from AGS Statistics
    rageQuits, _ := getStatistic(ctx, userID, "rse-rage-quit")
    losingStreak, _ := getStatistic(ctx, userID, "rse-current-losing-streak")
    
    // Check session decline from Redis
    sessionDecline := float64(state.Sessions.ThisWeek) / float64(state.Sessions.LastWeek)
    
    if rageQuits >= 3 || sessionDecline < 0.5 || losingStreak >= 5 {
        err := triggerIntervention(ctx, userID, "churn_detected")
        return true, err
    }
    
    return false, nil
}
```

---

## Reward Strategy

### Single Reward Type: Conditional Challenge

**Reward Configuration:**

```yaml
Intervention: "Comeback Challenge"

Trigger Condition:
  - Rage quit count >= 3 OR
  - Session decline > 50% OR
  - Losing streak >= 5

Challenge:
  - Action: "Win 3 matches (not just play, must win)"
  - Time Limit: 72 hours (3 days)
  - Progress Tracking: Via statItemUpdated "match_wins"

Rewards on Completion:
  - 1000 game coins
  - 1x Press Speed Booster (5 matches)
    Effect: Each press counts as 1.5 presses
  - "Comeback King" badge (displayed in profile)
  - Special terminal message: "You're back! 🎉"

Player Experience:
  1. After match, notification appears in terminal:
     "🎁 SPECIAL COMEBACK CHALLENGE! 🎁"
  2. Challenge description with countdown timer
  3. Progress tracked in main menu: [█░░] 1/3 wins
  4. Upon 3rd win: "CHALLENGE COMPLETE! Rewards granted!"
  5. Booster automatically applied to next 5 matches
```

**Why This Reward Type?**

1. **Encourages Re-engagement:** Player must continue playing to get reward
2. **No Exploitation:** Must actually WIN, not just participate (higher bar)
3. **Skill Validation:** Helps player prove to themselves they can win
4. **Achievable:** 3 wins in 72 hours = ~1 win per day (reasonable)
5. **Valuable:** Booster directly helps improve performance

**Why Require WINS (Not Just Participation)?**

In Ready, Set, Enter!, requiring wins is better than just participation:
- Player proves they're improving (not just grinding losses)
- More meaningful achievement = more satisfying
- Booster reward makes sense (helps maintain winning)
- Still achievable: Most players have >40% win rate

**Extend Event Handler (Go) - Direct Entitlement Grant:**

Instead of using AGS Reward Service, the Extend Event Handler directly grants entitlements:

```go
// In Extend Event Handler (Go)
func triggerIntervention(ctx context.Context, userID string, reason string) error {
    // 1. Create challenge state in Redis
    challenge := ChallengeState{
        Active:     true,
        WinsNeeded: 3,
        WinsCurrent: 0,
        ExpiresAt:  time.Now().Add(72 * time.Hour),
        TriggeredBy: reason,
    }
    
    intervention := InterventionState{
        LastTimestamp: time.Now(),
        CooldownUntil: time.Now().Add(7 * 24 * time.Hour),
        CountThisWeek: 1,
    }
    
    state := ChurnState{
        Challenge: challenge,
        Intervention: intervention,
    }
    
    // Save to Redis
    stateJSON, _ := json.Marshal(state)
    err := redis.HSet(ctx, fmt.Sprintf("churn_state:%s", userID), "state", stateJSON).Err()
    if err != nil {
        return err
    }
    
    // 2. Grant conditional booster via Entitlement API
    // Player gets it after completing challenge
    err = entitlementClient.GrantUserEntitlement(ctx, &platform.GrantUserEntitlementRequest{
        Namespace: namespace,
        UserID: userID,
        ItemID: "speed_booster_conditional",
        Quantity: 1,
        Source: "CHURN_INTERVENTION",
        Metadata: map[string]interface{}{
            "requires_wins": 3,
            "expires_at": time.Now().Add(72 * time.Hour).Unix(),
            "granted": false, // Not active yet
        },
    })
    
    return err
}

// Listen for match wins
func onMatchWin(ctx context.Context, userID string) error {
    // Read state from Redis
    var state ChurnState
    err := redis.HGet(ctx, fmt.Sprintf("churn_state:%s", userID), "state").Scan(&state)
    if err != nil {
        return err
    }
    
    if !state.Challenge.Active {
        return nil // No active challenge
    }
    
    // Check expiry
    if time.Now().After(state.Challenge.ExpiresAt) {
        state.Challenge.Active = false
        stateJSON, _ := json.Marshal(state)
        redis.HSet(ctx, fmt.Sprintf("churn_state:%s", userID), "state", stateJSON)
        return nil // Challenge expired
    }
    
    // Increment progress
    state.Challenge.WinsCurrent++
    
    // Check completion
    if state.Challenge.WinsCurrent >= state.Challenge.WinsNeeded {
        // Grant the booster
        err = entitlementClient.GrantUserEntitlement(ctx, &platform.GrantUserEntitlementRequest{
            Namespace: namespace,
            UserID: userID,
            ItemID: "speed_booster",
            Quantity: 1,
            Source: "CHALLENGE_COMPLETE",
            Metadata: map[string]interface{}{
                "matches_remaining": 5,
                "multiplier": 1.5,
            },
        })
        if err != nil {
            return err
        }
        
        // Mark challenge complete
        state.Challenge.Active = false
    }
    
    // Save updated state to Redis
    stateJSON, _ := json.Marshal(state)
    return redis.HSet(ctx, fmt.Sprintf("churn_state:%s", userID), "state", stateJSON).Err()
}
```

**Why Direct Entitlement (No Reward Service)?**
- ✅ Simpler: One less service to configure
- ✅ More control: Custom logic in Extend
- ✅ Flexible: Easy to modify challenge conditions
- ✅ Direct: No waiting for Reward Service processing

**Rate Limiting in Extend (Go with Redis):**
```go
// Check cooldown before triggering intervention
func canTriggerIntervention(ctx context.Context, userID string) (bool, error) {
    var state ChurnState
    err := redis.HGet(ctx, fmt.Sprintf("churn_state:%s", userID), "state").Scan(&state)
    if err == redis.Nil {
        return true, nil // No previous intervention
    }
    if err != nil {
        return false, err
    }
    
    // Check cooldown (7 days)
    if time.Now().Before(state.Intervention.CooldownUntil) {
        return false, nil // Still in cooldown
    }
    
    return true, nil
}
```

**How Booster Works in Game:**

```go
// During match, when player presses Enter
func onPlayerPress(userID string) {
    pressValue := 1.0
    
    // Check for active booster
    booster, err := getActiveBooster(userID)
    if err == nil && booster != nil && *booster.ItemID == "speed_booster" {
        pressValue = 1.5 // Each press counts as 1.5
    }
    
    playerPressCount += pressValue
    
    // Update display showing booster is active
    if booster != nil {
        displayBoosterIndicator(*booster.UseCount)
    }
}

// After match ends
func onMatchEnd(userID string, won bool) {
    booster, err := getActiveBooster(userID)
    if err != nil || booster == nil {
        return
    }
    
    // Decrement matches remaining
    _ = entitlementService.ConsumeUserEntitlement(&entitlement.ConsumeUserEntitlementParams{
        Namespace:     namespace,
        UserID:        userID,
        EntitlementID: *booster.ID,
        Body: &platformclientmodels.EntitlementDecrement{
            UseCount: int32Ptr(1),
        },
    })
    
    matchesRemaining := *booster.UseCount - 1
    if matchesRemaining <= 0 {
        showNotification("⚠️ Speed Booster depleted!")
    } else {
        showNotification(fmt.Sprintf("🚀 Booster: %d matches left", matchesRemaining))
    }
}
```

---

### Reward Delivery Flow

**Step-by-Step (Extend → Entitlement Direct):**

```
1. Player exhibits churn behavior
   → 3 rage quits this week
   
2. Anti-Churn Extend detects signal
   → rage_quit_count_weekly >= 3
   
3. Extend Event Handler triggers intervention
   → Checks rate limit (last intervention > 7 days ago)
   → Creates challenge statistics:
      • challenge_active = 1
      • challenge_wins_needed = 3
      • challenge_wins_current = 0
      • challenge_expires_at = timestamp
   
4. Extend directly calls Entitlement Service API
   → Creates conditional entitlement (not granted yet)
   → ItemId: speed_booster
   → Metadata: { requires_wins: 3, granted: false }
   
5. Game client checks stats on next login
   → Reads challenge_active statistic
   → Finds active challenge
   
6. Display challenge notification in terminal
   ╔════════════════════════════════════╗
   ║  🎁 SPECIAL COMEBACK CHALLENGE!   ║
   ║  Win 3 matches in 3 days          ║
   ║  Reward: Speed Booster (5 matches)║
   ╚════════════════════════════════════╝
   
7. Track progress in main menu
   Challenge Progress: [█░░] 1/3 wins
   Time Remaining: 2 days 14 hours
   
8. Player wins a match
   → Game updates match_wins statistic
   → Extend Event Handler listens to statItemUpdated
   
9. Extend checks challenge progress
   → Increments challenge_wins_current
   → If challenge_wins_current >= 3:
      → Calls Entitlement API to grant booster
      → Updates challenge_active = 0
      → Grants actual usable booster entitlement
   
10. Game client receives entitlement
    ╔════════════════════════════════════╗
    ║  🎉 CHALLENGE COMPLETE! 🎉        ║
    ║  Reward Granted:                   ║
    ║  • Speed Booster (5 matches)      ║
    ╚════════════════════════════════════╝
    
11. Booster automatically active next match
    → Visual indicator in match screen
    → Each press counts as 1.5
```

**Key Difference from Reward Service:**
- ❌ No AGS Reward Service involved
- ✅ Extend Event Handler has full control
- ✅ Direct API calls to Entitlement Service
- ✅ Custom logic in Extend code

---

### Alternative: If Player Doesn't Complete Challenge

**Scenario:** Player accepts challenge but only wins 1/3 matches before expiry

```
Challenge Status: EXPIRED
Player gets: Nothing (challenge failed)

What happens:
- Challenge removed from UI
- No rewards granted
- Player can trigger new intervention next week
- System logs: challenge_not_completed
```

**This is OK because:**
- Not everyone should get rewards (devalues them)
- Requires actual effort and improvement
- Makes completion more meaningful
- Still encourages engagement attempt

---

## Implementation Plan

### Week 1: Build CLI Game (Single-Player vs Bot)

**Complete details:** See [Ready, Set, Enter! - Game Design Document](ready-set-enter-game-design.md)

**Days 1-2: Core Game Foundation**
- Project setup + dependencies
- AccelByte IAM authentication
- Terminal UI framework (blessed)
- Keyboard input handling
- Press counting logic
- Match timer

**Days 3-4: Bot Implementation**
- Bot class with dynamic speed
- Single balanced difficulty (~8.3 press/sec)
- Natural variance + bursts + fatigue
- Match result logic
- Victory/defeat screens

**Days 5-6: AccelByte Integration**
- Create AGS statistics
- Update stats after matches
- MMR calculation
- Booster entitlement logic
- Reward checking on login
- Challenge tracking UI

**Day 7: Testing & Polish**
- Unit tests
- End-to-end testing
- Bot behavior validation
- UI polish
- Bug fixes

**Deliverables After Week 1:**
- ✅ Fully functional CLI game vs bot
- ✅ All AccelByte services integrated
- ✅ 3 bot difficulty levels working
- ✅ Ready for anti-churn system integration

---

### Week 2-3: Build Anti-Churn System

**Week 3: Event Processing & Detection**
- Day 1-2: Extend setup
  - Create Extend Event Handler app
  - Configure Kafka subscriptions
  - Set up local dev environment
  - Deploy to AccelByte
  
- Day 3-4: Rage quit detector
  - Listen to GameSessionEndedEvent
  - Detect mid-match forfeit
  - Calculate player deficit
  - Update rage_quit_count_weekly stat
  
- Day 5-7: Session frequency tracker
  - Listen to oauthTokenGenerated
  - Track logins per week
  - Calculate decline percentage
  - Implement weekly reset logic

**Week 4: Intervention Logic & Testing**
- Day 1-2: Losing streak detector
  - Track match_wins and match_losses
  - Calculate current_losing_streak
  - Reset on win
  
- Day 3-4: Combined trigger logic
  - Implement simple scoring (IF/THEN)
  - Cooldown and rate limiting
  - Update churn_intervention_trigger stat
  - Logging and monitoring
  
- Day 5-7: Integration testing
  - Test with game client
  - Verify reward triggering
  - Test cooldowns
  - End-to-end validation

**Deliverables After Week 4:**
- ✅ Anti-churn system deployed
- ✅ All 3 churn signals working
- ✅ Rewards triggering correctly
- ✅ Integration tested end-to-end

---

### Week 5: Bot-Based Validation

**Goal:** Validate system with automated bot testing

**Day 1-2: Bot Swarm Generation**
- Create bot population with diverse behaviors
  ```bash
  # Create 100 bot accounts
  go run scripts/spawn_bots.go --count=100 --distribution=diverse
  
  Distribution:
  - 30 bots: Rage quit behavior
  - 30 bots: Session decline behavior
  - 30 bots: Losing streak behavior
  - 10 bots: Normal play (baseline)
  ```

**Day 3-5: Run Bot Testing**
- Bots play automatically 24/7
- Generate ~1000 matches total
- Churn behaviors trigger interventions
- Challenges created and tracked
- Validate detection accuracy

**Day 6-7: Analysis & Fixes**
- Review detection accuracy (target: 80%+)
- Check challenge completion rates
- Verify entitlement grants
- Fix any issues found
- Document results

**Deliverables After Week 5:**
- ✅ System validated with bots
- ✅ Detection accuracy measured
- ✅ Challenge flow working correctly
- ✅ Entitlement grants functioning

---

### Week 6: Final Validation & Polish**
**Day 1-3: Extended Bot Testing**
- Continue running bots
- Monitor system stability
- Check Redis performance
- Validate Extend reliability (99%+ uptime)

**Day 4-5: Documentation & Review**
- Document all findings
- Prepare demo for stakeholders
- Create decision document
- Review with team

**Day 6-7: Buffer & Fixes**
- Fix any remaining issues
- Polish rough edges
- Prepare final presentation

**Deliverables After Week 6:**
- ✅ Fully validated system
- ✅ Detection accuracy report
- ✅ Challenge completion metrics
- ✅ System reliability confirmed
- ✅ Ready for decision

---

### Week 7: Final Decision

**Goal:** Decide whether to proceed to real player testing

**Day 1-3: Comprehensive Review**
- Review all validation results
- Check detection accuracy (target: 80%+)
- Verify zero critical bugs
- Assess system reliability
- Calculate implementation effort for production

**Day 4-5: Stakeholder Presentation**
- Present findings to leadership
- Demo the working system
- Show bot test results
- Discuss next steps

**Day 6-7: Decision & Planning**
- **Decision Point:** Proceed to real player testing?
  - If YES: Plan production rollout with real game
  - If NO: Document learnings, archive PoC

**Success Metrics Review:**
- ✅ Detection accuracy: 80%+ (measured with bots)
- ✅ Challenge flow: Working correctly
- ✅ Entitlement grants: 100% success rate
- ✅ Zero critical bugs
- ✅ Extend reliability: 99%+ uptime

**Deliverables After Week 7:**
- ✅ Go/No-Go decision made
- ✅ If GO: Production roadmap defined
- ✅ If NO-GO: Lessons learned documented

---

## Next Milestone (If PoC Succeeds)

**Real Player Testing with Production Game:**
- Deploy to actual game (not CLI PoC)
- A/B test with real players
- Measure actual retention improvement
- Validate ROI
- Scale to multi-game platform if successful

This PoC validates **functionality**. The next milestone validates **business impact**.

---

## Testing Approach

### Phase 1: Unit Testing (Week 2)

**Test Scenarios:**

1. **Rage Quit Detection:**
```
Test Case 1: Normal quit (no rage quit)
  - Player plays for 10 minutes
  - Dies 2 times
  - Quits 5 minutes after last death
  - Expected: rage_quit_count = 0

Test Case 2: Rage quit detected
  - Player plays for 5 minutes
  - Dies 3 times in last 2 minutes
  - Quits 30 seconds after 3rd death
  - Expected: rage_quit_count = 1

Test Case 3: Multiple rage quits trigger intervention
  - Day 1: 1 rage quit
  - Day 2: 1 rage quit  
  - Day 3: 1 rage quit (3rd this week)
  - Expected: Intervention triggered
```

2. **Session Frequency Decline:**
```
Test Case 1: Normal player (no decline)
  - Last week: 7 logins
  - This week: 6 logins
  - Expected: No intervention (decline < 50%)

Test Case 2: At-risk player
  - Last week: 8 logins
  - This week: 3 logins
  - Expected: Intervention triggered (62% decline)
```

---

### Phase 2: Integration Testing (Week 3)

**End-to-End Flow Test:**

1. Create test player account
2. Simulate rage quit behavior:
   - Play match, die 3+ times
   - Quit within 60 seconds
   - Repeat 3 times in a week
3. Verify:
   - System detects rage quits
   - Intervention triggered
   - Reward granted via AGS
   - Challenge tracked correctly
4. Complete challenge:
   - Play 3 matches
   - Verify reward completion
   - Check coins and items added

**AGS Integration Tests:**
- Statistics updates correctly (including cycles)
- Extend Event Handler triggers on statItemUpdated
- Entitlement API called directly by Extend
- No duplicate interventions (cooldown works)
- Challenge progress tracking accurate

---

### Phase 3: Live Pilot (Week 4-5)

**Monitoring Checklist:**

Daily checks:
- [ ] System uptime (should be 100%)
- [ ] Events being processed (log count)
- [ ] Interventions triggered (daily count)
- [ ] Errors in logs (should be zero)
- [ ] Intervention triggers (daily count)
- [ ] Challenge completions (track rate)
- [ ] Entitlement grants (100% success rate)
- [ ] System errors (should be zero)

Weekly checks:
- [ ] Detection accuracy (review false positives/negatives)
- [ ] Redis performance (latency checks)
- [ ] Extend reliability (uptime monitoring)

**Red Flags (Fix immediately if):**
- System errors > 1% of events
- Rewards granted incorrectly
- Challenge tracking broken
- Redis connection issues
- Extend app crashes

---

## Success Metrics

### Primary Metrics: Functional Correctness

**1. Detection Accuracy**
- **Definition:** % of simulated churn behaviors correctly detected
- **Target:** 80%+ detection rate
- **Measurement:** Run bots with known churn behaviors, count detections
- **Example:** 100 rage quit bots → 85 detected = 85% ✅ SUCCESS

**2. Challenge Flow Correctness**
- **Definition:** Challenges created, tracked, and completed without errors
- **Target:** 100% success rate
- **Measurement:** 
  - All interventions create challenges correctly
  - Progress tracked accurately (wins counted)
  - Rewards granted upon completion
- **Example:** 50 challenges created → 50 tracked correctly → 30 completed → 30 rewards granted = 100% ✅ SUCCESS

**3. System Reliability**
- **Definition:** Extend Event Handler uptime and error rate
- **Target:** 99%+ uptime, <1% error rate
- **Measurement:**
  - Monitor Extend logs
  - Track event processing failures
  - Measure response times
- **Example:** 10,000 events processed → 9,995 success → 99.95% ✅ SUCCESS

---

### Secondary Metrics

| Metric | Target | Purpose |
|--------|--------|---------|
| **Challenge Completion Rate** | >40% | % of triggered challenges that are completed |
| **False Positive Rate** | <20% | % of interventions triggered incorrectly |
| **Redis Latency** | <50ms | State read/write performance |
| **Event Processing Time** | <100ms | Extend handler response time |
| **Intervention Rate** | 10-20% of players | Appropriate trigger frequency |

---

### Validation Approach

**Bot Testing:**
- 100+ bot accounts with known churn behaviors
- Run for 1 week minimum
- Measure accuracy against ground truth

**What We'll Know After PoC:**
- Does the system detect churn accurately?
- Does the challenge flow work correctly?
- Does Extend + Redis + AGS integration work smoothly?
- Is Go a good choice for Extend apps?
- Are we ready for real player testing?

---

## Go/No-Go Decision

### Success Criteria (After Week 7)

**Proceed to Real Player Testing IF:**

1. ✅ **Detection Accuracy:** System correctly identifies 80%+ of simulated churn behaviors
2. ✅ **Challenge Flow:** 100% success rate (created, tracked, completed correctly)
3. ✅ **Technical Stability:** 99%+ uptime, <1% error rate
4. ✅ **System Performance:** Redis latency <50ms, event processing <100ms
5. ✅ **Zero Critical Bugs:** No show-stoppers found

**Example Success:**
- 100 rage quit bots → 85 detected = 85% ✅
- 50 challenges created → 50 tracked → 30 completed → 30 rewards = 100% ✅
- 10,000 events processed → 9,995 success = 99.95% ✅
- Average Redis latency: 12ms ✅
- Zero critical bugs ✅

**Recommendation:** Deploy to production game for real player A/B testing

---

**Stop/Pivot IF:**

1. ❌ **Low Detection:** <70% of churn behaviors detected
2. ❌ **Challenge Flow Broken:** Tracking errors, rewards not granted correctly
3. ❌ **Reliability Issues:** <95% uptime or frequent errors
4. ❌ **Performance Issues:** Redis latency >100ms or event processing >500ms
5. ❌ **Integration Problems:** Extend + Redis + AGS not working smoothly

**Example Failure:**
- 100 rage quit bots → only 55 detected = 55% ❌
- 50 challenges created → 10 tracking errors = 80% success ❌
- 10,000 events → 500 errors = 95% success ❌
- Average Redis latency: 150ms ❌

**Recommendation:** Abandon approach or significantly redesign

---

**Pivot (Mixed Results) IF:**

⚠️ Some positive signals but issues to address:
- Detection accuracy 70-80% (borderline)
- Challenge flow mostly works but has edge cases
- System stable but occasional errors
- Performance acceptable but not great

**Example Mixed:**
- Detection: 75% (okay but not great)
- Challenge flow: 95% success (few bugs)
- Uptime: 97% (stable but not 99%+)

**Recommendation:** Fix issues, re-test for 1 more week, then decide

**Recommendation:** Extend PoC for another 2 weeks with modifications

---

## Resource Requirements

### Team

**Minimum Team:**
- 1 Senior Full-Stack Engineer (Weeks 1-10, full-time)
  - Builds CLI game (Week 1-2)
  - Builds anti-churn system (Week 3-4)
  - Integration and testing (Week 5-6)
  - Integration testing (Week 4)
  - Bot validation (Week 5-6)
  - Final review (Week 7)

**Optional (Recommended):**
- 1 QA Engineer (Week 4-5, part-time)
  - Validate integration
  - Test with bot scenarios
  - Verify edge cases
  
- 1 Junior Engineer (Week 5-6, part-time)
  - Help with bot management
  - Monitor bot testing
  - Data collection

**Total Person-Hours:**
- Senior Engineer: 400 hours (10 weeks × 40 hours)
- Data Analyst: 40 hours (2 weeks × 20 hours part-time)
- QA Engineer (optional): 80 hours (2 weeks × 40 hours)
- Junior Engineer (optional): 40 hours (2 weeks × 20 hours part-time)

**Total: 480-560 person-hours**

---

### Infrastructure

**AccelByte Services Used:**
- **IAM Service** - Player authentication (game login)
- **Statistics Service** - Track player stats with cycles (CRITICAL)
- **Entitlement Service** - Grant and track boosters
- **Extend Service** - Run anti-churn detection logic and grant items

**All included in AccelByte subscription** - No additional cost for PoC testing.

**Note:** We don't use Reward Service - Extend Event Handler directly calls Entitlement Service to grant boosters.

**Additional Tools:**
- **Grafana** - Monitoring dashboard (included with Extend)
- **Git Repository** - Code version control (GitHub/GitLab)
- **Spreadsheet** - Analysis and reporting (Google Sheets)

**No additional infrastructure costs for PoC.**

---

### Development Environment

**Local Development:**
- Go 1.21+ (for CLI game and Extend Event Handler)
- IDE (VS Code or GoLand recommended)
- AccelByte CLI tools
- Git client
- Redis client (for local testing, optional)

**AccelByte Configuration:**
- Test namespace for PoC
- Extend app deployment (Go runtime)
- Statistics configured with cycles (weekly reset)
- Item definitions for boosters
- Entitlement service configured
- Redis instance (managed by AccelByte)

**Setup Time:** 1-2 hours per engineer

---

### Timeline Summary

| Week | Activities | Team | Deliverable |
|------|-----------|------|-------------|
| **1** | Build CLI game (single-player vs bot) | 1 Senior Eng | Playable game with AccelByte |
| **2-3** | Build anti-churn Extend app (Go) | 1 Senior Eng | Detection + intervention working |
| **4** | Integration testing | 1 Senior Eng | System validated end-to-end |
| **5** | Bot-based validation | 1 Senior Eng | Accuracy measured, bugs fixed |
| **6** | Extended testing + polish | 1 Senior Eng | System stable and ready |
| **7** | Final review & decision | 1 Senior Eng<br>+Stakeholders | Go/No-Go decision |

**Total Duration: 6-7 weeks**

**Time Saved by Going Single-Player:** Multiple weeks (no multiplayer networking complexity)

---

## What Happens After PoC?

### If SUCCESS → Real Player Testing

**Validated Learnings:**
- ✅ Churn detection works accurately (80%+ with bots)
- ✅ Challenge flow functions correctly (100% success rate)
- ✅ Extend + Redis + AGS integration is smooth
- ✅ Go is suitable for Extend apps
- ✅ System is reliable (99%+ uptime)
- ✅ CLI PoC proved technical feasibility

**Next Steps:**
1. **Deploy to Production Game**
   - Choose existing production game (not CLI)
   - Integrate anti-churn Extend app
   - A/B test with REAL players
   - Measure ACTUAL retention improvement
   - Validate business impact

2. **If Retention Improves (3%+)**
   - Build pluggable multi-game platform
   - Add more churn signals (5 total)
   - Add more reward types
   - Add player segmentation
   - Production-grade monitoring
   - Timeline: 3-4 months

3. **Scale to Multiple Games**
   - Deploy to Game 2, 3, etc.
   - Configure per-game rules
   - Timeline: 1 week per game

**Expected Timeline:**
- PoC complete: Week 7
- Decision made: Week 7
- Real player testing: +4 weeks
- Full platform (if retention improves): +12-16 weeks
- **Total: ~6 months from start to multi-game production**

---

### If FAILURE → Learn & Iterate

**Possible Learnings:**
- Wrong signals chosen (try different detection logic)
- Wrong reward type (try coupons or discounts instead of conditional)
- Wrong timing (intervene earlier or later)
- Thresholds too strict or too loose (adjust sensitivity)
- CLI game not representative (try with real game)
- Concept doesn't work for our players (abandon or pivot completely)

**Options:**
1. **Pivot:** Adjust approach based on data, extend PoC 2-4 weeks
2. **Abandon:** Clear data shows concept won't work, stop investment
3. **Try Different Game Type:** Perhaps casual/puzzle game better than competitive

---

### If PIVOT → Adjust & Retry

**Common Adjustments:**

**If detection too sensitive:**
- Increase thresholds (3 rage quits → 5 rage quits)
- Extend timeframes (weekly → 2 weeks)
- Add more signals to improve accuracy

**If rewards not motivating:**
- Increase reward value (1000 coins → 2000 coins)
- Change reward type (conditional → direct coupon)
- Reduce completion requirement (3 wins → 2 wins)

**If completion rate too low (<30%):**
- Make challenge easier (3 wins → "win 1 OR play 5")
- Extend time limit (72 hours → 1 week)
- Add progress milestones (reward after 1 win)

**If one signal works, others don't:**
- Focus on single best signal only
- Remove confusing signals
- Simplify trigger logic

**Timeline for Pivot:** +2-4 weeks, then re-test for 2 weeks

---

## Key Takeaways

### What Makes This PoC Successful

1. **Custom Game = Complete Control**
   - Build exactly what you need for testing
   - No dependencies on existing games
   - Fast iteration and changes
   - Perfect for generating test data

2. **Single-Player = Maximum Simplicity**
   - 1 week to build (vs 2+ weeks for multiplayer)
   - Zero networking complexity
   - No matchmaking delays
   - Instant matches with bot opponent
   - Single balanced difficulty (no decision paralysis)

3. **CLI Approach = Rapid Development**
   - 7 days to build full game
   - Easy to automate testing with bots
   - Zero deployment complexity
   - Focus on integration, not graphics

4. **Bot System = Reliable Data**
   - Generate 100+ player histories in days
   - Simulate exact churn behaviors
   - Consistent and reproducible
   - Scale testing as needed

5. **Simple Scope = Clear Results**
   - Only 3 churn signals (easy to understand)
   - Single reward type (no confusion)
   - Binary outcome (works or doesn't)
   - Fast decision cycle

6. **7 Weeks = Reasonable Investment**
   - Not too long (maintains momentum)
   - Not too short (enough data)
   - Includes game development time
   - Clear milestone gates

### Critical Success Factors

**For This PoC (Functional Validation):**
- ✅ System detects churn accurately (80%+ of simulated scenarios)
- ✅ Challenges created and tracked correctly (100% success rate)
- ✅ Rewards granted correctly upon completion
- ✅ Integration works smoothly
- ✅ Zero critical bugs during testing
- ✅ Extend Event Handler performs reliably (99%+ uptime)
- ✅ Clear decision: Ready for real player testing?

**For Production (Next Milestone - Separate):**
- ✅ 3%+ retention improvement (measured with real players)
- ✅ Positive or neutral player sentiment
- ✅ ROI projection is favorable
- ✅ Stakeholder confidence gained

**Organizational:**
- ✅ Team learned AccelByte platform deeply
- ✅ Insights applicable to future projects
- ✅ Clear decision framework established
- ✅ Foundation for full system in place

---

## Final Recommendation

### For Studios with 2+ Games (Current or Planned):

**Do This PoC** - The investment is worth it:
- 10 weeks of 1 engineer = modest investment
- Complete validation before bigger commitment
- Reusable learnings for multiple games
- CLI game can become testing platform
- Clear go/no-go decision with data

### For Studios with Exactly 1 Game:

**Consider This PoC** - But evaluate carefully:
- Is retention a critical problem? (>50% monthly churn)
- Can you afford 10 weeks? (or compress to 6-8 weeks)
- Will you expand to more games eventually?
- If all yes → Do the PoC
- If uncertain → Start with functional validation first (this PoC)

### For Studios Exploring Anti-Churn for First Time:

**This PoC is PERFECT** - Here's why:
- Proves concept with minimal risk
- Builds internal expertise
- Creates decision-making framework
- Provides real data (not guesses)
- Low cost compared to full implementation

**Cost comparison:**
- PoC: 9 weeks, 1 engineer, zero infrastructure = Low risk
- Full system without PoC: 3 months, 2 engineers, production deployment = High risk if wrong approach

**Risk mitigation:**
- PoC costs ~$25-35K in engineering time
- Full system costs ~$150-200K
- **If PoC fails, you saved $115-165K** by not building wrong thing
- **If PoC succeeds, you have validated approach** and clear path forward

---

**END OF DOCUMENT**

---

*This PoC is designed to validate assumptions quickly with a custom game that provides complete control over testing. The 10-week timeline includes game development, system building, and thorough validation with both bots and real players.*

**Related Documents:**
- [Ready, Set, Enter! - Game Design Document](ready-set-enter-game-design.md)
- [Anti-Churn Pluggable Architecture](anti-churn-pluggable-architecture.md) (for full system)
- [Anti-Churn Board Meeting Proposal](anti-churn-reward-system-proposal.md) (for full system)

## Appendix

### A. Event Examples (For CLI Game)

**Events Generated by Ready, Set, Enter!:**

1. **oauthTokenGenerated (Login)**
```json
{
  "userId": "user_123",
  "timestamp": "2024-12-04T08:00:00Z",
  "clientId": "ready-set-enter",
  "sessionId": "session_abc"
}
```
**Used for:** Session frequency tracking

2. **GameSessionEndedEvent (Match End)**
```json
{
  "userId": "user_123",
  "sessionId": "match_xyz",
  "timestamp": "2024-12-04T14:24:00Z",
  "payload": {
    "winner": "user_456",
    "scores": {
      "user_123": 67,
      "user_456": 100
    },
    "duration": 42.3,
    "forfeit": false
  }
}
```
**Used for:** Rage quit detection, match completion tracking

3. **statItemUpdated (Match Result)**
```json
{
  "userId": "user_123",
  "statCode": "match_wins",
  "latestValue": 29,
  "inc": 1,
  "timestamp": "2024-12-04T14:24:01Z"
}
```
**Used for:** Losing streak detection, challenge progress

4. **statItemUpdated (Rage Quit Counter)**
```json
{
  "userId": "user_123",
  "statCode": "rage_quit_count_weekly",
  "latestValue": 3,
  "inc": 1,
  "timestamp": "2024-12-04T14:24:00Z"
}
```
**Used for:** Trigger intervention when >= 3

5. **statisticCycleReset (Weekly Reset)**
```json
{
  "userId": "user_123",
  "statCode": "logins_current_week",
  "valueBeforeReset": 6,
  "timestamp": "2024-12-09T00:00:00Z"
}
```
**Used for:** Capture previous week's value to store in `logins_previous_week`

**Note:** This event is generated by AGS Statistics Service when a Statistic Cycle resets (Monday 00:00). Extend Event Handler captures `valueBeforeReset` and writes it to `logins_previous_week` stat code.

---

### B. CLI Game Commands

**Running the Game:**
```bash
# Download binary (no installation needed!)
# Windows: ready-set-enter.exe
# Mac: ready-set-enter-mac
# Linux: ready-set-enter-linux

# Just run it:
./ready-set-enter

# Or from source:
go run cmd/game/main.go
```

**Bot Commands (For Testing):**
```bash
# Spawn 100 diverse bots
go run scripts/spawn_bots.go --count=100

# Spawn specific bot types
go run scripts/spawn_bots.go --type=rage_quit --count=20
go run scripts/spawn_bots.go --type=declining --count=30
go run scripts/spawn_bots.go --type=losing --count=25

# Run bots for specific duration
go run scripts/run_bot_swarm.go --duration=7days
```

---

### C. Simple Detection Pseudo-Code

**Main Detection Loop:**
```python
# Rage Quit Detector
def onMatchEnd(userId, matchData):
    if matchData.forfeit:  # Player quit early
        playerScore = matchData.scores[userId]
        opponentId = getOpponent(matchData)
        opponentScore = matchData.scores[opponentId]
        
        deficit = opponentScore - playerScore
        
        if deficit >= 30:  # Losing badly
            # This is a rage quit!
            incrementStat(userId, "rage_quit_count_weekly")
            
            rageQuits = getStat(userId, "rage_quit_count_weekly")
            if rageQuits >= 3:
                triggerIntervention(userId, "rage_quit")


# Session Frequency Tracker
def onLogin(userId):
    incrementStat(userId, "logins_current_week")
    
    loginsThisWeek = getStat(userId, "logins_current_week")
    loginsLastWeek = getStat(userId, "logins_previous_week")
    
    if loginsLastWeek > 0:
        decline = 1 - (loginsThisWeek / loginsLastWeek)
        
        if decline > 0.5:  # 50% decline
            triggerIntervention(userId, "session_decline")


# Losing Streak Tracker
def onMatchEnd(userId, won):
    if won:
        updateStat(userId, "current_losing_streak", 0)  # Reset
    else:
        incrementStat(userId, "current_losing_streak")
        
        streak = getStat(userId, "current_losing_streak")
        if streak >= 5:
            triggerIntervention(userId, "losing_streak")


# Intervention Trigger
def triggerIntervention(userId, reason):
    # Check cooldown (from Redis)
    state = redis.get(f"churn_state:{userId}")
    if state.intervention.cooldownUntil > now:
        return  # Too soon
    
    # Check weekly limit
    if state.intervention.countThisWeek >= 1:
        return  # Already got one this week
    
    # Trigger intervention
    createChallenge(userId, reason)
    updateRedisState(userId, state)
    
    log(f"Intervention triggered for {userId}, reason: {reason}")
```

---

### D. Quick Reference

**CLI Game:**
- See: [Ready, Set, Enter! - Game Design Document](ready-set-enter-game-design.md)
- Development Time: 7 days (1 week - single-player vs bot)
- Technology: Go + Bubble Tea + AccelByte Go SDK
- Distribution: Single binary (no installation required!)

**Storage Architecture (Hybrid Approach):**

**Why Hybrid?**
- AGS Statistics = Admin visibility (stakeholders can monitor churn metrics in dashboard)
- Redis = Flexible backend logic (derived state, complex structures)
- Hybrid = Best of both worlds for PoC

**AGS Statistics (Game-Generated + Churn Aggregations):**
- `match_wins` - Total wins (for UI display)
- `match_losses` - Total losses (for UI display)
- `current_losing_streak` - Consecutive losses (for UI display)
- `rage_quit_count_weekly` - Rage quits (Statistic Cycle, Monday reset) **← Admin visibility**
- `logins_current_week` - Login count (Statistic Cycle, Monday reset) **← Admin visibility**

**Redis (Extend-Managed State):**
```json
churn_state:{userId} = {
  "sessions": {
    "thisWeek": 3,
    "lastWeek": 6,
    "lastReset": "2024-12-02T00:00:00Z"
  },
  "challenge": {
    "active": true,
    "winsNeeded": 3,
    "winsCurrent": 1,
    "expiresAt": "2024-12-07T10:00:00Z",
    "triggeredBy": "rage_quit"
  },
  "intervention": {
    "lastTimestamp": "2024-11-28T10:00:00Z",
    "cooldownUntil": "2024-12-05T10:00:00Z",
    "countThisWeek": 1
  }
}
```

**Stat Code Configuration (AGS only):**
```yaml
# Game stats with weekly cycles
rage_quit_count_weekly:
  type: INTEGER
  increment_only: true
  cycle: WEEKLY (reset Monday 00:00)
  
logins_current_week:
  type: INTEGER
  increment_only: true
  cycle: WEEKLY (reset Monday 00:00)

# Game stats (no cycle)
match_wins:
  type: INTEGER
  increment_only: true
  
match_losses:
  type: INTEGER
  increment_only: true
```

**Benefits of Redis for Extend State:**
- ✅ No Statistic Cycles workaround needed
- ✅ Flexible JSON structure
- ✅ Fast key-value access
- ✅ Managed by AccelByte (zero setup)
- ✅ Clean separation (game stats vs backend logic)
  type: INTEGER
  increment_only: false
  cycle: NONE
```

**Storage Strategy:**
- Everything stored in AGS Statistics Service (unified storage)
- Game writes some stats, Extend writes others
- Extend listens to `statisticCycleReset` to capture previous week values
- No external database needed for PoC

**Thresholds:**
- Rage quit: 3 per week
- Session decline: 50%
- Losing streak: 5 consecutive losses
- Cooldown: 48 hours between interventions
- Weekly limit: 1 intervention per week per player
- Challenge requirement: Win 3 matches in 72 hours

**Success Criteria (Functional Correctness):**
- Primary: 80%+ detection accuracy for churn signals
- Secondary: Challenge completion flow works correctly
- Tertiary: Zero critical bugs, 99%+ uptime
- Next Milestone: Retention improvement testing with real players

---

**END OF DOCUMENT**

---

*This is a Proof of Concept document focused on functional correctness. Includes custom single-player CLI game development. Implementation should be simple and fast. Optimize for learning and validation, not production quality. Retention improvement testing comes after PoC succeeds.*

**Document Version:** 4.0 (Go Extend, Redis storage, no A/B testing - functional validation)
