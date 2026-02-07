// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"

	"github.com/develeap/terraform-provider-hyperping/internal/client"
)

// mockOutageAPI implements client.OutageAPI for testing
type mockOutageAPI struct {
	getOutageFunc           func(ctx context.Context, uuid string) (*client.Outage, error)
	listOutagesFunc         func(ctx context.Context) ([]client.Outage, error)
	createOutageFunc        func(ctx context.Context, req client.CreateOutageRequest) (*client.Outage, error)
	acknowledgeOutageFunc   func(ctx context.Context, uuid string) (*client.OutageAction, error)
	unacknowledgeOutageFunc func(ctx context.Context, uuid string) (*client.OutageAction, error)
	resolveOutageFunc       func(ctx context.Context, uuid string) (*client.OutageAction, error)
	escalateOutageFunc      func(ctx context.Context, uuid string) (*client.OutageAction, error)
	deleteOutageFunc        func(ctx context.Context, uuid string) error
}

func (m *mockOutageAPI) GetOutage(ctx context.Context, uuid string) (*client.Outage, error) {
	if m.getOutageFunc != nil {
		return m.getOutageFunc(ctx, uuid)
	}
	return nil, nil
}

func (m *mockOutageAPI) ListOutages(ctx context.Context) ([]client.Outage, error) {
	if m.listOutagesFunc != nil {
		return m.listOutagesFunc(ctx)
	}
	return nil, nil
}

func (m *mockOutageAPI) CreateOutage(ctx context.Context, req client.CreateOutageRequest) (*client.Outage, error) {
	if m.createOutageFunc != nil {
		return m.createOutageFunc(ctx, req)
	}
	return nil, nil
}

func (m *mockOutageAPI) AcknowledgeOutage(ctx context.Context, uuid string) (*client.OutageAction, error) {
	if m.acknowledgeOutageFunc != nil {
		return m.acknowledgeOutageFunc(ctx, uuid)
	}
	return &client.OutageAction{Message: "acknowledged", UUID: uuid}, nil
}

func (m *mockOutageAPI) UnacknowledgeOutage(ctx context.Context, uuid string) (*client.OutageAction, error) {
	if m.unacknowledgeOutageFunc != nil {
		return m.unacknowledgeOutageFunc(ctx, uuid)
	}
	return &client.OutageAction{Message: "unacknowledged", UUID: uuid}, nil
}

func (m *mockOutageAPI) ResolveOutage(ctx context.Context, uuid string) (*client.OutageAction, error) {
	if m.resolveOutageFunc != nil {
		return m.resolveOutageFunc(ctx, uuid)
	}
	return &client.OutageAction{Message: "resolved", UUID: uuid}, nil
}

func (m *mockOutageAPI) EscalateOutage(ctx context.Context, uuid string) (*client.OutageAction, error) {
	if m.escalateOutageFunc != nil {
		return m.escalateOutageFunc(ctx, uuid)
	}
	return &client.OutageAction{Message: "escalated", UUID: uuid}, nil
}

func (m *mockOutageAPI) DeleteOutage(ctx context.Context, uuid string) error {
	if m.deleteOutageFunc != nil {
		return m.deleteOutageFunc(ctx, uuid)
	}
	return nil
}

func TestNewOutageResource(t *testing.T) {
	r := NewOutageResource()
	if r == nil {
		t.Fatal("expected resource, got nil")
	}
}

func TestOutageResource_Metadata(t *testing.T) {
	r := &OutageResource{}
	req := resource.MetadataRequest{ProviderTypeName: "hyperping"}
	resp := &resource.MetadataResponse{}

	r.Metadata(context.Background(), req, resp)

	if resp.TypeName != "hyperping_outage" {
		t.Errorf("expected TypeName 'hyperping_outage', got %s", resp.TypeName)
	}
}

func TestOutageResource_Schema(t *testing.T) {
	r := &OutageResource{}
	req := resource.SchemaRequest{}
	resp := &resource.SchemaResponse{}

	r.Schema(context.Background(), req, resp)

	if resp.Schema.Attributes == nil {
		t.Fatal("expected schema attributes, got nil")
	}

	// Verify key attributes exist
	requiredAttrs := []string{"id", "monitor_uuid", "start_date", "status_code", "description"}
	for _, attr := range requiredAttrs {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("expected attribute %q in schema", attr)
		}
	}
}

func TestOutageResource_ConfigureWrongType(t *testing.T) {
	r := &OutageResource{}
	req := resource.ConfigureRequest{
		ProviderData: "wrong type",
	}
	resp := &resource.ConfigureResponse{}

	r.Configure(context.Background(), req, resp)

	if !resp.Diagnostics.HasError() {
		t.Error("expected error for wrong provider data type")
	}
}

func TestOutageResource_ConfigureNilProviderData(t *testing.T) {
	r := &OutageResource{}
	req := resource.ConfigureRequest{
		ProviderData: nil,
	}
	resp := &resource.ConfigureResponse{}

	r.Configure(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("unexpected error: %v", resp.Diagnostics.Errors())
	}
}

func TestOutageResource_ConfigureValidClient(t *testing.T) {
	r := &OutageResource{}
	mockClient := &client.Client{}
	req := resource.ConfigureRequest{
		ProviderData: mockClient,
	}
	resp := &resource.ConfigureResponse{}

	r.Configure(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("unexpected error: %v", resp.Diagnostics.Errors())
	}

	if r.client == nil {
		t.Error("expected client to be set")
	}
}

func TestOutageResource_UpdateShouldNotBeCalled(t *testing.T) {
	// This test documents that Update should never be called
	// since all fields are ForceNew
	r := &OutageResource{}
	req := resource.UpdateRequest{}
	resp := &resource.UpdateResponse{}

	// Calling Update should add an error diagnostic
	r.Update(context.Background(), req, resp)

	if !resp.Diagnostics.HasError() {
		t.Error("expected error diagnostic when Update is called")
	}
}
