package storage

import (
	"crypto/sha256"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/adamtc007/KYC-DSL/internal/model"
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
		return nil, fmt.Errorf("postgres connection failed (host=%s, port=%s, dbname=%s): %w", host, port, dbname, err)
	}
	debugLog("Connection successful")

	// Test the connection
	if err := db.Ping(); err != nil {
		if closeErr := db.Close(); closeErr != nil {
			debugLog("Failed to close database after ping failure: %v", closeErr)
		}
		debugLog("Connection ping failed: %v", err)
		return nil, fmt.Errorf("postgres ping failed: %w", err)
	}

	// Set connection pool limits
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

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

	CREATE TABLE IF NOT EXISTS kyc_case_amendments (
		id SERIAL PRIMARY KEY,
		case_name TEXT NOT NULL,
		step TEXT NOT NULL,
		change_type TEXT NOT NULL,
		diff TEXT,
		created_at TIMESTAMP DEFAULT NOW()
	);
	`
	if _, err := db.Exec(schema); err != nil {
		if closeErr := db.Close(); closeErr != nil {
			debugLog("Failed to close database after schema error: %v", closeErr)
		}
		debugLog("Schema creation failed: %v", err)
		return nil, fmt.Errorf("schema creation failed: %w", err)
	}
	debugLog("Schema created/verified successfully")
	return db, nil
}

func InsertCase(db *sqlx.DB, name string) error {
	if name == "" {
		return fmt.Errorf("case name cannot be empty")
	}

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

// InsertAmendment logs a change to a case for audit trail.
func InsertAmendment(db *sqlx.DB, caseName, step, changeType, diff string) error {
	query := `INSERT INTO kyc_case_amendments (case_name, step, change_type, diff) VALUES ($1, $2, $3, $4)`
	_, err := db.Exec(query, caseName, step, changeType, diff)
	if err != nil {
		debugLog("InsertAmendment failed: %v", err)
		return fmt.Errorf("insert amendment failed: %w", err)
	}
	debugLog("Amendment logged for case=%s step=%s type=%s", caseName, step, changeType)
	return nil
}

// GetAmendments retrieves all amendments for a case.
func GetAmendments(db *sqlx.DB, caseName string) ([]Amendment, error) {
	var amendments []Amendment
	query := `SELECT id, case_name, step, change_type, diff, created_at
	          FROM kyc_case_amendments
	          WHERE case_name=$1
	          ORDER BY created_at DESC`
	err := db.Select(&amendments, query, caseName)
	if err != nil {
		return nil, fmt.Errorf("get amendments failed: %w", err)
	}
	return amendments, nil
}

// Amendment represents a logged change to a case.
type Amendment struct {
	ID         int       `db:"id"`
	CaseName   string    `db:"case_name"`
	Step       string    `db:"step"`
	ChangeType string    `db:"change_type"`
	Diff       string    `db:"diff"`
	CreatedAt  time.Time `db:"created_at"`
}

// CaseVersion represents a versioned snapshot of a case.
type CaseVersion struct {
	CaseName    string    `db:"case_name"`
	Version     int       `db:"version"`
	DslSnapshot string    `db:"dsl_snapshot"`
	Hash        string    `db:"hash"`
	CreatedAt   time.Time `db:"created_at"`
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

// GetLatestDSL fetches the most recent serialized DSL for a case.
func GetLatestDSL(db *sqlx.DB, caseName string) (string, error) {
	if db == nil {
		return "", fmt.Errorf("database connection is nil")
	}
	if caseName == "" {
		return "", fmt.Errorf("case name is required")
	}

	var dsl string
	err := db.Get(&dsl, `
		SELECT dsl_snapshot FROM kyc_case_versions
		WHERE case_name=$1
		ORDER BY version DESC LIMIT 1
	`, caseName)
	if err != nil {
		return "", fmt.Errorf("get latest DSL failed for case %s: %w", caseName, err)
	}
	return dsl, nil
}

// LogAmendment records an applied mutation step.
func LogAmendment(db *sqlx.DB, caseName, step, diff string) error {
	_, err := db.Exec(`
		INSERT INTO kyc_case_amendments (case_name, step, change_type, diff)
		VALUES ($1, $2, 'mutation', $3)
	`, caseName, step, diff)
	return err
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

// RecordValidationResult persists the outcome of validation for audit trail.
// Compliant with FCA SYSC, MAS 626 §4.2, HKMA AML §3.6, EU AMLD6 Article 30.
func RecordValidationResult(db *sqlx.DB, v model.CaseValidation) error {
	if db == nil {
		return fmt.Errorf("database connection is nil")
	}

	query := `
		INSERT INTO kyc_case_validations
		(case_name, version, grammar_version, ontology_version, validator_actor,
		 validation_status, error_message, total_checks, passed_checks, failed_checks)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		RETURNING id
	`
	var id int
	err := db.QueryRow(query,
		v.CaseName, v.Version, v.GrammarVersion, v.OntologyVersion, v.ValidatorActor,
		v.ValidationStatus, v.ErrorMessage, v.TotalChecks, v.PassedChecks, v.FailedChecks,
	).Scan(&id)

	if err != nil {
		debugLog("RecordValidationResult failed: %v", err)
		return fmt.Errorf("record validation result failed (case=%s): %w", v.CaseName, err)
	}

	debugLog("Validation recorded: case=%s, status=%s, id=%d", v.CaseName, v.ValidationStatus, id)
	return nil
}

// RecordValidationFinding records a detailed validation finding.
func RecordValidationFinding(db *sqlx.DB, f model.ValidationFinding) error {
	if db == nil {
		return fmt.Errorf("database connection is nil")
	}

	query := `
		INSERT INTO kyc_validation_findings
		(validation_id, check_type, check_name, check_status, check_message, entity_ref, severity)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
	`
	_, err := db.Exec(query,
		f.ValidationID, f.CheckType, f.CheckName, f.CheckStatus, f.CheckMessage, f.EntityRef, f.Severity)

	if err != nil {
		debugLog("RecordValidationFinding failed: %v", err)
		return fmt.Errorf("record validation finding failed: %w", err)
	}

	return nil
}

// GetValidationHistory retrieves validation audit trail for a case.
func GetValidationHistory(db *sqlx.DB, caseName string) ([]model.CaseValidation, error) {
	if db == nil {
		return nil, fmt.Errorf("database connection is nil")
	}
	if caseName == "" {
		return nil, fmt.Errorf("case name is required")
	}

	var validations []model.CaseValidation
	query := `
		SELECT id, case_name, version, validation_time, grammar_version, ontology_version,
		       validator_actor, validation_status, error_message, total_checks,
		       passed_checks, failed_checks, created_at
		FROM kyc_case_validations
		WHERE case_name = $1
		ORDER BY validation_time DESC
	`
	err := db.Select(&validations, query, caseName)
	if err != nil {
		return nil, fmt.Errorf("get validation history failed for case %s: %w", caseName, err)
	}
	return validations, nil
}

// RecordLineageEvaluation persists a lineage evaluation result for audit trail.
func RecordLineageEvaluation(db *sqlx.DB, caseName string, caseVersion int, result interface{}) error {
	if db == nil {
		return fmt.Errorf("database connection is nil")
	}

	// Type assertion to get evaluation result
	type EvalResult struct {
		DerivedCode string
		Value       interface{}
		Success     bool
		Error       string
		Rule        string
		Inputs      map[string]interface{}
	}

	r, ok := result.(EvalResult)
	if !ok {
		return fmt.Errorf("invalid evaluation result type")
	}

	// Determine value type
	valueType := "string"
	var valueStr string
	switch v := r.Value.(type) {
	case bool:
		valueType = "boolean"
		valueStr = fmt.Sprintf("%v", v)
	case int, int64, float64:
		valueType = "numeric"
		valueStr = fmt.Sprintf("%v", v)
	case string:
		valueType = "string"
		valueStr = v
	default:
		valueStr = fmt.Sprintf("%v", v)
	}

	// Convert inputs to JSON
	var inputsJSON interface{}
	if r.Inputs != nil {
		inputsJSON = r.Inputs
	}

	query := `
		INSERT INTO kyc_lineage_evaluations
		(case_name, case_version, derived_code, value, value_type, success, error, inputs, rule)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := db.Exec(query,
		caseName, caseVersion, r.DerivedCode, valueStr, valueType,
		r.Success, r.Error, inputsJSON, r.Rule)

	if err != nil {
		debugLog("RecordLineageEvaluation failed: %v", err)
		return fmt.Errorf("record lineage evaluation failed (case=%s, derived=%s): %w",
			caseName, r.DerivedCode, err)
	}

	debugLog("Lineage evaluation recorded: case=%s, derived=%s, success=%v",
		caseName, r.DerivedCode, r.Success)
	return nil
}

// GetLineageEvaluations retrieves evaluation history for a case.
func GetLineageEvaluations(db *sqlx.DB, caseName string) ([]map[string]interface{}, error) {
	if db == nil {
		return nil, fmt.Errorf("database connection is nil")
	}
	if caseName == "" {
		return nil, fmt.Errorf("case name is required")
	}

	var results []map[string]interface{}
	query := `
		SELECT id, case_name, case_version, derived_code, value, value_type,
		       success, error, inputs, rule, evaluated_at
		FROM kyc_lineage_evaluations
		WHERE case_name = $1
		ORDER BY evaluated_at DESC
	`

	rows, err := db.Queryx(query, caseName)
	if err != nil {
		return nil, fmt.Errorf("get lineage evaluations failed for case %s: %w", caseName, err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			debugLog("Failed to close rows: %v", closeErr)
		}
	}()

	for rows.Next() {
		result := make(map[string]interface{})
		if err := rows.MapScan(result); err != nil {
			return nil, fmt.Errorf("scan lineage evaluation failed: %w", err)
		}
		results = append(results, result)
	}

	return results, nil
}
