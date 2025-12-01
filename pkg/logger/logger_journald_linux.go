//go:build linux
// +build linux

package logger

import (
	"fmt"

	"github.com/coreos/go-systemd/v22/journal"
	"go.uber.org/zap/zapcore"
)

// journaldWriter wraps systemd journal writer
type journaldWriter struct{}

func (jw *journaldWriter) Write(p []byte) (n int, err error) {
	// Send to journald with INFO priority
	err = journal.Print(journal.PriInfo, "%s", string(p))
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

func (jw *journaldWriter) Sync() error {
	return nil
}

// getJournaldWriter creates a journald writer for Linux systems
func getJournaldWriter(cfg Config) (zapcore.WriteSyncer, func() error, error) {
	if !journal.Enabled() {
		return nil, nil, fmt.Errorf("systemd journal is not available on this system")
	}
	return &journaldWriter{}, nil, nil
}
