// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package generator

import (
	"fmt"
	"strings"

	"github.com/develeap/terraform-provider-hyperping/cmd/migrate-uptimerobot/converter"
	"github.com/develeap/terraform-provider-hyperping/cmd/migrate-uptimerobot/uptimerobot"
)

// GenerateManualSteps generates documentation for manual migration steps.
func GenerateManualSteps(result *converter.ConversionResult, alertContacts []uptimerobot.AlertContact) string {
	sections := []string{
		generateHeaderSection(result),
		generateEscalationPoliciesSection(result, alertContacts),
		generateHealthchecksSection(result),
		generateWarningsSection(result),
		generateSkippedResourcesSection(result),
		generateTestingSection(),
		generateDecommissioningSection(),
	}
	return strings.Join(sections, "\n")
}

// generateHeaderSection generates the title and table of contents.
func generateHeaderSection(result *converter.ConversionResult) string {
	var sb strings.Builder

	sb.WriteString("# Manual Migration Steps\n\n")
	sb.WriteString("This document outlines manual configuration steps required to complete the migration from UptimeRobot to Hyperping.\n\n")

	sb.WriteString("## Table of Contents\n\n")
	sb.WriteString("- [Escalation Policy Setup](#escalation-policy-setup)\n")
	if len(result.Healthchecks) > 0 {
		sb.WriteString("- [Healthcheck Configuration](#healthcheck-configuration)\n")
	}
	if hasWarnings(result) {
		sb.WriteString("- [Monitor Warnings](#monitor-warnings)\n")
	}
	if len(result.Skipped) > 0 {
		sb.WriteString("- [Skipped Resources](#skipped-resources)\n")
	}
	sb.WriteString("- [Testing and Validation](#testing-and-validation)\n")
	sb.WriteString("- [Decommissioning UptimeRobot](#decommissioning-uptimerobot)\n")

	return sb.String()
}

// generateEscalationPoliciesSection generates the escalation policy setup section.
func generateEscalationPoliciesSection(result *converter.ConversionResult, alertContacts []uptimerobot.AlertContact) string {
	_ = result
	var sb strings.Builder

	sb.WriteString("## Escalation Policy Setup\n\n")
	sb.WriteString("Hyperping uses escalation policies for alerting. You need to create these manually in the Hyperping dashboard.\n\n")

	if len(alertContacts) > 0 {
		sb.WriteString(generateAlertContactsSection(alertContacts))
	}

	sb.WriteString("### Steps to Create Escalation Policy\n\n")
	sb.WriteString("1. Log into your Hyperping dashboard: https://app.hyperping.io\n")
	sb.WriteString("2. Navigate to **Settings → Escalation Policies**\n")
	sb.WriteString("3. Click **Create Escalation Policy**\n")
	sb.WriteString("4. Configure notification channels:\n")
	sb.WriteString("   - Add email addresses from the list above\n")
	sb.WriteString("   - Configure webhook URLs if needed\n")
	sb.WriteString("   - Set up third-party integrations (Slack, PagerDuty, etc.)\n")
	sb.WriteString("5. Note the escalation policy UUID (format: `ep_xxxxx`)\n")
	sb.WriteString("6. Update `terraform.tfvars` with the UUID:\n\n")
	sb.WriteString("```hcl\n")
	sb.WriteString("escalation_policy = \"ep_your_uuid_here\"\n")
	sb.WriteString("```\n\n")
	sb.WriteString("7. Uncomment the `escalation_policy` line in each resource in `hyperping.tf`\n")

	return sb.String()
}

// generateAlertContactsSection generates the subsection listing UptimeRobot alert contacts.
func generateAlertContactsSection(alertContacts []uptimerobot.AlertContact) string {
	var sb strings.Builder
	contactInfo := converter.CategorizeAlertContacts(alertContacts)

	sb.WriteString("### Your UptimeRobot Alert Contacts\n\n")

	if len(contactInfo.Emails) > 0 {
		sb.WriteString("**Email Contacts:**\n\n")
		for _, email := range contactInfo.Emails {
			fmt.Fprintf(&sb, "- %s\n", email)
		}
		sb.WriteString("\n")
	}
	if len(contactInfo.SMSPhones) > 0 {
		sb.WriteString("**SMS Contacts:**\n\n")
		for _, phone := range contactInfo.SMSPhones {
			fmt.Fprintf(&sb, "- %s\n", phone)
		}
		sb.WriteString("\n")
	}
	if len(contactInfo.Webhooks) > 0 {
		sb.WriteString("**Webhook URLs:**\n\n")
		for _, webhook := range contactInfo.Webhooks {
			fmt.Fprintf(&sb, "- %s\n", webhook)
		}
		sb.WriteString("\n")
	}
	if len(contactInfo.Slack) > 0 {
		sb.WriteString("**Slack Integrations:**\n\n")
		for _, slack := range contactInfo.Slack {
			fmt.Fprintf(&sb, "- %s\n", slack)
		}
		sb.WriteString("\n")
	}
	if len(contactInfo.PagerDuty) > 0 {
		sb.WriteString("**PagerDuty Integrations:**\n\n")
		for _, pd := range contactInfo.PagerDuty {
			fmt.Fprintf(&sb, "- %s\n", pd)
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// generateHealthchecksSection generates the healthcheck configuration section.
func generateHealthchecksSection(result *converter.ConversionResult) string {
	if len(result.Healthchecks) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("## Healthcheck Configuration\n\n")
	sb.WriteString("Heartbeat monitors from UptimeRobot have been converted to Hyperping healthchecks. ")
	sb.WriteString("You need to update your scripts to ping the new Hyperping URLs.\n\n")

	for _, h := range result.Healthchecks {
		fmt.Fprintf(&sb, "### %s\n\n", h.Name)
		fmt.Fprintf(&sb, "**Original UptimeRobot Monitor ID:** %d\n\n", h.OriginalID)
		sb.WriteString("**Steps:**\n\n")
		sb.WriteString("1. Apply the Terraform configuration to create the healthcheck\n")
		sb.WriteString("2. Get the ping URL:\n\n")
		sb.WriteString("```bash\n")
		fmt.Fprintf(&sb, "terraform output -raw %s_ping_url\n", h.ResourceName)
		sb.WriteString("```\n\n")
		sb.WriteString("3. Update your script/cron job to use the new URL:\n\n")
		sb.WriteString("```bash\n")
		sb.WriteString("# Add this to your script (before UptimeRobot heartbeat)\n")
		sb.WriteString("HYPERPING_URL=$(terraform output -raw " + h.ResourceName + "_ping_url)\n")
		sb.WriteString("curl -fsS --retry 3 \"$HYPERPING_URL\" || echo \"Failed to ping Hyperping\"\n")
		sb.WriteString("```\n\n")
		sb.WriteString("4. Test the healthcheck by running your script manually\n")
		sb.WriteString("5. Verify in Hyperping dashboard that the ping was received\n")
		sb.WriteString("6. After successful testing, remove the UptimeRobot heartbeat ping\n\n")
	}

	return sb.String()
}

// generateWarningsSection generates the monitor warnings section.
func generateWarningsSection(result *converter.ConversionResult) string {
	if !hasWarnings(result) {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("## Monitor Warnings\n\n")
	sb.WriteString("The following monitors have warnings that require your attention:\n\n")

	for _, m := range result.Monitors {
		if len(m.Warnings) > 0 {
			fmt.Fprintf(&sb, "### %s\n\n", m.Name)
			for _, w := range m.Warnings {
				fmt.Fprintf(&sb, "- ⚠️ %s\n", w)
			}
			sb.WriteString("\n")
		}
	}

	for _, h := range result.Healthchecks {
		if len(h.Warnings) > 0 {
			fmt.Fprintf(&sb, "### %s (Healthcheck)\n\n", h.Name)
			for _, w := range h.Warnings {
				fmt.Fprintf(&sb, "- ⚠️ %s\n", w)
			}
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

// generateSkippedResourcesSection generates the skipped resources section.
func generateSkippedResourcesSection(result *converter.ConversionResult) string {
	if len(result.Skipped) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("## Skipped Resources\n\n")
	sb.WriteString("The following monitors could not be automatically migrated:\n\n")

	for _, s := range result.Skipped {
		fmt.Fprintf(&sb, "### %s (ID: %d)\n\n", s.Name, s.ID)
		fmt.Fprintf(&sb, "**Type:** %d\n\n", s.Type)
		fmt.Fprintf(&sb, "**Reason:** %s\n\n", s.Reason)
		sb.WriteString("**Action Required:** Manual configuration needed\n\n")
	}

	return sb.String()
}

// generateTestingSection generates the testing and validation section.
func generateTestingSection() string {
	var sb strings.Builder

	sb.WriteString("## Testing and Validation\n\n")
	sb.WriteString("### Pre-Migration Testing\n\n")
	sb.WriteString("Before switching from UptimeRobot to Hyperping:\n\n")
	sb.WriteString("1. **Parallel Operation:**\n")
	sb.WriteString("   - Keep UptimeRobot monitors active\n")
	sb.WriteString("   - Run Hyperping monitors in parallel\n")
	sb.WriteString("   - Compare alerting behavior for 1-2 weeks\n\n")
	sb.WriteString("2. **Test Alerting:**\n")
	sb.WriteString("   - Temporarily take down a test service\n")
	sb.WriteString("   - Verify Hyperping detects the outage\n")
	sb.WriteString("   - Verify alerts are received via escalation policy\n\n")
	sb.WriteString("3. **Validate Check Frequencies:**\n")
	sb.WriteString("   - Review adjusted check frequencies (see warnings above)\n")
	sb.WriteString("   - Adjust if needed for your SLAs\n\n")

	sb.WriteString("### Cutover Checklist\n\n")
	sb.WriteString("- [ ] All Terraform resources created successfully\n")
	sb.WriteString("- [ ] Escalation policies configured and tested\n")
	sb.WriteString("- [ ] Healthcheck ping URLs updated in scripts\n")
	sb.WriteString("- [ ] Parallel testing completed (1-2 weeks)\n")
	sb.WriteString("- [ ] Team trained on Hyperping dashboard\n")
	sb.WriteString("- [ ] Runbooks updated with new monitoring URLs\n")
	sb.WriteString("- [ ] On-call rotation updated with new alert channels\n\n")

	sb.WriteString("### Post-Migration Validation\n\n")
	sb.WriteString("```bash\n")
	sb.WriteString("# Verify all resources in Terraform state\n")
	sb.WriteString("terraform state list\n\n")
	sb.WriteString("# Check for configuration drift\n")
	sb.WriteString("terraform plan  # Should show no changes\n\n")
	sb.WriteString("# View monitor status via Hyperping API\n")
	sb.WriteString("curl -H \"Authorization: Bearer $HYPERPING_API_KEY\" \\\n")
	sb.WriteString("  https://api.hyperping.io/v1/monitors | jq '.[] | {name, down}'\n")
	sb.WriteString("```\n")

	return sb.String()
}

// generateDecommissioningSection generates the decommissioning UptimeRobot section.
func generateDecommissioningSection() string {
	var sb strings.Builder

	sb.WriteString("## Decommissioning UptimeRobot\n\n")
	sb.WriteString("After successful migration and validation:\n\n")
	sb.WriteString("### Week 1-2: Parallel Operation\n")
	sb.WriteString("- Run both systems in parallel\n")
	sb.WriteString("- Monitor for discrepancies\n")
	sb.WriteString("- Fine-tune Hyperping configuration\n\n")

	sb.WriteString("### Week 3: Primary Cutover\n")
	sb.WriteString("- Switch primary alerting to Hyperping\n")
	sb.WriteString("- Keep UptimeRobot as backup\n")
	sb.WriteString("- Update team documentation\n\n")

	sb.WriteString("### Week 4: Full Cutover\n")
	sb.WriteString("- Pause UptimeRobot monitors\n")
	sb.WriteString("- Export final UptimeRobot data for records\n")
	sb.WriteString("- Cancel UptimeRobot subscription\n\n")

	sb.WriteString("### Pausing UptimeRobot Monitors\n\n")
	sb.WriteString("```bash\n")
	sb.WriteString("# Pause all monitors via API\n")
	sb.WriteString("curl -X POST https://api.uptimerobot.com/v2/editMonitor \\\n")
	sb.WriteString("  -H \"Content-Type: application/json\" \\\n")
	sb.WriteString("  -d '{\n")
	sb.WriteString("    \"api_key\": \"'$UPTIMEROBOT_API_KEY'\",\n")
	sb.WriteString("    \"id\": \"MONITOR_ID\",\n")
	sb.WriteString("    \"status\": 0\n")
	sb.WriteString("  }'\n")
	sb.WriteString("```\n\n")

	sb.WriteString("## Need Help?\n\n")
	sb.WriteString("- **Hyperping Documentation:** https://hyperping.io/docs\n")
	sb.WriteString("- **Terraform Provider:** https://registry.terraform.io/providers/develeap/hyperping\n")
	sb.WriteString("- **Support:** support@hyperping.io\n")

	return sb.String()
}

// hasWarnings checks if any monitors or healthchecks have warnings.
func hasWarnings(result *converter.ConversionResult) bool {
	for _, m := range result.Monitors {
		if len(m.Warnings) > 0 {
			return true
		}
	}
	for _, h := range result.Healthchecks {
		if len(h.Warnings) > 0 {
			return true
		}
	}
	return false
}
