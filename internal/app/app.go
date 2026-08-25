package app

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/Mimist-Illusionard/mythings/config"
	"github.com/Mimist-Illusionard/mythings/internal/infrastructure/postgres"
	"github.com/Mimist-Illusionard/mythings/internal/interfaces/http/handlers"
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

	router := handlers.New(itemsRepo, tagsRepo)

	server := &http.Server{
		Addr:              ":" + cfg.HTTPPort,
		IdleTimeout:       60 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      5 * time.Second,
		Handler:           router.Router(),
	}

	errCh := make(chan error, 1)
	go func() {
		log.Printf("http server listening on %s", server.Addr)
		err := server.ListenAndServe()
		if errors.Is(err, http.ErrServerClosed) {
			err = nil
		}
		errCh <- err
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			_ = server.Close()
			return fmt.Errorf("shutdown http server: %w", err)
		}

		return <-errCh
	}
}
