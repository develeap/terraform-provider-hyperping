# Multi-Environment Module - Custom Configuration Tests

provider "hyperping" {
  api_key = "test_mock_api_key_for_plan_only"
}

variables {
  service_name = "PaymentAPI"
  name_format  = "[%s] %s Health Check"

  environments = {
    dev = {
      url       = "https://dev-payments.example.com/health"
      frequency = 300
      regions   = ["virginia"]
      paused    = true
    }
    staging = {
      url              = "https://staging-payments.example.com/health"
      frequency        = 120
      regions          = ["virginia", "london"]
      required_keyword = "healthy"
      alerts_wait      = 60
    }
    prod = {
      url               = "https://payments.example.com/health"
      frequency         = 30
      regions           = ["virginia", "london", "singapore"]
      alerts_wait       = 0
      escalation_policy = "pol_test123"
    }
  }

  default_method               = "POST"
  default_expected_status_code = "200"
  default_headers = {
    "Content-Type" = "application/json"
  }
}

run "applies_custom_name_format" {
  command = plan

  assert {
    condition     = hyperping_monitor.environment["dev"].name == "[DEV] PaymentAPI Health Check"
    error_message = "Monitor name should use custom format"
  }

  assert {
    condition     = hyperping_monitor.environment["staging"].name == "[STAGING] PaymentAPI Health Check"
    error_message = "Monitor name should use custom format"
  }
}

run "applies_environment_specific_frequency" {
  command = plan

  assert {
    condition     = hyperping_monitor.environment["dev"].check_frequency == 300
    error_message = "Dev should check every 300 seconds"
  }

  assert {
    condition     = hyperping_monitor.environment["staging"].check_frequency == 120
    error_message = "Staging should check every 120 seconds"
  }

  assert {
    condition     = hyperping_monitor.environment["prod"].check_frequency == 30
    error_message = "Prod should check every 30 seconds"
  }
}

run "applies_environment_specific_regions" {
  command = plan

  assert {
    condition     = length(hyperping_monitor.environment["dev"].regions) == 1
    error_message = "Dev should have 1 region"
  }

  assert {
    condition     = contains(hyperping_monitor.environment["dev"].regions, "virginia")
    error_message = "Dev should monitor from Virginia"
  }

  assert {
    condition     = length(hyperping_monitor.environment["staging"].regions) == 2
    error_message = "Staging should have 2 regions"
  }

  assert {
    condition     = length(hyperping_monitor.environment["prod"].regions) == 3
    error_message = "Prod should have 3 regions"
  }
}

run "applies_default_method_to_all_environments" {
  command = plan

  assert {
    condition     = hyperping_monitor.environment["dev"].http_method == "POST"
    error_message = "Dev should use POST method from defaults"
  }

  assert {
    condition     = hyperping_monitor.environment["staging"].http_method == "POST"
    error_message = "Staging should use POST method from defaults"
  }

  assert {
    condition     = hyperping_monitor.environment["prod"].http_method == "POST"
    error_message = "Prod should use POST method from defaults"
  }
}

run "applies_default_expected_status_code" {
  command = plan

  assert {
    condition     = hyperping_monitor.environment["dev"].expected_status_code == "200"
    error_message = "Should use custom default expected status code"
  }
}

run "applies_environment_specific_paused_state" {
  command = plan

  assert {
    condition     = hyperping_monitor.environment["dev"].paused == true
    error_message = "Dev should be paused"
  }

  assert {
    condition     = hyperping_monitor.environment["staging"].paused == false
    error_message = "Staging should not be paused"
  }

  assert {
    condition     = hyperping_monitor.environment["prod"].paused == false
    error_message = "Prod should not be paused"
  }
}

run "applies_environment_specific_required_keyword" {
  command = plan

  assert {
    condition     = hyperping_monitor.environment["staging"].required_keyword == "healthy"
    error_message = "Staging should have required keyword"
  }
}

run "applies_environment_specific_alerts_wait" {
  command = plan

  assert {
    condition     = hyperping_monitor.environment["staging"].alerts_wait == 60
    error_message = "Staging should wait 60 seconds before alerting"
  }

  assert {
    condition     = hyperping_monitor.environment["prod"].alerts_wait == 0
    error_message = "Prod should have immediate alerts (0 seconds)"
  }
}

run "applies_environment_specific_escalation_policy" {
  command = plan

  assert {
    condition     = hyperping_monitor.environment["prod"].escalation_policy == "pol_test123"
    error_message = "Prod should have escalation policy"
  }
}

run "applies_default_headers" {
  command = plan

  assert {
    condition     = length(hyperping_monitor.environment["dev"].request_headers) == 1
    error_message = "Should have 1 default header"
  }

  assert {
    condition     = hyperping_monitor.environment["dev"].request_headers[0].name == "Content-Type"
    error_message = "Should have Content-Type header"
  }

  assert {
    condition     = hyperping_monitor.environment["dev"].request_headers[0].value == "application/json"
    error_message = "Content-Type should be application/json"
  }
}
