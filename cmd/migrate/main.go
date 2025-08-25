package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"vector/internal/migrations"

	"github.com/joho/godotenv"
)

func main() {
	var (
		action = flag.String("action", "", "Action to perform: up, seed")
		help   = flag.Bool("help", false, "Show help")
	)
	flag.Parse()

	if *help || *action == "" {
		printHelp()
		return
	}
	_ = godotenv.Load()

	migrator, err := migrations.New()
	if err != nil {
		log.Fatalf("Failed to initialize migrator: %v", err)
	}

	switch *action {
	case "up":
		if err := migrator.MigrateAll(); err != nil {
			log.Fatalf("Migration failed: %v", err)
		}
		log.Println("✅ All migrations completed successfully")
	case "seed":
		if err := migrator.SeedUsers(); err != nil {
			log.Fatalf("Seeding failed: %v", err)
		}
		log.Println("✅ Seeding completed successfully")
	default:
		fmt.Printf("Unknown action: %s\n", *action)
		printHelp()
		os.Exit(1)
	}

}

func printHelp() {
	fmt.Println("Database Migration Tool")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  go run cmd/migrate/main.go -action=<action>")
	fmt.Println("  go run cmd/migrate/main.go -action=<action>")
	fmt.Println()
	fmt.Println("Actions:")
	fmt.Println("  up    - Run all database migrations")
	fmt.Println("  seed  - Seed default users")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  go run cmd/migrate/main.go -action=up")
	fmt.Println("  go run cmd/migrate/main.go -action=seed")
	fmt.Println()
	fmt.Println("Environment variables:")
	fmt.Println("  DB_HOST, DB_PORT, POSTGRES_USER, POSTGRES_PASSWORD, POSTGRES_DB")
}
