// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package converter

import (
	"strings"
	"testing"

	"github.com/develeap/terraform-provider-hyperping/cmd/migrate-pingdom/pingdom"
)

func TestGenerateName(t *testing.T) {
	tests := []struct {
		name  string
		check pingdom.Check
		want  string
	}{
		{
			name: "production api basic",
			check: pingdom.Check{
				Name: "API Health Check",
				Tags: []pingdom.Tag{{Name: "production"}, {Name: "api"}},
			},
			want: "[PROD]-API-ApiHealthCheck",
		},
		{
			name: "staging with customer prefix",
			check: pingdom.Check{
				Name: "Customer Portal",
				Tags: []pingdom.Tag{
					{Name: "staging"},
					{Name: "customer-acme"},
					{Name: "web"},
				},
			},
			want: "[STAGING-ACME]-Web-CustomerPortal",
		},
		{
			name: "tenant prefix",
			check: pingdom.Check{
				Name: "Worker Health",
				Tags: []pingdom.Tag{
					{Name: "prod"},
					{Name: "tenant-acme"},
					{Name: "worker"},
				},
			},
			want: "[PROD-ACME]-Worker-WorkerHealth",
		},
		{
			name:  "no tags falls back to UNKNOWN/Service",
			check: pingdom.Check{Name: "Simple Monitor"},
			want:  "[UNKNOWN]-Service-SimpleMonitor",
		},
		{
			name: "name has prefix Production - is stripped",
			check: pingdom.Check{
				Name: "Production - User Login",
				Tags: []pingdom.Tag{{Name: "prod"}, {Name: "api"}},
			},
			want: "[PROD]-API-UserLogin",
		},
		{
			name:  "empty service name falls back to Monitor",
			check: pingdom.Check{Name: "!!!"},
			want:  "[UNKNOWN]-Service-Monitor",
		},
		{
			name: "qa environment",
			check: pingdom.Check{
				Name: "Reports",
				Tags: []pingdom.Tag{{Name: "qa"}, {Name: "backend"}},
			},
			want: "[QA]-Backend-Reports",
		},
		{
			name: "test environment",
			check: pingdom.Check{
				Name: "Smoke",
				Tags: []pingdom.Tag{{Name: "test"}, {Name: "frontend"}},
			},
			want: "[TEST]-Frontend-Smoke",
		},
		{
			name: "dev environment, db category alias",
			check: pingdom.Check{
				Name: "Postgres",
				Tags: []pingdom.Tag{{Name: "dev"}, {Name: "db"}},
			},
			want: "[DEV]-Database-Postgres",
		},
		{
			name: "case insensitive tags",
			check: pingdom.Check{
				Name: "Cache",
				Tags: []pingdom.Tag{{Name: "PRODUCTION"}, {Name: "REDIS"}},
			},
			want: "[PROD]-Cache-Cache",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateName(tt.check)
			if got != tt.want {
				t.Errorf("GenerateName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTagsToString(t *testing.T) {
	tests := []struct {
		name string
		tags []pingdom.Tag
		want string
	}{
		{name: "nil", tags: nil, want: ""},
		{name: "empty", tags: []pingdom.Tag{}, want: ""},
		{
			name: "multiple",
			tags: []pingdom.Tag{{Name: "prod"}, {Name: "api"}},
			want: "prod, api",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := TagsToString(tt.tags); got != tt.want {
				t.Errorf("TagsToString = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestExtractCategory_AllAliases indirectly exercises the category aliases that
// GenerateName covers. Cases are kept in a slice rather than a map so subtest
// order is deterministic (helps when running with -shuffle and when reading
// failure output).
func TestExtractCategory_AllAliases(t *testing.T) {
	cases := []struct {
		tag, want string
	}{
		{"api", "API"},
		{"web", "Web"},
		{"website", "Web"},
		{"database", "Database"},
		{"cache", "Cache"},
		{"memcached", "Cache"},
		{"queue", "Queue"},
		{"cdn", "CDN"},
		{"dns", "DNS"},
		{"mail", "Mail"},
		{"smtp", "Mail"},
		{"email", "Mail"},
		{"service", "Service"},
		{"app", "App"},
		{"application", "App"},
	}
	for _, tc := range cases {
		t.Run(tc.tag, func(t *testing.T) {
			check := pingdom.Check{
				Name: "Foo",
				Tags: []pingdom.Tag{{Name: "prod"}, {Name: tc.tag}},
			}
			got := GenerateName(check)
			expected := "[PROD]-" + tc.want + "-Foo"
			if got != expected {
				t.Errorf("got %q, want %q", got, expected)
			}
		})
	}
}

// TestGenerateName_LongNameTruncated checks the length cap as a property
// rather than pinning a specific 30-char value: future tweaks to the cap or
// the casing rules don't require recomputing a hand-built expected string,
// only updating the cap constant in the assertion.
func TestGenerateName_LongNameTruncated(t *testing.T) {
	const cap = 30
	long := strings.Repeat("ABCD ", 20) // 100 chars before sanitisation
	got := GenerateName(pingdom.Check{Name: long})

	const prefix = "[UNKNOWN]-Service-"
	if !strings.HasPrefix(got, prefix) {
		t.Fatalf("got %q, missing env/category prefix %q", got, prefix)
	}
	svc := strings.TrimPrefix(got, prefix)
	if len(svc) > cap {
		t.Errorf("service-name length = %d, want <= %d", len(svc), cap)
	}
	// Confirm the long input actually exercised the truncation path,
	// otherwise this test would silently degrade if the cap moved.
	if len(svc) != cap {
		t.Errorf("expected truncation to exactly %d chars (input is far longer), got %d", cap, len(svc))
	}
	for _, r := range svc {
		if !((r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9')) {
			t.Errorf("unexpected non-alphanumeric char %q in service name %q", r, svc)
		}
	}
}
