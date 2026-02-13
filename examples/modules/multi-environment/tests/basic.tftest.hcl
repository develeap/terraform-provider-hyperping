# Multi-Environment Module - Basic Tests

provider "hyperping" {
  api_key = "test_mock_api_key_for_plan_only"
}

variables {
  service_name = "TestAPI"
  environments = {
    dev = {
      url = "https://dev-api.example.com/health"
    }
    staging = {
      url = "https://staging-api.example.com/health"
    }
    prod = {
      url = "https://api.example.com/health"
    }
  }
}

run "creates_monitors_for_all_environments" {
  command = plan

  assert {
    condition     = length(hyperping_monitor.environment) == 3
    error_message = "Should create 3 monitors (one per environment)"
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
    condition     = contains(keys(hyperping_monitor.environment), "prod")
    error_message = "Should create prod environment monitor"
  }
}

run "applies_correct_naming_convention" {
  command = plan

  assert {
    condition     = hyperping_monitor.environment["dev"].name == "[DEV] TestAPI"
    error_message = "Dev monitor should have correct name format"
  }

  assert {
    condition     = hyperping_monitor.environment["staging"].name == "[STAGING] TestAPI"
    error_message = "Staging monitor should have correct name format"
  }

  assert {
    condition     = hyperping_monitor.environment["prod"].name == "[PROD] TestAPI"
    error_message = "Prod monitor should have correct name format"
  }
}

run "applies_correct_urls" {
  command = plan

  assert {
    condition     = hyperping_monitor.environment["dev"].url == "https://dev-api.example.com/health"
    error_message = "Dev monitor should have correct URL"
  }

  assert {
    condition     = hyperping_monitor.environment["staging"].url == "https://staging-api.example.com/health"
    error_message = "Staging monitor should have correct URL"
  }

  assert {
    condition     = hyperping_monitor.environment["prod"].url == "https://api.example.com/health"
    error_message = "Prod monitor should have correct URL"
  }
}

run "applies_default_settings" {
  command = plan

  assert {
    condition     = hyperping_monitor.environment["dev"].http_method == "GET"
    error_message = "Default HTTP method should be GET"
  }

  assert {
    condition     = hyperping_monitor.environment["dev"].check_frequency == 60
    error_message = "Default frequency should be 60 seconds"
  }

  assert {
    condition     = hyperping_monitor.environment["dev"].expected_status_code == "2xx"
    error_message = "Default expected status should be 2xx"
  }

  assert {
    condition     = hyperping_monitor.environment["dev"].follow_redirects == true
    error_message = "Default follow_redirects should be true"
  }

  assert {
    condition     = hyperping_monitor.environment["dev"].paused == false
    error_message = "Default paused state should be false"
  }
}

run "outputs_contain_all_environments" {
  command = plan

  assert {
    condition     = length(keys(output.monitor_ids)) == 3
    error_message = "monitor_ids output should contain all 3 environments"
  }

  assert {
    condition     = length(output.monitor_ids_list) == 3
    error_message = "monitor_ids_list output should contain 3 UUIDs"
  }

  assert {
    condition     = output.environment_count == 3
    error_message = "environment_count should be 3"
  }

  assert {
    condition     = output.service_name == "TestAPI"
    error_message = "service_name output should match input"
  }
}
