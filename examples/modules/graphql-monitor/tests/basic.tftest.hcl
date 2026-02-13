# GraphQL Monitor Module - Basic Tests

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

run "creates_monitor_with_defaults" {
  command = plan

  assert {
    condition     = hyperping_monitor.graphql_query["health"].name == "GraphQL - health"
    error_message = "Monitor name should include query key"
  }

  assert {
    condition     = hyperping_monitor.graphql_query["health"].url == "https://api.example.com/graphql"
    error_message = "Monitor URL should match endpoint"
  }

  assert {
    condition     = hyperping_monitor.graphql_query["health"].http_method == "POST"
    error_message = "GraphQL monitors should use POST method"
  }

  assert {
    condition     = hyperping_monitor.graphql_query["health"].check_frequency == 120
    error_message = "Default frequency should be 120 seconds"
  }

  assert {
    condition     = hyperping_monitor.graphql_query["health"].protocol == "http"
    error_message = "Protocol should be http"
  }

  assert {
    condition     = hyperping_monitor.graphql_query["health"].required_keyword == "\"status\":\"ok\""
    error_message = "Required keyword should match expected response"
  }
}

run "creates_monitor_with_prefix" {
  command = plan

  variables {
    name_prefix = "PROD"
  }

  assert {
    condition     = hyperping_monitor.graphql_query["health"].name == "[PROD] GraphQL - health"
    error_message = "Monitor name should include prefix"
  }
}

run "creates_multiple_query_monitors" {
  command = plan

  variables {
    queries = {
      health = {
        query             = "{ health { status } }"
        expected_response = "\"status\":\"ok\""
      }
      users = {
        query             = "{ users { count } }"
        expected_response = "\"count\""
      }
      products = {
        query             = "{ products { total } }"
        expected_response = "\"total\""
      }
    }
  }

  assert {
    condition     = length(hyperping_monitor.graphql_query) == 3
    error_message = "Should create 3 query monitors"
  }

  assert {
    condition     = hyperping_monitor.graphql_query["health"].name == "GraphQL - health"
    error_message = "Health monitor should be created"
  }

  assert {
    condition     = hyperping_monitor.graphql_query["users"].name == "GraphQL - users"
    error_message = "Users monitor should be created"
  }

  assert {
    condition     = hyperping_monitor.graphql_query["products"].name == "GraphQL - products"
    error_message = "Products monitor should be created"
  }
}

run "validates_request_body_format" {
  command = plan

  assert {
    condition     = can(jsondecode(hyperping_monitor.graphql_query["health"].request_body))
    error_message = "Request body should be valid JSON"
  }

  assert {
    condition     = jsondecode(hyperping_monitor.graphql_query["health"].request_body).query == "{ health { status } }"
    error_message = "Request body should contain GraphQL query"
  }
}

run "sets_custom_frequency" {
  command = plan

  variables {
    queries = {
      health = {
        query             = "{ health { status } }"
        expected_response = "\"status\":\"ok\""
        frequency         = 60
      }
    }
  }

  assert {
    condition     = hyperping_monitor.graphql_query["health"].check_frequency == 60
    error_message = "Should use custom frequency"
  }
}

run "sets_custom_regions" {
  command = plan

  variables {
    queries = {
      health = {
        query             = "{ health { status } }"
        expected_response = "\"status\":\"ok\""
        regions           = ["virginia", "london", "singapore"]
      }
    }
  }

  assert {
    condition     = length(hyperping_monitor.graphql_query["health"].regions) == 3
    error_message = "Should use custom regions"
  }

  assert {
    condition     = contains(hyperping_monitor.graphql_query["health"].regions, "virginia")
    error_message = "Should include virginia region"
  }
}

run "includes_content_type_header" {
  command = plan

  assert {
    condition = anytrue([
      for header in hyperping_monitor.graphql_query["health"].request_headers :
      header.name == "Content-Type" && header.value == "application/json"
    ])
    error_message = "Should include Content-Type: application/json header"
  }
}
