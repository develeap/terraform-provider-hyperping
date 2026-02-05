# API Health Module - Custom Configuration Tests

provider "hyperping" {
  api_key = "test_mock_api_key_for_plan_only"
}

variables {
  name_prefix       = "PROD-API"
  name_format       = "[PROD-API] Health: %s"
  default_frequency = 30

  endpoints = {
    payments = {
      url                  = "https://api.example.com/v1/health"
      method               = "POST"
      expected_status_code = "201"
    }
  }

  default_regions = ["virginia", "london", "tokyo"]
  paused          = true
}

run "applies_custom_name_format" {
  command = plan

  assert {
    condition     = hyperping_monitor.endpoint["payments"].name == "[PROD-API] Health: payments"
    error_message = "Monitor name should use custom format"
  }
}

run "applies_custom_http_settings" {
  command = plan

  assert {
    condition     = hyperping_monitor.endpoint["payments"].http_method == "POST"
    error_message = "HTTP method should be POST"
  }

  assert {
    condition     = hyperping_monitor.endpoint["payments"].check_frequency == 30
    error_message = "Frequency should be 30 seconds"
  }

  assert {
    condition     = hyperping_monitor.endpoint["payments"].expected_status_code == "201"
    error_message = "Expected status code should be 201"
  }
}

run "applies_monitoring_settings" {
  command = plan

  assert {
    condition     = hyperping_monitor.endpoint["payments"].paused == true
    error_message = "Monitor should be paused"
  }
}
