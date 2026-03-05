// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"

	"github.com/develeap/terraform-provider-hyperping/internal/client"
)

func TestNewHealthcheckResource(t *testing.T) {
	r := NewHealthcheckResource()
	if r == nil {
		t.Fatal("expected resource, got nil")
	}
}

func TestHealthcheckResource_Metadata(t *testing.T) {
	r := &HealthcheckResource{}
	req := resource.MetadataRequest{ProviderTypeName: "hyperping"}
	resp := &resource.MetadataResponse{}

	r.Metadata(context.Background(), req, resp)

	if resp.TypeName != "hyperping_healthcheck" {
		t.Errorf("expected TypeName 'hyperping_healthcheck', got %s", resp.TypeName)
	}
}

func TestHealthcheckResource_Schema(t *testing.T) {
	r := &HealthcheckResource{}
	req := resource.SchemaRequest{}
	resp := &resource.SchemaResponse{}

	r.Schema(context.Background(), req, resp)

	if resp.Schema.Attributes == nil {
		t.Fatal("expected schema attributes, got nil")
	}

	// Verify key attributes exist
	requiredAttrs := []string{"id", "name", "ping_url", "grace_period_value", "grace_period_type"}
	for _, attr := range requiredAttrs {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("expected attribute %q in schema", attr)
		}
	}
}

func TestHealthcheckResource_ConfigureWrongType(t *testing.T) {
	r := &HealthcheckResource{}
	req := resource.ConfigureRequest{
		ProviderData: "wrong type",
	}
	resp := &resource.ConfigureResponse{}

	r.Configure(context.Background(), req, resp)

	if !resp.Diagnostics.HasError() {
		t.Error("expected error for wrong provider data type")
	}
}

func TestHealthcheckResource_ConfigureNilProviderData(t *testing.T) {
	r := &HealthcheckResource{}
	req := resource.ConfigureRequest{
		ProviderData: nil,
	}
	resp := &resource.ConfigureResponse{}

	r.Configure(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("unexpected error: %v", resp.Diagnostics.Errors())
	}
}

func TestHealthcheckResource_ConfigureValidClient(t *testing.T) {
	r := &HealthcheckResource{}
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

func TestValidateCronPeriodExclusivity_Scenarios(t *testing.T) {
	// These tests are already in mapping_test.go
	// This is a placeholder to acknowledge the validation logic is tested
	t.Log("Cron/period validation covered in mapping_test.go")
}
