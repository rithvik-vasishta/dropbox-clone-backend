package db

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
)

func RunMigrations(db *sql.DB) {
	files := []string{
		"migrations/01_create_migrations_table.sql",
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
