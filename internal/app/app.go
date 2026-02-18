// Copyright (c) 2025 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package app

import (
	"context"
	"fmt"
	"time"

	"github.com/AccelByte/extend-churn-intervention/internal/bootstrap"
	"github.com/AccelByte/extend-churn-intervention/internal/config"
	"github.com/AccelByte/extend-churn-intervention/internal/server"
	"github.com/AccelByte/extend-churn-intervention/pkg/pipeline"
	"github.com/AccelByte/extend-churn-intervention/pkg/service"
	"github.com/cenkalti/backoff/v4"

	"github.com/AccelByte/accelbyte-go-sdk/services-api/pkg/factory"
	"github.com/AccelByte/accelbyte-go-sdk/services-api/pkg/service/iam"
	"github.com/AccelByte/accelbyte-go-sdk/services-api/pkg/service/platform"
	"github.com/AccelByte/accelbyte-go-sdk/services-api/pkg/service/social"
	sdkAuth "github.com/AccelByte/accelbyte-go-sdk/services-api/pkg/utils/auth"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"

	actionBuiltin "github.com/AccelByte/extend-churn-intervention/pkg/action/builtin"
)

// App holds all application dependencies and manages the application lifecycle.
type App struct {
	cfg               *config.Config
	grpcServer        *server.GRPCServer
	metricsServer     *server.MetricsServer
	redisClient       *redis.Client
	shutdownTelemetry func(context.Context) error

	// AccelByte SDK repositories (shared across all services)
	configRepo *sdkAuth.ConfigRepositoryImpl
	tokenRepo  *sdkAuth.TokenRepositoryImpl
}

// New creates and initializes a new application instance.
//
// ============================================================
// DEVELOPER: Application initialization order
// ============================================================
// Components are initialized in dependency order:
// 1. AccelByte SDK (required for Platform API calls)
// 2. Redis (required for state storage)
// 3. Pipeline config (YAML configuration)
// 4. External services (state store, item granter, etc.)
// 5. Pipeline components (signal → rule → action)
// 6. Servers (gRPC, metrics)
// 7. Telemetry (OpenTelemetry tracing)
//
// If you add new external dependencies, initialize them in
// step 4 before bootstrapping pipeline components.
// ============================================================
func New(ctx context.Context, cfg *config.Config) (*App, error) {
	logrus.Info("initializing application...")

	app := &App{cfg: cfg}

	// ============================================================
	// Step 1: Initialize Client Auth using AccelByte SDK
	// ============================================================
	if err := app.initAccelByteSDKAuth(); err != nil {
		return nil, fmt.Errorf("failed to init AccelByte SDK: %w", err)
	}

	// ============================================================
	// Step 2: Initialize Redis
	// ============================================================
	if err := app.initRedis(ctx); err != nil {
		return nil, fmt.Errorf("failed to init Redis: %w", err)
	}

	// ============================================================
	// Step 3: Load pipeline configuration
	// ============================================================
	pipelineConfig, err := pipeline.LoadConfig(cfg.ConfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load pipeline config from %s: %w", cfg.ConfigPath, err)
	}
	logrus.Infof("loaded pipeline configuration from %s", cfg.ConfigPath)

	// ============================================================
	// Step 4: Initialize external services
	// ============================================================
	// DEVELOPER: Add custom external service initialization here.
	// Examples:
	// - notificationService := app.initNotificationService()
	// - analyticsClient := app.initAnalyticsClient()
	// - dbConnection := app.initDatabase(ctx)
	//
	// Then pass these services to the bootstrap functions below.
	// ============================================================
	stateStore := service.NewRedisChurnStateStore(app.redisClient, service.RedisChurnStateStoreConfig{})
	loginTrackingStore := service.NewRedisLoginSessionTrackingStore(app.redisClient, service.RedisLoginSessionTrackingStoreConfig{})
	itemGranter := app.initItemGranter()
	userStatUpdater := app.initStatisticService()

	// ============================================================
	// Step 5: Bootstrap pipeline components
	// ============================================================
	// The pipeline is built in this order:
	// Signal Processor → Rule Engine → Action Executor → Pipeline Manager
	//
	// DEVELOPER: If your custom actions need external services,
	// wrap them in signalBuiltin.Dependencies and pass them to InitSignalProcessor (see bootstrap/signal.go).
	// ============================================================
	processor := bootstrap.InitSignalProcessor(
		stateStore,
		loginTrackingStore,
		cfg.ABNamespace,
	)

	ruleEngine, ruleRegistry, err := bootstrap.InitRuleEngine(pipelineConfig, loginTrackingStore)
	if err != nil {
		return nil, fmt.Errorf("failed to init rule engine: %w", err)
	}

	// ============================================================
	// DEVELOPER: Action dependencies setup
	// ============================================================
	// If your custom actions need external services, add them
	// to the Dependencies struct in pkg/action/builtin/init.go
	// and pass them here.
	// ============================================================
	deps := &actionBuiltin.Dependencies{
		StateStore:         stateStore,
		EntitlementGranter: itemGranter,
		UserStatUpdater:    userStatUpdater,
		// DEVELOPER: Add custom service dependencies here
		// Example: NotificationService: myNotificationService,
	}

	actionExecutor, actionRegistry, err := bootstrap.InitActionExecutor(pipelineConfig, deps)
	if err != nil {
		return nil, fmt.Errorf("failed to init action executor: %w", err)
	}

	pipelineManager := bootstrap.InitPipeline(processor, ruleEngine, actionExecutor, pipelineConfig)

	// ============================================================
	// Validate pipeline wiring
	// ============================================================
	// This ensures all rule-to-action mappings in config/pipeline.yaml
	// reference valid rule and action IDs.
	// ============================================================
	if err := pipeline.ValidateWiring(ruleRegistry, actionRegistry, pipelineConfig); err != nil {
		return nil, fmt.Errorf("pipeline wiring validation failed: %w", err)
	}
	logrus.Info("pipeline wiring validation passed")

	// ============================================================
	// Step 6: Setup servers
	// ============================================================
	app.grpcServer = server.NewGRPCServer(cfg.GRPCPort, pipelineManager, cfg.ABNamespace)
	if err := app.grpcServer.Setup(); err != nil {
		return nil, fmt.Errorf("failed to setup gRPC server: %w", err)
	}

	app.metricsServer = server.NewMetricsServer(cfg.MetricsPort, "/metrics")
	if err := app.metricsServer.Setup(); err != nil {
		return nil, fmt.Errorf("failed to setup metrics server: %w", err)
	}

	// ============================================================
	// Step 7: Setup telemetry
	// ============================================================
	shutdownTelemetry, err := server.SetupTelemetry(ctx, cfg.ServiceName, cfg.Environment, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to setup telemetry: %w", err)
	}
	app.shutdownTelemetry = shutdownTelemetry

	logrus.Info("application initialized successfully")

	return app, nil
}

// initAccelByteSDKAuth initializes the AccelByte SDK auth by performing client login.
//
// ============================================================
// DEVELOPER: AccelByte Client Auth configuration
// ============================================================
// The Client Auth is configured via environment variables:
// - AB_BASE_URL: AccelByte platform base URL
// - AB_CLIENT_ID: OAuth2 client ID
// - AB_CLIENT_SECRET: OAuth2 client secret
// - AB_NAMESPACE: Game namespace
//
// The SDK uses automatic token refresh (RefreshRate: 0.8 = 80% of TTL).
//
// IMPORTANT: The configRepo and tokenRepo are stored in the App struct
// and must be reused by all AccelByte services to share authentication.
// ============================================================
func (a *App) initAccelByteSDKAuth() error {
	a.configRepo = sdkAuth.DefaultConfigRepositoryImpl()
	a.tokenRepo = sdkAuth.DefaultTokenRepositoryImpl()
	refreshRepo := &sdkAuth.RefreshTokenImpl{AutoRefresh: true, RefreshRate: 0.8}

	oauthService := iam.OAuth20Service{
		Client:                 factory.NewIamClient(a.configRepo),
		ConfigRepository:       a.configRepo,
		TokenRepository:        a.tokenRepo,
		RefreshTokenRepository: refreshRepo,
	}

	clientID := a.configRepo.GetClientId()
	clientSecret := a.configRepo.GetClientSecret()

	if err := oauthService.LoginClient(&clientID, &clientSecret); err != nil {
		return fmt.Errorf("unable to login using clientId and clientSecret: %w", err)
	}

	logrus.Info("AccelByte SDK initialized and authenticated")
	return nil
}

// initRedis initializes the Redis client.
func (a *App) initRedis(ctx context.Context) error {
	client := redis.NewClient(&redis.Options{
		Addr:         a.cfg.RedisHost + ":" + a.cfg.RedisPort,
		Password:     a.cfg.RedisPassword,
		DB:           0, // use default DB
		MaxRetries:   3,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	b := backoff.NewExponentialBackOff()
	maxRetries := backoff.WithMaxRetries(b, 5)

	err := backoff.Retry(
		func() error {
			_, err := client.Ping(ctx).Result()
			if err != nil {
				logrus.Warnf("Redis connection failed: %v, retrying...", err)
				return err
			}
			return nil
		},
		maxRetries,
	)

	if err != nil {
		return err
	}

	a.redisClient = client
	logrus.Info("Redis client initialized")
	return nil
}

// ============================================================
// DEVELOPER: Add custom service initializers here
// ============================================================
// If your custom actions require additional AccelByte Platform
// services, add initialization methods similar to initItemGranter or initStatisticService.
//
// IMPORTANT: Always reuse a.configRepo and a.tokenRepo to share the
// authenticated session. Do NOT call DefaultConfigRepositoryImpl() or
// DefaultTokenRepositoryImpl() again - this creates new empty instances!
//
// Example:
//
//	func (a *App) initNotificationService() service.NotificationSender {
//	    lobbyService := lobby.NotificationService{
//	        Client:           factory.NewLobbyClient(a.configRepo),
//	        ConfigRepository: a.configRepo,
//	        TokenRepository:  a.tokenRepo,  // ← Reuse existing repos!
//	    }
//	    return service.NewNotificationService(lobbyService, ...)
//	}
//
// ============================================================

// initItemGranter creates an entitlement service for granting items.
//
// IMPORTANT: Reuses a.configRepo and a.tokenRepo to share the authenticated
// session from initAccelByteSDK(). Do NOT create new repository instances.
func (a *App) initItemGranter() service.EntitlementGranter {
	fulfillmentService := &platform.FulfillmentService{
		Client:           factory.NewPlatformClient(a.configRepo),
		ConfigRepository: a.configRepo,
		TokenRepository:  a.tokenRepo,
	}

	return service.NewEntitlementService(fulfillmentService, service.EntitlementServiceConfig{
		Namespace: a.cfg.ABNamespace,
	})
}

// initStatisticService initializes the statistic service client.
//
// IMPORTANT: Reuses a.configRepo and a.tokenRepo to share the authenticated
// session from initAccelByteSDK(). Do NOT create new repository instances.
func (a *App) initStatisticService() service.UserStatisticUpdater {
	statisticService := &social.UserStatisticService{
		Client:           factory.NewSocialClient(a.configRepo),
		ConfigRepository: a.configRepo,
		TokenRepository:  a.tokenRepo,
	}

	return service.NewStatisticService(statisticService,
		service.StatisticServiceConfig{
			Namespace: a.cfg.ABNamespace,
		})
}
