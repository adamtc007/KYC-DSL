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
