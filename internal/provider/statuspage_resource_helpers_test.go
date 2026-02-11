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

	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/develeap/terraform-provider-hyperping/internal/client"
)

// Terraform configuration generators

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
    name      = "Test Status Page"
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
    name         = "Production Status"
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
    name      = "Test Status Page"
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
    name      = "Test Status Page"
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
    name      = "Test Status Page"
    languages = ["en"]
    theme     = %[2]q
  }
}
`, baseURL, theme)
}

// Test check functions

func testAccCheckStatusPageDisappears(server *mockStatusPageServer) func(*terraform.State) error {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources["hyperping_statuspage.test"]
		if !ok {
			return fmt.Errorf("Not found: hyperping_statuspage.test")
		}

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
	case r.Method == "GET" && strings.Contains(r.URL.Path, "/subscribers"):
		m.listSubscribers(w, r)
	case r.Method == "POST" && strings.Contains(r.URL.Path, "/subscribers"):
		m.addSubscriber(w, r)
	case r.Method == "DELETE" && strings.Contains(r.URL.Path, "/subscribers/"):
		m.deleteSubscriber(w, r)
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

	subdomain := "default-status"
	if s, ok := req["subdomain"].(string); ok && s != "" {
		subdomain = s
	}

	settingsName := "Status Page"
	if topName, ok := req["name"].(string); ok && topName != "" {
		settingsName = topName
	}

	settings := map[string]interface{}{
		"name":                     settingsName,
		"website":                  "",
		"theme":                    getOrDefault(req, "theme", "system"),
		"font":                     getOrDefault(req, "font", "Inter"),
		"accent_color":             getOrDefault(req, "accent_color", "#36b27e"),
		"default_language":         "en",
		"logo_height":              "32px",
		"auto_refresh":             getOrDefaultBool(req, "auto_refresh", false),
		"banner_header":            getOrDefaultBool(req, "banner_header", true),
		"hide_powered_by":          getOrDefaultBool(req, "hide_powered_by", false),
		"hide_from_search_engines": getOrDefaultBool(req, "hide_from_search_engines", false),
	}

	if website, ok := req["website"].(string); ok {
		settings["website"] = website
	}
	if languages, ok := req["languages"].([]interface{}); ok {
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

	description := map[string]string{
		"en": "",
		"fr": "",
		"de": "",
		"ru": "",
		"nl": "",
	}
	if reqDesc, ok := req["description"].(map[string]interface{}); ok {
		for lang, val := range reqDesc {
			if valStr, ok := val.(string); ok {
				description[lang] = valStr
			}
		}
	}
	settings["description"] = description
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
		settings["authentication"] = map[string]interface{}{
			"password_protection": false,
			"google_sso":          false,
			"saml_sso":            false,
			"allowed_domains":     []string{},
		}
	}

	page := map[string]interface{}{
		"uuid":               uuid,
		"name":               req["name"],
		"hostedsubdomain":    fmt.Sprintf("%s.hyperping.app", subdomain),
		"url":                fmt.Sprintf("https://%s.hyperping.app", subdomain),
		"password_protected": false,
		"settings":           settings,
		"sections":           []interface{}{},
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
					"services": nil,
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

	m.statusPages[uuid] = page

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

	if name, ok := req["name"]; ok {
		page["name"] = name
		settings, _ := page["settings"].(map[string]interface{})
		if settings == nil {
			settings = make(map[string]interface{})
			page["settings"] = settings
		}
		settings["name"] = name
	}
	if subdomain, ok := req["subdomain"].(string); ok {
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
					"services": nil,
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

	settings, _ := page["settings"].(map[string]interface{})
	if settings == nil {
		settings = make(map[string]interface{})
	}

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
		descMap := map[string]string{
			"en": "",
			"fr": "",
			"de": "",
			"ru": "",
			"nl": "",
		}
		if reqDesc, ok := description.(map[string]interface{}); ok {
			for lang, val := range reqDesc {
				if valStr, ok := val.(string); ok {
					descMap[lang] = valStr
				}
			}
		}
		settings["description"] = descMap
	}
	if subscribe, ok := req["subscribe"]; ok {
		settings["subscribe"] = subscribe
	}
	if authentication, ok := req["authentication"]; ok {
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
