// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	providerschema "github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func TestMonitorSchema_RequestHeaderValueIsWriteOnly(t *testing.T) {
	r := &MonitorResource{}
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	headersRaw, ok := resp.Schema.Attributes["request_headers"]
	if !ok {
		t.Fatal("request_headers attribute not found in monitor schema")
	}

	headersAttr, ok := headersRaw.(schema.ListNestedAttribute)
	if !ok {
		t.Fatalf("request_headers is not a ListNestedAttribute, got %T", headersRaw)
	}

	valueRaw, ok := headersAttr.NestedObject.Attributes["value"]
	if !ok {
		t.Fatal("value attribute not found in request_headers nested object")
	}

	valueAttr, ok := valueRaw.(schema.StringAttribute)
	if !ok {
		t.Fatalf("request_headers[].value is not a StringAttribute, got %T", valueRaw)
	}

	if !valueAttr.WriteOnly {
		t.Error("expected request_headers[].value to have WriteOnly: true (TF-09)")
	}
}

func TestSubscriberSchema_EmailIsWriteOnly(t *testing.T) {
	r := &StatusPageSubscriberResource{}
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	attrRaw, ok := resp.Schema.Attributes["email"]
	if !ok {
		t.Fatal("email attribute not found in subscriber schema")
	}

	attr, ok := attrRaw.(schema.StringAttribute)
	if !ok {
		t.Fatalf("email is not a StringAttribute, got %T", attrRaw)
	}

	if !attr.WriteOnly {
		t.Error("expected email to have WriteOnly: true (TF-09)")
	}
}

func TestSubscriberSchema_PhoneIsWriteOnly(t *testing.T) {
	r := &StatusPageSubscriberResource{}
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	attrRaw, ok := resp.Schema.Attributes["phone"]
	if !ok {
		t.Fatal("phone attribute not found in subscriber schema")
	}

	attr, ok := attrRaw.(schema.StringAttribute)
	if !ok {
		t.Fatalf("phone is not a StringAttribute, got %T", attrRaw)
	}

	if !attr.WriteOnly {
		t.Error("expected phone to have WriteOnly: true (TF-09)")
	}
}

func TestSubscriberSchema_TeamsWebhookURLIsWriteOnly(t *testing.T) {
	r := &StatusPageSubscriberResource{}
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	attrRaw, ok := resp.Schema.Attributes["teams_webhook_url"]
	if !ok {
		t.Fatal("teams_webhook_url attribute not found in subscriber schema")
	}

	attr, ok := attrRaw.(schema.StringAttribute)
	if !ok {
		t.Fatalf("teams_webhook_url is not a StringAttribute, got %T", attrRaw)
	}

	if !attr.WriteOnly {
		t.Error("expected teams_webhook_url to have WriteOnly: true (TF-09)")
	}
}

// TestProviderSchema_APIKeyIsNotWriteOnly documents the framework limitation:
// terraform-plugin-framework v1.19.0 does not support WriteOnly on provider schema
// attributes (IsWriteOnly() is hard-coded to return false). Tracked as TF-10.
func TestProviderSchema_APIKeyIsNotWriteOnly(t *testing.T) {
	p := &HyperpingProvider{}
	resp := &provider.SchemaResponse{}
	p.Schema(context.Background(), provider.SchemaRequest{}, resp)

	attrRaw, ok := resp.Schema.Attributes["api_key"]
	if !ok {
		t.Fatal("api_key attribute not found in provider schema")
	}

	attr, ok := attrRaw.(providerschema.StringAttribute)
	if !ok {
		t.Fatalf("api_key is not a provider StringAttribute, got %T", attrRaw)
	}

	if attr.IsWriteOnly() {
		t.Error("api_key should not be write-only: framework does not support WriteOnly on provider schemas (track TF-10)")
	}
}
