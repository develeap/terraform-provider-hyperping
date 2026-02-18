package extractor

import (
	"testing"
)

func TestIsRequestParameterSection(t *testing.T) {
	tests := []struct {
		title    string
		expected bool
	}{
		// Request sections - should return true
		{"Body", true},
		{"body", true},
		{"Path Parameters", true},
		{"path parameters", true},
		{"Query Parameters", true},
		{"Headers", true},
		{"Request Body", true},

		// Response sections - should return false
		{"Healthcheck Object Fields", false},
		{"Incident Object", false},
		{"Maintenance Window Object", false},
		{"Response", false},
		{"API Key Options", false},
		{"Permission Reference", false},
		{"Monitor Object Fields", false},

		// Unknown sections - should return false
		{"Something Random", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			result := isRequestParameterSection(tt.title)
			if result != tt.expected {
				t.Errorf("isRequestParameterSection(%q) = %v, want %v", tt.title, result, tt.expected)
			}
		})
	}
}

func TestExtractFromHTML(t *testing.T) {
	// Test HTML with Body parameters (should extract)
	htmlWithBody := `
		<div class="api-params-table">
			<div class="api-params-header">
				<span class="api-params-title">Body</span>
			</div>
			<div class="api-param-row">
				<div class="api-param-header">
					<span class="api-param-name">name</span>
					<span class="api-param-type">string</span>
					<span class="api-param-required">required</span>
				</div>
				<div class="api-param-description">The name of the resource</div>
			</div>
			<div class="api-param-row">
				<div class="api-param-header">
					<span class="api-param-name">enabled</span>
					<span class="api-param-type">boolean</span>
					<span class="api-param-optional">optional</span>
				</div>
			</div>
		</div>
	`

	params := ExtractFromHTML(htmlWithBody)
	if len(params) != 2 {
		t.Errorf("Expected 2 params, got %d", len(params))
	}

	if len(params) > 0 {
		if params[0].Name != "name" {
			t.Errorf("Expected first param name 'name', got '%s'", params[0].Name)
		}
		if !params[0].Required {
			t.Error("Expected 'name' to be required")
		}
	}

	if len(params) > 1 {
		if params[1].Name != "enabled" {
			t.Errorf("Expected second param name 'enabled', got '%s'", params[1].Name)
		}
		if params[1].Required {
			t.Error("Expected 'enabled' to be optional")
		}
	}
}

func TestExtractFromHTML_ResponseSection(t *testing.T) {
	// Test HTML with only response documentation (should NOT extract)
	htmlWithResponse := `
		<div class="api-params-table">
			<div class="api-params-header">
				<span class="api-params-title">Healthcheck Object Fields</span>
			</div>
			<div class="api-param-row">
				<div class="api-param-header">
					<span class="api-param-name">uuid</span>
					<span class="api-param-type">string</span>
				</div>
			</div>
		</div>
	`

	params := ExtractFromHTML(htmlWithResponse)
	if len(params) != 0 {
		t.Errorf("Expected 0 params from response section, got %d", len(params))
	}
}

func TestExtractFromHTML_MixedSections(t *testing.T) {
	// Test HTML with both Path Parameters and Response Object
	htmlMixed := `
		<div class="api-params-table">
			<div class="api-params-header">
				<span class="api-params-title">Path Parameters</span>
			</div>
			<div class="api-param-row">
				<div class="api-param-header">
					<span class="api-param-name">id</span>
					<span class="api-param-type">string</span>
					<span class="api-param-required">required</span>
				</div>
			</div>
		</div>
		<div class="api-params-table">
			<div class="api-params-header">
				<span class="api-params-title">Response Object</span>
			</div>
			<div class="api-param-row">
				<div class="api-param-header">
					<span class="api-param-name">created_at</span>
					<span class="api-param-type">string</span>
				</div>
			</div>
		</div>
	`

	params := ExtractFromHTML(htmlMixed)
	if len(params) != 1 {
		t.Errorf("Expected 1 param (only from Path Parameters), got %d", len(params))
	}

	if len(params) > 0 && params[0].Name != "id" {
		t.Errorf("Expected param name 'id', got '%s'", params[0].Name)
	}
}

func TestHasParameterTables(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		expected bool
	}{
		{
			name:     "with api-params-table",
			html:     `<div class="api-params-table">content</div>`,
			expected: true,
		},
		{
			name:     "with api-param-row",
			html:     `<div class="api-param-row">content</div>`,
			expected: true,
		},
		{
			name:     "without tables",
			html:     `<div class="other-content">no tables here</div>`,
			expected: false,
		},
		{
			name:     "empty",
			html:     ``,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HasParameterTablesInHTML(tt.html)
			if result != tt.expected {
				t.Errorf("HasParameterTables() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestNormalizeType(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"string", "string"},
		{"String", "string"},
		{"enum<string>", "enum"},
		{"array<string>", "array"},
		{"array<object>", "array"},
		{"object", "object"},
		{"boolean", "boolean"},
		{"number", "number"},
		{"integer", "number"},
		{"int", "number"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := normalizeType(tt.input)
			if result != tt.expected {
				t.Errorf("normalizeType(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
