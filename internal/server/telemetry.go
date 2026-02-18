// Copyright (c) 2025 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package server

import (
	"context"
	"fmt"

	"github.com/AccelByte/extend-churn-intervention/pkg/common"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/contrib/propagators/b3"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

// SetupTelemetry initializes OpenTelemetry tracer and propagators.
// Returns a shutdown function that should be called on application shutdown.
//
// ============================================================
// DEVELOPER: OpenTelemetry configuration
// ============================================================
// This sets up distributed tracing with OpenTelemetry.
// Traces are exported based on environment variables:
// - OTEL_EXPORTER_OTLP_ENDPOINT: OTLP collector endpoint
// - OTEL_EXPORTER_OTLP_PROTOCOL: Export protocol (grpc/http)
//
// The tracer provider configuration is in pkg/common/telemetry.go
// Modify that file to:
// - Change sampling strategy
// - Add custom resource attributes
// - Configure batch processing
// - Set up multiple exporters (Jaeger, Zipkin, etc.)
//
// This function configures trace context propagation to support
// distributed tracing across service boundaries using:
// - B3 (Zipkin) propagation
// - W3C TraceContext propagation
// - W3C Baggage propagation
// ============================================================
func SetupTelemetry(ctx context.Context, serviceName, environment string, id int) (func(context.Context) error, error) {
	// ============================================================
	// Create tracer provider with service metadata
	// ============================================================
	tracerProvider, err := common.NewTracerProvider(serviceName, environment, int64(id))
	if err != nil {
		return nil, fmt.Errorf("failed to create tracer provider: %w", err)
	}

	otel.SetTracerProvider(tracerProvider)
	logrus.Infof("set tracer provider: (name: %s environment: %s id: %d)", serviceName, environment, id)

	// ============================================================
	// Configure trace context propagation
	// ============================================================
	// DEVELOPER: Add custom propagators if needed
	// Example for AWS X-Ray:
	// import "go.opentelemetry.io/contrib/propagators/aws/xray"
	// xray.Propagator{}
	// ============================================================
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			b3.New(),                   // Zipkin B3 propagation
			propagation.TraceContext{}, // W3C Trace Context
			propagation.Baggage{},      // W3C Baggage
		),
	)
	logrus.Infof("set text map propagator")

	// Return cleanup function
	shutdown := func(ctx context.Context) error {
		logrus.Info("shutting down telemetry...")
		if err := tracerProvider.Shutdown(ctx); err != nil {
			return err
		}
		logrus.Info("telemetry stopped")
		return nil
	}

	return shutdown, nil
}
