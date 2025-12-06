# Ready, Set, Enter! - Concrete Implementation Plan

**Version:** 2.0  
**Date:** December 5, 2025  
**Target:** Single-player terminal game with AccelByte integration  
**UI Framework:** Bubble Tea (Elm-inspired architecture)

---

## Table of Contents

1. [Project Overview](#project-overview)
2. [Technology Stack](#technology-stack)
3. [Project Structure](#project-structure)
4. [AccelByte SDK Integration Pattern](#accelbyte-sdk-integration-pattern)
5. [Core Data Models](#core-data-models)
6. [Service Layer Architecture](#service-layer-architecture)
7. [Game Logic Implementation](#game-logic-implementation)
8. [Terminal UI Implementation](#terminal-ui-implementation)
9. [Bot Implementation](#bot-implementation)
10. [Development Workflow](#development-workflow)

---

## Project Overview

### Architecture Pattern

**Clean Architecture with Bubble Tea (Model-Update-View):**
```
┌─────────────────────────────────────────┐
│         CLI Entry Point (main)          │
│         tea.NewProgram(model)           │
└─────────────────┬───────────────────────┘
                  │
┌─────────────────▼───────────────────────┐
│    UI Layer (Bubble Tea MVU Pattern)    │
│                                         │
│  Model: State + Data                    │
│  Update: State Transitions + Logic      │
│  View: Render UI (lipgloss styling)     │
│                                         │
│  Screens: Login, Menu, Match, Result   │
└─────────────────┬───────────────────────┘
                  │
┌─────────────────▼───────────────────────┐
│       Service Layer (AccelByte)         │
│  - Auth, Stats, Entitlement, Reward     │
└─────────────────┬───────────────────────┘
                  │
┌─────────────────▼───────────────────────┐
│    AccelByte Go SDK (v0.85.0)           │
└─────────────────────────────────────────┘
```

### Key Principles

1. **Model-Update-View (MVU):** Elm-inspired architecture with unidirectional data flow
2. **Immutable State:** Models are immutable; updates return new models
3. **Message-Driven:** All state changes happen through messages (tea.Msg)
4. **Composable Components:** Each screen is a self-contained bubble
5. **Type-Safe:** Strong typing for messages and state transitions
6. **Context-Aware:** Use Go contexts for cancellation and timeouts
7. **Error Handling:** Consistent error wrapping and logging

---

## Technology Stack

### Core Dependencies

```go
require (
    // AccelByte SDK
    github.com/AccelByte/accelbyte-go-sdk v0.85.0
    
    // Terminal UI - Bubble Tea Stack
    github.com/charmbracelet/bubbletea v0.25.0  // TUI framework
    github.com/charmbracelet/lipgloss v0.9.1    // Styling
    github.com/charmbracelet/bubbles v0.18.0    // Common components
    
    // Utilities
    github.com/google/uuid v1.5.0       // For generating IDs
    github.com/sirupsen/logrus v1.9.3   // Logging
)
```

### SDK Modules We'll Use

```go
import (
    // IAM Service (Authentication)
    "github.com/AccelByte/accelbyte-go-sdk/iam-sdk/pkg/iamclient"
    "github.com/AccelByte/accelbyte-go-sdk/iam-sdk/pkg/iamclient/o_auth2_0"
    "github.com/AccelByte/accelbyte-go-sdk/iam-sdk/pkg/iamclientmodels"
    
    // Social Service (Statistics)
    "github.com/AccelByte/accelbyte-go-sdk/social-sdk/pkg/socialclient"
    "github.com/AccelByte/accelbyte-go-sdk/social-sdk/pkg/socialclient/user_statistic"
    "github.com/AccelByte/accelbyte-go-sdk/social-sdk/pkg/socialclientmodels"
    
    // Platform Service (Entitlements, Rewards)
    "github.com/AccelByte/accelbyte-go-sdk/platform-sdk/pkg/platformclient"
    "github.com/AccelByte/accelbyte-go-sdk/platform-sdk/pkg/platformclient/entitlement"
    "github.com/AccelByte/accelbyte-go-sdk/platform-sdk/pkg/platformclient/reward"
    "github.com/AccelByte/accelbyte-go-sdk/platform-sdk/pkg/platformclientmodels"
    
    // Services API (Authentication helpers)
    "github.com/AccelByte/accelbyte-go-sdk/services-api/pkg/factory"
    "github.com/AccelByte/accelbyte-go-sdk/services-api/pkg/service/iam"
)
```

---

## Project Structure

```
ready-set-enter/
├── cmd/
│   └── game/
│       └── main.go                      # Application entry point
│
├── internal/
│   ├── config/
│   │   ├── config.go                    # Configuration struct & loader
│   │   └── constants.go                 # Game constants
│   │
│   ├── models/
│   │   ├── player.go                    # Player data model
│   │   ├── match.go                     # Match data model
│   │   ├── bot.go                       # Bot data model
│   │   └── stats.go                     # Statistics data model
│   │
│   ├── services/
│   │   ├── auth_service.go              # Authentication service
│   │   ├── stats_service.go             # Statistics service
│   │   ├── entitlement_service.go       # Booster/items service
│   │   └── reward_service.go            # Reward checking service
│   │
│   ├── game/
│   │   ├── match_controller.go          # Match orchestration
│   │   ├── input_handler.go             # Keyboard input processing
│   │   ├── score_calculator.go          # Score & rating calculation
│   │   └── messages.go                  # Bubble Tea messages
│   │
│   ├── bot/
│   │   ├── bot.go                       # Main bot implementation
│   │   ├── bot_config.go                # Bot configuration
│   │   └── behavioral_bots.go           # Testing bots (rage quit, etc.)
│   │
│   └── ui/
│       ├── model.go                     # Root Bubble Tea model
│       ├── update.go                    # Root Update function
│       ├── view.go                      # Root View function
│       ├── login.go                     # Login screen model
│       ├── menu.go                      # Main menu screen model
│       ├── match.go                     # Match screen model
│       ├── result.go                    # Result screen model
│       ├── stats.go                     # Statistics screen model
│       └── styles.go                    # Lipgloss styling
│
├── pkg/
│   └── utils/
│       ├── logger.go                    # Logging utilities
│       └── errors.go                    # Error helpers
│
├── scripts/
│   └── test_bots/
│       └── spawn_bots.go                # Bot spawning for testing
│
├── .env.example                         # Environment variables template
├── go.mod
├── go.sum
└── README.md
```

---

## AccelByte SDK Integration Pattern

### Configuration Setup

**File: `internal/config/config.go`**

**Struct Fields:**
- AccelByte config: BaseURL, ClientID, ClientSecret, Namespace
- Game config: BotSpeed, TargetScore
- Logging: LogLevel

**Loading Strategy:**
- Use environment variables (AB_BASE_URL, AB_CLIENT_ID, AB_CLIENT_SECRET, AB_NAMESPACE)
- Provide defaults for game-specific values
- Validate all required fields are present

### SDK Initialization Pattern

**File: `internal/services/auth_service.go`**

**AuthService Structure:**
- Wraps `iam.OAuth20Service` from SDK
- Contains embedded structs implementing repository interfaces (configRepo, tokenRepo)
- Direct, simple implementation without external interfaces
- Stores reference to game config

**Initialization Steps:**
1. Create AuthService with game config
2. Initialize embedded configRepo with AccelByte credentials
3. Initialize embedded tokenRepo for storing access tokens
4. Create OAuth20Service with factory.NewIamClient()
5. All repositories are self-contained within the service

**Key Methods:**
- `Login(ctx, username, password)` - Uses LoginWithContext from SDK
- `GetUserID(ctx)` - Parse access token to extract user ID (from Sub field)
- `Logout(ctx)` - Call SDK Logout method
- `IsAuthenticated()` - Check if valid token exists

---

## Core Data Models

### Player Model

**File: `internal/models/player.go`**

**Player Struct:**
- Identity: UserID, Username, DisplayName
- Progression: Level (1-5 based on wins), TotalWins, TotalLosses, Rating, Rank
- Session: Current session stats (matches played, wins/losses today, streak)
- Booster: Reference to active booster (nullable)

**Rank Constants:**
- Bronze: 0-999 rating
- Silver: 1000-1499
- Gold: 1500-1999
- Platinum: 2000-2499
- Diamond: 2500+

**SessionStats Struct:**
- Track in-memory session data
- MatchesPlayed, WinsToday, LossesToday
- CurrentStreak (positive = wins, negative = losses)

**Booster Struct:**
- EntitlementID (from AccelByte)
- Type (BoosterType enum)
- Multiplier (e.g., 1.5x)
- MatchesLeft (decremented after each match)
- ExpiresAt (time-based expiry)

**Helper Methods:**
- `WinRate()` - Calculate percentage
- `GetRank()` - Determine rank from rating
- `CalculateLevel()` - Determine level from total wins

### Match Model

**File: `internal/models/match.go`**

**Match Struct:**
- IDs: Match ID (UUID), PlayerID, BotID
- Timing: StartTime, EndTime, Duration
- Scores: PlayerScore, BotScore (0-100+)
- Outcome: Winner (enum: PLAYER/BOT/NONE)
- Stats: Detailed performance for player and bot
- Flags: RageQuit (early quit detection), BoosterActive

**MatchWinner Enum:**
- WinnerPlayer, WinnerBot, WinnerNone (for incomplete matches)

**MatchPlayerStats:**
- TotalPresses (count)
- AverageSpeed, FastestSpeed, SlowestSpeed (presses/second)
- Track performance metrics for display

**MatchBotStats:**
- TotalPresses, AverageSpeed
- Simpler than player stats

**Winner Calculation Logic:**
- First to 100 wins
- If both reach 100 simultaneously, higher score wins
- RageQuit = automatic bot win

### Statistics Model

**File: `internal/models/stats.go`**

**StatCode Type:**
String-based constants matching AccelByte stat codes

**Stat Categories (Simplified for Anti-Churn PoC):**

1. **Match Results (needed for losing streak calculation):**
   - `rse-match-wins` - Increment on win, triggers losing streak reset
   - `rse-match-losses` - Increment on loss, triggers losing streak increment
   - `rse-match-played` - Total matches (also added to Daily Cycle)

2. **Churn Signals (Critical for anti-churn system):**
   - `rse-rage-quit` - SIGNAL 1: Rage quit detection (Weekly Cycle, auto-resets Monday)
   - `rse-session-login` - SIGNAL 2: Login frequency (Weekly Cycle, auto-resets Monday)
   - `rse-current-losing-streak` - SIGNAL 3: Losing streak (reset on win)

**Total: 6 stat codes only**

**Note:** Derived/complex state managed in Redis by Extend Event Handler:
- Previous week sessions (calculated from current week before reset)
- Challenge state (active, progress, expiry)
- Intervention cooldowns (timestamps, rate limiting)

**PlayerStatistics Struct:**
- UserID
- Stats map[StatCode]float64
- Provides type-safe access to AGS Statistics only

---

## Service Layer Architecture

### Statistics Service

**File: `internal/services/stats_service.go`**

**StatsService Structure:**
- Wraps social-sdk's JusticeSocialService
- Holds namespace and config
- Provides game-specific stat operations
- Direct struct implementation

**Initialization:**
- Use factory.NewSocialClient() with config
- Store namespace for all operations

**Key Methods:**

1. **GetUserStats(ctx, userID, statCodes)**
   - Use `BulkFetchStatItemsParams` from user_statistic
   - Join stat codes as comma-separated string
   - Parse response into PlayerStatistics model
   - Return map of StatCode -> value

2. **IncrementStat(ctx, userID, statCode, value)**
   - Use `IncStatItemValueParams` for single stat
   - Wrap value in StatItemInc body
   - For simple one-stat updates

3. **UpdateStats(ctx, userID, updates map)**
   - Use `BulkIncUserStatItem1Params` for batch updates
   - Build array of StatItemUpdate with INCREMENT strategy
   - More efficient than multiple IncrementStat calls

4. **UpdateMatchStats(ctx, userID, match)**
   - High-level method to update all stats after match
   - Logic: Determine which stats to update based on match outcome
   - Increment wins/losses, update streaks, track presses, playtime
   - Handle rage quit flag

**Update Strategy:**
- Use INCREMENT for base stat codes (wins, losses, matches_played)
- Stat codes added to Statistic Cycles in AGS Admin Portal auto-reset
- Game increments stat codes, AGS handles time-based resets
- Batch updates when possible for performance

**Hybrid Storage Approach (AGS Statistics + Redis):**

**Why Hybrid?**
- AGS Statistics = Admin visibility (dashboard, monitoring, alerts)
- Redis = Flexible backend logic (complex state, derived values)

**AGS Statistics (Churn Aggregations):**
- `logins_current_week` - Game increments on login, uses Weekly Statistic Cycle
- `rage_quit_count_weekly` - Game increments on rage quit, uses Weekly Statistic Cycle
- Admin dashboard shows real-time churn metrics without custom tooling

**Redis (Extend-Managed State):**
- Session tracking: `thisWeek`, `lastWeek`, `lastReset` (derived from AGS events)
- Challenge state: `active`, `winsNeeded`, `winsCurrent`, `expiresAt`
- Intervention cooldowns: `lastTimestamp`, `cooldownUntil`
- Managed by AccelByte (zero infrastructure setup)

**AGS Statistic Cycles Configuration:**
- Create base stat codes: rse-rage-quit, rse-session-login, rse-match-played
- In AGS Admin Portal:
  - Configure Weekly cycle → Add rse-rage-quit, rse-session-login
  - Configure Daily cycle → Add rse-match-played (creates matches_today)
- Note: Same stat code can be used for both lifetime total AND cycle aggregation
- Extend Event Handler listens to statisticCycleReset events
- Extend captures old values and stores in Redis before reset

### Entitlement Service

**File: `internal/services/entitlement_service.go`**

**EntitlementService Structure:**
- Wraps platform-sdk's JusticePlatformService
- Handles booster entitlements
- Manages use count tracking
- Direct struct implementation

**Initialization:**
- Use factory.NewPlatformClient() with config

**Key Methods:**

1. **GetActiveBooster(ctx, userID)**
   - Use `QueryUserEntitlementsParams` with filters:
     - EntitlementClazz: "ENTITLEMENT"
     - AppType: "GAME"
   - Loop through results to find ACTIVE with UseCount > 0
   - Parse into Booster model
   - Return nil if no active booster

2. **ConsumeBooster(ctx, userID, entitlementID)**
   - Use `DecrementUserEntitlementParams`
   - Decrement UseCount by 1
   - Called after each match completion
   - When UseCount reaches 0, booster expires

**Booster Lifecycle:**
- Granted by anti-churn system or purchased
- Checked before match starts
- Applied during match (1.5x multiplier)
- Consumed after match ends
- Expires when uses depleted

### Granted Items Service

**File: `internal/services/granted_items_service.go`**

**GrantedItemsService Structure:**
- Wraps platform-sdk entitlement endpoints
- Queries for newly granted items from Extend
- Displays notifications to player

**Key Methods:**

1. **CheckNewlyGrantedItems(ctx, userID)**
   - Use `QueryUserEntitlementsParams` to get all ACTIVE entitlements
   - Track last check timestamp in local state/file
   - Compare entitlement `grantedAt` or `createdAt` fields with last check time
   - Return entitlements created after last check timestamp
   - Update last check timestamp after processing

2. **DisplayItemNotifications(items)**
   - Show terminal notification for each new item
   - Display item name, type (booster/currency), quantity
   - Auto-dismiss or require key press to continue

**Simpler Approach - No "Seen" Tracking Needed:**
- Just query all active entitlements on login
- Display count of active boosters and currency balance
- No need to track individual "seen" status
- Player can view items anytime in menu

**Integration with Anti-Churn:**
- Anti-churn Extend app directly grants entitlements (no Reward Service)
- Game polls on login for newly granted items
- Displays notifications for boosters/coins received
- Simpler than Reward Service - just query entitlements

---

## Game Logic Implementation

### Match Controller

**File: `internal/game/match_controller.go`**

**MatchController Structure:**
- Orchestrates entire match lifecycle
- Holds references to: StatsService, EntitlementService, Player, Bot
- Manages match state and timing
- Uses channels for player input and bot updates

**Match Lifecycle:**

1. **Pre-Match:**
   - Check for active booster
   - Create Match model with UUID
   - Show 3-2-1-GO countdown
   - Start timer

2. **During Match:**
   - Listen on input channel for player presses
   - Run bot in goroutine with context
   - Update progress bars in real-time
   - Track performance metrics (speed calculations)
   - Apply booster multiplier if active
   - Check for ESC key (rage quit)

3. **Post-Match:**
   - Calculate winner
   - Calculate performance stats (avg speed, best time)
   - Update AccelByte statistics
   - Consume booster if active
   - Return match result

**Concurrency Pattern:**
- Player input: Main goroutine with keyboard listener
- Bot simulation: Separate goroutine
- UI updates: Ticker-based refresh (60 FPS)
- Cancellation: Context-based for early termination

### Input Handler

**File: `internal/game/input_handler.go`**

**InputHandler Structure:**
- Wraps tcell keyboard event handling
- Provides channel-based input system
- Debounces rapid presses if needed

**Key Responsibilities:**
- Listen for Enter key presses
- Detect ESC for rage quit
- Filter invalid input
- Send events to press channel
- Track press timing for speed calculation

**Press Event Model:**
- Timestamp of press
- Sequence number
- Time since last press (for speed calc)

### Score Calculator

**File: `internal/game/score_calculator.go`**

**ScoreCalculator Functions:**

1. **CalculateRating(playerRating, botRating, won)**
   - Implement Elo rating system
   - K-factor: 32 (standard)
   - Expected score formula: 1 / (1 + 10^((opponent - player)/400))
   - Win: +rating, Loss: -rating
   - Bot rating fixed at 1500

2. **CalculatePerformanceStats(presses []PressEvent)**
   - Average speed: total presses / duration
   - Fastest speed: min time between presses
   - Slowest speed: max time between presses
   - Consistency: standard deviation

3. **ApplyBoosterMultiplier(score, hasBooster)**
   - If booster active: score * 1.5
   - Round to nearest integer
   - Cap at 100 (target score)

### State Manager

**File: `internal/game/state_manager.go`**

**StateManager Structure:**
- Holds current game state
- Manages transitions between screens
- Stores player session data

**Game States:**
- StateLogin (authentication)
- StateMainMenu (menu navigation)
- StateMatchStarting (countdown)
- StateMatchInProgress (gameplay)
- StateMatchResult (showing results)
- StateStats (viewing statistics)
- StateQuit (cleanup)

**State Transitions:**
- Event-driven state machine
- Validate transitions
- Cleanup on state exit
- Persist session data

---

## Terminal UI Implementation with Bubble Tea

### Bubble Tea Architecture Overview

**The MVU Pattern:**
```
┌────────────────────────────────────┐
│           Model (State)            │
│  - Current screen                  │
│  - Player data                     │
│  - Match state                     │
│  - Services (auth, stats, etc.)    │
└────────────────┬───────────────────┘
                 │
                 ▼
┌────────────────────────────────────┐
│        Update (Msg → Model)        │
│  - Handle keyboard input           │
│  - Process async results           │
│  - State transitions               │
│  - Return (new model, command)     │
└────────────────┬───────────────────┘
                 │
                 ▼
┌────────────────────────────────────┐
│         View (Model → UI)          │
│  - Render current screen           │
│  - Apply lipgloss styling          │
│  - Return string for terminal      │
└────────────────────────────────────┘
```

**Why This Is Better:**
- **Predictable:** All state changes go through Update
- **Debuggable:** No hidden callbacks or event handlers
- **Testable:** Pure functions (model in → model out)
- **Composable:** Screens are independent models

---

## Bot Implementation

### Main Bot

**File: `internal/bot/bot.go`**

**Bot Structure:**
- pressCount: Current score (0-100)
- config: BotConfig (speed, variance, fatigue)
- state: inBurst, fatigueLevel
- pressHistory: Timestamps for speed tracking

**Bot Behavior:**

1. **Base Speed:** 120ms between presses (~8.3 press/sec)
2. **Random Variance:** ±15ms per press (realistic variation)
3. **Burst Phases:** 20% chance to enter burst
   - Burst lasts 3-5 presses
   - Speed multiplier: 0.7 (30% faster)
4. **Gradual Fatigue:** Slight slowdown as match progresses
   - +0.5ms every 10 presses
   - Max fatigue: +10ms

**Play Method:**
- Run in goroutine with context
- Loop until score >= 100 or context cancelled
- Calculate delay with variance and fatigue
- Sleep for calculated duration
- Increment score
- Send update via channel

### Bot Configuration

**File: `internal/bot/bot_config.go`**

**BotConfig Struct:**
- BaseSpeed: time.Duration (base delay between presses)
- Variance: float64 (random variation range)
- FatigueRate: float64 (how quickly bot slows)
- BurstChance: float64 (probability of burst)
- BurstDuration: int (presses in burst)
- BurstMultiplier: float64 (speed during burst)

**Balanced Difficulty Config:**
```
BaseSpeed: 120ms
Variance: 15ms
FatigueRate: 0.0005
BurstChance: 0.2
BurstDuration: 3-5 presses
BurstMultiplier: 0.7
```

**Why This Works:**
- Win rate ~45-55% for average players
- Feels realistic (not robotic)
- Challenging but beatable
- Natural variance prevents predictability

### Behavioral Bots (Testing)

**File: `internal/bot/behavioral_bots.go`**

**RageQuitBot:**
- Purpose: Test rage quit detection
- Behavior: Quit when falling behind by 20+ points
- Used to generate churn signals

**SessionDeclineBot:**
- Purpose: Test session decline detection
- Behavior: Play 8 sessions week 1, 3 sessions week 2
- Simulates declining engagement

**LosingStreakBot:**
- Purpose: Test losing streak detection
- Behavior: Intentionally lose 5+ matches in a row
- Triggers anti-churn interventions

**Implementation:**
- Extend base Bot with custom logic
- Override Play method with behavioral patterns
- Used only in scripts/test_bots

---

## Development Workflow

### Phase 1: Foundation & Project Setup

**Goal:** Complete project structure, models, and basic authentication

**Tasks:**

1. **Project Setup**
   - [x] Create directory structure (cmd/, internal/, pkg/)
   - [x] Initialize go.mod with Bubble Tea dependencies (bubbletea, lipgloss, bubbles)
   - [x] Create .env.example with required variables
   - [ ] Set up basic README.md

2. **Configuration & Utilities**
   - [x] Implement `internal/config/config.go` - Load from env vars
   - [x] Implement `internal/config/constants.go` - Game constants
   - [x] Implement `pkg/utils/logger.go` - Logging setup
   - [x] Implement `pkg/utils/errors.go` - Error helpers

3. **Data Models**
   - [x] Implement `internal/models/player.go` - Player struct with methods
   - [x] Implement `internal/models/match.go` - Match struct with methods
   - [x] Implement `internal/models/stats.go` - StatCode constants
   - [x] Implement `internal/models/bot.go` - Bot data model

4. **AccelByte Authentication**
   - [x] Implement `internal/services/auth_service.go`
   - [x] Initialize ConfigRepository, TokenRepository, RefreshTokenRepository
   - [x] Implement Login/Logout/GetUserID methods
   - [x] Test authentication with AccelByte credentials

5. **Bubble Tea UI Foundation**
   - [x] Implement `internal/ui/styles.go` - Lipgloss styling
   - [x] Implement `internal/ui/model.go` - Root model with screen management
   - [x] Implement `internal/ui/update.go` - Root Update function
   - [x] Implement `internal/ui/view.go` - Root View function
   - [x] Implement `internal/game/messages.go` - Custom tea.Msg types
   - [x] Implement `internal/ui/login.go` - Login screen with textinput.Model
   - [x] Test basic navigation and authentication flow

**Validation Checklist:**
- [x] Project structure complete
- [x] All models defined with helper methods
- [x] Configuration loads from environment variables
- [x] Can successfully login with AccelByte credentials
- [x] Token refresh works automatically
- [x] Logging system functional
- [x] Bubble Tea root model initialized
- [x] Login screen working with MVU pattern
- [x] Screen transitions work correctly

**Deliverable:** Solid foundation with Bubble Tea MVU architecture and working authentication

---

### Phase 2: Core Gameplay with Bubble Tea

**Goal:** Playable game with bot opponent using clean MVU architecture

**Tasks:**

1. **Bot Implementation with Messages**
   - [x] Implement `internal/bot/bot_config.go` - BotConfig struct
   - [x] Implement `internal/bot/bot.go` - Bot with dynamic speed
   - [x] Adapt bot to work with tea.Cmd instead of channels
   - [x] Create BotUpdateMsg to send score updates
   - [x] Test bot speed achieves ~45-55% win rate

2. **Game Logic with Bubble Tea Messages**
   - [x] Implement `internal/game/messages.go` - All custom tea.Msg types:
     - LoginSuccessMsg, StartMatchMsg, MatchCompleteMsg
     - BotUpdateMsg, MatchTickMsg, PlayerPressMsg
     - StatsLoadedMsg, BoosterLoadedMsg, ErrorMsg
   - [x] Implement `internal/game/score_calculator.go` - Elo rating, performance stats
   - [x] Refactor `internal/game/match_controller.go` - Return tea.Cmd
   - [x] Remove channel-based input handling (use messages instead)

3. **Bubble Tea Screens Implementation**
   - [x] Implement `internal/ui/menu.go` - MenuModel with Update/View
     - Cursor-based navigation
     - Display player stats
     - Handle "Start Match" selection
   - [x] Implement `internal/ui/match.go` - MatchModel with Update/View
     - Handle PlayerPressMsg (Enter key)
     - Handle BotUpdateMsg (bot score updates)
     - Handle MatchTickMsg (30 FPS ticker)
     - Render progress bars with lipgloss
     - Check win/lose conditions
   - [x] Implement `internal/ui/result.go` - ResultModel with Update/View
     - Display match results
     - Show rating changes
     - "Press Space to continue" back to menu (changed from Enter)

4. **Main Application with tea.Program**
   - [x] Update `cmd/game/main.go`:
     - Initialize root Model
     - Create tea.Program with tea.WithAltScreen()
     - Run program (blocks until quit)
   - [x] Test full flow: Login → Menu → Match → Result → Menu
   - [x] Test keyboard input (Enter, ESC, ctrl+c)
   - [x] Verify state transitions via messages

**Validation Checklist:**
- [x] Can start and complete a match vs bot
- [x] Progress bars update smoothly via MatchTickMsg (30 FPS)
- [x] Winner determined correctly
- [x] Bot feels natural and challenging
- [x] Keyboard input responsive (no lag, no callback issues)
- [x] Screen transitions clean via messages
- [x] No crashes or panics during gameplay
- [x] State is predictable (no hidden state in callbacks)
- [x] Easy to debug (add logging to Update function)
- [x] Performance acceptable on different terminals
- [x] Text input properly handles paste, cursor movement, special keys
- [x] Win rate calculation fixed (returns decimal, not percentage)
- [x] Result screen uses Space instead of Enter to prevent accidental match starts

**Deliverable:** Fully playable game with clean Bubble Tea MVU architecture

---

### Phase 3: AccelByte Integration & Statistics

**Goal:** Full integration with AccelByte services for stats, boosters, and rewards

**Tasks:**

1. **Statistics Service**
   - [x] Implement `internal/services/stats_service.go`
   - [x] Implement GetUserStats with BulkFetchStatItems
   - [x] Implement IncrementStat and UpdateStats methods
   - [x] Implement UpdateMatchStats high-level method
   - [ ] Test stat updates after matches

2. **AccelByte Portal Setup**
   - [x] Create base event stat codes in AccelByte Admin Portal:
     - [x] Match results: rse-match-wins, rse-match-losses, rse-match-played
     - [x] Churn signals: rse-rage-quit, rse-session-login, rse-current-losing-streak
     - [x] **Total: 6 stat codes only (simplified for Anti-Churn PoC)**
   - [x] Configure Statistic Cycles in AGS Admin Portal:
     - [x] Create Weekly Cycle:
       - [x] Add stat codes: rse-rage-quit, rse-session-login
       - [x] Set reset schedule: Weekly, Monday 00:00 UTC
       - [x] Set status: ACTIVE
       - [x] Result: Tracks rage quits and logins per week (auto-reset)
     - [x] Create Daily Cycle:
       - [x] Add stat code: rse-match-played
       - [x] Set reset schedule: Daily, 00:00 UTC
       - [x] Set status: ACTIVE
       - [x] Result: Tracks matches played today (daily auto-reset)
       - [x] Note: Same stat maintains both lifetime total AND daily count
   - [x] Verify cycles emit statisticCycleReset events (for Extend Event Handler)

3. **Match-Stats Integration**
   - [x] Update MatchController to call UpdateMatchStats
   - [x] Track rage quit detection (ESC during match)
   - [x] Update session statistics on login (increment rse-session-login)
   - [x] Fixed to use public endpoints instead of admin endpoints

4. **Entitlement Service**
   - [x] Implement `internal/services/entitlement_service.go`
   - [x] Implement GetActiveBooster method
   - [x] Implement ConsumeBooster method (using public endpoints)
   - [ ] Test booster lifecycle

5. **Booster Integration**
   - [ ] Update MatchController to check booster before match
   - [ ] Apply 1.5x multiplier during match
   - [ ] Update UI to show booster indicator
   - [ ] Consume booster after match completion

6. **Granted Items Service**
   - [ ] Implement `internal/services/granted_items_service.go`
   - [ ] Implement CheckNewlyGrantedItems method (with local timestamp tracking)
   - [ ] Store last check timestamp in config/state file
   - [ ] Display simple notification: "You have X active boosters!"
   - [ ] Show booster count in main menu
   - [ ] No complex "seen" tracking needed - just show current inventory

7. **UI - Statistics Screen**
   - [ ] Implement `internal/ui/stats.go` - Stats display screen
   - [ ] Fetch stats from StatsService
   - [ ] Display match history, performance, ranking, activity
   - [ ] Add to main menu navigation

8. **UI - Result Screen**
   - [ ] Implement `internal/ui/result.go` - Match result screen
   - [ ] Show victory/defeat with final scores
   - [ ] Display performance metrics and rating change
   - [ ] Show streak and win rate updates

**Validation Checklist:**
- [ ] Stats update in AccelByte after each match
- [ ] Can view all stats correctly in AccelByte Admin Portal
- [ ] Win/loss streaks tracked accurately
- [ ] Rage quits increment rage_quit_count_weekly
- [ ] Sessions tracked per week
- [ ] Boosters appear when granted
- [ ] Booster multiplier applies correctly in match
- [ ] Booster uses decremented after match
- [ ] Can view detailed stats in-game
- [ ] Result screen shows accurate information

**Deliverable:** Fully integrated game with AccelByte stats, boosters, and reward system

---

### Phase 4: Polish, Testing & Production Ready

**Goal:** Production-ready game with proper error handling, testing, and documentation

**Tasks:**

1. **State Management**
   - [ ] Implement `internal/game/state_manager.go`
   - [ ] Implement state machine for screen transitions
   - [ ] Add state validation and cleanup
   - [ ] Integrate with UI navigation

2. **Error Handling Improvements**
   - [ ] Add retry logic for network errors (exponential backoff)
   - [ ] Add user-friendly error messages
   - [ ] Implement error modals in UI
   - [ ] Add offline mode detection
   - [ ] Handle token expiration gracefully

3. **Logging & Debugging**
   - [ ] Add structured logging throughout application
   - [ ] Log all AccelByte API calls
   - [ ] Add debug mode with verbose output
   - [ ] Implement log levels (info, warn, error, debug)

4. **Testing - Unit Tests**
   - [ ] Test bot speed calculations and variance
   - [ ] Test score calculator (Elo rating)
   - [ ] Test model helper methods (WinRate, GetRank, etc.)
   - [ ] Test booster multiplier logic

5. **Testing - Integration Tests**
   - [ ] Test AuthService flow with test credentials
   - [ ] Test StatsService operations end-to-end
   - [ ] Test EntitlementService operations end-to-end
   - [ ] Use AccelByte test environment for integration tests

6. **Testing - Manual Testing**
   - [ ] Play 20+ matches end-to-end
   - [ ] Test rage quit scenario
   - [ ] Verify all stats in AccelByte portal
   - [ ] Test booster application and consumption
   - [ ] Test with slow network (latency simulation)
   - [ ] Test with network disconnect/reconnect

7. **UI Polish**
   - [ ] Add loading indicators for API calls
   - [ ] Add animations for transitions
   - [ ] Improve progress bar rendering
   - [ ] Add help text / instructions
   - [ ] Improve main menu layout
   - [ ] Add keyboard shortcut hints

8. **Behavioral Bots (Testing Only)**
   - [ ] Implement `internal/bot/behavioral_bots.go`
   - [ ] Implement RageQuitBot for testing
   - [ ] Implement SessionDeclineBot for testing
   - [ ] Implement LosingStreakBot for testing
   - [ ] Create `scripts/test_bots/spawn_bots.go`
   - [ ] Test spawning multiple bots to generate data

9. **Documentation**
   - [ ] Update README.md with:
     - [ ] Project description
     - [ ] Installation instructions
     - [ ] Configuration setup
     - [ ] How to run
     - [ ] How to build for distribution
   - [ ] Create CONTRIBUTING.md
   - [ ] Add code comments for complex logic
   - [ ] Document AccelByte setup requirements

10. **Build & Distribution**
    - [ ] Test build for Linux
    - [ ] Test build for macOS
    - [ ] Test build for Windows
    - [ ] Create build script
    - [ ] Test cross-platform execution
    - [ ] Optimize binary size

**Validation Checklist:**
- [ ] All unit tests pass
- [ ] All integration tests pass
- [ ] No crashes or panics in extended play
- [ ] Graceful error handling for all scenarios
- [ ] Stats accurately tracked across multiple sessions
- [ ] UI responsive and clear on all terminals
- [ ] Builds successfully for all platforms
- [ ] Documentation complete and accurate
- [ ] Ready for production deployment

**Deliverable:** Production-ready game with full testing, documentation, and multi-platform builds

### Testing Strategy

**Unit Tests:**
- Bot speed calculations
- Score calculator (Elo, stats)
- Model methods (WinRate, GetRank, etc.)

**Integration Tests:**
- AuthService with real AccelByte credentials (test environment)
- StatsService operations end-to-end
- EntitlementService operations end-to-end

**Manual Testing:**
- Play 10+ matches
- Test rage quit
- Verify stats in AccelByte portal
- Test booster application
- Check reward polling

**Bot Testing:**
- Spawn test bots to generate data
- Verify churn signals appear
- Test anti-churn system integration

### Environment Setup

**Required Environment Variables:**
```
AB_BASE_URL=https://demo.accelbyte.io
AB_CLIENT_ID=<your_client_id>
AB_CLIENT_SECRET=<your_client_secret>
AB_NAMESPACE=<your_namespace>
```

**Development Tools:**
- Go 1.23+
- AccelByte Admin Portal access
- Terminal with Unicode support

### Build & Run

**Development:**
```bash
go run cmd/game/main.go
```

**Build for Distribution:**
```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o ready-set-enter cmd/game/main.go

# macOS
GOOS=darwin GOARCH=amd64 go build -o ready-set-enter-mac cmd/game/main.go

# Windows
GOOS=windows GOARCH=amd64 go build -o ready-set-enter.exe cmd/game/main.go
```

---

## Implementation Notes

### SDK Usage Patterns

1. **Always use context:** Pass context to all service methods for cancellation
2. **Error wrapping:** Wrap SDK errors with context-specific messages
3. **Token management:** Let SDK handle token refresh automatically
4. **Batch operations:** Use bulk endpoints when updating multiple stats
5. **Timeouts:** Set reasonable timeouts for API calls (5-10 seconds)

### Performance Considerations

1. **UI Updates:** Cap at 60 FPS to prevent flickering
2. **Stats Updates:** Batch after match, not during
3. **Booster Checks:** Cache result, don't query every match
4. **Concurrent Operations:** Use goroutines for bot and input handling

### Error Handling Strategy

1. **Network Errors:** Retry with exponential backoff
2. **Auth Errors:** Force re-login, clear token
3. **Stat Errors:** Log but continue (don't block gameplay)
4. **UI Errors:** Show modal with error, allow retry

### Future Enhancements (Out of Scope)

- Multiple bot difficulties
- Player vs Player mode (with matchmaking)
- Leaderboards
- Achievements system
- Custom key bindings
- Sound effects (terminal beep)
- Match replay system

---

**End of Implementation Plan**
