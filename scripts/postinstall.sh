#!/bin/bash
# Post-installation script for Periphery

set -e

# Create periphery user if it doesn't exist
if ! id -u periphery >/dev/null 2>&1; then
    useradd --system --no-create-home --shell /bin/false periphery
fi

# Create config directory
mkdir -p /etc/periphery
chown periphery:periphery /etc/periphery

# Create systemd service file
cat > /etc/systemd/system/periphery.service << 'EOF'
[Unit]
Description=Periphery BGP Anycast Service
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=periphery
Group=periphery
ExecStart=/usr/bin/periphery --config /etc/periphery/config.yaml
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal
SyslogIdentifier=periphery

# Security
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/etc/periphery

[Install]
WantedBy=multi-user.target
EOF

# Reload systemd
systemctl daemon-reload

echo "Periphery installed successfully!"
echo ""
echo "Next steps:"
echo "1. Create your configuration: /etc/periphery/config.yaml"
echo "   Example: cp /etc/periphery/config.yaml.example /etc/periphery/config.yaml"
echo "2. Enable service: systemctl enable periphery"
echo "3. Start service: systemctl start periphery"
echo "4. Check status: systemctl status periphery"
