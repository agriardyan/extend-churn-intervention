#!/bin/bash
set -e

echo "[STARTUP] Anti-Churn Extend App Initializing..."

# Configuration
REDIS_HOST="${REDIS_HOST:-redis}"
REDIS_PORT="${REDIS_PORT:-6379}"
REDIS_PASSWORD="${REDIS_PASSWORD:-}"

# Colors for logging
GREEN='\033[0;32m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

# Display Redis configuration
log_info "═══════════════════════════════════════"
log_info "Redis Configuration:"
log_info "  Host: $REDIS_HOST"
log_info "  Port: $REDIS_PORT"
log_info "═══════════════════════════════════════"

# Start the main application
log_info "Starting Anti-Churn Extend App..."
exec /app/main
