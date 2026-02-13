---
page_title: "Best Practices Guide - Production-Ready Infrastructure"
subcategory: "Guides"
description: |-
  Comprehensive guide to production-ready monitoring infrastructure with the Hyperping Terraform provider. Covers naming conventions, state management, security, resource organization, CI/CD integration, and testing.
---

# Best Practices Guide - Production-Ready Infrastructure

This guide provides comprehensive best practices for building production-ready monitoring infrastructure with the Hyperping Terraform provider. Follow these patterns to ensure secure, maintainable, and scalable monitoring deployments.

## Table of Contents

- [Naming Conventions](#naming-conventions)
- [State Management](#state-management)
- [Security Best Practices](#security-best-practices)
- [Resource Organization](#resource-organization)
- [CI/CD Integration](#cicd-integration)
- [Testing and Validation](#testing-and-validation)
- [Anti-Patterns to Avoid](#anti-patterns-to-avoid)
- [Production Checklist](#production-checklist)

---

## Naming Conventions

Consistent naming is critical for large-scale monitoring deployments. A well-designed naming convention makes resources discoverable, maintainable, and auditable.

### The Environment-Category-Service Pattern

Use this three-part pattern for all monitor names:

```
[ENVIRONMENT]-[CATEGORY]-[SERVICE]
```

**Benefits:**
- Instant recognition of environment (prevent production incidents)
- Logical grouping in Hyperping dashboard
- Easy filtering in Terraform and Hyperping API
- Clear ownership and alerting boundaries

### Pattern Components

**1. Environment (Required)**

| Environment | Code | Description | Example |
|-------------|------|-------------|---------|
| Production | `PROD` | Live customer-facing services | `[PROD]-API-Payment` |
| Staging | `STG` | Pre-production testing | `[STG]-API-Payment` |
| Development | `DEV` | Development environment | `[DEV]-API-Payment` |
| Testing | `TEST` | Automated testing | `[TEST]-API-Mock` |
| Demo | `DEMO` | Sales demonstrations | `[DEMO]-App-Frontend` |

**2. Category (Required)**

| Category | Description | Example Services |
|----------|-------------|------------------|
| `API` | HTTP APIs and REST endpoints | Payment, Authentication, GraphQL |
| `Web` | Web applications and frontends | Marketing site, Admin panel |
| `DB` | Database servers | PostgreSQL, MySQL, Redis |
| `Cache` | Caching layers | Redis, Memcached, Varnish |
| `Queue` | Message queues | RabbitMQ, SQS, Kafka |
| `Storage` | Storage services | S3, MinIO, Object Storage |
| `CDN` | Content delivery networks | CloudFront, Cloudflare |
| `DNS` | DNS servers | Authoritative, Resolver |
| `VPN` | VPN endpoints | Client VPN, Site-to-Site |
| `Cron` | Scheduled jobs | Backups, Reports, Cleanup |

**3. Service (Required)**

The specific service or component being monitored. Use consistent names across your infrastructure.

**Examples:**
- `Payment` - Payment processing API
- `Auth` - Authentication service
- `GraphQL` - GraphQL gateway
- `Frontend` - User-facing web application
- `AdminPanel` - Internal administration interface

### Complete Examples

```hcl
# Production API monitors
resource "hyperping_monitor" "prod_api_payment" {
  name     = "[PROD]-API-Payment"
  url      = "https://api.example.com/v1/payment/health"
  protocol = "http"
}

resource "hyperping_monitor" "prod_api_auth" {
  name     = "[PROD]-API-Auth"
  url      = "https://api.example.com/v1/auth/health"
  protocol = "http"
}

# Staging monitors
resource "hyperping_monitor" "stg_api_payment" {
  name     = "[STG]-API-Payment"
  url      = "https://staging-api.example.com/v1/payment/health"
  protocol = "http"
}

# Database monitors
resource "hyperping_monitor" "prod_db_primary" {
  name     = "[PROD]-DB-Primary"
  url      = "db-primary.example.com"
  protocol = "port"
  port     = 5432
}

resource "hyperping_monitor" "prod_db_replica" {
  name     = "[PROD]-DB-Replica"
  url      = "db-replica.example.com"
  protocol = "port"
  port     = 5432
}

# Web application monitors
resource "hyperping_monitor" "prod_web_frontend" {
  name     = "[PROD]-Web-Frontend"
  url      = "https://www.example.com"
  protocol = "http"
}

resource "hyperping_monitor" "prod_web_admin" {
  name     = "[PROD]-Web-AdminPanel"
  url      = "https://admin.example.com"
  protocol = "http"
}

# Cron job monitors (healthchecks)
resource "hyperping_healthcheck" "prod_cron_backup" {
  name        = "[PROD]-Cron-DailyBackup"
  cron        = "0 2 * * *"
  timezone    = "America/New_York"
  http_method = "GET"
}
```

### Terraform Resource Naming

Use snake_case for Terraform resource names that mirror your monitor names:

```hcl
# Pattern: {environment}_{category}_{service}
resource "hyperping_monitor" "prod_api_payment" { ... }
resource "hyperping_monitor" "prod_db_primary" { ... }
resource "hyperping_monitor" "stg_web_frontend" { ... }
```

**Benefits:**
- Matches monitor name structure
- Easy to reference in dependencies
- Clear in Terraform state
- Searchable with `terraform state list`

### Variable Naming

```hcl
# variables.tf
variable "prod_api_urls" {
  description = "Production API URLs by service name"
  type        = map(string)
  default = {
    payment = "https://api.example.com/v1/payment/health"
    auth    = "https://api.example.com/v1/auth/health"
    graphql = "https://api.example.com/graphql"
  }
}

variable "environments" {
  description = "Environment configurations"
  type = map(object({
    code        = string
    api_base_url = string
    regions     = list(string)
  }))
  default = {
    production = {
      code         = "PROD"
      api_base_url = "https://api.example.com"
      regions      = ["virginia", "london", "singapore"]
    }
    staging = {
      code         = "STG"
      api_base_url = "https://staging-api.example.com"
      regions      = ["virginia"]
    }
  }
}
```

### Tag-Based Organization

If your organization uses tags, mirror the naming pattern:

```hcl
resource "hyperping_monitor" "prod_api_payment" {
  name     = "[PROD]-API-Payment"
  url      = "https://api.example.com/v1/payment/health"
  protocol = "http"

  # Tags for internal tracking (if supported by Hyperping)
  tags = {
    Environment = "production"
    Category    = "api"
    Service     = "payment"
    Team        = "payments-team"
    CostCenter  = "engineering"
    Criticality = "critical"
  }
}
```

### Advanced: Multi-Region Naming

For multi-region deployments, include region in the name:

```hcl
resource "hyperping_monitor" "prod_api_payment_us" {
  name     = "[PROD]-API-Payment-US"
  url      = "https://us.api.example.com/v1/payment/health"
  protocol = "http"
  regions  = ["virginia", "oregon"]
}

resource "hyperping_monitor" "prod_api_payment_eu" {
  name     = "[PROD]-API-Payment-EU"
  url      = "https://eu.api.example.com/v1/payment/health"
  protocol = "http"
  regions  = ["london", "frankfurt"]
}
```

### Advanced: Multi-Tenant Naming

For SaaS applications monitoring multiple tenants:

```hcl
resource "hyperping_monitor" "prod_api_tenant_acme" {
  name     = "[PROD]-API-Payment-TenantAcme"
  url      = "https://acme.api.example.com/health"
  protocol = "http"
}

resource "hyperping_monitor" "prod_api_tenant_globex" {
  name     = "[PROD]-API-Payment-TenantGlobex"
  url      = "https://globex.api.example.com/health"
  protocol = "http"
}
```

---

## State Management

Terraform state contains sensitive data and must be managed carefully. Follow these practices for reliable, secure state management.

### Use Remote State

**Never store state locally in production.** Always use a remote backend.

#### AWS S3 Backend (Recommended)

```hcl
# backend.tf
terraform {
  required_version = ">= 1.8"

  required_providers {
    hyperping = {
      source  = "develeap/hyperping"
      version = "~> 1.0"
    }
  }

  backend "s3" {
    bucket         = "my-company-terraform-state"
    key            = "hyperping/production/terraform.tfstate"
    region         = "us-east-1"
    encrypt        = true
    dynamodb_table = "terraform-state-lock"

    # Optional: Use versioning for rollback
    # Enable versioning on the S3 bucket
  }
}
```

**S3 Bucket Configuration:**

```hcl
# infrastructure/terraform-state-bucket/main.tf
resource "aws_s3_bucket" "terraform_state" {
  bucket = "my-company-terraform-state"

  lifecycle {
    prevent_destroy = true
  }
}

resource "aws_s3_bucket_versioning" "terraform_state" {
  bucket = aws_s3_bucket.terraform_state.id

  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "terraform_state" {
  bucket = aws_s3_bucket.terraform_state.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}

resource "aws_s3_bucket_public_access_block" "terraform_state" {
  bucket = aws_s3_bucket.terraform_state.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

resource "aws_dynamodb_table" "terraform_locks" {
  name         = "terraform-state-lock"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "LockID"

  attribute {
    name = "LockID"
    type = "S"
  }

  lifecycle {
    prevent_destroy = true
  }
}
```

#### Terraform Cloud Backend

```hcl
# backend.tf
terraform {
  cloud {
    organization = "my-company"

    workspaces {
      name = "hyperping-production"
    }
  }
}
```

**Benefits:**
- Built-in state locking
- Remote execution
- Team collaboration
- State versioning
- Run history
- Cost estimation

#### HashiCorp Consul Backend

```hcl
# backend.tf
terraform {
  backend "consul" {
    address = "consul.example.com:8500"
    scheme  = "https"
    path    = "terraform/hyperping/production"
    lock    = true
  }
}
```

### State Locking

State locking prevents concurrent modifications. Always enable locking.

**S3 + DynamoDB:**
```hcl
backend "s3" {
  dynamodb_table = "terraform-state-lock"  # Enables locking
}
```

**Terraform Cloud:**
- Locking enabled by default

**Consul:**
```hcl
backend "consul" {
  lock = true  # Enables locking
}
```

### Workspace Strategy

Use workspaces to manage multiple environments from the same configuration.

**Option 1: Separate Workspaces per Environment**

```bash
# Create workspaces
terraform workspace new production
terraform workspace new staging
terraform workspace new development

# Switch to production
terraform workspace select production
terraform apply

# Switch to staging
terraform workspace select staging
terraform apply
```

**Configuration with workspaces:**

```hcl
# main.tf
locals {
  environment = terraform.workspace

  env_config = {
    production = {
      code    = "PROD"
      regions = ["virginia", "london", "singapore"]
      frequency = 60
    }
    staging = {
      code    = "STG"
      regions = ["virginia"]
      frequency = 300
    }
    development = {
      code    = "DEV"
      regions = ["virginia"]
      frequency = 600
    }
  }

  current_env = local.env_config[local.environment]
}

resource "hyperping_monitor" "api_payment" {
  name            = "[${local.current_env.code}]-API-Payment"
  url             = "https://${local.environment}.api.example.com/health"
  protocol        = "http"
  check_frequency = local.current_env.frequency
  regions         = local.current_env.regions
}
```

**Option 2: Separate State Files per Environment**

```hcl
# production/backend.tf
terraform {
  backend "s3" {
    key = "hyperping/production/terraform.tfstate"
  }
}

# staging/backend.tf
terraform {
  backend "s3" {
    key = "hyperping/staging/terraform.tfstate"
  }
}
```

**Benefits:**
- Complete isolation between environments
- Different permissions per environment
- Independent apply/destroy operations
- Clear separation of concerns

### State Backup and Recovery

**Manual Backup:**

```bash
# Before major changes, backup state
terraform state pull > terraform.tfstate.backup-$(date +%Y%m%d-%H%M%S)

# Restore if needed
terraform state push terraform.tfstate.backup-20260213-143000
```

**Automated Backup (S3 Versioning):**

S3 bucket versioning provides automatic backups:

```bash
# List versions
aws s3api list-object-versions \
  --bucket my-company-terraform-state \
  --prefix hyperping/production/terraform.tfstate

# Restore previous version
aws s3api get-object \
  --bucket my-company-terraform-state \
  --key hyperping/production/terraform.tfstate \
  --version-id <version-id> \
  terraform.tfstate.restored
```

### State Cleanup

Remove obsolete resources from state:

```bash
# Remove individual resource
terraform state rm hyperping_monitor.old_api

# Remove all resources matching pattern
terraform state list | grep 'hyperping_monitor.old_' | xargs -n1 terraform state rm

# Remove entire module
terraform state rm module.deprecated_monitors
```

### State Migration

Migrating between backends:

```bash
# Step 1: Update backend configuration
# Edit backend.tf with new backend config

# Step 2: Reinitialize with migration
terraform init -migrate-state

# Step 3: Verify state
terraform plan
# Should show: "No changes. Infrastructure is up-to-date."
```

---

## Security Best Practices

Security is paramount when managing infrastructure as code. Follow these practices to protect your monitoring infrastructure and API credentials.

### API Key Management

**Never hardcode API keys in Terraform configuration.**

❌ **WRONG - Hardcoded Key:**
```hcl
provider "hyperping" {
  api_key = "sk_abc123def456..."  # NEVER DO THIS
}
```

✅ **CORRECT - Environment Variable:**
```bash
# Set in environment
export HYPERPING_API_KEY="sk_abc123def456..."

# Terraform uses it automatically
terraform plan
```

✅ **CORRECT - Variable with Sensitive Flag:**
```hcl
# variables.tf
variable "hyperping_api_key" {
  description = "Hyperping API key"
  type        = string
  sensitive   = true
}

# main.tf
provider "hyperping" {
  api_key = var.hyperping_api_key
}
```

```bash
# Set via command line
terraform apply -var="hyperping_api_key=sk_abc123..."

# Or via environment variable
export TF_VAR_hyperping_api_key="sk_abc123..."
terraform apply
```

### Secret Management Integration

#### AWS Secrets Manager

```hcl
# data.tf
data "aws_secretsmanager_secret_version" "hyperping_api_key" {
  secret_id = "production/hyperping/api-key"
}

# main.tf
provider "hyperping" {
  api_key = jsondecode(data.aws_secretsmanager_secret_version.hyperping_api_key.secret_string)["api_key"]
}
```

**Setup:**
```bash
# Store secret
aws secretsmanager create-secret \
  --name production/hyperping/api-key \
  --secret-string '{"api_key":"sk_abc123..."}'

# Update secret
aws secretsmanager update-secret \
  --secret-id production/hyperping/api-key \
  --secret-string '{"api_key":"sk_new456..."}'
```

#### HashiCorp Vault

```hcl
# data.tf
data "vault_generic_secret" "hyperping_api_key" {
  path = "secret/hyperping/production"
}

# main.tf
provider "hyperping" {
  api_key = data.vault_generic_secret.hyperping_api_key.data["api_key"]
}
```

**Setup:**
```bash
# Store secret
vault kv put secret/hyperping/production api_key="sk_abc123..."

# Read secret
vault kv get secret/hyperping/production
```

#### GitHub Actions Secrets

```yaml
# .github/workflows/terraform.yml
name: Terraform Apply

on:
  push:
    branches: [main]

jobs:
  terraform:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: hashicorp/setup-terraform@v3

      - name: Terraform Apply
        run: |
          terraform init
          terraform apply -auto-approve
        env:
          HYPERPING_API_KEY: ${{ secrets.HYPERPING_API_KEY }}
```

**Setup:**
1. Go to GitHub repository → Settings → Secrets and variables → Actions
2. Add secret: `HYPERPING_API_KEY` = `sk_abc123...`

#### GitLab CI/CD Variables

```yaml
# .gitlab-ci.yml
terraform:apply:
  script:
    - terraform init
    - terraform apply -auto-approve
  variables:
    HYPERPING_API_KEY: $HYPERPING_API_KEY
```

**Setup:**
1. Go to GitLab project → Settings → CI/CD → Variables
2. Add variable: `HYPERPING_API_KEY` = `sk_abc123...`
3. Check "Masked" and "Protected"

### Sensitive Data Handling

Mark sensitive outputs and variables:

```hcl
# variables.tf
variable "api_key_monitoring" {
  description = "API key for monitoring service"
  type        = string
  sensitive   = true  # Won't appear in logs
}

# outputs.tf
output "healthcheck_ping_url" {
  description = "Healthcheck ping URL (contains secret token)"
  value       = hyperping_healthcheck.backup.ping_url
  sensitive   = true  # Won't appear in plan/apply output
}

# To view sensitive output:
# terraform output -json | jq -r '.healthcheck_ping_url.value'
```

### Secret Scanning

Prevent accidental secret commits:

**Pre-commit Hook with git-secrets:**

```bash
# Install git-secrets
brew install git-secrets  # macOS
# or
apt-get install git-secrets  # Ubuntu

# Configure for repository
git secrets --install
git secrets --register-aws
git secrets --add 'sk_[a-zA-Z0-9]{40,}'  # Hyperping API key pattern

# Test
echo "HYPERPING_API_KEY=sk_abc123..." > test.txt
git add test.txt
# Error: secret detected!
```

**Pre-commit Framework:**

```yaml
# .pre-commit-config.yaml
repos:
  - repo: https://github.com/Yelp/detect-secrets
    rev: v1.4.0
    hooks:
      - id: detect-secrets
        args: ['--baseline', '.secrets.baseline']
        exclude: package.lock.json
```

```bash
# Install
pip install pre-commit
pre-commit install

# Generate baseline
detect-secrets scan > .secrets.baseline

# Test
git commit -m "test"
# Scans for secrets automatically
```

### Access Control

#### IAM Policies (AWS)

Restrict who can read Terraform state:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "s3:GetObject",
        "s3:PutObject"
      ],
      "Resource": "arn:aws:s3:::my-company-terraform-state/hyperping/production/*"
    },
    {
      "Effect": "Allow",
      "Action": [
        "dynamodb:GetItem",
        "dynamodb:PutItem",
        "dynamodb:DeleteItem"
      ],
      "Resource": "arn:aws:dynamodb:*:*:table/terraform-state-lock"
    }
  ]
}
```

#### Role-Based Access

```hcl
# Separate API keys per team/environment
# Production: Full access (only automation)
# Staging: Full access (engineering team)
# Development: Full access (all developers)

# CI/CD service account uses production key
# Developers use staging/dev keys for testing
```

### API Key Rotation

Rotate API keys regularly:

**Rotation Process:**

1. **Generate new API key** in Hyperping dashboard
2. **Update secret** in secret manager
3. **Test with new key** in non-production environment
4. **Deploy to production** via CI/CD
5. **Verify** with `terraform plan` (should show no changes)
6. **Revoke old key** after 7 days

**Automated Rotation Script:**

```bash
#!/bin/bash
# rotate-hyperping-key.sh

set -e

NEW_KEY="$1"
SECRET_NAME="production/hyperping/api-key"

echo "Updating Hyperping API key in AWS Secrets Manager..."

# Update secret
aws secretsmanager update-secret \
  --secret-id "$SECRET_NAME" \
  --secret-string "{\"api_key\":\"$NEW_KEY\"}"

echo "Testing new key..."

# Test with terraform plan
export HYPERPING_API_KEY="$NEW_KEY"
terraform plan -detailed-exitcode

echo "Key rotation successful!"
echo "Don't forget to revoke the old key in 7 days."
```

### Audit Logging

Enable audit logging for state changes:

**Terraform Cloud:**
- Automatically logs all runs
- View in UI: Organization → Settings → Audit Trails

**AWS CloudTrail (S3 State):**

```hcl
resource "aws_cloudtrail" "terraform_state_audit" {
  name           = "terraform-state-audit"
  s3_bucket_name = aws_s3_bucket.audit_logs.id

  event_selector {
    read_write_type           = "All"
    include_management_events = true

    data_resource {
      type   = "AWS::S3::Object"
      values = ["arn:aws:s3:::my-company-terraform-state/*"]
    }
  }
}
```

**Query audit logs:**

```bash
# Find who modified state in last 24 hours
aws cloudtrail lookup-events \
  --lookup-attributes AttributeKey=ResourceName,AttributeValue=my-company-terraform-state \
  --start-time "$(date -u -d '24 hours ago' '+%Y-%m-%dT%H:%M:%S')" \
  --end-time "$(date -u '+%Y-%m-%dT%H:%M:%S')"
```

---

## Resource Organization

Organize Terraform configuration for maintainability and scalability.

### File Structure for Small Deployments

For teams monitoring 1-20 services:

```
terraform/
├── main.tf              # Provider and monitors
├── variables.tf         # Input variables
├── outputs.tf           # Outputs
├── backend.tf           # Remote state config
├── terraform.tfvars     # Variable values (gitignored)
└── README.md            # Documentation
```

**Example main.tf:**

```hcl
terraform {
  required_version = ">= 1.8"
  required_providers {
    hyperping = {
      source  = "develeap/hyperping"
      version = "~> 1.0"
    }
  }
}

provider "hyperping" {
  api_key = var.hyperping_api_key
}

# Production monitors
resource "hyperping_monitor" "prod_api_payment" {
  name     = "[PROD]-API-Payment"
  url      = var.prod_api_urls.payment
  protocol = "http"
  regions  = ["virginia", "london", "singapore"]
}

resource "hyperping_monitor" "prod_api_auth" {
  name     = "[PROD]-API-Auth"
  url      = var.prod_api_urls.auth
  protocol = "http"
  regions  = ["virginia", "london", "singapore"]
}

resource "hyperping_monitor" "prod_db_primary" {
  name     = "[PROD]-DB-Primary"
  url      = var.prod_db_host
  protocol = "port"
  port     = 5432
}
```

### File Structure for Medium Deployments

For teams monitoring 20-100 services:

```
terraform/
├── backend.tf
├── providers.tf
├── variables.tf
├── outputs.tf
├── terraform.tfvars
├── monitors-api.tf      # API monitors
├── monitors-web.tf      # Web monitors
├── monitors-db.tf       # Database monitors
├── incidents.tf         # Incident management
├── maintenance.tf       # Maintenance windows
├── statuspage.tf        # Status pages
├── healthchecks.tf      # Cron job monitors
└── locals.tf            # Local values and data transformations
```

**Example monitors-api.tf:**

```hcl
# Production API Monitors
resource "hyperping_monitor" "prod_api_payment" {
  name            = "[PROD]-API-Payment"
  url             = "${var.prod_api_base_url}/v1/payment/health"
  protocol        = "http"
  check_frequency = 60
  regions         = local.prod_regions

  request_headers = {
    "X-Health-Check" = "terraform"
  }
}

resource "hyperping_monitor" "prod_api_auth" {
  name            = "[PROD]-API-Auth"
  url             = "${var.prod_api_base_url}/v1/auth/health"
  protocol        = "http"
  check_frequency = 60
  regions         = local.prod_regions
}

resource "hyperping_monitor" "prod_api_graphql" {
  name            = "[PROD]-API-GraphQL"
  url             = "${var.prod_api_base_url}/graphql"
  protocol        = "http"
  check_frequency = 60
  regions         = local.prod_regions
  request_method  = "POST"
  request_body    = jsonencode({
    query = "{ __typename }"
  })
}

# Staging API Monitors
resource "hyperping_monitor" "stg_api_payment" {
  name            = "[STG]-API-Payment"
  url             = "${var.stg_api_base_url}/v1/payment/health"
  protocol        = "http"
  check_frequency = 300
  regions         = ["virginia"]
}
```

### File Structure for Large Deployments

For teams monitoring 100+ services or multiple products:

```
terraform/
├── environments/
│   ├── production/
│   │   ├── backend.tf
│   │   ├── main.tf
│   │   ├── variables.tf
│   │   ├── terraform.tfvars
│   │   └── README.md
│   ├── staging/
│   │   ├── backend.tf
│   │   ├── main.tf
│   │   ├── variables.tf
│   │   └── terraform.tfvars
│   └── development/
│       ├── backend.tf
│       ├── main.tf
│       ├── variables.tf
│       └── terraform.tfvars
│
├── modules/
│   ├── api-monitor/
│   │   ├── main.tf
│   │   ├── variables.tf
│   │   ├── outputs.tf
│   │   └── README.md
│   ├── db-monitor/
│   │   ├── main.tf
│   │   ├── variables.tf
│   │   └── outputs.tf
│   ├── web-monitor/
│   │   ├── main.tf
│   │   ├── variables.tf
│   │   └── outputs.tf
│   └── statuspage/
│       ├── main.tf
│       ├── variables.tf
│       └── outputs.tf
│
└── README.md
```

**Example environment configuration:**

```hcl
# environments/production/main.tf
module "api_monitors" {
  source = "../../modules/api-monitor"

  environment         = "PROD"
  api_base_url        = "https://api.example.com"
  check_frequency     = 60
  regions             = ["virginia", "london", "singapore"]
  expected_status_code = "200"

  services = {
    payment = "/v1/payment/health"
    auth    = "/v1/auth/health"
    graphql = "/graphql"
    search  = "/v1/search/health"
  }
}

module "db_monitors" {
  source = "../../modules/db-monitor"

  environment = "PROD"
  databases = {
    primary = {
      host = "db-primary.example.com"
      port = 5432
    }
    replica_1 = {
      host = "db-replica-1.example.com"
      port = 5432
    }
    replica_2 = {
      host = "db-replica-2.example.com"
      port = 5432
    }
  }
}

module "statuspage" {
  source = "../../modules/statuspage"

  name             = "Production Status"
  subdomain        = "status"
  monitor_ids      = module.api_monitors.monitor_ids
  incident_webhook = var.slack_webhook_url
}
```

**Example module:**

```hcl
# modules/api-monitor/main.tf
variable "environment" {
  description = "Environment code (PROD, STG, DEV)"
  type        = string
}

variable "api_base_url" {
  description = "Base URL for API"
  type        = string
}

variable "services" {
  description = "Map of service name to health endpoint path"
  type        = map(string)
}

variable "check_frequency" {
  description = "Check frequency in seconds"
  type        = number
  default     = 60
}

variable "regions" {
  description = "Monitoring regions"
  type        = list(string)
  default     = ["virginia", "london"]
}

variable "expected_status_code" {
  description = "Expected HTTP status code"
  type        = string
  default     = "200"
}

resource "hyperping_monitor" "api" {
  for_each = var.services

  name                 = "[${var.environment}]-API-${title(each.key)}"
  url                  = "${var.api_base_url}${each.value}"
  protocol             = "http"
  check_frequency      = var.check_frequency
  regions              = var.regions
  expected_status_code = var.expected_status_code
}

output "monitor_ids" {
  description = "Map of service name to monitor UUID"
  value       = { for k, v in hyperping_monitor.api : k => v.id }
}

output "monitor_urls" {
  description = "Map of service name to Hyperping dashboard URL"
  value = {
    for k, v in hyperping_monitor.api :
    k => "https://app.hyperping.io/monitors/${v.id}"
  }
}
```

### Module Composition

Compose modules for complex scenarios:

```hcl
# environments/production/main.tf
module "payment_api" {
  source = "../../modules/api-monitor"

  environment      = "PROD"
  api_base_url     = "https://payment.example.com"
  check_frequency  = 30  # More frequent for critical service
  regions          = ["virginia", "london", "singapore", "tokyo"]

  services = {
    health   = "/health"
    ready    = "/ready"
    process  = "/v1/process"
    refund   = "/v1/refund"
  }
}

module "payment_db" {
  source = "../../modules/db-monitor"

  environment = "PROD"
  databases = {
    primary   = { host = "payment-db-1.example.com", port = 5432 }
    replica_1 = { host = "payment-db-2.example.com", port = 5432 }
  }
}

module "payment_statuspage" {
  source = "../../modules/statuspage"

  name        = "Payment Service Status"
  subdomain   = "payment-status"
  monitor_ids = concat(
    values(module.payment_api.monitor_ids),
    values(module.payment_db.monitor_ids)
  )
}

module "payment_incident_workflow" {
  source = "../../modules/incident-management"

  statuspage_id   = module.payment_statuspage.id
  slack_webhook   = var.payment_team_slack_webhook
  pagerduty_key   = var.payment_team_pagerduty_key
  escalation_time = 300  # 5 minutes
}
```

### Data-Driven Configuration

Use YAML or JSON to define monitors:

**monitors.yaml:**
```yaml
production:
  api:
    - name: Payment
      path: /v1/payment/health
      method: GET
      frequency: 60
      critical: true
    - name: Auth
      path: /v1/auth/health
      method: GET
      frequency: 60
      critical: true
    - name: GraphQL
      path: /graphql
      method: POST
      body: '{"query":"{ __typename }"}'
      frequency: 60
      critical: false

  databases:
    - name: Primary
      host: db-primary.example.com
      port: 5432
      critical: true
    - name: Replica
      host: db-replica.example.com
      port: 5432
      critical: false

staging:
  api:
    - name: Payment
      path: /v1/payment/health
      method: GET
      frequency: 300
      critical: false
```

**Load in Terraform:**

```hcl
# main.tf
locals {
  monitors_config = yamldecode(file("${path.module}/monitors.yaml"))

  prod_api_monitors = {
    for m in local.monitors_config.production.api :
    lower(m.name) => m
  }
}

resource "hyperping_monitor" "prod_api" {
  for_each = local.prod_api_monitors

  name            = "[PROD]-API-${each.value.name}"
  url             = "${var.prod_api_base_url}${each.value.path}"
  protocol        = "http"
  request_method  = each.value.method
  check_frequency = each.value.frequency

  request_body = try(each.value.body, null)

  regions = each.value.critical ? ["virginia", "london", "singapore"] : ["virginia"]
}
```

---

## CI/CD Integration

Automate Terraform workflows with continuous integration and deployment.

### GitHub Actions

Complete workflow for production deployments:

```yaml
# .github/workflows/terraform.yml
name: Terraform Hyperping Infrastructure

on:
  push:
    branches:
      - main
    paths:
      - 'terraform/**'
  pull_request:
    branches:
      - main
    paths:
      - 'terraform/**'
  workflow_dispatch:

env:
  TF_VERSION: "1.8.0"
  TF_WORKING_DIR: "./terraform"

jobs:
  validate:
    name: Validate Configuration
    runs-on: ubuntu-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Setup Terraform
        uses: hashicorp/setup-terraform@v3
        with:
          terraform_version: ${{ env.TF_VERSION }}

      - name: Terraform Format Check
        working-directory: ${{ env.TF_WORKING_DIR }}
        run: terraform fmt -check -recursive

      - name: Terraform Init
        working-directory: ${{ env.TF_WORKING_DIR }}
        run: terraform init -backend=false

      - name: Terraform Validate
        working-directory: ${{ env.TF_WORKING_DIR }}
        run: terraform validate

  plan:
    name: Plan Changes
    runs-on: ubuntu-latest
    needs: validate
    if: github.event_name == 'pull_request'
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Setup Terraform
        uses: hashicorp/setup-terraform@v3
        with:
          terraform_version: ${{ env.TF_VERSION }}

      - name: Terraform Init
        working-directory: ${{ env.TF_WORKING_DIR }}
        run: terraform init
        env:
          HYPERPING_API_KEY: ${{ secrets.HYPERPING_API_KEY }}

      - name: Terraform Plan
        working-directory: ${{ env.TF_WORKING_DIR }}
        run: |
          terraform plan -no-color -out=tfplan
          terraform show -no-color tfplan > plan.txt
        env:
          HYPERPING_API_KEY: ${{ secrets.HYPERPING_API_KEY }}

      - name: Comment Plan on PR
        uses: actions/github-script@v7
        with:
          script: |
            const fs = require('fs');
            const plan = fs.readFileSync('${{ env.TF_WORKING_DIR }}/plan.txt', 'utf8');
            const output = `#### Terraform Plan
            <details><summary>Show Plan</summary>

            \`\`\`
            ${plan}
            \`\`\`

            </details>

            **Pusher:** @${{ github.actor }}
            **Action:** \`${{ github.event_name }}\``;

            github.rest.issues.createComment({
              issue_number: context.issue.number,
              owner: context.repo.owner,
              repo: context.repo.repo,
              body: output
            });

  apply:
    name: Apply Changes
    runs-on: ubuntu-latest
    needs: validate
    if: github.ref == 'refs/heads/main' && github.event_name == 'push'
    environment:
      name: production
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Setup Terraform
        uses: hashicorp/setup-terraform@v3
        with:
          terraform_version: ${{ env.TF_VERSION }}

      - name: Terraform Init
        working-directory: ${{ env.TF_WORKING_DIR }}
        run: terraform init
        env:
          HYPERPING_API_KEY: ${{ secrets.HYPERPING_API_KEY }}

      - name: Terraform Apply
        working-directory: ${{ env.TF_WORKING_DIR }}
        run: terraform apply -auto-approve -parallelism=3
        env:
          HYPERPING_API_KEY: ${{ secrets.HYPERPING_API_KEY }}

      - name: Notify Success
        if: success()
        run: |
          echo "✅ Terraform apply completed successfully"
          # Add Slack notification here

      - name: Notify Failure
        if: failure()
        run: |
          echo "❌ Terraform apply failed"
          # Add Slack/PagerDuty alert here

  drift-detection:
    name: Detect Configuration Drift
    runs-on: ubuntu-latest
    if: github.event_name == 'schedule'
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Setup Terraform
        uses: hashicorp/setup-terraform@v3
        with:
          terraform_version: ${{ env.TF_VERSION }}

      - name: Terraform Init
        working-directory: ${{ env.TF_WORKING_DIR }}
        run: terraform init
        env:
          HYPERPING_API_KEY: ${{ secrets.HYPERPING_API_KEY }}

      - name: Terraform Plan (Drift Detection)
        working-directory: ${{ env.TF_WORKING_DIR }}
        run: |
          terraform plan -detailed-exitcode > /dev/null
          EXIT_CODE=$?
          if [ $EXIT_CODE -eq 2 ]; then
            echo "⚠️ Configuration drift detected!"
            terraform plan -no-color > drift.txt
            # Send alert
            exit 1
          else
            echo "✅ No drift detected"
          fi
        env:
          HYPERPING_API_KEY: ${{ secrets.HYPERPING_API_KEY }}
```

**Schedule drift detection:**

```yaml
on:
  schedule:
    - cron: '0 */6 * * *'  # Every 6 hours
```

### GitLab CI/CD

```yaml
# .gitlab-ci.yml
stages:
  - validate
  - plan
  - apply

variables:
  TF_VERSION: "1.8.0"
  TF_ROOT: ${CI_PROJECT_DIR}/terraform

.terraform_base:
  image:
    name: hashicorp/terraform:${TF_VERSION}
    entrypoint: [""]
  before_script:
    - cd ${TF_ROOT}
    - terraform --version

validate:
  extends: .terraform_base
  stage: validate
  script:
    - terraform fmt -check -recursive
    - terraform init -backend=false
    - terraform validate
  rules:
    - if: '$CI_PIPELINE_SOURCE == "merge_request_event"'
    - if: '$CI_COMMIT_BRANCH == "main"'

plan:
  extends: .terraform_base
  stage: plan
  script:
    - terraform init
    - terraform plan -out=tfplan
    - terraform show -json tfplan > plan.json
  artifacts:
    paths:
      - ${TF_ROOT}/tfplan
      - ${TF_ROOT}/plan.json
    expire_in: 1 day
  rules:
    - if: '$CI_PIPELINE_SOURCE == "merge_request_event"'

apply:
  extends: .terraform_base
  stage: apply
  script:
    - terraform init
    - terraform apply -auto-approve -parallelism=3
  environment:
    name: production
  rules:
    - if: '$CI_COMMIT_BRANCH == "main"'
      when: manual
  dependencies:
    - plan
```

### Pre-Apply Validation

Add custom validation before applying:

```bash
#!/bin/bash
# scripts/pre-apply-validate.sh

set -e

echo "Running pre-apply validation..."

# Check for critical resources
CRITICAL_MONITORS=$(terraform state list | grep -c 'hyperping_monitor.*prod.*payment' || true)
if [ "$CRITICAL_MONITORS" -eq 0 ]; then
  echo "⚠️  Warning: No production payment monitors found!"
fi

# Check for monitors without regions
terraform show -json | jq -r '
  .values.root_module.resources[]
  | select(.type == "hyperping_monitor")
  | select(.values.regions | length == 0)
  | .address
' | while read -r monitor; do
  echo "⚠️  Warning: Monitor $monitor has no regions configured"
done

# Check for high-frequency monitors (potential cost concern)
terraform show -json | jq -r '
  .values.root_module.resources[]
  | select(.type == "hyperping_monitor")
  | select(.values.check_frequency < 60)
  | "\(.address): \(.values.check_frequency)s"
' | while read -r line; do
  echo "💰 Cost alert: High-frequency monitor detected: $line"
done

echo "✅ Pre-apply validation complete"
```

**Integrate in CI:**

```yaml
- name: Pre-Apply Validation
  run: bash scripts/pre-apply-validate.sh
```

### Drift Detection

Automated drift detection with alerts:

```bash
#!/bin/bash
# scripts/drift-detection.sh

set -e

echo "Checking for configuration drift..."

# Run terraform plan with -detailed-exitcode
# Exit code 0 = no changes
# Exit code 1 = error
# Exit code 2 = changes detected (drift)
terraform plan -detailed-exitcode -no-color > plan.txt
EXIT_CODE=$?

if [ $EXIT_CODE -eq 2 ]; then
  echo "⚠️ Configuration drift detected!"

  # Send Slack notification
  curl -X POST "$SLACK_WEBHOOK_URL" \
    -H 'Content-Type: application/json' \
    -d "{
      \"text\": \":warning: Terraform drift detected in Hyperping infrastructure\",
      \"blocks\": [
        {
          \"type\": \"section\",
          \"text\": {
            \"type\": \"mrkdwn\",
            \"text\": \"*Terraform Drift Alert*\n\nConfiguration drift detected in production Hyperping monitors.\n\n<$CI_PIPELINE_URL|View Pipeline>\"
          }
        }
      ]
    }"

  # Upload plan to S3 for review
  aws s3 cp plan.txt "s3://my-company-terraform-drift/$(date +%Y%m%d-%H%M%S)-plan.txt"

  exit 2
elif [ $EXIT_CODE -eq 0 ]; then
  echo "✅ No drift detected - infrastructure matches code"
  exit 0
else
  echo "❌ Error running terraform plan"
  exit 1
fi
```

**Schedule in cron:**

```bash
# crontab -e
0 */6 * * * cd /path/to/terraform && /path/to/scripts/drift-detection.sh
```

### Cost Estimation

Estimate costs before applying changes (useful for understanding check frequency impact):

```bash
#!/bin/bash
# scripts/cost-estimate.sh

echo "Estimating monitoring costs..."

# Count monitors by frequency
terraform show -json | jq -r '
  .values.root_module.resources[]
  | select(.type == "hyperping_monitor")
  | "\(.values.check_frequency // 60)"
' | sort | uniq -c | while read -r count frequency; do
  checks_per_day=$((86400 / frequency))
  total_checks=$((count * checks_per_day))
  echo "- $count monitors at ${frequency}s = $total_checks checks/day"
done

# Count total checks per day
TOTAL_CHECKS=$(terraform show -json | jq -r '
  [
    .values.root_module.resources[]
    | select(.type == "hyperping_monitor")
    | (86400 / (.values.check_frequency // 60))
  ] | add
')

echo ""
echo "Total daily checks: $TOTAL_CHECKS"
echo "Monthly checks: $((TOTAL_CHECKS * 30))"
```

---

## Testing and Validation

Ensure your Terraform configuration is correct before applying to production.

### Native Terraform Tests

Use Terraform's built-in testing framework (Terraform 1.6+):

```hcl
# tests/monitor_creation.tftest.hcl
run "create_monitor" {
  command = plan

  assert {
    condition     = hyperping_monitor.test.name == "[TEST]-API-Sample"
    error_message = "Monitor name should follow naming convention"
  }

  assert {
    condition     = hyperping_monitor.test.protocol == "http"
    error_message = "Protocol should be http"
  }

  assert {
    condition     = length(hyperping_monitor.test.regions) >= 2
    error_message = "Monitors should use at least 2 regions for redundancy"
  }
}

run "validate_check_frequency" {
  command = plan

  assert {
    condition = contains(
      [10, 20, 30, 60, 120, 180, 300, 600, 1800, 3600],
      hyperping_monitor.test.check_frequency
    )
    error_message = "Check frequency must be a valid value"
  }
}
```

**Run tests:**

```bash
terraform test
```

### Contract Tests

Verify API behavior matches expectations:

```bash
#!/bin/bash
# tests/contract-tests.sh

set -e

echo "Running Hyperping API contract tests..."

API_KEY="${HYPERPING_API_KEY}"
BASE_URL="https://api.hyperping.io"

# Test 1: Create monitor
echo "Test 1: Create monitor..."
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/v1/monitors" \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "[TEST]-Contract-Test",
    "url": "https://httpstat.us/200",
    "protocol": "http",
    "check_frequency": 3600
  }')

HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')

if [ "$HTTP_CODE" -ne 201 ]; then
  echo "❌ Test 1 failed: Expected 201, got $HTTP_CODE"
  echo "$BODY"
  exit 1
fi

MONITOR_ID=$(echo "$BODY" | jq -r '.id // .monitorUuid // .uuid')
echo "✅ Test 1 passed: Created monitor $MONITOR_ID"

# Test 2: Read monitor
echo "Test 2: Read monitor..."
RESPONSE=$(curl -s -w "\n%{http_code}" -X GET "$BASE_URL/v1/monitors/$MONITOR_ID" \
  -H "Authorization: Bearer $API_KEY")

HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
if [ "$HTTP_CODE" -ne 200 ]; then
  echo "❌ Test 2 failed: Expected 200, got $HTTP_CODE"
  exit 1
fi

echo "✅ Test 2 passed: Read monitor"

# Test 3: Update monitor
echo "Test 3: Update monitor..."
RESPONSE=$(curl -s -w "\n%{http_code}" -X PUT "$BASE_URL/v1/monitors/$MONITOR_ID" \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "[TEST]-Contract-Test-Updated",
    "check_frequency": 1800
  }')

HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
if [ "$HTTP_CODE" -ne 200 ]; then
  echo "❌ Test 3 failed: Expected 200, got $HTTP_CODE"
  exit 1
fi

echo "✅ Test 3 passed: Updated monitor"

# Test 4: Delete monitor
echo "Test 4: Delete monitor..."
RESPONSE=$(curl -s -w "\n%{http_code}" -X DELETE "$BASE_URL/v1/monitors/$MONITOR_ID" \
  -H "Authorization: Bearer $API_KEY")

HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
if [ "$HTTP_CODE" -ne 204 ] && [ "$HTTP_CODE" -ne 200 ]; then
  echo "❌ Test 4 failed: Expected 204/200, got $HTTP_CODE"
  exit 1
fi

echo "✅ Test 4 passed: Deleted monitor"

echo "✅ All contract tests passed"
```

### Plan Validation

Validate plan output before applying:

```bash
#!/bin/bash
# scripts/validate-plan.sh

set -e

echo "Validating Terraform plan..."

# Generate plan in JSON format
terraform plan -out=tfplan
terraform show -json tfplan > plan.json

# Check for destructive changes
DESTROYS=$(jq -r '
  [
    .resource_changes[]?
    | select(.change.actions[] == "delete")
    | select(.address | startswith("hyperping_monitor"))
  ] | length
' plan.json)

if [ "$DESTROYS" -gt 0 ]; then
  echo "⚠️  Warning: Plan will destroy $DESTROYS monitors"
  echo "Destroyed monitors:"
  jq -r '
    .resource_changes[]?
    | select(.change.actions[] == "delete")
    | select(.address | startswith("hyperping_monitor"))
    | "  - \(.address): \(.change.before.name)"
  ' plan.json

  read -p "Continue? (yes/no): " CONFIRM
  if [ "$CONFIRM" != "yes" ]; then
    echo "Aborted"
    exit 1
  fi
fi

# Check for monitors without regions
NO_REGIONS=$(jq -r '
  [
    .resource_changes[]?
    | select(.type == "hyperping_monitor")
    | select(.change.after.regions | length == 0)
  ] | length
' plan.json)

if [ "$NO_REGIONS" -gt 0 ]; then
  echo "❌ Error: $NO_REGIONS monitors have no regions configured"
  jq -r '
    .resource_changes[]?
    | select(.type == "hyperping_monitor")
    | select(.change.after.regions | length == 0)
    | "  - \(.address)"
  ' plan.json
  exit 1
fi

# Check for naming convention violations
INVALID_NAMES=$(jq -r '
  [
    .resource_changes[]?
    | select(.type == "hyperping_monitor")
    | select(.change.after.name | test("^\\[(PROD|STG|DEV|TEST|DEMO)\\]-") | not)
  ] | length
' plan.json)

if [ "$INVALID_NAMES" -gt 0 ]; then
  echo "⚠️  Warning: $INVALID_NAMES monitors don't follow naming convention"
  jq -r '
    .resource_changes[]?
    | select(.type == "hyperping_monitor")
    | select(.change.after.name | test("^\\[(PROD|STG|DEV|TEST|DEMO)\\]-") | not)
    | "  - \(.address): \(.change.after.name)"
  ' plan.json
fi

echo "✅ Plan validation complete"
```

### Integration Tests

Test complete workflows:

```hcl
# tests/integration/main.tf
terraform {
  required_providers {
    hyperping = {
      source = "develeap/hyperping"
    }
  }
}

provider "hyperping" {
  api_key = var.api_key
}

variable "api_key" {
  type      = string
  sensitive = true
}

# Create test monitor
resource "hyperping_monitor" "integration_test" {
  name            = "[TEST]-Integration-${formatdate("YYYYMMDDhhmmss", timestamp())}"
  url             = "https://httpstat.us/200"
  protocol        = "http"
  check_frequency = 3600
  regions         = ["virginia"]
}

# Create test statuspage
resource "hyperping_statuspage" "integration_test" {
  name             = "Integration Test Status"
  hosted_subdomain = "integration-test-${formatdate("YYYYMMDDhhmmss", timestamp())}"

  settings = {
    name      = "Settings"
    languages = ["en"]
  }

  sections = [{
    name     = { en = "Services" }
    is_split = false
    services = [{
      monitor_uuid        = hyperping_monitor.integration_test.id
      show_uptime         = true
      show_response_times = false
    }]
  }]
}

# Verify monitor was created
output "monitor_id" {
  value = hyperping_monitor.integration_test.id
}

output "monitor_url" {
  value = "https://app.hyperping.io/monitors/${hyperping_monitor.integration_test.id}"
}

output "statuspage_url" {
  value = "https://${hyperping_statuspage.integration_test.hosted_subdomain}.hyperping.app"
}
```

**Run integration test:**

```bash
#!/bin/bash
# tests/integration/run.sh

set -e

cd tests/integration

echo "Running integration tests..."

# Apply
terraform init
terraform apply -auto-approve -var="api_key=$HYPERPING_API_KEY"

# Extract outputs
MONITOR_ID=$(terraform output -raw monitor_id)
STATUSPAGE_URL=$(terraform output -raw statuspage_url)

echo "Monitor created: $MONITOR_ID"
echo "Status page: $STATUSPAGE_URL"

# Verify via API
echo "Verifying monitor via API..."
curl -f -H "Authorization: Bearer $HYPERPING_API_KEY" \
  "https://api.hyperping.io/v1/monitors/$MONITOR_ID" > /dev/null

echo "✅ Monitor verified via API"

# Verify status page is accessible
echo "Verifying status page..."
curl -f "$STATUSPAGE_URL" > /dev/null

echo "✅ Status page verified"

# Cleanup
echo "Cleaning up..."
terraform destroy -auto-approve -var="api_key=$HYPERPING_API_KEY"

echo "✅ Integration tests passed"
```

---

## Anti-Patterns to Avoid

Learn from common mistakes.

### 1. Hardcoding API Keys

❌ **WRONG:**
```hcl
provider "hyperping" {
  api_key = "sk_abc123def456..."  # Committed to git!
}
```

✅ **CORRECT:**
```bash
export HYPERPING_API_KEY="sk_abc123..."
```

### 2. No State Locking

❌ **WRONG:**
```hcl
terraform {
  backend "s3" {
    bucket = "my-state"
    key    = "terraform.tfstate"
    # No DynamoDB table = no locking!
  }
}
```

✅ **CORRECT:**
```hcl
terraform {
  backend "s3" {
    bucket         = "my-state"
    key            = "terraform.tfstate"
    dynamodb_table = "terraform-locks"
  }
}
```

### 3. Inconsistent Naming

❌ **WRONG:**
```hcl
resource "hyperping_monitor" "prod_api" {
  name = "Production API"  # No environment prefix
}

resource "hyperping_monitor" "payment_monitor" {
  name = "payment-api"  # Different format
}
```

✅ **CORRECT:**
```hcl
resource "hyperping_monitor" "prod_api_payment" {
  name = "[PROD]-API-Payment"
}

resource "hyperping_monitor" "prod_api_auth" {
  name = "[PROD]-API-Auth"
}
```

### 4. Single Region Monitoring

❌ **WRONG:**
```hcl
resource "hyperping_monitor" "critical_api" {
  name    = "[PROD]-API-Critical"
  url     = "https://api.example.com"
  protocol = "http"
  regions = ["virginia"]  # Only one region!
}
```

✅ **CORRECT:**
```hcl
resource "hyperping_monitor" "critical_api" {
  name    = "[PROD]-API-Critical"
  url     = "https://api.example.com"
  protocol = "http"
  regions = ["virginia", "london", "singapore"]  # Multiple regions
}
```

### 5. No Output Documentation

❌ **WRONG:**
```hcl
output "id" {
  value = hyperping_monitor.api.id
}
```

✅ **CORRECT:**
```hcl
output "api_monitor_id" {
  description = "UUID of the API health monitor"
  value       = hyperping_monitor.api.id
}

output "api_monitor_dashboard_url" {
  description = "Link to monitor in Hyperping dashboard"
  value       = "https://app.hyperping.io/monitors/${hyperping_monitor.api.id}"
}
```

### 6. Overly High Check Frequency

❌ **WRONG:**
```hcl
resource "hyperping_monitor" "api" {
  name            = "[PROD]-API-Payment"
  url             = "https://api.example.com"
  protocol        = "http"
  check_frequency = 10  # Expensive! 8,640 checks/day
}
```

✅ **CORRECT:**
```hcl
resource "hyperping_monitor" "api" {
  name            = "[PROD]-API-Payment"
  url             = "https://api.example.com"
  protocol        = "http"
  check_frequency = 60  # Reasonable: 1,440 checks/day
}
```

### 7. No Lifecycle Protection for Critical Resources

❌ **WRONG:**
```hcl
resource "hyperping_monitor" "production_api" {
  name = "[PROD]-API-Critical"
  # No lifecycle protection
}
```

✅ **CORRECT:**
```hcl
resource "hyperping_monitor" "production_api" {
  name = "[PROD]-API-Critical"

  lifecycle {
    prevent_destroy = true
  }
}
```

### 8. Monolithic Configuration Files

❌ **WRONG:**
```hcl
# main.tf (5000+ lines)
# All monitors, incidents, maintenance, statuspages in one file
```

✅ **CORRECT:**
```
terraform/
├── monitors-api.tf
├── monitors-web.tf
├── monitors-db.tf
├── incidents.tf
├── maintenance.tf
└── statuspage.tf
```

### 9. No Validation in CI/CD

❌ **WRONG:**
```yaml
jobs:
  deploy:
    steps:
      - run: terraform apply -auto-approve
```

✅ **CORRECT:**
```yaml
jobs:
  validate:
    steps:
      - run: terraform fmt -check
      - run: terraform validate

  plan:
    needs: validate
    steps:
      - run: terraform plan

  apply:
    needs: plan
    steps:
      - run: terraform apply -auto-approve
```

### 10. Not Using Modules for Repeated Patterns

❌ **WRONG:**
```hcl
# Copy-pasted 50 times with slight variations
resource "hyperping_monitor" "api_1" {
  name     = "[PROD]-API-Service1"
  url      = "https://api.example.com/service1"
  protocol = "http"
  regions  = ["virginia", "london"]
}

resource "hyperping_monitor" "api_2" {
  name     = "[PROD]-API-Service2"
  url      = "https://api.example.com/service2"
  protocol = "http"
  regions  = ["virginia", "london"]
}
# ... repeated 48 more times
```

✅ **CORRECT:**
```hcl
module "api_monitors" {
  source = "./modules/api-monitor"

  environment = "PROD"
  services = {
    service1 = "/service1"
    service2 = "/service2"
    # ... 48 more
  }
}
```

---

## Production Checklist

Use this checklist before deploying to production.

### Configuration Review

- [ ] **Naming convention**: All monitors follow `[ENV]-[CATEGORY]-[SERVICE]` pattern
- [ ] **API keys**: No hardcoded keys, using environment variables or secret manager
- [ ] **State backend**: Remote state configured (S3, Terraform Cloud, Consul)
- [ ] **State locking**: DynamoDB table or equivalent configured
- [ ] **Regions**: Critical monitors use multiple regions (at least 3)
- [ ] **Check frequency**: Appropriate for service criticality (60s for critical, 300s+ for non-critical)
- [ ] **Lifecycle protection**: Critical resources have `prevent_destroy = true`
- [ ] **Validation**: All configuration passes `terraform validate`
- [ ] **Format**: All files formatted with `terraform fmt`

### Security Review

- [ ] **API key storage**: Keys stored in AWS Secrets Manager, Vault, or equivalent
- [ ] **API key rotation**: Process documented and tested
- [ ] **State encryption**: S3 bucket has encryption enabled
- [ ] **State access**: IAM policies restrict state bucket access
- [ ] **Secret scanning**: Pre-commit hooks prevent secret commits
- [ ] **Audit logging**: CloudTrail or equivalent enabled for state bucket
- [ ] **Sensitive outputs**: Marked with `sensitive = true`

### Testing Review

- [ ] **Validation tests**: `terraform validate` passes
- [ ] **Plan review**: `terraform plan` output reviewed for unexpected changes
- [ ] **Integration tests**: Tested in staging environment first
- [ ] **Contract tests**: API behavior validated
- [ ] **Rollback plan**: Documented process for reverting changes

### CI/CD Review

- [ ] **Automated validation**: CI runs `terraform validate` on every PR
- [ ] **Automated planning**: CI runs `terraform plan` and comments on PRs
- [ ] **Manual approval**: Production applies require manual approval
- [ ] **Drift detection**: Scheduled job checks for drift (every 6 hours)
- [ ] **Notifications**: Slack/email alerts for apply success/failure
- [ ] **Parallelism**: Set appropriately to avoid rate limits (`-parallelism=3`)

### Documentation Review

- [ ] **README**: Explains purpose, setup, and usage
- [ ] **Variables**: All variables documented with descriptions
- [ ] **Outputs**: All outputs documented with descriptions
- [ ] **Runbook**: Operational procedures documented
- [ ] **Disaster recovery**: State restore process documented
- [ ] **Team access**: Team members know where to find documentation

### Operations Review

- [ ] **Monitoring**: Monitors themselves are monitored (meta-monitoring)
- [ ] **Alerting**: Failures trigger appropriate alerts
- [ ] **Escalation**: Clear escalation path for failures
- [ ] **On-call**: Team members know how to respond to alerts
- [ ] **Maintenance windows**: Documented process for scheduling maintenance
- [ ] **Incident response**: Process for creating and managing incidents

### Cost Review

- [ ] **Check frequency**: Reviewed to balance cost vs. detection time
- [ ] **Number of regions**: Critical services use multiple regions, non-critical use fewer
- [ ] **Unused monitors**: Old monitors deleted or paused
- [ ] **Cost estimation**: Monthly cost calculated and approved

### Compliance Review

- [ ] **Change management**: Changes go through approval process
- [ ] **Audit trail**: All changes tracked in version control
- [ ] **Access control**: Only authorized personnel can apply changes
- [ ] **Data residency**: Monitor regions comply with data residency requirements
- [ ] **Retention**: State backups retained per compliance requirements

---

## Summary

Following these best practices ensures:

✅ **Maintainability** - Clear naming and organization
✅ **Security** - Protected credentials and encrypted state
✅ **Reliability** - Multi-region monitoring and state locking
✅ **Scalability** - Modular architecture for growth
✅ **Auditability** - Version control and audit logs
✅ **Automation** - CI/CD for consistent deployments
✅ **Quality** - Comprehensive testing and validation

**Next Steps:**

1. Review your existing configuration against this guide
2. Implement missing practices (start with security)
3. Document your specific patterns and conventions
4. Train team members on best practices
5. Schedule regular reviews to maintain standards

---

## Additional Resources

- [Naming Conventions Deep Dive](./naming-conventions.md)
- [State Management Guide](./state-management.md)
- [Security Hardening Checklist](./security.md)
- [CI/CD Pipeline Examples](./cicd-examples.md)
- [Testing Strategies](./testing-strategies.md)
- [Terraform Best Practices](https://developer.hashicorp.com/terraform/cloud-docs/recommended-practices)
- [Hyperping API Documentation](https://hyperping.io/docs/api)

---

**Need help implementing these practices?** Open an issue on [GitHub](https://github.com/develeap/terraform-provider-hyperping/issues) or contact your team's infrastructure lead.
