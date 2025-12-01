package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"

	"github.com/ahmet2mir/periphery/pkg/bfd"
	"github.com/ahmet2mir/periphery/pkg/config"
	"github.com/ahmet2mir/periphery/pkg/scheduler"
	"github.com/ahmet2mir/periphery/pkg/speaker"
)

func init() {
	zap.ReplaceGlobals(zap.Must(zap.NewProduction()))
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	c, err := config.New("config.test.yaml")
	if err != nil {
		zap.S().Fatal(err)
	}

	s, err := speaker.New(c, ctx)
	if err != nil {
		zap.S().Fatal(err)
	}

	go s.Serve()
	defer s.Stop()
	if err := s.Start(); err != nil {
		zap.S().Fatal(err)
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
