// Copyright (c) 2025 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/AccelByte/extends-anti-churn/pkg/common"
	"github.com/AccelByte/extends-anti-churn/pkg/handler"
	pb_iam "github.com/AccelByte/extends-anti-churn/pkg/pb/accelbyte-asyncapi/iam/oauth/v1"
	pb_social "github.com/AccelByte/extends-anti-churn/pkg/pb/accelbyte-asyncapi/social/statistic/v1"
	"github.com/AccelByte/extends-anti-churn/pkg/state"

	"github.com/AccelByte/accelbyte-go-sdk/services-api/pkg/factory"
	"github.com/AccelByte/accelbyte-go-sdk/services-api/pkg/service/iam"
	sdkAuth "github.com/AccelByte/accelbyte-go-sdk/services-api/pkg/utils/auth"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	prometheusGrpc "github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/contrib/propagators/b3"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

const (
	environment     = "dev"
	id              = 0
	serviceName     = "ExtendAntiChurnHandler"
	grpcServerPort  = 6565
	metricsPort     = 8080
	metricsEndpoint = "/metrics"
)

func main() {
	logrus.Infof("starting app server..")

	// Load .env file if it exists (for local development)
	// In production (Docker/K8s), environment variables are injected directly
	if err := godotenv.Load(); err != nil {
		logrus.Warnf("no .env file found or error loading it: %v (this is normal in production)", err)
	} else {
		logrus.Infof("loaded environment variables from .env file")
	}

	ctx := context.Background()

	// Configure logging
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetLevel(logrus.InfoLevel)

	// gRPC interceptors
	unaryServerInterceptors := []grpc.UnaryServerInterceptor{
		logging.UnaryServerInterceptor(common.InterceptorLogger(logrus.StandardLogger())),
	}
	streamServerInterceptors := []grpc.StreamServerInterceptor{
		logging.StreamServerInterceptor(common.InterceptorLogger(logrus.StandardLogger())),
	}

	// Create repositories
	configRepo := sdkAuth.DefaultConfigRepositoryImpl()
	tokenRepo := sdkAuth.DefaultTokenRepositoryImpl()
	var refreshRepo = &sdkAuth.RefreshTokenImpl{AutoRefresh: true, RefreshRate: 0.8}

	// Create gRPC Server
	s := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.ChainUnaryInterceptor(unaryServerInterceptors...),
		grpc.ChainStreamInterceptor(streamServerInterceptors...),
	)

	// Configure IAM authorization
	oauthService := iam.OAuth20Service{
		Client:                 factory.NewIamClient(configRepo),
		ConfigRepository:       configRepo,
		TokenRepository:        tokenRepo,
		RefreshTokenRepository: refreshRepo,
	}
	clientId := configRepo.GetClientId()
	clientSecret := configRepo.GetClientSecret()
	err := oauthService.LoginClient(&clientId, &clientSecret)
	if err != nil {
		logrus.Fatalf("Error unable to login using clientId and clientSecret: %v", err)
	}

	namespace := common.GetEnv("AB_NAMESPACE", "accelbyte")
	logrus.Infof("using namespace: %s", namespace)

	// Initialize Redis client
	redisClient, err := state.InitRedisClient(ctx)
	if err != nil {
		logrus.Fatalf("failed to initialize Redis client: %v", err)
	}
	defer redisClient.Close()
	logrus.Infof("Redis client initialized")

	// Register event listeners
	oauthListener := handler.NewOAuth(redisClient, namespace)
	pb_iam.RegisterOauthTokenOauthTokenGeneratedServiceServer(s, oauthListener)

	statisticListener := handler.NewStatistic(configRepo, tokenRepo, redisClient, namespace)
	pb_social.RegisterStatisticStatItemUpdatedServiceServer(s, statisticListener)

	logrus.Infof("registered event listeners: OAuth and Statistic")

	// Enable gRPC Reflection
	reflection.Register(s)
	logrus.Infof("gRPC reflection enabled")

	// Enable gRPC Health Check
	grpc_health_v1.RegisterHealthServer(s, health.NewServer())

	// Register Prometheus Metrics
	prometheusRegistry := prometheus.NewRegistry()
	prometheusRegistry.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		prometheusGrpc.NewCounter(prometheus.CounterOpts{
			Name: "grpc_requests_total",
			Help: "Total number of gRPC requests",
		}),
	)

	go func() {
		http.Handle(metricsEndpoint, promhttp.HandlerFor(prometheusRegistry, promhttp.HandlerOpts{}))
		logrus.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", metricsPort), nil))
	}()
	logrus.Infof("serving prometheus metrics at: (:%d%s)", metricsPort, metricsEndpoint)

	// Save Tracer Provider
	tracerProvider, err := common.NewTracerProvider(serviceName, environment, id)
	if err != nil {
		logrus.Fatalf("failed to create tracer provider: %v", err)
		return
	}
	otel.SetTracerProvider(tracerProvider)
	defer func(ctx context.Context) {
		if err := tracerProvider.Shutdown(ctx); err != nil {
			logrus.Fatal(err)
		}
	}(ctx)
	logrus.Infof("set tracer provider: (name: %s environment: %s id: %d)", serviceName, environment, id)

	// Save Text Map Propagator
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			b3.New(),
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	)
	logrus.Infof("set text map propagator")

	// Start gRPC Server
	logrus.Infof("starting gRPC server..")
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", grpcServerPort))
	if err != nil {
		logrus.Fatalf("failed to listen to tcp:%d: %v", grpcServerPort, err)
		return
	}
	go func() {
		if err = s.Serve(lis); err != nil {
			logrus.Fatalf("failed to run gRPC server: %v", err)
			return
		}
	}()
	logrus.Infof("gRPC server started on port %d", grpcServerPort)
	logrus.Infof("app server started")

	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()
	logrus.Infof("signal received, shutting down gracefully")
}
