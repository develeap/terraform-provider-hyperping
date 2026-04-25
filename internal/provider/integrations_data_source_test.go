// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"

	hyperping "github.com/develeap/hyperping-go"
)

func TestNewIntegrationsDataSource(t *testing.T) {
	ds := NewIntegrationsDataSource()
	if ds == nil {
		t.Fatal("NewIntegrationsDataSource returned nil")
	}
	if _, ok := ds.(*IntegrationsDataSource); !ok {
		t.Errorf("expected *IntegrationsDataSource, got %T", ds)
	}
}

func TestIntegrationsDataSource_Metadata(t *testing.T) {
	d := &IntegrationsDataSource{}

	req := datasource.MetadataRequest{
		ProviderTypeName: "hyperping",
	}
	resp := &datasource.MetadataResponse{}

	d.Metadata(context.Background(), req, resp)

	if resp.TypeName != "hyperping_integrations" {
		t.Errorf("expected type name 'hyperping_integrations', got '%s'", resp.TypeName)
	}
}

func TestIntegrationsDataSource_Schema(t *testing.T) {
	d := &IntegrationsDataSource{}

	req := datasource.SchemaRequest{}
	resp := &datasource.SchemaResponse{}

	d.Schema(context.Background(), req, resp)

	if _, ok := resp.Schema.Attributes["ids"]; !ok {
		t.Error("schema missing 'ids' attribute")
	}

	if _, ok := resp.Schema.Attributes["integrations"]; !ok {
		t.Error("schema missing 'integrations' attribute")
	}
}

func TestIntegrationsDataSource_Configure(t *testing.T) {
	t.Run("nil provider data", func(t *testing.T) {
		d := &IntegrationsDataSource{}

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
		d := &IntegrationsDataSource{}

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
		d := &IntegrationsDataSource{}

		transport, err := hyperping.NewMcpTransport("sk_test", "")
		if err != nil {
			t.Fatalf("NewMcpTransport: %v", err)
		}
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
