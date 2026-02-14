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

// TestBetterStackMigration_SmallScenario tests migration of 1-3 monitors and 1 heartbeat
func TestBetterStackMigration_SmallScenario(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	creds := integration.GetTestCredentials(t)
	integration.SkipIfCredentialsMissing(t, "BETTERSTACK_API_TOKEN", creds.BetterstackToken)
	integration.SkipIfCredentialsMissing(t, "HYPERPING_API_KEY", creds.HyperpingAPIKey)

	scenario := integration.TestScenario{
		Name:                 "Small",
		Description:          "1-3 monitors and 1 heartbeat",
		ExpectedMonitors:     1,
		ExpectedHealthchecks: 1,
		MinWarnings:          0,
		MaxWarnings:          5,
	}

	runBetterStackMigrationTest(t, creds, scenario)
}

// TestBetterStackMigration_MediumScenario tests migration of 5-10 monitors and 3 heartbeats
func TestBetterStackMigration_MediumScenario(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	creds := integration.GetTestCredentials(t)
	integration.SkipIfCredentialsMissing(t, "BETTERSTACK_API_TOKEN", creds.BetterstackToken)
	integration.SkipIfCredentialsMissing(t, "HYPERPING_API_KEY", creds.HyperpingAPIKey)

	scenario := integration.TestScenario{
		Name:                 "Medium",
		Description:          "5-10 monitors and 3 heartbeats",
		ExpectedMonitors:     5,
		ExpectedHealthchecks: 3,
		MinWarnings:          0,
		MaxWarnings:          10,
	}

	runBetterStackMigrationTest(t, creds, scenario)
}

// TestBetterStackMigration_LargeScenario tests migration of 20-30 monitors and 10 heartbeats
func TestBetterStackMigration_LargeScenario(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	creds := integration.GetTestCredentials(t)
	integration.SkipIfCredentialsMissing(t, "BETTERSTACK_API_TOKEN", creds.BetterstackToken)
	integration.SkipIfCredentialsMissing(t, "HYPERPING_API_KEY", creds.HyperpingAPIKey)

	scenario := integration.TestScenario{
		Name:                 "Large",
		Description:          "20-30 monitors and 10 heartbeats",
		ExpectedMonitors:     20,
		ExpectedHealthchecks: 10,
		MinWarnings:          0,
		MaxWarnings:          20,
	}

	runBetterStackMigrationTest(t, creds, scenario)
}

// TestBetterStackMigration_DryRun tests dry-run mode without creating files
func TestBetterStackMigration_DryRun(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	creds := integration.GetTestCredentials(t)
	integration.SkipIfCredentialsMissing(t, "BETTERSTACK_API_TOKEN", creds.BetterstackToken)
	integration.SkipIfCredentialsMissing(t, "HYPERPING_API_KEY", creds.HyperpingAPIKey)

	ctx, cancel := integration.CreateTestContext(t)
	defer cancel()

	tempDir := integration.CreateTempTestDir(t, "betterstack-dryrun")

	// Run migration in dry-run mode
	cmd := exec.CommandContext(ctx,
		"go", "run", "./cmd/migrate-betterstack",
		"--betterstack-token", creds.BetterstackToken,
		"--hyperping-api-key", creds.HyperpingAPIKey,
		"--output", filepath.Join(tempDir, "migrated-resources.tf"),
		"--import-script", filepath.Join(tempDir, "import.sh"),
		"--report", filepath.Join(tempDir, "migration-report.json"),
		"--manual-steps", filepath.Join(tempDir, "manual-steps.md"),
		"--dry-run",
		"--verbose",
	)

	output, err := cmd.CombinedOutput()
	t.Logf("Migration output:\n%s", string(output))
	require.NoError(t, err, "migration command failed")

	// Verify no files were created in dry-run mode
	files := []string{
		"migrated-resources.tf",
		"import.sh",
		"migration-report.json",
		"manual-steps.md",
	}

	for _, filename := range files {
		filePath := filepath.Join(tempDir, filename)
		_, err := os.Stat(filePath)
		assert.True(t, os.IsNotExist(err), "file %s should not exist in dry-run mode", filename)
	}

	t.Log("Dry-run test passed: no files created")
}

// TestBetterStackMigration_InvalidCredentials tests error handling with invalid credentials
func TestBetterStackMigration_InvalidCredentials(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cancel := integration.CreateTestContext(t)
	defer cancel()

	tempDir := integration.CreateTempTestDir(t, "betterstack-invalid")

	// Run with invalid credentials
	cmd := exec.CommandContext(ctx,
		"go", "run", "./cmd/migrate-betterstack",
		"--betterstack-token", "invalid_token_12345",
		"--hyperping-api-key", "invalid_key_67890",
		"--output", filepath.Join(tempDir, "migrated-resources.tf"),
		"--dry-run",
	)

	output, err := cmd.CombinedOutput()
	t.Logf("Migration output:\n%s", string(output))

	// Should fail with error
	require.Error(t, err, "migration should fail with invalid credentials")

	t.Log("Invalid credentials test passed")
}

// runBetterStackMigrationTest is a helper function that runs the complete migration test flow
func runBetterStackMigrationTest(t *testing.T, creds integration.TestCredentials, scenario integration.TestScenario) {
	t.Helper()

	ctx, cancel := integration.CreateTestContext(t)
	defer cancel()

	// Create temporary output directory
	tempDir := integration.CreateTempTestDir(t, "betterstack-"+scenario.Name)
	t.Logf("Test output directory: %s", tempDir)

	// Define output file paths
	outputFile := filepath.Join(tempDir, "migrated-resources.tf")
	importScript := filepath.Join(tempDir, "import.sh")
	reportFile := filepath.Join(tempDir, "migration-report.json")
	manualSteps := filepath.Join(tempDir, "manual-steps.md")

	// Step 1: API Connection Test - Fetch Better Stack resources
	t.Log("Step 1: Testing API connection to Better Stack")
	err := integration.RunWithRetry(ctx, t, "Better Stack API connection", func() error {
		cmd := exec.CommandContext(ctx,
			"go", "run", "./cmd/migrate-betterstack",
			"--betterstack-token", creds.BetterstackToken,
			"--hyperping-api-key", creds.HyperpingAPIKey,
			"--dry-run",
			"--verbose",
		)

		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Logf("API connection test output:\n%s", string(output))
			return err
		}

		return nil
	})
	require.NoError(t, err, "Better Stack API connection failed")
	t.Log("✅ API connection successful")

	// Step 2: Execute Migration Tool
	t.Logf("Step 2: Executing migration tool for scenario: %s", scenario.Name)
	err = integration.RunWithRetry(ctx, t, "migration execution", func() error {
		cmd := exec.CommandContext(ctx,
			"go", "run", "./cmd/migrate-betterstack",
			"--betterstack-token", creds.BetterstackToken,
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
	expectedFiles := []string{
		"migrated-resources.tf",
		"import.sh",
		"migration-report.json",
		"manual-steps.md",
	}
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

	// Set environment for provider
	planCmd.Env = append(os.Environ(),
		"HYPERPING_API_KEY="+creds.HyperpingAPIKey,
	)

	planOutput, err := planCmd.CombinedOutput()
	t.Logf("Terraform plan output:\n%s", string(planOutput))

	// Plan should succeed (0 errors)
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
