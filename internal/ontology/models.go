package ontology

import "time"

// Regulation represents a regulatory framework or law
type Regulation struct {
	ID            int        `db:"id"`
	Code          string     `db:"code"`
	Name          string     `db:"name"`
	Jurisdiction  string     `db:"jurisdiction"`
	Authority     string     `db:"authority"`
	Description   string     `db:"description"`
	EffectiveFrom time.Time  `db:"effective_from"`
	EffectiveTo   *time.Time `db:"effective_to"`
	CreatedAt     time.Time  `db:"created_at"`
}

// Document represents an evidence type that proves compliance attributes
type Document struct {
	ID             int       `db:"id"`
	Code           string    `db:"code"`
	Name           string    `db:"name"`
	Domain         string    `db:"domain"`
	Jurisdiction   string    `db:"jurisdiction"`
	RegulationCode string    `db:"regulation_code"`
	SourceType     string    `db:"source_type"`
	ValidityYears  int       `db:"validity_years"`
	Description    string    `db:"description"`
	CreatedAt      time.Time `db:"created_at"`
}

// Attribute represents a data point required for compliance
type Attribute struct {
	ID             int       `db:"id"`
	Code           string    `db:"code"`
	Name           string    `db:"name"`
	Domain         string    `db:"domain"`
	Description    string    `db:"description"`
	RiskCategory   string    `db:"risk_category"`
	IsPersonal     bool      `db:"is_personal_data"`
	AttributeClass string    `db:"attribute_class"` // Public or Private
	CreatedAt      time.Time `db:"created_at"`
}

// AttributeDocumentLink links attributes to documents that can evidence them
type AttributeDocumentLink struct {
	ID             int    `db:"id"`
	AttributeCode  string `db:"attribute_code"`
	DocumentCode   string `db:"document_code"`
	SourceTier     string `db:"source_tier"`
	IsMandatory    bool   `db:"is_mandatory"`
	Jurisdiction   string `db:"jurisdiction"`
	RegulationCode string `db:"regulation_code"`
	Notes          string `db:"notes"`
}

// DocumentRegulationLink links documents to regulations that require them
type DocumentRegulationLink struct {
	ID             int    `db:"id"`
	DocumentCode   string `db:"document_code"`
	RegulationCode string `db:"regulation_code"`
	Applicability  string `db:"applicability"`
	Jurisdiction   string `db:"jurisdiction"`
}

// AttributeDerivation represents how a private attribute is derived from public attributes
type AttributeDerivation struct {
	ID                   int       `db:"id"`
	DerivedAttributeCode string    `db:"derived_attribute_code"`
	SourceAttributeCode  string    `db:"source_attribute_code"`
	RuleExpression       string    `db:"rule_expression"`
	RuleType             string    `db:"rule_type"` // Boolean, Numeric, String, Lookup
	Description          string    `db:"description"`
	CreatedAt            time.Time `db:"created_at"`
}
