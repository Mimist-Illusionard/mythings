package app

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/Mimist-Illusionard/mythings/config"
	"github.com/Mimist-Illusionard/mythings/internal/infrastructure/postgres"
)

func Run(cfg *config.Config) error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	db, err := postgres.NewConnection(cfg)
	if err != nil {
		return fmt.Errorf("database connect: %w", err)
	}

	tagsRepo := postgres.NewTagsRepo(db)
	itemsRepo := postgres.NewItemsRepo(db)

	return nil
}
