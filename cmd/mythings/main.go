package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/Mimist-Illusionard/mythings/config"
	"github.com/Mimist-Illusionard/mythings/internal/app"

	_ "github.com/jackc/pgx/v5/stdlib"
)

var envPath string
var httpPort string

func main() {
	flag.StringVar(&envPath, "env", ".env", "path to .env file")
	flag.StringVar(&httpPort, "port", "8080", "application port")
	flag.Parse()

	cfg, err := config.New(httpPort, envPath)
	if err != nil {
		log.Fatal(fmt.Errorf("config load: %w", err))
	}

	if err := app.Run(cfg); err != nil {
		log.Fatal(fmt.Errorf("app run: %w", err))
	}
}

// дата приобретения
// курс доллара
// цена товара
// вcтавлять картинку перетаскиванием
