// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"

	"github.com/develeap/terraform-provider-hyperping/internal/client"
)

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
