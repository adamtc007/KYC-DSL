package amend

import (
	"fmt"
	"strings"

	pb "github.com/adamtc007/KYC-DSL/api/pb"
	"github.com/adamtc007/KYC-DSL/internal/model"
	"github.com/adamtc007/KYC-DSL/internal/rustclient"
	"github.com/adamtc007/KYC-DSL/internal/storage"
	"github.com/jmoiron/sqlx"
)

// ApplyAmendment loads the latest case version, applies a mutation, and saves the new version.
// For most amendments, this delegates to the Rust DSL service via gRPC.
// For ontology-aware amendments (like document-discovery), it uses local mutation functions.
//
// Flow:
//  1. Load latest serialized DSL from database
//  2. Apply mutation (via Rust or local function)
//  3. Validate the result
//  4. Save as next version
//  5. Log amendment
func ApplyAmendment(db *sqlx.DB, caseName string, step string, mutationFn func(*model.KycCase)) error {
	// Step 1: Load latest version
	latestVersion, err := getLatestVersion(db, caseName)
	if err != nil {
		return fmt.Errorf("failed to load latest version: %w", err)
	}

	oldSnapshot := latestVersion.DslSnapshot

	// Step 2: Apply mutation via local function (for ontology-aware steps)
	// This is called when we need direct DB access (e.g., document-discovery)
	if mutationFn != nil {
		// Parse via Rust to get structured case
		rustClient, err := rustclient.NewDslClient("")
		if err != nil {
			return fmt.Errorf("failed to connect to Rust DSL service: %w", err)
		}
		defer rustClient.Close()

		parseResp, err := rustClient.ParseDSL(oldSnapshot)
		if err != nil || !parseResp.Success {
			return fmt.Errorf("failed to parse DSL: %w", err)
		}

		if len(parseResp.Cases) == 0 {
			return fmt.Errorf("no cases found in DSL")
		}

		// Convert proto case to model (simplified - in real impl would need full conversion)
		kycCase := protoToModelCase(parseResp.Cases[0])

		// Apply local mutation
		mutationFn(kycCase)

		// Serialize back via Rust
		serializeResp, err := rustClient.SerializeCase(parseResp.Cases[0])
		if err != nil || !serializeResp.Success {
			return fmt.Errorf("failed to serialize case: %w", err)
		}

		newSnapshot := serializeResp.Dsl

		// Validate
		valResult, err := rustClient.ValidateDSL(newSnapshot)
		if err != nil || !valResult.Valid {
			return fmt.Errorf("validation failed after amendment: %v", valResult.Errors)
		}

		// Calculate diff
		diff := generateSimpleDiff(oldSnapshot, newSnapshot)

		// Save new version
		if err := storage.SaveCaseVersion(db, caseName, newSnapshot); err != nil {
			return fmt.Errorf("failed to save new version: %w", err)
		}

		// Log amendment
		changeType := detectChangeType(kycCase, step)
		if err := storage.InsertAmendment(db, caseName, step, changeType, diff); err != nil {
			return fmt.Errorf("failed to log amendment: %w", err)
		}

		fmt.Printf("✅ Amendment applied: %s → %s\n", caseName, step)
		return nil
	}

	// Step 3: For standard amendments, use Rust service directly
	rustClient, err := rustclient.NewDslClient("")
	if err != nil {
		return fmt.Errorf("failed to connect to Rust DSL service: %w", err)
	}
	defer rustClient.Close()

	amendResp, err := rustClient.AmendCase(caseName, step)
	if err != nil {
		return fmt.Errorf("amendment RPC failed: %w", err)
	}
	if !amendResp.Success {
		return fmt.Errorf("amendment failed: %s", amendResp.Message)
	}

	newSnapshot := amendResp.UpdatedDsl

	// Calculate diff
	diff := generateSimpleDiff(oldSnapshot, newSnapshot)

	// Save new version
	if err := storage.SaveCaseVersion(db, caseName, newSnapshot); err != nil {
		return fmt.Errorf("failed to save new version: %w", err)
	}

	// Log amendment
	changeType := step // Use step as change type for Rust-applied amendments
	if err := storage.InsertAmendment(db, caseName, step, changeType, diff); err != nil {
		return fmt.Errorf("failed to log amendment: %w", err)
	}

	fmt.Printf("✅ Amendment applied: %s → %s (via Rust service)\n", caseName, step)
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
	case "policy-discovery":
		return "policy-injection"
	case "document-solicitation":
		return "obligation-addition"
	case "document-discovery":
		return "document-discovery"
	case "ownership-discovery":
		return "ownership-tree"
	case "risk-assessment":
		return "risk-assessment"
	case "regulator-notify":
		return "regulator-notification"
	case "approve":
		if kycCase.Token != nil {
			return fmt.Sprintf("token-update:%s", kycCase.Token.Status)
		}
		return "finalization-approved"
	case "decline":
		return "finalization-declined"
	case "review":
		return "status-review"
	default:
		return "generic-amendment"
	}
}

// protoToModelCase converts a proto ParsedCase to internal KycCase model
// This is a simplified conversion - in production would need full field mapping
func protoToModelCase(protoCase *pb.ParsedCase) *model.KycCase {
	kycCase := &model.KycCase{
		Name:    protoCase.Name,
		Nature:  protoCase.Nature,
		Purpose: protoCase.Purpose,
	}

	// Add more field mappings as needed
	return kycCase
}
