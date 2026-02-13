# Cron Healthcheck Module Tests

This directory contains test configurations and examples for the cron-healthcheck module.

## Test Files

### basic.tf
Simple test with common cron job patterns.

**Usage:**
```bash
cd tests
terraform init
terraform plan
terraform apply

# View outputs
terraform output healthcheck_ids
terraform output -raw backup_ping_url

# Cleanup
terraform destroy
```

### advanced.tf
Complex test scenarios including:
- Multiple environments (production, staging, maintenance)
- Custom name formats
- Escalation policies
- Paused jobs

**Usage:**
```bash
cd tests
terraform init
terraform plan -out=advanced.tfplan
terraform apply advanced.tfplan

# View specific outputs
terraform output prod_healthcheck_ids
terraform output -json prod_ping_urls | jq

# Cleanup
terraform destroy
```

### validation.tf
Input validation tests - demonstrates valid configurations and includes commented-out invalid examples.

**Usage:**
```bash
cd tests
terraform validate  # Should pass for valid configurations

# To test validation failures, uncomment an invalid example and run:
terraform validate  # Should fail with appropriate error message
```

### integration.sh
Shell script with practical integration examples showing how to use ping URLs in real cron jobs.

**Usage:**
```bash
chmod +x integration.sh
./integration.sh  # Displays examples (doesn't execute them)
```

## Running Tests

### 1. Prerequisites

```bash
# Set API key
export HYPERPING_API_KEY="your-api-key-here"

# Or create terraform.tfvars (not recommended for production)
echo 'hyperping_api_key = "your-api-key-here"' > terraform.tfvars
```

### 2. Basic Test Flow

```bash
# Initialize
terraform init

# Validate configuration
terraform validate

# Plan (dry run)
terraform plan

# Apply (create resources)
terraform apply

# Test ping URLs manually
curl -v "$(terraform output -raw backup_ping_url)"

# Cleanup
terraform destroy
```

### 3. Testing Individual Scenarios

```bash
# Test only basic module
terraform plan -target=module.basic_cron_jobs

# Test only advanced module
terraform plan -target=module.production_cron_jobs

# Apply specific module
terraform apply -target=module.basic_cron_jobs
```

## Validation Tests

### Testing Valid Configurations

All modules in `validation.tf` should validate successfully:

```bash
terraform validate
# Success! The configuration is valid.
```

### Testing Invalid Configurations

Uncomment invalid examples in `validation.tf` to test error handling:

```bash
# Uncomment invalid_cron_format
terraform validate
# Error: Invalid value for variable
#   on validation.tf line X:
#   All cron schedules must be in standard 5-field format
```

## Integration Testing

### Manual Integration Test

1. Deploy the module:
```bash
terraform apply -auto-approve
```

2. Extract ping URL:
```bash
PING_URL=$(terraform output -raw backup_ping_url)
```

3. Create test script:
```bash
cat > /tmp/test-cron.sh << 'EOF'
#!/bin/bash
echo "Job running at $(date)"
curl -fsS "$PING_URL" > /dev/null
echo "Ping sent successfully"
EOF

chmod +x /tmp/test-cron.sh
```

4. Run test script:
```bash
PING_URL="$PING_URL" /tmp/test-cron.sh
```

5. Verify in Hyperping dashboard that ping was received.

### Automated Test Script

```bash
#!/bin/bash
# automated-test.sh

set -e

echo "=== Running Cron Healthcheck Module Tests ==="

# 1. Validate
echo "Step 1: Validating configuration..."
terraform validate
echo "✓ Validation passed"

# 2. Plan
echo "Step 2: Planning infrastructure..."
terraform plan -out=test.tfplan
echo "✓ Plan created"

# 3. Apply
echo "Step 3: Creating resources..."
terraform apply -auto-approve test.tfplan
echo "✓ Resources created"

# 4. Test ping URLs
echo "Step 4: Testing ping URLs..."
terraform output -json ping_urls | jq -r '.[]' | while read url; do
    if curl -fsS --max-time 5 "$url" > /dev/null 2>&1; then
        echo "✓ Ping successful: $url"
    else
        echo "✗ Ping failed: $url"
        exit 1
    fi
done

# 5. Verify outputs
echo "Step 5: Verifying outputs..."
COUNT=$(terraform output -raw healthcheck_count)
echo "✓ Created $COUNT healthchecks"

# 6. Cleanup
echo "Step 6: Cleaning up..."
terraform destroy -auto-approve
echo "✓ Resources destroyed"

echo ""
echo "=== All Tests Passed ==="
```

## Expected Outputs

### basic.tf

```
healthcheck_count = 3

healthcheck_ids = {
  "daily_backup"   = "hc_abc123def456"
  "hourly_sync"    = "hc_def456ghi789"
  "weekly_report"  = "hc_ghi789jkl012"
}

backup_ping_url = <sensitive>
all_ping_urls = <sensitive>
healthcheck_details = <sensitive>
```

### advanced.tf

```
prod_job_count = 5

prod_healthcheck_ids = {
  "payment_processing"  = "hc_prod001"
  "fraud_detection"     = "hc_prod002"
  "db_backup_postgres"  = "hc_prod003"
  "db_backup_mysql"     = "hc_prod004"
  "log_rotation"        = "hc_prod005"
}

prod_ping_urls = <sensitive>
staging_ping_urls = <sensitive>
all_healthchecks = <sensitive>
```

## Troubleshooting

### Error: API Key Not Set

```
Error: Missing required argument
  on provider.tf line X:
  The argument "api_key" is required
```

**Solution:**
```bash
export HYPERPING_API_KEY="your-api-key-here"
```

### Error: Invalid Cron Format

```
Error: Invalid value for variable
  All cron schedules must be in standard 5-field format
```

**Solution:** Verify cron expression has exactly 5 fields:
```
✓ "0 2 * * *"     (valid)
✗ "0 2 * *"       (invalid - missing weekday)
✗ "0 0 2 * * *"   (invalid - 6 fields)
```

Use [crontab.guru](https://crontab.guru) to validate expressions.

### Error: Grace Period Out of Range

```
Error: Invalid value for variable
  Grace period must be between 1 and 1440 minutes
```

**Solution:** Adjust grace period:
```hcl
grace = 30  # Valid: 1-1440 minutes
```

### Warning: Sensitive Output

If you see `<sensitive>` when running `terraform output`:

```bash
# Use -raw for single values
terraform output -raw backup_ping_url

# Use -json for maps/objects
terraform output -json ping_urls | jq

# Remove sensitive flag temporarily (not recommended)
# Edit outputs.tf and remove 'sensitive = true'
```

## CI/CD Integration

### GitHub Actions Example

```yaml
name: Test Cron Healthcheck Module

on: [pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Setup Terraform
        uses: hashicorp/setup-terraform@v2

      - name: Terraform Init
        run: |
          cd examples/modules/cron-healthcheck/tests
          terraform init

      - name: Terraform Validate
        run: |
          cd examples/modules/cron-healthcheck/tests
          terraform validate

      - name: Terraform Plan
        run: |
          cd examples/modules/cron-healthcheck/tests
          terraform plan
        env:
          HYPERPING_API_KEY: ${{ secrets.HYPERPING_API_KEY }}
```

## Best Practices for Testing

1. **Use test prefixes**: Add "TEST-" prefix to avoid confusion with production
2. **Clean up**: Always run `terraform destroy` after testing
3. **Test validation**: Verify both valid and invalid inputs
4. **Integration tests**: Manually ping URLs to verify end-to-end flow
5. **Document outputs**: Keep track of created healthcheck IDs
6. **Separate environments**: Use different API keys for test vs production

## Additional Resources

- [Module Documentation](../README.md)
- [Hyperping Healthcheck Resource](https://registry.terraform.io/providers/develeap/hyperping/latest/docs/resources/healthcheck)
- [Terraform Testing Best Practices](https://www.terraform.io/docs/language/modules/testing-experiment.html)
