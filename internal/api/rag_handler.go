package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/jmoiron/sqlx"

	"github.com/adamtc007/KYC-DSL/internal/ontology"
	"github.com/adamtc007/KYC-DSL/internal/rag"
)

// RagHandler handles RAG and vector search API endpoints
type RagHandler struct {
	DB       *sqlx.DB
	Embedder *rag.Embedder
}

// NewRagHandler creates a new RAG handler with OpenAI client
func NewRagHandler(db *sqlx.DB, embedder *rag.Embedder) *RagHandler {
	return &RagHandler{
		DB:       db,
		Embedder: embedder,
	}
}

// AttributeSearchResponse represents the API response
type AttributeSearchResponse struct {
	Query   string            `json:"query"`
	Limit   int               `json:"limit"`
	Count   int               `json:"count"`
	Results []AttributeResult `json:"results"`
}

// AttributeResult represents a single search result
type AttributeResult struct {
	Code                string   `json:"code"`
	RiskLevel           string   `json:"risk_level"`
	DataType            string   `json:"data_type"`
	Description         string   `json:"business_context"`
	Synonyms            []string `json:"synonyms,omitempty"`
	RegulatoryCitations []string `json:"regulatory_citations,omitempty"`
	ExampleValues       []string `json:"example_values,omitempty"`
	SimilarityScore     float64  `json:"similarity_score"`
	Distance            float64  `json:"distance"`
}

// SimilarAttributesResponse represents similar attributes API response
type SimilarAttributesResponse struct {
	SourceAttribute string            `json:"source_attribute"`
	Limit           int               `json:"limit"`
	Count           int               `json:"count"`
	Results         []AttributeResult `json:"results"`
}

// TextSearchResponse represents text search API response
type TextSearchResponse struct {
	SearchTerm string            `json:"search_term"`
	Count      int               `json:"count"`
	Results    []AttributeResult `json:"results"`
}

// StatsResponse represents metadata statistics API response
type StatsResponse struct {
	TotalAttributes          int                    `json:"total_attributes"`
	AttributesWithEmbeddings int                    `json:"attributes_with_embeddings"`
	EmbeddingCoveragePercent float64                `json:"embedding_coverage_percent"`
	RiskDistribution         []RiskDistributionItem `json:"risk_distribution"`
}

// RiskDistributionItem represents risk level counts
type RiskDistributionItem struct {
	RiskLevel string `json:"risk_level"`
	Count     int    `json:"count"`
}

// ErrorResponse represents an API error
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// HandleAttributeSearch performs semantic search on attributes
// GET /rag/attribute_search?q=<query>&limit=<limit>
func (h *RagHandler) HandleAttributeSearch(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	query := r.URL.Query().Get("q")
	if query == "" {
		h.sendError(w, http.StatusBadRequest, "missing 'q' query parameter")
		return
	}

	limit := 10
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	ctx := context.Background()

	// Generate embedding for query
	queryEmbedding, err := h.Embedder.GenerateEmbeddingFromText(ctx, query)
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, "failed to generate query embedding: "+err.Error())
		return
	}

	// Perform vector search
	repo := ontology.NewMetadataRepo(h.DB)
	results, err := repo.SearchByVector(ctx, queryEmbedding, limit)
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, "failed to search: "+err.Error())
		return
	}

	// Format response
	response := AttributeSearchResponse{
		Query:   query,
		Limit:   limit,
		Count:   len(results),
		Results: make([]AttributeResult, 0, len(results)),
	}

	for _, r := range results {
		response.Results = append(response.Results, AttributeResult{
			Code:                r.AttributeCode,
			RiskLevel:           r.RiskLevel,
			DataType:            r.DataType,
			Description:         strings.TrimSpace(r.BusinessContext),
			Synonyms:            r.Synonyms,
			RegulatoryCitations: r.RegulatoryCitations,
			ExampleValues:       r.ExampleValues,
			SimilarityScore:     r.SimilarityScore,
			Distance:            r.Distance,
		})
	}

	h.sendJSON(w, http.StatusOK, response)
}

// HandleSimilarAttributes finds attributes similar to a given attribute
// GET /rag/similar_attributes?code=<attribute_code>&limit=<limit>
func (h *RagHandler) HandleSimilarAttributes(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	attributeCode := r.URL.Query().Get("code")
	if attributeCode == "" {
		h.sendError(w, http.StatusBadRequest, "missing 'code' query parameter")
		return
	}

	limit := 10
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	ctx := context.Background()

	// Find similar attributes
	repo := ontology.NewMetadataRepo(h.DB)
	results, err := repo.FindSimilarAttributes(ctx, attributeCode, limit)
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, "failed to find similar attributes: "+err.Error())
		return
	}

	// Format response
	response := SimilarAttributesResponse{
		SourceAttribute: attributeCode,
		Limit:           limit,
		Count:           len(results),
		Results:         make([]AttributeResult, 0, len(results)),
	}

	for _, r := range results {
		response.Results = append(response.Results, AttributeResult{
			Code:                r.AttributeCode,
			RiskLevel:           r.RiskLevel,
			DataType:            r.DataType,
			Description:         strings.TrimSpace(r.BusinessContext),
			Synonyms:            r.Synonyms,
			RegulatoryCitations: r.RegulatoryCitations,
			ExampleValues:       r.ExampleValues,
			SimilarityScore:     r.SimilarityScore,
			Distance:            r.Distance,
		})
	}

	h.sendJSON(w, http.StatusOK, response)
}

// HandleTextSearch performs traditional text-based search
// GET /rag/text_search?term=<search_term>
func (h *RagHandler) HandleTextSearch(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	searchTerm := r.URL.Query().Get("term")
	if searchTerm == "" {
		h.sendError(w, http.StatusBadRequest, "missing 'term' query parameter")
		return
	}

	ctx := context.Background()

	// Perform text search
	repo := ontology.NewMetadataRepo(h.DB)
	results, err := repo.SearchByText(ctx, searchTerm)
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, "failed to search: "+err.Error())
		return
	}

	// Format response
	response := TextSearchResponse{
		SearchTerm: searchTerm,
		Count:      len(results),
		Results:    make([]AttributeResult, 0, len(results)),
	}

	for _, r := range results {
		response.Results = append(response.Results, AttributeResult{
			Code:                r.AttributeCode,
			RiskLevel:           r.RiskLevel,
			DataType:            r.DataType,
			Description:         strings.TrimSpace(r.BusinessContext),
			Synonyms:            r.Synonyms,
			RegulatoryCitations: r.RegulatoryCitations,
			ExampleValues:       r.ExampleValues,
		})
	}

	h.sendJSON(w, http.StatusOK, response)
}

// HandleMetadataStats returns metadata repository statistics
// GET /rag/stats
func (h *RagHandler) HandleMetadataStats(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// Get stats
	repo := ontology.NewMetadataRepo(h.DB)
	stats, err := repo.GetMetadataStats(ctx)
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, "failed to get stats: "+err.Error())
		return
	}

	// Format response
	response := StatsResponse{
		TotalAttributes:          stats["total_attributes"].(int),
		AttributesWithEmbeddings: stats["attributes_with_embeddings"].(int),
		EmbeddingCoveragePercent: stats["embedding_coverage_percent"].(float64),
		RiskDistribution:         make([]RiskDistributionItem, 0),
	}

	// Extract risk distribution
	if riskDist, ok := stats["risk_distribution"].([]struct {
		RiskLevel string `db:"risk_level"`
		Count     int    `db:"count"`
	}); ok {
		for _, rd := range riskDist {
			response.RiskDistribution = append(response.RiskDistribution, RiskDistributionItem{
				RiskLevel: rd.RiskLevel,
				Count:     rd.Count,
			})
		}
	}

	h.sendJSON(w, http.StatusOK, response)
}

// HandleGetAttribute retrieves metadata for a specific attribute
// GET /rag/attribute/<code>
func (h *RagHandler) HandleGetAttribute(w http.ResponseWriter, r *http.Request) {
	// Extract attribute code from URL path
	path := strings.TrimPrefix(r.URL.Path, "/rag/attribute/")
	attributeCode := strings.TrimSpace(path)

	if attributeCode == "" {
		h.sendError(w, http.StatusBadRequest, "missing attribute code in path")
		return
	}

	ctx := context.Background()

	// Get metadata
	repo := ontology.NewMetadataRepo(h.DB)
	metadata, err := repo.GetMetadata(ctx, attributeCode)
	if err != nil {
		h.sendError(w, http.StatusNotFound, "attribute not found: "+attributeCode)
		return
	}

	// Format response
	result := AttributeResult{
		Code:                metadata.AttributeCode,
		RiskLevel:           metadata.RiskLevel,
		DataType:            metadata.DataType,
		Description:         strings.TrimSpace(metadata.BusinessContext),
		Synonyms:            metadata.Synonyms,
		RegulatoryCitations: metadata.RegulatoryCitations,
		ExampleValues:       metadata.ExampleValues,
	}

	h.sendJSON(w, http.StatusOK, result)
}

// HandleHealth is a health check endpoint
// GET /rag/health
func (h *RagHandler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// Check database connection
	if err := h.DB.PingContext(ctx); err != nil {
		h.sendError(w, http.StatusServiceUnavailable, "database connection failed")
		return
	}

	// Check embeddings count
	repo := ontology.NewMetadataRepo(h.DB)
	count, err := repo.CountEmbeddings(ctx)
	if err != nil {
		h.sendError(w, http.StatusServiceUnavailable, "failed to check embeddings")
		return
	}

	h.sendJSON(w, http.StatusOK, map[string]interface{}{
		"status":               "healthy",
		"embeddings_count":     count,
		"embedding_model":      string(h.Embedder.GetModel()),
		"embedding_dimensions": h.Embedder.GetDimensions(),
	})
}

// sendJSON sends a JSON response
func (h *RagHandler) sendJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		// Log error but can't change response at this point
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

// sendError sends an error response
func (h *RagHandler) sendError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ErrorResponse{
		Error:   http.StatusText(statusCode),
		Message: message,
	})
}
