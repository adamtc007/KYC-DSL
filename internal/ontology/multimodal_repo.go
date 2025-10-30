package ontology

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"github.com/adamtc007/KYC-DSL/internal/model"
)

// MultiModalRepo handles multi-modal RAG queries across attributes, documents, and regulations
type MultiModalRepo struct {
	db *sqlx.DB
}

// NewMultiModalRepo creates a new multi-modal repository
func NewMultiModalRepo(db *sqlx.DB) *MultiModalRepo {
	return &MultiModalRepo{db: db}
}

// SearchAttributesAndDocs performs semantic search and enriches results with linked documents and regulations
func (r *MultiModalRepo) SearchAttributesAndDocs(ctx context.Context, vec []float32, limit int) ([]model.MultiModalResult, error) {
	// 1. Get top-matching attributes by vector similarity
	query := `
		SELECT
			id, attribute_code, synonyms, data_type, domain_values, risk_level,
			example_values, regulatory_citations, business_context, embedding, created_at
		FROM kyc_attribute_metadata
		WHERE embedding IS NOT NULL
		ORDER BY embedding <=> $1::vector
		LIMIT $2
	`

	var attrs []model.AttributeMetadata
	err := r.db.SelectContext(ctx, &attrs, query, pq.Array(vec), limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search attributes: %w", err)
	}

	results := make([]model.MultiModalResult, 0, len(attrs))

	// 2. For each attribute, fetch linked documents and regulations
	for _, attr := range attrs {
		result := model.MultiModalResult{
			Attribute:   attr,
			Documents:   []model.Document{},
			Regulations: []model.Regulation{},
		}

		// Fetch linked documents
		docQuery := `
			SELECT DISTINCT
				d.id, d.code, d.name,
				COALESCE(d.title, d.name) as title,
				d.domain, d.jurisdiction,
				COALESCE(d.doc_type, '') as doc_type,
				COALESCE(d.description, '') as description,
				d.embedding, d.created_at
			FROM kyc_documents d
			JOIN kyc_attr_doc_links l ON l.document_code = d.code
			WHERE l.attribute_code = $1
			ORDER BY COALESCE(l.relevance_score, 1.0) DESC
		`

		var docs []model.Document
		err = r.db.SelectContext(ctx, &docs, docQuery, attr.AttributeCode)
		if err != nil {
			// Log but don't fail - some attributes may not have linked documents
			fmt.Printf("Warning: failed to fetch documents for %s: %v\n", attr.AttributeCode, err)
		} else {
			result.Documents = docs
		}

		// Fetch linked regulations
		regQuery := `
			SELECT DISTINCT
				r.id, r.code, r.name,
				COALESCE(r.title, r.name) as title,
				COALESCE(r.region, r.jurisdiction) as region,
				r.jurisdiction, r.authority,
				COALESCE(r.citation, '') as citation,
				COALESCE(r.summary, r.description) as summary,
				r.description, r.embedding, r.created_at
			FROM kyc_regulations r
			JOIN kyc_attr_doc_links l ON l.regulation_code = r.code
			WHERE l.attribute_code = $1
			ORDER BY COALESCE(l.relevance_score, 1.0) DESC
		`

		var regs []model.Regulation
		err = r.db.SelectContext(ctx, &regs, regQuery, attr.AttributeCode)
		if err != nil {
			// Log but don't fail - some attributes may not have linked regulations
			fmt.Printf("Warning: failed to fetch regulations for %s: %v\n", attr.AttributeCode, err)
		} else {
			result.Regulations = regs
		}

		results = append(results, result)
	}

	return results, nil
}

// SearchDocuments performs semantic search on documents
func (r *MultiModalRepo) SearchDocuments(ctx context.Context, vec []float32, limit int) ([]model.DocumentSearchResult, error) {
	query := `
		SELECT
			id, code, name,
			COALESCE(title, name) as title,
			domain, jurisdiction,
			COALESCE(doc_type, '') as doc_type,
			COALESCE(description, '') as description,
			embedding, created_at,
			1 - (embedding <=> $1::vector) as similarity_score,
			embedding <=> $1::vector as distance
		FROM kyc_documents
		WHERE embedding IS NOT NULL
		ORDER BY embedding <=> $1::vector
		LIMIT $2
	`

	var results []model.DocumentSearchResult
	err := r.db.SelectContext(ctx, &results, query, pq.Array(vec), limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search documents: %w", err)
	}

	return results, nil
}

// SearchRegulations performs semantic search on regulations
func (r *MultiModalRepo) SearchRegulations(ctx context.Context, vec []float32, limit int) ([]model.RegulationSearchResult, error) {
	query := `
		SELECT
			id, code, name,
			COALESCE(title, name) as title,
			COALESCE(region, jurisdiction) as region,
			jurisdiction, authority,
			COALESCE(citation, '') as citation,
			COALESCE(summary, description) as summary,
			description, embedding, created_at,
			1 - (embedding <=> $1::vector) as similarity_score,
			embedding <=> $1::vector as distance
		FROM kyc_regulations
		WHERE embedding IS NOT NULL
		ORDER BY embedding <=> $1::vector
		LIMIT $2
	`

	var results []model.RegulationSearchResult
	err := r.db.SelectContext(ctx, &results, query, pq.Array(vec), limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search regulations: %w", err)
	}

	return results, nil
}

// GetDocumentsByAttribute retrieves all documents linked to an attribute
func (r *MultiModalRepo) GetDocumentsByAttribute(ctx context.Context, attributeCode string) ([]model.Document, error) {
	query := `
		SELECT DISTINCT
			d.id, d.code, d.name,
			COALESCE(d.title, d.name) as title,
			d.domain, d.jurisdiction,
			COALESCE(d.doc_type, '') as doc_type,
			COALESCE(d.description, '') as description,
			d.embedding, d.created_at
		FROM kyc_documents d
		JOIN kyc_attr_doc_links l ON l.document_code = d.code
		WHERE l.attribute_code = $1
		ORDER BY COALESCE(l.relevance_score, 1.0) DESC
	`

	var docs []model.Document
	err := r.db.SelectContext(ctx, &docs, query, attributeCode)
	if err != nil {
		return nil, fmt.Errorf("failed to get documents for attribute %s: %w", attributeCode, err)
	}

	return docs, nil
}

// GetRegulationsByAttribute retrieves all regulations linked to an attribute
func (r *MultiModalRepo) GetRegulationsByAttribute(ctx context.Context, attributeCode string) ([]model.Regulation, error) {
	query := `
		SELECT DISTINCT
			r.id, r.code, r.name,
			COALESCE(r.title, r.name) as title,
			COALESCE(r.region, r.jurisdiction) as region,
			r.jurisdiction, r.authority,
			COALESCE(r.citation, '') as citation,
			COALESCE(r.summary, r.description) as summary,
			r.description, r.embedding, r.created_at
		FROM kyc_regulations r
		JOIN kyc_attr_doc_links l ON l.regulation_code = r.code
		WHERE l.attribute_code = $1
		ORDER BY COALESCE(l.relevance_score, 1.0) DESC
	`

	var regs []model.Regulation
	err := r.db.SelectContext(ctx, &regs, query, attributeCode)
	if err != nil {
		return nil, fmt.Errorf("failed to get regulations for attribute %s: %w", attributeCode, err)
	}

	return regs, nil
}

// GetAttributesByDocument retrieves all attributes linked to a document
func (r *MultiModalRepo) GetAttributesByDocument(ctx context.Context, documentCode string) ([]model.AttributeMetadata, error) {
	query := `
		SELECT
			am.id, am.attribute_code, am.synonyms, am.data_type, am.domain_values,
			am.risk_level, am.example_values, am.regulatory_citations,
			am.business_context, am.embedding, am.created_at
		FROM kyc_attribute_metadata am
		JOIN kyc_attr_doc_links l ON l.attribute_code = am.attribute_code
		WHERE l.document_code = $1
		ORDER BY COALESCE(l.relevance_score, 1.0) DESC
	`

	var attrs []model.AttributeMetadata
	err := r.db.SelectContext(ctx, &attrs, query, documentCode)
	if err != nil {
		return nil, fmt.Errorf("failed to get attributes for document %s: %w", documentCode, err)
	}

	return attrs, nil
}

// UpsertDocumentEmbedding inserts or updates a document with its embedding
func (r *MultiModalRepo) UpsertDocumentEmbedding(ctx context.Context, doc model.Document) error {
	query := `
		INSERT INTO kyc_documents (code, name, title, domain, jurisdiction, doc_type, description, embedding)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (code)
		DO UPDATE SET
			name = EXCLUDED.name,
			title = EXCLUDED.title,
			domain = EXCLUDED.domain,
			jurisdiction = EXCLUDED.jurisdiction,
			doc_type = EXCLUDED.doc_type,
			description = EXCLUDED.description,
			embedding = EXCLUDED.embedding
		RETURNING id
	`

	var id int
	err := r.db.QueryRowContext(ctx, query,
		doc.Code, doc.Name, doc.Title, doc.Domain, doc.Jurisdiction,
		doc.DocType, doc.Description, pq.Array(doc.Embedding),
	).Scan(&id)

	if err != nil {
		return fmt.Errorf("failed to upsert document %s: %w", doc.Code, err)
	}

	return nil
}

// UpsertRegulationEmbedding inserts or updates a regulation with its embedding
func (r *MultiModalRepo) UpsertRegulationEmbedding(ctx context.Context, reg model.Regulation) error {
	query := `
		INSERT INTO kyc_regulations (code, name, title, jurisdiction, region, authority, citation, summary, description, embedding)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (code)
		DO UPDATE SET
			name = EXCLUDED.name,
			title = EXCLUDED.title,
			jurisdiction = EXCLUDED.jurisdiction,
			region = EXCLUDED.region,
			authority = EXCLUDED.authority,
			citation = EXCLUDED.citation,
			summary = EXCLUDED.summary,
			description = EXCLUDED.description,
			embedding = EXCLUDED.embedding
		RETURNING id
	`

	var id int
	err := r.db.QueryRowContext(ctx, query,
		reg.Code, reg.Name, reg.Title, reg.Jurisdiction, reg.Region,
		reg.Authority, reg.Citation, reg.Summary, reg.Description, pq.Array(reg.Embedding),
	).Scan(&id)

	if err != nil {
		return fmt.Errorf("failed to upsert regulation %s: %w", reg.Code, err)
	}

	return nil
}

// CountDocuments returns the total number of documents
func (r *MultiModalRepo) CountDocuments(ctx context.Context) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count, "SELECT COUNT(*) FROM kyc_documents")
	if err != nil {
		return 0, fmt.Errorf("failed to count documents: %w", err)
	}
	return count, nil
}

// CountRegulations returns the total number of regulations
func (r *MultiModalRepo) CountRegulations(ctx context.Context) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count, "SELECT COUNT(*) FROM kyc_regulations")
	if err != nil {
		return 0, fmt.Errorf("failed to count regulations: %w", err)
	}
	return count, nil
}

// CountDocumentEmbeddings returns the number of documents with embeddings
func (r *MultiModalRepo) CountDocumentEmbeddings(ctx context.Context) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count, "SELECT COUNT(*) FROM kyc_documents WHERE embedding IS NOT NULL")
	if err != nil {
		return 0, fmt.Errorf("failed to count document embeddings: %w", err)
	}
	return count, nil
}

// CountRegulationEmbeddings returns the number of regulations with embeddings
func (r *MultiModalRepo) CountRegulationEmbeddings(ctx context.Context) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count, "SELECT COUNT(*) FROM kyc_regulations WHERE embedding IS NOT NULL")
	if err != nil {
		return 0, fmt.Errorf("failed to count regulation embeddings: %w", err)
	}
	return count, nil
}
