# SSL Monitor Module - Basic Tests

provider "hyperping" {
  api_key = "test_mock_api_key_for_plan_only"
}

variables {
  domains = ["example.com", "api.example.com"]
}

run "creates_monitors_for_domains" {
  command = plan

  assert {
    condition     = length(hyperping_monitor.ssl) == 2
    error_message = "Should create 2 monitors for 2 domains"
  }
}

run "uses_https_protocol" {
  command = plan

  assert {
    condition     = hyperping_monitor.ssl["example.com"].url == "https://example.com"
    error_message = "Monitor URL should use https protocol"
  }

  assert {
    condition     = hyperping_monitor.ssl["api.example.com"].url == "https://api.example.com"
    error_message = "Monitor URL should use https protocol"
  }
}

run "applies_default_name_prefix" {
  command = plan

  assert {
    condition     = hyperping_monitor.ssl["example.com"].name == "[SSL] example.com"
    error_message = "Monitor name should include SSL prefix"
  }
}

run "uses_default_frequency" {
  command = plan

  assert {
    condition     = hyperping_monitor.ssl["example.com"].check_frequency == 3600
    error_message = "Default frequency should be 3600 seconds (1 hour)"
  }
}

run "uses_http_protocol_type" {
  command = plan

  assert {
    condition     = hyperping_monitor.ssl["example.com"].protocol == "http"
    error_message = "Protocol should be http"
  }
}
