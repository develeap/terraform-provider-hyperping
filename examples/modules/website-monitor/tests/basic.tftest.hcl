# Website Monitor Module - Basic Tests

provider "hyperping" {
  api_key = "test_mock_api_key_for_plan_only"
}

variables {
  domain = "example.com"
  pages = {
    homepage = {
      path          = "/"
      expected_text = "Welcome"
    }
  }
}

run "creates_monitor_with_defaults" {
  command = plan

  assert {
    condition     = hyperping_monitor.page["homepage"].name == "example.com - homepage"
    error_message = "Monitor name should include domain and page key"
  }

  assert {
    condition     = hyperping_monitor.page["homepage"].url == "https://example.com/"
    error_message = "Monitor URL should use https protocol and correct path"
  }

  assert {
    condition     = hyperping_monitor.page["homepage"].http_method == "GET"
    error_message = "Default HTTP method should be GET"
  }

  assert {
    condition     = hyperping_monitor.page["homepage"].check_frequency == 60
    error_message = "Default frequency should be 60 seconds"
  }

  assert {
    condition     = hyperping_monitor.page["homepage"].required_keyword == "Welcome"
    error_message = "Required keyword should match expected_text"
  }

  assert {
    condition     = hyperping_monitor.page["homepage"].protocol == "http"
    error_message = "Protocol should be http"
  }
}

run "creates_multiple_page_monitors" {
  command = plan

  variables {
    pages = {
      homepage = {
        path          = "/"
        expected_text = "Welcome"
      }
      login = {
        path          = "/login"
        expected_text = "Sign In"
      }
      about = {
        path          = "/about"
        expected_text = "About Us"
      }
    }
  }

  assert {
    condition     = length(hyperping_monitor.page) == 3
    error_message = "Should create 3 monitors"
  }

  assert {
    condition     = hyperping_monitor.page["login"].url == "https://example.com/login"
    error_message = "Login page URL should be correct"
  }

  assert {
    condition     = hyperping_monitor.page["about"].url == "https://example.com/about"
    error_message = "About page URL should be correct"
  }
}

run "applies_name_prefix" {
  command = plan

  variables {
    name_prefix = "PROD"
    pages = {
      homepage = {
        path = "/"
      }
    }
  }

  assert {
    condition     = hyperping_monitor.page["homepage"].name == "[PROD] example.com - homepage"
    error_message = "Monitor name should include prefix"
  }
}

run "uses_http_protocol" {
  command = plan

  variables {
    protocol = "http"
    pages = {
      homepage = {
        path = "/"
      }
    }
  }

  assert {
    condition     = hyperping_monitor.page["homepage"].url == "http://example.com/"
    error_message = "Should use http protocol when specified"
  }
}

run "validates_path_starts_with_slash" {
  command = plan

  variables {
    pages = {
      invalid = {
        path = "no-leading-slash"
      }
    }
  }

  expect_failures = [
    var.pages,
  ]
}
