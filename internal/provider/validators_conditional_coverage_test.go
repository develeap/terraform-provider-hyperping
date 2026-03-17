// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// buildSimpleConfig creates a tfsdk.Config for testing conditional validators
// that need to read other attributes via req.Config.GetAttribute().
func buildSimpleConfig(t *testing.T, s schema.Schema, vals map[string]tftypes.Value) tfsdk.Config {
	t.Helper()

	attrTypes := make(map[string]tftypes.Type)
	for name, attr := range s.Attributes {
		attrTypes[name] = attr.GetType().TerraformType(context.Background())
	}
	objType := tftypes.Object{AttributeTypes: attrTypes}

	allVals := make(map[string]tftypes.Value)
	for name, attrType := range attrTypes {
		allVals[name] = tftypes.NewValue(attrType, nil) // null by default
	}
	for k, v := range vals {
		allVals[k] = v
	}

	return tfsdk.Config{
		Schema: s,
		Raw:    tftypes.NewValue(objType, allVals),
	}
}

// --- requiredWhenValueIsValidator.ValidateString ---

func TestRequiredWhenValueIs_ValidateString_NullSkips(t *testing.T) {
	t.Parallel()

	s := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"type":  schema.StringAttribute{Optional: true},
			"email": schema.StringAttribute{Optional: true},
		},
	}
	config := buildSimpleConfig(t, s, map[string]tftypes.Value{
		"type": tftypes.NewValue(tftypes.String, "email"),
	})

	v := RequiredWhenValueIs(path.Root("type"), "email", "subscriber type")
	req := validator.StringRequest{
		Path:        path.Root("email"),
		Config:      config,
		ConfigValue: types.StringNull(),
	}
	resp := &validator.StringResponse{}
	v.ValidateString(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("expected no error for null value, got: %v", resp.Diagnostics)
	}
}

func TestRequiredWhenValueIs_ValidateString_UnknownSkips(t *testing.T) {
	t.Parallel()

	s := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"type":  schema.StringAttribute{Optional: true},
			"email": schema.StringAttribute{Optional: true},
		},
	}
	config := buildSimpleConfig(t, s, map[string]tftypes.Value{
		"type": tftypes.NewValue(tftypes.String, "email"),
	})

	v := RequiredWhenValueIs(path.Root("type"), "email", "subscriber type")
	req := validator.StringRequest{
		Path:        path.Root("email"),
		Config:      config,
		ConfigValue: types.StringUnknown(),
	}
	resp := &validator.StringResponse{}
	v.ValidateString(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("expected no error for unknown value, got: %v", resp.Diagnostics)
	}
}

func TestRequiredWhenValueIs_ValidateString_MatchEmptyError(t *testing.T) {
	t.Parallel()

	s := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"type":  schema.StringAttribute{Optional: true},
			"email": schema.StringAttribute{Optional: true},
		},
	}
	config := buildSimpleConfig(t, s, map[string]tftypes.Value{
		"type":  tftypes.NewValue(tftypes.String, "email"),
		"email": tftypes.NewValue(tftypes.String, ""),
	})

	v := RequiredWhenValueIs(path.Root("type"), "email", "subscriber type")
	req := validator.StringRequest{
		Path:        path.Root("email"),
		Config:      config,
		ConfigValue: types.StringValue(""),
	}
	resp := &validator.StringResponse{}
	v.ValidateString(context.Background(), req, resp)

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected error when match value equals target and field is empty")
	}
}

func TestRequiredWhenValueIs_ValidateString_MatchSetPass(t *testing.T) {
	t.Parallel()

	s := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"type":  schema.StringAttribute{Optional: true},
			"email": schema.StringAttribute{Optional: true},
		},
	}
	config := buildSimpleConfig(t, s, map[string]tftypes.Value{
		"type":  tftypes.NewValue(tftypes.String, "email"),
		"email": tftypes.NewValue(tftypes.String, "user@example.com"),
	})

	v := RequiredWhenValueIs(path.Root("type"), "email", "subscriber type")
	req := validator.StringRequest{
		Path:        path.Root("email"),
		Config:      config,
		ConfigValue: types.StringValue("user@example.com"),
	}
	resp := &validator.StringResponse{}
	v.ValidateString(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("expected no error when field is set, got: %v", resp.Diagnostics)
	}
}

func TestRequiredWhenValueIs_ValidateString_NoMatchPass(t *testing.T) {
	t.Parallel()

	s := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"type":  schema.StringAttribute{Optional: true},
			"email": schema.StringAttribute{Optional: true},
		},
	}
	config := buildSimpleConfig(t, s, map[string]tftypes.Value{
		"type":  tftypes.NewValue(tftypes.String, "sms"),
		"email": tftypes.NewValue(tftypes.String, ""),
	})

	v := RequiredWhenValueIs(path.Root("type"), "email", "subscriber type")
	req := validator.StringRequest{
		Path:        path.Root("email"),
		Config:      config,
		ConfigValue: types.StringValue(""),
	}
	resp := &validator.StringResponse{}
	v.ValidateString(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("expected no error when match value differs, got: %v", resp.Diagnostics)
	}
}

// --- requiredWhenProtocolPortValidator.ValidateInt64 ---

func TestRequiredWhenProtocolPort_ValidateInt64_UnknownSkips(t *testing.T) {
	t.Parallel()

	s := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"protocol": schema.StringAttribute{Optional: true},
			"port":     schema.Int64Attribute{Optional: true},
		},
	}
	config := buildSimpleConfig(t, s, map[string]tftypes.Value{
		"protocol": tftypes.NewValue(tftypes.String, "port"),
	})

	v := RequiredWhenProtocolPort(path.Root("protocol"))
	req := validator.Int64Request{
		Path:        path.Root("port"),
		Config:      config,
		ConfigValue: types.Int64Unknown(),
	}
	resp := &validator.Int64Response{}
	v.ValidateInt64(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("expected no error for unknown port, got: %v", resp.Diagnostics)
	}
}

func TestRequiredWhenProtocolPort_ValidateInt64_PortProtocolNullPort(t *testing.T) {
	t.Parallel()

	s := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"protocol": schema.StringAttribute{Optional: true},
			"port":     schema.Int64Attribute{Optional: true},
		},
	}
	config := buildSimpleConfig(t, s, map[string]tftypes.Value{
		"protocol": tftypes.NewValue(tftypes.String, "port"),
	})

	v := RequiredWhenProtocolPort(path.Root("protocol"))
	req := validator.Int64Request{
		Path:        path.Root("port"),
		Config:      config,
		ConfigValue: types.Int64Null(),
	}
	resp := &validator.Int64Response{}
	v.ValidateInt64(context.Background(), req, resp)

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected error when protocol is 'port' and port is null")
	}
}

func TestRequiredWhenProtocolPort_ValidateInt64_PortProtocolSetPort(t *testing.T) {
	t.Parallel()

	s := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"protocol": schema.StringAttribute{Optional: true},
			"port":     schema.Int64Attribute{Optional: true},
		},
	}
	config := buildSimpleConfig(t, s, map[string]tftypes.Value{
		"protocol": tftypes.NewValue(tftypes.String, "port"),
		"port":     tftypes.NewValue(tftypes.Number, 5432),
	})

	v := RequiredWhenProtocolPort(path.Root("protocol"))
	req := validator.Int64Request{
		Path:        path.Root("port"),
		Config:      config,
		ConfigValue: types.Int64Value(5432),
	}
	resp := &validator.Int64Response{}
	v.ValidateInt64(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("expected no error when port is set, got: %v", resp.Diagnostics)
	}
}

func TestRequiredWhenProtocolPort_ValidateInt64_HTTPProtocolNullPort(t *testing.T) {
	t.Parallel()

	s := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"protocol": schema.StringAttribute{Optional: true},
			"port":     schema.Int64Attribute{Optional: true},
		},
	}
	config := buildSimpleConfig(t, s, map[string]tftypes.Value{
		"protocol": tftypes.NewValue(tftypes.String, "http"),
	})

	v := RequiredWhenProtocolPort(path.Root("protocol"))
	req := validator.Int64Request{
		Path:        path.Root("port"),
		Config:      config,
		ConfigValue: types.Int64Null(),
	}
	resp := &validator.Int64Response{}
	v.ValidateInt64(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("expected no error for non-port protocol, got: %v", resp.Diagnostics)
	}
}

// --- statusPageHostnameOrSubdomainValidator.ValidateString ---

func TestAtLeastOneOf_ValidateString_HostnameSet(t *testing.T) {
	t.Parallel()

	s := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"hostname":  schema.StringAttribute{Optional: true},
			"subdomain": schema.StringAttribute{Optional: true},
		},
	}
	config := buildSimpleConfig(t, s, map[string]tftypes.Value{
		"hostname":  tftypes.NewValue(tftypes.String, "status.example.com"),
		"subdomain": tftypes.NewValue(tftypes.String, ""),
	})

	v := AtLeastOneOf()
	req := validator.StringRequest{
		Path:        path.Root("hostname"),
		Config:      config,
		ConfigValue: types.StringValue("status.example.com"),
	}
	resp := &validator.StringResponse{}
	v.ValidateString(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("expected no error when hostname is set, got: %v", resp.Diagnostics)
	}
}

func TestAtLeastOneOf_ValidateString_SubdomainSet(t *testing.T) {
	t.Parallel()

	s := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"hostname":  schema.StringAttribute{Optional: true},
			"subdomain": schema.StringAttribute{Optional: true},
		},
	}
	config := buildSimpleConfig(t, s, map[string]tftypes.Value{
		"hostname":  tftypes.NewValue(tftypes.String, ""),
		"subdomain": tftypes.NewValue(tftypes.String, "mypage"),
	})

	v := AtLeastOneOf()
	req := validator.StringRequest{
		Path:        path.Root("subdomain"),
		Config:      config,
		ConfigValue: types.StringValue("mypage"),
	}
	resp := &validator.StringResponse{}
	v.ValidateString(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("expected no error when subdomain is set, got: %v", resp.Diagnostics)
	}
}

func TestAtLeastOneOf_ValidateString_HostnameEmptySubdomainSet(t *testing.T) {
	t.Parallel()

	s := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"hostname":  schema.StringAttribute{Optional: true},
			"subdomain": schema.StringAttribute{Optional: true},
		},
	}
	config := buildSimpleConfig(t, s, map[string]tftypes.Value{
		"hostname":  tftypes.NewValue(tftypes.String, ""),
		"subdomain": tftypes.NewValue(tftypes.String, "mypage"),
	})

	v := AtLeastOneOf()
	req := validator.StringRequest{
		Path:        path.Root("hostname"),
		Config:      config,
		ConfigValue: types.StringValue(""),
	}
	resp := &validator.StringResponse{}
	v.ValidateString(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("expected no error when other field is set, got: %v", resp.Diagnostics)
	}
}

func TestAtLeastOneOf_ValidateString_SubdomainEmptyHostnameSet(t *testing.T) {
	t.Parallel()

	s := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"hostname":  schema.StringAttribute{Optional: true},
			"subdomain": schema.StringAttribute{Optional: true},
		},
	}
	config := buildSimpleConfig(t, s, map[string]tftypes.Value{
		"hostname":  tftypes.NewValue(tftypes.String, "status.example.com"),
		"subdomain": tftypes.NewValue(tftypes.String, ""),
	})

	v := AtLeastOneOf()
	req := validator.StringRequest{
		Path:        path.Root("subdomain"),
		Config:      config,
		ConfigValue: types.StringValue(""),
	}
	resp := &validator.StringResponse{}
	v.ValidateString(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("expected no error when other field is set, got: %v", resp.Diagnostics)
	}
}

func TestAtLeastOneOf_ValidateString_BothEmptyHostnameError(t *testing.T) {
	t.Parallel()

	s := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"hostname":  schema.StringAttribute{Optional: true},
			"subdomain": schema.StringAttribute{Optional: true},
		},
	}
	config := buildSimpleConfig(t, s, map[string]tftypes.Value{
		"hostname":  tftypes.NewValue(tftypes.String, ""),
		"subdomain": tftypes.NewValue(tftypes.String, ""),
	})

	v := AtLeastOneOf()
	req := validator.StringRequest{
		Path:        path.Root("hostname"),
		Config:      config,
		ConfigValue: types.StringValue(""),
	}
	resp := &validator.StringResponse{}
	v.ValidateString(context.Background(), req, resp)

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected error when both hostname and subdomain are empty")
	}
}

func TestAtLeastOneOf_ValidateString_BothEmptySubdomainError(t *testing.T) {
	t.Parallel()

	s := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"hostname":  schema.StringAttribute{Optional: true},
			"subdomain": schema.StringAttribute{Optional: true},
		},
	}
	config := buildSimpleConfig(t, s, map[string]tftypes.Value{
		"hostname":  tftypes.NewValue(tftypes.String, ""),
		"subdomain": tftypes.NewValue(tftypes.String, ""),
	})

	v := AtLeastOneOf()
	req := validator.StringRequest{
		Path:        path.Root("subdomain"),
		Config:      config,
		ConfigValue: types.StringValue(""),
	}
	resp := &validator.StringResponse{}
	v.ValidateString(context.Background(), req, resp)

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected error when both hostname and subdomain are empty")
	}
}

func TestAtLeastOneOf_ValidateString_NullHostnameSubdomainSet(t *testing.T) {
	t.Parallel()

	s := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"hostname":  schema.StringAttribute{Optional: true},
			"subdomain": schema.StringAttribute{Optional: true},
		},
	}
	config := buildSimpleConfig(t, s, map[string]tftypes.Value{
		"subdomain": tftypes.NewValue(tftypes.String, "mypage"),
	})

	v := AtLeastOneOf()
	req := validator.StringRequest{
		Path:        path.Root("hostname"),
		Config:      config,
		ConfigValue: types.StringNull(),
	}
	resp := &validator.StringResponse{}
	v.ValidateString(context.Background(), req, resp)

	// Null hostname → falls through to check subdomain.
	// GetAttribute into string with null subdomain would error, but subdomain="mypage"
	// is set, so GetAttribute succeeds and sees non-empty → no error.
	// However, GetAttribute reading a null attribute into a Go string fails.
	// Since hostname is null in the config, this falls through.
	// The null path: IsNull=true → not "" empty, falls through. Path is "hostname" → isHostname=true.
	// otherFieldPath = subdomain. GetAttribute("subdomain") into string: subdomain="mypage" → works.
	// otherValue="mypage" which is non-empty, so no error.
	if resp.Diagnostics.HasError() {
		t.Errorf("expected no error when null hostname but subdomain is set, got: %v", resp.Diagnostics)
	}
}
