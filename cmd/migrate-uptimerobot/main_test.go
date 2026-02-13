// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"testing"

	"github.com/develeap/terraform-provider-hyperping/cmd/migrate-uptimerobot/converter"
	"github.com/develeap/terraform-provider-hyperping/cmd/migrate-uptimerobot/uptimerobot"
)

func TestConverterHTTPMonitor(t *testing.T) {
	conv := converter.NewConverter()

	httpMethod := 1 // GET
	monitors := []uptimerobot.Monitor{
		{
			ID:           12345,
			FriendlyName: "Test API",
			URL:          "https://api.example.com/health",
			Type:         1, // HTTP
			HTTPMethod:   &httpMethod,
			Interval:     60,
		},
	}

	result := conv.Convert(monitors, nil)

	if len(result.Monitors) != 1 {
		t.Fatalf("Expected 1 monitor, got %d", len(result.Monitors))
	}

	m := result.Monitors[0]

	if m.Protocol != "http" {
		t.Errorf("Expected protocol 'http', got '%s'", m.Protocol)
	}

	if m.HTTPMethod != "GET" {
		t.Errorf("Expected HTTP method 'GET', got '%s'", m.HTTPMethod)
	}

	if m.CheckFrequency != 60 {
		t.Errorf("Expected check frequency 60, got %d", m.CheckFrequency)
	}

	if m.URL != "https://api.example.com/health" {
		t.Errorf("Expected URL 'https://api.example.com/health', got '%s'", m.URL)
	}
}

func TestConverterKeywordMonitor(t *testing.T) {
	conv := converter.NewConverter()

	keywordType := 1 // exists
	keywordValue := "healthy"
	monitors := []uptimerobot.Monitor{
		{
			ID:           12346,
			FriendlyName: "API Status Check",
			URL:          "https://api.example.com/status",
			Type:         2, // Keyword
			KeywordType:  &keywordType,
			KeywordValue: &keywordValue,
			Interval:     300,
		},
	}

	result := conv.Convert(monitors, nil)

	if len(result.Monitors) != 1 {
		t.Fatalf("Expected 1 monitor, got %d", len(result.Monitors))
	}

	m := result.Monitors[0]

	if m.Protocol != "http" {
		t.Errorf("Expected protocol 'http', got '%s'", m.Protocol)
	}

	if m.RequiredKeyword != "healthy" {
		t.Errorf("Expected required keyword 'healthy', got '%s'", m.RequiredKeyword)
	}
}

func TestConverterKeywordNotExists(t *testing.T) {
	conv := converter.NewConverter()

	keywordType := 2 // not exists
	keywordValue := "error"
	monitors := []uptimerobot.Monitor{
		{
			ID:           12347,
			FriendlyName: "Error Check",
			URL:          "https://api.example.com",
			Type:         2, // Keyword
			KeywordType:  &keywordType,
			KeywordValue: &keywordValue,
			Interval:     60,
		},
	}

	result := conv.Convert(monitors, nil)

	if len(result.Monitors) != 1 {
		t.Fatalf("Expected 1 monitor, got %d", len(result.Monitors))
	}

	m := result.Monitors[0]

	if len(m.Warnings) == 0 {
		t.Error("Expected warning for 'not exists' keyword check")
	}
}

func TestConverterPingMonitor(t *testing.T) {
	conv := converter.NewConverter()

	monitors := []uptimerobot.Monitor{
		{
			ID:           12348,
			FriendlyName: "Server Ping",
			URL:          "192.168.1.100",
			Type:         3, // Ping
			Interval:     60,
		},
	}

	result := conv.Convert(monitors, nil)

	if len(result.Monitors) != 1 {
		t.Fatalf("Expected 1 monitor, got %d", len(result.Monitors))
	}

	m := result.Monitors[0]

	if m.Protocol != "icmp" {
		t.Errorf("Expected protocol 'icmp', got '%s'", m.Protocol)
	}

	if m.URL != "192.168.1.100" {
		t.Errorf("Expected URL '192.168.1.100', got '%s'", m.URL)
	}
}

func TestConverterPortMonitor(t *testing.T) {
	conv := converter.NewConverter()

	port := 5432
	monitors := []uptimerobot.Monitor{
		{
			ID:           12349,
			FriendlyName: "PostgreSQL Port",
			URL:          "db.example.com",
			Type:         4, // Port
			Port:         &port,
			Interval:     120,
		},
	}

	result := conv.Convert(monitors, nil)

	if len(result.Monitors) != 1 {
		t.Fatalf("Expected 1 monitor, got %d", len(result.Monitors))
	}

	m := result.Monitors[0]

	if m.Protocol != "port" {
		t.Errorf("Expected protocol 'port', got '%s'", m.Protocol)
	}

	if m.Port != 5432 {
		t.Errorf("Expected port 5432, got %d", m.Port)
	}
}

func TestConverterHeartbeatMonitor(t *testing.T) {
	conv := converter.NewConverter()

	monitors := []uptimerobot.Monitor{
		{
			ID:           12350,
			FriendlyName: "Daily Backup",
			Type:         5, // Heartbeat
			Interval:     86400,
		},
	}

	result := conv.Convert(monitors, nil)

	if len(result.Healthchecks) != 1 {
		t.Fatalf("Expected 1 healthcheck, got %d", len(result.Healthchecks))
	}

	h := result.Healthchecks[0]

	if h.PeriodType != "days" {
		t.Errorf("Expected period type 'days', got '%s'", h.PeriodType)
	}

	if h.PeriodValue != 1 {
		t.Errorf("Expected period value 1, got %d", h.PeriodValue)
	}

	if len(h.Warnings) == 0 {
		t.Error("Expected warning for heartbeat conversion")
	}
}

func TestConverterFrequencyMapping(t *testing.T) {
	tests := []struct {
		input    int
		expected int
	}{
		{45, 30},    // Closest is 30 (diff=15) vs 60 (diff=15), picks first
		{90, 60},    // Closest is 60 (diff=30) vs 120 (diff=30), picks first
		{150, 120},  // Closest is 120 (diff=30) vs 180 (diff=30), picks first
		{350, 300},  // Closest is 300 (diff=50) vs 600 (diff=250)
		{500, 600},  // Closest is 600 (diff=100) vs 300 (diff=200)
		{1000, 600}, // Closest is 600 (diff=400) vs 1800 (diff=800)
	}

	conv := converter.NewConverter()

	for _, tt := range tests {
		monitors := []uptimerobot.Monitor{
			{
				ID:           1,
				FriendlyName: "Test",
				URL:          "https://example.com",
				Type:         1,
				Interval:     tt.input,
			},
		}

		result := conv.Convert(monitors, nil)

		if len(result.Monitors) != 1 {
			t.Fatalf("Expected 1 monitor, got %d", len(result.Monitors))
		}

		m := result.Monitors[0]

		if m.CheckFrequency != tt.expected {
			t.Errorf("Frequency mapping: input=%d, expected=%d, got=%d",
				tt.input, tt.expected, m.CheckFrequency)
		}

		if m.CheckFrequency != tt.input && len(m.Warnings) == 0 {
			t.Errorf("Expected warning for frequency adjustment from %d to %d",
				tt.input, m.CheckFrequency)
		}
	}
}

func TestTerraformNameGeneration(t *testing.T) {
	conv := converter.NewConverter()

	tests := []struct {
		input    string
		expected string
	}{
		{"Production API", "production_api"},
		{"[PROD]-API-Health", "prod_api_health"},
		{"Test_Monitor_123", "test_monitor_123"},
		{"Monitor (v2)", "monitor_v2"},
		{"123-Monitor", "r_123_monitor"},
		{"", "monitor"},
	}

	for _, tt := range tests {
		monitors := []uptimerobot.Monitor{
			{
				ID:           1,
				FriendlyName: tt.input,
				URL:          "https://example.com",
				Type:         1,
				Interval:     60,
			},
		}

		result := conv.Convert(monitors, nil)

		if len(result.Monitors) != 1 {
			t.Fatalf("Expected 1 monitor, got %d", len(result.Monitors))
		}

		m := result.Monitors[0]

		if m.ResourceName != tt.expected {
			t.Errorf("Terraform name: input='%s', expected='%s', got='%s'",
				tt.input, tt.expected, m.ResourceName)
		}
	}
}

func TestConverterMultipleMonitors(t *testing.T) {
	conv := converter.NewConverter()

	httpMethod := 1
	port := 443
	keywordType := 1
	keywordValue := "ok"

	monitors := []uptimerobot.Monitor{
		{
			ID:           1,
			FriendlyName: "HTTP Monitor",
			URL:          "https://example.com",
			Type:         1,
			HTTPMethod:   &httpMethod,
			Interval:     60,
		},
		{
			ID:           2,
			FriendlyName: "Keyword Monitor",
			URL:          "https://api.example.com",
			Type:         2,
			KeywordType:  &keywordType,
			KeywordValue: &keywordValue,
			Interval:     120,
		},
		{
			ID:           3,
			FriendlyName: "Ping Monitor",
			URL:          "10.0.0.1",
			Type:         3,
			Interval:     30,
		},
		{
			ID:           4,
			FriendlyName: "Port Monitor",
			URL:          "db.local",
			Type:         4,
			Port:         &port,
			Interval:     300,
		},
		{
			ID:           5,
			FriendlyName: "Heartbeat",
			Type:         5,
			Interval:     3600,
		},
	}

	result := conv.Convert(monitors, nil)

	if len(result.Monitors) != 4 {
		t.Errorf("Expected 4 monitors, got %d", len(result.Monitors))
	}

	if len(result.Healthchecks) != 1 {
		t.Errorf("Expected 1 healthcheck, got %d", len(result.Healthchecks))
	}

	if len(result.Skipped) != 0 {
		t.Errorf("Expected 0 skipped, got %d", len(result.Skipped))
	}
}

func TestAlertContactCategorization(t *testing.T) {
	contacts := []uptimerobot.AlertContact{
		{ID: "1", Type: 2, Value: "test@example.com"},
		{ID: "2", Type: 3, Value: "+1234567890"},
		{ID: "3", Type: 4, Value: "https://webhook.example.com"},
		{ID: "4", Type: 11, FriendlyName: "Slack Channel"},
		{ID: "5", Type: 14, FriendlyName: "PagerDuty"},
	}

	info := converter.CategorizeAlertContacts(contacts)

	if len(info.Emails) != 1 {
		t.Errorf("Expected 1 email, got %d", len(info.Emails))
	}

	if len(info.SMSPhones) != 1 {
		t.Errorf("Expected 1 SMS phone, got %d", len(info.SMSPhones))
	}

	if len(info.Webhooks) != 1 {
		t.Errorf("Expected 1 webhook, got %d", len(info.Webhooks))
	}

	if len(info.Slack) != 1 {
		t.Errorf("Expected 1 Slack integration, got %d", len(info.Slack))
	}

	if len(info.PagerDuty) != 1 {
		t.Errorf("Expected 1 PagerDuty integration, got %d", len(info.PagerDuty))
	}
}
