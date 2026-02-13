// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"testing"

	"github.com/develeap/terraform-provider-hyperping/cmd/migrate-pingdom/converter"
	"github.com/develeap/terraform-provider-hyperping/cmd/migrate-pingdom/pingdom"
)

func TestCheckConversion(t *testing.T) {
	tests := []struct {
		name      string
		check     pingdom.Check
		supported bool
		protocol  string
	}{
		{
			name: "HTTP check",
			check: pingdom.Check{
				ID:       1,
				Name:     "API Health",
				Type:     "http",
				Hostname: "api.example.com",
				URL:      "/health",
				Tags: []pingdom.Tag{
					{Name: "production", Type: "u"},
					{Name: "api", Type: "u"},
				},
			},
			supported: true,
			protocol:  "http",
		},
		{
			name: "HTTPS check",
			check: pingdom.Check{
				ID:         2,
				Name:       "Secure API",
				Type:       "https",
				Hostname:   "api.example.com",
				URL:        "/v1/status",
				Encryption: true,
				Tags: []pingdom.Tag{
					{Name: "production", Type: "u"},
				},
			},
			supported: true,
			protocol:  "http",
		},
		{
			name: "TCP check",
			check: pingdom.Check{
				ID:       3,
				Name:     "Database",
				Type:     "tcp",
				Hostname: "db.example.com",
				Port:     5432,
				Tags: []pingdom.Tag{
					{Name: "production", Type: "u"},
					{Name: "database", Type: "u"},
				},
			},
			supported: true,
			protocol:  "port",
		},
		{
			name: "PING check",
			check: pingdom.Check{
				ID:       4,
				Name:     "Server Ping",
				Type:     "ping",
				Hostname: "server.example.com",
				Tags: []pingdom.Tag{
					{Name: "staging", Type: "u"},
				},
			},
			supported: true,
			protocol:  "icmp",
		},
		{
			name: "DNS check (unsupported)",
			check: pingdom.Check{
				ID:       5,
				Name:     "DNS Resolution",
				Type:     "dns",
				Hostname: "example.com",
			},
			supported: false,
		},
		{
			name: "UDP check (unsupported)",
			check: pingdom.Check{
				ID:       6,
				Name:     "UDP Service",
				Type:     "udp",
				Hostname: "service.example.com",
				Port:     53,
			},
			supported: false,
		},
	}

	conv := converter.NewCheckConverter()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := conv.Convert(tt.check)

			if result.Supported != tt.supported {
				t.Errorf("expected supported=%v, got %v", tt.supported, result.Supported)
			}

			if tt.supported && result.Monitor != nil {
				if result.Monitor.Protocol != tt.protocol {
					t.Errorf("expected protocol=%s, got %s", tt.protocol, result.Monitor.Protocol)
				}
			}
		})
	}
}

func TestFrequencyConversion(t *testing.T) {
	tests := []struct {
		pingdomMinutes int
		expectedSecs   int
	}{
		{1, 60},    // 60 seconds
		{5, 300},   // 300 seconds
		{10, 600},  // 600 seconds
		{30, 1800}, // 1800 seconds exact
		{60, 3600}, // 3600 seconds exact
	}

	for _, tt := range tests {
		result := converter.ConvertFrequency(tt.pingdomMinutes)
		if result != tt.expectedSecs {
			t.Errorf("convertFrequency(%d) = %d, want %d", tt.pingdomMinutes, result, tt.expectedSecs)
		}
	}
}

func TestTagConversion(t *testing.T) {
	tests := []struct {
		name           string
		check          pingdom.Check
		expectedName   string
		expectEnv      bool
		expectCustomer bool
	}{
		{
			name: "Production API",
			check: pingdom.Check{
				Name: "API Health Check",
				Tags: []pingdom.Tag{
					{Name: "production", Type: "u"},
					{Name: "api", Type: "u"},
				},
			},
			expectedName: "[PROD]-API-ApiHealthCheck",
			expectEnv:    true,
		},
		{
			name: "Staging with customer",
			check: pingdom.Check{
				Name: "Customer Portal",
				Tags: []pingdom.Tag{
					{Name: "staging", Type: "u"},
					{Name: "customer-acme", Type: "u"},
					{Name: "web", Type: "u"},
				},
			},
			expectedName:   "[STAGING-ACME]-Web-CustomerPortal",
			expectEnv:      true,
			expectCustomer: true,
		},
		{
			name: "No tags",
			check: pingdom.Check{
				Name: "Simple Monitor",
				Tags: []pingdom.Tag{},
			},
			expectedName: "[UNKNOWN]-Service-SimpleMonitor",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			name := converter.GenerateName(tt.check)

			if name != tt.expectedName {
				t.Errorf("expected name=%s, got %s", tt.expectedName, name)
			}
		})
	}
}

func TestRegionConversion(t *testing.T) {
	tests := []struct {
		name          string
		probeFilters  []string
		expectRegions []string
	}{
		{
			name:          "North America",
			probeFilters:  []string{"region:NA"},
			expectRegions: []string{"virginia", "oregon"},
		},
		{
			name:          "Europe",
			probeFilters:  []string{"region:EU"},
			expectRegions: []string{"london", "frankfurt"},
		},
		{
			name:          "Multiple regions",
			probeFilters:  []string{"region:NA", "region:EU"},
			expectRegions: []string{"virginia", "oregon", "london", "frankfurt"},
		},
		{
			name:          "No filters (default)",
			probeFilters:  []string{},
			expectRegions: []string{"virginia", "london", "frankfurt", "singapore"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := converter.ConvertRegions(tt.probeFilters)

			// Check that all expected regions are present
			resultMap := make(map[string]bool)
			for _, r := range result {
				resultMap[r] = true
			}

			for _, expected := range tt.expectRegions {
				if !resultMap[expected] {
					t.Errorf("expected region %s not found in result: %v", expected, result)
				}
			}
		})
	}
}

// ConvertFrequency is exported for testing
func ConvertFrequency(minutes int) int {
	return converter.ConvertFrequency(minutes)
}

// ConvertRegions is exported for testing
func ConvertRegions(filters []string) []string {
	return converter.ConvertRegions(filters)
}
