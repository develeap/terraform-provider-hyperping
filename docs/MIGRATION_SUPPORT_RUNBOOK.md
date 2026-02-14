# Migration Support Runbook

**Version:** 1.0.0
**Last Updated:** 2026-02-14
**Audience:** Support Engineers, DevOps Teams, Internal Staff

---

## Purpose

This runbook provides step-by-step troubleshooting procedures for common issues encountered during migrations from Better Stack, UptimeRobot, and Pingdom to Hyperping using the automated migration tools.

---

## Table of Contents

1. [Quick Reference](#quick-reference)
2. [Common Issues & Solutions](#common-issues--solutions)
3. [Troubleshooting Flowcharts](#troubleshooting-flowcharts)
4. [Emergency Procedures](#emergency-procedures)
5. [Escalation Paths](#escalation-paths)
6. [Diagnostic Commands](#diagnostic-commands)
7. [Known Error Messages](#known-error-messages)

---

## Quick Reference

### Tool Versions

| Tool | Current Version | Minimum Terraform | Minimum Go |
|------|----------------|-------------------|------------|
| migrate-betterstack | 1.0.0+ | 1.8.0 | 1.21 |
| migrate-uptimerobot | 1.0.0+ | 1.8.0 | 1.21 |
| migrate-pingdom | 1.0.0+ | 1.8.0 | 1.21 |

### Support Contacts

| Issue Type | Contact | Response Time |
|------------|---------|---------------|
| Critical (migration failed) | support@hyperping.io | <2 hours |
| High (rollback needed) | support@hyperping.io | <4 hours |
| Medium (warnings/errors) | support@hyperping.io | <24 hours |
| Low (questions) | docs@hyperping.io | <48 hours |

### Essential Files

When troubleshooting, always request these files from the customer:

- `migration-report.json` - Detailed migration results
- `manual-steps.md` - Actions requiring manual intervention
- `debug.log` - Detailed execution log (if --debug used)
- Generated Terraform files (`*.tf`)
- Error output from terminal

---

## Common Issues & Solutions

### 1. Authentication Errors

#### Issue: "Invalid API Key" or "Unauthorized"

**Symptoms:**
```
Error: Authentication failed
Error: Invalid API key format
Error: 401 Unauthorized
```

**Diagnosis:**

1. Check API key format:
   - Better Stack: Should start with `BTU` or be a valid token
   - UptimeRobot: Should start with `u` followed by numbers and letters
   - Pingdom: Should be a valid API token (no specific prefix)
   - Hyperping: Should start with `sk_`

2. Verify API key is set correctly:
   ```bash
   # Check environment variables
   echo $BETTERSTACK_API_TOKEN
   echo $UPTIMEROBOT_API_KEY
   echo $PINGDOM_API_KEY
   echo $HYPERPING_API_KEY
   ```

3. Test API key manually:
   ```bash
   # Better Stack
   curl -H "Authorization: Bearer $BETTERSTACK_API_TOKEN" \
     https://betteruptime.com/api/v2/monitors

   # UptimeRobot
   curl -X POST https://api.uptimerobot.com/v2/getMonitors \
     -d "api_key=$UPTIMEROBOT_API_KEY" \
     -d "format=json"

   # Pingdom
   curl -H "Authorization: Bearer $PINGDOM_API_KEY" \
     https://api.pingdom.com/api/3.1/checks

   # Hyperping
   curl -H "Authorization: Bearer $HYPERPING_API_KEY" \
     https://api.hyperping.io/v1/monitors
   ```

**Solutions:**

**Solution 1: Fix API Key Format**
```bash
# Ensure no extra whitespace
export HYPERPING_API_KEY=$(echo $HYPERPING_API_KEY | xargs)

# Re-run migration
migrate-betterstack --verbose
```

**Solution 2: Generate New API Key**
1. Log into source/destination platform
2. Navigate to API settings
3. Generate new API key
4. Update environment variable
5. Re-run migration

**Solution 3: Check Account Permissions**
- Verify account has API access enabled
- Check if API key has required permissions (read for source, write for destination)
- Ensure account is not suspended or expired

**Prevention:**
- Always validate API keys before starting migration
- Use dry-run mode first: `--dry-run`
- Test with single monitor before bulk migration

---

### 2. API Rate Limiting

#### Issue: "Too Many Requests" or "Rate Limit Exceeded"

**Symptoms:**
```
Error: 429 Too Many Requests
Warning: Rate limit approached, throttling requests
Error: Rate limit exceeded, please try again later
```

**Diagnosis:**

1. Check platform rate limits:
   - Better Stack: 100 requests/minute
   - UptimeRobot: 10/min (free), 60/min (paid)
   - Pingdom: 30,000/month (varies by plan)
   - Hyperping: 100 requests/minute

2. Check migration scale:
   ```bash
   # Count monitors being migrated
   jq '.summary.total_monitors' migration-report.json
   ```

3. Review error log for retry attempts:
   ```bash
   grep "rate limit" debug.log
   grep "429" debug.log
   ```

**Solutions:**

**Solution 1: Wait and Retry**
```bash
# Check Retry-After header (if available)
# Usually 60 seconds

# Wait recommended time, then resume
sleep 60
migrate-betterstack --resume
```

**Solution 2: Use Resume Feature (Better Stack only)**
```bash
# Tool automatically saves checkpoint
# Resume from last checkpoint
migrate-betterstack --resume

# Or resume from specific checkpoint
migrate-betterstack --resume-id betterstack-20260214-120000
```

**Solution 3: Reduce Batch Size**
```bash
# For UptimeRobot free tier, migrate in smaller batches
# Export monitors to JSON first
migrate-uptimerobot --dry-run > monitors.json

# Split into batches and migrate separately
# (Manual process - filter JSON and re-run)
```

**Solution 4: Upgrade Plan (UptimeRobot)**
- Free tier: 10 requests/min
- Paid tier: 60 requests/min
- Consider temporary upgrade for migration

**Prevention:**
- Tools implement automatic backoff (no action needed)
- For large migrations (100+ monitors), expect 2-5 minute delays
- Schedule migrations during off-peak hours

---

### 3. Network Timeouts

#### Issue: "Context Deadline Exceeded" or "Connection Timeout"

**Symptoms:**
```
Error: context deadline exceeded
Error: Get "https://api.example.com": dial tcp: i/o timeout
Error: request timeout after 30 seconds
```

**Diagnosis:**

1. Check network connectivity:
   ```bash
   # Test connectivity to APIs
   ping betteruptime.com
   ping api.uptimerobot.com
   ping api.pingdom.com
   ping api.hyperping.io

   # Test HTTPS connectivity
   curl -v https://api.hyperping.io/v1/monitors
   ```

2. Check firewall/proxy settings:
   - Corporate firewall blocking API endpoints?
   - Proxy configuration needed?
   - VPN interfering with connections?

3. Check system time:
   ```bash
   # SSL certificates may fail with wrong time
   date
   timedatectl status
   ```

**Solutions:**

**Solution 1: Retry Operation**
```bash
# Simple retry (transient network issue)
migrate-betterstack --resume
```

**Solution 2: Configure Proxy**
```bash
# Set proxy if needed
export HTTP_PROXY=http://proxy.company.com:8080
export HTTPS_PROXY=http://proxy.company.com:8080

# Re-run migration
migrate-betterstack
```

**Solution 3: Increase Timeout**
```bash
# For slow networks, increase context timeout
# (Requires code modification - contact engineering)
```

**Solution 4: Check DNS Resolution**
```bash
# Verify DNS working
nslookup api.hyperping.io
dig api.hyperping.io

# Try alternate DNS (Google DNS)
export RESOLV_CONF=/etc/resolv.conf.google
```

**Prevention:**
- Run migration from stable network (avoid Wi-Fi if possible)
- Disable VPN if not required
- Test network connectivity before migration

---

### 4. Terraform Errors

#### Issue: Terraform Validation or Apply Failures

**Symptoms:**
```
Error: Invalid resource configuration
Error: terraform validate failed
Error: Error creating monitor: ...
```

**Diagnosis:**

1. Check Terraform version:
   ```bash
   terraform version
   # Requires 1.8.0+
   ```

2. Validate generated configuration:
   ```bash
   cd migration-output/
   terraform init
   terraform validate
   ```

3. Check for syntax errors:
   ```bash
   # Look for common issues
   grep -n "TODO\|FIXME\|XXX" *.tf
   ```

**Solutions:**

**Solution 1: Fix Terraform Version**
```bash
# Install correct version
# macOS
brew install terraform@1.8

# Linux
wget https://releases.hashicorp.com/terraform/1.8.0/terraform_1.8.0_linux_amd64.zip
unzip terraform_1.8.0_linux_amd64.zip
sudo mv terraform /usr/local/bin/
```

**Solution 2: Re-initialize Terraform**
```bash
cd migration-output/
rm -rf .terraform .terraform.lock.hcl
terraform init
terraform validate
```

**Solution 3: Fix Configuration Manually**
```bash
# Review terraform validate errors
terraform validate

# Edit generated .tf files to fix issues
# Common issues:
# - Invalid characters in resource names
# - Missing required fields
# - Invalid URLs
```

**Solution 4: Regenerate Configuration**
```bash
# If configuration is corrupted, regenerate
cd ..
rm -rf migration-output/
migrate-betterstack --output migration-output/
```

**Prevention:**
- Always run `terraform validate` before `terraform apply`
- Use `--validate` flag during migration
- Review generated config before applying

---

### 5. Import Script Failures

#### Issue: Import Script Returns Errors

**Symptoms:**
```
Error: Cannot import non-existent remote object
Error: Resource already exists
./import.sh: line 42: syntax error
```

**Diagnosis:**

1. Check if resources already exist:
   ```bash
   # List existing Hyperping monitors
   curl -H "Authorization: Bearer $HYPERPING_API_KEY" \
     https://api.hyperping.io/v1/monitors | jq '.monitors[].name'
   ```

2. Check import script syntax:
   ```bash
   bash -n import.sh  # Syntax check
   ```

3. Check Terraform state:
   ```bash
   terraform state list
   ```

**Solutions:**

**Solution 1: Skip Already Imported Resources**
```bash
# Edit import.sh to skip errors
# Change from:
#   terraform import ...
# To:
#   terraform import ... || echo "Already imported, skipping"
```

**Solution 2: Clean State and Re-import**
```bash
# Remove problematic state entries
terraform state rm hyperping_monitor.problematic_resource

# Re-run import for that resource
terraform import hyperping_monitor.problematic_resource mon_uuid
```

**Solution 3: Use Fresh State**
```bash
# Start with clean slate
rm terraform.tfstate terraform.tfstate.backup

# Re-run import script
bash import.sh
```

**Solution 4: Skip Import, Use Apply**
```bash
# If import not needed (creating new resources)
# Skip import.sh entirely
terraform plan
terraform apply
```

**Prevention:**
- Review import.sh before executing
- Make import.sh executable: `chmod +x import.sh`
- Run in dry-run mode first if tool supports it

---

### 6. Monitor Conversion Warnings

#### Issue: "Monitor Type Not Fully Supported" Warnings

**Symptoms:**
```
Warning: Heartbeat monitor requires manual cron review
Warning: Transaction check simplified to final step
Warning: Keyword pattern may need adjustment
```

**Diagnosis:**

1. Check migration report:
   ```bash
   jq '.warnings' migration-report.json
   ```

2. Review manual-steps.md:
   ```bash
   cat manual-steps.md
   ```

3. Identify affected monitors:
   ```bash
   jq '.monitors[] | select(.warnings | length > 0)' migration-report.json
   ```

**Solutions:**

**Solution 1: Review and Accept**
```bash
# If warnings are acceptable, proceed
# Document decisions in migration notes
terraform apply
```

**Solution 2: Manual Configuration**
```bash
# For critical monitors, configure manually
# Example: Heartbeat with complex cron

# Edit generated .tf file
vim monitors.tf

# Change from:
resource "hyperping_healthcheck" "my_heartbeat" {
  name = "My Heartbeat"
  schedule = "0 */6 * * *"  # Generated estimate
}

# To:
resource "hyperping_healthcheck" "my_heartbeat" {
  name = "My Heartbeat"
  schedule = "0 0,6,12,18 * * *"  # Actual cron expression
}
```

**Solution 3: Use Alternative Monitor Type**
```bash
# For transaction checks, create multiple simple monitors
# Instead of one complex transaction:

# Original Pingdom transaction:
# Step 1: GET /login
# Step 2: POST /login
# Step 3: GET /dashboard

# Create separate monitors:
resource "hyperping_monitor" "login_page" {
  name = "Login Page Available"
  url  = "https://example.com/login"
}

resource "hyperping_monitor" "dashboard_page" {
  name = "Dashboard Page Available"
  url  = "https://example.com/dashboard"
}
```

**Prevention:**
- Review manual-steps.md before applying
- Plan manual configuration for complex monitors
- Test critical monitors after migration

---

### 7. Checkpoint/Resume Issues

#### Issue: Cannot Resume from Checkpoint

**Symptoms:**
```
Error: No checkpoint found
Error: Checkpoint corrupted
Error: Cannot resume - migration already completed
```

**Diagnosis:**

1. List available checkpoints:
   ```bash
   migrate-betterstack --list-checkpoints
   ```

2. Check checkpoint directory:
   ```bash
   ls -la ~/.hyperping-migration/checkpoints/
   ```

3. Inspect checkpoint file:
   ```bash
   cat ~/.hyperping-migration/checkpoints/betterstack-20260214-120000.json
   ```

**Solutions:**

**Solution 1: Use Latest Checkpoint**
```bash
# Resume from most recent checkpoint
migrate-betterstack --resume
```

**Solution 2: Specify Checkpoint ID**
```bash
# List checkpoints
migrate-betterstack --list-checkpoints

# Resume from specific checkpoint
migrate-betterstack --resume-id betterstack-20260214-120000
```

**Solution 3: Start Fresh**
```bash
# If checkpoint corrupted, start over
migrate-betterstack
# (Will create new migration)
```

**Solution 4: Manually Recover**
```bash
# Extract partial data from checkpoint
jq '.state.completed_resources' \
  ~/.hyperping-migration/checkpoints/betterstack-20260214-120000.json

# Note which resources completed
# Manually skip or re-create as needed
```

**Prevention:**
- Don't delete checkpoint files until migration confirmed successful
- Use `--debug` for detailed checkpoint logging
- Keep multiple checkpoints for safety

---

### 8. Rollback Failures

#### Issue: Cannot Rollback Migration

**Symptoms:**
```
Error: Cannot find migration to rollback
Error: Resource not found during rollback
Error: Rollback failed - resources already deleted
```

**Diagnosis:**

1. List available migrations to rollback:
   ```bash
   migrate-betterstack --list-checkpoints
   ```

2. Check what resources exist in Hyperping:
   ```bash
   curl -H "Authorization: Bearer $HYPERPING_API_KEY" \
     https://api.hyperping.io/v1/monitors | jq '.monitors | length'
   ```

3. Review rollback checkpoint:
   ```bash
   jq '.state.created_resources' \
     ~/.hyperping-migration/checkpoints/betterstack-20260214-120000.json
   ```

**Solutions:**

**Solution 1: Rollback with Correct ID**
```bash
# Find migration ID
migrate-betterstack --list-checkpoints

# Rollback specific migration
migrate-betterstack --rollback --rollback-id betterstack-20260214-120000
```

**Solution 2: Force Rollback**
```bash
# Skip confirmation prompts
migrate-betterstack --rollback --rollback-id betterstack-20260214-120000 --force
```

**Solution 3: Manual Rollback**
```bash
# If automatic rollback fails, delete resources manually
# Get list of created resource IDs from checkpoint
jq -r '.state.created_resources[]' checkpoint.json

# Delete each resource
for id in $(jq -r '.state.created_resources[]' checkpoint.json); do
  curl -X DELETE \
    -H "Authorization: Bearer $HYPERPING_API_KEY" \
    https://api.hyperping.io/v1/monitors/$id
done
```

**Solution 4: Partial Rollback**
```bash
# If some resources already deleted, rollback will skip them
# Review output for "not found" errors (expected)
# Verify remaining resources deleted
```

**Prevention:**
- Always test migration with dry-run first
- Keep checkpoint files until migration fully validated
- Document created resource IDs for manual rollback if needed

---

### 9. Data Validation Failures

#### Issue: Migrated Monitors Have Incorrect Configuration

**Symptoms:**
- Monitor URLs are wrong
- Check frequencies don't match source
- Regions are incorrect
- Headers or body missing

**Diagnosis:**

1. Compare source and destination:
   ```bash
   # Export source data
   curl -H "Authorization: Bearer $BETTERSTACK_API_TOKEN" \
     https://betteruptime.com/api/v2/monitors/SOURCE_ID

   # Check Hyperping resource
   curl -H "Authorization: Bearer $HYPERPING_API_KEY" \
     https://api.hyperping.io/v1/monitors/DEST_ID
   ```

2. Review migration report:
   ```bash
   jq '.monitors[] | select(.name == "Monitor Name")' migration-report.json
   ```

3. Check generated Terraform:
   ```bash
   grep -A 20 "Monitor Name" monitors.tf
   ```

**Solutions:**

**Solution 1: Fix Terraform and Re-apply**
```bash
# Edit generated .tf file
vim monitors.tf

# Update configuration
terraform plan
terraform apply
```

**Solution 2: Regenerate Migration**
```bash
# If many errors, regenerate entirely
rm -rf migration-output/
migrate-betterstack --output migration-output/
cd migration-output/
terraform init
terraform plan
```

**Solution 3: Manual Correction in Hyperping**
```bash
# For minor issues, fix in Hyperping UI or API
curl -X PUT \
  -H "Authorization: Bearer $HYPERPING_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"url": "https://correct-url.com"}' \
  https://api.hyperping.io/v1/monitors/MONITOR_ID
```

**Prevention:**
- Always run `terraform plan` before `apply`
- Spot-check critical monitors after migration
- Use migration report to verify counts and configurations

---

### 10. Missing Features/Fields

#### Issue: Alert Contacts, Integrations Not Migrated

**Symptoms:**
```
Warning: Alert contacts not migrated (manual setup required)
Warning: Slack integration not migrated
Note: Contact groups must be configured manually
```

**Diagnosis:**

1. Check manual-steps.md:
   ```bash
   cat manual-steps.md
   grep -i "alert\|contact\|integration" manual-steps.md
   ```

2. Review migration report warnings:
   ```bash
   jq '.warnings[] | select(.category == "alert" or .category == "integration")' \
     migration-report.json
   ```

**Solutions:**

**Solution 1: Configure Alerts Manually**
```bash
# Document current alerting setup from source platform
# Example for Better Stack:
curl -H "Authorization: Bearer $BETTERSTACK_API_TOKEN" \
  https://betteruptime.com/api/v2/policies > alert-policies.json

# Configure in Hyperping:
# - Status pages for public notifications
# - Webhook integrations for Slack/PagerDuty
# - Email/SMS via external services
```

**Solution 2: Use Hyperping Status Pages**
```hcl
# Create status page for notifications
resource "hyperping_statuspage" "main" {
  name        = "Service Status"
  domain      = "status.example.com"
  is_public   = true
}

# Attach monitors to status page
resource "hyperping_statuspage_monitor" "monitor_attachment" {
  statuspage_id = hyperping_statuspage.main.id
  monitor_id    = hyperping_monitor.my_monitor.id
}

# Add subscribers
resource "hyperping_statuspage_subscriber" "team_email" {
  statuspage_id = hyperping_statuspage.main.id
  email        = "team@example.com"
}
```

**Solution 3: External Alerting**
```bash
# Use webhooks to connect to existing systems
# Configure Hyperping webhooks (if available) to:
# - PagerDuty integration URL
# - Slack webhook URL
# - Custom alerting service

# Terraform example (if webhook resource exists):
resource "hyperping_webhook" "pagerduty" {
  name = "PagerDuty Integration"
  url  = "https://events.pagerduty.com/integration/YOUR_KEY/enqueue"
}
```

**Prevention:**
- Document current alerting setup before migration
- Plan manual configuration time (30-60 min per integration)
- Test alerting after migration

---

## Troubleshooting Flowcharts

### Migration Failure Diagnostic Tree

```
Migration failed
├─ Authentication error?
│  ├─ YES → Check API keys (Issue #1)
│  └─ NO  → Continue
│
├─ Network timeout?
│  ├─ YES → Check connectivity (Issue #3)
│  └─ NO  → Continue
│
├─ Rate limit error?
│  ├─ YES → Wait and resume (Issue #2)
│  └─ NO  → Continue
│
├─ Terraform error?
│  ├─ YES → Validate config (Issue #4)
│  └─ NO  → Continue
│
├─ Conversion warning?
│  ├─ YES → Review manual steps (Issue #6)
│  └─ NO  → Contact support (Escalate)
```

### Rollback Decision Tree

```
Should I rollback?
├─ Migration completed but monitors wrong?
│  ├─ YES → Fix Terraform, re-apply (don't rollback)
│  └─ NO  → Continue
│
├─ Migration failed mid-process?
│  ├─ Can resume? → Resume (Issue #7)
│  └─ Cannot resume → Rollback (Issue #8)
│
├─ Customer changed mind?
│  └─ Rollback (Issue #8)
│
└─ Testing only?
   └─ Rollback (Issue #8)
```

### Performance Issue Tree

```
Migration taking too long
├─ Large number of monitors (100+)?
│  ├─ YES → Expected (see benchmarks)
│  └─ NO  → Continue
│
├─ Frequent rate limit warnings?
│  ├─ YES → Expected, tool handles (Issue #2)
│  └─ NO  → Continue
│
├─ Network slow?
│  ├─ YES → Check connectivity (Issue #3)
│  └─ NO  → Contact support
│
└─ Progress stalled?
   └─ Check debug log, contact support
```

---

## Emergency Procedures

### CRITICAL: Migration Created Duplicate Monitors

**Severity:** HIGH
**Impact:** Potential double-alerting, increased costs

**Immediate Actions:**

1. **Pause all new monitors in Hyperping**
   ```bash
   # Get list of created monitors
   jq -r '.state.created_resources[]' checkpoint.json > created-monitors.txt

   # Pause each monitor
   while read monitor_id; do
     curl -X PUT \
       -H "Authorization: Bearer $HYPERPING_API_KEY" \
       -H "Content-Type: application/json" \
       -d '{"paused": true}' \
       https://api.hyperping.io/v1/monitors/$monitor_id
   done < created-monitors.txt
   ```

2. **Verify duplication**
   ```bash
   # List all monitors
   curl -H "Authorization: Bearer $HYPERPING_API_KEY" \
     https://api.hyperping.io/v1/monitors | jq '.monitors[] | {id, name, url}'

   # Look for duplicates by name or URL
   ```

3. **Decide on action**
   - **Option A:** Rollback migration (if just completed)
   - **Option B:** Delete duplicates manually (if migration was days ago)

4. **Execute rollback if chosen**
   ```bash
   migrate-betterstack --rollback --rollback-id MIGRATION_ID --force
   ```

5. **Document incident**
   - What caused duplication?
   - Were monitors running in both platforms?
   - How was it resolved?

---

### CRITICAL: Migration Deleted Production Monitors

**Severity:** CRITICAL
**Impact:** Monitoring gaps, potential outages

**Immediate Actions:**

1. **DO NOT PANIC** - Most migrations don't delete from source platform

2. **Verify source platform**
   ```bash
   # Check if monitors still exist in source
   # Better Stack
   curl -H "Authorization: Bearer $BETTERSTACK_API_TOKEN" \
     https://betteruptime.com/api/v2/monitors

   # Count monitors
   jq '.data | length' monitors.json
   ```

3. **If monitors still in source - GOOD**
   - Migration tools don't delete from source by default
   - Re-run migration if needed

4. **If monitors missing from source - BAD**
   - Check platform's trash/archive (Better Stack has 30-day trash)
   - Contact source platform support immediately
   - Restore from backup if available

5. **Restore monitoring coverage immediately**
   ```bash
   # Use Terraform to create monitors from last known config
   terraform plan
   terraform apply

   # Or manually create critical monitors in Hyperping
   ```

6. **Escalate to engineering**
   - This should not happen with standard migration tools
   - Engineering investigation required

---

### CRITICAL: API Keys Compromised

**Severity:** CRITICAL
**Impact:** Security breach, unauthorized access

**Immediate Actions:**

1. **Rotate all API keys immediately**
   - Generate new Hyperping API key
   - Generate new source platform API key
   - Update all systems using old keys

2. **Review API access logs**
   ```bash
   # Check for suspicious activity in Hyperping
   # (Contact Hyperping support for access logs)
   ```

3. **Audit created resources**
   ```bash
   # List all monitors created in last 24 hours
   curl -H "Authorization: Bearer $NEW_HYPERPING_API_KEY" \
     https://api.hyperping.io/v1/monitors | \
     jq '.monitors[] | select(.created_at > "2026-02-13T00:00:00Z")'
   ```

4. **Document incident**
   - When were keys compromised?
   - What was accessed?
   - What actions were taken?

5. **Follow company security procedures**

---

## Escalation Paths

### When to Escalate

| Severity | Examples | Escalate To | Response Time |
|----------|----------|-------------|---------------|
| **CRITICAL** | Data loss, security breach, production down | Engineering + Security | Immediate |
| **HIGH** | Migration completely failed, rollback failed | Engineering | <2 hours |
| **MEDIUM** | Many monitors failed conversion, poor performance | Engineering | <24 hours |
| **LOW** | Single monitor issue, documentation question | Support | <48 hours |

### Escalation Procedure

**Step 1: Gather Information**

Collect these artifacts before escalating:
- [ ] Migration report (`migration-report.json`)
- [ ] Manual steps file (`manual-steps.md`)
- [ ] Debug log (if `--debug` was used)
- [ ] Generated Terraform files
- [ ] Terminal output (full error messages)
- [ ] Checkpoint files (if applicable)
- [ ] Customer description of issue
- [ ] Steps to reproduce

**Step 2: Create Support Ticket**

Template:
```
Subject: [SEVERITY] Migration Issue - [Brief Description]

Customer: [Name/ID]
Tool: migrate-[betterstack|uptimerobot|pingdom]
Version: [Tool Version]
Migration ID: [From checkpoint or report]

Issue Description:
[Detailed description of what went wrong]

Steps to Reproduce:
1. [Step 1]
2. [Step 2]
3. [Step 3]

Expected Behavior:
[What should have happened]

Actual Behavior:
[What actually happened]

Attachments:
- migration-report.json
- debug.log
- terminal-output.txt

Impact:
[How this affects the customer]

Attempted Solutions:
[What troubleshooting steps were already tried]
```

**Step 3: Engineering Handoff**

For HIGH/CRITICAL issues:
1. Notify engineering on-call via pager
2. Provide ticket number and brief summary
3. Make artifacts available (upload to secure location)
4. Be available for questions

**Step 4: Customer Communication**

Keep customer informed:
- Acknowledge issue immediately
- Provide realistic timeline
- Update every 2-4 hours (CRITICAL) or daily (MEDIUM)
- Explain resolution when complete

---

## Diagnostic Commands

### Environment Validation

```bash
# Check all prerequisites
echo "=== Tool Versions ==="
migrate-betterstack --version 2>/dev/null || echo "migrate-betterstack not installed"
migrate-uptimerobot --version 2>/dev/null || echo "migrate-uptimerobot not installed"
migrate-pingdom --version 2>/dev/null || echo "migrate-pingdom not installed"
terraform version

echo -e "\n=== Go Version ==="
go version

echo -e "\n=== API Keys Set ==="
[ -n "$BETTERSTACK_API_TOKEN" ] && echo "BETTERSTACK_API_TOKEN: SET" || echo "BETTERSTACK_API_TOKEN: NOT SET"
[ -n "$UPTIMEROBOT_API_KEY" ] && echo "UPTIMEROBOT_API_KEY: SET" || echo "UPTIMEROBOT_API_KEY: NOT SET"
[ -n "$PINGDOM_API_KEY" ] && echo "PINGDOM_API_KEY: SET" || echo "PINGDOM_API_KEY: NOT SET"
[ -n "$HYPERPING_API_KEY" ] && echo "HYPERPING_API_KEY: SET" || echo "HYPERPING_API_KEY: SET"

echo -e "\n=== Network Connectivity ==="
curl -s -o /dev/null -w "Better Stack: %{http_code}\n" https://betteruptime.com/api/v2/monitors
curl -s -o /dev/null -w "UptimeRobot: %{http_code}\n" https://api.uptimerobot.com
curl -s -o /dev/null -w "Pingdom: %{http_code}\n" https://api.pingdom.com
curl -s -o /dev/null -w "Hyperping: %{http_code}\n" https://api.hyperping.io

echo -e "\n=== Disk Space ==="
df -h . | tail -1
```

### Migration State Inspection

```bash
# List all checkpoints
echo "=== Available Checkpoints ==="
migrate-betterstack --list-checkpoints

# Inspect specific checkpoint
echo -e "\n=== Checkpoint Details ==="
CHECKPOINT_FILE=$(ls -t ~/.hyperping-migration/checkpoints/*.json | head -1)
echo "Latest checkpoint: $CHECKPOINT_FILE"
jq . $CHECKPOINT_FILE

# Summary of checkpoint
echo -e "\n=== Checkpoint Summary ==="
jq '{
  migration_id: .migration_id,
  tool: .tool,
  created_at: .created_at,
  completed_count: (.state.completed_resources | length),
  created_count: (.state.created_resources | length)
}' $CHECKPOINT_FILE
```

### Migration Report Analysis

```bash
# Parse migration report
echo "=== Migration Summary ==="
jq '{
  total_monitors: .summary.total_monitors,
  migrated: .summary.migrated_monitors,
  warnings: (.warnings | length),
  errors: (.errors | length)
}' migration-report.json

# List all warnings
echo -e "\n=== Warnings ==="
jq -r '.warnings[] | "\(.severity): \(.message)"' migration-report.json

# List all errors
echo -e "\n=== Errors ==="
jq -r '.errors[] | "\(.severity): \(.message)"' migration-report.json

# Monitors requiring manual steps
echo -e "\n=== Monitors Needing Manual Review ==="
jq -r '.monitors[] | select(.warnings | length > 0) | .name' migration-report.json
```

### Terraform State Analysis

```bash
# List all resources in state
echo "=== Terraform State ==="
terraform state list

# Count resources by type
echo -e "\n=== Resource Counts ==="
terraform state list | cut -d. -f1 | sort | uniq -c

# Show specific resource
echo -e "\n=== Resource Details ==="
terraform state show hyperping_monitor.example
```

---

## Known Error Messages

### Better Stack Errors

| Error Message | Meaning | Solution |
|--------------|---------|----------|
| `401 Unauthorized` | Invalid API token | Check API key format, regenerate if needed |
| `429 Too Many Requests` | Rate limit exceeded | Wait 60 seconds, resume migration |
| `404 Not Found` | Monitor doesn't exist | Check monitor ID, may have been deleted |
| `422 Unprocessable Entity` | Invalid data format | Check API request payload, report to engineering |
| `monitor_type not supported` | Unsupported monitor type | Review manual-steps.md for workaround |

### UptimeRobot Errors

| Error Message | Meaning | Solution |
|--------------|---------|----------|
| `Invalid API key` | Wrong API key format | Verify API key starts with 'u' |
| `Rate limit reached` | API rate limit hit | Wait (free: 60s, paid: varies), then resume |
| `No monitors found` | Account has no monitors | Verify correct account, check API permissions |
| `Invalid monitor type` | Unknown type code | Check UptimeRobot docs for type mapping |

### Pingdom Errors

| Error Message | Meaning | Solution |
|--------------|---------|----------|
| `Unauthorized` | Invalid credentials | Check API key, verify account active |
| `Forbidden` | Insufficient permissions | Verify API key has read permissions |
| `Not Found` | Check doesn't exist | Verify check ID, may be deleted |
| `Transaction check not supported` | Multi-step check | Will be simplified to final step |
| `DNS check cannot be migrated` | DNS resolution check | Use DNS-over-HTTPS workaround |

### Hyperping Errors

| Error Message | Meaning | Solution |
|--------------|---------|----------|
| `Invalid API key format` | Key doesn't start with sk_ | Regenerate API key from Hyperping dashboard |
| `Monitor limit reached` | Plan limit hit | Upgrade plan or delete unused monitors |
| `Invalid URL` | Malformed monitor URL | Fix URL in Terraform config |
| `Region not supported` | Invalid region code | Use valid region (see docs) |
| `Duplicate monitor name` | Name already exists | Use unique names or add suffix |

### Terraform Errors

| Error Message | Meaning | Solution |
|--------------|---------|----------|
| `Resource already exists` | Trying to create existing resource | Use `terraform import` instead |
| `Invalid resource name` | Name has invalid characters | Use alphanumeric + underscore only |
| `Provider not found` | Hyperping provider not installed | Run `terraform init` |
| `Error validating provider` | Provider version mismatch | Update provider version |

---

## Appendix: Useful Scripts

### Complete Diagnostic Script

Save as `diagnose-migration.sh`:

```bash
#!/bin/bash
# Migration diagnostic script

echo "=== Migration Diagnostic Report ==="
echo "Generated: $(date)"
echo ""

# Check environment
echo "### Environment"
echo "OS: $(uname -s)"
echo "Terraform: $(terraform version | head -1)"
echo "Go: $(go version)"
echo ""

# Check API connectivity
echo "### API Connectivity"
for api in betteruptime.com api.uptimerobot.com api.pingdom.com api.hyperping.io; do
  status=$(curl -s -o /dev/null -w "%{http_code}" https://$api 2>/dev/null || echo "FAIL")
  echo "$api: $status"
done
echo ""

# Check for migration artifacts
echo "### Migration Artifacts"
for file in migration-report.json manual-steps.md debug.log import.sh; do
  if [ -f "$file" ]; then
    echo "$file: EXISTS ($(wc -l < $file) lines)"
  else
    echo "$file: NOT FOUND"
  fi
done
echo ""

# Parse migration report if exists
if [ -f migration-report.json ]; then
  echo "### Migration Summary"
  jq -r '"Total monitors: " + (.summary.total_monitors | tostring)' migration-report.json
  jq -r '"Migrated: " + (.summary.migrated_monitors | tostring)' migration-report.json
  jq -r '"Warnings: " + ((.warnings | length) | tostring)' migration-report.json
  jq -r '"Errors: " + ((.errors | length) | tostring)' migration-report.json
  echo ""
fi

# Check Terraform state
if [ -f terraform.tfstate ]; then
  echo "### Terraform State"
  terraform state list | cut -d. -f1 | sort | uniq -c
  echo ""
fi

echo "=== End of Report ==="
```

---

*This runbook is a living document. Update as new issues are discovered and solutions are developed.*
