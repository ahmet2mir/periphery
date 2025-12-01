package service

import (
	"context"
	"fmt"

	"github.com/coreos/go-systemd/v22/dbus"
	"go.uber.org/zap"
)

type Service struct {
	Name string
	Type string
}

func NewService(name, t string) (*Service, error) {
	return &Service{Name: name, Type: t}, nil
}

func (s *Service) Status() error {
	return nil
}

func (s *Service) Started(ctx context.Context) (bool, error) {
	conn, err := dbus.NewSystemConnectionContext(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to connect to systemd: %w", err)
	}
	defer conn.Close()

	properties, err := conn.GetUnitPropertiesContext(context.Background(), s.Name)
	if err != nil {
		return false, fmt.Errorf("failed to get unit properties: %w", err)
	}

	zap.S().Debug(fmt.Sprintf("Service %s status: %s\n", s.Name, properties["ActiveState"]))

	return true, nil
}

func (s *Service) Restart(ctx context.Context) (bool, error) {
	conn, err := dbus.NewSystemConnectionContext(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to connect to systemd: %w", err)
	}
	defer conn.Close()

	_, err = conn.RestartUnitContext(ctx, s.Name, "replace", nil)
	if err != nil {
		return false, fmt.Errorf("failed to restart service %s: %w", s.Name, err)
	}

	return true, nil
}
