//go:build !windows
// +build !windows

package logger

import (
	"fmt"

	"go.uber.org/zap/zapcore"
)

// getWindowsWriter returns an error on non-Windows systems
func getWindowsWriter(cfg Config) (zapcore.WriteSyncer, func() error, error) {
	return nil, nil, fmt.Errorf("windows driver is only supported on Windows systems")
}
