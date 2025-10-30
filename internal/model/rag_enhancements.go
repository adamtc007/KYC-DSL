package model

import "time"

// ==================== Enhancement A: Feedback Loop ====================

// FeedbackType represents the type of feedback (positive or negative)
type FeedbackType string

const (
	FeedbackPositive FeedbackType = "positive"
	FeedbackNegative FeedbackType = "negative"
)

// RAGFeedback represents agent feedback on retrieval quality
type RAGFeedback struct {
	ID             int          `db:"id" json:"id"`
	QueryText      string       `db:"query_text" json:"query_text"`
	AttributeCode  string       `db:"attribute_code" json:"attribute_code,omitempty"`
	DocumentCode   string       `db:"document_code" json:"document_code,omitempty"`
	RegulationCode string       `db:"regulation_code" json:"regulation_code,omitempty"`
	Feedback       FeedbackType `db:"feedback" json:"feedback"`
	AgentName      string       `db:"agent_name" json:"agent_name,omitempty"`
	SessionID      string       `db:"session_id" json:"session_id,omitempty"`
	RelevanceScore float64      `db:"relevance_score" json:"relevance_score,omitempty"`
	Notes          string       `db:"notes" json:"notes,omitempty"`
	CreatedAt      time.Time    `db:"created_at" json:"created_at"`
}

// FeedbackRequest represents an API request to submit feedback
type FeedbackRequest struct {
	Query          string  `json:"query"`
	AttributeCode  string  `json:"attribute_code,omitempty"`
	DocumentCode   string  `json:"document_code,omitempty"`
	RegulationCode string  `json:"regulation_code,omitempty"`
	Feedback       string  `json:"feedback"` // "positive" or "negative"
	Agent          string  `json:"agent,omitempty"`
	SessionID      string  `json:"session_id,omitempty"`
	RelevanceScore float64 `json:"relevance_score,omitempty"`
	Notes          string  `json:"notes,omitempty"`
}

// FeedbackStats represents aggregated feedback statistics
type FeedbackStats struct {
	AttributeCode string    `db:"attribute_code" json:"attribute_code"`
	TotalFeedback int       `db:"total_feedback" json:"total_feedback"`
	PositiveCount int       `db:"positive_count" json:"positive_count"`
	NegativeCount int       `db:"negative_count" json:"negative_count"`
	PositivePct   float64   `db:"positive_pct" json:"positive_pct"`
	AvgRelevance  float64   `db:"avg_relevance" json:"avg_relevance"`
	LastFeedback  time.Time `db:"last_feedback" json:"last_feedback"`
}

// ==================== Enhancement C: Snippet-Level Retrieval ====================

// DocumentSection represents a fine-grained section of a document with embedding
type DocumentSection struct {
	ID            int       `db:"id" json:"id"`
	DocumentCode  string    `db:"document_code" json:"document_code"`
	SectionNumber string    `db:"section_number" json:"section_number,omitempty"`
	SectionTitle  string    `db:"section_title" json:"section_title,omitempty"`
	TextExcerpt   string    `db:"text_excerpt" json:"text_excerpt"`
	PageNumber    int       `db:"page_number" json:"page_number,omitempty"`
	Embedding     []float32 `db:"embedding" json:"embedding,omitempty"`
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
}

// DocumentSectionSearchResult represents a section search result with similarity
type DocumentSectionSearchResult struct {
	DocumentSection
	SimilarityScore float64 `db:"similarity_score" json:"similarity_score"`
	Distance        float64 `db:"distance" json:"distance"`
}

// DocumentSectionContext represents a section with full document context
type DocumentSectionContext struct {
	SectionID       int    `db:"section_id" json:"section_id"`
	SectionNumber   string `db:"section_number" json:"section_number,omitempty"`
	SectionTitle    string `db:"section_title" json:"section_title,omitempty"`
	TextExcerpt     string `db:"text_excerpt" json:"text_excerpt"`
	PageNumber      int    `db:"page_number" json:"page_number,omitempty"`
	DocumentCode    string `db:"document_code" json:"document_code"`
	DocumentTitle   string `db:"document_title" json:"document_title"`
	Jurisdiction    string `db:"jurisdiction" json:"jurisdiction,omitempty"`
	DocType         string `db:"doc_type" json:"doc_type,omitempty"`
	RegulationCode  string `db:"regulation_code" json:"regulation_code,omitempty"`
	RegulationTitle string `db:"regulation_title" json:"regulation_title,omitempty"`
}

// ToEmbeddingText converts document section to text suitable for embedding
func (s *DocumentSection) ToEmbeddingText() string {
	text := ""

	if s.DocumentCode != "" {
		text += s.DocumentCode
	}

	if s.SectionNumber != "" {
		text += " Section " + s.SectionNumber
	}

	if s.SectionTitle != "" {
		text += ". " + s.SectionTitle
	}

	if s.TextExcerpt != "" {
		text += ". " + s.TextExcerpt
	}

	return text
}

// ==================== Enhancement D: Semantic Clusters ====================

// RAGCluster represents a semantic cluster of related attributes
type RAGCluster struct {
	ID                   int       `db:"id" json:"id"`
	ClusterCode          string    `db:"cluster_code" json:"cluster_code"`
	ClusterName          string    `db:"cluster_name" json:"cluster_name"`
	Description          string    `db:"description" json:"description,omitempty"`
	Centroid             []float32 `db:"centroid" json:"centroid,omitempty"`
	MemberAttributeCodes []string  `db:"member_attribute_codes" json:"member_attribute_codes"`
	MemberCount          int       `db:"member_count" json:"member_count"`
	QualityScore         float64   `db:"quality_score" json:"quality_score"`
	LastComputed         time.Time `db:"last_computed" json:"last_computed"`
	CreatedAt            time.Time `db:"created_at" json:"created_at"`
}

// ClusterRecommendation represents a recommended cluster for a query
type ClusterRecommendation struct {
	ClusterCode string  `db:"cluster_code" json:"cluster_code"`
	ClusterName string  `db:"cluster_name" json:"cluster_name"`
	Similarity  float64 `db:"similarity" json:"similarity"`
	MemberCount int     `db:"member_count" json:"member_count"`
}

// ClusterDetails represents detailed cluster membership information
type ClusterDetails struct {
	ClusterCode   string    `db:"cluster_code" json:"cluster_code"`
	ClusterName   string    `db:"cluster_name" json:"cluster_name"`
	Description   string    `db:"description" json:"description,omitempty"`
	MemberCount   int       `db:"member_count" json:"member_count"`
	QualityScore  float64   `db:"quality_score" json:"quality_score"`
	LastComputed  time.Time `db:"last_computed" json:"last_computed"`
	AttributeCode string    `db:"attribute_code" json:"attribute_code"`
	AttributeName string    `db:"attribute_name" json:"attribute_name"`
	RiskLevel     string    `db:"risk_level" json:"risk_level,omitempty"`
}

// ==================== Enhancement E: RAG Audit Trail ====================

// RAGAuditLog represents a complete audit record of a RAG query
type RAGAuditLog struct {
	ID             int       `db:"id" json:"id"`
	QueryText      string    `db:"query_text" json:"query_text"`
	QueryEmbedding []float32 `db:"query_embedding" json:"query_embedding,omitempty"`
	Response       string    `db:"response" json:"response"` // JSONB stored as string
	ResultCount    int       `db:"result_count" json:"result_count"`
	AgentName      string    `db:"agent_name" json:"agent_name,omitempty"`
	SessionID      string    `db:"session_id" json:"session_id,omitempty"`
	Endpoint       string    `db:"endpoint" json:"endpoint,omitempty"`
	LatencyMs      int       `db:"latency_ms" json:"latency_ms,omitempty"`
	ErrorMessage   string    `db:"error_message" json:"error_message,omitempty"`
	IPAddress      string    `db:"ip_address" json:"ip_address,omitempty"`
	UserAgent      string    `db:"user_agent" json:"user_agent,omitempty"`
	CreatedAt      time.Time `db:"created_at" json:"created_at"`
}

// PopularQuery represents aggregated query statistics
type PopularQuery struct {
	QueryText    string    `db:"query_text" json:"query_text"`
	QueryCount   int       `db:"query_count" json:"query_count"`
	AvgLatencyMs float64   `db:"avg_latency_ms" json:"avg_latency_ms"`
	AvgResults   float64   `db:"avg_results" json:"avg_results"`
	LastQueried  time.Time `db:"last_queried" json:"last_queried"`
}

// AgentPerformance represents performance metrics for an agent
type AgentPerformance struct {
	AgentName    string  `db:"agent_name" json:"agent_name"`
	Sessions     int     `db:"sessions" json:"sessions"`
	TotalQueries int     `db:"total_queries" json:"total_queries"`
	AvgLatencyMs float64 `db:"avg_latency_ms" json:"avg_latency_ms"`
	ErrorCount   int     `db:"error_count" json:"error_count"`
	SuccessRate  float64 `db:"success_rate" json:"success_rate"`
}

// AuditLogRequest represents the data to log for a RAG query
type AuditLogRequest struct {
	QueryText      string      `json:"query_text"`
	QueryEmbedding []float32   `json:"query_embedding,omitempty"`
	Response       interface{} `json:"response"`
	AgentName      string      `json:"agent_name,omitempty"`
	SessionID      string      `json:"session_id,omitempty"`
	Endpoint       string      `json:"endpoint,omitempty"`
	LatencyMs      int         `json:"latency_ms,omitempty"`
	IPAddress      string      `json:"ip_address,omitempty"`
	UserAgent      string      `json:"user_agent,omitempty"`
}

// ==================== Enhancement B: Derived Attribute Awareness ====================

// DerivedAttributeInfo represents lineage information for a derived attribute
type DerivedAttributeInfo struct {
	DerivedAttributeCode string   `json:"derived_attribute_code"`
	RuleExpression       string   `json:"rule_expression"`
	RuleType             string   `json:"rule_type"`
	SourceAttributes     []string `json:"source_attributes"`
	Description          string   `json:"description,omitempty"`
}

// EnrichedAttributeResult extends AttributeResult with derivation info
type EnrichedAttributeResult struct {
	AttributeMetadata
	Derived *DerivedAttributeInfo `json:"derived,omitempty"`
}
