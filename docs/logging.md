# Logging Configuration

The herald application supports flexible logging configuration through various drivers and formats. All application logs, including GoBGP and BFD logs, are routed through the configured logger.

## Configuration Options

### Driver

The logging driver determines where logs are sent:

- **`file`** - Write logs to a file (default)
- **`syslog`** - Send logs to syslog (Unix/Linux only)
- **`journald`** - Send logs to systemd journal (Linux only)
- **`windows`** - Send logs to Windows Event Log (Windows only)
- **`none`** - Disable logging

### Format

The log format controls how log messages are structured:

- **`json`** - Structured JSON format (default, recommended for production)
- **`text`** - Human-readable text format (better for development)

### Level

The log level controls the verbosity of logging:

- **`debug`** - Most verbose, includes all log levels
- **`info`** - Normal operational messages (default)
- **`warn`** - Warning messages only
- **`error`** - Error messages only

### File

When using the `file` driver, this specifies the path to the log file.

## Configuration Examples

### File Logging with JSON (Production)

```yaml
logging:
  driver: file
  format: json
  level: info
  file: /var/log/herald/herald.log
```

### File Logging with Text (Development)

```yaml
logging:
  driver: file
  format: text
  level: debug
  file: herald.log
```

### Syslog (Unix/Linux)

```yaml
logging:
  driver: syslog
  format: json
  level: info
```

### Systemd Journal (Linux)

```yaml
logging:
  driver: journald
  format: json
  level: info
```

### Windows Event Log (Windows 2025)

```yaml
logging:
  driver: windows
  format: json
  level: info
```

### Disable Logging

```yaml
logging:
  driver: none
```

## Platform-Specific Notes

### Linux

On Linux systems, you can use any of these drivers:
- `file` - Standard file-based logging
- `syslog` - Traditional syslog daemon
- `journald` - Modern systemd journal (recommended for systemd-based systems)
- `none` - No logging

### Unix (macOS, BSD)

On Unix systems, you can use:
- `file` - Standard file-based logging
- `syslog` - Traditional syslog daemon
- `none` - No logging

### Windows 2025

On Windows systems, you can use:
- `file` - Standard file-based logging
- `windows` - Windows Event Log (recommended for Windows)
- `none` - No logging

**Note:** The Windows Event Log driver automatically registers the "herald" event source if it doesn't exist.

## Log Levels in Detail

### Debug Level
Includes detailed execution traces from:
- HTTP probe requests and responses
- TCP connection attempts
- gRPC health checks
- Command execution details

### Info Level
Includes operational status:
- Application startup/shutdown
- BGP neighbor configuration
- Path announcements/withdrawals
- Probe scheduling events

### Warn Level
Includes non-critical issues:
- Service status warnings
- Probe failures (within threshold)
- BGP operation warnings

### Error Level
Includes errors that need attention:
- Probe execution errors
- Service restart failures
- Path management errors

## Default Configuration

If the logging section is omitted from the configuration file, the following defaults are used:

```yaml
logging:
  driver: file
  format: json
  level: info
  file: herald.log
```

## Unified Logging

All components of herald use the configured logger:
- **Application logs**: Main application events, errors, and status
- **GoBGP logs**: BGP session establishment, route updates, neighbor events
- **BFD logs**: BFD session state changes and events
- **Probe logs**: Health check execution and results

This ensures all logs are consistently formatted and routed to the same destination, making monitoring and debugging easier.

## Viewing Logs

### File Logs

```bash
# Follow JSON logs
tail -f herald.log | jq

# Follow text logs
tail -f herald.log

# Search for errors
grep "error" herald.log
```

### Syslog

```bash
# View syslog
tail -f /var/log/syslog | grep herald

# Or on systems with rsyslog
tail -f /var/log/messages | grep herald
```

### Systemd Journal

```bash
# View herald logs
journalctl -u herald -f

# View with specific log level
journalctl -u herald -p err

# View recent logs
journalctl -u herald -n 100
```

### Windows Event Log

1. Open Event Viewer (eventvwr.msc)
2. Navigate to Windows Logs > Application
3. Filter by Source: "herald"

Or use PowerShell:

```powershell
# View recent herald events
Get-EventLog -LogName Application -Source herald -Newest 50

# Follow events
Get-EventLog -LogName Application -Source herald -After (Get-Date).AddMinutes(-5) | Format-List
```

## Best Practices

1. **Production**: Use `json` format with `info` level for structured logging
2. **Development**: Use `text` format with `debug` level for readability
3. **Systemd Services**: Use `journald` driver on Linux systems with systemd
4. **Windows Services**: Use `windows` driver for Windows Event Log integration
5. **Containers**: Use `file` driver with stdout redirection or log to `/dev/stdout`
6. **Log Rotation**: Configure logrotate for file-based logging

### Example Logrotate Configuration

```
/var/log/herald/herald.log {
    daily
    rotate 7
    compress
    delaycompress
    missingok
    notifempty
    create 0644 herald herald
    postrotate
        systemctl reload herald || true
    endscript
}
```
