package amend

import (
	"fmt"
	"strings"

	"github.com/adamtc007/KYC-DSL/internal/model"
	"github.com/adamtc007/KYC-DSL/internal/parser"
	"github.com/adamtc007/KYC-DSL/internal/storage"
	"github.com/jmoiron/sqlx"
)

// ApplyAmendment loads the latest case version, applies a mutation, and saves the new version.
// Flow:
//  1. Load latest serialized DSL from database
//  2. Parse + bind to model
//  3. Execute mutationFn to apply the change
//  4. Serialize → validate → save as next version → log amendment
func ApplyAmendment(db *sqlx.DB, caseName string, step string, mutationFn func(*model.KycCase)) error {
	// Step 1: Load latest version
	latestVersion, err := getLatestVersion(db, caseName)
	if err != nil {
		return fmt.Errorf("failed to load latest version: %w", err)
	}

	// Step 2: Parse and bind
	parsedDSL, err := parser.Parse(strings.NewReader(latestVersion.DslSnapshot))
	if err != nil {
		return fmt.Errorf("failed to parse DSL: %w", err)
	}

	cases, err := parser.Bind(parsedDSL)
	if err != nil {
		return fmt.Errorf("failed to bind DSL: %w", err)
	}

	if len(cases) == 0 {
		return fmt.Errorf("no cases found in DSL")
	}

	kycCase := cases[0]

	// Step 3: Apply mutation
	oldSnapshot := parser.SerializeCases([]*model.KycCase{kycCase})
	mutationFn(kycCase)
	newSnapshot := parser.SerializeCases([]*model.KycCase{kycCase})

	// Step 4: Validate
	if err := parser.ValidateDSL(db, []*model.KycCase{kycCase}, ""); err != nil {
		return fmt.Errorf("validation failed after amendment: %w", err)
	}

	// Step 5: Calculate diff
	diff := generateSimpleDiff(oldSnapshot, newSnapshot)

	// Step 6: Save new version
	if err := storage.SaveCaseVersion(db, caseName, newSnapshot); err != nil {
		return fmt.Errorf("failed to save new version: %w", err)
	}

	// Step 7: Log amendment
	changeType := detectChangeType(kycCase, step)
	if err := storage.InsertAmendment(db, caseName, step, changeType, diff); err != nil {
		return fmt.Errorf("failed to log amendment: %w", err)
	}

	fmt.Printf("✅ Amendment applied: %s → %s\n", caseName, step)
	return nil
}

// getLatestVersion retrieves the most recent version of a case from the database.
func getLatestVersion(db *sqlx.DB, caseName string) (*storage.CaseVersion, error) {
	var version storage.CaseVersion
	query := `SELECT case_name, version, dsl_snapshot, hash, created_at
	          FROM kyc_case_versions
	          WHERE case_name=$1
	          ORDER BY version DESC
	          LIMIT 1`
	err := db.Get(&version, query, caseName)
	if err != nil {
		return nil, fmt.Errorf("case not found: %w", err)
	}
	return &version, nil
}

// generateSimpleDiff creates a basic diff between old and new DSL snapshots.
func generateSimpleDiff(old, new string) string {
	if old == new {
		return "No changes"
	}

	oldLines := strings.Split(old, "\n")
	newLines := strings.Split(new, "\n")

	var diff strings.Builder
	maxLen := len(oldLines)
	if len(newLines) > maxLen {
		maxLen = len(newLines)
	}

	changes := 0
	for i := 0; i < maxLen; i++ {
		var oldLine, newLine string
		if i < len(oldLines) {
			oldLine = strings.TrimSpace(oldLines[i])
		}
		if i < len(newLines) {
			newLine = strings.TrimSpace(newLines[i])
		}

		if oldLine != newLine {
			if oldLine != "" && newLine == "" {
				diff.WriteString(fmt.Sprintf("- %s\n", oldLine))
				changes++
			} else if oldLine == "" && newLine != "" {
				diff.WriteString(fmt.Sprintf("+ %s\n", newLine))
				changes++
			} else if oldLine != newLine {
				diff.WriteString(fmt.Sprintf("- %s\n+ %s\n", oldLine, newLine))
				changes++
			}
		}
	}

	if changes == 0 {
		return "Structural changes only"
	}
	return diff.String()
}

// detectChangeType determines the type of change based on the step and case state.
func detectChangeType(kycCase *model.KycCase, step string) string {
	switch step {
	case "CASE-CREATION":
		return "initialization"
	case "POLICY-DISCOVERY":
		return "policy-injection"
	case "DOCUMENT-SOLICITATION":
		return "obligation-addition"
	case "OWNERSHIP-CONTROL":
		return "ownership-tree"
	case "RISK-REVIEW":
		return "risk-assessment"
	case "FINALIZATION":
		if kycCase.Token != nil {
			return fmt.Sprintf("token-update:%s", kycCase.Token.Status)
		}
		return "finalization"
	default:
		return "generic-amendment"
	}
}
