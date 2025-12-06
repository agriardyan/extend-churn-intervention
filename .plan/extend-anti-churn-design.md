# Anti-Churn Extend App Design Document

**Version:** 1.0  
**App Type:** Extend Event Handler  
**Platform:** AccelByte Gaming Services  
**Language:** Go  
**Purpose:** Automated churn detection and intervention for Ready, Set, Enter! game

---

## Table of Contents

1. [Overview](#overview)
2. [Architecture](#architecture)
3. [Event Subscriptions](#event-subscriptions)
4. [State Management](#state-management)
5. [Churn Detection Logic](#churn-detection-logic)
6. [Intervention System](#intervention-system)
7. [API Integration](#api-integration)
8. [Deployment](#deployment)

---

## Overview

### Purpose

This Extend Event Handler app automatically detects at-risk players and triggers interventions through direct entitlement grants. It listens to AGS events (triggered when the game client writes statistics), maintains player state in Redis, and orchestrates the challenge-reward system.

### Key Responsibilities

- **Listen to AGS events** (`oauthTokenGenerated` for logins, `statItemUpdated` for game stats)
- **Detect churn signals** (rage quits, session frequency decline, losing streaks)
- **Trigger interventions** (create challenge state in Redis, grant entitlements directly)
- **Track challenge progress** (monitor `rse-match-wins` stat, grant rewards on completion)
- **Enforce rate limits** (prevent intervention spam via cooldowns in Redis)

### Data Flow

```
Game Client → AGS Statistics (writes stats)
                    ↓
            AGS publishes statItemUpdated event
                    ↓
         Extend Event Handler (listens)
                    ↓
    Redis (reads/writes challenge state)
                    ↓
   AGS Platform API (grants entitlements)
                    ↓
  Game Client (queries entitlements on login)
```

### Technology Stack

```
Go 1.21+
├── AccelByte Go SDK (official)
├── gRPC (event handler protocol)
├── Redis Client (go-redis/redis)
└── Protocol Buffers (event descriptors)
```

---

## Architecture

### High-Level Structure

```
Game Client (Ready, Set, Enter!)
    │
    ├─► IAM: Login
    ├─► Statistics: Write stats (rage quits, wins, losses)
    └─► Entitlement: Query rewards
    
                ↓
                
AccelByte Gaming Services (AGS)
    │
    ├─► Statistics Service → Publishes events (Kafka)
    └─► Platform Service
    
                ↓
                
Extend Event Handler (Go)
    │
    ├─► Listen: oauthTokenGenerated, statItemUpdated
    ├─► Detect: Rage quits, session decline, losing streaks
    ├─► Store: Challenge state in Redis
    └─► Grant: Entitlements via Platform API
    
                ↓
                
Redis (Managed by AccelByte)
    └─► Store: Session tracking, challenges, cooldowns
```

### Component Overview

**gRPC Event Handlers**
- Receive events from AGS via Kafka Connect
- Deserialize Protobuf messages
- Route to appropriate detection logic

**Churn Detection Engine**
- Evaluates churn signals based on events and state
- Reads AGS Statistics for aggregated data
- Reads Redis for previous week sessions and derived state

**Intervention Orchestrator**
- Checks rate limits (cooldowns in Redis)
- Creates challenge state in Redis
- Grants entitlements via AGS Platform API
- Tracks challenge progress

**Redis State Store**
- Player churn state (session tracking, streaks)
- Active challenges (wins needed, progress, expiry)
- Intervention history (cooldowns, timestamps)

---

## Event Subscriptions

### Events to Listen

The Extend app subscribes to the following AGS events:

#### 1. `oauthTokenGenerated` (IAM Service)
**Purpose:** Track login sessions for frequency decline detection

**Trigger:** Player launches CLI and authenticates

**Payload Fields:**
- `userId` - Player ID
- `namespace` - Game namespace
- `timestamp` - Login time

**Handler Pattern (following template):**
```go
// In pkg/service/loginHandler.go
type LoginHandler struct {
    pb.UnimplementedUserAuthenticationUserLoggedInServiceServer
    redisClient *redis.Client
    namespace   string
}

func NewLoginHandler(
    configRepo repository.ConfigRepository,
    tokenRepo repository.TokenRepository,
    redisClient *redis.Client,
    namespace string,
) *LoginHandler {
    return &LoginHandler{
        redisClient: redisClient,
        namespace:   namespace,
    }
}

func (h *LoginHandler) OnMessage(ctx context.Context, msg *pb.UserLoggedIn) (*emptypb.Empty, error) {
    scope := common.GetScopeFromContext(ctx, "LoginHandler.OnMessage")
    defer scope.Finish()
    
    logrus.Infof("received login event: userId=%s", msg.UserId)
    
    // Increment this week's session count in Redis
    // Check if it's a new week (reset if needed)
    // Compare with last week's count
    // Trigger intervention if decline >= 50%
    
    if err := h.checkSessionDecline(ctx, msg.UserId); err != nil {
        return &emptypb.Empty{}, status.Errorf(codes.Internal, 
            "failed to check session decline: %v", err)
    }
    
    return &emptypb.Empty{}, nil
}
```

---

#### 2. `statItemUpdated` (Statistics Service)
**Purpose:** Track match results, rage quits, losing streaks

**Trigger:** Game updates statistics after match or rage quit

**Payload Fields:**
- `userId` - Player ID
- `statCode` - Stat name (e.g., `rse-match-wins`, `rse-rage-quit`)
- `value` - Updated value
- `additionalData` - Extra context

**Handler Pattern (following template):**
```go
// In pkg/service/statHandler.go
type StatHandler struct {
    pb.UnimplementedStatItemUpdatedServiceServer
    fulfillment platform.FulfillmentService
    statistics  social.StatisticService
    redisClient *redis.Client
    namespace   string
}

func NewStatHandler(
    configRepo repository.ConfigRepository,
    tokenRepo repository.TokenRepository,
    redisClient *redis.Client,
    namespace string,
) *StatHandler {
    // Initialize services...
}

func (h *StatHandler) OnMessage(ctx context.Context, msg *pb.StatItemUpdated) (*emptypb.Empty, error) {
    scope := common.GetScopeFromContext(ctx, "StatHandler.OnMessage")
    defer scope.Finish()
    
    logrus.Infof("received stat update: userId=%s statCode=%s value=%v", 
        msg.UserId, msg.StatCode, msg.Value)
    
    switch msg.StatCode {
    case "rse-rage-quit":
        // Read weekly count from AGS Statistics
        // Trigger intervention if count >= 3
        if err := h.handleRageQuit(ctx, msg.UserId); err != nil {
            return &emptypb.Empty{}, status.Errorf(codes.Internal, 
                "failed to handle rage quit: %v", err)
        }
        
    case "rse-match-wins":
        // Update challenge progress if active
        // Grant reward if challenge completed
        if err := h.handleMatchWin(ctx, msg.UserId, int(msg.Value)); err != nil {
            return &emptypb.Empty{}, status.Errorf(codes.Internal, 
                "failed to handle match win: %v", err)
        }
        
    case "rse-current-losing-streak":
        // Check losing streak from event value
        // Trigger intervention if streak >= 5
        if err := h.handleLosingStreak(ctx, msg.UserId, int(msg.Value)); err != nil {
            return &emptypb.Empty{}, status.Errorf(codes.Internal, 
                "failed to handle losing streak: %v", err)
        }
    }
    
    return &emptypb.Empty{}, nil
}
```

---

### Event Subscription Configuration

Events are subscribed via AGS Admin Portal when deploying the Extend app. The app listens to:

- Topic: `iam.oauthTokenGenerated`
- Topic: `social.statItemUpdated`

No code configuration needed - managed through AGS Admin Portal during deployment.

---

## State Management

### Redis Schema

All player state is stored in Redis using the following key structure:

**Key Pattern:** `extend_anti_churn:churn_state:{userId}`

**Value Structure (JSON):**
```json
{
  "sessions": {
    "thisWeek": 5,
    "lastWeek": 10,
    "lastReset": "2025-12-02T00:00:00Z"
  },
  "streaks": {
    "losing": 3,
    "lastMatchResult": "loss"
  },
  "challenge": {
    "active": true,
    "winsNeeded": 3,
    "winsCurrent": 1,
    "winsAtStart": 45,
    "expiresAt": "2025-12-09T10:30:00Z",
    "triggerReason": "rage_quit"
  },
  "intervention": {
    "lastTimestamp": "2025-12-06T10:30:00Z",
    "cooldownUntil": "2025-12-13T10:30:00Z",
    "totalTriggered": 2
  }
}
```

### State Operations

**Read State:**
```go
func getChurnState(ctx context.Context, userID string) (*ChurnState, error) {
    key := fmt.Sprintf("extend_anti_churn:churn_state:%s", userID)
    data, err := redisClient.Get(ctx, key).Result()
    if err == redis.Nil {
        return &ChurnState{}, nil // New player
    }
    var state ChurnState
    json.Unmarshal([]byte(data), &state)
    return &state, nil
}
```

**Update State:**
```go
func updateChurnState(ctx context.Context, userID string, state *ChurnState) error {
    key := fmt.Sprintf("extend_anti_churn:churn_state:%s", userID)
    data, _ := json.Marshal(state)
    return redisClient.Set(ctx, key, data, 30*24*time.Hour).Err() // 30 day TTL
}
```

**Weekly Reset Logic:**
```go
func checkWeeklyReset(state *ChurnState) bool {
    now := time.Now()
    lastReset := state.Sessions.LastReset
    
    // Check if we've crossed Monday 00:00
    if now.Weekday() == time.Monday && now.Sub(lastReset) > 24*time.Hour {
        state.Sessions.LastWeek = state.Sessions.ThisWeek
        state.Sessions.ThisWeek = 0
        state.Sessions.LastReset = now
        return true
    }
    return false
}
```

---

## Churn Detection Logic

### Signal 1: Rage Quit Detection

**Trigger:** `statItemUpdated` event with `statCode = "rse-rage-quit"`

**How it works:**
1. Game client detects rage quit (ESC pressed while losing badly)
2. Game client increments `rse-rage-quit` stat via AGS Statistics Service
3. AGS publishes `statItemUpdated` event to Kafka
4. Extend Event Handler receives event and checks threshold

**Detection Logic:**
```go
func onStatUpdated(ctx context.Context, event *StatItemUpdated) error {
    if event.StatCode != "rse-rage-quit" {
        return nil
    }
    
    // Read current weekly count from AGS Statistics
    // (Statistic Cycle auto-resets Monday, game writes to it)
    count, err := getWeeklyRageQuitCount(event.UserID)
    if err != nil {
        return err
    }
    
    // Threshold: 3 rage quits in a week
    if count >= 3 {
        return triggerIntervention(ctx, event.UserID, "rage_quit")
    }
    
    return nil
}
```

**AGS Statistics Read:**
```go
func getWeeklyRageQuitCount(userID string) (int, error) {
    // Query AGS Statistics Service
    // Statistic Cycle handles weekly aggregation automatically
    stats, err := statsService.GetUserStatItemsShort(&social.GetUserStatItemsParams{
        UserID:    userID,
        StatCodes: []string{"rse-rage-quit"},
    })
    
    if err != nil || len(stats) == 0 {
        return 0, err
    }
    
    return int(stats[0].Value), nil
}
```

---

### Signal 2: Session Frequency Decline

**Trigger:** `oauthTokenGenerated` event (login)

**Detection Logic:**
```go
func checkSessionDecline(ctx context.Context, userID string) (bool, error) {
    // Read state from Redis
    state, err := getChurnState(ctx, userID)
    if err != nil {
        return false, err
    }
    
    // Increment this week's count
    state.Sessions.ThisWeek++
    
    // Check for weekly reset
    checkWeeklyReset(state)
    
    // Calculate decline (need at least 1 week of history)
    if state.Sessions.LastWeek == 0 {
        updateChurnState(ctx, userID, state)
        return false, nil
    }
    
    decline := 1.0 - (float64(state.Sessions.ThisWeek) / float64(state.Sessions.LastWeek))
    
    // Save state
    updateChurnState(ctx, userID, state)
    
    // Threshold: 50% decline
    return decline >= 0.5, nil
}
```

---

### Signal 3: Losing Streak

**Trigger:** `statItemUpdated` event with `statCode = "rse-current-losing-streak"`

**How it works:**
1. Game client tracks wins/losses and maintains `rse-current-losing-streak` stat
2. On match loss: game increments `rse-current-losing-streak`
3. On match win: game resets `rse-current-losing-streak` to 0
4. AGS publishes `statItemUpdated` event to Kafka
5. Extend Event Handler receives event and checks threshold

**Detection Logic:**
```go
func onStatUpdated(ctx context.Context, event *StatItemUpdated) error {
    if event.StatCode != "rse-current-losing-streak" {
        return nil
    }
    
    // Read current losing streak from event (or query AGS Statistics)
    losingStreak := int(event.Value)
    
    // Threshold: 5 consecutive losses
    if losingStreak >= 5 {
        return triggerIntervention(ctx, event.UserID, "losing_streak")
    }
    
    return nil
}
```

**Note:** No need to maintain losing streak in Redis since the game client already tracks it in AGS Statistics (`rse-current-losing-streak`). The Extend app simply reads the value from events or queries AGS Statistics
```

---

## Intervention System

### Rate Limiting

**Cooldown Rules:**
- Max 1 intervention per player per 7 days
- Check cooldown before triggering any signal

**Implementation:**
```go
func canTriggerIntervention(ctx context.Context, userID string) (bool, error) {
    state, err := getChurnState(ctx, userID)
    if err != nil {
        return false, err
    }
    
    // Check if cooldown is still active
    now := time.Now()
    if state.Intervention.CooldownUntil.After(now) {
        return false, nil // Still in cooldown
    }
    
    return true, nil
}
```

---

### Challenge Creation

**Trigger:** Any churn signal detected + rate limit passed

**Implementation Pattern (following template):**
```go
// In pkg/service/statHandler.go
type StatHandler struct {
    pb.UnimplementedStatItemUpdatedServiceServer
    fulfillment platform.FulfillmentService
    statistics  social.StatisticService
    redisClient *redis.Client
    namespace   string
}

func NewStatHandler(
    configRepo repository.ConfigRepository,
    tokenRepo repository.TokenRepository,
    redisClient *redis.Client,
    namespace string,
) *StatHandler {
    return &StatHandler{
        fulfillment: platform.FulfillmentService{
            Client:           factory.NewPlatformClient(configRepo),
            ConfigRepository: configRepo,
            TokenRepository:  tokenRepo,
        },
        statistics: social.StatisticService{
            Client:           factory.NewSocialClient(configRepo),
            ConfigRepository: configRepo,
            TokenRepository:  tokenRepo,
        },
        redisClient: redisClient,
        namespace:   namespace,
    }
}

func (h *StatHandler) OnMessage(ctx context.Context, msg *pb.StatItemUpdated) (*emptypb.Empty, error) {
    scope := common.GetScopeFromContext(ctx, "StatHandler.OnMessage")
    defer scope.Finish()
    
    logrus.Infof("received stat update: userId=%s statCode=%s value=%v", 
        msg.UserId, msg.StatCode, msg.Value)
    
    // Delegate to detection logic based on stat code
    if msg.StatCode == "rse-rage-quit" {
        if err := h.handleRageQuit(ctx, msg.UserId); err != nil {
            return &emptypb.Empty{}, status.Errorf(codes.Internal, 
                "failed to handle rage quit: %v", err)
        }
    }
    
    // ... other stat codes
    
    return &emptypb.Empty{}, nil
}

func (h *StatHandler) handleRageQuit(ctx context.Context, userID string) error {
    // Read weekly count from AGS Statistics
    count, err := h.getWeeklyRageQuitCount(userID)
    if err != nil {
        return err
    }
    
    if count >= 3 {
        return h.triggerIntervention(ctx, userID, "rage_quit")
    }
    
    return nil
}

func (h *StatHandler) triggerIntervention(ctx context.Context, userID, reason string) error {
    // Read state from Redis
    state, err := getChurnState(ctx, h.redisClient, userID)
    if err != nil {
        return err
    }
    
    // Check cooldown
    if !canTriggerIntervention(state) {
        logrus.Infof("skipping intervention for %s: still in cooldown", userID)
        return nil
    }
    
    // Get current win count from AGS Statistics
    currentWins, err := h.getUserWinCount(userID)
    if err != nil {
        return err
    }
    
    // Create challenge in Redis
    state.Challenge = ChallengeState{
        Active:        true,
        WinsNeeded:    3,
        WinsCurrent:   0,
        WinsAtStart:   currentWins,
        ExpiresAt:     time.Now().Add(72 * time.Hour),
        TriggerReason: reason,
    }
    
    // Update intervention history
    state.Intervention.LastTimestamp = time.Now()
    state.Intervention.CooldownUntil = time.Now().Add(7 * 24 * time.Hour)
    state.Intervention.TotalTriggered++
    
    // Save state
    if err := updateChurnState(ctx, h.redisClient, userID, state); err != nil {
        return err
    }
    
    // Grant entitlement using shared helper (like template's grantEntitlement)
    itemID := common.GetEnv("COMEBACK_BOOSTER_ITEM_ID", "comeback_booster")
    return grantEntitlement(h.fulfillment, h.namespace, userID, itemID)
}

func (h *StatHandler) getUserWinCount(userID string) (int, error) {
    stats, err := h.statistics.GetUserStatItemsShort(&social.GetUserStatItemsParams{
        UserID:    userID,
        Namespace: h.namespace,
        StatCodes: []string{"rse-match-wins"},
    })
    
    if err != nil || len(stats) == 0 {
        return 0, err
    }
    
    return int(stats[0].Value), nil
}
```

---

### Challenge Progress Tracking

**Trigger:** `statItemUpdated` event with `statCode = "rse-match-wins"`

**How it works:**
1. Game client updates `rse-match-wins` stat after player wins a match
2. AGS publishes `statItemUpdated` event to Kafka
3. Extend Event Handler receives event and checks for active challenge
4. Compares current wins against challenge start value to calculate progress

**Implementation:**
```go
func onStatUpdated(ctx context.Context, event *StatItemUpdated) error {
    if event.StatCode != "rse-match-wins" {
        return nil
    }
    
    // Read challenge state from Redis
    state, err := getChurnState(ctx, event.UserID)
    if err != nil {
        return err
    }
    
    // Check if challenge is active
    if !state.Challenge.Active {
        return nil // No active challenge
    }
    
    // Check if expired
    if time.Now().After(state.Challenge.ExpiresAt) {
        state.Challenge.Active = false
        updateChurnState(ctx, event.UserID, state)
        return nil // Challenge expired
    }
    
    // Calculate progress: current wins - starting wins
    currentWins := int(event.Value)
    progress := currentWins - state.Challenge.WinsAtStart
    
    // Update progress
    state.Challenge.WinsCurrent = progress
    
    // Check if completed
    if state.Challenge.WinsCurrent >= state.Challenge.WinsNeeded {
        return completeChallenge(ctx, event.UserID, state)
    }
    
    // Save state
    return updateChurnState(ctx, event.UserID, state)
}
```

---

### Reward Granting

**Trigger:** Challenge completion (3 wins achieved)

**Implementation Pattern (using template's shared helper):**
```go
// In pkg/service/entitlement.go (shared helper like template)
func grantEntitlement(fulfillmentService platform.FulfillmentService, namespace string, userID string, itemID string) error {
    quantity := int32(1)
    fulfillmentResponse, err := fulfillmentService.FulfillItemShort(&fulfillment.FulfillItemParams{
        Namespace: namespace,
        UserID:    userID,
        Body: &platformclientmodels.FulfillmentRequest{
            ItemID:   itemID,
            Quantity: &quantity,
            Source:   platformclientmodels.EntitlementGrantSourceREWARD,
        },
    })

    if err != nil {
        return err
    }

    if fulfillmentResponse == nil || fulfillmentResponse.EntitlementSummaries == nil || len(fulfillmentResponse.EntitlementSummaries) <= 0 {
        return fmt.Errorf("could not grant item to user")
    }

    return nil
}

// In pkg/service/statHandler.go
func (h *StatHandler) completeChallenge(ctx context.Context, userID string, state *ChurnState) error {
    scope := common.GetScopeFromContext(ctx, "StatHandler.completeChallenge")
    defer scope.Finish()
    
    logrus.Infof("completing challenge for user %s", userID)
    
    // Mark challenge as complete
    state.Challenge.Active = false
    
    // Grant speed booster entitlement (item configured in AGS Admin Portal)
    speedBoosterItemID := common.GetEnv("SPEED_BOOSTER_ITEM_ID", "speed_booster")
    if err := grantEntitlement(h.fulfillment, h.namespace, userID, speedBoosterItemID); err != nil {
        return status.Errorf(codes.Internal, "failed to grant speed booster: %v", err)
    }
    
    logrus.Infof("granted speed booster to user %s", userID)
    
    // Save state
    if err := updateChurnState(ctx, h.redisClient, userID, state); err != nil {
        return status.Errorf(codes.Internal, "failed to save state: %v", err)
    }
    
    return nil
}
```

**Note:** Following template conventions:
- Use shared `grantEntitlement()` helper for consistency
- Item IDs configured via environment variables (like template's `ITEM_ID_TO_GRANT`)
- Return gRPC status errors with `status.Errorf(codes.Internal, ...)`
- Use `common.GetEnv()` for environment variable access

---

## API Integration

### AGS Services Used

#### 1. IAM Service (Authentication)
**Purpose:** App authentication to call other AGS services

**Usage Pattern (from template's main.go):**
```go
import (
    "github.com/AccelByte/accelbyte-go-sdk/services-api/pkg/service/iam"
    "github.com/AccelByte/accelbyte-go-sdk/services-api/pkg/factory"
    "github.com/AccelByte/accelbyte-go-sdk/services-api/pkg/repository"
    sdkAuth "github.com/AccelByte/accelbyte-go-sdk/services-api/pkg/utils/auth"
)

// In main.go initialization
func initIAMAuth() error {
    // Create repositories
    configRepo := sdkAuth.DefaultConfigRepositoryImpl()
    tokenRepo := sdkAuth.DefaultTokenRepositoryImpl()
    refreshRepo := &sdkAuth.RefreshTokenImpl{
        AutoRefresh: true,
        RefreshRate: 0.8,
    }
    
    // Create OAuth service
    oauthService := iam.OAuth20Service{
        Client:                 factory.NewIamClient(configRepo),
        ConfigRepository:       configRepo,
        TokenRepository:        tokenRepo,
        RefreshTokenRepository: refreshRepo,
    }
    
    // Login with client credentials
    clientId := configRepo.GetClientId()
    clientSecret := configRepo.GetClientSecret()
    err := oauthService.LoginClient(&clientId, &clientSecret)
    if err != nil {
        return fmt.Errorf("unable to login using clientId and clientSecret: %v", err)
    }
    
    return nil
}
```

---

#### 2. Statistics Service (Read Aggregations)
**Purpose:** Read churn signal aggregations (weekly rage quits, current wins)

**Usage Pattern (following template's service initialization):**
```go
import (
    "github.com/AccelByte/accelbyte-go-sdk/services-api/pkg/service/social"
    "github.com/AccelByte/accelbyte-go-sdk/services-api/pkg/factory"
    "github.com/AccelByte/accelbyte-go-sdk/social-sdk/pkg/socialclient/user_statistic"
)

// In handler constructor
func NewStatHandler(
    configRepo repository.ConfigRepository,
    tokenRepo repository.TokenRepository,
    redisClient *redis.Client,
    namespace string,
) *StatHandler {
    return &StatHandler{
        statistics: social.StatisticService{
            Client:           factory.NewSocialClient(configRepo),
            ConfigRepository: configRepo,
            TokenRepository:  tokenRepo,
        },
        // ... other fields
    }
}

// In handler methods
func (h *StatHandler) getWeeklyRageQuitCount(userID string) (int, error) {
    result, err := h.statistics.GetUserStatItemsShort(&user_statistic.GetUserStatItemsParams{
        UserID:    userID,
        Namespace: h.namespace,
        StatCodes: []string{"rse-rage-quit"},
    })
    
    if err != nil || len(result) == 0 {
        return 0, err
    }
    
    return int(result[0].Value), nil
}
```

---

#### 3. Platform Service (Grant Rewards)
**Purpose:** Grant booster entitlements directly via fulfillment

**Usage Pattern (using template's shared helper in `pkg/service/entitlement.go`):**
```go
import (
    "github.com/AccelByte/accelbyte-go-sdk/services-api/pkg/service/platform"
    "github.com/AccelByte/accelbyte-go-sdk/services-api/pkg/factory"
    "github.com/AccelByte/accelbyte-go-sdk/platform-sdk/pkg/platformclient/fulfillment"
    "github.com/AccelByte/accelbyte-go-sdk/platform-sdk/pkg/platformclientmodels"
)

// In handler constructor
func NewStatHandler(
    configRepo repository.ConfigRepository,
    tokenRepo repository.TokenRepository,
    redisClient *redis.Client,
    namespace string,
) *StatHandler {
    return &StatHandler{
        fulfillment: platform.FulfillmentService{
            Client:           factory.NewPlatformClient(configRepo),
            ConfigRepository: configRepo,
            TokenRepository:  tokenRepo,
        },
        // ... other fields
    }
}

// Shared helper in pkg/service/entitlement.go (like template)
func grantEntitlement(fulfillmentService platform.FulfillmentService, namespace string, userID string, itemID string) error {
    quantity := int32(1)
    fulfillmentResponse, err := fulfillmentService.FulfillItemShort(&fulfillment.FulfillItemParams{
        Namespace: namespace,
        UserID:    userID,
        Body: &platformclientmodels.FulfillmentRequest{
            ItemID:   itemID,
            Quantity: &quantity,
            Source:   platformclientmodels.EntitlementGrantSourceREWARD,
        },
    })

    if err != nil {
        return err
    }

    if fulfillmentResponse == nil || fulfillmentResponse.EntitlementSummaries == nil || len(fulfillmentResponse.EntitlementSummaries) <= 0 {
        return fmt.Errorf("could not grant item to user")
    }

    return nil
}
```

---

### Environment Configuration

**Required Environment Variables (following template `.env.template`):**

```bash
# AGS Configuration
AB_BASE_URL=https://demo.accelbyte.io
AB_NAMESPACE=your-namespace
AB_CLIENT_ID=your-client-id
AB_CLIENT_SECRET=your-client-secret

# Item IDs (configured in AGS Admin Portal)
COMEBACK_BOOSTER_ITEM_ID=comeback_booster
SPEED_BOOSTER_ITEM_ID=speed_booster

# Redis Configuration (auto-provided by AccelByte when deployed)
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=

# OpenTelemetry (optional, for local testing)
OTEL_EXPORTER_ZIPKIN_ENDPOINT=http://host.docker.internal:9411/api/v2/spans
OTEL_SERVICE_NAME=ExtendAntiChurnHandler

# gRPC Debug (optional)
# GRPC_GO_LOG_VERBOSITY_LEVEL=99
# GRPC_GO_LOG_SEVERITY_LEVEL=info
```

**Environment Variable Access Pattern (using template's `common.GetEnv`):**
```go
namespace := common.GetEnv("AB_NAMESPACE", "accelbyte")
comebackBoosterID := common.GetEnv("COMEBACK_BOOSTER_ITEM_ID", "comeback_booster")
grpcPort := common.GetEnvInt("GRPC_PORT", 6565)
```

---

## Deployment

### Build & Package

**Multi-Stage Dockerfile (following template):**

The template uses a 3-stage build process:

1. **Stage 1: Proto Code Generation**
   - Ubuntu 22.04 with protoc + Go compiler + protoc-gen-go plugins
   - Generates Go code from `.proto` files via `proto.sh`
   - Output: `pkg/pb/` directory with generated code

2. **Stage 2: Application Build**
   - golang:1.24 base image
   - Copies generated proto files from Stage 1
   - Builds Go binary with `go build -v -modcacherw`
   - Cross-compilation support via `TARGETOS` and `TARGETARCH`

3. **Stage 3: Runtime Container**
   - Minimal Alpine 3.22 image
   - Only contains final binary (`extend-event-handler`)
   - Exposes ports 6565 (gRPC) and 8080 (Prometheus)

**Build Commands:**
```bash
# Generate protobuf code (optional, done in Dockerfile)
make proto

# Build Docker image
docker build -t anti-churn-event-handler:v1.0.0 .

# Local development with docker-compose
docker compose up --build
```

---

### Deployment Steps

1. **Create Extend App in AGS Admin Portal**
   - Navigate to Extend > Event Handler
   - Create new app: "anti-churn-handler"
   - Configure resources (CPU, memory)

2. **Upload Docker Image**
   ```bash
   extend-helper-cli image-upload \
     --work-dir . \
     --namespace your-namespace \
     --app anti-churn-handler \
     --image-tag v1.0.0
   ```

3. **Configure Environment Variables**
   - Set `AB_CLIENT_ID`, `AB_CLIENT_SECRET` in Admin Portal
   - Redis configuration auto-provided by AccelByte

4. **Subscribe to Events**
   - Subscribe app to topics:
     - `iam.oauthTokenGenerated`
     - `social.statItemUpdated`

5. **Deploy**
   - Click "Deploy Latest Image"
   - Wait for status: RUNNING

---

### Testing

**Local Testing with Postman:**

1. Run app locally:
   ```bash
   docker compose up --build
   ```

2. Open Postman gRPC client

3. Connect to `localhost:6565`

4. Send test event (example `statItemUpdated`):
   ```json
   {
     "payload": {
       "userId": "test-user-123",
       "statCode": "rse-rage-quit",
       "value": 3
     },
     "namespace": "your-namespace",
     "timestamp": "2025-12-06T10:30:00Z"
   }
   ```

5. Verify response and check Redis state

---

### Monitoring

**Key Metrics to Track:**
- Event processing rate (events/sec)
- Intervention trigger rate (interventions/hour)
- Challenge completion rate (%)
- API call latency (ms)
- Redis operation latency (ms)

**Available via AGS Admin Portal:**
- App status (running/stopped)
- Resource usage (CPU, memory)
- Logs (stdout/stderr)
- Grafana dashboards (if enabled)

---

## Project Structure

Following AccelByte's official Extend Event Handler template structure:

```
extends-anti-churn/
├── main.go                       # Entry point with gRPC server setup
├── pkg/
│   ├── proto/
│   │   └── accelbyte-asyncapi/   # Proto definitions (from AccelByte API Proto repo)
│   │       ├── iam/
│   │       │   └── oauth/v1/
│   │       │       └── oauth.proto       # oauthTokenGenerated event
│   │       └── social/
│   │           └── statistic/v1/
│   │               └── statistic.proto   # statItemUpdated event
│   ├── pb/
│   │   └── accelbyte-asyncapi/   # Generated protobuf Go code
│   │       ├── iam/oauth/v1/
│   │       └── social/statistic/v1/
│   ├── service/                  # Event handlers
│   │   ├── loginHandler.go       # Handle oauthTokenGenerated
│   │   ├── statHandler.go        # Handle statItemUpdated
│   │   └── entitlement.go        # Shared entitlement logic
│   └── common/                   # Shared utilities
│       ├── logging.go            # gRPC logging interceptor
│       ├── scope.go              # OpenTelemetry tracing
│       └── utils.go              # Env helpers, random ID gen
├── proto.sh                      # Proto code generation script
├── go.mod
├── go.sum
├── Dockerfile                    # Multi-stage build (proto gen + builder + runtime)
├── docker-compose.yaml           # Local development setup
├── Makefile                      # Build commands
├── .env.template                 # Environment variable template
└── README.md
```

### Key Structure Notes (from Template)

**`main.go`:**
- gRPC server setup with interceptors (Prometheus, logging, OpenTelemetry)
- IAM OAuth20 client credentials authentication
- Service registration pattern: `pb.RegisterServiceServer(s, handler)`
- Two server ports: 6565 (gRPC), 8080 (Prometheus metrics)

**`pkg/service/` handlers:**
- Embed `UnimplementedXServiceServer` from generated protobuf
- Constructor pattern: `NewHandler(configRepo, tokenRepo, namespace)`
- Handler method signature: `OnMessage(ctx context.Context, msg *pb.Event) (*emptypb.Empty, error)`
- Use `common.GetScopeFromContext(ctx, "HandlerName.OnMessage")` for tracing
- Return errors as `status.Errorf(codes.Internal, "message: %v", err)`

**`pkg/common/` utilities:**
- `GetEnv(key, fallback)` - Environment variable helpers
- `GetScopeFromContext(ctx, name)` - OpenTelemetry span creation
- `InterceptorLogger(logger)` - Logrus adapter for gRPC interceptors

**Environment Variables (`.env.template`):**
```env
AB_BASE_URL=https://demo.accelbyte.io
AB_NAMESPACE=accelbyte
AB_CLIENT_ID=
AB_CLIENT_SECRET=
ITEM_ID_TO_GRANT=
```

**Docker Multi-Stage Build:**
- Stage 1: Proto code generation (protoc + Go plugins)
- Stage 2: Go application build (with generated proto files)
- Stage 3: Minimal Alpine runtime container

---

## Implementation Phases

### Phase 1: Foundation & Infrastructure (Days 1-2)

**Goal:** Set up project structure, dependencies, and basic gRPC server

**Tasks:**

1. **Project Scaffolding**
   - Copy template structure to main project
   - Update `go.mod` with project name: `extends-anti-churn`
   - Create directory structure: `pkg/service/`, `pkg/common/`, `pkg/proto/`, `pkg/pb/`
   - Copy shared utilities from template: `logging.go`, `scope.go`, `utils.go`

2. **Proto Definitions**
   - Download proto files from AccelByte API Proto repository:
     - IAM OAuth: `https://github.com/AccelByte/accelbyte-api-proto/tree/main/asyncapi/accelbyte/iam/oauth/v1/oauth.proto`
       - For `oauthTokenGenerated` event
       - Save to `pkg/proto/accelbyte-asyncapi/iam/oauth/v1/oauth.proto`
     - Social Statistic: `https://github.com/AccelByte/accelbyte-api-proto/tree/main/asyncapi/accelbyte/social/statistic/v1/statistic.proto`
       - For `statItemUpdated` event
       - Save to `pkg/proto/accelbyte-asyncapi/social/statistic/v1/statistic.proto`
   - Copy `proto.sh` script for code generation from template
   - Copy `Makefile` from template for build commands
   - Run `make proto` to generate Go code in `pkg/pb/`
   - Verify generated files: `pkg/pb/accelbyte-asyncapi/iam/oauth/v1/` and `pkg/pb/accelbyte-asyncapi/social/statistic/v1/`

3. **Main Server Setup**
   - Create `main.go` with gRPC server initialization
   - Add interceptors: Prometheus, logging, OpenTelemetry
   - Configure IAM OAuth20 client credentials authentication
   - Set up health check and reflection
   - Configure two ports: 6565 (gRPC), 8080 (metrics)

4. **Redis Integration**
   - Add `go-redis/redis` dependency using `go get github.com/redis/go-redis/v9`
   - Create `pkg/state/redis.go` with Redis client initialization
   - Implement basic connection and ping test
   - Add Redis configuration via environment variables

5. **Environment Configuration**
   - Create `.env.template` with all required variables
   - Create `.env` for local development (gitignored)
   - Update `docker-compose.yaml` with Redis service
   - Test local environment setup

**Deliverables:**
- ✅ Working gRPC server that starts successfully
- ✅ Redis connection established
- ✅ IAM authentication working
- ✅ Prometheus metrics available at `:8080/metrics`
- ✅ Health check responding at gRPC endpoint

**Validation:**
```bash
# Server starts without errors
go run main.go

# Metrics endpoint responds
curl http://localhost:8080/metrics

# Redis connection works
docker compose up redis
```

---

### Phase 2: State Management & Data Models (Days 3-4)

**Goal:** Implement Redis state management and churn detection data structures

**Tasks:**

1. **State Data Models**
   - Create `pkg/state/models.go` with structs:
     - `ChurnState` (sessions, streaks, challenge, intervention)
     - `SessionState` (thisWeek, lastWeek, lastReset)
     - `ChallengeState` (active, winsNeeded, winsCurrent, expiresAt, triggerReason)
     - `InterventionState` (lastTimestamp, cooldownUntil, totalTriggered)
   - Add JSON tags for Redis serialization
   - Add validation methods

2. **Redis State Operations**
   - Implement `getChurnState(ctx, userID)` in `pkg/state/redis.go`
   - Implement `updateChurnState(ctx, userID, state)`
   - Add key prefix constant: `extend_anti_churn:churn_state:`
   - Set TTL to 30 days for all keys
   - Add error handling for Redis operations

3. **Weekly Reset Logic**
   - Implement `checkWeeklyReset(state)` function
   - Detect Monday 00:00 crossover
   - Rotate thisWeek → lastWeek counts
   - Update lastReset timestamp

4. **Rate Limiting Logic**
   - Implement `canTriggerIntervention(ctx, userID)` in `pkg/state/cooldown.go`
   - Check cooldown expiry against current time
   - Return boolean + reason if blocked
   - Add logging for cooldown checks

5. **Unit Tests**
   - Test weekly reset logic with various dates
   - Test cooldown calculations
   - Test JSON serialization/deserialization
   - Test Redis key generation with different userIDs

**Deliverables:**
- ✅ All state models defined and documented
- ✅ CRUD operations for Redis state working
- ✅ Weekly reset logic tested and validated
- ✅ Rate limiting correctly enforces 7-day cooldown
- ✅ Unit tests passing with >80% coverage

**Validation:**
```bash
# Unit tests pass
go test ./pkg/state/...

# Manual Redis operations work
redis-cli GET extend_anti_churn:churn_state:test-user-123
```

---

### Phase 3: Event Handlers & Churn Detection (Days 5-7)

**Goal:** Implement event handlers and churn detection logic for all 3 signals

**Tasks:**

1. **Login Event Handler**
   - Create `pkg/service/loginHandler.go`
   - Implement `LoginHandler` struct with embedded `UnimplementedUserAuthenticationUserLoggedInServiceServer`
   - Implement `NewLoginHandler(configRepo, tokenRepo, redisClient, namespace)`
   - Implement `OnMessage(ctx, msg)` for `oauthTokenGenerated` event
   - Add session tracking: increment thisWeek count
   - Implement session decline detection (50% threshold)

2. **Stat Update Event Handler**
   - Create `pkg/service/statHandler.go`
   - Implement `StatHandler` struct with fulfillment, statistics, and Redis client
   - Implement `NewStatHandler(configRepo, tokenRepo, redisClient, namespace)`
   - Implement `OnMessage(ctx, msg)` for `statItemUpdated` event
   - Add switch statement for different stat codes

3. **Rage Quit Detection**
   - Implement `handleRageQuit(ctx, userID)` method
   - Query AGS Statistics for weekly `rse-rage-quit` count
   - Trigger intervention if count >= 3
   - Add logging for rage quit events

4. **Losing Streak Detection**
   - Implement `handleLosingStreak(ctx, userID, streakValue)` method
   - Read streak value from event (no AGS query needed)
   - Trigger intervention if streak >= 5
   - Add logging for losing streak events

5. **Match Win Progress Tracking**
   - Implement `handleMatchWin(ctx, userID, currentWins)` method
   - Read active challenge from Redis
   - Calculate progress: currentWins - winsAtStart
   - Complete challenge if progress >= winsNeeded
   - Add logging for challenge progress

6. **Service Registration**
   - Update `main.go` to register both handlers:
     - `pb.RegisterUserAuthenticationUserLoggedInServiceServer(s, loginHandler)`
     - `pb.RegisterStatItemUpdatedServiceServer(s, statHandler)`
   - Pass Redis client to constructors
   - Test handler initialization

**Deliverables:**
- ✅ Login handler detects session decline correctly
- ✅ Stat handler routes to correct detection logic
- ✅ All 3 churn signals detect correctly
- ✅ Challenge progress tracking works
- ✅ Handlers integrated with gRPC server

**Validation:**
```bash
# Send test events via Postman gRPC client
# Test rage quit detection
# Test losing streak detection
# Test session decline detection
# Verify Redis state updates correctly
```

---

### Phase 4: Intervention System & Reward Granting (Days 8-10)

**Goal:** Implement intervention orchestration, challenge creation, and reward granting

**Tasks:**

1. **Shared Entitlement Helper**
   - Create `pkg/service/entitlement.go`
   - Implement `grantEntitlement(fulfillmentService, namespace, userID, itemID)`
   - Use `FulfillItemShort` with `EntitlementGrantSourceREWARD`
   - Add error handling and response validation
   - Add logging for entitlement grants

2. **Intervention Orchestration**
   - Implement `triggerIntervention(ctx, userID, reason)` in `statHandler.go`
   - Check rate limit via `canTriggerIntervention()`
   - Query current win count from AGS Statistics
   - Create challenge state in Redis
   - Update intervention history (timestamp, cooldown, count)
   - Grant comeback booster entitlement
   - Add comprehensive logging

3. **Challenge Completion**
   - Implement `completeChallenge(ctx, userID, state)` in `statHandler.go`
   - Mark challenge as inactive
   - Grant speed booster entitlement (item ID from env var)
   - Update Redis state
   - Add success logging

4. **Challenge Expiry Handling**
   - Add expiry check in `handleMatchWin()`
   - Mark expired challenges as inactive
   - Log expired challenges (no penalty)

5. **Environment Variables**
   - Add `COMEBACK_BOOSTER_ITEM_ID` to `.env.template`
   - Add `SPEED_BOOSTER_ITEM_ID` to `.env.template`
   - Update `docker-compose.yaml` with new env vars
   - Document required AGS Admin Portal item configuration

6. **Integration Testing**
   - Test full intervention flow: detection → cooldown → challenge → completion
   - Test rate limiting prevents spam (7-day cooldown)
   - Test challenge expiry (72 hours)
   - Test AGS Platform API integration (entitlement grants)
   - Test AGS Statistics API integration (stat queries)

7. **Error Handling & Resilience**
   - Add retry logic for transient AGS API failures
   - Add fallback behavior for Redis connection issues
   - Add validation for environment variables on startup
   - Add graceful degradation if services unavailable

**Deliverables:**
- ✅ Full intervention flow working end-to-end
- ✅ Challenge creation and completion working
- ✅ Entitlement grants succeeding in AGS
- ✅ Rate limiting enforced correctly
- ✅ Error handling robust and tested
- ✅ Integration tests passing

**Validation:**
```bash
# Full flow test via Postman
# 1. Send 3 rage quit events → intervention triggered
# 2. Check Redis state → challenge created
# 3. Send 3 match win events → challenge completed
# 4. Check AGS Admin Portal → entitlements granted
# 5. Send another rage quit → intervention blocked (cooldown)

# Check logs for comprehensive tracking
docker compose logs -f app
```

---

### Phase 5: Deployment & Production Readiness (Days 11-12)

**Goal:** Build Docker image, deploy to AGS, and validate production setup

**Tasks:**

1. **Dockerfile Finalization**
   - Verify multi-stage build works correctly
   - Test proto generation stage
   - Test application build stage
   - Test runtime container starts successfully
   - Optimize image size (remove unnecessary layers)

2. **Build & Test**
   - Run `docker build -t extends-anti-churn:v1.0.0 .`
   - Test local Docker container with `docker compose up`
   - Verify all environment variables load correctly
   - Test Redis connectivity in Docker environment
   - Test AGS API calls from Docker container

3. **AGS Admin Portal Configuration**
   - Create items in Platform Service:
     - `comeback_booster` (consumable, 1 use)
     - `speed_booster` (consumable, 5 uses)
   - Create statistic configurations:
     - `rse-rage-quit` (weekly cycle, auto-reset Monday)
     - `rse-current-losing-streak` (persistent)
     - `rse-match-wins` (persistent, incrementing)
   - Create OAuth client for Extend app
   - Note client ID and secret for deployment

4. **Extend App Deployment**
   - Create Extend Event Handler app in AGS Admin Portal
   - Configure resources: CPU (1 core), Memory (512MB)
   - Upload Docker image via `extend-helper-cli`:
     ```bash
     extend-helper-cli image-upload \
       --work-dir . \
       --namespace <namespace> \
       --app extends-anti-churn \
       --image-tag v1.0.0
     ```
   - Configure environment variables in Admin Portal
   - Subscribe to events:
     - `iam.oauthTokenGenerated`
     - `social.statItemUpdated`

5. **Production Validation**
   - Deploy app and wait for RUNNING status
   - Monitor logs in AGS Admin Portal
   - Test with real game client login
   - Trigger rage quit in game → verify intervention
   - Complete challenge in game → verify reward
   - Check Grafana dashboards (if available)
   - Verify Prometheus metrics exposed correctly

6. **Documentation**
   - Update `README.md` with:
     - Setup instructions
     - Environment variable documentation
     - AGS configuration requirements
     - Deployment steps
     - Testing procedures
     - Troubleshooting guide
   - Document known limitations
   - Add monitoring and alerting recommendations

7. **Monitoring Setup**
   - Define key metrics to track:
     - Events processed per second
     - Interventions triggered per hour
     - Challenge completion rate (%)
     - API call latency (p50, p95, p99)
     - Redis operation latency
     - Error rate by handler
   - Set up alerts for critical issues:
     - High error rate (>5%)
     - API latency spike (>2s)
     - Redis connection failures

**Deliverables:**
- ✅ Docker image built and tested
- ✅ App deployed to AGS and running
- ✅ Event subscriptions configured
- ✅ Production testing completed successfully
- ✅ Documentation comprehensive and accurate
- ✅ Monitoring and alerting configured

**Validation:**
```bash
# Check app status in AGS Admin Portal
# Status: RUNNING
# Events: Processing successfully
# Logs: No errors

# Test end-to-end with game client
# 1. Play game and trigger churn signal
# 2. Verify intervention in Redis
# 3. Complete challenge
# 4. Verify reward in game client
```

---

## Phase Summary

| Phase | Duration | Focus | Key Deliverable |
|-------|----------|-------|-----------------|
| **Phase 1** | 2 days | Foundation | Working gRPC server with IAM auth |
| **Phase 2** | 2 days | State Management | Redis CRUD + rate limiting working |
| **Phase 3** | 3 days | Event Handlers | All 3 churn signals detecting correctly |
| **Phase 4** | 3 days | Intervention System | Full intervention flow end-to-end |
| **Phase 5** | 2 days | Deployment | Production app running in AGS |

**Total Timeline:** 12 days (can be parallelized with 2 developers)

**Critical Path:**
1. Phase 1 → Phase 2 (sequential, foundation required)
2. Phase 2 → Phase 3 (sequential, state required for detection)
3. Phase 3 → Phase 4 (sequential, detection required for intervention)
4. Phase 4 → Phase 5 (sequential, working app required for deployment)

**Risk Mitigation:**
- Each phase has clear validation criteria
- Unit tests prevent regressions
- Integration tests catch issues early
- Incremental deployment allows rollback

---

## Summary

This Extend Event Handler app provides:

✅ **Automated churn detection** via 3 signals (rage quits, session decline, losing streaks)  
✅ **Intelligent interventions** with rate limiting and cooldowns  
✅ **Challenge-based rewards** stored in Redis, tracked via stat events  
✅ **Direct entitlement grants** without Reward Service complexity  
✅ **Hybrid storage** (AGS Statistics for visibility, Redis for flexibility)  
✅ **Production-ready** architecture with proper error handling and monitoring

**Total Complexity:** Moderate - suitable for PoC with clear path to production scaling.
