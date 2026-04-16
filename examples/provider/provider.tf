terraform {
  required_version = ">= 1.0"

  required_providers {
    hyperping = {
      source  = "develeap/hyperping"
      version = "~> 1.0"
    }
  }
}

# Configure the Hyperping Provider
# API key can be set via HYPERPING_API_KEY environment variable
provider "hyperping" {
  # api_key = "sk_..." # Or use HYPERPING_API_KEY env var
}
