# Provider Backlog: Terraform Provider Hyperping

Unified backlog combining competitive analysis (7 providers), API gap audit, and undocumented API field discovery.

**Current version:** v1.9.2
**Last updated:** 2026-03-25

---

## Unified Backlog

### P0 — Ship Next

| ID | Item | Type | Effort | Source |
|----|------|------|--------|--------|
| B02 | **Report data source: expose MTTR** — `mttr` + `mttr_formatted` in Go struct, not in TF schema | Bug fix | ~20 lines | API audit |
| B03 | **Provider-level configuration** — `max_retries`, `retry_wait_min/max`, `request_timeout`, `rate_limit` | Feature | 2-3 days | Competitive (Better Uptime, Uptime.com) |

### P1 — High Value Enhancements

| ID | Item | Type | Effort | Source |
|----|------|------|--------|--------|
| B04 | **Outage: expose `confirmed_locations`, `acknowledged_at`** — already in Go struct | Schema gap | ~30 lines | API audit |
| B05 | **Maintenance: expose `status`** (upcoming/ongoing/completed) — already parsed | Schema gap | ~15 lines | API audit |
| B06 | **Maintenance: expose `timezone`, `created_at`, `created_by`, `updates`** | Schema gap | 1 day | API audit |
| B07 | **Report data source: outage details + SLA convenience fields** (`sla_nines`, `downtime_human`, `mttr`) | Feature | 1 day | Competitive (Uptime.com, Datadog) |
| B08 | **Outage lifecycle: `desired_state`** (acknowledged/resolved/escalated) via existing client methods | Feature | 1-2 days | API audit |
| B09 | **Healthcheck: expose `lastDowntime`, `due_date`** — operational visibility | Schema gap | 0.5 day | VCR cassette audit |
| B10 | **Healthcheck: writable `is_paused`** via pause/resume API | Feature | 0.5 day | API audit |
| B11 | **Error diagnostics: doc URL links** — append Hyperping docs link to error messages | Feature | 0.5 day | Competitive (Datadog, StatusCake) |
| B12 | **Outage: expose `alertedChannels`, `errorHeader`** — incident response data | Schema gap | 0.5 day | VCR cassette audit |
| B13 | **Cross-resource UUID validation** — validate monitor UUIDs in status pages/maintenance at plan time | Feature | 1 day | Creative |

### P2 — Architecture & DX

| ID | Item | Type | Effort | Source |
|----|------|------|--------|--------|
| B14 | **Schema attribute helpers** — DRY factory functions for common attributes | Refactor | 2-3 days | Competitive (Uptime.com, Checkly) |
| B15 | **PlanValuePreserver interface** — centralize write-only field handling | Refactor | 1-2 days | Competitive (Uptime.com) |
| B16 | **Import generator data source** — generate `terraform import` commands/blocks | Feature | 2-3 days | Unique advantage |
| B17 | **Maintenance recurrence** — investigate API support for recurring windows | Feature | 1-2 days | Competitive (Checkly, StatusCake, Uptime.com, Datadog) |
| B18 | **Status page enhancements** — section ordering, component count, subscription domain filtering | Feature | 1-2 days | Competitive (Better Uptime, Uptime.com) |
| B19 | **Escalation config enrichment** — expose escalation policy details from monitor GET | Feature | 1-2 days | Competitive (Uptime.com, Better Uptime, Checkly) |
| B20 | **Monitor `group_id` and `sort_order`** — if monitor groups become user-visible | Schema gap | 0.5 day | VCR cassette audit |
| B21 | **Maintenance notification tracking** — `scheduledNotificationStatus/SentAt/Breakdown` | Schema gap | 0.5 day | VCR cassette audit |

### P3 — Documentation, Modules & Future

| ID | Item | Type | Effort | Source |
|----|------|------|--------|--------|
| B22 | **Security & validation docs** — document our competitive advantages | Documentation | 1 day | Competitive analysis |
| B23 | **SLA tracking guide** — examples combining reports + outputs | Documentation | 0.5 day | Competitive (Uptime.com) |
| B24 | **Official Terraform modules** — `hyperping-monitored-service`, `hyperping-statuspage-complete` | Modules | 2-3 days | Creative |
| B25 | **Monitor fleet management examples** — YAML-driven `for_each` patterns | Examples | 0.5 day | Creative |
| B26 | **Status page DNS CNAME output** — help users set up custom domains with other providers | Feature | 0.5 day | Creative |
| B27 | **Healthcheck `slug`, `logs`** — human-readable ID and ping history | Schema gap | 0.5 day | VCR cassette audit |
| B28 | **Status page `sso_connection` full object** — expose beyond just UUID | Schema gap | 0.5 day | VCR cassette audit |
| B29 | **Escalation Policies resource** — CRUD for `/v2/escalation-policies`, currently referenced by UUID only | New resource | 2-3 days | API audit (API_COMPLETENESS_AUDIT.md) |
| B30 | **Notification Channels / Integrations resource** — Slack, PagerDuty, Webhook, Email, SMS | New resource | 3-5 days | API audit (API_COMPLETENESS_AUDIT.md) |
| B31 | **Projects / Workspaces resource** — `projectUuid` already in monitor model | New resource | 1-2 days | API audit (API_COMPLETENESS_AUDIT.md) |
| B32 | **Team Management resource** — user/RBAC management | New resource | 3-5 days | API audit (API_COMPLETENESS_AUDIT.md) |
| B33 | **E2E stress test: 50+ monitors** — large-scale scenario for performance validation | Testing | 1 day | test/e2e/README.md |
| B34 | **Migration CLI: visual progress indicators** — better UX for large migrations | Enhancement | 1 day | MIGRATION_CERTIFICATION.md |
| B35 | **Migration CLI: batch mode** — handle very large account migrations | Enhancement | 1-2 days | MIGRATION_CERTIFICATION.md |

---

## Completed Items

| ID | Item | Version | Date |
|----|------|---------|------|
| ~~Phase 2~~ | User-Agent versioning | v1.4.x | 2026-02 |
| ~~Phase 3~~ | Monitoring locations data source (17 regions) | v1.5.0 / v1.7.3 | 2026-03-16 / 2026-03-20 |
| ~~Phase 5~~ | Cross-field validators (monitor, maintenance, healthcheck) | v1.5.0 / v1.7.1 / v1.7.2 | 2026-03-16 / 2026-03-18 / 2026-03-19 |
| ~~Phase 6~~ | Enhanced error diagnostics (mostly done, doc URLs remaining → B11) | v1.5.0 / v1.7.1 | 2026-03-16 / 2026-03-18 |
| ~~Phase 8~~ | Bulk data source enhancements (`total`, `ids`) | v1.5.0 | 2026-03-16 |
| ~~Phase 15~~ | Cassette-based VCR testing (42 cassettes) | v1.6.0 | 2026-03-16 |
| ~~PR #88~~ | Maintenance update fields, dead code removal | v1.7.1 | 2026-03-18 |
| ~~PR #89~~ | Comprehensive review fixes (19 changes, security hardening) | v1.7.2 | 2026-03-19 |
| ~~PR #90~~ | 9 missing regions, notification_option "none", languages sync | v1.7.3 | 2026-03-20 |
| ~~PR #93~~ | API gaps: incident updates, sso_connection_uuid, service description | v1.8.0 | 2026-03-20 |
| ~~PR #94~~ | Write-only field preservation (nested services, required_keyword) | v1.8.1 | 2026-03-21 |
| ~~PR #95~~ | Nested service description localized map format | v1.8.2 | 2026-03-22 |
| ~~PR #96~~ | Plan-time warning for nested service description API limitation | v1.8.3 | 2026-03-22 |
| ~~B01~~ | Add `capetown` to AllowedRegions (18 regions) | v1.9.0 | 2026-03-23 |
| — | Outage severity/summary fields | v1.9.0 | 2026-03-23 |
| — | Language/notification_option enum defaults | v1.9.1 | 2026-03-24 |
| — | Subscriber pagination fix | v1.9.2 | 2026-03-25 |

---

## API Limitations (Requires Hyperping Server Changes)

These cannot be fixed provider-side:

| Bug # | Issue | Workaround |
|-------|-------|------------|
| #1 | Renderer uses v1 numeric IDs for status | UUID-to-numeric translation on every write |
| #6/#7 | Incident/maintenance `text` not returned on GET | Plan value preservation |
| #17 | `settings.name` overridden by `resource.name` on read | `replaceSettingsName()` restoration |
| #20 | Nested service description not persisted | Plan-time warning (v1.8.3) + state preservation |
| #21 | Nested service `show_response_times` defaults to true | Send on write + state preservation |
| #22 | Monitor `required_keyword` not returned on GET (possible regression) | State preservation (v1.8.1) |

## Out of Scope (Requires New API Endpoints)

- New monitor types (SMTP, IMAP, SSH, FTP, NTP, WHOIS)
- Integration resources (Slack, PagerDuty, Opsgenie, webhook destinations)
- SLO/SLA target resources
- On-call calendar/rotation resources
- Contact group management
- Severity/priority level resources
- Browser/transaction check support
- Real User Monitoring
- Multi-account/subaccount support
- Dashboard/reporting resources
