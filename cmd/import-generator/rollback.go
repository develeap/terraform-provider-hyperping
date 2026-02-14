// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"time"
)

const defaultImportLogFile = ".import-log"

// ImportLog tracks all imported resources for rollback capability.
type ImportLog struct {
	Timestamp time.Time        `json:"timestamp"`
	Resources []ImportLogEntry `json:"resources"`
}

// ImportLogEntry represents a single imported resource.
type ImportLogEntry struct {
	ResourceType string    `json:"resource_type"`
	ResourceName string    `json:"resource_name"`
	ResourceID   string    `json:"resource_id"`
	ImportedAt   time.Time `json:"imported_at"`
}

// NewImportLog creates a new import log.
func NewImportLog() *ImportLog {
	return &ImportLog{
		Timestamp: time.Now(),
		Resources: make([]ImportLogEntry, 0),
	}
}

// AddImport adds an imported resource to the log.
func (il *ImportLog) AddImport(resourceType, resourceName, resourceID string) {
	il.Resources = append(il.Resources, ImportLogEntry{
		ResourceType: resourceType,
		ResourceName: resourceName,
		ResourceID:   resourceID,
		ImportedAt:   time.Now(),
	})
}

// Save writes the import log to a file.
func (il *ImportLog) Save(filename string) error {
	if filename == "" {
		filename = defaultImportLogFile
	}

	data, err := json.MarshalIndent(il, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal import log: %w", err)
	}

	if err := os.WriteFile(filename, data, 0o600); err != nil {
		return fmt.Errorf("failed to write import log: %w", err)
	}

	return nil
}

// LoadImportLog reads an import log from a file.
func LoadImportLog(filename string) (*ImportLog, error) {
	if filename == "" {
		filename = defaultImportLogFile
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read import log: %w", err)
	}

	var log ImportLog
	if err := json.Unmarshal(data, &log); err != nil {
		return nil, fmt.Errorf("failed to unmarshal import log: %w", err)
	}

	return &log, nil
}

// ImportLogExists checks if an import log file exists.
func ImportLogExists(filename string) bool {
	if filename == "" {
		filename = defaultImportLogFile
	}

	_, err := os.Stat(filename)
	return err == nil
}

// RollbackManager handles rollback operations.
type RollbackManager struct {
	logFile string
	verbose bool
	dryRun  bool
}

// NewRollbackManager creates a new rollback manager.
func NewRollbackManager(logFile string, verbose, dryRun bool) *RollbackManager {
	if logFile == "" {
		logFile = defaultImportLogFile
	}

	return &RollbackManager{
		logFile: logFile,
		verbose: verbose,
		dryRun:  dryRun,
	}
}

// Rollback removes all imported resources from Terraform state.
func (rm *RollbackManager) Rollback(ctx context.Context) error {
	// Load import log
	log, err := LoadImportLog(rm.logFile)
	if err != nil {
		return fmt.Errorf("failed to load import log: %w", err)
	}

	if len(log.Resources) == 0 {
		fmt.Println("No resources to rollback")
		return nil
	}

	// Print summary
	fmt.Println("\n" + repeatString("=", 80))
	fmt.Println("ROLLBACK PLAN")
	fmt.Println(repeatString("=", 80))
	fmt.Printf("Import log created: %s\n", log.Timestamp.Format(time.RFC3339))
	fmt.Printf("Resources to remove: %d\n\n", len(log.Resources))

	// Group by resource type
	byType := make(map[string][]ImportLogEntry)
	for _, entry := range log.Resources {
		byType[entry.ResourceType] = append(byType[entry.ResourceType], entry)
	}

	for resourceType, entries := range byType {
		fmt.Printf("  %s: %d resource(s)\n", resourceType, len(entries))
	}

	fmt.Println(repeatString("=", 80))

	// Confirm rollback
	if !rm.dryRun {
		fmt.Print("\nThis will remove all listed resources from Terraform state.\n")
		fmt.Print("Are you sure you want to proceed? (yes/no): ")
		var response string
		//nolint:errcheck
		fmt.Scanln(&response)

		if response != "yes" {
			fmt.Println("Rollback cancelled")
			return nil
		}
	}

	// Execute rollback
	return rm.executeRollback(ctx, log)
}

// executeRollback performs the actual rollback operations.
func (rm *RollbackManager) executeRollback(ctx context.Context, log *ImportLog) error {
	fmt.Println("\n" + repeatString("=", 80))
	fmt.Println("EXECUTING ROLLBACK")
	fmt.Println(repeatString("=", 80))

	successCount := 0
	failureCount := 0

	// Remove resources in reverse order
	for i := len(log.Resources) - 1; i >= 0; i-- {
		entry := log.Resources[i]
		resourceAddress := fmt.Sprintf("%s.%s", entry.ResourceType, entry.ResourceName)

		if rm.dryRun {
			fmt.Printf("[DRY RUN] Would remove: %s\n", resourceAddress)
			successCount++
			continue
		}

		// Execute terraform state rm
		cmd := exec.CommandContext(ctx, "terraform", "state", "rm", resourceAddress)
		output, err := cmd.CombinedOutput()

		if err != nil {
			failureCount++
			fmt.Printf("✗ Failed to remove %s: %v\n", resourceAddress, err)
			if rm.verbose {
				fmt.Printf("  Output: %s\n", string(output))
			}
		} else {
			successCount++
			fmt.Printf("✓ Removed %s\n", resourceAddress)
			if rm.verbose {
				fmt.Printf("  Output: %s\n", string(output))
			}
		}
	}

	// Print summary
	fmt.Println("\n" + repeatString("=", 80))
	fmt.Println("ROLLBACK SUMMARY")
	fmt.Println(repeatString("=", 80))
	fmt.Printf("Successfully removed: %d\n", successCount)
	fmt.Printf("Failed to remove:     %d\n", failureCount)
	fmt.Println(repeatString("=", 80))

	// Delete import log if successful
	if !rm.dryRun && failureCount == 0 {
		if err := os.Remove(rm.logFile); err != nil {
			fmt.Printf("Warning: Failed to delete import log: %v\n", err)
		} else {
			fmt.Println("\nImport log deleted")
		}
	}

	if failureCount > 0 {
		return fmt.Errorf("rollback completed with %d error(s)", failureCount)
	}

	return nil
}

// ShowRollbackPlan displays what would be rolled back without executing.
func (rm *RollbackManager) ShowRollbackPlan() error {
	log, err := LoadImportLog(rm.logFile)
	if err != nil {
		return fmt.Errorf("failed to load import log: %w", err)
	}

	if len(log.Resources) == 0 {
		fmt.Println("No resources to rollback")
		return nil
	}

	fmt.Println("\n" + repeatString("=", 80))
	fmt.Println("ROLLBACK PLAN")
	fmt.Println(repeatString("=", 80))
	fmt.Printf("Import log created: %s\n", log.Timestamp.Format(time.RFC3339))
	fmt.Printf("Resources that would be removed: %d\n\n", len(log.Resources))

	for _, entry := range log.Resources {
		resourceAddress := fmt.Sprintf("%s.%s", entry.ResourceType, entry.ResourceName)
		fmt.Printf("  - %s (ID: %s, imported at: %s)\n",
			resourceAddress,
			entry.ResourceID,
			entry.ImportedAt.Format(time.RFC3339))
	}

	fmt.Println(repeatString("=", 80))

	return nil
}

// VerifyRollbackPreconditions checks if rollback can proceed.
func (rm *RollbackManager) VerifyRollbackPreconditions() error {
	// Check if import log exists
	if !ImportLogExists(rm.logFile) {
		return fmt.Errorf("import log not found: %s", rm.logFile)
	}

	// Check if terraform is available
	if _, err := exec.LookPath("terraform"); err != nil {
		return fmt.Errorf("terraform command not found in PATH")
	}

	// Check if we're in a terraform directory
	if _, err := os.Stat("terraform.tfstate"); err != nil {
		if _, err := os.Stat(".terraform"); err != nil {
			return fmt.Errorf("not in a Terraform directory (no .terraform or terraform.tfstate found)")
		}
	}

	return nil
}

// CleanupRollbackFiles removes rollback-related files.
func CleanupRollbackFiles(logFile string) error {
	if logFile == "" {
		logFile = defaultImportLogFile
	}

	if ImportLogExists(logFile) {
		if err := os.Remove(logFile); err != nil {
			return fmt.Errorf("failed to remove import log: %w", err)
		}
	}

	return nil
}
