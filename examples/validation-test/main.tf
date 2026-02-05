# Validation Test Examples
# This file contains valid configurations that should pass terraform validate

terraform {
  required_providers {
    hyperping = {
      source = "develeap/hyperping"
    }
  }
}

provider "hyperping" {
  # API key via environment variable HYPERPING_API_KEY
}

# Example 1: Valid monitor with all required fields
resource "hyperping_monitor" "api" {
  name     = "API Monitor"
  url      = "https://api.example.com/health"
  protocol = "http"
}

# Example 2: Valid monitor with optional fields
resource "hyperping_monitor" "api_full" {
  name            = "API Monitor Full"
  url             = "https://api.example.com/health"
  protocol        = "http"
  check_frequency = 60
  regions         = ["virginia", "london"]
}

# Example 3: Valid TCP monitor
resource "hyperping_monitor" "database" {
  name     = "Database"
  url      = "db.example.com"
  protocol = "port"
  port     = 5432
}

# Example 4: Valid statuspage (required for incidents)
resource "hyperping_statuspage" "main" {
  name             = "Status Page"
  hosted_subdomain = "status-example"

  settings = {
    name      = "Status Page Settings"
    languages = ["en"]
  }
}

# Example 5: Valid incident referencing statuspage
resource "hyperping_incident" "api_outage" {
  title        = "API Outage"
  text         = "We are investigating API connectivity issues."
  type         = "incident"
  status_pages = [hyperping_statuspage.main.id]
}
