package ontology

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"github.com/adamtc007/KYC-DSL/internal/model"
)

// MetadataRepo handles attribute metadata operations including vector search
type MetadataRepo struct {
	db *sqlx.DB
}

// NewMetadataRepo creates a new metadata repository
func NewMetadataRepo(db *sqlx.DB) *MetadataRepo {
	return &MetadataRepo{db: db}
}

// UpsertMetadata inserts or updates attribute metadata with embedding
func (r *MetadataRepo) UpsertMetadata(ctx context.Context, m model.AttributeMetadata) error {
	query := `
		INSERT INTO kyc_attribute_metadata
			(attribute_code, synonyms, data_type, domain_values, risk_level,
			 example_values, regulatory_citations, business_context, embedding)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (attribute_code)
		DO UPDATE SET
			synonyms = EXCLUDED.synonyms,
			data_type = EXCLUDED.data_type,
			domain_values = EXCLUDED.domain_values,
			risk_level = EXCLUDED.risk_level,
			example_values = EXCLUDED.example_values,
			regulatory_citations = EXCLUDED.regulatory_citations,
			business_context = EXCLUDED.business_context,
			embedding = EXCLUDED.embedding,
			updated_at = NOW()
		RETURNING id
	`

	var id int
	err := r.db.QueryRowContext(ctx, query,
		m.AttributeCode,
		pq.Array(m.Synonyms),
		m.DataType,
		pq.Array(m.DomainValues),
		m.RiskLevel,
		pq.Array(m.ExampleValues),
		pq.Array(m.RegulatoryCitations),
		m.BusinessContext,
		pq.Array(m.Embedding),
	).Scan(&id)

	if err != nil {
		return fmt.Errorf("failed to upsert metadata for %s: %w", m.AttributeCode, err)
	}

	return nil
}

// GetMetadata retrieves metadata for a specific attribute
func (r *MetadataRepo) GetMetadata(ctx context.Context, attributeCode string) (*model.AttributeMetadata, error) {
	query := `
		SELECT id, attribute_code, synonyms, data_type, domain_values, risk_level,
		       example_values, regulatory_citations, business_context, embedding, created_at
		FROM kyc_attribute_metadata
		WHERE attribute_code = $1
	`

	var m model.AttributeMetadata
	err := r.db.GetContext(ctx, &m, query, attributeCode)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("metadata not found for attribute: %s", attributeCode)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get metadata: %w", err)
	}

	return &m, nil
}

// ListAllMetadata retrieves all attribute metadata
func (r *MetadataRepo) ListAllMetadata(ctx context.Context) ([]model.AttributeMetadata, error) {
	query := `
		SELECT id, attribute_code, synonyms, data_type, domain_values, risk_level,
		       example_values, regulatory_citations, business_context, embedding, created_at
		FROM kyc_attribute_metadata
		ORDER BY attribute_code
	`

	var results []model.AttributeMetadata
	err := r.db.SelectContext(ctx, &results, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list metadata: %w", err)
	}

	return results, nil
}

// SearchByVector performs semantic search using vector similarity
func (r *MetadataRepo) SearchByVector(ctx context.Context, vec []float32, limit int) ([]model.AttributeSearchResult, error) {
	query := `
		SELECT
			id, attribute_code, synonyms, data_type, domain_values, risk_level,
			example_values, regulatory_citations, business_context, embedding, created_at,
			1 - (embedding <=> $1::vector) as similarity_score,
			embedding <=> $1::vector as distance
		FROM kyc_attribute_metadata
		WHERE embedding IS NOT NULL
		ORDER BY embedding <=> $1::vector
		LIMIT $2
	`

	var results []model.AttributeSearchResult
	err := r.db.SelectContext(ctx, &results, query, pq.Array(vec), limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search by vector: %w", err)
	}

	return results, nil
}

// SearchByText searches for attributes by synonym or keyword
func (r *MetadataRepo) SearchByText(ctx context.Context, searchTerm string) ([]model.AttributeMetadata, error) {
	query := `
		SELECT
			id, attribute_code, synonyms, data_type, domain_values, risk_level,
			example_values, regulatory_citations, business_context, embedding, created_at
		FROM kyc_attribute_metadata
		WHERE
			attribute_code ILIKE $1
			OR business_context ILIKE $1
			OR $2 = ANY(synonyms)
		ORDER BY attribute_code
	`

	pattern := "%" + searchTerm + "%"
	var results []model.AttributeMetadata
	err := r.db.SelectContext(ctx, &results, query, pattern, searchTerm)
	if err != nil {
		return nil, fmt.Errorf("failed to search by text: %w", err)
	}

	return results, nil
}

// GetAttributesWithoutEmbeddings returns attributes that don't have embeddings yet
func (r *MetadataRepo) GetAttributesWithoutEmbeddings(ctx context.Context) ([]model.AttributeMetadata, error) {
	query := `
		SELECT
			id, attribute_code, synonyms, data_type, domain_values, risk_level,
			example_values, regulatory_citations, business_context, embedding, created_at
		FROM kyc_attribute_metadata
		WHERE embedding IS NULL
		ORDER BY attribute_code
	`

	var results []model.AttributeMetadata
	err := r.db.SelectContext(ctx, &results, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get attributes without embeddings: %w", err)
	}

	return results, nil
}

// CountMetadata returns total count of attributes with metadata
func (r *MetadataRepo) CountMetadata(ctx context.Context) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count, "SELECT COUNT(*) FROM kyc_attribute_metadata")
	if err != nil {
		return 0, fmt.Errorf("failed to count metadata: %w", err)
	}
	return count, nil
}

// CountEmbeddings returns count of attributes with embeddings
func (r *MetadataRepo) CountEmbeddings(ctx context.Context) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count, "SELECT COUNT(*) FROM kyc_attribute_metadata WHERE embedding IS NOT NULL")
	if err != nil {
		return 0, fmt.Errorf("failed to count embeddings: %w", err)
	}
	return count, nil
}

// FindSimilarAttributes finds attributes similar to a given attribute code
func (r *MetadataRepo) FindSimilarAttributes(ctx context.Context, attributeCode string, limit int) ([]model.AttributeSearchResult, error) {
	// First get the embedding for the source attribute
	metadata, err := r.GetMetadata(ctx, attributeCode)
	if err != nil {
		return nil, err
	}

	if len(metadata.Embedding) == 0 {
		return nil, fmt.Errorf("source attribute %s has no embedding", attributeCode)
	}

	// Search for similar attributes (excluding itself)
	query := `
		SELECT
			id, attribute_code, synonyms, data_type, domain_values, risk_level,
			example_values, regulatory_citations, business_context, embedding, created_at,
			1 - (embedding <=> $1::vector) as similarity_score,
			embedding <=> $1::vector as distance
		FROM kyc_attribute_metadata
		WHERE embedding IS NOT NULL
		  AND attribute_code != $2
		ORDER BY embedding <=> $1::vector
		LIMIT $3
	`

	var results []model.AttributeSearchResult
	err = r.db.SelectContext(ctx, &results, query, pq.Array(metadata.Embedding), attributeCode, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to find similar attributes: %w", err)
	}

	return results, nil
}

// GetMetadataStats returns statistics about the metadata repository
func (r *MetadataRepo) GetMetadataStats(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Total count
	totalCount, err := r.CountMetadata(ctx)
	if err != nil {
		return nil, err
	}
	stats["total_attributes"] = totalCount

	// Embedding count
	embeddingCount, err := r.CountEmbeddings(ctx)
	if err != nil {
		return nil, err
	}
	stats["attributes_with_embeddings"] = embeddingCount
	stats["embedding_coverage_percent"] = float64(embeddingCount) / float64(totalCount) * 100

	// Risk level distribution
	var riskStats []struct {
		RiskLevel string `db:"risk_level"`
		Count     int    `db:"count"`
	}
	err = r.db.SelectContext(ctx, &riskStats, `
		SELECT risk_level, COUNT(*) as count
		FROM kyc_attribute_metadata
		WHERE risk_level IS NOT NULL
		GROUP BY risk_level
		ORDER BY count DESC
	`)
	if err != nil {
		return nil, err
	}
	stats["risk_distribution"] = riskStats

	return stats, nil
}
