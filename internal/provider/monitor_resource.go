// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/develeap/terraform-provider-hyperping/internal/client"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                = &MonitorResource{}
	_ resource.ResourceWithImportState = &MonitorResource{}
)

// NewMonitorResource creates a new monitor resource.
func NewMonitorResource() resource.Resource {
	return &MonitorResource{}
}

// MonitorResource defines the resource implementation.
type MonitorResource struct {
	client client.MonitorAPI
}

// MonitorResourceModel describes the resource data model.
type MonitorResourceModel struct {
	ID                 types.String `tfsdk:"id"`
	Name               types.String `tfsdk:"name"`
	URL                types.String `tfsdk:"url"`
	Protocol           types.String `tfsdk:"protocol"`
	HTTPMethod         types.String `tfsdk:"http_method"`
	CheckFrequency     types.Int64  `tfsdk:"check_frequency"`
	Regions            types.List   `tfsdk:"regions"`
	RequestHeaders     types.List   `tfsdk:"request_headers"`
	RequestBody        types.String `tfsdk:"request_body"`
	ExpectedStatusCode types.String `tfsdk:"expected_status_code"`
	FollowRedirects    types.Bool   `tfsdk:"follow_redirects"`
	Paused             types.Bool   `tfsdk:"paused"`
	Port               types.Int64  `tfsdk:"port"`
	AlertsWait         types.Int64  `tfsdk:"alerts_wait"`
	EscalationPolicy   types.String `tfsdk:"escalation_policy"`
	RequiredKeyword    types.String `tfsdk:"required_keyword"`
}

// Metadata returns the resource type name.
func (r *MonitorResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_monitor"
}

// Schema defines the schema for the resource.
func (r *MonitorResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Hyperping monitor for uptime monitoring.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The unique identifier (UUID) of the monitor.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the monitor.",
				Required:            true,
			},
			"url": schema.StringAttribute{
				MarkdownDescription: "The URL to monitor.",
				Required:            true,
			},
			"protocol": schema.StringAttribute{
				MarkdownDescription: "The protocol type. Valid values: `http`, `port`, `icmp`. Defaults to `http`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("http"),
				Validators: []validator.String{
					stringvalidator.OneOf(client.AllowedProtocols...),
				},
			},
			"http_method": schema.StringAttribute{
				MarkdownDescription: "HTTP method to use. Valid values: `GET`, `POST`, `PUT`, `PATCH`, `DELETE`, `HEAD`. Defaults to `GET`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("GET"),
				Validators: []validator.String{
					stringvalidator.OneOf(client.AllowedMethods...),
				},
			},
			"check_frequency": schema.Int64Attribute{
				MarkdownDescription: "Check frequency in seconds. Valid values: `10`, `20`, `30`, `60`, `120`, `180`, `300`, `600`, `1800`, `3600`, `21600`, `43200`, `86400`. Defaults to `60`.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(client.DefaultMonitorFrequency),
				Validators: []validator.Int64{
					int64validator.OneOf(10, 20, 30, 60, 120, 180, 300, 600, 1800, 3600, 21600, 43200, 86400),
				},
			},
			"regions": schema.ListAttribute{
				MarkdownDescription: "List of regions to check from. Valid values: `london`, `frankfurt`, `singapore`, `sydney`, `tokyo`, `virginia`, `saopaulo`, `bahrain`.",
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				Validators: []validator.List{
					listvalidator.ValueStringsAre(stringvalidator.OneOf(client.AllowedRegions...)),
				},
			},
			"request_headers": schema.ListNestedAttribute{
				MarkdownDescription: "Custom HTTP headers to send with the request.",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							MarkdownDescription: "The header name. Reserved headers (Authorization, Host, Cookie, etc.) are not allowed.",
							Required:            true,
							Validators: []validator.String{
								NoControlCharacters("Header name must not contain CR, LF, or NULL characters to prevent HTTP header injection."),
								ReservedHeaderName(),
							},
						},
						"value": schema.StringAttribute{
							MarkdownDescription: "The header value.",
							Required:            true,
							Validators: []validator.String{
								NoControlCharacters("Header value must not contain CR, LF, or NULL characters to prevent HTTP header injection."),
							},
						},
					},
				},
			},
			"request_body": schema.StringAttribute{
				MarkdownDescription: "Request body for POST/PUT/PATCH requests.",
				Optional:            true,
			},
			"expected_status_code": schema.StringAttribute{
				MarkdownDescription: "Expected HTTP status code pattern. Use `2xx` for any 2xx status, or specific like `200`, `201`. Defaults to `2xx`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("2xx"),
			},
			"follow_redirects": schema.BoolAttribute{
				MarkdownDescription: "Whether to follow HTTP redirects. Defaults to `true`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"paused": schema.BoolAttribute{
				MarkdownDescription: "Whether the monitor is paused. Defaults to `false`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"port": schema.Int64Attribute{
				MarkdownDescription: "Port number to check. Required when `protocol` is `port`.",
				Optional:            true,
			},
			"alerts_wait": schema.Int64Attribute{
				MarkdownDescription: "Seconds to wait before sending alerts after an outage is detected. Allows time for transient issues to resolve.",
				Optional:            true,
			},
			"escalation_policy": schema.StringAttribute{
				MarkdownDescription: "UUID of the escalation policy to link to this monitor.",
				Optional:            true,
			},
			"required_keyword": schema.StringAttribute{
				MarkdownDescription: "A keyword that must appear in the response body for the check to pass.",
				Optional:            true,
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *MonitorResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.Append(newUnexpectedConfigTypeError("*client.Client", req.ProviderData))
		return
	}

	r.client = c
}

// Create creates the resource and sets the initial Terraform state.
func (r *MonitorResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan MonitorResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build create request from plan
	createReq := r.buildCreateRequest(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save desired paused state (create API doesn't support paused field)
	wantPaused := !plan.Paused.IsNull() && plan.Paused.ValueBool()

	// Call API to create monitor
	createResp, err := r.client.CreateMonitor(ctx, createReq)
	if err != nil {
		resp.Diagnostics.Append(newCreateError("Monitor", err))
		return
	}

	// Read full monitor details (create response may be incomplete)
	monitor, err := r.client.GetMonitor(ctx, createResp.UUID)
	if err != nil {
		resp.Diagnostics.Append(newReadAfterCreateError("Monitor", createResp.UUID, err))
		return
	}

	// Map API response to Terraform state
	r.mapMonitorToModel(monitor, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Handle pause state via separate API call if needed
	if wantPaused {
		r.handlePostCreatePause(ctx, monitor.UUID, &plan, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *MonitorResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state MonitorResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	monitor, err := r.client.GetMonitor(ctx, state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.Append(newReadError("Monitor", state.ID.ValueString(), err))
		return
	}

	r.mapMonitorToModel(monitor, &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update updates the resource and sets the updated Terraform state.
func (r *MonitorResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan MonitorResourceModel
	var state MonitorResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build update request with only changed fields
	updateReq := r.buildUpdateRequest(ctx, &plan, &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Call API to update monitor
	monitor, err := r.client.UpdateMonitor(ctx, state.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.Append(newUpdateError("Monitor", state.ID.ValueString(), err))
		return
	}

	// Map API response to Terraform state
	r.mapMonitorToModel(monitor, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete deletes the resource.
func (r *MonitorResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state MonitorResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteMonitor(ctx, state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			// Already deleted, no error
			return
		}
		resp.Diagnostics.Append(newDeleteError("Monitor", state.ID.ValueString(), err))
		return
	}
}

// ImportState imports an existing resource into Terraform.
func (r *MonitorResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Validate the import ID before setting state (VULN-015)
	if err := client.ValidateResourceID(req.ID); err != nil {
		resp.Diagnostics.Append(newImportError("Monitor", err))
		return
	}
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// mapMonitorToModel maps a client.Monitor to the Terraform model.
func (r *MonitorResource) mapMonitorToModel(monitor *client.Monitor, model *MonitorResourceModel, diags *diag.Diagnostics) {
	model.ID = types.StringValue(monitor.UUID)
	model.Name = types.StringValue(monitor.Name)
	model.URL = types.StringValue(monitor.URL)
	model.Protocol = types.StringValue(monitor.Protocol)
	model.HTTPMethod = types.StringValue(monitor.HTTPMethod)
	model.CheckFrequency = types.Int64Value(int64(monitor.CheckFrequency))
	model.ExpectedStatusCode = types.StringValue(string(monitor.ExpectedStatusCode))
	model.FollowRedirects = types.BoolValue(monitor.FollowRedirects)
	model.Paused = types.BoolValue(monitor.Paused)

	// Handle regions
	model.Regions = mapStringSliceToList(monitor.Regions, diags)

	// Handle request headers
	if len(monitor.RequestHeaders) == 0 {
		model.RequestHeaders = types.ListNull(types.ObjectType{AttrTypes: RequestHeaderAttrTypes()})
	} else {
		values := make([]attr.Value, len(monitor.RequestHeaders))
		for i, h := range monitor.RequestHeaders {
			obj, objDiags := types.ObjectValue(RequestHeaderAttrTypes(), map[string]attr.Value{
				"name":  types.StringValue(h.Name),
				"value": types.StringValue(h.Value),
			})
			diags.Append(objDiags...)
			values[i] = obj
		}
		list, listDiags := types.ListValue(types.ObjectType{AttrTypes: RequestHeaderAttrTypes()}, values)
		diags.Append(listDiags...)
		model.RequestHeaders = list
	}

	// Handle request body
	if monitor.RequestBody != "" {
		model.RequestBody = types.StringValue(monitor.RequestBody)
	} else {
		model.RequestBody = types.StringNull()
	}

	// Handle port
	if monitor.Port != nil {
		model.Port = types.Int64Value(int64(*monitor.Port))
	} else {
		model.Port = types.Int64Null()
	}

	// Handle alerts_wait
	if monitor.AlertsWait > 0 {
		model.AlertsWait = types.Int64Value(int64(monitor.AlertsWait))
	} else {
		model.AlertsWait = types.Int64Null()
	}

	// Handle escalation_policy
	if monitor.EscalationPolicy != nil && *monitor.EscalationPolicy != "" {
		model.EscalationPolicy = types.StringValue(*monitor.EscalationPolicy)
	} else {
		model.EscalationPolicy = types.StringNull()
	}

	// Handle required_keyword
	if monitor.RequiredKeyword != nil && *monitor.RequiredKeyword != "" {
		model.RequiredKeyword = types.StringValue(*monitor.RequiredKeyword)
	} else {
		model.RequiredKeyword = types.StringNull()
	}
}

// buildCreateRequest constructs a CreateMonitorRequest from the Terraform plan.
// Extracts all required and optional fields from the plan model.
func (r *MonitorResource) buildCreateRequest(ctx context.Context, plan *MonitorResourceModel, diags *diag.Diagnostics) client.CreateMonitorRequest {
	createReq := client.CreateMonitorRequest{
		Name:               plan.Name.ValueString(),
		URL:                plan.URL.ValueString(),
		Protocol:           plan.Protocol.ValueString(),
		HTTPMethod:         plan.HTTPMethod.ValueString(),
		CheckFrequency:     int(plan.CheckFrequency.ValueInt64()),
		ExpectedStatusCode: plan.ExpectedStatusCode.ValueString(),
	}

	// Handle optional follow_redirects
	createReq.FollowRedirects = tfBoolToPtr(plan.FollowRedirects)

	// Handle optional regions
	if !isNullOrUnknown(plan.Regions) {
		var regions []string
		diags.Append(plan.Regions.ElementsAs(ctx, &regions, false)...)
		if !diags.HasError() {
			createReq.Regions = regions
		}
	}

	// Handle optional request headers
	if !isNullOrUnknown(plan.RequestHeaders) {
		createReq.RequestHeaders = mapTFListToRequestHeaders(plan.RequestHeaders, diags)
	}

	// Handle optional request body
	createReq.RequestBody = tfStringToPtr(plan.RequestBody)

	// Handle optional port
	createReq.Port = tfIntToPtr(plan.Port)

	// Handle optional alerts_wait
	createReq.AlertsWait = tfIntToPtr(plan.AlertsWait)

	// Handle optional escalation_policy
	createReq.EscalationPolicy = tfStringToPtr(plan.EscalationPolicy)

	// Handle optional required_keyword
	createReq.RequiredKeyword = tfStringToPtr(plan.RequiredKeyword)

	return createReq
}

// handlePostCreatePause pauses a newly created monitor if requested.
// The create API doesn't support the paused field, so this requires a separate API call.
func (r *MonitorResource) handlePostCreatePause(ctx context.Context, monitorID string, plan *MonitorResourceModel, diags *diag.Diagnostics) {
	_, err := r.client.PauseMonitor(ctx, monitorID)
	if err != nil {
		diags.Append(newUpdateError("Monitor", monitorID, fmt.Errorf("monitor created but failed to pause: %w", err)))
		return
	}
	plan.Paused = types.BoolValue(true)
}

// buildUpdateRequest constructs an UpdateMonitorRequest with only changed fields.
// Compares plan vs state and populates request with differences.
func (r *MonitorResource) buildUpdateRequest(ctx context.Context, plan *MonitorResourceModel, state *MonitorResourceModel, diags *diag.Diagnostics) client.UpdateMonitorRequest {
	updateReq := client.UpdateMonitorRequest{}

	// Handle simple string and numeric fields
	r.applySimpleFieldChanges(plan, state, &updateReq)

	// Handle complex fields (regions, headers, etc.)
	r.applyComplexFieldChanges(ctx, plan, state, &updateReq, diags)

	return updateReq
}

// applySimpleFieldChanges detects and applies changes for simple scalar fields.
// Includes: name, url, protocol, http_method, check_frequency, expected_status_code, follow_redirects, paused.
func (r *MonitorResource) applySimpleFieldChanges(plan *MonitorResourceModel, state *MonitorResourceModel, updateReq *client.UpdateMonitorRequest) {
	if !plan.Name.Equal(state.Name) {
		updateReq.Name = tfStringToPtr(plan.Name)
	}

	if !plan.URL.Equal(state.URL) {
		updateReq.URL = tfStringToPtr(plan.URL)
	}

	if !plan.Protocol.Equal(state.Protocol) {
		updateReq.Protocol = tfStringToPtr(plan.Protocol)
	}

	if !plan.HTTPMethod.Equal(state.HTTPMethod) {
		updateReq.HTTPMethod = tfStringToPtr(plan.HTTPMethod)
	}

	if !plan.CheckFrequency.Equal(state.CheckFrequency) {
		updateReq.CheckFrequency = tfIntToPtr(plan.CheckFrequency)
	}

	if !plan.ExpectedStatusCode.Equal(state.ExpectedStatusCode) {
		updateReq.ExpectedStatusCode = tfStringToPtr(plan.ExpectedStatusCode)
	}

	if !plan.FollowRedirects.Equal(state.FollowRedirects) {
		updateReq.FollowRedirects = tfBoolToPtr(plan.FollowRedirects)
	}

	if !plan.Paused.Equal(state.Paused) {
		updateReq.Paused = tfBoolToPtr(plan.Paused)
	}
}

// applyComplexFieldChanges detects and applies changes for complex fields.
// Handles: regions, request_headers, request_body, port, alerts_wait, escalation_policy, required_keyword.
func (r *MonitorResource) applyComplexFieldChanges(ctx context.Context, plan *MonitorResourceModel, state *MonitorResourceModel, updateReq *client.UpdateMonitorRequest, diags *diag.Diagnostics) {
	// Handle regions (skip if unknown)
	if !plan.Regions.IsUnknown() && !plan.Regions.Equal(state.Regions) {
		if plan.Regions.IsNull() {
			emptyRegions := []string{}
			updateReq.Regions = &emptyRegions
		} else {
			var regions []string
			diags.Append(plan.Regions.ElementsAs(ctx, &regions, false)...)
			if !diags.HasError() {
				updateReq.Regions = &regions
			}
		}
	}

	// Handle request headers (skip if unknown)
	if !plan.RequestHeaders.IsUnknown() && !plan.RequestHeaders.Equal(state.RequestHeaders) {
		if plan.RequestHeaders.IsNull() {
			emptyHeaders := []client.RequestHeader{}
			updateReq.RequestHeaders = &emptyHeaders
		} else {
			headers := mapTFListToRequestHeaders(plan.RequestHeaders, diags)
			if !diags.HasError() {
				updateReq.RequestHeaders = &headers
			}
		}
	}

	// Handle request body
	if !plan.RequestBody.Equal(state.RequestBody) {
		if plan.RequestBody.IsNull() {
			empty := ""
			updateReq.RequestBody = &empty
		} else {
			updateReq.RequestBody = tfStringToPtr(plan.RequestBody)
		}
	}

	// Handle port
	if !plan.Port.Equal(state.Port) {
		updateReq.Port = tfIntToPtr(plan.Port)
	}

	// Handle alerts_wait
	if !plan.AlertsWait.Equal(state.AlertsWait) {
		if plan.AlertsWait.IsNull() {
			zero := 0
			updateReq.AlertsWait = &zero
		} else {
			updateReq.AlertsWait = tfIntToPtr(plan.AlertsWait)
		}
	}

	// Handle escalation_policy
	if !plan.EscalationPolicy.Equal(state.EscalationPolicy) {
		if plan.EscalationPolicy.IsNull() {
			empty := ""
			updateReq.EscalationPolicy = &empty
		} else {
			updateReq.EscalationPolicy = tfStringToPtr(plan.EscalationPolicy)
		}
	}

	// Handle required_keyword
	if !plan.RequiredKeyword.Equal(state.RequiredKeyword) {
		if plan.RequiredKeyword.IsNull() {
			empty := ""
			updateReq.RequiredKeyword = &empty
		} else {
			updateReq.RequiredKeyword = tfStringToPtr(plan.RequiredKeyword)
		}
	}
}
