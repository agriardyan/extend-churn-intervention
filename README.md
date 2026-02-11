# Extend Anti-Churn Event Handler

Automated churn detection and intervention system for Ready, Set, Enter! game on AccelByte Gaming Services.

## Overview

This Extend Event Handler app automatically detects at-risk players and triggers interventions through direct entitlement grants. It listens to AGS events, maintains player state in Redis, and orchestrates the challenge-reward system.

## Features

- **Churn Detection**: Tracks rage quits, session frequency decline, and losing streaks
- **Smart Interventions**: Rate-limited comeback challenges with entitlement rewards
- **Real-time Processing**: Event-driven architecture via gRPC and Kafka
- **Scalable State**: Redis-based player state management

## Architecture

```
Game Client → AGS Statistics → Kafka Events → Extend Handler → Redis + Platform API
```

## Quick Start

### Prerequisites

- Go 1.21+
- Docker and Docker Compose
- Protocol Buffers compiler (protoc)
- AccelByte Gaming Services account

### Local Development

1. **Clone and Setup**
   ```bash
   git clone <repo-url>
   cd extends-anti-churn
   cp .env.template .env
   # Edit .env with your AGS credentials
   ```

2. **Generate Proto Code**
   ```bash
   make proto
   ```

3. **Run Locally (without Docker)**
   ```bash
   # The app will automatically load .env file
   go run main.go
   ```

4. **Run with Docker Compose**
   ```bash
   docker compose up --build
   ```

4. **Verify Services**
   - gRPC server: `localhost:6565`
   - Prometheus metrics: `http://localhost:8080/metrics`
   - Redis: `localhost:6379`

## Project Structure

```
extends-anti-churn/
├── main.go                 # Entry point with gRPC server
├── pkg/
│   ├── common/             # Shared utilities (logging, tracing, env)
│   ├── service/            # Event handlers
│   ├── state/              # Redis state management
│   ├── proto/              # Proto definitions
│   └── pb/                 # Generated proto code
├── docker-compose.yaml     # Local development setup
├── Dockerfile              # Multi-stage build
└── .env.template           # Environment variables template
```

## Environment Variables

See `.env.template` for required configuration:

- `AB_CLIENT_ID` / `AB_CLIENT_SECRET`: OAuth credentials
- `AB_BASE_URL`: AccelByte API endpoint
- `AB_NAMESPACE`: Your game namespace
- `REDIS_HOST` / `REDIS_PORT`: Redis connection
- Item IDs for rewards (configured in AGS Admin Portal)

## Development Status

**Phase 1: Foundation ✅ COMPLETE**
- Project structure and dependencies
- Proto code generation
- gRPC server with IAM authentication
- Redis integration scaffolding
- Environment configuration

**Phase 2-5: Coming Soon**
- State management and data models
- Event handlers and churn detection
- Intervention system and rewards
- Production deployment

## Testing

```bash
# Run unit tests
go test ./...

# Check metrics endpoint
curl http://localhost:8080/metrics

# Test gRPC health check
grpcurl -plaintext localhost:6565 grpc.health.v1.Health/Check
```

## Documentation

See `.plan/extend-anti-churn-design.md` for comprehensive design documentation.

## License

Copyright (c) 2025 AccelByte Inc. All Rights Reserved.
