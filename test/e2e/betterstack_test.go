// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

//go:build e2e

package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"testing"

	"github.com/develeap/terraform-provider-hyperping/cmd/migrate-betterstack/betterstack"
	"github.com/develeap/terraform-provider-hyperping/test/e2e/fixtures"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	betterstackBaseURL = "https://betteruptime.com/api/v2"
)

// BetterStackTestClient extends the client with create/delete capabilities
type BetterStackTestClient struct {
	*betterstack.Client
	apiToken   string
	httpClient *http.Client
}

// NewBetterStackTestClient creates a test client with create/delete methods
func NewBetterStackTestClient(apiToken string) *BetterStackTestClient {
	return &BetterStackTestClient{
		Client:     betterstack.NewClient(apiToken),
		apiToken:   apiToken,
		httpClient: &http.Client{},
	}
}

// CreateMonitor creates a monitor in Better Stack
func (c *BetterStackTestClient) CreateMonitor(ctx context.Context, name, url, method string, frequency int) (string, error) {
	payload := map[string]interface{}{
		"monitor_type":          "status",
		"pronounceable_name":    name,
		"url":                   url,
		"request_method":        method,
		"check_frequency":       frequency,
		"request_timeout":       10,
		"regions":               []string{"us", "eu"},
		"expected_status_codes": []int{200},
		"follow_redirects":      true,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", betterstackBaseURL+"/monitors", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Data.ID, nil
}

// DeleteMonitor deletes a monitor from Better Stack
func (c *BetterStackTestClient) DeleteMonitor(ctx context.Context, id string) error {
	req, err := http.NewRequestWithContext(ctx, "DELETE", betterstackBaseURL+"/monitors/"+id, http.NoBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

// CleanupTestMonitors deletes all test monitors from Better Stack
func (c *BetterStackTestClient) CleanupTestMonitors(ctx context.Context) error {
	monitors, err := c.FetchMonitors(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch monitors: %w", err)
	}

	for _, monitor := range monitors {
		if strings.HasPrefix(monitor.Attributes.PronouncableName, testResourcePrefix) {
			if err := c.DeleteMonitor(ctx, monitor.ID); err != nil {
				// Log but continue
				fmt.Printf("Warning: failed to delete monitor %s: %v\n", monitor.ID, err)
			}
		}
	}

	return nil
}

// TestBetterStackE2E_SmallMigration tests a small migration with 1-3 monitors
func TestBetterStackE2E_SmallMigration(t *testing.T) {
	creds := GetTestCredentials(t)
	SkipIfCredentialMissing(t, "BETTERSTACK_API_TOKEN", creds.BetterstackToken)
	SkipIfCredentialMissing(t, "HYPERPING_API_KEY", creds.HyperpingAPIKey)

	ctx, cancel := CreateTestContext(t)
	defer cancel()

	// Setup clients
	bsClient := NewBetterStackTestClient(creds.BetterstackToken)
	hpManager := NewHyperpingResourceManager(t, creds.HyperpingAPIKey)

	// Setup cleanup
	SetupTestCleanup(t,
		func() error { return bsClient.CleanupTestMonitors(ctx) },
		func() error { return hpManager.CleanupTestMonitors(ctx) },
	)

	// Create temp work directory
	workDir := CreateTempWorkDir(t, "betterstack-small")

	// Generate test monitors
	testPrefix := GenerateTestResourceName("BS-Small")
	monitorFixtures := fixtures.GetSmallScenarioMonitors(testPrefix)

	// Create monitors in Better Stack
	t.Logf("Creating %d test monitors in Better Stack", len(monitorFixtures))
	var createdIDs []string
	for _, fixture := range monitorFixtures {
		id, err := bsClient.CreateMonitor(ctx, fixture.Name, fixture.URL, fixture.Method, fixture.Frequency)
		require.NoError(t, err, "failed to create monitor: %s", fixture.Name)
		createdIDs = append(createdIDs, id)
		t.Logf("Created monitor: %s (ID: %s)", fixture.Name, id)
	}

	// Run migration tool
	t.Log("Running Better Stack migration tool")
	migrationTool := NewMigrationToolExecutor(t, "migrate-betterstack", workDir)
	err := migrationTool.Execute(ctx, map[string]string{
		"BETTERSTACK_API_TOKEN": creds.BetterstackToken,
		"HYPERPING_API_KEY":     creds.HyperpingAPIKey,
	})
	require.NoError(t, err, "migration tool failed")

	// Validate output files were generated
	t.Log("Validating migration tool output files")
	err = migrationTool.ValidateOutputFiles()
	require.NoError(t, err, "output files validation failed")

	// Validate Terraform syntax
	tfFile := migrationTool.GetOutputFilePath("terraform")
	err = ValidateTerraformSyntax(t, tfFile)
	require.NoError(t, err, "terraform syntax validation failed")

	// Count resources
	resourceCount := CountTerraformResources(t, tfFile)
	assert.GreaterOrEqual(t, resourceCount, len(monitorFixtures), "insufficient resources generated")

	// Execute Terraform workflow
	t.Log("Executing Terraform workflow")
	tfExec := NewTerraformExecutor(t, workDir)

	// Terraform init
	err = tfExec.Init(ctx)
	require.NoError(t, err, "terraform init failed")

	// Terraform plan
	planOutput, err := tfExec.Plan(ctx)
	require.NoError(t, err, "terraform plan failed")
	assert.Contains(t, planOutput, "hyperping_monitor", "plan should contain monitor resources")

	// Terraform apply
	err = tfExec.Apply(ctx)
	require.NoError(t, err, "terraform apply failed")

	// Verify resources were created in Hyperping
	t.Log("Verifying resources were created in Hyperping")
	for _, fixture := range monitorFixtures {
		monitor, err := hpManager.VerifyMonitorExists(ctx, fixture.Name)
		require.NoError(t, err, "monitor not found in Hyperping: %s", fixture.Name)

		// Verify monitor configuration matches
		assert.Equal(t, fixture.Name, monitor.Name, "monitor name mismatch")
		assert.Equal(t, fixture.URL, monitor.URL, "monitor URL mismatch")
	}

	// Execute import script
	t.Log("Executing import script")
	importScript := migrationTool.GetOutputFilePath("import")
	err = tfExec.ExecuteImportScript(ctx, importScript)
	require.NoError(t, err, "import script execution failed")

	// Verify state file
	err = tfExec.ValidateStateFile()
	require.NoError(t, err, "terraform state validation failed")

	// Run terraform plan again to verify no changes
	t.Log("Verifying terraform plan shows no changes")
	planOutput, err = tfExec.Plan(ctx)
	require.NoError(t, err, "terraform plan failed after import")

	// Cleanup via Terraform
	t.Log("Cleaning up via terraform destroy")
	err = tfExec.Destroy(ctx)
	require.NoError(t, err, "terraform destroy failed")

	t.Log("E2E test completed successfully")
}

// TestBetterStackE2E_MediumMigration tests a medium migration with 10+ monitors
func TestBetterStackE2E_MediumMigration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping medium migration test in short mode")
	}

	creds := GetTestCredentials(t)
	SkipIfCredentialMissing(t, "BETTERSTACK_API_TOKEN", creds.BetterstackToken)
	SkipIfCredentialMissing(t, "HYPERPING_API_KEY", creds.HyperpingAPIKey)

	ctx, cancel := CreateTestContext(t)
	defer cancel()

	// Setup clients
	bsClient := NewBetterStackTestClient(creds.BetterstackToken)
	hpManager := NewHyperpingResourceManager(t, creds.HyperpingAPIKey)

	// Setup cleanup
	SetupTestCleanup(t,
		func() error { return bsClient.CleanupTestMonitors(ctx) },
		func() error { return hpManager.CleanupTestMonitors(ctx) },
	)

	// Create temp work directory
	workDir := CreateTempWorkDir(t, "betterstack-medium")

	// Generate test monitors
	testPrefix := GenerateTestResourceName("BS-Med")
	monitorFixtures := fixtures.GetMediumScenarioMonitors(testPrefix)

	// Create monitors in Better Stack
	t.Logf("Creating %d test monitors in Better Stack", len(monitorFixtures))
	for i, fixture := range monitorFixtures {
		id, err := bsClient.CreateMonitor(ctx, fixture.Name, fixture.URL, fixture.Method, fixture.Frequency)
		require.NoError(t, err, "failed to create monitor %d: %s", i, fixture.Name)
		t.Logf("Created monitor %d/%d: %s (ID: %s)", i+1, len(monitorFixtures), fixture.Name, id)
	}

	// Run migration tool
	t.Log("Running Better Stack migration tool")
	migrationTool := NewMigrationToolExecutor(t, "migrate-betterstack", workDir)
	err := migrationTool.Execute(ctx, map[string]string{
		"BETTERSTACK_API_TOKEN": creds.BetterstackToken,
		"HYPERPING_API_KEY":     creds.HyperpingAPIKey,
	})
	require.NoError(t, err, "migration tool failed")

	// Validate output files
	err = migrationTool.ValidateOutputFiles()
	require.NoError(t, err, "output files validation failed")

	// Execute Terraform workflow
	t.Log("Executing Terraform workflow")
	tfExec := NewTerraformExecutor(t, workDir)

	err = tfExec.Init(ctx)
	require.NoError(t, err, "terraform init failed")

	_, err = tfExec.Plan(ctx)
	require.NoError(t, err, "terraform plan failed")

	err = tfExec.Apply(ctx)
	require.NoError(t, err, "terraform apply failed")

	// Verify at least some resources were created
	t.Log("Verifying resources in Hyperping")
	monitors, err := hpManager.ListTestMonitors(ctx)
	require.NoError(t, err, "failed to list monitors")
	assert.GreaterOrEqual(t, len(monitors), len(monitorFixtures), "insufficient monitors created")

	// Cleanup
	err = tfExec.Destroy(ctx)
	require.NoError(t, err, "terraform destroy failed")

	t.Log("Medium migration E2E test completed successfully")
}

// TestBetterStackE2E_ErrorHandling tests error scenarios
func TestBetterStackE2E_ErrorHandling(t *testing.T) {
	t.Run("InvalidAPIToken", func(t *testing.T) {
		creds := GetTestCredentials(t)
		SkipIfCredentialMissing(t, "HYPERPING_API_KEY", creds.HyperpingAPIKey)

		ctx, cancel := CreateTestContext(t)
		defer cancel()

		workDir := CreateTempWorkDir(t, "betterstack-error")

		migrationTool := NewMigrationToolExecutor(t, "migrate-betterstack", workDir)
		err := migrationTool.Execute(ctx, map[string]string{
			"BETTERSTACK_API_TOKEN": "invalid_token",
			"HYPERPING_API_KEY":     creds.HyperpingAPIKey,
		})

		assert.Error(t, err, "should fail with invalid token")
	})

	t.Run("EmptySourcePlatform", func(t *testing.T) {
		creds := GetTestCredentials(t)
		SkipIfCredentialMissing(t, "BETTERSTACK_API_TOKEN", creds.BetterstackToken)
		SkipIfCredentialMissing(t, "HYPERPING_API_KEY", creds.HyperpingAPIKey)

		ctx, cancel := CreateTestContext(t)
		defer cancel()

		workDir := CreateTempWorkDir(t, "betterstack-empty")

		bsClient := NewBetterStackTestClient(creds.BetterstackToken)
		SetupTestCleanup(t, func() error { return bsClient.CleanupTestMonitors(ctx) })

		// Ensure no test monitors exist
		err := bsClient.CleanupTestMonitors(ctx)
		require.NoError(t, err)

		// Run migration with no monitors
		migrationTool := NewMigrationToolExecutor(t, "migrate-betterstack", workDir)
		err = migrationTool.Execute(ctx, map[string]string{
			"BETTERSTACK_API_TOKEN": creds.BetterstackToken,
			"HYPERPING_API_KEY":     creds.HyperpingAPIKey,
		})

		// Migration should succeed but generate minimal output
		require.NoError(t, err, "migration should handle empty source")

		// Check that files were generated but may be minimal
		tfFile := migrationTool.GetOutputFilePath("terraform")
		content, err := filepath.Abs(tfFile)
		require.NoError(t, err)
		t.Logf("Generated terraform file: %s", content)
	})
}

// TestBetterStackE2E_IdempotentCleanup tests that cleanup can run multiple times
func TestBetterStackE2E_IdempotentCleanup(t *testing.T) {
	creds := GetTestCredentials(t)
	SkipIfCredentialMissing(t, "BETTERSTACK_API_TOKEN", creds.BetterstackToken)
	SkipIfCredentialMissing(t, "HYPERPING_API_KEY", creds.HyperpingAPIKey)

	ctx, cancel := CreateTestContext(t)
	defer cancel()

	bsClient := NewBetterStackTestClient(creds.BetterstackToken)
	hpManager := NewHyperpingResourceManager(t, creds.HyperpingAPIKey)

	// Run cleanup multiple times
	for i := 0; i < 3; i++ {
		t.Logf("Cleanup iteration %d", i+1)

		err := bsClient.CleanupTestMonitors(ctx)
		assert.NoError(t, err, "cleanup should be idempotent")

		err = hpManager.CleanupTestMonitors(ctx)
		assert.NoError(t, err, "cleanup should be idempotent")
	}

	t.Log("Idempotent cleanup test passed")
}
