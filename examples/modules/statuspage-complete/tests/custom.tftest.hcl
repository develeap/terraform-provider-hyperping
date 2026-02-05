# Status Page Complete Module - Custom Configuration Tests

provider "hyperping" {
  api_key = "test_mock_api_key_for_plan_only"
}

variables {
  name             = "Acme Corp Status"
  hosted_subdomain = "acme-status"
  hostname         = "status.acme.com"

  services = {
    api = {
      url                  = "https://api.acme.com/health"
      method               = "POST"
      frequency            = 30
      expected_status_code = "201"
    }
    payments = {
      url = "https://payments.acme.com/health"
    }
  }

  theme                = "dark"
  accent_color         = "#3B82F6"
  languages            = ["en", "es", "fr"]
  regions              = ["virginia", "london", "singapore"]
  hide_powered_by      = true
  enable_subscriptions = false
}

run "applies_custom_hostname" {
  command = plan

  assert {
    condition     = hyperping_statuspage.main.hostname == "status.acme.com"
    error_message = "Hostname should be set"
  }
}

run "applies_service_custom_settings" {
  command = plan

  assert {
    condition     = hyperping_monitor.service["api"].http_method == "POST"
    error_message = "API monitor should use POST method"
  }

  assert {
    condition     = hyperping_monitor.service["api"].check_frequency == 30
    error_message = "API monitor frequency should be 30 seconds"
  }

  assert {
    condition     = hyperping_monitor.service["api"].expected_status_code == "201"
    error_message = "API monitor should expect 201 status"
  }
}

run "creates_correct_number_of_monitors" {
  command = plan

  assert {
    condition     = length(hyperping_monitor.service) == 2
    error_message = "Should create 2 monitors"
  }
}
