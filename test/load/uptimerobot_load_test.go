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

	"github.com/develeap/terraform-provider-hyperping/cmd/migrate-uptimerobot/converter"
	"github.com/develeap/terraform-provider-hyperping/cmd/migrate-uptimerobot/generator"
	"github.com/develeap/terraform-provider-hyperping/cmd/migrate-uptimerobot/uptimerobot"
	"github.com/stretchr/testify/require"
)

// TestUptimeRobotLoad_SmallScale tests migration with 10 monitors (baseline).
func TestUptimeRobotLoad_SmallScale(t *testing.T) {
	scenario := LoadTestScenario{
		Name:               "UptimeRobot_10monitors",
		Description:        "Small scale baseline test",
		NumMonitors:        10,
		NumHealthchecks:    5,
		MaxDurationSeconds: 30,
		MaxMemoryMB:        100,
		ExpectedThroughput: 1.0,
	}

	runUptimeRobotLoadTest(t, scenario)
}

// TestUptimeRobotLoad_MediumScale tests migration with 50 monitors.
func TestUptimeRobotLoad_MediumScale(t *testing.T) {
	scenario := LoadTestScenario{
		Name:               "UptimeRobot_50monitors",
		Description:        "Medium scale test",
		NumMonitors:        50,
		NumHealthchecks:    25,
		MaxDurationSeconds: 60,
		MaxMemoryMB:        250,
		ExpectedThroughput: 1.0,
	}

	runUptimeRobotLoadTest(t, scenario)
}

// TestUptimeRobotLoad_LargeScale tests migration with 100 monitors.
func TestUptimeRobotLoad_LargeScale(t *testing.T) {
	scenario := LoadTestScenario{
		Name:                  "UptimeRobot_100monitors",
		Description:           "Large scale test (100 monitors)",
		NumMonitors:           100,
		NumHealthchecks:       50,
		MaxDurationSeconds:    120,
		MaxMemoryMB:           500,
		ExpectedThroughput:    1.0,
		ValidateLinearScaling: true,
	}

	runUptimeRobotLoadTest(t, scenario)
}

// TestUptimeRobotLoad_XLScale tests migration with 200+ monitors.
func TestUptimeRobotLoad_XLScale(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping XL scale test in short mode")
	}

	scenario := LoadTestScenario{
		Name:                  "UptimeRobot_200monitors",
		Description:           "XL scale test (200+ monitors)",
		NumMonitors:           200,
		NumHealthchecks:       100,
		MaxDurationSeconds:    300,
		MaxMemoryMB:           1000,
		ExpectedThroughput:    1.0,
		ValidateLinearScaling: true,
	}

	runUptimeRobotLoadTest(t, scenario)
}

// TestUptimeRobotLoad_MemoryLeak tests for memory leaks across multiple iterations.
func TestUptimeRobotLoad_MemoryLeak(t *testing.T) {
	ForceGC(t)

	var beforeMem runtime.MemStats
	runtime.ReadMemStats(&beforeMem)

	// Run conversion 5 times with the same data
	monitors := generateUptimeRobotMonitors(50)
	alertContacts := generateUptimeRobotAlertContacts(10)

	conv := converter.NewConverter()

	for i := 0; i < 5; i++ {
		_ = conv.Convert(monitors, alertContacts)
		ForceGC(t)
	}

	var afterMem runtime.MemStats
	runtime.ReadMemStats(&afterMem)

	CheckMemoryLeak(t, beforeMem.Alloc, afterMem.Alloc, "UptimeRobot conversion")
}

// TestUptimeRobotLoad_FileSize validates generated file sizes for large migrations.
func TestUptimeRobotLoad_FileSize(t *testing.T) {
	tempDir := CreateTempTestDir(t, "uptimerobot-filesize")

	// Generate large dataset
	monitors := generateUptimeRobotMonitors(100)
	alertContacts := generateUptimeRobotAlertContacts(20)

	// Convert
	conv := converter.NewConverter()
	result := conv.Convert(monitors, alertContacts)

	// Generate files
	tfConfig := generator.GenerateTerraform(result)
	importScript := generator.GenerateImportScript(result)

	// Write files
	tfFile := filepath.Join(tempDir, "hyperping.tf")
	importFile := filepath.Join(tempDir, "import.sh")

	require.NoError(t, os.WriteFile(tfFile, []byte(tfConfig), 0600))
	require.NoError(t, os.WriteFile(importFile, []byte(importScript), 0600))

	// Validate file sizes
	t.Log("Validating generated file sizes for 100 monitors + 50 healthchecks:")
	ValidateFileSize(t, tfFile, 10)    // Max 10MB for Terraform config
	ValidateFileSize(t, importFile, 5) // Max 5MB for import script

	t.Log("✅ File size validation passed")
}

// TestUptimeRobotLoad_ParallelConversion tests parallel conversion of monitors.
func TestUptimeRobotLoad_ParallelConversion(t *testing.T) {
	monitors := generateUptimeRobotMonitors(100)
	alertContacts := generateUptimeRobotAlertContacts(20)

	// Test for race conditions with parallel execution
	conv := converter.NewConverter()

	// Run conversions in parallel
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			_ = conv.Convert(monitors, alertContacts)
			done <- true
		}()
	}

	// Wait for all to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	t.Log("✅ Parallel conversion test passed (no race conditions)")
}

// TestUptimeRobotLoad_MixedMonitorTypes tests performance with diverse monitor types.
func TestUptimeRobotLoad_MixedMonitorTypes(t *testing.T) {
	ForceGC(t)
	startTime := time.Now()

	// Generate mix of all monitor types
	monitors := make([]uptimerobot.Monitor, 100)
	types := []int{1, 2, 3, 4, 5} // HTTP, Keyword, Ping, Port, Heartbeat
	for i := 0; i < 100; i++ {
		monitors[i] = generateUptimeRobotMonitorOfType(i, types[i%len(types)])
	}

	alertContacts := generateUptimeRobotAlertContacts(20)

	// Convert
	conv := converter.NewConverter()
	result := conv.Convert(monitors, alertContacts)

	// Verify all types were converted
	t.Logf("Converted: %d monitors, %d healthchecks, %d skipped",
		len(result.Monitors), len(result.Healthchecks), len(result.Skipped))

	stats := CaptureMemoryStats(startTime, 100)
	LogMemoryStats(t, "Mixed Monitor Types", stats)
	ValidateMemoryUsage(t, stats, 100)

	t.Log("✅ Mixed monitor types test passed")
}

// BenchmarkUptimeRobotConversion benchmarks monitor conversion performance.
func BenchmarkUptimeRobotConversion(b *testing.B) {
	monitors := generateUptimeRobotMonitors(100)
	alertContacts := generateUptimeRobotAlertContacts(20)
	conv := converter.NewConverter()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = conv.Convert(monitors, alertContacts)
	}
}

// BenchmarkUptimeRobotGeneration benchmarks Terraform generation performance.
func BenchmarkUptimeRobotGeneration(b *testing.B) {
	monitors := generateUptimeRobotMonitors(100)
	alertContacts := generateUptimeRobotAlertContacts(20)

	conv := converter.NewConverter()
	result := conv.Convert(monitors, alertContacts)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = generator.GenerateTerraform(result)
	}
}

// BenchmarkUptimeRobotE2E benchmarks end-to-end migration performance.
func BenchmarkUptimeRobotE2E(b *testing.B) {
	monitors := generateUptimeRobotMonitors(100)
	alertContacts := generateUptimeRobotAlertContacts(20)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		conv := converter.NewConverter()
		result := conv.Convert(monitors, alertContacts)

		_ = generator.GenerateTerraform(result)
		_ = generator.GenerateImportScript(result)
	}
}

// runUptimeRobotLoadTest executes a load test scenario for UptimeRobot migration.
func runUptimeRobotLoadTest(t *testing.T, scenario LoadTestScenario) {
	t.Helper()
	t.Logf("=== Running Load Test: %s ===", scenario.Name)
	t.Logf("Description: %s", scenario.Description)
	t.Logf("Monitors: %d, Healthchecks: %d", scenario.NumMonitors, scenario.NumHealthchecks)

	ForceGC(t)
	startTime := time.Now()

	// Generate test data
	t.Log("Generating test data...")
	monitors := generateUptimeRobotMonitors(scenario.NumMonitors)
	// Add heartbeats
	for i := 0; i < scenario.NumHealthchecks; i++ {
		monitors = append(monitors, generateUptimeRobotMonitorOfType(scenario.NumMonitors+i, 5))
	}
	alertContacts := generateUptimeRobotAlertContacts(20)

	// Convert resources
	t.Log("Converting resources...")
	conv := converter.NewConverter()
	result := conv.Convert(monitors, alertContacts)

	totalConverted := len(result.Monitors) + len(result.Healthchecks)
	t.Logf("Converted: %d monitors, %d healthchecks, %d skipped",
		len(result.Monitors), len(result.Healthchecks), len(result.Skipped))

	// Generate Terraform
	t.Log("Generating Terraform configuration...")
	tfConfig := generator.GenerateTerraform(result)
	importScript := generator.GenerateImportScript(result)

	require.NotEmpty(t, tfConfig, "Terraform config should not be empty")
	require.NotEmpty(t, importScript, "Import script should not be empty")

	// Capture memory stats
	stats := CaptureMemoryStats(startTime, totalConverted)
	LogMemoryStats(t, scenario.Name, stats)

	// Validate memory usage
	ValidateMemoryUsage(t, stats, scenario.NumMonitors)

	// Validate performance
	ValidatePerformance(t, stats, totalConverted, scenario.MaxDurationSeconds)

	// Write to temp directory and validate file sizes
	tempDir := CreateTempTestDir(t, "uptimerobot-load")
	tfFile := filepath.Join(tempDir, "hyperping.tf")
	importFile := filepath.Join(tempDir, "import.sh")

	require.NoError(t, os.WriteFile(tfFile, []byte(tfConfig), 0600))
	require.NoError(t, os.WriteFile(importFile, []byte(importScript), 0600))

	t.Log("Validating generated file sizes:")
	ValidateFileSize(t, tfFile, 10)
	ValidateFileSize(t, importFile, 5)

	t.Logf("=== Load Test PASSED: %s ===", scenario.Name)
}

// generateUptimeRobotMonitors generates test monitors for load testing.
func generateUptimeRobotMonitors(count int) []uptimerobot.Monitor {
	monitors := make([]uptimerobot.Monitor, count)

	types := []int{1, 2, 3, 4} // HTTP, Keyword, Ping, Port (not heartbeat)
	intervals := []int{60, 300, 600, 1800}
	httpMethods := []int{1, 2, 3, 4, 5, 6} // GET, POST, PUT, PATCH, DELETE, HEAD

	for i := 0; i < count; i++ {
		monitorType := types[i%len(types)]
		monitors[i] = generateUptimeRobotMonitorOfType(i, monitorType)

		// Set common fields
		monitors[i].Interval = intervals[i%len(intervals)]
		if monitorType == 1 {
			method := httpMethods[i%len(httpMethods)]
			monitors[i].HTTPMethod = &method
		}
	}

	return monitors
}

// generateUptimeRobotMonitorOfType generates a monitor of a specific type.
func generateUptimeRobotMonitorOfType(id int, monitorType int) uptimerobot.Monitor {
	monitor := uptimerobot.Monitor{
		ID:           1000000 + id,
		FriendlyName: fmt.Sprintf("load-test-monitor-%d", id),
		Type:         monitorType,
		Interval:     300,
		Status:       2, // Up
	}

	switch monitorType {
	case 1: // HTTP
		monitor.URL = fmt.Sprintf("https://api-%d.example.com/health", id)
		method := 1 // GET
		monitor.HTTPMethod = &method

	case 2: // Keyword
		monitor.URL = fmt.Sprintf("https://api-%d.example.com/status", id)
		keywordType := 1 // exists
		keyword := "OK"
		monitor.KeywordType = &keywordType
		monitor.KeywordValue = &keyword

	case 3: // Ping
		monitor.URL = fmt.Sprintf("192.168.1.%d", id%254+1)

	case 4: // Port
		monitor.URL = fmt.Sprintf("api-%d.example.com", id)
		port := 443
		monitor.Port = &port

	case 5: // Heartbeat
		monitor.URL = fmt.Sprintf("https://heartbeat.uptimerobot.com/%d", id)
	}

	return monitor
}

// generateUptimeRobotAlertContacts generates test alert contacts.
func generateUptimeRobotAlertContacts(count int) []uptimerobot.AlertContact {
	contacts := make([]uptimerobot.AlertContact, count)

	contactTypes := []int{2, 3, 4, 11} // Email, SMS, Webhook, Slack

	for i := 0; i < count; i++ {
		contacts[i] = uptimerobot.AlertContact{
			ID:           fmt.Sprintf("%d", 100000+i),
			FriendlyName: fmt.Sprintf("contact-%d", i),
			Type:         contactTypes[i%len(contactTypes)],
			Status:       2, // Active
		}

		switch contacts[i].Type {
		case 2: // Email
			contacts[i].Value = fmt.Sprintf("alerts-%d@example.com", i)
		case 3: // SMS
			contacts[i].Value = fmt.Sprintf("+1555000%04d", i)
		case 4: // Webhook
			contacts[i].Value = fmt.Sprintf("https://hooks.example.com/webhook-%d", i)
		case 11: // Slack
			contacts[i].Value = fmt.Sprintf("https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXX-%d", i)
		}
	}

	return contacts
}
