// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"regexp"
	"testing"

	tfresource "github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccStatusPageResource_withGroupedServices verifies create, update (group name),
// and update (children count) for a status page that contains a nested group service.
func TestAccStatusPageResource_withGroupedServices(t *testing.T) {
	server := newMockStatusPageServer(t)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			// Step 1: Create a status page with one section containing one group service
			// that has two nested monitor children.
			{
				Config: testAccNestedGroupsConfig_initial(server.URL),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "sections.#", "1"),
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "sections.0.name.en", "Infrastructure"),
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "sections.0.services.#", "1"),
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "sections.0.services.0.is_group", "true"),
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "sections.0.services.0.name.en", "Payment Processing"),
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "sections.0.services.0.services.#", "2"),
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "sections.0.services.0.services.0.name.en", "Payment API"),
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "sections.0.services.0.services.1.name.en", "Payment DB"),
					tfresource.TestCheckResourceAttrSet("hyperping_statuspage.test", "id"),
				),
			},
			// Step 2: Update the group name; verify children are still present.
			{
				Config: testAccNestedGroupsConfig_updatedGroupName(server.URL),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "sections.0.services.0.is_group", "true"),
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "sections.0.services.0.name.en", "Core Services"),
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "sections.0.services.0.services.#", "2"),
				),
			},
			// Step 3: Update children to just one entry; verify update is applied.
			{
				Config: testAccNestedGroupsConfig_oneChild(server.URL),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "sections.0.services.0.is_group", "true"),
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "sections.0.services.0.services.#", "1"),
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "sections.0.services.0.services.0.name.en", "Payment API"),
				),
			},
		},
	})
}

// TestAccStatusPageResource_flatAndGroupMixed verifies a section with both a flat
// service and a group service can be created without drift on re-plan.
func TestAccStatusPageResource_flatAndGroupMixed(t *testing.T) {
	server := newMockStatusPageServer(t)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config: testAccNestedGroupsConfig_flatAndGroup(server.URL),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "sections.#", "1"),
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "sections.0.services.#", "2"),
					// First service: flat monitor
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "sections.0.services.0.is_group", "false"),
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "sections.0.services.0.name.en", "API Service"),
					// Second service: group with nested children
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "sections.0.services.1.is_group", "true"),
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "sections.0.services.1.name.en", "Infrastructure Group"),
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "sections.0.services.1.services.#", "1"),
				),
			},
			// Re-apply the same config to verify no drift.
			{
				Config:   testAccNestedGroupsConfig_flatAndGroup(server.URL),
				PlanOnly: true,
			},
		},
	})
}

// TestAccStatusPageResource_groupValidationErrors verifies that invalid configurations
// produce the expected validation errors.
func TestAccStatusPageResource_groupValidationErrors(t *testing.T) {
	server := newMockStatusPageServer(t)
	defer server.Close()

	// A group service with an empty services list must be rejected.
	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config:      testAccNestedGroupsConfig_emptyGroupServices(server.URL),
				ExpectError: regexp.MustCompile(`group service must have at least one nested service`),
			},
		},
	})
}

// TestAccStatusPageResource_flatServiceNoUUIDError verifies that a flat (non-group)
// service without a uuid produces a validation error.
func TestAccStatusPageResource_flatServiceNoUUIDError(t *testing.T) {
	server := newMockStatusPageServer(t)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config:      testAccNestedGroupsConfig_flatNoUUID(server.URL),
				ExpectError: regexp.MustCompile(`uuid required for non-group service`),
			},
		},
	})
}

// =============================================================================
// Config helpers
// =============================================================================

func testAccNestedGroupsConfig_initial(baseURL string) string {
	return testAccStatusPageProviderConfig(baseURL) + `
resource "hyperping_statuspage" "test" {
  name             = "Test Status Page"
  hosted_subdomain = "test-nested-groups"

  settings = {
    name      = "Test Status Page"
    languages = ["en"]
  }

  sections = [
    {
      name = {
        en = "Infrastructure"
      }
      services = [
        {
          is_group = true
          name = {
            en = "Payment Processing"
          }
          services = [
            {
              uuid = "mon_child_1"
              name = {
                en = "Payment API"
              }
            },
            {
              uuid = "mon_child_2"
              name = {
                en = "Payment DB"
              }
            },
          ]
        }
      ]
    }
  ]
}
`
}

func testAccNestedGroupsConfig_updatedGroupName(baseURL string) string {
	return testAccStatusPageProviderConfig(baseURL) + `
resource "hyperping_statuspage" "test" {
  name             = "Test Status Page"
  hosted_subdomain = "test-nested-groups"

  settings = {
    name      = "Test Status Page"
    languages = ["en"]
  }

  sections = [
    {
      name = {
        en = "Infrastructure"
      }
      services = [
        {
          is_group = true
          name = {
            en = "Core Services"
          }
          services = [
            {
              uuid = "mon_child_1"
              name = {
                en = "Payment API"
              }
            },
            {
              uuid = "mon_child_2"
              name = {
                en = "Payment DB"
              }
            },
          ]
        }
      ]
    }
  ]
}
`
}

func testAccNestedGroupsConfig_oneChild(baseURL string) string {
	return testAccStatusPageProviderConfig(baseURL) + `
resource "hyperping_statuspage" "test" {
  name             = "Test Status Page"
  hosted_subdomain = "test-nested-groups"

  settings = {
    name      = "Test Status Page"
    languages = ["en"]
  }

  sections = [
    {
      name = {
        en = "Infrastructure"
      }
      services = [
        {
          is_group = true
          name = {
            en = "Core Services"
          }
          services = [
            {
              uuid = "mon_child_1"
              name = {
                en = "Payment API"
              }
            },
          ]
        }
      ]
    }
  ]
}
`
}

func testAccNestedGroupsConfig_flatAndGroup(baseURL string) string {
	return testAccStatusPageProviderConfig(baseURL) + `
resource "hyperping_statuspage" "test" {
  name             = "Mixed Services Status Page"
  hosted_subdomain = "test-mixed-services"

  settings = {
    name      = "Mixed Services Status Page"
    languages = ["en"]
  }

  sections = [
    {
      name = {
        en = "All Services"
      }
      services = [
        {
          uuid     = "mon_flat_01"
          is_group = false
          name = {
            en = "API Service"
          }
        },
        {
          is_group = true
          name = {
            en = "Infrastructure Group"
          }
          services = [
            {
              uuid = "mon_nested_01"
              name = {
                en = "Database"
              }
            },
          ]
        },
      ]
    }
  ]
}
`
}

func testAccNestedGroupsConfig_emptyGroupServices(baseURL string) string {
	return testAccStatusPageProviderConfig(baseURL) + `
resource "hyperping_statuspage" "test" {
  name             = "Validation Test Status Page"
  hosted_subdomain = "test-validation-group"

  settings = {
    name      = "Validation Test Status Page"
    languages = ["en"]
  }

  sections = [
    {
      name = {
        en = "Services"
      }
      services = [
        {
          is_group = true
          name = {
            en = "Empty Group"
          }
          services = []
        },
      ]
    }
  ]
}
`
}

func testAccNestedGroupsConfig_flatNoUUID(baseURL string) string {
	return testAccStatusPageProviderConfig(baseURL) + `
resource "hyperping_statuspage" "test" {
  name             = "Validation Test Status Page"
  hosted_subdomain = "test-validation-flat"

  settings = {
    name      = "Validation Test Status Page"
    languages = ["en"]
  }

  sections = [
    {
      name = {
        en = "Services"
      }
      services = [
        {
          is_group = false
          name = {
            en = "Missing UUID Service"
          }
        },
      ]
    }
  ]
}
`
}
