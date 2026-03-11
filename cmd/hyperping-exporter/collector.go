// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/develeap/terraform-provider-hyperping/internal/client"
)

const namespace = "hyperping"

// HyperpingAPI defines the Hyperping API methods used by the collector.
type HyperpingAPI interface {
	ListMonitors(ctx context.Context) ([]client.Monitor, error)
	ListHealthchecks(ctx context.Context) ([]client.Healthcheck, error)
}

// Collector fetches Hyperping data on a background timer and serves
// cached results as Prometheus metrics. It implements prometheus.Collector.
type Collector struct {
	api      HyperpingAPI
	cacheTTL time.Duration
	logger   *slog.Logger

	// Cache (protected by mu).
	mu             sync.RWMutex
	monitors       []client.Monitor
	healthchecks   []client.Healthcheck
	lastScrapeTime time.Time
	lastScrapeOK   bool
	lastScrapeDur  time.Duration

	// Metric descriptors (immutable after construction).
	monitorUp            *prometheus.Desc
	monitorPaused        *prometheus.Desc
	monitorSSLExpDays    *prometheus.Desc
	monitorCheckInterval *prometheus.Desc
	monitorInfo          *prometheus.Desc
	healthcheckUp        *prometheus.Desc
	healthcheckPaused    *prometheus.Desc
	healthcheckPeriod    *prometheus.Desc
	monitorsTotal        *prometheus.Desc
	healthchecksTotal    *prometheus.Desc
	scrapeDurationDesc   *prometheus.Desc
	scrapeSuccessDesc    *prometheus.Desc
}

// Verify Collector implements prometheus.Collector at compile time.
var _ prometheus.Collector = (*Collector)(nil)

// NewCollector creates a new Hyperping metrics collector.
func NewCollector(api HyperpingAPI, cacheTTL time.Duration, logger *slog.Logger) *Collector {
	monitorLabels := []string{"uuid", "name"}
	healthcheckLabels := []string{"uuid", "name"}

	return &Collector{
		api:      api,
		cacheTTL: cacheTTL,
		logger:   logger,

		monitorUp: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "monitor", "up"),
			"Whether the monitor is up (1) or down (0).",
			monitorLabels, nil,
		),
		monitorPaused: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "monitor", "paused"),
			"Whether the monitor is paused (1) or active (0).",
			monitorLabels, nil,
		),
		monitorSSLExpDays: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "monitor", "ssl_expiration_days"),
			"Days until SSL certificate expiration.",
			monitorLabels, nil,
		),
		monitorCheckInterval: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "monitor", "check_interval_seconds"),
			"Monitor check frequency in seconds.",
			monitorLabels, nil,
		),
		monitorInfo: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "monitor", "info"),
			"Monitor metadata (value is always 1).",
			[]string{"uuid", "name", "protocol", "url", "project_uuid", "http_method"}, nil,
		),
		healthcheckUp: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "healthcheck", "up"),
			"Whether the healthcheck is up (1) or down (0).",
			healthcheckLabels, nil,
		),
		healthcheckPaused: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "healthcheck", "paused"),
			"Whether the healthcheck is paused (1) or active (0).",
			healthcheckLabels, nil,
		),
		healthcheckPeriod: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "healthcheck", "period_seconds"),
			"Expected healthcheck ping period in seconds.",
			healthcheckLabels, nil,
		),
		monitorsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "monitors"),
			"Total number of monitors.",
			nil, nil,
		),
		healthchecksTotal: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "healthchecks"),
			"Total number of healthchecks.",
			nil, nil,
		),
		scrapeDurationDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "scrape", "duration_seconds"),
			"Duration of the last API scrape in seconds.",
			nil, nil,
		),
		scrapeSuccessDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "scrape", "success"),
			"Whether the last API scrape succeeded (1) or failed (0).",
			nil, nil,
		),
	}
}

// Start begins the background cache refresh loop. It blocks until ctx is cancelled.
func (c *Collector) Start(ctx context.Context) {
	c.Refresh(ctx)

	ticker := time.NewTicker(c.cacheTTL)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.Refresh(ctx)
		}
	}
}

// Refresh performs a single API scrape and updates the cache.
// Both ListMonitors and ListHealthchecks are called in parallel.
func (c *Collector) Refresh(ctx context.Context) {
	start := time.Now()

	var (
		monitors     []client.Monitor
		healthchecks []client.Healthcheck
		monErr       error
		hcErr        error
		wg           sync.WaitGroup
	)

	wg.Add(2)
	go func() {
		defer wg.Done()
		monitors, monErr = c.api.ListMonitors(ctx)
	}()
	go func() {
		defer wg.Done()
		healthchecks, hcErr = c.api.ListHealthchecks(ctx)
	}()
	wg.Wait()

	dur := time.Since(start)

	c.mu.Lock()
	defer c.mu.Unlock()

	c.lastScrapeTime = time.Now()
	c.lastScrapeDur = dur

	if monErr != nil {
		c.logger.Error("failed to list monitors", "error", monErr)
		c.lastScrapeOK = false
		return
	}
	if hcErr != nil {
		c.logger.Error("failed to list healthchecks", "error", hcErr)
		c.lastScrapeOK = false
		return
	}

	c.monitors = monitors
	c.healthchecks = healthchecks
	c.lastScrapeOK = true

	c.logger.Info("cache refreshed",
		"monitors", len(monitors),
		"healthchecks", len(healthchecks),
		"duration", dur,
	)
}

// IsReady returns true after at least one successful API scrape.
func (c *Collector) IsReady() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.lastScrapeOK
}

// Describe implements prometheus.Collector.
func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.monitorUp
	ch <- c.monitorPaused
	ch <- c.monitorSSLExpDays
	ch <- c.monitorCheckInterval
	ch <- c.monitorInfo
	ch <- c.healthcheckUp
	ch <- c.healthcheckPaused
	ch <- c.healthcheckPeriod
	ch <- c.monitorsTotal
	ch <- c.healthchecksTotal
	ch <- c.scrapeDurationDesc
	ch <- c.scrapeSuccessDesc
}

// Collect implements prometheus.Collector.
func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for _, m := range c.monitors {
		ch <- prometheus.MustNewConstMetric(c.monitorUp, prometheus.GaugeValue,
			boolToFloat64(m.Status == "up"), m.UUID, m.Name)
		ch <- prometheus.MustNewConstMetric(c.monitorPaused, prometheus.GaugeValue,
			boolToFloat64(m.Paused), m.UUID, m.Name)
		ch <- prometheus.MustNewConstMetric(c.monitorCheckInterval, prometheus.GaugeValue,
			float64(m.CheckFrequency), m.UUID, m.Name)
		ch <- prometheus.MustNewConstMetric(c.monitorInfo, prometheus.GaugeValue, 1,
			m.UUID, m.Name, m.Protocol, m.URL, m.ProjectUUID, m.HTTPMethod)

		if m.SSLExpiration != nil {
			ch <- prometheus.MustNewConstMetric(c.monitorSSLExpDays, prometheus.GaugeValue,
				float64(*m.SSLExpiration), m.UUID, m.Name)
		}
	}

	for _, hc := range c.healthchecks {
		ch <- prometheus.MustNewConstMetric(c.healthcheckUp, prometheus.GaugeValue,
			boolToFloat64(!hc.IsDown), hc.UUID, hc.Name)
		ch <- prometheus.MustNewConstMetric(c.healthcheckPaused, prometheus.GaugeValue,
			boolToFloat64(hc.IsPaused), hc.UUID, hc.Name)
		ch <- prometheus.MustNewConstMetric(c.healthcheckPeriod, prometheus.GaugeValue,
			float64(hc.Period), hc.UUID, hc.Name)
	}

	ch <- prometheus.MustNewConstMetric(c.monitorsTotal, prometheus.GaugeValue,
		float64(len(c.monitors)))
	ch <- prometheus.MustNewConstMetric(c.healthchecksTotal, prometheus.GaugeValue,
		float64(len(c.healthchecks)))
	ch <- prometheus.MustNewConstMetric(c.scrapeDurationDesc, prometheus.GaugeValue,
		c.lastScrapeDur.Seconds())
	ch <- prometheus.MustNewConstMetric(c.scrapeSuccessDesc, prometheus.GaugeValue,
		boolToFloat64(c.lastScrapeOK))
}

func boolToFloat64(b bool) float64 {
	if b {
		return 1
	}
	return 0
}
