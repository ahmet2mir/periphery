package scheduler

import (
	"context"
	"time"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"

	"github.com/ahmet2mir/periphery/pkg/config"
	"github.com/ahmet2mir/periphery/pkg/probe"
	"github.com/ahmet2mir/periphery/pkg/speaker"
)

func RunScheduler(ctx context.Context, p config.Prefix, s *speaker.Speaker) {
	cron := cron.New(cron.WithSeconds())

	svc, err := p.Service.Started(ctx)
	if err != nil || !svc {
		zap.S().Warn(err)
		return
	}

	if p.StartupProbe != nil {
		if p.StartupProbe.InitialDelaySeconds > 0 {
			zap.S().Info("p.StartupProbe.InitialDelaySeconds", p.StartupProbe.InitialDelaySeconds)
			time.Sleep(p.StartupProbe.InitialDelaySeconds)
		}

		pm := probe.NewProbeManager()
		_, err := pm.Run(ctx, p.StartupProbe, p.Service)
		if err != nil {
			zap.S().Warn(err)
			zap.S().Info("p.StartupProbe.PeriodSeconds", p.StartupProbe.PeriodSeconds)
			time.Sleep(p.StartupProbe.PeriodSeconds)
			return
		}
	}

	if p.LivenessProbe != nil {
		if p.LivenessProbe.InitialDelaySeconds > 0 {
			zap.S().Info("p.LivenessProbe.InitialDelaySeconds", p.LivenessProbe.InitialDelaySeconds)
			time.Sleep(p.LivenessProbe.InitialDelaySeconds)
		}

		pm := probe.NewProbeManager()
		_, err := cron.AddFunc("@every "+p.LivenessProbe.PeriodSeconds.String(), func() {
			svc, err := p.Service.Started(ctx)
			if err != nil || !svc {
				zap.S().Warn(err)
				if _, restartErr := p.Service.Restart(ctx); restartErr != nil {
					zap.S().Error("Failed to restart service", restartErr)
				}
			} else if status, err := pm.Run(ctx, p.LivenessProbe, p.Service); err != nil {
				zap.S().Error("SchedulerProbeError: LivenessProbe", err, "  ", pm.NumberFailure, "==", pm.NumberSuccess)
				if _, restartErr := p.Service.Restart(ctx); restartErr != nil {
					zap.S().Error("Failed to restart service", restartErr)
				}
			} else {
				zap.S().Info("SchedulerProbe: LivenessProbe", status.Status, "  ", pm.NumberFailure, "==", pm.NumberSuccess)
			}
		})
		if err != nil {
			zap.S().Error("SchedulerProbeError: Failed to schedule LivenessProbe", err)
		}
	}

	if p.ReadinessProbe != nil {
		if p.ReadinessProbe.InitialDelaySeconds > 0 {
			zap.S().Info("p.ReadinessProbe.InitialDelaySeconds", p.ReadinessProbe.InitialDelaySeconds)
			time.Sleep(p.ReadinessProbe.InitialDelaySeconds)
		}

		_, err := cron.AddFunc("@every "+p.ReadinessProbe.PeriodSeconds.String(), func() {
			pm := probe.NewProbeManager()
			if status, err := pm.Run(ctx, p.ReadinessProbe, p.Service); err != nil {
				zap.S().Error("SchedulerProbeError: ReadinessProbe => %w", err, "  ", pm.NumberFailure, "==", pm.NumberSuccess)
				if delErr := s.DeletePath(p); delErr != nil {
					zap.S().Error("Failed to delete path", delErr)
				}

			} else {
				zap.S().Info("SchedulerProbe: ReadinessProbe => %s", status.Status, "  ", pm.NumberFailure, "==", pm.NumberSuccess, "AddPath")
				if err := s.AddPath(p); err != nil {
					zap.S().Error("SchedulerProbeError: Failed to addpath", err)
				}
			}
		})
		if err != nil {
			zap.S().Error("SchedulerProbeError: Failed to schedule ReadinessProbe => %w", err)
			if delErr := s.DeletePath(p); delErr != nil {
				zap.S().Error("Failed to delete path", delErr)
			}
		}
	}

	cron.Start()
}
