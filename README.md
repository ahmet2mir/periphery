# Periphery

> A lightweight BGP anycast service with Kubernetes-inspired health probes and BFD support

[![Go Report Card](https://goreportcard.com/badge/github.com/ahmet2mir/periphery)](https://goreportcard.com/report/github.com/ahmet2mir/periphery)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)

Periphery is a modern BGP speaker designed as a drop-in replacement for ExaBGP in simple anycast use cases. It provides intelligent route announcement based on service health checks, combining BGP routing with Kubernetes-style probe mechanisms.

## Features

- **BGP Speaker**: Full bidirectional BGP session support via GoBGP
- **Kubernetes-Inspired Probes**: Three probe types for comprehensive health checking
  - **Startup Probe**: Wait for service initialization before health checks
  - **Liveness Probe**: Detect service failures and trigger restarts
  - **Readiness Probe**: Control route announcement based on service availability
- **Multiple Probe Types**: HTTP, TCP, gRPC, and Exec probes
- **BFD Support**: Bidirectional Forwarding Detection for fast failure detection
- **Systemd Integration**: Direct service management via systemd D-Bus
- **Graceful Restart**: BGP graceful restart support for zero-downtime maintenance
- **Zero Dependencies**: Single binary with no external dependencies

## Quick Start

### Installation

Download the latest release for your platform:

```bash
# Linux AMD64
wget https://github.com/ahmet2mir/periphery/releases/latest/download/periphery_linux_amd64

# Make executable
chmod +x periphery_linux_amd64
sudo mv periphery_linux_amd64 /usr/local/bin/periphery
```

Or build from source:

```bash
git clone https://github.com/ahmet2mir/periphery.git
cd periphery
make build
```

### Basic Usage

1. Create a configuration file `config.yaml`:

```yaml
logging:
  driver: file
  format: json
  level: info
  file: periphery.log

speaker:
  asn: 64600
  routerId: "10.0.0.1"
  gracefulRestartEnabled: true
  gracefulRestartRestartTime: 120

api:
  listenAddress: "127.0.0.1"
  listenPort: 50051

neighbors:
  - address: "10.0.0.254"
    asn: 64599
    ebgpMultihopEnabled: false

prefixes:
  - ipAddress: "192.0.2.1/32"
    communities:
      - '65000:100'
    nextHop: "10.0.0.1"

    service:
      name: nginx.service
      type: systemd

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

2. Run Periphery:

```bash
periphery --config config.yaml
```

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                      Periphery                          │
├─────────────────────────────────────────────────────────┤
│                                                         │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐ │
│  │   Startup    │  │  Liveness    │  │  Readiness   │ │
│  │    Probe     │  │    Probe     │  │    Probe     │ │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘ │
│         │                 │                  │         │
│         └─────────┬───────┴──────────┬───────┘         │
│                   │                  │                 │
│            ┌──────▼──────┐    ┌──────▼──────┐         │
│            │  Service    │    │     BGP     │         │
│            │  Manager    │    │   Speaker   │         │
│            │  (systemd)  │    │   (GoBGP)   │         │
│            └─────────────┘    └──────┬──────┘         │
│                                      │                 │
│                               ┌──────▼──────┐         │
│                               │     BFD     │         │
│                               │    Agent    │         │
│                               └─────────────┘         │
└─────────────────────────────────────────────────────────┘
                                │
                         ┌──────▼──────┐
                         │   Network   │
                         │  (BGP Peer) │
                         └─────────────┘
```

### How It Works

1. **Service Health Monitoring**:
   - **Startup Probe**: Ensures service is fully initialized before other checks begin
   - **Liveness Probe**: Monitors service health; triggers restart on consecutive failures
   - **Readiness Probe**: Controls BGP route announcement based on service availability

2. **BGP Route Management**:
   - Routes are announced only when readiness probes succeed
   - Routes are withdrawn when probes fail (based on `failureThreshold`)
   - Supports graceful restart for maintenance windows

3. **BFD Integration**:
   - Fast failure detection (sub-second)
   - Complements BGP keepalives
   - Configurable intervals and multipliers

## Probe Types

### HTTP Probe

```yaml
readinessProbe:
  periodSeconds: "10s"
  timeoutSeconds: "5s"
  http:
    host: localhost
    port: 8080
    path: /healthz
    scheme: https
    expectedStatus: [200, 204]
    httpHeaders:
      - name: Authorization
        value: "Bearer token123"
```

### TCP Probe

```yaml
readinessProbe:
  periodSeconds: "10s"
  tcp:
    host: localhost
    port: 3306
    timeout: "3s"
```

### gRPC Probe

```yaml
readinessProbe:
  periodSeconds: "10s"
  grpc:
    host: localhost
    port: 9090
    service: myapp.health.v1.Health
    timeout: "5s"
```

### Exec Probe

```yaml
startupProbe:
  periodSeconds: "5s"
  exec:
    command: /usr/local/bin/check-ready
    args:
      - "--strict"
    exitCodes: [0]
```

## Configuration Reference

See [docs/configuration.md](docs/configuration.md) for complete configuration reference.

## Use Cases

### Load Balancer Health Checking

Announce anycast VIP only when backend services are healthy:

```yaml
prefixes:
  - ipAddress: "192.0.2.100/32"
    readinessProbe:
      http:
        port: 80
        path: /health
```

### Database High Availability

Announce database VIP based on replication lag:

```yaml
prefixes:
  - ipAddress: "192.0.2.200/32"
    readinessProbe:
      exec:
        command: /usr/local/bin/check-replication-lag
        exitCodes: [0]
```

### API Gateway Failover

Multiple instances announcing same VIP with health-based routing:

```yaml
prefixes:
  - ipAddress: "192.0.2.50/32"
    readinessProbe:
      http:
        port: 8080
        path: /api/health
    livenessProbe:
      tcp:
        port: 8080
```

## BFD Configuration

Enable fast failure detection with BFD:

```yaml
bfd:
  enabled: true
  listenAddress: "0.0.0.0"
  listenPort: 3784
  minimumReceptionInterval: 300ms
  minimumTransmissionInterval: 300ms
  detectionMultiplier: 3
  passive: false
```

## Logging Configuration

Periphery supports flexible logging with multiple drivers and formats:

```yaml
logging:
  driver: file        # Options: file, syslog, journald, windows, none
  format: json        # Options: json, text
  level: info         # Options: debug, info, warn, error
  file: periphery.log # Required for file driver
```

### Logging Drivers

- **file**: Write logs to a file
- **syslog**: Send logs to syslog (Unix/Linux)
- **journald**: Send logs to systemd journal (Linux)
- **windows**: Send logs to Windows Event Log (Windows 2025)
- **none**: Disable logging

See [docs/logging.md](docs/logging.md) for detailed logging configuration and examples.

## Development

### Prerequisites

- Go 1.21 or later
- golangci-lint
- gosec
- goreleaser

### Building

```bash
# Install dependencies
make setup

# Run tests
make test

# Run linting
make lint

# Run security checks
make security

# Build binaries
make build
```

### Project Structure

```
periphery/
├── main.go              # Application entry point
├── pkg/
│   ├── bfd/            # BFD agent implementation
│   ├── config/         # Configuration structures
│   ├── logger/         # Logging infrastructure
│   ├── probe/          # Health probe implementations
│   │   ├── probe.go          # Probe interface and manager
│   │   ├── probe_http.go     # HTTP probe
│   │   ├── probe_tcp.go      # TCP probe
│   │   ├── probe_grpc.go     # gRPC probe
│   │   └── probe_exec.go     # Exec probe
│   ├── scheduler/      # Probe scheduling logic
│   ├── service/        # Service management (systemd)
│   └── speaker/        # BGP speaker (GoBGP wrapper)
└── docs/               # Documentation
    ├── examples/       # Configuration examples
    └── logging.md      # Logging documentation
```

## Contributing

Contributions are welcome! Please read [CONTRIBUTING.md](CONTRIBUTING.md) for details.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- **[GoBGP](https://github.com/osrg/gobgp)**: BGP implementation
- **[ExaBGP](https://github.com/Exa-Networks/exabgp)**: Inspiration for concepts and design
- **[gobfd](https://github.com/rhgb/gobfd)**: BFD implementation
- **[bgp-speaker](https://github.com/sir-sukhov/bgp-speaker)**: Code inspiration
- **Kubernetes**: Probe design and conventions

## Support

- Documentation: [https://ahmet2mir.github.io/periphery](https://ahmet2mir.github.io/periphery)
- Issues: [GitHub Issues](https://github.com/ahmet2mir/periphery/issues)
- Discussions: [GitHub Discussions](https://github.com/ahmet2mir/periphery/discussions)
