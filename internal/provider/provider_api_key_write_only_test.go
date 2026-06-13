// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	providerschema "github.com/hashicorp/terraform-plugin-framework/provider/schema"
)

// TestProviderSchema_APIKeyNotWriteOnly documents the framework limitation:
// terraform-plugin-framework v1.19.0 does not support WriteOnly on provider schema
// attributes (IsWriteOnly() is hard-coded to return false). Tracked as TF-10.
func TestProviderSchema_APIKeyNotWriteOnly(t *testing.T) {
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
