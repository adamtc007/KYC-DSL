package ontology

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

type Repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) ListRegulations() ([]Regulation, error) {
	var regs []Regulation
	err := r.db.Select(&regs, "SELECT * FROM kyc_regulations ORDER BY code")
	return regs, err
}

func (r *Repository) GetRegulationByCode(code string) (*Regulation, error) {
	var reg Regulation
	err := r.db.Get(&reg, "SELECT * FROM kyc_regulations WHERE code=$1", code)
	if err != nil {
		return nil, err
	}
	return &reg, nil
}

func (r *Repository) ListDocuments() ([]Document, error) {
	var docs []Document
	err := r.db.Select(&docs, "SELECT * FROM kyc_documents ORDER BY code")
	return docs, err
}

func (r *Repository) ListDocumentsByRegulation(regCode string) ([]Document, error) {
	var docs []Document
	err := r.db.Select(&docs, "SELECT * FROM kyc_documents WHERE regulation_code=$1 ORDER BY code", regCode)
	return docs, err
}

func (r *Repository) GetDocumentByCode(code string) (*Document, error) {
	var doc Document
	err := r.db.Get(&doc, "SELECT * FROM kyc_documents WHERE code=$1", code)
	if err != nil {
		return nil, err
	}
	return &doc, nil
}

func (r *Repository) ListAttributes() ([]Attribute, error) {
	var attrs []Attribute
	err := r.db.Select(&attrs, "SELECT * FROM kyc_attributes ORDER BY code")
	return attrs, err
}

func (r *Repository) GetAttributeByCode(code string) (*Attribute, error) {
	var attr Attribute
	err := r.db.Get(&attr, "SELECT * FROM kyc_attributes WHERE code=$1", code)
	if err != nil {
		return nil, err
	}
	return &attr, nil
}

func (r *Repository) GetAttributesForDocument(docCode string) ([]Attribute, error) {
	query := `
		SELECT a.* FROM kyc_attributes a
		JOIN kyc_attr_doc_links l ON l.attribute_code=a.code
		WHERE l.document_code=$1
		ORDER BY a.code
	`
	var attrs []Attribute
	err := r.db.Select(&attrs, query, docCode)
	return attrs, err
}

func (r *Repository) GetDocumentSources(attributeCode string) ([]AttributeDocumentLink, error) {
	var links []AttributeDocumentLink
	err := r.db.Select(&links, `
		SELECT * FROM kyc_attr_doc_links WHERE attribute_code=$1 ORDER BY source_tier
	`, attributeCode)
	return links, err
}

func (r *Repository) GetDocumentsForAttribute(attributeCode string) ([]Document, error) {
	query := `
		SELECT d.* FROM kyc_documents d
		JOIN kyc_attr_doc_links l ON l.document_code=d.code
		WHERE l.attribute_code=$1
		ORDER BY l.source_tier, d.code
	`
	var docs []Document
	err := r.db.Select(&docs, query, attributeCode)
	return docs, err
}

func (r *Repository) GetDocumentRegulationLinks(docCode string) ([]DocumentRegulationLink, error) {
	var links []DocumentRegulationLink
	err := r.db.Select(&links, `
		SELECT * FROM kyc_doc_reg_links WHERE document_code=$1
	`, docCode)
	return links, err
}

func (r *Repository) GetRegulationsForDocument(docCode string) ([]Regulation, error) {
	query := `
		SELECT r.* FROM kyc_regulations r
		JOIN kyc_doc_reg_links l ON l.regulation_code=r.code
		WHERE l.document_code=$1
		ORDER BY r.code
	`
	var regs []Regulation
	err := r.db.Select(&regs, query, docCode)
	return regs, err
}

func (r *Repository) DebugPrintOntologySummary() error {
	regs, err := r.ListRegulations()
	if err != nil {
		return err
	}
	fmt.Println("\n=== Regulatory Data Ontology Summary ===")
	for _, reg := range regs {
		fmt.Printf("📘 %s — %s (%s)\n", reg.Code, reg.Name, reg.Jurisdiction)
		docs, _ := r.ListDocumentsByRegulation(reg.Code)
		for _, d := range docs {
			fmt.Printf("   📄 %s — %s\n", d.Code, d.Name)
		}
	}
	return nil
}

// AllDocumentCodes returns all document codes for validation
func (r *Repository) AllDocumentCodes() ([]string, error) {
	var codes []string
	err := r.db.Select(&codes, "SELECT code FROM kyc_documents ORDER BY code")
	return codes, err
}

// AllAttributeCodes returns all attribute codes for validation
func (r *Repository) AllAttributeCodes() ([]string, error) {
	var codes []string
	err := r.db.Select(&codes, "SELECT code FROM kyc_attributes ORDER BY code")
	return codes, err
}

// AllRegulationCodes returns all regulation codes for validation
func (r *Repository) AllRegulationCodes() ([]string, error) {
	var codes []string
	err := r.db.Select(&codes, "SELECT code FROM kyc_regulations ORDER BY code")
	return codes, err
}

// DocumentLinkedToRegulation checks if a document is linked to at least one regulation
func (r *Repository) DocumentLinkedToRegulation(docCode string) (bool, error) {
	var count int
	err := r.db.Get(&count, `
		SELECT COUNT(*) FROM kyc_doc_reg_links WHERE document_code=$1
	`, docCode)
	return count > 0, err
}

// ListPublicAttributes returns all public attributes
func (r *Repository) ListPublicAttributes() ([]Attribute, error) {
	var attrs []Attribute
	err := r.db.Select(&attrs, `
		SELECT * FROM kyc_attributes
		WHERE attribute_class = 'Public'
		ORDER BY code
	`)
	return attrs, err
}

// ListPrivateAttributes returns all private (derived) attributes
func (r *Repository) ListPrivateAttributes() ([]Attribute, error) {
	var attrs []Attribute
	err := r.db.Select(&attrs, `
		SELECT * FROM kyc_attributes
		WHERE attribute_class = 'Private'
		ORDER BY code
	`)
	return attrs, err
}

// GetAttributeDerivations returns all derivations for a private attribute
func (r *Repository) GetAttributeDerivations(derivedAttrCode string) ([]AttributeDerivation, error) {
	var derivations []AttributeDerivation
	err := r.db.Select(&derivations, `
		SELECT * FROM kyc_attribute_derivations
		WHERE derived_attribute_code=$1
		ORDER BY source_attribute_code
	`, derivedAttrCode)
	return derivations, err
}

// GetAttributeLineage returns the full lineage view for a derived attribute
func (r *Repository) GetAttributeLineage(derivedAttrCode string) ([]AttributeDerivation, error) {
	var lineage []AttributeDerivation
	err := r.db.Select(&lineage, `
		SELECT
			d.id,
			d.derived_attribute_code,
			d.source_attribute_code,
			d.rule_expression,
			d.rule_type,
			d.description,
			d.created_at
		FROM kyc_attribute_derivations d
		WHERE d.derived_attribute_code=$1
		ORDER BY d.source_attribute_code
	`, derivedAttrCode)
	return lineage, err
}

// InsertAttributeDerivation adds a new derivation rule
func (r *Repository) InsertAttributeDerivation(d AttributeDerivation) error {
	query := `
		INSERT INTO kyc_attribute_derivations
		(derived_attribute_code, source_attribute_code, rule_expression, rule_type, description)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.db.Exec(query, d.DerivedAttributeCode, d.SourceAttributeCode,
		d.RuleExpression, d.RuleType, d.Description)
	return err
}

// ValidateDerivationSources checks that all source attributes exist and are public
func (r *Repository) ValidateDerivationSources(sourceAttrCodes []string) error {
	for _, code := range sourceAttrCodes {
		var attrClass string
		err := r.db.Get(&attrClass, `
			SELECT attribute_class FROM kyc_attributes WHERE code=$1
		`, code)
		if err != nil {
			return fmt.Errorf("source attribute '%s' not found", code)
		}
		if attrClass != "Public" {
			return fmt.Errorf("source attribute '%s' must be Public, got '%s'", code, attrClass)
		}
	}
	return nil
}

// GetAttributeMetadata retrieves metadata for an attribute
func (r *Repository) GetAttributeMetadata(attrCode string) (*AttributeMetadata, error) {
	var metadata AttributeMetadata
	err := r.db.Get(&metadata, `
		SELECT * FROM kyc_attribute_metadata WHERE attribute_code=$1
	`, attrCode)
	if err != nil {
		return nil, err
	}
	return &metadata, nil
}

// ListAttributeClusters returns all attribute clusters
func (r *Repository) ListAttributeClusters() ([]AttributeCluster, error) {
	var clusters []AttributeCluster
	err := r.db.Select(&clusters, `
		SELECT * FROM kyc_attribute_clusters ORDER BY priority, cluster_name
	`)
	return clusters, err
}

// GetAttributeCluster retrieves a specific cluster by code
func (r *Repository) GetAttributeCluster(clusterCode string) (*AttributeCluster, error) {
	var cluster AttributeCluster
	err := r.db.Get(&cluster, `
		SELECT * FROM kyc_attribute_clusters WHERE cluster_code=$1
	`, clusterCode)
	if err != nil {
		return nil, err
	}
	return &cluster, nil
}

// GetClusterAttributes returns all attributes in a cluster
func (r *Repository) GetClusterAttributes(clusterCode string) ([]Attribute, error) {
	var attrs []Attribute
	err := r.db.Select(&attrs, `
		SELECT a.* FROM kyc_attributes a
		WHERE a.code = ANY(
			SELECT unnest(attribute_codes)
			FROM kyc_attribute_clusters
			WHERE cluster_code=$1
		)
		ORDER BY a.code
	`, clusterCode)
	return attrs, err
}

// GetAttributeProfile returns complete profile with metadata
func (r *Repository) GetAttributeProfile(attrCode string) (*AttributeProfile, error) {
	var profile AttributeProfile
	err := r.db.Get(&profile, `
		SELECT
			a.code,
			a.name,
			a.domain,
			a.description,
			a.risk_category,
			a.is_personal_data,
			a.attribute_class,
			m.synonyms,
			m.data_type,
			m.domain_values,
			m.risk_level,
			m.example_values,
			m.regulatory_citations,
			m.data_sensitivity,
			m.retention_period_days
		FROM kyc_attributes a
		LEFT JOIN kyc_attribute_metadata m ON m.attribute_code = a.code
		WHERE a.code=$1
	`, attrCode)
	if err != nil {
		return nil, err
	}
	return &profile, nil
}

// FindAttributeBySynonym searches for attributes by synonym
func (r *Repository) FindAttributeBySynonym(synonym string) ([]Attribute, error) {
	var attrs []Attribute
	err := r.db.Select(&attrs, `
		SELECT a.* FROM kyc_attributes a
		JOIN kyc_attribute_metadata m ON m.attribute_code = a.code
		WHERE $1 = ANY(m.synonyms)
		ORDER BY a.code
	`, synonym)
	return attrs, err
}

// GetAttributeRelationships returns relationships for an attribute
func (r *Repository) GetAttributeRelationships(attrCode string) ([]AttributeRelationship, error) {
	var rels []AttributeRelationship
	err := r.db.Select(&rels, `
		SELECT * FROM kyc_attribute_relationships
		WHERE source_attribute_code=$1 OR target_attribute_code=$1
		ORDER BY strength DESC
	`, attrCode)
	return rels, err
}

// FindRelatedAttributes finds semantically related attributes
func (r *Repository) FindRelatedAttributes(attrCode string, maxDepth int) ([]Attribute, error) {
	var attrs []Attribute
	err := r.db.Select(&attrs, `
		SELECT DISTINCT a.*
		FROM find_related_attributes($1, $2) f
		JOIN kyc_attributes a ON a.code = f.related_code
		ORDER BY f.depth, a.code
	`, attrCode, maxDepth)
	return attrs, err
}
