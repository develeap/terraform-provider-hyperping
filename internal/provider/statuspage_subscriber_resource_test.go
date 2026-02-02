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

func TestAccStatusPageSubscriberResource_email(t *testing.T) {
	server := newMockStatusPageServer(t)
	defer server.Close()

	tfresource.Test(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			// Create status page first
			{
				Config: testAccStatusPageSubscriberResourceConfig_email(server.URL),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_statuspage_subscriber.email", "type", "email"),
					tfresource.TestCheckResourceAttr("hyperping_statuspage_subscriber.email", "email", "team@example.com"),
					tfresource.TestCheckResourceAttr("hyperping_statuspage_subscriber.email", "language", "en"),
					tfresource.TestCheckResourceAttrSet("hyperping_statuspage_subscriber.email", "id"),
					tfresource.TestCheckResourceAttrSet("hyperping_statuspage_subscriber.email", "created_at"),
				),
			},
			// ImportState testing
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

func TestAccStatusPageSubscriberResource_sms(t *testing.T) {
	server := newMockStatusPageServer(t)
	defer server.Close()

	tfresource.Test(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config: testAccStatusPageSubscriberResourceConfig_sms(server.URL),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_statuspage_subscriber.sms", "type", "sms"),
					tfresource.TestCheckResourceAttr("hyperping_statuspage_subscriber.sms", "phone", "+1234567890"),
					tfresource.TestCheckResourceAttrSet("hyperping_statuspage_subscriber.sms", "id"),
				),
			},
		},
	})
}

func TestAccStatusPageSubscriberResource_teams(t *testing.T) {
	server := newMockStatusPageServer(t)
	defer server.Close()

	tfresource.Test(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config: testAccStatusPageSubscriberResourceConfig_teams(server.URL),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_statuspage_subscriber.teams", "type", "teams"),
					tfresource.TestCheckResourceAttr("hyperping_statuspage_subscriber.teams", "teams_webhook_url", "https://outlook.office.com/webhook/test"),
					tfresource.TestCheckResourceAttrSet("hyperping_statuspage_subscriber.teams", "id"),
				),
			},
		},
	})
}

func TestAccStatusPageSubscriberResource_slackRejected(t *testing.T) {
	server := newMockStatusPageServer(t)
	defer server.Close()

	tfresource.Test(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config:      testAccStatusPageSubscriberResourceConfig_slack(server.URL),
				ExpectError: regexp.MustCompile("Slack subscribers cannot be added via the Terraform provider"),
			},
		},
	})
}

func TestAccStatusPageSubscriberResource_multipleSubscribers(t *testing.T) {
	server := newMockStatusPageServer(t)
	defer server.Close()

	tfresource.Test(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config: testAccStatusPageSubscriberResourceConfig_multiple(server.URL),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_statuspage_subscriber.email", "type", "email"),
					tfresource.TestCheckResourceAttr("hyperping_statuspage_subscriber.sms", "type", "sms"),
					tfresource.TestCheckResourceAttr("hyperping_statuspage_subscriber.teams", "type", "teams"),
				),
			},
		},
	})
}

// Helper functions

func testAccStatusPageSubscriberResourceConfig_email(baseURL string) string {
	return fmt.Sprintf(`
provider "hyperping" {
  api_key  = "test_api_key"
  base_url = %[1]q
}

resource "hyperping_statuspage" "test" {
  name             = "Test Status Page"
  hosted_subdomain = "test-status"

  settings = {
    name      = "Test Settings"
    languages = ["en"]
  }
}

resource "hyperping_statuspage_subscriber" "email" {
  statuspage_uuid = hyperping_statuspage.test.id
  type            = "email"
  email           = "team@example.com"
  language        = "en"
}
`, baseURL)
}

func testAccStatusPageSubscriberResourceConfig_sms(baseURL string) string {
	return fmt.Sprintf(`
provider "hyperping" {
  api_key  = "test_api_key"
  base_url = %[1]q
}

resource "hyperping_statuspage" "test" {
  name             = "Test Status Page"
  hosted_subdomain = "test-status"

  settings = {
    name      = "Test Settings"
    languages = ["en"]
  }
}

resource "hyperping_statuspage_subscriber" "sms" {
  statuspage_uuid = hyperping_statuspage.test.id
  type            = "sms"
  phone           = "+1234567890"
  language        = "en"
}
`, baseURL)
}

func testAccStatusPageSubscriberResourceConfig_teams(baseURL string) string {
	return fmt.Sprintf(`
provider "hyperping" {
  api_key  = "test_api_key"
  base_url = %[1]q
}

resource "hyperping_statuspage" "test" {
  name             = "Test Status Page"
  hosted_subdomain = "test-status"

  settings = {
    name      = "Test Settings"
    languages = ["en"]
  }
}

resource "hyperping_statuspage_subscriber" "teams" {
  statuspage_uuid = hyperping_statuspage.test.id
  type            = "teams"
  teams_webhook_url = "https://outlook.office.com/webhook/test"
}
`, baseURL)
}

func testAccStatusPageSubscriberResourceConfig_slack(baseURL string) string {
	return fmt.Sprintf(`
provider "hyperping" {
  api_key  = "test_api_key"
  base_url = %[1]q
}

resource "hyperping_statuspage" "test" {
  name             = "Test Status Page"
  hosted_subdomain = "test-status"

  settings = {
    name      = "Test Settings"
    languages = ["en"]
  }
}

resource "hyperping_statuspage_subscriber" "slack" {
  statuspage_uuid = hyperping_statuspage.test.id
  type            = "slack"
}
`, baseURL)
}

func testAccStatusPageSubscriberResourceConfig_multiple(baseURL string) string {
	return fmt.Sprintf(`
provider "hyperping" {
  api_key  = "test_api_key"
  base_url = %[1]q
}

resource "hyperping_statuspage" "test" {
  name             = "Test Status Page"
  hosted_subdomain = "test-status"

  settings = {
    name      = "Test Settings"
    languages = ["en"]
  }
}

resource "hyperping_statuspage_subscriber" "email" {
  statuspage_uuid = hyperping_statuspage.test.id
  type            = "email"
  email           = "team@example.com"
  language        = "en"
}

resource "hyperping_statuspage_subscriber" "sms" {
  statuspage_uuid = hyperping_statuspage.test.id
  type            = "sms"
  phone           = "+1234567890"
  language        = "en"
}

resource "hyperping_statuspage_subscriber" "teams" {
  statuspage_uuid = hyperping_statuspage.test.id
  type            = "teams"
  teams_webhook_url = "https://outlook.office.com/webhook/test"
}
`, baseURL)
}

// Validation tests for cross-field requirements

func TestAccStatusPageSubscriberResource_emailRequired(t *testing.T) {
	server := newMockStatusPageServer(t)
	defer server.Close()

	tfresource.Test(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config:      testAccStatusPageSubscriberResourceConfig_emailMissing(server.URL),
				ExpectError: regexp.MustCompile("is required when type is"),
			},
		},
	})
}

func TestAccStatusPageSubscriberResource_phonRequired(t *testing.T) {
	server := newMockStatusPageServer(t)
	defer server.Close()

	tfresource.Test(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config:      testAccStatusPageSubscriberResourceConfig_phoneMissing(server.URL),
				ExpectError: regexp.MustCompile("is required when type is"),
			},
		},
	})
}

func TestAccStatusPageSubscriberResource_teamsWebhookRequired(t *testing.T) {
	server := newMockStatusPageServer(t)
	defer server.Close()

	tfresource.Test(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config:      testAccStatusPageSubscriberResourceConfig_teamsWebhookMissing(server.URL),
				ExpectError: regexp.MustCompile("is required when type is"),
			},
		},
	})
}

// Helper configs for validation tests

func testAccStatusPageSubscriberResourceConfig_emailMissing(baseURL string) string {
	return fmt.Sprintf(`
provider "hyperping" {
  api_key  = "test_api_key"
  base_url = %[1]q
}

resource "hyperping_statuspage" "test" {
  name             = "Test Status Page"
  hosted_subdomain = "test-status"

  settings = {
    name      = "Test Settings"
    languages = ["en"]
  }
}

resource "hyperping_statuspage_subscriber" "invalid" {
  statuspage_uuid = hyperping_statuspage.test.id
  type            = "email"
  # email intentionally missing - should fail validation
}
`, baseURL)
}

func testAccStatusPageSubscriberResourceConfig_phoneMissing(baseURL string) string {
	return fmt.Sprintf(`
provider "hyperping" {
  api_key  = "test_api_key"
  base_url = %[1]q
}

resource "hyperping_statuspage" "test" {
  name             = "Test Status Page"
  hosted_subdomain = "test-status"

  settings = {
    name      = "Test Settings"
    languages = ["en"]
  }
}

resource "hyperping_statuspage_subscriber" "invalid" {
  statuspage_uuid = hyperping_statuspage.test.id
  type            = "sms"
  # phone intentionally missing - should fail validation
}
`, baseURL)
}

func testAccStatusPageSubscriberResourceConfig_teamsWebhookMissing(baseURL string) string {
	return fmt.Sprintf(`
provider "hyperping" {
  api_key  = "test_api_key"
  base_url = %[1]q
}

resource "hyperping_statuspage" "test" {
  name             = "Test Status Page"
  hosted_subdomain = "test-status"

  settings = {
    name      = "Test Settings"
    languages = ["en"]
  }
}

resource "hyperping_statuspage_subscriber" "invalid" {
  statuspage_uuid = hyperping_statuspage.test.id
  type            = "teams"
  # teams_webhook_url intentionally missing - should fail validation
}
`, baseURL)
}
