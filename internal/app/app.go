package app

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/Mimist-Illusionard/mythings/config"
	"github.com/Mimist-Illusionard/mythings/internal/infrastructure/repository/postgres"
)

func Run(cfg *config.Config) error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	db, err := postgres.NewDatabase(cfg)
	if err != nil {
		return fmt.Errorf("database connect: %w", err)
	}

	return nil
}
