# Configuration Reference

This document provides a complete reference for Periphery configuration.

## Configuration File Format

Periphery uses YAML for configuration. By default, it looks for `config.yaml` in the current directory.

```bash
periphery --config /etc/periphery/config.yaml
```

## Top-Level Structure

```yaml
speaker:      # BGP speaker configuration
bfd:          # BFD configuration (optional)
api:          # gRPC API configuration
neighbors:    # BGP neighbors
prefixes:     # Routes to announce with health checks
```

## Speaker Configuration

BGP speaker global settings.

```yaml
speaker:
  asn: 64600                              # Required: Local AS number
  routerId: "10.0.0.1"                   # Required: BGP router ID
  gracefulRestartEnabled: true            # Optional: Enable graceful restart
  gracefulRestartRestartTime: 120         # Optional: GR restart time (seconds)
```

### Fields

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `asn` | uint32 | Yes | - | Local Autonomous System number |
| `routerId` | string | Yes | - | BGP router ID (IPv4 format) |
| `gracefulRestartEnabled` | bool | No | false | Enable BGP graceful restart |
| `gracefulRestartRestartTime` | uint32 | No | 0 | Graceful restart time in seconds |

## BFD Configuration

Bidirectional Forwarding Detection settings.

```yaml
bfd:
  enabled: true                                    # Enable BFD
  listenAddress: "0.0.0.0"                        # Listen address
  listenPort: 3784                                # BFD port (standard: 3784)
  minimumReceptionInterval: 300ms                  # Min RX interval
  minimumTransmissionInterval: 300ms               # Min TX interval
  detectionMultiplier: 3                          # Detection multiplier
  passive: false                                   # Passive mode
```

### Fields

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `enabled` | bool | No | false | Enable BFD support |
| `listenAddress` | string | No | "0.0.0.0" | Address to listen on |
| `listenPort` | int | No | 3784 | UDP port for BFD |
| `minimumReceptionInterval` | duration | No | 1s | Min interval between received BFD packets |
| `minimumTransmissionInterval` | duration | No | 1s | Min interval between transmitted BFD packets |
| `detectionMultiplier` | uint8 | No | 3 | Number of missed packets before declaring peer down |
| `passive` | bool | No | false | Passive BFD mode |

**Detection Time**: `minimumReceptionInterval * detectionMultiplier`

Example: `300ms * 3 = 900ms` to detect failure

## API Configuration

gRPC API server settings.

```yaml
api:
  listenAddress: "127.0.0.1"    # API listen address
  listenPort: 50051              # API listen port
```

### Fields

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `listenAddress` | string | Yes | - | Address for API server |
| `listenPort` | int | Yes | - | Port for API server |

## Neighbors Configuration

BGP neighbor definitions.

```yaml
neighbors:
  - address: "10.0.0.254"           # Neighbor IP
    asn: 64599                       # Neighbor AS
    ebgpMultihopEnabled: false       # Enable eBGP multihop
```

### Fields

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `address` | string | Yes | - | BGP neighbor IP address |
| `asn` | uint32 | Yes | - | Neighbor AS number |
| `ebgpMultihopEnabled` | bool | No | false | Enable eBGP multihop |

## Prefixes Configuration

Routes to announce with health check configuration.

```yaml
prefixes:
  - ipAddress: "192.0.2.1/32"              # Prefix to announce
    communities:                            # BGP communities
      - '65000:100'
      - '65000:200'
    nextHop: "10.0.0.1"                    # Next hop
    asn: 64600                              # Override AS
    multiExitDescriminator: 100             # MED value
    asPathPrepend: []                       # AS path prepend
    withdrawOnDown: true                    # Withdraw on failure
    maintenance: /etc/maintenance/enabled   # Maintenance file

    service:                                # Service to monitor
      name: nginx.service
      type: systemd

    startupProbe:                           # Startup health check
      # ... probe configuration

    livenessProbe:                          # Liveness check
      # ... probe configuration

    readinessProbe:                         # Readiness check
      # ... probe configuration
```

### Prefix Fields

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `ipAddress` | string | Yes | - | IP prefix in CIDR notation |
| `communities` | []string | No | [] | BGP communities (format: `ASN:value`) |
| `nextHop` | string | Yes | - | Next hop IP address |
| `asn` | uint32 | No | speaker.asn | Override AS number |
| `multiExitDescriminator` | uint32 | No | 0 | BGP MED attribute |
| `asPathPrepend` | []uint32 | No | [] | AS path prepend list |
| `withdrawOnDown` | bool | No | true | Withdraw route when unhealthy |
| `maintenance` | string | No | "" | Path to maintenance flag file |

### Service Configuration

```yaml
service:
  name: nginx.service    # Service name
  type: systemd          # Service type (currently only systemd)
```

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `name` | string | Yes | - | Systemd service unit name |
| `type` | string | Yes | - | Service manager type (systemd) |

## Probe Configuration

All probe types share common timing and threshold settings.

### Common Probe Fields

```yaml
probe:
  initialDelaySeconds: "5s"       # Delay before first check
  terminationGracePeriodSeconds: "30s"  # Graceful shutdown time
  periodSeconds: "10s"             # Check interval
  timeoutSeconds: "5s"            # Check timeout
  failureThreshold: 3              # Consecutive failures
  successThreshold: 1              # Consecutive successes
```

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `initialDelaySeconds` | duration | No | 0s | Wait before starting checks |
| `terminationGracePeriodSeconds` | duration | No | 0s | Grace period for shutdown |
| `periodSeconds` | duration | No | 10s | How often to perform probe |
| `timeoutSeconds` | duration | No | 1s | Probe timeout |
| `failureThreshold` | int32 | No | 3 | Failures before unhealthy |
| `successThreshold` | int32 | No | 1 | Successes before healthy |

### HTTP Probe

```yaml
http:
  host: localhost               # Target host
  port: 8080                    # Target port
  path: /healthz               # HTTP path
  scheme: http                  # http or https
  expectedStatus: [200, 204]   # Expected status codes
  httpHeaders:                  # Custom headers
    - name: Authorization
      value: "Bearer token"
  requestTimeout: "3s"          # Request timeout
```

### TCP Probe

```yaml
tcp:
  host: localhost    # Target host
  port: 3306         # Target port
  timeout: "3s"      # Connection timeout
```

### gRPC Probe

```yaml
grpc:
  host: localhost                    # Target host
  port: 9090                         # Target port
  service: myapp.health.v1.Health   # gRPC service name
  timeout: "5s"                      # Request timeout
```

### Exec Probe

```yaml
exec:
  command: /usr/local/bin/check    # Command to execute
  args:                             # Command arguments
    - "--strict"
    - "--timeout=5"
  exitCodes: [0]                    # Expected exit codes
```

## Duration Format

Durations are specified as strings with units:

- `ns` - nanoseconds
- `us` or `µs` - microseconds
- `ms` - milliseconds
- `s` - seconds
- `m` - minutes
- `h` - hours

Examples:
- `"300ms"` - 300 milliseconds
- `"1.5s"` - 1.5 seconds
- `"2m30s"` - 2 minutes 30 seconds
- `"1h"` - 1 hour

## Complete Example

See [examples/complete.yaml](examples/complete.yaml) for a complete configuration example.
