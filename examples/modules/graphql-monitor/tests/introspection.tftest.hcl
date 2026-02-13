# GraphQL Monitor Module - Introspection Tests

provider "hyperping" {
  api_key = "test_mock_api_key_for_plan_only"
}

variables {
  endpoint = "https://api.example.com/graphql"
  queries = {
    health = {
      query             = "{ health { status } }"
      expected_response = "\"status\":\"ok\""
    }
  }
}

run "introspection_disabled_by_default" {
  command = plan

  assert {
    condition     = length(hyperping_monitor.introspection) == 0
    error_message = "Introspection monitor should not be created by default"
  }
}

run "introspection_enabled" {
  command = plan

  variables {
    enable_introspection_check = true
  }

  assert {
    condition     = length(hyperping_monitor.introspection) == 1
    error_message = "Should create introspection monitor when enabled"
  }

  assert {
    condition     = hyperping_monitor.introspection[0].name == "GraphQL - introspection"
    error_message = "Introspection monitor should have correct name"
  }

  assert {
    condition     = hyperping_monitor.introspection[0].http_method == "POST"
    error_message = "Introspection monitor should use POST method"
  }

  assert {
    condition     = hyperping_monitor.introspection[0].url == "https://api.example.com/graphql"
    error_message = "Introspection monitor should use same endpoint"
  }
}

run "introspection_default_query" {
  command = plan

  variables {
    enable_introspection_check = true
  }

  assert {
    condition     = can(jsondecode(hyperping_monitor.introspection[0].request_body))
    error_message = "Introspection request body should be valid JSON"
  }

  assert {
    condition     = jsondecode(hyperping_monitor.introspection[0].request_body).query == "{ __schema { queryType { name } } }"
    error_message = "Should use default introspection query"
  }

  assert {
    condition     = hyperping_monitor.introspection[0].required_keyword == "queryType"
    error_message = "Should use default expected response"
  }

  assert {
    condition     = hyperping_monitor.introspection[0].check_frequency == 3600
    error_message = "Should use default frequency of 3600 seconds"
  }
}

run "introspection_custom_query" {
  command = plan

  variables {
    enable_introspection_check     = true
    introspection_query            = "{ __type(name: \"Query\") { name } }"
    introspection_expected_response = "\"name\":\"Query\""
  }

  assert {
    condition     = jsondecode(hyperping_monitor.introspection[0].request_body).query == "{ __type(name: \"Query\") { name } }"
    error_message = "Should use custom introspection query"
  }

  assert {
    condition     = hyperping_monitor.introspection[0].required_keyword == "\"name\":\"Query\""
    error_message = "Should use custom expected response"
  }
}

run "introspection_custom_frequency" {
  command = plan

  variables {
    enable_introspection_check = true
    introspection_frequency    = 600
  }

  assert {
    condition     = hyperping_monitor.introspection[0].check_frequency == 600
    error_message = "Should use custom frequency"
  }
}

run "introspection_with_custom_headers" {
  command = plan

  variables {
    enable_introspection_check = true
    custom_headers = {
      "X-Environment" = "production"
      "X-Request-ID"  = "test-123"
    }
  }

  assert {
    condition = anytrue([
      for header in hyperping_monitor.introspection[0].request_headers :
      header.name == "X-Environment" && header.value == "production"
    ])
    error_message = "Introspection monitor should include custom headers"
  }

  assert {
    condition = anytrue([
      for header in hyperping_monitor.introspection[0].request_headers :
      header.name == "Content-Type" && header.value == "application/json"
    ])
    error_message = "Introspection monitor should include Content-Type header"
  }
}

run "introspection_with_prefix" {
  command = plan

  variables {
    enable_introspection_check = true
    name_prefix                = "PROD"
  }

  assert {
    condition     = hyperping_monitor.introspection[0].name == "[PROD] GraphQL - introspection"
    error_message = "Introspection monitor should include prefix"
  }
}

run "introspection_custom_status_code" {
  command = plan

  variables {
    enable_introspection_check     = true
    introspection_expected_status  = "2xx"
  }

  assert {
    condition     = hyperping_monitor.introspection[0].expected_status_code == "2xx"
    error_message = "Should use custom expected status code"
  }
}

run "total_monitor_count_with_introspection" {
  command = plan

  variables {
    enable_introspection_check = true
    queries = {
      health = {
        query             = "{ health { status } }"
        expected_response = "\"status\":\"ok\""
      }
      users = {
        query             = "{ users { count } }"
        expected_response = "\"count\""
      }
    }
  }

  assert {
    condition     = length(hyperping_monitor.graphql_query) == 2
    error_message = "Should create 2 query monitors"
  }

  assert {
    condition     = length(hyperping_monitor.introspection) == 1
    error_message = "Should create 1 introspection monitor"
  }
}

run "introspection_follows_global_paused_setting" {
  command = plan

  variables {
    enable_introspection_check = true
    paused                     = true
  }

  assert {
    condition     = hyperping_monitor.introspection[0].paused == true
    error_message = "Introspection monitor should respect global paused setting"
  }
}
