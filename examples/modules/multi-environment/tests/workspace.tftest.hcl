# Multi-Environment Module - Workspace Integration Tests

provider "hyperping" {
  api_key = "test_mock_api_key_for_plan_only"
}

variables {
  service_name       = "WorkspaceAPI"
  use_workspace_name = true

  environments = {
    current = {
      url       = "https://api.example.com/health"
      frequency = 60
    }
  }
}

run "uses_workspace_name_when_enabled" {
  command = plan

  # Note: In tests, workspace is typically "default"
  # In real usage, this would be "dev", "staging", "prod" etc.
  assert {
    condition     = hyperping_monitor.environment["current"].name == "[DEFAULT] WorkspaceAPI"
    error_message = "Should use workspace name (DEFAULT in tests) when use_workspace_name is true"
  }
}

run "creates_single_monitor_for_workspace" {
  command = plan

  assert {
    condition     = length(hyperping_monitor.environment) == 1
    error_message = "Should create 1 monitor for current workspace"
  }

  assert {
    condition     = contains(keys(hyperping_monitor.environment), "current")
    error_message = "Monitor key should be 'current'"
  }
}
