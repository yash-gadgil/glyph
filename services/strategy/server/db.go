package server

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func InitDB() *sql.DB {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		os.Getenv("STRATEGY_DB_USER"),
		os.Getenv("STRATEGY_DB_PASSWORD"),
		os.Getenv("STRATEGY_DB_HOST"),
		os.Getenv("STRATEGY_DB_PORT"),
		os.Getenv("STRATEGY_DB"),
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("failed to open db: %v", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}

	log.Println("connected to postgres")
	return db
}
