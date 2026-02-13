// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package generator

import (
	"fmt"
	"strings"

	"github.com/develeap/terraform-provider-hyperping/cmd/migrate-uptimerobot/converter"
)

// GenerateImportScript generates a shell script for importing existing Hyperping resources.
func GenerateImportScript(result *converter.ConversionResult) string {
	var sb strings.Builder

	// Shebang and header
	sb.WriteString("#!/bin/bash\n")
	sb.WriteString("# Terraform import script generated from UptimeRobot migration\n")
	sb.WriteString("#\n")
	sb.WriteString("# This script is only needed if resources already exist in Hyperping\n")
	sb.WriteString("# and you want to import them into Terraform state.\n")
	sb.WriteString("#\n")
	sb.WriteString("# If you're creating new resources, you don't need this script.\n")
	sb.WriteString("# Just run: terraform apply\n")
	sb.WriteString("#\n")
	sb.WriteString("# Prerequisites:\n")
	sb.WriteString("#   1. Terraform initialized: terraform init\n")
	sb.WriteString("#   2. Resources created in Hyperping manually or via API\n")
	sb.WriteString("#   3. HYPERPING_API_KEY environment variable set\n")
	sb.WriteString("#\n")
	sb.WriteString(fmt.Sprintf("# Total resources to import: %d\n", len(result.Monitors)+len(result.Healthchecks)))
	sb.WriteString("#\n\n")

	sb.WriteString("set -e  # Exit on error\n\n")

	sb.WriteString("# Colors for output\n")
	sb.WriteString("GREEN='\\033[0;32m'\n")
	sb.WriteString("RED='\\033[0;31m'\n")
	sb.WriteString("YELLOW='\\033[1;33m'\n")
	sb.WriteString("NC='\\033[0m' # No Color\n\n")

	sb.WriteString("echo \"Starting Terraform import process...\"\n")
	sb.WriteString("echo \"\"\n\n")

	// Check prerequisites
	sb.WriteString("# Check prerequisites\n")
	sb.WriteString("if [ -z \"$HYPERPING_API_KEY\" ]; then\n")
	sb.WriteString("  echo \"${RED}Error: HYPERPING_API_KEY environment variable not set${NC}\"\n")
	sb.WriteString("  exit 1\n")
	sb.WriteString("fi\n\n")

	sb.WriteString("if ! command -v terraform &> /dev/null; then\n")
	sb.WriteString("  echo \"${RED}Error: terraform command not found${NC}\"\n")
	sb.WriteString("  exit 1\n")
	sb.WriteString("fi\n\n")

	sb.WriteString("if [ ! -f \"hyperping.tf\" ]; then\n")
	sb.WriteString("  echo \"${YELLOW}Warning: hyperping.tf not found in current directory${NC}\"\n")
	sb.WriteString("  echo \"Make sure you're in the correct directory\"\n")
	sb.WriteString("fi\n\n")

	// Track success/failure
	sb.WriteString("# Track import results\n")
	sb.WriteString("SUCCESS_COUNT=0\n")
	sb.WriteString("FAIL_COUNT=0\n")
	sb.WriteString("SKIPPED_COUNT=0\n\n")

	// Import monitors
	if len(result.Monitors) > 0 {
		sb.WriteString("# ============================================\n")
		sb.WriteString("# Import Monitors\n")
		sb.WriteString("# ============================================\n\n")

		for _, m := range result.Monitors {
			sb.WriteString(fmt.Sprintf("echo \"Importing monitor: %s...\"\n", escapeShellString(m.Name)))
			sb.WriteString("# Note: You need the actual Hyperping monitor UUID\n")
			sb.WriteString("# Replace 'mon_PLACEHOLDER' with the actual UUID from Hyperping\n")
			sb.WriteString(fmt.Sprintf("if terraform import 'hyperping_monitor.%s' 'mon_PLACEHOLDER_%d' 2>/dev/null; then\n", m.ResourceName, m.OriginalID))
			sb.WriteString("  echo \"${GREEN}✓ Successfully imported${NC}\"\n")
			sb.WriteString("  ((SUCCESS_COUNT++))\n")
			sb.WriteString("else\n")
			sb.WriteString("  echo \"${YELLOW}⚠ Import failed or resource doesn't exist - will be created on apply${NC}\"\n")
			sb.WriteString("  ((SKIPPED_COUNT++))\n")
			sb.WriteString("fi\n")
			sb.WriteString("echo \"\"\n\n")
		}
	}

	// Import healthchecks
	if len(result.Healthchecks) > 0 {
		sb.WriteString("# ============================================\n")
		sb.WriteString("# Import Healthchecks\n")
		sb.WriteString("# ============================================\n\n")

		for _, h := range result.Healthchecks {
			sb.WriteString(fmt.Sprintf("echo \"Importing healthcheck: %s...\"\n", escapeShellString(h.Name)))
			sb.WriteString("# Note: You need the actual Hyperping healthcheck UUID\n")
			sb.WriteString("# Replace 'hc_PLACEHOLDER' with the actual UUID from Hyperping\n")
			sb.WriteString(fmt.Sprintf("if terraform import 'hyperping_healthcheck.%s' 'hc_PLACEHOLDER_%d' 2>/dev/null; then\n", h.ResourceName, h.OriginalID))
			sb.WriteString("  echo \"${GREEN}✓ Successfully imported${NC}\"\n")
			sb.WriteString("  ((SUCCESS_COUNT++))\n")
			sb.WriteString("else\n")
			sb.WriteString("  echo \"${YELLOW}⚠ Import failed or resource doesn't exist - will be created on apply${NC}\"\n")
			sb.WriteString("  ((SKIPPED_COUNT++))\n")
			sb.WriteString("fi\n")
			sb.WriteString("echo \"\"\n\n")
		}
	}

	// Summary
	sb.WriteString("# ============================================\n")
	sb.WriteString("# Summary\n")
	sb.WriteString("# ============================================\n\n")
	sb.WriteString("echo \"\"\n")
	sb.WriteString("echo \"Import process complete!\"\n")
	sb.WriteString("echo \"${GREEN}Successfully imported: $SUCCESS_COUNT${NC}\"\n")
	sb.WriteString("echo \"${YELLOW}Skipped (will be created): $SKIPPED_COUNT${NC}\"\n")
	sb.WriteString("if [ $FAIL_COUNT -gt 0 ]; then\n")
	sb.WriteString("  echo \"${RED}Failed: $FAIL_COUNT${NC}\"\n")
	sb.WriteString("fi\n")
	sb.WriteString("echo \"\"\n")
	sb.WriteString("echo \"Next steps:\"\n")
	sb.WriteString("echo \"  1. Run: terraform plan\"\n")
	sb.WriteString("echo \"  2. Review the plan\"\n")
	sb.WriteString("echo \"  3. Run: terraform apply\"\n")

	return sb.String()
}

// escapeShellString escapes a string for use in shell scripts.
func escapeShellString(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "$", "\\$")
	s = strings.ReplaceAll(s, "`", "\\`")
	return s
}
