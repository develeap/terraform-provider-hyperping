// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

// NOTE: VCR cassettes for nested service groups are not yet recorded.
// Recording requires a real Hyperping API instance with grouped status pages.
// These tests use httptest mock servers to verify the request/response shapes.

package client

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// =============================================================================
// GetStatusPage — nested service groups (read path)
// =============================================================================

func TestGetStatusPage_WithNestedGroups(t *testing.T) {
	mockBody := map[string]interface{}{
		"statuspage": map[string]interface{}{
			"uuid": "sp_nested_test",
			"name": "Nested Groups Page",
			"settings": map[string]interface{}{
				"name":             "Nested Groups Page",
				"theme":            "light",
				"font":             "Inter",
				"accent_color":     "#0000ff",
				"languages":        []string{"en"},
				"default_language": "en",
			},
			"sections": []interface{}{
				map[string]interface{}{
					"name":     map[string]string{"en": "Main Section"},
					"is_split": false,
					"services": []interface{}{
						// Flat service: ID is a string UUID
						map[string]interface{}{
							"id":       "mon_abc123",
							"uuid":     "mon_abc123",
							"name":     map[string]string{"en": "Flat Service"},
							"is_group": false,
						},
						// Group header: no top-level monitor, is_group=true, has children
						map[string]interface{}{
							"uuid":     "",
							"name":     map[string]string{"en": "Payment Processing Group"},
							"is_group": true,
							"services": []interface{}{
								map[string]interface{}{
									"id":       117122,
									"uuid":     "child_uuid_1",
									"name":     map[string]string{"en": "Child Service 1"},
									"is_group": false,
								},
								map[string]interface{}{
									"id":       117123,
									"uuid":     "child_uuid_2",
									"name":     map[string]string{"en": "Child Service 2"},
									"is_group": false,
								},
							},
						},
					},
				},
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(mockBody); err != nil {
			t.Errorf("failed to encode mock response: %v", err)
		}
	}))
	defer server.Close()

	client := NewClient("test_key", WithBaseURL(server.URL), WithMaxRetries(0))

	result, err := client.GetStatusPage(context.Background(), "sp_nested_test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Sections) == 0 {
		t.Fatal("expected at least one section")
	}

	section := result.Sections[0]
	if len(section.Services) != 2 {
		t.Fatalf("expected 2 services in section[0], got %d", len(section.Services))
	}

	// --- Flat service assertions ---
	flatSvc := section.Services[0]

	flatID, ok := flatSvc.ID.(string)
	if !ok {
		t.Fatalf("expected flat service ID to be string, got %T", flatSvc.ID)
	}
	if flatID != "mon_abc123" {
		t.Errorf("expected flat service ID %q, got %q", "mon_abc123", flatID)
	}
	if flatSvc.IsGroup {
		t.Error("expected flat service IsGroup=false")
	}

	// --- Group service assertions ---
	groupSvc := section.Services[1]

	if !groupSvc.IsGroup {
		t.Error("expected group service IsGroup=true")
	}
	if len(groupSvc.Services) != 2 {
		t.Fatalf("expected 2 child services in group, got %d", len(groupSvc.Services))
	}

	// Child 1: ID should unmarshal as float64 (JSON number → interface{} = float64)
	child1 := groupSvc.Services[0]
	child1ID, ok := child1.ID.(float64)
	if !ok {
		t.Fatalf("expected child service ID to be float64 (JSON number), got %T", child1.ID)
	}
	if child1ID != 117122 {
		t.Errorf("expected child1 ID 117122, got %v", child1ID)
	}
	if child1.UUID != "child_uuid_1" {
		t.Errorf("expected child1 UUID %q, got %q", "child_uuid_1", child1.UUID)
	}

	// Child 2: verify UUID is present and distinct
	child2 := groupSvc.Services[1]
	if child2.UUID != "child_uuid_2" {
		t.Errorf("expected child2 UUID %q, got %q", "child_uuid_2", child2.UUID)
	}
}

// =============================================================================
// CreateStatusPage — group service payload (write path)
// =============================================================================

func TestCreateStatusPage_GroupServicePayload(t *testing.T) {
	var capturedBody []byte

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		var readErr error
		capturedBody, readErr = io.ReadAll(r.Body)
		if readErr != nil {
			t.Errorf("failed to read request body: %v", readErr)
		}

		// Return a minimal 201 response with a generated status page
		response := map[string]interface{}{
			"message": "Status page created",
			"statuspage": map[string]interface{}{
				"uuid": "sp_generated_001",
				"name": "Test Page",
				"settings": map[string]interface{}{
					"name":             "Test Page",
					"theme":            "light",
					"font":             "Inter",
					"accent_color":     "#000000",
					"languages":        []string{"en"},
					"default_language": "en",
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			t.Errorf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	client := NewClient("test_key", WithBaseURL(server.URL), WithMaxRetries(0))

	req := CreateStatusPageRequest{
		Name:      "Test Page",
		Subdomain: stringPtr("test-sub"),
		Languages: []string{"en"},
		Sections: []CreateStatusPageSection{
			{
				Name: "Infrastructure",
				Services: []CreateStatusPageService{
					{
						// Group header — no MonitorUUID, has children
						IsGroup:   boolPtr(true),
						NameShown: stringPtr("Payment Processing"),
						Services: []CreateStatusPageService{
							{
								UUID: stringPtr("child_uuid_1"),
								Name: map[string]string{"en": "Child 1"},
							},
							{
								UUID: stringPtr("child_uuid_2"),
								Name: map[string]string{"en": "Child 2"},
							},
						},
					},
				},
			},
		},
	}

	result, err := client.CreateStatusPage(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.UUID != "sp_generated_001" {
		t.Errorf("expected UUID %q, got %q", "sp_generated_001", result.UUID)
	}

	// Parse the captured request body and verify the payload shape
	if capturedBody == nil {
		t.Fatal("captured request body is nil; cannot verify payload")
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(capturedBody, &payload); err != nil {
		t.Fatalf("failed to parse captured request body: %v", err)
	}

	sections, ok := payload["sections"].([]interface{})
	if !ok || len(sections) == 0 {
		t.Fatal("expected sections array in payload")
	}

	section, ok := sections[0].(map[string]interface{})
	if !ok {
		t.Fatal("expected section to be an object")
	}

	services, ok := section["services"].([]interface{})
	if !ok || len(services) == 0 {
		t.Fatal("expected services array in section")
	}

	groupSvc, ok := services[0].(map[string]interface{})
	if !ok {
		t.Fatal("expected group service to be an object")
	}

	// Assert is_group == true
	isGroup, _ := groupSvc["is_group"].(bool)
	if !isGroup {
		t.Errorf("expected group service to have is_group=true, got %v", groupSvc["is_group"])
	}

	// Assert monitor_uuid is absent or null
	if monUUID, exists := groupSvc["monitor_uuid"]; exists && monUUID != nil {
		t.Errorf("expected group service to have no monitor_uuid, got %v", monUUID)
	}

	// Assert services array with 2 children
	children, ok := groupSvc["services"].([]interface{})
	if !ok {
		t.Fatal("expected group service to have a services array")
	}
	if len(children) != 2 {
		t.Fatalf("expected 2 children in group service, got %d", len(children))
	}

	// Assert each child has "uuid" key and "name" map
	expectedChildUUIDs := []string{"child_uuid_1", "child_uuid_2"}
	expectedChildNames := []string{"Child 1", "Child 2"}

	for i, rawChild := range children {
		child, ok := rawChild.(map[string]interface{})
		if !ok {
			t.Fatalf("child[%d] is not an object", i)
		}

		// "uuid" key must be present (not "monitor_uuid")
		childUUID, uuidOk := child["uuid"].(string)
		if !uuidOk {
			t.Errorf("child[%d] expected uuid key (string), got %T", i, child["uuid"])
		} else if childUUID != expectedChildUUIDs[i] {
			t.Errorf("child[%d] expected uuid %q, got %q", i, expectedChildUUIDs[i], childUUID)
		}

		// "monitor_uuid" must not be present on nested children
		if monUUID, exists := child["monitor_uuid"]; exists && monUUID != nil {
			t.Errorf("child[%d] should not have monitor_uuid, got %v", i, monUUID)
		}

		// "name" must be a map (localized), not a string
		nameMap, nameOk := child["name"].(map[string]interface{})
		if !nameOk {
			t.Errorf("child[%d] expected name to be a map, got %T", i, child["name"])
		} else {
			enName, _ := nameMap["en"].(string)
			if enName != expectedChildNames[i] {
				t.Errorf("child[%d] expected name[en]=%q, got %q", i, expectedChildNames[i], enName)
			}
		}

		// "name_shown" must not be present on nested children
		if nameShown, exists := child["name_shown"]; exists && nameShown != nil {
			t.Errorf("child[%d] should not have name_shown, got %v", i, nameShown)
		}
	}
}
