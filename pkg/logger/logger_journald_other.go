//go:build !linux
// +build !linux

package logger

import (
	"fmt"

	"go.uber.org/zap/zapcore"
)

// getJournaldWriter returns an error on non-Linux systems
func getJournaldWriter(cfg Config) (zapcore.WriteSyncer, func() error, error) {
	return nil, nil, fmt.Errorf("journald driver is only supported on Linux systems")
}
