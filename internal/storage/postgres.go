package storage

import (
	"crypto/sha256"
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

	CREATE TABLE IF NOT EXISTS kyc_case_versions (
		id SERIAL PRIMARY KEY,
		case_name TEXT NOT NULL,
		version INT NOT NULL,
		dsl_snapshot TEXT,
		hash TEXT,
		created_at TIMESTAMP DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS kyc_grammar (
		id SERIAL PRIMARY KEY,
		name TEXT UNIQUE,
		version TEXT,
		ebnf TEXT,
		created_at TIMESTAMP DEFAULT NOW()
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

// InsertVersion stores a DSL snapshot with its hash for audit trail.
func InsertVersion(db *sqlx.DB, caseName string, version int, dsl string) error {
	hash := sha256Hex(dsl)
	query := `INSERT INTO kyc_case_versions (case_name, version, dsl_snapshot, hash) VALUES ($1, $2, $3, $4)`
	_, err := db.Exec(query, caseName, version, dsl, hash)
	if err != nil {
		debugLog("InsertVersion failed: %v", err)
		return err
	}
	debugLog("Version inserted for case=%s version=%d hash=%s", caseName, version, hash)
	return nil
}

func sha256Hex(input string) string {
	sum := sha256.Sum256([]byte(input))
	return fmt.Sprintf("%x", sum)
}

// GetNextVersion returns the next version number for a given case.
func GetNextVersion(db *sqlx.DB, caseName string) (int, error) {
	var current int
	err := db.Get(&current, "SELECT COALESCE(MAX(version), 0) FROM kyc_case_versions WHERE case_name=$1", caseName)
	if err != nil {
		return 1, err
	}
	return current + 1, nil
}

// SaveCaseVersion handles auto-versioning and persistence of a serialized DSL snapshot.
func SaveCaseVersion(db *sqlx.DB, caseName, dsl string) error {
	nextVer, err := GetNextVersion(db, caseName)
	if err != nil {
		return fmt.Errorf("failed to get next version: %w", err)
	}
	hash := sha256Hex(dsl)
	query := `INSERT INTO kyc_case_versions (case_name, version, dsl_snapshot, hash) VALUES ($1, $2, $3, $4)`
	_, err = db.Exec(query, caseName, nextVer, dsl, hash)
	if err != nil {
		return fmt.Errorf("insert version failed: %w", err)
	}
	fmt.Printf("📜 Case %s saved version %d (hash=%s)\n", caseName, nextVer, hash[:12])
	return nil
}

func InsertGrammar(db *sqlx.DB, name, version, ebnf string) error {
	query := `
		INSERT INTO kyc_grammar (name, version, ebnf)
		VALUES ($1, $2, $3)
		ON CONFLICT (name) DO UPDATE
		SET version = EXCLUDED.version, ebnf = EXCLUDED.ebnf, created_at = NOW();
	`
	_, err := db.Exec(query, name, version, ebnf)
	if err != nil {
		return fmt.Errorf("insert grammar failed: %w", err)
	}
	fmt.Printf("📘 Grammar '%s' (v%s) stored in Postgres.\n", name, version)
	return nil
}

func GetGrammar(db *sqlx.DB, name string) (string, error) {
	var ebnf string
	err := db.Get(&ebnf, "SELECT ebnf FROM kyc_grammar WHERE name=$1", name)
	if err != nil {
		return "", err
	}
	return ebnf, nil
}
