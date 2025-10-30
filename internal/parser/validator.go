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
		"BUILD-OWNERSHIP-TREE",
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

	// ------------------ Ownership & Control checks ------------------
	// Only validate ownership if BUILD-OWNERSHIP-TREE function is present
	hasOwnershipFunction := false
	for _, f := range c.Functions {
		if f.Action == "BUILD-OWNERSHIP-TREE" || f.Action == "VERIFY-OWNERSHIP" {
			hasOwnershipFunction = true
			break
		}
	}

	if hasOwnershipFunction {
		if err := validateOwnershipAndControl(c); err != nil {
			return err
		}
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

// validateOwnershipAndControl performs structural and numeric checks on ownership and control data.
func validateOwnershipAndControl(c *model.KycCase) error {
	if len(c.Ownership) == 0 {
		return fmt.Errorf("no ownership-structure defined")
	}

	var legalTotal, beneficialTotal float64
	var ownerCount, beneficialCount, controllerCount int
	seenOwners := make(map[string]bool)
	seenBeneficial := make(map[string]bool)
	seenControllers := make(map[string]bool)

	for _, o := range c.Ownership {
		if o.Owner != "" {
			ownerCount++
			if seenOwners[o.Owner] {
				return fmt.Errorf("duplicate owner entry for '%s'", o.Owner)
			}
			seenOwners[o.Owner] = true
			legalTotal += o.OwnershipPercent
		}
		if o.BeneficialOwner != "" {
			beneficialCount++
			if seenBeneficial[o.BeneficialOwner] {
				return fmt.Errorf("duplicate beneficial-owner entry for '%s'", o.BeneficialOwner)
			}
			seenBeneficial[o.BeneficialOwner] = true
			beneficialTotal += o.OwnershipPercent
		}
		if o.Controller != "" {
			controllerCount++
			if seenControllers[o.Controller] {
				return fmt.Errorf("duplicate controller entry for '%s'", o.Controller)
			}
			seenControllers[o.Controller] = true
		}
	}

	// Require at least one owner or controller
	if ownerCount == 0 && controllerCount == 0 {
		return fmt.Errorf("ownership-structure must include at least one owner or controller")
	}

	// Check legal ownership percentage (within tolerance)
	if ownerCount > 0 {
		if legalTotal < 99.5 || legalTotal > 100.5 {
			return fmt.Errorf("legal ownership percentages must sum to 100%% ± 0.5 (got %.2f%%)", legalTotal)
		}
	}

	// Beneficial ownership can be any amount (typically less than or equal to legal)
	// No strict validation on beneficial ownership totals

	// Require controller if not 100% owned by one entity
	if controllerCount == 0 && ownerCount > 1 {
		return fmt.Errorf("at least one controller must be specified when multiple owners exist")
	}

	return nil
}
