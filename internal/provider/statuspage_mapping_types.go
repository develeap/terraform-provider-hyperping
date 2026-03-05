// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/develeap/terraform-provider-hyperping/internal/client"
)

// =============================================================================
// Subscriber Mapping
// =============================================================================

// mapSubscriberToTF converts API subscriber to Terraform model fields.
func mapSubscriberToTF(sub *client.StatusPageSubscriber, _ *diag.Diagnostics) SubscriberCommonFields {
	if sub == nil {
		return SubscriberCommonFields{
			ID:           types.Int64Null(),
			Type:         types.StringNull(),
			Value:        types.StringNull(),
			Email:        types.StringNull(),
			Phone:        types.StringNull(),
			SlackChannel: types.StringNull(),
			CreatedAt:    types.StringNull(),
		}
	}

	result := SubscriberCommonFields{
		ID:        types.Int64Value(int64(sub.ID)),
		Type:      types.StringValue(sub.Type),
		Value:     types.StringValue(sub.Value),
		CreatedAt: types.StringValue(sub.CreatedAt),
	}

	// Handle optional email
	if sub.Email != nil && *sub.Email != "" {
		result.Email = types.StringValue(*sub.Email)
	} else {
		result.Email = types.StringNull()
	}

	// Handle optional phone
	if sub.Phone != nil && *sub.Phone != "" {
		result.Phone = types.StringValue(*sub.Phone)
	} else {
		result.Phone = types.StringNull()
	}

	// Handle optional slack_channel
	if sub.SlackChannel != nil && *sub.SlackChannel != "" {
		result.SlackChannel = types.StringValue(*sub.SlackChannel)
	} else {
		result.SlackChannel = types.StringNull()
	}

	return result
}

// SubscriberCommonFields contains fields shared between subscriber resource and data sources.
type SubscriberCommonFields struct {
	ID           types.Int64  `tfsdk:"id"`
	Type         types.String `tfsdk:"type"`
	Value        types.String `tfsdk:"value"`
	Email        types.String `tfsdk:"email"`
	Phone        types.String `tfsdk:"phone"`
	SlackChannel types.String `tfsdk:"slack_channel"`
	CreatedAt    types.String `tfsdk:"created_at"`
}

// =============================================================================
// AttrTypes Helpers (for schema definitions)
// =============================================================================

// StatusPageSettingsAttrTypes returns the attribute types for settings nested object.
func StatusPageSettingsAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":                     types.StringType,
		"website":                  types.StringType,
		"description":              types.StringType,
		"languages":                types.ListType{ElemType: types.StringType},
		"default_language":         types.StringType,
		"theme":                    types.StringType,
		"font":                     types.StringType,
		"accent_color":             types.StringType,
		"auto_refresh":             types.BoolType,
		"banner_header":            types.BoolType,
		"logo":                     types.StringType,
		"logo_height":              types.StringType,
		"favicon":                  types.StringType,
		"hide_powered_by":          types.BoolType,
		"hide_from_search_engines": types.BoolType,
		"google_analytics":         types.StringType,
		"subscribe":                types.ObjectType{AttrTypes: SubscribeSettingsAttrTypes()},
		"authentication":           types.ObjectType{AttrTypes: AuthenticationSettingsAttrTypes()},
	}
}

// SubscribeSettingsAttrTypes returns the attribute types for subscribe settings nested object.
func SubscribeSettingsAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"enabled": types.BoolType,
		"email":   types.BoolType,
		"slack":   types.BoolType,
		"teams":   types.BoolType,
		"sms":     types.BoolType,
	}
}

// AuthenticationSettingsAttrTypes returns the attribute types for authentication settings nested object.
func AuthenticationSettingsAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"password_protection": types.BoolType,
		"google_sso":          types.BoolType,
		"saml_sso":            types.BoolType,
		"allowed_domains":     types.ListType{ElemType: types.StringType},
	}
}

// SectionAttrTypes returns the attribute types for section nested object.
func SectionAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":     types.MapType{ElemType: types.StringType},
		"is_split": types.BoolType,
		"services": types.ListType{ElemType: types.ObjectType{AttrTypes: ServiceAttrTypes()}},
	}
}

// NestedServiceAttrTypes returns attribute types for services nested inside a group.
// This is one level deep -- nested services are not groups themselves.
func NestedServiceAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":                  types.StringType,
		"uuid":                types.StringType,
		"name":                types.MapType{ElemType: types.StringType},
		"is_group":            types.BoolType,
		"show_uptime":         types.BoolType,
		"show_response_times": types.BoolType,
	}
}

// ServiceAttrTypes returns the attribute types for top-level services in a section.
// Top-level services can be group entries with a nested services list.
func ServiceAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":                  types.StringType,
		"uuid":                types.StringType,
		"name":                types.MapType{ElemType: types.StringType},
		"is_group":            types.BoolType,
		"show_uptime":         types.BoolType,
		"show_response_times": types.BoolType,
		"services":            types.ListType{ElemType: types.ObjectType{AttrTypes: NestedServiceAttrTypes()}},
	}
}
