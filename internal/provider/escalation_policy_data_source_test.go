// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	tfresource "github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccEscalationPolicyDataSource_basic(t *testing.T) {
	// This fixture points mcp_url at an httptest server on 127.0.0.1, which the
	// provider rejects by default. The opt-in env var is scoped to this test so
	// the production gate (default-deny localhost for mcp_url) remains intact
	// for real users; it is only relaxed here, where the loopback target is the
	// test server we just spun up in-process.
	t.Setenv("HYPERPING_ALLOW_LOCAL", "1")

	// Strict fixture: rejects the v0.6.x nil-arguments wire shape the way
	// the live /v1/mcp endpoint does. ListEscalationPolicies in
	// hyperping-go v0.7.1 sends arguments:{} so this path stays green;
	// any future regression to omitted/null arguments would fail here
	// instead of silently going green like the pre-TF-06 fixture.
	server := newStrictMCPTestServer(t, map[string]strictMCPTool{
		"list_escalation_policies": {
			Handler: func(_ map[string]any) (any, error) {
				return []any{
					map[string]any{
						"uuid": "ep_123",
						"name": "SRE-Policy",
						"team": "SRE-Team",
						"steps": []any{
							map[string]any{
								"delay":       5,
								"target_type": "user",
								"target_id":   "u_1",
							},
						},
					},
				}, nil
			},
		},
	})
	defer server.Close()

	// Note: using tfresource.Test rather than ParallelTest because t.Setenv
	// above is incompatible with t.Parallel (the Go testing runtime panics).
	tfresource.Test(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config: fmt.Sprintf(`
provider "hyperping" {
  api_key = "sk_test"
  mcp_url = %[1]q
}

data "hyperping_escalation_policy" "test" {
  name = "SRE-Policy"
}

data "hyperping_escalation_policies" "all" {}
`, server.URL),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("data.hyperping_escalation_policy.test", "id", "ep_123"),
					tfresource.TestCheckResourceAttr("data.hyperping_escalation_policy.test", "name", "SRE-Policy"),
					tfresource.TestCheckResourceAttr("data.hyperping_escalation_policy.test", "team", "SRE-Team"),
					tfresource.TestCheckResourceAttr("data.hyperping_escalation_policy.test", "steps.#", "1"),
					tfresource.TestCheckResourceAttr("data.hyperping_escalation_policy.test", "steps.0.delay", "5"),
					tfresource.TestCheckResourceAttr("data.hyperping_escalation_policy.test", "steps.0.target_type", "user"),
					tfresource.TestCheckResourceAttr("data.hyperping_escalation_policy.test", "steps.0.target_id", "u_1"),
					tfresource.TestCheckResourceAttr("data.hyperping_escalation_policies.all", "policies.#", "1"),
					tfresource.TestCheckResourceAttr("data.hyperping_escalation_policies.all", "ids.0", "ep_123"),
				),
			},
		},
	})
}
