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
			name: "long name truncated to 30 chars",
			check: pingdom.Check{
				Name: strings.Repeat("ABCD ", 20), // many words, very long
			},
			// Each "ABCD " becomes "Abcd"; concatenated to "AbcdAbcd...". 30-char cutoff.
			want: "[UNKNOWN]-Service-" + strings.Repeat("Abcd", 8)[:30],
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

// Indirectly exercise the category aliases that GenerateName covers.
func TestExtractCategory_AllAliases(t *testing.T) {
	cases := map[string]string{
		"api":         "API",
		"web":         "Web",
		"website":     "Web",
		"database":    "Database",
		"cache":       "Cache",
		"memcached":   "Cache",
		"queue":       "Queue",
		"cdn":         "CDN",
		"dns":         "DNS",
		"mail":        "Mail",
		"smtp":        "Mail",
		"email":       "Mail",
		"service":     "Service",
		"app":         "App",
		"application": "App",
	}
	for tag, want := range cases {
		t.Run(tag, func(t *testing.T) {
			check := pingdom.Check{
				Name: "Foo",
				Tags: []pingdom.Tag{{Name: "prod"}, {Name: tag}},
			}
			got := GenerateName(check)
			expected := "[PROD]-" + want + "-Foo"
			if got != expected {
				t.Errorf("got %q, want %q", got, expected)
			}
		})
	}
}
