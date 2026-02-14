// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

const defaultCheckpointFile = ".import-checkpoint"

// ImportCheckpoint holds the state of an import operation for resume capability.
type ImportCheckpoint struct {
	Timestamp      time.Time          `json:"timestamp"`
	TotalResources int                `json:"total_resources"`
	ImportedIDs    []ImportedResource `json:"imported_ids"`
	FailedIDs      []string           `json:"failed_ids"`
	CurrentIndex   int                `json:"current_index"`
	Completed      bool               `json:"completed"`
	FilterSummary  string             `json:"filter_summary,omitempty"`
}

// ImportedResource represents a resource that was successfully imported.
type ImportedResource struct {
	ID           string `json:"id"`
	ResourceType string `json:"resource_type"`
	ResourceName string `json:"resource_name"`
}

// NewImportCheckpoint creates a new checkpoint.
func NewImportCheckpoint(totalResources int) *ImportCheckpoint {
	return &ImportCheckpoint{
		Timestamp:      time.Now(),
		TotalResources: totalResources,
		ImportedIDs:    make([]ImportedResource, 0),
		FailedIDs:      make([]string, 0),
		CurrentIndex:   0,
		Completed:      false,
	}
}

// AddImported adds an imported resource to the checkpoint.
func (ic *ImportCheckpoint) AddImported(id, resourceType, resourceName string) {
	ic.ImportedIDs = append(ic.ImportedIDs, ImportedResource{
		ID:           id,
		ResourceType: resourceType,
		ResourceName: resourceName,
	})
}

// AddFailed adds a failed resource ID to the checkpoint.
func (ic *ImportCheckpoint) AddFailed(id string) {
	ic.FailedIDs = append(ic.FailedIDs, id)
}

// IsImported checks if a resource ID has already been imported.
func (ic *ImportCheckpoint) IsImported(id string) bool {
	for _, imported := range ic.ImportedIDs {
		if imported.ID == id {
			return true
		}
	}
	return false
}

// IsFailed checks if a resource ID has failed import.
func (ic *ImportCheckpoint) IsFailed(id string) bool {
	for _, failedID := range ic.FailedIDs {
		if failedID == id {
			return true
		}
	}
	return false
}

// Save writes the checkpoint to a file.
func (ic *ImportCheckpoint) Save(filename string) error {
	if filename == "" {
		filename = defaultCheckpointFile
	}

	data, err := json.MarshalIndent(ic, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal checkpoint: %w", err)
	}

	if err := os.WriteFile(filename, data, 0o600); err != nil {
		return fmt.Errorf("failed to write checkpoint file: %w", err)
	}

	return nil
}

// LoadCheckpoint reads a checkpoint from a file.
func LoadCheckpoint(filename string) (*ImportCheckpoint, error) {
	if filename == "" {
		filename = defaultCheckpointFile
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read checkpoint file: %w", err)
	}

	var checkpoint ImportCheckpoint
	if err := json.Unmarshal(data, &checkpoint); err != nil {
		return nil, fmt.Errorf("failed to unmarshal checkpoint: %w", err)
	}

	return &checkpoint, nil
}

// CheckpointExists checks if a checkpoint file exists.
func CheckpointExists(filename string) bool {
	if filename == "" {
		filename = defaultCheckpointFile
	}

	_, err := os.Stat(filename)
	return err == nil
}

// DeleteCheckpoint removes a checkpoint file.
func DeleteCheckpoint(filename string) error {
	if filename == "" {
		filename = defaultCheckpointFile
	}

	if !CheckpointExists(filename) {
		return nil
	}

	return os.Remove(filename)
}

// PrintCheckpointSummary prints a summary of the checkpoint.
func (ic *ImportCheckpoint) PrintCheckpointSummary() {
	fmt.Println("\nCheckpoint Information:")
	fmt.Println("=======================")
	fmt.Printf("Created:          %s\n", ic.Timestamp.Format(time.RFC3339))
	fmt.Printf("Total resources:  %d\n", ic.TotalResources)
	fmt.Printf("Imported:         %d\n", len(ic.ImportedIDs))
	fmt.Printf("Failed:           %d\n", len(ic.FailedIDs))
	fmt.Printf("Progress:         %d/%d (%.1f%%)\n",
		ic.CurrentIndex, ic.TotalResources,
		float64(ic.CurrentIndex)/float64(ic.TotalResources)*100)
	fmt.Printf("Completed:        %t\n", ic.Completed)

	if ic.FilterSummary != "" {
		fmt.Printf("Filters:          %s\n", ic.FilterSummary)
	}

	fmt.Println()
}

// CheckpointManager handles checkpoint operations.
type CheckpointManager struct {
	filename string
	enabled  bool
}

// NewCheckpointManager creates a new checkpoint manager.
func NewCheckpointManager(filename string, enabled bool) *CheckpointManager {
	if filename == "" {
		filename = defaultCheckpointFile
	}

	return &CheckpointManager{
		filename: filename,
		enabled:  enabled,
	}
}

// Save saves a checkpoint if enabled.
func (cm *CheckpointManager) Save(checkpoint *ImportCheckpoint) error {
	if !cm.enabled {
		return nil
	}
	return checkpoint.Save(cm.filename)
}

// Load loads a checkpoint.
func (cm *CheckpointManager) Load() (*ImportCheckpoint, error) {
	return LoadCheckpoint(cm.filename)
}

// Exists checks if a checkpoint exists.
func (cm *CheckpointManager) Exists() bool {
	return CheckpointExists(cm.filename)
}

// Delete deletes the checkpoint file.
func (cm *CheckpointManager) Delete() error {
	return DeleteCheckpoint(cm.filename)
}

// FilterJobsForResume filters import jobs based on checkpoint state.
func FilterJobsForResume(jobs []ImportJob, checkpoint *ImportCheckpoint) []ImportJob {
	if checkpoint == nil {
		return jobs
	}

	filtered := make([]ImportJob, 0, len(jobs))
	for _, job := range jobs {
		// Skip already imported or failed resources
		if !checkpoint.IsImported(job.ResourceID) && !checkpoint.IsFailed(job.ResourceID) {
			filtered = append(filtered, job)
		}
	}

	return filtered
}

// PromptForResume asks the user if they want to resume from checkpoint.
func PromptForResume(checkpoint *ImportCheckpoint) bool {
	checkpoint.PrintCheckpointSummary()

	fmt.Printf("Resume from checkpoint? (y/N): ")
	var response string
	//nolint:errcheck
	fmt.Scanln(&response)

	return response == "y" || response == "Y"
}
