# Website Monitor Module - Advanced Tests

provider "hyperping" {
  api_key = "test_mock_api_key_for_plan_only"
}

variables {
  domain = "shop.example.com"
  pages = {
    homepage = {
      path = "/"
    }
  }
}

run "applies_performance_threshold_globally" {
  command = plan

  variables {
    performance_threshold_ms = 2000
    pages = {
      homepage = {
        path          = "/"
        expected_text = "Welcome"
      }
      products = {
        path          = "/products"
        expected_text = "Browse"
      }
    }
  }

  assert {
    condition     = hyperping_monitor.page["homepage"].response_time_threshold == 2000
    error_message = "Homepage should have global performance threshold"
  }

  assert {
    condition     = hyperping_monitor.page["products"].response_time_threshold == 2000
    error_message = "Products page should have global performance threshold"
  }
}

run "overrides_performance_threshold_per_page" {
  command = plan

  variables {
    performance_threshold_ms = 2000
    pages = {
      homepage = {
        path          = "/"
        expected_text = "Welcome"
      }
      checkout = {
        path                     = "/checkout"
        expected_text            = "Cart"
        performance_threshold_ms = 1000
      }
    }
  }

  assert {
    condition     = hyperping_monitor.page["homepage"].response_time_threshold == 2000
    error_message = "Homepage should use global threshold"
  }

  assert {
    condition     = hyperping_monitor.page["checkout"].response_time_threshold == 1000
    error_message = "Checkout should use page-specific threshold"
  }
}

run "applies_custom_headers" {
  command = plan

  variables {
    default_headers = {
      "Authorization" = "Bearer test-token"
      "User-Agent"    = "TestMonitor"
    }
    pages = {
      dashboard = {
        path          = "/dashboard"
        expected_text = "Dashboard"
      }
    }
  }

  assert {
    condition     = length(hyperping_monitor.page["dashboard"].request_headers) == 2
    error_message = "Should have 2 headers"
  }

  assert {
    condition = anytrue([
      for h in hyperping_monitor.page["dashboard"].request_headers :
      h.name == "Authorization" && h.value == "Bearer test-token"
    ])
    error_message = "Should include Authorization header"
  }
}

run "overrides_headers_per_page" {
  command = plan

  variables {
    default_headers = {
      "User-Agent" = "DefaultAgent"
    }
    pages = {
      api = {
        path = "/api"
        headers = {
          "X-Custom" = "CustomValue"
        }
      }
    }
  }

  assert {
    condition     = length(hyperping_monitor.page["api"].request_headers) == 1
    error_message = "Should only have page-specific headers"
  }

  assert {
    condition = anytrue([
      for h in hyperping_monitor.page["api"].request_headers :
      h.name == "X-Custom" && h.value == "CustomValue"
    ])
    error_message = "Should include custom header"
  }
}

run "supports_different_methods" {
  command = plan

  variables {
    pages = {
      health = {
        path   = "/health"
        method = "HEAD"
      }
      webhook = {
        path   = "/webhook"
        method = "POST"
        body   = jsonencode({ test = true })
      }
    }
  }

  assert {
    condition     = hyperping_monitor.page["health"].http_method == "HEAD"
    error_message = "Health check should use HEAD method"
  }

  assert {
    condition     = hyperping_monitor.page["webhook"].http_method == "POST"
    error_message = "Webhook should use POST method"
  }

  assert {
    condition     = hyperping_monitor.page["webhook"].request_body != null
    error_message = "Webhook should have request body"
  }
}

run "supports_custom_frequency_per_page" {
  command = plan

  variables {
    frequency = 60
    pages = {
      homepage = {
        path = "/"
      }
      critical = {
        path      = "/checkout"
        frequency = 30
      }
    }
  }

  assert {
    condition     = hyperping_monitor.page["homepage"].check_frequency == 60
    error_message = "Homepage should use default frequency"
  }

  assert {
    condition     = hyperping_monitor.page["critical"].check_frequency == 30
    error_message = "Critical page should use custom frequency"
  }
}

run "supports_custom_regions_per_page" {
  command = plan

  variables {
    regions = ["virginia", "london"]
    pages = {
      global = {
        path = "/"
      }
      asia_only = {
        path    = "/asia"
        regions = ["singapore", "tokyo"]
      }
    }
  }

  assert {
    condition     = length(hyperping_monitor.page["global"].regions) == 2
    error_message = "Global page should use default regions"
  }

  assert {
    condition     = length(hyperping_monitor.page["asia_only"].regions) == 2
    error_message = "Asia-only page should use custom regions"
  }

  assert {
    condition     = contains(hyperping_monitor.page["asia_only"].regions, "singapore")
    error_message = "Asia-only should include singapore"
  }
}

run "supports_expected_status_codes" {
  command = plan

  variables {
    pages = {
      normal = {
        path            = "/"
        expected_status = "200"
      }
      redirect = {
        path            = "/old-page"
        expected_status = "3xx"
      }
    }
  }

  assert {
    condition     = hyperping_monitor.page["normal"].expected_status_code == "200"
    error_message = "Normal page should expect 200"
  }

  assert {
    condition     = hyperping_monitor.page["redirect"].expected_status_code == "3xx"
    error_message = "Redirect page should expect 3xx"
  }
}
