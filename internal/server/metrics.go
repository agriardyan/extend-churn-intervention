// Copyright (c) 2025 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

// MetricsServer manages the Prometheus metrics HTTP server.
type MetricsServer struct {
	server   *http.Server
	port     int
	endpoint string
}

// NewMetricsServer creates a new metrics server instance.
func NewMetricsServer(port int, endpoint string) *MetricsServer {
	return &MetricsServer{
		port:     port,
		endpoint: endpoint,
	}
}

// Setup configures the metrics server and registers collectors.
//
// ============================================================
// DEVELOPER: Register custom Prometheus metrics here
// ============================================================
// By default, we expose Go runtime and process metrics.
// To add custom application metrics:
//
// 1. Define your metrics in a separate package (e.g., pkg/metrics/)
//    Example:
//    var RuleTriggersTotal = prometheus.NewCounterVec(
//        prometheus.CounterOpts{
//            Name: "anti_churn_rule_triggers_total",
//            Help: "Total number of rule triggers",
//        },
//        []string{"rule_id", "rule_type"},
//    )
//
// 2. Register them here:
//    registry.MustRegister(metrics.RuleTriggersTotal)
//
// 3. Increment them in your code:
//    metrics.RuleTriggersTotal.WithLabelValues(ruleID, ruleType).Inc()
//
// See: https://prometheus.io/docs/guides/go-application/
// ============================================================
func (m *MetricsServer) Setup() error {
	registry := prometheus.NewRegistry()

	// Register default collectors
	registry.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)

	// ============================================================
	// DEVELOPER: Register custom metrics below
	// ============================================================
	// Example:
	// registry.MustRegister(metrics.RuleTriggersTotal)
	// registry.MustRegister(metrics.ActionExecutionDuration)
	// ============================================================

	mux := http.NewServeMux()
	mux.Handle(m.endpoint, promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))

	m.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", m.port),
		Handler: mux,
	}

	return nil
}

// Start begins serving metrics on the configured port.
func (m *MetricsServer) Start(ctx context.Context) error {
	go func() {
		logrus.Infof("metrics server listening on port %d%s", m.port, m.endpoint)
		if err := m.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Fatalf("metrics server failed: %v", err)
		}
	}()
	return nil
}

// Shutdown gracefully stops the metrics server.
func (m *MetricsServer) Shutdown(ctx context.Context) error {
	logrus.Info("shutting down metrics server...")
	if err := m.server.Shutdown(ctx); err != nil {
		return err
	}
	logrus.Info("metrics server stopped")
	return nil
}
