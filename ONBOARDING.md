# Onboarding Guide: Extend Anti-Churn Event Handler

Welcome to the **Extend Anti-Churn Event Handler** project! This guide will help you understand the codebase and get started with development.

## Table of Contents

1. [Project Overview](#project-overview)
2. [Architecture](#architecture)
3. [Key Concepts](#key-concepts)
4. [Project Structure](#project-structure)
5. [Getting Started](#getting-started)
6. [Development Workflow](#development-workflow)
7. [Testing](#testing)
8. [How It Works](#how-it-works)
9. [Configuration](#configuration)
10. [Deployment](#deployment)

---

## Project Overview

This is an **AccelByte Gaming Services (AGS) Extend Event Handler** application designed to detect and prevent player churn in the "Ready, Set, Enter!" game. The system:

- **Listens** to real-time game events via gRPC/Kafka
- **Detects** at-risk players based on behavioral patterns
- **Intervenes** by creating comeback challenges with rewards
- **Grants** in-game items (Speed Boosters) when challenges are completed

**Technology Stack:**
- Go 1.23+
- gRPC for event streaming
- Redis for player state management
- Protocol Buffers for event definitions
- Docker for containerization
- AccelByte Go SDK for AGS integration

---

## Architecture

```
┌─────────────┐     Events      ┌──────────────────┐     State      ┌───────┐
│   AGS       │ ───────────────> │  Extend Event    │ ────────────> │ Redis │
│   Platform  │   (gRPC/Kafka)  │  Handler (gRPC)  │               └───────┘
│             │                  │                  │
│  - IAM      │                  │  - OAuth Handler │
│  - Social   │                  │  - Stat Handler  │
│  - Platform │ <──────────────  │  - Entitlement   │
│             │   API Calls      │    Grant Logic   │
└─────────────┘                  └──────────────────┘
```

**Event Flow:**
1. Player performs actions in the game
2. AGS publishes events to Kafka
3. Events are streamed to our gRPC handler via Extend framework
4. Handler updates Redis state and detects churn patterns
5. When churn detected, creates challenge and grants rewards via AGS Platform API

---

## Key Concepts

### Churn Detection Triggers

The system detects three types of at-risk behavior:

1. **Session Frequency Decline** (OAuth events)
   - Player was active last week but hasn't logged in this week
   - Detected after 7+ days since last reset
   - Triggers intervention on next login

2. **Rage Quits** (Statistic events: `rse-rage-quit`)
   - Player rage quit 3+ times
   - Indicates frustration with the game

3. **Losing Streaks** (Statistic events: `rse-current-losing-streak`)
   - Player lost 5+ consecutive matches
   - Sign of skill gap or discouragement

### Comeback Challenges

When churn is detected, the system creates a challenge:
- **Goal**: Win 3 matches within 7 days
- **Reward**: Speed Booster item (granted via Platform API)
- **Cooldown**: 48 hours between interventions
- **Progress Tracking**: Monitored via `rse-match-wins` statistic

### State Management

All player state is stored in Redis with 30-day TTL:

```go
type ChurnState struct {
    Sessions     SessionState      // Weekly login tracking
    Challenge    ChallengeState    // Active comeback challenge
    Intervention InterventionState // Cooldown & history
}
```

---

## Project Structure

```
extends-anti-churn/
├── main.go                      # Entry point - gRPC server setup
├── go.mod, go.sum               # Go dependencies
├── Dockerfile                   # Multi-stage build
├── docker-compose.yaml          # Local dev environment
├── Makefile                     # Build automation
├── proto.sh                     # Protobuf code generation
│
├── pkg/
│   ├── common/                  # Shared utilities
│   │   ├── logging.go           # Logrus logging interceptor
│   │   ├── scope.go             # OpenTelemetry tracing scope
│   │   ├── tracerProvider.go   # Zipkin tracing setup
│   │   └── utils.go             # Env var helpers
│   │
│   ├── service/                 # Event handlers
│   │   ├── handler.go           # Service structs & constants
│   │   ├── oauth_handler.go     # OAuth event processing
│   │   ├── statistic_handler.go # Stat event processing
│   │   ├── entitlement.go       # AGS Platform API integration
│   │   └── handler_test.go      # Unit tests
│   │
│   ├── state/                   # Redis state management
│   │   ├── models.go            # ChurnState data structures
│   │   ├── redis.go             # Redis client & CRUD operations
│   │   ├── logic.go             # Churn detection & challenge logic
│   │   ├── redis_test.go        # Redis tests (with miniredis)
│   │   └── logic_test.go        # Business logic unit tests
│   │
│   ├── proto/                   # Proto definitions (source)
│   │   └── accelbyte-asyncapi/
│   │       ├── iam/oauth/v1/    # OAuth token events
│   │       └── social/statistic/v1/ # Statistic update events
│   │
│   └── pb/                      # Generated protobuf code
│       └── accelbyte-asyncapi/  # (auto-generated, don't edit)
│
├── bin/                         # Binary tools
│   └── extend-helper-cli-linux_amd64
│
└── example/                     # Reference implementation
    └── extend-event-handler-go/ # AccelByte's official example
```

---

## Getting Started

### Prerequisites

- **Go 1.23+** (`go version`)
- **Docker & Docker Compose** (`docker --version`)
- **Protocol Buffers compiler** (optional, Docker handles this)
- **AccelByte Gaming Services account** with namespace configured
- **Items configured in AGS Admin Portal** (Speed Booster)

### Initial Setup

1. **Clone the repository**
   ```bash
   cd /home/k6/code/extends-app-contest/extends-anti-churn
   ```

2. **Create environment configuration**
   
   Create a `.env` file (not in repo for security) with:
   ```bash
   # AccelByte OAuth2 credentials
   AB_CLIENT_ID=your_client_id
   AB_CLIENT_SECRET=your_client_secret
   AB_BASE_URL=https://demo.accelbyte.io
   AB_NAMESPACE=your_namespace

   # Item IDs (configure in AGS Admin Portal first)
   SPEED_BOOSTER_ITEM_ID=speed_booster_item_uuid

   # Redis (defaults work for Docker Compose)
   REDIS_HOST=redis
   REDIS_PORT=6379
   REDIS_PASSWORD=
   REDIS_MAX_RETRIES=5
   REDIS_RETRY_DELAY_MS=1000

   # Optional: OpenTelemetry tracing
   OTEL_EXPORTER_ZIPKIN_ENDPOINT=http://host.docker.internal:9411/api/v2/spans
   ```

3. **Install dependencies**
   ```bash
   go mod download
   ```

### Running Locally

#### Option A: Direct Go Execution (Fastest for Development)

```bash
# The app auto-loads .env file
go run main.go
```

Services available at:
- gRPC server: `localhost:6565`
- Prometheus metrics: `http://localhost:8080/metrics`
- Health check: gRPC health protocol

**Note:** You'll need Redis running separately:
```bash
docker run -p 6379:6379 redis:7.2-alpine redis-server --requirepass ""
```

#### Option B: Docker Compose (Full Environment)

```bash
# Build and start all services
docker compose up --build

# Or run in detached mode
docker compose up -d --build

# View logs
docker compose logs -f app

# Stop services
docker compose down
```

---

## Development Workflow

### Making Code Changes

1. **Edit source files** in `pkg/service/` or `pkg/state/`
2. **Run tests** to verify changes:
   ```bash
   make test
   # or
   go test -v ./...
   ```
3. **Test locally** with `go run main.go` or Docker Compose
4. **Rebuild** if needed: `make build`

### Regenerating Protocol Buffers

If you modify `.proto` files in `pkg/proto/`:

```bash
# Using Make (recommended)
make proto

# Or manually with Docker
docker build --target proto-builder -t proto-builder .
docker run --rm -v $(pwd):/build -w /build proto-builder ./proto.sh
```

This generates Go code in `pkg/pb/` - **never edit generated files manually**.

### Code Style Guidelines

- **Error handling**: Always check and log errors with context
- **Logging**: Use `logrus.Infof/Warnf/Errorf` with structured fields
- **Testing**: Write unit tests for business logic (see `*_test.go` files)
- **Constants**: Define magic numbers as named constants in `handler.go`
- **State mutations**: Always follow get → modify → save pattern for Redis

---

## Testing

### Unit Tests

The codebase has comprehensive unit tests:

```bash
# Run all tests
go test ./... -v

# Run with coverage
go test ./... -v -cover

# Run specific package
go test ./pkg/state -v
go test ./pkg/service -v

# Run specific test
go test ./pkg/service -v -run TestOAuthHandler_OnMessage_NewPlayer
```

**Test Infrastructure:**
- Uses `miniredis` for in-memory Redis mocking
- No external dependencies needed for tests
- Tests cover: OAuth events, stat events, churn detection, challenges

### Integration Testing

For end-to-end testing with AGS:

1. **Deploy to AGS Extend** (see Deployment section)
2. **Use Postman collection** in `example/extend-event-handler-go/demo/`
3. **Trigger events** from game client or AGS Admin Portal
4. **Monitor logs**: `docker compose logs -f app`
5. **Check Redis state**: 
   ```bash
   docker compose exec redis redis-cli
   > KEYS extends_anti_churn:user_state:*
   > GET extends_anti_churn:user_state:USER_ID_HERE
   ```

---

## How It Works

### 1. OAuth Event Processing (Login Tracking)

**File:** [`pkg/service/oauth_handler.go`](pkg/service/oauth_handler.go)

```go
// Event: Player logs in and gets OAuth token
OauthTokenGenerated → OAuthHandler.OnMessage()
  ↓
1. Get churn state from Redis
2. Check if weekly reset needed (7+ days)
3. Increment Sessions.ThisWeek
4. Check if player is churning:
   - LastWeek > 0 && ThisWeek == 0 && 7+ days passed
5. If churning → Create challenge + Set cooldown
6. Save state to Redis
```

**Key Functions:**
- `state.CheckWeeklyReset()` - Resets weekly counters
- `state.ShouldTriggerIntervention()` - Checks churn conditions
- `state.CreateChallenge()` - Sets up comeback challenge

### 2. Statistic Event Processing

**File:** [`pkg/service/statistic_handler.go`](pkg/service/statistic_handler.go)

#### A. Rage Quit Handler

```go
// Event: Player rage quits (disconnect without finishing match)
StatItemUpdated(rse-rage-quit) → handleRageQuit()
  ↓
1. Get churn state
2. Check weekly reset
3. If rageQuitCount >= 3 (threshold):
   - Fetch current wins from AGS
   - Create challenge (3 wins in 7 days)
   - Set 48h cooldown
4. Save state
```

#### B. Match Win Handler

```go
// Event: Player wins a match
StatItemUpdated(rse-match-wins) → handleMatchWin()
  ↓
1. Get churn state
2. If no active challenge → Skip
3. Update challenge progress:
   - WinsCurrent = totalWins - WinsAtStart
4. If WinsCurrent >= WinsNeeded:
   - Grant Speed Booster via Platform API
   - Deactivate challenge
5. Save state
```

#### C. Losing Streak Handler

```go
// Event: Player's losing streak updated
StatItemUpdated(rse-current-losing-streak) → handleLosingStreak()
  ↓
Similar to rage quit handler, triggers at 5+ losses
```

### 3. State Management

**File:** [`pkg/state/redis.go`](pkg/state/redis.go)

**Redis Key Pattern:** `extends_anti_churn:user_state:{userID}`

**TTL:** 30 days (auto-cleanup of inactive players)

**State Structure:**
```json
{
  "sessions": {
    "thisWeek": 3,
    "lastWeek": 7,
    "lastReset": "2026-01-29T10:00:00Z"
  },
  "challenge": {
    "active": true,
    "winsNeeded": 3,
    "winsCurrent": 1,
    "winsAtStart": 42,
    "expiresAt": "2026-02-12T10:00:00Z",
    "triggerReason": "rage_quit"
  },
  "intervention": {
    "lastTimestamp": "2026-02-05T10:00:00Z",
    "cooldownUntil": "2026-02-07T10:00:00Z",
    "totalTriggered": 2
  }
}
```

### 4. Entitlement Grant

**File:** [`pkg/service/entitlement.go`](pkg/service/entitlement.go)

```go
// When challenge completed
grantEntitlement(fulfillmentService, namespace, userID, itemID)
  ↓
1. Prepare FulfillmentRequest
2. Call AGS Platform API: FulfillItemShort()
3. User receives Speed Booster in their inventory
```

---

## Configuration

### Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `AB_CLIENT_ID` | ✅ | - | OAuth2 client ID from AGS Admin Portal |
| `AB_CLIENT_SECRET` | ✅ | - | OAuth2 client secret |
| `AB_BASE_URL` | ✅ | - | AGS API endpoint (e.g., `https://demo.accelbyte.io`) |
| `AB_NAMESPACE` | ✅ | `accelbyte` | Your game namespace |
| `SPEED_BOOSTER_ITEM_ID` | ✅ | `speed_booster` | Item UUID from AGS Admin Portal |
| `REDIS_HOST` | No | `localhost` | Redis hostname |
| `REDIS_PORT` | No | `6379` | Redis port |
| `REDIS_PASSWORD` | No | `""` | Redis password (empty for no auth) |
| `REDIS_MAX_RETRIES` | No | `5` | Connection retry attempts |
| `REDIS_RETRY_DELAY_MS` | No | `1000` | Delay between retries (ms) |
| `OTEL_EXPORTER_ZIPKIN_ENDPOINT` | No | - | Zipkin endpoint for distributed tracing |

**Note:** AccelByte Extend platform provides managed Redis services. Always use external Redis (`REDIS_HOST` and `REDIS_PORT`) in production deployments.

### Tunable Parameters

**File:** [`pkg/service/handler.go`](pkg/service/handler.go)

```go
const (
    // Challenge requirements
    ChallengeWinsNeeded   = 3    // Wins needed to complete challenge
    ChallengeDurationDays = 7    // Days before challenge expires

    // Intervention rules
    InterventionCooldownHours = 48  // Hours between interventions
    RageQuitThreshold         = 3   // Rage quits to trigger intervention
    LosingStreakThreshold     = 5   // Consecutive losses to trigger

    // Stat codes (must match game implementation)
    StatCodeRageQuit     = "rse-rage-quit"
    StatCodeMatchWins    = "rse-match-wins"
    StatCodeLosingStreak = "rse-current-losing-streak"
)
```

**Weekly Reset Logic:**
- Sessions reset every 7 days from `LastReset` timestamp
- Active challenges are cancelled on reset
- Cooldowns persist across resets

---

## Deployment

### Building for Production

```bash
# Build Docker image
make build

# Or with custom tag
IMAGE_TAG=v1.0.0 make build

# Push to registry
docker tag extends-anti-churn:latest your-registry/extends-anti-churn:v1.0.0
docker push your-registry/extends-anti-churn:v1.0.0
```

### Docker Image Stages

The `Dockerfile` uses multi-stage builds:

1. **proto-builder**: Generates protobuf code from `.proto` files
2. **builder**: Compiles Go binary with dependencies
3. **final**: Minimal Alpine image with binary only

**Image size:** ~15-20MB (optimized)

### Deploying to AccelByte Extend

1. **Build and push image** to container registry
2. **Use Extend Helper CLI**:
   ```bash
   ./bin/extend-helper-cli-linux_amd64 deploy \
     --namespace your_namespace \
     --image your-registry/extends-anti-churn:v1.0.0
   ```
3. **Configure environment variables** in AGS Admin Portal
4. **Subscribe to events** in AGS Event Handler configuration:
   - `iam.oauth.oauthTokenGenerated`
   - `social.statistic.statItemUpdated`
5. **Monitor via AGS dashboards** and Prometheus metrics

### Production Checklist

- [ ] Configure proper `AB_CLIENT_ID` and `AB_CLIENT_SECRET`
- [ ] Set up Redis (managed service recommended)
- [ ] Configure item IDs in `SPEED_BOOSTER_ITEM_ID`
- [ ] Enable Zipkin/Jaeger for distributed tracing
- [ ] Set up Prometheus for metrics scraping
- [ ] Configure log aggregation (e.g., ELK stack)
- [ ] Set resource limits in Kubernetes/Docker
- [ ] Enable health checks in orchestrator
- [ ] Test event subscriptions in staging environment

---

## Troubleshooting

### Common Issues

**Problem:** `failed to connect to Redis`
- **Solution**: Check `REDIS_HOST` and `REDIS_PORT`, ensure Redis is running
- **Docker**: Use `redis` as hostname (service name in docker-compose)
- **Local**: Use `localhost` and ensure Redis is installed/running

**Problem:** `Error unable to login using clientId and clientSecret`
- **Solution**: Verify `AB_CLIENT_ID` and `AB_CLIENT_SECRET` in `.env`
- Check credentials in AGS Admin Portal
- Ensure client has proper permissions (Platform, Social scopes)

**Problem:** Events not being received
- **Solution**: Check AGS Event Handler subscription configuration
- Verify namespace matches `AB_NAMESPACE`
- Check gRPC server logs for connection errors
- Ensure port 6565 is accessible

**Problem:** Challenges not granting rewards
- **Solution**: Verify `SPEED_BOOSTER_ITEM_ID` exists in AGS catalog
- Check Platform API permissions for client
- Review logs for `failed to grant speed booster` errors

### Debug Mode

Enable verbose logging:

```go
// In main.go, change:
logrus.SetLevel(logrus.DebugLevel)
```

### Inspecting Redis State

```bash
# Access Redis CLI
docker compose exec redis redis-cli

# List all player states
> KEYS extends_anti_churn:user_state:*

# Get specific player
> GET extends_anti_churn:user_state:USER_ID_HERE

# Delete player state (for testing)
> DEL extends_anti_churn:user_state:USER_ID_HERE
```

---

## Further Reading

- **AccelByte Extend Docs**: https://docs.accelbyte.io/extend/
- **gRPC Go Tutorial**: https://grpc.io/docs/languages/go/
- **Redis Go Client**: https://github.com/go-redis/redis
- **AccelByte Go SDK**: https://github.com/AccelByte/accelbyte-go-sdk

## Contributing

1. Write unit tests for new features
2. Follow existing code structure and naming conventions
3. Update this ONBOARDING.md if adding new concepts
4. Run `make test` before committing
5. Keep business logic in `pkg/state/logic.go` separate from handlers

---

**Happy coding! 🚀**

*Last updated: February 2026*
