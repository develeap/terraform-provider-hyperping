// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	tfresource "github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/develeap/terraform-provider-hyperping/internal/client"
)

func TestAccMaintenanceWindowDataSource_basic(t *testing.T) {
	server := newMockMaintenanceServer(t)
	defer server.Close()

	now := time.Now().UTC()
	start := now.Add(24 * time.Hour).Format(time.RFC3339)
	end := now.Add(26 * time.Hour).Format(time.RFC3339)

	// Pre-create a maintenance window
	server.maintenanceWindows["mw_test1"] = map[string]interface{}{
		"uuid":       "mw_test1",
		"name":       "Test Maintenance",
		"title":      map[string]interface{}{"en": "Test Maintenance Title"},
		"text":       map[string]interface{}{"en": "Test description"},
		"start_date": start,
		"end_date":   end,
		"monitors":   []string{"mon_123"},
	}

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config: testAccMaintenanceWindowDataSourceConfig(server.URL, "mw_test1"),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("data.hyperping_maintenance_window.test", "id", "mw_test1"),
					tfresource.TestCheckResourceAttr("data.hyperping_maintenance_window.test", "name", "Test Maintenance"),
					tfresource.TestCheckResourceAttr("data.hyperping_maintenance_window.test", "title", "Test Maintenance Title"),
				),
			},
		},
	})
}

func testAccMaintenanceWindowDataSourceConfig(baseURL, id string) string {
	return fmt.Sprintf(`
provider "hyperping" {
  api_key  = "test_api_key"
  base_url = %[1]q
}

data "hyperping_maintenance_window" "test" {
  id = %[2]q
}
`, baseURL, id)
}

// Mock server for maintenance windows

type mockMaintenanceServer struct {
	*httptest.Server
	t                  *testing.T
	maintenanceWindows map[string]map[string]interface{}
}

func newMockMaintenanceServer(t *testing.T) *mockMaintenanceServer {
	m := &mockMaintenanceServer{
		t:                  t,
		maintenanceWindows: make(map[string]map[string]interface{}),
	}

	m.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.handleRequest(w, r)
	}))

	return m
}

func (m *mockMaintenanceServer) handleRequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(client.HeaderContentType, client.ContentTypeJSON)
	basePath := client.MaintenanceBasePath
	basePathWithSlash := basePath + "/"

	switch {
	case r.Method == "GET" && r.URL.Path == basePath:
		m.listMaintenanceWindows(w)
	case r.Method == "GET" && strings.HasPrefix(r.URL.Path, basePathWithSlash):
		m.getMaintenanceWindow(w, r)
	default:
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Not found"})
	}
}

func (m *mockMaintenanceServer) listMaintenanceWindows(w http.ResponseWriter) {
	windows := make([]map[string]interface{}, 0, len(m.maintenanceWindows))
	for _, window := range m.maintenanceWindows {
		windows = append(windows, window)
	}
	json.NewEncoder(w).Encode(windows)
}

func (m *mockMaintenanceServer) getMaintenanceWindow(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, client.MaintenanceBasePath+"/")

	window, exists := m.maintenanceWindows[id]
	if !exists {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Maintenance window not found"})
		return
	}

	json.NewEncoder(w).Encode(window)
}
