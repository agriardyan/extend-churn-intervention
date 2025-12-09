# Copyright (c) 2025 AccelByte Inc. All Rights Reserved.
# This is licensed software from AccelByte Inc, for limitations
# and restrictions contact your company contract manager.

# ----------------------------------------
# Stage 0: Redis Binary Extraction
# ----------------------------------------
FROM redis:7-alpine AS redis-binary
# We only need the redis-server and redis-cli binaries

# ----------------------------------------
# Stage 1: Protoc Code Generation
# ----------------------------------------
FROM golang:1.24-alpine AS proto-builder

# Install build dependencies and protoc
RUN apk add --no-cache \
    bash \
    ca-certificates \
    curl \
    git \
    protobuf \
    protobuf-dev

# Install protoc Go tools and plugins
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest \
    && go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Set working directory
WORKDIR /build

# Copy proto sources and generator script
COPY proto.sh .
COPY pkg/proto/ pkg/proto/

# Generate protobuf files.
RUN chmod +x proto.sh && \
    ./proto.sh



# ----------------------------------------
# Stage 2: gRPC Server Builder
# ----------------------------------------
FROM golang:1.24-alpine AS builder

ARG TARGETOS=linux
ARG TARGETARCH=amd64

ARG GOOS=$TARGETOS
ARG GOARCH=$TARGETARCH
ARG CGO_ENABLED=0

# Set working directory
WORKDIR /build

# Copy and download the dependencies for application
COPY go.mod go.sum ./
RUN go mod download

# Copy application code
COPY . .

# Copy generated protobuf files from stage 1
COPY --from=proto-builder /build/pkg/pb pkg/pb

# Build the Go application binary for the target OS and architecture
RUN go build -v -modcacherw -o /output/extends-anti-churn .


# ----------------------------------------
# Stage 3: Runtime Container
# ----------------------------------------
FROM alpine:3.22

# Install runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    bash \
    curl

# Set working directory.
WORKDIR /app

# Copy Redis binaries from redis-binary stage
COPY --from=redis-binary /usr/local/bin/redis-server /usr/local/bin/redis-server
COPY --from=redis-binary /usr/local/bin/redis-cli /usr/local/bin/redis-cli

# Create directories for Redis data and config
RUN mkdir -p /data/redis /etc/redis

# Copy Redis configuration
COPY docker/redis.conf /etc/redis/redis.conf

# Copy startup script
COPY docker/start.sh /app/start.sh
RUN chmod +x /app/start.sh

# Copy build
COPY --from=builder /output/extends-anti-churn /app/main

# Plugin Arch gRPC Server Port.
EXPOSE 6565

# Prometheus /metrics Web Server Port.
EXPOSE 8080

# Entrypoint - use startup script instead of direct binary
ENTRYPOINT ["/app/start.sh"]
