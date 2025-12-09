#!/bin/bash
set -e

echo "[STARTUP] Anti-Churn Extend App Initializing..."

# Configuration
REDIS_HOST="${REDIS_HOST:-localhost}"
REDIS_PORT="${REDIS_PORT:-6379}"
REDIS_PASSWORD="${REDIS_PASSWORD:-}"
REDIS_MODE="${REDIS_MODE:-auto}"  # auto, external, embedded

# Colors for logging
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to test Redis connection
test_redis_connection() {
    local host=$1
    local port=$2
    local max_attempts=5
    local attempt=1
    
    log_info "Testing Redis connection at $host:$port..."
    
    while [ $attempt -le $max_attempts ]; do
        if redis-cli -h "$host" -p "$port" ping > /dev/null 2>&1; then
            log_info "Redis is ready at $host:$port"
            return 0
        fi
        log_warn "Redis not ready (attempt $attempt/$max_attempts), waiting..."
        sleep 2
        attempt=$((attempt + 1))
    done
    
    log_error "Redis is not available at $host:$port after $max_attempts attempts"
    return 1
}

# Function to start embedded Redis
start_embedded_redis() {
    log_info "Starting embedded Redis server..."
    
    # Start Redis in background
    redis-server /etc/redis/redis.conf --daemonize yes
    
    # Wait for Redis to be ready
    if test_redis_connection "localhost" "6379"; then
        log_info "Embedded Redis started successfully"
        export REDIS_HOST="localhost"
        export REDIS_PORT="6379"
        return 0
    else
        log_error "Failed to start embedded Redis"
        return 1
    fi
}

# Main startup logic
main() {
    log_info "Redis Mode: $REDIS_MODE"
    
    if [ "$REDIS_MODE" = "embedded" ]; then
        # Force embedded mode
        log_info "Embedded mode forced, starting local Redis..."
        if ! start_embedded_redis; then
            log_error "Failed to start embedded Redis, exiting"
            exit 1
        fi
        
    elif [ "$REDIS_MODE" = "external" ]; then
        # Force external mode
        log_info "External mode forced, connecting to $REDIS_HOST:$REDIS_PORT..."
        if ! test_redis_connection "$REDIS_HOST" "$REDIS_PORT"; then
            log_error "External Redis not available, exiting"
            exit 1
        fi
        
    else
        # Auto mode: try external first, fallback to embedded
        log_info "Auto mode: trying external Redis at $REDIS_HOST:$REDIS_PORT..."
        
        if test_redis_connection "$REDIS_HOST" "$REDIS_PORT"; then
            log_info "Using external Redis at $REDIS_HOST:$REDIS_PORT"
        else
            log_warn "External Redis not available, falling back to embedded Redis"
            if ! start_embedded_redis; then
                log_error "Both external and embedded Redis failed, exiting"
                exit 1
            fi
        fi
    fi
    
    # Display Redis configuration
    log_info "═══════════════════════════════════════"
    log_info "Redis Configuration:"
    log_info "  Host: $REDIS_HOST"
    log_info "  Port: $REDIS_PORT"
    log_info "  Mode: $([ "$REDIS_HOST" = "localhost" ] && echo "Embedded" || echo "External")"
    log_info "═══════════════════════════════════════"
    
    # Start the main application
    log_info "Starting Anti-Churn Extend App..."
    exec /app/main
}

# Trap signals for graceful shutdown
trap 'log_info "Shutting down..."; exit 0' SIGTERM SIGINT

# Run main
main
