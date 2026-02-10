# Advanced Patterns - Hyperping Terraform Provider
#
# This example demonstrates advanced monitoring patterns and best practices:
# - Regional health checks with fallback
# - Environment-based configurations
# - Dynamic monitor creation from data
# - Conditional alerting strategies
# - Integration with status pages and incidents
#
# Prerequisites:
#   export HYPERPING_API_KEY="sk_your_api_key"
#
# Usage:
#   terraform init
#   terraform plan
#   terraform apply

terraform {
  required_version = ">= 1.0"

  required_providers {
    hyperping = {
      source  = "develeap/hyperping"
      version = "~> 1.0"
    }
  }
}

provider "hyperping" {}

# =============================================================================
# Variables and Configuration
# =============================================================================

variable "environment" {
  description = "Environment name"
  type        = string
  default     = "production"
}

variable "services" {
  description = "Map of services to monitor"
  type = map(object({
    url             = string
    method          = string
    frequency       = number
    expected_status = string
    critical        = bool
    regions         = list(string)
  }))
  default = {
    api = {
      url             = "https://api.example.com/health"
      method          = "GET"
      frequency       = 30
      expected_status = "200"
      critical        = true
      regions         = ["london", "virginia", "singapore", "tokyo"]
    }
    website = {
      url             = "https://www.example.com"
      method          = "GET"
      frequency       = 60
      expected_status = "2xx"
      critical        = false
      regions         = ["london", "virginia"]
    }
    admin = {
      url             = "https://admin.example.com"
      method          = "GET"
      frequency       = 300
      expected_status = "200"
      critical        = false
      regions         = ["virginia"]
    }
  }
}

variable "status_page_id" {
  description = "Status page UUID for incident management"
  type        = string
  default     = ""
}

# =============================================================================
# Local Values and Computed Configuration
# =============================================================================

locals {
  # Name prefix with environment tag
  name_prefix = "[${upper(var.environment)}]"

  # Critical services for priority monitoring
  critical_services = {
    for k, v in var.services : k => v if v.critical
  }

  # Non-critical services
  standard_services = {
    for k, v in var.services : k => v if !v.critical
  }

  # Default regions by tier
  tier1_regions = ["london", "virginia", "tokyo"]      # Critical
  tier2_regions = ["frankfurt", "singapore"]            # Important
  tier3_regions = ["tokyo", "sydney"]                  # Standard

  # Common request headers
  json_headers = [
    { name = "Accept", value = "application/json" },
    { name = "Content-Type", value = "application/json" }
  ]
}

# =============================================================================
# Pattern 1: Dynamic Monitor Creation from Data Structure
# =============================================================================

# Create monitors dynamically from services map
resource "hyperping_monitor" "services" {
  for_each = var.services

  name                 = "${local.name_prefix} ${title(each.key)} Service"
  url                  = each.value.url
  protocol             = "http"
  http_method          = each.value.method
  check_frequency      = each.value.frequency
  expected_status_code = each.value.expected_status
  follow_redirects     = true

  # Use service-specific regions or defaults based on criticality
  regions = length(each.value.regions) > 0 ? each.value.regions : (
    each.value.critical ? local.tier1_regions : local.tier2_regions
  )

  # Add JSON headers for API endpoints
  request_headers = can(regex("/api/", each.value.url)) ? local.json_headers : []
}

# =============================================================================
# Pattern 2: Regional Redundancy Checks
# =============================================================================

# Primary region check (US East)
resource "hyperping_monitor" "regional_primary" {
  name                 = "${local.name_prefix} Regional Check - US East"
  url                  = "https://us-east.api.example.com/health"
  protocol             = "http"
  http_method          = "GET"
  check_frequency      = 30
  expected_status_code = "200"
  regions              = ["virginia"]
}

# Secondary region check (EU)
resource "hyperping_monitor" "regional_secondary" {
  name                 = "${local.name_prefix} Regional Check - EU West"
  url                  = "https://eu-west.api.example.com/health"
  protocol             = "http"
  http_method          = "GET"
  check_frequency      = 30
  expected_status_code = "200"
  regions              = ["london"]
}

# Tertiary region check (Asia Pacific)
resource "hyperping_monitor" "regional_tertiary" {
  name                 = "${local.name_prefix} Regional Check - Asia Pacific"
  url                  = "https://ap-south.api.example.com/health"
  protocol             = "http"
  http_method          = "GET"
  check_frequency      = 30
  expected_status_code = "200"
  regions              = ["singapore"]
}

# =============================================================================
# Pattern 3: Multi-Protocol Monitoring
# =============================================================================

# HTTP endpoint
resource "hyperping_monitor" "http_endpoint" {
  name                 = "${local.name_prefix} HTTP API"
  url                  = "https://api.example.com/v1/status"
  protocol             = "http"
  http_method          = "GET"
  check_frequency      = 60
  expected_status_code = "200"
}

# TCP port check (database)
resource "hyperping_monitor" "database_port" {
  name            = "${local.name_prefix} Database Port"
  url             = "tcp://db.example.com:5432"
  protocol        = "port"
  port            = 5432
  check_frequency = 120
  regions         = ["virginia", "tokyo"]
}

# TCP port check (Redis)
resource "hyperping_monitor" "cache_port" {
  name            = "${local.name_prefix} Redis Cache"
  url             = "tcp://cache.example.com:6379"
  protocol        = "port"
  port            = 6379
  check_frequency = 60
  regions         = ["virginia"]
}

# =============================================================================
# Pattern 4: Authenticated Endpoint Monitoring
# =============================================================================

# API with Bearer token
resource "hyperping_monitor" "authenticated_api" {
  name                 = "${local.name_prefix} Authenticated API"
  url                  = "https://api.example.com/v1/private/health"
  protocol             = "http"
  http_method          = "GET"
  check_frequency      = 120
  expected_status_code = "200"

  request_headers = [
    { name = "Authorization", value = "Bearer ${var.environment}_health_check_token" },
    { name = "Accept", value = "application/json" }
  ]
}

# API with custom authentication
resource "hyperping_monitor" "custom_auth_api" {
  name                 = "${local.name_prefix} Custom Auth API"
  url                  = "https://internal.example.com/health"
  protocol             = "http"
  http_method          = "GET"
  check_frequency      = 300
  expected_status_code = "200"

  request_headers = [
    { name = "X-API-Key", value = "monitoring_key_${var.environment}" },
    { name = "X-Environment", value = var.environment }
  ]
}

# =============================================================================
# Pattern 5: Data Source Queries and Filtering
# =============================================================================

# Query all monitors
data "hyperping_monitors" "all" {
  depends_on = [
    hyperping_monitor.services,
    hyperping_monitor.regional_primary,
    hyperping_monitor.http_endpoint,
  ]
}

# =============================================================================
# Pattern 6: Conditional Incident Management
# =============================================================================

# Create incident only if status page is configured
resource "hyperping_incident" "service_degradation" {
  count = var.status_page_id != "" ? 1 : 0

  title        = "Service Performance Monitoring"
  text         = "Monitoring service performance after deployment in ${var.environment}."
  type         = "incident"
  status_pages = [var.status_page_id]
}

# =============================================================================
# Pattern 7: Scheduled Maintenance with Dynamic Monitors
# =============================================================================

# Maintenance window for critical services only
resource "hyperping_maintenance" "critical_upgrade" {
  count = var.status_page_id != "" ? 1 : 0

  name  = "critical-services-maintenance-${var.environment}"
  title = "Critical Services Maintenance"
  text  = "Scheduled maintenance for critical infrastructure components."

  start_date = "2026-02-15T02:00:00.000Z"
  end_date   = "2026-02-15T04:00:00.000Z"

  # Only affect critical service monitors
  monitors = [
    for k, v in hyperping_monitor.services : v.id
    if var.services[k].critical
  ]

  status_pages         = [var.status_page_id]
  notification_option  = "scheduled"
  notification_minutes = 120 # Notify 2 hours in advance
}

# =============================================================================
# Outputs with Advanced Filtering and Analysis
# =============================================================================

output "environment_summary" {
  description = "High-level environment summary"
  value = {
    environment      = var.environment
    total_monitors   = length(var.services) + 6 # services + additional monitors
    critical_count   = length(local.critical_services)
    standard_count   = length(local.standard_services)
    status_page_id   = var.status_page_id
  }
}

output "service_monitors" {
  description = "Dynamically created service monitors"
  value = {
    for k, v in hyperping_monitor.services : k => {
      id        = v.id
      name      = v.name
      url       = v.url
      frequency = v.check_frequency
      regions   = v.regions
      critical  = var.services[k].critical
    }
  }
}

output "critical_monitors" {
  description = "Monitor IDs for critical services"
  value = [
    for k, v in hyperping_monitor.services : v.id
    if var.services[k].critical
  ]
}

output "regional_monitors" {
  description = "Regional redundancy monitor IDs"
  value = {
    us_east      = hyperping_monitor.regional_primary.id
    eu_west      = hyperping_monitor.regional_secondary.id
    asia_pacific = hyperping_monitor.regional_tertiary.id
  }
}

output "monitor_by_frequency" {
  description = "Monitors grouped by check frequency"
  value = {
    high_frequency = [
      for k, v in hyperping_monitor.services : {
        name      = v.name
        frequency = v.check_frequency
      }
      if v.check_frequency <= 60
    ]
    standard_frequency = [
      for k, v in hyperping_monitor.services : {
        name      = v.name
        frequency = v.check_frequency
      }
      if v.check_frequency > 60 && v.check_frequency <= 300
    ]
    low_frequency = [
      for k, v in hyperping_monitor.services : {
        name      = v.name
        frequency = v.check_frequency
      }
      if v.check_frequency > 300
    ]
  }
}

output "regional_coverage" {
  description = "Unique regions used across all monitors"
  value = distinct(flatten([
    for k, v in hyperping_monitor.services : v.regions
  ]))
}

output "maintenance_window" {
  description = "Scheduled maintenance window details"
  value = var.status_page_id != "" ? {
    id                   = hyperping_maintenance.critical_upgrade[0].id
    name                 = hyperping_maintenance.critical_upgrade[0].name
    start_date           = hyperping_maintenance.critical_upgrade[0].start_date
    end_date             = hyperping_maintenance.critical_upgrade[0].end_date
    affected_monitors    = length(hyperping_maintenance.critical_upgrade[0].monitors)
    notification_minutes = hyperping_maintenance.critical_upgrade[0].notification_minutes
  } : null
}

output "monitor_urls_by_criticality" {
  description = "Monitor URLs grouped by criticality"
  value = {
    critical = [
      for k, v in hyperping_monitor.services : v.url
      if var.services[k].critical
    ]
    standard = [
      for k, v in hyperping_monitor.services : v.url
      if !var.services[k].critical
    ]
  }
}

# Cost estimation output (approximate)
output "estimated_monthly_checks" {
  description = "Estimated total checks per month"
  value = sum([
    for k, v in hyperping_monitor.services :
    (2592000 / v.check_frequency) * length(v.regions) # 30 days in seconds
  ])
}
