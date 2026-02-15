// Copyright (c) 2025 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package app

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"
)

// Run starts the application and blocks until a shutdown signal is received.
func (a *App) Run(ctx context.Context) error {
	// Start servers
	if err := a.grpcServer.Start(ctx); err != nil {
		return err
	}
	if err := a.metricsServer.Start(ctx); err != nil {
		return err
	}

	logrus.Info("application started successfully")

	// Wait for shutdown signal
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()

	logrus.Info("shutdown signal received")
	return a.Shutdown(ctx)
}

// Shutdown gracefully shuts down all application components.
//
// ============================================================
// DEVELOPER: Shutdown order is critical
// ============================================================
// Components are shut down in reverse dependency order:
// 1. Stop accepting new requests (gRPC + metrics servers)
// 2. Close external connections (Redis, databases)
// 3. Flush telemetry data (OpenTelemetry)
//
// If you added custom services in app.New(), shut them down here
// following this order. For example, if you added a database
// connection, close it after Redis but before telemetry.
//
// IMPORTANT: Shutdown errors are logged but don't stop the
// shutdown sequence. Each component gets a chance to clean up.
// ============================================================
func (a *App) Shutdown(ctx context.Context) error {
	logrus.Info("shutting down application...")

	// ============================================================
	// Step 1: Shutdown servers (stop accepting new requests)
	// ============================================================
	if err := a.grpcServer.Shutdown(ctx); err != nil {
		logrus.Errorf("gRPC server shutdown error: %v", err)
	}
	if err := a.metricsServer.Shutdown(ctx); err != nil {
		logrus.Errorf("metrics server shutdown error: %v", err)
	}

	// ============================================================
	// Step 2: Close external connections
	// ============================================================
	// DEVELOPER: Add custom service cleanup here
	// Example:
	// if a.dbConnection != nil {
	//     if err := a.dbConnection.Close(); err != nil {
	//         logrus.Errorf("database close error: %v", err)
	//     }
	// }
	// ============================================================
	if a.redisClient != nil {
		if err := a.redisClient.Close(); err != nil {
			logrus.Errorf("Redis close error: %v", err)
		}
	}

	// ============================================================
	// Step 3: Flush telemetry data
	// ============================================================
	if a.shutdownTelemetry != nil {
		if err := a.shutdownTelemetry(ctx); err != nil {
			logrus.Errorf("telemetry shutdown error: %v", err)
		}
	}

	logrus.Info("application shutdown complete")
	return nil
}
