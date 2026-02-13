#!/bin/bash
# Cron Healthcheck Integration Examples
#
# This script demonstrates how to integrate Hyperping healthcheck ping URLs
# into actual cron jobs and scripts.

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}=== Hyperping Cron Healthcheck Integration Examples ===${NC}\n"

# ------------------------------------------------------------------------------
# Example 1: Extract Ping URLs from Terraform
# ------------------------------------------------------------------------------
echo -e "${YELLOW}Example 1: Extracting Ping URLs${NC}"
echo "Run these commands to get ping URLs from Terraform:"
echo ""
echo "  # Get all ping URLs as JSON"
echo "  terraform output -json ping_urls"
echo ""
echo "  # Get specific ping URL"
echo "  terraform output -raw backup_ping_url"
echo ""
echo "  # Save to secure location"
echo "  terraform output -raw backup_ping_url > /etc/cron-secrets/backup-ping-url"
echo "  chmod 600 /etc/cron-secrets/backup-ping-url"
echo ""

# ------------------------------------------------------------------------------
# Example 2: Simple Bash Script Integration
# ------------------------------------------------------------------------------
echo -e "${YELLOW}Example 2: Bash Script Integration${NC}"
cat << 'EOF'

#!/bin/bash
# /usr/local/bin/daily-backup.sh

PING_URL="https://ping.hyperping.io/YOUR_HEALTHCHECK_ID"

# Run backup
/usr/local/bin/backup.sh

# Ping Hyperping on success
if [ $? -eq 0 ]; then
    curl -fsS --retry 3 --max-time 10 "$PING_URL" > /dev/null
    echo "Backup completed and Hyperping notified"
else
    echo "Backup failed - Hyperping will alert" >&2
    exit 1
fi
EOF
echo ""

# ------------------------------------------------------------------------------
# Example 3: Secure Ping URL Loading
# ------------------------------------------------------------------------------
echo -e "${YELLOW}Example 3: Secure Ping URL Loading${NC}"
cat << 'EOF'

#!/bin/bash
# /usr/local/bin/secure-backup.sh

# Load ping URL from secure file
PING_URL=$(cat /etc/cron-secrets/backup-ping-url)

if [ -z "$PING_URL" ]; then
    echo "ERROR: Ping URL not configured" >&2
    exit 1
fi

# Run backup
/usr/local/bin/backup.sh

# Notify on success
if [ $? -eq 0 ]; then
    curl -fsS --retry 3 "$PING_URL" > /dev/null || {
        echo "WARNING: Failed to ping Hyperping" >&2
    }
fi
EOF
echo ""

# ------------------------------------------------------------------------------
# Example 4: Python Script Integration
# ------------------------------------------------------------------------------
echo -e "${YELLOW}Example 4: Python Script Integration${NC}"
cat << 'EOF'

#!/usr/bin/env python3
# /usr/local/bin/data-sync.py

import os
import sys
import requests
import logging

PING_URL = os.environ.get("HYPERPING_URL")

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

def notify_success():
    """Ping Hyperping on successful job completion"""
    if not PING_URL:
        logger.warning("HYPERPING_URL not set, skipping notification")
        return

    try:
        response = requests.get(PING_URL, timeout=10)
        response.raise_for_status()
        logger.info("Successfully notified Hyperping")
    except requests.RequestException as e:
        logger.error(f"Failed to notify Hyperping: {e}")

def main():
    try:
        # Your job logic here
        logger.info("Starting data sync...")

        # ... do work ...

        logger.info("Data sync completed successfully")
        notify_success()

    except Exception as e:
        logger.error(f"Job failed: {e}")
        sys.exit(1)

if __name__ == "__main__":
    main()
EOF
echo ""

# ------------------------------------------------------------------------------
# Example 5: Crontab Examples
# ------------------------------------------------------------------------------
echo -e "${YELLOW}Example 5: Crontab Integration${NC}"
cat << 'EOF'

# /etc/crontab or user crontab (crontab -e)

# Method 1: Direct URL in crontab (less secure)
0 2 * * * root /usr/local/bin/backup.sh && curl -fsS "https://ping.hyperping.io/abc123" > /dev/null

# Method 2: Environment variable
BACKUP_PING_URL=https://ping.hyperping.io/abc123
0 2 * * * root /usr/local/bin/backup.sh && curl -fsS "$BACKUP_PING_URL" > /dev/null

# Method 3: Script with embedded ping
0 2 * * * root /usr/local/bin/backup-and-ping.sh

# Method 4: Load from file
0 2 * * * root /usr/local/bin/backup.sh && curl -fsS "$(cat /etc/cron-secrets/backup-ping)" > /dev/null

# Multiple jobs with different ping URLs
0 2 * * * root /usr/local/bin/db-backup.sh && curl -fsS "$DB_BACKUP_PING" > /dev/null
0 3 * * * root /usr/local/bin/log-rotation.sh && curl -fsS "$LOG_ROTATION_PING" > /dev/null
EOF
echo ""

# ------------------------------------------------------------------------------
# Example 6: Advanced Error Handling
# ------------------------------------------------------------------------------
echo -e "${YELLOW}Example 6: Advanced Error Handling${NC}"
cat << 'EOF'

#!/bin/bash
# /usr/local/bin/robust-backup.sh

PING_URL="https://ping.hyperping.io/YOUR_ID"
MAX_RETRIES=3
RETRY_DELAY=5

backup_and_notify() {
    local exit_code

    # Run backup
    /usr/local/bin/backup.sh
    exit_code=$?

    if [ $exit_code -eq 0 ]; then
        # Success - ping Hyperping with retry logic
        for i in $(seq 1 $MAX_RETRIES); do
            if curl -fsS --max-time 10 "$PING_URL" > /dev/null 2>&1; then
                echo "Backup successful and Hyperping notified"
                return 0
            else
                echo "Attempt $i: Failed to ping Hyperping, retrying..." >&2
                sleep $RETRY_DELAY
            fi
        done
        echo "WARNING: Backup succeeded but failed to notify Hyperping" >&2
        return 0  # Don't fail the job
    else
        echo "ERROR: Backup failed with exit code $exit_code" >&2
        # Don't ping Hyperping - let it alert
        return $exit_code
    fi
}

backup_and_notify
EOF
echo ""

# ------------------------------------------------------------------------------
# Example 7: Using with systemd timers
# ------------------------------------------------------------------------------
echo -e "${YELLOW}Example 7: systemd Timer Integration${NC}"
cat << 'EOF'

# /etc/systemd/system/backup.service
[Unit]
Description=Daily Backup Job
After=network.target

[Service]
Type=oneshot
ExecStart=/usr/local/bin/backup.sh
ExecStartPost=/bin/bash -c 'curl -fsS "https://ping.hyperping.io/abc123" > /dev/null || true'
User=backup
Environment="PATH=/usr/local/bin:/usr/bin"

[Install]
WantedBy=multi-user.target

---

# /etc/systemd/system/backup.timer
[Unit]
Description=Daily Backup Timer
Requires=backup.service

[Timer]
OnCalendar=daily
OnCalendar=02:00
Persistent=true

[Install]
WantedBy=timers.target

---

# Enable and start
sudo systemctl enable backup.timer
sudo systemctl start backup.timer
sudo systemctl list-timers backup.timer
EOF
echo ""

# ------------------------------------------------------------------------------
# Example 8: Testing Ping URLs
# ------------------------------------------------------------------------------
echo -e "${YELLOW}Example 8: Testing Ping URLs${NC}"
cat << 'EOF'

#!/bin/bash
# test-ping.sh - Test all ping URLs

# Get ping URLs from Terraform
PING_URLS=$(terraform output -json ping_urls | jq -r '.[]')

echo "Testing all ping URLs..."

for url in $PING_URLS; do
    echo -n "Testing $url ... "
    if curl -fsS --max-time 5 "$url" > /dev/null 2>&1; then
        echo "✓ OK"
    else
        echo "✗ FAILED"
    fi
done
EOF
echo ""

# ------------------------------------------------------------------------------
# Example 9: AWS Systems Manager Parameter Store Integration
# ------------------------------------------------------------------------------
echo -e "${YELLOW}Example 9: AWS SSM Parameter Store${NC}"
cat << 'EOF'

# Store ping URL in SSM Parameter Store
aws ssm put-parameter \
  --name "/prod/cron/backup-ping-url" \
  --value "$(terraform output -raw backup_ping_url)" \
  --type "SecureString" \
  --description "Hyperping healthcheck URL for backup job"

# Retrieve in script
#!/bin/bash
PING_URL=$(aws ssm get-parameter \
  --name "/prod/cron/backup-ping-url" \
  --with-decryption \
  --query "Parameter.Value" \
  --output text)

/usr/local/bin/backup.sh && curl -fsS "$PING_URL" > /dev/null
EOF
echo ""

# ------------------------------------------------------------------------------
# Example 10: Docker Container Integration
# ------------------------------------------------------------------------------
echo -e "${YELLOW}Example 10: Docker Container${NC}"
cat << 'EOF'

# Dockerfile
FROM ubuntu:22.04
RUN apt-get update && apt-get install -y curl
COPY backup.sh /usr/local/bin/
ENV HYPERPING_URL=""
CMD ["/usr/local/bin/backup.sh"]

# backup.sh
#!/bin/bash
# ... backup logic ...
[ -n "$HYPERPING_URL" ] && curl -fsS "$HYPERPING_URL" > /dev/null

# Run with environment variable
docker run --rm \
  -e HYPERPING_URL="https://ping.hyperping.io/abc123" \
  backup-container

# Or use docker-compose
services:
  backup:
    image: backup-container
    environment:
      HYPERPING_URL: https://ping.hyperping.io/abc123
EOF
echo ""

echo -e "${GREEN}=== Integration Examples Complete ===${NC}"
echo ""
echo "Next steps:"
echo "1. Deploy the module with: terraform apply"
echo "2. Extract ping URLs with: terraform output ping_urls"
echo "3. Integrate ping URLs into your cron scripts"
echo "4. Test manually before production deployment"
echo ""
