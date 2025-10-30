package storage

import (
	"fmt"
	"os"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func ConnectPostgres() (*sqlx.DB, error) {
	host := os.Getenv("PGHOST")
	if host == "" {
		host = "localhost"
	}
	port := os.Getenv("PGPORT")
	if port == "" {
		port = "5432"
	}
	user := os.Getenv("PGUSER")
	if user == "" {
		user = "adamtc007"
	}
	password := os.Getenv("PGPASSWORD")
	dbname := os.Getenv("PGDATABASE")
	if dbname == "" {
		dbname = "kyc_dsl"
	}

	connStr := fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=disable", host, port, user, dbname)
	if password != "" {
		connStr = fmt.Sprintf("%s password=%s", connStr, password)
	}
	db, err := sqlx.Connect("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("connect failed: %w", err)
	}

	schema := `
	CREATE TABLE IF NOT EXISTS kyc_cases (
		id SERIAL PRIMARY KEY,
		name TEXT NOT NULL,
		version INT DEFAULT 1,
		status TEXT DEFAULT 'pending',
		last_updated TIMESTAMP DEFAULT NOW()
	);
	`
	db.MustExec(schema)
	return db, nil
}

func InsertCase(db *sqlx.DB, name string) error {
	query := `INSERT INTO kyc_cases (name, status, last_updated) VALUES ($1, 'pending', $2)`
	_, err := db.Exec(query, name, time.Now())
	return err
}
