package analyzer

import (
	"time"
)

// CoverageGapType categorizes the type of discrepancy found.
type CoverageGapType string

const (
	GapMissing         CoverageGapType = "missing"          // API has field, provider doesn't
	GapStale           CoverageGapType = "stale"            // Provider has field, API doesn't document
	GapTypeMismatch    CoverageGapType = "type_mismatch"    // Types differ
	GapRequiredChanged CoverageGapType = "required_changed" // Required/optional differs
)

// CoverageGap represents a discrepancy between API docs and provider.
type CoverageGap struct {
	Type       CoverageGapType
	Resource   string
	APIField   string
	TFField    string
	APIType    string
	TFType     string
	Details    string
	Severity   string
	Suggestion string
	FilePath   string
	CodeHint   string
}

// ResourceCoverage represents coverage for a single resource.
type ResourceCoverage struct {
	Resource          string
	TerraformResource string
	APIFields         int
	ImplementedFields int
	MissingFields     int
	StaleFields       int
	CoveragePercent   float64
}

// CoverageReport aggregates all coverage analysis.
type CoverageReport struct {
	Timestamp       time.Time
	Resources       []ResourceCoverage
	TotalAPIFields  int
	CoveredFields   int
	MissingFields   int
	StaleFields     int
	CoveragePercent float64
	Gaps            []CoverageGap
}
