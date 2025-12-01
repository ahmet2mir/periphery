# Getting Started with Herald

This guide will help you get Herald up and running quickly.

## Prerequisites

- Linux system with systemd
- BGP router/peer to establish sessions with
- Go 1.21+ (if building from source)

## Installation

### Option 1: Download Binary

Download the latest release for your platform:

```bash
# Download for Linux AMD64
wget https://github.com/ahmet2mir/herald/releases/latest/download/herald_linux_amd64

# Make it executable
chmod +x herald_linux_amd64

# Move to system path
sudo mv herald_linux_amd64 /usr/local/bin/herald

# Verify installation
herald --version
```

### Option 2: Build from Source

```bash
# Clone repository
git clone https://github.com/ahmet2mir/herald.git
cd herald

# Install build dependencies
make setup

# Build
make build

# Binary will be in dist/herald_linux_amd64_v1/herald
sudo cp dist/herald_linux_amd64_v1/herald /usr/local/bin/
```

## Basic Configuration

Create a configuration file at `/etc/herald/config.yaml`:

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

## Running Herald

### Foreground Mode

```bash
herald --config /etc/herald/config.yaml
```

### Systemd Service

Create `/etc/systemd/system/herald.service`:

```ini
[Unit]
Description=Herald BGP Anycast Service
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=herald
Group=herald
ExecStart=/usr/local/bin/herald --config /etc/herald/config.yaml
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal
SyslogIdentifier=herald

[Install]
WantedBy=multi-user.target
```

Create user and start service:

```bash
# Create user
sudo useradd -r -s /bin/false herald

# Create config directory
sudo mkdir -p /etc/herald
sudo chown herald:herald /etc/herald

# Copy config
sudo cp config.yaml /etc/herald/
sudo chown herald:herald /etc/herald/config.yaml

# Enable and start service
sudo systemctl daemon-reload
sudo systemctl enable herald
sudo systemctl start herald

# Check status
sudo systemctl status herald
```

## Verifying BGP Session

Check BGP neighbor status:

```bash
# If you have GoBGP CLI installed
gobgp neighbor

# Check logs
sudo journalctl -u herald -f
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
sudo journalctl -u herald -n 100
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
