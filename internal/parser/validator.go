package parser

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/adamtc007/KYC-DSL/internal/model"
	"github.com/jmoiron/sqlx"
)

// ValidateDSL checks syntactic and semantic correctness of parsed DSL.
func ValidateDSL(db *sqlx.DB, cases []*model.KycCase, ebnf string) error {
	for _, c := range cases {
		if err := validateCaseStructure(c); err != nil {
			return fmt.Errorf("case %s structural error: %w", c.Name, err)
		}
		if err := validateCaseSemantics(db, c); err != nil {
			return fmt.Errorf("case %s semantic error: %w", c.Name, err)
		}
	}
	return nil
}

// ------------------ Structure checks ------------------

func validateCaseStructure(c *model.KycCase) error {
	if c.Nature == "" || c.Purpose == "" {
		return fmt.Errorf("missing nature or purpose section")
	}
	if c.CBU.Name == "" {
		return fmt.Errorf("missing client-business-unit section")
	}
	if c.Token == nil {
		return fmt.Errorf("missing kyc-token section")
	}
	return nil
}

// ------------------ Semantic checks ------------------

func validateCaseSemantics(db *sqlx.DB, c *model.KycCase) error {
	validFunctions := []string{
		"DISCOVER-POLICIES",
		"SOLICIT-DOCUMENTS",
		"EXTRACT-DATA",
		"VERIFY-OWNERSHIP",
		"ASSESS-RISK",
		"REGULATOR-NOTIFY",
	}

	// Validate function names
	for _, f := range c.Functions {
		if !contains(validFunctions, f.Action) {
			return fmt.Errorf("unknown function '%s'", f.Action)
		}
	}

	// Validate policy codes exist in DB
	for _, p := range c.Policies {
		var count int
		err := db.Get(&count, "SELECT COUNT(*) FROM kyc_policies WHERE code=$1", p.Code)
		if err != nil {
			return fmt.Errorf("policy lookup failed: %v", err)
		}
		if count == 0 {
			return fmt.Errorf("policy code '%s' not found in registry", p.Code)
		}
	}

	// Validate token state
	validToken := regexp.MustCompile(`^(pending|approved|declined|review)$`)
	if c.Token != nil && !validToken.MatchString(strings.ToLower(c.Token.Status)) {
		return fmt.Errorf("invalid token state '%s'", c.Token.Status)
	}

	return nil
}

// ------------------ Helper ------------------
func contains(list []string, val string) bool {
	for _, v := range list {
		if v == val {
			return true
		}
	}
	return false
}
