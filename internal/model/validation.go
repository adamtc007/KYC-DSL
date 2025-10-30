package model

import "time"

// CaseValidation represents a validation audit record
type CaseValidation struct {
	ID               int       `db:"id"`
	CaseName         string    `db:"case_name"`
	Version          int       `db:"version"`
	ValidationTime   time.Time `db:"validation_time"`
	GrammarVersion   string    `db:"grammar_version"`
	OntologyVersion  string    `db:"ontology_version"`
	ValidatorActor   string    `db:"validator_actor"`
	ValidationStatus string    `db:"validation_status"`
	ErrorMessage     string    `db:"error_message"`
	TotalChecks      int       `db:"total_checks"`
	PassedChecks     int       `db:"passed_checks"`
	FailedChecks     int       `db:"failed_checks"`
	CreatedAt        time.Time `db:"created_at"`
}

// ValidationFinding represents a detailed finding from validation
type ValidationFinding struct {
	ID           int       `db:"id"`
	ValidationID int       `db:"validation_id"`
	CheckType    string    `db:"check_type"`
	CheckName    string    `db:"check_name"`
	CheckStatus  string    `db:"check_status"`
	CheckMessage string    `db:"check_message"`
	EntityRef    string    `db:"entity_ref"`
	Severity     string    `db:"severity"`
	CreatedAt    time.Time `db:"created_at"`
}
