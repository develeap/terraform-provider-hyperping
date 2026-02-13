# Multi-Environment Module - Main
#
# Deploys identical monitors across multiple environments (dev, staging, prod)
# with environment-specific configuration for frequency, regions, and alerts.
#
# Usage:
#   module "api_monitoring" {
#     source = "path/to/modules/multi-environment"
#
#     service_name = "UserAPI"
#     environments = {
#       dev = {
#         url       = "https://dev-api.example.com"
#         frequency = 300
#         regions   = ["virginia"]
#       }
#       staging = {
#         url       = "https://staging-api.example.com"
#         frequency = 120
#         regions   = ["virginia", "london"]
#       }
#       prod = {
#         url       = "https://api.example.com"
#         frequency = 60
#         regions   = ["virginia", "london", "singapore"]
#       }
#     }
#   }

locals {
  # Flatten environments into a map keyed by "environment-name"
  # This enables for_each on resources
  monitors = {
    for env_name, env_config in var.environments : env_name => merge(
      {
        name                 = var.use_workspace_name ? terraform.workspace : env_name
        url                  = env_config.url
        method               = coalesce(env_config.method, var.default_method)
        frequency            = coalesce(env_config.frequency, var.default_frequency)
        regions              = coalesce(env_config.regions, var.default_regions)
        expected_status_code = coalesce(env_config.expected_status_code, var.default_expected_status_code)
        follow_redirects     = coalesce(env_config.follow_redirects, var.default_follow_redirects)
        headers              = coalesce(env_config.headers, var.default_headers)
        body                 = coalesce(env_config.body, var.default_body)
        required_keyword     = coalesce(env_config.required_keyword, var.default_required_keyword)
        alerts_wait          = coalesce(env_config.alerts_wait, var.default_alerts_wait)
        escalation_policy    = coalesce(env_config.escalation_policy, var.default_escalation_policy)
        paused               = coalesce(env_config.paused, var.default_paused)
        enabled              = coalesce(env_config.enabled, true)
      },
      env_config
    )
  }

  # Filter to only enabled environments
  enabled_monitors = {
    for env_name, config in local.monitors : env_name => config
    if config.enabled
  }

  # Build monitor name using configurable format
  name_format = var.name_format != "" ? var.name_format : "[%s] %s"
}

resource "hyperping_monitor" "environment" {
  for_each = local.enabled_monitors

  name     = format(local.name_format, upper(each.value.name), var.service_name)
  url      = each.value.url
  protocol = "http"

  http_method          = each.value.method
  check_frequency      = each.value.frequency
  expected_status_code = each.value.expected_status_code
  follow_redirects     = each.value.follow_redirects
  regions              = each.value.regions
  paused               = each.value.paused
  alerts_wait          = each.value.alerts_wait

  # Optional: required keyword for content validation
  required_keyword = each.value.required_keyword

  # Optional: escalation policy
  escalation_policy = each.value.escalation_policy

  # Request headers as list of objects (convert from map)
  request_headers = each.value.headers != null ? [
    for k, v in each.value.headers : {
      name  = k
      value = v
    }
  ] : null

  # Optional request body (for POST/PUT/PATCH methods)
  request_body = each.value.body

  lifecycle {
    create_before_destroy = true
  }
}
