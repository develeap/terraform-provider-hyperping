# Examples

This directory contains examples for using the Hyperping Terraform provider.

## Available Examples

| Example | Description |
|---------|-------------|
| [provider](./provider/) | Basic provider configuration |
| [resources/hyperping_monitor](./resources/hyperping_monitor/) | Monitor resource examples |
| [resources/hyperping_incident](./resources/hyperping_incident/) | Incident resource examples |
| [resources/hyperping_maintenance](./resources/hyperping_maintenance/) | Maintenance window examples |
| [resources/hyperping_statuspage](./resources/hyperping_statuspage/) | Status page configuration examples |
| [data-sources/hyperping_monitors](./data-sources/hyperping_monitors/) | Monitors data source examples |
| [complete](./complete/) | Complete end-to-end example with all features |
| [advanced-patterns](./advanced-patterns/) | Production-ready patterns: dynamic monitors, regional redundancy, conditional resources |
| [multi-tenant](./multi-tenant/) | Multi-tenant monitoring setup with modules |

## Usage

Each example can be run independently:

```bash
cd examples/<example-name>
export HYPERPING_API_KEY="sk_your_api_key"
terraform init
terraform plan
terraform apply
```

## Documentation Generation

These examples are used by `terraform-plugin-docs` to generate provider documentation:

- `provider/provider.tf` - Used for provider index page
- `resources/<name>/resource.tf` - Used for resource documentation pages
- `data-sources/<name>/data-source.tf` - Used for data source documentation pages

The documentation tool looks for these specific file paths and names. Additional files in example directories are ignored by the documentation tool but can be useful for testing.

## Testing Examples

To validate all examples:

```bash
for dir in examples/*/; do
  if [ -f "$dir/main.tf" ] || [ -f "$dir/resource.tf" ] || [ -f "$dir/data-source.tf" ] || [ -f "$dir/provider.tf" ]; then
    echo "Validating $dir"
    (cd "$dir" && terraform init -backend=false && terraform validate)
  fi
done
```
