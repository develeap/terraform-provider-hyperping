# Status Page Complete Module - Custom Configuration Tests

variables {
  name      = "Acme Corp Status"
  subdomain = "acme-status"
  hostname  = "status.acme.com"

  services = {
    api = {
      url                  = "https://api.acme.com/health"
      method               = "POST"
      frequency            = 30
      expected_status_code = "201"
    }
    payments = {
      url = "https://payments.acme.com/health"
      headers = {
        "X-Health-Check" = "true"
      }
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

run "applies_branding_settings" {
  command = plan

  assert {
    condition     = hyperping_statuspage.main.theme == "dark"
    error_message = "Theme should be dark"
  }

  assert {
    condition     = hyperping_statuspage.main.accent_color == "#3B82F6"
    error_message = "Accent color should be custom value"
  }

  assert {
    condition     = hyperping_statuspage.main.hide_powered_by == true
    error_message = "Should hide powered by branding"
  }
}

run "applies_language_settings" {
  command = plan

  assert {
    condition     = tolist(hyperping_statuspage.main.languages) == tolist(["en", "es", "fr"])
    error_message = "Languages should match custom configuration"
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

run "disables_subscriptions" {
  command = plan

  assert {
    condition     = hyperping_statuspage.main.enable_subscriptions == false
    error_message = "Subscriptions should be disabled"
  }
}
