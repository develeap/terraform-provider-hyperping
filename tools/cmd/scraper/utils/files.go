package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// File permission constants
const (
	FilePermPublic  = 0644 // Public docs, reports
	FilePermPrivate = 0600 // Cache, snapshots, logs (sensitive data)
)

// FileExists checks if a file exists
func FileExists(filepath string) bool {
	_, err := os.Stat(filepath)
	return err == nil
}

// SaveToFile writes a string to a file with public permissions (atomic)
// Use for non-sensitive content like reports
func SaveToFile(filename, content string) error {
	if err := AtomicWriteFile(filename, []byte(content), FilePermPublic); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}
	return nil
}

// SaveJSON writes a struct as JSON to a file with private permissions (atomic)
// Uses private permissions for security (may contain sensitive data)
func SaveJSON(filename string, data interface{}) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	if err := AtomicWriteFile(filename, jsonData, FilePermPrivate); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// AtomicWriteFile writes data to a file atomically using temp file + rename
// This prevents corruption if the process is killed during write
func AtomicWriteFile(filename string, data []byte, perm os.FileMode) error {
	// Create temp file in same directory (ensures same filesystem for atomic rename)
	dir := filepath.Dir(filename)
	tmpFile, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return fmt.Errorf("create temp: %w", err)
	}
	tmpPath := tmpFile.Name()

	// Cleanup temp file on any error
	defer func() {
		tmpFile.Close()
		if _, err := os.Stat(tmpPath); err == nil {
			os.Remove(tmpPath)
		}
	}()

	// Write data
	if _, err := tmpFile.Write(data); err != nil {
		return fmt.Errorf("write temp: %w", err)
	}

	// Sync to disk (durability guarantee)
	if err := tmpFile.Sync(); err != nil {
		return fmt.Errorf("sync: %w", err)
	}

	// Close before rename
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("close: %w", err)
	}

	// Set permissions
	if err := os.Chmod(tmpPath, perm); err != nil {
		return fmt.Errorf("chmod: %w", err)
	}

	// Atomic rename (this is the atomic operation on Unix)
	if err := os.Rename(tmpPath, filename); err != nil {
		return fmt.Errorf("rename: %w", err)
	}

	return nil
}

// LoadJSON reads a JSON file into a struct
func LoadJSON(filename string, target interface{}) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	if err := json.Unmarshal(data, target); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	return nil
}
