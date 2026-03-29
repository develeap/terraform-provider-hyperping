// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/develeap/terraform-provider-hyperping/internal/client"
)

// mockAPI implements HyperpingAPI for testing.
type mockAPI struct {
	monitors        []client.Monitor
	healthchecks    []client.Healthcheck
	outages         []client.Outage
	reports         []client.MonitorReport
	monitorsErr     error
	healthchecksErr error
	outagesErr      error
	reportsErr      error
}

func (m *mockAPI) ListMonitors(_ context.Context) ([]client.Monitor, error) {
	return m.monitors, m.monitorsErr
}

func (m *mockAPI) ListHealthchecks(_ context.Context) ([]client.Healthcheck, error) {
	return m.healthchecks, m.healthchecksErr
}

func (m *mockAPI) ListOutages(_ context.Context) ([]client.Outage, error) {
	return m.outages, m.outagesErr
}

func (m *mockAPI) ListMonitorReports(_ context.Context, _, _ string) ([]client.MonitorReport, error) {
	return m.reports, m.reportsErr
}

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestNewCollector(t *testing.T) {
	c := NewCollector(&mockAPI{}, 60*time.Second, newTestLogger())

	assert.NotNil(t, c)
	assert.False(t, c.IsReady())
}

func TestDescribe(t *testing.T) {
	c := NewCollector(&mockAPI{}, 60*time.Second, newTestLogger())

	ch := make(chan *prometheus.Desc, 30)
	c.Describe(ch)
	close(ch)

	var descs []*prometheus.Desc
	for d := range ch {
		descs = append(descs, d)
	}
	// 12 original + 13 new (OPS-31/32/33/34/39) = 25
	assert.Len(t, descs, 25)
}

func TestRefresh_Success(t *testing.T) {
	sslDays := 90
	api := &mockAPI{
		monitors: []client.Monitor{
			{
				UUID:           "mon_123",
				Name:           "API Monitor",
				URL:            "https://api.example.com",
				Protocol:       "http",
				Status:         "up",
				CheckFrequency: 60,
				SSLExpiration:  &sslDays,
				ProjectUUID:    "proj_abc",
				HTTPMethod:     "GET",
			},
		},
		healthchecks: []client.Healthcheck{
			{UUID: "tok_456", Name: "Cron Job", Period: 300},
		},
	}

	c := NewCollector(api, 60*time.Second, newTestLogger())
	c.Refresh(context.Background())

	assert.True(t, c.IsReady())
}

func TestRefresh_MonitorError(t *testing.T) {
	api := &mockAPI{
		monitorsErr:  errors.New("api error"),
		healthchecks: []client.Healthcheck{},
	}

	c := NewCollector(api, 60*time.Second, newTestLogger())
	c.Refresh(context.Background())

	assert.False(t, c.IsReady())
}

func TestRefresh_HealthcheckError(t *testing.T) {
	api := &mockAPI{
		monitors:        []client.Monitor{},
		healthchecksErr: errors.New("api error"),
	}

	c := NewCollector(api, 60*time.Second, newTestLogger())
	c.Refresh(context.Background())

	assert.False(t, c.IsReady())
}

func TestRefresh_OutageErrorIsNonFatal(t *testing.T) {
	// Outage failures should not mark the scrape as failed.
	api := &mockAPI{
		monitors:     []client.Monitor{{UUID: "mon_1", Name: "Web", HTTPMethod: "GET", Status: "up"}},
		healthchecks: []client.Healthcheck{},
		outagesErr:   errors.New("outage api error"),
	}

	c := NewCollector(api, 60*time.Second, newTestLogger())
	c.Refresh(context.Background())

	assert.True(t, c.IsReady())
}

func TestRefresh_PreservesOldCacheOnError(t *testing.T) {
	api := &mockAPI{
		monitors:     []client.Monitor{{UUID: "mon_1", Name: "Web", HTTPMethod: "GET"}},
		healthchecks: []client.Healthcheck{},
	}

	c := NewCollector(api, 60*time.Second, newTestLogger())
	c.Refresh(context.Background())
	require.True(t, c.IsReady())

	api.monitorsErr = errors.New("temporary failure")
	c.Refresh(context.Background())

	assert.False(t, c.IsReady())

	// Old monitor data remains; lastSuccessTime was set on first scrape so data_age IS emitted.
	// Per monitor: up + paused + check_interval + info + outage_active + status_code + tier = 7
	// Summary: 4, Tenant: up_ratio + active_outages + data_age = 3 (health_score omitted: no reports)
	// Total: 7 + 4 + 3 = 14
	count := testutil.CollectAndCount(c)
	assert.Equal(t, 14, count)
}

func TestCollect_WithMonitorsAndHealthchecks(t *testing.T) {
	sslDays := 30
	api := &mockAPI{
		monitors: []client.Monitor{
			{
				UUID:           "mon_1",
				Name:           "Web",
				URL:            "https://example.com",
				Protocol:       "http",
				Status:         "up",
				CheckFrequency: 60,
				SSLExpiration:  &sslDays,
				ProjectUUID:    "proj_1",
				HTTPMethod:     "GET",
			},
			{
				UUID:           "mon_2",
				Name:           "API",
				URL:            "https://api.example.com",
				Protocol:       "http",
				Status:         "down",
				Paused:         true,
				CheckFrequency: 30,
				ProjectUUID:    "proj_1",
				HTTPMethod:     "POST",
			},
		},
		healthchecks: []client.Healthcheck{
			{UUID: "tok_1", Name: "Backup", Period: 3600},
		},
	}

	c := NewCollector(api, 60*time.Second, newTestLogger())
	c.Refresh(context.Background())

	// mon_1: up + paused + interval + info + ssl + outage_active + status_code + tier = 8
	// mon_2: up + paused + interval + info + outage_active + status_code + tier = 7
	// hc: up + paused + period = 3
	// Summary: monitors + healthchecks + scrape_duration + scrape_success = 4
	// Tenant: up_ratio + active_outages + data_age = 3 (health_score omitted: no reports in mock)
	// Total = 25
	count := testutil.CollectAndCount(c)
	assert.Equal(t, 25, count)
}

func TestCollect_EmptyCache(t *testing.T) {
	c := NewCollector(&mockAPI{}, 60*time.Second, newTestLogger())

	// No refresh: 4 summary + 2 tenant (up_ratio + active_outages; no data_age, no health_score) = 6
	count := testutil.CollectAndCount(c)
	assert.Equal(t, 6, count)
}

func TestCollect_NoSSLExpiration(t *testing.T) {
	api := &mockAPI{
		monitors: []client.Monitor{
			{
				UUID:           "mon_1",
				Name:           "TCP Monitor",
				Protocol:       "tcp",
				Status:         "up",
				CheckFrequency: 60,
				HTTPMethod:     "GET",
			},
		},
		healthchecks: []client.Healthcheck{},
	}

	c := NewCollector(api, 60*time.Second, newTestLogger())
	c.Refresh(context.Background())

	// Monitor: up + paused + interval + info + outage_active + status_code + tier = 7
	// Summary: 4, Tenant: up_ratio + active_outages + data_age = 3 (health_score omitted: no reports)
	// Total = 14
	count := testutil.CollectAndCount(c)
	assert.Equal(t, 14, count)
}

func TestCollect_SummaryMetricValues(t *testing.T) {
	api := &mockAPI{
		monitors: []client.Monitor{
			{UUID: "mon_1", Name: "Web", Protocol: "http", HTTPMethod: "GET", CheckFrequency: 60, Status: "up"},
			{UUID: "mon_2", Name: "API", Protocol: "http", HTTPMethod: "GET", CheckFrequency: 30, Status: "down"},
		},
		healthchecks: []client.Healthcheck{
			{UUID: "tok_1", Name: "Job", Period: 300},
		},
	}

	c := NewCollector(api, 60*time.Second, newTestLogger())
	c.Refresh(context.Background())

	expected := `
# HELP hyperping_monitors Total number of monitors.
# TYPE hyperping_monitors gauge
hyperping_monitors 2
# HELP hyperping_healthchecks Total number of healthchecks.
# TYPE hyperping_healthchecks gauge
hyperping_healthchecks 1
# HELP hyperping_scrape_success Whether the last API scrape succeeded (1) or failed (0).
# TYPE hyperping_scrape_success gauge
hyperping_scrape_success 1
`
	err := testutil.CollectAndCompare(c, strings.NewReader(expected),
		"hyperping_monitors",
		"hyperping_healthchecks",
		"hyperping_scrape_success",
	)
	require.NoError(t, err)
}

func TestCollect_MonitorMetricValues(t *testing.T) {
	sslDays := 45
	api := &mockAPI{
		monitors: []client.Monitor{
			{
				UUID:           "mon_1",
				Name:           "Web",
				URL:            "https://example.com",
				Protocol:       "http",
				Status:         "up",
				CheckFrequency: 120,
				SSLExpiration:  &sslDays,
				ProjectUUID:    "proj_1",
				HTTPMethod:     "GET",
			},
		},
		healthchecks: []client.Healthcheck{},
	}

	c := NewCollector(api, 60*time.Second, newTestLogger())
	c.Refresh(context.Background())

	expected := `
# HELP hyperping_monitor_up Whether the monitor is up (1) or down (0).
# TYPE hyperping_monitor_up gauge
hyperping_monitor_up{name="Web",uuid="mon_1"} 1
# HELP hyperping_monitor_paused Whether the monitor is paused (1) or active (0).
# TYPE hyperping_monitor_paused gauge
hyperping_monitor_paused{name="Web",uuid="mon_1"} 0
# HELP hyperping_monitor_check_interval_seconds Monitor check frequency in seconds.
# TYPE hyperping_monitor_check_interval_seconds gauge
hyperping_monitor_check_interval_seconds{name="Web",uuid="mon_1"} 120
# HELP hyperping_monitor_ssl_expiration_days Days until SSL certificate expiration.
# TYPE hyperping_monitor_ssl_expiration_days gauge
hyperping_monitor_ssl_expiration_days{name="Web",uuid="mon_1"} 45
`
	err := testutil.CollectAndCompare(c, strings.NewReader(expected),
		"hyperping_monitor_up",
		"hyperping_monitor_paused",
		"hyperping_monitor_check_interval_seconds",
		"hyperping_monitor_ssl_expiration_days",
	)
	require.NoError(t, err)
}

func TestCollect_HealthcheckMetricValues(t *testing.T) {
	api := &mockAPI{
		monitors: []client.Monitor{},
		healthchecks: []client.Healthcheck{
			{UUID: "tok_1", Name: "Daily Backup", IsDown: true, IsPaused: false, Period: 86400},
			{UUID: "tok_2", Name: "Hourly Sync", IsDown: false, IsPaused: true, Period: 3600},
		},
	}

	c := NewCollector(api, 60*time.Second, newTestLogger())
	c.Refresh(context.Background())

	expected := `
# HELP hyperping_healthcheck_up Whether the healthcheck is up (1) or down (0).
# TYPE hyperping_healthcheck_up gauge
hyperping_healthcheck_up{name="Daily Backup",uuid="tok_1"} 0
hyperping_healthcheck_up{name="Hourly Sync",uuid="tok_2"} 1
# HELP hyperping_healthcheck_paused Whether the healthcheck is paused (1) or active (0).
# TYPE hyperping_healthcheck_paused gauge
hyperping_healthcheck_paused{name="Daily Backup",uuid="tok_1"} 0
hyperping_healthcheck_paused{name="Hourly Sync",uuid="tok_2"} 1
# HELP hyperping_healthcheck_period_seconds Expected healthcheck ping period in seconds.
# TYPE hyperping_healthcheck_period_seconds gauge
hyperping_healthcheck_period_seconds{name="Daily Backup",uuid="tok_1"} 86400
hyperping_healthcheck_period_seconds{name="Hourly Sync",uuid="tok_2"} 3600
`
	err := testutil.CollectAndCompare(c, strings.NewReader(expected),
		"hyperping_healthcheck_up",
		"hyperping_healthcheck_paused",
		"hyperping_healthcheck_period_seconds",
	)
	require.NoError(t, err)
}

func TestCollect_ScrapeFailureMetric(t *testing.T) {
	api := &mockAPI{
		monitorsErr: errors.New("network error"),
	}

	c := NewCollector(api, 60*time.Second, newTestLogger())
	c.Refresh(context.Background())

	expected := `
# HELP hyperping_scrape_success Whether the last API scrape succeeded (1) or failed (0).
# TYPE hyperping_scrape_success gauge
hyperping_scrape_success 0
`
	err := testutil.CollectAndCompare(c, strings.NewReader(expected),
		"hyperping_scrape_success",
	)
	require.NoError(t, err)
}

func TestCollect_ActiveOutageMetrics(t *testing.T) {
	endDate := "2026-03-29T12:00:00Z"
	api := &mockAPI{
		monitors: []client.Monitor{
			{UUID: "mon_1", Name: "Web", Protocol: "http", HTTPMethod: "GET", Status: "down"},
			{UUID: "mon_2", Name: "API", Protocol: "http", HTTPMethod: "GET", Status: "up"},
		},
		healthchecks: []client.Healthcheck{},
		outages: []client.Outage{
			// Active outage on mon_1 (EndDate nil, IsResolved false)
			{
				UUID:       "out_1",
				IsResolved: false,
				EndDate:    nil,
				StatusCode: 503,
				Monitor:    client.MonitorReference{UUID: "mon_1", Name: "Web"},
				OutageType: "automatic",
				StartDate:  "2026-03-29T10:00:00Z",
			},
			// Resolved outage on mon_2 (should not be flagged as active)
			{
				UUID:       "out_2",
				IsResolved: true,
				EndDate:    &endDate,
				StatusCode: 500,
				Monitor:    client.MonitorReference{UUID: "mon_2", Name: "API"},
				OutageType: "automatic",
				StartDate:  "2026-03-29T09:00:00Z",
			},
		},
	}

	c := NewCollector(api, 60*time.Second, newTestLogger())
	c.Refresh(context.Background())

	expected := `
# HELP hyperping_monitor_outage_active Whether the monitor has an active (unresolved) outage (1) or not (0).
# TYPE hyperping_monitor_outage_active gauge
hyperping_monitor_outage_active{name="API",uuid="mon_2"} 0
hyperping_monitor_outage_active{name="Web",uuid="mon_1"} 1
# HELP hyperping_monitor_active_outage_status_code HTTP status code of the current active outage; 0 when no active outage.
# TYPE hyperping_monitor_active_outage_status_code gauge
hyperping_monitor_active_outage_status_code{name="API",uuid="mon_2"} 0
hyperping_monitor_active_outage_status_code{name="Web",uuid="mon_1"} 503
`
	err := testutil.CollectAndCompare(c, strings.NewReader(expected),
		"hyperping_monitor_outage_active",
		"hyperping_monitor_active_outage_status_code",
	)
	require.NoError(t, err)
}

func TestCollect_EscalationTierMetrics(t *testing.T) {
	policyUUID := "policy_abc"
	api := &mockAPI{
		monitors: []client.Monitor{
			{UUID: "mon_1", Name: "Core", Protocol: "http", HTTPMethod: "GET", EscalationPolicy: &policyUUID},
			{UUID: "mon_2", Name: "Edge", Protocol: "http", HTTPMethod: "GET", EscalationPolicy: nil},
		},
		healthchecks: []client.Healthcheck{},
	}

	c := NewCollector(api, 60*time.Second, newTestLogger())
	c.Refresh(context.Background())

	expected := `
# HELP hyperping_monitor_escalation_tier Escalation tier info (always 1). Join on uuid+name; use tier label to filter core/noncore.
# TYPE hyperping_monitor_escalation_tier gauge
hyperping_monitor_escalation_tier{name="Core",tier="core",uuid="mon_1"} 1
hyperping_monitor_escalation_tier{name="Edge",tier="noncore",uuid="mon_2"} 1
`
	err := testutil.CollectAndCompare(c, strings.NewReader(expected),
		"hyperping_monitor_escalation_tier",
	)
	require.NoError(t, err)
}

func TestCollect_SLAReportMetrics(t *testing.T) {
	api := &mockAPI{
		monitors: []client.Monitor{
			{UUID: "mon_1", Name: "Web", Protocol: "http", HTTPMethod: "GET", Status: "up"},
		},
		healthchecks: []client.Healthcheck{},
		reports: []client.MonitorReport{
			{
				UUID:     "mon_1",
				Name:     "Web",
				Protocol: "http",
				SLA:      99.5,
				MTTR:     120,
				Outages: client.OutageStats{
					Count:         2,
					TotalDowntime: 300,
					LongestOutage: 240,
				},
			},
		},
	}

	c := NewCollector(api, 60*time.Second, newTestLogger())
	c.Refresh(context.Background())

	// Reports are fetched for 3 periods; each period returns the same mock data.
	expected := `
# HELP hyperping_monitor_sla_ratio Monitor SLA as a ratio (0–1) over the labelled period.
# TYPE hyperping_monitor_sla_ratio gauge
hyperping_monitor_sla_ratio{name="Web",period="24h",uuid="mon_1"} 0.995
hyperping_monitor_sla_ratio{name="Web",period="7d",uuid="mon_1"} 0.995
hyperping_monitor_sla_ratio{name="Web",period="30d",uuid="mon_1"} 0.995
`
	err := testutil.CollectAndCompare(c, strings.NewReader(expected),
		"hyperping_monitor_sla_ratio",
	)
	require.NoError(t, err)
}

func TestCollect_TenantHealthMetrics(t *testing.T) {
	api := &mockAPI{
		monitors: []client.Monitor{
			{UUID: "mon_1", Name: "A", Protocol: "http", HTTPMethod: "GET", Status: "up"},
			{UUID: "mon_2", Name: "B", Protocol: "http", HTTPMethod: "GET", Status: "up"},
		},
		healthchecks: []client.Healthcheck{},
		reports: []client.MonitorReport{
			{UUID: "mon_1", Name: "A", SLA: 100.0},
			{UUID: "mon_2", Name: "B", SLA: 98.0},
		},
	}

	c := NewCollector(api, 60*time.Second, newTestLogger())
	c.Refresh(context.Background())

	// 2 monitors up, 0 active outages → up_ratio=1.0
	// avgSLA = (1.0+0.98)/2 = 0.99 → health_score = 1.0*60 + 0.99*40 = 99.6
	expected := `
# HELP hyperping_tenant_monitors_up_ratio Fraction of monitors currently up (0–1).
# TYPE hyperping_tenant_monitors_up_ratio gauge
hyperping_tenant_monitors_up_ratio 1
# HELP hyperping_tenant_active_outages Total number of active (unresolved) outages across all monitors.
# TYPE hyperping_tenant_active_outages gauge
hyperping_tenant_active_outages 0
`
	err := testutil.CollectAndCompare(c, strings.NewReader(expected),
		"hyperping_tenant_monitors_up_ratio",
		"hyperping_tenant_active_outages",
	)
	require.NoError(t, err)
}

func TestCollect_Lint(t *testing.T) {
	policyUUID := "policy_123"
	endDate := "2026-03-29T08:00:00Z"
	api := &mockAPI{
		monitors: []client.Monitor{
			{
				UUID: "mon_1", Name: "Web", Protocol: "http",
				HTTPMethod: "GET", CheckFrequency: 60, Status: "up",
				EscalationPolicy: &policyUUID,
			},
		},
		healthchecks: []client.Healthcheck{
			{UUID: "tok_1", Name: "Job", Period: 300},
		},
		outages: []client.Outage{
			{
				UUID: "out_1", IsResolved: true, EndDate: &endDate, StatusCode: 200,
				Monitor: client.MonitorReference{UUID: "mon_1", Name: "Web"},
			},
		},
		reports: []client.MonitorReport{
			{UUID: "mon_1", Name: "Web", SLA: 99.9},
		},
	}

	c := NewCollector(api, 60*time.Second, newTestLogger())
	c.Refresh(context.Background())

	problems, err := testutil.CollectAndLint(c)
	require.NoError(t, err)
	assert.Empty(t, problems)
}

func TestComputeHealthScore(t *testing.T) {
	tests := []struct {
		name          string
		upRatio       float64
		avgSLA        float64
		activeOutages int
		totalMonitors int
		expectedMin   float64
		expectedMax   float64
	}{
		{"all healthy", 1.0, 1.0, 0, 10, 99, 101},
		{"all down", 0.0, 0.0, 10, 10, 0, 1},
		{"partial", 0.8, 0.9, 1, 10, 50, 90},
		{"no monitors", 0.0, 0.0, 0, 0, 0, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := computeHealthScore(tt.upRatio, tt.avgSLA, tt.activeOutages, tt.totalMonitors)
			assert.GreaterOrEqual(t, score, tt.expectedMin)
			assert.LessOrEqual(t, score, tt.expectedMax)
		})
	}
}

func TestBuildActiveOutageIndex(t *testing.T) {
	endDate := "2026-03-29T12:00:00Z"
	outages := []client.Outage{
		{UUID: "a", IsResolved: false, EndDate: nil, Monitor: client.MonitorReference{UUID: "mon_1"}},
		{UUID: "b", IsResolved: true, EndDate: &endDate, Monitor: client.MonitorReference{UUID: "mon_2"}},
		{UUID: "c", IsResolved: false, EndDate: &endDate, Monitor: client.MonitorReference{UUID: "mon_3"}},
	}

	idx := buildActiveOutageIndex(outages)

	assert.Len(t, idx, 1)
	assert.Contains(t, idx, "mon_1")
	assert.NotContains(t, idx, "mon_2")
	assert.NotContains(t, idx, "mon_3")
}

func TestEscalationTier(t *testing.T) {
	policyUUID := "uuid-123"
	emptyUUID := ""

	assert.Equal(t, "core", escalationTier(client.Monitor{EscalationPolicy: &policyUUID}))
	assert.Equal(t, "noncore", escalationTier(client.Monitor{EscalationPolicy: nil}))
	assert.Equal(t, "noncore", escalationTier(client.Monitor{EscalationPolicy: &emptyUUID}))
}
