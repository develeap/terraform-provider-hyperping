# Hyperping API Capability Audit
**For Backend Development Prioritization**

**Document Version:** 1.0
**Date:** February 13, 2026
**Audience:** Leo (Backend Developer)
**Purpose:** Identify API gaps blocking Terraform provider roadmap features

---

## Executive Summary

This audit analyzes the Hyperping API against the Tier 1 and Tier 2 features defined in the [Competitive Strategy Roadmap](./COMPETITIVE_STRATEGY_ROADMAP.md). It categorizes features into:

1. **Implementable NOW** (no API changes needed)
2. **Needs API Enhancement** (endpoints exist but need additional capabilities)
3. **Requires New Endpoints** (completely new API functionality)

### Critical Finding

**60% of Tier 1 and Tier 2 features can be implemented immediately** with the current API. The remaining 40% require backend enhancements, primarily in:
- Server-side filtering (query parameters)
- Resource tagging system
- Alert/notification channel management
- SLA tracking infrastructure

---

## Part 1: Features Implementable NOW (No API Changes Required)

These features can be implemented purely in the Terraform provider without any backend API modifications.

### 1.1 Complete Import Support (Tier 1)
**Status:** ✅ Can implement now
**Effort:** 40 hours (provider-only)
**API Support:** All resources support GET by ID

**Current API Capabilities:**
```
GET /v1/monitors/{uuid}                ✅ Exists
GET /v3/incidents/{uuid}               ✅ Exists
GET /v1/maintenance-windows/{uuid}     ✅ Exists
GET /v2/statuspages/{uuid}             ✅ Exists
GET /v2/healthchecks/{uuid}            ✅ Exists
GET /v2/outages/{uuid}                 ✅ Exists
```

**Implementation:** Add `ImportState` method to remaining resources. No backend work needed.

---

### 1.2 Enhanced Error Messages (Tier 1)
**Status:** ✅ Can implement now
**Effort:** 32 hours (provider-only)
**API Support:** API already returns error details

**Current API Error Format:**
```json
{
  "error": "Validation Error",
  "message": "The request body contains invalid data",
  "details": [
    {"field": "url", "message": "Invalid URL format"}
  ]
}
```

**Implementation:** Enhance client-side error parsing and add user-friendly messages. No backend changes needed.

---

### 1.3 Client-Side Validation Enhancement (Tier 1)
**Status:** ✅ Can implement now
**Effort:** 40 hours (provider-only)
**API Support:** N/A (client-side only)

**Validation Types to Add:**
- Enum validators (already have constants in `models_common.go`)
- URL format validation
- ISO8601 datetime validation
- Range validators (frequency, timeout)
- Cross-field validation

**Implementation:** Pure Terraform provider code. No API involvement.

---

### 1.4 Documentation Overhaul (Tier 1)
**Status:** ✅ Can implement now
**Effort:** 60 hours (documentation-only)
**API Support:** N/A

**Implementation:** Write comprehensive docs, examples, guides. No backend work.

---

### 1.5 Basic Data Source Filtering - Name Only (Tier 1)
**Status:** ✅ Can implement now (client-side filtering)
**Effort:** 24 hours (provider-only)
**API Support:** List endpoints return all resources

**Current Approach:**
```go
// Client-side filtering after fetching all resources
monitors, err := client.ListMonitors(ctx)
filtered := filterByName(monitors, namePattern)
```

**Limitation:** Must fetch ALL resources then filter client-side. Acceptable for small datasets (<1000 resources).

**Note:** Server-side filtering would be better (see Part 2.1) but not required for basic functionality.

---

### 1.6 Terraform Modules Library (Tier 2)
**Status:** ✅ Can implement now
**Effort:** 80 hours (modules-only)
**API Support:** N/A

**Implementation:** Create reusable Terraform modules using existing resources. No API changes.

---

### 1.7 Migration Tools and Guides (Tier 2)
**Status:** ✅ Can implement now
**Effort:** 64 hours (tooling + documentation)
**API Support:** N/A

**Implementation:** Scripts to convert competitor configs to Terraform. No API involvement.

---

**Part 1 Summary:**
- **Features:** 7 out of 11 Tier 1+2 features (64%)
- **Total Effort:** 340 hours
- **Blocker:** None - all can start immediately

---

## Part 2: Features Needing API Enhancements

These features have partial API support but need additional query parameters, fields, or capabilities.

### 2.1 Advanced Data Source Filtering - Server-Side (Tier 1 + Tier 2)

**Feature:** Filter resources by name, status, tags using server-side queries
**Current API:** List endpoints exist but no query parameters
**Gap:** No server-side filtering support

**API Changes Needed:**

#### For Monitors (`GET /v1/monitors`)
```http
# Proposed query parameters
GET /v1/monitors?name=production&status=up&paused=false
GET /v1/monitors?tags=environment:prod,team:platform
GET /v1/monitors?name_regex=.*-api-.*
```

**Request Schema Addition:**
```go
type ListMonitorsRequest struct {
    Name       *string   `url:"name,omitempty"`
    NameRegex  *string   `url:"name_regex,omitempty"`
    Status     *string   `url:"status,omitempty"`      // up, down
    Paused     *bool     `url:"paused,omitempty"`
    Tags       []string  `url:"tags,omitempty"`        // key:value format
    Page       *int      `url:"page,omitempty"`
    Limit      *int      `url:"limit,omitempty"`
}
```

**Response Format:** Same as current (array or wrapped array)

#### For Incidents (`GET /v3/incidents`)
```http
GET /v3/incidents?status=investigating&severity=major
GET /v3/incidents?monitor_uuid=mon_abc123
```

**Request Schema:**
```go
type ListIncidentsRequest struct {
    Status      *string   `url:"status,omitempty"`      // investigating, identified, monitoring, resolved
    Severity    *string   `url:"severity,omitempty"`    // minor, major, critical
    MonitorUUID *string   `url:"monitor_uuid,omitempty"`
    Page        *int      `url:"page,omitempty"`
}
```

#### For Maintenance Windows (`GET /v1/maintenance-windows`)
```http
GET /v1/maintenance-windows?status=scheduled&monitor_uuid=mon_abc123
```

**Request Schema:**
```go
type ListMaintenanceRequest struct {
    Status      *string   `url:"status,omitempty"`      // scheduled, in_progress, completed
    MonitorUUID *string   `url:"monitor_uuid,omitempty"`
    StartAfter  *string   `url:"start_after,omitempty"` // ISO8601
    EndBefore   *string   `url:"end_before,omitempty"`  // ISO8601
    Page        *int      `url:"page,omitempty"`
}
```

**Priority:** Tier 1 (high impact)
**Impact:** Enable efficient filtering for large deployments (100+ monitors)
**Effort Estimate:** 40-60 backend hours
**Benefits:**
- Reduce API response sizes by 90%
- Enable dynamic monitor management
- Support for large-scale deployments (1000+ monitors)

---

### 2.2 Pagination Support for List Endpoints

**Feature:** Paginated responses for all list endpoints
**Current API:** Only `/v2/statuspages` supports pagination
**Gap:** Most list endpoints return all resources (unbounded)

**API Changes Needed:**

#### Standardize Pagination Across All Endpoints
```http
GET /v1/monitors?page=0&limit=50
GET /v3/incidents?page=1&limit=100
GET /v1/maintenance-windows?page=0&limit=25
```

**Response Format (standardize):**
```json
{
  "data": [...],
  "pagination": {
    "page": 0,
    "limit": 50,
    "total": 523,
    "hasNext": true,
    "hasPrev": false
  }
}
```

**Current State:**
- ✅ `/v2/statuspages` - Has pagination
- ❌ `/v1/monitors` - Returns all monitors
- ❌ `/v3/incidents` - Returns all incidents
- ❌ `/v1/maintenance-windows` - Has pagination metadata but inconsistent
- ❌ `/v2/healthchecks` - Returns all
- ❌ `/v2/outages` - Returns all

**Priority:** Tier 1 (scalability requirement)
**Impact:** Essential for accounts with 100+ resources
**Effort Estimate:** 24 backend hours

---

### 2.3 Computed Attributes Enhancement

**Feature:** Return more read-only computed fields in responses
**Current API:** Basic fields returned
**Gap:** Missing useful computed data

**API Changes Needed:**

#### For Monitors
**Current Response:**
```json
{
  "uuid": "mon_abc123",
  "name": "Production API",
  "status": "up",
  "ssl_expiration": 89
}
```

**Enhanced Response:**
```json
{
  "uuid": "mon_abc123",
  "name": "Production API",
  "status": "up",
  "ssl_expiration": 89,
  "created_at": "2024-01-15T10:30:00Z",     // ADD
  "updated_at": "2024-02-13T08:00:00Z",     // ADD
  "last_check_at": "2024-02-13T14:25:00Z",  // ADD
  "last_success_at": "2024-02-13T14:25:00Z", // ADD
  "last_failure_at": "2024-02-10T03:15:00Z", // ADD
  "uptime_percentage_30d": 99.87,           // ADD
  "avg_response_time_24h": 245              // ADD (milliseconds)
}
```

**Priority:** Tier 2 (enhances UX)
**Impact:** Richer Terraform state, better monitoring insights
**Effort Estimate:** 16 backend hours

---

### 2.4 Monitor Report Enhancements

**Feature:** More granular reporting data
**Current API:** `/v2/reporting/monitor-reports` exists but limited
**Gap:** No real-time uptime percentage, limited historical data

**API Changes Needed:**

**Enhance `/v2/reporting/monitor-reports/{uuid}`:**
```json
{
  "monitor_uuid": "mon_abc123",
  "period": {
    "from": "2024-01-13T00:00:00Z",
    "to": "2024-02-13T00:00:00Z"
  },
  "uptime_percentage": 99.87,
  "total_checks": 43200,
  "failed_checks": 56,
  "avg_response_time": 245,
  "p95_response_time": 420,     // ADD
  "p99_response_time": 680,     // ADD
  "incidents": [                 // ADD
    {
      "start": "2024-02-10T03:15:00Z",
      "end": "2024-02-10T03:45:00Z",
      "duration_seconds": 1800
    }
  ],
  "sla_compliance": {            // ADD (Part 2.6)
    "target": 99.9,
    "actual": 99.87,
    "met": false
  }
}
```

**Priority:** Tier 2
**Impact:** Enable SLA tracking (prerequisite for Feature 2.6)
**Effort Estimate:** 32 backend hours

---

### 2.5 Status Page Component Support

**Feature:** Manage status page components via API
**Current API:** Status pages exist but no component management
**Gap:** Missing component CRUD endpoints

**API Changes Needed:**

**New Endpoints:**
```
POST   /v2/statuspages/{uuid}/components
GET    /v2/statuspages/{uuid}/components
GET    /v2/statuspages/{uuid}/components/{component_id}
PUT    /v2/statuspages/{uuid}/components/{component_id}
DELETE /v2/statuspages/{uuid}/components/{component_id}
```

**Component Schema:**
```json
{
  "id": "comp_abc123",
  "name": "API Gateway",
  "description": "Main API endpoint",
  "status": "operational",        // operational, degraded_performance, partial_outage, major_outage
  "group_id": "grp_backend",
  "monitor_uuids": ["mon_abc123", "mon_def456"],
  "created_at": "2024-01-15T10:00:00Z",
  "updated_at": "2024-02-13T08:00:00Z"
}
```

**Component Status Automation:**
```json
{
  "auto_update": true,
  "update_rules": {
    "any_monitor_down": "major_outage",
    "half_monitors_down": "partial_outage",
    "slow_response_time": "degraded_performance"
  }
}
```

**Priority:** Tier 2 (status page differentiation)
**Impact:** Enable full status page automation
**Effort Estimate:** 60-80 backend hours

---

### 2.6 SLA Tracking System

**Feature:** Define and track SLA targets per monitor
**Current API:** Reporting exists but no SLA configuration
**Gap:** No SLA target storage or violation tracking

**API Changes Needed:**

**New Resource: SLA Targets**
```
POST   /v1/monitors/{uuid}/sla
GET    /v1/monitors/{uuid}/sla
PUT    /v1/monitors/{uuid}/sla
DELETE /v1/monitors/{uuid}/sla
```

**SLA Schema:**
```json
{
  "monitor_uuid": "mon_abc123",
  "target_uptime_percentage": 99.9,
  "measurement_window": "monthly",  // hourly, daily, weekly, monthly, yearly
  "alert_on_violation": true,
  "current_compliance": {
    "period_start": "2024-02-01T00:00:00Z",
    "period_end": "2024-02-13T23:59:59Z",
    "actual_uptime": 99.87,
    "target_uptime": 99.9,
    "met": false,
    "remaining_downtime_minutes": 0
  }
}
```

**SLA Reporting Data Source:**
```
GET /v1/sla-reports?from=2024-01-01&to=2024-02-13
```

**Response:**
```json
{
  "period": {
    "from": "2024-01-01T00:00:00Z",
    "to": "2024-02-13T23:59:59Z"
  },
  "monitors": [
    {
      "monitor_uuid": "mon_abc123",
      "monitor_name": "Production API",
      "sla_target": 99.9,
      "actual_uptime": 99.87,
      "met": false,
      "violations": 1,
      "total_downtime_minutes": 56
    }
  ]
}
```

**Priority:** Tier 2 (enterprise requirement)
**Impact:** Enable SLA compliance tracking for enterprise customers
**Effort Estimate:** 80-100 backend hours

---

**Part 2 Summary:**
- **Features:** 6 enhancements needed
- **Total Backend Effort:** 252-356 hours
- **Priority:** Mix of Tier 1 (filtering, pagination) and Tier 2 (SLA, components)

---

## Part 3: New API Endpoints Required

These features require completely new backend functionality.

### 3.1 Resource Tagging System (Tier 2)

**Feature:** Add tags (key-value pairs) to all resources
**Current API:** No tag support on any resource
**Gap:** Missing tags field and tag-based querying

**API Changes Needed:**

#### Add Tags to All Resource Schemas

**Monitor Example:**
```json
{
  "uuid": "mon_abc123",
  "name": "Production API",
  "url": "https://api.example.com",
  "tags": {                          // ADD
    "environment": "production",
    "team": "platform",
    "service": "api-gateway",
    "criticality": "high"
  }
}
```

**Required for:**
- Monitors (`/v1/monitors`)
- Incidents (`/v3/incidents`)
- Maintenance Windows (`/v1/maintenance-windows`)
- Status Pages (`/v2/statuspages`)
- Healthchecks (`/v2/healthchecks`)

#### Tag-Based Filtering
```http
GET /v1/monitors?tags=environment:production,team:platform
GET /v1/monitors?tags=environment:production&tags=team:platform  # AND logic
```

#### Request/Response Changes
**Create/Update Requests:** Add optional `tags` field
```json
{
  "name": "Production API",
  "url": "https://api.example.com",
  "tags": {
    "environment": "production",
    "team": "platform"
  }
}
```

**All GET Responses:** Include `tags` field

**Database Schema Changes:**
- Add `tags` JSONB column to all resource tables
- Index tags for efficient querying
- Support tag validation (max keys, max value length)

**Priority:** Tier 2 (high user demand)
**Impact:** Essential for large-scale infrastructure organization
**Effort Estimate:** 100-120 backend hours
**Complexity:** Medium (schema changes + indexing)

---

### 3.2 Alert Channel / Notification Management (Tier 2)

**Feature:** Configure notification channels and routing rules
**Current API:** No alert channel management via API
**Gap:** Completely missing

**API Changes Needed:**

**New Resource: Alert Channels**
```
POST   /v1/alert-channels
GET    /v1/alert-channels
GET    /v1/alert-channels/{id}
PUT    /v1/alert-channels/{id}
DELETE /v1/alert-channels/{id}
```

**Alert Channel Schema:**
```json
{
  "id": "ch_slack001",
  "name": "Platform Team Slack",
  "type": "slack",
  "config": {
    "webhook_url": "https://hooks.slack.com/...",
    "channel": "#alerts-production"
  },
  "enabled": true,
  "created_at": "2024-01-15T10:00:00Z"
}
```

**Channel Types:**
```json
{
  "type": "slack",
  "config": {
    "webhook_url": "https://hooks.slack.com/...",
    "channel": "#alerts"
  }
}

{
  "type": "email",
  "config": {
    "recipients": ["team@example.com", "oncall@example.com"]
  }
}

{
  "type": "webhook",
  "config": {
    "url": "https://api.example.com/alerts",
    "method": "POST",
    "headers": {"Authorization": "Bearer xxx"},
    "payload_template": "{\"text\": \"{{message}}\"}"
  }
}

{
  "type": "pagerduty",
  "config": {
    "integration_key": "abc123..."
  }
}

{
  "type": "opsgenie",
  "config": {
    "api_key": "xyz789...",
    "team": "platform"
  }
}
```

**New Resource: Notification Rules**
```
POST   /v1/monitors/{uuid}/notification-rules
GET    /v1/monitors/{uuid}/notification-rules
PUT    /v1/monitors/{uuid}/notification-rules/{rule_id}
DELETE /v1/monitors/{uuid}/notification-rules/{rule_id}
```

**Notification Rule Schema:**
```json
{
  "id": "rule_001",
  "monitor_uuid": "mon_abc123",
  "channel_ids": ["ch_slack001", "ch_email001"],
  "conditions": {
    "on_down": true,
    "on_up": false,
    "on_slow_response": false
  },
  "enabled": true
}
```

**Priority:** Tier 2 (critical gap)
**Impact:** Monitoring without alerting is incomplete
**Effort Estimate:** 160-200 backend hours
**Complexity:** High (integrations, webhook handling, retry logic)

---

### 3.3 Monitor Type Expansion (Tier 3 - Future)

**Feature:** Support non-HTTP monitor types
**Current API:** Only HTTP/HTTPS monitoring
**Gap:** Missing monitor types: Ping, TCP, DNS, SSL, Heartbeat

**API Changes Needed:**

**Extend Monitor Schema with Type Discriminator:**
```json
{
  "uuid": "mon_abc123",
  "name": "Database Server",
  "type": "tcp",              // ADD: http, ping, tcp, dns, ssl, heartbeat
  "config": {                 // Type-specific config
    "host": "db.example.com",
    "port": 5432,
    "timeout": 10
  }
}
```

**Type-Specific Configurations:**

**Ping Monitor:**
```json
{
  "type": "ping",
  "config": {
    "host": "server.example.com",
    "packet_count": 4,
    "acceptable_packet_loss": 25
  }
}
```

**TCP Port Monitor:**
```json
{
  "type": "tcp",
  "config": {
    "host": "db.example.com",
    "port": 5432,
    "timeout": 10
  }
}
```

**SSL Certificate Monitor:**
```json
{
  "type": "ssl",
  "config": {
    "hostname": "api.example.com",
    "port": 443,
    "warn_days_before_expiry": 30
  }
}
```

**DNS Monitor:**
```json
{
  "type": "dns",
  "config": {
    "hostname": "example.com",
    "expected_records": ["1.2.3.4", "5.6.7.8"],
    "record_type": "A"
  }
}
```

**Heartbeat/Cron Monitor:**
```json
{
  "type": "heartbeat",
  "config": {
    "schedule": "0 */6 * * *",
    "grace_period_minutes": 15
  }
}
```

**Priority:** Tier 3 (future differentiation)
**Impact:** Feature parity with competitors
**Effort Estimate:** 400+ backend hours (50-80 hours per type)
**Complexity:** Very High (requires different checking infrastructure per type)

---

**Part 3 Summary:**
- **Features:** 3 completely new capabilities
- **Total Backend Effort:** 660-720+ hours
- **Priority:** Tier 2 (tags, alerts), Tier 3 (monitor types)
- **Recommendation:** Implement tags and alerts first (Tier 2), defer monitor types to 2027

---

## Part 4: Priority Questions for Leo

These are YES/NO questions to clarify current API capabilities.

### High Priority (Tier 1 - Blocks Immediate Work)

**Filtering & Pagination:**

- [ ] **Q1:** Does the API support server-side filtering on `/v1/monitors`?
  - Needed for: Efficient monitor querying
  - Example: `GET /v1/monitors?name=production&status=up`
  - If YES: What query parameters are supported?
  - If NO: Backend effort estimate?

- [ ] **Q2:** Does the API support pagination on `/v1/monitors`?
  - Needed for: Scalability (100+ monitors)
  - Example: `GET /v1/monitors?page=0&limit=50`
  - If YES: What's the response format?
  - If NO: Backend effort estimate?

- [ ] **Q3:** Does the API support pagination on `/v3/incidents`?
  - Same as Q2 but for incidents

- [ ] **Q4:** Does the API support pagination on `/v1/maintenance-windows`?
  - Current code suggests partial support - confirm format

**Computed Fields:**

- [ ] **Q5:** Do monitor responses include `created_at` and `updated_at` timestamps?
  - Needed for: Terraform state tracking
  - Currently: Not visible in API responses

- [ ] **Q6:** Do monitor responses include last check time and uptime percentage?
  - Needed for: Enhanced monitoring insights
  - Example fields: `last_check_at`, `uptime_percentage_30d`

### Medium Priority (Tier 2 - Needed for Competitive Parity)

**Tagging:**

- [ ] **Q7:** Do any resources currently support tags/labels?
  - Needed for: Resource organization
  - Check: Monitors, Incidents, Maintenance, Status Pages
  - If YES: What's the schema? (map vs array)
  - If NO: Is this planned in backend roadmap?

- [ ] **Q8:** Is there database infrastructure for storing tags?
  - If YES: JSONB column? Separate table?
  - If NO: Estimated effort to add?

**Alert Channels:**

- [ ] **Q9:** Does an alert channel/notification API exist (not in Terraform provider)?
  - Needed for: Complete infrastructure-as-code
  - Check: Any `/v1/alert-channels` or `/v1/integrations` endpoint?
  - If YES: Provide API documentation
  - If NO: Is this planned?

- [ ] **Q10:** Can monitors be linked to notification channels via API?
  - If YES: What's the endpoint and schema?
  - If NO: How are alerts currently configured?

**SLA Tracking:**

- [ ] **Q11:** Does the reporting API include SLA compliance data?
  - Check: `/v2/reporting/monitor-reports` response
  - If YES: What fields are available?
  - If NO: Is SLA calculation infrastructure planned?

- [ ] **Q12:** Can SLA targets be configured per monitor?
  - If YES: What's the endpoint?
  - If NO: Backend effort estimate?

**Status Pages:**

- [ ] **Q13:** Does status page API support components?
  - Needed for: Status page automation
  - Check: `/v2/statuspages/{uuid}/components` endpoint?
  - If YES: CRUD operations supported?
  - If NO: Estimated effort to add?

- [ ] **Q14:** Can status page components auto-update from monitor status?
  - If YES: What's the configuration mechanism?
  - If NO: Is this feasible?

### Low Priority (Good to Know)

**Performance:**

- [ ] **Q15:** What's the maximum number of resources the API can return in a single list request?
  - Needed for: Understanding scale limitations
  - Example: 1000 monitors? 10,000?

- [ ] **Q16:** Are there rate limits on list endpoints?
  - Needed for: Retry logic design
  - Current: General rate limits known, but endpoint-specific?

**Monitor Types:**

- [ ] **Q17:** Is there backend infrastructure for non-HTTP monitor types?
  - Example: Ping, TCP, DNS monitoring
  - If YES: Which types are supported or in progress?
  - If NO: Is this planned for 2026?

---

## Part 5: API Feature Matrix

| Feature | Current API | Needed API Enhancement | Priority | Blocker? |
|---------|-------------|------------------------|----------|----------|
| **Import Support** | ✅ GET by ID exists | None | Tier 1 | No |
| **Enhanced Errors** | ✅ Error details returned | None | Tier 1 | No |
| **Client Validation** | ✅ N/A (client-side) | None | Tier 1 | No |
| **Documentation** | ✅ N/A | None | Tier 1 | No |
| **Basic Name Filtering** | ✅ List endpoints exist | None (client-side) | Tier 1 | No |
| **Server-Side Filtering** | ❌ No query params | Add filter query params | Tier 1 | **Yes** |
| **Pagination** | ⚠️ Only status pages | Add to all list endpoints | Tier 1 | **Yes** |
| **Computed Timestamps** | ❌ Missing fields | Add created_at, updated_at | Tier 1 | No |
| **Resource Tagging** | ❌ No tag support | Add tags field + filtering | Tier 2 | **Yes** |
| **Alert Channels** | ❌ No endpoint | New alert channel API | Tier 2 | **Yes** |
| **SLA Tracking** | ❌ No SLA config | New SLA target API | Tier 2 | Partial |
| **Status Components** | ❌ No components | New component endpoints | Tier 2 | Partial |
| **Terraform Modules** | ✅ N/A | None | Tier 2 | No |
| **Migration Tools** | ✅ N/A | None | Tier 2 | No |
| **Monitor Types** | ⚠️ HTTP only | New monitor types | Tier 3 | No |

**Legend:**
- ✅ = Exists and sufficient
- ⚠️ = Partial support
- ❌ = Missing
- **Blocker?** = Does it block Terraform provider implementation?

---

## Part 6: Recommended API Roadmap

Aligned with Terraform Provider Roadmap priorities.

### Phase 1: Month 1 (Enable Tier 1 Features)

**Focus:** Scalability and usability foundations

**Backend Tasks:**
1. **Add server-side filtering** (40 hours)
   - Implement query parameters for `/v1/monitors`
   - Support: `name`, `status`, `paused` filters
   - Add basic regex matching for `name`

2. **Standardize pagination** (24 hours)
   - Add pagination to `/v1/monitors`
   - Add pagination to `/v3/incidents`
   - Standardize response format across all list endpoints

3. **Add computed timestamps** (16 hours)
   - Add `created_at`, `updated_at` to all resources
   - Ensure consistent ISO8601 format

**Total Effort:** ~80 backend hours
**Impact:** Unblocks Tier 1 Terraform provider features
**Risk:** Low (standard REST API enhancements)

---

### Phase 2: Months 2-3 (Enable Tier 2 Features)

**Focus:** Enterprise capabilities and competitive parity

**Backend Tasks:**
1. **Resource tagging system** (100 hours)
   - Add `tags` field to all resources
   - Implement tag-based filtering
   - Add database indexes for tag queries
   - Support tag validation (max 50 tags, max length)

2. **Alert channel management** (160 hours)
   - Design alert channel schema
   - Implement CRUD endpoints for channels
   - Support 5 integration types (Slack, email, webhook, PagerDuty, Opsgenie)
   - Implement notification routing logic
   - Add retry/failure handling for webhooks

3. **Enhanced reporting** (32 hours)
   - Add percentile metrics (p95, p99 response times)
   - Include incident history in reports
   - Add real-time uptime percentage

**Total Effort:** ~292 backend hours
**Impact:** Achieve competitive parity with Better Stack
**Risk:** Medium (alert integrations are complex)

---

### Phase 3: Months 4-5 (Enable Advanced Features)

**Focus:** Differentiation and enterprise requirements

**Backend Tasks:**
1. **SLA tracking system** (100 hours)
   - Design SLA target storage schema
   - Implement SLA configuration endpoints
   - Build SLA compliance calculation engine
   - Create SLA reporting endpoints
   - Add SLA violation alerting

2. **Status page components** (80 hours)
   - Design component schema
   - Implement component CRUD endpoints
   - Build component-monitor linking logic
   - Add automated status updates
   - Support component grouping

**Total Effort:** ~180 backend hours
**Impact:** Enable enterprise SLA requirements
**Risk:** Medium (calculation complexity)

---

### Phase 4: Future (Tier 3 - 2027)

**Backend Tasks:**
1. **Monitor type expansion** (400+ hours)
   - Ping/ICMP monitoring
   - TCP port monitoring
   - SSL certificate monitoring
   - DNS monitoring
   - Heartbeat/cron monitoring
   - Each requires separate checking infrastructure

**Total Effort:** 400+ backend hours
**Impact:** Feature parity with full-featured competitors
**Risk:** High (complex infrastructure changes)

---

**Total Backend Roadmap Effort:**
- Phase 1: 80 hours (Month 1)
- Phase 2: 292 hours (Months 2-3)
- Phase 3: 180 hours (Months 4-5)
- **Total for Tier 1+2:** ~552 backend hours (~14 weeks with 1 backend developer)

---

## Part 7: API Design Recommendations

Based on competitive analysis and REST best practices.

### 7.1 Pagination Best Practices

**Recommended Approach:** Offset-based pagination (simplest)

**Standard Query Parameters:**
```
?page=0        # 0-indexed page number
&limit=50      # Results per page (default: 50, max: 200)
```

**Standard Response Format:**
```json
{
  "data": [...],
  "pagination": {
    "page": 0,
    "limit": 50,
    "total": 523,
    "total_pages": 11,
    "has_next": true,
    "has_prev": false
  }
}
```

**Why Not Cursor-Based?**
- More complex to implement
- Offset-based is sufficient for Hyperping's scale
- Better Stack uses offset-based successfully

---

### 7.2 Filtering Query Parameter Patterns

**Recommended Patterns:**

**Simple Equality:**
```
?status=up
?paused=false
```

**Multiple Values (OR logic):**
```
?status=up,down       # status = 'up' OR status = 'down'
```

**Tag Filtering:**
```
?tags=environment:production          # Single tag
?tags=environment:production,team:platform  # Multiple tags (AND logic)
```

**Regex Matching:**
```
?name_regex=.*-api-.*    # URL-encoded regex
```

**Negation (optional enhancement):**
```
?status!=paused          # Exclude paused monitors
```

**Existing Implementation Reference:**
- Status Pages API already uses `?search=term` and `?page=N`
- Reports API uses `?from=date&to=date`
- Standardize on these patterns

---

### 7.3 Sorting Capabilities

**Recommended:**
```
?sort=name              # Ascending by name
?sort=-created_at       # Descending by created_at (- prefix)
```

**Useful Sort Fields:**
- `name`, `-name`
- `created_at`, `-created_at`
- `updated_at`, `-updated_at`
- `status` (alphabetical: down, up)

---

### 7.4 Field Selection (Sparse Fieldsets)

**Optional Enhancement:**
```
?fields=uuid,name,status    # Return only specified fields
```

**Benefits:**
- Reduce response size by 70%
- Faster parsing for clients
- Bandwidth savings

**Priority:** Low (optimization, not required for MVP)

---

### 7.5 Batch Operations Support

**Use Case:** Create 100 monitors at once

**Recommended Approach:**
```
POST /v1/monitors/batch
{
  "monitors": [
    {"name": "API 1", "url": "..."},
    {"name": "API 2", "url": "..."},
    ...
  ]
}
```

**Response:**
```json
{
  "created": [
    {"uuid": "mon_001", "name": "API 1"},
    ...
  ],
  "failed": [
    {"index": 5, "error": "Invalid URL"}
  ]
}
```

**Priority:** Tier 3 (nice to have, not urgent)

---

### 7.6 Consistency Guidelines

**URL Patterns:**
- Use plural nouns: `/monitors`, `/incidents`, `/alert-channels`
- Use hyphens for multi-word: `/maintenance-windows`, `/alert-channels`
- Consistent versioning: `/v1`, `/v2`, `/v3`

**Datetime Format:**
- Always ISO8601 UTC: `2024-02-13T14:30:00Z`
- Never Unix timestamps

**Boolean Fields:**
- Use `true`/`false` (not `1`/`0`)

**ID Fields:**
- Use `uuid` consistently (not `id` sometimes, `uuid` other times)

**Wrapping:**
- Standardize on `{"data": [...], "pagination": {...}}` for lists
- Use `{"resource_name": {...}}` for single resource GET only if needed for metadata

---

## Part 8: Critical Path for Terraform Provider Development

**What Can We Build NOW (Week 1-2):**
1. Complete import support
2. Enhanced error messages
3. Client-side validation
4. Documentation overhaul
5. Basic name filtering (client-side)
6. Terraform modules
7. Migration tools

**Blocked Until Backend Phase 1 Complete (Month 1):**
1. Server-side filtering
2. Efficient pagination for large datasets
3. Computed timestamp tracking

**Blocked Until Backend Phase 2 Complete (Months 2-3):**
1. Resource tagging
2. Tag-based filtering
3. Alert channel management
4. Notification routing

**Blocked Until Backend Phase 3 Complete (Months 4-5):**
1. SLA tracking
2. Status page components
3. Advanced reporting

---

## Appendix A: Current API Endpoint Inventory

**Working Endpoints (Confirmed via Code):**

```
Monitors (v1):
  GET    /v1/monitors
  GET    /v1/monitors/{uuid}
  POST   /v1/monitors
  PUT    /v1/monitors/{uuid}
  DELETE /v1/monitors/{uuid}

Incidents (v3):
  GET    /v3/incidents
  GET    /v3/incidents/{uuid}
  POST   /v3/incidents
  PUT    /v3/incidents/{uuid}
  DELETE /v3/incidents/{uuid}
  POST   /v3/incidents/{uuid}/updates

Maintenance (v1):
  GET    /v1/maintenance-windows
  GET    /v1/maintenance-windows/{uuid}
  POST   /v1/maintenance-windows
  PUT    /v1/maintenance-windows/{uuid}
  DELETE /v1/maintenance-windows/{uuid}

Status Pages (v2):
  GET    /v2/statuspages?page=N&search=term
  GET    /v2/statuspages/{uuid}
  POST   /v2/statuspages
  PUT    /v2/statuspages/{uuid}
  DELETE /v2/statuspages/{uuid}
  GET    /v2/statuspages/{uuid}/subscribers?page=N&type=email
  POST   /v2/statuspages/{uuid}/subscribers
  DELETE /v2/statuspages/{uuid}/subscribers/{id}

Healthchecks (v2):
  GET    /v2/healthchecks
  GET    /v2/healthchecks/{uuid}
  POST   /v2/healthchecks
  PUT    /v2/healthchecks/{uuid}
  DELETE /v2/healthchecks/{uuid}
  POST   /v2/healthchecks/{uuid}/pause
  POST   /v2/healthchecks/{uuid}/resume

Outages (v2):
  GET    /v2/outages
  GET    /v2/outages/{uuid}
  POST   /v2/outages
  DELETE /v2/outages/{uuid}
  POST   /v2/outages/{uuid}/acknowledge
  POST   /v2/outages/{uuid}/unacknowledge
  POST   /v2/outages/{uuid}/resolve
  POST   /v2/outages/{uuid}/escalate

Reports (v2):
  GET    /v2/reporting/monitor-reports?from=date&to=date
  GET    /v2/reporting/monitor-reports/{uuid}?from=date&to=date
```

**Total Endpoints:** ~45 (comprehensive coverage)

---

## Appendix B: Competitor API Feature Comparison

**Query Parameters Support:**

| Provider | Filtering | Pagination | Sorting | Field Selection |
|----------|-----------|------------|---------|-----------------|
| Better Stack | ✅ Yes | ✅ Offset | ✅ Yes | ❌ No |
| StatusPage.io | ✅ Yes | ✅ Offset | ⚠️ Limited | ❌ No |
| PagerDuty | ✅ Yes | ✅ Offset | ✅ Yes | ❌ No |
| Datadog | ✅ Yes | ✅ Cursor | ✅ Yes | ✅ Yes |
| **Hyperping** | ⚠️ Partial | ⚠️ Partial | ❌ No | ❌ No |

**Recommendation:** Match Better Stack (offset pagination + filtering) as minimum viable parity.

---

## Next Steps for Leo

1. **Answer Priority Questions** (Part 4) - especially Q1-Q6
2. **Review Phase 1 Tasks** (Part 6) - confirm feasibility and effort estimates
3. **Prioritize Backend Roadmap** - align with business priorities
4. **Schedule API Enhancement Work** - coordinate with Terraform provider development

**Questions?** Contact Terraform provider team for clarification on specific use cases.

---

**Document Prepared By:** Claude (Terraform Provider Analysis)
**Last Updated:** February 13, 2026
**Next Review:** After Leo's feedback on Priority Questions
