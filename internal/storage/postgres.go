package storage

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

const DEBUG = true

func debugLog(format string, args ...interface{}) {
	if DEBUG {
		log.Printf("[STORAGE DEBUG] "+format, args...)
	}
}

func ConnectPostgres() (*sqlx.DB, error) {
	debugLog("=== STORAGE BREAKPOINT 1: ConnectPostgres called ===")
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

	debugLog("Connection parameters: host=%s, port=%s, user=%s, dbname=%s", host, port, user, dbname)

	connStr := fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=disable", host, port, user, dbname)
	if password != "" {
		connStr = fmt.Sprintf("%s password=%s", connStr, password)
	}

	debugLog("=== STORAGE BREAKPOINT 2: Attempting to connect ===")
	db, err := sqlx.Connect("postgres", connStr)
	if err != nil {
		debugLog("Connection failed: %v", err)
		return nil, fmt.Errorf("connect failed: %w", err)
	}
	debugLog("Connection successful")

	debugLog("=== STORAGE BREAKPOINT 3: Creating schema ===")
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
	debugLog("Schema created/verified successfully")
	return db, nil
}

func InsertCase(db *sqlx.DB, name string) error {
	debugLog("=== STORAGE BREAKPOINT 4: InsertCase called with name='%s' ===", name)
	query := `INSERT INTO kyc_cases (name, status, last_updated) VALUES ($1, 'pending', $2)`
	debugLog("Executing query: %s", query)
	debugLog("Parameters: name=%s, timestamp=%v", name, time.Now())

	result, err := db.Exec(query, name, time.Now())
	if err != nil {
		debugLog("Insert failed with error: %v", err)
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	debugLog("=== STORAGE BREAKPOINT 5: Insert successful, rows affected: %d ===", rowsAffected)
	return nil
}
