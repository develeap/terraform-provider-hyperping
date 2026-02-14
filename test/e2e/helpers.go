// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

//go:build e2e

package e2e

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/develeap/terraform-provider-hyperping/internal/client"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

const (
	testTimeout        = 30 * time.Minute
	cleanupTimeout     = 10 * time.Minute
	terraformTimeout   = 15 * time.Minute
	maxRetries         = 3
	retryDelay         = 2 * time.Second
	testResourcePrefix = "E2E-Test-"
)

// TestCredentials holds all API credentials needed for E2E tests
type TestCredentials struct {
	BetterstackToken  string
	UptimeRobotAPIKey string
	PingdomAPIKey     string
	HyperpingAPIKey   string
}

// GetTestCredentials retrieves credentials from environment variables
func GetTestCredentials(t *testing.T) TestCredentials {
	t.Helper()

	return TestCredentials{
		BetterstackToken:  os.Getenv("BETTERSTACK_API_TOKEN"),
		UptimeRobotAPIKey: os.Getenv("UPTIMEROBOT_API_KEY"),
		PingdomAPIKey:     getEnvOrAlternate("PINGDOM_API_KEY", "PINGDOM_API_TOKEN"),
		HyperpingAPIKey:   os.Getenv("HYPERPING_API_KEY"),
	}
}

// getEnvOrAlternate returns the value of the primary env var, or the alternate if empty
func getEnvOrAlternate(primary, alternate string) string {
	if val := os.Getenv(primary); val != "" {
		return val
	}
	return os.Getenv(alternate)
}

// SkipIfCredentialMissing skips the test if a required credential is missing
func SkipIfCredentialMissing(t *testing.T, credName, credValue string) {
	t.Helper()

	if credValue == "" {
		t.Skipf("Skipping E2E test: %s not set (required)", credName)
	}
}

// CreateTestContext creates a context with timeout for E2E tests
func CreateTestContext(t *testing.T) (context.Context, context.CancelFunc) {
	t.Helper()
	return context.WithTimeout(context.Background(), testTimeout)
}

// CreateTempWorkDir creates a temporary working directory for test
func CreateTempWorkDir(t *testing.T, prefix string) string {
	t.Helper()

	tempDir, err := os.MkdirTemp("", fmt.Sprintf("%s-e2e-*", prefix))
	require.NoError(t, err, "failed to create temp directory")

	t.Cleanup(func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("Warning: failed to cleanup temp directory %s: %v", tempDir, err)
		}
	})

	t.Logf("Created temp work directory: %s", tempDir)
	return tempDir
}

// GenerateTestResourceName generates a unique test resource name with timestamp
func GenerateTestResourceName(prefix string) string {
	timestamp := time.Now().Format("20060102-150405")
	randomID := uuid.New().String()[:8]
	return fmt.Sprintf("%s%s-%s-%s", testResourcePrefix, prefix, timestamp, randomID)
}

// TerraformExecutor handles Terraform command execution
type TerraformExecutor struct {
	workDir string
	t       *testing.T
}

// NewTerraformExecutor creates a new Terraform executor
func NewTerraformExecutor(t *testing.T, workDir string) *TerraformExecutor {
	t.Helper()
	return &TerraformExecutor{
		workDir: workDir,
		t:       t,
	}
}

// Init runs terraform init
func (te *TerraformExecutor) Init(ctx context.Context) error {
	te.t.Helper()
	te.t.Logf("Running terraform init in %s", te.workDir)

	cmd := exec.CommandContext(ctx, "terraform", "init", "-no-color")
	cmd.Dir = te.workDir
	cmd.Env = append(os.Environ(),
		"TF_LOG=",
		"TF_IN_AUTOMATION=1",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		te.t.Logf("Terraform init output:\n%s", string(output))
		return fmt.Errorf("terraform init failed: %w", err)
	}

	te.t.Logf("Terraform init succeeded")
	return nil
}

// Plan runs terraform plan and returns the output
func (te *TerraformExecutor) Plan(ctx context.Context) (string, error) {
	te.t.Helper()
	te.t.Logf("Running terraform plan in %s", te.workDir)

	cmd := exec.CommandContext(ctx, "terraform", "plan", "-no-color", "-detailed-exitcode")
	cmd.Dir = te.workDir
	cmd.Env = append(os.Environ(),
		"TF_LOG=",
		"TF_IN_AUTOMATION=1",
	)

	output, err := cmd.CombinedOutput()
	outputStr := string(output)
	te.t.Logf("Terraform plan output:\n%s", outputStr)

	// Exit code 2 means changes are present (not an error)
	if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 2 {
		te.t.Logf("Terraform plan shows changes to apply")
		return outputStr, nil
	}

	if err != nil {
		return outputStr, fmt.Errorf("terraform plan failed: %w", err)
	}

	te.t.Logf("Terraform plan succeeded (no changes)")
	return outputStr, nil
}

// Apply runs terraform apply with auto-approve
func (te *TerraformExecutor) Apply(ctx context.Context) error {
	te.t.Helper()
	te.t.Logf("Running terraform apply in %s", te.workDir)

	cmd := exec.CommandContext(ctx, "terraform", "apply", "-auto-approve", "-no-color")
	cmd.Dir = te.workDir
	cmd.Env = append(os.Environ(),
		"TF_LOG=",
		"TF_IN_AUTOMATION=1",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		te.t.Logf("Terraform apply output:\n%s", string(output))
		return fmt.Errorf("terraform apply failed: %w", err)
	}

	te.t.Logf("Terraform apply succeeded")
	return nil
}

// Destroy runs terraform destroy with auto-approve
func (te *TerraformExecutor) Destroy(ctx context.Context) error {
	te.t.Helper()
	te.t.Logf("Running terraform destroy in %s", te.workDir)

	cmd := exec.CommandContext(ctx, "terraform", "destroy", "-auto-approve", "-no-color")
	cmd.Dir = te.workDir
	cmd.Env = append(os.Environ(),
		"TF_LOG=",
		"TF_IN_AUTOMATION=1",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		te.t.Logf("Terraform destroy output:\n%s", string(output))
		return fmt.Errorf("terraform destroy failed: %w", err)
	}

	te.t.Logf("Terraform destroy succeeded")
	return nil
}

// ExecuteImportScript executes the generated import script
func (te *TerraformExecutor) ExecuteImportScript(ctx context.Context, scriptPath string) error {
	te.t.Helper()
	te.t.Logf("Executing import script: %s", scriptPath)

	// Make script executable
	if err := os.Chmod(scriptPath, 0755); err != nil {
		return fmt.Errorf("failed to make script executable: %w", err)
	}

	cmd := exec.CommandContext(ctx, "bash", scriptPath)
	cmd.Dir = te.workDir
	cmd.Env = append(os.Environ(),
		"TF_LOG=",
		"TF_IN_AUTOMATION=1",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		te.t.Logf("Import script output:\n%s", string(output))
		return fmt.Errorf("import script failed: %w", err)
	}

	te.t.Logf("Import script executed successfully")
	return nil
}

// ValidateStateFile checks that terraform.tfstate exists and is not empty
func (te *TerraformExecutor) ValidateStateFile() error {
	te.t.Helper()

	statePath := filepath.Join(te.workDir, "terraform.tfstate")
	stat, err := os.Stat(statePath)
	if err != nil {
		return fmt.Errorf("terraform state file not found: %w", err)
	}

	if stat.Size() == 0 {
		return fmt.Errorf("terraform state file is empty")
	}

	te.t.Logf("Terraform state file validated: %s (%d bytes)", statePath, stat.Size())
	return nil
}

// MigrationToolExecutor handles running migration tools
type MigrationToolExecutor struct {
	tool        string
	workDir     string
	outputFiles map[string]string
	t           *testing.T
}

// NewMigrationToolExecutor creates a new migration tool executor
func NewMigrationToolExecutor(t *testing.T, tool, workDir string) *MigrationToolExecutor {
	t.Helper()

	return &MigrationToolExecutor{
		tool:    tool,
		workDir: workDir,
		outputFiles: map[string]string{
			"terraform":   filepath.Join(workDir, "migrated-resources.tf"),
			"import":      filepath.Join(workDir, "import.sh"),
			"report":      filepath.Join(workDir, "migration-report.json"),
			"manualSteps": filepath.Join(workDir, "manual-steps.md"),
		},
		t: t,
	}
}

// Execute runs the migration tool with provided environment variables
func (mte *MigrationToolExecutor) Execute(ctx context.Context, envVars map[string]string) error {
	mte.t.Helper()
	mte.t.Logf("Executing migration tool: %s", mte.tool)

	// Build command based on tool type
	var cmd *exec.Cmd
	toolBinary := filepath.Join(
		filepath.Dir(mte.workDir),
		"..",
		"..",
		"cmd",
		mte.tool,
		mte.tool,
	)

	// Check if binary exists, if not build it
	if _, err := os.Stat(toolBinary); os.IsNotExist(err) {
		mte.t.Logf("Building migration tool: %s", mte.tool)
		buildCmd := exec.CommandContext(ctx, "go", "build", "-o", toolBinary,
			fmt.Sprintf("./cmd/%s", mte.tool))
		buildCmd.Dir = filepath.Join(mte.workDir, "..", "..")

		if output, err := buildCmd.CombinedOutput(); err != nil {
			mte.t.Logf("Build output:\n%s", string(output))
			return fmt.Errorf("failed to build migration tool: %w", err)
		}
	}

	cmd = exec.CommandContext(ctx, toolBinary,
		"--output", mte.outputFiles["terraform"],
		"--import-script", mte.outputFiles["import"],
		"--report", mte.outputFiles["report"],
		"--manual-steps", mte.outputFiles["manualSteps"],
	)

	cmd.Dir = mte.workDir

	// Set environment variables
	env := os.Environ()
	for k, v := range envVars {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}
	cmd.Env = env

	output, err := cmd.CombinedOutput()
	if err != nil {
		mte.t.Logf("Migration tool output:\n%s", string(output))
		return fmt.Errorf("migration tool failed: %w", err)
	}

	mte.t.Logf("Migration tool executed successfully")
	return nil
}

// ValidateOutputFiles validates that all expected output files were created
func (mte *MigrationToolExecutor) ValidateOutputFiles() error {
	mte.t.Helper()

	for name, path := range mte.outputFiles {
		stat, err := os.Stat(path)
		if err != nil {
			return fmt.Errorf("output file %s not found: %w", name, err)
		}

		if stat.Size() == 0 {
			return fmt.Errorf("output file %s is empty", name)
		}

		mte.t.Logf("Validated output file: %s (%d bytes)", name, stat.Size())
	}

	return nil
}

// GetOutputFilePath returns the path to a specific output file
func (mte *MigrationToolExecutor) GetOutputFilePath(name string) string {
	return mte.outputFiles[name]
}

// HyperpingResourceManager handles Hyperping resource operations
type HyperpingResourceManager struct {
	client *client.Client
	t      *testing.T
}

// NewHyperpingResourceManager creates a new resource manager
func NewHyperpingResourceManager(t *testing.T, apiKey string) *HyperpingResourceManager {
	t.Helper()

	return &HyperpingResourceManager{
		client: client.NewClient(apiKey),
		t:      t,
	}
}

// ListTestMonitors lists all monitors with test prefix
func (hrm *HyperpingResourceManager) ListTestMonitors(ctx context.Context) ([]client.Monitor, error) {
	hrm.t.Helper()

	monitors, err := hrm.client.ListMonitors(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list monitors: %w", err)
	}

	// Filter to test monitors only
	var testMonitors []client.Monitor
	for _, m := range monitors {
		if strings.HasPrefix(m.Name, testResourcePrefix) {
			testMonitors = append(testMonitors, m)
		}
	}

	hrm.t.Logf("Found %d test monitors", len(testMonitors))
	return testMonitors, nil
}

// CleanupTestMonitors deletes all monitors with test prefix
func (hrm *HyperpingResourceManager) CleanupTestMonitors(ctx context.Context) error {
	hrm.t.Helper()
	hrm.t.Logf("Cleaning up test monitors from Hyperping")

	monitors, err := hrm.ListTestMonitors(ctx)
	if err != nil {
		return err
	}

	for _, monitor := range monitors {
		hrm.t.Logf("Deleting test monitor: %s (UUID: %s)", monitor.Name, monitor.UUID)
		if err := hrm.client.DeleteMonitor(ctx, monitor.UUID); err != nil {
			hrm.t.Logf("Warning: failed to delete monitor %s: %v", monitor.UUID, err)
			// Continue with other deletions
		}
	}

	hrm.t.Logf("Cleanup completed: deleted %d monitors", len(monitors))
	return nil
}

// VerifyMonitorExists checks if a monitor with the given name exists
func (hrm *HyperpingResourceManager) VerifyMonitorExists(ctx context.Context, name string) (*client.Monitor, error) {
	hrm.t.Helper()

	monitors, err := hrm.client.ListMonitors(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list monitors: %w", err)
	}

	for _, monitor := range monitors {
		if monitor.Name == name {
			hrm.t.Logf("Found monitor: %s (UUID: %s)", name, monitor.UUID)
			return &monitor, nil
		}
	}

	return nil, fmt.Errorf("monitor not found: %s", name)
}

// RunWithRetry executes a function with retry logic
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

// SetupTestCleanup registers cleanup handlers for test resources
func SetupTestCleanup(t *testing.T, cleanupFuncs ...func() error) {
	t.Helper()

	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), cleanupTimeout)
		defer cancel()

		for i, cleanupFn := range cleanupFuncs {
			if err := RunWithRetry(ctx, t, fmt.Sprintf("cleanup-%d", i), cleanupFn); err != nil {
				t.Logf("Warning: cleanup function %d failed: %v", i, err)
			}
		}
	})
}

// ValidateTerraformSyntax validates the Terraform configuration syntax
func ValidateTerraformSyntax(t *testing.T, tfFile string) error {
	t.Helper()

	content, err := os.ReadFile(tfFile)
	if err != nil {
		return fmt.Errorf("failed to read terraform file: %w", err)
	}

	tfContent := string(content)

	// Basic syntax checks
	if !strings.Contains(tfContent, "resource \"") {
		return fmt.Errorf("terraform file contains no resource blocks")
	}

	// Count braces
	openBraces := strings.Count(tfContent, "{")
	closeBraces := strings.Count(tfContent, "}")
	if openBraces != closeBraces {
		return fmt.Errorf("unbalanced braces in terraform file: %d open, %d close", openBraces, closeBraces)
	}

	t.Logf("Terraform syntax validation passed")
	return nil
}

// CountTerraformResources counts resource blocks in a Terraform file
func CountTerraformResources(t *testing.T, tfFile string) int {
	t.Helper()

	content, err := os.ReadFile(tfFile)
	require.NoError(t, err, "failed to read terraform file")

	count := strings.Count(string(content), "resource \"")
	t.Logf("Found %d resource blocks in %s", count, tfFile)
	return count
}
