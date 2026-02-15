// Copyright (c) 2025 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package main

import (
	"context"

	"github.com/AccelByte/extends-anti-churn/internal/app"
	"github.com/AccelByte/extends-anti-churn/internal/config"
	"github.com/sirupsen/logrus"
)

func main() {
	// Setup logging
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetLevel(logrus.InfoLevel)
	logrus.Info("starting anti-churn service...")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logrus.Fatalf("failed to load config: %v", err)
	}

	if err := cfg.Validate(); err != nil {
		logrus.Fatalf("invalid config: %v", err)
	}

	// Create and run application
	ctx := context.Background()
	application, err := app.New(ctx, cfg)
	if err != nil {
		logrus.Fatalf("failed to create application: %v", err)
	}

	if err := application.Run(ctx); err != nil {
		logrus.Fatalf("application error: %v", err)
	}
}
