package amend

import (
	"fmt"

	"github.com/adamtc007/KYC-DSL/internal/model"
	"github.com/adamtc007/KYC-DSL/internal/ontology"
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

	// Add ownership nodes
	c.Ownership = append(c.Ownership,
		model.OwnershipNode{
			Entity: "BLACKROCK-GLOBAL-FUNDS",
		},
		model.OwnershipNode{
			Owner:            "BLACKROCK-PLC",
			OwnershipPercent: 100,
		},
		model.OwnershipNode{
			BeneficialOwner:  "LARRY-FINK",
			OwnershipPercent: 35,
		},
		model.OwnershipNode{
			Controller: "JANE-DOE",
			Role:       "Senior Managing Official",
		},
		model.OwnershipNode{
			Controller: "JOHN-SMITH",
			Role:       "Director",
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

// AddDocumentDiscovery performs ontology-aware document discovery
// This function queries the regulatory ontology database to automatically
// populate document requirements and data dictionary mappings based on
// jurisdiction and applicable regulations.
func AddDocumentDiscovery(c *model.KycCase, repo *ontology.Repository) error {
	fmt.Println("🔍 Performing document discovery based on jurisdiction and regulation...")

	// Example: pull documents from ontology DB based on regulation
	docs, err := repo.ListDocumentsByRegulation("AMLD5")
	if err != nil {
		return fmt.Errorf("failed to retrieve documents from ontology: %w", err)
	}

	// Create document requirement for EU jurisdiction
	dr := model.DocumentRequirement{Jurisdiction: "EU"}
	for _, d := range docs {
		dr.Documents = append(dr.Documents, model.DocumentRef{
			Code: d.Code,
			Name: d.Name,
		})
	}
	c.DocumentRequirements = append(c.DocumentRequirements, dr)

	// Add sample data dictionary links from ontology for key attributes
	// Query the ontology for attribute-document mappings
	attrCodes := []string{"UBO_NAME", "REGISTERED_NAME", "TAX_RESIDENCY_COUNTRY"}
	for _, attrCode := range attrCodes {
		attrDocs, err := repo.GetDocumentSources(attrCode)
		if err != nil {
			fmt.Printf("Warning: failed to get document sources for %s: %v\n", attrCode, err)
			continue
		}

		if len(attrDocs) > 0 {
			src := model.AttributeSource{
				AttributeCode: attrCode,
			}
			// Assign sources based on tier
			for _, link := range attrDocs {
				switch link.SourceTier {
				case "Primary":
					if src.PrimarySource == "" {
						src.PrimarySource = link.DocumentCode
					}
				case "Secondary":
					if src.SecondarySource == "" {
						src.SecondarySource = link.DocumentCode
					}
				case "Tertiary":
					if src.TertiarySource == "" {
						src.TertiarySource = link.DocumentCode
					}
				}
			}
			c.DataDictionary = append(c.DataDictionary, src)
		}
	}

	fmt.Printf("✅ Added %d document requirements and %d data dictionary entries\n",
		len(dr.Documents), len(c.DataDictionary))

	return nil
}
