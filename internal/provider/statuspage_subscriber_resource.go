// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/develeap/terraform-provider-hyperping/internal/client"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &StatusPageSubscriberResource{}
var _ resource.ResourceWithImportState = &StatusPageSubscriberResource{}

func NewStatusPageSubscriberResource() resource.Resource {
	return &StatusPageSubscriberResource{}
}

// StatusPageSubscriberResource defines the resource implementation.
type StatusPageSubscriberResource struct {
	client client.HyperpingAPI
}

// StatusPageSubscriberResourceModel describes the resource data model.
type StatusPageSubscriberResourceModel struct {
	ID              types.Int64  `tfsdk:"id"`
	StatusPageUUID  types.String `tfsdk:"statuspage_uuid"`
	Type            types.String `tfsdk:"type"`
	Email           types.String `tfsdk:"email"`
	Phone           types.String `tfsdk:"phone"`
	TeamsWebhookURL types.String `tfsdk:"teams_webhook_url"`
	Language        types.String `tfsdk:"language"`
	CreatedAt       types.String `tfsdk:"created_at"`
	Value           types.String `tfsdk:"value"`
}

func (r *StatusPageSubscriberResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_statuspage_subscriber"
}

func (r *StatusPageSubscriberResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a subscriber for a Hyperping status page.\n\n" +
			"Subscribers receive notifications about incidents and maintenance. " +
			"Supported types: email, sms, teams. " +
			"Note: Slack subscribers must be added via OAuth flow and cannot be managed via API.\n\n" +
			"All fields are immutable - any change will recreate the subscriber.",

		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				MarkdownDescription: "Subscriber ID (computed)",
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"statuspage_uuid": schema.StringAttribute{
				MarkdownDescription: "UUID of the status page",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Subscriber type: email, sms, or teams (slack not supported via API)",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("email", "sms", "teams"),
					NoSlackSubscriberType(),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"email": schema.StringAttribute{
				MarkdownDescription: "Email address (required when type=email)",
				Optional:            true,
				Validators: []validator.String{
					RequiredWhenValueIs(path.Root("type"), "email", "type"),
					EmailFormat(),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"phone": schema.StringAttribute{
				MarkdownDescription: "Phone number (required when type=sms)",
				Optional:            true,
				Validators: []validator.String{
					RequiredWhenValueIs(path.Root("type"), "sms", "type"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"teams_webhook_url": schema.StringAttribute{
				MarkdownDescription: "Microsoft Teams webhook URL (required when type=teams)",
				Optional:            true,
				Validators: []validator.String{
					RequiredWhenValueIs(path.Root("type"), "teams", "type"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"language": schema.StringAttribute{
				MarkdownDescription: "Preferred language code (default: en)",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("en"),
				Validators: []validator.String{
					stringvalidator.OneOf("en", "fr", "de", "ru", "nl", "es", "it", "pt", "ja", "zh"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "Creation timestamp (computed)",
				Computed:            true,
			},
			"value": schema.StringAttribute{
				MarkdownDescription: "Display value (computed)",
				Computed:            true,
			},
		},
	}
}

func (r *StatusPageSubscriberResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	apiClient, ok := req.ProviderData.(client.HyperpingAPI)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected client.HyperpingAPI, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = apiClient
}

func (r *StatusPageSubscriberResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan StatusPageSubscriberResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate status page UUID
	if err := client.ValidateResourceID(plan.StatusPageUUID.ValueString()); err != nil {
		resp.Diagnostics.AddError(
			"Invalid Status Page UUID",
			fmt.Sprintf("Status page UUID must be valid: %s", err.Error()),
		)
		return
	}

	// Validate conditional requirements
	if err := r.validateConditionalFields(&plan); err != nil {
		resp.Diagnostics.AddError("Validation Error", err.Error())
		return
	}

	// Build add subscriber request
	addReq := r.buildAddSubscriberRequest(&plan)

	// Add subscriber via API
	subscriber, err := r.client.AddSubscriber(ctx, plan.StatusPageUUID.ValueString(), *addReq)
	if err != nil {
		resp.Diagnostics.AddError("Error adding subscriber", err.Error())
		return
	}

	// Map response to state
	r.mapSubscriberToModel(subscriber, &plan)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *StatusPageSubscriberResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state StatusPageSubscriberResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// List subscribers to find this one (API doesn't have GetSubscriber endpoint)
	subscriberID := int(state.ID.ValueInt64())
	paginatedResp, err := r.client.ListSubscribers(ctx, state.StatusPageUUID.ValueString(), nil, nil)
	if err != nil {
		if client.IsNotFound(err) {
			// Status page was deleted
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading subscriber", err.Error())
		return
	}

	// Find the subscriber by ID
	var foundSubscriber *client.StatusPageSubscriber
	for _, sub := range paginatedResp.Subscribers {
		if sub.ID == subscriberID {
			foundSubscriber = &sub
			break
		}
	}

	if foundSubscriber == nil {
		// Subscriber was deleted outside Terraform
		resp.State.RemoveResource(ctx)
		return
	}

	// Map response to state
	r.mapSubscriberToModel(foundSubscriber, &state)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *StatusPageSubscriberResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// All fields are ForceNew, so Update should never be called
	// If it is called, just read the state back (no-op)
	var state StatusPageSubscriberResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *StatusPageSubscriberResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state StatusPageSubscriberResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete subscriber via API
	subscriberID := int(state.ID.ValueInt64())
	err := r.client.DeleteSubscriber(ctx, state.StatusPageUUID.ValueString(), subscriberID)
	if err != nil {
		if !client.IsNotFound(err) {
			resp.Diagnostics.AddError("Error deleting subscriber", err.Error())
			return
		}
		// Already deleted, continue
	}
}

func (r *StatusPageSubscriberResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import ID format: statuspage_uuid:subscriber_id
	parts := strings.Split(req.ID, ":")
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Import ID must be in format 'statuspage_uuid:subscriber_id', got: %s", req.ID),
		)
		return
	}

	statuspageUUID := parts[0]
	subscriberIDStr := parts[1]

	// Validate status page UUID
	if err := client.ValidateResourceID(statuspageUUID); err != nil {
		resp.Diagnostics.AddError(
			"Invalid Status Page UUID",
			fmt.Sprintf("Status page UUID must be valid: %s", err.Error()),
		)
		return
	}

	// Validate subscriber ID
	subscriberID, err := strconv.Atoi(subscriberIDStr)
	if err != nil || subscriberID <= 0 {
		resp.Diagnostics.AddError(
			"Invalid Subscriber ID",
			fmt.Sprintf("Subscriber ID must be a positive integer, got: %s", subscriberIDStr),
		)
		return
	}

	// Set the attributes
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("statuspage_uuid"), statuspageUUID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), int64(subscriberID))...)
}

// validateConditionalFields validates that required fields are present based on type.
func (r *StatusPageSubscriberResource) validateConditionalFields(model *StatusPageSubscriberResourceModel) error {
	subscriberType := model.Type.ValueString()

	switch subscriberType {
	case "email":
		if model.Email.IsNull() || model.Email.ValueString() == "" {
			return fmt.Errorf("email is required when type is 'email'")
		}
	case "sms":
		if model.Phone.IsNull() || model.Phone.ValueString() == "" {
			return fmt.Errorf("phone is required when type is 'sms'")
		}
	case "teams":
		if model.TeamsWebhookURL.IsNull() || model.TeamsWebhookURL.ValueString() == "" {
			return fmt.Errorf("teams_webhook_url is required when type is 'teams'")
		}
	}

	return nil
}

// buildAddSubscriberRequest builds an AddSubscriberRequest from the Terraform plan.
func (r *StatusPageSubscriberResource) buildAddSubscriberRequest(plan *StatusPageSubscriberResourceModel) *client.AddSubscriberRequest {
	req := &client.AddSubscriberRequest{
		Type: plan.Type.ValueString(),
	}

	// Email (optional, required when type=email)
	if !isNullOrUnknown(plan.Email) {
		email := plan.Email.ValueString()
		req.Email = &email
	}

	// Phone (optional, required when type=sms)
	if !isNullOrUnknown(plan.Phone) {
		phone := plan.Phone.ValueString()
		req.Phone = &phone
	}

	// Teams webhook URL (optional, required when type=teams)
	if !isNullOrUnknown(plan.TeamsWebhookURL) {
		url := plan.TeamsWebhookURL.ValueString()
		req.TeamsWebhookURL = &url
	}

	// Language (optional, default: en)
	if !isNullOrUnknown(plan.Language) {
		lang := plan.Language.ValueString()
		req.Language = &lang
	}

	return req
}

// mapSubscriberToModel maps API response to Terraform model.
func (r *StatusPageSubscriberResource) mapSubscriberToModel(sub *client.StatusPageSubscriber, model *StatusPageSubscriberResourceModel) {
	model.ID = types.Int64Value(int64(sub.ID))
	model.Type = types.StringValue(sub.Type)
	model.Value = types.StringValue(sub.Value)
	model.Language = types.StringValue(sub.Language)
	model.CreatedAt = types.StringValue(sub.CreatedAt)

	// Map optional fields
	if sub.Email != nil && *sub.Email != "" {
		model.Email = types.StringValue(*sub.Email)
	} else {
		model.Email = types.StringNull()
	}

	if sub.Phone != nil && *sub.Phone != "" {
		model.Phone = types.StringValue(*sub.Phone)
	} else {
		model.Phone = types.StringNull()
	}

	// Note: API doesn't return teams_webhook_url, so we keep the planned value
	// Slack channel is read-only (can't be created via API)
}

// NoSlackSubscriberType returns a validator that rejects "slack" type.
type noSlackSubscriberTypeValidator struct{}

func (v noSlackSubscriberTypeValidator) Description(ctx context.Context) string {
	return "Slack subscribers cannot be added via API - they must use the OAuth flow"
}

func (v noSlackSubscriberTypeValidator) MarkdownDescription(ctx context.Context) string {
	return "Slack subscribers cannot be added via API - they must use the OAuth flow through the Hyperping dashboard"
}

func (v noSlackSubscriberTypeValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	if req.ConfigValue.ValueString() == "slack" {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Slack Subscribers Not Supported",
			"Slack subscribers cannot be added via the Terraform provider. "+
				"They must be configured through the Hyperping OAuth flow in the dashboard. "+
				"Supported types are: email, sms, teams.",
		)
	}
}

// NoSlackSubscriberType returns a validator that rejects "slack" type.
func NoSlackSubscriberType() validator.String {
	return noSlackSubscriberTypeValidator{}
}
