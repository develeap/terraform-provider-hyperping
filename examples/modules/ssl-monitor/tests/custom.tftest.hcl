# SSL Monitor Module - Custom Configuration Tests

provider "hyperping" {
  api_key = "test_mock_api_key_for_plan_only"
}

variables {
  domains         = ["payments.example.com", "api.example.com"]
  name_prefix     = "SSL-PROD"
  check_frequency = 1800
  regions         = ["virginia", "london", "tokyo"]
  paused          = true
}

run "applies_custom_name_prefix" {
  command = plan

  assert {
    condition     = hyperping_monitor.ssl["payments.example.com"].name == "[SSL-PROD] payments.example.com"
    error_message = "Monitor name should use custom prefix"
  }
}

run "applies_custom_frequency" {
  command = plan

  assert {
    condition     = hyperping_monitor.ssl["payments.example.com"].check_frequency == 1800
    error_message = "Frequency should be 1800 seconds (30 min)"
  }
}

run "applies_paused_state" {
  command = plan

  assert {
    condition     = hyperping_monitor.ssl["payments.example.com"].paused == true
    error_message = "Monitor should be paused"
  }
}

run "follows_redirects_by_default" {
  command = plan

  assert {
    condition     = hyperping_monitor.ssl["payments.example.com"].follow_redirects == true
    error_message = "Should follow redirects by default"
  }
}
