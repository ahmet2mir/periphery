package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/ahmet2mir/periphery/pkg/bfd"
	"github.com/ahmet2mir/periphery/pkg/config"
	"github.com/ahmet2mir/periphery/pkg/logger"
	"github.com/ahmet2mir/periphery/pkg/metrics"
	"github.com/ahmet2mir/periphery/pkg/scheduler"
	"github.com/ahmet2mir/periphery/pkg/speaker"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	c, err := config.New("config.test.yaml")
	if err != nil {
		tempLogger := zap.Must(zap.NewProduction())
		tempLogger.Sugar().Fatal(err)
	}

	cleanup, err := logger.Initialize(c.Logging)
	if err != nil {
		tempLogger := zap.Must(zap.NewProduction())
		tempLogger.Sugar().Fatalf("Failed to initialize logger: %v", err)
	}
	defer cleanup()

	s, err := speaker.New(c, ctx)
	if err != nil {
		zap.S().Fatal(err)
	}

	go s.Serve()
	defer s.Stop()
	if err := s.Start(); err != nil {
		zap.S().Fatal(err)
	}

	if c.Metrics != nil && c.Metrics.Enabled {
		metricsServer := metrics.NewServer(c.Metrics.ListenAddress, c.Metrics.ListenPort)
		go func() {
			if err := metricsServer.Start(ctx); err != nil {
				zap.S().Error("Metrics server error:", err)
			}
		}()
		defer func() {
			if err := metricsServer.Stop(); err != nil {
				zap.S().Error("Metrics server stop error:", err)
			}
		}()

		interval := c.Metrics.Interval
		if interval == 0 {
			interval = 15 * time.Second
		}
		collector := metrics.NewGoBGPCollector(s.Server, ctx, interval)
		collector.Start()
		defer collector.Stop()
	}

	for _, p := range c.Prefixes {
		go scheduler.RunScheduler(ctx, p, s)
	}

	go func() {
		<-ctx.Done()
		zap.S().Info("Shutting down gracefully...")
		cancel()
	}()

	if c.BFD != nil && c.BFD.Enabled {
		go func() {
			if err := bfd.Run(c.BFD); err != nil {
				zap.S().Error("BFD error:", err)
			}
		}()
	}

	select {}
}
