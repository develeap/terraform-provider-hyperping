# Basic HTTP monitor - just URL and name
resource "hyperping_monitor" "basic" {
  name     = "My Website"
  url      = "https://example.com"
  protocol = "http"
}

# Full-featured HTTP monitor with all options
resource "hyperping_monitor" "api_health" {
  name                 = "API Health Check"
  url                  = "https://api.example.com/health"
  protocol             = "http"
  http_method          = "POST"
  check_frequency      = 300 # 5 minutes
  expected_status_code = "201"
  follow_redirects     = false

  regions = ["london", "virginia", "singapore"]

  request_headers = [
    {
      name  = "Content-Type"
      value = "application/json"
    },
    {
      name  = "Authorization"
      value = "Bearer ${var.api_token}"
    }
  ]

  request_body = jsonencode({
    check = "health"
  })
}

# Monitor with pause capability
resource "hyperping_monitor" "maintenance" {
  name     = "Service Under Maintenance"
  url      = "https://maintenance.example.com"
  protocol = "http"
  paused   = true
}
