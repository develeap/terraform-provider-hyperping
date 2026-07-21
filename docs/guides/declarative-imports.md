---
page_title: "Declarative Import Blocks"
subcategory: "Guides"
description: |-
  Use Terraform 1.5+ declarative import blocks and for_each fleet patterns to bring existing Hyperping resources under IaC management without running terraform import commands.
---

# Declarative Import Blocks

Terraform 1.5 introduced `import {}` blocks as a code-first alternative to the `terraform import` CLI command. Terraform 1.7 extended those blocks with `for_each`, making it possible to import an entire fleet of resources in a single plan/apply cycle.

This guide covers:

- Why declarative imports are preferable to `terraform import` shell commands
- Single-resource import blocks for monitors, status pages, and incidents
- Fleet imports using `for_each` for bulk migration scenarios
- A recommended workflow for zero-downtime migration to IaC

## Why Declarative Imports

| Approach | Terraform version | Review-able in VCS | Repeatable | Parallel |
|---|---|---|---|---|
| `terraform import` CLI | any | no | no | no |
| `import {}` block | >= 1.5 | yes | yes | yes |
| `import { for_each }` | >= 1.7 | yes | yes | yes |

Declarative `import {}` blocks live in your `.tf` files alongside the resources they import. They are code-reviewed, version-controlled, and idempotent: re-running `terraform apply` is safe once the resource is already in state because Terraform silently skips the import step.

## Prerequisites

- Terraform 1.7 or later (for `for_each` support)
- Provider configured with a valid `HYPERPING_API_KEY`
- Resource IDs from the Hyperping dashboard (see ID formats below)

## ID Formats

| Resource | ID prefix | Example |
|---|---|---|
| Monitor | `mon_` | `mon_abc123def456` |
| Healthcheck | `hc_` | `hc_xyz789abc123` |
| Status page | `sp_` | `sp_status123` |
| Incident | `inc_` | `inc_incident123` |
| Maintenance | `maint_` | `maint_window123` |
| Status page subscriber | `{sp_id}:{subscriber_id}` | `sp_abc:42` |
| Incident update | `{inc_id}/{upd_id}` | `inc_abc/upd_xyz` |

## Single-Resource Import Block

The simplest pattern: one `import {}` block per resource.

### Monitor

```hcl
import {
  to = hyperping_monitor.api
  id = "mon_abc123def456"
}

resource "hyperping_monitor" "api" {
  name                 = "Production API"
  url                  = "https://api.example.com/health"
  protocol             = "http"
  http_method          = "GET"
  check_frequency      = 300
  expected_status_code = "200"

  regions = ["london", "virginia", "singapore"]
}
```

### Status Page

```hcl
import {
  to = hyperping_statuspage.main
  id = "sp_status123abc"
}

resource "hyperping_statuspage" "main" {
  name             = "Production Status"
  hosted_subdomain = "status"

  settings = {
    name      = "Production Status"
    languages = ["en"]
  }
}
```

### Incident

```hcl
import {
  to = hyperping_incident.api_degradation
  id = "inc_incident123abc"
}

resource "hyperping_incident" "api_degradation" {
  title        = "API Performance Degradation"
  text         = "We are investigating reports of slow API response times."
  type         = "incident"
  status_pages = ["sp_prod111aaa"]
}
```

## Fleet Import with for_each

When migrating dozens or hundreds of resources, use `for_each` to express all imports in a compact local map. Terraform resolves every entry in parallel during `terraform apply`.

### Monitor Fleet

```hcl
locals {
  monitor_ids = {
    api_health = "mon_aaabbbccc111"
    web_home   = "mon_dddeeefff222"
    db_primary = "mon_ggghhh333444"
    cdn_assets = "mon_iiijjj555666"
  }
}

import {
  for_each = local.monitor_ids
  to       = hyperping_monitor.fleet[each.key]
  id       = each.value
}

resource "hyperping_monitor" "fleet" {
  for_each = local.monitor_ids

  name     = each.key
  url      = "https://placeholder.example.com"
  protocol = "http"
}
```

After `terraform plan`, the plan output shows the real attribute values from the Hyperping API. Replace the placeholder values in the `resource` block to match, then run `terraform apply`. The next `terraform plan` should show no changes.

### Status Page Fleet

```hcl
locals {
  statuspage_ids = {
    production = "sp_prod111aaa"
    staging    = "sp_stage222bbb"
    internal   = "sp_int333ccc"
  }
}

import {
  for_each = local.statuspage_ids
  to       = hyperping_statuspage.fleet[each.key]
  id       = each.value
}

resource "hyperping_statuspage" "fleet" {
  for_each = local.statuspage_ids

  name             = each.key
  hosted_subdomain = each.key

  settings = {
    name      = each.key
    languages = ["en"]
  }
}
```

### Incident Fleet

Use this pattern when seeding a new Terraform workspace from an existing account that already has open or historical incidents.

```hcl
locals {
  incident_ids = {
    api_degradation  = "inc_aaa111bbb222"
    database_latency = "inc_ccc333ddd444"
    cdn_outage       = "inc_eee555fff666"
  }
}

import {
  for_each = local.incident_ids
  to       = hyperping_incident.fleet[each.key]
  id       = each.value
}

resource "hyperping_incident" "fleet" {
  for_each = local.incident_ids

  title        = each.key
  text         = "Imported incident"
  type         = "incident"
  status_pages = []
}
```

## Recommended Workflow

1. **Inventory** your existing resources. The Hyperping dashboard URL contains the resource ID for each monitor (`/monitors/{id}`), status page (`/statuspages/{id}`), and incident (`/incidents/{id}`).

2. **Build the local map** with the IDs gathered in step 1.

3. **Write placeholder `resource` blocks** that satisfy the provider schema. Attributes that do not have defaults must be supplied; use any valid placeholder value for now.

4. **Run `terraform plan`**. Terraform prints the full set of real attribute values it read from the API alongside the diff against your placeholder config.

5. **Update the `resource` blocks** to match the values shown in the plan output. Pay particular attention to:
   - `check_frequency` (defaults differ across monitors)
   - `regions` (defaults to all regions except Bahrain)
   - `expected_status_code`
   - `request_headers` ordering

6. **Run `terraform apply`**. Terraform imports each resource and reconciles any remaining differences.

7. **Verify** with `terraform plan`. The output should be `No changes`.

8. **Commit** your `.tf` files. The `import {}` blocks are idempotent and safe to leave in place; Terraform skips them once resources are already in state.

## Generating the ID Map Automatically

The `import-generator` tool bundled in `cmd/import-generator/` fetches all resources from the Hyperping API and emits a ready-to-paste local map:

```bash
cd cmd/import-generator
go run . -format=import -resources=monitors
```

See [docs/IMPORT_GENERATOR_GUIDE.md](../IMPORT_GENERATOR_GUIDE.md) for full usage.

## Relation to terraform import CLI

The `terraform import` command and `import {}` blocks are functionally equivalent for a single resource. The CLI command is useful for one-off ad-hoc imports; `import {}` blocks are preferable for anything that will be code-reviewed, repeated, or shared with a team.

For existing `import.sh` scripts generated by the import-generator or written manually, you can mechanically convert each line:

```bash
# CLI form
terraform import hyperping_monitor.api mon_abc123

# Equivalent declarative block
import {
  to = hyperping_monitor.api
  id = "mon_abc123"
}
```

Both forms can coexist in the same workspace; there is no migration cost to switching gradually.

## Related Guides

- [Importing Resources (CLI)](./importing-resources.md): imperative `terraform import` reference
- [Import Generator Guide](../IMPORT_GENERATOR_GUIDE.md): automated bulk import tooling
- [Best Practices](./best-practices.md): team workflow and state management
