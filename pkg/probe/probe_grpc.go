package probe

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health/grpc_health_v1"
)

// Ensure implements interface.
var _ ProbeInterface = (*ProbeGRPC)(nil)

type ProbeGRPC struct {
	Host    string        `yaml:"host"`
	Port    int           `yaml:"port"`
	Service string        `yaml:"service"`
	Timeout time.Duration `yaml:"timeout"`
}

func (p *ProbeGRPC) Run(ctx context.Context) (*ProbeStatus, error) {
	// Set defaults
	if p.Host == "" {
		p.Host = "localhost"
	}
	if p.Timeout == 0 {
		p.Timeout = 1 * time.Second
	}

	address := fmt.Sprintf("%s:%d", p.Host, p.Port)
	zap.S().Debug("ProbeGRPC Run", "address", address, "service", p.Service)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, p.Timeout)
	defer cancel()

	// Create gRPC connection
	conn, err := grpc.DialContext(ctx, address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("ProbeGRPC Run: Dial %w", err)
	}
	defer func() {
		if closeErr := conn.Close(); closeErr != nil {
			zap.S().Debug("ProbeGRPC Run: error closing connection", closeErr)
		}
	}()

	// Create health check client
	healthClient := grpc_health_v1.NewHealthClient(conn)

	// Perform health check
	resp, err := healthClient.Check(ctx, &grpc_health_v1.HealthCheckRequest{
		Service: p.Service,
	})
	if err != nil {
		return nil, fmt.Errorf("ProbeGRPC Run: Check %w", err)
	}

	// Check if service is serving
	if resp.Status != grpc_health_v1.HealthCheckResponse_SERVING {
		return nil, fmt.Errorf("ProbeGRPC Run: Service not serving, status: %v", resp.Status)
	}

	zap.S().Debug("ProbeGRPC Run", "address", address, "service", p.Service, "success", true)
	return &ProbeStatus{Status: "success"}, nil
}
