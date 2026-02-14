#!/bin/bash
# Copyright (c) 2026 Develeap
# SPDX-License-Identifier: MPL-2.0

set -e

# E2E Test Runner Script
# Runs end-to-end tests for migration tools with proper setup and cleanup

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

# Default values
TEST_TIMEOUT="30m"
TEST_PATTERN=""
SHORT_MODE=""
VERBOSE="-v"
PARALLEL=""

# Usage information
usage() {
    cat <<EOF
Usage: $0 [OPTIONS]

Run end-to-end tests for migration tools

OPTIONS:
    -p, --pattern PATTERN    Run tests matching pattern (e.g., SmallMigration, BetterStack)
    -s, --short              Run in short mode (skip slow tests)
    -t, --timeout DURATION   Set test timeout (default: 30m)
    -q, --quiet              Reduce output verbosity
    -j, --parallel N         Run N tests in parallel
    -h, --help               Show this help message

EXAMPLES:
    # Run all E2E tests
    $0

    # Run only small migration tests
    $0 -p SmallMigration

    # Run Better Stack tests only
    $0 -p BetterStack

    # Run in short mode with 15 minute timeout
    $0 -s -t 15m

    # Run tests in parallel
    $0 -j 4

ENVIRONMENT VARIABLES:
    Required for all tests:
        HYPERPING_API_KEY            Hyperping API key

    Required for specific platform tests:
        BETTERSTACK_API_TOKEN        Better Stack API token
        UPTIMEROBOT_API_KEY          UptimeRobot API key
        PINGDOM_API_KEY              Pingdom API key

EOF
    exit 1
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -p|--pattern)
            TEST_PATTERN="$2"
            shift 2
            ;;
        -s|--short)
            SHORT_MODE="-short"
            shift
            ;;
        -t|--timeout)
            TEST_TIMEOUT="$2"
            shift 2
            ;;
        -q|--quiet)
            VERBOSE=""
            shift
            ;;
        -j|--parallel)
            PARALLEL="-p $2"
            shift 2
            ;;
        -h|--help)
            usage
            ;;
        *)
            echo -e "${RED}Unknown option: $1${NC}"
            usage
            ;;
    esac
done

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo -e "${RED}Error: Go is not installed${NC}"
    exit 1
fi

# Check if Terraform is installed
if ! command -v terraform &> /dev/null; then
    echo -e "${RED}Error: Terraform is not installed${NC}"
    exit 1
fi

# Print environment status
echo -e "${BLUE}=== E2E Test Runner ===${NC}"
echo -e "Go version: $(go version)"
echo -e "Terraform version: $(terraform version -json | grep -o '"terraform_version":"[^"]*"' | cut -d'"' -f4)"
echo ""

# Check for API keys
echo -e "${BLUE}=== Checking API Credentials ===${NC}"
MISSING_CREDS=""

if [[ -z "${HYPERPING_API_KEY}" ]]; then
    echo -e "${RED}✗ HYPERPING_API_KEY not set${NC}"
    MISSING_CREDS="yes"
else
    echo -e "${GREEN}✓ HYPERPING_API_KEY set${NC}"
fi

if [[ -z "${BETTERSTACK_API_TOKEN}" ]]; then
    echo -e "${YELLOW}⚠ BETTERSTACK_API_TOKEN not set (Better Stack tests will be skipped)${NC}"
else
    echo -e "${GREEN}✓ BETTERSTACK_API_TOKEN set${NC}"
fi

if [[ -z "${UPTIMEROBOT_API_KEY}" ]]; then
    echo -e "${YELLOW}⚠ UPTIMEROBOT_API_KEY not set (UptimeRobot tests will be skipped)${NC}"
else
    echo -e "${GREEN}✓ UPTIMEROBOT_API_KEY set${NC}"
fi

if [[ -z "${PINGDOM_API_KEY}" ]]; then
    echo -e "${YELLOW}⚠ PINGDOM_API_KEY not set (Pingdom tests will be skipped)${NC}"
else
    echo -e "${GREEN}✓ PINGDOM_API_KEY set${NC}"
fi

echo ""

if [[ -n "${MISSING_CREDS}" ]]; then
    echo -e "${RED}Error: HYPERPING_API_KEY is required for all E2E tests${NC}"
    echo -e "Set it with: export HYPERPING_API_KEY=your_key"
    exit 1
fi

# Change to project root
cd "${PROJECT_ROOT}"

# Build test command
TEST_CMD="go test -tags=e2e ./test/e2e"

if [[ -n "${TEST_PATTERN}" ]]; then
    TEST_CMD="${TEST_CMD} -run ${TEST_PATTERN}"
fi

if [[ -n "${SHORT_MODE}" ]]; then
    TEST_CMD="${TEST_CMD} ${SHORT_MODE}"
fi

if [[ -n "${VERBOSE}" ]]; then
    TEST_CMD="${TEST_CMD} ${VERBOSE}"
fi

if [[ -n "${PARALLEL}" ]]; then
    TEST_CMD="${TEST_CMD} ${PARALLEL}"
fi

TEST_CMD="${TEST_CMD} -timeout=${TEST_TIMEOUT}"

# Print test configuration
echo -e "${BLUE}=== Test Configuration ===${NC}"
echo -e "Command: ${TEST_CMD}"
echo -e "Timeout: ${TEST_TIMEOUT}"
echo -e "Pattern: ${TEST_PATTERN:-all tests}"
echo -e "Short mode: ${SHORT_MODE:-disabled}"
echo -e "Parallel: ${PARALLEL:-sequential}"
echo ""

# Run tests
echo -e "${BLUE}=== Running E2E Tests ===${NC}"
echo ""

START_TIME=$(date +%s)

if eval "${TEST_CMD}"; then
    END_TIME=$(date +%s)
    DURATION=$((END_TIME - START_TIME))
    echo ""
    echo -e "${GREEN}=== Tests Passed ===${NC}"
    echo -e "Duration: ${DURATION} seconds"
    exit 0
else
    END_TIME=$(date +%s)
    DURATION=$((END_TIME - START_TIME))
    echo ""
    echo -e "${RED}=== Tests Failed ===${NC}"
    echo -e "Duration: ${DURATION} seconds"
    echo ""
    echo -e "${YELLOW}Tip: Check test logs above for details${NC}"
    echo -e "${YELLOW}Tip: Run cleanup if resources were left behind:${NC}"
    echo -e "  go test -tags=e2e ./test/e2e -run IdempotentCleanup -v"
    exit 1
fi
