// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/develeap/terraform-provider-hyperping/internal/client"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ datasource.DataSource              = &HealthchecksDataSource{}
	_ datasource.DataSourceWithConfigure = &HealthchecksDataSource{}
)

// NewHealthchecksDataSource creates a new healthchecks data source.
func NewHealthchecksDataSource() datasource.DataSource {
	return &HealthchecksDataSource{}
}

// HealthchecksDataSource defines the data source implementation.
type HealthchecksDataSource struct {
	client client.HealthcheckAPI
}

// HealthchecksDataSourceModel describes the data source data model.
type HealthchecksDataSourceModel struct {
	Healthchecks []HealthcheckDataModel `tfsdk:"healthchecks"`
}

// HealthcheckDataModel describes a single healthcheck in the list data source.
type HealthcheckDataModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	PingURL          types.String `tfsdk:"ping_url"`
	Cron             types.String `tfsdk:"cron"`
	Timezone         types.String `tfsdk:"timezone"`
	PeriodValue      types.Int64  `tfsdk:"period_value"`
	PeriodType       types.String `tfsdk:"period_type"`
	GracePeriodValue types.Int64  `tfsdk:"grace_period_value"`
	GracePeriodType  types.String `tfsdk:"grace_period_type"`
	EscalationPolicy types.String `tfsdk:"escalation_policy"`
	IsPaused         types.Bool   `tfsdk:"is_paused"`
	IsDown           types.Bool   `tfsdk:"is_down"`
	Period           types.Int64  `tfsdk:"period"`
	GracePeriod      types.Int64  `tfsdk:"grace_period"`
	LastPing         types.String `tfsdk:"last_ping"`
	CreatedAt        types.String `tfsdk:"created_at"`
}

// Metadata returns the data source type name.
func (d *HealthchecksDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_healthchecks"
}

// Schema defines the schema for the data source.
func (d *HealthchecksDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches the list of all Hyperping healthchecks.",

		Attributes: map[string]schema.Attribute{
			"healthchecks": schema.ListNestedAttribute{
				MarkdownDescription: "List of healthchecks.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "The unique identifier (UUID) of the healthcheck.",
							Computed:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "The name of the healthcheck.",
							Computed:            true,
						},
						"ping_url": schema.StringAttribute{
							MarkdownDescription: "The auto-generated ping URL.",
							Computed:            true,
						},
						"cron": schema.StringAttribute{
							MarkdownDescription: "Cron expression defining the schedule.",
							Computed:            true,
						},
						"timezone": schema.StringAttribute{
							MarkdownDescription: "Timezone for the cron expression.",
							Computed:            true,
						},
						"period_value": schema.Int64Attribute{
							MarkdownDescription: "Numeric value for the expected interval.",
							Computed:            true,
						},
						"period_type": schema.StringAttribute{
							MarkdownDescription: "Unit for period_value.",
							Computed:            true,
						},
						"grace_period_value": schema.Int64Attribute{
							MarkdownDescription: "Numeric value for the grace period buffer.",
							Computed:            true,
						},
						"grace_period_type": schema.StringAttribute{
							MarkdownDescription: "Unit for grace_period_value.",
							Computed:            true,
						},
						"escalation_policy": schema.StringAttribute{
							MarkdownDescription: "UUID of the escalation policy linked to this healthcheck.",
							Computed:            true,
						},
						"is_paused": schema.BoolAttribute{
							MarkdownDescription: "Whether the healthcheck is paused.",
							Computed:            true,
						},
						"is_down": schema.BoolAttribute{
							MarkdownDescription: "Whether the healthcheck is currently in a failure state.",
							Computed:            true,
						},
						"period": schema.Int64Attribute{
							MarkdownDescription: "Calculated period in seconds.",
							Computed:            true,
						},
						"grace_period": schema.Int64Attribute{
							MarkdownDescription: "Calculated grace period in seconds.",
							Computed:            true,
						},
						"last_ping": schema.StringAttribute{
							MarkdownDescription: "Timestamp of the last ping received in ISO 8601 format.",
							Computed:            true,
						},
						"created_at": schema.StringAttribute{
							MarkdownDescription: "Creation timestamp in ISO 8601 format.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *HealthchecksDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = c
}

// Read refreshes the Terraform state with the latest data.
func (d *HealthchecksDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state HealthchecksDataSourceModel

	healthchecks, err := d.client.ListHealthchecks(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading healthchecks",
			fmt.Sprintf("Could not list healthchecks: %s", err),
		)
		return
	}

	state.Healthchecks = make([]HealthcheckDataModel, len(healthchecks))
	for i, hc := range healthchecks {
		d.mapHealthcheckToDataModel(&hc, &state.Healthchecks[i])
		// Currently mapHealthcheckToDataModel doesn't produce errors, but checking
		// provides symmetry with outages data source and future-proofs against changes.
		if resp.Diagnostics.HasError() {
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// mapHealthcheckToDataModel maps a client.Healthcheck to the list data model
// using the shared HealthcheckCommonFields mapping.
func (d *HealthchecksDataSource) mapHealthcheckToDataModel(hc *client.Healthcheck, model *HealthcheckDataModel) {
	f := MapHealthcheckCommonFields(hc)
	model.ID = f.ID
	model.Name = f.Name
	model.PingURL = f.PingURL
	model.Cron = f.Cron
	model.Timezone = f.Timezone
	model.PeriodValue = f.PeriodValue
	model.PeriodType = f.PeriodType
	model.GracePeriodValue = f.GracePeriodValue
	model.GracePeriodType = f.GracePeriodType
	model.EscalationPolicy = f.EscalationPolicy
	model.IsPaused = f.IsPaused
	model.IsDown = f.IsDown
	model.Period = f.Period
	model.GracePeriod = f.GracePeriod
	model.LastPing = f.LastPing
	model.CreatedAt = f.CreatedAt
}
