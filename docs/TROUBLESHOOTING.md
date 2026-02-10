# Troubleshooting Guide

Common issues and solutions when using the Hyperping Terraform Provider.

## Authentication Issues

### Error: "401 Unauthorized" or "Invalid API key"

**Symptoms:**
```
Error: Failed to create monitor
  API error: 401 Unauthorized - Invalid or missing API key
```

**Solutions:**
1. Verify your API key starts with `sk_`
2. Check the environment variable is set:
   ```bash
   echo $HYPERPING_API_KEY
   ```
3. Ensure no extra whitespace in the key
4. Generate a new API key from Hyperping dashboard if needed

### Error: "403 Forbidden"

**Symptoms:**
```
Error: API error: 403 Forbidden
```

**Solutions:**
- Verify your API key has the required permissions
- Check if the resource belongs to your account
- Contact Hyperping support if the issue persists

---

## Monitor Issues

### Error: "Invalid URL format"

**Symptoms:**
```
Error: Invalid attribute value
  url: Invalid URL format
```

**Solutions:**
- Ensure URL includes the protocol: `https://example.com` not `example.com`
- For TCP monitors, use format: `tcp://hostname:port`
- Verify no trailing spaces in the URL

### Error: "Invalid check_frequency"

**Symptoms:**
```
Error: Invalid attribute value
  check_frequency must be one of: [10 20 30 60 120 180 300 600 1800 3600]
```

**Solutions:**
- Use only allowed frequency values (in seconds)
- Common values: `60` (1 min), `300` (5 min), `3600` (1 hour)

### Error: "Invalid regions"

**Symptoms:**
```
Error: Invalid region specified
```

**Solutions:**
Valid regions are:
- `london`, `frankfurt`, `singapore`, `sydney`
- `virginia`, `tokyo`, `saopaulo`, `tokyo`, `bahrain`

---

## Import Issues

### Error: "Resource not found" during import

**Symptoms:**
```
Error: Cannot import non-existent remote object
```

**Solutions:**
1. Verify the resource UUID is correct
2. Check the resource exists in Hyperping dashboard
3. Use the correct ID format:
   - Monitors: `mon_xxxxx`
   - Incidents: `inc_xxxxx`
   - Maintenance: `mw_xxxxx`
   - Status pages: `sp_xxxxx`
   - Healthchecks: `hc_xxxxx`

### Error: "Import ID format invalid"

**Symptoms:**
```
Error: Invalid import ID format
```

**Solutions:**
For composite IDs, use the correct format:
- Incident updates: `incident_id/update_id`
- Subscribers: `statuspage_uuid:subscriber_id`

---

## State Issues

### Drift Detection: State doesn't match remote

**Symptoms:**
```
~ resource will be updated in-place
```

**Solutions:**
1. Run `terraform refresh` to sync state
2. If intentional, run `terraform apply` to update
3. For persistent drift, check if changes are made outside Terraform

### Error: "Resource already exists"

**Symptoms:**
```
Error: Resource already exists with this name/URL
```

**Solutions:**
1. Import the existing resource: `terraform import hyperping_monitor.name "mon_xxx"`
2. Or delete the existing resource first
3. Or use a different name/URL

---

## Rate Limiting

### Error: "429 Too Many Requests"

**Symptoms:**
```
Error: API error: 429 Too Many Requests
```

**Solutions:**
1. The provider automatically retries with backoff
2. Reduce parallelism: `terraform apply -parallelism=2`
3. Add delays between operations in CI/CD pipelines
4. Contact Hyperping for rate limit increases if needed

---

## Maintenance Windows

### Error: "Invalid date format"

**Symptoms:**
```
Error: Invalid start_date format
```

**Solutions:**
- Use ISO 8601 format with timezone: `2026-02-01T02:00:00.000Z`
- Ensure `end_date` is after `start_date`
- Dates must be in the future for new maintenance windows

---

## Status Pages

### Error: "Subdomain already taken"

**Symptoms:**
```
Error: Subdomain is already in use
```

**Solutions:**
- Choose a unique subdomain
- Check if you own an existing status page with that subdomain
- Import the existing status page if it's yours

### Error: "Invalid section configuration"

**Symptoms:**
```
Error: Invalid sections configuration
```

**Solutions:**
- Ensure all referenced `monitor_uuid` values exist
- Check section names are provided as maps: `name = { en = "Name" }`
- Verify `services` is a list of valid service objects

---

## Healthchecks

### Error: "cron and period_value are mutually exclusive"

**Symptoms:**
```
Error: Conflicting attributes specified
```

**Solutions:**
- Use either `cron` + `timezone` OR `period_value` + `period_type`
- Don't specify both scheduling methods

### Error: "Invalid cron expression"

**Symptoms:**
```
Error: Invalid cron expression format
```

**Solutions:**
- Use standard 5-field cron format: `minute hour day month weekday`
- Example: `0 2 * * *` (2 AM daily)
- Validate with tools like crontab.guru

---

## Getting Help

If you're still experiencing issues:

1. **Check the logs:** Run with `TF_LOG=DEBUG terraform apply`
2. **GitHub Issues:** https://github.com/develeap/terraform-provider-hyperping/issues
3. **Hyperping Support:** Contact via dashboard
4. **API Documentation:** https://hyperping.io/docs/api
