package db

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

func RunMigrations(db *sql.DB) {
	//files := []string{
	//	"migrations/01_create_migrations_table.sql",
	//}
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get current working directory: %v", err)
	}
	migrationPattern := filepath.Join(cwd, "migrations", "*.sql")
	files, err := filepath.Glob(migrationPattern)
	fmt.Println("Found %d migrations\n", len(files))
	if err != nil {
		log.Fatalf("failed to read migration files: %v", err)
	}

	for _, file := range files {
		fmt.Println("Running migration:", file)
		sqlBytes, err := ioutil.ReadFile(file)
		if err != nil {
			log.Fatalf("Failed to read migration %s: %v", file, err)
		}

		if _, err := db.Exec(string(sqlBytes)); err != nil {
			log.Fatalf("Failed to execute migration %s: %v", file, err)
		}
	}
}
