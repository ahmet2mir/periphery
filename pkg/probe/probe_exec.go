package probe

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"slices"
	"sync"
	"time"

	"go.uber.org/zap"
)

// Ensure implements interface.
var _ ProbeInterface = (*ProbeExec)(nil)

type ProbeExec struct {
	Command   string        `yaml:"command"`
	Args      []string      `yaml:"args"`
	User      string        `yaml:"user"`
	Timeout   time.Duration `yaml:"timeout"`
	ExitCodes []int         `yaml:"exitCodes"`
}

func (p *ProbeExec) Run(ctx context.Context) (*ProbeStatus, error) {
	if p.ExitCodes == nil {
		p.ExitCodes = []int{0}
	}
	if p.Args == nil {
		p.Args = []string{}
	}
	zap.S().Debug("ProbeExec Run", "command", p.Command, "args", p.Args, "exitCodes", p.ExitCodes)

	// #nosec G204 -- Command execution is intentional for health check probes
	cmd := exec.CommandContext(ctx, p.Command, p.Args...)

	// Get the stdout and stderr pipes
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("ProbeExec Run: StdoutPipe %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("ProbeExec Run: StderrPipe %w", err)
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("ProbeExec Run: Start %w", err)
	}

	// Use a WaitGroup to wait for both goroutines to finish
	var wg sync.WaitGroup
	wg.Add(2)

	// Goroutine to stream stdout
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			zap.S().Debug("ProbeExec Run", scanner.Text())
		}
	}()

	// Goroutine to stream stderr
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			zap.S().Debug("ProbeExec Run", scanner.Text())
		}
	}()

	// Wait for both goroutines to finish
	wg.Wait()

	// Wait for the command to exit and get the exit code
	if err := cmd.Wait(); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode := exitError.ExitCode()
			if !slices.Contains(p.ExitCodes, exitCode) {
				return nil, fmt.Errorf("ProbeExec Run: Unexpected exit code %d, expect in '%v'", exitCode, p.ExitCodes)
			} else {
				return &ProbeStatus{Status: "success"}, nil
			}
		}
		return nil, fmt.Errorf("ProbeExec Run: Unwrap ExitCode %w", err)
	}
	return &ProbeStatus{Status: "success"}, nil
}
