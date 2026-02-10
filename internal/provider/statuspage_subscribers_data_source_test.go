// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	tfresource "github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccStatusPageSubscribersDataSource_listAll(t *testing.T) {
	server := newMockStatusPageServer(t)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config: testAccStatusPageSubscribersDataSourceConfig_listAll(server.URL),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("data.hyperping_statuspage_subscribers.all", "subscribers.#", "4"),
					tfresource.TestCheckResourceAttr("data.hyperping_statuspage_subscribers.all", "total", "4"),
				),
			},
		},
	})
}

func TestAccStatusPageSubscribersDataSource_filterByType(t *testing.T) {
	server := newMockStatusPageServer(t)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config: testAccStatusPageSubscribersDataSourceConfig_filterByType(server.URL),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					// Note: Mock server doesn't implement actual type filtering,
					// but we test the parameter is accepted
					tfresource.TestCheckResourceAttr("data.hyperping_statuspage_subscribers.email_only", "type", "email"),
				),
			},
		},
	})
}

func TestAccStatusPageSubscribersDataSource_empty(t *testing.T) {
	server := newMockStatusPageServer(t)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config: testAccStatusPageSubscribersDataSourceConfig_empty(server.URL),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("data.hyperping_statuspage_subscribers.empty", "subscribers.#", "0"),
					tfresource.TestCheckResourceAttr("data.hyperping_statuspage_subscribers.empty", "total", "0"),
				),
			},
		},
	})
}

// Helper functions

func testAccStatusPageSubscribersDataSourceConfig_listAll(baseURL string) string {
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

resource "hyperping_statuspage_subscriber" "email1" {
  statuspage_uuid = hyperping_statuspage.test.id
  type            = "email"
  email           = "team@example.com"
}

resource "hyperping_statuspage_subscriber" "email2" {
  statuspage_uuid = hyperping_statuspage.test.id
  type            = "email"
  email           = "admin@example.com"
}

resource "hyperping_statuspage_subscriber" "sms" {
  statuspage_uuid = hyperping_statuspage.test.id
  type            = "sms"
  phone           = "+1234567890"
}

resource "hyperping_statuspage_subscriber" "teams" {
  statuspage_uuid = hyperping_statuspage.test.id
  type            = "teams"
  teams_webhook_url = "https://outlook.office.com/webhook/test"
}

data "hyperping_statuspage_subscribers" "all" {
  statuspage_uuid = hyperping_statuspage.test.id
  depends_on = [
    hyperping_statuspage_subscriber.email1,
    hyperping_statuspage_subscriber.email2,
    hyperping_statuspage_subscriber.sms,
    hyperping_statuspage_subscriber.teams,
  ]
}
`, baseURL)
}

func testAccStatusPageSubscribersDataSourceConfig_filterByType(baseURL string) string {
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
}

resource "hyperping_statuspage_subscriber" "sms" {
  statuspage_uuid = hyperping_statuspage.test.id
  type            = "sms"
  phone           = "+1234567890"
}

data "hyperping_statuspage_subscribers" "email_only" {
  statuspage_uuid = hyperping_statuspage.test.id
  type            = "email"
  depends_on = [
    hyperping_statuspage_subscriber.email,
    hyperping_statuspage_subscriber.sms,
  ]
}
`, baseURL)
}

func testAccStatusPageSubscribersDataSourceConfig_empty(baseURL string) string {
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

data "hyperping_statuspage_subscribers" "empty" {
  statuspage_uuid = hyperping_statuspage.test.id
}
`, baseURL)
}
