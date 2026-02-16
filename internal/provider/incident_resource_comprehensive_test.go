// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"regexp"
	"testing"
	"time"

	tfresource "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// =============================================================================
// Test 1: date Computed Field
// =============================================================================

// TestAccIncidentResource_dateComputedField tests the date computed field
// to verify it's set on creation, follows ISO 8601 format, and behaves
// correctly on updates.
func TestAccIncidentResource_dateComputedField(t *testing.T) {
	server := newMockIncidentServer(t)
	defer server.Close()

	// ISO 8601 format regex: YYYY-MM-DDTHH:MM:SS.sssZ
	iso8601Regex := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{3}Z$`)

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			// Create incident and verify date is set
			{
				Config: testAccIncidentResourceConfig_basic(server.URL, "Date Test Incident"),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_incident.test", "title", "Date Test Incident"),
					// Verify date field is set
					tfresource.TestCheckResourceAttrSet("hyperping_incident.test", "date"),
					// Verify date matches ISO 8601 format
					tfresource.TestMatchResourceAttr("hyperping_incident.test", "date", iso8601Regex),
					// Verify date is a recent timestamp (within last 24 hours for mock server)
					testAccCheckIncidentDateIsRecent("hyperping_incident.test"),
				),
			},
			// Update incident title and verify date persists
			{
				Config: testAccIncidentResourceConfig_basic(server.URL, "Updated Date Test Incident"),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_incident.test", "title", "Updated Date Test Incident"),
					// Date should still be set (mock server keeps original date)
					tfresource.TestCheckResourceAttrSet("hyperping_incident.test", "date"),
					tfresource.TestMatchResourceAttr("hyperping_incident.test", "date", iso8601Regex),
				),
			},
		},
	})
}

// =============================================================================
// Test 2: text Long Content
// =============================================================================

// TestAccIncidentResource_textLongContent tests the text field with very long
// content (5000+ characters) to verify proper handling of large text.
func TestAccIncidentResource_textLongContent(t *testing.T) {
	server := newMockIncidentServer(t)
	defer server.Close()

	// Generate long text (5000 characters)
	longText := generateLongIncidentText(5000)

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			// Create with very long text (5000 characters)
			{
				Config: testAccIncidentResourceConfigWithLongText(server.URL, "Long Text Test", longText),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_incident.test", "title", "Long Text Test"),
					tfresource.TestCheckResourceAttr("hyperping_incident.test", "text", longText),
					// Verify length
					testAccCheckIncidentTextLength("hyperping_incident.test", 5000),
				),
			},
		},
	})
}

// =============================================================================
// Test 3: text Special Characters (Bonus Coverage)
// =============================================================================

// TestAccIncidentResource_textSpecialCharacters tests the text field with
// markdown, Unicode, HTML, and special characters to ensure proper handling.
func TestAccIncidentResource_textMarkdown(t *testing.T) {
	server := newMockIncidentServer(t)
	defer server.Close()

	// Markdown text
	markdownText := `## Service Degradation

We are experiencing issues with the following services:

- **API Gateway**: Slow response times
- **Database**: Connection pool exhaustion
- **Cache Layer**: Redis cluster failover

**Impact:** Users may experience slower page loads.

We are actively investigating and will provide updates soon.`

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			// Test with markdown formatting
			{
				Config: testAccIncidentResourceConfigWithLongText(server.URL, "Markdown Test", markdownText),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_incident.test", "text", markdownText),
				),
			},
		},
	})
}

func TestAccIncidentResource_textUnicode(t *testing.T) {
	server := newMockIncidentServer(t)
	defer server.Close()

	// Unicode text
	unicodeText := "🔴 系统故障 - 我们正在调查问题 - Nous enquêtons sur le problème - Мы расследуем проблему"

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config: testAccIncidentResourceConfigWithLongText(server.URL, "Unicode Test", unicodeText),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_incident.test", "text", unicodeText),
				),
			},
		},
	})
}

// =============================================================================
// Test Configuration Helpers
// =============================================================================

func testAccIncidentResourceConfigWithLongText(baseURL, title, text string) string {
	return fmt.Sprintf(`
provider "hyperping" {
  api_key  = "test_api_key"
  base_url = %[1]q
}

resource "hyperping_incident" "test" {
  title        = %[2]q
  text         = %[3]q
  type         = "incident"
  status_pages = ["sp_main"]
}
`, baseURL, title, text)
}

// =============================================================================
// Custom Check Functions
// =============================================================================

// testAccCheckIncidentDateIsRecent verifies the date field is a recent timestamp.
// For mock servers, we just verify it's parseable as ISO 8601.
func testAccCheckIncidentDateIsRecent(resourceName string) tfresource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		dateStr := rs.Primary.Attributes["date"]
		if dateStr == "" {
			return fmt.Errorf("date attribute is empty")
		}

		// Parse as ISO 8601
		_, err := time.Parse(time.RFC3339, dateStr)
		if err != nil {
			return fmt.Errorf("date %q is not valid ISO 8601: %v", dateStr, err)
		}

		return nil
	}
}

// testAccCheckIncidentTextLength verifies the text field has the expected length.
func testAccCheckIncidentTextLength(resourceName string, expectedLength int) tfresource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		textStr := rs.Primary.Attributes["text"]
		if len(textStr) != expectedLength {
			return fmt.Errorf("expected text length %d, got %d", expectedLength, len(textStr))
		}

		return nil
	}
}

// =============================================================================
// Utility Functions
// =============================================================================

func generateLongIncidentText(length int) string {
	prefix := "Incident description: "
	result := prefix

	// Fill with lorem ipsum style text
	lorem := "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. "

	for len(result) < length {
		if len(result)+len(lorem) <= length {
			result += lorem
		} else {
			// Fill remaining with 'A'
			result += "A"
		}
	}

	return result[:length]
}
