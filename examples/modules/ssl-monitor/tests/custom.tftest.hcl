# SSL Monitor Module - Custom Configuration Tests

variables {
  domains         = ["payments.example.com", "api.example.com"]
  name_prefix     = "SSL-PROD"
  check_frequency = 1800
  regions         = ["virginia", "london", "tokyo"]
  alerts_wait     = 2
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

run "applies_custom_regions" {
  command = plan

  assert {
    condition     = tolist(hyperping_monitor.ssl["payments.example.com"].regions) == tolist(["virginia", "london", "tokyo"])
    error_message = "Regions should match custom configuration"
  }
}

run "applies_alerting_settings" {
  command = plan

  assert {
    condition     = hyperping_monitor.ssl["payments.example.com"].alerts_wait == 2
    error_message = "Alerts wait should be 2"
  }

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
