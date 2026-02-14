// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package dryrun

import (
	"fmt"
	"time"
)

// BridgeConverter provides a generic interface for conversion results.
type BridgeConverter interface {
	GetResourceName() string
	GetResourceType() string
	GetIssues() []string
	GetSourceData() map[string]interface{}
	GetTargetData() map[string]interface{}
}

// BuildComparisons creates resource comparisons from conversion results.
func BuildComparisons(converters []BridgeConverter) []ResourceComparison {
	comparisons := make([]ResourceComparison, 0, len(converters))

	for _, conv := range converters {
		comp := ResourceComparison{
			ResourceName:    conv.GetResourceName(),
			ResourceType:    conv.GetResourceType(),
			SourceData:      conv.GetSourceData(),
			TargetData:      conv.GetTargetData(),
			Transformations: extractTransformations(conv.GetSourceData(), conv.GetTargetData()),
			Unsupported:     []string{},
			HasWarnings:     len(conv.GetIssues()) > 0,
			HasErrors:       false,
		}

		comparisons = append(comparisons, comp)
	}

	return comparisons
}

func extractTransformations(source, target map[string]interface{}) []Transformation {
	var transforms []Transformation

	for key, targetVal := range target {
		sourceVal, sourceExists := source[key]

		if !sourceExists {
			transforms = append(transforms, Transformation{
				SourceField: "",
				TargetField: key,
				SourceValue: nil,
				TargetValue: targetVal,
				Action:      "defaulted",
				Notes:       fmt.Sprintf("Default value set: %v", targetVal),
			})
			continue
		}

		action := "preserved"
		notes := ""

		if fmt.Sprintf("%v", sourceVal) != fmt.Sprintf("%v", targetVal) {
			action = "mapped"
			notes = fmt.Sprintf("Converted from %v to %v", sourceVal, targetVal)
		}

		transforms = append(transforms, Transformation{
			SourceField: key,
			TargetField: key,
			SourceValue: sourceVal,
			TargetValue: targetVal,
			Action:      action,
			Notes:       notes,
		})
	}

	return transforms
}

// EstimatePerformance calculates performance estimates.
func EstimatePerformance(
	resourceCount int,
	sourceAPICalls int,
	tfContentSize int,
) PerformanceEstimates {
	// Base estimates
	migrationTime := time.Duration(resourceCount*2) * time.Second
	if migrationTime < 30*time.Second {
		migrationTime = 30 * time.Second
	}

	// Terraform operations
	tfPlanTime := time.Duration(resourceCount/5) * time.Second
	if tfPlanTime < 5*time.Second {
		tfPlanTime = 5 * time.Second
	}

	tfApplyTime := time.Duration(resourceCount) * time.Second
	if tfApplyTime < 10*time.Second {
		tfApplyTime = 10 * time.Second
	}

	// File sizes
	importScriptSize := int64(resourceCount * 80)
	manualStepsSize := int64(2048)
	reportSize := int64(resourceCount * 200)

	return PerformanceEstimates{
		MigrationTime:      migrationTime,
		SourceAPICalls:     sourceAPICalls,
		TargetAPICalls:     0,
		TerraformPlanTime:  tfPlanTime,
		TerraformApplyTime: tfApplyTime,
		TerraformFileSize:  int64(tfContentSize),
		ImportScriptSize:   importScriptSize,
		ManualStepsSize:    manualStepsSize,
		ReportSize:         reportSize,
	}
}

// BuildSummary creates a migration summary.
func BuildSummary(
	monitorCount int,
	healthcheckCount int,
	tfContent string,
	resourceBreakdown map[string]int,
) Summary {
	lines := 0
	if tfContent != "" {
		for _, c := range tfContent {
			if c == '\n' {
				lines++
			}
		}
	}

	return Summary{
		TotalMonitors:       monitorCount,
		TotalHealthchecks:   healthcheckCount,
		ExpectedTFResources: monitorCount + healthcheckCount,
		ExpectedTFLines:     lines,
		ExpectedTFSizeBytes: int64(len(tfContent)),
		ResourceBreakdown:   resourceBreakdown,
	}
}
