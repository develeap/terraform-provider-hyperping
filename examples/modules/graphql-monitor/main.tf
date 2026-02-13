# GraphQL Monitor Module - Main
#
# Creates POST monitors for GraphQL endpoint health checking with query validation.
#
# NOTE: This module is designed for public GraphQL endpoints or endpoints where
# authentication can be handled via custom (non-reserved) headers. The Hyperping
# provider blocks reserved headers (Authorization, Cookie, Host, etc.) for security.
#
# Usage:
#   module "graphql_monitors" {
#     source = "path/to/modules/graphql-monitor"
#
#     endpoint = "https://api.example.com/graphql"
#     queries = {
#       "health" = {
#         query = "{ health { status } }"
#         expected_response = "\"status\":\"ok\""
#       }
#     }
#   }

locals {
  # Build the full name with format string
  name_format = var.name_format != "" ? var.name_format : (var.name_prefix != "" ? "[${upper(var.name_prefix)}] GraphQL - %s" : "GraphQL - %s")

  # Build headers for GraphQL requests (Content-Type + custom headers)
  request_headers = merge(
    { "Content-Type" = "application/json" },
    var.custom_headers != null ? var.custom_headers : {}
  )

  # Build GraphQL request body for each query
  query_bodies = {
    for key, config in var.queries : key => jsonencode({
      query     = config.query
      variables = config.variables != null ? config.variables : null
    })
  }
}

resource "hyperping_monitor" "graphql_query" {
  for_each = var.queries

  name     = format(local.name_format, each.key)
  url      = var.endpoint
  protocol = "http"

  http_method          = "POST"
  check_frequency      = coalesce(each.value.frequency, var.default_frequency)
  expected_status_code = coalesce(each.value.expected_status_code, "200")
  follow_redirects     = coalesce(each.value.follow_redirects, true)
  regions              = coalesce(each.value.regions, var.default_regions)
  paused               = coalesce(each.value.paused, var.paused)
  alerts_wait          = var.alerts_wait

  # Validate response contains expected content
  required_keyword = each.value.expected_response

  # Optional: escalation policy
  escalation_policy = var.escalation_policy

  # Request headers (Content-Type + custom headers only)
  request_headers = [
    for k, v in local.request_headers : {
      name  = k
      value = v
    }
  ]

  # GraphQL query as JSON request body
  request_body = local.query_bodies[each.key]

  lifecycle {
    create_before_destroy = true
  }
}

# Optional: Introspection check monitor (if enabled)
resource "hyperping_monitor" "introspection" {
  count = var.enable_introspection_check ? 1 : 0

  name     = format(local.name_format, "introspection")
  url      = var.endpoint
  protocol = "http"

  http_method          = "POST"
  check_frequency      = var.introspection_frequency
  expected_status_code = var.introspection_expected_status
  follow_redirects     = true
  regions              = var.default_regions
  paused               = var.paused
  alerts_wait          = var.alerts_wait

  # Check that introspection query returns valid response
  required_keyword = var.introspection_expected_response

  escalation_policy = var.escalation_policy

  request_headers = [
    for k, v in local.request_headers : {
      name  = k
      value = v
    }
  ]

  # Standard GraphQL introspection query
  request_body = jsonencode({
    query = var.introspection_query
  })

  lifecycle {
    create_before_destroy = true
  }
}
