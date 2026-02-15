// Copyright (c) 2025 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package server

import (
	"context"
	"fmt"
	"net"

	"github.com/AccelByte/extends-anti-churn/pkg/common"
	"github.com/AccelByte/extends-anti-churn/pkg/handler"
	pb_iam "github.com/AccelByte/extends-anti-churn/pkg/pb/accelbyte-asyncapi/iam/oauth/v1"
	pb_social "github.com/AccelByte/extends-anti-churn/pkg/pb/accelbyte-asyncapi/social/statistic/v1"
	"github.com/AccelByte/extends-anti-churn/pkg/pipeline"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

// GRPCServer manages the gRPC server lifecycle.
type GRPCServer struct {
	server    *grpc.Server
	port      int
	manager   *pipeline.Manager
	namespace string
}

// NewGRPCServer creates a new gRPC server instance.
func NewGRPCServer(port int, manager *pipeline.Manager, namespace string) *GRPCServer {
	return &GRPCServer{
		port:      port,
		manager:   manager,
		namespace: namespace,
	}
}

// Setup configures the gRPC server with interceptors and registers handlers.
//
// ============================================================
// DEVELOPER: gRPC server configuration
// ============================================================
// This method sets up:
// 1. Interceptors (logging, auth, rate limiting, etc.)
// 2. Event handlers (OAuth, Statistic, etc.)
// 3. Server features (reflection, health checks)
// ============================================================
func (s *GRPCServer) Setup() error {
	// ============================================================
	// DEVELOPER: Add custom gRPC interceptors here
	// ============================================================
	// Interceptors wrap all gRPC calls for cross-cutting concerns.
	// Add interceptors for:
	// - Authentication/authorization
	// - Rate limiting
	// - Request validation
	// - Custom logging
	// - Error handling
	//
	// Example:
	// unaryInterceptors = append(unaryInterceptors, myAuthInterceptor)
	// ============================================================
	unaryInterceptors := []grpc.UnaryServerInterceptor{
		logging.UnaryServerInterceptor(common.InterceptorLogger(logrus.StandardLogger())),
	}
	streamInterceptors := []grpc.StreamServerInterceptor{
		logging.StreamServerInterceptor(common.InterceptorLogger(logrus.StandardLogger())),
	}

	// Create server with OpenTelemetry instrumentation
	s.server = grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.ChainUnaryInterceptor(unaryInterceptors...),
		grpc.ChainStreamInterceptor(streamInterceptors...),
	)

	// ============================================================
	// DEVELOPER: Register event handlers here
	// ============================================================
	// Event handlers receive gRPC events from Kafka Connect and
	// feed them into the pipeline.
	//
	// To add a new event type:
	// 1. Generate protobuf definitions (make proto)
	// 2. Create a handler in pkg/handler/ (see oauth.go, statistic.go)
	// 3. Register the handler below
	//
	// Example:
	// sessionHandler := handler.NewSession(s.manager, s.namespace)
	// pb_session.RegisterSessionServiceServer(s.server, sessionHandler)
	// ============================================================
	oauthHandler := handler.NewOAuth(s.manager, s.namespace)
	pb_iam.RegisterOauthTokenOauthTokenGeneratedServiceServer(s.server, oauthHandler)

	statisticHandler := handler.NewStatistic(s.manager, s.namespace)
	pb_social.RegisterStatisticStatItemUpdatedServiceServer(s.server, statisticHandler)

	logrus.Infof("registered event listeners: OAuth and Statistic")

	// ============================================================
	// Enable gRPC server features
	// ============================================================
	// - Reflection: allows tools like grpcurl to inspect services
	// - Health check: for Kubernetes liveness/readiness probes
	// ============================================================
	reflection.Register(s.server)
	grpc_health_v1.RegisterHealthServer(s.server, health.NewServer())

	logrus.Infof("gRPC reflection and health check enabled")

	return nil
}

// Start begins listening and serving gRPC requests.
func (s *GRPCServer) Start(ctx context.Context) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		return fmt.Errorf("failed to listen on port %d: %w", s.port, err)
	}

	go func() {
		logrus.Infof("gRPC server listening on port %d", s.port)
		if err := s.server.Serve(lis); err != nil {
			logrus.Fatalf("gRPC server failed: %v", err)
		}
	}()

	return nil
}

// Shutdown gracefully stops the gRPC server.
func (s *GRPCServer) Shutdown(ctx context.Context) error {
	logrus.Info("shutting down gRPC server...")
	s.server.GracefulStop()
	logrus.Info("gRPC server stopped")
	return nil
}
