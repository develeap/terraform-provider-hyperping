# API Health Module - Custom Configuration Tests

variables {
  name_prefix = "PROD-API"
  name_format = "[%s] Health: %s"

  endpoints = {
    payments = {
      url                  = "https://api.example.com/v1/health"
      method               = "POST"
      frequency            = 30
      expected_status_code = "201"
      timeout              = 15
      headers = {
        "Authorization" = "Bearer test-token"
        "Content-Type"  = "application/json"
      }
    }
  }

  regions     = ["virginia", "london", "tokyo"]
  alerts_wait = 2
  paused      = true
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
    condition     = tolist(hyperping_monitor.endpoint["payments"].regions) == tolist(["virginia", "london", "tokyo"])
    error_message = "Regions should match custom configuration"
  }

  assert {
    condition     = hyperping_monitor.endpoint["payments"].alerts_wait == 2
    error_message = "Alerts wait should be 2"
  }

  assert {
    condition     = hyperping_monitor.endpoint["payments"].paused == true
    error_message = "Monitor should be paused"
  }
}
