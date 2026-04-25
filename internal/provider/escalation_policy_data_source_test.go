// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	tfresource "github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccEscalationPolicyDataSource_basic(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var rpcReq struct {
			Method string `json:"method"`
		}
		json.NewDecoder(r.Body).Decode(&rpcReq)

		if rpcReq.Method == "initialize" {
			fmt.Fprintf(w, `{"jsonrpc":"2.0","id":1,"result":{"protocolVersion":"2025-03-26","capabilities":{},"serverInfo":{"name":"test","version":"1.0"}}}`)
			return
		}

		if rpcReq.Method == "tools/call" {
			// Simulate ListEscalationPolicies tool call
			// Note: hyperping-go transport expects tool response to be JSON string in content[0].text
			resultJSON := `[{"uuid":"ep_123","name":"SRE-Policy","team":"SRE-Team","steps":[{"delay":5,"target_type":"user","target_id":"u_1"}]}]`
			escaped, _ := json.Marshal(resultJSON)
			fmt.Fprintf(w, `{"jsonrpc":"2.0","id":2,"result":{"content":[{"type":"text","text":%s}]}}`, escaped)
			return
		}
	}))
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
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
