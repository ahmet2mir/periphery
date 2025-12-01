# Prometheus Metrics

Herald exposes Prometheus metrics for monitoring prefix availability, BGP peer status, probe execution, and more.

## Configuration

Enable metrics by adding a `metrics` section to your configuration:

```yaml
metrics:
  enabled: true
  listenAddress: "0.0.0.0"
  listenPort: 9091
  interval: 15s
```

### Configuration Options

- **`enabled`**: Enable or disable metrics collection (default: false)
- **`listenAddress`**: Address to bind the metrics HTTP server (default: "127.0.0.1")
- **`listenPort`**: Port for the metrics HTTP server (default: 9091)
- **`interval`**: BGP metrics collection interval (default: 15s)

## Endpoints

- **`/metrics`**: Prometheus metrics endpoint
- **`/health`**: Health check endpoint (returns HTTP 200 OK)

## Metrics

### Prefix Metrics

#### `herald_prefix_up`
**Type:** Gauge
**Labels:** `prefix`, `name`
**Description:** Prefix announcement status (1=announced, 0=withdrawn)

Indicates whether a prefix is currently announced via BGP based on readiness probe results.

```promql
# Check if prefix is announced
herald_prefix_up{prefix="10.138.39.183/32",name="web.example.com"}

# Count announced prefixes
sum(herald_prefix_up)

# Prefixes that are down
herald_prefix_up == 0
```

### Probe Metrics

#### `herald_probe_success_total`
**Type:** Counter
**Labels:** `prefix`, `probe_type`, `name`
**Description:** Total number of successful probe executions

```promql
# Success rate by probe type
rate(herald_probe_success_total[5m])

# Success count by service
sum by (name) (herald_probe_success_total)
```

#### `herald_probe_failure_total`
**Type:** Counter
**Labels:** `prefix`, `probe_type`, `name`
**Description:** Total number of failed probe executions

```promql
# Failure rate
rate(herald_probe_failure_total[5m])

# Services with failures
sum by (name) (herald_probe_failure_total) > 0
```

#### `herald_probe_duration_seconds`
**Type:** Histogram
**Labels:** `prefix`, `probe_type`, `name`
**Description:** Duration of probe execution in seconds

```promql
# Average probe duration
rate(herald_probe_duration_seconds_sum[5m]) / rate(herald_probe_duration_seconds_count[5m])

# 95th percentile probe duration
histogram_quantile(0.95, rate(herald_probe_duration_seconds_bucket[5m]))

# Slow probes (>1s)
herald_probe_duration_seconds > 1
```

### BGP Peer Metrics

#### `herald_bgp_peer_up`
**Type:** Gauge
**Labels:** `peer_address`, `peer_asn`
**Description:** BGP peer status (1=established, 0=down)

```promql
# Check peer status
herald_bgp_peer_up{peer_address="10.134.21.74",peer_asn="64599"}

# Count established peers
sum(herald_bgp_peer_up)

# Alert on peer down
herald_bgp_peer_up == 0
```

#### `herald_bgp_peer_state`
**Type:** Gauge
**Labels:** `peer_address`, `peer_asn`
**Description:** BGP peer session state

**States:**
- `0` = Unknown
- `1` = Idle
- `2` = Connect
- `3` = Active
- `4` = OpenSent
- `5` = OpenConfirm
- `6` = Established

```promql
# Peers not in established state
herald_bgp_peer_state != 6

# Peer state changes
changes(herald_bgp_peer_state[1h])
```

#### `herald_bgp_peer_messages_sent_total`
**Type:** Counter
**Labels:** `peer_address`, `peer_asn`, `message_type`
**Description:** Total BGP messages sent to peer

**Message Types:** `update`, `notification`, `open`, `keepalive`

```promql
# Update messages sent per second
rate(herald_bgp_peer_messages_sent_total{message_type="update"}[5m])

# Notification rate (potential issues)
rate(herald_bgp_peer_messages_sent_total{message_type="notification"}[5m])
```

#### `herald_bgp_peer_messages_received_total`
**Type:** Counter
**Labels:** `peer_address`, `peer_asn`, `message_type`
**Description:** Total BGP messages received from peer

```promql
# Messages received per second
rate(herald_bgp_peer_messages_received_total[5m])

# Update message rate
rate(herald_bgp_peer_messages_received_total{message_type="update"}[5m])
```

#### `herald_bgp_route_count`
**Type:** Gauge
**Labels:** `route_table`
**Description:** Number of BGP routes in routing table

```promql
# Total routes
herald_bgp_route_count{route_table="global"}

# Route count changes
delta(herald_bgp_route_count[5m])
```

### Service Metrics

#### `herald_service_restarts_total`
**Type:** Counter
**Labels:** `name`
**Description:** Total number of service restarts triggered by failed probes

```promql
# Restart rate
rate(herald_service_restarts_total[1h])

# Services with recent restarts
increase(herald_service_restarts_total[10m]) > 0

# Total restarts by service
sum by (name) (herald_service_restarts_total)
```

## Example Prometheus Configuration

```yaml
scrape_configs:
  - job_name: 'herald'
    static_configs:
      - targets: ['localhost:9091']
    scrape_interval: 15s
    scrape_timeout: 10s
```

## Grafana Dashboard Queries

### Service Availability Panel

```promql
herald_prefix_up{name="web.example.com"}
```

### Probe Success Rate

```promql
sum(rate(herald_probe_success_total[5m])) by (name, probe_type)
/
(
  sum(rate(herald_probe_success_total[5m])) by (name, probe_type)
  +
  sum(rate(herald_probe_failure_total[5m])) by (name, probe_type)
) * 100
```

### BGP Peer Status Overview

```promql
sum(herald_bgp_peer_up) by (peer_address)
```

### Service Restart Frequency

```promql
increase(herald_service_restarts_total[1h])
```

## Alerting Rules

### Prefix Down Alert

```yaml
groups:
  - name: herald
    rules:
      - alert: PrefixDown
        expr: herald_prefix_up == 0
        for: 2m
        labels:
          severity: critical
        annotations:
          summary: "Prefix {{ $labels.prefix }} is withdrawn"
          description: "Service {{ $labels.name }} prefix {{ $labels.prefix }} has been withdrawn for more than 2 minutes"
```

### BGP Peer Down Alert

```yaml
- alert: BGPPeerDown
  expr: herald_bgp_peer_up == 0
  for: 1m
  labels:
    severity: warning
  annotations:
    summary: "BGP peer {{ $labels.peer_address }} is down"
    description: "BGP session to {{ $labels.peer_address }} (AS{{ $labels.peer_asn }}) is not established"
```

### High Probe Failure Rate

```yaml
- alert: HighProbeFailureRate
  expr: |
    (
      sum(rate(herald_probe_failure_total[5m])) by (name)
      /
      (
        sum(rate(herald_probe_success_total[5m])) by (name)
        +
        sum(rate(herald_probe_failure_total[5m])) by (name)
      )
    ) > 0.1
  for: 5m
  labels:
    severity: warning
  annotations:
    summary: "High probe failure rate for {{ $labels.name }}"
    description: "{{ $labels.name }} has >10% probe failure rate"
```

### Frequent Service Restarts

```yaml
- alert: FrequentServiceRestarts
  expr: increase(herald_service_restarts_total[10m]) > 3
  labels:
    severity: warning
  annotations:
    summary: "Frequent restarts for {{ $labels.name }}"
    description: "{{ $labels.name }} has been restarted {{ $value }} times in the last 10 minutes"
```

## Best Practices

1. **Collection Interval**: Use 15-30 second intervals for BGP metrics to balance accuracy and performance
2. **Retention**: Keep detailed metrics for at least 7 days, aggregated metrics for 30+ days
3. **Labels**: Use the `name` label for service identification in multi-service deployments
4. **Alerting**: Set up alerts for prefix withdrawal, BGP peer down, and high failure rates
5. **Dashboards**: Create per-service dashboards using the `name` label for filtering

## Integration with GoBGP

Herald automatically collects BGP metrics from GoBGP, including:
- Peer session states and connectivity
- Message counts (sent/received by type)
- Route table statistics
- Session state transitions

These metrics are collected at the configured interval and provide deep visibility into BGP operations without requiring external exporters.
