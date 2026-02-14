// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"fmt"

	"github.com/develeap/terraform-provider-hyperping/pkg/checkpoint"
	"github.com/develeap/terraform-provider-hyperping/pkg/recovery"
)

//nolint:unused
const (
	checkpointInterval = 10 // Save checkpoint every N resources
	toolName           = "pingdom"
)

// migrationState tracks the state of the migration
//
//nolint:unused
type migrationState struct {
	checkpoint          *checkpoint.Checkpoint
	manager             *checkpoint.Manager
	logger              *recovery.Logger
	resourceCount       int
	lastCheckpointSaved int
}

// newMigrationState creates a new migration state
//
//nolint:unused
func newMigrationState(migrationID string, totalResources int, logger *recovery.Logger) (*migrationState, error) {
	mgr, err := checkpoint.NewManager()
	if err != nil {
		return nil, fmt.Errorf("failed to create checkpoint manager: %w", err)
	}

	cp := &checkpoint.Checkpoint{
		MigrationID:      migrationID,
		Tool:             toolName,
		Status:           checkpoint.StatusInProgress,
		TotalResources:   totalResources,
		Processed:        0,
		Failed:           0,
		ProcessedIDs:     []string{},
		FailedResources:  []checkpoint.FailedResource{},
		HyperpingCreated: []string{},
		Metadata:         make(map[string]string),
	}

	return &migrationState{
		checkpoint:          cp,
		manager:             mgr,
		logger:              logger,
		resourceCount:       0,
		lastCheckpointSaved: 0,
	}, nil
}

// resumeFromCheckpoint resumes from an existing checkpoint
//
//nolint:unused
func resumeFromCheckpoint(migrationID string, logger *recovery.Logger) (*migrationState, error) {
	mgr, err := checkpoint.NewManager()
	if err != nil {
		return nil, fmt.Errorf("failed to create checkpoint manager: %w", err)
	}

	cp, err := mgr.Load(migrationID)
	if err != nil {
		return nil, fmt.Errorf("failed to load checkpoint: %w", err)
	}

	logger.Info("Resuming from checkpoint: %s", migrationID)
	logger.Info("Previous progress: %d/%d processed, %d failed", cp.Processed, cp.TotalResources, cp.Failed)

	return &migrationState{
		checkpoint:          cp,
		manager:             mgr,
		logger:              logger,
		resourceCount:       cp.Processed,
		lastCheckpointSaved: cp.Processed,
	}, nil
}

// markResourceProcessed marks a resource as successfully processed
//
//nolint:unused
func (s *migrationState) markResourceProcessed(resourceID string) {
	s.checkpoint.MarkProcessed(resourceID)
	s.resourceCount++
	s.maybeCheckpoint()
}

// markResourceFailed marks a resource as failed
//
//nolint:unused
func (s *migrationState) markResourceFailed(resourceID, resourceType, resourceName, errorMsg string) {
	s.checkpoint.MarkFailed(checkpoint.FailedResource{
		ID:    resourceID,
		Type:  resourceType,
		Name:  resourceName,
		Error: errorMsg,
	})
	s.resourceCount++
	s.maybeCheckpoint()
}

// addHyperpingResource tracks a created Hyperping resource
//
//nolint:unused
func (s *migrationState) addHyperpingResource(uuid string) {
	s.checkpoint.AddHyperpingResource(uuid)
}

// maybeCheckpoint saves checkpoint if interval reached
//
//nolint:unused
func (s *migrationState) maybeCheckpoint() {
	if s.resourceCount-s.lastCheckpointSaved >= checkpointInterval {
		s.saveCheckpoint()
	}
}

// saveCheckpoint saves the current checkpoint
//
//nolint:unused
func (s *migrationState) saveCheckpoint() {
	s.logger.Debug("Saving checkpoint (processed: %d/%d, failed: %d)",
		s.checkpoint.Processed, s.checkpoint.TotalResources, s.checkpoint.Failed)

	if err := s.manager.Save(s.checkpoint); err != nil {
		s.logger.Warn("Failed to save checkpoint: %v", err)
	} else {
		s.lastCheckpointSaved = s.resourceCount
	}
}

// finalize saves final checkpoint with completion status
//
//nolint:unused
func (s *migrationState) finalize(success bool) {
	if success {
		s.checkpoint.Status = checkpoint.StatusCompleted
	} else {
		s.checkpoint.Status = checkpoint.StatusFailed
	}

	s.logger.Info("Saving final checkpoint")
	if err := s.manager.Save(s.checkpoint); err != nil {
		s.logger.Error("Failed to save final checkpoint: %v", err)
	}
}

// isProcessed checks if a resource was already processed
//
//nolint:unused
func (s *migrationState) isProcessed(resourceID string) bool {
	return s.checkpoint.IsProcessed(resourceID)
}

// getFailureReport generates a failure report
//
//nolint:unused
func (s *migrationState) getFailureReport() string {
	if len(s.checkpoint.FailedResources) == 0 {
		return ""
	}

	report := fmt.Sprintf("\n=== Failed Resources (%d) ===\n", len(s.checkpoint.FailedResources))
	for i, failed := range s.checkpoint.FailedResources {
		report += fmt.Sprintf("\n%d. %s (%s)\n", i+1, failed.Name, failed.Type)
		report += fmt.Sprintf("   ID: %s\n", failed.ID)
		report += fmt.Sprintf("   Error: %s\n", failed.Error)
	}
	return report
}
