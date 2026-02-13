# CDN Monitor Module - Advanced Configuration Tests

provider "hyperping" {
  api_key = "test_mock_api_key_for_plan_only"
}

variables {
  cdn_domain  = "static.example.com"
  name_prefix = "PROD"
  name_format = "[PROD-CDN] Asset: %s"
  protocol    = "http"

  assets = {
    js     = "/dist/app.min.js"
    css    = "/dist/style.min.css"
    images = "/images/hero.webp"
  }

  regions         = ["virginia", "london", "tokyo"]
  check_frequency = 600

  monitor_root_domain  = true
  root_expected_status = "404"

  paused = true
}

run "applies_custom_name_format" {
  command = plan

  assert {
    condition     = hyperping_monitor.cdn_asset["js"].name == "[PROD-CDN] Asset: js"
    error_message = "Monitor name should use custom format"
  }

  assert {
    condition     = hyperping_monitor.cdn_asset["css"].name == "[PROD-CDN] Asset: css"
    error_message = "CSS monitor name should use custom format"
  }
}

run "uses_http_protocol" {
  command = plan

  assert {
    condition     = hyperping_monitor.cdn_asset["js"].url == "http://static.example.com/dist/app.min.js"
    error_message = "Should use HTTP protocol when specified"
  }

  assert {
    condition     = hyperping_monitor.cdn_asset["css"].url == "http://static.example.com/dist/style.min.css"
    error_message = "CSS URL should use HTTP protocol"
  }
}

run "applies_custom_regions" {
  command = plan

  assert {
    condition     = length(hyperping_monitor.cdn_asset["js"].regions) == 3
    error_message = "Should use 3 custom regions"
  }

  assert {
    condition = (
      contains(hyperping_monitor.cdn_asset["js"].regions, "virginia") &&
      contains(hyperping_monitor.cdn_asset["js"].regions, "london") &&
      contains(hyperping_monitor.cdn_asset["js"].regions, "tokyo")
    )
    error_message = "Should use specified regions"
  }
}

run "applies_custom_check_frequency" {
  command = plan

  assert {
    condition     = hyperping_monitor.cdn_asset["js"].check_frequency == 600
    error_message = "Check frequency should be 600 seconds"
  }
}

run "creates_root_domain_monitor" {
  command = plan

  assert {
    condition     = length(hyperping_monitor.cdn_root) == 1
    error_message = "Should create root domain monitor when enabled"
  }

  assert {
    condition     = hyperping_monitor.cdn_root[0].url == "http://static.example.com"
    error_message = "Root monitor should use CDN domain without path"
  }

  assert {
    condition     = hyperping_monitor.cdn_root[0].expected_status_code == "404"
    error_message = "Root monitor should expect configured status code"
  }

  assert {
    condition     = hyperping_monitor.cdn_root[0].name == "[PROD-CDN] Asset: root"
    error_message = "Root monitor should use name format"
  }
}

run "applies_paused_state" {
  command = plan

  assert {
    condition     = hyperping_monitor.cdn_asset["js"].paused == true
    error_message = "Asset monitors should be paused"
  }

  assert {
    condition     = hyperping_monitor.cdn_root[0].paused == true
    error_message = "Root monitor should be paused"
  }
}

run "creates_all_asset_monitors" {
  command = plan

  assert {
    condition     = length(hyperping_monitor.cdn_asset) == 3
    error_message = "Should create monitors for all 3 assets"
  }
}
