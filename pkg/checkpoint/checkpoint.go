// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

// Package checkpoint provides checkpoint management for migration tools
package checkpoint

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	// StatusInProgress indicates migration is currently running
	StatusInProgress = "in_progress"
	// StatusCompleted indicates migration finished successfully
	StatusCompleted = "completed"
	// StatusFailed indicates migration failed
	StatusFailed = "failed"
)

// FailedResource represents a resource that failed to convert or migrate
type FailedResource struct {
	ID    string `json:"id"`
	Type  string `json:"type"`
	Error string `json:"error"`
	Name  string `json:"name,omitempty"`
}

// Checkpoint represents the state of a migration at a point in time
type Checkpoint struct {
	MigrationID      string            `json:"migration_id"`
	Tool             string            `json:"tool"`
	Timestamp        time.Time         `json:"timestamp"`
	Status           string            `json:"status"`
	TotalResources   int               `json:"total_resources"`
	Processed        int               `json:"processed"`
	Failed           int               `json:"failed"`
	ProcessedIDs     []string          `json:"processed_ids"`
	FailedResources  []FailedResource  `json:"failed_resources"`
	HyperpingCreated []string          `json:"hyperping_created,omitempty"`
	Metadata         map[string]string `json:"metadata,omitempty"`
}

// Manager handles checkpoint file operations
type Manager struct {
	checkpointDir string
}

// NewManager creates a new checkpoint manager
func NewManager() (*Manager, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %w", err)
	}

	checkpointDir := filepath.Join(homeDir, ".hyperping-migrate", "checkpoints")
	if err := os.MkdirAll(checkpointDir, 0o700); err != nil {
		return nil, fmt.Errorf("failed to create checkpoint directory: %w", err)
	}

	return &Manager{
		checkpointDir: checkpointDir,
	}, nil
}

// Save saves a checkpoint to disk atomically
func (m *Manager) Save(checkpoint *Checkpoint) error {
	if checkpoint == nil {
		return errors.New("checkpoint cannot be nil")
	}

	checkpoint.Timestamp = time.Now().UTC()

	data, err := json.MarshalIndent(checkpoint, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal checkpoint: %w", err)
	}

	filename := m.getCheckpointFilename(checkpoint.MigrationID)
	tempFile := filename + ".tmp"

	// Write to temporary file
	if err := os.WriteFile(tempFile, data, 0o600); err != nil {
		return fmt.Errorf("failed to write temporary checkpoint file: %w", err)
	}

	// Atomic rename
	if err := os.Rename(tempFile, filename); err != nil {
		os.Remove(tempFile) // Clean up temp file on error
		return fmt.Errorf("failed to rename checkpoint file: %w", err)
	}

	return nil
}

// Load loads a checkpoint from disk
func (m *Manager) Load(migrationID string) (*Checkpoint, error) {
	filename := m.getCheckpointFilename(migrationID)

	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("checkpoint not found: %s", migrationID)
		}
		return nil, fmt.Errorf("failed to read checkpoint file: %w", err)
	}

	var checkpoint Checkpoint
	if err := json.Unmarshal(data, &checkpoint); err != nil {
		return nil, fmt.Errorf("failed to unmarshal checkpoint: %w", err)
	}

	return &checkpoint, nil
}

// Delete removes a checkpoint file
func (m *Manager) Delete(migrationID string) error {
	filename := m.getCheckpointFilename(migrationID)
	if err := os.Remove(filename); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete checkpoint file: %w", err)
	}
	return nil
}

// List returns all checkpoint files
func (m *Manager) List() ([]*Checkpoint, error) {
	entries, err := os.ReadDir(m.checkpointDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read checkpoint directory: %w", err)
	}

	var checkpoints []*Checkpoint
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		migrationID := entry.Name()[:len(entry.Name())-5] // Remove .json extension
		checkpoint, err := m.Load(migrationID)
		if err != nil {
			continue // Skip invalid checkpoint files
		}

		checkpoints = append(checkpoints, checkpoint)
	}

	return checkpoints, nil
}

// FindLatest finds the most recent checkpoint for a tool
func (m *Manager) FindLatest(tool string) (*Checkpoint, error) {
	checkpoints, err := m.List()
	if err != nil {
		return nil, err
	}

	var latest *Checkpoint
	for _, cp := range checkpoints {
		if cp.Tool != tool {
			continue
		}
		if latest == nil || cp.Timestamp.After(latest.Timestamp) {
			latest = cp
		}
	}

	if latest == nil {
		return nil, fmt.Errorf("no checkpoint found for tool: %s", tool)
	}

	return latest, nil
}

// Exists checks if a checkpoint exists
func (m *Manager) Exists(migrationID string) bool {
	filename := m.getCheckpointFilename(migrationID)
	_, err := os.Stat(filename)
	return err == nil
}

// getCheckpointFilename returns the full path to a checkpoint file
func (m *Manager) getCheckpointFilename(migrationID string) string {
	return filepath.Join(m.checkpointDir, migrationID+".json")
}

// GenerateMigrationID generates a unique migration ID
func GenerateMigrationID(tool string) string {
	timestamp := time.Now().UTC().Format("20060102-150405")
	return fmt.Sprintf("%s-%s", tool, timestamp)
}

// IsProcessed checks if a resource ID has been processed
func (c *Checkpoint) IsProcessed(id string) bool {
	for _, processedID := range c.ProcessedIDs {
		if processedID == id {
			return true
		}
	}
	return false
}

// MarkProcessed marks a resource as processed
func (c *Checkpoint) MarkProcessed(id string) {
	if !c.IsProcessed(id) {
		c.ProcessedIDs = append(c.ProcessedIDs, id)
		c.Processed = len(c.ProcessedIDs)
	}
}

// MarkFailed marks a resource as failed
func (c *Checkpoint) MarkFailed(failed FailedResource) {
	c.FailedResources = append(c.FailedResources, failed)
	c.Failed = len(c.FailedResources)
}

// AddHyperpingResource tracks a created Hyperping resource
func (c *Checkpoint) AddHyperpingResource(uuid string) {
	c.HyperpingCreated = append(c.HyperpingCreated, uuid)
}
