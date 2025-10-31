package storage

import (
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

// CaseVersionInfo holds metadata about a case version
type CaseVersionInfo struct {
	Version   int       `db:"version"`
	Hash      string    `db:"hash"`
	CreatedAt time.Time `db:"created_at"`
}

// CaseSummary holds summary information about a case
type CaseSummary struct {
	Name         string    `db:"name"`
	VersionCount int       `db:"version_count"`
	Status       string    `db:"status"`
	LastUpdated  time.Time `db:"last_updated"`
}

// GetCaseVersion retrieves a specific version of a case
func GetCaseVersion(db *sqlx.DB, caseName string, version int) (string, string, error) {
	if db == nil {
		return "", "", fmt.Errorf("database connection is nil")
	}
	if caseName == "" {
		return "", "", fmt.Errorf("case name is required")
	}
	if version <= 0 {
		return "", "", fmt.Errorf("version must be positive")
	}

	var result struct {
		DslSnapshot string `db:"dsl_snapshot"`
		Hash        string `db:"hash"`
	}

	query := `
		SELECT dsl_snapshot, hash
		FROM kyc_case_versions
		WHERE case_name = $1 AND version = $2
	`

	err := db.Get(&result, query, caseName, version)
	if err != nil {
		return "", "", fmt.Errorf("failed to get case '%s' version %d: %w", caseName, version, err)
	}

	return result.DslSnapshot, result.Hash, nil
}

// GetLatestCaseWithMetadata retrieves the latest version with metadata
func GetLatestCaseWithMetadata(db *sqlx.DB, caseName string) (string, int, string, error) {
	if db == nil {
		return "", 0, "", fmt.Errorf("database connection is nil")
	}
	if caseName == "" {
		return "", 0, "", fmt.Errorf("case name is required")
	}

	var result struct {
		DslSnapshot string `db:"dsl_snapshot"`
		Version     int    `db:"version"`
		Hash        string `db:"hash"`
	}

	query := `
		SELECT dsl_snapshot, version, hash
		FROM kyc_case_versions
		WHERE case_name = $1
		ORDER BY version DESC
		LIMIT 1
	`

	err := db.Get(&result, query, caseName)
	if err != nil {
		return "", 0, "", fmt.Errorf("failed to get latest case '%s': %w", caseName, err)
	}

	return result.DslSnapshot, result.Version, result.Hash, nil
}

// ListCaseVersions lists all versions of a specific case
func ListCaseVersions(db *sqlx.DB, caseName string) ([]CaseVersionInfo, error) {
	if db == nil {
		return nil, fmt.Errorf("database connection is nil")
	}
	if caseName == "" {
		return nil, fmt.Errorf("case name is required")
	}

	var versions []CaseVersionInfo

	query := `
		SELECT version, hash, created_at
		FROM kyc_case_versions
		WHERE case_name = $1
		ORDER BY version ASC
	`

	err := db.Select(&versions, query, caseName)
	if err != nil {
		return nil, fmt.Errorf("failed to list versions for case '%s': %w", caseName, err)
	}

	return versions, nil
}

// ListAllCases lists all cases with summary information
func ListAllCases(db *sqlx.DB) ([]CaseSummary, error) {
	if db == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	var cases []CaseSummary

	query := `
		SELECT
			c.name,
			COUNT(v.version) as version_count,
			c.status,
			c.last_updated
		FROM kyc_cases c
		LEFT JOIN kyc_case_versions v ON c.name = v.case_name
		GROUP BY c.name, c.status, c.last_updated
		ORDER BY c.last_updated DESC
	`

	err := db.Select(&cases, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list cases: %w", err)
	}

	return cases, nil
}

// GetCaseByName retrieves case metadata by name
func GetCaseByName(db *sqlx.DB, caseName string) (*CaseSummary, error) {
	if db == nil {
		return nil, fmt.Errorf("database connection is nil")
	}
	if caseName == "" {
		return nil, fmt.Errorf("case name is required")
	}

	var caseInfo CaseSummary

	query := `
		SELECT
			c.name,
			COUNT(v.version) as version_count,
			c.status,
			c.last_updated
		FROM kyc_cases c
		LEFT JOIN kyc_case_versions v ON c.name = v.case_name
		WHERE c.name = $1
		GROUP BY c.name, c.status, c.last_updated
	`

	err := db.Get(&caseInfo, query, caseName)
	if err != nil {
		return nil, fmt.Errorf("failed to get case '%s': %w", caseName, err)
	}

	return &caseInfo, nil
}

// GetCaseVersionCount returns the number of versions for a case
func GetCaseVersionCount(db *sqlx.DB, caseName string) (int, error) {
	if db == nil {
		return 0, fmt.Errorf("database connection is nil")
	}
	if caseName == "" {
		return 0, fmt.Errorf("case name is required")
	}

	var count int
	query := `SELECT COUNT(*) FROM kyc_case_versions WHERE case_name = $1`

	err := db.Get(&count, query, caseName)
	if err != nil {
		return 0, fmt.Errorf("failed to get version count for case '%s': %w", caseName, err)
	}

	return count, nil
}

// CaseExists checks if a case exists in the database
func CaseExists(db *sqlx.DB, caseName string) (bool, error) {
	if db == nil {
		return false, fmt.Errorf("database connection is nil")
	}
	if caseName == "" {
		return false, fmt.Errorf("case name is required")
	}

	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM kyc_cases WHERE name = $1)`

	err := db.Get(&exists, query, caseName)
	if err != nil {
		return false, fmt.Errorf("failed to check if case '%s' exists: %w", caseName, err)
	}

	return exists, nil
}
