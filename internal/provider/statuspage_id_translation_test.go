// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"

	"github.com/develeap/terraform-provider-hyperping/internal/client"
)

func TestWarnUnresolvedNumericUUIDs_DetectsDrift(t *testing.T) {
	sp := &client.StatusPage{
		Sections: []client.StatusPageSection{
			{
				Services: []client.StatusPageService{
					{UUID: "117896"},     // drifted
					{UUID: "mon_abc123"}, // correct
					{UUID: "118001"},     // drifted
				},
			},
		},
	}

	var diags diag.Diagnostics
	warnUnresolvedNumericUUIDs(sp, &diags)

	if diags.WarningsCount() != 1 {
		t.Fatalf("expected 1 warning, got %d", diags.WarningsCount())
	}

	detail := diags[0].Detail()
	if !strings.Contains(detail, "117896") || !strings.Contains(detail, "118001") {
		t.Errorf("expected warning to list both drifted UUIDs, got: %s", detail)
	}
	if !strings.Contains(detail, "2 service(s)") {
		t.Errorf("expected count of 2, got: %s", detail)
	}
}

func TestWarnUnresolvedNumericUUIDs_NoWarningWhenClean(t *testing.T) {
	sp := &client.StatusPage{
		Sections: []client.StatusPageSection{
			{
				Services: []client.StatusPageService{
					{UUID: "mon_abc123"},
					{UUID: "mon_def456"},
				},
			},
		},
	}

	var diags diag.Diagnostics
	warnUnresolvedNumericUUIDs(sp, &diags)

	if diags.WarningsCount() > 0 {
		t.Errorf("expected no warnings, got %d", diags.WarningsCount())
	}
}

func TestWarnUnresolvedNumericUUIDs_NestedDrift(t *testing.T) {
	sp := &client.StatusPage{
		Sections: []client.StatusPageSection{
			{
				Services: []client.StatusPageService{
					{
						IsGroup: true,
						UUID:    "",
						Services: []client.StatusPageService{
							{UUID: "117896"},
						},
					},
				},
			},
		},
	}

	var diags diag.Diagnostics
	warnUnresolvedNumericUUIDs(sp, &diags)

	if diags.WarningsCount() != 1 {
		t.Errorf("expected 1 warning, got %d", diags.WarningsCount())
	}
}

func TestWarnUnresolvedNumericUUIDs_NilStatusPage(t *testing.T) {
	var diags diag.Diagnostics
	warnUnresolvedNumericUUIDs(nil, &diags)

	if diags.WarningsCount() > 0 {
		t.Errorf("expected no warnings for nil status page, got %d", diags.WarningsCount())
	}
}

func TestWarnUnresolvedNumericUUIDs_EmptyUUIDSkipped(t *testing.T) {
	sp := &client.StatusPage{
		Sections: []client.StatusPageSection{
			{
				Services: []client.StatusPageService{
					{UUID: "", IsGroup: true}, // group header, no UUID
					{UUID: "mon_abc123"},
				},
			},
		},
	}

	var diags diag.Diagnostics
	warnUnresolvedNumericUUIDs(sp, &diags)

	if diags.WarningsCount() > 0 {
		t.Errorf("expected no warnings, got %d", diags.WarningsCount())
	}
}

func TestIsNumericString(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"115746", true},
		{"0", true},
		{"mon_abc123", false},
		{"", false},
		{"12abc", false},
		{"sp_xyz", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := isNumericString(tt.input); got != tt.want {
				t.Errorf("isNumericString(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
