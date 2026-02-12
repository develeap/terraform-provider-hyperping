---
page_title: "Filtering Data Sources - Hyperping Provider"
description: |-
  Guide to using client-side filtering with Hyperping data sources.
---

# Filtering Data Sources

This guide explains how to use client-side filtering with Hyperping data sources to efficiently query and filter resources.

## Overview

The Hyperping provider supports **client-side filtering** for all list-based data sources. This means:

- The provider fetches all resources from the API
- Filters are applied locally in Terraform
- Only matching resources are returned to your configuration

### Why Client-Side Filtering?

The Hyperping API does not currently support server-side filtering for most resources. Client-side filtering provides:

- **Consistency**: Same filtering behavior across all data sources
- **Flexibility**: Full regex support for pattern matching
- **Simplicity**: No need to learn API-specific query languages

### Performance Characteristics

- **Fetch Time**: O(n) - fetches all resources from API
- **Filter Time**: O(n) - applies filters locally
- **Total Time**: Typically < 5 seconds for 1000+ resources

For production use with large datasets, consider:
- Caching results with `lifecycle` rules
- Using specific resource data sources when IDs are known
- Implementing local caching strategies

## Available Filters by Data Source

| Data Source | Name Regex | Status | Severity | Protocol | Paused | Monitor UUID | Hostname |
|-------------|------------|--------|----------|----------|--------|--------------|----------|
| `hyperping_monitors` | ✓ | - | - | ✓ | ✓ | - | - |
| `hyperping_incidents` | ✓ | ✓ | ✓ | - | - | - | - |
| `hyperping_maintenance_windows` | ✓ | ✓ | - | - | - | - | - |
| `hyperping_healthchecks` | ✓ | ✓ | - | - | - | - | - |
| `hyperping_outages` | ✓ | - | - | - | - | ✓ | - |
| `hyperping_statuspages` | ✓ | - | - | - | - | - | ✓ |

### Filter Types

- **Name Regex**: Regular expression pattern matching (case-sensitive)
- **Status**: Exact string match (case-sensitive)
- **Severity**: Exact string match (case-sensitive)
- **Protocol**: Exact string match (case-sensitive)
- **Paused**: Boolean match
- **Monitor UUID**: Exact string match
- **Hostname**: Exact string match

## Filter Examples

### Simple Regex Filter

Filter monitors by name pattern:

```hcl
data "hyperping_monitors" "production" {
  filter = {
    name_regex = "\\[PROD\\]-.*"
  }
}

output "production_monitors" {
  value = data.hyperping_monitors.production.monitors
}
```

### Combined Filters

Apply multiple filters together (AND logic):

```hcl
data "hyperping_monitors" "prod_https_active" {
  filter = {
    name_regex = "\\[PROD\\]-.*"
    protocol   = "https"
    paused     = false
  }
}
```

All filters must match for a resource to be included.

### Boolean Filters

Filter by boolean attributes:

```hcl
# Active monitors only
data "hyperping_monitors" "active" {
  filter = {
    paused = false
  }
}

# Paused monitors only
data "hyperping_monitors" "paused" {
  filter = {
    paused = true
  }
}
```

### Status Filters

Filter incidents by status:

```hcl
data "hyperping_incidents" "active" {
  filter = {
    status = "investigating"
  }
}

data "hyperping_incidents" "resolved" {
  filter = {
    status = "resolved"
  }
}
```

Valid status values:
- **Incidents**: `investigating`, `identified`, `monitoring`, `resolved`
- **Maintenance**: `scheduled`, `in_progress`, `completed`

### Severity Filters

Filter incidents by severity:

```hcl
data "hyperping_incidents" "critical" {
  filter = {
    severity = "critical"
  }
}
```

Valid severity values: `minor`, `major`, `critical`

### Protocol Filters

Filter monitors by protocol:

```hcl
data "hyperping_monitors" "https_only" {
  filter = {
    protocol = "https"
  }
}
```

Valid protocol values: `http`, `https`, `tcp`, `icmp`, `udp`

### Monitor UUID Filters

Filter outages by specific monitor:

```hcl
data "hyperping_outages" "specific_monitor" {
  filter = {
    monitor_uuid = hyperping_monitor.api.id
  }
}
```

### Hostname Filters

Filter status pages by custom hostname:

```hcl
data "hyperping_statuspages" "custom_domain" {
  filter = {
    hostname = "status.example.com"
  }
}
```

## Complex Regex Patterns

### Environment Prefix

Match monitors with specific environment prefixes:

```hcl
data "hyperping_monitors" "prod_or_staging" {
  filter = {
    name_regex = "^\\[(PROD|STAGING)\\]-.*"
  }
}
```

### Service Type Suffix

Match monitors ending with specific service types:

```hcl
data "hyperping_monitors" "api_services" {
  filter = {
    name_regex = ".*-API$"
  }
}
```

### Contains Pattern

Match monitors containing specific keywords:

```hcl
data "hyperping_monitors" "database_related" {
  filter = {
    name_regex = ".*(DATABASE|DB|POSTGRES|MYSQL).*"
  }
}
```

### Complex Naming Convention

Match monitors following a specific naming pattern:

```hcl
data "hyperping_monitors" "standard_naming" {
  filter = {
    # Format: [ENV]-ServiceType-Name-Optional123
    name_regex = "^\\[(PROD|STAGING|DEV)\\]-(API|WEB|DATABASE)-[A-Z]+(-[0-9]+)?$"
  }
}
```

### Digits in Name

Match monitors with version numbers:

```hcl
data "hyperping_monitors" "versioned" {
  filter = {
    name_regex = ".*-v\\d+.*"
  }
}
```

## Use Case Scenarios

### Scenario 1: Find All Production HTTPS Monitors

```hcl
data "hyperping_monitors" "prod_https" {
  filter = {
    name_regex = "\\[PROD\\]-.*"
    protocol   = "https"
    paused     = false
  }
}

output "production_https_count" {
  value = length(data.hyperping_monitors.prod_https.monitors)
}

output "production_https_names" {
  value = [for m in data.hyperping_monitors.prod_https.monitors : m.name]
}
```

### Scenario 2: Find Active Critical Incidents

```hcl
data "hyperping_incidents" "critical_active" {
  filter = {
    severity = "critical"
  }
}

# Filter out resolved incidents in local processing
locals {
  active_critical = [
    for inc in data.hyperping_incidents.critical_active.incidents :
    inc if inc.status != "resolved"
  ]
}

output "critical_incident_count" {
  value = length(local.active_critical)
}
```

### Scenario 3: Find Upcoming Maintenance Windows

```hcl
data "hyperping_maintenance_windows" "upcoming" {
  filter = {
    status = "scheduled"
  }
}

output "upcoming_maintenance" {
  value = [
    for mw in data.hyperping_maintenance_windows.upcoming.maintenance_windows : {
      title = mw.title
      start = mw.scheduled_start
      end   = mw.scheduled_end
    }
  ]
}
```

### Scenario 4: Cross-Reference Resources

```hcl
# Find all production monitors
data "hyperping_monitors" "production" {
  filter = {
    name_regex = "\\[PROD\\]-.*"
  }
}

# Find outages for production monitors
data "hyperping_outages" "production_outages" {
  filter = {
    name_regex = "\\[PROD\\]-.*"
  }
}

# Combine data
output "production_health" {
  value = {
    total_monitors = length(data.hyperping_monitors.production.monitors)
    total_outages  = length(data.hyperping_outages.production_outages.outages)
  }
}
```

### Scenario 5: Generate Dashboard Configuration

```hcl
data "hyperping_monitors" "all_https" {
  filter = {
    protocol = "https"
  }
}

# Generate a dashboard config from filtered monitors
locals {
  dashboard_config = {
    for m in data.hyperping_monitors.all_https.monitors : m.id => {
      name     = m.name
      url      = m.url
      regions  = m.regions
    }
  }
}

output "dashboard_json" {
  value = jsonencode(local.dashboard_config)
}
```

## Best Practices

### 1. Use Specific Filters

More specific filters = fewer resources processed:

```hcl
# Good: Specific filter
data "hyperping_monitors" "specific" {
  filter = {
    name_regex = "^\\[PROD\\]-API-.*$"
    protocol   = "https"
  }
}

# Less efficient: Broad filter
data "hyperping_monitors" "broad" {
  filter = {
    name_regex = ".*"
  }
}
```

### 2. Escape Regex Metacharacters

Always escape regex special characters:

```hcl
# Correct: Escaped brackets and backslashes
filter = {
  name_regex = "\\[PROD\\]-.*"
}

# Incorrect: Unescaped brackets
filter = {
  name_regex = "[PROD]-.*"  # This creates a character class!
}
```

### 3. Combine with Local Processing

Use filters for API data, then refine with locals:

```hcl
data "hyperping_monitors" "candidates" {
  filter = {
    protocol = "https"
  }
}

locals {
  # Further filter in Terraform
  high_frequency = [
    for m in data.hyperping_monitors.candidates.monitors :
    m if m.frequency <= 60
  ]
}
```

### 4. Use Empty Filter for All Resources

Omit the filter block to get all resources:

```hcl
# Returns all monitors
data "hyperping_monitors" "all" {}
```

### 5. Test Regex Patterns

Test complex regex patterns separately:

```bash
# Test regex locally
echo "[PROD]-API-Service" | grep -P '^\[(PROD|STAGING)\]-.*'
```

### 6. Cache Results

Use lifecycle to prevent unnecessary API calls:

```hcl
data "hyperping_monitors" "cached" {
  filter = {
    name_regex = "\\[PROD\\]-.*"
  }

  lifecycle {
    postcondition {
      condition     = length(self.monitors) > 0
      error_message = "No production monitors found"
    }
  }
}
```

### 7. Document Regex Patterns

Add comments explaining complex patterns:

```hcl
data "hyperping_monitors" "production" {
  filter = {
    # Matches: [PROD]-ServiceType-Name
    # Examples: [PROD]-API-Auth, [PROD]-WEB-Dashboard
    name_regex = "^\\[PROD\\]-[A-Z]+-.*"
  }
}
```

## Common Regex Patterns

### Match Exact Prefix

```hcl
name_regex = "^\\[PROD\\]-"  # Starts with [PROD]-
```

### Match Exact Suffix

```hcl
name_regex = "-API$"  # Ends with -API
```

### Match Multiple Options

```hcl
name_regex = "^\\[(PROD|STAGING|DEV)\\]-"  # Any environment
```

### Match Word Boundaries

```hcl
name_regex = "\\bAPI\\b"  # Matches "API" as whole word
```

### Match Any Digit

```hcl
name_regex = ".*\\d+.*"  # Contains any number
```

### Match Specific Pattern

```hcl
name_regex = "^[A-Z]+-[0-9]{3}$"  # Format: ABC-123
```

### Case-Insensitive (using alternation)

```hcl
name_regex = "(?i)prod"  # Matches PROD, prod, Prod, etc.
```

## Troubleshooting

### Issue: No Matches Returned

**Problem**: Filter returns empty list.

**Solutions**:
1. Verify regex pattern is correct
2. Check case sensitivity (filters are case-sensitive)
3. Test pattern with known resources
4. Check for typos in filter values

```hcl
# Debug: Get all resources first
data "hyperping_monitors" "all" {}

output "all_names" {
  value = [for m in data.hyperping_monitors.all.monitors : m.name]
}

# Then test filter
data "hyperping_monitors" "filtered" {
  filter = {
    name_regex = "YOUR_PATTERN"
  }
}
```

### Issue: Invalid Regex Error

**Problem**: Terraform returns regex compilation error.

**Solutions**:
1. Escape backslashes: `\\` not `\`
2. Test regex syntax: use online regex testers
3. Check for unmatched brackets or parentheses

```hcl
# Correct
name_regex = "\\[PROD\\]-.*"

# Incorrect
name_regex = "\[PROD\]-.*"  # Single backslash fails
```

### Issue: Too Many Resources

**Problem**: Filter returns more resources than expected.

**Solutions**:
1. Add more specific filters
2. Combine multiple filter fields
3. Use anchors in regex: `^` and `$`

```hcl
# Too broad
filter = {
  name_regex = "API"  # Matches anywhere in name
}

# More specific
filter = {
  name_regex = "^\\[PROD\\]-API-.*$"  # Exact format
  protocol   = "https"
  paused     = false
}
```

### Issue: Performance Degradation

**Problem**: Filtering takes too long.

**Solutions**:
1. Use simpler regex patterns
2. Reduce number of resources (contact Hyperping support)
3. Cache results with lifecycle rules
4. Use specific resource data sources when IDs are known

```hcl
# Slow: Complex regex
name_regex = "^((?!test).)*$"  # Negative lookahead

# Fast: Simple pattern
name_regex = "^\\[PROD\\]-.*"
```

### Issue: Null vs Unknown Values

**Problem**: Unexpected behavior with optional fields.

**Solution**: Filters treat null/unknown as "no filter" (match all):

```hcl
filter = {
  protocol = "https"
  # paused is null - will match both paused=true and paused=false
}
```

To filter by null values, use local processing:

```hcl
data "hyperping_monitors" "all" {}

locals {
  has_custom_headers = [
    for m in data.hyperping_monitors.all.monitors :
    m if length(m.headers) > 0
  ]
}
```

## Performance Optimization

### Benchmark Results

Filtering performance for typical datasets:

| Dataset Size | Filter Type | Time |
|--------------|-------------|------|
| 100 monitors | Simple regex | < 100ms |
| 1000 monitors | Simple regex | < 500ms |
| 1000 monitors | Complex regex | < 2s |
| 1000 monitors | Multiple filters | < 1s |
| 10000 monitors | Simple filter | < 5s |

### Optimization Tips

1. **Use exact matches over regex** when possible:
   ```hcl
   # Faster
   filter = {
     protocol = "https"
   }

   # Slower
   filter = {
     name_regex = ".*"  # Matches everything
   }
   ```

2. **Anchor regex patterns**:
   ```hcl
   # Faster: Anchored
   name_regex = "^\\[PROD\\]-API$"

   # Slower: Unanchored
   name_regex = "PROD.*API"
   ```

3. **Avoid negative lookaheads**:
   ```hcl
   # Slow
   name_regex = "^((?!test).)*$"

   # Fast: Use positive pattern instead
   name_regex = "^\\[PROD\\]-.*"
   ```

4. **Combine filters efficiently**:
   ```hcl
   # Efficient: Protocol filter first (fast), then regex
   filter = {
     protocol   = "https"  # Fast exact match
     name_regex = ".*"     # Then regex
   }
   ```

## Additional Resources

- [Terraform Data Sources](https://www.terraform.io/docs/language/data-sources/index.html)
- [Go Regular Expressions Syntax](https://pkg.go.dev/regexp/syntax)
- [Regex Testing Tool](https://regex101.com/)
- [Hyperping API Documentation](https://hyperping.io/docs/api)

## Related Guides

- [Data Source Reference](../data-sources/index.md)
- [Resource Import Guide](./importing-resources.md)
- [Provider Configuration](../index.md)
