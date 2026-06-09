// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package generator

import (
	"strings"
	"testing"

	hyperping "github.com/develeap/hyperping-go"

	"github.com/develeap/terraform-provider-hyperping/cmd/migrate-pingdom/converter"
	"github.com/develeap/terraform-provider-hyperping/cmd/migrate-pingdom/pingdom"
	"github.com/develeap/terraform-provider-hyperping/pkg/migrate"
)

func intPtr(n int) *int    { return &n }
func boolPtr(b bool) *bool { return &b }
func strPtr(s string) *string {
	return &s
}

func TestTerraformName(t *testing.T) {
	tests := []struct {
		name   string
		prefix string
		input  string
		want   string
	}{
		{name: "simple", input: "api", want: "api"},
		{name: "spaces", input: "API Health", want: "api_health"},
		{name: "brackets stripped", input: "[PROD]-API-Service", want: "api_service"},
		{name: "leading number", input: "123-foo", want: "monitor_123_foo"},
		{name: "with prefix", prefix: "p_", input: "API", want: "p_api"},
		{name: "empty becomes monitor", input: "", want: "monitor"},
		{name: "only special", input: "!!!", want: "monitor"},
		{name: "trims underscores", input: "---abc---", want: "abc"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewTerraformGenerator(tt.prefix)
			got := g.terraformName(tt.input)
			if got != tt.want {
				t.Errorf("terraformName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// The local escapeHCL helper that this test originally covered was removed
// in PR #138 (sec: HCL injection hardening) in favour of the shared
// migrate.EscapeHCL/QuoteHCL pair. The generator now calls migrate.QuoteHCL
// directly, so we exercise that here to keep coverage on the HCL-quoting
// path the package depends on.
func TestEscapeHCL(t *testing.T) {
	if got := migrate.EscapeHCL(`he said "hi"`); got != `he said \"hi\"` {
		t.Errorf("EscapeHCL quotes = %q", got)
	}
	if got := migrate.EscapeHCL("a\\b"); got != `a\\b` {
		t.Errorf("EscapeHCL backslash = %q", got)
	}
}

func TestFormatStringList(t *testing.T) {
	tests := []struct {
		in   []string
		want string
	}{
		{nil, "[]"},
		{[]string{}, "[]"},
		{[]string{"a"}, `["a"]`},
		{[]string{"a", "b"}, `["a", "b"]`},
	}
	for _, tt := range tests {
		if got := formatStringList(tt.in); got != tt.want {
			t.Errorf("formatStringList(%v) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestBuildOptionalHelpers(t *testing.T) {
	tests := []struct {
		name string
		fn   func(*hyperping.CreateMonitorRequest) string
		mon  *hyperping.CreateMonitorRequest
		want string
	}{
		{"http_method default GET omitted", buildOptionalHTTPMethod, &hyperping.CreateMonitorRequest{HTTPMethod: "GET"}, ""},
		{"http_method empty omitted", buildOptionalHTTPMethod, &hyperping.CreateMonitorRequest{}, ""},
		{"http_method POST emitted", buildOptionalHTTPMethod, &hyperping.CreateMonitorRequest{HTTPMethod: "POST"}, "  http_method = \"POST\"\n"},

		{"frequency 60 omitted", buildOptionalCheckFrequency, &hyperping.CreateMonitorRequest{CheckFrequency: 60}, ""},
		{"frequency 300 emitted", buildOptionalCheckFrequency, &hyperping.CreateMonitorRequest{CheckFrequency: 300}, "  check_frequency = 300\n"},

		{"empty regions omitted", buildOptionalRegions, &hyperping.CreateMonitorRequest{}, ""},
		{"regions emitted", buildOptionalRegions, &hyperping.CreateMonitorRequest{Regions: []string{"london", "virginia"}}, "  regions = [\"london\", \"virginia\"]\n"},

		{"port nil omitted", buildOptionalPort, &hyperping.CreateMonitorRequest{}, ""},
		{"port zero omitted", buildOptionalPort, &hyperping.CreateMonitorRequest{Port: intPtr(0)}, ""},
		{"port emitted", buildOptionalPort, &hyperping.CreateMonitorRequest{Port: intPtr(5432)}, "  port = 5432\n"},

		{"follow nil omitted", buildOptionalFollowRedirects, &hyperping.CreateMonitorRequest{}, ""},
		{"follow true omitted", buildOptionalFollowRedirects, &hyperping.CreateMonitorRequest{FollowRedirects: boolPtr(true)}, ""},
		{"follow false emitted", buildOptionalFollowRedirects, &hyperping.CreateMonitorRequest{FollowRedirects: boolPtr(false)}, "  follow_redirects = false\n"},

		{"status default omitted", buildOptionalExpectedStatus, &hyperping.CreateMonitorRequest{ExpectedStatusCode: "200"}, ""},
		{"status empty omitted", buildOptionalExpectedStatus, &hyperping.CreateMonitorRequest{}, ""},
		{"status 201 emitted", buildOptionalExpectedStatus, &hyperping.CreateMonitorRequest{ExpectedStatusCode: "201"}, "  expected_status_code = \"201\"\n"},

		{"keyword nil omitted", buildOptionalRequiredKeyword, &hyperping.CreateMonitorRequest{}, ""},
		{"keyword empty omitted", buildOptionalRequiredKeyword, &hyperping.CreateMonitorRequest{RequiredKeyword: strPtr("")}, ""},
		{"keyword emitted", buildOptionalRequiredKeyword, &hyperping.CreateMonitorRequest{RequiredKeyword: strPtr("ok")}, "  required_keyword = \"ok\"\n"},

		{"body nil omitted", buildOptionalRequestBody, &hyperping.CreateMonitorRequest{}, ""},
		{"body empty omitted", buildOptionalRequestBody, &hyperping.CreateMonitorRequest{RequestBody: strPtr("")}, ""},
		{"body emitted (ASCII-safe)", buildOptionalRequestBody, &hyperping.CreateMonitorRequest{RequestBody: strPtr("hello")}, "  request_body = \"hello\"\n"},
		// The earlier double-escape bug (terraform.go formatting an already-escaped
		// string with %q) was fixed in PR #138, which switched the generator to
		// migrate.QuoteHCL. For input {"a":1} the correct HCL output is
		// `"{\"a\":1}"` (1 backslash before each quote), encoded here as the
		// Go literal "  request_body = \"{\\\"a\\\":1}\"\n".
		{"body emitted (json escaped once)", buildOptionalRequestBody, &hyperping.CreateMonitorRequest{RequestBody: strPtr(`{"a":1}`)}, "  request_body = \"{\\\"a\\\":1}\"\n"},

		{"paused false omitted", buildOptionalPaused, &hyperping.CreateMonitorRequest{}, ""},
		{"paused true emitted", buildOptionalPaused, &hyperping.CreateMonitorRequest{Paused: true}, "  paused = true\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.fn(tt.mon); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestBuildOptionalRequestHeaders(t *testing.T) {
	if got := buildOptionalRequestHeaders(&hyperping.CreateMonitorRequest{}); got != "" {
		t.Errorf("expected empty for no headers, got %q", got)
	}
	mon := &hyperping.CreateMonitorRequest{
		RequestHeaders: []hyperping.RequestHeader{
			{Name: "X-Foo", Value: "bar"},
		},
	}
	got := buildOptionalRequestHeaders(mon)
	want := "  request_headers = [\n    {\n      name  = \"X-Foo\"\n      value = \"bar\"\n    },\n  ]\n"
	if got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}

// TestGenerateHCL_Golden snapshots a representative mix of supported and
// unsupported checks. Inputs are deliberately ASCII-only and avoid any field
// that would introduce non-deterministic output (request headers — map
// iteration; probe-filter-derived regions — set→slice). A field-ordering or
// whitespace regression here would silently ship broken HCL to users, so we
// pin the full output rather than substring-asserting.
func TestGenerateHCL_Golden(t *testing.T) {
	checks := []pingdom.Check{
		{
			ID:         1,
			Name:       "API Health",
			Type:       "https",
			Hostname:   "api.example.com",
			URL:        "/health",
			Encryption: true,
			Resolution: 5,
			Tags:       []pingdom.Tag{{Name: "production"}, {Name: "api"}},
		},
		{
			ID:         2,
			Name:       "Database",
			Type:       "tcp",
			Hostname:   "db.example.com",
			Port:       5432,
			Resolution: 1,
			Tags:       []pingdom.Tag{{Name: "production"}, {Name: "database"}},
		},
		{
			ID:       3,
			Name:     "DNS Lookup",
			Type:     "dns",
			Hostname: "example.com",
		},
	}
	conv := converter.NewCheckConverter()
	results := make([]converter.ConversionResult, len(checks))
	for i, c := range checks {
		results[i] = conv.Convert(c)
	}

	got := NewTerraformGenerator("").GenerateHCL(checks, results)
	goldenAssert(t, "monitors.tf.golden", got)
}

func TestGenerateHCL_Unsupported(t *testing.T) {
	check := pingdom.Check{ID: 7, Name: "DNS Check", Type: "dns", Hostname: "x"}
	results := []converter.ConversionResult{
		converter.NewCheckConverter().Convert(check),
	}
	hcl := NewTerraformGenerator("").GenerateHCL([]pingdom.Check{check}, results)

	if !strings.Contains(hcl, "# UNSUPPORTED: dns") {
		t.Errorf("missing unsupported comment:\n%s", hcl)
	}
	if strings.Contains(hcl, `resource "hyperping_monitor"`) {
		t.Errorf("should not emit a resource for unsupported check:\n%s", hcl)
	}
}

func TestGenerateHCL_PrefixApplied(t *testing.T) {
	check := pingdom.Check{
		ID:       2,
		Name:     "Web",
		Type:     "http",
		Hostname: "site.example.com",
		Tags:     []pingdom.Tag{{Name: "production"}, {Name: "web"}},
	}
	results := []converter.ConversionResult{
		converter.NewCheckConverter().Convert(check),
	}
	hcl := NewTerraformGenerator("pd_").GenerateHCL([]pingdom.Check{check}, results)
	if !strings.Contains(hcl, `resource "hyperping_monitor" "pd_`) {
		t.Errorf("expected prefix in resource name:\n%s", hcl)
	}
}

func TestGenerateHCL_Empty(t *testing.T) {
	hcl := NewTerraformGenerator("").GenerateHCL(nil, nil)
	if !strings.HasPrefix(hcl, "# Generated from Pingdom export") {
		t.Errorf("expected header for empty input, got:\n%s", hcl)
	}
}

func TestGenerateHCL_Notes(t *testing.T) {
	check := pingdom.Check{ID: 3, Name: "Mail", Type: "smtp", Hostname: "mail.example.com"}
	results := []converter.ConversionResult{
		converter.NewCheckConverter().Convert(check),
	}
	hcl := NewTerraformGenerator("").GenerateHCL([]pingdom.Check{check}, results)
	if !strings.Contains(hcl, "  # NOTE:") {
		t.Errorf("expected NOTE in HCL:\n%s", hcl)
	}
}
