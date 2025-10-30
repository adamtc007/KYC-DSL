package amend

import (
	"fmt"

	"github.com/adamtc007/KYC-DSL/internal/model"
)

// LifecyclePhase represents a stage in the KYC case lifecycle
type LifecyclePhase string

const (
	PhaseCreation         LifecyclePhase = "CASE-CREATION"
	PhasePolicyDiscovery  LifecyclePhase = "POLICY-DISCOVERY"
	PhaseDocSolicitation  LifecyclePhase = "DOCUMENT-SOLICITATION"
	PhaseOwnershipControl LifecyclePhase = "OWNERSHIP-CONTROL"
	PhaseRiskReview       LifecyclePhase = "RISK-REVIEW"
	PhaseFinalization     LifecyclePhase = "FINALIZATION"
)

// PhaseDefinition describes what happens in each lifecycle phase
type PhaseDefinition struct {
	Phase       LifecyclePhase
	Description string
	Additions   []string
	Functions   []string
	NextPhases  []LifecyclePhase
}

// LifecycleDefinitions maps each phase to its definition
var LifecycleDefinitions = map[LifecyclePhase]PhaseDefinition{
	PhaseCreation: {
		Phase:       PhaseCreation,
		Description: "Initialize the case with nature, purpose, and client business unit",
		Additions: []string{
			"kyc-case",
			"nature-purpose",
			"client-business-unit",
		},
		Functions: []string{},
		NextPhases: []LifecyclePhase{
			PhasePolicyDiscovery,
		},
	},
	PhasePolicyDiscovery: {
		Phase:       PhasePolicyDiscovery,
		Description: "Discover which policies apply based on product & jurisdiction",
		Additions: []string{
			"policy nodes (auto-injected)",
		},
		Functions: []string{
			"DISCOVER-POLICIES",
		},
		NextPhases: []LifecyclePhase{
			PhaseDocSolicitation,
		},
	},
	PhaseDocSolicitation: {
		Phase:       PhaseDocSolicitation,
		Description: "Request proofs (W8/W9, UBO declarations, etc.)",
		Additions: []string{
			"obligation nodes",
		},
		Functions: []string{
			"SOLICIT-DOCUMENTS",
		},
		NextPhases: []LifecyclePhase{
			PhaseOwnershipControl,
		},
	},
	PhaseOwnershipControl: {
		Phase:       PhaseOwnershipControl,
		Description: "Build legal & beneficial ownership graph + operational control roles",
		Additions: []string{
			"ownership-structure",
			"legal owners",
			"beneficial owners",
			"controllers",
			"operational roles",
		},
		Functions: []string{
			"BUILD-OWNERSHIP-TREE",
			"VERIFY-OWNERSHIP",
		},
		NextPhases: []LifecyclePhase{
			PhaseRiskReview,
		},
	},
	PhaseRiskReview: {
		Phase:       PhaseRiskReview,
		Description: "Compute KYC score; escalate or approve",
		Additions: []string{
			"risk assessment results",
		},
		Functions: []string{
			"ASSESS-RISK",
			"REGULATOR-NOTIFY",
		},
		NextPhases: []LifecyclePhase{
			PhaseFinalization,
			PhaseDocSolicitation, // Can loop back for additional docs
		},
	},
	PhaseFinalization: {
		Phase:       PhaseFinalization,
		Description: "Token issuance / completion",
		Additions: []string{
			"kyc-token status update",
		},
		Functions:  []string{},
		NextPhases: []LifecyclePhase{
			// Terminal phase - can only reopen if needed
		},
	},
}

// ValidateTransition checks if a phase transition is allowed
func ValidateTransition(currentPhase, nextPhase LifecyclePhase) error {
	def, exists := LifecycleDefinitions[currentPhase]
	if !exists {
		return fmt.Errorf("unknown current phase: %s", currentPhase)
	}

	// Check if next phase is allowed
	for _, allowed := range def.NextPhases {
		if allowed == nextPhase {
			return nil
		}
	}

	return fmt.Errorf("invalid transition from %s to %s", currentPhase, nextPhase)
}

// GetCurrentPhase determines the current phase of a KYC case based on its state
func GetCurrentPhase(kycCase *model.KycCase) LifecyclePhase {
	// Check for finalization (token exists and is not pending)
	if kycCase.Token != nil && kycCase.Token.Status != "pending" {
		return PhaseFinalization
	}

	// Check for risk review functions
	hasRiskAssessment := false
	hasOwnershipTree := false
	hasDocSolicitation := false
	hasPolicyDiscovery := false

	for _, fn := range kycCase.Functions {
		switch fn.Action {
		case "ASSESS-RISK", "REGULATOR-NOTIFY":
			hasRiskAssessment = true
		case "BUILD-OWNERSHIP-TREE", "VERIFY-OWNERSHIP":
			hasOwnershipTree = true
		case "SOLICIT-DOCUMENTS":
			hasDocSolicitation = true
		case "DISCOVER-POLICIES":
			hasPolicyDiscovery = true
		}
	}

	// Determine phase based on what's been done
	if hasRiskAssessment {
		return PhaseRiskReview
	}
	if hasOwnershipTree || kycCase.Ownership != nil {
		return PhaseOwnershipControl
	}
	if hasDocSolicitation || len(kycCase.Obligations) > 0 {
		return PhaseDocSolicitation
	}
	if hasPolicyDiscovery || len(kycCase.Policies) > 0 {
		return PhasePolicyDiscovery
	}

	// Default to creation phase
	return PhaseCreation
}

// GetNextAllowedPhases returns the list of phases that can follow the current phase
func GetNextAllowedPhases(currentPhase LifecyclePhase) []LifecyclePhase {
	def, exists := LifecycleDefinitions[currentPhase]
	if !exists {
		return []LifecyclePhase{}
	}
	return def.NextPhases
}

// PhaseMetadata returns human-readable information about a phase
func PhaseMetadata(phase LifecyclePhase) (description string, additions []string, functions []string) {
	def, exists := LifecycleDefinitions[phase]
	if !exists {
		return "Unknown phase", []string{}, []string{}
	}
	return def.Description, def.Additions, def.Functions
}

// IsTerminalPhase returns true if the phase is a final state
func IsTerminalPhase(phase LifecyclePhase) bool {
	return phase == PhaseFinalization
}

// CanAddFunction checks if a function is allowed in the current phase
func CanAddFunction(phase LifecyclePhase, functionName string) bool {
	def, exists := LifecycleDefinitions[phase]
	if !exists {
		return false
	}

	for _, allowed := range def.Functions {
		if allowed == functionName {
			return true
		}
	}

	return false
}
