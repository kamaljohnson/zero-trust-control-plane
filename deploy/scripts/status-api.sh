#!/bin/bash
# status-api.sh: Returns deployment status as JSON
# Called by nginx via fastcgi or as a simple script

DEPLOY_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
STATUS_FILE="$DEPLOY_DIR/.deployment-status.json"

# Default status if file doesn't exist
if [ ! -f "$STATUS_FILE" ]; then
    cat << 'EOF'
{
  "status": "unknown",
  "message": "Deployment status not available",
  "timestamp": "",
  "steps": []
}
EOF
    exit 0
fi

# Return the status file
cat "$STATUS_FILE"
