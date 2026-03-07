# Basic status page example
resource "hyperping_statuspage" "basic" {
  name             = "My Status Page"
  hosted_subdomain = "status"

  settings = {
    name      = "My Status Page"
    languages = ["en"]
  }
}

# Advanced status page with all features
resource "hyperping_statuspage" "production" {
  name             = "Production Status"
  hosted_subdomain = "prod-status"

  # Optional: Use custom domain instead of hosted subdomain
  # hostname = "status.example.com"

  # Optional: Password protect the page
  # password = "secret"

  settings = {
    name             = "Production Status"
    website          = "https://example.com"
    languages        = ["en", "fr"]
    default_language = "en"

    # Theme and branding
    theme        = "dark"     # Options: system, light, dark
    font         = "Inter"    # Options: Inter, Roboto, Poppins, Lato, etc.
    accent_color = "#0066cc"  # Brand color (hex)

    # Multi-language description
    description = "Production system status and uptime"

    # Subscription settings
    subscribe = {
      enabled = true
      email   = true
      sms     = true
      slack   = false # Configured via Hyperping OAuth
      teams   = true
    }

    # Authentication settings
    authentication = {
      password_protection = false
      google_sso          = true
      allowed_domains     = ["example.com", "partner.com"]
    }
  }

  # Status page sections with monitors
  sections = [
    {
      name = {
        en = "Core API Services"
        fr = "Services API principaux"
      }
      is_split = true # Show individual service status
      services = [
        {
          uuid = hyperping_monitor.api.id
          name = {
            en = "Main API"
          }
          show_uptime         = true
          show_response_times = true
        },
        {
          uuid = hyperping_monitor.auth.id
          name = {
            en = "Authentication API"
          }
          show_uptime         = true
          show_response_times = false
        }
      ]
    },
    {
      name = {
        en = "Infrastructure"
        fr = "Infrastructure"
      }
      is_split = false # Show aggregated status
      services = [
        {
          uuid     = hyperping_monitor.database.id
          is_group = true
          name = {
            en = "Database Cluster"
          }
          services = [
            {
              uuid = hyperping_monitor.db_primary.id
              name = {
                en = "Primary DB"
              }
            },
            {
              uuid = hyperping_monitor.db_replica.id
              name = {
                en = "Replica DB"
              }
            }
          ]
        }
      ]
    }
  ]
}

# Example monitors (referenced in status page)
resource "hyperping_monitor" "api" {
  name                 = "Production API"
  url                  = "https://api.example.com/health"
  protocol             = "http"
  http_method          = "GET"
  check_frequency      = 60
  expected_status_code = "2xx"
}

resource "hyperping_monitor" "auth" {
  name            = "Auth Service"
  url             = "https://auth.example.com/health"
  protocol        = "http"
  http_method     = "GET"
  check_frequency = 60
}

resource "hyperping_monitor" "database" {
  name            = "Database Health"
  url             = "https://db.example.com/health"
  protocol        = "http"
  http_method     = "GET"
  check_frequency = 300
}

resource "hyperping_monitor" "db_primary" {
  name            = "DB Primary"
  url             = "tcp://db-primary.example.com:5432"
  protocol        = "port"
  port            = 5432
  check_frequency = 60
}

resource "hyperping_monitor" "db_replica" {
  name            = "DB Replica"
  url             = "tcp://db-replica.example.com:5432"
  protocol        = "port"
  port            = 5432
  check_frequency = 60
}

# Output the status page URL
output "status_page_url" {
  value       = hyperping_statuspage.production.url
  description = "Public URL of the status page"
}

output "status_page_id" {
  value       = hyperping_statuspage.production.id
  description = "UUID of the status page for use with subscribers"
}
