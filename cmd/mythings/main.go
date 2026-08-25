package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"

	"github.com/Mimist-Illusionard/mythings/internal/config"
	"github.com/Mimist-Illusionard/mythings/internal/repository/postgres"

	_ "github.com/jackc/pgx/v5/stdlib"
)

var envPath string
var httpPort string

func main() {
	flag.StringVar(&envPath, "env", ".env", "path to .env file")
	flag.StringVar(&httpPort, "http", "8080", "application port")
	flag.Parse()

	cfg, err := config.New(httpPort, envPath)
	if err != nil {
		log.Fatal(fmt.Errorf("config load: %w", err))
	}

	db, err := postgres.NewDatabase(cfg)
	if err != nil {
		log.Fatal(fmt.Errorf("database connect: %w", err))
	}
}
