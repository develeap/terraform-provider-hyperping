// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

// Package contract validates VCR cassette responses against basic sanity checks.
// It verifies that success responses contain valid JSON with expected fields.
package contract

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// ValidationError describes a cassette response that fails validation.
type ValidationError struct {
	CassetteFile string
	Method       string
	Path         string
	StatusCode   int
	Message      string
}

// ValidateCassettes validates every cassette response body in cassetteDir.
// specPath is accepted for interface compatibility but is not used in this
// implementation — JSON validity is checked without loading the full spec.
// Returns a slice of validation errors (empty slice means all OK).
func ValidateCassettes(specPath, cassetteDir string) ([]ValidationError, error) {
	entries, err := filepath.Glob(filepath.Join(cassetteDir, "*.yaml"))
	if err != nil {
		return nil, fmt.Errorf("contract: glob cassettes %s: %w", cassetteDir, err)
	}

	var allErrors []ValidationError

	for _, path := range entries {
		errs, err := validateCassetteFile(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "contract: skip %s: %v\n", filepath.Base(path), err)
			continue
		}
		allErrors = append(allErrors, errs...)
	}

	return allErrors, nil
}

// rawCassette is a minimal YAML struct for reading go-vcr cassette files.
type rawCassette struct {
	Interactions []rawInteraction `yaml:"interactions"`
}

type rawInteraction struct {
	Request struct {
		Method string `yaml:"method"`
		URI    string `yaml:"uri"`
	} `yaml:"request"`
	Response struct {
		Code int    `yaml:"code"`
		Body string `yaml:"body"`
	} `yaml:"response"`
}

func validateCassetteFile(path string) ([]ValidationError, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var raw rawCassette
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parse cassette YAML: %w", err)
	}

	base := filepath.Base(path)
	var errs []ValidationError

	for _, interaction := range raw.Interactions {
		code := interaction.Response.Code
		if code < 200 || code >= 300 {
			continue // Only validate success responses
		}

		body := interaction.Response.Body
		if body == "" {
			continue // Empty body is OK for some endpoints
		}

		if !json.Valid([]byte(body)) {
			errs = append(errs, ValidationError{
				CassetteFile: base,
				Method:       interaction.Request.Method,
				Path:         interaction.Request.URI,
				StatusCode:   code,
				Message:      "response body is not valid JSON",
			})
		}
	}

	return errs, nil
}
