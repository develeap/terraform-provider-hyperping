# CDN Monitor Module - Basic Tests

provider "hyperping" {
  api_key = "test_mock_api_key_for_plan_only"
}

variables {
  cdn_domain = "cdn.example.com"
  assets = {
    logo = "/images/logo.png"
    css  = "/styles/main.css"
  }
}

run "creates_monitors_for_all_assets" {
  command = plan

  assert {
    condition     = length(hyperping_monitor.cdn_asset) == 2
    error_message = "Should create 2 monitors for 2 assets"
  }

  assert {
    condition     = hyperping_monitor.cdn_asset["logo"].name == "CDN: logo"
    error_message = "Logo monitor name should use default format"
  }

  assert {
    condition     = hyperping_monitor.cdn_asset["css"].name == "CDN: css"
    error_message = "CSS monitor name should use default format"
  }
}

run "builds_correct_asset_urls" {
  command = plan

  assert {
    condition     = hyperping_monitor.cdn_asset["logo"].url == "https://cdn.example.com/images/logo.png"
    error_message = "Logo URL should combine domain and path with HTTPS"
  }

  assert {
    condition     = hyperping_monitor.cdn_asset["css"].url == "https://cdn.example.com/styles/main.css"
    error_message = "CSS URL should combine domain and path with HTTPS"
  }
}

run "applies_default_settings" {
  command = plan

  assert {
    condition     = hyperping_monitor.cdn_asset["logo"].http_method == "GET"
    error_message = "HTTP method should default to GET"
  }

  assert {
    condition     = hyperping_monitor.cdn_asset["logo"].check_frequency == 300
    error_message = "Check frequency should default to 300 seconds"
  }

  assert {
    condition     = hyperping_monitor.cdn_asset["logo"].expected_status_code == "2xx"
    error_message = "Expected status should be 2xx"
  }

  assert {
    condition     = hyperping_monitor.cdn_asset["logo"].follow_redirects == false
    error_message = "Should not follow redirects by default for CDN assets"
  }
}

run "uses_default_regions" {
  command = plan

  assert {
    condition     = length(hyperping_monitor.cdn_asset["logo"].regions) == 5
    error_message = "Should use 5 default regions"
  }

  assert {
    condition     = contains(hyperping_monitor.cdn_asset["logo"].regions, "virginia")
    error_message = "Should include virginia region"
  }

  assert {
    condition     = contains(hyperping_monitor.cdn_asset["logo"].regions, "singapore")
    error_message = "Should include singapore region"
  }
}

run "does_not_create_root_monitor_by_default" {
  command = plan

  assert {
    condition     = length(hyperping_monitor.cdn_root) == 0
    error_message = "Should not create root domain monitor by default"
  }
}
