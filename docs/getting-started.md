# Getting Started with Periphery

This guide will help you get Periphery up and running quickly.

## Prerequisites

- Linux system with systemd
- BGP router/peer to establish sessions with
- Go 1.21+ (if building from source)

## Installation

### Option 1: Download Binary

Download the latest release for your platform:

```bash
# Download for Linux AMD64
wget https://github.com/ahmet2mir/periphery/releases/latest/download/periphery_linux_amd64

# Make it executable
chmod +x periphery_linux_amd64

# Move to system path
sudo mv periphery_linux_amd64 /usr/local/bin/periphery

# Verify installation
periphery --version
```

### Option 2: Build from Source

```bash
# Clone repository
git clone https://github.com/ahmet2mir/periphery.git
cd periphery

# Install build dependencies
make setup

# Build
make build

# Binary will be in dist/periphery_linux_amd64_v1/periphery
sudo cp dist/periphery_linux_amd64_v1/periphery /usr/local/bin/
```

## Basic Configuration

Create a configuration file at `/etc/periphery/config.yaml`:

```yaml
# BGP Speaker Configuration
speaker:
  asn: 64600                      # Your AS number
  routerId: "10.0.0.1"           # Your router ID
  gracefulRestartEnabled: true
  gracefulRestartRestartTime: 120

# API Server
api:
  listenAddress: "127.0.0.1"
  listenPort: 50051

# BGP Neighbors
neighbors:
  - address: "10.0.0.254"        # Your BGP peer
    asn: 64599                    # Peer AS number
    ebgpMultihopEnabled: false

# Prefixes to Announce
prefixes:
  - ipAddress: "192.0.2.1/32"    # Your anycast IP
    communities:
      - '65000:100'
    nextHop: "10.0.0.1"

    # Service to Monitor
    service:
      name: nginx.service
      type: systemd

    # Health Check
    readinessProbe:
      initialDelaySeconds: "5s"
      periodSeconds: "10s"
      timeoutSeconds: "5s"
      failureThreshold: 3
      successThreshold: 1
      http:
        host: localhost
        port: 80
        path: /health
        expectedStatus: [200]
```

## Running Periphery

### Foreground Mode

```bash
periphery --config /etc/periphery/config.yaml
```

### Systemd Service

Create `/etc/systemd/system/periphery.service`:

```ini
[Unit]
Description=Periphery BGP Anycast Service
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=periphery
Group=periphery
ExecStart=/usr/local/bin/periphery --config /etc/periphery/config.yaml
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal
SyslogIdentifier=periphery

[Install]
WantedBy=multi-user.target
```

Create user and start service:

```bash
# Create user
sudo useradd -r -s /bin/false periphery

# Create config directory
sudo mkdir -p /etc/periphery
sudo chown periphery:periphery /etc/periphery

# Copy config
sudo cp config.yaml /etc/periphery/
sudo chown periphery:periphery /etc/periphery/config.yaml

# Enable and start service
sudo systemctl daemon-reload
sudo systemctl enable periphery
sudo systemctl start periphery

# Check status
sudo systemctl status periphery
```

## Verifying BGP Session

Check BGP neighbor status:

```bash
# If you have GoBGP CLI installed
gobgp neighbor

# Check logs
sudo journalctl -u periphery -f
```

You should see log messages indicating:
1. BGP session establishment
2. Probe execution
3. Route announcement

## Next Steps

- [Configure different probe types](probes.md)
- [Enable BFD](bfd.md)
- [Advanced BGP configuration](bgp.md)
- [See example configurations](examples/)

## Troubleshooting

### BGP Session Won't Establish

Check:
1. Firewall rules allow BGP (TCP port 179)
2. Correct neighbor IP and AS number
3. BGP configuration on peer router

```bash
# Check if BGP port is reachable
telnet 10.0.0.254 179
```

### Routes Not Announced

Check:
1. Health probes are passing
2. Service is running
3. `withdrawOnDown` setting

```bash
# Check service status
systemctl status nginx.service

# View logs
sudo journalctl -u periphery -n 100
```

### Probes Failing

1. Verify probe target is accessible
2. Check probe timeout settings
3. Verify expected response

```bash
# Test HTTP probe manually
curl -v http://localhost:80/health

# Test TCP probe manually
nc -zv localhost 3306
```

For more troubleshooting tips, see [Troubleshooting Guide](reference/troubleshooting.md).
