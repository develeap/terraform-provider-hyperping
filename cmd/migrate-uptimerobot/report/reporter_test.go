// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package report

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/develeap/terraform-provider-hyperping/cmd/migrate-uptimerobot/converter"
	"github.com/develeap/terraform-provider-hyperping/cmd/migrate-uptimerobot/uptimerobot"
)

func sampleResult() *converter.ConversionResult {
	return &converter.ConversionResult{
		Monitors: []converter.HyperpingMonitor{
			{OriginalID: 1, Name: "API", ResourceName: "api", Protocol: "http"},
			{OriginalID: 2, Name: "Keyword", ResourceName: "kw", Protocol: "http", RequiredKeyword: "ok", Warnings: []string{"freq adjusted"}},
			{OriginalID: 3, Name: "Ping", ResourceName: "ping", Protocol: "icmp"},
			{OriginalID: 4, Name: "Port", ResourceName: "port", Protocol: "port"},
		},
		Healthchecks: []converter.HyperpingHealthcheck{
			{OriginalID: 5, Name: "HB", ResourceName: "hb", Warnings: []string{"check script"}},
		},
		Skipped: []converter.SkippedMonitor{
			{ID: 6, Name: "Bad", Type: 99, Reason: "unsupported monitor type: 99"},
		},
		ContactsMap: map[string][]string{},
	}
}

func TestGenerate_Counts(t *testing.T) {
	monitors := []uptimerobot.Monitor{
		{ID: 1}, {ID: 2}, {ID: 3}, {ID: 4}, {ID: 5}, {ID: 6},
	}
	r := Generate(monitors, nil, sampleResult())

	if r.Summary.TotalMonitors != 6 {
		t.Errorf("TotalMonitors = %d, want 6", r.Summary.TotalMonitors)
	}
	if r.Summary.MigratedMonitors != 4 {
		t.Errorf("MigratedMonitors = %d, want 4", r.Summary.MigratedMonitors)
	}
	if r.Summary.MigratedHealthchecks != 1 {
		t.Errorf("MigratedHealthchecks = %d, want 1", r.Summary.MigratedHealthchecks)
	}
	if r.Summary.SkippedMonitors != 1 {
		t.Errorf("SkippedMonitors = %d, want 1", r.Summary.SkippedMonitors)
	}
	if r.Summary.MonitorsWithWarnings != 2 { // keyword monitor + healthcheck
		t.Errorf("MonitorsWithWarnings = %d, want 2", r.Summary.MonitorsWithWarnings)
	}
}

func TestGenerate_OriginalTypeInference(t *testing.T) {
	r := Generate(nil, nil, sampleResult())

	byName := map[string]int{}
	for _, m := range r.Monitors {
		byName[m.OriginalName] = m.OriginalType
	}
	tests := []struct {
		name string
		want int
	}{
		{"API", 1},     // http no keyword → HTTP
		{"Keyword", 2}, // http with required_keyword → Keyword
		{"Ping", 3},    // icmp → Ping
		{"Port", 4},    // port → Port
		{"HB", 5},      // healthcheck → Heartbeat
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := byName[tt.name]; got != tt.want {
				t.Errorf("OriginalType for %s = %d, want %d", tt.name, got, tt.want)
			}
		})
	}
}

func TestGenerate_SkippedRecordedAsErrorAndMonitor(t *testing.T) {
	r := Generate(nil, nil, sampleResult())

	if len(r.Errors) != 1 {
		t.Fatalf("Errors = %d, want 1", len(r.Errors))
	}
	if r.Errors[0].Resource != "Bad" {
		t.Errorf("Errors[0].Resource = %q, want Bad", r.Errors[0].Resource)
	}

	var skipped *MonitorReport
	for i := range r.Monitors {
		if r.Monitors[i].MigrationStatus == "skipped" {
			skipped = &r.Monitors[i]
			break
		}
	}
	if skipped == nil {
		t.Fatal("expected one skipped MonitorReport")
	}
	if skipped.ResourceType != "" || skipped.ResourceName != "" {
		t.Errorf("skipped report should leave Resource{Type,Name} empty, got %+v", skipped)
	}
	if len(skipped.Warnings) == 0 || skipped.Warnings[0] != "unsupported monitor type: 99" {
		t.Errorf("skipped Warnings = %v", skipped.Warnings)
	}
}

func TestGenerate_TimestampFormat(t *testing.T) {
	r := Generate(nil, nil, &converter.ConversionResult{ContactsMap: map[string][]string{}})
	if _, err := time.Parse(time.RFC3339, r.Timestamp); err != nil {
		t.Errorf("Timestamp = %q is not RFC3339: %v", r.Timestamp, err)
	}
}

func TestGenerate_JSONRoundTrip(t *testing.T) {
	r := Generate(nil, nil, sampleResult())
	b, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var back Report
	if err := json.Unmarshal(b, &back); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if back.Summary.MigratedMonitors != r.Summary.MigratedMonitors {
		t.Errorf("MigratedMonitors roundtrip: got %d, want %d", back.Summary.MigratedMonitors, r.Summary.MigratedMonitors)
	}
	if len(back.Monitors) != len(r.Monitors) {
		t.Errorf("Monitors length roundtrip: got %d, want %d", len(back.Monitors), len(r.Monitors))
	}
}

func TestGenerate_WarningsCollected(t *testing.T) {
	r := Generate(nil, nil, sampleResult())

	var seenKW, seenHB bool
	for _, w := range r.Warnings {
		if w.Resource == "Keyword" && w.Message == "freq adjusted" {
			seenKW = true
		}
		if w.Resource == "HB" && w.Message == "check script" {
			seenHB = true
		}
	}
	if !seenKW || !seenHB {
		t.Errorf("expected both keyword and healthcheck warnings, got: %+v", r.Warnings)
	}
}
