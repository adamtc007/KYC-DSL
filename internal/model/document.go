package model

import "time"

// Document represents a KYC/compliance document type
type Document struct {
	ID           int       `db:"id"`
	Code         string    `db:"code"`
	Name         string    `db:"name"`
	Title        string    `db:"title"`
	Domain       string    `db:"domain"`
	Jurisdiction string    `db:"jurisdiction"`
	DocType      string    `db:"doc_type"`
	Description  string    `db:"description"`
	Embedding    []float32 `db:"embedding"`
	CreatedAt    time.Time `db:"created_at"`
}

// Regulation represents a regulatory framework or law
type Regulation struct {
	ID           int       `db:"id"`
	Code         string    `db:"code"`
	Name         string    `db:"name"`
	Title        string    `db:"title"`
	Region       string    `db:"region"`
	Jurisdiction string    `db:"jurisdiction"`
	Authority    string    `db:"authority"`
	Citation     string    `db:"citation"`
	Summary      string    `db:"summary"`
	Description  string    `db:"description"`
	Embedding    []float32 `db:"embedding"`
	CreatedAt    time.Time `db:"created_at"`
}

// AttributeDocumentLink represents a relationship between an attribute and a document
type AttributeDocumentLink struct {
	ID             int     `db:"id"`
	AttributeCode  string  `db:"attribute_code"`
	DocumentCode   string  `db:"document_code"`
	RegulationCode string  `db:"regulation_code"`
	SourceTier     string  `db:"source_tier"`
	RelevanceScore float64 `db:"relevance_score"`
	Notes          string  `db:"notes"`
}

// MultiModalResult combines attribute, documents, and regulations
type MultiModalResult struct {
	Attribute   AttributeMetadata
	Documents   []Document
	Regulations []Regulation
}

// DocumentSearchResult represents a document search result with similarity score
type DocumentSearchResult struct {
	Document
	SimilarityScore float64 `db:"similarity_score"`
	Distance        float64 `db:"distance"`
}

// RegulationSearchResult represents a regulation search result with similarity score
type RegulationSearchResult struct {
	Regulation
	SimilarityScore float64 `db:"similarity_score"`
	Distance        float64 `db:"distance"`
}

// ToEmbeddingText converts document metadata to text suitable for embedding
func (d *Document) ToEmbeddingText() string {
	text := d.Code

	if d.Title != "" {
		text += ". Title: " + d.Title
	}

	if d.Name != "" && d.Name != d.Title {
		text += ". Name: " + d.Name
	}

	if d.DocType != "" {
		text += ". Type: " + d.DocType
	}

	if d.Jurisdiction != "" {
		text += ". Jurisdiction: " + d.Jurisdiction
	}

	if d.Domain != "" {
		text += ". Domain: " + d.Domain
	}

	if d.Description != "" {
		text += ". " + d.Description
	}

	return text
}

// ToEmbeddingText converts regulation metadata to text suitable for embedding
func (r *Regulation) ToEmbeddingText() string {
	text := r.Code

	if r.Title != "" {
		text += ". " + r.Title
	}

	if r.Name != "" && r.Name != r.Title {
		text += ". " + r.Name
	}

	if r.Citation != "" {
		text += ". Citation: " + r.Citation
	}

	if r.Region != "" {
		text += ". Region: " + r.Region
	}

	if r.Authority != "" {
		text += ". Authority: " + r.Authority
	}

	if r.Summary != "" {
		text += ". " + r.Summary
	} else if r.Description != "" {
		text += ". " + r.Description
	}

	return text
}
