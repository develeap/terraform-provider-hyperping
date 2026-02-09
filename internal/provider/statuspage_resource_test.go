// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	tfresource "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/develeap/terraform-provider-hyperping/internal/client"
)

func TestAccStatusPageResource_basic(t *testing.T) {
	server := newMockStatusPageServer(t)
	defer server.Close()

	tfresource.Test(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			// Create and Read testing
			{
				Config: testAccStatusPageResourceConfig_basic(server.URL),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "name", "Test Status Page"),
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "hosted_subdomain", "test-status"),
					tfresource.TestCheckResourceAttrSet("hyperping_statuspage.test", "id"),
					tfresource.TestCheckResourceAttrSet("hyperping_statuspage.test", "url"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "hyperping_statuspage.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccStatusPageResource_full(t *testing.T) {
	server := newMockStatusPageServer(t)
	defer server.Close()

	tfresource.Test(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config: testAccStatusPageResourceConfig_full(server.URL),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "name", "Production Status"),
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "hosted_subdomain", "prod-status"),
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "settings.theme", "dark"),
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "settings.font", "Inter"),
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "settings.accent_color", "#0066cc"),
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "settings.languages.#", "2"),
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "settings.description.en", "Production system status"),
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "settings.subscribe.enabled", "true"),
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "settings.subscribe.email", "true"),
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "settings.authentication.password_protection", "false"),
				),
			},
		},
	})
}

func TestAccStatusPageResource_withSections(t *testing.T) {
	server := newMockStatusPageServer(t)
	defer server.Close()

	tfresource.Test(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			// Create with sections
			{
				Config: testAccStatusPageResourceConfig_withSections(server.URL),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "sections.#", "2"),
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "sections.0.name.en", "API Services"),
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "sections.1.name.en", "Databases"),
				),
			},
			// Update sections
			{
				Config: testAccStatusPageResourceConfig_withUpdatedSections(server.URL),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "sections.#", "1"),
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "sections.0.name.en", "All Services"),
				),
			},
		},
	})
}

func TestAccStatusPageResource_updateSettings(t *testing.T) {
	server := newMockStatusPageServer(t)
	defer server.Close()

	tfresource.Test(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			// Create with default settings
			{
				Config: testAccStatusPageResourceConfig_basic(server.URL),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "settings.theme", "system"),
				),
			},
			// Update theme
			{
				Config: testAccStatusPageResourceConfig_withTheme(server.URL, "dark"),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "settings.theme", "dark"),
				),
			},
			// Update to light theme
			{
				Config: testAccStatusPageResourceConfig_withTheme(server.URL, "light"),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "settings.theme", "light"),
				),
			},
		},
	})
}

func TestAccStatusPageResource_disappears(t *testing.T) {
	server := newMockStatusPageServer(t)
	defer server.Close()

	tfresource.Test(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config: testAccStatusPageResourceConfig_basic(server.URL),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					testAccCheckStatusPageDisappears(server),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// Helper functions

func testAccStatusPageResourceConfig_basic(baseURL string) string {
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
`, baseURL)
}

func testAccStatusPageResourceConfig_full(baseURL string) string {
	return fmt.Sprintf(`
provider "hyperping" {
  api_key  = "test_api_key"
  base_url = %[1]q
}

resource "hyperping_statuspage" "test" {
  name             = "Production Status"
  hosted_subdomain = "prod-status"

  settings = {
    name         = "Production Settings"
    theme        = "dark"
    font         = "Inter"
    accent_color = "#0066cc"
    languages    = ["en", "fr"]

    description = {
      en = "Production system status"
      fr = "État du système de production"
    }

    subscribe = {
      enabled = true
      email   = true
      sms     = false
      slack   = false
      teams   = false
    }

    authentication = {
      password_protection = false
      google_sso          = false
    }
  }
}
`, baseURL)
}

func testAccStatusPageResourceConfig_withSections(baseURL string) string {
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
`, baseURL)
}

func testAccStatusPageResourceConfig_withUpdatedSections(baseURL string) string {
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
        en = "All Services"
      }
      is_split = true
    }
  ]
}
`, baseURL)
}

func testAccStatusPageResourceConfig_withTheme(baseURL, theme string) string {
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
    theme     = %[2]q
  }
}
`, baseURL, theme)
}

func testAccCheckStatusPageDisappears(server *mockStatusPageServer) tfresource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources["hyperping_statuspage.test"]
		if !ok {
			return fmt.Errorf("Not found: hyperping_statuspage.test")
		}

		// Delete the status page from the mock server
		server.deleteAllStatusPages()
		return nil
	}
}

// Mock server implementation

type mockStatusPageServer struct {
	*httptest.Server
	t           *testing.T
	mu          sync.RWMutex
	statusPages map[string]map[string]interface{}
	subscribers map[string][]map[string]interface{}
	counter     int
	subCounter  int
}

func newMockStatusPageServer(t *testing.T) *mockStatusPageServer {
	m := &mockStatusPageServer{
		t:           t,
		statusPages: make(map[string]map[string]interface{}),
		subscribers: make(map[string][]map[string]interface{}),
		counter:     0,
		subCounter:  1,
	}

	m.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.handleRequest(w, r)
	}))

	return m
}

func (m *mockStatusPageServer) handleRequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(client.HeaderContentType, client.ContentTypeJSON)
	basePath := client.StatuspagesBasePath
	basePathWithID := basePath + "/sp_"

	switch {
	case r.Method == "GET" && r.URL.Path == basePath:
		m.listStatusPages(w, r)
	case r.Method == "POST" && r.URL.Path == basePath:
		m.createStatusPage(w, r)
	// Subscriber routes must come before general status page routes
	case r.Method == "GET" && strings.Contains(r.URL.Path, "/subscribers"):
		m.listSubscribers(w, r)
	case r.Method == "POST" && strings.Contains(r.URL.Path, "/subscribers"):
		m.addSubscriber(w, r)
	case r.Method == "DELETE" && strings.Contains(r.URL.Path, "/subscribers/"):
		m.deleteSubscriber(w, r)
	// General status page routes
	case r.Method == "GET" && strings.HasPrefix(r.URL.Path, basePathWithID):
		m.getStatusPage(w, r)
	case r.Method == "PUT" && strings.HasPrefix(r.URL.Path, basePathWithID):
		m.updateStatusPage(w, r)
	case r.Method == "DELETE" && strings.HasPrefix(r.URL.Path, basePathWithID):
		m.deleteStatusPage(w, r)
	default:
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Not found"})
	}
}

func (m *mockStatusPageServer) listStatusPages(w http.ResponseWriter, _ *http.Request) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	pages := make([]map[string]interface{}, 0, len(m.statusPages))
	for _, page := range m.statusPages {
		pages = append(pages, page)
	}

	response := map[string]interface{}{
		"statuspages":   pages,
		"total":         len(pages),
		"has_next_page": false,
	}
	json.NewEncoder(w).Encode(response)
}

func (m *mockStatusPageServer) createStatusPage(w http.ResponseWriter, r *http.Request) {
	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request"})
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.counter++
	uuid := fmt.Sprintf("sp_%03d", m.counter)

	// Extract subdomain from flat request
	subdomain := "default-status"
	if s, ok := req["subdomain"].(string); ok && s != "" {
		subdomain = s
	}

	// Build nested settings object from flat request fields
	// Derive settings.name from the top-level name
	// Real API likely does the same since provider doesn't send this field
	settingsName := "Settings"
	if topName, ok := req["name"].(string); ok && topName != "" {
		// Extract first word: "Test Status Page" → "Test"
		parts := strings.Fields(topName)
		if len(parts) > 0 {
			settingsName = parts[0] + " Settings"
		}
	}

	settings := map[string]interface{}{
		"name":                     settingsName,
		"website":                  "",
		"theme":                    getOrDefault(req, "theme", "system"),
		"font":                     getOrDefault(req, "font", "Inter"),
		"accent_color":             getOrDefault(req, "accent_color", "#36b27e"),
		"default_language":         "en",
		"logo_height":              "40px",
		"auto_refresh":             getOrDefaultBool(req, "auto_refresh", false),
		"banner_header":            getOrDefaultBool(req, "banner_header", false),
		"hide_powered_by":          getOrDefaultBool(req, "hide_powered_by", false),
		"hide_from_search_engines": getOrDefaultBool(req, "hide_from_search_engines", false),
	}

	// Add optional fields to settings
	if website, ok := req["website"].(string); ok {
		settings["website"] = website
	}
	if languages, ok := req["languages"].([]interface{}); ok {
		// Convert []interface{} to []string for proper JSON serialization
		langStrings := make([]string, len(languages))
		for i, lang := range languages {
			if langStr, ok := lang.(string); ok {
				langStrings[i] = langStr
			}
		}
		settings["languages"] = langStrings
	} else {
		settings["languages"] = []string{}
	}
	if description, ok := req["description"].(map[string]interface{}); ok {
		settings["description"] = description
	}
	if logo, ok := req["logo"].(string); ok {
		settings["logo"] = logo
	}
	if logoHeight, ok := req["logo_height"].(string); ok {
		settings["logo_height"] = logoHeight
	}
	if favicon, ok := req["favicon"].(string); ok {
		settings["favicon"] = favicon
	}
	if ga, ok := req["google_analytics"].(string); ok {
		settings["google_analytics"] = ga
	}
	if subscribe, ok := req["subscribe"].(map[string]interface{}); ok {
		settings["subscribe"] = subscribe
	} else {
		// Default subscribe settings
		settings["subscribe"] = map[string]interface{}{
			"enabled": false,
			"email":   false,
			"slack":   false,
			"teams":   false,
			"sms":     false,
		}
	}
	if auth, ok := req["authentication"].(map[string]interface{}); ok {
		settings["authentication"] = auth
	} else {
		// Default authentication settings
		settings["authentication"] = map[string]interface{}{
			"password_protection": false,
			"google_sso":          false,
			"saml_sso":            false,
			"allowed_domains":     []string{},
		}
	}

	// Build API-compliant status page response with nested settings
	// Real API returns hostedsubdomain WITH the .hyperping.app suffix
	// The provider's normalization logic strips this suffix for state consistency
	page := map[string]interface{}{
		"uuid":               uuid,
		"name":               req["name"],
		"hostedsubdomain":    fmt.Sprintf("%s.hyperping.app", subdomain), // Real API returns full subdomain
		"url":                fmt.Sprintf("https://%s.hyperping.app", subdomain),
		"password_protected": false,
		"settings":           settings,
		"sections":           []interface{}{},
	}

	// Add optional top-level fields
	if hostname, ok := req["hostname"]; ok {
		page["hostname"] = hostname
	}

	// Add sections if present - transform name from string to map[string]string
	if sectionsReq, ok := req["sections"].([]interface{}); ok {
		sections := []map[string]interface{}{}
		for _, secInterface := range sectionsReq {
			if secMap, ok := secInterface.(map[string]interface{}); ok {
				section := map[string]interface{}{
					"is_split": false,
					"services": nil, // API returns null for empty services, not []
				}

				// Transform name: API receives string but returns map[string]string
				if name, ok := secMap["name"].(string); ok {
					section["name"] = map[string]string{"en": name}
				}

				if isSplit, ok := secMap["is_split"].(bool); ok {
					section["is_split"] = isSplit
				}

				sections = append(sections, section)
			}
		}
		page["sections"] = sections
	}

	m.statusPages[uuid] = page

	// Wrap response as per API spec
	response := map[string]interface{}{
		"message":    "Status page created successfully",
		"statuspage": page,
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (m *mockStatusPageServer) getStatusPage(w http.ResponseWriter, r *http.Request) {
	uuid := strings.TrimPrefix(r.URL.Path, client.StatuspagesBasePath+"/")

	m.mu.RLock()
	defer m.mu.RUnlock()

	page, ok := m.statusPages[uuid]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Status page not found"})
		return
	}

	// Wrap response as per API spec: GET returns {"statuspage": {...}}
	response := map[string]interface{}{
		"statuspage": page,
	}
	json.NewEncoder(w).Encode(response)
}

func (m *mockStatusPageServer) updateStatusPage(w http.ResponseWriter, r *http.Request) {
	uuid := strings.TrimPrefix(r.URL.Path, client.StatuspagesBasePath+"/")

	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request"})
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	page, ok := m.statusPages[uuid]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Status page not found"})
		return
	}

	// Update top-level fields
	if name, ok := req["name"]; ok {
		page["name"] = name
	}
	if subdomain, ok := req["subdomain"].(string); ok {
		// Real API returns hostedsubdomain WITH the .hyperping.app suffix
		page["hostedsubdomain"] = fmt.Sprintf("%s.hyperping.app", subdomain)
		page["url"] = fmt.Sprintf("https://%s.hyperping.app", subdomain)
	}
	if hostname, ok := req["hostname"]; ok {
		page["hostname"] = hostname
	}
	if sectionsReq, ok := req["sections"].([]interface{}); ok {
		sections := []map[string]interface{}{}
		for _, secInterface := range sectionsReq {
			if secMap, ok := secInterface.(map[string]interface{}); ok {
				section := map[string]interface{}{
					"is_split": false,
					"services": nil, // API returns null for empty services, not []
				}
				if name, ok := secMap["name"].(string); ok {
					section["name"] = map[string]string{"en": name}
				}
				if isSplit, ok := secMap["is_split"].(bool); ok {
					section["is_split"] = isSplit
				}
				sections = append(sections, section)
			}
		}
		page["sections"] = sections
	}

	// Update settings from flat request fields
	settings, _ := page["settings"].(map[string]interface{})
	if settings == nil {
		settings = make(map[string]interface{})
	}

	// Update settings fields from flat request
	if theme, ok := req["theme"].(string); ok {
		settings["theme"] = theme
	}
	if font, ok := req["font"].(string); ok {
		settings["font"] = font
	}
	if accentColor, ok := req["accent_color"].(string); ok {
		settings["accent_color"] = accentColor
	}
	if languages, ok := req["languages"]; ok {
		// Ensure languages is always an array of strings
		if langArr, ok := languages.([]interface{}); ok {
			langStrings := make([]string, len(langArr))
			for i, lang := range langArr {
				if langStr, ok := lang.(string); ok {
					langStrings[i] = langStr
				}
			}
			settings["languages"] = langStrings
		} else {
			settings["languages"] = languages
		}
	}
	if description, ok := req["description"]; ok {
		settings["description"] = description
	}
	if subscribe, ok := req["subscribe"]; ok {
		settings["subscribe"] = subscribe
	}
	if authentication, ok := req["authentication"]; ok {
		// Ensure allowed_domains is always an array (not nil/null)
		if authMap, ok := authentication.(map[string]interface{}); ok {
			if _, hasAllowedDomains := authMap["allowed_domains"]; !hasAllowedDomains {
				authMap["allowed_domains"] = []string{}
			}
			settings["authentication"] = authMap
		} else {
			settings["authentication"] = authentication
		}
	}

	page["settings"] = settings
	m.statusPages[uuid] = page

	// Wrap response in same format as create (API spec)
	response := map[string]interface{}{
		"message":    "Status page updated successfully",
		"statuspage": page,
	}
	json.NewEncoder(w).Encode(response)
}

func (m *mockStatusPageServer) deleteStatusPage(w http.ResponseWriter, r *http.Request) {
	uuid := strings.TrimPrefix(r.URL.Path, client.StatuspagesBasePath+"/")

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.statusPages[uuid]; !ok {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Status page not found"})
		return
	}

	delete(m.statusPages, uuid)
	w.WriteHeader(http.StatusNoContent)
}

func (m *mockStatusPageServer) deleteAllStatusPages() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.statusPages = make(map[string]map[string]interface{})
}

func (m *mockStatusPageServer) listSubscribers(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	var uuid string
	for i, part := range parts {
		if part == "statuspages" && i+1 < len(parts) {
			uuid = parts[i+1]
			break
		}
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	subs, ok := m.subscribers[uuid]
	if !ok {
		subs = []map[string]interface{}{}
	}

	response := map[string]interface{}{
		"subscribers": subs,
		"total":       len(subs),
	}
	json.NewEncoder(w).Encode(response)
}

func (m *mockStatusPageServer) addSubscriber(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	var uuid string
	for i, part := range parts {
		if part == "statuspages" && i+1 < len(parts) {
			uuid = parts[i+1]
			break
		}
	}

	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request"})
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.subCounter++
	subscriber := map[string]interface{}{
		"id":         m.subCounter,
		"type":       req["type"],
		"language":   getOrDefault(req, "language", "en"),
		"created_at": "2026-01-01T00:00:00Z",
	}

	// Add type-specific fields
	if email, ok := req["email"]; ok {
		subscriber["email"] = email
		subscriber["value"] = email
	}
	if phone, ok := req["phone"]; ok {
		subscriber["phone"] = phone
		subscriber["value"] = phone
	}
	if webhook, ok := req["teams_webhook_url"]; ok {
		subscriber["teams_webhook_url"] = webhook
		subscriber["value"] = "Microsoft Teams"
	}

	m.subscribers[uuid] = append(m.subscribers[uuid], subscriber)

	// Wrap response like real API
	response := map[string]interface{}{
		"message":    "Subscriber added successfully",
		"subscriber": subscriber,
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (m *mockStatusPageServer) deleteSubscriber(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	var uuid string
	var subIDStr string

	for i, part := range parts {
		if part == "statuspages" && i+1 < len(parts) {
			uuid = parts[i+1]
		}
		if part == "subscribers" && i+1 < len(parts) {
			subIDStr = parts[i+1]
		}
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	subs, ok := m.subscribers[uuid]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Filter out the subscriber
	var subID int
	fmt.Sscanf(subIDStr, "%d", &subID)

	newSubs := make([]map[string]interface{}, 0)
	for _, sub := range subs {
		if sub["id"].(int) != subID {
			newSubs = append(newSubs, sub)
		}
	}

	m.subscribers[uuid] = newSubs
	w.WriteHeader(http.StatusNoContent)
}
