terraform {
  required_providers {
    hyperping = {
      source = "develeap/hyperping"
    }
  }
}

# Configure the Hyperping Provider
# API key can be set via HYPERPING_API_KEY environment variable
provider "hyperping" {
  # api_key = "hp_..." # Or use HYPERPING_API_KEY env var
}
