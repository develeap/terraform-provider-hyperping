# Multi-Environment Module - Selective Deployment Tests

provider "hyperping" {
  api_key = "test_mock_api_key_for_plan_only"
}

variables {
  service_name = "SelectiveAPI"

  environments = {
    dev = {
      url     = "https://dev-api.example.com/health"
      enabled = true
    }
    staging = {
      url     = "https://staging-api.example.com/health"
      enabled = true
    }
    prod = {
      url     = "https://api.example.com/health"
      enabled = false # Disabled
    }
  }
}

run "only_creates_enabled_environments" {
  command = plan

  assert {
    condition     = length(hyperping_monitor.environment) == 2
    error_message = "Should only create 2 monitors (prod is disabled)"
  }

  assert {
    condition     = contains(keys(hyperping_monitor.environment), "dev")
    error_message = "Should create dev environment monitor"
  }

  assert {
    condition     = contains(keys(hyperping_monitor.environment), "staging")
    error_message = "Should create staging environment monitor"
  }

  assert {
    condition     = !contains(keys(hyperping_monitor.environment), "prod")
    error_message = "Should NOT create prod environment monitor"
  }
}

run "outputs_only_include_enabled_environments" {
  command = plan

  assert {
    condition     = length(keys(output.monitor_ids)) == 2
    error_message = "monitor_ids should only contain 2 enabled environments"
  }

  assert {
    condition     = output.environment_count == 2
    error_message = "environment_count should be 2 (only enabled)"
  }

  assert {
    condition     = length(output.environments) == 2
    error_message = "environments list should contain 2 items"
  }
}

run "production_monitor_ids_empty_when_disabled" {
  command = plan

  assert {
    condition     = length(output.production_monitor_ids) == 0
    error_message = "production_monitor_ids should be empty when prod is disabled"
  }
}

run "non_production_monitor_ids_contains_enabled_only" {
  command = plan

  assert {
    condition     = length(output.non_production_monitor_ids) == 2
    error_message = "non_production_monitor_ids should contain 2 monitors (dev and staging)"
  }
}
