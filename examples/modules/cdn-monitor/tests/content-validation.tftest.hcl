# CDN Monitor Module - Content Validation Tests

provider "hyperping" {
  api_key = "test_mock_api_key_for_plan_only"
}

variables {
  cdn_domain = "cdn.example.com"

  assets = {
    html = "/index.html"
    json = "/api/data.json"
    css  = "/styles/main.css"
  }

  asset_keywords = {
    html = "<!DOCTYPE html>"
    json = "\"version\""
  }

  regions = ["virginia", "london", "singapore"]
}

run "applies_required_keywords" {
  command = plan

  assert {
    condition     = hyperping_monitor.cdn_asset["html"].required_keyword == "<!DOCTYPE html>"
    error_message = "HTML monitor should validate DOCTYPE keyword"
  }

  assert {
    condition     = hyperping_monitor.cdn_asset["json"].required_keyword == "\"version\""
    error_message = "JSON monitor should validate version keyword"
  }
}

run "no_keyword_validation_when_not_specified" {
  command = plan

  assert {
    condition     = hyperping_monitor.cdn_asset["css"].required_keyword == null
    error_message = "CSS monitor should not have keyword validation"
  }
}

run "creates_monitors_with_content_validation" {
  command = plan

  assert {
    condition     = length(hyperping_monitor.cdn_asset) == 3
    error_message = "Should create all 3 monitors"
  }

  assert {
    condition = (
      hyperping_monitor.cdn_asset["html"].url == "https://cdn.example.com/index.html" &&
      hyperping_monitor.cdn_asset["json"].url == "https://cdn.example.com/api/data.json"
    )
    error_message = "URLs should be correctly formatted"
  }
}
