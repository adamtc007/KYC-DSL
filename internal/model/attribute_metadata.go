package model

import "time"

// AttributeMetadata represents rich metadata for an attribute with vector embeddings
type AttributeMetadata struct {
	ID                  int       `db:"id"`
	AttributeCode       string    `db:"attribute_code"`
	Synonyms            []string  `db:"synonyms"`
	DataType            string    `db:"data_type"`
	DomainValues        []string  `db:"domain_values"`
	RiskLevel           string    `db:"risk_level"`
	ExampleValues       []string  `db:"example_values"`
	RegulatoryCitations []string  `db:"regulatory_citations"`
	BusinessContext     string    `db:"business_context"`
	Embedding           []float32 `db:"embedding"`
	CreatedAt           time.Time `db:"created_at"`
}

// AttributeSearchResult represents a search result with similarity score
type AttributeSearchResult struct {
	AttributeMetadata
	SimilarityScore float64 `db:"similarity_score"`
	Distance        float64 `db:"distance"`
}

// EmbeddingRequest represents a request to generate embeddings
type EmbeddingRequest struct {
	AttributeCode       string
	Synonyms            []string
	BusinessContext     string
	RegulatoryCitations []string
	ExampleValues       []string
}

// ToEmbeddingText converts metadata to text suitable for embedding
func (m *AttributeMetadata) ToEmbeddingText() string {
	text := m.AttributeCode

	if m.BusinessContext != "" {
		text += ". Definition: " + m.BusinessContext
	}

	if len(m.Synonyms) > 0 {
		text += ". Synonyms: "
		for i, syn := range m.Synonyms {
			if i > 0 {
				text += ", "
			}
			text += syn
		}
	}

	if len(m.RegulatoryCitations) > 0 {
		text += ". Citations: "
		for i, cit := range m.RegulatoryCitations {
			if i > 0 {
				text += ", "
			}
			text += cit
		}
	}

	if len(m.ExampleValues) > 0 {
		text += ". Examples: "
		for i, ex := range m.ExampleValues {
			if i > 0 {
				text += ", "
			}
			text += ex
		}
	}

	return text
}
