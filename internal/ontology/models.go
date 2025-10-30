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

// AttributeMetadata represents rich metadata for an attribute
type AttributeMetadata struct {
	ID                  int       `db:"id"`
	AttributeCode       string    `db:"attribute_code"`
	Synonyms            []string  `db:"synonyms"`
	DataType            string    `db:"data_type"`
	DomainValues        []string  `db:"domain_values"`
	ValidationPattern   string    `db:"validation_pattern"`
	RiskLevel           string    `db:"risk_level"`
	ExampleValues       []string  `db:"example_values"`
	RegulatoryCitations []string  `db:"regulatory_citations"`
	BusinessGlossaryURL string    `db:"business_glossary_url"`
	DataSensitivity     string    `db:"data_sensitivity"`
	RetentionPeriodDays int       `db:"retention_period_days"`
	CreatedAt           time.Time `db:"created_at"`
	UpdatedAt           time.Time `db:"updated_at"`
}

// AttributeCluster represents a logical grouping of related attributes
type AttributeCluster struct {
	ID             int       `db:"id"`
	ClusterCode    string    `db:"cluster_code"`
	ClusterName    string    `db:"cluster_name"`
	AttributeCodes []string  `db:"attribute_codes"`
	Description    string    `db:"description"`
	UseCase        string    `db:"use_case"`
	Priority       int       `db:"priority"`
	CreatedAt      time.Time `db:"created_at"`
	UpdatedAt      time.Time `db:"updated_at"`
}

// AttributeRelationship represents semantic relationships between attributes
type AttributeRelationship struct {
	ID                  int       `db:"id"`
	SourceAttributeCode string    `db:"source_attribute_code"`
	TargetAttributeCode string    `db:"target_attribute_code"`
	RelationshipType    string    `db:"relationship_type"`
	Strength            float64   `db:"strength"`
	Description         string    `db:"description"`
	CreatedAt           time.Time `db:"created_at"`
}

// AttributeProfile combines attribute with its metadata for complete view
type AttributeProfile struct {
	Code                string
	Name                string
	Domain              string
	Description         string
	RiskCategory        string
	IsPersonalData      bool
	AttributeClass      string
	Synonyms            []string
	DataType            string
	DomainValues        []string
	RiskLevel           string
	ExampleValues       []string
	RegulatoryCitations []string
	DataSensitivity     string
	RetentionPeriodDays int
}
