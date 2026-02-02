# Multi-Tenant Terraform Structure

> ⚠️ **Advanced Pattern** - This example demonstrates sophisticated multi-environment management patterns. If you're just getting started, see the simpler examples in [`examples/resources/`](../resources/) first.

This directory demonstrates a recommended structure for managing multiple environments or tenants with Terraform modules and workspaces.

## Directory Structure

```
terraform-hyperping-tenants/
├── modules/
│   └── tenant/
│       ├── main.tf           # Monitor resources
│       ├── variables.tf      # Input variables
│       ├── outputs.tf        # Output values
│       └── versions.tf       # Provider requirements
├── shared/
│   ├── main.tf               # Shared monitors (Auth, CDN, etc.)
│   └── outputs.tf            # Shared monitor IDs
├── tenants/
│   ├── acme/
│   │   ├── main.tf           # Module call
│   │   ├── terraform.tfvars  # Tenant-specific values
│   │   └── backend.tf        # State configuration
│   ├── globex/
│   │   └── ...
│   └── .../
├── scripts/
│   ├── apply-all.sh          # Apply all tenants
│   ├── plan-all.sh           # Plan all tenants
│   └── import-tenant.sh      # Import existing tenant
└── config/
    └── tenants.yaml          # Optional: Generate tfvars from YAML
```

## Usage

### Initialize a tenant workspace
```bash
cd tenants/acme
terraform init
terraform workspace new acme  # Or use existing
terraform plan
terraform apply
```

### Apply all tenants
```bash
./scripts/apply-all.sh
```

### Import existing resources
```bash
./scripts/import-tenant.sh acme
```
