# Client-Side Feature Enhancement Plan
## Terraform Provider for Hyperping - No API Changes Required

## Context

Following the successful completion of v1.0.9 (Import Enhancement, Client-Side Filtering, Enhanced Error Messages), this plan implements 4 remaining client-side features that provide competitive differentiation without requiring backend API changes.

**Why these features?**
Based on competitive analysis of 13 providers (Better Stack, StatusPage.io, UptimeRobot, Checkly, Cronitor, PagerDuty, Opsgenie, Grafana OnCall, ilert, Site24x7, Datadog, New Relic, Pingdom), we identified critical gaps:

1. **Validation**: Competitors catch errors at plan time; we catch at apply time (poor UX)
2. **Documentation**: Better Stack has 5-minute quick start; we take 30+ minutes
3. **Modules**: Better Stack has 10+ modules; we have 3
4. **Migration**: Competitors offer migration tools; we don't

**Already Complete (v1.0.9):**
- ✅ Import support and testing
- ✅ Enhanced error messages with troubleshooting
- ✅ Client-side filtering for 12 data sources

**Remaining Work:**
1. Client-side validation enhancements (60h)
2. Documentation enhancements (40h)
3. Terraform modules expansion (80h)
4. Migration tools for competitors (40h)

**Total Effort:** 220 hours over 6 phases

---

## Phase 1: Critical Validation Layer (30h)

### Objective
Implement high-priority validators to catch 80% of common errors at plan time instead of apply time.

### Why This Phase First?
- **Immediate user value**: Prevents wasted time on failed applies
- **Competitive gap**: Better Stack, Checkly, and UptimeRobot all validate at plan time
- **Low risk**: Pure addition, no breaking changes

### Tasks

#### 1.1 URL Format Validator (8h)
**Files:**
- `internal/provider/validators.go` - Add URLFormat() validator
- `internal/provider/validators_test.go` - 100% test coverage
- `internal/provider/monitor_resource.go` - Apply to monitor.url
- `internal/provider/statuspage_resource.go` - Apply to hostname, website

**Implementation:**
```go
type urlFormatValidator struct{}

func (v urlFormatValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
    if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
        return
    }
    value := req.ConfigValue.ValueString()
    u, err := url.Parse(value)
    if err != nil || (u.Scheme != "http" && u.Scheme != "https") {
        resp.Diagnostics.AddAttributeError(
            req.Path,
            "Invalid URL Format",
            fmt.Sprintf("The value %q must be a valid HTTP or HTTPS URL", value),
        )
    }
}
```

**Test Cases:**
- ✅ Valid: `https://example.com`, `http://api.example.com/health`
- ❌ Invalid: `ftp://example.com`, `example.com`, `htp://typo.com`

#### 1.2 String Length Validator (6h)
**Files:**
- `internal/provider/validators.go` - Add StringLength(min, max)
- Apply to: monitor.name, incident.title, maintenance.title, statuspage.name, healthcheck.name

**Constraint Reference:**
- Name fields: 1-255 characters (from `internal/client/models_common.go`)
- Message fields: 1-10,000 characters

#### 1.3 Cron Expression Validator (10h)
**Files:**
- `internal/provider/validators.go` - Add CronExpression() validator
- `go.mod` - Add dependency: `github.com/robfig/cron/v3`
- `internal/provider/healthcheck_resource.go` - Apply to cron field

**Implementation:**
```go
import "github.com/robfig/cron/v3"

type cronExpressionValidator struct{}

func (v cronExpressionValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
    if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
        return
    }
    value := req.ConfigValue.ValueString()
    parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
    _, err := parser.Parse(value)
    if err != nil {
        resp.Diagnostics.AddAttributeError(
            req.Path,
            "Invalid Cron Expression",
            fmt.Sprintf("The value %q is not a valid cron expression: %v\nExpected format: 'minute hour day month weekday' (e.g., '0 0 * * *')", value, err),
        )
    }
}
```

**Test Cases:**
- ✅ Valid: `0 0 * * *`, `*/15 * * * *`, `0 9-17 * * 1-5`
- ❌ Invalid: `invalid`, `60 * * * *`, `* * 32 * *`

#### 1.4 Timezone Validator (6h)
**Files:**
- `internal/provider/validators.go` - Add Timezone() validator
- `internal/provider/healthcheck_resource.go` - Apply to timezone field

**Implementation:**
```go
import "time"

type timezoneValidator struct{}

func (v timezoneValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
    if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
        return
    }
    value := req.ConfigValue.ValueString()
    _, err := time.LoadLocation(value)
    if err != nil {
        resp.Diagnostics.AddAttributeError(
            req.Path,
            "Invalid Timezone",
            fmt.Sprintf("The value %q is not a valid IANA timezone (e.g., 'America/New_York', 'Europe/London')", value),
        )
    }
}
```

**Test Cases:**
- ✅ Valid: `America/New_York`, `Europe/London`, `UTC`, `Asia/Tokyo`
- ❌ Invalid: `EST`, `PST`, `invalid`, `New York`

### Definition of Done
- [ ] 4 validators implemented (URL, StringLength, Cron, Timezone)
- [ ] 100% test coverage for new validators (go test -coverprofile)
- [ ] Applied to all relevant resource attributes
- [ ] All existing tests still passing (0 regressions)
- [ ] Documentation strings added (godoc format)
- [ ] Example Terraform configs demonstrating validation errors

### Testing & Verification

**Unit Tests:**
```bash
# Test validators
go test ./internal/provider -v -run "TestURLFormat|TestStringLength|TestCronExpression|TestTimezone"

# Coverage check (target: 90%+)
go test ./internal/provider -coverprofile=coverage.out
go tool cover -func=coverage.out | grep validators
```

**Integration Tests:**
```bash
# Create test config with invalid values
cat > /tmp/test-validation.tf << 'EOF'
resource "hyperping_monitor" "test" {
  name = "Test"
  url  = "ftp://invalid.com"  # Should fail validation
}

resource "hyperping_healthcheck" "cron_test" {
  name     = "Cron Test"
  url      = "https://example.com"
  cron     = "invalid cron"  # Should fail validation
  timezone = "EST"           # Should fail validation
}
EOF

# Validate - should show clear errors
cd /tmp
terraform init
terraform validate  # Should catch errors at plan time (not apply)
```

**Verification Success Criteria:**
- ✅ `terraform validate` catches all 4 types of errors
- ✅ Error messages are clear and actionable
- ✅ Suggests correct format in error message
- ✅ No false positives (valid values pass)

### Agent Team
- **Implementation**: general-purpose (Sonnet) - Complex validator logic
- **Testing**: tdd-guide - Ensure test-first approach
- **Review**: code-reviewer - Validate implementation quality
- **Security**: security-reviewer - Check URL parsing for security issues (SSRF, injection)

### Risks & Mitigation
| Risk | Impact | Mitigation |
|------|--------|------------|
| Cron library adds build size | Low | robfig/cron is industry standard, minimal size |
| Validators too strict vs API | Medium | Keep lenient, document edge cases, let API be authoritative |
| Breaking change for existing configs | Low | Only apply to new resources, don't validate existing state |

---

## Phase 2: Extended Validation & Security (30h)

### Objective
Complete validation coverage for all input types (port, color, email, dates).

### Tasks

#### 2.1 Port Range Validator (4h)
**Implementation:**
```go
type portRangeValidator struct{}

func (v portRangeValidator) ValidateInt64(...) {
    value := req.ConfigValue.ValueInt64()
    if value < 1 || value > 65535 {
        resp.Diagnostics.AddAttributeError(req.Path, "Invalid Port Number",
            fmt.Sprintf("Port must be between 1 and 65535, got %d", value))
    }
}
```
**Apply to:** `monitor.port`

#### 2.2 Hex Color Validator (6h)
**Implementation:**
```go
type hexColorValidator struct{}

func (v hexColorValidator) ValidateString(...) {
    matched, _ := regexp.MatchString(`^#[0-9A-Fa-f]{6}$`, value)
    if !matched {
        resp.Diagnostics.AddAttributeError(req.Path, "Invalid Hex Color",
            fmt.Sprintf("The value %q must be a 6-digit hex color (e.g., '#ff5733')", value))
    }
}
```
**Apply to:** `statuspage.settings.accent_color`

#### 2.3 Email Format Validator (6h)
**Implementation:**
```go
type emailFormatValidator struct{}

func (v emailFormatValidator) ValidateString(...) {
    // RFC 5322 simplified regex
    matched, _ := regexp.MatchString(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`, value)
    if !matched {
        resp.Diagnostics.AddAttributeError(req.Path, "Invalid Email Format",
            fmt.Sprintf("The value %q is not a valid email address", value))
    }
}
```
**Apply to:** `statuspage_subscriber.email` (conditional - when type is "email")

#### 2.4 Cross-Field Date Validation (8h)
**Enhancement to maintenance resource:**
- Validate scheduled_end > scheduled_start
- Validate start date is in future (warning, not error)
- Validate duration is reasonable (< 7 days warning)

#### 2.5 Validation Documentation (6h)
**Update:** `docs/guides/validation.md`

**Add sections:**
- Complete validator reference table
- Common validation errors and fixes
- Best practices for early error detection
- CI/CD integration examples

### Definition of Done
- [ ] 4 additional validators implemented
- [ ] Applied to all relevant attributes
- [ ] Tests passing with 90%+ coverage
- [ ] Validation guide updated
- [ ] Example configs demonstrating all validators

### Testing Strategy
Same as Phase 1 + additional integration tests for cross-field validation.

### Verification
- All 9 validator types working
- `terraform validate` catches 95%+ of invalid inputs
- Error messages provide actionable guidance

---

## Phase 3: Getting Started Documentation (20h)

### Objective
Create step-by-step tutorial to reduce time-to-first-monitor from 30min to 5min.

**Competitive Gap:** Better Stack has excellent quick start; we don't.

### Tasks

#### 3.1 Quick Start Tutorial (8h)
**Create:** `docs/guides/quickstart.md`

**Structure:**
```markdown
# Quick Start - 5 Minute Guide

## Prerequisites (30 seconds)
- [ ] Terraform 1.8+ installed (`terraform version`)
- [ ] Hyperping account (sign up at hyperping.io)
- [ ] API key (Settings → API Keys → Create)

## Step 1: Configure Provider (2 minutes)
```hcl
terraform {
  required_providers {
    hyperping = {
      source  = "develeap/hyperping"
      version = "~> 1.0"
    }
  }
}

provider "hyperping" {
  api_key = var.hyperping_api_key  # Export HYPERPING_API_KEY
}
```

## Step 2: Create Your First Monitor (1 minute)
[Copy-paste ready configuration with explanation]

## Step 3: Apply and Verify (1 minute)
```bash
terraform init
terraform plan
terraform apply
```

[Screenshot of expected output]

## Step 4: Check Dashboard (30 seconds)
[Screenshot showing monitor in Hyperping dashboard]

## Next Steps
- [Add more monitors]
- [Create status page]
- [Set up alerts]
```

**Visual Elements:**
- Screenshots of Hyperping dashboard
- Mermaid architecture diagram
- Expected vs actual output
- Common pitfalls section

#### 3.2 Use Case Guides (8h)
**Create 3 guides:**

**A. Microservices Monitoring** (`docs/guides/use-case-microservices.md`)
- Multiple service endpoints
- Service dependency monitoring
- Regional deployments
- Dynamic configuration with for_each

**B. Kubernetes Monitoring** (`docs/guides/use-case-kubernetes.md`)
- Ingress health endpoints
- Service mesh integration
- Pod readiness checks
- Helm chart integration

**C. API Gateway Monitoring** (`docs/guides/use-case-api-gateway.md`)
- REST API health checks
- GraphQL endpoints
- Authentication patterns
- Rate limit handling

**Each guide includes:**
- Problem statement
- Architecture diagram (Mermaid)
- Complete working example (tested)
- Customization guide
- Troubleshooting section

#### 3.3 Landing Page Improvement (4h)
**Update:** `docs/index.md`

**Sections:**
- Hero: "Monitor your infrastructure with Terraform"
- Quick links: "I want to..." navigation
- Feature highlights with links
- Getting started CTA
- Community resources

### Definition of Done
- [ ] Quick Start guide created and tested (<5 min completion)
- [ ] 3 use case guides created with working examples
- [ ] All code examples validated (`terraform validate`)
- [ ] Screenshots added (6+ images)
- [ ] Landing page redesigned
- [ ] Internal linking between guides
- [ ] SEO optimization (titles, meta descriptions)

### Testing & Verification

**Code Validation:**
```bash
# Extract and test all code blocks from guides
python scripts/test-docs-examples.py docs/guides/quickstart.md
python scripts/test-docs-examples.py docs/guides/use-case-*.md

# Expected: All examples pass terraform validate
```

**User Testing:**
- Recruit 3-5 new Terraform users
- Have them follow Quick Start guide
- Time to completion (target: < 5 minutes)
- Collect feedback via survey
- Iterate based on feedback

**Success Metrics:**
- ✅ Time-to-first-monitor: 30min → 5min (83% reduction)
- ✅ Quick Start completion rate: 80%+
- ✅ Support tickets reduced by 40%

### Agent Team
- **Writing**: doc-updater - Structure and content
- **Review**: code-reviewer - Validate examples work
- **UX**: Architect - Navigation and information flow

---

## Phase 4: Best Practices & Migration Guides (20h)

### Objective
Consolidate patterns and provide competitor migration paths to accelerate adoption.

### Tasks

#### 4.1 Best Practices Guide (8h)
**Create:** `docs/guides/best-practices.md`

**Sections:**

**1. Naming Conventions**
```hcl
# Good: Environment-Category-Service pattern
name = "[PROD]-API-UserService"

# Bad: Inconsistent naming
name = "prod user api"
```

**2. State Management**
- Remote state (S3, Terraform Cloud)
- State locking
- Workspace strategies

**3. Security**
- API key management (env vars)
- Sensitive data handling
- Secret scanning

**4. Resource Organization**
- File structure for large deployments
- Module composition
- Dependencies

**5. CI/CD Integration**
- GitHub Actions examples
- Pre-apply validation
- Drift detection

**6. Testing**
- Native Terraform tests
- Contract tests
- Plan validation

#### 4.2 Competitor Migration Guides (12h)
**Create 3 guides:**

**A. Better Stack Migration** (`docs/guides/migrate-from-betterstack.md`)
- Why migrate section
- Resource mapping table
- Configuration conversion
- Import workflow
- Feature parity comparison

**B. UptimeRobot Migration** (`docs/guides/migrate-from-uptimerobot.md`)
- Monitor type mapping
- Alert contact migration
- Import script

**C. Pingdom Migration** (`docs/guides/migrate-from-pingdom.md`)
- Check type conversion
- Tag to naming convention
- Bulk import

**Common structure for each:**
```markdown
# Migrating from [Competitor] to Hyperping

## Why Migrate?
[Hyperping advantages]

## Prerequisites
- Export from [Competitor]
- Hyperping API key
- Terraform 1.8+

## Step-by-Step Migration
1. Export configuration
2. Convert to Hyperping format
3. Import resources
4. Validate

## Resource Mapping
| [Competitor] | Hyperping | Notes |
|--------------|-----------|-------|

## Troubleshooting
[Common issues]
```

### Definition of Done
- [ ] Best practices guide covering 6 topics
- [ ] 3 competitor migration guides
- [ ] All code examples tested
- [ ] Resource mapping tables complete
- [ ] Migration success stories (optional testimonials)

### Testing & Verification
- Extract and validate all code examples
- Manual migration test for each competitor
- Verify 95%+ conversion success rate

### Agent Team
- **Writing**: doc-updater
- **Review**: Architect - Validate patterns
- **Validation**: code-reviewer

---

## Phase 5: Terraform Modules Expansion (80h)

### Objective
Expand from 3 to 10 production-ready modules covering 80% of common use cases.

**Competitive Gap:** Better Stack has 10+ modules; we have 3.

### Existing Modules
1. ✅ api-health - HTTP/HTTPS API monitoring
2. ✅ ssl-monitor - SSL certificate monitoring
3. ✅ statuspage-complete - Full status page setup

### New Modules (7 total)

#### 5.1 Database Monitoring Module (12h)
**Path:** `examples/modules/database-monitor/`

**Features:**
- TCP port monitoring for PostgreSQL, MySQL, Redis, MongoDB
- Multiple database instances
- Regional redundancy
- Configurable check frequency

**Example Usage:**
```hcl
module "databases" {
  source = "./modules/database-monitor"

  name_prefix = "PROD"
  databases = {
    postgres = {
      host = "db.example.com"
      port = 5432
      type = "postgresql"
    }
    redis = {
      host = "cache.example.com"
      port = 6379
      type = "redis"
    }
  }
  regions = ["virginia", "london"]
}
```

#### 5.2 CDN Monitoring Module (10h)
**Features:**
- Monitor CDN edge locations
- Cache performance checks
- Multi-region validation
- Static asset availability

#### 5.3 Cron Healthcheck Module (10h)
**Features:**
- Dead man's switch for cron jobs
- Multiple job schedules
- Grace period configuration
- Timezone support

**Example:**
```hcl
module "cron_jobs" {
  source = "./modules/cron-healthcheck"

  jobs = {
    daily_backup = {
      cron     = "0 2 * * *"
      timezone = "America/New_York"
      grace    = 30  # minutes
    }
  }
}
```

#### 5.4 Multi-Environment Module (12h)
**Features:**
- Deploy same monitors across dev/staging/prod
- Environment-specific configuration
- Workspace integration
- Consistent naming

#### 5.5 Incident Management Module (10h)
**Features:**
- Pre-configured incident templates
- Status page integration
- Maintenance window scheduling
- Update workflow

#### 5.6 Website Monitoring Module (8h)
**Features:**
- Homepage availability
- Critical page paths
- Expected content validation
- Performance thresholds

#### 5.7 GraphQL Monitoring Module (8h)
**Features:**
- GraphQL endpoint health
- Query validation
- Introspection checks
- Authentication support

#### 5.8 Module Registry Preparation (10h)
**Tasks:**
- GitHub releases automation
- Version tagging
- Registry documentation
- Publish first 3 modules

**Files:**
- `.github/workflows/release-modules.yml`
- `terraform-registry.md`

### Definition of Done
- [ ] 7 new modules created (total 10)
- [ ] Each module has: README, tests, examples, versions.tf
- [ ] All tests passing (`terraform test`)
- [ ] Modules published to Terraform Registry
- [ ] Module catalog page created
- [ ] Integration tests for module composition

### Testing Strategy
```bash
# Test each module
for module in examples/modules/*/; do
    cd "$module"
    terraform init
    terraform test
    cd -
done

# Test module composition (use multiple together)
cd examples/complete-stack
terraform init
terraform validate
terraform plan
```

### Verification
- ✅ 10 production modules
- ✅ 80% use case coverage
- ✅ Registry downloads > 0
- ✅ 100% test coverage for modules

### Agent Team
- **Design**: Planner - Module interfaces and APIs
- **Implementation**: general-purpose (Sonnet)
- **Testing**: tdd-guide
- **Review**: code-reviewer

---

## Phase 6: Competitor Migration Tools (40h)

### Objective
Build automated migration tools for top 3 competitors to reduce migration friction.

**Competitive Advantage:** No other provider offers migration FROM competitors TO their platform.

### Tasks

#### 6.1 Better Stack Migration Tool (15h)
**Path:** `cmd/migrate-betterstack/`

**Features:**
- Export Better Stack monitors via API
- Convert to Hyperping Terraform config
- Generate import commands
- Handle alert mapping
- Migration report

**Implementation:**
```go
type BetterStackMonitor struct {
    ID              string   `json:"id"`
    URL             string   `json:"url"`
    Name            string   `json:"pronounceable_name"`
    CheckFrequency  int      `json:"check_frequency"`
    MonitorType     string   `json:"monitor_type"`
    Regions         []string `json:"regions"`
}

func convertBetterStackMonitor(bs BetterStackMonitor) HyperpingMonitorConfig {
    return HyperpingMonitorConfig{
        Name:      bs.Name,
        URL:       bs.URL,
        Frequency: bs.CheckFrequency,
        Protocol:  mapMonitorType(bs.MonitorType),
        Regions:   mapRegions(bs.Regions),
    }
}
```

**Usage:**
```bash
export BETTERSTACK_API_TOKEN="..."
export HYPERPING_API_KEY="..."

go run ./cmd/migrate-betterstack \
    --output=migrated-resources.tf \
    --import-script=import.sh \
    --validate
```

**Output:**
- `migrated-resources.tf` - Terraform config
- `import.sh` - Import commands
- `migration-report.json` - Mapping and warnings
- `manual-steps.md` - Manual migration items

#### 6.2 UptimeRobot Migration Tool (12h)
**Path:** `cmd/migrate-uptimerobot/`

**Features:**
- Parse UptimeRobot API export
- Convert monitor types (HTTP, Keyword, Port, Ping)
- Map alert contacts
- Handle maintenance windows
- Migration report

**Type Mapping:**
```
UptimeRobot     → Hyperping
HTTP(s)         → protocol=http
Keyword         → protocol=http (with keyword)
Port            → protocol=port
Ping            → protocol=icmp
```

#### 6.3 Pingdom Migration Tool (13h)
**Path:** `cmd/migrate-pingdom/`

**Features:**
- Export Pingdom checks via API
- Convert check types (HTTP, TCP, DNS, SMTP)
- Map tags to naming conventions
- Handle integrations
- Support transaction checks

**Check Mapping:**
```
Pingdom         → Hyperping
HTTP            → protocol=http
HTTPCUSTOM      → protocol=http (headers/body)
TCP             → protocol=port
DNS             → Manual (not supported)
SMTP            → protocol=port (port=25)
```

#### 6.4 Testing & Documentation (10h)

**Testing:**
1. Create test accounts in each service
2. Set up diverse monitors
3. Run migration tool
4. Validate output
5. Verify import success
6. Check for drift

**Documentation:**
- `docs/guides/automated-migration.md`
- Tool-specific READMEs
- Troubleshooting guides
- Video walkthrough (optional)

### Definition of Done
- [ ] 3 migration tools implemented
- [ ] Each tool handles primary resource types
- [ ] Validation and dry-run modes
- [ ] Migration reports generated
- [ ] Documentation complete
- [ ] Test suite passing
- [ ] 95%+ success rate for common configs

### Testing Strategy
```bash
# Unit tests
go test ./cmd/migrate-betterstack -v
go test ./cmd/migrate-uptimerobot -v
go test ./cmd/migrate-pingdom -v

# Integration test
export BETTERSTACK_API_TOKEN="test"
export HYPERPING_API_KEY="test"
go run ./cmd/migrate-betterstack --dry-run --validate

# Verify output
cd /tmp/migration-test
terraform init
terraform validate
terraform plan  # Should show resources to create

# After import
terraform plan  # Should show no changes (0 drift)
```

### Verification
- ✅ 95%+ migration success rate
- ✅ <10 minutes for 100 monitors
- ✅ 0 post-migration drift
- ✅ User satisfaction survey

### Agent Team
- **Design**: Planner - Tool architecture
- **Implementation**: general-purpose (Sonnet)
- **Testing**: tdd-guide
- **Documentation**: doc-updater

---

## Execution Strategy

### Parallel Execution Tracks

**Track 1: Validation (Developer A)**
- Week 1-2: Phase 1 (Critical validators)
- Week 3-4: Phase 2 (Extended validation)

**Track 2: Documentation (Developer B / Tech Writer)**
- Week 1-2: Phase 3 (Quick start)
- Week 3-4: Phase 4 (Best practices & migration)

**Track 3: Modules (Developer C)**
- Week 1-6: Phase 5 (7 new modules)

**Track 4: Migration Tools (Developer D)**
- Week 5-7: Phase 6 (Competitor migrations)

**Timeline:**
- Weeks 1-2: Tracks 1, 2 start (Phases 1, 3)
- Weeks 3-4: Tracks 1, 2, 3 active (Phases 2, 4, 5)
- Weeks 5-6: Tracks 3, 4 active (Phases 5, 6)
- Week 7: Integration, testing, polish

### Sequential Dependencies
- Phase 2 depends on Phase 1 (validation foundation)
- Phase 4 references Phase 3 (links to quick start)
- All others can run in parallel

### Recommended Order (If Sequential)
1. **Phase 1 + Phase 3** (parallel) - Maximum immediate value
2. **Phase 2 + Phase 4 + Phase 5** (parallel) - Comprehensive coverage
3. **Phase 6** - Competitive advantage
4. Integration & ship

---

## Success Metrics

### Phase 1-2: Validation
- **Plan-time error detection:** 80%+ of errors caught before apply
- **Validator coverage:** 90%+ of string/int64 attributes
- **Test coverage:** 90%+ for validator code
- **User feedback:** "Terraform validate catches my mistakes"

### Phase 3-4: Documentation
- **Time-to-first-monitor:** 30min → 5min (83% reduction)
- **Guide completion rate:** 80%+ (analytics tracking)
- **Support ticket reduction:** 40% fewer "how do I..." questions
- **SEO ranking:** Top 3 for "Hyperping Terraform tutorial"

### Phase 5: Modules
- **Module count:** 3 → 10 modules
- **Use case coverage:** 80% of common scenarios
- **Registry downloads:** 100+ in first month
- **Test coverage:** 100% (all modules have tests)

### Phase 6: Migration
- **Migration success:** 95%+ for standard monitors
- **Migration speed:** <10 min for 100 monitors
- **Post-migration drift:** 0 resources
- **Competitive conversions:** 20% increase from Better Stack

### Overall Success
- **Feature parity:** Match Better Stack (leading competitor)
- **User satisfaction:** +15 NPS points
- **Adoption velocity:** +30% new provider users
- **Community engagement:** 10+ community contributions

---

## Risk Assessment

### High Risk
| Risk | Impact | Mitigation |
|------|--------|------------|
| Cron library breaks builds | High | Use well-maintained robfig/cron, test extensively |
| Migration tools handle edge cases poorly | Medium | Extensive testing, clear docs on limitations |

### Medium Risk
| Risk | Impact | Mitigation |
|------|--------|------------|
| Validators too strict vs API | Low | Keep lenient, let API validate edge cases |
| Documentation becomes outdated | Medium | Version docs, CI validation checks |
| Module complexity too high | Medium | Simple + advanced examples, progressive disclosure |

### Low Risk
| Risk | Impact | Mitigation |
|------|--------|------------|
| Competitor API changes | Low | Version migration tools, document API versions |
| Breaking changes | Low | Only add validators to new attributes |

---

## Critical Files

### Phase 1-2: Validation
- `internal/provider/validators.go` - 9 new validators
- `internal/provider/validators_test.go` - Test coverage
- `internal/provider/monitor_resource.go` - Apply validators
- `internal/provider/healthcheck_resource.go` - Cron/timezone
- `internal/provider/statuspage_resource.go` - URL/color

### Phase 3-4: Documentation
- `docs/guides/quickstart.md` - 5-minute tutorial
- `docs/guides/use-case-microservices.md`
- `docs/guides/use-case-kubernetes.md`
- `docs/guides/use-case-api-gateway.md`
- `docs/guides/best-practices.md`
- `docs/guides/migrate-from-betterstack.md`
- `docs/guides/migrate-from-uptimerobot.md`
- `docs/guides/migrate-from-pingdom.md`

### Phase 5: Modules
- `examples/modules/database-monitor/`
- `examples/modules/cdn-monitor/`
- `examples/modules/cron-healthcheck/`
- `examples/modules/multi-environment/`
- `examples/modules/incident-mgmt/`
- `examples/modules/website-monitor/`
- `examples/modules/graphql-monitor/`

### Phase 6: Migration Tools
- `cmd/migrate-betterstack/main.go`
- `cmd/migrate-uptimerobot/main.go`
- `cmd/migrate-pingdom/main.go`
- `docs/guides/automated-migration.md`

---

## Verification Approach

### Phase-by-Phase Verification

**Phase 1-2:**
```bash
# Unit tests
go test ./internal/provider -run "TestValidator" -coverprofile=coverage.out
go tool cover -func=coverage.out | grep "validators.go"

# Integration test
terraform validate  # Should catch errors at plan time
```

**Phase 3-4:**
```bash
# Code validation
python scripts/test-docs-examples.py docs/guides/*.md

# User testing
# - Time 5 users following Quick Start
# - Target: <5 minutes average
```

**Phase 5:**
```bash
# Module tests
for module in examples/modules/*/; do
    cd "$module"
    terraform test
done

# Composition test
cd examples/complete-stack
terraform plan
```

**Phase 6:**
```bash
# Migration tool tests
go test ./cmd/migrate-* -v

# End-to-end migration
go run ./cmd/migrate-betterstack --dry-run
terraform init && terraform plan
```

### Continuous Integration
All phases include CI/CD validation in GitHub Actions (see plan agent output for workflow YAML).

---

## Summary

This plan delivers **4 high-value client-side features** over **220 hours** across **6 phases**:

1. **Phase 1** (30h): Critical validation - Immediate user value
2. **Phase 2** (30h): Extended validation - Comprehensive coverage
3. **Phase 3** (20h): Quick start docs - Major UX improvement
4. **Phase 4** (20h): Best practices & migration - Reduced support burden
5. **Phase 5** (80h): Module expansion - Competitive differentiation
6. **Phase 6** (40h): Migration tools - Accelerate adoption

**Key Differentiators:**
- ✅ Plan-time error detection (vs competitors' apply-time failures)
- ✅ 5-minute quick start (vs 30+ minutes)
- ✅ 10 production modules (vs Better Stack's current lead)
- ✅ Automated migration from top 3 competitors (unique offering)

**Each phase is independently shippable** - continuous value delivery.

**Recommended execution:** Phases 1+3 → 2+4+5 → 6 → Integration & ship
