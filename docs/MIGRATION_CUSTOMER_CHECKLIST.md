# Migration Customer Checklist

**Version:** 1.0.0
**Last Updated:** 2026-02-14
**Audience:** Customers, DevOps Engineers, Site Reliability Engineers

---

## Purpose

This checklist guides you through a successful migration from Better Stack, UptimeRobot, or Pingdom to Hyperping using our automated migration tools. Follow each step carefully to ensure a smooth transition with minimal risk.

**Estimated Time:** 30-90 minutes (depending on monitor count)

---

## Overview

### Migration Phases

1. **Pre-Migration** (15-30 min) - Preparation and validation
2. **Migration Execution** (5-30 min) - Running the migration tool
3. **Post-Migration** (10-30 min) - Validation and verification
4. **Cutover** (5-10 min) - Switching to production monitoring

---

## Pre-Migration Checklist

Complete these steps before running the migration tool.

### 1. Prerequisites

#### Required Software

- [ ] **Terraform 1.8.0 or later installed**
  ```bash
  terraform version
  # Should show: Terraform v1.8.0 or higher
  ```

  Install if needed:
  - macOS: `brew install terraform`
  - Linux: Download from [terraform.io](https://terraform.io)
  - Windows: Use Chocolatey `choco install terraform`

- [ ] **Go 1.21 or later installed** (for building migration tools)
  ```bash
  go version
  # Should show: go1.21 or higher
  ```

- [ ] **Git installed** (for cloning repositories)
  ```bash
  git version
  ```

- [ ] **curl and jq installed** (for API testing)
  ```bash
  curl --version
  jq --version
  ```

#### Required Access

- [ ] **Source Platform Account Access**
  - Login credentials for Better Stack / UptimeRobot / Pingdom
  - Admin or API access permissions
  - Ability to generate API keys

- [ ] **Hyperping Account Created**
  - Sign up at [hyperping.io](https://hyperping.io)
  - Account verified and active
  - Appropriate plan selected (based on monitor count)

- [ ] **Network Connectivity**
  - Access to all required API endpoints
  - No corporate firewall blocking API calls
  - Stable internet connection (not on flaky Wi-Fi)

### 2. Gather API Credentials

#### Better Stack API Token

- [ ] **Obtain Better Stack API token**
  1. Log into Better Stack dashboard
  2. Go to Settings → API Tokens
  3. Create new token with "Read" permissions
  4. Copy token (starts with `BTU` or is a long alphanumeric string)
  5. Store securely (you'll need it later)

- [ ] **Test API token**
  ```bash
  export BETTERSTACK_API_TOKEN="your_token_here"
  curl -H "Authorization: Bearer $BETTERSTACK_API_TOKEN" \
    https://betteruptime.com/api/v2/monitors | jq .
  # Should return list of monitors, not an error
  ```

#### UptimeRobot API Key

- [ ] **Obtain UptimeRobot API key**
  1. Log into UptimeRobot dashboard
  2. Go to My Settings → API Settings
  3. Generate "Main API Key"
  4. Copy key (starts with `u` followed by numbers/letters)
  5. Store securely

- [ ] **Test API key**
  ```bash
  export UPTIMEROBOT_API_KEY="your_key_here"
  curl -X POST https://api.uptimerobot.com/v2/getMonitors \
    -d "api_key=$UPTIMEROBOT_API_KEY" \
    -d "format=json" | jq .
  # Should return monitor list
  ```

#### Pingdom API Token

- [ ] **Obtain Pingdom API token**
  1. Log into Pingdom dashboard
  2. Go to Settings → Pingdom API
  3. Create new API token
  4. Copy token
  5. Store securely

- [ ] **Test API token**
  ```bash
  export PINGDOM_API_KEY="your_token_here"
  curl -H "Authorization: Bearer $PINGDOM_API_KEY" \
    https://api.pingdom.com/api/3.1/checks | jq .
  # Should return check list
  ```

#### Hyperping API Key

- [ ] **Obtain Hyperping API key**
  1. Log into Hyperping dashboard
  2. Go to Settings → API Keys
  3. Create new API key
  4. Copy key (starts with `sk_`)
  5. Store securely

- [ ] **Test API key**
  ```bash
  export HYPERPING_API_KEY="sk_your_key_here"
  curl -H "Authorization: Bearer $HYPERPING_API_KEY" \
    https://api.hyperping.io/v1/monitors | jq .
  # Should return monitor list (may be empty)
  ```

### 3. Backup Current Configuration

**CRITICAL: Always backup before migration**

- [ ] **Export current monitors from source platform**

  **Better Stack:**
  ```bash
  curl -H "Authorization: Bearer $BETTERSTACK_API_TOKEN" \
    https://betteruptime.com/api/v2/monitors > betterstack-backup.json
  ```

  **UptimeRobot:**
  ```bash
  curl -X POST https://api.uptimerobot.com/v2/getMonitors \
    -d "api_key=$UPTIMEROBOT_API_KEY" \
    -d "format=json" > uptimerobot-backup.json
  ```

  **Pingdom:**
  ```bash
  curl -H "Authorization: Bearer $PINGDOM_API_KEY" \
    https://api.pingdom.com/api/3.1/checks > pingdom-backup.json
  ```

- [ ] **Document current alert configurations**
  - List of alert contacts (email, SMS, Slack, PagerDuty)
  - Escalation policies
  - On-call schedules
  - Integration webhooks
  - Save screenshots of critical configurations

- [ ] **Document integrations**
  - Slack channels receiving alerts
  - PagerDuty integration keys
  - Webhook URLs
  - Status page configurations

- [ ] **Save backup files to secure location**
  ```bash
  mkdir -p ~/migration-backup-$(date +%Y%m%d)/
  mv *-backup.json ~/migration-backup-$(date +%Y%m%d)/
  ```

### 4. Plan Migration Scope

- [ ] **Count monitors to migrate**
  ```bash
  # Better Stack
  jq '.data | length' betterstack-backup.json

  # UptimeRobot
  jq '.monitors | length' uptimerobot-backup.json

  # Pingdom
  jq '.checks | length' pingdom-backup.json
  ```

- [ ] **Identify critical monitors**
  - List production-critical monitors
  - List monitors that MUST work immediately after migration
  - Consider migrating critical monitors first as a test

- [ ] **Identify complex monitors**
  - Heartbeat/cron monitors (Better Stack)
  - Transaction checks (Pingdom)
  - UDP monitors (UptimeRobot)
  - Monitors with complex alert logic
  - Note: These may require manual review after migration

- [ ] **Verify Hyperping plan supports monitor count**
  - Check your Hyperping plan monitor limit
  - Upgrade plan if necessary BEFORE migration
  - Leave headroom (don't migrate exactly to limit)

### 5. Read Documentation

- [ ] **Read automated migration guide**
  - [Automated Migration Tools Guide](./guides/automated-migration.md)
  - Understand what gets automated vs manual

- [ ] **Read platform-specific migration guide**
  - [Better Stack Migration Guide](./guides/migrate-from-betterstack.md)
  - [UptimeRobot Migration Guide](./guides/migrate-from-uptimerobot.md)
  - [Pingdom Migration Guide](./guides/migrate-from-pingdom.md)

- [ ] **Review known limitations**
  - Understand what cannot be migrated automatically
  - Plan manual steps for unsupported features
  - See [Migration Certification Report](../MIGRATION_CERTIFICATION.md) for full limitations list

### 6. Prepare Workspace

- [ ] **Create migration workspace directory**
  ```bash
  mkdir -p ~/hyperping-migration
  cd ~/hyperping-migration
  ```

- [ ] **Install migration tool**

  **Better Stack:**
  ```bash
  go install github.com/develeap/terraform-provider-hyperping/cmd/migrate-betterstack@latest
  ```

  **UptimeRobot:**
  ```bash
  go install github.com/develeap/terraform-provider-hyperping/cmd/migrate-uptimerobot@latest
  ```

  **Pingdom:**
  ```bash
  go install github.com/develeap/terraform-provider-hyperping/cmd/migrate-pingdom@latest
  ```

- [ ] **Verify tool installation**
  ```bash
  # Should show help text
  migrate-betterstack --help
  # or
  migrate-uptimerobot --help
  # or
  migrate-pingdom --help
  ```

- [ ] **Set up environment variables**
  ```bash
  # Add to ~/.bashrc or ~/.zshrc for persistence
  export BETTERSTACK_API_TOKEN="your_token"  # If migrating from Better Stack
  export UPTIMEROBOT_API_KEY="your_key"      # If migrating from UptimeRobot
  export PINGDOM_API_KEY="your_token"        # If migrating from Pingdom
  export HYPERPING_API_KEY="sk_your_key"     # Always needed
  ```

### 7. Test Migration (Dry Run)

**IMPORTANT: Always test before actual migration**

- [ ] **Run migration in dry-run mode**

  **Better Stack:**
  ```bash
  migrate-betterstack --dry-run --verbose
  ```

  **UptimeRobot:**
  ```bash
  migrate-uptimerobot --dry-run --verbose
  ```

  **Pingdom:**
  ```bash
  migrate-pingdom --dry-run --verbose
  ```

- [ ] **Review dry-run output**
  - Check for errors or warnings
  - Verify monitor count matches expectations
  - Note any unsupported features

- [ ] **Review generated files (dry-run creates them locally)**
  ```bash
  ls -la
  # Should see:
  # - monitors.tf (or similar)
  # - migration-report.json
  # - manual-steps.md
  ```

- [ ] **Review migration report**
  ```bash
  cat migration-report.json | jq .
  # Check:
  # - Total monitors
  # - Migrated count
  # - Warnings
  # - Errors
  ```

- [ ] **Review manual steps file**
  ```bash
  cat manual-steps.md
  # Note any monitors requiring manual configuration
  ```

### 8. Schedule Migration Window

- [ ] **Choose low-traffic time window**
  - Off-peak hours (e.g., weekend, late night)
  - Avoid during critical business periods
  - Allow 2-4 hours for complete process (migration + validation)

- [ ] **Notify stakeholders**
  - Inform team of migration schedule
  - Explain potential monitoring gaps
  - Provide rollback plan if needed

- [ ] **Prepare rollback plan**
  - Keep source platform monitors running during validation period
  - Don't delete old monitors until migration fully validated
  - Have source platform credentials ready to re-enable monitors if needed

---

## Migration Execution Checklist

Complete these steps during the actual migration.

### 1. Final Pre-Migration Verification

- [ ] **Verify all API keys still valid**
  ```bash
  # Test each API key
  curl -H "Authorization: Bearer $BETTERSTACK_API_TOKEN" \
    https://betteruptime.com/api/v2/monitors | jq '.data | length'

  curl -H "Authorization: Bearer $HYPERPING_API_KEY" \
    https://api.hyperping.io/v1/monitors | jq '.monitors | length'
  ```

- [ ] **Verify network connectivity**
  ```bash
  ping api.hyperping.io -c 3
  ping betteruptime.com -c 3  # or api.uptimerobot.com or api.pingdom.com
  ```

- [ ] **Verify sufficient disk space**
  ```bash
  df -h .
  # Should have at least 100 MB free
  ```

- [ ] **Clear previous migration attempts (if any)**
  ```bash
  rm -rf migration-output/
  rm -f migration-report.json manual-steps.md
  ```

### 2. Run Migration

- [ ] **Execute migration tool**

  **Better Stack (recommended flags):**
  ```bash
  migrate-betterstack \
    --output ./migration-output \
    --report migration-report.json \
    --manual-steps manual-steps.md \
    --verbose
  ```

  **UptimeRobot:**
  ```bash
  migrate-uptimerobot \
    --output ./migration-output/hyperping.tf \
    --report migration-report.json \
    --manual-steps manual-steps.md \
    --verbose
  ```

  **Pingdom:**
  ```bash
  migrate-pingdom \
    --output ./migration-output \
    --verbose
  ```

- [ ] **Monitor migration progress**
  - Watch terminal output for errors
  - Note any warnings
  - Migration should complete in 5-30 minutes depending on monitor count

- [ ] **Wait for completion**
  - Don't interrupt the process
  - If rate limiting occurs, tool will automatically wait and retry
  - If migration fails, check error message and consult [Support Runbook](../MIGRATION_SUPPORT_RUNBOOK.md)

### 3. Review Migration Output

- [ ] **Check migration completion status**
  ```bash
  echo $?
  # Should be 0 (success)
  ```

- [ ] **Review summary output**
  - Total monitors processed
  - Successfully migrated count
  - Warnings count
  - Errors count

- [ ] **Review migration report**
  ```bash
  cat migration-report.json | jq .
  ```

  Verify:
  - [ ] Total monitors matches expected count
  - [ ] Migrated count is acceptable (95%+ for standard monitors)
  - [ ] Warnings are understood and acceptable
  - [ ] No critical errors

- [ ] **Review manual steps document**
  ```bash
  cat manual-steps.md
  ```

  Note:
  - [ ] Monitors requiring manual configuration
  - [ ] Alert contacts that need setup
  - [ ] Integrations that need reconfiguration

- [ ] **Review generated Terraform configuration**
  ```bash
  cat migration-output/monitors.tf | less
  ```

  Spot check:
  - [ ] Monitor names look correct
  - [ ] URLs are correct
  - [ ] Check frequencies are reasonable
  - [ ] No obvious errors (syntax, placeholder values, etc.)

### 4. Validate Terraform Configuration

- [ ] **Initialize Terraform**
  ```bash
  cd migration-output/
  terraform init
  ```

  Expected output: "Terraform has been successfully initialized!"

- [ ] **Validate configuration syntax**
  ```bash
  terraform validate
  ```

  Expected output: "Success! The configuration is valid."

  If errors:
  - [ ] Review error messages
  - [ ] Check for syntax issues in .tf files
  - [ ] Consult [Support Runbook](../MIGRATION_SUPPORT_RUNBOOK.md)

- [ ] **Plan Terraform changes**
  ```bash
  terraform plan
  ```

  Review plan output:
  - [ ] Number of resources to create matches expected monitor count
  - [ ] No unexpected deletions or modifications
  - [ ] Resource configurations look correct

  **IMPORTANT:** Do NOT proceed to apply if plan looks incorrect

### 5. Apply Terraform Configuration

**This step creates resources in Hyperping**

- [ ] **Double-check plan one more time**
  ```bash
  terraform plan | grep "Plan:"
  # Should show: Plan: X to add, 0 to change, 0 to destroy
  ```

- [ ] **Apply configuration**
  ```bash
  terraform apply
  ```

  When prompted:
  - [ ] Review summary one final time
  - [ ] Type `yes` to confirm
  - [ ] Wait for completion (may take 5-15 minutes for many monitors)

- [ ] **Verify successful apply**
  - [ ] No errors in output
  - [ ] All resources created successfully
  - [ ] Check last line: "Apply complete! Resources: X added, 0 changed, 0 destroyed."

- [ ] **Save Terraform state safely**
  ```bash
  # Terraform state file contains resource IDs
  # Back it up!
  cp terraform.tfstate terraform.tfstate.backup-$(date +%Y%m%d-%H%M%S)
  ```

---

## Post-Migration Verification Checklist

Complete these steps to ensure migration was successful.

### 1. Verify Resources in Hyperping

- [ ] **Login to Hyperping dashboard**
  - Navigate to [app.hyperping.io](https://app.hyperping.io)
  - Go to Monitors page

- [ ] **Verify monitor count**
  ```bash
  # API check
  curl -H "Authorization: Bearer $HYPERPING_API_KEY" \
    https://api.hyperping.io/v1/monitors | jq '.monitors | length'

  # Should match expected count
  ```

- [ ] **Verify monitors are visible in UI**
  - [ ] All monitors appear in dashboard
  - [ ] Monitor names are correct
  - [ ] No duplicate monitors

- [ ] **Verify monitors are active (not paused)**
  ```bash
  # Check for paused monitors
  curl -H "Authorization: Bearer $HYPERPING_API_KEY" \
    https://api.hyperping.io/v1/monitors | \
    jq '.monitors[] | select(.paused == true) | .name'

  # Should return empty list (no paused monitors)
  ```

### 2. Verify Monitor Configurations

**Spot check critical monitors:**

- [ ] **Check monitor URLs**
  - Select 5-10 critical monitors
  - Verify URLs match source platform
  - Check for http vs https correctness

- [ ] **Check check frequencies**
  - Verify intervals match source platform
  - Confirm frequencies are as expected (60s, 300s, etc.)

- [ ] **Check regions**
  - Verify monitoring regions are appropriate
  - Confirm critical monitors have multiple regions

- [ ] **Check HTTP settings (if applicable)**
  - Request method (GET, POST, etc.)
  - Custom headers
  - Request body (for POST monitors)
  - Expected status codes
  - Follow redirects setting

- [ ] **Check SSL certificate monitoring (if applicable)**
  - SSL monitoring enabled where expected
  - Certificate expiry thresholds set correctly

### 3. Verify Monitors Are Running

**Wait 5-10 minutes for first checks to run**

- [ ] **Check monitor status in dashboard**
  - Monitors should show "UP" or "DOWN" status (not "PENDING")
  - Last check time should be recent (< 5 minutes ago)

- [ ] **Verify check results**
  ```bash
  # Get recent check results for a monitor
  curl -H "Authorization: Bearer $HYPERPING_API_KEY" \
    https://api.hyperping.io/v1/monitors/MONITOR_ID/checks | jq .
  ```

- [ ] **Check for any immediate failures**
  - Review monitors showing DOWN status
  - Verify if DOWN status is accurate or configuration issue
  - Fix configuration if needed

### 4. Configure Alert Contacts

**IMPORTANT: Alerts are NOT migrated automatically**

- [ ] **Review alert contact requirements**
  - Check backup documentation of source platform alerts
  - List all email addresses, phone numbers, integrations needed

- [ ] **Configure email alerts**
  - Option 1: Use Hyperping status pages with email subscribers
  - Option 2: Use external alerting (webhook to email service)

- [ ] **Configure Slack alerts** (if needed)
  - Set up Slack webhook integration
  - Test with a test monitor

- [ ] **Configure PagerDuty alerts** (if needed)
  - Set up PagerDuty integration key
  - Test escalation flow

- [ ] **Test alerting**
  - Create a test monitor that will fail
  - Verify alerts are received
  - Fix configuration if alerts not working

### 5. Configure Status Page (Optional)

If you were using status pages in source platform:

- [ ] **Create Hyperping status page**
  ```hcl
  # Add to Terraform config
  resource "hyperping_statuspage" "main" {
    name      = "Service Status"
    domain    = "status.yourcompany.com"
    is_public = true
  }
  ```

- [ ] **Attach monitors to status page**
  ```hcl
  resource "hyperping_statuspage_monitor" "monitor_attachment" {
    statuspage_id = hyperping_statuspage.main.id
    monitor_id    = hyperping_monitor.critical_api.id
  }
  ```

- [ ] **Add subscribers**
  ```hcl
  resource "hyperping_statuspage_subscriber" "team" {
    statuspage_id = hyperping_statuspage.main.id
    email        = "team@yourcompany.com"
  }
  ```

- [ ] **Test status page**
  - Visit status page URL
  - Verify monitors are displayed
  - Test subscriber notifications

### 6. Compare with Source Platform

**Parallel monitoring validation period: 24-72 hours recommended**

- [ ] **Keep source platform monitors running**
  - Do NOT delete old monitors yet
  - Run both platforms in parallel for validation

- [ ] **Spot check uptime data**
  - Compare uptime percentages
  - Verify both platforms detecting similar issues
  - Investigate discrepancies

- [ ] **Compare alert timing**
  - When an outage occurs, verify both platforms detect it
  - Check alert timing (should be similar)

- [ ] **Monitor for missed checks**
  - Verify Hyperping catching all outages
  - Check for false positives/negatives

### 7. Handle Manual Steps

- [ ] **Review manual-steps.md file**
  ```bash
  cat manual-steps.md
  ```

- [ ] **Complete all manual steps listed**
  - For each item in manual-steps.md:
    - [ ] Understand the requirement
    - [ ] Complete the manual configuration
    - [ ] Test the configuration
    - [ ] Document what was done

- [ ] **Common manual steps:**

  **Better Stack heartbeats:**
  - [ ] Review cron expressions
  - [ ] Adjust Hyperping healthcheck schedules if needed

  **Pingdom transaction checks:**
  - [ ] Review simplified monitors
  - [ ] Consider creating separate monitors for each step

  **UptimeRobot UDP monitors:**
  - [ ] Determine alternative monitoring approach
  - [ ] Implement TCP or HTTP monitoring where possible

### 8. Update Documentation

- [ ] **Update runbooks**
  - Update operational runbooks with new Hyperping monitor IDs
  - Update incident response procedures
  - Update contact information

- [ ] **Update architecture diagrams**
  - Replace old monitoring platform with Hyperping in diagrams
  - Update monitoring architecture documentation

- [ ] **Document migration**
  - Record migration date
  - Document any manual changes made
  - Note any monitors that couldn't be migrated

- [ ] **Update team knowledge base**
  - How to access Hyperping
  - How to add new monitors
  - How to configure alerts
  - Who to contact for issues

---

## Cutover and Cleanup Checklist

Complete these steps after successful validation period (24-72 hours).

### 1. Final Validation

- [ ] **Confirm Hyperping is working reliably**
  - No missed outages
  - Alerts working correctly
  - Team familiar with dashboard

- [ ] **Verify all critical monitors**
  - Review list of critical monitors from planning phase
  - Verify each is working in Hyperping
  - Confirm alerts configured correctly

- [ ] **Get team approval**
  - DevOps team confirms migration successful
  - On-call team confirms alerts working
  - Management approves cutover

### 2. Disable Source Platform Monitors

**Do this gradually, not all at once**

- [ ] **Start with non-critical monitors**
  - Pause or delete non-production monitors first
  - Monitor for any issues

- [ ] **Disable critical monitors in batches**
  - Disable 10-20% at a time
  - Wait 24 hours between batches
  - Monitor for issues

- [ ] **Fully disable all monitors**
  - Once confident, disable all source platform monitors
  - OR delete monitors to stop billing

### 3. Clean Up

- [ ] **Archive migration artifacts**
  ```bash
  mkdir -p ~/migration-archive-$(date +%Y%m%d)/
  mv ~/hyperping-migration/* ~/migration-archive-$(date +%Y%m%d)/
  ```

- [ ] **Secure Terraform state**
  - Store terraform.tfstate in secure location
  - Consider using Terraform Cloud or S3 backend
  - Do NOT commit state to Git

- [ ] **Revoke temporary API keys**
  - If you created temporary keys for migration, revoke them
  - Keep production Hyperping API key secure

- [ ] **Cancel source platform subscription** (if desired)
  - Review final billing period
  - Export historical data if needed
  - Cancel subscription

### 4. Monitoring and Feedback

- [ ] **Monitor Hyperping for 30 days**
  - Watch for any issues
  - Fine-tune configurations
  - Adjust alert thresholds

- [ ] **Collect team feedback**
  - Survey team on migration experience
  - Note any pain points
  - Document lessons learned

- [ ] **Report issues to Hyperping support**
  - Report any bugs or issues encountered
  - Suggest improvements
  - Share feedback on migration tools

---

## Success Criteria

Your migration is successful when:

- [x] All monitors migrated or accounted for (95%+ success rate expected)
- [x] All critical monitors working in Hyperping
- [x] Monitors actively checking (not paused)
- [x] Alert contacts configured and tested
- [x] Team trained on Hyperping dashboard
- [x] No missed outages compared to source platform
- [x] Terraform state saved securely
- [x] Documentation updated
- [x] Source platform monitors disabled
- [x] Team satisfied with migration outcome

---

## Rollback Procedure

If migration goes wrong, follow these steps:

### Immediate Rollback (Better Stack tool only)

```bash
# Rollback migration (deletes Hyperping resources)
migrate-betterstack --rollback --rollback-id MIGRATION_ID

# Or rollback latest migration
migrate-betterstack --rollback
```

### Manual Rollback (All tools)

1. **Keep source platform monitors running**
   - Don't disable until migration validated

2. **Delete Hyperping resources**
   ```bash
   cd migration-output/
   terraform destroy
   ```

3. **Re-enable source platform monitors** (if paused)
   - Verify all monitors active
   - Test alerts working

4. **Notify team of rollback**
   - Explain what went wrong
   - Plan remediation or retry

---

## Getting Help

### Resources

- **Documentation:**
  - [Automated Migration Guide](./guides/automated-migration.md)
  - [Better Stack Migration](./guides/migrate-from-betterstack.md)
  - [UptimeRobot Migration](./guides/migrate-from-uptimerobot.md)
  - [Pingdom Migration](./guides/migrate-from-pingdom.md)
  - [Support Runbook](../MIGRATION_SUPPORT_RUNBOOK.md)

- **Support:**
  - Email: support@hyperping.io
  - Documentation: https://docs.hyperping.io
  - GitHub Issues: https://github.com/develeap/terraform-provider-hyperping/issues

### Before Contacting Support

Gather these artifacts:

- [ ] migration-report.json
- [ ] manual-steps.md
- [ ] debug.log (if you used --debug flag)
- [ ] Terminal output (copy/paste error messages)
- [ ] Description of what went wrong
- [ ] Steps to reproduce the issue

---

## Appendix: Quick Reference Commands

### Installation

```bash
# Install migration tools
go install github.com/develeap/terraform-provider-hyperping/cmd/migrate-betterstack@latest
go install github.com/develeap/terraform-provider-hyperping/cmd/migrate-uptimerobot@latest
go install github.com/develeap/terraform-provider-hyperping/cmd/migrate-pingdom@latest
```

### Environment Setup

```bash
# Set API keys
export BETTERSTACK_API_TOKEN="your_token"
export UPTIMEROBOT_API_KEY="your_key"
export PINGDOM_API_KEY="your_token"
export HYPERPING_API_KEY="sk_your_key"
```

### Migration Commands

```bash
# Dry run (no changes)
migrate-betterstack --dry-run --verbose

# Full migration
migrate-betterstack --output ./output --verbose

# Resume from checkpoint
migrate-betterstack --resume

# Rollback
migrate-betterstack --rollback
```

### Terraform Commands

```bash
# Initialize
terraform init

# Validate
terraform validate

# Plan (preview changes)
terraform plan

# Apply (create resources)
terraform apply

# Destroy (delete resources)
terraform destroy
```

### Verification Commands

```bash
# Count monitors in Hyperping
curl -H "Authorization: Bearer $HYPERPING_API_KEY" \
  https://api.hyperping.io/v1/monitors | jq '.monitors | length'

# List monitor names
curl -H "Authorization: Bearer $HYPERPING_API_KEY" \
  https://api.hyperping.io/v1/monitors | jq '.monitors[].name'

# Check for paused monitors
curl -H "Authorization: Bearer $HYPERPING_API_KEY" \
  https://api.hyperping.io/v1/monitors | \
  jq '.monitors[] | select(.paused == true) | .name'
```

---

**Good luck with your migration!**

If you have questions or run into issues, consult the [Support Runbook](../MIGRATION_SUPPORT_RUNBOOK.md) or contact support@hyperping.io.

---

*This checklist is a living document. Feedback welcome!*
