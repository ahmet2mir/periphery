//go:build !windows
// +build !windows

package logger

import (
	"fmt"
	"log/syslog"

	"go.uber.org/zap/zapcore"
)

// getSyslogWriter creates a syslog writer for Unix systems
func getSyslogWriter(cfg Config) (zapcore.WriteSyncer, func() error, error) {
	writer, err := syslog.New(syslog.LOG_INFO|syslog.LOG_DAEMON, "herald")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to syslog: %w", err)
	}
	return zapcore.AddSync(writer), func() error { return writer.Close() }, nil
}
