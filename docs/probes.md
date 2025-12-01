# Health Probes

Periphery uses Kubernetes-inspired health probes to monitor service health and control BGP route announcements.

## Probe Types

There are three types of health probes:

### Startup Probe

**Purpose**: Verify that a service has successfully started before other probes begin.

**Behavior**:
- Runs once after `initialDelaySeconds`
- If it fails, other probes don't start
- Useful for slow-starting services

**Example**:
```yaml
startupProbe:
  initialDelaySeconds: "10s"
  periodSeconds: "5s"
  timeoutSeconds: "5s"
  failureThreshold: 10      # Allow up to 50 seconds (10 * 5s)
  successThreshold: 1
  exec:
    command: /usr/local/bin/service-ready
    exitCodes: [0]
```

### Liveness Probe

**Purpose**: Detect when a service has entered a broken state and needs to be restarted.

**Behavior**:
- Runs periodically after startup probe succeeds
- On failure (after `failureThreshold` consecutive failures):
  - Service is restarted via systemd
- Use for detecting deadlocks, infinite loops, or corruption

**Example**:
```yaml
livenessProbe:
  initialDelaySeconds: "30s"
  periodSeconds: "10s"
  timeoutSeconds: "5s"
  failureThreshold: 3       # Restart after 30 seconds of failures
  successThreshold: 1
  http:
    host: localhost
    port: 8080
    path: /healthz
```

### Readiness Probe

**Purpose**: Control when traffic should be routed to the service (via BGP route announcement).

**Behavior**:
- Runs periodically after startup probe succeeds
- On success (after `successThreshold` consecutive successes):
  - BGP route is announced
- On failure (after `failureThreshold` consecutive failures):
  - BGP route is withdrawn
- Use for services that temporarily can't handle traffic

**Example**:
```yaml
readinessProbe:
  initialDelaySeconds: "15s"
  periodSeconds: "10s"
  timeoutSeconds: "5s"
  failureThreshold: 2       # Withdraw after 20 seconds of failures
  successThreshold: 1
  http:
    host: localhost
    port: 8080
    path: /ready
```

## Probe Execution Timeline

```
Service Starts
    │
    ├─ initialDelaySeconds (startup)
    │
    ├─ Startup Probe (every periodSeconds)
    │   │
    │   ├─ Success ─┐
    │   │           │
    │   └─ Failure ─┘ (retry until successThreshold or give up after failureThreshold)
    │
    ├─ Startup Probe Succeeded
    │
    ├─ initialDelaySeconds (liveness)
    │
    ├─ Liveness Probe (every periodSeconds)
    │   │
    │   ├─ Success ─┐
    │   │           │
    │   └─ Failure ─┘ (count failures, restart service after failureThreshold)
    │
    ├─ initialDelaySeconds (readiness)
    │
    └─ Readiness Probe (every periodSeconds)
        │
        ├─ Success ─┐ Announce BGP Route
        │           │
        └─ Failure ─┘ Withdraw BGP Route (after failureThreshold)
```

## Common Probe Fields

All probes share these configuration fields:

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `initialDelaySeconds` | duration | 0s | Delay before first probe |
| `periodSeconds` | duration | 10s | Interval between probes |
| `timeoutSeconds` | duration | 1s | Probe timeout |
| `failureThreshold` | int32 | 3 | Consecutive failures before action |
| `successThreshold` | int32 | 1 | Consecutive successes before healthy |
| `terminationGracePeriodSeconds` | duration | 0s | Grace period for shutdown |

## Probe Mechanisms

### HTTP Probe

Performs an HTTP GET request.

```yaml
http:
  host: localhost           # Target host
  port: 8080               # Target port
  path: /healthz          # HTTP path
  scheme: http            # http or https
  expectedStatus: [200]   # Expected status codes
  httpHeaders:            # Optional custom headers
    - name: Authorization
      value: "Bearer token"
  requestTimeout: "3s"    # Request timeout
```

**Success**: Response status code matches `expectedStatus`

**Failure**: Request fails, times out, or status code doesn't match

### TCP Probe

Attempts to establish a TCP connection.

```yaml
tcp:
  host: localhost    # Target host
  port: 3306        # Target port
  timeout: "3s"     # Connection timeout
```

**Success**: TCP connection established

**Failure**: Connection refused, timeout, or network error

### gRPC Probe

Uses gRPC health checking protocol.

```yaml
grpc:
  host: localhost                    # Target host
  port: 9090                        # Target port
  service: myapp.health.v1.Health  # gRPC service (optional)
  timeout: "5s"                    # Request timeout
```

**Success**: Health check returns `SERVING` status

**Failure**: Request fails, times out, or status is not `SERVING`

**Note**: Service must implement [gRPC Health Checking Protocol](https://github.com/grpc/grpc/blob/master/doc/health-checking.md)

### Exec Probe

Executes a command in the container.

```yaml
exec:
  command: /usr/local/bin/check    # Command path
  args:                            # Command arguments
    - "--strict"
  exitCodes: [0]                  # Expected exit codes
```

**Success**: Command exits with code in `exitCodes`

**Failure**: Command exits with unexpected code, times out, or fails to execute

## Best Practices

### Startup Probes

1. **Use for slow-starting services**: If initialization takes >30s
2. **Set generous timeouts**: `failureThreshold * periodSeconds` should cover startup time
3. **Keep it simple**: Just verify service is ready, not full health

```yaml
startupProbe:
  periodSeconds: "5s"
  failureThreshold: 30     # Allow 150 seconds for startup
  exec:
    command: /usr/local/bin/is-ready
```

### Liveness Probes

1. **Don't restart for transient failures**: Set `failureThreshold` >= 3
2. **Test critical functionality**: Verify service can do its job
3. **Keep checks lightweight**: Avoid expensive operations
4. **Don't check dependencies**: Only check the service itself

```yaml
livenessProbe:
  periodSeconds: "10s"
  failureThreshold: 3       # 30 seconds of failures before restart
  timeoutSeconds: "5s"
  http:
    path: /healthz          # Lightweight health check
```

### Readiness Probes

1. **Check all dependencies**: Database, cache, external APIs
2. **Use conservative thresholds**: Avoid route flapping
3. **Consider impact**: Withdrawing route affects all traffic

```yaml
readinessProbe:
  periodSeconds: "10s"
  failureThreshold: 2       # Quick withdrawal on failure
  successThreshold: 2       # Confirm stability before announcing
  http:
    path: /ready            # Thorough readiness check
```

## Example: Complete Probe Configuration

```yaml
prefixes:
  - ipAddress: "192.0.2.1/32"

    service:
      name: myapp.service
      type: systemd

    # 1. Verify service has started
    startupProbe:
      initialDelaySeconds: "10s"
      periodSeconds: "5s"
      timeoutSeconds: "5s"
      failureThreshold: 20
      successThreshold: 1
      exec:
        command: /usr/local/bin/startup-check
        exitCodes: [0]

    # 2. Restart if service becomes unhealthy
    livenessProbe:
      initialDelaySeconds: "30s"
      periodSeconds: "10s"
      timeoutSeconds: "5s"
      failureThreshold: 3
      successThreshold: 1
      http:
        host: localhost
        port: 8080
        path: /healthz
        expectedStatus: [200]

    # 3. Control BGP route announcement
    readinessProbe:
      initialDelaySeconds: "30s"
      periodSeconds: "10s"
      timeoutSeconds: "5s"
      failureThreshold: 2
      successThreshold: 2
      http:
        host: localhost
        port: 8080
        path: /ready
        expectedStatus: [200, 204]
```

## Troubleshooting

### Probes Always Failing

1. Check probe configuration matches service
2. Verify service is actually healthy
3. Check firewall/network connectivity
4. Increase `timeoutSeconds`
5. Check service logs

### Route Flapping

1. Increase `failureThreshold`
2. Increase `successThreshold`
3. Increase `periodSeconds`
4. Check for intermittent failures
5. Use more stable health check

### Service Not Restarting

1. Verify liveness probe is configured
2. Check systemd service permissions
3. Review liveness probe logs
4. Confirm `failureThreshold` is being reached
