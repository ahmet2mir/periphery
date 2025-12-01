package probe

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"slices"
	"time"

	"go.uber.org/zap"
)

// Ensure implements interface.
var _ ProbeInterface = (*ProbeHTTP)(nil)

type HTTPHeader struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value"`
}

type ProbeHTTP struct {
	Host           string        `yaml:"host"`
	Port           int           `yaml:"port"`
	Path           string        `yaml:"path"`
	Scheme         string        `yaml:"scheme"`
	HTTPHeaders    []HTTPHeader  `yaml:"httpHeaders"`
	ExpectedStatus []int         `yaml:"expectedStatus"`
	RequestTimeout time.Duration `yaml:"requestTimeout"`
}

func (p *ProbeHTTP) Run(ctx context.Context) (*ProbeStatus, error) {
	// Set defaults
	if p.Scheme == "" {
		p.Scheme = "http"
	}
	if p.Path == "" {
		p.Path = "/"
	}
	if p.Host == "" {
		p.Host = "localhost"
	}
	if p.ExpectedStatus == nil {
		p.ExpectedStatus = []int{200}
	}
	if p.RequestTimeout == 0 {
		p.RequestTimeout = 1 * time.Second
	}

	url := fmt.Sprintf("%s://%s:%d%s", p.Scheme, p.Host, p.Port, p.Path)
	zap.S().Debug("ProbeHTTP Run", "url", url)

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: p.RequestTimeout,
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("ProbeHTTP Run: NewRequest %w", err)
	}

	// Add custom headers
	for _, header := range p.HTTPHeaders {
		req.Header.Add(header.Name, header.Value)
	}

	// Execute request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ProbeHTTP Run: Do %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			zap.S().Debug("ProbeHTTP Run: error closing response body", closeErr)
		}
	}()

	// Drain and close body to allow connection reuse
	if _, copyErr := io.Copy(io.Discard, resp.Body); copyErr != nil {
		zap.S().Debug("ProbeHTTP Run: error draining response body", copyErr)
	}

	// Check if status code is expected
	if !slices.Contains(p.ExpectedStatus, resp.StatusCode) {
		return nil, fmt.Errorf("ProbeHTTP Run: Unexpected status code %d, expected one of %v", resp.StatusCode, p.ExpectedStatus)
	}

	zap.S().Debug("ProbeHTTP Run", "status", resp.StatusCode, "success", true)
	return &ProbeStatus{Status: "success"}, nil
}
