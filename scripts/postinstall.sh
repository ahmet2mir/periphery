#!/bin/bash
# Post-installation script for Herald

set -e

# Create herald user if it doesn't exist
if ! id -u herald >/dev/null 2>&1; then
    useradd --system --no-create-home --shell /bin/false herald
fi

# Create config directory
mkdir -p /etc/herald
chown herald:herald /etc/herald

# Create systemd service file
cat > /etc/systemd/system/herald.service << 'EOF'
[Unit]
Description=Herald BGP Anycast Service
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=herald
Group=herald
ExecStart=/usr/bin/herald --config /etc/herald/config.yaml
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal
SyslogIdentifier=herald

# Security
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/etc/herald

[Install]
WantedBy=multi-user.target
EOF

# Reload systemd
systemctl daemon-reload

echo "Herald installed successfully!"
echo ""
echo "Next steps:"
echo "1. Create your configuration: /etc/herald/config.yaml"
echo "   Example: cp /etc/herald/config.yaml.example /etc/herald/config.yaml"
echo "2. Enable service: systemctl enable herald"
echo "3. Start service: systemctl start herald"
echo "4. Check status: systemctl status herald"
