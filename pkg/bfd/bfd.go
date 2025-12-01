package bfd

import (
	"fmt"
	"strings"

	"github.com/ahmet2mir/periphery/pkg/config"
	lbfd "github.com/rhgb/gobfd/bfd"
	ludp "github.com/rhgb/gobfd/udp"
	"go.uber.org/zap"
)

func Run(cfg *config.BFDConfig) error {
	if cfg == nil {
		return fmt.Errorf("BFD configuration is nil")
	}

	if !cfg.Enabled {
		zap.S().Info("BFD is disabled in configuration")
		return nil
	}

	zap.S().Info("Starting BFD agent", "listenAddress", cfg.GetListenURI())

	loggerInstance, err := zap.NewDevelopment()
	if err != nil {
		return fmt.Errorf("failed to create BFD logger: %w", err)
	}
	defer func() {
		if syncErr := loggerInstance.Sync(); syncErr != nil {
			zap.S().Debug("BFD logger sync error", syncErr)
		}
	}()
	logger := loggerInstance.Sugar()

	c := ludp.AgentConfig{}
	c.IPv4Only = true
	c.ListenAddress = cfg.GetListenURI()
	c.PeerAddresses = lbfd.UniqueStringsSorted(strings.Split("", ","))
	c.DesiredMinTxInterval = uint32(cfg.MinimumTransmissionInterval.Milliseconds())
	c.RequiredMinRxInterval = uint32(cfg.MinimumReceptionInterval.Milliseconds())
	c.DetectMult = cfg.DetectionMultiplier

	_, err = ludp.NewAgent(c, logger)
	if err != nil {
		return fmt.Errorf("error creating BFD agent: %w", err)
	}

	zap.S().Info("BFD agent started successfully")
	return nil
}
