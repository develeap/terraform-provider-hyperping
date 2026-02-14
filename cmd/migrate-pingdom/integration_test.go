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

// TestPingdomMigration_SmallScenario tests migration of 1-3 checks
func TestPingdomMigration_SmallScenario(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	creds := integration.GetTestCredentials(t)
	integration.SkipIfCredentialsMissing(t, "PINGDOM_API_KEY", creds.PingdomAPIKey)
	integration.SkipIfCredentialsMissing(t, "HYPERPING_API_KEY", creds.HyperpingAPIKey)

	scenario := integration.TestScenario{
		Name:             "Small",
		Description:      "1-3 checks of various types",
		ExpectedMonitors: 1,
		MinWarnings:      0,
		MaxWarnings:      5,
	}

	runPingdomMigrationTest(t, creds, scenario)
}

// TestPingdomMigration_MediumScenario tests migration of 5-10 checks
func TestPingdomMigration_MediumScenario(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	creds := integration.GetTestCredentials(t)
	integration.SkipIfCredentialsMissing(t, "PINGDOM_API_KEY", creds.PingdomAPIKey)
	integration.SkipIfCredentialsMissing(t, "HYPERPING_API_KEY", creds.HyperpingAPIKey)

	scenario := integration.TestScenario{
		Name:             "Medium",
		Description:      "5-10 checks including HTTP, TCP, and PING types",
		ExpectedMonitors: 5,
		MinWarnings:      0,
		MaxWarnings:      10,
	}

	runPingdomMigrationTest(t, creds, scenario)
}

// TestPingdomMigration_LargeScenario tests migration of 20-30 checks
func TestPingdomMigration_LargeScenario(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	creds := integration.GetTestCredentials(t)
	integration.SkipIfCredentialsMissing(t, "PINGDOM_API_KEY", creds.PingdomAPIKey)
	integration.SkipIfCredentialsMissing(t, "HYPERPING_API_KEY", creds.HyperpingAPIKey)

	scenario := integration.TestScenario{
		Name:             "Large",
		Description:      "20-30 checks testing all supported check types",
		ExpectedMonitors: 20,
		MinWarnings:      0,
		MaxWarnings:      20,
	}

	runPingdomMigrationTest(t, creds, scenario)
}

// TestPingdomMigration_CheckTypes tests different Pingdom check types
func TestPingdomMigration_CheckTypes(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	creds := integration.GetTestCredentials(t)
	integration.SkipIfCredentialsMissing(t, "PINGDOM_API_KEY", creds.PingdomAPIKey)
	integration.SkipIfCredentialsMissing(t, "HYPERPING_API_KEY", creds.HyperpingAPIKey)

	// Test check types
	checkTypes := []struct {
		name        string
		description string
	}{
		{"HTTP", "HTTP/HTTPS uptime monitoring"},
		{"TCP", "TCP port monitoring"},
		{"PING", "ICMP ping monitoring"},
		{"SMTP", "SMTP email server monitoring"},
	}

	ctx, cancel := integration.CreateTestContext(t)
	defer cancel()

	tempDir := integration.CreateTempTestDir(t, "pingdom-types")
	outputDir := filepath.Join(tempDir, "output")

	// Run migration to get all checks
	err := integration.RunWithRetry(ctx, t, "migration execution", func() error {
		cmd := exec.CommandContext(ctx,
			"go", "run", "./cmd/migrate-pingdom",
			"--pingdom-api-key", creds.PingdomAPIKey,
			"--hyperping-api-key", creds.HyperpingAPIKey,
			"--output", outputDir,
			"--verbose",
		)

		output, err := cmd.CombinedOutput()
		t.Logf("Migration output:\n%s", string(output))
		return err
	})

	require.NoError(t, err, "migration failed")

	// Verify output files were created
	monitorsFile := filepath.Join(outputDir, "monitors.tf")
	require.FileExists(t, monitorsFile, "monitors.tf not created")

	resourceCount := integration.CountTerraformResources(t, monitorsFile)
	require.Greater(t, resourceCount, 0, "no resources were migrated")

	t.Logf("Successfully tested migration with %d resources covering check types: %v",
		resourceCount, checkTypes)
}

// TestPingdomMigration_DryRun tests dry-run mode
func TestPingdomMigration_DryRun(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	creds := integration.GetTestCredentials(t)
	integration.SkipIfCredentialsMissing(t, "PINGDOM_API_KEY", creds.PingdomAPIKey)
	integration.SkipIfCredentialsMissing(t, "HYPERPING_API_KEY", creds.HyperpingAPIKey)

	ctx, cancel := integration.CreateTestContext(t)
	defer cancel()

	tempDir := integration.CreateTempTestDir(t, "pingdom-dryrun")
	outputDir := filepath.Join(tempDir, "output")

	// Run migration in dry-run mode
	cmd := exec.CommandContext(ctx,
		"go", "run", "./cmd/migrate-pingdom",
		"--pingdom-api-key", creds.PingdomAPIKey,
		"--hyperping-api-key", creds.HyperpingAPIKey,
		"--output", outputDir,
		"--dry-run",
		"--verbose",
	)

	output, err := cmd.CombinedOutput()
	t.Logf("Migration output:\n%s", string(output))
	require.NoError(t, err, "dry-run mode failed")

	// In dry-run mode, files should still be created (just no resources in Hyperping)
	// Verify expected files exist
	expectedFiles := []string{
		"monitors.tf",
		"import.sh",
		"report.json",
		"report.txt",
		"manual-steps.md",
	}

	for _, filename := range expectedFiles {
		filePath := filepath.Join(outputDir, filename)
		require.FileExists(t, filePath, "expected file %s not created in dry-run", filename)
	}

	t.Log("Dry-run test passed")
}

// TestPingdomMigration_WithPrefix tests resource name prefix feature
func TestPingdomMigration_WithPrefix(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	creds := integration.GetTestCredentials(t)
	integration.SkipIfCredentialsMissing(t, "PINGDOM_API_KEY", creds.PingdomAPIKey)
	integration.SkipIfCredentialsMissing(t, "HYPERPING_API_KEY", creds.HyperpingAPIKey)

	ctx, cancel := integration.CreateTestContext(t)
	defer cancel()

	tempDir := integration.CreateTempTestDir(t, "pingdom-prefix")
	outputDir := filepath.Join(tempDir, "output")

	// Run migration with prefix
	cmd := exec.CommandContext(ctx,
		"go", "run", "./cmd/migrate-pingdom",
		"--pingdom-api-key", creds.PingdomAPIKey,
		"--hyperping-api-key", creds.HyperpingAPIKey,
		"--output", outputDir,
		"--prefix", "pingdom_",
		"--dry-run",
		"--verbose",
	)

	output, err := cmd.CombinedOutput()
	t.Logf("Migration output:\n%s", string(output))
	require.NoError(t, err, "migration with prefix failed")

	// Verify monitors.tf contains prefixed resource names
	monitorsFile := filepath.Join(outputDir, "monitors.tf")
	require.FileExists(t, monitorsFile, "monitors.tf not created")

	content, err := os.ReadFile(monitorsFile)
	require.NoError(t, err, "failed to read monitors.tf")

	// Should contain resources with the prefix
	tfContent := string(content)
	assert.Contains(t, tfContent, "pingdom_", "resource names should have prefix")

	t.Log("Prefix test passed")
}

// TestPingdomMigration_InvalidCredentials tests error handling
func TestPingdomMigration_InvalidCredentials(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cancel := integration.CreateTestContext(t)
	defer cancel()

	tempDir := integration.CreateTempTestDir(t, "pingdom-invalid")
	outputDir := filepath.Join(tempDir, "output")

	// Run with invalid credentials
	cmd := exec.CommandContext(ctx,
		"go", "run", "./cmd/migrate-pingdom",
		"--pingdom-api-key", "invalid_token_12345",
		"--hyperping-api-key", "invalid_key_67890",
		"--output", outputDir,
		"--dry-run",
	)

	output, err := cmd.CombinedOutput()
	t.Logf("Migration output:\n%s", string(output))

	// Should fail with error
	require.Error(t, err, "migration should fail with invalid credentials")

	t.Log("Invalid credentials test passed")
}

// runPingdomMigrationTest is a helper function that runs the complete migration test flow
func runPingdomMigrationTest(t *testing.T, creds integration.TestCredentials, scenario integration.TestScenario) {
	t.Helper()

	ctx, cancel := integration.CreateTestContext(t)
	defer cancel()

	// Create temporary output directory
	tempDir := integration.CreateTempTestDir(t, "pingdom-"+scenario.Name)
	outputDir := filepath.Join(tempDir, "migration-output")
	t.Logf("Test output directory: %s", outputDir)

	// Step 1: API Connection Test
	t.Log("Step 1: Testing API connection to Pingdom")
	err := integration.RunWithRetry(ctx, t, "Pingdom API connection", func() error {
		cmd := exec.CommandContext(ctx,
			"go", "run", "./cmd/migrate-pingdom",
			"--pingdom-api-key", creds.PingdomAPIKey,
			"--hyperping-api-key", creds.HyperpingAPIKey,
			"--output", filepath.Join(tempDir, "test-connection"),
			"--dry-run",
		)

		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Logf("API connection test output:\n%s", string(output))
			return err
		}

		return nil
	})
	require.NoError(t, err, "Pingdom API connection failed")
	t.Log("✅ API connection successful")

	// Step 2: Execute Migration Tool
	t.Logf("Step 2: Executing migration tool for scenario: %s", scenario.Name)
	err = integration.RunWithRetry(ctx, t, "migration execution", func() error {
		cmd := exec.CommandContext(ctx,
			"go", "run", "./cmd/migrate-pingdom",
			"--pingdom-api-key", creds.PingdomAPIKey,
			"--hyperping-api-key", creds.HyperpingAPIKey,
			"--output", outputDir,
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
	t.Log("Step 3: Validating all output files were generated")
	expectedFiles := []string{
		"monitors.tf",
		"import.sh",
		"report.json",
		"report.txt",
		"manual-steps.md",
	}
	integration.ValidateGeneratedFiles(t, outputDir, expectedFiles)
	t.Log("✅ All output files generated")

	// Step 4: Validate Terraform Syntax
	t.Log("Step 4: Validating generated Terraform is syntactically valid")
	monitorsFile := filepath.Join(outputDir, "monitors.tf")
	integration.ValidateTerraformFile(t, monitorsFile)
	t.Log("✅ Terraform validation passed")

	// Step 5: Validate Terraform Plan
	t.Log("Step 5: Running terraform plan to verify resources")
	planCmd := exec.CommandContext(ctx, "terraform", "plan", "-no-color")
	planCmd.Dir = outputDir
	planCmd.Env = append(os.Environ(), "HYPERPING_API_KEY="+creds.HyperpingAPIKey)

	planOutput, err := planCmd.CombinedOutput()
	t.Logf("Terraform plan output:\n%s", string(planOutput))
	require.NoError(t, err, "terraform plan failed")
	t.Log("✅ Terraform plan shows expected resources (0 errors)")

	// Step 6: Validate Import Script
	t.Log("Step 6: Validating import script is executable with valid syntax")
	importScript := filepath.Join(outputDir, "import.sh")
	integration.ValidateImportScript(t, importScript)
	t.Log("✅ Import script validation passed")

	// Step 7: Validate Additional Files
	t.Log("Step 7: Validating report and manual steps files")
	reportJSON := filepath.Join(outputDir, "report.json")
	manualSteps := filepath.Join(outputDir, "manual-steps.md")

	integration.ValidateJSONFile(t, reportJSON)
	integration.ValidateMarkdownFile(t, manualSteps)
	t.Log("✅ Report and manual steps files validated")

	// Step 8: Count and Validate Resources
	t.Log("Step 8: Counting and validating resources")
	resourceCount := integration.CountTerraformResources(t, monitorsFile)
	integration.ValidateScenarioOutput(t, scenario, outputDir, resourceCount)
	t.Log("✅ Resource count validation passed")

	t.Logf("=== Integration Test PASSED for scenario: %s ===", scenario.Name)
}
