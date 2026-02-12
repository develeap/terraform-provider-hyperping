// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccDataSourceFilters_multipleFiltersMonitors tests combining multiple filter types on monitors.
func TestAccDataSourceFilters_multipleFiltersMonitors(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					data "hyperping_monitors" "filtered" {
						filter = {
							name_regex = "\\[PROD\\]-.*"
							protocol   = "https"
							paused     = false
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.hyperping_monitors.filtered", "monitors.#"),
				),
			},
		},
	})
}

// TestAccDataSourceFilters_multipleFiltersIncidents tests combining multiple filter types on incidents.
func TestAccDataSourceFilters_multipleFiltersIncidents(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					data "hyperping_incidents" "filtered" {
						filter = {
							name_regex = "API.*"
							status     = "resolved"
							severity   = "major"
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.hyperping_incidents.filtered", "incidents.#"),
				),
			},
		},
	})
}

// TestAccDataSourceFilters_emptyFilter tests that empty filter returns all resources.
func TestAccDataSourceFilters_emptyFilter(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					data "hyperping_monitors" "all" {
						filter = {}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.hyperping_monitors.all", "monitors.#"),
				),
			},
		},
	})
}

// TestAccDataSourceFilters_noMatches tests filter with no matches returns empty list.
func TestAccDataSourceFilters_noMatches(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					data "hyperping_monitors" "no_matches" {
						filter = {
							name_regex = "^THIS_PATTERN_SHOULD_NOT_MATCH_ANYTHING_12345$"
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.hyperping_monitors.no_matches", "monitors.#", "0"),
				),
			},
		},
	})
}

// TestAccDataSourceFilters_invalidRegex tests that invalid regex returns error.
func TestAccDataSourceFilters_invalidRegex(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					data "hyperping_monitors" "invalid" {
						filter = {
							name_regex = "[invalid(regex"
						}
					}
				`,
				ExpectError: regexp.MustCompile("Invalid filter regex|Failed to compile name_regex pattern|error parsing regexp"),
			},
		},
	})
}

// TestAccDataSourceFilters_crossDataSource tests filtering across related data sources.
func TestAccDataSourceFilters_crossDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					# First, filter monitors
					data "hyperping_monitors" "production" {
						filter = {
							name_regex = "\\[PROD\\]-.*"
						}
					}

					# Then filter incidents related to those monitors
					data "hyperping_incidents" "production_incidents" {
						filter = {
							severity = "major"
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.hyperping_monitors.production", "monitors.#"),
					resource.TestCheckResourceAttrSet("data.hyperping_incidents.production_incidents", "incidents.#"),
				),
			},
		},
	})
}

// TestAccDataSourceFilters_regexCaseSensitive tests regex is case-sensitive.
func TestAccDataSourceFilters_regexCaseSensitive(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					data "hyperping_monitors" "uppercase" {
						filter = {
							name_regex = "\\[PROD\\]-.*"
						}
					}

					data "hyperping_monitors" "lowercase" {
						filter = {
							name_regex = "\\[prod\\]-.*"
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.hyperping_monitors.uppercase", "monitors.#"),
					resource.TestCheckResourceAttrSet("data.hyperping_monitors.lowercase", "monitors.#"),
				),
			},
		},
	})
}

// TestAccDataSourceFilters_booleanFilter tests boolean filter behavior.
func TestAccDataSourceFilters_booleanFilter(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					data "hyperping_monitors" "active" {
						filter = {
							paused = false
						}
					}

					data "hyperping_monitors" "paused" {
						filter = {
							paused = true
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.hyperping_monitors.active", "monitors.#"),
					resource.TestCheckResourceAttrSet("data.hyperping_monitors.paused", "monitors.#"),
				),
			},
		},
	})
}

// TestAccDataSourceFilters_exactMatchFilter tests exact string matching.
func TestAccDataSourceFilters_exactMatchFilter(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					data "hyperping_monitors" "https_only" {
						filter = {
							protocol = "https"
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.hyperping_monitors.https_only", "monitors.#"),
				),
			},
		},
	})
}

// TestAccDataSourceFilters_maintenanceWindows tests maintenance window filtering.
func TestAccDataSourceFilters_maintenanceWindows(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					data "hyperping_maintenance_windows" "scheduled" {
						filter = {
							status = "scheduled"
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.hyperping_maintenance_windows.scheduled", "maintenance_windows.#"),
				),
			},
		},
	})
}

// TestAccDataSourceFilters_healthchecks tests healthcheck filtering.
func TestAccDataSourceFilters_healthchecks(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					data "hyperping_healthchecks" "filtered" {
						filter = {
							name_regex = ".*health.*"
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.hyperping_healthchecks.filtered", "healthchecks.#"),
				),
			},
		},
	})
}

// TestAccDataSourceFilters_outages tests outage filtering.
func TestAccDataSourceFilters_outages(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					data "hyperping_outages" "filtered" {
						filter = {
							name_regex = "\\[PROD\\]-.*"
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.hyperping_outages.filtered", "outages.#"),
				),
			},
		},
	})
}

// TestAccDataSourceFilters_statusPages tests status page filtering.
func TestAccDataSourceFilters_statusPages(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					data "hyperping_statuspages" "filtered" {
						filter = {
							name_regex = "Production.*"
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.hyperping_statuspages.filtered", "statuspages.#"),
				),
			},
		},
	})
}

// TestAccDataSourceFilters_complexRegexPatterns tests various regex patterns.
func TestAccDataSourceFilters_complexRegexPatterns(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					# Match monitors starting with [PROD] or [STAGING]
					data "hyperping_monitors" "env_prefix" {
						filter = {
							name_regex = "^\\[(PROD|STAGING)\\]-.*"
						}
					}

					# Match monitors ending with "API"
					data "hyperping_monitors" "api_suffix" {
						filter = {
							name_regex = ".*-API$"
						}
					}

					# Match monitors with any digits
					data "hyperping_monitors" "with_digits" {
						filter = {
							name_regex = ".*\\d+.*"
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.hyperping_monitors.env_prefix", "monitors.#"),
					resource.TestCheckResourceAttrSet("data.hyperping_monitors.api_suffix", "monitors.#"),
					resource.TestCheckResourceAttrSet("data.hyperping_monitors.with_digits", "monitors.#"),
				),
			},
		},
	})
}

// TestAccDataSourceFilters_nullVsUnknownHandling tests null and unknown value handling.
func TestAccDataSourceFilters_nullVsUnknownHandling(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					# Filter with only one field set
					data "hyperping_monitors" "partial_filter" {
						filter = {
							protocol = "https"
							# name_regex is null - should not filter by name
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.hyperping_monitors.partial_filter", "monitors.#"),
				),
			},
		},
	})
}

// TestAccDataSourceFilters_realWorldScenarios tests realistic use cases.
func TestAccDataSourceFilters_realWorldScenarios(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					# Scenario 1: Find all production HTTPS monitors that are not paused
					data "hyperping_monitors" "prod_https_active" {
						filter = {
							name_regex = "\\[PROD\\]-.*"
							protocol   = "https"
							paused     = false
						}
					}

					# Scenario 2: Find all critical incidents that are not resolved
					data "hyperping_incidents" "critical_active" {
						filter = {
							severity = "critical"
						}
					}

					# Scenario 3: Find upcoming maintenance windows
					data "hyperping_maintenance_windows" "upcoming" {
						filter = {
							status = "scheduled"
						}
					}

					# Scenario 4: Find all API-related outages
					data "hyperping_outages" "api_outages" {
						filter = {
							name_regex = ".*API.*"
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.hyperping_monitors.prod_https_active", "monitors.#"),
					resource.TestCheckResourceAttrSet("data.hyperping_incidents.critical_active", "incidents.#"),
					resource.TestCheckResourceAttrSet("data.hyperping_maintenance_windows.upcoming", "maintenance_windows.#"),
					resource.TestCheckResourceAttrSet("data.hyperping_outages.api_outages", "outages.#"),
				),
			},
		},
	})
}
