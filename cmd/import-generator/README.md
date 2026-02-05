# Hyperping Import Generator

Generate Terraform import commands and HCL configurations from existing Hyperping resources.

## Why This Tool?

`terraform import` requires you to write HCL configuration *before* importing. This tool solves the chicken-and-egg problem by:

1. **Discovering** all your existing Hyperping resources via API
2. **Generating** the HCL configuration blocks you need
3. **Generating** the `terraform import` commands to run

## Quick Start

```bash
export HYPERPING_API_KEY="your_api_key"

# Generate both import commands and HCL
go run ./cmd/import-generator

# Save to files
go run ./cmd/import-generator -format=hcl -output=resources.tf
go run ./cmd/import-generator -format=import -output=import.sh
```

## Options

| Flag | Default | Description |
|------|---------|-------------|
| `-format` | `both` | Output: `import`, `hcl`, or `both` |
| `-output` | stdout | Output file path |
| `-resources` | `all` | Filter: `monitors`, `healthchecks`, `statuspages`, `incidents`, `maintenance`, `outages` |
| `-prefix` | (none) | Prefix for resource names (e.g., `prod_`) |

## Example Output

**Import commands:**
```bash
terraform import hyperping_monitor.production_api "mon_abc123"
terraform import hyperping_statuspage.main_status "sp_xyz789"
```

**HCL configuration:**
```hcl
resource "hyperping_monitor" "production_api" {
  name            = "Production API"
  url             = "https://api.example.com/health"
  protocol        = "http"
  check_frequency = 60
  regions         = ["virginia", "london"]
}
```

## Workflow

```bash
# 1. Generate files
go run ./cmd/import-generator -format=hcl -output=resources.tf
go run ./cmd/import-generator -format=import -output=import.sh

# 2. Review and adjust resources.tf as needed

# 3. Run imports
chmod +x import.sh && ./import.sh

# 4. Verify
terraform plan  # Should show no changes
```
