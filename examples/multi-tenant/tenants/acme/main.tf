# ACME Corporation Tenant
#
# Usage:
#   cd tenants/acme
#   terraform init
#   terraform plan
#   terraform apply

terraform {
  required_version = ">= 1.0"

  required_providers {
    hyperping = {
      source  = "develeap/hyperping"
      version = "~> 1.0"
    }
  }

  # Uncomment for remote state
  # backend "s3" {
  #   bucket = "hyperping-terraform-state"
  #   key    = "tenants/acme/terraform.tfstate"
  #   region = "us-east-1"
  # }
}

provider "hyperping" {
  # Uses HYPERPING_API_KEY environment variable
}

# Load tenant configuration from module
module "tenant" {
  source = "../../modules/tenant"

  tenant_id   = var.tenant_id
  tenant_name = var.tenant_name
  active      = var.active
  status_page = var.status_page
  monitors    = var.monitors
  tags        = var.tags

  # Reference shared monitors if needed
  shared_monitor_ids = var.shared_monitor_ids
}

# Outputs for registry integration
output "registry_data" {
  description = "Complete tenant data for external integrations"
  value       = module.tenant.registry_data
}

output "monitor_ids" {
  description = "All monitor IDs"
  value       = module.tenant.monitor_ids
}
