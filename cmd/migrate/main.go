package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"be-eventgate/config"
)

func main() {
	cmd := flag.String("cmd", "up", "Command to run: 'up' to apply migrations, 'down' to rollback migrations")
	flag.Parse()

	if len(os.Args) > 1 && (os.Args[1] == "up" || os.Args[1] == "down") {
		*cmd = os.Args[1]
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load environment configuration: %v", err)
	}

	db, err := config.InitDB(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v\nMake sure PostgreSQL is running and DB credentials in .env are correct.", err)
	}
	defer db.Close()

	var migrationFiles []string
	if *cmd == "up" {
		migrationFiles = []string{
			filepath.Join("migrations", "000001_create_initial_schema.up.sql"),
			filepath.Join("migrations", "000002_seed_initial_roles.up.sql"),
			filepath.Join("migrations", "000003_add_event_pricing_quota.up.sql"),
			filepath.Join("migrations", "000004_add_event_version_to_events.up.sql"),
			filepath.Join("migrations", "000005_update_question_type_options.up.sql"),
			filepath.Join("migrations", "000006_rename_pending_payment_status.up.sql"),
		}
		log.Println("Starting database migration UP...")
	} else if *cmd == "down" {
		migrationFiles = []string{
			filepath.Join("migrations", "000006_rename_pending_payment_status.down.sql"),
			filepath.Join("migrations", "000005_update_question_type_options.down.sql"),
			filepath.Join("migrations", "000004_add_event_version_to_events.down.sql"),
			filepath.Join("migrations", "000003_add_event_pricing_quota.down.sql"),
			filepath.Join("migrations", "000002_seed_initial_roles.down.sql"),
			filepath.Join("migrations", "000001_create_initial_schema.down.sql"),
		}
		log.Println("Starting database migration DOWN...")
	} else {
		log.Fatalf("Unknown migration command '%s'. Valid commands: 'up', 'down'", *cmd)
	}

	for _, file := range migrationFiles {
		log.Printf("Executing migration file: %s", file)
		content, err := os.ReadFile(file)
		if err != nil {
			log.Fatalf("Failed to read migration file %s: %v", file, err)
		}

		tx, err := db.Begin()
		if err != nil {
			log.Fatalf("Failed to begin transaction for %s: %v", file, err)
		}

		if _, err := tx.Exec(string(content)); err != nil {
			_ = tx.Rollback()
			log.Fatalf("Failed to execute SQL in %s: %v", file, err)
		}

		if err := tx.Commit(); err != nil {
			log.Fatalf("Failed to commit transaction for %s: %v", file, err)
		}

		log.Printf("Successfully executed %s", file)
	}

	fmt.Printf("\nDatabase migration '%s' completed successfully!\n", *cmd)
}
