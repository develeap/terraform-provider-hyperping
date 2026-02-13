# Website Monitor Module - E-commerce Example Tests

provider "hyperping" {
  api_key = "test_mock_api_key_for_plan_only"
}

variables {
  domain      = "shop.example.com"
  name_prefix = "PROD"
  frequency   = 30

  performance_threshold_ms = 2000

  pages = {
    homepage = {
      path          = "/"
      expected_text = "Shop Now"
    }
    login = {
      path            = "/login"
      expected_text   = "Sign In"
      expected_status = "200"
    }
    checkout = {
      path                     = "/checkout"
      expected_text            = "Shopping Cart"
      performance_threshold_ms = 1000
    }
    products = {
      path          = "/products"
      expected_text = "Browse Products"
    }
    account = {
      path          = "/account"
      expected_text = "My Account"
    }
  }

  regions = ["virginia", "london", "singapore"]
}

run "creates_all_ecommerce_pages" {
  command = plan

  assert {
    condition     = length(hyperping_monitor.page) == 5
    error_message = "Should create 5 monitors for e-commerce pages"
  }
}

run "applies_name_prefix_to_all" {
  command = plan

  assert {
    condition = alltrue([
      for k, v in hyperping_monitor.page : startswith(v.name, "[PROD]")
    ])
    error_message = "All monitors should have PROD prefix"
  }
}

run "homepage_has_correct_config" {
  command = plan

  assert {
    condition     = hyperping_monitor.page["homepage"].url == "https://shop.example.com/"
    error_message = "Homepage URL should be correct"
  }

  assert {
    condition     = hyperping_monitor.page["homepage"].required_keyword == "Shop Now"
    error_message = "Homepage should check for 'Shop Now'"
  }

  assert {
    condition     = hyperping_monitor.page["homepage"].check_frequency == 30
    error_message = "Homepage should check every 30 seconds"
  }

  assert {
    condition     = hyperping_monitor.page["homepage"].response_time_threshold == 2000
    error_message = "Homepage should have 2000ms threshold"
  }
}

run "checkout_has_stricter_threshold" {
  command = plan

  assert {
    condition     = hyperping_monitor.page["checkout"].response_time_threshold == 1000
    error_message = "Checkout should have stricter 1000ms threshold"
  }

  assert {
    condition     = hyperping_monitor.page["checkout"].required_keyword == "Shopping Cart"
    error_message = "Checkout should verify cart text"
  }
}

run "login_expects_200_status" {
  command = plan

  assert {
    condition     = hyperping_monitor.page["login"].expected_status_code == "200"
    error_message = "Login should expect exact 200 status"
  }

  assert {
    condition     = hyperping_monitor.page["login"].required_keyword == "Sign In"
    error_message = "Login should verify sign in text"
  }
}

run "all_pages_use_custom_regions" {
  command = plan

  assert {
    condition = alltrue([
      for k, v in hyperping_monitor.page : length(v.regions) == 3
    ])
    error_message = "All monitors should use 3 regions"
  }

  assert {
    condition = alltrue([
      for k, v in hyperping_monitor.page : contains(v.regions, "virginia")
    ])
    error_message = "All monitors should include virginia region"
  }

  assert {
    condition = alltrue([
      for k, v in hyperping_monitor.page : contains(v.regions, "singapore")
    ])
    error_message = "All monitors should include singapore region"
  }
}

run "all_pages_validate_content" {
  command = plan

  assert {
    condition = alltrue([
      for k, v in hyperping_monitor.page : v.required_keyword != null && v.required_keyword != ""
    ])
    error_message = "All pages should have content validation"
  }
}

run "all_pages_have_performance_monitoring" {
  command = plan

  assert {
    condition = alltrue([
      for k, v in hyperping_monitor.page : v.response_time_threshold != null
    ])
    error_message = "All pages should have performance thresholds"
  }
}
