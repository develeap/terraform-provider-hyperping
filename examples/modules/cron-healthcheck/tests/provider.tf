# Provider configuration for tests
#
# Set HYPERPING_API_KEY environment variable before running tests

terraform {
  required_version = ">= 1.0"

  required_providers {
    hyperping = {
      source  = "develeap/hyperping"
      version = ">= 1.0"
    }
  }
}

provider "hyperping" {
  # API key will be read from HYPERPING_API_KEY environment variable
}
