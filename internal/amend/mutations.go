package amend

import (
	"github.com/adamtc007/KYC-DSL/internal/model"
)

// Predefined amendments for each lifecycle phase

// AddPolicyDiscovery adds policy discovery function and injects policies
func AddPolicyDiscovery(c *model.KycCase) {
	c.Functions = append(c.Functions, model.Function{
		Action: "DISCOVER-POLICIES",
		Status: model.Pending,
	})
	c.Policies = append(c.Policies,
		model.KycPolicy{Code: "KYCPOL-UK-2025"},
		model.KycPolicy{Code: "KYCPOL-EU-2025"},
		model.KycPolicy{Code: "AML-GLOBAL-BASE"},
	)
}

// AddDocumentSolicitation adds document solicitation function and obligations
func AddDocumentSolicitation(c *model.KycCase) {
	c.Functions = append(c.Functions, model.Function{
		Action: "SOLICIT-DOCUMENTS",
		Status: model.Pending,
	})
	c.Obligations = append(c.Obligations,
		model.KycObligation{PolicyCode: "OBL-W8BEN"},
		model.KycObligation{PolicyCode: "OBL-W9"},
		model.KycObligation{PolicyCode: "OBL-UBO-DECLARATION"},
		model.KycObligation{PolicyCode: "OBL-PEP-001"},
	)
}

// AddOwnershipStructure adds ownership tree building function and ownership data
func AddOwnershipStructure(c *model.KycCase) {
	c.Functions = append(c.Functions, model.Function{
		Action: "BUILD-OWNERSHIP-TREE",
		Status: model.Pending,
	})
	c.Functions = append(c.Functions, model.Function{
		Action: "VERIFY-OWNERSHIP",
		Status: model.Pending,
	})

	// Initialize ownership structure if not present
	if c.Ownership == nil {
		c.Ownership = &model.OwnershipStructure{
			Entity: c.Name,
		}
	}

	// Add legal owners
	c.Ownership.LegalOwners = append(c.Ownership.LegalOwners,
		model.Owner{
			Name:       "BLACKROCK-PLC",
			Percentage: 100.0,
		},
	)

	// Add beneficial owners
	c.Ownership.BeneficialOwners = append(c.Ownership.BeneficialOwners,
		model.BeneficialOwner{
			Name:       "LARRY-FINK",
			Percentage: 35.0,
			Interest:   "voting rights",
		},
	)

	// Add controllers
	c.Ownership.Controllers = append(c.Ownership.Controllers,
		model.Controller{
			Name: "JANE-DOE",
			Role: "Senior Managing Official",
		},
		model.Controller{
			Name: "JOHN-SMITH",
			Role: "Director",
		},
	)

	// Add operational roles
	c.Ownership.OperationalRoles = append(c.Ownership.OperationalRoles,
		model.OperationalRole{
			Name:     "MARY-JONES",
			Title:    "Chief Compliance Officer",
			Function: "compliance oversight",
		},
	)
}

// AddRiskAssessment adds risk assessment function
func AddRiskAssessment(c *model.KycCase) {
	c.Functions = append(c.Functions, model.Function{
		Action: "ASSESS-RISK",
		Status: model.Pending,
	})
}

// AddRegulatorNotification adds regulator notification function
func AddRegulatorNotification(c *model.KycCase) {
	c.Functions = append(c.Functions, model.Function{
		Action: "REGULATOR-NOTIFY",
		Status: model.Pending,
	})
}

// FinalizeCase updates the token status to approved
func FinalizeCase(status string) func(*model.KycCase) {
	return func(c *model.KycCase) {
		if c.Token == nil {
			c.Token = &model.KycToken{}
		}
		c.Token.Status = status
		c.Status = model.Complete
	}
}

// ApproveCase finalizes the case with approved status
func ApproveCase(c *model.KycCase) {
	FinalizeCase("approved")(c)
}

// DeclineCase finalizes the case with declined status
func DeclineCase(c *model.KycCase) {
	FinalizeCase("declined")(c)
}

// RequestReviewCase sets the case to review status
func RequestReviewCase(c *model.KycCase) {
	FinalizeCase("review")(c)
}
