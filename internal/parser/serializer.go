package parser

import (
	"fmt"
	"strings"
	"time"

	"github.com/adamtc007/KYC-DSL/internal/model"
)

// SerializeCases converts typed model structs back to S-expression DSL text.
func SerializeCases(cases []*model.KycCase) string {
	var sb strings.Builder
	for _, c := range cases {
		sb.WriteString(serializeCase(c))
		sb.WriteString("\n\n")
	}
	return sb.String()
}

func serializeCase(c *model.KycCase) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("(kyc-case %s\n", c.Name))

	// Nature and Purpose
	if c.Nature != "" || c.Purpose != "" {
		sb.WriteString("  (nature-purpose\n")
		if c.Nature != "" {
			sb.WriteString(fmt.Sprintf("    (nature \"%s\")\n", c.Nature))
		}
		if c.Purpose != "" {
			sb.WriteString(fmt.Sprintf("    (purpose \"%s\")\n", c.Purpose))
		}
		sb.WriteString("  )\n")
	}

	// CBU
	if c.CBU.Name != "" {
		sb.WriteString(fmt.Sprintf("  (client-business-unit %s)\n", c.CBU.Name))
	}

	// Policies
	for _, p := range c.Policies {
		sb.WriteString(fmt.Sprintf("  (policy %s)\n", p.Code))
	}

	// Obligations
	for _, o := range c.Obligations {
		sb.WriteString(fmt.Sprintf("  (obligation %s)\n", o.PolicyCode))
	}

	// Functions
	for _, f := range c.Functions {
		sb.WriteString(fmt.Sprintf("  (function %s)\n", f.Action))
	}

	// Ownership Structure
	if len(c.Ownership) > 0 {
		sb.WriteString("  (ownership-structure\n")
		for _, o := range c.Ownership {
			switch {
			case o.Entity != "":
				sb.WriteString(fmt.Sprintf("    (entity %s)\n", o.Entity))
			case o.Owner != "":
				sb.WriteString(fmt.Sprintf("    (owner %s %.0f%%)\n", o.Owner, o.OwnershipPercent))
			case o.BeneficialOwner != "":
				sb.WriteString(fmt.Sprintf("    (beneficial-owner %s %.0f%%)\n", o.BeneficialOwner, o.OwnershipPercent))
			case o.Controller != "":
				sb.WriteString(fmt.Sprintf("    (controller %s \"%s\")\n", o.Controller, o.Role))
			}
		}
		sb.WriteString("  )\n")
	}

	// Data Dictionary
	if len(c.DataDictionary) > 0 {
		sb.WriteString("  (data-dictionary\n")
		for _, src := range c.DataDictionary {
			sb.WriteString(fmt.Sprintf("    (attribute %s\n", src.AttributeCode))
			if src.PrimarySource != "" {
				sb.WriteString(fmt.Sprintf("      (primary-source (document %s))\n", src.PrimarySource))
			}
			if src.SecondarySource != "" {
				sb.WriteString(fmt.Sprintf("      (secondary-source (document %s))\n", src.SecondarySource))
			}
			if src.TertiarySource != "" {
				sb.WriteString(fmt.Sprintf("      (tertiary-source \"%s\")\n", src.TertiarySource))
			}
			sb.WriteString("    )\n")
		}
		sb.WriteString("  )\n")
	}

	// Document Requirements
	if len(c.DocumentRequirements) > 0 {
		for _, dr := range c.DocumentRequirements {
			sb.WriteString("  (document-requirements\n")
			sb.WriteString(fmt.Sprintf("    (jurisdiction %s)\n", dr.Jurisdiction))
			sb.WriteString("    (required\n")
			for _, d := range dr.Documents {
				sb.WriteString(fmt.Sprintf("      (document %s \"%s\")\n", d.Code, d.Name))
			}
			sb.WriteString("    )\n  )\n")
		}
	}

	// Derived Attributes
	if len(c.DerivedAttributes) > 0 {
		sb.WriteString("  (derived-attributes\n")
		for _, da := range c.DerivedAttributes {
			sb.WriteString(fmt.Sprintf("    (attribute %s\n", da.DerivedAttribute))
			if len(da.SourceAttributes) > 0 {
				sb.WriteString("      (sources (")
				for i, src := range da.SourceAttributes {
					if i > 0 {
						sb.WriteString(" ")
					}
					sb.WriteString(src)
				}
				sb.WriteString("))\n")
			}
			if da.RuleExpression != "" {
				sb.WriteString(fmt.Sprintf("      (rule \"%s\")\n", da.RuleExpression))
			}
			if da.Jurisdiction != "" {
				sb.WriteString(fmt.Sprintf("      (jurisdiction %s)\n", da.Jurisdiction))
			}
			if da.RegulationCode != "" {
				sb.WriteString(fmt.Sprintf("      (regulation %s)\n", da.RegulationCode))
			}
			sb.WriteString("    )\n")
		}
		sb.WriteString("  )\n")
	}

	// Token
	if c.Token != nil {
		sb.WriteString(fmt.Sprintf("  (kyc-token \"%s\")\n", c.Token.Status))
	}

	sb.WriteString(")\n")
	return sb.String()
}

// Optional: deterministic metadata header for version control
func SerializeCaseHeader(c *model.KycCase) string {
	return fmt.Sprintf("; Generated %s | Case %s v%d\n",
		time.Now().Format(time.RFC3339), c.Name, c.Version)
}
