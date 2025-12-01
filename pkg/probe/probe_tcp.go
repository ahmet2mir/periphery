package probe

import (
	"context"
	"fmt"
	"net"
	"time"

	"go.uber.org/zap"
)

// Ensure implements interface.
var _ ProbeInterface = (*ProbeTCP)(nil)

type ProbeTCP struct {
	Host    string        `yaml:"host"`
	Port    int           `yaml:"port"`
	Timeout time.Duration `yaml:"timeout"`
}

func (p *ProbeTCP) Run(ctx context.Context) (*ProbeStatus, error) {
	// Set defaults
	if p.Host == "" {
		p.Host = "localhost"
	}
	if p.Timeout == 0 {
		p.Timeout = 1 * time.Second
	}

	address := fmt.Sprintf("%s:%d", p.Host, p.Port)
	zap.S().Debug("ProbeTCP Run", "address", address)

	// Create dialer with timeout
	dialer := &net.Dialer{
		Timeout: p.Timeout,
	}

	// Attempt to connect
	conn, err := dialer.DialContext(ctx, "tcp", address)
	if err != nil {
		return nil, fmt.Errorf("ProbeTCP Run: Dial %w", err)
	}
	defer func() {
		if closeErr := conn.Close(); closeErr != nil {
			zap.S().Debug("ProbeTCP Run: error closing connection", closeErr)
		}
	}()

	zap.S().Debug("ProbeTCP Run", "address", address, "success", true)
	return &ProbeStatus{Status: "success"}, nil
}
