// Package main provides a CLI tool for running database migrations.
package main

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	if len(os.Args) < 2 {
		log.Fatal("usage: migrate <up|down|version>")
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	m, err := migrate.New("file://migrations", dbURL)
	if err != nil {
		log.Fatalf("init migrate: %v", err)
	}
	defer func() { _, _ = m.Close() }()

	switch os.Args[1] {
	case "up":
		if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			log.Fatalf("migrate up: %v", err)
		}
		log.Println("migrations applied")
	case "down":
		if err := m.Down(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			log.Fatalf("migrate down: %v", err)
		}
		log.Println("migrations rolled back")
	case "version":
		version, dirty, err := m.Version()
		if err != nil {
			log.Fatalf("migrate version: %v", err)
		}
		fmt.Printf("version: %d, dirty: %v\n", version, dirty)
	default:
		log.Fatalf("unknown command: %s (use up, down, or version)", os.Args[1])
	}
}
