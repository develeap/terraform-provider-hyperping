// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package terraflyerrors

import (
	"fmt"
	"testing"

	"github.com/develeap/terraform-provider-hyperping/internal/client"
)

// TestDemo_AuthenticationError demonstrates an authentication error.
func TestDemo_AuthenticationError(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping demo test in short mode")
	}

	err := client.ErrUnauthorized
	enhanced := EnhanceClientError(err, "create", "hyperping_monitor.prod_api", "")

	fmt.Println("\n" + repeatStr("=", 80))
	fmt.Println("DEMO: Authentication Error")
	fmt.Println(repeatStr("=", 80))
	fmt.Println(enhanced.Error())
	fmt.Println(repeatStr("=", 80) + "\n")
}

// TestDemo_FrequencyValidation demonstrates a frequency validation error.
func TestDemo_FrequencyValidation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping demo test in short mode")
	}

	err := FrequencySuggestion(45)

	fmt.Println("\n" + repeatStr("=", 80))
	fmt.Println("DEMO: Frequency Validation Error")
	fmt.Println(repeatStr("=", 80))
	fmt.Println(err.Error())
	fmt.Println(repeatStr("=", 80) + "\n")
}

// TestDemo_RegionTypo demonstrates a region typo error.
func TestDemo_RegionTypo(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping demo test in short mode")
	}

	err := RegionSuggestion("frenkfurt")

	fmt.Println("\n" + repeatStr("=", 80))
	fmt.Println("DEMO: Region Typo Error")
	fmt.Println(repeatStr("=", 80))
	fmt.Println(err.Error())
	fmt.Println(repeatStr("=", 80) + "\n")
}

// TestDemo_RateLimit demonstrates a rate limit error with retry.
func TestDemo_RateLimit(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping demo test in short mode")
	}

	apiErr := client.NewRateLimitError(30)
	enhanced := EnhanceClientError(apiErr, "create", "hyperping_monitor.staging_api", "")

	fmt.Println("\n" + repeatStr("=", 80))
	fmt.Println("DEMO: Rate Limit Error")
	fmt.Println(repeatStr("=", 80))
	fmt.Println(enhanced.Error())
	fmt.Println(repeatStr("=", 80) + "\n")
}

// repeatStr repeats "=" 80 times
func repeatStr(_ string, _ int) string {
	result := ""
	for i := 0; i < 80; i++ {
		result += "="
	}
	return result
}
