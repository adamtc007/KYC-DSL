package parser

import (
	"strconv"
	"strings"
	"time"

	"github.com/adamtc007/KYC-DSL/internal/model"
)

const (
	nodeTypeDocument = "document"
)

// Bind converts a parsed DSL AST into typed model structs.
func Bind(dsl *DSL) ([]*model.KycCase, error) {
	var cases []*model.KycCase
	for _, c := range dsl.Cases {
		caseObj := &model.KycCase{
			Name:        c.Name,
			Status:      model.Pending,
			LastUpdated: time.Now(),
		}

		// Walk the case body
		for _, node := range c.Body.Args {
			switch node.Head {
			case "nature-purpose":
				nature, purpose := parseNaturePurpose(node)
				caseObj.Nature = nature
				caseObj.Purpose = purpose

			case "client-business-unit":
				cbu := parseCBU(node)
				caseObj.CBU = cbu

			case "policy":
				caseObj.Policies = append(caseObj.Policies, model.KycPolicy{
					Code: trimQuotes(node.Args[0].Head),
				})

			case "function":
				if len(node.Args) > 0 {
					caseObj.Functions = append(caseObj.Functions, model.Function{
						Action: node.Args[0].Head,
						Status: model.Pending,
					})
				}

			case "obligation":
				caseObj.Obligations = append(caseObj.Obligations, model.KycObligation{
					PolicyCode: trimQuotes(node.Args[0].Head),
				})

			case "kyc-token":
				caseObj.Token = &model.KycToken{
					Status: trimQuotes(node.Args[0].Head),
				}

			case "ownership-structure":
				for _, n := range node.Args {
					switch n.Head {
					case "entity":
						if len(n.Args) > 0 {
							caseObj.Ownership = append(caseObj.Ownership, model.OwnershipNode{
								Entity: n.Args[0].Head,
							})
						}
					case "owner":
						if len(n.Args) >= 2 {
							caseObj.Ownership = append(caseObj.Ownership, model.OwnershipNode{
								Owner:            n.Args[0].Head,
								OwnershipPercent: parsePercent(n.Args[1].Head),
							})
						}
					case "beneficial-owner":
						if len(n.Args) >= 2 {
							caseObj.Ownership = append(caseObj.Ownership, model.OwnershipNode{
								BeneficialOwner:  n.Args[0].Head,
								OwnershipPercent: parsePercent(n.Args[1].Head),
							})
						}
					case "controller":
						if len(n.Args) >= 2 {
							caseObj.Ownership = append(caseObj.Ownership, model.OwnershipNode{
								Controller: n.Args[0].Head,
								Role:       trimQuotes(n.Args[1].Head),
							})
						}
					}
				}

			case "data-dictionary":
				for _, attrNode := range node.Args {
					if attrNode.Head != "attribute" {
						continue
					}
					if len(attrNode.Args) == 0 {
						continue
					}
					src := model.AttributeSource{AttributeCode: attrNode.Args[0].Head}
					for _, s := range attrNode.Args[1:] {
						switch s.Head {
						case "primary-source":
							if len(s.Args) > 0 {
								if s.Args[0].Head == nodeTypeDocument && len(s.Args[0].Args) > 0 {
									src.PrimarySource = s.Args[0].Args[0].Head
								} else {
									src.PrimarySource = trimQuotes(s.Args[0].Head)
								}
							}
						case "secondary-source":
							if len(s.Args) > 0 {
								if s.Args[0].Head == nodeTypeDocument && len(s.Args[0].Args) > 0 {
									src.SecondarySource = s.Args[0].Args[0].Head
								} else {
									src.SecondarySource = trimQuotes(s.Args[0].Head)
								}
							}
						case "tertiary-source":
							if len(s.Args) > 0 {
								if s.Args[0].Head == nodeTypeDocument && len(s.Args[0].Args) > 0 {
									src.TertiarySource = s.Args[0].Args[0].Head
								} else {
									src.TertiarySource = trimQuotes(s.Args[0].Head)
								}
							}
						}
					}
					caseObj.DataDictionary = append(caseObj.DataDictionary, src)
				}

			case "document-requirements":
				var dr model.DocumentRequirement
				for _, arg := range node.Args {
					switch arg.Head {
					case "jurisdiction":
						if len(arg.Args) > 0 {
							dr.Jurisdiction = arg.Args[0].Head
						}
					case "required":
						for _, doc := range arg.Args {
							if doc.Head == nodeTypeDocument && len(doc.Args) >= 2 {
								dr.Documents = append(dr.Documents, model.DocumentRef{
									Code: doc.Args[0].Head,
									Name: trimQuotes(doc.Args[1].Head),
								})
							}
						}
					}
				}
				caseObj.DocumentRequirements = append(caseObj.DocumentRequirements, dr)

			default:
				// Unknown node types ignored
			}
		}
		cases = append(cases, caseObj)
	}
	return cases, nil
}

func parseNaturePurpose(node *Expr) (nature, purpose string) {
	for _, arg := range node.Args {
		switch arg.Head {
		case "nature":
			if len(arg.Args) > 0 {
				nature = trimQuotes(arg.Args[0].Head)
			}
		case "purpose":
			if len(arg.Args) > 0 {
				purpose = trimQuotes(arg.Args[0].Head)
			}
		}
	}
	return
}

func parseCBU(node *Expr) model.ClientBusinessUnit {
	cbu := model.ClientBusinessUnit{Name: "UNKNOWN"}
	if len(node.Args) > 0 {
		cbu.Name = node.Args[0].Head
	}
	return cbu
}

func trimQuotes(s string) string {
	return strings.Trim(s, "\"")
}

func parsePercent(s string) float64 {
	s = strings.TrimSuffix(s, "%")
	val, _ := strconv.ParseFloat(s, 64)
	return val
}
