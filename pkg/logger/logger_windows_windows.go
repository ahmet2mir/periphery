//go:build windows
// +build windows

package logger

import (
	"fmt"

	"go.uber.org/zap/zapcore"
	"golang.org/x/sys/windows/svc/eventlog"
)

// windowsEventLogWriter wraps Windows Event Log
type windowsEventLogWriter struct {
	log *eventlog.Log
}

func (w *windowsEventLogWriter) Write(p []byte) (n int, err error) {
	// Write to Windows Event Log as Info
	err = w.log.Info(1, string(p))
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

func (w *windowsEventLogWriter) Sync() error {
	return nil
}

// getWindowsWriter creates a Windows Event Log writer
func getWindowsWriter(cfg Config) (zapcore.WriteSyncer, func() error, error) {
	// Open or create event log source
	log, err := eventlog.Open("herald")
	if err != nil {
		// Try to install the event source
		err = eventlog.InstallAsEventCreate("herald", eventlog.Info|eventlog.Warning|eventlog.Error)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to install Windows Event Log source: %w", err)
		}
		log, err = eventlog.Open("herald")
		if err != nil {
			return nil, nil, fmt.Errorf("failed to open Windows Event Log: %w", err)
		}
	}

	writer := &windowsEventLogWriter{log: log}
	return writer, func() error { return log.Close() }, nil
}
