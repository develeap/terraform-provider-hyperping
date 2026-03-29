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

// reportPeriods defines the SLA/outage report windows fetched on each refresh.
var reportPeriods = []string{"24h", "7d", "30d"}

// reportDurations maps period labels to their lookback durations.
var reportDurations = map[string]time.Duration{
	"24h": 24 * time.Hour,
	"7d":  7 * 24 * time.Hour,
	"30d": 30 * 24 * time.Hour,
}

// HyperpingAPI defines the Hyperping API methods used by the collector.
type HyperpingAPI interface {
	ListMonitors(ctx context.Context) ([]client.Monitor, error)
	ListHealthchecks(ctx context.Context) ([]client.Healthcheck, error)
	ListOutages(ctx context.Context) ([]client.Outage, error)
	ListMonitorReports(ctx context.Context, from, to string) ([]client.MonitorReport, error)
}

// Collector fetches Hyperping data on a background timer and serves
// cached results as Prometheus metrics. It implements prometheus.Collector.
type Collector struct {
	api      HyperpingAPI
	cacheTTL time.Duration
	logger   *slog.Logger

	// Cache (protected by mu).
	mu              sync.RWMutex
	monitors        []client.Monitor
	healthchecks    []client.Healthcheck
	outages         []client.Outage
	reportsByPeriod map[string][]client.MonitorReport
	lastScrapeTime  time.Time
	lastSuccessTime time.Time // OPS-31: tracks age of fresh data
	lastScrapeOK    bool
	lastScrapeDur   time.Duration

	// Existing core metric descriptors.
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

	// OPS-31: Cache freshness.
	dataAgeDesc *prometheus.Desc

	// OPS-32 + OPS-33: Outage and SLA metrics (with period label where applicable).
	monitorOutageActive       *prometheus.Desc
	monitorActiveOutageStatus *prometheus.Desc
	monitorSLA                *prometheus.Desc
	monitorOutages            *prometheus.Desc
	monitorDowntime           *prometheus.Desc
	monitorLongestOutage      *prometheus.Desc
	monitorMTTR               *prometheus.Desc

	// OPS-34: Tenant-wide health metrics.
	tenantHealthScore   *prometheus.Desc
	tenantUpRatio       *prometheus.Desc
	tenantActiveOutages *prometheus.Desc
	tenantAvgSLA        *prometheus.Desc

	// OPS-39: Escalation tier info metric.
	monitorTier *prometheus.Desc
}

// Verify Collector implements prometheus.Collector at compile time.
var _ prometheus.Collector = (*Collector)(nil)

// NewCollector creates a new Hyperping metrics collector.
func NewCollector(api HyperpingAPI, cacheTTL time.Duration, logger *slog.Logger) *Collector {
	monitorLabels := []string{"uuid", "name"}
	monitorPeriodLabels := []string{"uuid", "name", "period"}

	return &Collector{
		api:             api,
		cacheTTL:        cacheTTL,
		logger:          logger,
		reportsByPeriod: make(map[string][]client.MonitorReport),

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
			[]string{"uuid", "name"}, nil,
		),
		healthcheckPaused: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "healthcheck", "paused"),
			"Whether the healthcheck is paused (1) or active (0).",
			[]string{"uuid", "name"}, nil,
		),
		healthcheckPeriod: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "healthcheck", "period_seconds"),
			"Expected healthcheck ping period in seconds.",
			[]string{"uuid", "name"}, nil,
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
		dataAgeDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "data", "age_seconds"),
			"Seconds elapsed since the last successful API cache refresh.",
			nil, nil,
		),
		monitorOutageActive: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "monitor", "outage_active"),
			"Whether the monitor has an active (unresolved) outage (1) or not (0).",
			monitorLabels, nil,
		),
		monitorActiveOutageStatus: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "monitor", "active_outage_status_code"),
			"HTTP status code of the current active outage; 0 when no active outage.",
			monitorLabels, nil,
		),
		monitorSLA: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "monitor", "sla_ratio"),
			"Monitor SLA as a ratio (0–1) over the labelled period.",
			monitorPeriodLabels, nil,
		),
		monitorOutages: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "monitor", "outages"),
			"Number of outages over the labelled period.",
			monitorPeriodLabels, nil,
		),
		monitorDowntime: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "monitor", "downtime_seconds"),
			"Total downtime in seconds over the labelled period.",
			monitorPeriodLabels, nil,
		),
		monitorLongestOutage: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "monitor", "longest_outage_seconds"),
			"Duration of the longest single outage in seconds over the labelled period.",
			monitorPeriodLabels, nil,
		),
		monitorMTTR: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "monitor", "mttr_seconds"),
			"Mean Time To Recovery in seconds over the labelled period.",
			monitorPeriodLabels, nil,
		),
		tenantHealthScore: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "tenant", "health_score"),
			"Composite tenant health score from 0 to 100.",
			nil, nil,
		),
		tenantUpRatio: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "tenant", "monitors_up_ratio"),
			"Fraction of monitors currently up (0–1).",
			nil, nil,
		),
		tenantActiveOutages: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "tenant", "active_outages"),
			"Total number of active (unresolved) outages across all monitors.",
			nil, nil,
		),
		tenantAvgSLA: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "tenant", "avg_sla_ratio"),
			"Average SLA ratio across all monitors for the labelled period.",
			[]string{"period"}, nil,
		),
		monitorTier: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "monitor", "escalation_tier"),
			"Escalation tier info (always 1). Join on uuid+name; use tier label to filter core/noncore.",
			[]string{"uuid", "name", "tier"}, nil,
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
// Monitors and healthchecks are required; outages and reports are best-effort.
func (c *Collector) Refresh(ctx context.Context) {
	start := time.Now()
	now := start.UTC()

	var (
		monitors     []client.Monitor
		healthchecks []client.Healthcheck
		outages      []client.Outage
		monErr       error
		hcErr        error
		outageErr    error
		wg           sync.WaitGroup
	)

	wg.Add(3)
	go func() {
		defer wg.Done()
		monitors, monErr = c.api.ListMonitors(ctx)
	}()
	go func() {
		defer wg.Done()
		healthchecks, hcErr = c.api.ListHealthchecks(ctx)
	}()
	go func() {
		defer wg.Done()
		outages, outageErr = c.api.ListOutages(ctx)
	}()
	wg.Wait()

	// Fetch SLA reports for all periods in parallel (non-fatal on failure).
	reportResults := make(map[string][]client.MonitorReport, len(reportPeriods))
	var (
		reportMu sync.Mutex
		reportWg sync.WaitGroup
	)
	for _, p := range reportPeriods {
		dur := reportDurations[p]
		from := now.Add(-dur).Format(time.RFC3339)
		to := now.Format(time.RFC3339)
		reportWg.Add(1)
		go func(period, fromStr, toStr string) {
			defer reportWg.Done()
			reports, err := c.api.ListMonitorReports(ctx, fromStr, toStr)
			if err != nil {
				c.logger.Warn("failed to fetch monitor reports", "period", period, "error", err)
				return
			}
			reportMu.Lock()
			reportResults[period] = reports
			reportMu.Unlock()
		}(p, from, to)
	}
	reportWg.Wait()

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

	if outageErr != nil {
		c.logger.Warn("failed to list outages; outage metrics will use stale data", "error", outageErr)
	} else {
		c.outages = outages
	}

	c.monitors = monitors
	c.healthchecks = healthchecks
	// Merge successful period results into the cache; periods that failed this
	// cycle retain their previous data rather than being dropped entirely.
	for period, reports := range reportResults {
		c.reportsByPeriod[period] = reports
	}
	c.lastScrapeOK = true
	c.lastSuccessTime = time.Now()

	c.logger.Info("cache refreshed",
		"monitors", len(monitors),
		"healthchecks", len(healthchecks),
		"outages", len(c.outages),
		"report_periods", len(reportResults),
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
	ch <- c.dataAgeDesc
	ch <- c.monitorOutageActive
	ch <- c.monitorActiveOutageStatus
	ch <- c.monitorSLA
	ch <- c.monitorOutages
	ch <- c.monitorDowntime
	ch <- c.monitorLongestOutage
	ch <- c.monitorMTTR
	ch <- c.tenantHealthScore
	ch <- c.tenantUpRatio
	ch <- c.tenantActiveOutages
	ch <- c.tenantAvgSLA
	ch <- c.monitorTier
}

// Collect implements prometheus.Collector.
func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	activeOutageByMonitor := buildActiveOutageIndex(c.outages)

	upCount := 0
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

		// OPS-32: active outage state and HTTP status code.
		activeOutage, hasActive := activeOutageByMonitor[m.UUID]
		ch <- prometheus.MustNewConstMetric(c.monitorOutageActive, prometheus.GaugeValue,
			boolToFloat64(hasActive), m.UUID, m.Name)
		statusCode := 0
		if hasActive {
			statusCode = activeOutage.StatusCode
		}
		ch <- prometheus.MustNewConstMetric(c.monitorActiveOutageStatus, prometheus.GaugeValue,
			float64(statusCode), m.UUID, m.Name)

		// OPS-39: escalation tier info.
		ch <- prometheus.MustNewConstMetric(c.monitorTier, prometheus.GaugeValue, 1,
			m.UUID, m.Name, escalationTier(m))

		if m.Status == "up" {
			upCount++
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

	// OPS-32 + OPS-33: per-monitor SLA/outage stats per report period.
	for _, period := range reportPeriods {
		reports := c.reportsByPeriod[period]
		slaSum := 0.0
		for _, r := range reports {
			sla := r.SLA / 100.0 // API returns 0–100; expose as 0–1
			ch <- prometheus.MustNewConstMetric(c.monitorSLA, prometheus.GaugeValue,
				sla, r.UUID, r.Name, period)
			ch <- prometheus.MustNewConstMetric(c.monitorOutages, prometheus.GaugeValue,
				float64(r.Outages.Count), r.UUID, r.Name, period)
			ch <- prometheus.MustNewConstMetric(c.monitorDowntime, prometheus.GaugeValue,
				float64(r.Outages.TotalDowntime), r.UUID, r.Name, period)
			ch <- prometheus.MustNewConstMetric(c.monitorLongestOutage, prometheus.GaugeValue,
				float64(r.Outages.LongestOutage), r.UUID, r.Name, period)
			if r.MTTR > 0 {
				ch <- prometheus.MustNewConstMetric(c.monitorMTTR, prometheus.GaugeValue,
					float64(r.MTTR), r.UUID, r.Name, period)
			}
			slaSum += sla
		}
		if len(reports) > 0 {
			ch <- prometheus.MustNewConstMetric(c.tenantAvgSLA, prometheus.GaugeValue,
				slaSum/float64(len(reports)), period)
		}
	}

	// Summary metrics.
	ch <- prometheus.MustNewConstMetric(c.monitorsTotal, prometheus.GaugeValue,
		float64(len(c.monitors)))
	ch <- prometheus.MustNewConstMetric(c.healthchecksTotal, prometheus.GaugeValue,
		float64(len(c.healthchecks)))
	ch <- prometheus.MustNewConstMetric(c.scrapeDurationDesc, prometheus.GaugeValue,
		c.lastScrapeDur.Seconds())
	ch <- prometheus.MustNewConstMetric(c.scrapeSuccessDesc, prometheus.GaugeValue,
		boolToFloat64(c.lastScrapeOK))

	// OPS-31: data age — only after at least one successful scrape.
	if !c.lastSuccessTime.IsZero() {
		ch <- prometheus.MustNewConstMetric(c.dataAgeDesc, prometheus.GaugeValue,
			time.Since(c.lastSuccessTime).Seconds())
	}

	// OPS-34: tenant-wide health metrics.
	upRatio := 0.0
	if len(c.monitors) > 0 {
		upRatio = float64(upCount) / float64(len(c.monitors))
	}
	ch <- prometheus.MustNewConstMetric(c.tenantUpRatio, prometheus.GaugeValue, upRatio)
	ch <- prometheus.MustNewConstMetric(c.tenantActiveOutages, prometheus.GaugeValue,
		float64(len(activeOutageByMonitor)))
	// Health score requires 30d SLA data; omit until reports are loaded to avoid
	// misleadingly low scores (upRatio×60 + 0×40 = 60 even for a healthy fleet).
	if reports30d := c.reportsByPeriod["30d"]; len(reports30d) > 0 {
		avg30dSLA := avgSLAForPeriod(reports30d)
		ch <- prometheus.MustNewConstMetric(c.tenantHealthScore, prometheus.GaugeValue,
			computeHealthScore(upRatio, avg30dSLA, len(activeOutageByMonitor), len(c.monitors)))
	}
}

// buildActiveOutageIndex returns a map of monitor UUID → active Outage.
// An outage is active when IsResolved is false and EndDate is nil (ongoing).
func buildActiveOutageIndex(outages []client.Outage) map[string]client.Outage {
	idx := make(map[string]client.Outage, len(outages))
	for _, o := range outages {
		if !o.IsResolved && o.EndDate == nil {
			idx[o.Monitor.UUID] = o
		}
	}
	return idx
}

// escalationTier returns "core" when the monitor has an escalation policy set,
// "noncore" otherwise. The binary classification is intentional: the Hyperping
// API does not expose a "shared infrastructure" concept at the monitor level;
// tenant grouping is a concern of the Python automation layer, not this exporter.
func escalationTier(m client.Monitor) string {
	if m.EscalationPolicy != nil && *m.EscalationPolicy != "" {
		return "core"
	}
	return "noncore"
}

// avgSLAForPeriod computes the mean SLA ratio (0–1) across a set of reports.
func avgSLAForPeriod(reports []client.MonitorReport) float64 {
	if len(reports) == 0 {
		return 0
	}
	sum := 0.0
	for _, r := range reports {
		sum += r.SLA / 100.0
	}
	return sum / float64(len(reports))
}

// computeHealthScore returns a composite 0–100 health score.
// Base: upRatio×60 + avgSLA×40. Penalty: (activeOutages/total)×30, capped at 100/0.
func computeHealthScore(upRatio, avgSLA float64, activeOutages, totalMonitors int) float64 {
	base := upRatio*60.0 + avgSLA*40.0
	if totalMonitors > 0 {
		penalty := float64(activeOutages) / float64(totalMonitors) * 30.0
		base -= penalty
	}
	if base < 0 {
		return 0
	}
	if base > 100 {
		return 100
	}
	return base
}

func boolToFloat64(b bool) float64 {
	if b {
		return 1
	}
	return 0
}
