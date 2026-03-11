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
	monitorsErr     error
	healthchecksErr error
}

func (m *mockAPI) ListMonitors(_ context.Context) ([]client.Monitor, error) {
	return m.monitors, m.monitorsErr
}

func (m *mockAPI) ListHealthchecks(_ context.Context) ([]client.Healthcheck, error) {
	return m.healthchecks, m.healthchecksErr
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

	ch := make(chan *prometheus.Desc, 20)
	c.Describe(ch)
	close(ch)

	var descs []*prometheus.Desc
	for d := range ch {
		descs = append(descs, d)
	}
	assert.Len(t, descs, 12)
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

func TestRefresh_PreservesOldCacheOnError(t *testing.T) {
	api := &mockAPI{
		monitors:     []client.Monitor{{UUID: "mon_1", Name: "Web", HTTPMethod: "GET"}},
		healthchecks: []client.Healthcheck{},
	}

	c := NewCollector(api, 60*time.Second, newTestLogger())
	c.Refresh(context.Background())
	require.True(t, c.IsReady())

	// Now make the API fail
	api.monitorsErr = errors.New("temporary failure")
	c.Refresh(context.Background())

	// Should not be ready (last scrape failed), but old data remains
	assert.False(t, c.IsReady())

	// Collect should still emit metrics from old cache
	count := testutil.CollectAndCount(c)
	// 1 monitor: up + paused + check_interval + info = 4, no SSL
	// 4 summary metrics
	assert.Equal(t, 8, count)
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

	// mon_1: up + paused + check_interval + info + ssl = 5
	// mon_2: up + paused + check_interval + info = 4 (no SSL)
	// 1 healthcheck: up + paused + period = 3
	// Summary: 4 (monitors, healthchecks, scrape_duration, scrape_success)
	// Total = 16
	count := testutil.CollectAndCount(c)
	assert.Equal(t, 16, count)
}

func TestCollect_EmptyCache(t *testing.T) {
	c := NewCollector(&mockAPI{}, 60*time.Second, newTestLogger())

	// Before any refresh: only 4 summary/exporter metrics
	count := testutil.CollectAndCount(c)
	assert.Equal(t, 4, count)
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

	// 1 monitor: up + paused + check_interval + info = 4 (no SSL)
	// Summary: 4
	// Total = 8
	count := testutil.CollectAndCount(c)
	assert.Equal(t, 8, count)
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

func TestCollect_Lint(t *testing.T) {
	api := &mockAPI{
		monitors: []client.Monitor{
			{UUID: "mon_1", Name: "Web", Protocol: "http", HTTPMethod: "GET", CheckFrequency: 60, Status: "up"},
		},
		healthchecks: []client.Healthcheck{
			{UUID: "tok_1", Name: "Job", Period: 300},
		},
	}

	c := NewCollector(api, 60*time.Second, newTestLogger())
	c.Refresh(context.Background())

	problems, err := testutil.CollectAndLint(c)
	require.NoError(t, err)
	assert.Empty(t, problems)
}
