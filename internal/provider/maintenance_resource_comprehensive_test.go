// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	tfresource "github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// =============================================================================
// Test 1: notification_option Immediate
// =============================================================================

// TestAccMaintenanceResource_notificationImmediate tests notification_option
// "immediate" mode and verifies notification_minutes is properly set with defaults.
func TestAccMaintenanceResource_notificationImmediate(t *testing.T) {
	startStr, endStr := generateMaintenanceTimeRange(24, 2)

	fixture := &maintenanceTestFixture{
		UUID:                "mw_immediate_123",
		Name:                "Immediate Notification Test",
		Title:               "Immediate Notification Test",
		Text:                "Testing immediate notification option",
		StartDate:           startStr,
		EndDate:             endStr,
		Monitors:            []string{"mon_123"},
		NotificationOption:  "immediate",
		NotificationMinutes: 60, // API default
	}

	server := newSimpleMaintenanceServer(fixture)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			// Create with notification_option="immediate"
			{
				Config: testAccMaintenanceResourceConfigWithNotificationImmediate(server.URL, fixture.Name, fixture.Title, fixture.Text, startStr, endStr, fixture.Monitors),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_maintenance.test", "notification_option", "immediate"),
					// notification_minutes should be computed default (60)
					tfresource.TestCheckResourceAttr("hyperping_maintenance.test", "notification_minutes", "60"),
				),
			},
		},
	})
}

// =============================================================================
// Test 2: notification_minutes Edge Cases
// =============================================================================

// TestAccMaintenanceResource_notificationMinutesMinimum tests notification_minutes
// with minimum value (1 minute) to verify proper handling of edge case.
func TestAccMaintenanceResource_notificationMinutesMinimum(t *testing.T) {
	startStr, endStr := generateMaintenanceTimeRange(48, 4)

	fixture := &maintenanceTestFixture{
		UUID:                "mw_minutes_min_123",
		Name:                "Notification Minutes Min Test",
		Title:               "Notification Minutes Min Test",
		Text:                "Testing notification_minutes=1",
		StartDate:           startStr,
		EndDate:             endStr,
		Monitors:            []string{"mon_123"},
		NotificationOption:  "scheduled",
		NotificationMinutes: 1,
	}

	server := newSimpleMaintenanceServer(fixture)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			// Test with notification_minutes=1 (minimum)
			{
				Config: testAccMaintenanceResourceConfigWithNotificationScheduled(server.URL, fixture.Name, fixture.Title, fixture.Text, startStr, endStr, fixture.Monitors, 1),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_maintenance.test", "notification_minutes", "1"),
				),
			},
		},
	})
}

// TestAccMaintenanceResource_notificationMinutesMaximum tests notification_minutes
// with maximum value (10080 = 7 days) to verify proper handling of large values.
func TestAccMaintenanceResource_notificationMinutesMaximum(t *testing.T) {
	startStr, endStr := generateMaintenanceTimeRange(96, 4)

	fixture := &maintenanceTestFixture{
		UUID:                "mw_minutes_max_123",
		Name:                "Notification Minutes Max Test",
		Title:               "Notification Minutes Max Test",
		Text:                "Testing notification_minutes=10080",
		StartDate:           startStr,
		EndDate:             endStr,
		Monitors:            []string{"mon_123"},
		NotificationOption:  "scheduled",
		NotificationMinutes: 10080,
	}

	server := newSimpleMaintenanceServer(fixture)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			// Test with notification_minutes=10080 (7 days, maximum common value)
			{
				Config: testAccMaintenanceResourceConfigWithNotificationScheduled(server.URL, fixture.Name, fixture.Title, fixture.Text, startStr, endStr, fixture.Monitors, 10080),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_maintenance.test", "notification_minutes", "10080"),
				),
			},
		},
	})
}

// =============================================================================
// Test 3: text Field Special Characters
// =============================================================================

// TestAccMaintenanceResource_textMarkdown tests the text field with markdown formatting.
func TestAccMaintenanceResource_textMarkdown(t *testing.T) {
	startStr, endStr := generateMaintenanceTimeRange(24, 2)

	markdownText := `## Maintenance Window

We will be performing routine maintenance:

- Database optimization
- Cache clearing
- Security patches

**Expected downtime:** 2 hours`

	fixture := &maintenanceTestFixture{
		UUID:      "mw_markdown_123",
		Name:      "Text Markdown Test",
		Title:     "Text Markdown Test",
		Text:      markdownText,
		StartDate: startStr,
		EndDate:   endStr,
		Monitors:  []string{"mon_123"},
	}

	server := newSimpleMaintenanceServer(fixture)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config: testAccMaintenanceResourceConfigWithText(server.URL, fixture.Name, fixture.Title, markdownText, startStr, endStr, fixture.Monitors),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_maintenance.test", "text", markdownText),
				),
			},
		},
	})
}

// TestAccMaintenanceResource_textUnicode tests the text field with Unicode characters.
func TestAccMaintenanceResource_textUnicode(t *testing.T) {
	startStr, endStr := generateMaintenanceTimeRange(48, 2)

	unicodeText := "数据库维护窗口 - 预计停机时间：2小时 🔧"

	fixture := &maintenanceTestFixture{
		UUID:      "mw_unicode_123",
		Name:      "Text Unicode Test",
		Title:     "Text Unicode Test",
		Text:      unicodeText,
		StartDate: startStr,
		EndDate:   endStr,
		Monitors:  []string{"mon_123"},
	}

	server := newSimpleMaintenanceServer(fixture)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config: testAccMaintenanceResourceConfigWithText(server.URL, fixture.Name, fixture.Title, unicodeText, startStr, endStr, fixture.Monitors),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_maintenance.test", "text", unicodeText),
				),
			},
		},
	})
}

// TestAccMaintenanceResource_textLongContent tests the text field with very long content (5000+ chars).
func TestAccMaintenanceResource_textLongContent(t *testing.T) {
	startStr, endStr := generateMaintenanceTimeRange(72, 2)

	longText := generateLongMaintenanceText(5000)

	fixture := &maintenanceTestFixture{
		UUID:      "mw_long_123",
		Name:      "Text Long Test",
		Title:     "Text Long Test",
		Text:      longText,
		StartDate: startStr,
		EndDate:   endStr,
		Monitors:  []string{"mon_123"},
	}

	server := newSimpleMaintenanceServer(fixture)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config: testAccMaintenanceResourceConfigWithText(server.URL, fixture.Name, fixture.Title, longText, startStr, endStr, fixture.Monitors),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_maintenance.test", "text", longText),
				),
			},
		},
	})
}

// =============================================================================
// Test Configuration Helpers
// =============================================================================

func testAccMaintenanceResourceConfigWithNotificationImmediate(baseURL, name, title, text, startDate, endDate string, monitors []string) string {
	monitorsList := `["` + monitors[0] + `"]`
	if len(monitors) > 1 {
		monitorsList = `[` + fmt.Sprintf("%q", monitors[0])
		for _, m := range monitors[1:] {
			monitorsList += fmt.Sprintf(", %q", m)
		}
		monitorsList += `]`
	}

	return fmt.Sprintf(`
provider "hyperping" {
  api_key  = "test_api_key"
  base_url = %[1]q
}

resource "hyperping_maintenance" "test" {
  name                = %[2]q
  title               = %[3]q
  text                = %[4]q
  start_date          = %[5]q
  end_date            = %[6]q
  monitors            = %[7]s
  notification_option = "immediate"
}
`, baseURL, name, title, text, startDate, endDate, monitorsList)
}

func testAccMaintenanceResourceConfigWithNotificationScheduled(baseURL, name, title, text, startDate, endDate string, monitors []string, notificationMinutes int) string {
	monitorsList := `["` + monitors[0] + `"]`
	if len(monitors) > 1 {
		monitorsList = `[` + fmt.Sprintf("%q", monitors[0])
		for _, m := range monitors[1:] {
			monitorsList += fmt.Sprintf(", %q", m)
		}
		monitorsList += `]`
	}

	return fmt.Sprintf(`
provider "hyperping" {
  api_key  = "test_api_key"
  base_url = %[1]q
}

resource "hyperping_maintenance" "test" {
  name                 = %[2]q
  title                = %[3]q
  text                 = %[4]q
  start_date           = %[5]q
  end_date             = %[6]q
  monitors             = %[7]s
  notification_option  = "scheduled"
  notification_minutes = %[8]d
}
`, baseURL, name, title, text, startDate, endDate, monitorsList, notificationMinutes)
}

func testAccMaintenanceResourceConfigWithText(baseURL, name, title, text, startDate, endDate string, monitors []string) string {
	monitorsList := `["` + monitors[0] + `"]`
	if len(monitors) > 1 {
		monitorsList = `[` + fmt.Sprintf("%q", monitors[0])
		for _, m := range monitors[1:] {
			monitorsList += fmt.Sprintf(", %q", m)
		}
		monitorsList += `]`
	}

	return fmt.Sprintf(`
provider "hyperping" {
  api_key  = "test_api_key"
  base_url = %[1]q
}

resource "hyperping_maintenance" "test" {
  name       = %[2]q
  title      = %[3]q
  text       = %[4]q
  start_date = %[5]q
  end_date   = %[6]q
  monitors   = %[7]s
}
`, baseURL, name, title, text, startDate, endDate, monitorsList)
}

// =============================================================================
// Utility Functions
// =============================================================================

func generateLongMaintenanceText(length int) string {
	result := "Maintenance description: "
	for i := len(result); i < length; i++ {
		result += "A"
	}
	return result
}
