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
	"strings"
	"testing"

	"github.com/develeap/terraform-provider-hyperping/cmd/migrate-pingdom/pingdom"
	"github.com/develeap/terraform-provider-hyperping/test/e2e/fixtures"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	pingdomBaseURL = "https://api.pingdom.com/api/3.1"
)

// PingdomTestClient extends the client with create/delete capabilities
type PingdomTestClient struct {
	*pingdom.Client
	apiToken   string
	httpClient *http.Client
}

// NewPingdomTestClient creates a test client with create/delete methods
func NewPingdomTestClient(apiToken string) *PingdomTestClient {
	return &PingdomTestClient{
		Client:     pingdom.NewClient(apiToken),
		apiToken:   apiToken,
		httpClient: &http.Client{},
	}
}

// CreateCheck creates a check in Pingdom
func (c *PingdomTestClient) CreateCheck(ctx context.Context, name, hostname, checkType string, resolution int) (int, error) {
	// Build request based on check type
	payload := map[string]interface{}{
		"name":       name,
		"type":       checkType, // http, tcp, ping, dns
		"host":       hostname,
		"resolution": resolution, // Check frequency in minutes
	}

	// Add URL for HTTP checks
	if checkType == "http" {
		payload["url"] = "/"
		payload["encryption"] = hostname[:5] == "https"
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", pingdomBaseURL+"/checks", bytes.NewBuffer(jsonData))
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return 0, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result struct {
		Check struct {
			ID int `json:"id"`
		} `json:"check"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Check.ID, nil
}

// DeleteCheck deletes a check from Pingdom
func (c *PingdomTestClient) DeleteCheck(ctx context.Context, id int) error {
	url := fmt.Sprintf("%s/checks/%d", pingdomBaseURL, id)
	req, err := http.NewRequestWithContext(ctx, "DELETE", url, http.NoBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

// CleanupTestChecks deletes all test checks from Pingdom
func (c *PingdomTestClient) CleanupTestChecks(ctx context.Context) error {
	checks, err := c.ListChecks(ctx)
	if err != nil {
		return fmt.Errorf("failed to list checks: %w", err)
	}

	for _, check := range checks {
		if strings.HasPrefix(check.Name, testResourcePrefix) {
			if err := c.DeleteCheck(ctx, check.ID); err != nil {
				fmt.Printf("Warning: failed to delete check %d: %v\n", check.ID, err)
			}
		}
	}

	return nil
}

// TestPingdomE2E_SmallMigration tests a small migration with 1-3 checks
func TestPingdomE2E_SmallMigration(t *testing.T) {
	creds := GetTestCredentials(t)
	SkipIfCredentialMissing(t, "PINGDOM_API_KEY", creds.PingdomAPIKey)
	SkipIfCredentialMissing(t, "HYPERPING_API_KEY", creds.HyperpingAPIKey)

	ctx, cancel := CreateTestContext(t)
	defer cancel()

	// Setup clients
	pdClient := NewPingdomTestClient(creds.PingdomAPIKey)
	hpManager := NewHyperpingResourceManager(t, creds.HyperpingAPIKey)

	// Setup cleanup
	SetupTestCleanup(t,
		func() error { return pdClient.CleanupTestChecks(ctx) },
		func() error { return hpManager.CleanupTestMonitors(ctx) },
	)

	// Create temp work directory
	workDir := CreateTempWorkDir(t, "pingdom-small")

	// Generate test checks
	testPrefix := GenerateTestResourceName("PD-Small")
	monitorFixtures := fixtures.GetSmallScenarioMonitors(testPrefix)

	// Create checks in Pingdom
	t.Logf("Creating %d test checks in Pingdom", len(monitorFixtures))
	var createdIDs []int
	for _, fixture := range monitorFixtures {
		// Convert frequency from seconds to minutes for Pingdom
		resolutionMinutes := fixture.Frequency / 60
		if resolutionMinutes < 1 {
			resolutionMinutes = 1
		}

		id, err := pdClient.CreateCheck(ctx, fixture.Name, fixture.URL, "http", resolutionMinutes)
		require.NoError(t, err, "failed to create check: %s", fixture.Name)
		createdIDs = append(createdIDs, id)
		t.Logf("Created check: %s (ID: %d)", fixture.Name, id)
	}

	// Run migration tool
	t.Log("Running Pingdom migration tool")
	migrationTool := NewMigrationToolExecutor(t, "migrate-pingdom", workDir)
	err := migrationTool.Execute(ctx, map[string]string{
		"PINGDOM_API_KEY":   creds.PingdomAPIKey,
		"HYPERPING_API_KEY": creds.HyperpingAPIKey,
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
	}

	// Execute import script
	t.Log("Executing import script")
	importScript := migrationTool.GetOutputFilePath("import")
	err = tfExec.ExecuteImportScript(ctx, importScript)
	require.NoError(t, err, "import script execution failed")

	// Verify state file
	err = tfExec.ValidateStateFile()
	require.NoError(t, err, "terraform state validation failed")

	// Cleanup via Terraform
	t.Log("Cleaning up via terraform destroy")
	err = tfExec.Destroy(ctx)
	require.NoError(t, err, "terraform destroy failed")

	t.Log("E2E test completed successfully")
}

// TestPingdomE2E_MediumMigration tests a medium migration with 10+ checks
func TestPingdomE2E_MediumMigration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping medium migration test in short mode")
	}

	creds := GetTestCredentials(t)
	SkipIfCredentialMissing(t, "PINGDOM_API_KEY", creds.PingdomAPIKey)
	SkipIfCredentialMissing(t, "HYPERPING_API_KEY", creds.HyperpingAPIKey)

	ctx, cancel := CreateTestContext(t)
	defer cancel()

	// Setup clients
	pdClient := NewPingdomTestClient(creds.PingdomAPIKey)
	hpManager := NewHyperpingResourceManager(t, creds.HyperpingAPIKey)

	// Setup cleanup
	SetupTestCleanup(t,
		func() error { return pdClient.CleanupTestChecks(ctx) },
		func() error { return hpManager.CleanupTestMonitors(ctx) },
	)

	// Create temp work directory
	workDir := CreateTempWorkDir(t, "pingdom-medium")

	// Generate test checks
	testPrefix := GenerateTestResourceName("PD-Med")
	monitorFixtures := fixtures.GetMediumScenarioMonitors(testPrefix)

	// Create checks in Pingdom
	t.Logf("Creating %d test checks in Pingdom", len(monitorFixtures))
	for i, fixture := range monitorFixtures {
		resolutionMinutes := fixture.Frequency / 60
		if resolutionMinutes < 1 {
			resolutionMinutes = 1
		}

		id, err := pdClient.CreateCheck(ctx, fixture.Name, fixture.URL, "http", resolutionMinutes)
		require.NoError(t, err, "failed to create check %d: %s", i, fixture.Name)
		t.Logf("Created check %d/%d: %s (ID: %d)", i+1, len(monitorFixtures), fixture.Name, id)
	}

	// Run migration tool
	t.Log("Running Pingdom migration tool")
	migrationTool := NewMigrationToolExecutor(t, "migrate-pingdom", workDir)
	err := migrationTool.Execute(ctx, map[string]string{
		"PINGDOM_API_KEY":   creds.PingdomAPIKey,
		"HYPERPING_API_KEY": creds.HyperpingAPIKey,
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

// TestPingdomE2E_ErrorHandling tests error scenarios
func TestPingdomE2E_ErrorHandling(t *testing.T) {
	t.Run("InvalidAPIKey", func(t *testing.T) {
		creds := GetTestCredentials(t)
		SkipIfCredentialMissing(t, "HYPERPING_API_KEY", creds.HyperpingAPIKey)

		ctx, cancel := CreateTestContext(t)
		defer cancel()

		workDir := CreateTempWorkDir(t, "pingdom-error")

		migrationTool := NewMigrationToolExecutor(t, "migrate-pingdom", workDir)
		err := migrationTool.Execute(ctx, map[string]string{
			"PINGDOM_API_KEY":   "invalid_key",
			"HYPERPING_API_KEY": creds.HyperpingAPIKey,
		})

		assert.Error(t, err, "should fail with invalid API key")
	})
}

// TestPingdomE2E_IdempotentCleanup tests that cleanup can run multiple times
func TestPingdomE2E_IdempotentCleanup(t *testing.T) {
	creds := GetTestCredentials(t)
	SkipIfCredentialMissing(t, "PINGDOM_API_KEY", creds.PingdomAPIKey)
	SkipIfCredentialMissing(t, "HYPERPING_API_KEY", creds.HyperpingAPIKey)

	ctx, cancel := CreateTestContext(t)
	defer cancel()

	pdClient := NewPingdomTestClient(creds.PingdomAPIKey)
	hpManager := NewHyperpingResourceManager(t, creds.HyperpingAPIKey)

	// Run cleanup multiple times
	for i := 0; i < 3; i++ {
		t.Logf("Cleanup iteration %d", i+1)

		err := pdClient.CleanupTestChecks(ctx)
		assert.NoError(t, err, "cleanup should be idempotent")

		err = hpManager.CleanupTestMonitors(ctx)
		assert.NoError(t, err, "cleanup should be idempotent")
	}

	t.Log("Idempotent cleanup test passed")
}
