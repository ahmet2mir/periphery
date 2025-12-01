package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/ahmet2mir/periphery/pkg/probe"
	"github.com/ahmet2mir/periphery/pkg/service"
)

type ConfigAPI struct {
	ListenAddress string `yaml:"listenAddress"`
	ListenPort    int    `yaml:"listenPort"`
}

func (ca *ConfigAPI) GetURI() string {
	return fmt.Sprintf("%s:%d", ca.ListenAddress, ca.ListenPort)
}

type Config struct {
	Speaker   Speaker    `yaml:"speaker"`
	BFD       *BFDConfig `yaml:"bfd"`
	API       ConfigAPI  `yaml:"api"`
	Neighbors []Neighbor `yaml:"neighbors"`
	Prefixes  []Prefix   `yaml:"prefixes"`
}

type Speaker struct {
	ASN                        uint32 `yaml:"asn"`
	RouterID                   string `yaml:"routerId"`
	GracefulRestartEnabled     bool   `yaml:"gracefulRestartEnabled"`
	GracefulRestartRestartTime uint32 `yaml:"gracefulRestartRestartTime"`
}

type BFDConfig struct {
	Enabled                     bool          `yaml:"enabled"`
	ListenAddress               string        `yaml:"listenAddress"`
	ListenPort                  int           `yaml:"listenPort"`
	MinimumReceptionInterval    time.Duration `yaml:"minimumReceptionInterval"`
	MinimumTransmissionInterval time.Duration `yaml:"minimumTransmissionInterval"`
	DetectionMultiplier         uint8         `yaml:"detectionMultiplier"`
	Passive                     bool          `yaml:"passive"`
}

func (bc *BFDConfig) GetListenURI() string {
	return fmt.Sprintf("%s:%d", bc.ListenAddress, bc.ListenPort)
}

type Neighbor struct {
	Address             string `yaml:"address"`
	ASN                 uint32 `yaml:"asn"`
	EbgpMultihopEnabled bool   `yaml:"ebgpMultihopEnabled"`
}

type Prefix struct {
	IPAddress              string   `yaml:"ipAddress"`
	Communities            []string `yaml:"communities"`
	NextHop                string   `yaml:"nextHop"`
	ASN                    uint32   `yaml:"asn"`
	MultiExitDescriminator uint32   `yaml:"multiExitDescriminator"`
	AsPathPrepend          []uint32 `yaml:"asPathPrepend"`
	WithdrawOnDown         bool     `yaml:"withdrawOnDown"`
	Maintenance            string   `yaml:"maintenance"`

	Service *service.Service `yaml:"service"`

	LivenessProbe  *probe.Probe `yaml:"livenessProbe"`
	StartupProbe   *probe.Probe `yaml:"startupProbe"`
	ReadinessProbe *probe.Probe `yaml:"readinessProbe"`
}

func New(configPath string) (*Config, error) {
	c := &Config{}
	configBytes, err := os.ReadFile(filepath.Clean(configPath))
	if err != nil {
		return nil, fmt.Errorf("NewConfigFromFile error when reading file %s: %w", configPath, err)
	}

	if err := yaml.Unmarshal(configBytes, &c); err != nil {
		return nil, fmt.Errorf("NewConfigFromFile error unmarshal file %s: %w", configPath, err)
	}

	return c, nil
}
