// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/develeap/terraform-provider-hyperping/internal/client"
)

var version = "dev"

func main() {
	os.Exit(run())
}

func run() int {
	var (
		listenAddr  = flag.String("listen-address", ":9312", "Address to listen on for metrics")
		metricsPath = flag.String("metrics-path", "/metrics", "Path under which to expose metrics")
		apiKey      = flag.String("api-key", "", "Hyperping API key (env: HYPERPING_API_KEY)")
		cacheTTL    = flag.Duration("cache-ttl", 60*time.Second, "How often to refresh data from the API")
		logLevel    = flag.String("log-level", "info", "Log level (debug, info, warn, error)")
		logFormat   = flag.String("log-format", "text", "Log format (text, json)")
	)
	flag.Parse()

	if *apiKey == "" {
		*apiKey = os.Getenv("HYPERPING_API_KEY")
	}
	if *apiKey == "" {
		fmt.Fprintln(os.Stderr, "error: API key required (use --api-key or HYPERPING_API_KEY)")
		return 1
	}

	logger := setupLogger(*logLevel, *logFormat)

	apiClient := client.NewClient(*apiKey, client.WithMaxRetries(2))

	collector := NewCollector(apiClient, *cacheTTL, logger)
	registry := prometheus.NewRegistry()
	registry.MustRegister(collector)
	registry.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	registry.MustRegister(collectors.NewGoCollector())

	mux := http.NewServeMux()
	mux.Handle(*metricsPath, promhttp.HandlerFor(registry, promhttp.HandlerOpts{
		EnableOpenMetrics: true,
	}))
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "ok")
	})
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, _ *http.Request) {
		if collector.IsReady() {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, "ready")
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprintln(w, "not ready")
		}
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, `<html><head><title>Hyperping Exporter</title></head>
<body><h1>Hyperping Exporter</h1><p><a href="%s">Metrics</a></p>
<p>Version: %s</p></body></html>`, *metricsPath, version)
	})

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go collector.Start(ctx)

	srv := &http.Server{
		Addr:              *listenAddr,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		<-ctx.Done()
		logger.Info("shutting down")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			logger.Error("shutdown error", "error", err)
		}
	}()

	logger.Info("starting hyperping exporter",
		"version", version,
		"address", *listenAddr,
		"metrics_path", *metricsPath,
		"cache_ttl", *cacheTTL,
	)

	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		logger.Error("server error", "error", err)
		return 1
	}
	return 0
}

func setupLogger(level, format string) *slog.Logger {
	var lvl slog.Level
	switch level {
	case "debug":
		lvl = slog.LevelDebug
	case "warn":
		lvl = slog.LevelWarn
	case "error":
		lvl = slog.LevelError
	default:
		lvl = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{Level: lvl}
	var handler slog.Handler
	if format == "json" {
		handler = slog.NewJSONHandler(os.Stderr, opts)
	} else {
		handler = slog.NewTextHandler(os.Stderr, opts)
	}
	return slog.New(handler)
}
