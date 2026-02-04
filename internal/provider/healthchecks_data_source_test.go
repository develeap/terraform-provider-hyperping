// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"

	"github.com/develeap/terraform-provider-hyperping/internal/client"
)

func TestNewHealthchecksDataSource(t *testing.T) {
	ds := NewHealthchecksDataSource()
	if ds == nil {
		t.Fatal("NewHealthchecksDataSource returned nil")
	}
	if _, ok := ds.(*HealthchecksDataSource); !ok {
		t.Errorf("expected *HealthchecksDataSource, got %T", ds)
	}
}

func TestHealthchecksDataSource_Metadata(t *testing.T) {
	d := &HealthchecksDataSource{}

	req := datasource.MetadataRequest{
		ProviderTypeName: "hyperping",
	}
	resp := &datasource.MetadataResponse{}

	d.Metadata(context.Background(), req, resp)

	if resp.TypeName != "hyperping_healthchecks" {
		t.Errorf("Expected type name 'hyperping_healthchecks', got '%s'", resp.TypeName)
	}
}

func TestHealthchecksDataSource_Schema(t *testing.T) {
	d := &HealthchecksDataSource{}

	req := datasource.SchemaRequest{}
	resp := &datasource.SchemaResponse{}

	d.Schema(context.Background(), req, resp)

	if _, ok := resp.Schema.Attributes["healthchecks"]; !ok {
		t.Error("Schema missing 'healthchecks' attribute")
	}
}

func TestHealthchecksDataSource_Configure(t *testing.T) {
	t.Run("valid client", func(t *testing.T) {
		d := &HealthchecksDataSource{}
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
		d := &HealthchecksDataSource{}

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
		d := &HealthchecksDataSource{}

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

func TestHealthchecksDataSource_mapHealthcheckToDataModel(t *testing.T) {
	d := &HealthchecksDataSource{}

	t.Run("all fields populated", func(t *testing.T) {
		periodVal := 120
		hc := &client.Healthcheck{
			UUID:             "hc-full",
			Name:             "Full Healthcheck",
			PingURL:          "https://ping.hyperping.io/full",
			Cron:             "*/15 * * * *",
			Timezone:         "America/Los_Angeles",
			PeriodValue:      &periodVal,
			PeriodType:       "minutes",
			GracePeriodValue: 15,
			GracePeriodType:  "minutes",
			EscalationPolicy: &client.EscalationPolicyReference{
				UUID: "ep-abc",
				Name: "Test Policy",
			},
			IsPaused:    true,
			IsDown:      false,
			Period:      7200,
			GracePeriod: 900,
			LastPing:    "2026-01-20T15:30:00Z",
			CreatedAt:   "2026-01-01T12:00:00Z",
		}

		model := &HealthcheckDataModel{}
		d.mapHealthcheckToDataModel(hc, model)

		if model.ID.ValueString() != "hc-full" {
			t.Errorf("Expected ID 'hc-full', got %s", model.ID.ValueString())
		}
		if model.Name.ValueString() != "Full Healthcheck" {
			t.Errorf("Expected name 'Full Healthcheck', got %s", model.Name.ValueString())
		}
		if model.Cron.ValueString() != "*/15 * * * *" {
			t.Errorf("Expected cron '*/15 * * * *', got %s", model.Cron.ValueString())
		}
		if model.Period.ValueInt64() != 7200 {
			t.Errorf("Expected period 7200, got %d", model.Period.ValueInt64())
		}
	})

	t.Run("minimal fields", func(t *testing.T) {
		hc := &client.Healthcheck{
			UUID:             "hc-min",
			Name:             "Minimal",
			PingURL:          "https://ping.hyperping.io/min",
			GracePeriodValue: 1,
			GracePeriodType:  "seconds",
			GracePeriod:      1,
		}

		model := &HealthcheckDataModel{}
		d.mapHealthcheckToDataModel(hc, model)

		if model.ID.ValueString() != "hc-min" {
			t.Errorf("Expected ID 'hc-min', got %s", model.ID.ValueString())
		}
		if !model.Cron.IsNull() {
			t.Error("Expected Cron to be null for empty value")
		}
		if !model.LastPing.IsNull() {
			t.Error("Expected LastPing to be null for empty value")
		}
	})
}
