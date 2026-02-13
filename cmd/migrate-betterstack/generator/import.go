// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package generator

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/develeap/terraform-provider-hyperping/cmd/migrate-betterstack/converter"
)

// GenerateImportScript generates a bash script for importing resources.
func (g *Generator) GenerateImportScript(monitors []converter.ConvertedMonitor, healthchecks []converter.ConvertedHealthcheck) string {
	var sb strings.Builder

	sb.WriteString("#!/bin/bash\n")
	sb.WriteString("# Auto-generated Terraform import script\n")
	sb.WriteString("# Generated at: " + getCurrentTimestamp() + "\n\n")

	sb.WriteString("set -e  # Exit on error\n")
	sb.WriteString("set -u  # Exit on undefined variable\n\n")

	sb.WriteString("# Colors for output\n")
	sb.WriteString("RED='\\033[0;31m'\n")
	sb.WriteString("GREEN='\\033[0;32m'\n")
	sb.WriteString("YELLOW='\\033[1;33m'\n")
	sb.WriteString("NC='\\033[0m' # No Color\n\n")

	sb.WriteString("echo \"${YELLOW}Starting Terraform import...${NC}\"\n")
	sb.WriteString("echo \"\"\n\n")

	sb.WriteString("# Check prerequisites\n")
	sb.WriteString("if ! command -v terraform &> /dev/null; then\n")
	sb.WriteString("    echo \"${RED}Error: terraform not found${NC}\"\n")
	sb.WriteString("    exit 1\n")
	sb.WriteString("fi\n\n")

	sb.WriteString("if [ -z \"${HYPERPING_API_KEY:-}\" ]; then\n")
	sb.WriteString("    echo \"${RED}Error: HYPERPING_API_KEY environment variable not set${NC}\"\n")
	sb.WriteString("    exit 1\n")
	sb.WriteString("fi\n\n")

	sb.WriteString("# Initialize Terraform if needed\n")
	sb.WriteString("if [ ! -d \".terraform\" ]; then\n")
	sb.WriteString("    echo \"${YELLOW}Initializing Terraform...${NC}\"\n")
	sb.WriteString("    terraform init\n")
	sb.WriteString("    echo \"\"\n")
	sb.WriteString("fi\n\n")

	// Counters
	sb.WriteString("TOTAL=0\n")
	sb.WriteString("SUCCESS=0\n")
	sb.WriteString("FAILED=0\n\n")

	// Import monitors
	if len(monitors) > 0 {
		sb.WriteString("# ===== IMPORTING MONITORS =====\n")
		sb.WriteString("echo \"${YELLOW}Importing monitors...${NC}\"\n\n")

		for _, m := range monitors {
			sb.WriteString(fmt.Sprintf("# Monitor: %s\n", m.Name))
			sb.WriteString("TOTAL=$((TOTAL + 1))\n")
			sb.WriteString(fmt.Sprintf("if terraform import \"hyperping_monitor.%s\" \"PLACEHOLDER_UUID\"; then\n", m.ResourceName))
			sb.WriteString("    SUCCESS=$((SUCCESS + 1))\n")
			sb.WriteString(fmt.Sprintf("    echo \"${GREEN}✓ Imported monitor: %s${NC}\"\n", m.Name))
			sb.WriteString("else\n")
			sb.WriteString("    FAILED=$((FAILED + 1))\n")
			sb.WriteString(fmt.Sprintf("    echo \"${RED}✗ Failed to import monitor: %s${NC}\"\n", m.Name))
			sb.WriteString("fi\n")
			sb.WriteString("echo \"\"\n\n")
		}
	}

	// Import healthchecks
	if len(healthchecks) > 0 {
		sb.WriteString("# ===== IMPORTING HEALTHCHECKS =====\n")
		sb.WriteString("echo \"${YELLOW}Importing healthchecks...${NC}\"\n\n")

		for _, h := range healthchecks {
			sb.WriteString(fmt.Sprintf("# Healthcheck: %s\n", h.Name))
			sb.WriteString("TOTAL=$((TOTAL + 1))\n")
			sb.WriteString(fmt.Sprintf("if terraform import \"hyperping_healthcheck.%s\" \"PLACEHOLDER_UUID\"; then\n", h.ResourceName))
			sb.WriteString("    SUCCESS=$((SUCCESS + 1))\n")
			sb.WriteString(fmt.Sprintf("    echo \"${GREEN}✓ Imported healthcheck: %s${NC}\"\n", h.Name))
			sb.WriteString("else\n")
			sb.WriteString("    FAILED=$((FAILED + 1))\n")
			sb.WriteString(fmt.Sprintf("    echo \"${RED}✗ Failed to import healthcheck: %s${NC}\"\n", h.Name))
			sb.WriteString("fi\n")
			sb.WriteString("echo \"\"\n\n")
		}
	}

	// Summary
	sb.WriteString("# ===== SUMMARY =====\n")
	sb.WriteString("echo \"${YELLOW}Import Summary:${NC}\"\n")
	sb.WriteString("echo \"Total resources: $TOTAL\"\n")
	sb.WriteString("echo \"${GREEN}Successful: $SUCCESS${NC}\"\n")
	sb.WriteString("echo \"${RED}Failed: $FAILED${NC}\"\n\n")

	sb.WriteString("if [ $FAILED -eq 0 ]; then\n")
	sb.WriteString("    echo \"${GREEN}All imports completed successfully!${NC}\"\n")
	sb.WriteString("    exit 0\n")
	sb.WriteString("else\n")
	sb.WriteString("    echo \"${RED}Some imports failed. Please review the errors above.${NC}\"\n")
	sb.WriteString("    exit 1\n")
	sb.WriteString("fi\n")

	return sb.String()
}

// GenerateManualSteps generates documentation for manual migration steps.
func (g *Generator) GenerateManualSteps(monitorIssues, healthcheckIssues []converter.ConversionIssue) string {
	var sb strings.Builder

	sb.WriteString("# Manual Migration Steps\n\n")
	sb.WriteString("This document outlines manual steps required to complete the migration from Better Stack to Hyperping.\n\n")

	sb.WriteString("## Before You Begin\n\n")
	sb.WriteString("1. **Review the generated Terraform configuration** (`migrated-resources.tf`)\n")
	sb.WriteString("2. **Update the import script** with actual resource UUIDs after creating resources in Hyperping\n")
	sb.WriteString("3. **Set environment variables**:\n")
	sb.WriteString("   ```bash\n")
	sb.WriteString("   export HYPERPING_API_KEY=\"sk_your_api_key\"\n")
	sb.WriteString("   ```\n\n")

	// Collect issues by severity
	warnings := []converter.ConversionIssue{}
	errors := []converter.ConversionIssue{}

	for _, issue := range monitorIssues {
		if issue.Severity == "error" {
			errors = append(errors, issue)
		} else {
			warnings = append(warnings, issue)
		}
	}

	for _, issue := range healthcheckIssues {
		if issue.Severity == "error" {
			errors = append(errors, issue)
		} else {
			warnings = append(warnings, issue)
		}
	}

	// Critical issues
	if len(errors) > 0 {
		sb.WriteString("## ⚠️ Critical Issues (Must Fix)\n\n")
		for _, issue := range errors {
			sb.WriteString(fmt.Sprintf("### %s (%s)\n", issue.ResourceName, issue.ResourceType))
			sb.WriteString(fmt.Sprintf("- **Issue**: %s\n\n", issue.Message))
		}
	}

	// Warnings
	if len(warnings) > 0 {
		sb.WriteString("## ⚠️ Warnings (Review Recommended)\n\n")

		// Group by resource
		resourceWarnings := make(map[string][]converter.ConversionIssue)
		for _, issue := range warnings {
			key := issue.ResourceName + "::" + issue.ResourceType
			resourceWarnings[key] = append(resourceWarnings[key], issue)
		}

		for key, issues := range resourceWarnings {
			parts := strings.Split(key, "::")
			resourceName := parts[0]
			resourceType := parts[1]

			sb.WriteString(fmt.Sprintf("### %s (%s)\n", resourceName, resourceType))
			for _, issue := range issues {
				sb.WriteString(fmt.Sprintf("- %s\n", issue.Message))
			}
			sb.WriteString("\n")
		}
	}

	// Update import script
	sb.WriteString("## Update Import Script\n\n")
	sb.WriteString("The import script contains placeholder UUIDs. You need to:\n\n")
	sb.WriteString("1. **Create resources in Hyperping** by running `terraform apply`\n")
	sb.WriteString("2. **Note down the UUIDs** from the Terraform output\n")
	sb.WriteString("3. **Update the import script** replacing `PLACEHOLDER_UUID` with actual UUIDs\n\n")
	sb.WriteString("Alternatively, skip the import step if you're creating new resources (not importing existing ones).\n\n")

	// Notification setup
	sb.WriteString("## Configure Notifications\n\n")
	sb.WriteString("Better Stack notification channels are not automatically migrated. You need to:\n\n")
	sb.WriteString("1. **Email notifications**: Configure in Hyperping dashboard\n")
	sb.WriteString("2. **Slack integration**: Set up webhooks in Hyperping\n")
	sb.WriteString("3. **PagerDuty**: Configure webhook integration\n")
	sb.WriteString("4. **SMS alerts**: Not supported in Hyperping (use email/Slack instead)\n\n")

	// Status pages
	sb.WriteString("## Status Pages\n\n")
	sb.WriteString("If you have Better Stack status pages:\n\n")
	sb.WriteString("1. **Create status pages** manually in Hyperping dashboard or using Terraform\n")
	sb.WriteString("2. **Add monitors** to status page sections\n")
	sb.WriteString("3. **Configure branding** (logo, colors, custom domain)\n")
	sb.WriteString("4. **Add subscribers** for status updates\n\n")

	// On-call schedules
	sb.WriteString("## On-Call Schedules\n\n")
	sb.WriteString("Better Stack on-call schedules are not supported in Hyperping:\n\n")
	sb.WriteString("1. **Use PagerDuty** or similar tool for on-call management\n")
	sb.WriteString("2. **Configure webhook alerts** to your on-call system\n\n")

	// Testing
	sb.WriteString("## Testing\n\n")
	sb.WriteString("Before decommissioning Better Stack:\n\n")
	sb.WriteString("1. **Run both systems in parallel** for at least 1 week\n")
	sb.WriteString("2. **Compare uptime metrics** between Better Stack and Hyperping\n")
	sb.WriteString("3. **Test alert delivery** by triggering test incidents\n")
	sb.WriteString("4. **Verify all monitors** are checking correctly\n\n")

	// Final steps
	sb.WriteString("## Final Steps\n\n")
	sb.WriteString("Once Hyperping is fully validated:\n\n")
	sb.WriteString("1. **Pause Better Stack monitors** (don't delete immediately)\n")
	sb.WriteString("2. **Monitor for 48 hours** to ensure no issues\n")
	sb.WriteString("3. **Export Better Stack data** for historical records\n")
	sb.WriteString("4. **Cancel Better Stack subscription**\n")
	sb.WriteString("5. **Update documentation** with Hyperping URLs and procedures\n\n")

	return sb.String()
}

// ValidateTerraform runs terraform validate on the generated configuration.
func (g *Generator) ValidateTerraform(configFile string) error {
	cmd := exec.Command("terraform", "validate")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("validation failed: %s\n%s", err, string(output))
	}
	return nil
}
