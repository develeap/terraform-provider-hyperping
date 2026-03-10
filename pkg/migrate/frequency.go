// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package migrate

// AllowedFrequencies lists the check frequencies (in seconds) supported by Hyperping.
var AllowedFrequencies = []int{10, 20, 30, 60, 120, 180, 300, 600, 1800, 3600, 21600, 43200, 86400}

// MapFrequency maps an arbitrary interval (in seconds) to the nearest
// Hyperping-supported check frequency.
func MapFrequency(interval int) int {
	closest := AllowedFrequencies[0]
	minDiff := abs(interval - closest)

	for _, freq := range AllowedFrequencies {
		diff := abs(interval - freq)
		if diff < minDiff {
			minDiff = diff
			closest = freq
		}
	}

	return closest
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
