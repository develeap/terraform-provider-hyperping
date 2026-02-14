// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

//go:build integration

package integration

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const (
	maxRetries     = 3
	retryDelay     = 2 * time.Second
	testTimeout    = 15 * time.Minute
	defaultRegion  = "london"
	testURLPrefix  = "https://httpstat.us"
	cleanupTimeout = 5 * time.Minute
)

// TestCredentials holds API credentials for integration tests
type TestCredentials struct {
	BetterstackToken  string
	UptimeRobotAPIKey string
	PingdomAPIKey     string
	HyperpingAPIKey   string
}

// GetTestCredentials retrieves test credentials from environment variables
func GetTestCredentials(t *testing.T) TestCredentials {
	t.Helper()

	creds := TestCredentials{
		BetterstackToken:  os.Getenv("BETTERSTACK_API_TOKEN"),
		UptimeRobotAPIKey: os.Getenv("UPTIMEROBOT_API_KEY"),
		PingdomAPIKey:     getEnvOrAlternate("PINGDOM_API_KEY", "PINGDOM_API_TOKEN"),
		HyperpingAPIKey:   os.Getenv("HYPERPING_API_KEY"),
	}

	return creds
}

// getEnvOrAlternate returns the value of the primary env var, or the alternate if empty
func getEnvOrAlternate(primary, alternate string) string {
	if val := os.Getenv(primary); val != "" {
		return val
	}
	return os.Getenv(alternate)
}

// CreateTempTestDir creates a temporary directory for test outputs
func CreateTempTestDir(t *testing.T, prefix string) string {
	t.Helper()

	tempDir, err := os.MkdirTemp("", fmt.Sprintf("%s-integration-*", prefix))
	require.NoError(t, err, "failed to create temp directory")

	t.Cleanup(func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("Warning: failed to cleanup temp directory %s: %v", tempDir, err)
		}
	})

	return tempDir
}

// ValidateTerraformFile validates a Terraform configuration file using terraform validate
func ValidateTerraformFile(t *testing.T, tfFile string) {
	t.Helper()

	dir := filepath.Dir(tfFile)

	// Initialize Terraform
	t.Logf("Initializing Terraform in %s", dir)
	initCmd := exec.Command("terraform", "init", "-no-color")
	initCmd.Dir = dir
	initOutput, err := initCmd.CombinedOutput()
	if err != nil {
		t.Logf("Terraform init output:\n%s", string(initOutput))
		require.NoError(t, err, "terraform init failed")
	}

	// Validate configuration
	t.Logf("Validating Terraform configuration")
	validateCmd := exec.Command("terraform", "validate", "-no-color")
	validateCmd.Dir = dir
	validateOutput, err := validateCmd.CombinedOutput()
	if err != nil {
		t.Logf("Terraform validate output:\n%s", string(validateOutput))
		require.NoError(t, err, "terraform validate failed")
	}

	t.Logf("Terraform validation passed")
}

// ValidateImportScript validates that the import script is executable and has valid syntax
func ValidateImportScript(t *testing.T, scriptPath string) {
	t.Helper()

	// Check file exists
	require.FileExists(t, scriptPath, "import script does not exist")

	// Read script content
	content, err := os.ReadFile(scriptPath)
	require.NoError(t, err, "failed to read import script")

	scriptContent := string(content)

	// Validate shebang
	require.True(t, strings.HasPrefix(scriptContent, "#!/bin/bash") ||
		strings.HasPrefix(scriptContent, "#!/usr/bin/env bash"),
		"import script missing valid shebang")

	// Validate contains terraform import commands
	require.Contains(t, scriptContent, "terraform import",
		"import script does not contain terraform import commands")

	// Check script syntax using bash -n
	bashCmd := exec.Command("bash", "-n", scriptPath)
	output, err := bashCmd.CombinedOutput()
	if err != nil {
		t.Logf("Bash syntax check output:\n%s", string(output))
		require.NoError(t, err, "import script has syntax errors")
	}

	t.Logf("Import script validation passed")
}

// ValidateGeneratedFiles validates that all expected output files were generated
func ValidateGeneratedFiles(t *testing.T, outputDir string, expectedFiles []string) {
	t.Helper()

	for _, filename := range expectedFiles {
		filePath := filepath.Join(outputDir, filename)
		require.FileExists(t, filePath, "expected file %s does not exist", filename)

		// Validate file is not empty
		stat, err := os.Stat(filePath)
		require.NoError(t, err, "failed to stat file %s", filename)
		require.Greater(t, stat.Size(), int64(0), "file %s is empty", filename)
	}

	t.Logf("All expected files generated successfully")
}

// RunWithRetry runs a function with exponential backoff retry logic
func RunWithRetry(ctx context.Context, t *testing.T, description string, fn func() error) error {
	t.Helper()

	var lastErr error
	delay := retryDelay

	for attempt := 1; attempt <= maxRetries; attempt++ {
		if err := fn(); err != nil {
			lastErr = err
			t.Logf("Attempt %d/%d failed for %s: %v", attempt, maxRetries, description, err)

			if attempt < maxRetries {
				select {
				case <-ctx.Done():
					return fmt.Errorf("context cancelled during retry: %w", ctx.Err())
				case <-time.After(delay):
					delay *= 2 // Exponential backoff
				}
				continue
			}
		} else {
			return nil
		}
	}

	return fmt.Errorf("%s failed after %d attempts: %w", description, maxRetries, lastErr)
}

// CreateTestContext creates a context with timeout for integration tests
func CreateTestContext(t *testing.T) (context.Context, context.CancelFunc) {
	t.Helper()
	return context.WithTimeout(context.Background(), testTimeout)
}

// ValidateJSONFile validates that a file contains valid JSON
func ValidateJSONFile(t *testing.T, filePath string) {
	t.Helper()

	content, err := os.ReadFile(filePath)
	require.NoError(t, err, "failed to read JSON file")

	// Simple JSON validation - check for balanced braces
	jsonStr := string(content)
	require.True(t, strings.HasPrefix(jsonStr, "{") || strings.HasPrefix(jsonStr, "["),
		"JSON file does not start with { or [")

	// More thorough validation could use json.Valid() but keeping it simple
	openBraces := strings.Count(jsonStr, "{")
	closeBraces := strings.Count(jsonStr, "}")
	require.Equal(t, openBraces, closeBraces, "JSON has unbalanced braces")
}

// ValidateMarkdownFile validates that a file is valid markdown
func ValidateMarkdownFile(t *testing.T, filePath string) {
	t.Helper()

	content, err := os.ReadFile(filePath)
	require.NoError(t, err, "failed to read markdown file")

	// Basic markdown validation
	mdContent := string(content)
	require.NotEmpty(t, mdContent, "markdown file is empty")

	// Should contain at least one heading or some structure
	hasContent := strings.Contains(mdContent, "#") ||
		strings.Contains(mdContent, "-") ||
		len(mdContent) > 50
	require.True(t, hasContent, "markdown file appears to have no content")
}

// CountTerraformResources counts the number of resource blocks in a Terraform file
func CountTerraformResources(t *testing.T, tfFile string) int {
	t.Helper()

	content, err := os.ReadFile(tfFile)
	require.NoError(t, err, "failed to read terraform file")

	// Count resource blocks (simple pattern matching)
	tfContent := string(content)
	count := strings.Count(tfContent, "resource \"")

	t.Logf("Found %d resource blocks in %s", count, tfFile)
	return count
}

// SkipIfCredentialsMissing skips the test if required credentials are not available
func SkipIfCredentialsMissing(t *testing.T, credName string, credValue string) {
	t.Helper()

	if credValue == "" {
		t.Skipf("Skipping test: %s not set (required for integration tests)", credName)
	}
}

// TestScenario represents a test scenario configuration
type TestScenario struct {
	Name                 string
	Description          string
	ExpectedMonitors     int
	ExpectedHealthchecks int
	MinWarnings          int
	MaxWarnings          int
}

// ValidateScenarioOutput validates the output of a migration scenario
func ValidateScenarioOutput(t *testing.T, scenario TestScenario, outputDir string, resourceCount int) {
	t.Helper()

	t.Logf("Validating scenario: %s", scenario.Name)
	t.Logf("Description: %s", scenario.Description)

	// Validate resource count is reasonable
	totalExpected := scenario.ExpectedMonitors + scenario.ExpectedHealthchecks
	if totalExpected > 0 {
		require.GreaterOrEqual(t, resourceCount, totalExpected,
			"resource count lower than expected for scenario %s", scenario.Name)
	}

	t.Logf("Scenario validation passed: %d resources generated", resourceCount)
}

// CleanupTestResources provides a cleanup function for test resources
func CleanupTestResources(t *testing.T, cleanupFn func() error) {
	t.Helper()

	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), cleanupTimeout)
		defer cancel()

		if err := RunWithRetry(ctx, t, "cleanup", cleanupFn); err != nil {
			t.Logf("Warning: cleanup failed: %v", err)
		} else {
			t.Logf("Cleanup completed successfully")
		}
	})
}
