// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/develeap/terraform-provider-hyperping/internal/client"
)

// mockIncidentUpdateAPI implements client.IncidentAPI for unit testing
// the IncidentUpdateResource CRUD methods.
type mockIncidentUpdateAPI struct {
	getIncidentFunc       func(ctx context.Context, id string) (*client.Incident, error)
	addIncidentUpdateFunc func(ctx context.Context, id string, req client.AddIncidentUpdateRequest) (*client.Incident, error)
	// Unused methods required to satisfy the interface.
	listIncidentsFunc   func(ctx context.Context) ([]client.Incident, error)
	createIncidentFunc  func(ctx context.Context, req client.CreateIncidentRequest) (*client.Incident, error)
	updateIncidentFunc  func(ctx context.Context, id string, req client.UpdateIncidentRequest) (*client.Incident, error)
	deleteIncidentFunc  func(ctx context.Context, id string) error
	resolveIncidentFunc func(ctx context.Context, uuid string, message string) (*client.Incident, error)
}

func (m *mockIncidentUpdateAPI) ListIncidents(ctx context.Context) ([]client.Incident, error) {
	if m.listIncidentsFunc != nil {
		return m.listIncidentsFunc(ctx)
	}
	return nil, errors.New("ListIncidents not implemented")
}

func (m *mockIncidentUpdateAPI) GetIncident(ctx context.Context, id string) (*client.Incident, error) {
	if m.getIncidentFunc != nil {
		return m.getIncidentFunc(ctx, id)
	}
	return nil, errors.New("GetIncident not implemented")
}

func (m *mockIncidentUpdateAPI) CreateIncident(ctx context.Context, req client.CreateIncidentRequest) (*client.Incident, error) {
	if m.createIncidentFunc != nil {
		return m.createIncidentFunc(ctx, req)
	}
	return nil, errors.New("CreateIncident not implemented")
}

func (m *mockIncidentUpdateAPI) UpdateIncident(ctx context.Context, id string, req client.UpdateIncidentRequest) (*client.Incident, error) {
	if m.updateIncidentFunc != nil {
		return m.updateIncidentFunc(ctx, id, req)
	}
	return nil, errors.New("UpdateIncident not implemented")
}

func (m *mockIncidentUpdateAPI) DeleteIncident(ctx context.Context, id string) error {
	if m.deleteIncidentFunc != nil {
		return m.deleteIncidentFunc(ctx, id)
	}
	return errors.New("DeleteIncident not implemented")
}

func (m *mockIncidentUpdateAPI) AddIncidentUpdate(ctx context.Context, uuid string, req client.AddIncidentUpdateRequest) (*client.Incident, error) {
	if m.addIncidentUpdateFunc != nil {
		return m.addIncidentUpdateFunc(ctx, uuid, req)
	}
	return nil, errors.New("AddIncidentUpdate not implemented")
}

func (m *mockIncidentUpdateAPI) ResolveIncident(ctx context.Context, uuid string, message string) (*client.Incident, error) {
	if m.resolveIncidentFunc != nil {
		return m.resolveIncidentFunc(ctx, uuid, message)
	}
	return nil, errors.New("ResolveIncident not implemented")
}

// getIncidentUpdateSchema returns the IncidentUpdateResource schema for testing.
func getIncidentUpdateSchema() resource.SchemaResponse {
	r := &IncidentUpdateResource{}
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)
	return *resp
}

// incidentUpdateTFType returns the tftypes.Object type matching the schema.
func incidentUpdateTFType() tftypes.Object {
	return tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"id":          tftypes.String,
			"incident_id": tftypes.String,
			"text":        tftypes.String,
			"type":        tftypes.String,
			"date":        tftypes.String,
		},
	}
}

// buildIncidentUpdatePlan creates a tfsdk.Plan for testing Create.
func buildIncidentUpdatePlan(incidentID, text, updateType, date string) tfsdk.Plan {
	schemaResp := getIncidentUpdateSchema()
	objType := incidentUpdateTFType()

	dateVal := tftypes.NewValue(tftypes.String, nil)
	if date != "" {
		dateVal = tftypes.NewValue(tftypes.String, date)
	}

	raw := tftypes.NewValue(objType, map[string]tftypes.Value{
		"id":          tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		"incident_id": tftypes.NewValue(tftypes.String, incidentID),
		"text":        tftypes.NewValue(tftypes.String, text),
		"type":        tftypes.NewValue(tftypes.String, updateType),
		"date":        dateVal,
	})

	return tfsdk.Plan{
		Schema: schemaResp.Schema,
		Raw:    raw,
	}
}

// buildIncidentUpdateState creates a tfsdk.State for testing Read/Update/Delete.
func buildIncidentUpdateState(id, incidentID, text, updateType, date string) tfsdk.State {
	schemaResp := getIncidentUpdateSchema()
	objType := incidentUpdateTFType()

	dateVal := tftypes.NewValue(tftypes.String, nil)
	if date != "" {
		dateVal = tftypes.NewValue(tftypes.String, date)
	}

	raw := tftypes.NewValue(objType, map[string]tftypes.Value{
		"id":          tftypes.NewValue(tftypes.String, id),
		"incident_id": tftypes.NewValue(tftypes.String, incidentID),
		"text":        tftypes.NewValue(tftypes.String, text),
		"type":        tftypes.NewValue(tftypes.String, updateType),
		"date":        dateVal,
	})

	return tfsdk.State{
		Schema: schemaResp.Schema,
		Raw:    raw,
	}
}

func TestNewIncidentUpdateResource(t *testing.T) {
	r := NewIncidentUpdateResource()
	if r == nil {
		t.Error("Expected non-nil resource")
	}
}

func TestIncidentUpdateResource_Metadata(t *testing.T) {
	r := &IncidentUpdateResource{}

	req := resource.MetadataRequest{
		ProviderTypeName: "hyperping",
	}
	resp := &resource.MetadataResponse{}

	r.Metadata(context.Background(), req, resp)

	if resp.TypeName != "hyperping_incident_update" {
		t.Errorf("Expected type name 'hyperping_incident_update', got '%s'", resp.TypeName)
	}
}

func TestIncidentUpdateResource_Schema(t *testing.T) {
	r := &IncidentUpdateResource{}

	req := resource.SchemaRequest{}
	resp := &resource.SchemaResponse{}

	r.Schema(context.Background(), req, resp)

	// Verify essential attributes exist
	requiredAttrs := []string{"id", "incident_id", "text", "type", "date"}
	for _, attr := range requiredAttrs {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("Schema missing '%s' attribute", attr)
		}
	}
}

func TestIncidentUpdateResource_ConfigureWrongType(t *testing.T) {
	r := &IncidentUpdateResource{}

	req := resource.ConfigureRequest{
		ProviderData: "wrong type",
	}
	resp := &resource.ConfigureResponse{}

	r.Configure(context.Background(), req, resp)

	if !resp.Diagnostics.HasError() {
		t.Error("Expected error for wrong type")
	}
}

func TestIncidentUpdateResource_ConfigureNilProviderData(t *testing.T) {
	r := &IncidentUpdateResource{}

	req := resource.ConfigureRequest{
		ProviderData: nil,
	}
	resp := &resource.ConfigureResponse{}

	r.Configure(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Error("Expected no error when provider data is nil")
	}
}

func TestIncidentUpdateResource_ConfigureValidClient(t *testing.T) {
	r := &IncidentUpdateResource{}

	c := client.NewClient("test_api_key")

	req := resource.ConfigureRequest{
		ProviderData: c,
	}
	resp := &resource.ConfigureResponse{}

	r.Configure(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("Unexpected error: %v", resp.Diagnostics)
	}

	if r.client == nil {
		t.Error("Expected client to be set")
	}
}

func TestParseIncidentUpdateID(t *testing.T) {
	tests := []struct {
		name             string
		id               string
		expectedIncident string
		expectedUpdate   string
	}{
		{
			name:             "valid composite ID",
			id:               "inci_123/update_456",
			expectedIncident: "inci_123",
			expectedUpdate:   "update_456",
		},
		{
			name:             "ID with multiple slashes",
			id:               "inci_123/update_456/extra",
			expectedIncident: "inci_123",
			expectedUpdate:   "update_456/extra",
		},
		{
			name:             "empty ID",
			id:               "",
			expectedIncident: "",
			expectedUpdate:   "",
		},
		{
			name:             "no slash",
			id:               "invalid",
			expectedIncident: "",
			expectedUpdate:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			incidentID, updateID := parseIncidentUpdateID(tt.id)
			if incidentID != tt.expectedIncident {
				t.Errorf("Expected incident ID '%s', got '%s'", tt.expectedIncident, incidentID)
			}
			if updateID != tt.expectedUpdate {
				t.Errorf("Expected update ID '%s', got '%s'", tt.expectedUpdate, updateID)
			}
		})
	}
}

// =============================================================================
// Create Tests
// =============================================================================

func TestIncidentUpdateResource_Create_Success(t *testing.T) {
	mock := &mockIncidentUpdateAPI{
		addIncidentUpdateFunc: func(_ context.Context, id string, req client.AddIncidentUpdateRequest) (*client.Incident, error) {
			if id != "incident-uuid-1" {
				t.Errorf("expected incident ID 'incident-uuid-1', got %q", id)
			}
			if req.Text.En != "We are investigating the issue." {
				t.Errorf("unexpected text: %q", req.Text.En)
			}
			if req.Type != "investigating" {
				t.Errorf("unexpected type: %q", req.Type)
			}
			return &client.Incident{
				UUID: "incident-uuid-1",
				Updates: []client.IncidentUpdate{
					{UUID: "update-uuid-old", Date: "2026-01-01T00:00:00Z", Text: client.LocalizedText{En: "Old update"}, Type: "identified"},
					{UUID: "update-uuid-1", Date: "2026-01-02T10:00:00Z", Text: client.LocalizedText{En: "We are investigating the issue."}, Type: "investigating"},
				},
			}, nil
		},
	}

	r := &IncidentUpdateResource{client: mock}
	ctx := context.Background()

	req := resource.CreateRequest{
		Plan: buildIncidentUpdatePlan("incident-uuid-1", "We are investigating the issue.", "investigating", ""),
	}
	resp := &resource.CreateResponse{
		State: buildIncidentUpdateState("", "", "", "", ""),
	}

	r.Create(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected errors: %v", resp.Diagnostics.Errors())
	}

	// Verify the state was set correctly
	var state IncidentUpdateResourceModel
	diags := resp.State.Get(ctx, &state)
	if diags.HasError() {
		t.Fatalf("failed to read state: %v", diags.Errors())
	}

	expectedID := "incident-uuid-1/update-uuid-1"
	if state.ID.ValueString() != expectedID {
		t.Errorf("expected ID %q, got %q", expectedID, state.ID.ValueString())
	}
	if state.Date.ValueString() != "2026-01-02T10:00:00Z" {
		t.Errorf("expected date '2026-01-02T10:00:00Z', got %q", state.Date.ValueString())
	}
	if state.IncidentID.ValueString() != "incident-uuid-1" {
		t.Errorf("expected incident_id 'incident-uuid-1', got %q", state.IncidentID.ValueString())
	}
}

func TestIncidentUpdateResource_Create_WithDate(t *testing.T) {
	mock := &mockIncidentUpdateAPI{
		addIncidentUpdateFunc: func(_ context.Context, _ string, req client.AddIncidentUpdateRequest) (*client.Incident, error) {
			if req.Date != "2026-03-15T12:00:00Z" {
				t.Errorf("expected date '2026-03-15T12:00:00Z', got %q", req.Date)
			}
			return &client.Incident{
				UUID: "incident-uuid-1",
				Updates: []client.IncidentUpdate{
					{UUID: "update-uuid-1", Date: "2026-03-15T12:00:00Z", Text: req.Text, Type: req.Type},
				},
			}, nil
		},
	}

	r := &IncidentUpdateResource{client: mock}
	ctx := context.Background()

	req := resource.CreateRequest{
		Plan: buildIncidentUpdatePlan("incident-uuid-1", "Identified the root cause.", "identified", "2026-03-15T12:00:00Z"),
	}
	resp := &resource.CreateResponse{
		State: buildIncidentUpdateState("", "", "", "", ""),
	}

	r.Create(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected errors: %v", resp.Diagnostics.Errors())
	}

	var state IncidentUpdateResourceModel
	diags := resp.State.Get(ctx, &state)
	if diags.HasError() {
		t.Fatalf("failed to read state: %v", diags.Errors())
	}

	if state.Date.ValueString() != "2026-03-15T12:00:00Z" {
		t.Errorf("expected date '2026-03-15T12:00:00Z', got %q", state.Date.ValueString())
	}
}

func TestIncidentUpdateResource_Create_APIError(t *testing.T) {
	mock := &mockIncidentUpdateAPI{
		addIncidentUpdateFunc: func(_ context.Context, _ string, _ client.AddIncidentUpdateRequest) (*client.Incident, error) {
			return nil, fmt.Errorf("API connection refused")
		},
	}

	r := &IncidentUpdateResource{client: mock}
	ctx := context.Background()

	req := resource.CreateRequest{
		Plan: buildIncidentUpdatePlan("incident-uuid-1", "Some update text", "investigating", ""),
	}
	resp := &resource.CreateResponse{
		State: buildIncidentUpdateState("", "", "", "", ""),
	}

	r.Create(ctx, req, resp)

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected error diagnostic when API returns error")
	}

	found := false
	for _, d := range resp.Diagnostics.Errors() {
		if d.Summary() == "Error creating incident update" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'Error creating incident update' diagnostic, got: %v", resp.Diagnostics.Errors())
	}
}

func TestIncidentUpdateResource_Create_FallbackToLastUpdate(t *testing.T) {
	// When no matching type+text is found (e.g., API normalizes text),
	// the resource should fall back to the last update.
	mock := &mockIncidentUpdateAPI{
		addIncidentUpdateFunc: func(_ context.Context, _ string, _ client.AddIncidentUpdateRequest) (*client.Incident, error) {
			return &client.Incident{
				UUID: "incident-uuid-1",
				Updates: []client.IncidentUpdate{
					{UUID: "update-uuid-old", Date: "2026-01-01T00:00:00Z", Text: client.LocalizedText{En: "Old"}, Type: "investigating"},
					// The API normalized the text, so it doesn't match the request exactly
					{UUID: "update-uuid-normalized", Date: "2026-01-02T10:00:00Z", Text: client.LocalizedText{En: "Normalized text by API"}, Type: "update"},
				},
			}, nil
		},
	}

	r := &IncidentUpdateResource{client: mock}
	ctx := context.Background()

	req := resource.CreateRequest{
		Plan: buildIncidentUpdatePlan("incident-uuid-1", "Original text that API normalized", "monitoring", ""),
	}
	resp := &resource.CreateResponse{
		State: buildIncidentUpdateState("", "", "", "", ""),
	}

	r.Create(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected errors: %v", resp.Diagnostics.Errors())
	}

	var state IncidentUpdateResourceModel
	diags := resp.State.Get(ctx, &state)
	if diags.HasError() {
		t.Fatalf("failed to read state: %v", diags.Errors())
	}

	// Should fall back to the last update
	expectedID := "incident-uuid-1/update-uuid-normalized"
	if state.ID.ValueString() != expectedID {
		t.Errorf("expected fallback ID %q, got %q", expectedID, state.ID.ValueString())
	}
	if state.Date.ValueString() != "2026-01-02T10:00:00Z" {
		t.Errorf("expected fallback date, got %q", state.Date.ValueString())
	}
}

func TestIncidentUpdateResource_Create_EmptyUpdatesNoDate(t *testing.T) {
	// Edge case: API returns incident with no updates at all.
	mock := &mockIncidentUpdateAPI{
		addIncidentUpdateFunc: func(_ context.Context, _ string, _ client.AddIncidentUpdateRequest) (*client.Incident, error) {
			return &client.Incident{
				UUID:    "incident-uuid-1",
				Updates: []client.IncidentUpdate{},
			}, nil
		},
	}

	r := &IncidentUpdateResource{client: mock}
	ctx := context.Background()

	req := resource.CreateRequest{
		Plan: buildIncidentUpdatePlan("incident-uuid-1", "Some text", "investigating", ""),
	}
	resp := &resource.CreateResponse{
		State: buildIncidentUpdateState("", "", "", "", ""),
	}

	r.Create(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected errors: %v", resp.Diagnostics.Errors())
	}

	var state IncidentUpdateResourceModel
	diags := resp.State.Get(ctx, &state)
	if diags.HasError() {
		t.Fatalf("failed to read state: %v", diags.Errors())
	}

	// Should have empty update UUID and null date
	expectedID := "incident-uuid-1/"
	if state.ID.ValueString() != expectedID {
		t.Errorf("expected ID %q (empty update UUID), got %q", expectedID, state.ID.ValueString())
	}
	if !state.Date.IsNull() {
		t.Errorf("expected null date when no updates, got %q", state.Date.ValueString())
	}
}

// =============================================================================
// Read Tests
// =============================================================================

func TestIncidentUpdateResource_Read_Success(t *testing.T) {
	mock := &mockIncidentUpdateAPI{
		getIncidentFunc: func(_ context.Context, id string) (*client.Incident, error) {
			if id != "incident-uuid-1" {
				t.Errorf("expected incident ID 'incident-uuid-1', got %q", id)
			}
			return &client.Incident{
				UUID: "incident-uuid-1",
				Updates: []client.IncidentUpdate{
					{UUID: "update-uuid-1", Date: "2026-01-02T10:00:00Z", Text: client.LocalizedText{En: "Investigating issue"}, Type: "investigating"},
					{UUID: "update-uuid-2", Date: "2026-01-02T11:00:00Z", Text: client.LocalizedText{En: "Found root cause"}, Type: "identified"},
				},
			}, nil
		},
	}

	r := &IncidentUpdateResource{client: mock}
	ctx := context.Background()

	req := resource.ReadRequest{
		State: buildIncidentUpdateState("incident-uuid-1/update-uuid-2", "incident-uuid-1", "old text", "identified", "2026-01-01T00:00:00Z"),
	}
	resp := &resource.ReadResponse{
		State: buildIncidentUpdateState("incident-uuid-1/update-uuid-2", "incident-uuid-1", "old text", "identified", "2026-01-01T00:00:00Z"),
	}

	r.Read(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected errors: %v", resp.Diagnostics.Errors())
	}

	var state IncidentUpdateResourceModel
	diags := resp.State.Get(ctx, &state)
	if diags.HasError() {
		t.Fatalf("failed to read state: %v", diags.Errors())
	}

	// Should have refreshed values from the API
	if state.Text.ValueString() != "Found root cause" {
		t.Errorf("expected text 'Found root cause', got %q", state.Text.ValueString())
	}
	if state.Type.ValueString() != "identified" {
		t.Errorf("expected type 'identified', got %q", state.Type.ValueString())
	}
	if state.Date.ValueString() != "2026-01-02T11:00:00Z" {
		t.Errorf("expected date '2026-01-02T11:00:00Z', got %q", state.Date.ValueString())
	}
}

func TestIncidentUpdateResource_Read_IncidentNotFound(t *testing.T) {
	mock := &mockIncidentUpdateAPI{
		getIncidentFunc: func(_ context.Context, _ string) (*client.Incident, error) {
			return nil, fmt.Errorf("failed to get incident: %w", client.ErrNotFound)
		},
	}

	r := &IncidentUpdateResource{client: mock}
	ctx := context.Background()

	initialState := buildIncidentUpdateState("incident-uuid-1/update-uuid-1", "incident-uuid-1", "text", "investigating", "2026-01-01T00:00:00Z")

	req := resource.ReadRequest{
		State: initialState,
	}
	resp := &resource.ReadResponse{
		State: initialState,
	}

	r.Read(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected errors (404 should remove resource, not error): %v", resp.Diagnostics.Errors())
	}

	// State should be removed (Raw should be null)
	if !resp.State.Raw.IsNull() {
		t.Error("expected state to be removed when incident is not found (404)")
	}
}

func TestIncidentUpdateResource_Read_APIError(t *testing.T) {
	mock := &mockIncidentUpdateAPI{
		getIncidentFunc: func(_ context.Context, _ string) (*client.Incident, error) {
			return nil, fmt.Errorf("internal server error")
		},
	}

	r := &IncidentUpdateResource{client: mock}
	ctx := context.Background()

	state := buildIncidentUpdateState("incident-uuid-1/update-uuid-1", "incident-uuid-1", "text", "investigating", "2026-01-01T00:00:00Z")

	req := resource.ReadRequest{
		State: state,
	}
	resp := &resource.ReadResponse{
		State: state,
	}

	r.Read(ctx, req, resp)

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected error diagnostic when API returns non-404 error")
	}

	found := false
	for _, d := range resp.Diagnostics.Errors() {
		if d.Summary() == "Error reading incident" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'Error reading incident' diagnostic, got: %v", resp.Diagnostics.Errors())
	}
}

func TestIncidentUpdateResource_Read_UpdateNotFound(t *testing.T) {
	mock := &mockIncidentUpdateAPI{
		getIncidentFunc: func(_ context.Context, _ string) (*client.Incident, error) {
			return &client.Incident{
				UUID: "incident-uuid-1",
				Updates: []client.IncidentUpdate{
					{UUID: "update-uuid-other", Date: "2026-01-02T10:00:00Z", Text: client.LocalizedText{En: "Other update"}, Type: "investigating"},
				},
			}, nil
		},
	}

	r := &IncidentUpdateResource{client: mock}
	ctx := context.Background()

	initialState := buildIncidentUpdateState("incident-uuid-1/update-uuid-deleted", "incident-uuid-1", "text", "investigating", "2026-01-01T00:00:00Z")

	req := resource.ReadRequest{
		State: initialState,
	}
	resp := &resource.ReadResponse{
		State: initialState,
	}

	r.Read(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected errors (missing update should remove resource, not error): %v", resp.Diagnostics.Errors())
	}

	// State should be removed
	if !resp.State.Raw.IsNull() {
		t.Error("expected state to be removed when update is not found in incident")
	}
}

func TestIncidentUpdateResource_Read_InvalidCompositeID(t *testing.T) {
	mock := &mockIncidentUpdateAPI{}

	r := &IncidentUpdateResource{client: mock}
	ctx := context.Background()

	state := buildIncidentUpdateState("invalid-no-slash", "incident-uuid-1", "text", "investigating", "2026-01-01T00:00:00Z")

	req := resource.ReadRequest{
		State: state,
	}
	resp := &resource.ReadResponse{
		State: state,
	}

	r.Read(ctx, req, resp)

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected error diagnostic when composite ID is invalid")
	}

	found := false
	for _, d := range resp.Diagnostics.Errors() {
		if d.Summary() == "Invalid Resource ID" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'Invalid Resource ID' diagnostic, got: %v", resp.Diagnostics.Errors())
	}
}

func TestIncidentUpdateResource_Read_EmptyUpdatesInIncident(t *testing.T) {
	mock := &mockIncidentUpdateAPI{
		getIncidentFunc: func(_ context.Context, _ string) (*client.Incident, error) {
			return &client.Incident{
				UUID:    "incident-uuid-1",
				Updates: []client.IncidentUpdate{},
			}, nil
		},
	}

	r := &IncidentUpdateResource{client: mock}
	ctx := context.Background()

	initialState := buildIncidentUpdateState("incident-uuid-1/update-uuid-1", "incident-uuid-1", "text", "investigating", "2026-01-01T00:00:00Z")

	req := resource.ReadRequest{
		State: initialState,
	}
	resp := &resource.ReadResponse{
		State: initialState,
	}

	r.Read(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected errors: %v", resp.Diagnostics.Errors())
	}

	// State should be removed because the update was not found
	if !resp.State.Raw.IsNull() {
		t.Error("expected state to be removed when incident has no updates")
	}
}

// =============================================================================
// Update Tests
// =============================================================================

func TestIncidentUpdateResource_Update_Success(t *testing.T) {
	mock := &mockIncidentUpdateAPI{}

	r := &IncidentUpdateResource{client: mock}
	ctx := context.Background()

	plan := buildIncidentUpdatePlan("incident-uuid-2", "Updated text", "identified", "")
	state := buildIncidentUpdateState("incident-uuid-2/update-uuid-1", "incident-uuid-2", "Old text", "investigating", "2026-01-01T00:00:00Z")

	req := resource.UpdateRequest{
		Plan:  plan,
		State: state,
	}
	resp := &resource.UpdateResponse{
		State: state,
	}

	r.Update(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected errors: %v", resp.Diagnostics.Errors())
	}

	// Verify state preservation: ID and Date come from previous state
	var resultState IncidentUpdateResourceModel
	diags := resp.State.Get(ctx, &resultState)
	if diags.HasError() {
		t.Fatalf("failed to read result state: %v", diags.Errors())
	}

	if resultState.ID.ValueString() != "incident-uuid-2/update-uuid-1" {
		t.Errorf("expected ID preserved from state, got %q", resultState.ID.ValueString())
	}
	if resultState.Date.ValueString() != "2026-01-01T00:00:00Z" {
		t.Errorf("expected date preserved from state, got %q", resultState.Date.ValueString())
	}
	// Plan values for text and type should be in the new state
	if resultState.Text.ValueString() != "Updated text" {
		t.Errorf("expected text 'Updated text' from plan, got %q", resultState.Text.ValueString())
	}
	if resultState.Type.ValueString() != "identified" {
		t.Errorf("expected type 'identified' from plan, got %q", resultState.Type.ValueString())
	}
}

func TestIncidentUpdateResource_Update_PreservesState(t *testing.T) {
	mock := &mockIncidentUpdateAPI{}

	r := &IncidentUpdateResource{client: mock}
	ctx := context.Background()

	// Update preserves computed fields (ID, date) from state regardless of ID format.
	// With RequiresReplace on text and type, Update is only called for date-only changes.
	plan := buildIncidentUpdatePlan("incident-uuid-3", "Same text", "investigating", "")
	state := buildIncidentUpdateState("incident-uuid-3/update-uuid-1", "incident-uuid-3", "Same text", "investigating", "2026-01-15T10:00:00Z")

	req := resource.UpdateRequest{
		Plan:  plan,
		State: state,
	}
	resp := &resource.UpdateResponse{
		State: state,
	}

	r.Update(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected errors: %v", resp.Diagnostics.Errors())
	}

	var resultState IncidentUpdateResourceModel
	diags := resp.State.Get(ctx, &resultState)
	if diags.HasError() {
		t.Fatalf("failed to read result state: %v", diags.Errors())
	}

	if resultState.ID.ValueString() != "incident-uuid-3/update-uuid-1" {
		t.Errorf("expected ID preserved from state, got %q", resultState.ID.ValueString())
	}
	if resultState.Date.ValueString() != "2026-01-15T10:00:00Z" {
		t.Errorf("expected date preserved from state, got %q", resultState.Date.ValueString())
	}
}

// =============================================================================
// Delete Tests
// =============================================================================

func TestIncidentUpdateResource_Delete_Success(t *testing.T) {
	mock := &mockIncidentUpdateAPI{}

	r := &IncidentUpdateResource{client: mock}
	ctx := context.Background()

	state := buildIncidentUpdateState("incident-uuid-1/update-uuid-1", "incident-uuid-1", "Some text", "investigating", "2026-01-01T00:00:00Z")

	req := resource.DeleteRequest{
		State: state,
	}
	resp := &resource.DeleteResponse{
		State: state,
	}

	r.Delete(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected errors: %v", resp.Diagnostics.Errors())
	}

	// Should have a warning about the update persisting in Hyperping
	hasWarning := false
	for _, d := range resp.Diagnostics {
		if d.Severity() == diag.SeverityWarning && d.Summary() == "Incident Update Not Deleted From Hyperping" {
			hasWarning = true
			break
		}
	}
	if !hasWarning {
		t.Error("expected 'Incident Update Not Deleted From Hyperping' warning diagnostic")
	}
}

func TestIncidentUpdateResource_Delete_InvalidID(t *testing.T) {
	mock := &mockIncidentUpdateAPI{}

	r := &IncidentUpdateResource{client: mock}
	ctx := context.Background()

	state := buildIncidentUpdateState("no-slash-invalid", "incident-uuid-1", "Some text", "investigating", "2026-01-01T00:00:00Z")

	req := resource.DeleteRequest{
		State: state,
	}
	resp := &resource.DeleteResponse{
		State: state,
	}

	r.Delete(ctx, req, resp)

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected error diagnostic when ID is invalid")
	}

	found := false
	for _, d := range resp.Diagnostics.Errors() {
		if d.Summary() == "Invalid Resource ID" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'Invalid Resource ID' diagnostic, got: %v", resp.Diagnostics.Errors())
	}
}

func TestIncidentUpdateResource_Delete_WarningContainsIDs(t *testing.T) {
	mock := &mockIncidentUpdateAPI{}

	r := &IncidentUpdateResource{client: mock}
	ctx := context.Background()

	compositeID := "abc-incident/xyz-update"
	state := buildIncidentUpdateState(compositeID, "abc-incident", "text", "investigating", "2026-01-01T00:00:00Z")

	req := resource.DeleteRequest{
		State: state,
	}
	resp := &resource.DeleteResponse{
		State: state,
	}

	r.Delete(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected errors: %v", resp.Diagnostics.Errors())
	}

	// Verify the warning message contains both the composite ID and the incident ID
	for _, d := range resp.Diagnostics {
		if d.Summary() == "Incident Update Not Deleted From Hyperping" {
			detail := d.Detail()
			if !containsSubstr(detail, compositeID) {
				t.Errorf("expected warning detail to contain composite ID %q, got: %s", compositeID, detail)
			}
			if !containsSubstr(detail, "abc-incident") {
				t.Errorf("expected warning detail to contain incident ID 'abc-incident', got: %s", detail)
			}
			return
		}
	}
	t.Error("expected 'Incident Update Not Deleted From Hyperping' warning diagnostic")
}

// containsSubstr checks if a string contains a substring.
func containsSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
