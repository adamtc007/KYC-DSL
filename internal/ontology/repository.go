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
