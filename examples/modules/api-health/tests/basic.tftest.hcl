# API Health Module - Basic Tests

variables {
  name_prefix = "TEST"
  endpoints = {
    api = {
      url = "https://httpstat.us/200"
    }
  }
}

run "creates_monitor_with_defaults" {
  command = plan

  assert {
    condition     = hyperping_monitor.endpoint["api"].name == "[TEST] api"
    error_message = "Monitor name should include prefix and endpoint key"
  }

  assert {
    condition     = hyperping_monitor.endpoint["api"].url == "https://httpstat.us/200"
    error_message = "Monitor URL should match input"
  }

  assert {
    condition     = hyperping_monitor.endpoint["api"].http_method == "GET"
    error_message = "Default HTTP method should be GET"
  }

  assert {
    condition     = hyperping_monitor.endpoint["api"].check_frequency == 60
    error_message = "Default frequency should be 60 seconds"
  }
}

run "creates_multiple_monitors" {
  command = plan

  variables {
    endpoints = {
      api = {
        url = "https://api.example.com/health"
      }
      web = {
        url = "https://www.example.com"
      }
      admin = {
        url = "https://admin.example.com/health"
      }
    }
  }

  assert {
    condition     = length(hyperping_monitor.endpoint) == 3
    error_message = "Should create 3 monitors"
  }
}
