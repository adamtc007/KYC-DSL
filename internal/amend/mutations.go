package amend

import (
	"fmt"

	"github.com/adamtc007/KYC-DSL/internal/model"
	"github.com/adamtc007/KYC-DSL/internal/ontology"
)

// AddDocumentDiscovery performs ontology-aware document discovery
// This function queries the regulatory ontology database to automatically
// populate document requirements and data dictionary mappings based on
// jurisdiction and applicable regulations.
//
// This is the only Go-side mutation function still used. All other amendments
// (policy-discovery, document-solicitation, ownership-discovery, risk-assessment,
// approve, decline, review) are now handled by the Rust DSL service.
func AddDocumentDiscovery(c *model.KycCase, repo *ontology.Repository) error {
	fmt.Println("🔍 Performing document discovery based on jurisdiction and regulation...")

	// Pull documents from ontology DB based on regulation
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
