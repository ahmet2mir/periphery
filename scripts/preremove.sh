#!/bin/bash
# Pre-removal script for Herald

set -e

# Stop and disable service if running
if systemctl is-active --quiet herald; then
    systemctl stop herald
fi

if systemctl is-enabled --quiet herald; then
    systemctl disable herald
fi

echo "Herald service stopped and disabled"
