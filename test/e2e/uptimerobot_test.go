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

	"github.com/develeap/terraform-provider-hyperping/cmd/migrate-uptimerobot/uptimerobot"
	"github.com/develeap/terraform-provider-hyperping/test/e2e/fixtures"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	uptimerobotBaseURL = "https://api.uptimerobot.com/v2"
)

// UptimeRobotTestClient extends the client with create/delete capabilities
type UptimeRobotTestClient struct {
	*uptimerobot.Client
	apiKey     string
	httpClient *http.Client
}

// NewUptimeRobotTestClient creates a test client with create/delete methods
func NewUptimeRobotTestClient(apiKey string) *UptimeRobotTestClient {
	return &UptimeRobotTestClient{
		Client:     uptimerobot.NewClient(apiKey),
		apiKey:     apiKey,
		httpClient: &http.Client{},
	}
}

// CreateMonitor creates a monitor in UptimeRobot
func (c *UptimeRobotTestClient) CreateMonitor(ctx context.Context, name, url string, monitorType, interval int) (int, error) {
	payload := map[string]interface{}{
		"api_key":       c.apiKey,
		"format":        "json",
		"type":          monitorType, // 1=HTTP(s), 2=Keyword, 3=Ping, 4=Port
		"friendly_name": name,
		"url":           url,
		"interval":      interval, // in seconds
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", uptimerobotBaseURL+"/newMonitor", bytes.NewBuffer(jsonData))
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Stat    string `json:"stat"`
		Monitor struct {
			ID int `json:"id"`
		} `json:"monitor"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("failed to decode response: %w", err)
	}

	if result.Stat != "ok" {
		return 0, fmt.Errorf("API returned error status: %s", result.Stat)
	}

	return result.Monitor.ID, nil
}

// DeleteMonitor deletes a monitor from UptimeRobot
func (c *UptimeRobotTestClient) DeleteMonitor(ctx context.Context, id int) error {
	payload := map[string]interface{}{
		"api_key": c.apiKey,
		"format":  "json",
		"id":      id,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", uptimerobotBaseURL+"/deleteMonitor", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Stat string `json:"stat"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if result.Stat != "ok" {
		return fmt.Errorf("API returned error status: %s", result.Stat)
	}

	return nil
}

// CleanupTestMonitors deletes all test monitors from UptimeRobot
func (c *UptimeRobotTestClient) CleanupTestMonitors(ctx context.Context) error {
	monitors, err := c.GetMonitors(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch monitors: %w", err)
	}

	for _, monitor := range monitors {
		if strings.HasPrefix(monitor.FriendlyName, testResourcePrefix) {
			if err := c.DeleteMonitor(ctx, monitor.ID); err != nil {
				fmt.Printf("Warning: failed to delete monitor %d: %v\n", monitor.ID, err)
			}
		}
	}

	return nil
}

// TestUptimeRobotE2E_SmallMigration tests a small migration with 1-3 monitors
func TestUptimeRobotE2E_SmallMigration(t *testing.T) {
	creds := GetTestCredentials(t)
	SkipIfCredentialMissing(t, "UPTIMEROBOT_API_KEY", creds.UptimeRobotAPIKey)
	SkipIfCredentialMissing(t, "HYPERPING_API_KEY", creds.HyperpingAPIKey)

	ctx, cancel := CreateTestContext(t)
	defer cancel()

	// Setup clients
	urClient := NewUptimeRobotTestClient(creds.UptimeRobotAPIKey)
	hpManager := NewHyperpingResourceManager(t, creds.HyperpingAPIKey)

	// Setup cleanup
	SetupTestCleanup(t,
		func() error { return urClient.CleanupTestMonitors(ctx) },
		func() error { return hpManager.CleanupTestMonitors(ctx) },
	)

	// Create temp work directory
	workDir := CreateTempWorkDir(t, "uptimerobot-small")

	// Generate test monitors
	testPrefix := GenerateTestResourceName("UR-Small")
	monitorFixtures := fixtures.GetSmallScenarioMonitors(testPrefix)

	// Create monitors in UptimeRobot
	t.Logf("Creating %d test monitors in UptimeRobot", len(monitorFixtures))
	var createdIDs []int
	for _, fixture := range monitorFixtures {
		// UptimeRobot type: 1=HTTP(s)
		id, err := urClient.CreateMonitor(ctx, fixture.Name, fixture.URL, 1, fixture.Frequency)
		require.NoError(t, err, "failed to create monitor: %s", fixture.Name)
		createdIDs = append(createdIDs, id)
		t.Logf("Created monitor: %s (ID: %d)", fixture.Name, id)
	}

	// Run migration tool
	t.Log("Running UptimeRobot migration tool")
	migrationTool := NewMigrationToolExecutor(t, "migrate-uptimerobot", workDir)
	err := migrationTool.Execute(ctx, map[string]string{
		"UPTIMEROBOT_API_KEY": creds.UptimeRobotAPIKey,
		"HYPERPING_API_KEY":   creds.HyperpingAPIKey,
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

	// Cleanup via Terraform
	t.Log("Cleaning up via terraform destroy")
	err = tfExec.Destroy(ctx)
	require.NoError(t, err, "terraform destroy failed")

	t.Log("E2E test completed successfully")
}

// TestUptimeRobotE2E_MediumMigration tests a medium migration with 10+ monitors
func TestUptimeRobotE2E_MediumMigration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping medium migration test in short mode")
	}

	creds := GetTestCredentials(t)
	SkipIfCredentialMissing(t, "UPTIMEROBOT_API_KEY", creds.UptimeRobotAPIKey)
	SkipIfCredentialMissing(t, "HYPERPING_API_KEY", creds.HyperpingAPIKey)

	ctx, cancel := CreateTestContext(t)
	defer cancel()

	// Setup clients
	urClient := NewUptimeRobotTestClient(creds.UptimeRobotAPIKey)
	hpManager := NewHyperpingResourceManager(t, creds.HyperpingAPIKey)

	// Setup cleanup
	SetupTestCleanup(t,
		func() error { return urClient.CleanupTestMonitors(ctx) },
		func() error { return hpManager.CleanupTestMonitors(ctx) },
	)

	// Create temp work directory
	workDir := CreateTempWorkDir(t, "uptimerobot-medium")

	// Generate test monitors
	testPrefix := GenerateTestResourceName("UR-Med")
	monitorFixtures := fixtures.GetMediumScenarioMonitors(testPrefix)

	// Create monitors in UptimeRobot
	t.Logf("Creating %d test monitors in UptimeRobot", len(monitorFixtures))
	for i, fixture := range monitorFixtures {
		id, err := urClient.CreateMonitor(ctx, fixture.Name, fixture.URL, 1, fixture.Frequency)
		require.NoError(t, err, "failed to create monitor %d: %s", i, fixture.Name)
		t.Logf("Created monitor %d/%d: %s (ID: %d)", i+1, len(monitorFixtures), fixture.Name, id)
	}

	// Run migration tool
	t.Log("Running UptimeRobot migration tool")
	migrationTool := NewMigrationToolExecutor(t, "migrate-uptimerobot", workDir)
	err := migrationTool.Execute(ctx, map[string]string{
		"UPTIMEROBOT_API_KEY": creds.UptimeRobotAPIKey,
		"HYPERPING_API_KEY":   creds.HyperpingAPIKey,
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

// TestUptimeRobotE2E_ErrorHandling tests error scenarios
func TestUptimeRobotE2E_ErrorHandling(t *testing.T) {
	t.Run("InvalidAPIKey", func(t *testing.T) {
		creds := GetTestCredentials(t)
		SkipIfCredentialMissing(t, "HYPERPING_API_KEY", creds.HyperpingAPIKey)

		ctx, cancel := CreateTestContext(t)
		defer cancel()

		workDir := CreateTempWorkDir(t, "uptimerobot-error")

		migrationTool := NewMigrationToolExecutor(t, "migrate-uptimerobot", workDir)
		err := migrationTool.Execute(ctx, map[string]string{
			"UPTIMEROBOT_API_KEY": "invalid_key",
			"HYPERPING_API_KEY":   creds.HyperpingAPIKey,
		})

		assert.Error(t, err, "should fail with invalid API key")
	})
}

// TestUptimeRobotE2E_IdempotentCleanup tests that cleanup can run multiple times
func TestUptimeRobotE2E_IdempotentCleanup(t *testing.T) {
	creds := GetTestCredentials(t)
	SkipIfCredentialMissing(t, "UPTIMEROBOT_API_KEY", creds.UptimeRobotAPIKey)
	SkipIfCredentialMissing(t, "HYPERPING_API_KEY", creds.HyperpingAPIKey)

	ctx, cancel := CreateTestContext(t)
	defer cancel()

	urClient := NewUptimeRobotTestClient(creds.UptimeRobotAPIKey)
	hpManager := NewHyperpingResourceManager(t, creds.HyperpingAPIKey)

	// Run cleanup multiple times
	for i := 0; i < 3; i++ {
		t.Logf("Cleanup iteration %d", i+1)

		err := urClient.CleanupTestMonitors(ctx)
		assert.NoError(t, err, "cleanup should be idempotent")

		err = hpManager.CleanupTestMonitors(ctx)
		assert.NoError(t, err, "cleanup should be idempotent")
	}

	t.Log("Idempotent cleanup test passed")
}
