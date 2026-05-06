// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package converter

import (
	"strings"
	"testing"

	"github.com/develeap/terraform-provider-hyperping/cmd/migrate-uptimerobot/uptimerobot"
)

func flexInt(n int) *uptimerobot.FlexibleInt {
	v := uptimerobot.FlexibleInt(n)
	return &v
}

func strPtr(s string) *string { return &s }

// TestConvert_DispatchByType walks every type the switch handles and checks the
// branch lands in the right output bucket (Monitors vs Healthchecks vs Skipped)
// with the right protocol where applicable.
func TestConvert_DispatchByType(t *testing.T) {
	tests := []struct {
		name           string
		monitorType    int
		wantMonitors   int
		wantHealthchks int
		wantSkipped    int
		wantProtocol   string
	}{
		{"http", 1, 1, 0, 0, "http"},
		{"keyword", 2, 1, 0, 0, "http"},
		{"ping", 3, 1, 0, 0, "icmp"},
		{"port", 4, 1, 0, 0, "port"},
		{"heartbeat", 5, 0, 1, 0, ""},
		{"unsupported", 99, 0, 0, 1, ""},
	}
	c := NewConverter()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := c.Convert([]uptimerobot.Monitor{
				{ID: 1, FriendlyName: "x", URL: "https://x.example.com", Type: tt.monitorType, Interval: 60},
			}, nil)
			if got := len(r.Monitors); got != tt.wantMonitors {
				t.Errorf("Monitors = %d, want %d", got, tt.wantMonitors)
			}
			if got := len(r.Healthchecks); got != tt.wantHealthchks {
				t.Errorf("Healthchecks = %d, want %d", got, tt.wantHealthchks)
			}
			if got := len(r.Skipped); got != tt.wantSkipped {
				t.Errorf("Skipped = %d, want %d", got, tt.wantSkipped)
			}
			if tt.wantProtocol != "" && r.Monitors[0].Protocol != tt.wantProtocol {
				t.Errorf("Protocol = %q, want %q", r.Monitors[0].Protocol, tt.wantProtocol)
			}
			if tt.monitorType == 99 {
				if !strings.Contains(r.Skipped[0].Reason, "unsupported monitor type") {
					t.Errorf("Skipped reason = %q", r.Skipped[0].Reason)
				}
			}
		})
	}
}

func TestConvertHTTPMonitor_Fields(t *testing.T) {
	c := NewConverter()
	r := c.Convert([]uptimerobot.Monitor{
		{
			ID:           7,
			FriendlyName: "Health Check",
			URL:          "https://api.example.com/health",
			Type:         1,
			HTTPMethod:   flexInt(2), // POST
			Interval:     300,
		},
	}, nil)
	if len(r.Monitors) != 1 {
		t.Fatalf("got %d monitors, want 1", len(r.Monitors))
	}
	m := r.Monitors[0]
	if m.HTTPMethod != "POST" {
		t.Errorf("HTTPMethod = %q, want POST", m.HTTPMethod)
	}
	if m.CheckFrequency != 300 {
		t.Errorf("CheckFrequency = %d, want 300", m.CheckFrequency)
	}
	if m.ExpectedStatusCode != "2xx" {
		t.Errorf("ExpectedStatusCode = %q, want 2xx", m.ExpectedStatusCode)
	}
	if !m.FollowRedirects {
		t.Error("FollowRedirects = false, want true (default)")
	}
	if m.OriginalID != 7 {
		t.Errorf("OriginalID = %d, want 7", m.OriginalID)
	}
	if len(m.Warnings) != 0 {
		t.Errorf("expected no warnings for matching frequency, got %v", m.Warnings)
	}
}

func TestConvertHTTPMonitor_FrequencyAdjustedWarning(t *testing.T) {
	c := NewConverter()
	r := c.Convert([]uptimerobot.Monitor{
		{ID: 1, FriendlyName: "X", URL: "https://x.example.com", Type: 1, Interval: 45},
	}, nil)
	if len(r.Monitors) != 1 {
		t.Fatalf("monitors = %d", len(r.Monitors))
	}
	m := r.Monitors[0]
	// 45s rounds to 30 (closest in AllowedFrequencies). A warning must fire.
	if len(m.Warnings) == 0 {
		t.Error("expected frequency-adjusted warning")
	}
}

func TestConvertKeywordMonitor_ExistsVsNotExists(t *testing.T) {
	tests := []struct {
		name        string
		keywordType *uptimerobot.FlexibleInt
		keywordVal  *string
		wantKeyword string
		wantWarn    bool
	}{
		{
			name:        "exists supported",
			keywordType: flexInt(1),
			keywordVal:  strPtr("ok"),
			wantKeyword: "ok",
			wantWarn:    false,
		},
		{
			name:        "not-exists warns and drops keyword",
			keywordType: flexInt(2),
			keywordVal:  strPtr("error"),
			wantKeyword: "",
			wantWarn:    true,
		},
		{
			name:        "no value, no keyword",
			keywordType: flexInt(1),
			keywordVal:  nil,
			wantKeyword: "",
			wantWarn:    false,
		},
		{
			name:        "empty value treated as missing",
			keywordType: flexInt(1),
			keywordVal:  strPtr(""),
			wantKeyword: "",
			wantWarn:    false,
		},
	}
	c := NewConverter()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := c.Convert([]uptimerobot.Monitor{
				{
					ID: 1, FriendlyName: "K", URL: "https://k.example.com", Type: 2, Interval: 60,
					KeywordType:  tt.keywordType,
					KeywordValue: tt.keywordVal,
				},
			}, nil)
			m := r.Monitors[0]
			if m.RequiredKeyword != tt.wantKeyword {
				t.Errorf("RequiredKeyword = %q, want %q", m.RequiredKeyword, tt.wantKeyword)
			}
			hasWarn := false
			for _, w := range m.Warnings {
				if strings.Contains(w, "must not exist") {
					hasWarn = true
				}
			}
			if hasWarn != tt.wantWarn {
				t.Errorf("must-not-exist warning = %v, want %v (warnings: %v)", hasWarn, tt.wantWarn, m.Warnings)
			}
		})
	}
}

func TestConvertPingMonitor_AddsScheme(t *testing.T) {
	c := NewConverter()
	r := c.Convert([]uptimerobot.Monitor{
		{ID: 1, FriendlyName: "P", URL: "host.example.com", Type: 3, Interval: 60},
	}, nil)
	m := r.Monitors[0]
	if m.Protocol != "icmp" {
		t.Errorf("Protocol = %q, want icmp", m.Protocol)
	}
	if !strings.HasPrefix(m.URL, "https://") {
		t.Errorf("URL = %q, expected scheme prefix", m.URL)
	}
}

func TestConvertPortMonitor_PortResolution(t *testing.T) {
	tests := []struct {
		name     string
		port     *uptimerobot.FlexibleInt
		subType  *uptimerobot.FlexibleInt
		wantPort int
	}{
		{"explicit port wins", flexInt(8080), flexInt(3 /* HTTPS */), 8080},
		{"sub-type HTTPS = 443", nil, flexInt(3), 443},
		{"sub-type HTTP = 80", nil, flexInt(2), 80},
		{"sub-type FTP = 21", nil, flexInt(4), 21},
		{"sub-type SMTP = 25", nil, flexInt(5), 25},
		{"sub-type POP3 = 110", nil, flexInt(6), 110},
		{"sub-type IMAP = 143", nil, flexInt(7), 143},
		{"sub-type unknown = 80", nil, flexInt(99), 80},
		{"no port no sub-type = 80", nil, nil, 80},
	}
	c := NewConverter()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := c.Convert([]uptimerobot.Monitor{
				{
					ID: 1, FriendlyName: "PortMon", URL: "host.example.com", Type: 4, Interval: 60,
					Port: tt.port, SubType: tt.subType,
				},
			}, nil)
			if got := r.Monitors[0].Port; got != tt.wantPort {
				t.Errorf("Port = %d, want %d", got, tt.wantPort)
			}
		})
	}
}

func TestConvertHeartbeatMonitor_PeriodMapping(t *testing.T) {
	tests := []struct {
		name        string
		intervalSec int
		wantValue   int
		wantType    string
	}{
		{"days", 172800, 2, "days"},
		{"hours", 7200, 2, "hours"},
		{"minutes", 600, 10, "minutes"},
		{"seconds fallback", 30, 30, "seconds"},
		{"exactly 1 day", 86400, 1, "days"},
		{"exactly 1 hour", 3600, 1, "hours"},
		{"exactly 1 minute", 60, 1, "minutes"},
	}
	c := NewConverter()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := c.Convert([]uptimerobot.Monitor{
				{ID: 1, FriendlyName: "HB", Type: 5, Interval: tt.intervalSec},
			}, nil)
			if len(r.Healthchecks) != 1 {
				t.Fatalf("Healthchecks = %d", len(r.Healthchecks))
			}
			h := r.Healthchecks[0]
			if h.PeriodValue != tt.wantValue || h.PeriodType != tt.wantType {
				t.Errorf("Period = %d %s, want %d %s", h.PeriodValue, h.PeriodType, tt.wantValue, tt.wantType)
			}
			if h.GracePeriodValue != 1 || h.GracePeriodType != "hours" {
				t.Errorf("Grace = %d %s, want 1 hours", h.GracePeriodValue, h.GracePeriodType)
			}
			if len(h.Warnings) == 0 {
				t.Error("expected note warning on heartbeat conversion")
			}
		})
	}
}

func TestConvert_ContactsMapPopulated(t *testing.T) {
	c := NewConverter()
	r := c.Convert(nil, []uptimerobot.AlertContact{
		{ID: "1", Type: 2, Value: "a@example.com"},
		{ID: "1", Type: 2, Value: "b@example.com"},
		{ID: "2", Type: 4, Value: "https://hook"},
	})
	if got := len(r.ContactsMap["1"]); got != 2 {
		t.Errorf("ContactsMap[1] len = %d, want 2", got)
	}
	if got := len(r.ContactsMap["2"]); got != 1 {
		t.Errorf("ContactsMap[2] len = %d, want 1", got)
	}
}

func TestConvert_ResourceNameDeduplication(t *testing.T) {
	c := NewConverter()
	r := c.Convert([]uptimerobot.Monitor{
		{ID: 1, FriendlyName: "Same Name", URL: "https://a.example.com", Type: 1, Interval: 60},
		{ID: 2, FriendlyName: "Same Name", URL: "https://b.example.com", Type: 1, Interval: 60},
		{ID: 3, FriendlyName: "Same Name", URL: "https://c.example.com", Type: 1, Interval: 60},
	}, nil)
	if len(r.Monitors) != 3 {
		t.Fatalf("got %d", len(r.Monitors))
	}
	seen := map[string]bool{}
	for _, m := range r.Monitors {
		if seen[m.ResourceName] {
			t.Errorf("duplicate ResourceName: %s", m.ResourceName)
		}
		seen[m.ResourceName] = true
	}
}

func TestConvertHTTPMethod(t *testing.T) {
	tests := []struct {
		method *uptimerobot.FlexibleInt
		want   string
	}{
		{nil, "GET"},
		{flexInt(1), "GET"},
		{flexInt(2), "POST"},
		{flexInt(3), "PUT"},
		{flexInt(4), "PATCH"},
		{flexInt(5), "DELETE"},
		{flexInt(6), "HEAD"},
		{flexInt(99), "GET"}, // unknown falls back to GET
	}
	for _, tt := range tests {
		got := convertHTTPMethod(tt.method)
		if got != tt.want {
			t.Errorf("convertHTTPMethod(%v) = %q, want %q", tt.method, got, tt.want)
		}
	}
}

func TestMapSubTypeToPort(t *testing.T) {
	cases := map[int]int{1: 80, 2: 80, 3: 443, 4: 21, 5: 25, 6: 110, 7: 143, 999: 80}
	for in, want := range cases {
		if got := mapSubTypeToPort(in); got != want {
			t.Errorf("mapSubTypeToPort(%d) = %d, want %d", in, got, want)
		}
	}
}

func TestTerraformName_DigitPrefix(t *testing.T) {
	// Property: digit-leading names get the "r_" prefix.
	got := terraformName("123abc")
	if !strings.HasPrefix(got, "r_") {
		t.Errorf("terraformName(\"123abc\") = %q, want r_ prefix", got)
	}
}

func TestTerraformName_EmptyFallback(t *testing.T) {
	got := terraformName("!!!")
	if got != "monitor" {
		t.Errorf("terraformName(\"!!!\") = %q, want \"monitor\"", got)
	}
}

func TestEnsureURLScheme(t *testing.T) {
	if got := ensureURLScheme("example.com"); !strings.HasPrefix(got, "https://") {
		t.Errorf("expected scheme prefix, got %q", got)
	}
	if got := ensureURLScheme("http://example.com"); got != "http://example.com" {
		t.Errorf("scheme should be preserved, got %q", got)
	}
}

func TestMapFrequency_DelegatesToPkg(t *testing.T) {
	// Sanity: 60 is in the allowed list, so this must round-trip exactly.
	if got := mapFrequency(60); got != 60 {
		t.Errorf("mapFrequency(60) = %d, want 60", got)
	}
}
