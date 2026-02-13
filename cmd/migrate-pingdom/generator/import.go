// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package generator

import (
	"fmt"
	"strings"

	"github.com/develeap/terraform-provider-hyperping/cmd/migrate-pingdom/converter"
	"github.com/develeap/terraform-provider-hyperping/cmd/migrate-pingdom/pingdom"
)

// ImportGenerator generates Terraform import scripts.
type ImportGenerator struct {
	prefix string
}

// NewImportGenerator creates a new ImportGenerator.
func NewImportGenerator(prefix string) *ImportGenerator {
	return &ImportGenerator{
		prefix: prefix,
	}
}

// GenerateImportScript generates a shell script for importing resources.
func (g *ImportGenerator) GenerateImportScript(checks []pingdom.Check, results []converter.ConversionResult, createdResources map[int]string) string {
	var sb strings.Builder

	sb.WriteString("#!/bin/bash\n")
	sb.WriteString("# Generated Terraform import script for Pingdom -> Hyperping migration\n")
	sb.WriteString("# Run this after applying the Terraform configuration\n\n")
	sb.WriteString("set -e\n\n")

	sb.WriteString("echo \"Importing Hyperping resources into Terraform state...\"\n")
	sb.WriteString("echo \"\"\n\n")

	importCount := 0
	for i, check := range checks {
		result := results[i]

		if !result.Supported {
			continue
		}

		uuid, ok := createdResources[check.ID]
		if !ok {
			sb.WriteString(fmt.Sprintf("# Skipping Pingdom Check %d (not yet created in Hyperping)\n", check.ID))
			continue
		}

		if result.Monitor != nil {
			tfName := g.terraformName(result.Monitor.Name)
			sb.WriteString(fmt.Sprintf("# Pingdom Check %d: %s\n", check.ID, check.Name))
			sb.WriteString(fmt.Sprintf("echo \"Importing hyperping_monitor.%s...\"\n", tfName))
			sb.WriteString(fmt.Sprintf("terraform import hyperping_monitor.%s %q || echo \"Warning: Import failed for %s\"\n", tfName, uuid, tfName))
			sb.WriteString("echo \"\"\n\n")
			importCount++
		}
	}

	sb.WriteString(fmt.Sprintf("echo \"Import complete! Imported %d resources.\"\n", importCount))
	sb.WriteString("echo \"Run 'terraform plan' to verify the state matches your configuration.\"\n")

	return sb.String()
}

// GenerateImportCommands generates raw import commands without shell script wrapper.
func (g *ImportGenerator) GenerateImportCommands(checks []pingdom.Check, results []converter.ConversionResult, createdResources map[int]string) string {
	var sb strings.Builder

	sb.WriteString("# Terraform Import Commands\n")
	sb.WriteString("# Run these commands to import Hyperping resources into Terraform state\n\n")

	for i, check := range checks {
		result := results[i]

		if !result.Supported {
			continue
		}

		uuid, ok := createdResources[check.ID]
		if !ok {
			continue
		}

		if result.Monitor != nil {
			tfName := g.terraformName(result.Monitor.Name)
			sb.WriteString(fmt.Sprintf("# Pingdom Check %d: %s\n", check.ID, check.Name))
			sb.WriteString(fmt.Sprintf("terraform import hyperping_monitor.%s %q\n\n", tfName, uuid))
		}
	}

	return sb.String()
}

func (g *ImportGenerator) terraformName(name string) string {
	tg := NewTerraformGenerator(g.prefix)
	return tg.terraformName(name)
}
