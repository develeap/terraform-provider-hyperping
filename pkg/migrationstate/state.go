// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

// Package migrationstate provides shared migration state tracking for migration tools.
package migrationstate

import (
	"fmt"

	"github.com/develeap/terraform-provider-hyperping/pkg/checkpoint"
	"github.com/develeap/terraform-provider-hyperping/pkg/recovery"
)

// CheckpointInterval is the number of resources processed between checkpoint saves.
const CheckpointInterval = 10

// State tracks the state of a migration run.
type State struct {
	Checkpoint          *checkpoint.Checkpoint
	Manager             *checkpoint.Manager
	Logger              *recovery.Logger
	resourceCount       int
	lastCheckpointSaved int
}

// New creates a new migration state for the given tool and migration ID.
func New(toolName, migrationID string, totalResources int, logger *recovery.Logger) (*State, error) {
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
		HyperpingCreated: []checkpoint.CreatedResource{},
		Metadata:         make(map[string]string),
	}

	return &State{
		Checkpoint:          cp,
		Manager:             mgr,
		Logger:              logger,
		resourceCount:       0,
		lastCheckpointSaved: 0,
	}, nil
}

// Resume resumes from an existing checkpoint.
func Resume(migrationID string, logger *recovery.Logger) (*State, error) {
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

	return &State{
		Checkpoint:          cp,
		Manager:             mgr,
		Logger:              logger,
		resourceCount:       cp.Processed,
		lastCheckpointSaved: cp.Processed,
	}, nil
}

// MarkResourceProcessed marks a resource as successfully processed.
func (s *State) MarkResourceProcessed(resourceID string) {
	s.Checkpoint.MarkProcessed(resourceID)
	s.resourceCount++
	s.maybeCheckpoint()
}

// MarkResourceFailed marks a resource as failed.
func (s *State) MarkResourceFailed(resourceID, resourceType, resourceName, errorMsg string) {
	s.Checkpoint.MarkFailed(checkpoint.FailedResource{
		ID:    resourceID,
		Type:  resourceType,
		Name:  resourceName,
		Error: errorMsg,
	})
	s.resourceCount++
	s.maybeCheckpoint()
}

// AddHyperpingResource tracks a created Hyperping resource UUID and type.
// resourceType should be "monitor" or "healthcheck".
func (s *State) AddHyperpingResource(uuid, resourceType string) {
	s.Checkpoint.AddHyperpingResource(uuid, resourceType)
}

// maybeCheckpoint saves a checkpoint if the interval has been reached.
func (s *State) maybeCheckpoint() {
	if s.resourceCount-s.lastCheckpointSaved >= CheckpointInterval {
		s.SaveCheckpoint()
	}
}

// SaveCheckpoint saves the current checkpoint to disk.
func (s *State) SaveCheckpoint() {
	s.Logger.Debug("Saving checkpoint (processed: %d/%d, failed: %d)",
		s.Checkpoint.Processed, s.Checkpoint.TotalResources, s.Checkpoint.Failed)

	if err := s.Manager.Save(s.Checkpoint); err != nil {
		s.Logger.Warn("Failed to save checkpoint: %v", err)
	} else {
		s.lastCheckpointSaved = s.resourceCount
	}
}

// Finalize saves the final checkpoint with a completed or failed status.
func (s *State) Finalize(success bool) {
	if success {
		s.Checkpoint.Status = checkpoint.StatusCompleted
	} else {
		s.Checkpoint.Status = checkpoint.StatusFailed
	}

	s.Logger.Info("Saving final checkpoint")
	if err := s.Manager.Save(s.Checkpoint); err != nil {
		s.Logger.Error("Failed to save final checkpoint: %v", err)
	}
}

// IsProcessed returns true if the given resource ID was already processed.
func (s *State) IsProcessed(resourceID string) bool {
	return s.Checkpoint.IsProcessed(resourceID)
}

// GetFailureReport generates a human-readable report of failed resources.
func (s *State) GetFailureReport() string {
	if len(s.Checkpoint.FailedResources) == 0 {
		return ""
	}

	report := fmt.Sprintf("\n=== Failed Resources (%d) ===\n", len(s.Checkpoint.FailedResources))
	for i, failed := range s.Checkpoint.FailedResources {
		report += fmt.Sprintf("\n%d. %s (%s)\n", i+1, failed.Name, failed.Type)
		report += fmt.Sprintf("   ID: %s\n", failed.ID)
		report += fmt.Sprintf("   Error: %s\n", failed.Error)
	}
	return report
}
