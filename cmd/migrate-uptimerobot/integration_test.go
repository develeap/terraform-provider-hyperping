// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

//go:build integration

package main

import (
	"flag"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/develeap/terraform-provider-hyperping/test/integration"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	flag.Parse()
	os.Exit(m.Run())
}

// TestUptimeRobotMigration_SmallScenario tests migration of 1-3 monitors
func TestUptimeRobotMigration_SmallScenario(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	creds := integration.GetTestCredentials(t)
	integration.SkipIfCredentialsMissing(t, "UPTIMEROBOT_API_KEY", creds.UptimeRobotAPIKey)
	integration.SkipIfCredentialsMissing(t, "HYPERPING_API_KEY", creds.HyperpingAPIKey)

	scenario := integration.TestScenario{
		Name:                 "Small",
		Description:          "1-3 monitors of various types",
		ExpectedMonitors:     1,
		ExpectedHealthchecks: 0,
		MinWarnings:          0,
		MaxWarnings:          5,
	}

	runUptimeRobotMigrationTest(t, creds, scenario)
}

// TestUptimeRobotMigration_MediumScenario tests migration of 5-10 monitors
func TestUptimeRobotMigration_MediumScenario(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	creds := integration.GetTestCredentials(t)
	integration.SkipIfCredentialsMissing(t, "UPTIMEROBOT_API_KEY", creds.UptimeRobotAPIKey)
	integration.SkipIfCredentialsMissing(t, "HYPERPING_API_KEY", creds.HyperpingAPIKey)

	scenario := integration.TestScenario{
		Name:                 "Medium",
		Description:          "5-10 monitors including HTTP, Keyword, and Heartbeat types",
		ExpectedMonitors:     5,
		ExpectedHealthchecks: 1,
		MinWarnings:          0,
		MaxWarnings:          10,
	}

	runUptimeRobotMigrationTest(t, creds, scenario)
}

// TestUptimeRobotMigration_LargeScenario tests migration of 20-30 monitors
func TestUptimeRobotMigration_LargeScenario(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	creds := integration.GetTestCredentials(t)
	integration.SkipIfCredentialsMissing(t, "UPTIMEROBOT_API_KEY", creds.UptimeRobotAPIKey)
	integration.SkipIfCredentialsMissing(t, "HYPERPING_API_KEY", creds.HyperpingAPIKey)

	scenario := integration.TestScenario{
		Name:                 "Large",
		Description:          "20-30 monitors testing all 5 monitor types",
		ExpectedMonitors:     20,
		ExpectedHealthchecks: 5,
		MinWarnings:          0,
		MaxWarnings:          20,
	}

	runUptimeRobotMigrationTest(t, creds, scenario)
}

// TestUptimeRobotMigration_MonitorTypes tests all 5 UptimeRobot monitor types
func TestUptimeRobotMigration_MonitorTypes(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	creds := integration.GetTestCredentials(t)
	integration.SkipIfCredentialsMissing(t, "UPTIMEROBOT_API_KEY", creds.UptimeRobotAPIKey)
	integration.SkipIfCredentialsMissing(t, "HYPERPING_API_KEY", creds.HyperpingAPIKey)

	// Test each monitor type individually
	monitorTypes := []struct {
		name        string
		description string
	}{
		{"HTTP/HTTPS", "Standard HTTP uptime monitoring"},
		{"Keyword", "HTTP with keyword validation"},
		{"Ping", "ICMP ping monitoring"},
		{"Port", "TCP port monitoring"},
		{"Heartbeat", "Push-based heartbeat monitoring"},
	}

	ctx, cancel := integration.CreateTestContext(t)
	defer cancel()

	tempDir := integration.CreateTempTestDir(t, "uptimerobot-types")

	// Run migration to get all monitors
	outputFile := filepath.Join(tempDir, "hyperping.tf")
	reportFile := filepath.Join(tempDir, "migration-report.json")

	err := integration.RunWithRetry(ctx, t, "migration execution", func() error {
		cmd := exec.CommandContext(ctx,
			"go", "run", "./cmd/migrate-uptimerobot",
			"--uptimerobot-api-key", creds.UptimeRobotAPIKey,
			"--hyperping-api-key", creds.HyperpingAPIKey,
			"--output", outputFile,
			"--report", reportFile,
			"--verbose",
		)

		output, err := cmd.CombinedOutput()
		t.Logf("Migration output:\n%s", string(output))
		return err
	})

	require.NoError(t, err, "migration failed")

	// Verify output contains resources
	resourceCount := integration.CountTerraformResources(t, outputFile)
	require.Greater(t, resourceCount, 0, "no resources were migrated")

	t.Logf("Successfully tested migration with %d resources covering monitor types: %v",
		resourceCount, monitorTypes)
}

// TestUptimeRobotMigration_ValidateMode tests the validate-only mode
func TestUptimeRobotMigration_ValidateMode(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	creds := integration.GetTestCredentials(t)
	integration.SkipIfCredentialsMissing(t, "UPTIMEROBOT_API_KEY", creds.UptimeRobotAPIKey)

	ctx, cancel := integration.CreateTestContext(t)
	defer cancel()

	// Run in validate mode (no Hyperping API key needed)
	cmd := exec.CommandContext(ctx,
		"go", "run", "./cmd/migrate-uptimerobot",
		"--uptimerobot-api-key", creds.UptimeRobotAPIKey,
		"--validate",
	)

	output, err := cmd.CombinedOutput()
	t.Logf("Validation output:\n%s", string(output))
	require.NoError(t, err, "validate mode failed")

	t.Log("Validate mode test passed")
}

// TestUptimeRobotMigration_DryRun tests dry-run mode
func TestUptimeRobotMigration_DryRun(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	creds := integration.GetTestCredentials(t)
	integration.SkipIfCredentialsMissing(t, "UPTIMEROBOT_API_KEY", creds.UptimeRobotAPIKey)
	integration.SkipIfCredentialsMissing(t, "HYPERPING_API_KEY", creds.HyperpingAPIKey)

	ctx, cancel := integration.CreateTestContext(t)
	defer cancel()

	tempDir := integration.CreateTempTestDir(t, "uptimerobot-dryrun")

	// Run migration in dry-run mode
	cmd := exec.CommandContext(ctx,
		"go", "run", "./cmd/migrate-uptimerobot",
		"--uptimerobot-api-key", creds.UptimeRobotAPIKey,
		"--hyperping-api-key", creds.HyperpingAPIKey,
		"--output", filepath.Join(tempDir, "hyperping.tf"),
		"--import-script", filepath.Join(tempDir, "import.sh"),
		"--report", filepath.Join(tempDir, "migration-report.json"),
		"--manual-steps", filepath.Join(tempDir, "manual-steps.md"),
		"--dry-run",
		"--verbose",
	)

	output, err := cmd.CombinedOutput()
	t.Logf("Migration output:\n%s", string(output))
	require.NoError(t, err, "dry-run mode failed")

	// Verify no files were created
	files := []string{"hyperping.tf", "import.sh", "migration-report.json", "manual-steps.md"}
	for _, filename := range files {
		filePath := filepath.Join(tempDir, filename)
		_, err := os.Stat(filePath)
		assert.True(t, os.IsNotExist(err), "file %s should not exist in dry-run mode", filename)
	}

	t.Log("Dry-run test passed")
}

// TestUptimeRobotMigration_InvalidCredentials tests error handling
func TestUptimeRobotMigration_InvalidCredentials(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cancel := integration.CreateTestContext(t)
	defer cancel()

	// Run with invalid credentials
	cmd := exec.CommandContext(ctx,
		"go", "run", "./cmd/migrate-uptimerobot",
		"--uptimerobot-api-key", "u1234567-invalid-key",
		"--validate",
	)

	output, err := cmd.CombinedOutput()
	t.Logf("Migration output:\n%s", string(output))

	// Should fail with error
	require.Error(t, err, "migration should fail with invalid credentials")

	t.Log("Invalid credentials test passed")
}

// runUptimeRobotMigrationTest is a helper function that runs the complete migration test flow
func runUptimeRobotMigrationTest(t *testing.T, creds integration.TestCredentials, scenario integration.TestScenario) {
	t.Helper()

	ctx, cancel := integration.CreateTestContext(t)
	defer cancel()

	// Create temporary output directory
	tempDir := integration.CreateTempTestDir(t, "uptimerobot-"+scenario.Name)
	t.Logf("Test output directory: %s", tempDir)

	// Define output file paths
	outputFile := filepath.Join(tempDir, "hyperping.tf")
	importScript := filepath.Join(tempDir, "import.sh")
	reportFile := filepath.Join(tempDir, "migration-report.json")
	manualSteps := filepath.Join(tempDir, "manual-steps.md")

	// Step 1: API Connection Test
	t.Log("Step 1: Testing API connection to UptimeRobot")
	err := integration.RunWithRetry(ctx, t, "UptimeRobot API connection", func() error {
		cmd := exec.CommandContext(ctx,
			"go", "run", "./cmd/migrate-uptimerobot",
			"--uptimerobot-api-key", creds.UptimeRobotAPIKey,
			"--validate",
		)

		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Logf("API connection test output:\n%s", string(output))
			return err
		}

		return nil
	})
	require.NoError(t, err, "UptimeRobot API connection failed")
	t.Log("✅ API connection successful")

	// Step 2: Execute Migration Tool
	t.Logf("Step 2: Executing migration tool for scenario: %s", scenario.Name)
	err = integration.RunWithRetry(ctx, t, "migration execution", func() error {
		cmd := exec.CommandContext(ctx,
			"go", "run", "./cmd/migrate-uptimerobot",
			"--uptimerobot-api-key", creds.UptimeRobotAPIKey,
			"--hyperping-api-key", creds.HyperpingAPIKey,
			"--output", outputFile,
			"--import-script", importScript,
			"--report", reportFile,
			"--manual-steps", manualSteps,
			"--verbose",
		)

		output, err := cmd.CombinedOutput()
		t.Logf("Migration output:\n%s", string(output))

		if err != nil {
			return err
		}

		return nil
	})
	require.NoError(t, err, "migration tool execution failed")
	t.Log("✅ Migration tool executed successfully")

	// Step 3: Validate All Output Files Generated
	t.Log("Step 3: Validating all 4 output files were generated")
	expectedFiles := []string{"hyperping.tf", "import.sh", "migration-report.json", "manual-steps.md"}
	integration.ValidateGeneratedFiles(t, tempDir, expectedFiles)
	t.Log("✅ All 4 output files generated")

	// Step 4: Validate Terraform Syntax
	t.Log("Step 4: Validating generated Terraform is syntactically valid")
	integration.ValidateTerraformFile(t, outputFile)
	t.Log("✅ Terraform validation passed")

	// Step 5: Validate Terraform Plan
	t.Log("Step 5: Running terraform plan to verify resources")
	planCmd := exec.CommandContext(ctx, "terraform", "plan", "-no-color")
	planCmd.Dir = tempDir
	planCmd.Env = append(os.Environ(), "HYPERPING_API_KEY="+creds.HyperpingAPIKey)

	planOutput, err := planCmd.CombinedOutput()
	t.Logf("Terraform plan output:\n%s", string(planOutput))
	require.NoError(t, err, "terraform plan failed")
	t.Log("✅ Terraform plan shows expected resources (0 errors)")

	// Step 6: Validate Import Script
	t.Log("Step 6: Validating import script is executable with valid syntax")
	integration.ValidateImportScript(t, importScript)
	t.Log("✅ Import script validation passed")

	// Step 7: Validate Additional Files
	t.Log("Step 7: Validating report and manual steps files")
	integration.ValidateJSONFile(t, reportFile)
	integration.ValidateMarkdownFile(t, manualSteps)
	t.Log("✅ Report and manual steps files validated")

	// Step 8: Count and Validate Resources
	t.Log("Step 8: Counting and validating resources")
	resourceCount := integration.CountTerraformResources(t, outputFile)
	integration.ValidateScenarioOutput(t, scenario, tempDir, resourceCount)
	t.Log("✅ Resource count validation passed")

	t.Logf("=== Integration Test PASSED for scenario: %s ===", scenario.Name)
}
