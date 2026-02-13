# GraphQL Monitor Module - Query Variables Tests

provider "hyperping" {
  api_key = "test_mock_api_key_for_plan_only"
}

variables {
  endpoint = "https://api.example.com/graphql"
}

run "query_without_variables" {
  command = plan

  variables {
    queries = {
      health = {
        query             = "{ health { status } }"
        expected_response = "\"status\":\"ok\""
      }
    }
  }

  assert {
    condition     = jsondecode(hyperping_monitor.graphql_query["health"].request_body).variables == null
    error_message = "Should not include variables field when not specified"
  }
}

run "query_with_simple_variables" {
  command = plan

  variables {
    queries = {
      user = {
        query = "query GetUser($id: ID!) { user(id: $id) { name } }"
        variables = {
          id = "user-123"
        }
        expected_response = "\"name\""
      }
    }
  }

  assert {
    condition     = jsondecode(hyperping_monitor.graphql_query["user"].request_body).variables.id == "user-123"
    error_message = "Should include variables in request body"
  }
}

run "query_with_multiple_variables" {
  command = plan

  variables {
    queries = {
      search = {
        query = <<-GRAPHQL
          query SearchProducts($category: String!, $limit: Int!, $active: Boolean!) {
            products(category: $category, limit: $limit, active: $active) {
              id
              name
            }
          }
        GRAPHQL
        variables = {
          category = "electronics"
          limit    = 10
          active   = true
        }
        expected_response = "\"products\""
      }
    }
  }

  assert {
    condition     = jsondecode(hyperping_monitor.graphql_query["search"].request_body).variables.category == "electronics"
    error_message = "Should include category variable"
  }

  assert {
    condition     = jsondecode(hyperping_monitor.graphql_query["search"].request_body).variables.limit != null
    error_message = "Should include limit variable"
  }

  assert {
    condition     = jsondecode(hyperping_monitor.graphql_query["search"].request_body).variables.active != null
    error_message = "Should include active variable"
  }
}

run "query_with_nested_variables" {
  command = plan

  variables {
    queries = {
      create_user = {
        query = <<-GRAPHQL
          mutation CreateUser($input: CreateUserInput!) {
            createUser(input: $input) {
              id
              name
            }
          }
        GRAPHQL
        variables = {
          input = {
            name  = "Test User"
            email = "test@example.com"
            role  = "admin"
          }
        }
        expected_response = "\"createUser\""
      }
    }
  }

  assert {
    condition     = jsondecode(hyperping_monitor.graphql_query["create_user"].request_body).variables.input.name == "Test User"
    error_message = "Should include nested input variable"
  }

  assert {
    condition     = jsondecode(hyperping_monitor.graphql_query["create_user"].request_body).variables.input.email == "test@example.com"
    error_message = "Should include nested email field"
  }
}

run "multiple_queries_with_different_variables" {
  command = plan

  variables {
    queries = {
      user_query = {
        query = "query GetUser($id: ID!) { user(id: $id) { name } }"
        variables = {
          id = "user-1"
        }
        expected_response = "\"name\""
      }
      product_query = {
        query = "query GetProduct($sku: String!) { product(sku: $sku) { name } }"
        variables = {
          sku = "PROD-001"
        }
        expected_response = "\"name\""
      }
    }
  }

  assert {
    condition     = jsondecode(hyperping_monitor.graphql_query["user_query"].request_body).variables.id == "user-1"
    error_message = "User query should have correct variable"
  }

  assert {
    condition     = jsondecode(hyperping_monitor.graphql_query["product_query"].request_body).variables.sku == "PROD-001"
    error_message = "Product query should have correct variable"
  }
}

run "query_with_array_variables" {
  command = plan

  variables {
    queries = {
      multi_user = {
        query = "query GetUsers($ids: [ID!]!) { users(ids: $ids) { id name } }"
        variables = {
          ids = ["user-1", "user-2", "user-3"]
        }
        expected_response = "\"users\""
      }
    }
  }

  assert {
    condition     = length(jsondecode(hyperping_monitor.graphql_query["multi_user"].request_body).variables.ids) == 3
    error_message = "Should include array variable with 3 items"
  }

  assert {
    condition     = jsondecode(hyperping_monitor.graphql_query["multi_user"].request_body).variables.ids[0] == "user-1"
    error_message = "Should include correct array values"
  }
}

run "mutation_query_with_variables" {
  command = plan

  variables {
    queries = {
      update_status = {
        query = <<-GRAPHQL
          mutation UpdateStatus($id: ID!, $status: String!) {
            updateStatus(id: $id, status: $status) {
              success
            }
          }
        GRAPHQL
        variables = {
          id     = "item-123"
          status = "active"
        }
        expected_response = "\"success\""
      }
    }
  }

  assert {
    condition     = can(regex("^\\s*mutation", jsondecode(hyperping_monitor.graphql_query["update_status"].request_body).query))
    error_message = "Should support mutation queries"
  }

  assert {
    condition     = jsondecode(hyperping_monitor.graphql_query["update_status"].request_body).variables.status == "active"
    error_message = "Mutation should include variables"
  }
}

run "validates_request_body_structure" {
  command = plan

  variables {
    queries = {
      test = {
        query = "{ test { value } }"
        variables = {
          key = "value"
        }
        expected_response = "\"value\""
      }
    }
  }

  assert {
    condition     = can(jsondecode(hyperping_monitor.graphql_query["test"].request_body).query)
    error_message = "Request body should have query field"
  }

  assert {
    condition     = can(jsondecode(hyperping_monitor.graphql_query["test"].request_body).variables)
    error_message = "Request body should have variables field"
  }

  assert {
    condition     = length(keys(jsondecode(hyperping_monitor.graphql_query["test"].request_body))) == 2
    error_message = "Request body should only have query and variables fields"
  }
}
