// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package migrate

import "strings"

// DefaultRegions returns the default set of regions used when no mapping is available.
func DefaultRegions() []string {
	return []string{"london", "virginia", "singapore"}
}

// RegionAliases maps common cloud provider region identifiers to Hyperping region names.
var RegionAliases = map[string]string{
	"us":             "virginia",
	"us-east":        "virginia",
	"us-east-1":      "virginia",
	"us-west":        "oregon",
	"us-west-1":      "oregon",
	"eu":             "london",
	"eu-west":        "london",
	"eu-west-1":      "london",
	"eu-central":     "frankfurt",
	"eu-central-1":   "frankfurt",
	"asia":           "singapore",
	"ap-southeast":   "singapore",
	"ap-southeast-1": "singapore",
	"ap-northeast":   "tokyo",
	"ap-northeast-1": "tokyo",
	"au":             "sydney",
	"au-southeast":   "sydney",
	"sa":             "saopaulo",
	"sa-east-1":      "saopaulo",
	"me":             "bahrain",
	"me-south":       "bahrain",
	"me-south-1":     "bahrain",
}

// MapRegions converts a list of source region identifiers to Hyperping region names.
// Unknown regions are silently skipped. Duplicates are removed.
func MapRegions(sourceRegions []string) []string {
	var regions []string
	seen := make(map[string]bool)

	for _, region := range sourceRegions {
		normalized := strings.ToLower(strings.TrimSpace(region))
		if mapped, ok := RegionAliases[normalized]; ok {
			if !seen[mapped] {
				regions = append(regions, mapped)
				seen[mapped] = true
			}
		}
	}

	return regions
}
