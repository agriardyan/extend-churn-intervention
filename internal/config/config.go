// Copyright (c) 2025 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package config

// Config holds all application configuration loaded from environment variables.
// This struct uses github.com/caarlos0/env for automatic environment variable parsing.
//
// ============================================================
// DEVELOPER: Add new configuration fields here.
// ============================================================
// Use struct tags to define:
// - `env:"VAR_NAME"` - the environment variable name
// - `env:",required"` - make it required
// - `envDefault:"value"` - set a default value
//
// Example:
//   NewFeature bool `env:"ENABLE_NEW_FEATURE" envDefault:"false"`
//
// After adding fields here, update loader.go Validate() if custom
// validation is needed.
// ============================================================
type Config struct {
	// ============================================================
	// Server configuration
	// ============================================================
	GRPCPort    int    `env:"GRPC_PORT" envDefault:"6565"`
	MetricsPort int    `env:"METRICS_PORT" envDefault:"8080"`
	Environment string `env:"ENVIRONMENT" envDefault:"dev"`
	ServiceName string `env:"SERVICE_NAME" envDefault:"ExtendAntiChurnHandler"`

	// ============================================================
	// AccelByte configuration (REQUIRED)
	// ============================================================
	ABNamespace    string `env:"AB_NAMESPACE,required"`
	ABBaseURL      string `env:"AB_BASE_URL,required"`
	ABClientID     string `env:"AB_CLIENT_ID,required"`
	ABClientSecret string `env:"AB_CLIENT_SECRET,required"`

	// ============================================================
	// Redis configuration
	// ============================================================
	RedisHost         string `env:"REDIS_HOST" envDefault:"localhost"`
	RedisPort         string `env:"REDIS_PORT" envDefault:"6379"`
	RedisPassword     string `env:"REDIS_PASSWORD"`
	RedisMaxRetries   int    `env:"REDIS_MAX_RETRIES" envDefault:"5"`
	RedisRetryDelayMs int    `env:"REDIS_RETRY_DELAY_MS" envDefault:"1000"`

	// ============================================================
	// Pipeline configuration
	// ============================================================
	ConfigPath string `env:"CONFIG_PATH" envDefault:"config/pipeline.yaml"`

	// ============================================================
	// Telemetry configuration
	// ============================================================
	OtelEnabled     bool   `env:"OTEL_ENABLED" envDefault:"true"`
	OtelEndpoint    string `env:"OTEL_EXPORTER_OTLP_ENDPOINT"`
	OtelServiceName string `env:"OTEL_SERVICE_NAME" envDefault:"extends-anti-churn"`

	// ============================================================
	// DEVELOPER: Add your custom configuration fields below
	// ============================================================
	// Example:
	// MyCustomFeature string `env:"MY_CUSTOM_FEATURE" envDefault:"default-value"`
	// ============================================================
}
