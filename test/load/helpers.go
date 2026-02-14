// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

//go:build load

package load

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const (
	defaultTimeout = 10 * time.Minute
	maxMemoryMB    = 500 // Maximum memory usage for 100 monitors
)

// MemoryStats captures memory usage statistics.
type MemoryStats struct {
	AllocMB        float64
	TotalAllocMB   float64
	SysMB          float64
	NumGC          uint32
	HeapObjectsK   uint64
	StartTime      time.Time
	EndTime        time.Time
	DurationMS     int64
	MonitorsPerSec float64
}

// CaptureMemoryStats captures current memory statistics.
func CaptureMemoryStats(startTime time.Time, numMonitors int) MemoryStats {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	endTime := time.Now()
	durationMS := endTime.Sub(startTime).Milliseconds()
	monitorsPerSec := float64(0)
	if durationMS > 0 {
		monitorsPerSec = float64(numMonitors) / (float64(durationMS) / 1000.0)
	}

	return MemoryStats{
		AllocMB:        float64(m.Alloc) / 1024 / 1024,
		TotalAllocMB:   float64(m.TotalAlloc) / 1024 / 1024,
		SysMB:          float64(m.Sys) / 1024 / 1024,
		NumGC:          m.NumGC,
		HeapObjectsK:   m.HeapObjects / 1000,
		StartTime:      startTime,
		EndTime:        endTime,
		DurationMS:     durationMS,
		MonitorsPerSec: monitorsPerSec,
	}
}

// LogMemoryStats logs memory statistics to test output.
func LogMemoryStats(t *testing.T, scenario string, stats MemoryStats) {
	t.Helper()
	t.Logf("=== Memory Stats for %s ===", scenario)
	t.Logf("  Allocated: %.2f MB", stats.AllocMB)
	t.Logf("  Total Allocated: %.2f MB", stats.TotalAllocMB)
	t.Logf("  System Memory: %.2f MB", stats.SysMB)
	t.Logf("  GC Runs: %d", stats.NumGC)
	t.Logf("  Heap Objects: %dk", stats.HeapObjectsK)
	t.Logf("  Duration: %dms", stats.DurationMS)
	t.Logf("  Throughput: %.2f monitors/sec", stats.MonitorsPerSec)
}

// ValidateMemoryUsage ensures memory usage is within acceptable limits.
func ValidateMemoryUsage(t *testing.T, stats MemoryStats, numMonitors int) {
	t.Helper()

	// Check peak memory usage
	if numMonitors >= 100 {
		require.LessOrEqual(t, stats.AllocMB, float64(maxMemoryMB),
			"Memory usage exceeds %d MB for %d monitors", maxMemoryMB, numMonitors)
	}

	// Check for reasonable memory usage per monitor
	memPerMonitor := stats.AllocMB / float64(numMonitors)
	require.LessOrEqual(t, memPerMonitor, 5.0,
		"Memory per monitor (%.2f MB) exceeds 5 MB", memPerMonitor)

	t.Logf("✅ Memory validation passed: %.2f MB for %d monitors (%.2f MB/monitor)",
		stats.AllocMB, numMonitors, memPerMonitor)
}

// ValidatePerformance ensures performance is within acceptable limits.
func ValidatePerformance(t *testing.T, stats MemoryStats, numMonitors int, maxDurationSeconds int) {
	t.Helper()

	durationSeconds := stats.DurationMS / 1000

	// Validate total execution time
	require.LessOrEqual(t, durationSeconds, int64(maxDurationSeconds),
		"Execution time (%ds) exceeds maximum (%ds) for %d monitors",
		durationSeconds, maxDurationSeconds, numMonitors)

	// Validate throughput is reasonable (at least 0.5 monitors per second)
	// Note: Very fast operations (< 1ms) may show 0 throughput due to timing precision
	if stats.DurationMS > 0 {
		require.GreaterOrEqual(t, stats.MonitorsPerSec, 0.5,
			"Throughput (%.2f monitors/sec) is too low", stats.MonitorsPerSec)
	}

	t.Logf("✅ Performance validation passed: %dms for %d monitors (%.2f monitors/sec)",
		stats.DurationMS, numMonitors, stats.MonitorsPerSec)
}

// ValidateFileSize ensures generated files are within reasonable size limits.
func ValidateFileSize(t *testing.T, filePath string, maxSizeMB int) int64 {
	t.Helper()

	stat, err := os.Stat(filePath)
	require.NoError(t, err, "failed to stat file %s", filePath)

	sizeMB := float64(stat.Size()) / 1024 / 1024
	require.LessOrEqual(t, sizeMB, float64(maxSizeMB),
		"File %s size (%.2f MB) exceeds maximum (%d MB)",
		filepath.Base(filePath), sizeMB, maxSizeMB)

	t.Logf("  %s: %.2f MB", filepath.Base(filePath), sizeMB)
	return stat.Size()
}

// LoadTestScenario represents a load testing scenario.
type LoadTestScenario struct {
	Name                  string
	Description           string
	NumMonitors           int
	NumHealthchecks       int
	MaxDurationSeconds    int
	MaxMemoryMB           int
	ExpectedThroughput    float64 // monitors/sec
	ValidateLinearScaling bool
}

// BenchmarkResult captures benchmark results for comparison.
type BenchmarkResult struct {
	Scenario         string
	NumMonitors      int
	NumHealthchecks  int
	DurationMS       int64
	MemoryMB         float64
	ThroughputPerSec float64
	FileSize         map[string]int64
	Timestamp        time.Time
}

// SaveBenchmarkResults saves benchmark results to JSON file.
func SaveBenchmarkResults(t *testing.T, results []BenchmarkResult, outputPath string) {
	t.Helper()

	data, err := json.MarshalIndent(results, "", "  ")
	require.NoError(t, err, "failed to marshal benchmark results")

	err = os.WriteFile(outputPath, data, 0600)
	require.NoError(t, err, "failed to write benchmark results to %s", outputPath)

	t.Logf("Benchmark results saved to %s", outputPath)
}

// CreateTestContext creates a context with timeout for load tests.
func CreateTestContext(t *testing.T) (context.Context, context.CancelFunc) {
	t.Helper()
	return context.WithTimeout(context.Background(), defaultTimeout)
}

// CreateTempTestDir creates a temporary directory for test outputs.
func CreateTempTestDir(t *testing.T, prefix string) string {
	t.Helper()

	tempDir, err := os.MkdirTemp("", fmt.Sprintf("%s-load-*", prefix))
	require.NoError(t, err, "failed to create temp directory")

	t.Cleanup(func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("Warning: failed to cleanup temp directory %s: %v", tempDir, err)
		}
	})

	return tempDir
}

// RateLimitingStats captures rate limiting behavior.
type RateLimitingStats struct {
	TotalRequests    int
	RateLimited      int
	RetriesAttempted int
	TotalWaitTimeMS  int64
	MaxRetryDelayMS  int64
}

// MockRateLimitServer creates a mock server that simulates rate limiting.
func MockRateLimitServer(t *testing.T, rateLimitAfter int, retryAfterSeconds int) *httptest.Server {
	t.Helper()

	requestCount := 0
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++

		if requestCount > rateLimitAfter {
			// Reset after rate limit to allow eventual success
			if requestCount > rateLimitAfter+3 {
				requestCount = 0
			}

			w.Header().Set("Retry-After", fmt.Sprintf("%d", retryAfterSeconds))
			w.Header().Set("X-RateLimit-Remaining", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"error": "Too Many Requests"}`))
			return
		}

		// Successful response
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", rateLimitAfter-requestCount))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data": []}`))
	}))
}

// GenerateTestFixture generates a large test fixture file for load testing.
func GenerateTestFixture(t *testing.T, outputPath string, numMonitors int, numHealthchecks int, dataGenerator func(int, int) interface{}) {
	t.Helper()

	data := dataGenerator(numMonitors, numHealthchecks)

	jsonData, err := json.MarshalIndent(data, "", "  ")
	require.NoError(t, err, "failed to marshal test fixture")

	err = os.WriteFile(outputPath, jsonData, 0600)
	require.NoError(t, err, "failed to write test fixture to %s", outputPath)

	stat, _ := os.Stat(outputPath)
	sizeMB := float64(stat.Size()) / 1024 / 1024
	t.Logf("Generated test fixture: %s (%.2f MB, %d monitors, %d healthchecks)",
		outputPath, sizeMB, numMonitors, numHealthchecks)
}

// ProfileInfo captures profiling information for analysis.
type ProfileInfo struct {
	CPUProfilePath   string
	MemProfilePath   string
	GoroutineProfile string
	BlockProfilePath string
}

// ForceGC triggers garbage collection and waits for it to complete.
func ForceGC(t *testing.T) {
	t.Helper()
	runtime.GC()
	time.Sleep(100 * time.Millisecond)
}

// CheckMemoryLeak compares memory usage before and after operations.
func CheckMemoryLeak(t *testing.T, beforeAlloc, afterAlloc uint64, operationName string) {
	t.Helper()

	diff := int64(afterAlloc - beforeAlloc)
	diffMB := float64(diff) / 1024 / 1024

	// Allow some tolerance (10 MB) for memory fluctuations
	if diffMB > 10 {
		t.Errorf("Potential memory leak detected in %s: %.2f MB increase", operationName, diffMB)
	} else {
		t.Logf("✅ No memory leak detected in %s (%.2f MB change)", operationName, diffMB)
	}
}

// CompareScalingLinear validates that execution time scales linearly with data size.
func CompareScalingLinear(t *testing.T, baselineMonitors int, baselineTime int64, currentMonitors int, currentTime int64) {
	t.Helper()

	// Skip validation if times are too small to measure accurately (< 1ms)
	if baselineTime < 1 && currentTime < 1 {
		t.Logf("Scaling analysis skipped: execution times too small to measure accurately (< 1ms)")
		return
	}

	// Calculate expected time based on linear scaling
	scale := float64(currentMonitors) / float64(baselineMonitors)
	expectedTime := float64(baselineTime) * scale

	// Allow 50% tolerance for variance, or at least 1ms for very fast operations
	tolerance := 0.5
	if expectedTime < 2 {
		tolerance = 2.0 // Allow more variance for sub-2ms operations
	}

	lowerBound := expectedTime * (1.0 - tolerance)
	if lowerBound < 0 {
		lowerBound = 0
	}
	upperBound := expectedTime * (1.0 + tolerance)

	actualTime := float64(currentTime)

	t.Logf("Scaling analysis: %dx data, expected %dms (±%.0f%%), actual %dms",
		int(scale), int(expectedTime), tolerance*100, currentTime)

	if baselineTime > 0 {
		require.GreaterOrEqual(t, actualTime, lowerBound,
			"Execution time (%dms) is suspiciously fast (expected ~%dms)",
			currentTime, int(expectedTime))

		require.LessOrEqual(t, actualTime, upperBound,
			"Execution time (%dms) exceeds linear scaling expectation (%dms)",
			currentTime, int(expectedTime))

		timeRatio := actualTime / float64(baselineTime)
		if baselineTime > 0 {
			t.Logf("✅ Performance scales linearly: %dx data took %.2fx time", int(scale), timeRatio)
		}
	} else {
		t.Logf("✅ Performance excellent: sub-millisecond execution")
	}
}
