# GitHub Actions Workflows for Terraform Hyperping

Ready-to-use GitHub Actions workflows for managing Hyperping resources with Terraform.

## Quick Start

### 1. Copy the Workflow

Copy `terraform-hyperping.yml` to your repository:

```bash
mkdir -p .github/workflows
cp terraform-hyperping.yml .github/workflows/
```

### 2. Set Up Secrets

Add your Hyperping API key to repository secrets:

1. Go to **Settings** > **Secrets and variables** > **Actions**
2. Click **New repository secret**
3. Name: `HYPERPING_API_KEY`
4. Value: Your API key (starts with `sk_`)

### 3. Configure State Backend (Recommended)

For team environments, configure a remote backend. Add to your `main.tf`:

```hcl
terraform {
  backend "s3" {
    bucket         = "your-terraform-state-bucket"
    key            = "hyperping/terraform.tfstate"
    region         = "us-east-1"
    dynamodb_table = "terraform-locks"
    encrypt        = true
  }
}
```

Or use Terraform Cloud:

```hcl
terraform {
  cloud {
    organization = "your-org"
    workspaces {
      name = "hyperping-production"
    }
  }
}
```

## Workflow Behavior

### Triggers

| Event | Action |
|-------|--------|
| Push to `main` | Plan + Apply (if changes) |
| Pull Request | Plan only (comment on PR) |
| Manual (workflow_dispatch) | Choose: plan, apply, or destroy |

### Jobs

| Job | Purpose | When |
|-----|---------|------|
| `validate` | Format check, syntax validation | Always |
| `plan` | Generate execution plan | Always |
| `apply` | Apply changes | Main branch or manual |
| `destroy` | Destroy all resources | Manual only |
| `drift-check` | Detect configuration drift | Scheduled (disabled by default) |

## Configuration Options

### Change Parallelism

Edit the environment variables to adjust rate limiting:

```yaml
env:
  TF_CLI_ARGS_plan: '-parallelism=5'   # Reduce for large deployments
  TF_CLI_ARGS_apply: '-parallelism=5'
```

### Enable Drift Detection

Uncomment the schedule trigger and enable the drift-check job:

```yaml
on:
  schedule:
    - cron: '0 3 * * *'  # Daily at 3 AM UTC

jobs:
  drift-check:
    if: github.event_name == 'schedule'  # Change from 'false'
```

### Require Approval for Apply

The workflow uses the `production` environment. Configure approval in:

**Settings** > **Environments** > **production** > **Required reviewers**

### Filter Paths

By default, the workflow only runs when `.tf` files change:

```yaml
paths:
  - '**.tf'
  - '.github/workflows/terraform-hyperping.yml'
```

## Example Repository Structure

```
your-repo/
├── .github/
│   └── workflows/
│       └── terraform-hyperping.yml
├── main.tf
├── monitors.tf
├── statuspages.tf
├── variables.tf
├── outputs.tf
└── README.md
```

## Workflow Files

| File | Description |
|------|-------------|
| `terraform-hyperping.yml` | Full CI/CD workflow with plan, apply, destroy, and drift detection |

## Security Best Practices

1. **Never commit API keys** - Always use secrets
2. **Enable branch protection** - Require PR reviews before merging to main
3. **Use environment protection** - Require approval for production applies
4. **Limit secret access** - Only allow workflows on protected branches to access secrets

## Troubleshooting

### Error: 429 Too Many Requests

Reduce parallelism in the workflow:

```yaml
env:
  TF_CLI_ARGS_plan: '-parallelism=2'
  TF_CLI_ARGS_apply: '-parallelism=2'
```

### Error: State Lock Timeout

Another workflow is running. Wait for it to complete or manually unlock:

```bash
terraform force-unlock LOCK_ID
```

### Plan Comment Too Long

GitHub has a 65,535 character limit for comments. The workflow truncates long plans automatically.

## Additional Resources

- [Rate Limits Guide](../../docs/guides/rate-limits.md)
- [Migration Guide](../../docs/guides/migration.md)
- [Troubleshooting Guide](../../docs/TROUBLESHOOTING.md)
