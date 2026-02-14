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

	bsconverter "github.com/develeap/terraform-provider-hyperping/cmd/migrate-betterstack/converter"
	pdconverter "github.com/develeap/terraform-provider-hyperping/cmd/migrate-pingdom/converter"
	urconverter "github.com/develeap/terraform-provider-hyperping/cmd/migrate-uptimerobot/converter"
)

// TestLoadSuite runs a comprehensive load test suite across all platforms.
func TestLoadSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping comprehensive load test suite in short mode")
	}

	t.Log("=== Running Comprehensive Load Test Suite ===")
	startTime := time.Now()

	var allResults []BenchmarkResult

	// Test scenarios for each platform
	scenarios := []struct {
		name     string
		platform string
		run      func(*testing.T, LoadTestScenario)
	}{
		{"BetterStack_10", "BetterStack", runBetterStackLoadTest},
		{"BetterStack_50", "BetterStack", runBetterStackLoadTest},
		{"BetterStack_100", "BetterStack", runBetterStackLoadTest},
		{"UptimeRobot_10", "UptimeRobot", runUptimeRobotLoadTest},
		{"UptimeRobot_50", "UptimeRobot", runUptimeRobotLoadTest},
		{"UptimeRobot_100", "UptimeRobot", runUptimeRobotLoadTest},
		{"Pingdom_10", "Pingdom", runPingdomLoadTest},
		{"Pingdom_50", "Pingdom", runPingdomLoadTest},
		{"Pingdom_100", "Pingdom", runPingdomLoadTest},
	}

	scenarioConfigs := map[string]LoadTestScenario{
		"BetterStack_10": {
			Name:               "BetterStack_10monitors",
			NumMonitors:        10,
			NumHealthchecks:    5,
			MaxDurationSeconds: 30,
		},
		"BetterStack_50": {
			Name:               "BetterStack_50monitors",
			NumMonitors:        50,
			NumHealthchecks:    25,
			MaxDurationSeconds: 60,
		},
		"BetterStack_100": {
			Name:               "BetterStack_100monitors",
			NumMonitors:        100,
			NumHealthchecks:    50,
			MaxDurationSeconds: 120,
		},
		"UptimeRobot_10": {
			Name:               "UptimeRobot_10monitors",
			NumMonitors:        10,
			NumHealthchecks:    5,
			MaxDurationSeconds: 30,
		},
		"UptimeRobot_50": {
			Name:               "UptimeRobot_50monitors",
			NumMonitors:        50,
			NumHealthchecks:    25,
			MaxDurationSeconds: 60,
		},
		"UptimeRobot_100": {
			Name:               "UptimeRobot_100monitors",
			NumMonitors:        100,
			NumHealthchecks:    50,
			MaxDurationSeconds: 120,
		},
		"Pingdom_10": {
			Name:               "Pingdom_10checks",
			NumMonitors:        10,
			NumHealthchecks:    0,
			MaxDurationSeconds: 30,
		},
		"Pingdom_50": {
			Name:               "Pingdom_50checks",
			NumMonitors:        50,
			NumHealthchecks:    0,
			MaxDurationSeconds: 60,
		},
		"Pingdom_100": {
			Name:               "Pingdom_100checks",
			NumMonitors:        100,
			NumHealthchecks:    0,
			MaxDurationSeconds: 120,
		},
	}

	// Run all scenarios
	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			config := scenarioConfigs[scenario.name]

			testStartTime := time.Now()
			scenario.run(t, config)
			testDuration := time.Since(testStartTime)

			// Capture result
			result := BenchmarkResult{
				Scenario:        scenario.name,
				NumMonitors:     config.NumMonitors,
				NumHealthchecks: config.NumHealthchecks,
				DurationMS:      testDuration.Milliseconds(),
				Timestamp:       time.Now(),
			}
			allResults = append(allResults, result)
		})
	}

	// Generate summary report
	totalDuration := time.Since(startTime)
	generateLoadTestReport(t, allResults, totalDuration)

	t.Logf("=== Load Test Suite Complete: %s ===", totalDuration.Round(time.Second))
}

// TestScalingComparison compares scaling characteristics across platforms.
func TestScalingComparison(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping scaling comparison in short mode")
	}

	t.Log("=== Scaling Comparison Test ===")

	// Test data points: 10, 50, 100 monitors
	scales := []int{10, 50, 100}
	results := make(map[string][]MemoryStats)

	// Better Stack
	t.Run("BetterStack_Scaling", func(t *testing.T) {
		var scaleResults []MemoryStats
		for _, scale := range scales {
			ForceGC(t)
			startTime := time.Now()

			monitors := generateBetterStackMonitors(scale)
			heartbeats := generateBetterStackHeartbeats(scale / 2)

			conv := bsconverter.New()
			_, _ = conv.ConvertMonitors(monitors)
			_, _ = conv.ConvertHeartbeats(heartbeats)

			stats := CaptureMemoryStats(startTime, scale)
			scaleResults = append(scaleResults, stats)
			t.Logf("  %d monitors: %.2f MB, %dms", scale, stats.AllocMB, stats.DurationMS)
		}
		results["BetterStack"] = scaleResults

		// Validate linear scaling
		if len(scaleResults) >= 2 {
			baseline := scaleResults[0]
			for i := 1; i < len(scaleResults); i++ {
				CompareScalingLinear(t, scales[0], baseline.DurationMS, scales[i], scaleResults[i].DurationMS)
			}
		}
	})

	// UptimeRobot
	t.Run("UptimeRobot_Scaling", func(t *testing.T) {
		var scaleResults []MemoryStats
		for _, scale := range scales {
			ForceGC(t)
			startTime := time.Now()

			monitors := generateUptimeRobotMonitors(scale)
			alertContacts := generateUptimeRobotAlertContacts(10)

			conv := urconverter.NewConverter()
			_ = conv.Convert(monitors, alertContacts)

			stats := CaptureMemoryStats(startTime, scale)
			scaleResults = append(scaleResults, stats)
			t.Logf("  %d monitors: %.2f MB, %dms", scale, stats.AllocMB, stats.DurationMS)
		}
		results["UptimeRobot"] = scaleResults

		// Validate linear scaling
		if len(scaleResults) >= 2 {
			baseline := scaleResults[0]
			for i := 1; i < len(scaleResults); i++ {
				CompareScalingLinear(t, scales[0], baseline.DurationMS, scales[i], scaleResults[i].DurationMS)
			}
		}
	})

	// Pingdom
	t.Run("Pingdom_Scaling", func(t *testing.T) {
		var scaleResults []MemoryStats
		for _, scale := range scales {
			ForceGC(t)
			startTime := time.Now()

			checks := generatePingdomChecks(scale)

			conv := pdconverter.NewCheckConverter()
			for _, check := range checks {
				_ = conv.Convert(check)
			}

			stats := CaptureMemoryStats(startTime, scale)
			scaleResults = append(scaleResults, stats)
			t.Logf("  %d checks: %.2f MB, %dms", scale, stats.AllocMB, stats.DurationMS)
		}
		results["Pingdom"] = scaleResults

		// Validate linear scaling
		if len(scaleResults) >= 2 {
			baseline := scaleResults[0]
			for i := 1; i < len(scaleResults); i++ {
				CompareScalingLinear(t, scales[0], baseline.DurationMS, scales[i], scaleResults[i].DurationMS)
			}
		}
	})

	// Generate comparison report
	generateScalingReport(t, scales, results)

	t.Log("✅ Scaling comparison complete")
}

// TestMemoryLeakDetection runs extended memory leak detection across all platforms.
func TestMemoryLeakDetection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping memory leak detection in short mode")
	}

	t.Log("=== Memory Leak Detection ===")

	iterations := 10
	checkInterval := 50 // monitors per iteration

	// Better Stack
	t.Run("BetterStack_MemoryLeak", func(t *testing.T) {
		ForceGC(t)
		var startMem, prevMem runtime.MemStats
		runtime.ReadMemStats(&startMem)
		prevMem = startMem

		monitors := generateBetterStackMonitors(checkInterval)
		heartbeats := generateBetterStackHeartbeats(checkInterval / 2)
		conv := bsconverter.New()

		for i := 0; i < iterations; i++ {
			_, _ = conv.ConvertMonitors(monitors)
			_, _ = conv.ConvertHeartbeats(heartbeats)
			ForceGC(t)

			var currentMem runtime.MemStats
			runtime.ReadMemStats(&currentMem)

			increase := float64(currentMem.Alloc-prevMem.Alloc) / 1024 / 1024
			t.Logf("  Iteration %d: %.2f MB (increase: %.2f MB)", i+1,
				float64(currentMem.Alloc)/1024/1024, increase)

			prevMem = currentMem
		}

		var finalMem runtime.MemStats
		runtime.ReadMemStats(&finalMem)
		CheckMemoryLeak(t, startMem.Alloc, finalMem.Alloc, "BetterStack (extended)")
	})

	// UptimeRobot
	t.Run("UptimeRobot_MemoryLeak", func(t *testing.T) {
		ForceGC(t)
		var startMem, prevMem runtime.MemStats
		runtime.ReadMemStats(&startMem)
		prevMem = startMem

		monitors := generateUptimeRobotMonitors(checkInterval)
		alertContacts := generateUptimeRobotAlertContacts(10)
		conv := urconverter.NewConverter()

		for i := 0; i < iterations; i++ {
			_ = conv.Convert(monitors, alertContacts)
			ForceGC(t)

			var currentMem runtime.MemStats
			runtime.ReadMemStats(&currentMem)

			increase := float64(currentMem.Alloc-prevMem.Alloc) / 1024 / 1024
			t.Logf("  Iteration %d: %.2f MB (increase: %.2f MB)", i+1,
				float64(currentMem.Alloc)/1024/1024, increase)

			prevMem = currentMem
		}

		var finalMem runtime.MemStats
		runtime.ReadMemStats(&finalMem)
		CheckMemoryLeak(t, startMem.Alloc, finalMem.Alloc, "UptimeRobot (extended)")
	})

	// Pingdom
	t.Run("Pingdom_MemoryLeak", func(t *testing.T) {
		ForceGC(t)
		var startMem, prevMem runtime.MemStats
		runtime.ReadMemStats(&startMem)
		prevMem = startMem

		checks := generatePingdomChecks(checkInterval)
		conv := pdconverter.NewCheckConverter()

		for i := 0; i < iterations; i++ {
			for _, check := range checks {
				_ = conv.Convert(check)
			}
			ForceGC(t)

			var currentMem runtime.MemStats
			runtime.ReadMemStats(&currentMem)

			increase := float64(currentMem.Alloc-prevMem.Alloc) / 1024 / 1024
			t.Logf("  Iteration %d: %.2f MB (increase: %.2f MB)", i+1,
				float64(currentMem.Alloc)/1024/1024, increase)

			prevMem = currentMem
		}

		var finalMem runtime.MemStats
		runtime.ReadMemStats(&finalMem)
		CheckMemoryLeak(t, startMem.Alloc, finalMem.Alloc, "Pingdom (extended)")
	})

	t.Log("✅ Memory leak detection complete")
}

// generateLoadTestReport generates a comprehensive report of load test results.
func generateLoadTestReport(t *testing.T, results []BenchmarkResult, totalDuration time.Duration) {
	t.Helper()

	t.Log("\n=== LOAD TEST REPORT ===")
	t.Logf("Total Duration: %s", totalDuration.Round(time.Second))
	t.Logf("Total Scenarios: %d", len(results))

	// Group by platform
	platforms := make(map[string][]BenchmarkResult)
	for _, r := range results {
		platform := ""
		if len(r.Scenario) > 0 {
			// Extract platform name from scenario
			for _, p := range []string{"BetterStack", "UptimeRobot", "Pingdom"} {
				if len(r.Scenario) >= len(p) && r.Scenario[:len(p)] == p {
					platform = p
					break
				}
			}
		}
		if platform != "" {
			platforms[platform] = append(platforms[platform], r)
		}
	}

	// Print per-platform summary
	for platform, platformResults := range platforms {
		t.Logf("\n--- %s ---", platform)
		for _, r := range platformResults {
			totalResources := r.NumMonitors + r.NumHealthchecks
			throughput := float64(0)
			if r.DurationMS > 0 {
				throughput = float64(totalResources) / (float64(r.DurationMS) / 1000.0)
			}
			t.Logf("  %s: %d resources in %dms (%.2f resources/sec)",
				r.Scenario, totalResources, r.DurationMS, throughput)
		}
	}

	// Save results to file
	tempDir, err := os.MkdirTemp("", "load-test-results-*")
	if err != nil {
		t.Logf("Warning: could not create temp dir for results: %v", err)
		return
	}

	resultsFile := filepath.Join(tempDir, fmt.Sprintf("load-test-results-%s.json",
		time.Now().Format("20060102-150405")))

	SaveBenchmarkResults(t, results, resultsFile)
	t.Logf("\nDetailed results saved to: %s", resultsFile)
}

// generateScalingReport generates a scaling comparison report.
func generateScalingReport(t *testing.T, scales []int, results map[string][]MemoryStats) {
	t.Helper()

	t.Log("\n=== SCALING ANALYSIS ===")

	for platform, stats := range results {
		t.Logf("\n%s:", platform)
		t.Log("Scale | Memory (MB) | Duration (ms) | Throughput (res/sec)")
		t.Log("------|-------------|---------------|----------------------")

		for i, scale := range scales {
			if i < len(stats) {
				t.Logf("%5d | %11.2f | %13d | %20.2f",
					scale, stats[i].AllocMB, stats[i].DurationMS, stats[i].MonitorsPerSec)
			}
		}

		// Calculate scaling factors
		if len(stats) >= 2 {
			memScaling := stats[len(stats)-1].AllocMB / stats[0].AllocMB
			timeScaling := float64(stats[len(stats)-1].DurationMS) / float64(stats[0].DurationMS)
			dataScaling := float64(scales[len(scales)-1]) / float64(scales[0])

			t.Logf("\nScaling factors (%dx data):", int(dataScaling))
			t.Logf("  Memory: %.2fx (expected: %.2fx) - %s", memScaling, dataScaling,
				scalingVerdict(memScaling, dataScaling))
			t.Logf("  Time: %.2fx (expected: %.2fx) - %s", timeScaling, dataScaling,
				scalingVerdict(timeScaling, dataScaling))
		}
	}
}

// scalingVerdict provides a verdict on scaling performance.
func scalingVerdict(actual, expected float64) string {
	ratio := actual / expected
	if ratio < 0.8 {
		return "✅ Better than linear (excellent)"
	} else if ratio < 1.2 {
		return "✅ Linear (good)"
	} else if ratio < 1.5 {
		return "⚠️ Slightly worse than linear"
	}
	return "❌ Much worse than linear (investigate)"
}
