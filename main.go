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

	"github.com/AccelByte/extends-anti-churn/pkg/action"
	actionBuiltin "github.com/AccelByte/extends-anti-churn/pkg/action/builtin"
	"github.com/AccelByte/extends-anti-churn/pkg/common"
	"github.com/AccelByte/extends-anti-churn/pkg/handler"
	pb_iam "github.com/AccelByte/extends-anti-churn/pkg/pb/accelbyte-asyncapi/iam/oauth/v1"
	pb_social "github.com/AccelByte/extends-anti-churn/pkg/pb/accelbyte-asyncapi/social/statistic/v1"
	"github.com/AccelByte/extends-anti-churn/pkg/pipeline"
	"github.com/AccelByte/extends-anti-churn/pkg/rule"
	ruleBuiltin "github.com/AccelByte/extends-anti-churn/pkg/rule/builtin"
	"github.com/AccelByte/extends-anti-churn/pkg/service"
	signalPkg "github.com/AccelByte/extends-anti-churn/pkg/signal"
	signalBuiltin "github.com/AccelByte/extends-anti-churn/pkg/signal/builtin"
	"github.com/AccelByte/extends-anti-churn/pkg/state"

	"github.com/AccelByte/accelbyte-go-sdk/services-api/pkg/factory"
	"github.com/AccelByte/accelbyte-go-sdk/services-api/pkg/service/iam"
	"github.com/AccelByte/accelbyte-go-sdk/services-api/pkg/service/platform"
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

	// Load pipeline configuration
	configPath := common.GetEnv("CONFIG_PATH", "config/pipeline.yaml")
	pipelineConfig, err := pipeline.LoadConfig(configPath)
	if err != nil {
		logrus.Fatalf("failed to load pipeline configuration from %s: %v", configPath, err)
	}
	logrus.Infof("loaded pipeline configuration from %s", configPath)

	// Initialize signal processor with Redis state store
	redisStateStore := service.NewRedisStateStore(
		redisClient,
		service.RedisStateStoreConfig{},
	)
	processor := signalPkg.NewProcessor(redisStateStore, namespace)

	// ============================================================
	// DEVELOPER: Register your event processors here.
	// Each event processor handles a specific event type or stat code.
	// ============================================================
	signalBuiltin.RegisterEventProcessors(
		processor.GetEventProcessorRegistry(),
		processor.GetStateStore(),
		processor.GetNamespace(),
	)
	logrus.Infof("initialized signal processor with %d event processors",
		processor.GetEventProcessorRegistry().Count())

	// Convert pipeline rule configs to rule package configs
	ruleConfigs := make([]rule.RuleConfig, len(pipelineConfig.Rules))
	for i, rc := range pipelineConfig.Rules {
		ruleConfigs[i] = rule.RuleConfig{
			ID:         rc.ID,
			Type:       rc.Type,
			Enabled:    rc.Enabled,
			Parameters: rc.Parameters,
		}
	}

	// ============================================================
	// DEVELOPER: Register your rule types here.
	// Each rule type defines how to evaluate a signal.
	// ============================================================
	ruleBuiltin.RegisterRules()

	// Initialize rule registry and register rules
	ruleRegistry := rule.NewRegistry()
	if err := rule.RegisterRules(ruleRegistry, ruleConfigs); err != nil {
		logrus.Fatalf("failed to register rules: %v", err)
	}
	logrus.Infof("registered %d rules", len(ruleConfigs))

	// Initialize rule engine
	ruleEngine := rule.NewEngine(ruleRegistry)
	logrus.Infof("initialized rule engine")

	// Initialize action registry and register actions
	actionRegistry := action.NewRegistry()

	// ============================================================
	// DEVELOPER: Register your action types here.
	// Set up dependencies and register action factories.
	// ============================================================
	fulfillmentService := platform.FulfillmentService{
		Client:           factory.NewPlatformClient(configRepo),
		ConfigRepository: configRepo,
		TokenRepository:  tokenRepo,
	}
	itemGranter := service.NewEntitlementService(
		fulfillmentService,
		service.EntitlementServiceConfig{
			Namespace: namespace,
		},
	)
	deps := &actionBuiltin.Dependencies{
		StateStore:         redisStateStore,
		EntitlementGranter: itemGranter,
	}
	actionBuiltin.RegisterActions(deps)

	// Convert pipeline action configs to action package configs
	actionConfigs := make([]action.ActionConfig, len(pipelineConfig.Actions))
	for i, ac := range pipelineConfig.Actions {
		actionConfigs[i] = action.ActionConfig{
			ID:         ac.ID,
			Type:       ac.Type,
			Enabled:    ac.Enabled,
			Parameters: ac.Parameters,
		}
	}

	// Register actions from config
	if err := action.RegisterActions(actionRegistry, actionConfigs); err != nil {
		logrus.Fatalf("failed to register actions: %v", err)
	}
	logrus.Infof("registered %d actions", len(actionConfigs))

	// Initialize action executor
	actionExecutor := action.NewExecutor(actionRegistry)
	logrus.Infof("initialized action executor")

	// Build rule-to-actions mapping from pipeline config
	ruleActions := make(map[string][]string)
	for _, rc := range pipelineConfig.Rules {
		if len(rc.Actions) > 0 {
			ruleActions[rc.ID] = rc.Actions
		}
	}
	logrus.Infof("configured %d rule-to-action mappings", len(ruleActions))

	// Initialize pipeline manager with rule-to-actions mapping
	pipelineManager := pipeline.NewManager(processor, ruleEngine, actionExecutor, ruleActions, nil)
	logrus.Infof("initialized pipeline manager")

	// Validate pipeline wiring - ensures all enabled rules and actions are registered
	if err := pipeline.ValidateWiring(ruleRegistry, actionRegistry, pipelineConfig); err != nil {
		logrus.Fatalf("pipeline wiring validation failed: %v", err)
	}
	logrus.Infof("pipeline wiring validation passed")

	// ============================================================
	// DEVELOPER: Register your gRPC event handlers here.
	// Each handler receives events from a specific AGS service
	// and passes them to the pipeline manager for processing.
	// ============================================================
	oauthListener := handler.NewOAuth(pipelineManager, namespace)
	pb_iam.RegisterOauthTokenOauthTokenGeneratedServiceServer(s, oauthListener)

	statisticListener := handler.NewStatistic(pipelineManager, namespace)
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
