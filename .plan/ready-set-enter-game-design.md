# Ready, Set, Enter!
## CLI Single-Player Game Design Document

**Version:** 1.0  
**Game Type:** CLI/Terminal Single-Player  
**Platform:** Cross-platform (Windows, Mac, Linux)  
**Purpose:** Testing platform for Anti-Churn Reward System PoC  
**Development Timeline:** 1 week

---

## Table of Contents

1. [Game Overview](#game-overview)
2. [Technical Architecture](#technical-architecture)
3. [Game Mechanics](#game-mechanics)
4. [AccelByte Integration](#accelbyte-integration)
5. [Terminal UI Design](#terminal-ui-design)
6. [Statistics & Progression](#statistics--progression)
7. [Reward System](#reward-system)
8. [Bot Implementation](#bot-implementation)
9. [Development Phases](#development-phases)
10. [Code Structure](#code-structure)

---

## Game Overview

### Concept

**Ready, Set, Enter!** is a simple, single-player terminal-based game where the player races against a bot to be the first to press Enter 100 times. The bot has dynamic, natural-feeling speed that creates realistic competitive matches. The game provides a perfect testing environment for churn detection algorithms without the complexity of multiplayer networking.

### Core Loop

```
1. Launch game CLI
2. Login with AccelByte account (required - no guest option)
3. Start match (instant - no selection needed)
4. Race to 100 presses against bot
5. View results and stats
6. Repeat or quit
```

**Important:** All players must register/login with real AccelByte accounts. No guest login is supported because:
- Anti-churn system requires persistent user tracking
- Statistics need stable user IDs for week-over-week comparisons
- Rewards and boosters require account persistence
- Anonymous play defeats the PoC's core purpose

### Target Metrics

- **Match Duration:** 30-60 seconds
- **Matches per Session:** 5-10 matches
- **Session Duration:** 5-15 minutes
- **Daily Active Pattern:** 2-3 sessions per day

---

## Technical Architecture

### Technology Stack

**Go (Golang)**

```
Dependencies:
├── AccelByte Go SDK (official SDK)
├── Bubble Tea (terminal UI framework - charmbracelet/bubbletea)
├── Lip Gloss (styling - charmbracelet/lipgloss)
└── Standard library (keyboard input, concurrency)
```

**Why Go:**
- **Single binary distribution** - no installation required!
- Official AccelByte Go SDK with full support
- Cross-compile for Windows/Mac/Linux from one machine
- Excellent terminal UI library (Bubble Tea - modern, easy Elm Architecture)
- Fast compilation and execution
- Small executable size (~10-15MB)
- Easy async with goroutines
- Perfect for CLI applications

**Distribution:**
```bash
# Build for all platforms
GOOS=windows GOARCH=amd64 go build -o ready-set-enter.exe
GOOS=darwin GOARCH=amd64 go build -o ready-set-enter-mac
GOOS=linux GOARCH=amd64 go build -o ready-set-enter-linux

# Players just download and run - NO installation needed!
./ready-set-enter
```

**Bubble Tea Framework:**

[Bubble Tea](https://github.com/charmbracelet/bubbletea) is a powerful, modern TUI framework based on The Elm Architecture.

**Why Bubble Tea:**
- ✅ **Simple & Elegant:** Clean Model-Update-View pattern (Elm Architecture)
- ✅ **Easy to Learn:** Much simpler than tview/tcell
- ✅ **Great Documentation:** Extensive examples and tutorials
- ✅ **Active Development:** Well-maintained by Charm (charmbracelet)
- ✅ **Message-based:** Clean event handling via commands and messages
- ✅ **Styling with Lip Gloss:** Powerful, chainable styling library
- ✅ **Perfect for Games:** Natural fit for game loops and state management

**Bubble Tea Pattern:**
```go
// 1. Define your Model (game state)
type model struct {
    playerScore int
    botScore    int
}

// 2. Initialize
func (m model) Init() tea.Cmd {
    return nil
}

// 3. Update (handle events)
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    // Handle key presses, messages, timers
    return m, nil
}

// 4. View (render UI)
func (m model) View() string {
    return "Your terminal UI here"
}
```

### System Architecture

```
┌─────────────────────────────────────────────────────────┐
│                   CLI Game Client                       │
│                (Single-Player vs Bot)                   │
│                                                         │
│  ┌──────────────┐  ┌──────────────┐  ┌─────────────┐ │
│  │  Terminal UI │  │  Game Logic  │  │  Bot Engine │ │
│  │ (Bubble Tea) │  │  (Press cnt) │  │  (Dynamic)  │ │
│  └──────────────┘  └──────────────┘  └─────────────┘ │
│         │                  │                  │         │
│         └──────────────────┴──────────────────┘         │
│                            │                            │
└────────────────────────────┼────────────────────────────┘
                             │
                             ↓
              ┌──────────────────────────┐
              │  AccelByte Gaming Services│
              ├──────────────────────────┤
              │ • IAM (Authentication)   │
              │ • Statistics Service     │
              │ • Entitlement (Boosters) │
              │ • Reward Service         │
              └──────────────┬───────────┘
                             │
                             ↓
              ┌──────────────────────────┐
              │   Anti-Churn System      │
              │   (Extend Event Handler) │
              └──────────────────────────┘
```

**Key Simplifications:**
- ✅ No matchmaking needed (instant bot opponent)
- ✅ No session management (single-player)
- ✅ No networking/WebSocket (local bot)
- ✅ No lobby service (no multiplayer)
- ✅ Faster development (1 week vs 2 weeks)
- ✅ Zero network dependency or latency issues

---

## Game Mechanics

### Core Gameplay

#### Match Flow

1. **Pre-Match (Instant)**
   - Player clicks "Start Match"
   - Bot opponent is created instantly (no wait!)
   - 3-second countdown: "3... 2... 1... GO!"

2. **Match Phase (30-60 seconds)**
   - Player presses Enter as fast as possible
   - Bot presses at balanced speed (natural variance)
   - Real-time display of both player's and bot's progress
   - First to 100 presses wins

3. **Post-Match (5 seconds)**
   - Display final scores
   - Show winner/loser
   - Update statistics
   - Return to main menu

#### Press Detection (Bubble Tea Model)

```go
import (
    tea "github.com/charmbracelet/bubbletea"
)

// Model - game state
type model struct {
    playerPresses int
    botPresses    int
    target        int
    gameStarted   bool
    gameOver      bool
    winner        string
}

// Init - initialize the model
func (m model) Init() tea.Cmd {
    return nil
}

// Update - handle messages (key presses, bot updates)
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        if msg.String() == "enter" && m.gameStarted && !m.gameOver {
            m.playerPresses++
            
            if m.playerPresses >= m.target {
                m.gameOver = true
                m.winner = "player"
            }
        }
    case botPressMsg:
        m.botPresses++
        
        if m.botPresses >= m.target {
            m.gameOver = true
            m.winner = "bot"
        }
        
        // Schedule next bot press
        if !m.gameOver {
            return m, botPressCmd()
        }
    }
    
    return m, nil
}

// View - render the UI
func (m model) View() string {
    // Render terminal UI (see UI Design section)
    return renderGameScreen(m)
}
```

**Note:** No anti-cheat or rate limiting needed. This is a PoC for testing the anti-churn system, not a competitive game. We trust players and focus on the Extend integration.

---

#### Bot Opponent with Dynamic Speed

**Purpose:** Create realistic, challenging matches without multiplayer complexity

**Bot Configuration (Single Balanced Difficulty):**

```go
type BotConfig struct {
    BaseSpeed        time.Duration // Time between presses
    Variance         time.Duration // Random variance
    FatigueRate      float64       // Fatigue multiplier per press
    BurstChance      float64       // Probability of burst
    BurstMultiplier  float64       // Speed multiplier during burst
}

var botConfig = BotConfig{
    BaseSpeed:       120 * time.Millisecond, // ~8.3 press/sec
    Variance:        30 * time.Millisecond,  // ±30ms
    FatigueRate:     0.001,                  // Minimal fatigue
    BurstChance:     0.10,                   // 10% chance
    BurstMultiplier: 0.7,                    // 30% faster
}
```

**Why This Speed:**
- Creates balanced matches (~45-55% win rate for average players)
- Fast enough to be challenging
- Slow enough to be beatable
- Natural variance makes each match feel different

**Bot Implementation (Bubble Tea Commands):**

```go
// Bot message for Bubble Tea
type botPressMsg struct{}

// Bot command - sends bot press message after delay
func botPressCmd() tea.Cmd {
    return func() tea.Msg {
        // Calculate delay with natural variance
        delay := calculateBotDelay()
        time.Sleep(delay)
        return botPressMsg{}
    }
}

type BotConfig struct {
    BaseSpeed        time.Duration // Time between presses
    Variance         time.Duration // Random variance
    FatigueRate      float64       // Fatigue multiplier per press
    BurstChance      float64       // Probability of burst
    BurstMultiplier  float64       // Speed multiplier during burst
}

var botConfig = BotConfig{
    BaseSpeed:       120 * time.Millisecond, // ~8.3 press/sec
    Variance:        30 * time.Millisecond,  // ±30ms
    FatigueRate:     0.001,                  // Minimal fatigue
    BurstChance:     0.10,                   // 10% chance
    BurstMultiplier: 0.7,                    // 30% faster
}

func calculateBotDelay() time.Duration {
    // Base delay
    delay := botConfig.BaseSpeed
    
    // Add natural human-like variance
    variance := time.Duration(rand.Float64()*2-1) * botConfig.Variance
    delay += variance
    
    // Check for burst (moments of focus/excitement)
    if rand.Float64() < botConfig.BurstChance {
        delay = time.Duration(float64(delay) * botConfig.BurstMultiplier)
    }
    
    // Apply fatigue (humans slow down slightly)
    fatigueMultiplier := 1.0 + (float64(botPressCount) * botConfig.FatigueRate)
    delay = time.Duration(float64(delay) * fatigueMultiplier)
    
    return delay
}
```

**Why This Creates Natural Feel:**

1. **Random Variance:** No two presses are exactly the same interval
2. **Burst Behavior:** Occasional fast sequences (like human focus moments)
3. **Gradual Fatigue:** Slightly slower as match progresses (realistic)
4. **Balanced Challenge:** ~45-55% win rate creates engagement without frustration

**Bot Behavior Example:**

```
Press 1: 125ms delay (base 120 + variance +5)
Press 2: 115ms delay (base 120 + variance -5)
Press 3:  84ms delay (BURST! base 120 * 0.7)
Press 4: 122ms delay (back to normal)
...
Press 95: 131ms delay (slight fatigue)
```

**Expected Win Rate Distribution:**
- New players: ~30-40% win rate (learning)
- Average players: ~45-55% win rate (balanced)
- Skilled players: ~60-70% win rate (mastery)

This creates enough losses to trigger churn signals while keeping the game engaging.

---

### Player Progression

#### Experience System

**Levels Based on Total Wins:**
```
Level 1: 0 - 10 wins        (Beginner)
Level 2: 11 - 30 wins       (Amateur)
Level 3: 31 - 75 wins       (Intermediate)
Level 4: 76 - 150 wins      (Advanced)
Level 5: 151+ wins          (Master)
```

**Visual Indicator:**
```
Player: FastFingers [Lvl 3] ⭐⭐⭐
        └─ 45 total wins
```

#### Ranking System

**Based on Win Rate (Last 20 Matches):**
```
Bronze:   0-40% win rate   🥉
Silver:   41-55% win rate  🥈
Gold:     56-70% win rate  🥇
Platinum: 71-85% win rate  💎
Diamond:  86%+ win rate    💠
```

---

### Boosters & Power-ups

#### Press Speed Booster

**Effect:** Each press counts as 1.5 presses  
**Duration:** 5 matches  
**How to Get:** Reward from anti-churn system or purchase

**Implementation:**
```go
func applyBooster(pressCount float64, hasActiveBooster bool) float64 {
    if hasActiveBooster {
        return pressCount * 1.5
    }
    return pressCount
}
```

**Visual Indicator:**
```
╔════════════════════════════════════════╗
║  🚀 BOOSTER ACTIVE: 1.5x Speed        ║
║  Remaining: 3 matches                  ║
╚════════════════════════════════════════╝
```

#### Auto-Press Simulator (Testing Only)

**Effect:** Automatically presses at configurable rate  
**Purpose:** Generate test data, simulate player behaviors  
**Not available to real players**

---

## AccelByte Integration

### Required Services (Simplified)

For a single-player game, we only need 3 AccelByte services:

1. **IAM Service** - Player authentication (login)
2. **Statistics Service** - Track player stats (CRITICAL for churn detection, admin visibility)
3. **Entitlement Service** - Track boosters/items granted by Extend

**Additional Backend (Extend Event Handler):**
- **Redis** - Managed by AccelByte, stores derived state (previous week sessions, challenge tracking, cooldowns)

**Note:** 
- No Reward Service needed - Extend directly grants entitlements
- No matchmaking, session, or lobby services needed for single-player!
- Hybrid approach: AGS Statistics for visibility, Redis for complex logic

---

#### 1. IAM Service (Identity & Access Management)

**Purpose:** Player authentication, registration, and login (required - no guest access)

**Implementation:**
```go
import (
    "github.com/AccelByte/accelbyte-go-sdk/iam-sdk/pkg/iamclient"
    "github.com/AccelByte/accelbyte-go-sdk/services-api/pkg/service/iam"
)

// Login flow
func login(username, password string) (*iamclientmodels.OauthmodelTokenResponse, error) {
    input := &o_auth2_0.TokenGrantV3Params{
        GrantType: "password",
        Username:  &username,
        Password:  &password,
    }
    
    token, err := iamService.TokenGrantV3(input)
    if err != nil {
        return nil, err
    }
    
    // Store token for subsequent API calls
    setAuthToken(token.AccessToken)
    
    return token, nil
}
```

**Note:** No guest login support. All players must have real AccelByte accounts for persistent tracking and churn detection.

---

#### 2. Statistics Service

**Purpose:** Track player stats and trigger churn detection

**Statistics to Track:**

```go
// Stat codes
const (
    // Match results
    MATCH_WINS   = "match_wins"
    MATCH_LOSSES = "match_losses"
    MATCH_PLAYED = "match_played"
    
    // Performance metrics
    TOTAL_PRESSES        = "total_presses"
    AVG_PRESSES_PER_MATCH = "avg_presses_per_match"
    BEST_TIME_TO_100     = "best_time_to_100"
    
    // Churn signals (game-generated)
    RAGE_QUIT_COUNT      = "rage_quit_count_weekly"
    LOGINS_CURRENT_WEEK  = "logins_current_week"
    LOSING_STREAK        = "current_losing_streak"
    LAST_PLAYED_DATE     = "last_played_date"
)

// Note: Previous week sessions, challenge state, and intervention cooldowns
// are managed by Extend Event Handler in Redis, not in AGS Statistics
```

**Update Statistics:**
```go
import (
    "github.com/AccelByte/accelbyte-go-sdk/services-api/pkg/service/social"
)

func updateStats(userID string, matchResult MatchResult) error {
    // Update match counts
    if matchResult.Won {
        err := statsService.IncrementUserStatItem(&social.IncrementUserStatItemParams{
            Namespace: namespace,
            StatCode:  MATCH_WINS,
            UserID:    userID,
            Inc: &socialclientmodels.StatItemInc{
                Inc: floatPtr(1),
            },
        })
        if err != nil {
            return err
        }
        
        // Reset losing streak
        err = statsService.UpdateUserStatItemValue(&social.UpdateUserStatItemValueParams{
            Namespace: namespace,
            StatCode:  LOSING_STREAK,
            UserID:    userID,
            Body: &socialclientmodels.StatItemUpdate{
                UpdateStrategy: stringPtr("OVERRIDE"),
                Value:         floatPtr(0),
            },
        })
        if err != nil {
            return err
        }
    } else {
        // Increment losses and losing streak
        _ = statsService.IncrementUserStatItem(&social.IncrementUserStatItemParams{
            Namespace: namespace,
            StatCode:  MATCH_LOSSES,
            UserID:    userID,
            Inc: &socialclientmodels.StatItemInc{
                Inc: floatPtr(1),
            },
        })
        
        _ = statsService.IncrementUserStatItem(&social.IncrementUserStatItemParams{
            Namespace: namespace,
            StatCode:  LOSING_STREAK,
            UserID:    userID,
            Inc: &socialclientmodels.StatItemInc{
                Inc: floatPtr(1),
            },
        })
    }
    
    // Update press count
    _ = statsService.IncrementUserStatItem(&social.IncrementUserStatItemParams{
        Namespace: namespace,
        StatCode:  TOTAL_PRESSES,
        UserID:    userID,
        Inc: &socialclientmodels.StatItemInc{
            Inc: floatPtr(float64(matchResult.FinalPressCount)),
        },
    })
    
    return nil
    
    return nil
}
```

**Track Rage Quits:**
```go
func trackRageQuit(userID string) error {
    return statsService.IncrementUserStatItem(&social.IncrementUserStatItemParams{
        Namespace: namespace,
        StatCode:  RAGE_QUIT_COUNT,
        UserID:    userID,
        Inc: &socialclientmodels.StatItemInc{
            Inc: floatPtr(1),
        },
    })
}
```

---

#### 3. Entitlement Service

**Purpose:** Grant and track boosters/items

**Booster Entitlement:**
```go
import (
    "github.com/AccelByte/accelbyte-go-sdk/services-api/pkg/service/platform"
)

func grantBooster(userID, boosterType string) error {
    input := &entitlement.GrantUserEntitlementParams{
        Namespace: namespace,
        UserID:    userID,
        Body: &platformclientmodels.EntitlementGrant{
            ItemID:     &boosterType, // 'speed_booster'
            Quantity:   int32Ptr(1),
            Source:     stringPtr("REWARD"),
            ItemNamespace: &namespace,
        },
    }
    
    _, err := entitlementService.GrantUserEntitlement(input)
    return err
}

func checkActiveBooster(userID string) (*platformclientmodels.EntitlementInfo, error) {
    input := &entitlement.QueryUserEntitlementsParams{
        Namespace:  namespace,
        UserID:     userID,
        ItemID:     stringPtr("speed_booster"),
        ActiveOnly: boolPtr(true),
    }
    
    result, err := entitlementService.QueryUserEntitlements(input)
    if err != nil {
        return nil, err
    }
    
    if len(result.Data) > 0 {
        return result.Data[0], nil
    }
    
    return nil, nil
}
```

---

#### 4. Reward Service (Handled by Anti-Churn System)

**Purpose:** Automatically grant rewards when churn interventions are triggered

**Implementation:** The Reward Service is configured and triggered by the Anti-Churn Extend app. The game client simply checks for pending rewards on login and displays them to the player.

```go
// Check for pending rewards
func checkPendingRewards(userID string) ([]*platformclientmodels.RewardInfo, error) {
    input := &reward.GetRewardsParams{
        Namespace: namespace,
        UserID:    userID,
        Status:    stringPtr("PENDING"),
    }
    
    result, err := rewardService.GetRewards(input)
    if err != nil {
        return nil, err
    }
    
    for _, reward := range result.Data {
        if reward.Type != nil && *reward.Type == "CONDITIONAL" {
            displayChallengeNotification(reward)
            trackChallengeProgress(reward)
        }
    }
    
    return result.Data, nil
}
```
    status: 'PENDING'
  });
  
  for (const reward of rewards) {
    if (reward.type === 'CONDITIONAL') {
      displayChallengeNotification(reward);
      trackChallengeProgress(reward);
    }
  }
}
```

---

## Terminal UI Design

### Main Menu

```
╔════════════════════════════════════════════════════════╗
║                                                        ║
║          🏁  READY, SET, ENTER!  🏁                   ║
║                                                        ║
║              Main Menu                                 ║
║                                                        ║
╠════════════════════════════════════════════════════════╣
║                                                        ║
║  Player: SpeedDemon [Lvl 3] 🥈 Silver                ║
║  Win Rate: 58% (45W / 32L)                            ║
║  Current Streak: 3 losses 📉                          ║
║                                                        ║
║  🚀 Active Booster: 1.5x Speed (3 matches left)      ║
║                                                        ║
╠════════════════════════════════════════════════════════╣
║                                                        ║
║  1. Start Match                                       ║
║  2. View Stats                                        ║
║  3. Check Rewards                                     ║
║  4. Settings                                          ║
║  5. Quit                                              ║
║                                                        ║
╚════════════════════════════════════════════════════════╝

Enter choice (1-5): _
```

---

### Match Starting Screen

```
╔════════════════════════════════════════════════════════╗
║              Match Starting!                           ║
╠════════════════════════════════════════════════════════╣
║                                                        ║
║  👤 You:       SpeedDemon [Lvl 3] 🥈                  ║
║     Win Rate:  58%                                     ║
║                                                        ║
║     VS                                                 ║
║                                                        ║
║  🤖 Bot:       Press Bot                              ║
║     Speed:     ~8.3 presses/sec                       ║
║                                                        ║
║  Starting in: 3... 2... 1... GO!                      ║
║                                                        ║
╚════════════════════════════════════════════════════════╝
```

---

### Match Screen (Live Game)

```
╔════════════════════════════════════════════════════════╗
║        READY, SET, ENTER! - MATCH IN PROGRESS         ║
╠════════════════════════════════════════════════════════╣
║                                                        ║
║  YOU (SpeedDemon):                                     ║
║  [████████████████████░░░░░░░░░░]  67/100            ║
║  Time: 34.2s | Speed: 1.96 press/sec                  ║
║                                                        ║
║  BOT (Medium):                                         ║
║  [████████████████████████░░░░░░]  75/100            ║
║  Time: 34.2s | Speed: 2.19 press/sec                  ║
║                                                        ║
║  🚀 BOOSTER ACTIVE: 1.5x Speed                        ║
║                                                        ║
╠════════════════════════════════════════════════════════╣
║                                                        ║
║          >>> PRESS ENTER AS FAST AS YOU CAN! <<<      ║
║                                                        ║
║              Press ESC to forfeit match                ║
║                                                        ║
╚════════════════════════════════════════════════════════╝

Bot progress updated in real-time (local simulation)...
```

---

### Match Result Screen

**Victory:**
```
╔════════════════════════════════════════════════════════╗
║                    🏆 VICTORY! 🏆                      ║
╠════════════════════════════════════════════════════════╣
║                                                        ║
║  Final Score:                                          ║
║  • You:       100 presses in 42.8 seconds             ║
║  • Bot:        89 presses in 42.8 seconds             ║
║                                                        ║
║  Performance:                                          ║
║  • Average Speed: 2.34 press/sec                      ║
║  • Personal Best: No (Best: 2.89 press/sec)           ║
║                                                        ║
║  Rewards:                                              ║
║  • +10 Rating (1450 → 1460)                           ║
║  • +50 coins                                           ║
║  • Losing streak broken! 🎉                           ║
║                                                        ║
║  Win Streak: 1                                         ║
║  Updated Win Rate: 60% (46W / 31L)                    ║
║                                                        ║
╚════════════════════════════════════════════════════════╝

Press ENTER to continue...
```

**Defeat:**
```
╔════════════════════════════════════════════════════════╗
║                    😞 DEFEAT 😞                        ║
╠════════════════════════════════════════════════════════╣
║                                                        ║
║  Final Score:                                          ║
║  • You:        78 presses in 45.3 seconds             ║
║  • Bot:       100 presses in 45.3 seconds             ║
║                                                        ║
║  Performance:                                          ║
║  • Average Speed: 1.72 press/sec                      ║
║  • Below your average (2.1 press/sec)                 ║
║                                                        ║
║  Results:                                              ║
║  • -5 Rating (1450 → 1445)                            ║
║  • +10 coins (participation)                           ║
║                                                        ║
║  ⚠️  Losing Streak: 4 matches                         ║
║  Updated Win Rate: 58% (45W / 33L)                    ║
║                                                        ║
║  💡 Tip: Take a break and come back fresh!            ║
║                                                        ║
╚════════════════════════════════════════════════════════╝

Press ENTER to continue...
```

---

### Churn Intervention Screen

**Appears after match if intervention triggered:**

```
╔════════════════════════════════════════════════════════╗
║           🎁 SPECIAL COMEBACK CHALLENGE! 🎁           ║
╠════════════════════════════════════════════════════════╣
║                                                        ║
║  We noticed you've had a tough streak lately!          ║
║  Here's a special challenge to help you bounce back:   ║
║                                                        ║
║  CHALLENGE: Win 3 matches in the next 3 days          ║
║                                                        ║
║  REWARDS when completed:                               ║
║  • 1000 coins 💰                                       ║
║  • 1.5x Press Speed Booster (5 matches) 🚀            ║
║  • "Comeback King" badge 👑                           ║
║                                                        ║
║  Progress:                                             ║
║  [░░░░░░░░░░░░] 0/3 wins                              ║
║                                                        ║
║  Time Remaining: 3 days 0 hours                        ║
║                                                        ║
║  This challenge will expire on Dec 7, 2024            ║
║                                                        ║
╚════════════════════════════════════════════════════════╝

Press ENTER to accept challenge...
```

**Challenge Progress (in main menu):**
```
╔════════════════════════════════════════════════════════╗
║         📊 ACTIVE CHALLENGE: Comeback King            ║
╠════════════════════════════════════════════════════════╣
║                                                        ║
║  Progress: [████████░░░░] 2/3 wins                    ║
║                                                        ║
║  Time Remaining: 1 day 14 hours                        ║
║                                                        ║
║  Keep going! You're almost there! 💪                  ║
║                                                        ║
╚════════════════════════════════════════════════════════╝
```

---

### Statistics Screen

```
╔════════════════════════════════════════════════════════╗
║                  Player Statistics                      ║
╠════════════════════════════════════════════════════════╣
║                                                        ║
║  Player: SpeedDemon [Lvl 3] 🥈 Silver                ║
║  Total Playtime: 12 hours 34 minutes                   ║
║  Member Since: Nov 15, 2024                            ║
║                                                        ║
║  ───────────────────────────────────────────────────  ║
║                                                        ║
║  MATCH HISTORY                                         ║
║  • Total Matches: 77                                   ║
║  • Wins: 45 (58%)                                      ║
║  • Losses: 32 (42%)                                    ║
║  • Win Streak (Best): 7                               ║
║  • Current Streak: 4 losses 📉                        ║
║                                                        ║
║  PERFORMANCE                                           ║
║  • Total Presses: 4,250                               ║
║  • Avg Presses/Match: 85                              ║
║  • Best Match: 100 in 38.2s (2.62 press/sec)          ║
║  • Avg Speed: 2.1 press/sec                           ║
║                                                        ║
║  RANKING                                               ║
║  • Rating: 1450 (Silver Tier)                         ║
║  • Progress to Gold: 50 points needed                 ║
║                                                        ║
║  ACTIVITY                                              ║
║  • Sessions This Week: 5                              ║
║  • Sessions Last Week: 9 📉                           ║
║  • Last Played: 2 hours ago                           ║
║                                                        ║
╚════════════════════════════════════════════════════════╝

Press ENTER to return to menu...
```

---

## Statistics & Progression

### Core Statistics Tracked (Simplified for Anti-Churn PoC)

| Statistic | AGS Stat Code | Purpose | Reset Cycle |
|-----------|---------------|---------|-------------|
| **Match Results (for losing streak calculation)** |
| Total Wins | `rse-match-wins` | Calculate win rate, reset losing streak | Never |
| Total Losses | `rse-match-losses` | Calculate win rate, increment losing streak | Never |
| Total Matches | `rse-match-played` | Track engagement, added to Daily Cycle | Never (Daily Cycle provides daily count) |
| **Anti-Churn Signals** |
| Rage Quits | `rse-rage-quit` | **SIGNAL 1:** Rage quit detection | Statistic Cycle (Weekly, Monday) |
| Session Logins | `rse-session-login` | **SIGNAL 2:** Login frequency | Statistic Cycle (Weekly, Monday) |
| Losing Streak | `rse-current-losing-streak` | **SIGNAL 3:** Consecutive losses | Reset on win |

**Note on Churn State Storage (Hybrid Approach):**  

The game writes statistics to AGS Statistics Service. The Extend Event Handler (Go) maintains additional state in Redis. We use BOTH for different purposes:

**Why Hybrid Storage?**
- **AGS Statistics** = Admin visibility (stakeholders can see churn metrics in real-time dashboard)
- **Redis** = Flexible backend logic (derived state, complex structures invisible to admins)
- **Hybrid** = Best of both for PoC (monitoring + flexibility)

**AGS Statistics (6 Base Event Stats):**
- `rse-session-login` - Uses Weekly Statistic Cycle (auto-resets Monday) **← Admin can monitor**
- `rse-rage-quit` - Uses Weekly Statistic Cycle (auto-resets Monday) **← Admin can monitor**
- `rse-current-losing-streak` - Reset on win **← Admin can monitor**
- `rse-match-wins`, `rse-match-losses`, `rse-match-played` - Match tracking

**Redis (Extend-managed Backend State):**
- Session tracking: `thisWeek`, `lastWeek`, `lastReset` (derived from login events)
- Challenge state: `active`, `winsNeeded`, `winsCurrent`, `expiresAt`
- Intervention cooldowns: `lastTimestamp`, `cooldownUntil`

This separation allows:
1. **Admin visibility** - Stakeholders see churn signals without custom dashboards
2. **Clean architecture** - Game generates stats, Extend manages backend logic
3. **Flexibility** - Redis stores complex JSON structures not possible in Statistics
4. **No workarounds** - Extend handles weekly resets and derived state naturally
5. **Managed by AccelByte** - Zero infrastructure setup for both services
6. **Simplified scope** - Only 6 stats needed for Anti-Churn PoC

---

### Player Rating System (Elo-based)

**Starting Rating:** 1000

**Win/Loss Adjustments:**
```go
func calculateRatingChange(playerRating, botRating int, won bool) int {
    const K = 32 // K-factor (standard for Elo)
    
    // Expected win probability
    expectedScore := 1.0 / (1.0 + math.Pow(10, float64(botRating-playerRating)/400.0))
    
    // Actual score (1 for win, 0 for loss)
    var actualScore float64
    if won {
        actualScore = 1.0
    } else {
        actualScore = 0.0
    }
    
    // Rating change
    ratingChange := int(math.Round(K * (actualScore - expectedScore)))
    
    return ratingChange
}
```

**Example:**
- Player Rating: 1450
- Bot Rating: 1500 (bots use fixed 1500 rating)
- Player wins: +12 Rating
- Player loses: -18 Rating

**Tier Boundaries:**
```
Bronze:   0 - 999
Silver:   1000 - 1499
Gold:     1500 - 1999
Platinum: 2000 - 2499
Diamond:  2500+
```

---

## Reward System

### Integration with Anti-Churn System

**Flow:**
```
1. Player exhibits churn signals (rage quits, losing streak, session decline)
   ↓
2. Anti-Churn Extend app detects signal
   ↓
3. Updates AGS Statistic: churn_intervention_trigger = 1
   ↓
4. AGS Reward Service triggers
   ↓
5. Creates pending reward (conditional challenge)
   ↓
6. Game client polls for pending rewards
   ↓
7. Display challenge notification
   ↓
8. Track progress via match_wins statistic
   ↓
9. Auto-grant rewards when challenge complete
```

### Reward Types in Game

#### 1. Conditional Challenge

**"Comeback Challenge"**
```yaml
trigger:
  - rage_quit_count >= 3
  - OR losing_streak >= 5
  - OR session_decline > 50%

challenge:
  task: "Win 3 matches"
  time_limit: 72 hours
  track_via: match_wins statistic

rewards_on_complete:
  - coins: 1000
  - item: speed_booster (5 matches)
  - badge: comeback_king

display:
  title: "Special Comeback Challenge!"
  message: "We noticed you've had a tough streak. Complete this challenge for rewards!"
```

**In-Game Implementation:**
```go
// Check for pending challenges on login
func checkPendingRewards(userID string) error {
    input := &reward.GetRewardsParams{
        Namespace: namespace,
        UserID:    userID,
        Status:    stringPtr("PENDING"),
    }
    
    rewards, err := rewardService.GetRewards(input)
    if err != nil {
        return err
    }
    
    for _, reward := range rewards.Data {
        if reward.Type != nil && *reward.Type == "CONDITIONAL" {
            displayChallengeNotification(reward)
            go trackChallengeProgress(userID, reward)
        }
    }
    
    return nil
}

// Track progress
func trackChallengeProgress(userID string, challenge *platformclientmodels.RewardInfo) {
    stats, err := statisticsService.GetUserStat(&social.GetUserStatParams{
        Namespace: namespace,
        UserID:    userID,
        StatCode:  "match_wins",
    })
    if err != nil {
        return
    }
    
    progress := int(*stats.Value) - challenge.StartingValue
    updateChallengeUI(progress, challenge.TargetValue)
    
    if progress >= challenge.TargetValue {
        // Challenge complete!
        _ = rewardService.ClaimReward(&reward.ClaimRewardParams{
            Namespace: namespace,
            UserID:    userID,
            RewardID:  *challenge.ID,
        })
        displayRewardGrantedScreen(challenge.Rewards)
    }
}
```

---

#### 2. Booster Items

**Speed Booster:**
```go
// When player starts match with active booster
func applyBoosterEffect(userID string, pressCount float64) float64 {
    booster, err := getActiveBooster(userID)
    if err != nil || booster == nil {
        return pressCount
    }
    
    if booster.ItemID != nil && *booster.ItemID == "speed_booster" {
        // Each press counts as 1.5
        return pressCount * 1.5
    }
    
    return pressCount
}

// Decrement matches remaining after match ends
func decrementBoosterMatches(userID, boosterID string) error {
    // Consume one "use" of the booster
    return entitlementService.ConsumeUserEntitlement(&entitlement.ConsumeUserEntitlementParams{
        Namespace:     namespace,
        UserID:        userID,
        EntitlementID: boosterID,
        Body: &platformclientmodels.EntitlementDecrement{
            UseCount: int32Ptr(1),
        },
    })
}
```

**Booster Consumption:**
```go
func handleBoosterConsumption(userID string, matchEnded bool) error {
    if !matchEnded {
        return nil
    }
    
    booster, err := getActiveBooster(userID)
    if err != nil || booster == nil {
        return err
    }
    
    matchesRemaining := booster.UseCount
    if matchesRemaining != nil && *matchesRemaining > 0 {
        err := decrementBoosterMatches(userID, *booster.ID)
        if err != nil {
            return err
        }
        
        remaining := *matchesRemaining - 1
        if remaining <= 0 {
            showNotification("Speed Booster depleted!")
        } else {
            showNotification(fmt.Sprintf("Speed Booster: %d matches left", remaining))
        }
    }
    
    return nil
}
```

---

#### 3. Badges & Achievements

**Comeback King Badge:**
```go
// Granted when challenge completed
func grantBadge(userID, badgeID string) error {
    return achievementService.UnlockAchievement(&achievement.UnlockAchievementParams{
        Namespace:     namespace,
        UserID:        userID,
        AchievementID: badgeID,
    })
    
    // Display in player profile
    // Visual indicator in main menu
}
```

**Display:**
```
Player: SpeedDemon [Lvl 3] 🥈 👑
                            └─ Comeback King badge
```

---

## Bot Implementation

### Purpose

1. **Testing:** Simulate player behaviors for PoC testing
2. **Gameplay:** Single opponent for all matches
3. **Data Generation:** Create realistic player data at scale

---

### Bot Types

#### 1. Main Game Bot

Already defined earlier in the document - see "Bot Opponent with Dynamic Speed" section. This is the bot used for all matches with balanced difficulty (~8.3 presses/sec).

---

#### 2. Behavioral Bots (For PoC Testing)

**Rage Quit Bot:**
```go
type RageQuitBot struct {
    *Bot
    rageQuitThreshold float64
}

func NewRageQuitBot() *RageQuitBot {
    return &RageQuitBot{
        Bot:               New(),
        rageQuitThreshold: 0.7, // Quit when losing by 70%
    }
}

func (b *RageQuitBot) PlayMatch(ctx context.Context, opponentProgressChan <-chan int) error {
    opponentProgress := 0
    
    for b.pressCount < 100 {
        select {
        case <-ctx.Done():
            return ctx.Err()
        case opponent := <-opponentProgressChan:
            opponentProgress = opponent
        default:
            b.simulatePress()
            
            // Check if losing badly
            deficit := opponentProgress - b.pressCount
            deficitRatio := float64(deficit) / 100.0
            
            if deficitRatio > b.rageQuitThreshold {
                // Rage quit!
                b.logRageQuit()
                return fmt.Errorf("rage quit")
            }
        }
    }
    return nil
}
```

**Session Decline Bot:**
```go
type DecliningPlayerBot struct {
    *Bot
    sessionsThisWeek int
}

func NewDecliningPlayerBot() *DecliningPlayerBot {
    return &DecliningPlayerBot{
        Bot:              New(),
        sessionsThisWeek: 8, // Started active
    }
}

func (b *DecliningPlayerBot) Simulate7Days(ctx context.Context) error {
    // Week 1: Play normally (8 sessions)
    for i := 0; i < 8; i++ {
        if err := b.playSession(ctx); err != nil {
            return err
        }
        time.Sleep(randomDuration(2, 6) * time.Hour)
    }
    
    // Week 2: Decline (3 sessions) - 62% decline
    for i := 0; i < 3; i++ {
        if err := b.playSession(ctx); err != nil {
            return err
        }
        time.Sleep(randomDuration(1, 2) * 24 * time.Hour)
    }
    
    // Week 3: Should trigger intervention
    return nil
}
```

**Losing Streak Bot:**
```go
type LosingStreakBot struct {
    *Bot
    userID string
}

func NewLosingStreakBot(userID string) *LosingStreakBot {
    return &LosingStreakBot{
        Bot:    New(),
        userID: userID,
    }
}

func (b *LosingStreakBot) SimulateLosses(ctx context.Context, count int) error {
    for i := 0; i < count; i++ {
        // Play slowly to guarantee loss
        if err := b.PlayMatch(ctx); err != nil {
            return err
        }
        
        // Log loss
        _ = statisticsService.IncrementUserStatItem(&social.IncrementUserStatItemParams{
            Namespace: namespace,
            StatCode:  "match_losses",
            UserID:    b.userID,
            Inc: &socialclientmodels.StatItemInc{
                Inc: floatPtr(1),
            },
        })
        
        _ = statisticsService.IncrementUserStatItem(&social.IncrementUserStatItemParams{
            Namespace: namespace,
            StatCode:  "current_losing_streak",
            UserID:    b.userID,
            Inc: &socialclientmodels.StatItemInc{
                Inc: floatPtr(1),
            },
        })
        
        time.Sleep(randomDuration(5, 15) * time.Minute)
    }
    
    // After 5 losses, should trigger intervention
  }
}
```

---

#### 3. Data Generation Bot Swarm

**Create Large Dataset:**
```go
// scripts/spawn_bots.go

func spawnBotSwarm(config BotSwarmConfig) error {
    var bots []*Bot
    
    // Create diverse bot population
    for i := 0; i < config.TotalBots; i++ {
        botType := selectRandomBotType()
        bot := createBot(botType, fmt.Sprintf("bot_%d", i))
        bots = append(bots, bot)
    }
    
    // Run all bots concurrently
    var wg sync.WaitGroup
    for _, bot := range bots {
        wg.Add(1)
        go func(b *Bot) {
            defer wg.Done()
            b.Simulate()
        }(bot)
    }
    wg.Wait()
    
    fmt.Printf("Generated data for %d bots\n", config.TotalBots)
    return nil
}

// Usage:
spawnBotSwarm(BotSwarmConfig{
    TotalBots: 100,
    Distribution: map[string]int{
    rageQuit: 20,      // 20 bots that rage quit
    declining: 30,      // 30 bots with session decline
    losing: 25,         // 25 bots with losing streaks
    normal: 25          // 25 normal bots (control)
  },
  durationDays: 14
});
```

---

## Development Phases

### Single Week Timeline (7 Days)

**Total: 7 working days (1 week) for complete single-player game**

---

### Days 1-2: Core Game Foundation

**Deliverables:**
- Project structure created
- AccelByte SDK integrated
- Login flow implemented (IAM)
- Basic terminal UI framework
- Press detection and counting
- Match timer

**Tasks:**
```
✅ Initialize Go project (go mod init)
✅ Install dependencies (Bubble Tea, Lip Gloss, AccelByte Go SDK)
✅ Create main menu structure
✅ Implement IAM login
✅ Test login with AccelByte accounts
✅ Build keyboard input handling
✅ Implement press counter
✅ Create match timer (countdown from 0 to 100)
```

---

### Days 3-4: Bot Implementation & Game Logic

**Deliverables:**
- Dynamic bot with 3 difficulty levels
- Bot with natural speed variance
- Match result screens
- Victory/defeat logic
- Bot difficulty selection UI

**Tasks:**
```
✅ Implement Bot class with natural speed
✅ Add variance, fatigue, burst behavior
✅ Create Easy/Medium/Hard difficulties
✅ Build bot difficulty selection screen
✅ Implement match result calculations
✅ Create victory/defeat screens
✅ Test bot behaviors
```

---

### Days 5-6: AccelByte Integration & Rewards

**Deliverables:**
- Statistics integration complete
- All stats tracked in AGS
- Player rating system implemented
- Reward checking
- Challenge tracking
- Player profile/stats screen

**Tasks:**
```
✅ Create all required AGS statistics
✅ Update stats after each match
✅ Implement rating calculation
✅ Build statistics display screen
✅ Implement booster entitlement logic
✅ Apply booster effects during match
✅ Poll for pending rewards on login
✅ Display challenge UI
✅ Track challenge progress
✅ Auto-claim on completion
```

---

### Day 7: Testing & Polish

**Deliverables:**
- Bug-free gameplay
- Bot testing validated
- Ready for PoC testing
- Documentation complete

**Tasks:**
```
✅ Unit tests for core logic
✅ End-to-end testing
✅ Test bot behavior
✅ Test reward system
✅ Performance optimization
✅ UI polish
✅ Documentation
✅ Final QA
```

---

### Total Timeline Summary

| Phase | Duration | Key Deliverable |
|-------|----------|-----------------|
| **Core Game** | 2 days | Login + press counting working |
| **Bot & Logic** | 2 days | Playable vs bot |
| **Integration** | 2 days | Stats + rewards fully integrated |
| **Polish** | 1 day | Production-ready |

**Total: 7 days (1 week)**

---

### Why This is Faster Than Multiplayer

**Time Saved:**
- ❌ No Matchmaking V2 setup (2 days saved)
- ❌ No Session Service integration (1 day saved)
- ❌ No Lobby WebSocket (1 day saved)
- ❌ No real-time sync complexity (2 days saved)
- ❌ No network debugging (1 day saved)

**Total Time Saved: 7 days** (2 weeks → 1 week)

---

## Code Structure

### Project Organization (Simplified)

```
ready-set-enter/
├── cmd/
│   └── game/
│       └── main.go              # Entry point
├── internal/
│   ├── config/
│   │   ├── accelbyte.go        # AccelByte configuration
│   │   └── game.go             # Game constants
│   ├── services/
│   │   ├── auth.go             # Authentication (IAM)
│   │   ├── statistics.go       # Statistics tracking
│   │   └── entitlement.go      # Booster/items
│   ├── game/
│   │   ├── match.go            # Match logic
│   │   ├── input.go            # Keyboard handling
│   │   ├── scoring.go          # Score calculation
│   │   ├── rating.go           # Player rating system
│   │   └── booster.go          # Booster effects
│   ├── bot/
│   │   ├── bot.go              # Bot with dynamic speed
│   │   ├── rage_quit_bot.go    # Rage quit behavior (testing)
│   │   ├── declining_bot.go    # Session decline (testing)
│   │   └── bot_manager.go      # Bot spawning for testing
│   └── ui/
│       ├── menu.go             # Main menu (Bubble Tea model)
│       ├── match_screen.go     # Live match UI (Bubble Tea model)
│       ├── result_screen.go    # Match results (Bubble Tea model)
│       ├── stats_screen.go     # Player stats (Bubble Tea model)
│       ├── reward_screen.go    # Challenge notifications (Bubble Tea model)
│       ├── styles.go           # Lip Gloss styling
│       └── components.go       # Reusable UI components
├── pkg/
│   └── utils/
│       ├── logger.go           # Logging
│       └── helpers.go          # Utility functions
├── scripts/
│   ├── spawn_bots.go           # Generate test data
│   └── reset_stats.go          # Reset weekly stats
├── go.mod
├── go.sum
├── .env.example                # Environment variables
└── README.md
```

**Note:** Much simpler than multiplayer - no matchmaking, session, or lobby components!

**Build Commands:**
```bash
# Development
go run cmd/game/main.go

# Build for current platform
go build -o ready-set-enter cmd/game/main.go

# Build for all platforms
GOOS=windows GOARCH=amd64 go build -o ready-set-enter.exe cmd/game/main.go
GOOS=darwin GOARCH=amd64 go build -o ready-set-enter-mac cmd/game/main.go
GOOS=linux GOARCH=amd64 go build -o ready-set-enter-linux cmd/game/main.go
```

**Main Entry Point (Bubble Tea):**
```go
// cmd/game/main.go
package main

import (
    "fmt"
    "os"
    
    tea "github.com/charmbracelet/bubbletea"
    "github.com/your-org/ready-set-enter/internal/ui"
    "github.com/your-org/ready-set-enter/internal/services"
)

func main() {
    // Initialize AccelByte services
    if err := services.InitializeAccelByte(); err != nil {
        fmt.Printf("Error initializing AccelByte: %v\n", err)
        os.Exit(1)
    }
    
    // Create initial model (main menu)
    initialModel := ui.NewMenuModel()
    
    // Start Bubble Tea program
    p := tea.NewProgram(
        initialModel,
        tea.WithAltScreen(),       // Use alternate screen buffer
        tea.WithMouseCellMotion(), // Enable mouse support
    )
    
    // Run the program
    if _, err := p.Run(); err != nil {
        fmt.Printf("Error running game: %v\n", err)
        os.Exit(1)
    }
}
```

---

## Testing Strategy

### Unit Tests

**Test Coverage:**
- Rating calculation
- Booster application
- Bot behavior
- Statistics updates
- Match completion logic

**Example Test:**
```go
func TestMatch(t *testing.T) {
    t.Run("should apply 1.5x booster correctly", func(t *testing.T) {
        player := &Player{
            ID: "player1",
            ActiveBooster: &Booster{Type: "speed_booster"},
        }
        bot := NewBot()
        match := NewMatch(player, bot)
        
        pressValue := match.ApplyBoosterEffect(1.0)
        
        assert.Equal(t, 1.5, pressValue)
    })
    
    t.Run("should determine winner correctly", func(t *testing.T) {
        player := &Player{ID: "player1", PressCount: 100}
        bot := &Bot{PressCount: 89}
        match := NewMatch(player, bot)
        
        result := match.End()
        
        assert.Equal(t, "player1", result.Winner)
        assert.True(t, result.PlayerWon)
    })
}
```

---

### Integration Tests

**Test Scenarios:**
1. Complete match flow (login → play → results)
2. Rage quit detection
3. Challenge tracking and completion
4. Booster consumption

---

### PoC Testing

**Automated Bot Testing:**
```bash
# Generate 100 players with diverse behaviors
go run scripts/spawn_bots.go --count=100 --duration=14days

# Expected outcomes:
# - 20 players trigger rage quit intervention
# - 30 players trigger session decline intervention
# - 25 players trigger losing streak intervention
# - 25 players remain normal (no intervention triggered)
```

---

## Deployment

### Local Development

```bash
# Clone repository
git clone https://github.com/your-org/ready-set-enter.git
cd ready-set-enter

# Install Go dependencies
go mod download

# Configure AccelByte credentials
cp .env.example .env
# Edit .env with your AccelByte credentials

# Run game
go run cmd/game/main.go
```

---

### Distribution

**Option 1: Binary Releases (Recommended)**
```bash
# Build for all platforms
make build-all

# Or manually:
GOOS=windows GOARCH=amd64 go build -o ready-set-enter.exe cmd/game/main.go
GOOS=darwin GOARCH=amd64 go build -o ready-set-enter-mac cmd/game/main.go
GOOS=linux GOARCH=amd64 go build -o ready-set-enter-linux cmd/game/main.go

# Create releases on GitHub
# Players download and run - NO installation required!
```

**Why This is Perfect:**
- ✅ Single binary - no dependencies
- ✅ No runtime installation (Node.js, Python, etc.)
- ✅ Cross-platform from one build machine
- ✅ Small file size (~10-15MB)
- ✅ Just download and run

**Player Setup:**
```bash
# Windows
download ready-set-enter.exe
double-click to run

# Mac
download ready-set-enter-mac
chmod +x ready-set-enter-mac
./ready-set-enter-mac

# Linux
download ready-set-enter-linux
chmod +x ready-set-enter-linux
./ready-set-enter-linux
```

**Option 2: Git Repository (For Developers)**
```bash
git clone https://github.com/your-org/ready-set-enter.git
cd ready-set-enter
go mod download
go run cmd/game/main.go
```

---

## Appendix

### Environment Variables

```bash
# .env
ACCELBYTE_BASE_URL=https://demo.accelbyte.io
ACCELBYTE_CLIENT_ID=your_client_id
ACCELBYTE_CLIENT_SECRET=your_client_secret
ACCELBYTE_NAMESPACE=your_namespace

# Optional
LOG_LEVEL=info
MATCH_TIMEOUT=60
```

---

### FAQ

**Q: Can I play offline?**  
A: No, the game requires AccelByte services for authentication and stats tracking.

**Q: Is this multiplayer?**  
A: No, it's single-player only. You always play against a bot opponent.

**Q: How are boosters tracked?**  
A: Via AccelByte Entitlement Service, persistent across sessions.

**Q: Can I create a custom account?**  
A: Yes, via AccelByte IAM registration. All players must have real accounts for churn tracking.

**Q: Why no guest login?**  
A: The PoC requires persistent user tracking for churn detection. Guest accounts can't be tracked consistently.

**Q: How difficult is the bot?**  
A: Balanced at ~8.3 presses/sec. Most players have a 45-55% win rate.

---

## Next Phase After PoC

### Features for Production Version (If PoC Succeeds)

**1. Telemetry Integration**

Currently excluded from PoC to keep scope minimal. Performance metrics should go to a proper telemetry system:

**Telemetry Events to Track:**
```json
{
  "event": "match_completed",
  "timestamp": "2024-01-15T10:30:00Z",
  "user_id": "user123",
  "match_id": "match456",
  "performance": {
    "total_presses": 100,
    "avg_presses_per_second": 2.34,
    "best_time_to_100": 42.8,
    "press_speed_distribution": [1.2, 2.5, 2.8, 2.1, ...],
    "reaction_times": [120, 115, 118, ...],
    "input_errors": 2
  },
  "match_details": {
    "duration": 42.8,
    "result": "win",
    "bot_difficulty": "medium",
    "booster_active": true
  }
}
```

**Why Telemetry Instead of Statistics Service:**
- ✅ **High volume data:** Telemetry handles millions of events
- ✅ **Detailed analytics:** Track every press, not just aggregates
- ✅ **Separate concerns:** Statistics Service for churn signals only
- ✅ **Performance:** Async event logging doesn't block gameplay
- ✅ **Flexible querying:** Use tools like BigQuery, DataDog, or Splunk

**Recommended Telemetry Services:**
- AccelByte Game Telemetry (if available)
- Google Analytics for Gaming
- AWS CloudWatch / Kinesis
- DataDog
- Custom ELK stack (Elasticsearch, Logstash, Kibana)

**Implementation Example:**
```go
// After PoC, add telemetry client
import "github.com/your-org/telemetry"

func (m *Match) End() {
    // Update Statistics Service (churn signals only)
    statsService.UpdateMatchStats(ctx, userID, match)
    
    // Send detailed performance to Telemetry
    telemetry.TrackEvent("match_completed", map[string]interface{}{
        "user_id": userID,
        "total_presses": match.PlayerStats.TotalPresses,
        "avg_press_speed": match.PlayerStats.AverageSpeed,
        "best_time": match.PlayerStats.BestTime,
        "press_distribution": match.PlayerStats.PressTimestamps,
        "duration": match.Duration.Seconds(),
        "result": match.Winner,
    })
}
```

---

**2. Advanced Churn Detection**

After validating basic signals in PoC, add:

- **Machine Learning Models:** Predict churn probability (0-100%)
- **Behavioral Clustering:** Segment players by play patterns
- **Time-series Analysis:** Detect gradual engagement decline
- **Sentiment Analysis:** Parse rage quit patterns (e.g., multiple attempts to close)
- **Social Network Effects:** Track friend churn correlation

---

**3. Enhanced Reward System**

Expand beyond simple conditional rewards:

- **Multiple reward types:** Coins, boosters, cosmetics, badges
- **Dynamic reward sizing:** Adjust based on player value (whales vs F2P)
- **Personalized challenges:** Match difficulty to player skill
- **Reward scheduling:** Time-limited offers, weekend bonuses
- **Reward A/B testing:** Test different reward amounts

---

**4. Multiplayer Mode**

If single-player PoC succeeds, validate with real competitive gameplay:

- **Real-time matchmaking:** Match players by skill rating
- **Ranked mode:** Competitive ladder with seasons
- **Friend challenges:** Direct 1v1 invites
- **Tournaments:** Weekly competitions with prizes
- **Spectator mode:** Watch top players compete

---

**5. Production Features**

Additional features for full launch:

- **Cross-platform progression:** Sync stats across devices
- **Achievements system:** Unlock badges and titles
- **Leaderboards:** Global, regional, friend rankings
- **Cosmetic customization:** Themes, sound packs, animations
- **Social features:** Friend lists, chat, clubs
- **Battle pass:** Seasonal progression with exclusive rewards
- **In-game store:** Purchase boosters and cosmetics
- **Admin dashboard:** Monitor churn metrics in real-time

---

### Migration Path from PoC to Production

**Phase 1: PoC (Current)**
- ✅ Statistics Service for churn signals only
- ✅ 3 churn signals (rage quit, session decline, losing streak)
- ✅ 1 reward type (conditional challenge)
- ✅ Simple A/B test
- ✅ Bot testing

**Phase 2: Enhanced PoC (If successful)**
- ➕ Add telemetry integration
- ➕ Migrate performance metrics to telemetry
- ➕ Keep Statistics Service lean (churn signals only)
- ➕ Add more sophisticated churn models
- ➕ Expand reward types

**Phase 3: Production (If Phase 2 validates)**
- ➕ Multiplayer mode
- ➕ Advanced features (leaderboards, achievements, etc.)
- ➕ Full monitoring and analytics
- ➕ Pluggable architecture for multi-game support
- ➕ Scale to millions of players

---

### Key Principle: Start Simple, Prove Value, Then Expand

**PoC Focus:**
- Minimal stats (churn detection only)
- Single-player (no networking complexity)
- Simple rewards (prove the concept)
- Fast iteration (1-week game development)

**Production Evolution:**
- Add telemetry for detailed analytics
- Keep Statistics Service focused on churn
- Expand when PoC proves ROI
- Don't over-engineer upfront

---

**END OF DOCUMENT**

---

*This game is designed specifically for PoC testing of the Anti-Churn Reward System. Keep it simple, focus on generating quality test data.*
