package db

import (
	"database/sql"
	_ "github.com/lib/pq"
	"log"
)

var DB *sql.DB

func Init() {
	connStr := "host=localhost port=5432 user=rithvik password=root dbname=dropbox_clone sslmode=disable"

	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("DB connection error:", err)
	}

	if err = DB.Ping(); err != nil {
		log.Fatal("DB ping error:", err)
	}

	RunMigrations(DB)
}
