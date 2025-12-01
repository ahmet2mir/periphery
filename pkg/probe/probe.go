package probe

// highly inspired by
// https://kubernetes.io/docs/reference/kubernetes-api/workload-resources/pod-v1/#Probe

import (
	"context"
	"fmt"
	"time"

	"github.com/ahmet2mir/herald/pkg/service"
)

type ProbeInterface interface {
	Run(ctx context.Context) (*ProbeStatus, error)
}

type ProbeStatus struct {
	Status string `yaml:"status"`
}

type Probe struct {
	// Number of seconds after the service has started before liveness probes are initiated
	//nolint:staticcheck // Field name matches Kubernetes API convention
	InitialDelaySeconds time.Duration `yaml:"initialDelaySeconds"`

	// Optional duration in seconds the service needs to terminate gracefully upon probe failure.
	// Used by the scheduler to wait before forcefully terminating/restarting a failed service.
	//nolint:staticcheck // Field name matches Kubernetes API convention
	TerminationGracePeriodSeconds time.Duration `yaml:"terminationGracePeriodSeconds"`

	// How often (in seconds) to perform the probe.
	// Default to 10 seconds. Minimum value is 1.
	//nolint:staticcheck // Field name matches Kubernetes API convention
	PeriodSeconds time.Duration `yaml:"periodSeconds"`

	// Number of seconds after which the probe times out.
	// Defaults to 1 second. Minimum value is 1.
	// Applied as a context timeout for the entire probe operation.
	//nolint:staticcheck // Field name matches Kubernetes API convention
	TimeoutSeconds time.Duration `yaml:"timeoutSeconds"`

	// Minimum consecutive failures for the probe to be considered failed after having succeeded.
	// Defaults to 3. Minimum value is 1.
	FailureThreshold int32 `yaml:"failureThreshold"`

	// Minimum consecutive successes for the probe to be considered successful after having failed.
	// Defaults to 1. Must be 1 for liveness and startup. Minimum value is 1.
	SuccessThreshold int32 `yaml:"successThreshold"`

	ProbeHTTP *ProbeHTTP `yaml:"http"`
	ProbeGRPC *ProbeGRPC `yaml:"grpc"`
	ProbeExec *ProbeExec `yaml:"exec"`
	ProbeTCP  *ProbeTCP  `yaml:"tcp"`
}

type ProbeManager struct {
	NumberFailure int32
	NumberSuccess int32
}

func NewProbeManager() *ProbeManager {
	return &ProbeManager{0, 0}
}
func (pm *ProbeManager) Run(ctx context.Context, p *Probe, s *service.Service) (*ProbeStatus, error) {
	var ps *ProbeStatus
	var err error

	// Apply global timeout if specified
	if p.TimeoutSeconds > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, p.TimeoutSeconds)
		defer cancel()
	}

	if p.ProbeHTTP != nil {
		ps, err = p.ProbeHTTP.Run(ctx)
	} else if p.ProbeGRPC != nil {
		ps, err = p.ProbeGRPC.Run(ctx)
	} else if p.ProbeExec != nil {
		ps, err = p.ProbeExec.Run(ctx)
	} else if p.ProbeTCP != nil {
		ps, err = p.ProbeTCP.Run(ctx)
	} else {
		return nil, fmt.Errorf("no probe configured")
	}

	if err != nil && pm.NumberFailure < p.FailureThreshold {
		pm.NumberFailure++
	} else if pm.NumberSuccess < p.SuccessThreshold {
		pm.NumberSuccess++
	}
	return ps, err
}
