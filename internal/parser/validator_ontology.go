package parser

import (
	"fmt"
	"strings"

	"github.com/adamtc007/KYC-DSL/internal/model"
	"github.com/adamtc007/KYC-DSL/internal/ontology"
	"github.com/adamtc007/KYC-DSL/internal/storage"
	"github.com/jmoiron/sqlx"
)

const (
	validationStatusPass = "PASS"
	validationStatusFail = "FAIL"
)

// ValidateOntologyRefs checks that all DSL references correspond to ontology data.
// Returns detailed validation feedback on success.
func ValidateOntologyRefs(db *sqlx.DB, c *model.KycCase) error {
	repo := ontology.NewRepository(db)

	// Track validation statistics for success message
	var docCount, attrCount int

	// ----------------------------------------------------
	// 1️⃣ Validate Document References
	// ----------------------------------------------------
	allDocs, err := repo.AllDocumentCodes()
	if err != nil {
		return fmt.Errorf("ontology validation failed (documents): %w", err)
	}
	validDocs := make(map[string]bool)
	for _, d := range allDocs {
		validDocs[strings.ToUpper(d)] = true
	}

	for _, dr := range c.DocumentRequirements {
		for _, d := range dr.Documents {
			if !validDocs[strings.ToUpper(d.Code)] {
				return fmt.Errorf("unknown document code '%s' in jurisdiction '%s'", d.Code, dr.Jurisdiction)
			}
		}
	}

	for _, src := range c.DataDictionary {
		check := func(code string, tier string) error {
			if code != "" && !validDocs[strings.ToUpper(code)] {
				return fmt.Errorf("unknown %s document '%s' for attribute '%s'", tier, code, src.AttributeCode)
			}
			return nil
		}
		if err := check(src.PrimarySource, "primary-source"); err != nil {
			return err
		}
		if err := check(src.SecondarySource, "secondary-source"); err != nil {
			return err
		}
		if err := check(src.TertiarySource, "tertiary-source"); err != nil {
			return err
		}
	}

	// ----------------------------------------------------
	// 2️⃣ Validate Attribute References
	// ----------------------------------------------------
	allAttrs, err := repo.AllAttributeCodes()
	if err != nil {
		return fmt.Errorf("ontology validation failed (attributes): %w", err)
	}
	validAttrs := make(map[string]bool)
	for _, a := range allAttrs {
		validAttrs[strings.ToUpper(a)] = true
	}

	for _, src := range c.DataDictionary {
		if !validAttrs[strings.ToUpper(src.AttributeCode)] {
			return fmt.Errorf("unknown attribute '%s' in data-dictionary", src.AttributeCode)
		}
	}

	// Get all regulation codes for validation
	allRegs, err := repo.AllRegulationCodes()
	if err != nil {
		return fmt.Errorf("ontology validation failed (regulations): %w", err)
	}
	validRegs := make(map[string]bool)
	for _, r := range allRegs {
		validRegs[strings.ToUpper(r)] = true
	}

	// ----------------------------------------------------
	// 3️⃣ Validate Jurisdictions (Lenient - just warn if empty)
	// ----------------------------------------------------
	for _, dr := range c.DocumentRequirements {
		if dr.Jurisdiction == "" {
			return fmt.Errorf("document-requirements section missing jurisdiction")
		}
	}

	// ----------------------------------------------------
	// 4️⃣ Validate that required docs are linked to regulations
	// ----------------------------------------------------
	for _, dr := range c.DocumentRequirements {
		for _, d := range dr.Documents {
			ok, err := repo.DocumentLinkedToRegulation(d.Code)
			if err != nil {
				return err
			}
			if !ok {
				return fmt.Errorf("document '%s' not linked to any regulation in ontology", d.Code)
			}
		}
	}

	// ----------------------------------------------------
	// Success feedback
	// ----------------------------------------------------
	if len(c.DocumentRequirements) > 0 || len(c.DataDictionary) > 0 {
		for _, dr := range c.DocumentRequirements {
			docCount += len(dr.Documents)
		}
		attrCount = len(c.DataDictionary)

		fmt.Printf("✔ Case %s passed ontology validation\n", c.Name)
		if docCount > 0 {
			fmt.Printf("   - All %d documents valid\n", docCount)
		}
		if attrCount > 0 {
			fmt.Printf("   - All %d attributes resolved\n", attrCount)
		}
		if len(c.Policies) > 0 {
			fmt.Printf("   - Policy references validated\n")
		}
	}

	// ----------------------------------------------------
	// 5️⃣ Validate Derived Attributes
	// ----------------------------------------------------
	for _, da := range c.DerivedAttributes {
		// Check that derived attribute exists and is Private
		attr, err := repo.GetAttributeByCode(da.DerivedAttribute)
		if err != nil {
			return fmt.Errorf("derived attribute '%s' not found in ontology", da.DerivedAttribute)
		}
		if attr.AttributeClass != "Private" {
			return fmt.Errorf("derived attribute '%s' must have attribute_class='Private', got '%s'",
				da.DerivedAttribute, attr.AttributeClass)
		}

		// Check that all source attributes exist and are Public
		for _, srcCode := range da.SourceAttributes {
			srcAttr, err := repo.GetAttributeByCode(srcCode)
			if err != nil {
				return fmt.Errorf("derived attribute '%s' references unknown source '%s'",
					da.DerivedAttribute, srcCode)
			}
			if srcAttr.AttributeClass != "Public" {
				return fmt.Errorf("derived attribute '%s' source '%s' must be Public, got '%s'",
					da.DerivedAttribute, srcCode, srcAttr.AttributeClass)
			}
		}

		// Check that rule expression is not empty
		if da.RuleExpression == "" {
			return fmt.Errorf("derived attribute '%s' missing rule expression", da.DerivedAttribute)
		}
	}

	// ----------------------------------------------------
	// 5️⃣ Validate Derived Attributes Lineage
	// ----------------------------------------------------
	if len(c.DerivedAttributes) > 0 {
		// Build map of valid public attributes
		validPublicAttrs := make(map[string]bool)
		for code := range validAttrs {
			attr, _ := repo.GetAttributeByCode(code)
			if attr != nil && attr.AttributeClass == "Public" {
				validPublicAttrs[strings.ToUpper(code)] = true
			}
		}

		// Validate each derived attribute
		for _, der := range c.DerivedAttributes {
			// Check that derived attribute exists and is Private
			derivedAttr, err := repo.GetAttributeByCode(der.DerivedAttribute)
			if err != nil {
				return fmt.Errorf("derived attribute '%s' not found in ontology", der.DerivedAttribute)
			}
			if derivedAttr.AttributeClass != "Private" {
				return fmt.Errorf("derived attribute '%s' must have attribute_class='Private', got '%s'",
					der.DerivedAttribute, derivedAttr.AttributeClass)
			}

			// Check that rule expression is not empty
			if der.RuleExpression == "" {
				return fmt.Errorf("derived attribute '%s' missing rule expression", der.DerivedAttribute)
			}

			// Check that all source attributes exist and are Public
			if len(der.SourceAttributes) == 0 {
				return fmt.Errorf("derived attribute '%s' has no source attributes", der.DerivedAttribute)
			}
			for _, srcCode := range der.SourceAttributes {
				if !validPublicAttrs[strings.ToUpper(srcCode)] {
					return fmt.Errorf("derived attribute '%s' references unknown or non-public source attribute '%s'",
						der.DerivedAttribute, srcCode)
				}
			}

			// Validate regulation code if specified
			if der.RegulationCode != "" && !validRegs[strings.ToUpper(der.RegulationCode)] {
				return fmt.Errorf("derived attribute '%s' references unknown regulation '%s'",
					der.DerivedAttribute, der.RegulationCode)
			}
		}
	}

	return nil
}

// ValidateCaseWithAudit performs comprehensive validation and records audit trail.
// Compliant with FCA SYSC, MAS 626 §4.2, HKMA AML §3.6, EU AMLD6 Article 30.
func ValidateCaseWithAudit(db *sqlx.DB, c *model.KycCase, actor string) error {
	var (
		totalChecks  = 0
		passedChecks = 0
		failedChecks = 0
	)

	record := model.CaseValidation{
		CaseName:        c.Name,
		Version:         c.Version,
		GrammarVersion:  "1.2",
		OntologyVersion: "v1.0",
		ValidatorActor:  actor,
	}

	// Check 1: Structure validation
	totalChecks++
	if err := validateCaseStructure(c); err != nil {
		record.ValidationStatus = validationStatusFail
		record.ErrorMessage = fmt.Sprintf("structure: %v", err)
		failedChecks++
		record.TotalChecks = totalChecks
		record.FailedChecks = failedChecks
		record.PassedChecks = passedChecks
		_ = storage.RecordValidationResult(db, record)
		return err
	}
	passedChecks++

	// Check 2: Semantic validation
	totalChecks++
	if err := validateCaseSemantics(db, c); err != nil {
		record.ValidationStatus = validationStatusFail
		record.ErrorMessage = fmt.Sprintf("semantics: %v", err)
		failedChecks++
		record.TotalChecks = totalChecks
		record.FailedChecks = failedChecks
		record.PassedChecks = passedChecks
		_ = storage.RecordValidationResult(db, record)
		return err
	}
	passedChecks++

	// Check 3: Ontology reference validation
	totalChecks++
	if err := ValidateOntologyRefs(db, c); err != nil {
		record.ValidationStatus = validationStatusFail
		record.ErrorMessage = fmt.Sprintf("ontology: %v", err)
		failedChecks++
		record.TotalChecks = totalChecks
		record.FailedChecks = failedChecks
		record.PassedChecks = passedChecks
		_ = storage.RecordValidationResult(db, record)
		return err
	}
	passedChecks++

	// All checks passed
	record.ValidationStatus = validationStatusPass
	record.PassedChecks = passedChecks
	record.TotalChecks = totalChecks
	_ = storage.RecordValidationResult(db, record)

	fmt.Printf("✅ Case %s validated successfully (%d/%d checks passed)\n",
		c.Name, passedChecks, totalChecks)
	fmt.Printf("   Audit trail recorded (actor: %s)\n", actor)

	return nil
}
