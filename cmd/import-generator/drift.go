// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

// DriftDetector handles drift detection operations.
type DriftDetector struct {
	verbose bool
}

// NewDriftDetector creates a new drift detector.
func NewDriftDetector(verbose bool) *DriftDetector {
	return &DriftDetector{
		verbose: verbose,
	}
}

// DriftResult holds the results of drift detection.
type DriftResult struct {
	HasDrift         bool
	DriftedResources []DriftedResource
	PlanOutput       string
	Error            error
}

// DriftedResource represents a resource with detected drift.
type DriftedResource struct {
	Address     string
	ChangeType  string // "update", "delete", "create"
	Description string
}

// DetectDrift runs terraform plan to detect configuration drift.
func (dd *DriftDetector) DetectDrift(ctx context.Context) (*DriftResult, error) {
	result := &DriftResult{
		DriftedResources: make([]DriftedResource, 0),
	}

	fmt.Println("\n" + repeatString("=", 80))
	fmt.Println("DRIFT DETECTION")
	fmt.Println(repeatString("=", 80))
	fmt.Println("Running terraform plan to detect configuration drift...")
	fmt.Println()

	// Run terraform plan with detailed exit code
	cmd := exec.CommandContext(ctx, "terraform", "plan", "-detailed-exitcode", "-no-color")
	output, err := cmd.CombinedOutput()
	result.PlanOutput = string(output)

	// Handle exit codes
	// 0 = no changes
	// 1 = error
	// 2 = changes detected
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			switch exitErr.ExitCode() {
			case 0:
				// No drift
				result.HasDrift = false
				fmt.Println("✓ No drift detected - state matches configuration")
				return result, nil

			case 2:
				// Drift detected
				result.HasDrift = true
				result.DriftedResources = dd.parsePlanOutput(result.PlanOutput)
				dd.printDriftSummary(result)
				return result, nil

			default:
				// Error
				result.Error = fmt.Errorf("terraform plan failed with exit code %d", exitErr.ExitCode())
				fmt.Printf("✗ Terraform plan failed: %v\n", result.Error)
				if dd.verbose {
					fmt.Printf("\nOutput:\n%s\n", result.PlanOutput)
				}
				return result, result.Error
			}
		}

		result.Error = fmt.Errorf("terraform plan failed: %w", err)
		return result, result.Error
	}

	// No drift (exit code 0)
	result.HasDrift = false
	fmt.Println("✓ No drift detected - state matches configuration")
	return result, nil
}

// parsePlanOutput extracts drifted resources from terraform plan output.
func (dd *DriftDetector) parsePlanOutput(output string) []DriftedResource {
	drifted := make([]DriftedResource, 0)

	// Regex patterns for different change types
	updatePattern := regexp.MustCompile(`^\s*#\s+(\S+)\s+will be updated in-place`)
	deletePattern := regexp.MustCompile(`^\s*#\s+(\S+)\s+will be destroyed`)
	createPattern := regexp.MustCompile(`^\s*#\s+(\S+)\s+will be created`)
	replacePattern := regexp.MustCompile(`^\s*#\s+(\S+)\s+must be replaced`)

	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()

		if matches := updatePattern.FindStringSubmatch(line); matches != nil {
			drifted = append(drifted, DriftedResource{
				Address:     matches[1],
				ChangeType:  "update",
				Description: "will be updated in-place",
			})
		} else if matches := deletePattern.FindStringSubmatch(line); matches != nil {
			drifted = append(drifted, DriftedResource{
				Address:     matches[1],
				ChangeType:  "delete",
				Description: "will be destroyed",
			})
		} else if matches := createPattern.FindStringSubmatch(line); matches != nil {
			drifted = append(drifted, DriftedResource{
				Address:     matches[1],
				ChangeType:  "create",
				Description: "will be created",
			})
		} else if matches := replacePattern.FindStringSubmatch(line); matches != nil {
			drifted = append(drifted, DriftedResource{
				Address:     matches[1],
				ChangeType:  "replace",
				Description: "must be replaced",
			})
		}
	}

	return drifted
}

// printDriftSummary prints a formatted summary of detected drift.
func (dd *DriftDetector) printDriftSummary(result *DriftResult) {
	fmt.Println("⚠️  Configuration drift detected!")
	fmt.Println()

	if len(result.DriftedResources) > 0 {
		// Group by change type
		byType := make(map[string][]DriftedResource)
		for _, dr := range result.DriftedResources {
			byType[dr.ChangeType] = append(byType[dr.ChangeType], dr)
		}

		// Print summary by type
		for changeType, resources := range byType {
			fmt.Printf("  %s (%d):\n", strings.ToUpper(changeType), len(resources))
			for _, r := range resources {
				fmt.Printf("    - %s (%s)\n", r.Address, r.Description)
			}
			fmt.Println()
		}
	}

	fmt.Printf("Total drifted resources: %d\n", len(result.DriftedResources))
	fmt.Println(repeatString("=", 80))
}

// PromptToContinueWithDrift asks the user if they want to continue despite drift.
func PromptToContinueWithDrift() bool {
	fmt.Println()
	fmt.Println("⚠️  WARNING: Importing resources with existing drift may cause unexpected results.")
	fmt.Println("It's recommended to resolve drift before importing new resources.")
	fmt.Println()
	fmt.Print("Do you want to continue anyway? (yes/no): ")

	var response string
	//nolint:errcheck
	fmt.Scanln(&response)

	return response == "yes"
}

// VerifyTerraformInit checks if terraform has been initialized.
func VerifyTerraformInit() error {
	cmd := exec.Command("terraform", "version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("terraform command not found or not working: %w", err)
	}

	// Check for .terraform directory
	cmd = exec.Command("test", "-d", ".terraform")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("terraform not initialized (run 'terraform init')")
	}

	return nil
}

// RefreshState runs terraform refresh to update state from remote.
func RefreshState(ctx context.Context) error {
	fmt.Println("Refreshing Terraform state...")

	cmd := exec.CommandContext(ctx, "terraform", "refresh")
	output, err := cmd.CombinedOutput()

	if err != nil {
		return fmt.Errorf("terraform refresh failed: %w\nOutput: %s", err, string(output))
	}

	fmt.Println("✓ State refreshed successfully")
	return nil
}

// ValidateTerraformConfig runs terraform validate.
func ValidateTerraformConfig(ctx context.Context) error {
	fmt.Println("Validating Terraform configuration...")

	cmd := exec.CommandContext(ctx, "terraform", "validate")
	output, err := cmd.CombinedOutput()

	if err != nil {
		return fmt.Errorf("terraform validate failed: %w\nOutput: %s", err, string(output))
	}

	fmt.Println("✓ Configuration is valid")
	return nil
}

// PostImportDriftCheck runs drift detection after import to verify clean import.
func (dd *DriftDetector) PostImportDriftCheck(ctx context.Context) error {
	fmt.Println("\n" + repeatString("=", 80))
	fmt.Println("POST-IMPORT DRIFT CHECK")
	fmt.Println(repeatString("=", 80))
	fmt.Println("Verifying imported resources match configuration...")
	fmt.Println()

	result, err := dd.DetectDrift(ctx)
	if err != nil {
		return fmt.Errorf("post-import drift check failed: %w", err)
	}

	if result.HasDrift {
		fmt.Println()
		fmt.Println("⚠️  Post-import drift detected!")
		fmt.Println("This may indicate:")
		fmt.Println("  - HCL configuration doesn't match actual resource state")
		fmt.Println("  - Resources were modified during import")
		fmt.Println("  - Optional attributes have different defaults")
		fmt.Println()
		fmt.Println("Recommended actions:")
		fmt.Println("  1. Review drift details above")
		fmt.Println("  2. Update HCL configuration to match actual state")
		fmt.Println("  3. Run 'terraform apply' to align configuration")
		fmt.Println()

		return fmt.Errorf("drift detected after import")
	}

	fmt.Println()
	fmt.Println("✓ Post-import drift check passed")
	fmt.Println("  All imported resources match their HCL configuration")
	fmt.Println()

	return nil
}

// DriftDetectionOptions holds options for drift detection.
type DriftDetectionOptions struct {
	Enabled      bool
	AbortOnDrift bool
	Verbose      bool
	RefreshFirst bool
}

// RunPreImportDriftCheck performs drift detection before import with options.
func RunPreImportDriftCheck(ctx context.Context, opts DriftDetectionOptions) error {
	if !opts.Enabled {
		return nil
	}

	dd := NewDriftDetector(opts.Verbose)

	// Optionally refresh state first
	if opts.RefreshFirst {
		if err := RefreshState(ctx); err != nil {
			return err
		}
	}

	// Detect drift
	result, err := dd.DetectDrift(ctx)
	if err != nil {
		return err
	}

	if !result.HasDrift {
		return nil
	}

	// Drift detected
	if opts.AbortOnDrift {
		return fmt.Errorf("aborting due to detected drift (use --no-abort-on-drift to continue)")
	}

	// Prompt user
	if !PromptToContinueWithDrift() {
		return fmt.Errorf("import cancelled by user")
	}

	return nil
}
