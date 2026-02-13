# CDN Monitor Module - Multi-Region Tests

provider "hyperping" {
  api_key = "test_mock_api_key_for_plan_only"
}

variables {
  cdn_domain = "global-cdn.example.com"

  assets = {
    app_js = "/js/app.bundle.js"
  }

  # Global coverage across continents
  regions = [
    "virginia",  # North America
    "london",    # Europe West
    "frankfurt", # Europe Central
    "singapore", # Asia Southeast
    "tokyo",     # Asia East
    "sydney",    # Oceania
    "saopaulo",  # South America
    "bahrain"    # Middle East
  ]

  check_frequency = 300
}

run "uses_global_regions" {
  command = plan

  assert {
    condition     = length(hyperping_monitor.cdn_asset["app_js"].regions) == 8
    error_message = "Should use 8 global regions"
  }

  assert {
    condition     = contains(hyperping_monitor.cdn_asset["app_js"].regions, "virginia")
    error_message = "Should include virginia region"
  }

  assert {
    condition     = contains(hyperping_monitor.cdn_asset["app_js"].regions, "singapore")
    error_message = "Should include singapore region"
  }

  assert {
    condition     = contains(hyperping_monitor.cdn_asset["app_js"].regions, "saopaulo")
    error_message = "Should include saopaulo region"
  }

  assert {
    condition     = contains(hyperping_monitor.cdn_asset["app_js"].regions, "sydney")
    error_message = "Should include sydney region"
  }

  assert {
    condition     = contains(hyperping_monitor.cdn_asset["app_js"].regions, "bahrain")
    error_message = "Should include bahrain region"
  }
}

run "validates_global_cdn_coverage" {
  command = plan

  assert {
    condition     = hyperping_monitor.cdn_asset["app_js"].check_frequency == 300
    error_message = "Should check every 5 minutes for global monitoring"
  }

  assert {
    condition     = hyperping_monitor.cdn_asset["app_js"].url == "https://global-cdn.example.com/js/app.bundle.js"
    error_message = "URL should be correctly formatted"
  }
}
