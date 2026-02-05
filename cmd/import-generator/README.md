# Hyperping Import Generator

Generate Terraform import commands and HCL configurations from existing Hyperping resources.

## Usage

```bash
# Set your API key
export HYPERPING_API_KEY="sk_your_api_key"

# Generate both import commands and HCL (default)
go run ./cmd/import-generator

# Generate only import commands
go run ./cmd/import-generator -format=import

# Generate only HCL configuration
go run ./cmd/import-generator -format=hcl

# Save to file
go run ./cmd/import-generator -output=imported.tf

# Import specific resource types
go run ./cmd/import-generator -resources=monitors,healthchecks

# Add prefix to resource names
go run ./cmd/import-generator -prefix=prod_
```

## Options

| Flag | Default | Description |
|------|---------|-------------|
| `-format` | `both` | Output format: `import`, `hcl`, or `both` |
| `-output` | stdout | Output file path |
| `-resources` | `all` | Comma-separated list: `monitors`, `healthchecks`, `statuspages`, `incidents`, `maintenance`, `outages` |
| `-prefix` | (none) | Prefix for Terraform resource names |
| `-base-url` | `https://api.hyperping.io` | Hyperping API base URL |

## Example Output

### Import Commands

```bash
terraform import hyperping_monitor.production_api "mon_abc123"
terraform import hyperping_monitor.database_health "mon_def456"
terraform import hyperping_statuspage.main_status "sp_xyz789"
```

### HCL Configuration

```hcl
resource "hyperping_monitor" "production_api" {
  name            = "Production API"
  url             = "https://api.example.com/health"
  protocol        = "http"
  check_frequency = 60
  regions         = ["virginia", "london"]
}

resource "hyperping_statuspage" "main_status" {
  name             = "Main Status"
  hosted_subdomain = "status"

  settings = {
    name      = "Main Status"
    languages = ["en"]
  }
}
```

## Workflow

1. **Export existing resources:**
   ```bash
   go run ./cmd/import-generator -output=import.sh -format=import
   go run ./cmd/import-generator -output=resources.tf -format=hcl
   ```

2. **Review and adjust the generated HCL** as needed

3. **Run the import commands:**
   ```bash
   chmod +x import.sh
   ./import.sh
   ```

4. **Verify no drift:**
   ```bash
   terraform plan
   ```

## Notes

- Resource names are automatically converted to valid Terraform identifiers
- Names with special characters become underscores (e.g., "My API Monitor" → `my_api_monitor`)
- The generated HCL includes only configurable fields (not computed/read-only fields)
- Status page sections are noted but need manual configuration after import
- Outages are mostly read-only; the generated HCL is minimal
