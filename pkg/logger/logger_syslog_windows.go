//go:build windows
// +build windows

package logger

import (
	"fmt"

	"go.uber.org/zap/zapcore"
)

// getSyslogWriter returns an error on Windows as syslog is not supported
func getSyslogWriter(cfg Config) (zapcore.WriteSyncer, func() error, error) {
	return nil, nil, fmt.Errorf("syslog driver is not supported on Windows, use 'windows' driver instead")
}
