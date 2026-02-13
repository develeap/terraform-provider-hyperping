# Website Monitor Module - Example Usage Test
#
# This test validates the exact usage pattern from the module requirements

provider "hyperping" {
  api_key = "test_mock_api_key_for_plan_only"
}

# Exact usage from requirements
variables {
  domain = "example.com"

  pages = {
    homepage = {
      path            = "/"
      expected_text   = "Welcome"
      expected_status = "200"
    }
    login = {
      path            = "/login"
      expected_text   = "Sign In"
      expected_status = "200"
    }
    checkout = {
      path            = "/checkout"
      expected_text   = "Shopping Cart"
      expected_status = "200"
    }
  }

  frequency = 60
  regions   = ["virginia", "london", "singapore"]
}

run "validates_example_usage" {
  command = plan

  assert {
    condition     = length(hyperping_monitor.page) == 3
    error_message = "Should create 3 monitors (homepage, login, checkout)"
  }
}

run "homepage_configuration" {
  command = plan

  assert {
    condition     = hyperping_monitor.page["homepage"].url == "https://example.com/"
    error_message = "Homepage URL incorrect"
  }

  assert {
    condition     = hyperping_monitor.page["homepage"].required_keyword == "Welcome"
    error_message = "Homepage should validate 'Welcome' text"
  }

  assert {
    condition     = hyperping_monitor.page["homepage"].expected_status_code == "200"
    error_message = "Homepage should expect 200 status"
  }

  assert {
    condition     = hyperping_monitor.page["homepage"].check_frequency == 60
    error_message = "Homepage should check every 60 seconds"
  }

  assert {
    condition     = length(hyperping_monitor.page["homepage"].regions) == 3
    error_message = "Homepage should monitor from 3 regions"
  }
}

run "login_configuration" {
  command = plan

  assert {
    condition     = hyperping_monitor.page["login"].url == "https://example.com/login"
    error_message = "Login URL incorrect"
  }

  assert {
    condition     = hyperping_monitor.page["login"].required_keyword == "Sign In"
    error_message = "Login should validate 'Sign In' text"
  }

  assert {
    condition     = hyperping_monitor.page["login"].expected_status_code == "200"
    error_message = "Login should expect 200 status"
  }
}

run "checkout_configuration" {
  command = plan

  assert {
    condition     = hyperping_monitor.page["checkout"].url == "https://example.com/checkout"
    error_message = "Checkout URL incorrect"
  }

  assert {
    condition     = hyperping_monitor.page["checkout"].required_keyword == "Shopping Cart"
    error_message = "Checkout should validate 'Shopping Cart' text"
  }

  assert {
    condition     = hyperping_monitor.page["checkout"].expected_status_code == "200"
    error_message = "Checkout should expect 200 status"
  }
}

run "all_use_specified_regions" {
  command = plan

  assert {
    condition = alltrue([
      for k, v in hyperping_monitor.page : contains(v.regions, "virginia")
    ])
    error_message = "All should monitor from virginia"
  }

  assert {
    condition = alltrue([
      for k, v in hyperping_monitor.page : contains(v.regions, "london")
    ])
    error_message = "All should monitor from london"
  }

  assert {
    condition = alltrue([
      for k, v in hyperping_monitor.page : contains(v.regions, "singapore")
    ])
    error_message = "All should monitor from singapore"
  }
}

run "all_use_same_frequency" {
  command = plan

  assert {
    condition = alltrue([
      for k, v in hyperping_monitor.page : v.check_frequency == 60
    ])
    error_message = "All should check every 60 seconds"
  }
}
