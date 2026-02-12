// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"regexp"
	"testing"

	tfresource "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccStatusPageSubscriberResource_import(t *testing.T) {
	server := newMockStatusPageServer(t)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			// Step 1: Create the resource
			{
				Config: testAccStatusPageSubscriberResourceConfig_email(server.URL),
				Check: tfresource.ComposeTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_statuspage_subscriber.email", "type", "email"),
					tfresource.TestCheckResourceAttr("hyperping_statuspage_subscriber.email", "email", "team@example.com"),
					tfresource.TestCheckResourceAttr("hyperping_statuspage_subscriber.email", "language", "en"),
					tfresource.TestCheckResourceAttrSet("hyperping_statuspage_subscriber.email", "id"),
					tfresource.TestCheckResourceAttrSet("hyperping_statuspage_subscriber.email", "created_at"),
				),
			},
			// Step 2: Import it and verify all fields match
			{
				ResourceName:      "hyperping_statuspage_subscriber.email",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs := s.RootModule().Resources["hyperping_statuspage_subscriber.email"]
					return fmt.Sprintf("%s:%s", rs.Primary.Attributes["statuspage_uuid"], rs.Primary.Attributes["id"]), nil
				},
			},
		},
	})
}

func TestAccStatusPageSubscriberResource_importSMS(t *testing.T) {
	server := newMockStatusPageServer(t)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			// Create resource
			{
				Config: testAccStatusPageSubscriberResourceConfig_sms(server.URL),
				Check: tfresource.ComposeTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_statuspage_subscriber.sms", "type", "sms"),
					tfresource.TestCheckResourceAttr("hyperping_statuspage_subscriber.sms", "phone", "+1234567890"),
					tfresource.TestCheckResourceAttrSet("hyperping_statuspage_subscriber.sms", "id"),
				),
			},
			// Import and verify
			{
				ResourceName:      "hyperping_statuspage_subscriber.sms",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs := s.RootModule().Resources["hyperping_statuspage_subscriber.sms"]
					return fmt.Sprintf("%s:%s", rs.Primary.Attributes["statuspage_uuid"], rs.Primary.Attributes["id"]), nil
				},
			},
		},
	})
}

func TestAccStatusPageSubscriberResource_importTeams(t *testing.T) {
	server := newMockStatusPageServer(t)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			// Create resource
			{
				Config: testAccStatusPageSubscriberResourceConfig_teams(server.URL),
				Check: tfresource.ComposeTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_statuspage_subscriber.teams", "type", "teams"),
					tfresource.TestCheckResourceAttr("hyperping_statuspage_subscriber.teams", "teams_webhook_url", "https://outlook.office.com/webhook/test"),
					tfresource.TestCheckResourceAttrSet("hyperping_statuspage_subscriber.teams", "id"),
				),
			},
			// Import and verify
			{
				ResourceName:      "hyperping_statuspage_subscriber.teams",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"teams_webhook_url", // Webhook URLs are sensitive and may not be returned by the API
				},
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs := s.RootModule().Resources["hyperping_statuspage_subscriber.teams"]
					return fmt.Sprintf("%s:%s", rs.Primary.Attributes["statuspage_uuid"], rs.Primary.Attributes["id"]), nil
				},
			},
		},
	})
}

func TestAccStatusPageSubscriberResource_importNotFound(t *testing.T) {
	server := newMockStatusPageServer(t)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			// Create resource first
			{
				Config: testAccStatusPageSubscriberResourceConfig_email(server.URL),
			},
			// Try to import non-existent resource with invalid subscriber ID format
			{
				ResourceName:  "hyperping_statuspage_subscriber.email",
				ImportState:   true,
				ImportStateId: "sp_test:sub_nonexistent",
				ExpectError:   regexp.MustCompile("Subscriber not found|Cannot import non-existent|Invalid Subscriber ID"),
			},
		},
	})
}
