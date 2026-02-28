// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/develeap/terraform-provider-hyperping/internal/client"
)

func TestTranslateSectionsToNumericIDs_TopLevelService(t *testing.T) {
	uuid := "mon_abc123"
	uuidToID := map[string]int{"mon_abc123": 115746}

	sections := []client.CreateStatusPageSection{
		{
			Name: "API",
			Services: []client.CreateStatusPageService{
				{MonitorUUID: &uuid},
			},
		},
	}

	result := translateSectionsToNumericIDs(sections, uuidToID)

	if result[0].Services[0].MonitorUUID == nil {
		t.Fatal("expected MonitorUUID to be set")
	}
	if *result[0].Services[0].MonitorUUID != "115746" {
		t.Errorf("expected '115746', got %q", *result[0].Services[0].MonitorUUID)
	}

	// Verify original was not mutated
	if *sections[0].Services[0].MonitorUUID != "mon_abc123" {
		t.Error("original section was mutated")
	}
}

func TestTranslateSectionsToNumericIDs_NestedService(t *testing.T) {
	uuid := "mon_def456"
	isGroup := true
	uuidToID := map[string]int{"mon_def456": 115747}

	sections := []client.CreateStatusPageSection{
		{
			Name: "Group",
			Services: []client.CreateStatusPageService{
				{
					IsGroup: &isGroup,
					Services: []client.CreateStatusPageService{
						{UUID: &uuid},
					},
				},
			},
		},
	}

	result := translateSectionsToNumericIDs(sections, uuidToID)

	nested := result[0].Services[0].Services[0]
	if nested.UUID == nil {
		t.Fatal("expected UUID to be set on nested service")
	}
	if *nested.UUID != "115747" {
		t.Errorf("expected '115747', got %q", *nested.UUID)
	}
}

func TestTranslateSectionsToNumericIDs_UnknownUUID(t *testing.T) {
	unknownUUID := "mon_unknown"
	uuidToID := map[string]int{"mon_abc123": 115746}

	sections := []client.CreateStatusPageSection{
		{
			Name: "API",
			Services: []client.CreateStatusPageService{
				{MonitorUUID: &unknownUUID},
			},
		},
	}

	result := translateSectionsToNumericIDs(sections, uuidToID)

	if *result[0].Services[0].MonitorUUID != "mon_unknown" {
		t.Errorf("expected unknown UUID to pass through, got %q", *result[0].Services[0].MonitorUUID)
	}
}

func TestTranslateSectionsToNumericIDs_EmptyMap(t *testing.T) {
	uuid := "mon_abc123"
	sections := []client.CreateStatusPageSection{
		{
			Name: "API",
			Services: []client.CreateStatusPageService{
				{MonitorUUID: &uuid},
			},
		},
	}

	result := translateSectionsToNumericIDs(sections, map[string]int{})

	// Should return original sections unchanged
	if *result[0].Services[0].MonitorUUID != "mon_abc123" {
		t.Errorf("expected UUID unchanged, got %q", *result[0].Services[0].MonitorUUID)
	}
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

func TestTranslateRoundTrip(t *testing.T) {
	uuid := "mon_abc123"
	uuidToID := map[string]int{"mon_abc123": 115746}
	idToUUID := map[string]string{"115746": "mon_abc123"}

	// Forward: UUID -> numeric ID
	sections := []client.CreateStatusPageSection{
		{
			Name: "API",
			Services: []client.CreateStatusPageService{
				{MonitorUUID: &uuid},
			},
		},
	}

	translated := translateSectionsToNumericIDs(sections, uuidToID)
	numericID := *translated[0].Services[0].MonitorUUID

	if numericID != "115746" {
		t.Fatalf("forward translation failed: expected '115746', got %q", numericID)
	}

	// Reverse: numeric ID -> UUID
	sp := &client.StatusPage{
		Sections: []client.StatusPageSection{
			{
				Services: []client.StatusPageService{
					{UUID: numericID, ID: numericID},
				},
			},
		},
	}

	translateStatusPageToUUIDs(sp, idToUUID)

	if sp.Sections[0].Services[0].UUID != "mon_abc123" {
		t.Errorf("round-trip failed: expected 'mon_abc123', got %q", sp.Sections[0].Services[0].UUID)
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
