#!/bin/bash
# Example migration workflow with dry-run preview
# This script demonstrates the recommended approach for migrating from Better Stack

set -e

echo "================================================"
echo "Hyperping Migration Workflow Example"
echo "================================================"
echo

# Check prerequisites
if [ -z "$BETTERSTACK_API_TOKEN" ]; then
    echo "Error: BETTERSTACK_API_TOKEN environment variable not set"
    echo "Export your Better Stack API token:"
    echo "  export BETTERSTACK_API_TOKEN='your_token'"
    exit 1
fi

# Step 1: Dry-run preview
echo "Step 1: Running dry-run preview..."
echo "================================================"
echo

migrate-betterstack --dry-run --format > dry-run-report.json

# Extract compatibility score
SCORE=$(jq -r '.compatibility.overall_score' dry-run-report.json)
COMPLEXITY=$(jq -r '.compatibility.complexity' dry-run-report.json)
WARNING_COUNT=$(jq -r '.compatibility.warning_count' dry-run-report.json)

echo "Compatibility Score: $SCORE%"
echo "Complexity: $COMPLEXITY"
echo "Warnings: $WARNING_COUNT"
echo

# Step 2: Decision based on score
if (( $(echo "$SCORE < 75" | bc -l) )); then
    echo "⚠️  Compatibility score is below 75%"
    echo "Review dry-run-report.json carefully before proceeding"
    echo
    echo "Run 'migrate-betterstack --dry-run --verbose' for detailed analysis"
    exit 1
fi

if (( $(echo "$SCORE < 90" | bc -l) )); then
    echo "⚠️  Migration will require some manual steps"
    echo "Warnings found: $WARNING_COUNT"
    echo
    read -p "Do you want to review warnings? (y/n) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        migrate-betterstack --dry-run --verbose | less
        echo
        read -p "Proceed with migration? (y/n) " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            echo "Migration cancelled"
            exit 0
        fi
    fi
fi

# Step 3: Execute migration
echo
echo "Step 2: Executing migration..."
echo "================================================"
echo

migrate-betterstack \
    --betterstack-token "$BETTERSTACK_API_TOKEN" \
    --output hyperping.tf \
    --import-script import.sh \
    --report migration-report.json \
    --manual-steps manual-steps.md

echo
echo "✅ Migration completed successfully!"
echo
echo "Generated files:"
echo "  - hyperping.tf         (Terraform configuration)"
echo "  - import.sh            (Import script)"
echo "  - migration-report.json (Detailed report)"
echo "  - manual-steps.md      (Manual steps required)"
echo

# Step 4: Next steps
echo "Next steps:"
echo "================================================"
echo
echo "1. Review generated files:"
echo "   cat hyperping.tf"
echo "   cat manual-steps.md"
echo
echo "2. Initialize Terraform:"
echo "   terraform init"
echo
echo "3. Review the plan:"
echo "   terraform plan"
echo
echo "4. Apply the configuration:"
echo "   terraform apply"
echo
echo "5. Complete manual steps (if any):"
echo "   See manual-steps.md"
echo

echo "Migration workflow complete!"
