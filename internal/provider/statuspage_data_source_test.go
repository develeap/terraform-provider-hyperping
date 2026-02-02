// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"regexp"
	"testing"

	tfresource "github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccStatusPageDataSource_basic(t *testing.T) {
	server := newMockStatusPageServer(t)
	defer server.Close()

	tfresource.Test(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config: testAccStatusPageDataSourceConfig_basic(server.URL),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("data.hyperping_statuspage.test", "name", "Test Status Page"),
					tfresource.TestCheckResourceAttr("data.hyperping_statuspage.test", "hosted_subdomain", "test-status"),
					tfresource.TestCheckResourceAttrSet("data.hyperping_statuspage.test", "id"),
					tfresource.TestCheckResourceAttrSet("data.hyperping_statuspage.test", "url"),
				),
			},
		},
	})
}

func TestAccStatusPageDataSource_withSections(t *testing.T) {
	server := newMockStatusPageServer(t)
	defer server.Close()

	tfresource.Test(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config: testAccStatusPageDataSourceConfig_withSections(server.URL),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("data.hyperping_statuspage.test", "sections.#", "2"),
					tfresource.TestCheckResourceAttr("data.hyperping_statuspage.test", "sections.0.name.en", "API Services"),
				),
			},
		},
	})
}

func TestAccStatusPageDataSource_notFound(t *testing.T) {
	server := newMockStatusPageServer(t)
	defer server.Close()

	tfresource.Test(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config:      testAccStatusPageDataSourceConfig_notFound(server.URL),
				ExpectError: regexp.MustCompile("Status page not found"),
			},
		},
	})
}

// Helper functions

func testAccStatusPageDataSourceConfig_basic(baseURL string) string {
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

data "hyperping_statuspage" "test" {
  id = hyperping_statuspage.test.id
}
`, baseURL)
}

func testAccStatusPageDataSourceConfig_withSections(baseURL string) string {
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

  sections = [
    {
      name = {
        en = "API Services"
      }
      is_split = true
    },
    {
      name = {
        en = "Databases"
      }
      is_split = false
    }
  ]
}

data "hyperping_statuspage" "test" {
  id = hyperping_statuspage.test.id
}
`, baseURL)
}

func testAccStatusPageDataSourceConfig_notFound(baseURL string) string {
	return fmt.Sprintf(`
provider "hyperping" {
  api_key  = "test_api_key"
  base_url = %[1]q
}

data "hyperping_statuspage" "test" {
  id = "sp_nonexistent"
}
`, baseURL)
}
