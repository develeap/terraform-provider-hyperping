// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// SafeReadFile reads a file after validating it's within the expected base directory.
// This prevents directory traversal attacks (CWE-22).
func SafeReadFile(baseDir, filePath string) ([]byte, error) {
	// Resolve both paths to absolute
	absBase, err := filepath.Abs(baseDir)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve base directory: %w", err)
	}

	absFile, err := filepath.Abs(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve file path: %w", err)
	}

	// Clean paths to remove any .. or . components
	absBase = filepath.Clean(absBase)
	absFile = filepath.Clean(absFile)

	// Ensure the file path is within the base directory
	if !strings.HasPrefix(absFile, absBase+string(filepath.Separator)) && absFile != absBase {
		return nil, fmt.Errorf("path %q is outside base directory %q", filePath, baseDir)
	}

	// Now safe to read the file
	return os.ReadFile(absFile) // #nosec G304 -- path validated above
}
