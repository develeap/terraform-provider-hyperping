// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"

	"github.com/develeap/terraform-provider-hyperping/internal/client"
)

func TestNewHealthcheckDataSource(t *testing.T) {
	ds := NewHealthcheckDataSource()
	if ds == nil {
		t.Fatal("NewHealthcheckDataSource returned nil")
	}
	if _, ok := ds.(*HealthcheckDataSource); !ok {
		t.Errorf("expected *HealthcheckDataSource, got %T", ds)
	}
}

func TestHealthcheckDataSource_Metadata(t *testing.T) {
	d := &HealthcheckDataSource{}

	req := datasource.MetadataRequest{
		ProviderTypeName: "hyperping",
	}
	resp := &datasource.MetadataResponse{}

	d.Metadata(context.Background(), req, resp)

	if resp.TypeName != "hyperping_healthcheck" {
		t.Errorf("Expected type name 'hyperping_healthcheck', got '%s'", resp.TypeName)
	}
}

func TestHealthcheckDataSource_Schema(t *testing.T) {
	d := &HealthcheckDataSource{}

	req := datasource.SchemaRequest{}
	resp := &datasource.SchemaResponse{}

	d.Schema(context.Background(), req, resp)

	if _, ok := resp.Schema.Attributes["id"]; !ok {
		t.Error("Schema missing 'id' attribute")
	}

	computedAttrs := []string{
		"name", "ping_url", "cron", "timezone", "period_value", "period_type",
		"grace_period_value", "grace_period_type", "escalation_policy",
		"is_paused", "is_down", "period", "grace_period", "last_ping", "created_at",
	}
	for _, attr := range computedAttrs {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("Schema missing '%s' attribute", attr)
		}
	}
}

func TestHealthcheckDataSource_Configure(t *testing.T) {
	t.Run("valid client", func(t *testing.T) {
		d := &HealthcheckDataSource{}
		c := &client.Client{}

		req := datasource.ConfigureRequest{
			ProviderData: c,
		}
		resp := &datasource.ConfigureResponse{}

		d.Configure(context.Background(), req, resp)

		if resp.Diagnostics.HasError() {
			t.Errorf("Unexpected error: %v", resp.Diagnostics)
		}

		if d.client == nil {
			t.Error("Expected client to be set")
		}
	})

	t.Run("nil provider data", func(t *testing.T) {
		d := &HealthcheckDataSource{}

		req := datasource.ConfigureRequest{
			ProviderData: nil,
		}
		resp := &datasource.ConfigureResponse{}

		d.Configure(context.Background(), req, resp)

		if resp.Diagnostics.HasError() {
			t.Error("Expected no error when provider data is nil")
		}
	})

	t.Run("wrong type", func(t *testing.T) {
		d := &HealthcheckDataSource{}

		req := datasource.ConfigureRequest{
			ProviderData: "wrong type",
		}
		resp := &datasource.ConfigureResponse{}

		d.Configure(context.Background(), req, resp)

		if !resp.Diagnostics.HasError() {
			t.Fatal("Expected error when provider data is wrong type")
		}
	})
}

func TestMapHealthcheckToDataSourceModel(t *testing.T) {
	t.Run("all fields populated", func(t *testing.T) {
		periodVal := 60
		hc := &client.Healthcheck{
			UUID:             "hc-full",
			Name:             "Full Healthcheck",
			PingURL:          "https://ping.hyperping.io/full",
			Cron:             "*/5 * * * *",
			Timezone:         "America/New_York",
			PeriodValue:      &periodVal,
			PeriodType:       "minutes",
			GracePeriodValue: 10,
			GracePeriodType:  "minutes",
			EscalationPolicy: &client.EscalationPolicyReference{
				UUID: "ep-789",
				Name: "Test Policy",
			},
			IsPaused:    true,
			IsDown:      true,
			Period:      3600,
			GracePeriod: 600,
			LastPing:    "2026-01-15T12:34:56Z",
			CreatedAt:   "2026-01-01T00:00:00Z",
		}

		model := &HealthcheckDataSourceModel{}
		mapHealthcheckToDataSourceModel(hc, model)

		if model.ID.ValueString() != "hc-full" {
			t.Errorf("Expected ID 'hc-full', got %s", model.ID.ValueString())
		}
		if model.Name.ValueString() != "Full Healthcheck" {
			t.Errorf("Expected name 'Full Healthcheck', got %s", model.Name.ValueString())
		}
		if model.Cron.ValueString() != "*/5 * * * *" {
			t.Errorf("Expected cron '*/5 * * * *', got %s", model.Cron.ValueString())
		}
		if model.IsPaused.ValueBool() != true {
			t.Error("Expected IsPaused to be true")
		}
		if model.IsDown.ValueBool() != true {
			t.Error("Expected IsDown to be true")
		}
	})

	t.Run("minimal fields", func(t *testing.T) {
		hc := &client.Healthcheck{
			UUID:             "hc-min",
			Name:             "Minimal",
			PingURL:          "https://ping.hyperping.io/min",
			GracePeriodValue: 1,
			GracePeriodType:  "minutes",
			GracePeriod:      60,
		}

		model := &HealthcheckDataSourceModel{}
		mapHealthcheckToDataSourceModel(hc, model)

		if model.ID.ValueString() != "hc-min" {
			t.Errorf("Expected ID 'hc-min', got %s", model.ID.ValueString())
		}
		if !model.Cron.IsNull() {
			t.Error("Expected Cron to be null for empty value")
		}
		if !model.EscalationPolicy.IsNull() {
			t.Error("Expected EscalationPolicy to be null for nil reference")
		}
	})

	t.Run("null escalation policy", func(t *testing.T) {
		hc := &client.Healthcheck{
			UUID:             "hc-no-ep",
			Name:             "No Escalation Policy",
			PingURL:          "https://ping.hyperping.io/no-ep",
			GracePeriodValue: 5,
			GracePeriodType:  "minutes",
			GracePeriod:      300,
			EscalationPolicy: nil,
		}

		model := &HealthcheckDataSourceModel{}
		mapHealthcheckToDataSourceModel(hc, model)

		if !model.EscalationPolicy.IsNull() {
			t.Error("Expected EscalationPolicy to be null when nil in response")
		}
	})
}
