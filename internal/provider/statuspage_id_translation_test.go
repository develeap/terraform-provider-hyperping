// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"

	"github.com/develeap/terraform-provider-hyperping/internal/client"
)

// mockMonitorAPI implements client.MonitorAPI for testing buildMonitorIDToUUIDMap.
type mockMonitorAPI struct {
	listMonitorsFunc func(ctx context.Context) ([]client.Monitor, error)
}

func (m *mockMonitorAPI) ListMonitors(ctx context.Context) ([]client.Monitor, error) {
	if m.listMonitorsFunc != nil {
		return m.listMonitorsFunc(ctx)
	}
	return nil, nil
}

func (m *mockMonitorAPI) GetMonitor(ctx context.Context, uuid string) (*client.Monitor, error) {
	return nil, nil
}

func (m *mockMonitorAPI) CreateMonitor(ctx context.Context, req client.CreateMonitorRequest) (*client.Monitor, error) {
	return nil, nil
}

func (m *mockMonitorAPI) UpdateMonitor(ctx context.Context, uuid string, req client.UpdateMonitorRequest) (*client.Monitor, error) {
	return nil, nil
}

func (m *mockMonitorAPI) DeleteMonitor(ctx context.Context, uuid string) error {
	return nil
}

func (m *mockMonitorAPI) PauseMonitor(ctx context.Context, uuid string) (*client.Monitor, error) {
	return nil, nil
}

func (m *mockMonitorAPI) ResumeMonitor(ctx context.Context, uuid string) (*client.Monitor, error) {
	return nil, nil
}

func TestTranslateStatusPageToUUIDs_NumericString(t *testing.T) {
	idToUUID := map[string]string{"115746": "mon_abc123"}

	sp := &client.StatusPage{
		Sections: []client.StatusPageSection{
			{
				Services: []client.StatusPageService{
					{UUID: "115746", ID: float64(115746)},
				},
			},
		},
	}

	translateStatusPageToUUIDs(sp, idToUUID)

	if sp.Sections[0].Services[0].UUID != "mon_abc123" {
		t.Errorf("expected 'mon_abc123', got %q", sp.Sections[0].Services[0].UUID)
	}
}

func TestTranslateStatusPageToUUIDs_Float64ID(t *testing.T) {
	idToUUID := map[string]string{"115746": "mon_abc123"}

	sp := &client.StatusPage{
		Sections: []client.StatusPageSection{
			{
				Services: []client.StatusPageService{
					{UUID: "", ID: float64(115746)},
				},
			},
		},
	}

	translateStatusPageToUUIDs(sp, idToUUID)

	if sp.Sections[0].Services[0].UUID != "mon_abc123" {
		t.Errorf("expected 'mon_abc123', got %q", sp.Sections[0].Services[0].UUID)
	}
}

func TestTranslateStatusPageToUUIDs_NestedServices(t *testing.T) {
	idToUUID := map[string]string{"115747": "mon_def456"}

	sp := &client.StatusPage{
		Sections: []client.StatusPageSection{
			{
				Services: []client.StatusPageService{
					{
						IsGroup: true,
						Services: []client.StatusPageService{
							{UUID: "115747", ID: float64(115747)},
						},
					},
				},
			},
		},
	}

	translateStatusPageToUUIDs(sp, idToUUID)

	nested := sp.Sections[0].Services[0].Services[0]
	if nested.UUID != "mon_def456" {
		t.Errorf("expected 'mon_def456', got %q", nested.UUID)
	}
}

func TestTranslateStatusPageToUUIDs_EmptyMap(t *testing.T) {
	sp := &client.StatusPage{
		Sections: []client.StatusPageSection{
			{
				Services: []client.StatusPageService{
					{UUID: "115746", ID: float64(115746)},
				},
			},
		},
	}

	translateStatusPageToUUIDs(sp, map[string]string{})

	// Should remain unchanged
	if sp.Sections[0].Services[0].UUID != "115746" {
		t.Errorf("expected '115746' unchanged, got %q", sp.Sections[0].Services[0].UUID)
	}
}

func TestTranslateStatusPageToUUIDs_AlreadyUUID(t *testing.T) {
	idToUUID := map[string]string{"115746": "mon_abc123"}

	sp := &client.StatusPage{
		Sections: []client.StatusPageSection{
			{
				Services: []client.StatusPageService{
					{UUID: "mon_xyz789"},
				},
			},
		},
	}

	translateStatusPageToUUIDs(sp, idToUUID)

	// Should remain unchanged — already a valid UUID
	if sp.Sections[0].Services[0].UUID != "mon_xyz789" {
		t.Errorf("expected 'mon_xyz789', got %q", sp.Sections[0].Services[0].UUID)
	}
}

func TestServiceIDToNumericString(t *testing.T) {
	tests := []struct {
		name string
		id   interface{}
		want string
	}{
		{"float64", float64(115746), "115746"},
		{"numeric string", "115746", "115746"},
		{"uuid string", "mon_abc123", ""},
		{"nil", nil, ""},
		{"empty string", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := serviceIDToNumericString(tt.id)
			if got != tt.want {
				t.Errorf("serviceIDToNumericString(%v) = %q, want %q", tt.id, got, tt.want)
			}
		})
	}
}

func TestWarnUnresolvedNumericUUIDs_DetectsDrift(t *testing.T) {
	sp := &client.StatusPage{
		Sections: []client.StatusPageSection{
			{
				Services: []client.StatusPageService{
					{UUID: "117896"}, // drifted
					{UUID: "mon_abc123"},
					{UUID: "118001"}, // drifted
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

func TestBuildMonitorIDToUUIDMap_Success(t *testing.T) {
	mock := &mockMonitorAPI{
		listMonitorsFunc: func(ctx context.Context) ([]client.Monitor, error) {
			return []client.Monitor{
				{UUID: "mon_abc123", ID: 115746},
				{UUID: "mon_def456", ID: 115747},
			}, nil
		},
	}

	var diags diag.Diagnostics
	result := buildMonitorIDToUUIDMap(context.Background(), mock, &diags)

	if diags.HasError() || diags.WarningsCount() > 0 {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if result["115746"] != "mon_abc123" {
		t.Errorf("expected mon_abc123 for 115746, got %q", result["115746"])
	}
	if result["115747"] != "mon_def456" {
		t.Errorf("expected mon_def456 for 115747, got %q", result["115747"])
	}
}

func TestBuildMonitorIDToUUIDMap_ErrorReturnsEmpty(t *testing.T) {
	mock := &mockMonitorAPI{
		listMonitorsFunc: func(ctx context.Context) ([]client.Monitor, error) {
			return nil, fmt.Errorf("connection refused")
		},
	}

	var diags diag.Diagnostics
	result := buildMonitorIDToUUIDMap(context.Background(), mock, &diags)

	if len(result) != 0 {
		t.Errorf("expected empty map on error, got %d entries", len(result))
	}
	if diags.WarningsCount() != 1 {
		t.Errorf("expected 1 warning, got %d", diags.WarningsCount())
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
