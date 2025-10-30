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
