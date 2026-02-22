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

    description = "Production system status"

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

// isSubscriberPath reports whether the URL path refers to a subscribers resource.
func isSubscriberPath(path string) bool {
	return strings.Contains(path, "/subscribers")
}

// isSubscriberDeletePath reports whether the URL path refers to a specific subscriber.
func isSubscriberDeletePath(path string) bool {
	return strings.Contains(path, "/subscribers/")
}

func (m *mockStatusPageServer) handleRequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(client.HeaderContentType, client.ContentTypeJSON)
	basePath := client.StatuspagesBasePath
	basePathWithID := basePath + "/sp_"
	isSubscriber := isSubscriberPath(r.URL.Path)

	switch {
	case r.Method == "GET" && r.URL.Path == basePath:
		m.listStatusPages(w, r)
	case r.Method == "POST" && r.URL.Path == basePath:
		m.createStatusPage(w, r)
	case isSubscriber:
		m.handleSubscriberRequest(w, r)
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

// handleSubscriberRequest dispatches subscriber-related requests (list, add, delete).
func (m *mockStatusPageServer) handleSubscriberRequest(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == "GET":
		m.listSubscribers(w, r)
	case r.Method == "POST":
		m.addSubscriber(w, r)
	case r.Method == "DELETE" && isSubscriberDeletePath(r.URL.Path):
		m.deleteSubscriber(w, r)
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

// buildMockService converts a service map from the request into the mock response format.
// For group services (is_group=true) with nested children, integer IDs are assigned to
// the children to exercise the real API behavior where nested child IDs are integers.
func buildMockService(svcMap map[string]interface{}) map[string]interface{} {
	isGroup := getOrDefaultBool(svcMap, "is_group", false)
	service := map[string]interface{}{
		"id":                  "svc_001",
		"uuid":                getOrDefault(svcMap, "monitor_uuid", "mon_default"),
		"is_group":            isGroup,
		"show_uptime":         getOrDefaultBool(svcMap, "show_uptime", true),
		"show_response_times": getOrDefaultBool(svcMap, "show_response_times", true),
	}

	if svcName, ok := svcMap["name"].(map[string]interface{}); ok {
		nameStrMap := make(map[string]string)
		for k, v := range svcName {
			if vStr, ok := v.(string); ok {
				nameStrMap[k] = vStr
			}
		}
		service["name"] = nameStrMap
	} else if nameShown, ok := svcMap["name_shown"].(string); ok {
		service["name"] = map[string]string{"en": nameShown}
	}

	// For group services, build nested children with integer IDs to match real API behavior.
	// The Hyperping API returns integer IDs for nested child services inside groups.
	if isGroup {
		service["services"] = buildMockGroupChildren(svcMap)
	}

	return service
}

// buildMockGroupChildren builds the nested children for a group service.
// Assigns integer IDs starting at 117122 to match real API behavior for nested children.
func buildMockGroupChildren(svcMap map[string]interface{}) []map[string]interface{} {
	children := []map[string]interface{}{}

	nestedServices, ok := svcMap["services"].([]interface{})
	if !ok || len(nestedServices) == 0 {
		return children
	}

	startID := 117122
	for i, childInterface := range nestedServices {
		childMap, ok := childInterface.(map[string]interface{})
		if !ok {
			continue
		}

		child := map[string]interface{}{
			"id":                  startID + i, // integer ID — matches real API behavior
			"is_group":            false,
			"show_uptime":         true,
			"show_response_times": true,
		}

		// Nested children carry their uuid string value alongside the integer id
		if uuid, ok := childMap["uuid"].(string); ok && uuid != "" {
			child["uuid"] = uuid
		}

		// Nested children use name as a localized map (not name_shown string)
		if childName, ok := childMap["name"].(map[string]interface{}); ok {
			nameStrMap := make(map[string]string)
			for k, v := range childName {
				if vStr, ok := v.(string); ok {
					nameStrMap[k] = vStr
				}
			}
			child["name"] = nameStrMap
		} else if nameShown, ok := childMap["name_shown"].(string); ok {
			child["name"] = map[string]string{"en": nameShown}
		} else {
			child["name"] = map[string]string{"en": ""}
		}

		children = append(children, child)
	}

	return children
}

// buildMockSections converts a sections slice from the request into the mock response format.
func buildMockSections(sectionsReq []interface{}) []map[string]interface{} {
	sections := []map[string]interface{}{}
	for _, secInterface := range sectionsReq {
		secMap, ok := secInterface.(map[string]interface{})
		if !ok {
			continue
		}

		section := map[string]interface{}{
			"is_split": false,
			"services": nil,
		}

		if name, ok := secMap["name"].(string); ok {
			section["name"] = map[string]string{"en": name}
		}
		if nameMap, ok := secMap["name"].(map[string]interface{}); ok {
			nameStrMap := make(map[string]string)
			for k, v := range nameMap {
				if vStr, ok := v.(string); ok {
					nameStrMap[k] = vStr
				}
			}
			section["name"] = nameStrMap
		}
		if isSplit, ok := secMap["is_split"].(bool); ok {
			section["is_split"] = isSplit
		}

		if servicesReq, ok := secMap["services"].([]interface{}); ok {
			services := []map[string]interface{}{}
			for _, svcInterface := range servicesReq {
				if svcMap, ok := svcInterface.(map[string]interface{}); ok {
					services = append(services, buildMockService(svcMap))
				}
			}
			section["services"] = services
		}

		sections = append(sections, section)
	}
	return sections
}

// buildDescriptionMap builds the multi-language description map from a request value.
// The real API accepts a plain string on write and returns a localized map on read.
// When a plain string is received it is stored under the "en" key (all others remain empty).
func buildDescriptionMap(description interface{}) map[string]string {
	descMap := map[string]string{"en": "", "fr": "", "de": "", "ru": "", "nl": ""}
	switch v := description.(type) {
	case string:
		descMap["en"] = v
	case map[string]interface{}:
		for lang, val := range v {
			if valStr, ok := val.(string); ok {
				descMap[lang] = valStr
			}
		}
	}
	return descMap
}

// applyAuthFields merges an authentication sub-map into settings,
// ensuring the allowed_domains field is always present.
func applyAuthFields(authMap map[string]interface{}, settings map[string]interface{}) {
	if _, hasAllowedDomains := authMap["allowed_domains"]; !hasAllowedDomains {
		authMap["allowed_domains"] = []string{}
	}
	settings["authentication"] = authMap
}

// applySettingsFields applies all scalar and structured settings fields from a
// request map onto the target settings map.
func applySettingsFields(req map[string]interface{}, settings map[string]interface{}) {
	stringFields := []string{
		"theme", "font", "accent_color", "default_language",
		"website", "logo", "logo_height", "favicon", "google_analytics",
	}
	for _, field := range stringFields {
		if val, ok := req[field].(string); ok {
			settings[field] = val
		}
	}

	boolFields := []string{
		"auto_refresh", "banner_header", "hide_powered_by", "hide_from_search_engines",
	}
	for _, field := range boolFields {
		if val, ok := req[field].(bool); ok {
			settings[field] = val
		}
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
		settings["description"] = buildDescriptionMap(description)
	}

	if subscribe, ok := req["subscribe"]; ok {
		settings["subscribe"] = subscribe
	}

	if authentication, ok := req["authentication"]; ok {
		if authMap, ok := authentication.(map[string]interface{}); ok {
			applyAuthFields(authMap, settings)
		} else {
			settings["authentication"] = authentication
		}
	}
}

// buildMockCreateSettings builds the initial settings map for a newly created status page.
func buildMockCreateSettings(req map[string]interface{}, settingsName string) map[string]interface{} {
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
	if defaultLang, ok := req["default_language"].(string); ok {
		settings["default_language"] = defaultLang
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

	settings["description"] = buildDescriptionMap(req["description"])

	for _, optKey := range []string{"logo", "logo_height", "favicon", "google_analytics"} {
		if val, ok := req[optKey].(string); ok {
			settings[optKey] = val
		}
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

	return settings
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

	settings := buildMockCreateSettings(req, settingsName)

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
		page["sections"] = buildMockSections(sectionsReq)
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
		page["sections"] = buildMockSections(sectionsReq)
	}

	settings, _ := page["settings"].(map[string]interface{})
	if settings == nil {
		settings = make(map[string]interface{})
	}

	applySettingsFields(req, settings)

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
