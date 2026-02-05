# Status Page Complete Module - Basic Tests

variables {
  name      = "Test Status Page"
  subdomain = "test-status"

  services = {
    api = {
      url = "https://api.example.com/health"
    }
    web = {
      url = "https://www.example.com"
    }
  }
}

run "creates_status_page" {
  command = plan

  assert {
    condition     = hyperping_statuspage.main.name == "Test Status Page"
    error_message = "Status page name should match input"
  }

  assert {
    condition     = hyperping_statuspage.main.subdomain == "test-status"
    error_message = "Status page subdomain should match input"
  }
}

run "creates_service_monitors" {
  command = plan

  assert {
    condition     = length(hyperping_monitor.service) == 2
    error_message = "Should create 2 monitors for 2 services"
  }
}

run "monitors_have_correct_urls" {
  command = plan

  assert {
    condition     = hyperping_monitor.service["api"].url == "https://api.example.com/health"
    error_message = "API monitor URL should match service URL"
  }

  assert {
    condition     = hyperping_monitor.service["web"].url == "https://www.example.com"
    error_message = "Web monitor URL should match service URL"
  }
}

run "applies_default_theme" {
  command = plan

  assert {
    condition     = hyperping_statuspage.main.theme == "system"
    error_message = "Default theme should be system"
  }
}

run "enables_subscriptions_by_default" {
  command = plan

  assert {
    condition     = hyperping_statuspage.main.enable_subscriptions == true
    error_message = "Subscriptions should be enabled by default"
  }
}
