// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

// Acceptance tests that exercise the four remaining MCP-backed data
// sources (escalation_policies, integrations, on_call_schedules,
// on_call_schedule) end-to-end against the strict MCP fixture introduced
// for TF-06. The pre-TF-06 test suite had no httptest-backed coverage
// for these four data sources, so the v0.6.x silent breakage was
// invisible in CI. With the strict fixture in place, any future
// regression to the nil-arguments wire shape (or to a stale result
// envelope) would fail one of these tests.

package provider

import (
	"fmt"
	"testing"

	tfresource "github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccEscalationPoliciesDataSource_basic(t *testing.T) {
	t.Setenv("HYPERPING_ALLOW_LOCAL", "1")

	server := newStrictMCPTestServer(t, map[string]strictMCPTool{
		"list_escalation_policies": {
			Handler: func(_ map[string]any) (any, error) {
				return []any{
					map[string]any{
						"uuid":  "ep_a",
						"name":  "A-Policy",
						"team":  "SRE",
						"steps": []any{},
					},
					map[string]any{
						"uuid": "ep_b",
						"name": "B-Policy",
						"team": "Platform",
						"steps": []any{
							map[string]any{
								"delay":       10,
								"target_type": "schedule",
								"target_id":   "sch_1",
							},
						},
					},
				}, nil
			},
		},
	})
	defer server.Close()

	tfresource.Test(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config: fmt.Sprintf(`
provider "hyperping" {
  api_key = "sk_test"
  mcp_url = %[1]q
}

data "hyperping_escalation_policies" "all" {}
`, server.URL),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("data.hyperping_escalation_policies.all", "policies.#", "2"),
					tfresource.TestCheckResourceAttr("data.hyperping_escalation_policies.all", "ids.#", "2"),
					tfresource.TestCheckResourceAttr("data.hyperping_escalation_policies.all", "ids.0", "ep_a"),
					tfresource.TestCheckResourceAttr("data.hyperping_escalation_policies.all", "ids.1", "ep_b"),
				),
			},
		},
	})
}

func TestAccIntegrationsDataSource_basic(t *testing.T) {
	t.Setenv("HYPERPING_ALLOW_LOCAL", "1")

	// list_integrations returns a top-level array per the v0.7.1 SDK
	// contract (mcp_client.go ListIntegrations does result.([]any)).
	server := newStrictMCPTestServer(t, map[string]strictMCPTool{
		"list_integrations": {
			Handler: func(_ map[string]any) (any, error) {
				return []any{
					map[string]any{
						"uuid":         "int_1",
						"name":         "Slack-Prod",
						"type":         "slack",
						"enabled":      true,
						"last_test_at": "2026-06-08T00:00:00Z",
						"created_at":   "2026-01-15T00:00:00Z",
					},
				}, nil
			},
		},
	})
	defer server.Close()

	tfresource.Test(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config: fmt.Sprintf(`
provider "hyperping" {
  api_key = "sk_test"
  mcp_url = %[1]q
}

data "hyperping_integrations" "all" {}
`, server.URL),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("data.hyperping_integrations.all", "integrations.#", "1"),
					tfresource.TestCheckResourceAttr("data.hyperping_integrations.all", "integrations.0.id", "int_1"),
					tfresource.TestCheckResourceAttr("data.hyperping_integrations.all", "integrations.0.name", "Slack-Prod"),
					tfresource.TestCheckResourceAttr("data.hyperping_integrations.all", "integrations.0.type", "slack"),
					tfresource.TestCheckResourceAttr("data.hyperping_integrations.all", "integrations.0.enabled", "true"),
					tfresource.TestCheckResourceAttr("data.hyperping_integrations.all", "ids.0", "int_1"),
				),
			},
		},
	})
}

// onCallSchedulesStub returns the result envelope shape the v0.7.1 SDK
// expects from list_on_call_schedules: an object with a "schedules" key
// whose value is an array of schedule records. ListOnCallSchedules in
// mcp_client.go does data["schedules"].([]any), so a top-level array
// would silently return zero schedules. Centralizing the envelope here
// keeps the two acceptance tests below in lockstep with that contract.
func onCallSchedulesStub() any {
	return map[string]any{
		"schedules": []any{
			map[string]any{
				"uuid":           "sch_1",
				"name":           "Primary-Rotation",
				"team":           "SRE",
				"current_oncall": "alice@example.com",
				"next_oncall":    "bob@example.com",
				"rotation_start": "2026-06-01T00:00:00Z",
				"rotation_end":   "2026-06-08T00:00:00Z",
			},
		},
	}
}

func TestAccOnCallSchedulesDataSource_basic(t *testing.T) {
	t.Setenv("HYPERPING_ALLOW_LOCAL", "1")

	server := newStrictMCPTestServer(t, map[string]strictMCPTool{
		"list_on_call_schedules": {
			Handler: func(_ map[string]any) (any, error) { return onCallSchedulesStub(), nil },
		},
	})
	defer server.Close()

	tfresource.Test(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config: fmt.Sprintf(`
provider "hyperping" {
  api_key = "sk_test"
  mcp_url = %[1]q
}

data "hyperping_on_call_schedules" "all" {}
`, server.URL),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("data.hyperping_on_call_schedules.all", "schedules.#", "1"),
					tfresource.TestCheckResourceAttr("data.hyperping_on_call_schedules.all", "schedules.0.id", "sch_1"),
					tfresource.TestCheckResourceAttr("data.hyperping_on_call_schedules.all", "schedules.0.name", "Primary-Rotation"),
					tfresource.TestCheckResourceAttr("data.hyperping_on_call_schedules.all", "schedules.0.current_oncall", "alice@example.com"),
					tfresource.TestCheckResourceAttr("data.hyperping_on_call_schedules.all", "ids.0", "sch_1"),
				),
			},
		},
	})
}

func TestAccOnCallScheduleDataSource_basic(t *testing.T) {
	t.Setenv("HYPERPING_ALLOW_LOCAL", "1")

	// OnCallScheduleDataSource.Read actually calls
	// ListOnCallSchedules and filters client-side; see
	// on_call_schedule_data_source.go:125. So the strict server only
	// needs the list tool wired, not get_on_call_schedule.
	server := newStrictMCPTestServer(t, map[string]strictMCPTool{
		"list_on_call_schedules": {
			Handler: func(_ map[string]any) (any, error) { return onCallSchedulesStub(), nil },
		},
	})
	defer server.Close()

	tfresource.Test(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config: fmt.Sprintf(`
provider "hyperping" {
  api_key = "sk_test"
  mcp_url = %[1]q
}

data "hyperping_on_call_schedule" "primary" {
  name = "Primary-Rotation"
}
`, server.URL),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("data.hyperping_on_call_schedule.primary", "id", "sch_1"),
					tfresource.TestCheckResourceAttr("data.hyperping_on_call_schedule.primary", "name", "Primary-Rotation"),
					tfresource.TestCheckResourceAttr("data.hyperping_on_call_schedule.primary", "team", "SRE"),
					tfresource.TestCheckResourceAttr("data.hyperping_on_call_schedule.primary", "current_oncall", "alice@example.com"),
				),
			},
		},
	})
}
