// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"

	hyperping "github.com/develeap/hyperping-go"
)

func TestNewEscalationPoliciesDataSource(t *testing.T) {
	ds := NewEscalationPoliciesDataSource()
	if ds == nil {
		t.Fatal("NewEscalationPoliciesDataSource returned nil")
	}
	if _, ok := ds.(*EscalationPoliciesDataSource); !ok {
		t.Errorf("expected *EscalationPoliciesDataSource, got %T", ds)
	}
}

func TestEscalationPoliciesDataSource_Metadata(t *testing.T) {
	d := &EscalationPoliciesDataSource{}

	req := datasource.MetadataRequest{
		ProviderTypeName: "hyperping",
	}
	resp := &datasource.MetadataResponse{}

	d.Metadata(context.Background(), req, resp)

	if resp.TypeName != "hyperping_escalation_policies" {
		t.Errorf("expected type name 'hyperping_escalation_policies', got '%s'", resp.TypeName)
	}
}

func TestEscalationPoliciesDataSource_Schema(t *testing.T) {
	d := &EscalationPoliciesDataSource{}

	req := datasource.SchemaRequest{}
	resp := &datasource.SchemaResponse{}

	d.Schema(context.Background(), req, resp)

	if _, ok := resp.Schema.Attributes["ids"]; !ok {
		t.Error("schema missing 'ids' attribute")
	}

	if _, ok := resp.Schema.Attributes["policies"]; !ok {
		t.Error("schema missing 'policies' attribute")
	}
}

func TestEscalationPoliciesDataSource_Configure(t *testing.T) {
	t.Run("nil provider data", func(t *testing.T) {
		d := &EscalationPoliciesDataSource{}

		req := datasource.ConfigureRequest{
			ProviderData: nil,
		}
		resp := &datasource.ConfigureResponse{}

		d.Configure(context.Background(), req, resp)

		if resp.Diagnostics.HasError() {
			t.Error("expected no error when provider data is nil")
		}
		if d.client != nil {
			t.Error("expected client to remain nil when provider data is nil")
		}
	})

	t.Run("wrong type provider data", func(t *testing.T) {
		d := &EscalationPoliciesDataSource{}

		req := datasource.ConfigureRequest{
			ProviderData: "wrong type",
		}
		resp := &datasource.ConfigureResponse{}

		d.Configure(context.Background(), req, resp)

		if !resp.Diagnostics.HasError() {
			t.Fatal("expected error when provider data is wrong type")
		}
	})

	t.Run("valid provider data", func(t *testing.T) {
		d := &EscalationPoliciesDataSource{}

		transport := hyperping.NewMcpTransport("sk_test", "")
		mcpClient := hyperping.NewMCPClient(transport)
		clients := &hyperpingClients{MCP: mcpClient}

		req := datasource.ConfigureRequest{
			ProviderData: clients,
		}
		resp := &datasource.ConfigureResponse{}

		d.Configure(context.Background(), req, resp)

		if resp.Diagnostics.HasError() {
			t.Errorf("unexpected error: %v", resp.Diagnostics)
		}
		if d.client == nil {
			t.Error("expected client to be set after valid configure")
		}
	})
}
