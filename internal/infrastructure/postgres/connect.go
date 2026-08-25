package postgres

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/Mimist-Illusionard/mythings/config"
)

func NewConnection(cfg *config.Config) (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
		cfg.DBParams.Host, cfg.DBParams.User, cfg.DBParams.Pass, cfg.DBParams.Name, cfg.DBParams.Port,
	)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("db connect: %w", err)
	}

	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(10 * time.Minute)
	db.SetConnMaxIdleTime(30 * time.Minute)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("db ping: %w", err)
	}

	migrationDSN := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.DBParams.User,
		cfg.DBParams.Pass,
		cfg.DBParams.Host,
		cfg.DBParams.Port,
		cfg.DBParams.Name,
	)

	if err := RunMigration(migrationDSN); err != nil {
		return nil, fmt.Errorf("db migration: %w", err)
	}

	return db, nil
}
