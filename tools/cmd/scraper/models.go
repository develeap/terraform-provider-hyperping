package main

import (
	"time"

	"github.com/develeap/terraform-provider-hyperping/tools/scraper/extractor"
)

// CacheEntry stores metadata for change detection
type CacheEntry struct {
	Filename     string    `json:"filename"`
	URL          string    `json:"url"`
	ContentHash  string    `json:"content_hash"`
	Size         int       `json:"size"`
	LastModified time.Time `json:"last_modified"`
}

// Cache stores all cache entries with creation timestamp
type Cache struct {
	Entries   map[string]CacheEntry `json:"entries"`
	CreatedAt time.Time             `json:"created_at"`
}

// ParameterChange represents a modification to an existing parameter
type ParameterChange struct {
	Name        string      `json:"name"`
	OldType     string      `json:"old_type"`
	NewType     string      `json:"new_type"`
	OldRequired bool        `json:"old_required"`
	NewRequired bool        `json:"new_required"`
	OldDefault  interface{} `json:"old_default,omitempty"`
	NewDefault  interface{} `json:"new_default,omitempty"`
}

// APIDiff represents changes detected between two versions of an API endpoint
type APIDiff struct {
	Section          string                   `json:"section"` // monitors, statuspages, incidents, etc.
	Endpoint         string                   `json:"endpoint"`
	Method           string                   `json:"method"`
	AddedParams      []extractor.APIParameter `json:"added_params,omitempty"`
	RemovedParams    []extractor.APIParameter `json:"removed_params,omitempty"`
	ModifiedParams   []ParameterChange        `json:"modified_params,omitempty"`
	Breaking         bool                     `json:"breaking"`
	RawContentChange bool                     `json:"raw_content_change"` // True if hash changed but no semantic diff found
}

// DiffReport aggregates all changes for a scraping run
type DiffReport struct {
	Timestamp      time.Time `json:"timestamp"`
	TotalPages     int       `json:"total_pages"`
	ChangedPages   int       `json:"changed_pages"`
	UnchangedPages int       `json:"unchanged_pages"`
	APIDiffs       []APIDiff `json:"api_diffs,omitempty"`
	Breaking       bool      `json:"breaking"` // True if any diff is breaking
	Summary        string    `json:"summary"`
}

// DiscoveredURL represents a URL found during navigation discovery
type DiscoveredURL struct {
	URL      string `json:"url"`
	Section  string `json:"section"`   // Parent section (monitors, statuspages, etc.)
	Method   string `json:"method"`    // HTTP method or operation (list, create, update, etc.)
	IsParent bool   `json:"is_parent"` // True if this is a parent page with children
}

// =============================================================================
// Coverage Analysis Types
// =============================================================================

// CoverageGapType categorizes the type of discrepancy found
type CoverageGapType string

const (
	GapMissing         CoverageGapType = "missing"          // API has field, provider doesn't
	GapStale           CoverageGapType = "stale"            // Provider has field, API doesn't document
	GapTypeMismatch    CoverageGapType = "type_mismatch"    // Types differ between API and provider
	GapRequiredChanged CoverageGapType = "required_changed" // Required/optional status differs
	GapValidValues     CoverageGapType = "valid_values"     // Allowed values differ
)

// SchemaField represents a field extracted from the Terraform provider schema
type SchemaField struct {
	TerraformName string      `json:"terraform_name"` // tfsdk:"name" value
	APIName       string      `json:"api_name"`       // json:"name" value from client models
	Type          string      `json:"type"`           // string, int64, bool, list, object
	Required      bool        `json:"required"`
	Optional      bool        `json:"optional"`
	Computed      bool        `json:"computed"`
	Default       interface{} `json:"default,omitempty"`
	ValidValues   []string    `json:"valid_values,omitempty"` // From OneOf validators
	Description   string      `json:"description,omitempty"`
	Deprecated    bool        `json:"deprecated,omitempty"`
}

// ResourceSchema represents the extracted schema for a single Terraform resource
type ResourceSchema struct {
	Name       string        `json:"name"`        // e.g., "hyperping_monitor"
	APISection string        `json:"api_section"` // e.g., "monitors"
	Fields     []SchemaField `json:"fields"`
	FilePath   string        `json:"file_path"` // Source file path
}

// CoverageGap represents a discrepancy between API docs and provider schema
type CoverageGap struct {
	Type       CoverageGapType `json:"type"`
	Resource   string          `json:"resource"`             // e.g., "monitor"
	APIField   string          `json:"api_field,omitempty"`  // Field name in API docs
	TFField    string          `json:"tf_field,omitempty"`   // Field name in Terraform
	APIType    string          `json:"api_type,omitempty"`   // Type in API docs
	TFType     string          `json:"tf_type,omitempty"`    // Type in Terraform
	Details    string          `json:"details"`              // Human-readable explanation
	Severity   string          `json:"severity"`             // error, warning, info
	Suggestion string          `json:"suggestion,omitempty"` // Recommended action
	FilePath   string          `json:"file_path,omitempty"`  // File to modify
	CodeHint   string          `json:"code_hint,omitempty"`  // Suggested code addition
}

// ResourceCoverage represents coverage statistics for a single resource
type ResourceCoverage struct {
	Resource          string  `json:"resource"`
	TerraformResource string  `json:"terraform_resource"` // e.g., "hyperping_monitor"
	APIFields         int     `json:"api_fields"`
	ImplementedFields int     `json:"implemented_fields"`
	MissingFields     int     `json:"missing_fields"`
	StaleFields       int     `json:"stale_fields"`
	CoveragePercent   float64 `json:"coverage_percent"`
}

// CoverageReport aggregates all coverage analysis results
type CoverageReport struct {
	Timestamp       time.Time          `json:"timestamp"`
	ProviderVersion string             `json:"provider_version,omitempty"`
	Resources       []ResourceCoverage `json:"resources"`
	TotalAPIFields  int                `json:"total_api_fields"`
	CoveredFields   int                `json:"covered_fields"`
	MissingFields   int                `json:"missing_fields"`
	StaleFields     int                `json:"stale_fields"`
	CoveragePercent float64            `json:"coverage_percent"`
	Gaps            []CoverageGap      `json:"gaps,omitempty"`
}
