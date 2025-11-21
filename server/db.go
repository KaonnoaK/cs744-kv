package main

import (
	"database/sql"
	"log"
	"time"
)

func InitDB(cfg Config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.DbURL)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(50)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(time.Minute * 30)

	for i := 0; i < 5; i++ {
		if err = db.Ping(); err == nil {
			return db, nil
		}
		log.Printf("Retrying DB connection...")
		time.Sleep(2 * time.Second)
	}
	return nil, err
}

func ensureSchema(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS kv (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL
		);
	`)
	return err
}

