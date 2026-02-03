# Hyperping API Documentation Structure

## Complete API Hierarchy (Discovered from Navigation)

### 1. Monitors API (`/v1/monitors`)
**Parent URL**: https://hyperping.com/docs/api/monitors

**Child Pages**:
- GET List monitors → https://hyperping.com/docs/api/monitors/list
- GET Get monitor → https://hyperping.com/docs/api/monitors/get
- POST Create monitor → https://hyperping.com/docs/api/monitors/create
- PUT Update monitor → https://hyperping.com/docs/api/monitors/update
- DELETE Delete monitor → https://hyperping.com/docs/api/monitors/delete

**Total Pages**: 1 parent + 5 children = **6 pages**

---

### 2. Status Pages API (`/v2/statuspages`)
**Parent URL**: https://hyperping.com/docs/api/statuspages

**Child Pages**:
- GET List status pages → https://hyperping.com/docs/api/statuspages/list
- GET Get status page → https://hyperping.com/docs/api/statuspages/get
- POST Create status page → https://hyperping.com/docs/api/statuspages/create
- PUT Update status page → https://hyperping.com/docs/api/statuspages/update
- DELETE Delete status page → https://hyperping.com/docs/api/statuspages/delete
- GET List subscribers → https://hyperping.com/docs/api/statuspages/subscribers/list
- POST Add subscriber → https://hyperping.com/docs/api/statuspages/subscribers/create
- DELETE Delete subscriber → https://hyperping.com/docs/api/statuspages/subscribers/delete

**Total Pages**: 1 parent + 8 children = **9 pages**

---

### 3. Maintenance API (`/v1/maintenance-windows`)
**Parent URL**: https://hyperping.com/docs/api/maintenance

**Child Pages**:
- GET List maintenance → https://hyperping.com/docs/api/maintenance/list
- GET Get maintenance → https://hyperping.com/docs/api/maintenance/get
- POST Create maintenance → https://hyperping.com/docs/api/maintenance/create
- PUT Update maintenance → https://hyperping.com/docs/api/maintenance/update
- DELETE Delete maintenance → https://hyperping.com/docs/api/maintenance/delete

**Total Pages**: 1 parent + 5 children = **6 pages**

---

### 4. Incidents API (`/v3/incidents`)
**Parent URL**: https://hyperping.com/docs/api/incidents

**Child Pages**:
- GET List incidents → https://hyperping.com/docs/api/incidents/list
- GET Get incident → https://hyperping.com/docs/api/incidents/get
- POST Create incident → https://hyperping.com/docs/api/incidents/create
- PUT Update incident → https://hyperping.com/docs/api/incidents/update
- DELETE Delete incident → https://hyperping.com/docs/api/incidents/delete
- POST Add update → https://hyperping.com/docs/api/incidents/add-update

**Total Pages**: 1 parent + 6 children = **7 pages**

---

### 5. Outages API (`/v1/outages`)
**Parent URL**: https://hyperping.com/docs/api/outages

**Child Pages**:
- GET List outages → https://hyperping.com/docs/api/outages/list
- GET Get outage → https://hyperping.com/docs/api/outages/get
- POST Create incident → https://hyperping.com/docs/api/outages/create
- POST Acknowledge → https://hyperping.com/docs/api/outages/acknowledge
- POST Unacknowledge → https://hyperping.com/docs/api/outages/unacknowledge
- POST Resolve → https://hyperping.com/docs/api/outages/resolve
- POST Escalate → https://hyperping.com/docs/api/outages/escalate
- DELETE Delete incident → https://hyperping.com/docs/api/outages/delete

**Total Pages**: 1 parent + 8 children = **9 pages**

---

### 6. Healthchecks API (`/v2/healthchecks`)
**Parent URL**: https://hyperping.com/docs/api/healthchecks

**Child Pages**:
- GET List healthchecks → https://hyperping.com/docs/api/healthchecks/list
- GET Get healthcheck → https://hyperping.com/docs/api/healthchecks/get
- POST Create healthcheck → https://hyperping.com/docs/api/healthchecks/create
- PUT Update healthcheck → https://hyperping.com/docs/api/healthchecks/update
- DELETE Delete healthcheck → https://hyperping.com/docs/api/healthchecks/delete
- POST Pause healthcheck → https://hyperping.com/docs/api/healthchecks/pause
- POST Resume healthcheck → https://hyperping.com/docs/api/healthchecks/resume

**Total Pages**: 1 parent + 7 children = **8 pages**

---

### 7. Reports API (`/v2/reporting/monitor-reports`)
**Parent URL**: https://hyperping.com/docs/api/reports

**Child Pages**:
- GET List reports → https://hyperping.com/docs/api/reports/list
- GET Get report → https://hyperping.com/docs/api/reports/get

**Total Pages**: 1 parent + 2 children = **3 pages**

---

### 8. Overview (Authentication, Base Info)
**URL**: https://hyperping.com/docs/api/overview

**Child Pages**: None (single page)

**Total Pages**: **1 page**

---

## Summary

### Current Scraper Status
**Currently Scraped**: 9 pages (8 parent pages + 1 Notion page)
- ❌ Parents only, no child endpoints
- ❌ Missing detailed parameters, request/response schemas
- ❌ Missing code examples from child pages

### Complete API Coverage Needed
**Total Pages to Scrape**:
- Parent pages: 8
- Child pages: 5 + 8 + 5 + 6 + 8 + 7 + 2 = **41 children**
- **TOTAL: 49 pages** (+ 1 Notion page = 50 pages)

### Missing Pages
**Currently Missing**: 41 child endpoint pages with:
- Request parameters (types, required/optional, defaults, validation rules)
- Response schemas (field types, examples)
- Code examples (curl, language SDKs)
- Error responses
- Rate limiting details per endpoint

---

## URL Pattern Analysis

### Confirmed Pattern
```
Parent:  https://hyperping.com/docs/api/{section}
Child:   https://hyperping.com/docs/api/{section}/{method}
         https://hyperping.com/docs/api/{section}/{subsection}/{method}
```

### Examples
```
Parent:  /api/monitors
Child:   /api/monitors/list
         /api/monitors/create

Parent:  /api/statuspages
Child:   /api/statuspages/list
         /api/statuspages/subscribers/list  (nested)
```

---

## Next Steps for Scraper Enhancement

1. **Discover child URLs automatically** from parent page navigation
2. **Scrape all 49 pages** (currently only 9)
3. **Extract structured data**: parameters, schemas, examples
4. **Store in hierarchical format**: parent → children relationships
5. **Detect changes at endpoint level**: compare child pages independently
