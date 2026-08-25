package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/Mimist-Illusionard/mythings/config"
	"github.com/Mimist-Illusionard/mythings/internal/infrastructure/postgres"
	httpHandlers "github.com/Mimist-Illusionard/mythings/internal/interfaces/http/handlers"
)

func Run(cfg *config.Config) error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	db, err := postgres.NewConnection(cfg)
	if err != nil {
		return fmt.Errorf("database connect: %w", err)
	}
	defer db.Close()

	tagsRepo := postgres.NewTagsRepo(db)
	itemsRepo := postgres.NewItemsRepo(db)

	handler := httpHandlers.New(itemsRepo, tagsRepo)
	router := handler.Router()

	static := http.FileServer(http.Dir("web/static"))
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", static))
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, "web/index.html")
	}).Methods(http.MethodGet)

	server := &http.Server{
		Addr:              ":" + cfg.HTTPPort,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("http shutdown: %w", err)
		}
		return nil
	case err := <-errCh:
		return fmt.Errorf("http server: %w", err)
	}
}
