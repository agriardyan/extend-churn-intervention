// Copyright (c) 2025 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package config

import (
	"fmt"

	"github.com/caarlos0/env/v10"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

// Load reads configuration from environment variables.
// It attempts to load from .env file first (for local development),
// then parses environment variables into the Config struct.
func Load() (*Config, error) {
	// Load .env file if it exists (for local development)
	// In production (Docker/K8s), environment variables are injected directly
	if err := godotenv.Load(); err != nil {
		logrus.Warnf("no .env file found or error loading it: %v (this is normal in production)", err)
	} else {
		logrus.Infof("loaded environment variables from .env file")
	}

	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config from environment: %w", err)
	}

	return cfg, nil
}

// Validate performs custom validation on the configuration.
//
// ============================================================
// DEVELOPER: Add custom validation logic here.
// ============================================================
// This function is called after environment variables are parsed.
// Add validation for:
// - Value ranges (e.g., port numbers must be 1-65535)
// - Business logic constraints
// - Cross-field validation
// - Format validation (URLs, emails, etc.)
//
// Example:
//   if c.MyTimeout < 0 {
//       return fmt.Errorf("timeout must be non-negative")
//   }
// ============================================================
func (c *Config) Validate() error {
	// Validate server ports
	if c.GRPCPort < 1 || c.GRPCPort > 65535 {
		return fmt.Errorf("invalid GRPC_PORT: %d (must be 1-65535)", c.GRPCPort)
	}

	if c.MetricsPort < 1 || c.MetricsPort > 65535 {
		return fmt.Errorf("invalid METRICS_PORT: %d (must be 1-65535)", c.MetricsPort)
	}

	// Validate required fields
	if c.ABNamespace == "" {
		return fmt.Errorf("AB_NAMESPACE is required")
	}

	// ============================================================
	// DEVELOPER: Add your custom validation below
	// ============================================================

	return nil
}
