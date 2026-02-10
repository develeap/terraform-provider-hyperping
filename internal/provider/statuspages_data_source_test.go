// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	tfresource "github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccStatusPagesDataSource_listAll(t *testing.T) {
	server := newMockStatusPageServer(t)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config: testAccStatusPagesDataSourceConfig_listAll(server.URL),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("data.hyperping_statuspages.all", "statuspages.#", "3"),
					tfresource.TestCheckResourceAttr("data.hyperping_statuspages.all", "total", "3"),
					tfresource.TestCheckResourceAttr("data.hyperping_statuspages.all", "has_next_page", "false"),
				),
			},
		},
	})
}

func TestAccStatusPagesDataSource_search(t *testing.T) {
	server := newMockStatusPageServer(t)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config: testAccStatusPagesDataSourceConfig_search(server.URL),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					// Note: Mock server doesn't implement actual search filtering,
					// but we test the parameter is accepted
					tfresource.TestCheckResourceAttr("data.hyperping_statuspages.filtered", "search", "prod"),
				),
			},
		},
	})
}

func TestAccStatusPagesDataSource_empty(t *testing.T) {
	server := newMockStatusPageServer(t)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config: testAccStatusPagesDataSourceConfig_empty(server.URL),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("data.hyperping_statuspages.empty", "statuspages.#", "0"),
					tfresource.TestCheckResourceAttr("data.hyperping_statuspages.empty", "total", "0"),
				),
			},
		},
	})
}

// Helper functions

func testAccStatusPagesDataSourceConfig_listAll(baseURL string) string {
	return fmt.Sprintf(`
provider "hyperping" {
  api_key  = "test_api_key"
  base_url = %[1]q
}

resource "hyperping_statuspage" "prod" {
  name             = "Production Status"
  hosted_subdomain = "prod-status"

  settings = {
    name      = "Production Settings"
    languages = ["en"]
  }
}

resource "hyperping_statuspage" "dev" {
  name             = "Development Status"
  hosted_subdomain = "dev-status"

  settings = {
    name      = "Development Settings"
    languages = ["en"]
  }
}

resource "hyperping_statuspage" "staging" {
  name             = "Staging Status"
  hosted_subdomain = "staging-status"

  settings = {
    name      = "Staging Settings"
    languages = ["en"]
  }
}

data "hyperping_statuspages" "all" {
  depends_on = [
    hyperping_statuspage.prod,
    hyperping_statuspage.dev,
    hyperping_statuspage.staging,
  ]
}
`, baseURL)
}

func testAccStatusPagesDataSourceConfig_search(baseURL string) string {
	return fmt.Sprintf(`
provider "hyperping" {
  api_key  = "test_api_key"
  base_url = %[1]q
}

resource "hyperping_statuspage" "prod" {
  name             = "Production Status"
  hosted_subdomain = "prod-status"

  settings = {
    name      = "Production Settings"
    languages = ["en"]
  }
}

resource "hyperping_statuspage" "dev" {
  name             = "Development Status"
  hosted_subdomain = "dev-status"

  settings = {
    name      = "Development Settings"
    languages = ["en"]
  }
}

data "hyperping_statuspages" "filtered" {
  search = "prod"
  depends_on = [
    hyperping_statuspage.prod,
    hyperping_statuspage.dev,
  ]
}
`, baseURL)
}

func testAccStatusPagesDataSourceConfig_empty(baseURL string) string {
	return fmt.Sprintf(`
provider "hyperping" {
  api_key  = "test_api_key"
  base_url = %[1]q
}

data "hyperping_statuspages" "empty" {}
`, baseURL)
}
