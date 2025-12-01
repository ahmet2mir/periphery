#!/bin/bash
# Pre-removal script for Periphery

set -e

# Stop and disable service if running
if systemctl is-active --quiet periphery; then
    systemctl stop periphery
fi

if systemctl is-enabled --quiet periphery; then
    systemctl disable periphery
fi

echo "Periphery service stopped and disabled"
