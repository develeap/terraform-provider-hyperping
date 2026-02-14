// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package dryrun

import "time"

// Report contains comprehensive dry-run analysis results.
type Report struct {
	Summary          Summary
	Comparison       []ResourceComparison
	TerraformPreview string
	Compatibility    CompatibilityScore
	Warnings         []Warning
	Estimates        PerformanceEstimates
	SourcePlatform   string
	TargetPlatform   string
	Timestamp        time.Time
}

// Summary contains high-level migration statistics.
type Summary struct {
	TotalMonitors         int
	TotalHealthchecks     int
	ExpectedTFResources   int
	ExpectedTFLines       int
	ExpectedTFSizeBytes   int64
	ResourceBreakdown     map[string]int
	FrequencyDistribution map[int]int
	RegionDistribution    map[string]int
}

// ResourceComparison shows side-by-side comparison of source and target.
type ResourceComparison struct {
	ResourceName    string
	ResourceType    string
	SourceData      map[string]interface{}
	TargetData      map[string]interface{}
	Transformations []Transformation
	Unsupported     []string
	HasWarnings     bool
	HasErrors       bool
}

// Transformation describes a field mapping from source to target.
type Transformation struct {
	SourceField string
	TargetField string
	SourceValue interface{}
	TargetValue interface{}
	Action      string // "preserved", "mapped", "rounded", "defaulted", "dropped"
	Notes       string
}

// CompatibilityScore provides migration compatibility analysis.
type CompatibilityScore struct {
	OverallScore    float64
	ByType          map[string]float64
	CleanMigrations int
	WarningCount    int
	ErrorCount      int
	Complexity      string
	Details         string
}

// Warning represents a migration warning or manual step.
type Warning struct {
	Severity     string // "critical", "warning", "info"
	ResourceName string
	ResourceType string
	Category     string
	Message      string
	Action       string
	Impact       string
}

// PerformanceEstimates provides time and resource estimates.
type PerformanceEstimates struct {
	MigrationTime       time.Duration
	SourceAPICalls      int
	TargetAPICalls      int
	TerraformPlanTime   time.Duration
	TerraformApplyTime  time.Duration
	TerraformFileSize   int64
	ImportScriptSize    int64
	ManualStepsSize     int64
	ReportSize          int64
	EstimatedManualTime time.Duration
}

// Options configures dry-run behavior.
type Options struct {
	Verbose          bool
	ShowAllResources bool
	PreviewLimit     int
	FormatJSON       bool
	IncludeExamples  bool
}
