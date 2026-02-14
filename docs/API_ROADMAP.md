# API Feature Roadmap

**Version:** v1.1.0 Planning
**Last Updated:** 2026-02-13
**Status:** Draft

## Overview

This document outlines planned features for future releases of terraform-provider-hyperping based on the [API Completeness Audit](./API_COMPLETENESS_AUDIT.md).

Current coverage: **95%** of core API endpoints
Target coverage: **100%** including advanced features

---

## Release Planning

### v1.0.8 (Patch Release) - Documentation & Bug Fixes

**Target Date:** 2026-02-20
**Focus:** Documentation accuracy and minor improvements

#### Tasks

1. **Update CLAUDE.md** (P1)
   - Remove outdated fields (`timeout`, `notify_subscribers`, etc.)
   - Add documentation for healthchecks
   - Add documentation for outages
   - Add documentation for status pages
   - Fix field name discrepancies

2. **Add Missing Monitor Field** (P1)
   - Add `project_uuid` field to `hyperping_monitor` resource
   - Update schema with optional `project` attribute
   - Test project assignment
   - Update examples

**Estimated Effort:** 2-3 days
**Breaking Changes:** None

---

### v1.1.0 (Minor Release) - Escalation Policies

**Target Date:** 2026-03-15
**Focus:** First-class escalation policy management

#### 1. Escalation Policy Resource (P1)

**Status:** ❌ NOT IMPLEMENTED
**Complexity:** Medium
**Estimated Effort:** 1 week

**API Endpoints to Implement:**
```
GET    /v2/escalation-policies
GET    /v2/escalation-policies/{uuid}
POST   /v2/escalation-policies
PUT    /v2/escalation-policies/{uuid}
DELETE /v2/escalation-policies/{uuid}
```

**Proposed Terraform Resource:**
```hcl
resource "hyperping_escalation_policy" "oncall" {
  name        = "On-Call Team Escalation"
  description = "Escalate to on-call engineer after 5 minutes"

  steps {
    wait_minutes = 0
    notify       = ["integration_uuid_slack"]
  }

  steps {
    wait_minutes = 5
    notify       = ["integration_uuid_pagerduty"]
  }

  steps {
    wait_minutes = 10
    notify       = ["integration_uuid_phone"]
  }
}

resource "hyperping_monitor" "api" {
  name               = "API Health"
  url                = "https://api.example.com"
  escalation_policy  = hyperping_escalation_policy.oncall.id
}
```

**Implementation Checklist:**
- [ ] Research actual API request/response format
- [ ] Create `models_escalation_policy.go`
- [ ] Implement client methods in `escalation_policies.go`
- [ ] Create `escalation_policy_resource.go`
- [ ] Create `escalation_policies_data_source.go`
- [ ] Create `escalation_policy_data_source.go`
- [ ] Write contract tests
- [ ] Write acceptance tests
- [ ] Generate documentation
- [ ] Add examples

**Benefits:**
- ✅ Manage on-call schedules as code
- ✅ Version control escalation procedures
- ✅ Link policies to multiple monitors
- ✅ Remove manual dashboard configuration

**Risks:**
- ⚠️ API may not exist or be undocumented
- ⚠️ Complex nested structure (steps)
- ⚠️ Integration UUIDs must reference notification channels

---

#### 2. Notification Channels/Integrations Resource (P1)

**Status:** ❌ NOT IMPLEMENTED
**Complexity:** Medium-High
**Estimated Effort:** 1-2 weeks

**Investigation Required:**
- Determine if API endpoints exist
- Document endpoint paths
- Understand authentication models (Slack OAuth, PagerDuty API keys, etc.)

**Potential Terraform Resource:**
```hcl
resource "hyperping_integration_slack" "alerts" {
  name         = "Team Slack Channel"
  webhook_url  = var.slack_webhook_url
  channel      = "#alerts"
}

resource "hyperping_integration_pagerduty" "oncall" {
  name            = "PagerDuty On-Call"
  integration_key = var.pagerduty_key
  service_id      = "PXXXXXX"
}

resource "hyperping_integration_webhook" "custom" {
  name        = "Custom Webhook"
  url         = "https://example.com/webhook"
  http_method = "POST"
  headers = {
    "Authorization" = "Bearer ${var.webhook_token}"
  }
}
```

**Implementation Checklist:**
- [ ] **API Discovery** - Determine if endpoints exist
- [ ] Document API request/response formats
- [ ] Implement for Slack
- [ ] Implement for PagerDuty
- [ ] Implement for Webhook
- [ ] Implement for Email
- [ ] Implement for SMS (Twilio)
- [ ] Write tests for each integration type
- [ ] Generate documentation
- [ ] Add examples

**Benefits:**
- ✅ Manage alert destinations as code
- ✅ Automate Slack/PagerDuty setup
- ✅ Secure credential management
- ✅ Audit integration changes

**Risks:**
- 🔴 HIGH: API may not exist publicly
- ⚠️ Complex OAuth flows for Slack
- ⚠️ Different schemas per integration type

**Mitigation:**
- Start with simple webhook integration
- Use conditional implementation based on API availability
- Document manual workarounds if API unavailable

---

### v1.2.0 (Minor Release) - Enhanced Features

**Target Date:** 2026-Q2
**Focus:** Quality of life improvements

#### 3. Read-Only Monitor Fields (P2)

**Status:** ⚠️ PARTIAL
**Complexity:** Low
**Estimated Effort:** 2-3 days

**Missing Fields:**
- `status` (string) - "up" or "down"
- `ssl_expiration` (int) - Days until SSL cert expires

**Implementation:**
```hcl
resource "hyperping_monitor" "api" {
  name = "API Health"
  url  = "https://api.example.com"

  # Computed fields
  # status         = "up"           # Read-only
  # ssl_expiration = 45             # Read-only
}

# Use in outputs
output "api_status" {
  value = hyperping_monitor.api.status
}

output "ssl_warning" {
  value = hyperping_monitor.api.ssl_expiration < 30 ? "SSL cert expiring soon!" : "OK"
}
```

**Implementation Checklist:**
- [ ] Add fields to monitor resource model
- [ ] Mark as `Computed: true` in schema
- [ ] Update `mapMonitorToModel()` function
- [ ] Write tests
- [ ] Update documentation
- [ ] Add example use case

**Benefits:**
- ✅ Access real-time monitor status in Terraform
- ✅ Build alerts based on SSL expiration
- ✅ Export status to other systems

---

#### 4. Project/Workspace Support (P2)

**Status:** ⚠️ PARTIAL (field exists, not documented)
**Complexity:** Low
**Estimated Effort:** 2-3 days

**Current State:**
- API model has `projectUuid` field
- Not exposed in Terraform resource

**Proposed Implementation:**
```hcl
# Option 1: Direct UUID assignment
resource "hyperping_monitor" "api" {
  name       = "API Health"
  url        = "https://api.example.com"
  project_id = "proj_abc123"  # Manual UUID
}

# Option 2: With project resource (if API supports)
resource "hyperping_project" "production" {
  name        = "Production Services"
  description = "All production monitors"
}

resource "hyperping_monitor" "api" {
  name       = "API Health"
  url        = "https://api.example.com"
  project_id = hyperping_project.production.id
}
```

**Implementation Checklist:**
- [ ] Add `project_id` to monitor schema (Optional)
- [ ] Update create/update requests
- [ ] Test with real Hyperping projects
- [ ] **Research project API** (if exists)
- [ ] Implement project resource (if API available)
- [ ] Write tests
- [ ] Update documentation

**Benefits:**
- ✅ Organize monitors by team/project
- ✅ Multi-tenancy support
- ✅ Better resource organization

---

#### 5. Bulk Operations (P2)

**Status:** ❌ NOT IMPLEMENTED
**Complexity:** Medium
**Estimated Effort:** 1 week

**Use Cases:**
1. Pause all monitors during maintenance
2. Resume all monitors after maintenance
3. Update tags on multiple monitors
4. Bulk delete test monitors

**Proposed API:**
```hcl
# Option 1: Utility resource
resource "hyperping_bulk_action" "pause_all" {
  action  = "pause"
  filters = {
    name_contains = "staging"
  }
}

# Option 2: Data source + for_each
data "hyperping_monitors" "staging" {
  filter {
    name_regex = "^\\[STAGING\\]"
  }
}

resource "hyperping_monitor" "pause_staging" {
  for_each = { for m in data.hyperping_monitors.staging.monitors : m.id => m }

  id     = each.value.id
  paused = true
}
```

**Implementation Options:**
1. **Custom resource** - `hyperping_bulk_action`
2. **CLI tool** - `tfhyperping bulk-pause --filter "staging"`
3. **Provider-level operation** - Not standard in Terraform

**Recommendation:** Use native Terraform `for_each` with data sources

**Implementation Checklist:**
- [ ] Document best practices for bulk operations
- [ ] Add examples for common scenarios
- [ ] Consider CLI tool if demand is high

---

### v1.3.0 (Minor Release) - Advanced Features

**Target Date:** 2026-Q3
**Focus:** Enterprise and advanced use cases

#### 6. Team Management (P2)

**Status:** ❌ NOT IMPLEMENTED
**Complexity:** High
**Estimated Effort:** 2-3 weeks

**Prerequisite:** API endpoints must exist

**Proposed Resources:**
```hcl
resource "hyperping_team" "sre" {
  name        = "SRE Team"
  description = "Site Reliability Engineering"
}

resource "hyperping_team_member" "alice" {
  team_id = hyperping_team.sre.id
  email   = "alice@example.com"
  role    = "admin"
}

resource "hyperping_team_member" "bob" {
  team_id = hyperping_team.sre.id
  email   = "bob@example.com"
  role    = "member"
}
```

**Implementation Checklist:**
- [ ] **Research Hyperping Teams API** (if exists)
- [ ] Document endpoint paths
- [ ] Implement team resource
- [ ] Implement team member resource
- [ ] Handle role-based permissions
- [ ] Write tests
- [ ] Generate documentation

**Benefits:**
- ✅ Automated team provisioning
- ✅ Audit team membership changes
- ✅ Sync with HR systems

**Risks:**
- 🔴 API may not exist
- 🔴 May require enterprise plan
- ⚠️ Complex permission models

---

#### 7. Advanced Reporting (P2)

**Status:** ⚠️ PARTIAL (single monitor only)
**Complexity:** Medium
**Estimated Effort:** 1 week

**Current State:**
- `hyperping_monitor_report` - Single monitor reports only

**Proposed Enhancements:**
```hcl
# Aggregate report across monitors
data "hyperping_aggregate_report" "production" {
  from = "2026-02-01T00:00:00Z"
  to   = "2026-02-28T23:59:59Z"

  monitors = [
    hyperping_monitor.api.id,
    hyperping_monitor.web.id,
    hyperping_monitor.db.id,
  ]

  # Computed outputs
  # overall_uptime   = 99.95
  # total_downtime   = 3600  # seconds
  # worst_monitor    = hyperping_monitor.db.id
}

# Export to file
data "hyperping_report_export" "monthly" {
  from   = "2026-02-01T00:00:00Z"
  to     = "2026-02-28T23:59:59Z"
  format = "csv"  # or "pdf", "json"

  # Downloads to local file
  output_path = "${path.module}/reports/february-2026.csv"
}
```

**Implementation Checklist:**
- [ ] Research aggregate reporting API
- [ ] Implement aggregate data source
- [ ] Research export API
- [ ] Implement export functionality
- [ ] Support CSV, PDF, JSON formats
- [ ] Write tests
- [ ] Generate documentation

**Benefits:**
- ✅ SLA reporting across services
- ✅ Automated report generation
- ✅ Integration with dashboards

---

#### 8. Recurring Maintenance Windows (P3)

**Status:** ❌ NOT IMPLEMENTED
**Complexity:** Medium
**Estimated Effort:** 1 week

**Current State:**
- One-time maintenance windows only

**Proposed Enhancement:**
```hcl
resource "hyperping_maintenance" "weekly_db_backup" {
  name        = "Weekly Database Backup"
  description = "Automated weekly backup maintenance"

  # Recurring schedule
  recurrence {
    frequency = "weekly"
    day       = "sunday"
    time      = "02:00:00"
    duration  = "2h"
    timezone  = "America/New_York"
  }

  monitors = [hyperping_monitor.db.id]
}

resource "hyperping_maintenance" "monthly_patching" {
  name = "Monthly Security Patching"

  recurrence {
    frequency     = "monthly"
    day_of_month  = 1  # First day of month
    time          = "03:00:00"
    duration      = "4h"
  }

  monitors = [
    hyperping_monitor.api.id,
    hyperping_monitor.web.id,
  ]
}
```

**Implementation Checklist:**
- [ ] **Research if API supports recurrence**
- [ ] If not, implement client-side scheduling
- [ ] Add recurrence schema
- [ ] Implement recurrence logic
- [ ] Handle timezone conversion
- [ ] Write tests (including edge cases)
- [ ] Generate documentation

**Benefits:**
- ✅ Automate recurring maintenance
- ✅ Reduce manual scheduling
- ✅ Consistent maintenance windows

**Risks:**
- 🔴 API may not support recurrence
- ⚠️ Complex timezone handling
- ⚠️ May require external scheduler if API doesn't support

---

## Feature Priority Matrix

| Feature | Priority | Complexity | User Impact | API Availability | Target Release |
|---------|----------|------------|-------------|------------------|----------------|
| Project field | P1 | Low | Medium | ✅ Confirmed | v1.0.8 |
| CLAUDE.md update | P1 | Low | High | N/A | v1.0.8 |
| Escalation policies | P1 | Medium | High | ⚠️ Unknown | v1.1.0 |
| Notification channels | P1 | High | High | ⚠️ Unknown | v1.1.0 |
| Read-only monitor fields | P2 | Low | Low | ✅ Confirmed | v1.2.0 |
| Bulk operations | P2 | Medium | Medium | N/A | v1.2.0 |
| Team management | P2 | High | Medium | ❌ Unknown | v1.3.0 |
| Advanced reporting | P2 | Medium | Medium | ⚠️ Unknown | v1.3.0 |
| Recurring maintenance | P3 | Medium | Low | ❌ Unlikely | v2.0.0 |

---

## API Research Needed

Before implementing the following features, we need to research if Hyperping provides API endpoints:

### High Priority Research

1. **Escalation Policies**
   - Endpoint paths
   - Request/response schemas
   - Step configuration format
   - Integration linking

2. **Notification Channels/Integrations**
   - Supported integration types
   - Authentication methods (OAuth, API keys)
   - Webhook configuration
   - Rate limits

### Medium Priority Research

3. **Projects API**
   - CRUD operations
   - Member assignment
   - Permission models

4. **Teams API**
   - User management endpoints
   - Role definitions
   - Invitation flow

5. **Aggregate Reporting**
   - Multi-monitor reports
   - Export formats
   - Calculation methods

### How to Research

**Methods:**
1. **Hyperping Documentation** - Check official docs
2. **Browser DevTools** - Inspect dashboard network calls
3. **Support Contact** - Ask Hyperping team directly
4. **Community** - Check GitHub issues, forums

**Template for Recording Findings:**
```markdown
## [Feature Name] API Research

**Date:** YYYY-MM-DD
**Researcher:** Name

### Endpoints Found
- `GET /vX/resource` - List all
- `POST /vX/resource` - Create

### Request Format
```json
{
  "field": "value"
}
```

### Response Format
```json
{
  "id": "uuid",
  "field": "value"
}
```

### Notes
- Any special behaviors
- Rate limits
- Authentication requirements
```

---

## Community Feedback Integration

We welcome feature requests from the community:

**How to Request Features:**
1. Open GitHub issue with label `feature-request`
2. Describe use case and priority
3. Provide example Terraform configuration
4. Vote on existing requests with 👍

**Prioritization Criteria:**
- Number of community votes
- Alignment with roadmap
- Implementation complexity
- API availability
- Maintainability

---

## Breaking Changes Policy

We follow semantic versioning (semver):

- **Patch (v1.0.x):** Bug fixes, documentation, no breaking changes
- **Minor (v1.x.0):** New features, backward compatible
- **Major (v2.0.0):** Breaking changes allowed

**Breaking changes will only occur in major releases.**

**Deprecation Process:**
1. Mark feature as deprecated in v1.x
2. Provide migration guide
3. Keep deprecated feature for 6+ months
4. Remove in next major version

---

## Implementation Guidelines

For contributors implementing roadmap features:

### Before Starting

1. ✅ Check if API endpoint exists
2. ✅ Read this roadmap document
3. ✅ Create GitHub issue linking to roadmap item
4. ✅ Discuss approach in issue comments

### During Implementation

1. ✅ Follow existing code patterns
2. ✅ Write tests first (TDD)
3. ✅ Document API quirks
4. ✅ Add examples to `examples/`
5. ✅ Update CHANGELOG.md

### Before PR

1. ✅ Run full test suite
2. ✅ Run golangci-lint
3. ✅ Generate documentation (`tfplugindocs generate`)
4. ✅ Test with real Hyperping account
5. ✅ Update this roadmap (mark as completed)

---

## Success Metrics

We track the following metrics to measure roadmap success:

| Metric | Current | Target |
|--------|---------|--------|
| API Coverage | 95% | 100% |
| Test Coverage | 50.8% | 60% |
| Community PRs | 0 | 5+ per quarter |
| GitHub Stars | - | 100+ |
| Terraform Registry Downloads | - | 1000+ per month |
| Open Issues (bugs) | - | < 5 |
| Documentation Completeness | 90% | 100% |

---

## Long-Term Vision (v2.0+)

### Potential Major Features

1. **Terraform Cloud Integration**
   - Remote state management
   - Sentinel policies for monitoring standards
   - Cost estimation for Hyperping plans

2. **Terraform CDK Support**
   - Generate TypeScript/Python bindings
   - CDK examples and patterns

3. **Provider Caching**
   - Reduce API calls during planning
   - Improve terraform plan performance

4. **Custom Resource Importers**
   - Automated import from Better Stack
   - Import from UptimeRobot
   - Import from Pingdom
   - Import from existing Hyperping account

5. **Monitoring as Code Patterns**
   - Reusable module library
   - Best practices documentation
   - Reference architectures

---

## Questions or Feedback?

- **GitHub Issues:** https://github.com/develeap/terraform-provider-hyperping/issues
- **Discussions:** https://github.com/develeap/terraform-provider-hyperping/discussions
- **Email:** maintainers@develeap.com

---

**Roadmap Last Updated:** 2026-02-13
**Next Review:** 2026-03-01
