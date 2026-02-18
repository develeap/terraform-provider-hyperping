// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package migrationstate

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/develeap/terraform-provider-hyperping/internal/client"
	"github.com/develeap/terraform-provider-hyperping/pkg/checkpoint"
	"github.com/develeap/terraform-provider-hyperping/pkg/recovery"
)

// PerformRollback deletes Hyperping resources created during a migration run.
func PerformRollback(migrationID string, hyperpingAPIKey string, force bool, logger *recovery.Logger) int {
	logger.Info("Starting rollback for migration: %s", migrationID)

	mgr, err := checkpoint.NewManager()
	if err != nil {
		logger.Error("Failed to create checkpoint manager: %v", err)
		return 1
	}

	cp, err := mgr.Load(migrationID)
	if err != nil {
		logger.Error("Failed to load checkpoint: %v", err)
		fmt.Fprintf(os.Stderr, "Error: Could not find checkpoint for migration: %s\n", migrationID)
		fmt.Fprintf(os.Stderr, "Use --list-checkpoints to see available migrations\n")
		return 1
	}

	if len(cp.HyperpingCreated) == 0 {
		logger.Info("No Hyperping resources to delete")
		fmt.Fprintln(os.Stderr, "No Hyperping resources were created in this migration")
		return 0
	}

	if !force {
		fmt.Fprintf(os.Stderr, "\nThis will delete %d resources from Hyperping:\n", len(cp.HyperpingCreated))
		for i, uuid := range cp.HyperpingCreated {
			if i < 10 {
				fmt.Fprintf(os.Stderr, "  - %s\n", uuid)
			} else if i == 10 {
				fmt.Fprintf(os.Stderr, "  ... and %d more\n", len(cp.HyperpingCreated)-10)
				break
			}
		}
		fmt.Fprintln(os.Stderr)

		if !recovery.ConfirmAction("Are you sure you want to delete these resources?", false) {
			logger.Info("Rollback cancelled by user")
			fmt.Fprintln(os.Stderr, "Rollback cancelled")
			return 0
		}
	}

	hpClient := client.NewClient(hyperpingAPIKey)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	backoff := recovery.DefaultBackoff()
	deletedCount := 0
	failedCount := 0

	logger.Info("Deleting %d Hyperping resources...", len(cp.HyperpingCreated))

	for _, uuid := range cp.HyperpingCreated {
		logger.Debug("Deleting resource: %s", uuid)

		err := backoff.Retry(ctx, func() error {
			return hpClient.DeleteMonitor(ctx, uuid)
		})

		if err != nil {
			logger.Error("Failed to delete resource %s: %v", uuid, err)
			fmt.Fprintf(os.Stderr, "Warning: Failed to delete %s: %v\n", uuid, err)
			failedCount++
			continue
		}

		deletedCount++
		logger.Debug("Successfully deleted resource: %s", uuid)
	}

	logger.Info("Rollback complete: %d deleted, %d failed", deletedCount, failedCount)

	if failedCount == 0 {
		if err := mgr.Delete(migrationID); err != nil {
			logger.Warn("Failed to delete checkpoint file: %v", err)
		} else {
			logger.Info("Checkpoint file deleted")
		}
	}

	fmt.Fprintln(os.Stderr, "\n=== Rollback Complete ===")
	fmt.Fprintf(os.Stderr, "Deleted: %d resources\n", deletedCount)
	if failedCount > 0 {
		fmt.Fprintf(os.Stderr, "Failed: %d resources\n", failedCount)
		fmt.Fprintln(os.Stderr, "\nSome resources could not be deleted. You may need to delete them manually.")
		return 1
	}

	fmt.Fprintln(os.Stderr, "\nAll resources successfully deleted")
	return 0
}

// ListCheckpoints displays available checkpoints, optionally filtered by tool name.
func ListCheckpoints(tool string) int {
	mgr, err := checkpoint.NewManager()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to create checkpoint manager: %v\n", err)
		return 1
	}

	checkpoints, err := mgr.List()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to list checkpoints: %v\n", err)
		return 1
	}

	if len(checkpoints) == 0 {
		fmt.Fprintln(os.Stderr, "No checkpoints found")
		return 0
	}

	var filtered []*checkpoint.Checkpoint
	for _, cp := range checkpoints {
		if tool == "" || cp.Tool == tool {
			filtered = append(filtered, cp)
		}
	}

	if len(filtered) == 0 {
		fmt.Fprintf(os.Stderr, "No checkpoints found for tool: %s\n", tool)
		return 0
	}

	fmt.Fprintf(os.Stderr, "Available checkpoints:\n\n")
	for _, cp := range filtered {
		fmt.Fprintf(os.Stderr, "Migration ID: %s\n", cp.MigrationID)
		fmt.Fprintf(os.Stderr, "  Tool: %s\n", cp.Tool)
		fmt.Fprintf(os.Stderr, "  Status: %s\n", cp.Status)
		fmt.Fprintf(os.Stderr, "  Timestamp: %s\n", cp.Timestamp.Format("2006-01-02 15:04:05"))
		fmt.Fprintf(os.Stderr, "  Progress: %d/%d processed (%d failed)\n", cp.Processed, cp.TotalResources, cp.Failed)
		if len(cp.HyperpingCreated) > 0 {
			fmt.Fprintf(os.Stderr, "  Hyperping resources: %d created\n", len(cp.HyperpingCreated))
		}
		fmt.Fprintln(os.Stderr)
	}

	return 0
}
