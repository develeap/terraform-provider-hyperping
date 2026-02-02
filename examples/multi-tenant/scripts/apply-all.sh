#!/usr/bin/env bash
# Apply Terraform for all tenants
#
# Usage:
#   ./scripts/apply-all.sh [plan|apply|destroy]
#
# Environment:
#   HYPERPING_API_KEY - Required
#   TF_PARALLELISM    - Optional, default 5

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"
TENANTS_DIR="$ROOT_DIR/tenants"

ACTION="${1:-plan}"
PARALLELISM="${TF_PARALLELISM:-5}"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() { echo -e "${GREEN}[INFO]${NC} $*"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $*"; }
log_error() { echo -e "${RED}[ERROR]${NC} $*"; }

# Check API key
if [[ -z "${HYPERPING_API_KEY:-}" ]]; then
    log_error "HYPERPING_API_KEY not set"
    exit 1
fi

# Get list of tenants
tenants=()
for dir in "$TENANTS_DIR"/*/; do
    if [[ -f "$dir/main.tf" ]]; then
        tenants+=("$(basename "$dir")")
    fi
done

if [[ ${#tenants[@]} -eq 0 ]]; then
    log_warn "No tenant directories found in $TENANTS_DIR"
    exit 0
fi

log_info "Found ${#tenants[@]} tenant(s): ${tenants[*]}"
log_info "Action: $ACTION"
echo ""

# Track results
declare -A results
failed=0

for tenant in "${tenants[@]}"; do
    tenant_dir="$TENANTS_DIR/$tenant"
    log_info "Processing tenant: $tenant"

    cd "$tenant_dir"

    # Initialize if needed
    if [[ ! -d ".terraform" ]]; then
        log_info "  Initializing Terraform..."
        terraform init -input=false -no-color > /dev/null 2>&1
    fi

    # Run action
    case "$ACTION" in
        plan)
            if terraform plan -input=false -no-color -detailed-exitcode > /tmp/tf-plan-$tenant.log 2>&1; then
                results[$tenant]="no-changes"
                log_info "  ✓ No changes"
            elif [[ $? -eq 2 ]]; then
                results[$tenant]="has-changes"
                log_warn "  ⚡ Has changes (see /tmp/tf-plan-$tenant.log)"
            else
                results[$tenant]="error"
                log_error "  ✗ Error (see /tmp/tf-plan-$tenant.log)"
                ((failed++))
            fi
            ;;
        apply)
            if terraform apply -input=false -no-color -auto-approve -parallelism="$PARALLELISM" > /tmp/tf-apply-$tenant.log 2>&1; then
                results[$tenant]="applied"
                log_info "  ✓ Applied successfully"
            else
                results[$tenant]="error"
                log_error "  ✗ Apply failed (see /tmp/tf-apply-$tenant.log)"
                ((failed++))
            fi
            ;;
        destroy)
            log_warn "  Destroying tenant $tenant..."
            if terraform destroy -input=false -no-color -auto-approve > /tmp/tf-destroy-$tenant.log 2>&1; then
                results[$tenant]="destroyed"
                log_info "  ✓ Destroyed"
            else
                results[$tenant]="error"
                log_error "  ✗ Destroy failed"
                ((failed++))
            fi
            ;;
        *)
            log_error "Unknown action: $ACTION"
            log_info "Usage: $0 [plan|apply|destroy]"
            exit 1
            ;;
    esac

    cd "$ROOT_DIR"
    echo ""
done

# Summary
echo "========================================"
echo "Summary: ${#tenants[@]} tenants processed"
for tenant in "${tenants[@]}"; do
    status="${results[$tenant]:-unknown}"
    case "$status" in
        applied|no-changes|destroyed)
            echo -e "  ${GREEN}✓${NC} $tenant: $status"
            ;;
        has-changes)
            echo -e "  ${YELLOW}⚡${NC} $tenant: $status"
            ;;
        *)
            echo -e "  ${RED}✗${NC} $tenant: $status"
            ;;
    esac
done
echo "========================================"

exit $failed
