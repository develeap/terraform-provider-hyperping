// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"strings"
	"testing"

	tfresource "github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// Test 1: Custom Domain (hostname)
func TestAccStatuspageResource_customDomain(t *testing.T) {
	server := newMockStatusPageServer(t)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config: testAccStatusPageResourceConfig_customDomain(server.URL, "status.example.com"),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "name", "Custom Domain Status"),
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "hostname", "status.example.com"),
					tfresource.TestCheckResourceAttrSet("hyperping_statuspage.test", "id"),
				),
			},
			{
				Config: testAccStatusPageResourceConfig_customDomain(server.URL, "status2.example.com"),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "hostname", "status2.example.com"),
				),
			},
			{
				Config: testAccStatusPageResourceConfig_basic(server.URL),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "name", "Test Status Page"),
				),
			},
		},
	})
}

// Test 2: Password Protection
func TestAccStatuspageResource_passwordProtection(t *testing.T) {
	server := newMockStatusPageServer(t)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config: testAccStatusPageResourceConfig_passwordProtection(server.URL, "secret123", true),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "password", "secret123"),
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "settings.authentication.password_protection", "true"),
				),
			},
			{
				Config: testAccStatusPageResourceConfig_passwordProtection(server.URL, "newsecret456", true),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "password", "newsecret456"),
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "settings.authentication.password_protection", "true"),
				),
			},
			{
				Config: testAccStatusPageResourceConfig_passwordProtection(server.URL, "", false),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "settings.authentication.password_protection", "false"),
				),
			},
		},
	})
}

// Test 3: Branding Settings
func TestAccStatuspageResource_brandingSettings(t *testing.T) {
	server := newMockStatusPageServer(t)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config: testAccStatusPageResourceConfig_branding(
					server.URL,
					"https://example.com/logo.png",
					"40px",
					"https://example.com/favicon.ico",
					true,
				),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "settings.logo", "https://example.com/logo.png"),
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "settings.logo_height", "40px"),
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "settings.favicon", "https://example.com/favicon.ico"),
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "settings.hide_powered_by", "true"),
				),
			},
			{
				Config: testAccStatusPageResourceConfig_branding(
					server.URL,
					"https://example.com/newlogo.png",
					"50px",
					"https://example.com/favicon.ico",
					true,
				),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "settings.logo", "https://example.com/newlogo.png"),
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "settings.logo_height", "50px"),
				),
			},
		},
	})
}

// Test 4: SEO and Analytics
func TestAccStatuspageResource_seoAndAnalytics(t *testing.T) {
	server := newMockStatusPageServer(t)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config: testAccStatusPageResourceConfig_seoAnalytics(server.URL, true, "UA-123456-1"),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "settings.hide_from_search_engines", "true"),
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "settings.google_analytics", "UA-123456-1"),
				),
			},
			{
				Config: testAccStatusPageResourceConfig_seoAnalytics(server.URL, false, "UA-789012-3"),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "settings.hide_from_search_engines", "false"),
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "settings.google_analytics", "UA-789012-3"),
				),
			},
		},
	})
}

// Test 5: Auto-refresh and Banner
func TestAccStatuspageResource_autoRefreshBanner(t *testing.T) {
	server := newMockStatusPageServer(t)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config: testAccStatusPageResourceConfig_autoRefreshBanner(server.URL, true, true),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "settings.auto_refresh", "true"),
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "settings.banner_header", "true"),
				),
			},
			{
				Config: testAccStatusPageResourceConfig_autoRefreshBanner(server.URL, false, false),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "settings.auto_refresh", "false"),
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "settings.banner_header", "false"),
				),
			},
		},
	})
}

// Test 6: Subscription Methods
func TestAccStatuspageResource_subscriptionMethods(t *testing.T) {
	server := newMockStatusPageServer(t)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config: testAccStatusPageResourceConfig_subscriptionMethods(
					server.URL,
					true, // email
					true, // sms
					true, // slack
					true, // teams
				),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "settings.subscribe.enabled", "true"),
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "settings.subscribe.email", "true"),
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "settings.subscribe.sms", "true"),
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "settings.subscribe.slack", "true"),
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "settings.subscribe.teams", "true"),
				),
			},
			{
				Config: testAccStatusPageResourceConfig_subscriptionMethods(
					server.URL,
					true,  // email
					false, // sms
					true,  // slack
					false, // teams
				),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "settings.subscribe.email", "true"),
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "settings.subscribe.sms", "false"),
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "settings.subscribe.slack", "true"),
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "settings.subscribe.teams", "false"),
				),
			},
		},
	})
}

// Test 7: SSO Authentication
func TestAccStatuspageResource_ssoAuthentication(t *testing.T) {
	server := newMockStatusPageServer(t)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config: testAccStatusPageResourceConfig_ssoAuthentication(
					server.URL,
					true,
					true,
					[]string{"example.com", "test.com"},
				),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "settings.authentication.google_sso", "true"),
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "settings.authentication.saml_sso", "true"),
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "settings.authentication.allowed_domains.#", "2"),
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "settings.authentication.allowed_domains.0", "example.com"),
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "settings.authentication.allowed_domains.1", "test.com"),
				),
			},
			{
				Config: testAccStatusPageResourceConfig_ssoAuthentication(
					server.URL,
					false,
					true,
					[]string{"newdomain.com"},
				),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "settings.authentication.google_sso", "false"),
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "settings.authentication.saml_sso", "true"),
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "settings.authentication.allowed_domains.#", "1"),
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "settings.authentication.allowed_domains.0", "newdomain.com"),
				),
			},
		},
	})
}

// Test 8: Settings Name and Website
func TestAccStatuspageResource_settingsNameWebsite(t *testing.T) {
	server := newMockStatusPageServer(t)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config: testAccStatusPageResourceConfig_settingsNameWebsite(
					server.URL,
					"My Status Page",
					"https://example.com",
				),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "settings.name", "My Status Page"),
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "settings.website", "https://example.com"),
				),
			},
			{
				Config: testAccStatusPageResourceConfig_settingsNameWebsite(
					server.URL,
					"Updated Status Page",
					"https://newexample.com",
				),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "settings.name", "Updated Status Page"),
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "settings.website", "https://newexample.com"),
				),
			},
		},
	})
}

// Test 9: Sections with Services
func TestAccStatuspageResource_sectionsWithServices(t *testing.T) {
	server := newMockStatusPageServer(t)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config: testAccStatusPageResourceConfig_sectionsWithServices(server.URL),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "sections.#", "1"),
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "sections.0.is_split", "true"),
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "sections.0.services.#", "2"),
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "sections.0.services.0.name.en", "API Service"),
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "sections.0.services.0.is_group", "true"),
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "sections.0.services.0.show_uptime", "true"),
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "sections.0.services.0.show_response_times", "true"),
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "sections.0.services.1.name.en", "Database Service"),
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "sections.0.services.1.is_group", "false"),
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "sections.0.services.1.show_uptime", "false"),
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "sections.0.services.1.show_response_times", "false"),
				),
			},
		},
	})
}

// Test 10: Default Language and Fonts
func TestAccStatuspageResource_defaultLanguageFonts(t *testing.T) {
	server := newMockStatusPageServer(t)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config: testAccStatusPageResourceConfig_defaultLanguageFonts(server.URL, "es", "Roboto"),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "settings.default_language", "es"),
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "settings.font", "Roboto"),
				),
			},
			{
				Config: testAccStatusPageResourceConfig_defaultLanguageFonts(server.URL, "fr", "Open Sans"),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "settings.default_language", "fr"),
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "settings.font", "Open Sans"),
				),
			},
			{
				Config: testAccStatusPageResourceConfig_defaultLanguageFonts(server.URL, "en", "Inter"),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "settings.default_language", "en"),
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "settings.font", "Inter"),
				),
			},
		},
	})
}

// Config helper functions

func testAccStatusPageResourceConfig_customDomain(baseURL, hostname string) string {
	return testAccStatusPageProviderConfig(baseURL) + `
resource "hyperping_statuspage" "test" {
  name             = "Custom Domain Status"
  hosted_subdomain = "custom-domain-test"
  hostname         = "` + hostname + `"

  settings = {
    name      = "Custom Domain Status"
    languages = ["en"]
  }
}
`
}

func testAccStatusPageResourceConfig_passwordProtection(baseURL, password string, enabled bool) string {
	passwordField := ""
	if password != "" {
		passwordField = `password = "` + password + `"`
	}

	return testAccStatusPageProviderConfig(baseURL) + `
resource "hyperping_statuspage" "test" {
  name             = "Password Protected Status"
  hosted_subdomain = "password-test"
  ` + passwordField + `

  settings = {
    name      = "Password Protected Status"
    languages = ["en"]

    authentication = {
      password_protection = ` + boolToString(enabled) + `
    }
  }
}
`
}

func testAccStatusPageResourceConfig_branding(baseURL, logo, logoHeight, favicon string, hidePoweredBy bool) string {
	return testAccStatusPageProviderConfig(baseURL) + `
resource "hyperping_statuspage" "test" {
  name             = "Branded Status Page"
  hosted_subdomain = "branding-test"

  settings = {
    name           = "Branded Status Page"
    languages      = ["en"]
    logo           = "` + logo + `"
    logo_height    = "` + logoHeight + `"
    favicon        = "` + favicon + `"
    hide_powered_by = ` + boolToString(hidePoweredBy) + `
  }
}
`
}

func testAccStatusPageResourceConfig_seoAnalytics(baseURL string, hideFromSearch bool, gaTrackingID string) string {
	return testAccStatusPageProviderConfig(baseURL) + `
resource "hyperping_statuspage" "test" {
  name             = "SEO Analytics Status"
  hosted_subdomain = "seo-test"

  settings = {
    name                     = "SEO Analytics Status"
    languages                = ["en"]
    hide_from_search_engines = ` + boolToString(hideFromSearch) + `
    google_analytics         = "` + gaTrackingID + `"
  }
}
`
}

func testAccStatusPageResourceConfig_autoRefreshBanner(baseURL string, autoRefresh, bannerHeader bool) string {
	return testAccStatusPageProviderConfig(baseURL) + `
resource "hyperping_statuspage" "test" {
  name             = "Auto Refresh Status"
  hosted_subdomain = "autorefresh-test"

  settings = {
    name          = "Auto Refresh Status"
    languages     = ["en"]
    auto_refresh  = ` + boolToString(autoRefresh) + `
    banner_header = ` + boolToString(bannerHeader) + `
  }
}
`
}

func testAccStatusPageResourceConfig_subscriptionMethods(baseURL string, email, sms, slack, teams bool) string {
	return testAccStatusPageProviderConfig(baseURL) + `
resource "hyperping_statuspage" "test" {
  name             = "Subscription Methods Status"
  hosted_subdomain = "subscription-test"

  settings = {
    name      = "Subscription Methods Status"
    languages = ["en"]

    subscribe = {
      enabled = true
      email   = ` + boolToString(email) + `
      sms     = ` + boolToString(sms) + `
      slack   = ` + boolToString(slack) + `
      teams   = ` + boolToString(teams) + `
    }
  }
}
`
}

func testAccStatusPageResourceConfig_ssoAuthentication(baseURL string, googleSSO, samlSSO bool, allowedDomains []string) string {
	domainsJSON := "[]"
	if len(allowedDomains) > 0 {
		domains := make([]string, len(allowedDomains))
		for i := range allowedDomains {
			domains[i] = `"` + allowedDomains[i] + `"`
		}
		domainsJSON = `[` + strings.Join(domains, ", ") + `]`
	}

	return testAccStatusPageProviderConfig(baseURL) + `
resource "hyperping_statuspage" "test" {
  name             = "SSO Status Page"
  hosted_subdomain = "sso-test"

  settings = {
    name      = "SSO Status Page"
    languages = ["en"]

    authentication = {
      google_sso      = ` + boolToString(googleSSO) + `
      saml_sso        = ` + boolToString(samlSSO) + `
      allowed_domains = ` + domainsJSON + `
    }
  }
}
`
}

func testAccStatusPageResourceConfig_settingsNameWebsite(baseURL, settingsName, website string) string {
	return testAccStatusPageProviderConfig(baseURL) + `
resource "hyperping_statuspage" "test" {
  name             = "Settings Name Website"
  hosted_subdomain = "settings-test"

  settings = {
    name      = "` + settingsName + `"
    website   = "` + website + `"
    languages = ["en"]
  }
}
`
}

func testAccStatusPageResourceConfig_sectionsWithServices(baseURL string) string {
	return testAccStatusPageProviderConfig(baseURL) + `
resource "hyperping_statuspage" "test" {
  name             = "Sections With Services"
  hosted_subdomain = "services-test"

  settings = {
    name      = "Sections With Services"
    languages = ["en"]
  }

  sections = [
    {
      name = {
        en = "Services Section"
      }
      is_split = true
      services = [
        {
          uuid                = "mon_test001"
          name = {
            en = "API Service"
          }
          is_group            = true
          show_uptime         = true
          show_response_times = true
        },
        {
          uuid                = "mon_test002"
          name = {
            en = "Database Service"
          }
          is_group            = false
          show_uptime         = false
          show_response_times = false
        }
      ]
    }
  ]
}
`
}

func testAccStatusPageResourceConfig_defaultLanguageFonts(baseURL, defaultLang, font string) string {
	return testAccStatusPageProviderConfig(baseURL) + `
resource "hyperping_statuspage" "test" {
  name             = "Language Font Status"
  hosted_subdomain = "language-test"

  settings = {
    name             = "Language Font Status"
    languages        = ["en", "es", "fr"]
    default_language = "` + defaultLang + `"
    font             = "` + font + `"
  }
}
`
}

// Helper function to format provider config
func testAccStatusPageProviderConfig(baseURL string) string {
	return `
provider "hyperping" {
  api_key  = "test_api_key"
  base_url = "` + baseURL + `"
}
`
}

// Helper function to convert bool to string for HCL
func boolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
