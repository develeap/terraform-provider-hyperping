# Complete Example - Hyperping Terraform Provider
#
# This example demonstrates all features of the Hyperping provider:
# - Provider configuration
# - Multiple monitors with different configurations
# - Data source to query existing monitors
# - Incident management linked to status pages
# - Maintenance window scheduling
#
# Prerequisites:
#   export HYPERPING_API_KEY="hp_your_api_key"
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

# Configure the Hyperping provider
# API key is read from HYPERPING_API_KEY environment variable
provider "hyperping" {}

# =============================================================================
# Variables
# =============================================================================

variable "environment" {
  description = "Environment name (e.g., production, staging)"
  type        = string
  default     = "production"
}

variable "api_token" {
  description = "API token for authenticated health checks"
  type        = string
  sensitive   = true
  default     = ""
}

variable "status_page_id" {
  description = "Status page UUID for incidents and maintenance"
  type        = string
  default     = "your-status-page-uuid"
}

# =============================================================================
# Monitors
# =============================================================================

# Basic website monitor
resource "hyperping_monitor" "website" {
  name                 = "Website - ${var.environment}"
  url                  = "https://example.com"
  protocol             = "http"
  http_method          = "GET"
  check_frequency      = 60
  expected_status_code = "200"
  follow_redirects     = true

  regions = ["london", "virginia", "singapore", "tokyo"]
}

# API health check with authentication
resource "hyperping_monitor" "api_health" {
  name                 = "API Health - ${var.environment}"
  url                  = "https://api.example.com/health"
  protocol             = "http"
  http_method          = "GET"
  check_frequency      = 30
  expected_status_code = "200"

  regions = ["london", "frankfurt", "virginia", "sydney"]

  request_headers = var.api_token != "" ? [
    {
      name  = "Accept"
      value = "application/json"
    },
    {
      name  = "Authorization"
      value = "Bearer ${var.api_token}"
    }
  ] : [
    {
      name  = "Accept"
      value = "application/json"
    }
  ]
}

# API endpoint with POST request
resource "hyperping_monitor" "api_endpoint" {
  name                 = "API Create Endpoint - ${var.environment}"
  url                  = "https://api.example.com/v1/ping"
  protocol             = "http"
  http_method          = "POST"
  check_frequency      = 300
  expected_status_code = "200"
  follow_redirects     = false

  regions = ["london", "virginia"]

  request_headers = [
    {
      name  = "Content-Type"
      value = "application/json"
    },
    {
      name  = "Accept"
      value = "application/json"
    }
  ]

  request_body = jsonencode({
    test      = true
    timestamp = "check"
  })
}

# Database connectivity monitor (can be paused during maintenance)
resource "hyperping_monitor" "database" {
  name                 = "Database Health - ${var.environment}"
  url                  = "https://api.example.com/db/health"
  protocol             = "http"
  http_method          = "GET"
  check_frequency      = 120
  expected_status_code = "200"
  paused               = false

  regions = ["virginia", "sydney"]
}

# =============================================================================
# Data Source - Query All Monitors
# =============================================================================

# Fetch all monitors to use in other configurations
data "hyperping_monitors" "all" {
  depends_on = [
    hyperping_monitor.website,
    hyperping_monitor.api_health,
    hyperping_monitor.api_endpoint,
    hyperping_monitor.database,
  ]
}

# =============================================================================
# Incidents
# =============================================================================

# Example: Declare an incident for status page
resource "hyperping_incident" "api_degradation" {
  title        = "API Performance Monitoring"
  text         = "Monitoring API response times after recent deployment."
  type         = "incident"
  status_pages = [var.status_page_id]
}

# Example: Add an update to the incident
resource "hyperping_incident_update" "investigating" {
  incident_id = hyperping_incident.api_degradation.id
  text        = "We have identified the cause and are implementing a fix."
  type        = "identified"
}

# =============================================================================
# Maintenance Windows
# =============================================================================

# Scheduled maintenance window
resource "hyperping_maintenance" "database_upgrade" {
  name  = "database-maintenance-${var.environment}"
  title = "Database Maintenance Window"
  text  = "Scheduled database maintenance for performance optimization and security updates."

  # Schedule for future date (adjust as needed)
  start_date = "2026-02-01T02:00:00.000Z"
  end_date   = "2026-02-01T04:00:00.000Z"

  # Affects database-related monitors
  monitors = [
    hyperping_monitor.database.id,
    hyperping_monitor.api_health.id,
  ]

  status_pages         = [var.status_page_id]
  notification_option  = "scheduled"
  notification_minutes = 60
}

# =============================================================================
# Outputs
# =============================================================================

output "website_monitor" {
  description = "Website monitor details"
  value = {
    id       = hyperping_monitor.website.id
    name     = hyperping_monitor.website.name
    url      = hyperping_monitor.website.url
    protocol = hyperping_monitor.website.protocol
    status   = hyperping_monitor.website.paused ? "PAUSED" : "ACTIVE"
  }
}

output "api_monitors" {
  description = "API monitor IDs"
  value = {
    health   = hyperping_monitor.api_health.id
    endpoint = hyperping_monitor.api_endpoint.id
    database = hyperping_monitor.database.id
  }
}

output "total_monitors" {
  description = "Total number of monitors in account"
  value       = length(data.hyperping_monitors.all.monitors)
}

output "active_monitors" {
  description = "List of active (non-paused) monitor names"
  value = [
    for m in data.hyperping_monitors.all.monitors : m.name
    if !m.paused
  ]
}

output "monitors_by_status" {
  description = "Monitor counts by status"
  value = {
    total  = length(data.hyperping_monitors.all.monitors)
    active = length([for m in data.hyperping_monitors.all.monitors : m if !m.paused])
    paused = length([for m in data.hyperping_monitors.all.monitors : m if m.paused])
  }
}

output "incident_info" {
  description = "Incident tracking information"
  value = {
    id   = hyperping_incident.api_degradation.id
    type = hyperping_incident.api_degradation.type
    date = hyperping_incident.api_degradation.date
  }
}

output "maintenance_info" {
  description = "Maintenance window information"
  value = {
    id         = hyperping_maintenance.database_upgrade.id
    name       = hyperping_maintenance.database_upgrade.name
    start_date = hyperping_maintenance.database_upgrade.start_date
    end_date   = hyperping_maintenance.database_upgrade.end_date
  }
}
