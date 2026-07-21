# Status Page Complete Module - Provider Requirements

terraform {
  required_version = ">= 1.11"

  required_providers {
    hyperping = {
      source  = "develeap/hyperping"
      version = "~> 1.0"
    }
  }
}
