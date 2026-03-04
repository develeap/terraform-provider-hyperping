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

// --- Tests for buildMonitorIDMaps ---

func TestBuildMonitorIDMaps_HappyPath(t *testing.T) {
	monitors := []client.Monitor{
		{ID: 117896, UUID: "mon_abc123"},
		{ID: 118001, UUID: "mon_def456"},
	}
	listFn := func(_ context.Context) ([]client.Monitor, error) {
		return monitors, nil
	}

	maps, err := buildMonitorIDMaps(context.Background(), listFn)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if maps.uuidToNumericID["mon_abc123"] != "117896" {
		t.Errorf("expected mon_abc123 -> 117896, got %s", maps.uuidToNumericID["mon_abc123"])
	}
	if maps.uuidToNumericID["mon_def456"] != "118001" {
		t.Errorf("expected mon_def456 -> 118001, got %s", maps.uuidToNumericID["mon_def456"])
	}
	if maps.numericIDToUUID["117896"] != "mon_abc123" {
		t.Errorf("expected 117896 -> mon_abc123, got %s", maps.numericIDToUUID["117896"])
	}
	if maps.numericIDToUUID["118001"] != "mon_def456" {
		t.Errorf("expected 118001 -> mon_def456, got %s", maps.numericIDToUUID["118001"])
	}
}

func TestBuildMonitorIDMaps_EmptyList(t *testing.T) {
	listFn := func(_ context.Context) ([]client.Monitor, error) {
		return []client.Monitor{}, nil
	}

	maps, err := buildMonitorIDMaps(context.Background(), listFn)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(maps.uuidToNumericID) != 0 || len(maps.numericIDToUUID) != 0 {
		t.Errorf("expected empty maps, got %d and %d entries",
			len(maps.uuidToNumericID), len(maps.numericIDToUUID))
	}
}

func TestBuildMonitorIDMaps_APIError(t *testing.T) {
	listFn := func(_ context.Context) ([]client.Monitor, error) {
		return nil, fmt.Errorf("API unavailable")
	}

	maps, err := buildMonitorIDMaps(context.Background(), listFn)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if maps != nil {
		t.Errorf("expected nil maps on error, got %+v", maps)
	}
	if !strings.Contains(err.Error(), "failed to list monitors") {
		t.Errorf("expected wrapped error message, got: %s", err.Error())
	}
}

// --- Tests for translateSectionsUUIDsToNumericIDs ---

func TestTranslateSectionsUUIDsToNumericIDs_TopLevel(t *testing.T) {
	uuidToID := map[string]string{
		"mon_abc123": "117896",
		"mon_def456": "118001",
	}
	monUUID1 := "mon_abc123"
	monUUID2 := "mon_def456"
	sections := []client.CreateStatusPageSection{
		{
			Name: "Services",
			Services: []client.CreateStatusPageService{
				{MonitorUUID: &monUUID1},
				{MonitorUUID: &monUUID2},
			},
		},
	}

	var diags diag.Diagnostics
	translateSectionsUUIDsToNumericIDs(sections, uuidToID, &diags)

	if diags.HasError() {
		t.Fatalf("unexpected error: %s", diags.Errors()[0].Detail())
	}
	if *sections[0].Services[0].MonitorUUID != "117896" {
		t.Errorf("expected 117896, got %s", *sections[0].Services[0].MonitorUUID)
	}
	if *sections[0].Services[1].MonitorUUID != "118001" {
		t.Errorf("expected 118001, got %s", *sections[0].Services[1].MonitorUUID)
	}
}

func TestTranslateSectionsUUIDsToNumericIDs_NestedServices(t *testing.T) {
	uuidToID := map[string]string{
		"mon_nested1": "200",
		"mon_nested2": "201",
	}
	nestedUUID1 := "mon_nested1"
	nestedUUID2 := "mon_nested2"
	isGroup := true
	sections := []client.CreateStatusPageSection{
		{
			Name: "Groups",
			Services: []client.CreateStatusPageService{
				{
					IsGroup: &isGroup,
					Services: []client.CreateStatusPageService{
						{UUID: &nestedUUID1},
						{UUID: &nestedUUID2},
					},
				},
			},
		},
	}

	var diags diag.Diagnostics
	translateSectionsUUIDsToNumericIDs(sections, uuidToID, &diags)

	if diags.HasError() {
		t.Fatalf("unexpected error: %s", diags.Errors()[0].Detail())
	}
	if *sections[0].Services[0].Services[0].UUID != "200" {
		t.Errorf("expected 200, got %s", *sections[0].Services[0].Services[0].UUID)
	}
	if *sections[0].Services[0].Services[1].UUID != "201" {
		t.Errorf("expected 201, got %s", *sections[0].Services[0].Services[1].UUID)
	}
}

func TestTranslateSectionsUUIDsToNumericIDs_UnresolvableError(t *testing.T) {
	uuidToID := map[string]string{
		"mon_abc123": "117896",
	}
	knownUUID := "mon_abc123"
	unknownUUID := "mon_unknown"
	sections := []client.CreateStatusPageSection{
		{
			Name: "Services",
			Services: []client.CreateStatusPageService{
				{MonitorUUID: &knownUUID},
				{MonitorUUID: &unknownUUID},
			},
		},
	}

	var diags diag.Diagnostics
	translateSectionsUUIDsToNumericIDs(sections, uuidToID, &diags)

	if !diags.HasError() {
		t.Fatal("expected error for unresolvable UUID")
	}
	detail := diags.Errors()[0].Detail()
	if !strings.Contains(detail, "mon_unknown") {
		t.Errorf("expected unresolved UUID in error, got: %s", detail)
	}
	if !strings.Contains(detail, "1 monitor UUID(s)") {
		t.Errorf("expected count in error, got: %s", detail)
	}
	// Known UUID should still be translated
	if *sections[0].Services[0].MonitorUUID != "117896" {
		t.Errorf("expected known UUID to be translated, got %s", *sections[0].Services[0].MonitorUUID)
	}
}

func TestTranslateSectionsUUIDsToNumericIDs_EmptySections(t *testing.T) {
	var diags diag.Diagnostics
	translateSectionsUUIDsToNumericIDs(nil, map[string]string{}, &diags)

	if diags.HasError() {
		t.Fatalf("unexpected error for nil sections: %s", diags.Errors()[0].Detail())
	}
}

func TestTranslateCreateServicesToNumericIDs_MixedFields(t *testing.T) {
	uuidToID := map[string]string{
		"mon_top":    "100",
		"mon_nested": "200",
	}
	topUUID := "mon_top"
	nestedUUID := "mon_nested"
	isGroup := true
	services := []client.CreateStatusPageService{
		{MonitorUUID: &topUUID},
		{
			IsGroup: &isGroup,
			Services: []client.CreateStatusPageService{
				{UUID: &nestedUUID},
			},
		},
	}

	var unresolved []string
	translateCreateServicesToNumericIDs(services, uuidToID, &unresolved)

	if len(unresolved) > 0 {
		t.Errorf("expected no unresolved UUIDs, got %v", unresolved)
	}
	if *services[0].MonitorUUID != "100" {
		t.Errorf("expected 100, got %s", *services[0].MonitorUUID)
	}
	if *services[1].Services[0].UUID != "200" {
		t.Errorf("expected 200, got %s", *services[1].Services[0].UUID)
	}
}

// --- Tests for translateResponseNumericIDsToUUIDs ---

func TestTranslateResponseNumericIDsToUUIDs_NumericToMon(t *testing.T) {
	idToUUID := map[string]string{
		"117896": "mon_abc123",
		"118001": "mon_def456",
	}
	sp := &client.StatusPage{
		Sections: []client.StatusPageSection{
			{
				Services: []client.StatusPageService{
					{UUID: "117896"},
					{UUID: "118001"},
				},
			},
		},
	}

	var diags diag.Diagnostics
	translateResponseNumericIDsToUUIDs(sp, idToUUID, &diags)

	if diags.HasError() {
		t.Fatalf("unexpected error: %s", diags.Errors()[0].Detail())
	}
	if sp.Sections[0].Services[0].UUID != "mon_abc123" {
		t.Errorf("expected mon_abc123, got %s", sp.Sections[0].Services[0].UUID)
	}
	if sp.Sections[0].Services[1].UUID != "mon_def456" {
		t.Errorf("expected mon_def456, got %s", sp.Sections[0].Services[1].UUID)
	}
}

func TestTranslateResponseNumericIDsToUUIDs_AlreadyMon(t *testing.T) {
	idToUUID := map[string]string{
		"117896": "mon_abc123",
	}
	sp := &client.StatusPage{
		Sections: []client.StatusPageSection{
			{
				Services: []client.StatusPageService{
					{UUID: "mon_abc123"}, // already mon_xxx, should not be touched
				},
			},
		},
	}

	var diags diag.Diagnostics
	translateResponseNumericIDsToUUIDs(sp, idToUUID, &diags)

	if diags.WarningsCount() > 0 {
		t.Errorf("expected no warnings, got %d", diags.WarningsCount())
	}
	if sp.Sections[0].Services[0].UUID != "mon_abc123" {
		t.Errorf("expected mon_abc123 to be unchanged, got %s", sp.Sections[0].Services[0].UUID)
	}
}

func TestTranslateResponseNumericIDsToUUIDs_UnresolvableWarns(t *testing.T) {
	idToUUID := map[string]string{
		"117896": "mon_abc123",
	}
	sp := &client.StatusPage{
		Sections: []client.StatusPageSection{
			{
				Services: []client.StatusPageService{
					{UUID: "117896"},
					{UUID: "999999"}, // unknown numeric ID
				},
			},
		},
	}

	var diags diag.Diagnostics
	translateResponseNumericIDsToUUIDs(sp, idToUUID, &diags)

	if diags.WarningsCount() != 1 {
		t.Fatalf("expected 1 warning, got %d", diags.WarningsCount())
	}
	detail := diags[0].Detail()
	if !strings.Contains(detail, "999999") {
		t.Errorf("expected unresolved ID in warning, got: %s", detail)
	}
	// Known one should still be translated
	if sp.Sections[0].Services[0].UUID != "mon_abc123" {
		t.Errorf("expected mon_abc123, got %s", sp.Sections[0].Services[0].UUID)
	}
}

func TestTranslateResponseNumericIDsToUUIDs_NilStatusPage(t *testing.T) {
	var diags diag.Diagnostics
	translateResponseNumericIDsToUUIDs(nil, map[string]string{}, &diags)

	if diags.HasError() || diags.WarningsCount() > 0 {
		t.Errorf("expected no diagnostics for nil status page")
	}
}

func TestTranslateResponseNumericIDsToUUIDs_NestedServices(t *testing.T) {
	idToUUID := map[string]string{
		"300": "mon_nested",
	}
	sp := &client.StatusPage{
		Sections: []client.StatusPageSection{
			{
				Services: []client.StatusPageService{
					{
						IsGroup: true,
						UUID:    "",
						Services: []client.StatusPageService{
							{UUID: "300"},
						},
					},
				},
			},
		},
	}

	var diags diag.Diagnostics
	translateResponseNumericIDsToUUIDs(sp, idToUUID, &diags)

	if diags.WarningsCount() > 0 {
		t.Errorf("unexpected warnings: %s", diags[0].Detail())
	}
	if sp.Sections[0].Services[0].Services[0].UUID != "mon_nested" {
		t.Errorf("expected mon_nested, got %s", sp.Sections[0].Services[0].Services[0].UUID)
	}
}
