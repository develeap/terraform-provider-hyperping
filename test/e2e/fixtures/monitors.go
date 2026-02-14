// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

//go:build e2e

package fixtures

import (
	"fmt"
)

// MonitorFixture represents a test monitor configuration
type MonitorFixture struct {
	Name       string
	URL        string
	Method     string
	Frequency  int
	ExpectCode int
}

// GetSmallScenarioMonitors returns 1-3 monitors for small migration tests
func GetSmallScenarioMonitors(namePrefix string) []MonitorFixture {
	return []MonitorFixture{
		{
			Name:       fmt.Sprintf("%s-API-Health", namePrefix),
			URL:        "https://httpstat.us/200",
			Method:     "GET",
			Frequency:  60,
			ExpectCode: 200,
		},
		{
			Name:       fmt.Sprintf("%s-Homepage", namePrefix),
			URL:        "https://httpstat.us/200?sleep=100",
			Method:     "GET",
			Frequency:  120,
			ExpectCode: 200,
		},
	}
}

// GetMediumScenarioMonitors returns 10+ monitors with various configurations
func GetMediumScenarioMonitors(namePrefix string) []MonitorFixture {
	monitors := []MonitorFixture{
		{
			Name:       fmt.Sprintf("%s-API-v1-Health", namePrefix),
			URL:        "https://httpstat.us/200",
			Method:     "GET",
			Frequency:  60,
			ExpectCode: 200,
		},
		{
			Name:       fmt.Sprintf("%s-API-v2-Health", namePrefix),
			URL:        "https://httpstat.us/200?sleep=50",
			Method:     "GET",
			Frequency:  60,
			ExpectCode: 200,
		},
		{
			Name:       fmt.Sprintf("%s-Auth-Service", namePrefix),
			URL:        "https://httpstat.us/200",
			Method:     "GET",
			Frequency:  120,
			ExpectCode: 200,
		},
		{
			Name:       fmt.Sprintf("%s-Payment-Gateway", namePrefix),
			URL:        "https://httpstat.us/200?sleep=100",
			Method:     "GET",
			Frequency:  30,
			ExpectCode: 200,
		},
		{
			Name:       fmt.Sprintf("%s-Database-Health", namePrefix),
			URL:        "https://httpstat.us/200",
			Method:     "GET",
			Frequency:  60,
			ExpectCode: 200,
		},
		{
			Name:       fmt.Sprintf("%s-Cache-Service", namePrefix),
			URL:        "https://httpstat.us/200?sleep=25",
			Method:     "GET",
			Frequency:  120,
			ExpectCode: 200,
		},
		{
			Name:       fmt.Sprintf("%s-Queue-Worker", namePrefix),
			URL:        "https://httpstat.us/200",
			Method:     "GET",
			Frequency:  300,
			ExpectCode: 200,
		},
		{
			Name:       fmt.Sprintf("%s-Notification-Service", namePrefix),
			URL:        "https://httpstat.us/200?sleep=75",
			Method:     "GET",
			Frequency:  180,
			ExpectCode: 200,
		},
		{
			Name:       fmt.Sprintf("%s-Search-API", namePrefix),
			URL:        "https://httpstat.us/200",
			Method:     "GET",
			Frequency:  60,
			ExpectCode: 200,
		},
		{
			Name:       fmt.Sprintf("%s-Analytics-Endpoint", namePrefix),
			URL:        "https://httpstat.us/200?sleep=150",
			Method:     "GET",
			Frequency:  300,
			ExpectCode: 200,
		},
		{
			Name:       fmt.Sprintf("%s-CDN-Status", namePrefix),
			URL:        "https://httpstat.us/200",
			Method:     "HEAD",
			Frequency:  120,
			ExpectCode: 200,
		},
		{
			Name:       fmt.Sprintf("%s-Admin-Dashboard", namePrefix),
			URL:        "https://httpstat.us/200?sleep=50",
			Method:     "GET",
			Frequency:  180,
			ExpectCode: 200,
		},
	}

	return monitors
}

// GetLargeScenarioMonitors returns 50+ monitors for stress testing
func GetLargeScenarioMonitors(namePrefix string) []MonitorFixture {
	monitors := GetMediumScenarioMonitors(namePrefix)

	// Add more monitors to reach 50+
	services := []string{
		"Frontend", "Backend", "GraphQL", "REST-API", "WebSocket",
		"Mobile-API", "Partner-API", "Internal-API", "Webhook-Handler",
		"Email-Service", "SMS-Gateway", "Push-Notifications",
		"Image-Processor", "Video-Transcoder", "File-Storage",
		"Backup-Service", "Monitoring-Agent", "Log-Aggregator",
		"Metrics-Collector", "Alert-Manager", "Incident-Manager",
		"Config-Server", "Feature-Flags", "AB-Testing",
		"Recommendation-Engine", "ML-Model-Server", "Data-Pipeline",
		"ETL-Service", "Data-Warehouse", "Business-Intelligence",
		"Customer-Portal", "Vendor-Portal", "Support-Portal",
		"Documentation-Site", "Blog", "Marketing-Site",
		"E-commerce-Cart", "Checkout-Service", "Inventory-Service",
		"Shipping-Tracker", "Returns-Portal", "Reviews-Service",
	}

	frequencies := []int{30, 60, 120, 180, 300}
	methods := []string{"GET", "HEAD"}
	sleeps := []string{"", "?sleep=25", "?sleep=50", "?sleep=100"}

	for i, service := range services {
		monitor := MonitorFixture{
			Name:       fmt.Sprintf("%s-%s", namePrefix, service),
			URL:        fmt.Sprintf("https://httpstat.us/200%s", sleeps[i%len(sleeps)]),
			Method:     methods[i%len(methods)],
			Frequency:  frequencies[i%len(frequencies)],
			ExpectCode: 200,
		}
		monitors = append(monitors, monitor)
	}

	return monitors
}

// HeartbeatFixture represents a test heartbeat/healthcheck configuration
type HeartbeatFixture struct {
	Name   string
	Period int
	Grace  int
}

// GetSmallScenarioHeartbeats returns 1-2 heartbeats for small migration tests
func GetSmallScenarioHeartbeats(namePrefix string) []HeartbeatFixture {
	return []HeartbeatFixture{
		{
			Name:   fmt.Sprintf("%s-Cron-Job", namePrefix),
			Period: 3600,
			Grace:  300,
		},
	}
}

// GetMediumScenarioHeartbeats returns 5+ heartbeats
func GetMediumScenarioHeartbeats(namePrefix string) []HeartbeatFixture {
	return []HeartbeatFixture{
		{
			Name:   fmt.Sprintf("%s-Hourly-Job", namePrefix),
			Period: 3600,
			Grace:  300,
		},
		{
			Name:   fmt.Sprintf("%s-Daily-Backup", namePrefix),
			Period: 86400,
			Grace:  3600,
		},
		{
			Name:   fmt.Sprintf("%s-Every-5min", namePrefix),
			Period: 300,
			Grace:  60,
		},
		{
			Name:   fmt.Sprintf("%s-Data-Sync", namePrefix),
			Period: 1800,
			Grace:  180,
		},
		{
			Name:   fmt.Sprintf("%s-Report-Generator", namePrefix),
			Period: 43200,
			Grace:  1800,
		},
	}
}
