// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
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

	hyperping "github.com/develeap/hyperping-go"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                   = &MonitorResource{}
	_ resource.ResourceWithImportState    = &MonitorResource{}
	_ resource.ResourceWithValidateConfig = &MonitorResource{}
)

// NewMonitorResource creates a new monitor resource.
func NewMonitorResource() resource.Resource {
	return &MonitorResource{}
}

// MonitorResource defines the resource implementation.
type MonitorResource struct {
	client hyperping.MonitorAPI
}

// MonitorResourceModel describes the resource data model.
type MonitorResourceModel struct {
	ID                   types.String `tfsdk:"id"`
	Name                 types.String `tfsdk:"name"`
	URL                  types.String `tfsdk:"url"`
	Protocol             types.String `tfsdk:"protocol"`
	HTTPMethod           types.String `tfsdk:"http_method"`
	CheckFrequency       types.Int64  `tfsdk:"check_frequency"`
	Regions              types.List   `tfsdk:"regions"`
	RequestHeaders       types.List   `tfsdk:"request_headers"`
	RequestBody          types.String `tfsdk:"request_body"`
	ExpectedStatusCode   types.String `tfsdk:"expected_status_code"`
	FollowRedirects      types.Bool   `tfsdk:"follow_redirects"`
	Paused               types.Bool   `tfsdk:"paused"`
	Port                 types.Int64  `tfsdk:"port"`
	AlertsWait           types.Int64  `tfsdk:"alerts_wait"`
	EscalationPolicy     types.String `tfsdk:"escalation_policy"`
	EscalationPolicyName types.String `tfsdk:"escalation_policy_name"`
	DNSRecordType        types.String `tfsdk:"dns_record_type"`
	DNSNameserver        types.String `tfsdk:"dns_nameserver"`
	DNSExpectedAnswer    types.String `tfsdk:"dns_expected_answer"`
	RequiredKeyword      types.String `tfsdk:"required_keyword"`
	Status               types.String `tfsdk:"status"`
	IsDown               types.Bool   `tfsdk:"is_down"`
	SSLExpiration        types.Int64  `tfsdk:"ssl_expiration"`
	ProjectUUID          types.String `tfsdk:"project_uuid"`
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
				MarkdownDescription: "The display name of the monitor. Must be 1-255 characters.",
				Required:            true,
				Validators: []validator.String{
					StringLength(1, 255),
				},
			},
			"url": schema.StringAttribute{
				MarkdownDescription: "The URL to monitor. For HTTP/ICMP/port protocols, must include scheme " +
					"(e.g., `https://api.example.com/health`). For DNS protocol, use a bare domain (e.g., `example.com`).",
				Required: true,
				Validators: []validator.String{
					StringLength(1, 2048),
				},
			},
			"protocol": schema.StringAttribute{
				MarkdownDescription: "The protocol type. Valid values: `http`, `port`, `icmp`, `dns`. Defaults to `http`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("http"),
				Validators: []validator.String{
					stringvalidator.OneOf(hyperping.AllowedProtocols...),
				},
			},
			"http_method": schema.StringAttribute{
				MarkdownDescription: "HTTP method to use. Only valid when protocol is `http`. Valid values: `GET`, `POST`, `PUT`, `PATCH`, `DELETE`, `HEAD`, `OPTIONS`. Defaults to `GET`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("GET"),
				Validators: []validator.String{
					stringvalidator.OneOf(hyperping.AllowedMethods...),
				},
			},
			"check_frequency": schema.Int64Attribute{
				MarkdownDescription: "Check frequency in seconds. Valid values: `10`, `20`, `30`, `60`, `120`, `180`, `300`, `600`, `1800`, `3600`, `21600`, `43200`, `86400`. Defaults to `60`.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(hyperping.DefaultMonitorFrequency),
				Validators: []validator.Int64{
					int64validator.OneOf(10, 20, 30, 60, 120, 180, 300, 600, 1800, 3600, 21600, 43200, 86400),
				},
			},
			"regions": schema.ListAttribute{
				MarkdownDescription: "List of monitoring regions. Use the `hyperping_monitoring_locations` data source to discover available locations.",
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				Validators: []validator.List{
					listvalidator.ValueStringsAre(stringvalidator.OneOf(hyperping.AllowedRegions...)),
				},
			},
			"request_headers": schema.ListNestedAttribute{
				MarkdownDescription: "Custom HTTP headers to send with the request. Only valid when protocol is `http`.",
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
				MarkdownDescription: "HTTP request body. Only valid when protocol is `http` and http_method is `POST`, `PUT`, or `PATCH`.",
				Optional:            true,
			},
			"expected_status_code": schema.StringAttribute{
				MarkdownDescription: "Expected HTTP status code pattern. " +
					"Use a specific code like `200`, a wildcard like `2xx` (200-299), " +
					"or a range like `1xx-3xx` (100-399). Defaults to `2xx`.",
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("2xx"),
				Validators: []validator.String{
					StatusCodePattern(),
				},
			},
			"follow_redirects": schema.BoolAttribute{
				MarkdownDescription: "Whether to follow HTTP redirects. Only applies to `http` protocol monitors. Defaults to `true`.",
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
				MarkdownDescription: "TCP port number (1-65535). Required when protocol is `port`. Examples: `443` (HTTPS), `5432` (PostgreSQL), `6379` (Redis).",
				Optional:            true,
				Validators: []validator.Int64{
					PortRange(),
				},
			},
			"alerts_wait": schema.Int64Attribute{
				MarkdownDescription: "Minutes to wait before sending alerts after an outage is detected. " +
					"Must be one of: `-1` (disabled), `0`, `1`, `2`, `3`, `5`, `10`, `30`, `60`.",
				Optional: true,
				Validators: []validator.Int64{
					AlertsWait(),
				},
			},
			"escalation_policy": schema.StringAttribute{
				MarkdownDescription: "UUID of the escalation policy to link to this monitor.",
				Optional:            true,
				Validators: []validator.String{
					UUIDFormat(),
				},
			},
			"escalation_policy_name": schema.StringAttribute{
				MarkdownDescription: "Human-readable name of the assigned escalation policy.",
				Computed:            true,
			},
			"dns_record_type": schema.StringAttribute{
				MarkdownDescription: "DNS record type to check. Only valid when protocol is `dns`. " +
					"Valid values: `A`, `AAAA`, `CNAME`, `MX`, `NS`, `TXT`, `SOA`, `SRV`, `CAA`, `PTR`. " +
					"Defaults to `A` (set by the API if omitted).",
				Optional: true,
				Computed: true,
				Validators: []validator.String{
					stringvalidator.OneOf(hyperping.AllowedDNSRecordTypes...),
				},
			},
			"dns_nameserver": schema.StringAttribute{
				MarkdownDescription: "Nameserver to query against (e.g., `8.8.8.8`). " +
					"Only valid when protocol is `dns`. Leave empty to use default resolvers.",
				Optional: true,
			},
			"dns_expected_answer": schema.StringAttribute{
				MarkdownDescription: "Expected DNS answer to validate against. " +
					"Only valid when protocol is `dns`. Monitor fails if the resolved value does not contain this string.",
				Optional: true,
			},
			"required_keyword": schema.StringAttribute{
				MarkdownDescription: "A keyword that must appear in the HTTP response body for the check to pass. Only valid when protocol is `http`.",
				Optional:            true,
			},
			"status": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Current monitor status. Either `up` or `down`.",
			},
			"is_down": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether the monitor is currently reporting as down.",
			},
			"ssl_expiration": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Days until the SSL certificate expires.",
			},
			"project_uuid": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "UUID of the Hyperping project this monitor belongs to.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *MonitorResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	c, ok := req.ProviderData.(*hyperping.Client)
	if !ok {
		resp.Diagnostics.Append(newUnexpectedConfigTypeError("*hyperping.Client", req.ProviderData))
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
		resp.Diagnostics.Append(NewCreateErrorWithContext("Monitor", err))
		return
	}

	// Write the ID to state immediately to prevent orphaned resources if read-back fails.
	plan.ID = types.StringValue(createResp.UUID)

	// Read full monitor details (create response may be incomplete)
	monitor, err := r.client.GetMonitor(ctx, createResp.UUID)
	if err != nil {
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
		resp.Diagnostics.Append(newReadAfterCreateError("Monitor", createResp.UUID, err))
		return
	}

	// Save write-only fields before mapping (API doesn't return these)
	saved := saveHTTPFields(&plan)
	planRequiredKeyword := plan.RequiredKeyword

	// Map API response to Terraform state
	r.mapMonitorToModel(monitor, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	restoreHTTPFieldsForNonHTTP(monitor.Protocol, &plan, saved)

	// Restore required_keyword: API accepts on write but doesn't return on GET
	if !planRequiredKeyword.IsNull() && plan.RequiredKeyword.IsNull() {
		plan.RequiredKeyword = planRequiredKeyword
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
		if hyperping.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.Append(NewReadErrorWithContext("Monitor", state.ID.ValueString(), err))
		return
	}

	// Save write-only fields before mapping (API doesn't return these)
	saved := saveHTTPFields(&state)
	priorRequiredKeyword := state.RequiredKeyword

	r.mapMonitorToModel(monitor, &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	restoreHTTPFieldsForNonHTTP(monitor.Protocol, &state, saved)

	// Restore required_keyword: API accepts on write but doesn't return on GET
	if !priorRequiredKeyword.IsNull() && state.RequiredKeyword.IsNull() {
		state.RequiredKeyword = priorRequiredKeyword
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
		resp.Diagnostics.Append(NewUpdateErrorWithContext("Monitor", state.ID.ValueString(), err))
		return
	}

	// Save write-only fields before mapping (API doesn't return these)
	saved := saveHTTPFields(&plan)
	planRequiredKeyword := plan.RequiredKeyword

	// Map API response to Terraform state
	r.mapMonitorToModel(monitor, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	restoreHTTPFieldsForNonHTTP(monitor.Protocol, &plan, saved)

	// Restore required_keyword: API accepts on write but doesn't return on GET
	if !planRequiredKeyword.IsNull() && plan.RequiredKeyword.IsNull() {
		plan.RequiredKeyword = planRequiredKeyword
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
		if hyperping.IsNotFound(err) {
			// Already deleted, no error
			return
		}
		resp.Diagnostics.Append(NewDeleteErrorWithContext("Monitor", state.ID.ValueString(), err))
		return
	}
}

// ImportState imports an existing resource into Terraform.
func (r *MonitorResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Validate the import ID before setting state (VULN-015)
	if err := hyperping.ValidateResourceID(req.ID); err != nil {
		resp.Diagnostics.Append(newImportError("Monitor", err))
		return
	}
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// savedHTTPFields holds HTTP-specific field values saved before mapMonitorToModel
// so they can be restored for non-HTTP monitors (ISS-ICMP-002). The API returns
// empty/null for these fields on non-HTTP monitors, but Terraform needs them
// preserved to avoid unnecessary diffs.
type savedHTTPFields struct {
	httpMethod         types.String
	expectedStatusCode types.String
	followRedirects    types.Bool
}

// saveHTTPFields captures the HTTP-specific fields from a model before mapping.
func saveHTTPFields(model *MonitorResourceModel) savedHTTPFields {
	return savedHTTPFields{
		httpMethod:         model.HTTPMethod,
		expectedStatusCode: model.ExpectedStatusCode,
		followRedirects:    model.FollowRedirects,
	}
}

// restoreHTTPFieldsForNonHTTP restores saved HTTP-specific fields when the
// monitor protocol is not "http". The API returns empty/null values for these
// fields on non-HTTP monitors, so we restore the plan/state values to prevent
// spurious Terraform diffs.
func restoreHTTPFieldsForNonHTTP(protocol string, model *MonitorResourceModel, saved savedHTTPFields) {
	if protocol == "http" {
		return
	}
	if !saved.httpMethod.IsNull() {
		model.HTTPMethod = saved.httpMethod
	}
	if !saved.expectedStatusCode.IsNull() {
		model.ExpectedStatusCode = saved.expectedStatusCode
	}
	if !saved.followRedirects.IsNull() {
		model.FollowRedirects = saved.followRedirects
	}
}

// mapMonitorToModel maps a hyperping.Monitor to the Terraform model.
// Delegates to the shared MapMonitorCommonFields to avoid duplication with data sources.
//
// The field-by-field copy is intentional: MonitorResourceModel embeds additional
// resource-only fields (e.g. Timeouts) that MonitorCommonFields doesn't have,
// so a direct struct assignment isn't possible.
func (r *MonitorResource) mapMonitorToModel(monitor *hyperping.Monitor, model *MonitorResourceModel, diags *diag.Diagnostics) {
	common := MapMonitorCommonFields(monitor, diags)
	model.ID = common.ID
	model.Name = common.Name
	model.URL = common.URL
	model.Protocol = common.Protocol
	model.HTTPMethod = common.HTTPMethod
	model.CheckFrequency = common.CheckFrequency
	model.Regions = common.Regions
	model.RequestHeaders = common.RequestHeaders
	model.RequestBody = common.RequestBody
	model.ExpectedStatusCode = common.ExpectedStatusCode
	model.FollowRedirects = common.FollowRedirects
	model.Paused = common.Paused
	model.Port = common.Port
	model.AlertsWait = common.AlertsWait
	model.EscalationPolicy = common.EscalationPolicy
	model.EscalationPolicyName = common.EscalationPolicyName
	model.DNSRecordType = common.DNSRecordType
	model.DNSNameserver = common.DNSNameserver
	model.DNSExpectedAnswer = common.DNSExpectedAnswer
	model.RequiredKeyword = common.RequiredKeyword
	model.Status = common.Status
	model.IsDown = common.IsDown
	model.SSLExpiration = common.SSLExpiration
	model.ProjectUUID = common.ProjectUUID
}

// buildCreateRequest constructs a CreateMonitorRequest from the Terraform plan.
// Extracts all required and optional fields from the plan model.
func (r *MonitorResource) buildCreateRequest(ctx context.Context, plan *MonitorResourceModel, diags *diag.Diagnostics) hyperping.CreateMonitorRequest {
	createReq := hyperping.CreateMonitorRequest{
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

	// Handle optional DNS fields
	createReq.DNSRecordType = tfStringToPtr(plan.DNSRecordType)
	createReq.DNSNameserver = tfStringToPtr(plan.DNSNameserver)
	createReq.DNSExpectedAnswer = tfStringToPtr(plan.DNSExpectedAnswer)

	// Handle optional project_uuid
	createReq.ProjectUUID = plan.ProjectUUID.ValueString()

	return createReq
}

// handlePostCreatePause pauses a newly created monitor if requested.
// The create API doesn't support the paused field, so this requires a separate API call.
func (r *MonitorResource) handlePostCreatePause(ctx context.Context, monitorID string, plan *MonitorResourceModel, diags *diag.Diagnostics) {
	_, err := r.client.PauseMonitor(ctx, monitorID)
	if err != nil {
		diags.Append(NewUpdateErrorWithContext("Monitor", monitorID, fmt.Errorf("monitor created but failed to pause: %w", err)))
		return
	}
	plan.Paused = types.BoolValue(true)
}

// buildUpdateRequest constructs an UpdateMonitorRequest with only changed fields.
// Compares plan vs state and populates request with differences.
func (r *MonitorResource) buildUpdateRequest(ctx context.Context, plan *MonitorResourceModel, state *MonitorResourceModel, diags *diag.Diagnostics) hyperping.UpdateMonitorRequest {
	updateReq := hyperping.UpdateMonitorRequest{}

	// Handle simple string and numeric fields
	r.applySimpleFieldChanges(plan, state, &updateReq)

	// Handle complex fields (regions, headers, etc.)
	r.applyComplexFieldChanges(ctx, plan, state, &updateReq, diags)

	return updateReq
}

// applySimpleFieldChanges detects and applies changes for simple scalar fields.
// Includes: name, url, protocol, http_method, check_frequency, expected_status_code, follow_redirects, paused.
func (r *MonitorResource) applySimpleFieldChanges(plan *MonitorResourceModel, state *MonitorResourceModel, updateReq *hyperping.UpdateMonitorRequest) {
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

	if !plan.ProjectUUID.Equal(state.ProjectUUID) {
		updateReq.ProjectUUID = tfStringToPtr(plan.ProjectUUID)
	}
}

// applyHTTPFieldChanges handles HTTP-specific field changes for monitor updates.
// Handles: request_headers, request_body, expected_status_code (via UpdateMonitorRequest).
// Note: http_method, follow_redirects are handled in applySimpleFieldChanges.
func applyHTTPFieldChanges(plan *MonitorResourceModel, state *MonitorResourceModel, updateReq *hyperping.UpdateMonitorRequest, diags *diag.Diagnostics) {
	// Handle request headers (skip if unknown)
	if !plan.RequestHeaders.IsUnknown() && !plan.RequestHeaders.Equal(state.RequestHeaders) {
		if plan.RequestHeaders.IsNull() {
			emptyHeaders := []hyperping.RequestHeader{}
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
}

// applyMonitoringFieldChanges handles monitoring-behavior field changes for monitor updates.
// Handles: regions, alerts_wait, escalation_policy, required_keyword.
func applyMonitoringFieldChanges(ctx context.Context, plan *MonitorResourceModel, state *MonitorResourceModel, updateReq *hyperping.UpdateMonitorRequest, diags *diag.Diagnostics) {
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
	// The Hyperping API uses a magic value "none" as the canonical way to unlink
	// (disassociate) an escalation policy from a monitor. Sending a valid UUID
	// links that policy; sending "none" removes any existing link. The API
	// rejects an empty string "", so "none" is the only supported unlink value.
	if !plan.EscalationPolicy.Equal(state.EscalationPolicy) {
		if plan.EscalationPolicy.IsNull() {
			none := "none"
			updateReq.EscalationPolicy = &none
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

	// Handle port
	if !plan.Port.Equal(state.Port) {
		updateReq.Port = tfIntToPtr(plan.Port)
	}

	// Handle DNS fields
	if !plan.DNSRecordType.Equal(state.DNSRecordType) {
		updateReq.DNSRecordType = tfStringToPtr(plan.DNSRecordType)
	}
	if !plan.DNSNameserver.Equal(state.DNSNameserver) {
		if plan.DNSNameserver.IsNull() {
			empty := ""
			updateReq.DNSNameserver = &empty
		} else {
			updateReq.DNSNameserver = tfStringToPtr(plan.DNSNameserver)
		}
	}
	if !plan.DNSExpectedAnswer.Equal(state.DNSExpectedAnswer) {
		if plan.DNSExpectedAnswer.IsNull() {
			empty := ""
			updateReq.DNSExpectedAnswer = &empty
		} else {
			updateReq.DNSExpectedAnswer = tfStringToPtr(plan.DNSExpectedAnswer)
		}
	}
}

// applyComplexFieldChanges detects and applies changes for complex fields.
// Dispatches to applyHTTPFieldChanges and applyMonitoringFieldChanges.
func (r *MonitorResource) applyComplexFieldChanges(ctx context.Context, plan *MonitorResourceModel, state *MonitorResourceModel, updateReq *hyperping.UpdateMonitorRequest, diags *diag.Diagnostics) {
	applyHTTPFieldChanges(plan, state, updateReq, diags)
	applyMonitoringFieldChanges(ctx, plan, state, updateReq, diags)
}
