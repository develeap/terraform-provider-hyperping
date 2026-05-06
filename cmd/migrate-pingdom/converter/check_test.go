// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package converter

import (
	"sort"
	"testing"

	"github.com/develeap/terraform-provider-hyperping/cmd/migrate-pingdom/pingdom"
)

func TestConvert_DispatchByType(t *testing.T) {
	tests := []struct {
		name             string
		check            pingdom.Check
		wantSupported    bool
		wantProtocol     string
		wantUnsupported  string
		wantNotesNonZero bool
	}{
		{
			name:          "http",
			check:         pingdom.Check{Type: "http", Hostname: "a.example.com"},
			wantSupported: true,
			wantProtocol:  "http",
		},
		{
			name:          "https",
			check:         pingdom.Check{Type: "https", Hostname: "a.example.com", Encryption: true},
			wantSupported: true,
			wantProtocol:  "http",
		},
		{
			name:          "tcp",
			check:         pingdom.Check{Type: "tcp", Hostname: "db.example.com", Port: 5432},
			wantSupported: true,
			wantProtocol:  "port",
		},
		{
			name:          "ping",
			check:         pingdom.Check{Type: "ping", Hostname: "host.example.com"},
			wantSupported: true,
			wantProtocol:  "icmp",
		},
		{
			name:             "smtp",
			check:            pingdom.Check{Type: "smtp", Hostname: "mail.example.com"},
			wantSupported:    true,
			wantProtocol:     "port",
			wantNotesNonZero: true,
		},
		{
			name:             "pop3",
			check:            pingdom.Check{Type: "pop3", Hostname: "mail.example.com"},
			wantSupported:    true,
			wantProtocol:     "port",
			wantNotesNonZero: true,
		},
		{
			name:             "imap",
			check:            pingdom.Check{Type: "imap", Hostname: "mail.example.com"},
			wantSupported:    true,
			wantProtocol:     "port",
			wantNotesNonZero: true,
		},
		{
			name:             "dns unsupported",
			check:            pingdom.Check{Type: "dns", Hostname: "example.com"},
			wantSupported:    false,
			wantUnsupported:  "dns",
			wantNotesNonZero: true,
		},
		{
			name:             "udp unsupported",
			check:            pingdom.Check{Type: "udp", Hostname: "example.com"},
			wantSupported:    false,
			wantUnsupported:  "udp",
			wantNotesNonZero: true,
		},
		{
			name:             "transaction unsupported",
			check:            pingdom.Check{Type: "transaction", Hostname: "example.com"},
			wantSupported:    false,
			wantUnsupported:  "transaction",
			wantNotesNonZero: true,
		},
		{
			name:             "unknown type",
			check:            pingdom.Check{Type: "weirdo", Hostname: "example.com"},
			wantSupported:    false,
			wantUnsupported:  "weirdo",
			wantNotesNonZero: true,
		},
	}

	c := NewCheckConverter()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := c.Convert(tt.check)

			if result.Supported != tt.wantSupported {
				t.Fatalf("Supported = %v, want %v", result.Supported, tt.wantSupported)
			}
			if tt.wantSupported {
				if result.Monitor == nil {
					t.Fatal("expected Monitor non-nil for supported type")
				}
				if result.Monitor.Protocol != tt.wantProtocol {
					t.Errorf("Protocol = %q, want %q", result.Monitor.Protocol, tt.wantProtocol)
				}
			} else if result.UnsupportedType != tt.wantUnsupported {
				t.Errorf("UnsupportedType = %q, want %q", result.UnsupportedType, tt.wantUnsupported)
			}
			if tt.wantNotesNonZero && len(result.Notes) == 0 {
				t.Error("expected at least one note")
			}
		})
	}
}

func TestConvertHTTPCheck_FieldMapping(t *testing.T) {
	postBody := `{"hello":"world"}`
	check := pingdom.Check{
		Type:              "https",
		Hostname:          "api.example.com",
		URL:               "/v1/health",
		Encryption:        true,
		Resolution:        5,
		Paused:            true,
		PostData:          postBody,
		ShouldContain:     "ok",
		VerifyCertificate: false,
		RequestHeaders: map[string]string{
			"X-Test": "value",
		},
		Tags: []pingdom.Tag{
			{Name: "production", Type: "u"},
			{Name: "api", Type: "u"},
		},
		ProbeFilters: []string{"region:NA"},
	}

	result := NewCheckConverter().Convert(check)
	if !result.Supported {
		t.Fatalf("expected supported, got unsupported (%s)", result.UnsupportedType)
	}
	m := result.Monitor
	if m == nil {
		t.Fatal("expected non-nil monitor")
	}
	if got, want := m.URL, "https://api.example.com/v1/health"; got != want {
		t.Errorf("URL = %q, want %q", got, want)
	}
	if m.HTTPMethod != "POST" {
		t.Errorf("HTTPMethod = %q, want POST (because PostData was set)", m.HTTPMethod)
	}
	if m.RequestBody == nil || *m.RequestBody != postBody {
		t.Errorf("RequestBody = %v, want %q", m.RequestBody, postBody)
	}
	if m.RequiredKeyword == nil || *m.RequiredKeyword != "ok" {
		t.Errorf("RequiredKeyword = %v, want %q", m.RequiredKeyword, "ok")
	}
	if m.ExpectedStatusCode != "200" {
		t.Errorf("ExpectedStatusCode = %q, want 200", m.ExpectedStatusCode)
	}
	if !m.Paused {
		t.Error("Paused = false, want true")
	}
	if len(m.RequestHeaders) != 1 || m.RequestHeaders[0].Name != "X-Test" || m.RequestHeaders[0].Value != "value" {
		t.Errorf("RequestHeaders = %#v", m.RequestHeaders)
	}
	// HTTPS branch overrides FollowRedirects with !VerifyCertificate.
	// VerifyCertificate=false here, so FollowRedirects must be true.
	if m.FollowRedirects == nil || !*m.FollowRedirects {
		t.Errorf("FollowRedirects = %v, want non-nil pointing to true (VerifyCertificate=false on HTTPS)", m.FollowRedirects)
	}
}

// TestConvertHTTPCheck_VerifyCertificateFlipsFollowRedirects locks in the
// (admittedly odd) production behaviour where VerifyCertificate on an HTTPS
// check is what controls FollowRedirects. Tested as a property — flipping
// VerifyCertificate must flip FollowRedirects — so the test does not pretend
// to endorse the semantic mapping.
func TestConvertHTTPCheck_VerifyCertificateFlipsFollowRedirects(t *testing.T) {
	tests := []struct {
		name              string
		verifyCertificate bool
		wantFollow        bool
	}{
		{"verify=false → follow=true", false, true},
		{"verify=true → follow=false", true, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewCheckConverter().Convert(pingdom.Check{
				Type:              "https",
				Hostname:          "api.example.com",
				URL:               "/",
				Encryption:        true,
				VerifyCertificate: tt.verifyCertificate,
			}).Monitor
			if m.FollowRedirects == nil {
				t.Fatal("FollowRedirects is nil, want non-nil for HTTPS check")
			}
			if *m.FollowRedirects != tt.wantFollow {
				t.Errorf("FollowRedirects = %v, want %v", *m.FollowRedirects, tt.wantFollow)
			}
		})
	}
}

// TestConvertHTTPCheck_NoEncryptionFollowsRedirects covers the non-HTTPS path
// where the initial true value survives unchanged.
func TestConvertHTTPCheck_NoEncryptionFollowsRedirects(t *testing.T) {
	m := NewCheckConverter().Convert(pingdom.Check{
		Type:              "http",
		Hostname:          "site.example.com",
		Encryption:        false,
		VerifyCertificate: true, // ignored when Encryption is false
	}).Monitor
	if m.FollowRedirects == nil || !*m.FollowRedirects {
		t.Errorf("FollowRedirects = %v, want non-nil pointing to true (HTTP default)", m.FollowRedirects)
	}
}

func TestConvertHTTPCheck_HTTPProtocolNoEncryption(t *testing.T) {
	check := pingdom.Check{
		Type:       "http",
		Hostname:   "site.example.com",
		URL:        "/",
		Encryption: false,
	}
	m := NewCheckConverter().Convert(check).Monitor
	if m.URL != "http://site.example.com/" {
		t.Errorf("URL = %q, want http://site.example.com/", m.URL)
	}
	if m.HTTPMethod != "GET" {
		t.Errorf("HTTPMethod = %q, want GET", m.HTTPMethod)
	}
	if m.RequestBody != nil {
		t.Errorf("RequestBody = %v, want nil", m.RequestBody)
	}
}

func TestConvertTCPCheck_PortDefault(t *testing.T) {
	tests := []struct {
		name     string
		port     int
		wantPort int
	}{
		{"explicit port", 5432, 5432},
		{"zero defaults to 80", 0, 80},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			check := pingdom.Check{Type: "tcp", Hostname: "db.example.com", Port: tt.port}
			m := NewCheckConverter().Convert(check).Monitor
			if m.Port == nil || *m.Port != tt.wantPort {
				t.Errorf("Port = %v, want %d", m.Port, tt.wantPort)
			}
		})
	}
}

func TestConvertSMTPCheck_PortDefaults(t *testing.T) {
	tests := []struct {
		name       string
		encryption bool
		explicit   int
		wantPort   int
	}{
		{"plain default", false, 0, 25},
		{"encrypted default", true, 0, 587},
		{"explicit overrides", false, 2525, 2525},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewCheckConverter().Convert(pingdom.Check{
				Type:       "smtp",
				Hostname:   "mail.example.com",
				Port:       tt.explicit,
				Encryption: tt.encryption,
			}).Monitor
			if m.Port == nil || *m.Port != tt.wantPort {
				t.Errorf("Port = %v, want %d", m.Port, tt.wantPort)
			}
		})
	}
}

func TestConvertPOP3Check_PortDefaults(t *testing.T) {
	tests := []struct {
		name       string
		encryption bool
		wantPort   int
	}{
		{"plain", false, 110},
		{"encrypted", true, 995},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewCheckConverter().Convert(pingdom.Check{
				Type:       "pop3",
				Hostname:   "mail.example.com",
				Encryption: tt.encryption,
			}).Monitor
			if m.Port == nil || *m.Port != tt.wantPort {
				t.Errorf("Port = %v, want %d", m.Port, tt.wantPort)
			}
		})
	}
}

func TestConvertIMAPCheck_PortDefaults(t *testing.T) {
	tests := []struct {
		name       string
		encryption bool
		wantPort   int
	}{
		{"plain", false, 143},
		{"encrypted", true, 993},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewCheckConverter().Convert(pingdom.Check{
				Type:       "imap",
				Hostname:   "mail.example.com",
				Encryption: tt.encryption,
			}).Monitor
			if m.Port == nil || *m.Port != tt.wantPort {
				t.Errorf("Port = %v, want %d", m.Port, tt.wantPort)
			}
		})
	}
}

func TestConvertPingCheck_Fields(t *testing.T) {
	check := pingdom.Check{
		Type:         "ping",
		Hostname:     "host.example.com",
		Resolution:   1,
		Paused:       true,
		ProbeFilters: []string{"region:EU"},
	}
	m := NewCheckConverter().Convert(check).Monitor
	if m.Protocol != "icmp" {
		t.Errorf("Protocol = %q, want icmp", m.Protocol)
	}
	if m.URL != "host.example.com" {
		t.Errorf("URL = %q, want host.example.com", m.URL)
	}
	if !m.Paused {
		t.Error("Paused = false, want true")
	}
}

func TestConvertFrequency(t *testing.T) {
	tests := []struct {
		minutes int
		want    int
	}{
		{1, 60},
		{5, 300},
		{10, 600},
		{30, 1800},
		{60, 3600},
		// Pingdom supports 15-minute checks too; should round to nearest allowed.
		{15, 600}, // 900s; allowed list jumps 600 -> 1800; 600 is closer.
	}
	for _, tt := range tests {
		got := ConvertFrequency(tt.minutes)
		if got != tt.want {
			t.Errorf("ConvertFrequency(%d) = %d, want %d", tt.minutes, got, tt.want)
		}
	}
}

func TestConvertRegions(t *testing.T) {
	sortedSet := func(in []string) []string {
		out := append([]string{}, in...)
		sort.Strings(out)
		return out
	}

	tests := []struct {
		name    string
		filters []string
		want    []string
	}{
		{
			name:    "empty defaults",
			filters: nil,
			want:    []string{"frankfurt", "london", "singapore", "virginia"},
		},
		{
			name:    "NA",
			filters: []string{"region:NA"},
			want:    []string{"oregon", "virginia"},
		},
		{
			name:    "EU",
			filters: []string{"region:EU"},
			want:    []string{"frankfurt", "london"},
		},
		{
			name:    "APAC",
			filters: []string{"region:APAC"},
			want:    []string{"singapore", "sydney", "tokyo"},
		},
		{
			name:    "LATAM",
			filters: []string{"region:LATAM"},
			want:    []string{"saopaulo"},
		},
		{
			name:    "NA+EU dedup",
			filters: []string{"region:NA", "region:EU", "region:NA"},
			want:    []string{"frankfurt", "london", "oregon", "virginia"},
		},
		{
			name:    "unknown filter falls back",
			filters: []string{"region:MARS"},
			want:    []string{"london", "virginia"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sortedSet(ConvertRegions(tt.filters))
			if len(got) != len(tt.want) {
				t.Fatalf("regions = %v, want %v", got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Fatalf("regions = %v, want %v", got, tt.want)
				}
			}
		})
	}
}
