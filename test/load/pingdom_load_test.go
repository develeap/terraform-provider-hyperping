// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

//go:build load

package load

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/develeap/terraform-provider-hyperping/cmd/migrate-pingdom/converter"
	"github.com/develeap/terraform-provider-hyperping/cmd/migrate-pingdom/generator"
	"github.com/develeap/terraform-provider-hyperping/cmd/migrate-pingdom/pingdom"
	"github.com/stretchr/testify/require"
)

// TestPingdomLoad_SmallScale tests migration with 10 checks (baseline).
func TestPingdomLoad_SmallScale(t *testing.T) {
	scenario := LoadTestScenario{
		Name:               "Pingdom_10checks",
		Description:        "Small scale baseline test",
		NumMonitors:        10,
		NumHealthchecks:    0,
		MaxDurationSeconds: 30,
		MaxMemoryMB:        100,
		ExpectedThroughput: 1.0,
	}

	runPingdomLoadTest(t, scenario)
}

// TestPingdomLoad_MediumScale tests migration with 50 checks.
func TestPingdomLoad_MediumScale(t *testing.T) {
	scenario := LoadTestScenario{
		Name:               "Pingdom_50checks",
		Description:        "Medium scale test",
		NumMonitors:        50,
		NumHealthchecks:    0,
		MaxDurationSeconds: 60,
		MaxMemoryMB:        250,
		ExpectedThroughput: 1.0,
	}

	runPingdomLoadTest(t, scenario)
}

// TestPingdomLoad_LargeScale tests migration with 100 checks.
func TestPingdomLoad_LargeScale(t *testing.T) {
	scenario := LoadTestScenario{
		Name:                  "Pingdom_100checks",
		Description:           "Large scale test (100 checks)",
		NumMonitors:           100,
		NumHealthchecks:       0,
		MaxDurationSeconds:    120,
		MaxMemoryMB:           500,
		ExpectedThroughput:    1.0,
		ValidateLinearScaling: true,
	}

	runPingdomLoadTest(t, scenario)
}

// TestPingdomLoad_XLScale tests migration with 200+ checks.
func TestPingdomLoad_XLScale(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping XL scale test in short mode")
	}

	scenario := LoadTestScenario{
		Name:                  "Pingdom_200checks",
		Description:           "XL scale test (200+ checks)",
		NumMonitors:           200,
		NumHealthchecks:       0,
		MaxDurationSeconds:    300,
		MaxMemoryMB:           1000,
		ExpectedThroughput:    1.0,
		ValidateLinearScaling: true,
	}

	runPingdomLoadTest(t, scenario)
}

// TestPingdomLoad_MemoryLeak tests for memory leaks across multiple iterations.
func TestPingdomLoad_MemoryLeak(t *testing.T) {
	ForceGC(t)

	var beforeMem runtime.MemStats
	runtime.ReadMemStats(&beforeMem)

	// Run conversion 5 times with the same data
	checks := generatePingdomChecks(50)
	conv := converter.NewCheckConverter()

	for i := 0; i < 5; i++ {
		for _, check := range checks {
			_ = conv.Convert(check)
		}
		ForceGC(t)
	}

	var afterMem runtime.MemStats
	runtime.ReadMemStats(&afterMem)

	CheckMemoryLeak(t, beforeMem.Alloc, afterMem.Alloc, "Pingdom conversion")
}

// TestPingdomLoad_FileSize validates generated file sizes for large migrations.
func TestPingdomLoad_FileSize(t *testing.T) {
	tempDir := CreateTempTestDir(t, "pingdom-filesize")

	// Generate large dataset
	checks := generatePingdomChecks(100)

	// Convert
	conv := converter.NewCheckConverter()
	var results []converter.ConversionResult

	for _, check := range checks {
		result := conv.Convert(check)
		results = append(results, result)
	}

	// Generate files
	tfGen := generator.NewTerraformGenerator("pingdom")
	tfConfig := tfGen.GenerateHCL(checks, results)

	importGen := generator.NewImportGenerator("pingdom")
	importScript := importGen.GenerateImportScript(checks, results, make(map[int]string))

	// Write files
	tfFile := filepath.Join(tempDir, "hyperping.tf")
	importFile := filepath.Join(tempDir, "import.sh")

	require.NoError(t, os.WriteFile(tfFile, []byte(tfConfig), 0600))
	require.NoError(t, os.WriteFile(importFile, []byte(importScript), 0600))

	// Validate file sizes
	t.Log("Validating generated file sizes for 100 checks:")
	ValidateFileSize(t, tfFile, 10)    // Max 10MB for Terraform config
	ValidateFileSize(t, importFile, 5) // Max 5MB for import script

	t.Log("✅ File size validation passed")
}

// TestPingdomLoad_ParallelConversion tests parallel conversion of checks.
func TestPingdomLoad_ParallelConversion(t *testing.T) {
	checks := generatePingdomChecks(100)

	// Test for race conditions with parallel execution
	conv := converter.NewCheckConverter()

	// Run conversions in parallel
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			for _, check := range checks {
				_ = conv.Convert(check)
			}
			done <- true
		}()
	}

	// Wait for all to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	t.Log("✅ Parallel conversion test passed (no race conditions)")
}

// TestPingdomLoad_MixedCheckTypes tests performance with diverse check types.
func TestPingdomLoad_MixedCheckTypes(t *testing.T) {
	ForceGC(t)
	startTime := time.Now()

	// Generate mix of all check types
	checks := make([]pingdom.Check, 100)
	types := []string{"http", "tcp", "ping", "smtp", "pop3", "imap"}
	for i := 0; i < 100; i++ {
		checks[i] = generatePingdomCheckOfType(i, types[i%len(types)])
	}

	// Convert
	conv := converter.NewCheckConverter()
	supported := 0
	unsupported := 0

	for _, check := range checks {
		result := conv.Convert(check)
		if result.Supported {
			supported++
		} else {
			unsupported++
		}
	}

	t.Logf("Converted: %d supported, %d unsupported", supported, unsupported)

	stats := CaptureMemoryStats(startTime, 100)
	LogMemoryStats(t, "Mixed Check Types", stats)
	ValidateMemoryUsage(t, stats, 100)

	t.Log("✅ Mixed check types test passed")
}

// TestPingdomLoad_ComplexHTTPChecks tests performance with complex HTTP checks.
func TestPingdomLoad_ComplexHTTPChecks(t *testing.T) {
	ForceGC(t)
	startTime := time.Now()

	// Generate HTTP checks with headers, POST data, etc.
	checks := make([]pingdom.Check, 100)
	for i := 0; i < 100; i++ {
		checks[i] = pingdom.Check{
			ID:         1000000 + i,
			Name:       fmt.Sprintf("complex-http-check-%d", i),
			Type:       "http",
			Hostname:   fmt.Sprintf("api-%d.example.com", i),
			URL:        fmt.Sprintf("/api/v1/resource/%d", i),
			Encryption: true,
			Resolution: 5,
			Paused:     false,
			RequestHeaders: map[string]string{
				"Authorization":   fmt.Sprintf("Bearer token-%d", i),
				"Content-Type":    "application/json",
				"X-Request-ID":    fmt.Sprintf("req-%d", i),
				"X-Custom-Header": fmt.Sprintf("value-%d", i),
			},
			PostData:      fmt.Sprintf(`{"test": "data-%d", "id": %d}`, i, i),
			ShouldContain: "success",
			Tags: []pingdom.Tag{
				{Name: "production", Type: "u"},
				{Name: fmt.Sprintf("service-%d", i%10), Type: "u"},
			},
		}
	}

	// Convert
	conv := converter.NewCheckConverter()
	for _, check := range checks {
		_ = conv.Convert(check)
	}

	stats := CaptureMemoryStats(startTime, 100)
	LogMemoryStats(t, "Complex HTTP Checks", stats)
	ValidateMemoryUsage(t, stats, 100)

	t.Log("✅ Complex HTTP checks test passed")
}

// BenchmarkPingdomConversion benchmarks check conversion performance.
func BenchmarkPingdomConversion(b *testing.B) {
	checks := generatePingdomChecks(100)
	conv := converter.NewCheckConverter()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		for _, check := range checks {
			_ = conv.Convert(check)
		}
	}
}

// BenchmarkPingdomGeneration benchmarks Terraform generation performance.
func BenchmarkPingdomGeneration(b *testing.B) {
	checks := generatePingdomChecks(100)

	conv := converter.NewCheckConverter()
	var results []converter.ConversionResult

	for _, check := range checks {
		result := conv.Convert(check)
		results = append(results, result)
	}

	tfGen := generator.NewTerraformGenerator("pingdom")

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = tfGen.GenerateHCL(checks, results)
	}
}

// BenchmarkPingdomE2E benchmarks end-to-end migration performance.
func BenchmarkPingdomE2E(b *testing.B) {
	checks := generatePingdomChecks(100)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		conv := converter.NewCheckConverter()
		var results []converter.ConversionResult

		for _, check := range checks {
			result := conv.Convert(check)
			results = append(results, result)
		}

		tfGen := generator.NewTerraformGenerator("pingdom")
		_ = tfGen.GenerateHCL(checks, results)

		importGen := generator.NewImportGenerator("pingdom")
		_ = importGen.GenerateImportScript(checks, results, make(map[int]string))
	}
}

// runPingdomLoadTest executes a load test scenario for Pingdom migration.
func runPingdomLoadTest(t *testing.T, scenario LoadTestScenario) {
	t.Helper()
	t.Logf("=== Running Load Test: %s ===", scenario.Name)
	t.Logf("Description: %s", scenario.Description)
	t.Logf("Checks: %d", scenario.NumMonitors)

	ForceGC(t)
	startTime := time.Now()

	// Generate test data
	t.Log("Generating test data...")
	checks := generatePingdomChecks(scenario.NumMonitors)

	// Convert resources
	t.Log("Converting resources...")
	conv := converter.NewCheckConverter()
	var results []converter.ConversionResult

	supported := 0
	for _, check := range checks {
		result := conv.Convert(check)
		results = append(results, result)
		if result.Supported {
			supported++
		}
	}

	t.Logf("Converted: %d supported of %d total", supported, len(checks))

	// Generate Terraform
	t.Log("Generating Terraform configuration...")
	tfGen := generator.NewTerraformGenerator("pingdom")
	tfConfig := tfGen.GenerateHCL(checks, results)

	importGen := generator.NewImportGenerator("pingdom")
	importScript := importGen.GenerateImportScript(checks, results, make(map[int]string))

	require.NotEmpty(t, tfConfig, "Terraform config should not be empty")
	require.NotEmpty(t, importScript, "Import script should not be empty")

	// Capture memory stats
	stats := CaptureMemoryStats(startTime, supported)
	LogMemoryStats(t, scenario.Name, stats)

	// Validate memory usage
	ValidateMemoryUsage(t, stats, scenario.NumMonitors)

	// Validate performance
	ValidatePerformance(t, stats, supported, scenario.MaxDurationSeconds)

	// Write to temp directory and validate file sizes
	tempDir := CreateTempTestDir(t, "pingdom-load")
	tfFile := filepath.Join(tempDir, "hyperping.tf")
	importFile := filepath.Join(tempDir, "import.sh")

	require.NoError(t, os.WriteFile(tfFile, []byte(tfConfig), 0600))
	require.NoError(t, os.WriteFile(importFile, []byte(importScript), 0600))

	t.Log("Validating generated file sizes:")
	ValidateFileSize(t, tfFile, 10)
	ValidateFileSize(t, importFile, 5)

	t.Logf("=== Load Test PASSED: %s ===", scenario.Name)
}

// generatePingdomChecks generates test checks for load testing.
func generatePingdomChecks(count int) []pingdom.Check {
	checks := make([]pingdom.Check, count)

	types := []string{"http", "tcp", "ping"}
	resolutions := []int{1, 5, 15, 30, 60}

	for i := 0; i < count; i++ {
		checkType := types[i%len(types)]
		checks[i] = generatePingdomCheckOfType(i, checkType)
		checks[i].Resolution = resolutions[i%len(resolutions)]
	}

	return checks
}

// generatePingdomCheckOfType generates a check of a specific type.
func generatePingdomCheckOfType(id int, checkType string) pingdom.Check {
	check := pingdom.Check{
		ID:         1000000 + id,
		Name:       fmt.Sprintf("load-test-check-%d", id),
		Type:       checkType,
		Resolution: 5,
		Paused:     false,
		Tags: []pingdom.Tag{
			{Name: "load-test", Type: "u"},
		},
	}

	switch checkType {
	case "http", "https":
		check.Hostname = fmt.Sprintf("api-%d.example.com", id)
		check.URL = fmt.Sprintf("/health/%d", id)
		check.Encryption = checkType == "https"
		check.RequestHeaders = map[string]string{
			"X-Test-ID": fmt.Sprintf("%d", id),
		}

	case "tcp":
		check.Hostname = fmt.Sprintf("server-%d.example.com", id)
		check.Port = 443

	case "ping":
		check.Hostname = fmt.Sprintf("192.168.1.%d", id%254+1)

	case "smtp":
		check.Hostname = fmt.Sprintf("mail-%d.example.com", id)
		check.Port = 25
		check.Encryption = false

	case "pop3":
		check.Hostname = fmt.Sprintf("mail-%d.example.com", id)
		check.Port = 110
		check.Encryption = false

	case "imap":
		check.Hostname = fmt.Sprintf("mail-%d.example.com", id)
		check.Port = 143
		check.Encryption = false
	}

	return check
}
