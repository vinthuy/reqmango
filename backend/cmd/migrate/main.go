package main

import (
	"fmt"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbUser := envOrDefault("DB_USER", "reqmango")
		dbPass := envOrDefault("DB_PASSWORD", "reqmango")
		dbHost := envOrDefault("DB_HOST", "localhost")
		dbPort := envOrDefault("DB_PORT", "5432")
		dbName := envOrDefault("DB_NAME", "reqmango")
		dbURL = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
			dbUser, dbPass, dbHost, dbPort, dbName)
	}

	m, err := migrate.New("file://migrations", dbURL)
	if err != nil {
		log.Fatalf("migrate: init failed: %v", err)
	}

	cmd := "up"
	if len(os.Args) > 1 {
		cmd = os.Args[1]
	}

	switch cmd {
	case "up":
		err = m.Up()
	case "down":
		err = m.Down()
	case "down-all":
		err = m.Down()
		for err == nil {
			err = m.Down()
		}
		if err == migrate.ErrNoChange {
			err = nil
		}
	case "version":
		v, dirty, verr := m.Version()
		if verr != nil {
			log.Fatalf("migrate: version failed: %v", verr)
		}
		fmt.Printf("version=%d dirty=%v\n", v, dirty)
		return
	case "force":
		if len(os.Args) < 3 {
			log.Fatal("migrate: force requires version number")
		}
		var ver int
		fmt.Sscanf(os.Args[2], "%d", &ver)
		err = m.Force(ver)
	case "steps":
		if len(os.Args) < 3 {
			log.Fatal("migrate: steps requires number")
		}
		var n int
		fmt.Sscanf(os.Args[2], "%d", &n)
		err = m.Steps(n)
	default:
		log.Fatalf("migrate: unknown command %q (use: up, down, version, force, steps)", cmd)
	}

	if err != nil && err != migrate.ErrNoChange {
		log.Fatalf("migrate: %s failed: %v", cmd, err)
	}
	fmt.Printf("migrate %s: OK\n", cmd)
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
