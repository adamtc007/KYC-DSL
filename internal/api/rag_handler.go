package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/jmoiron/sqlx"

	"github.com/adamtc007/KYC-DSL/internal/model"
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

// MultiModalResponse represents enriched search results with documents and regulations
type MultiModalResponse struct {
	Query   string                      `json:"query"`
	Limit   int                         `json:"limit"`
	Count   int                         `json:"count"`
	Results []MultiModalAttributeResult `json:"results"`
}

// MultiModalAttributeResult represents an attribute with linked documents and regulations
type MultiModalAttributeResult struct {
	Attribute   AttributeResultSimple `json:"attribute"`
	Documents   []DocumentResult      `json:"documents"`
	Regulations []RegulationResult    `json:"regulations"`
}

// AttributeResultSimple is a simplified attribute result for multi-modal queries
type AttributeResultSimple struct {
	Code        string `json:"code"`
	RiskLevel   string `json:"risk_level"`
	Description string `json:"business_context"`
}

// DocumentResult represents a document in search results
type DocumentResult struct {
	Code         string `json:"code"`
	Title        string `json:"title"`
	Jurisdiction string `json:"jurisdiction"`
	Description  string `json:"description"`
	DocType      string `json:"doc_type,omitempty"`
}

// RegulationResult represents a regulation in search results
type RegulationResult struct {
	Code     string `json:"code"`
	Title    string `json:"title"`
	Citation string `json:"citation"`
	Summary  string `json:"summary"`
	Region   string `json:"region,omitempty"`
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

// HandleEnrichedAttributeSearch performs multi-modal semantic search with documents and regulations
// GET /rag/attribute_search_enriched?q=<query>&limit=<limit>
func (h *RagHandler) HandleEnrichedAttributeSearch(w http.ResponseWriter, r *http.Request) {
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

	// Perform multi-modal search
	repo := ontology.NewMultiModalRepo(h.DB)
	results, err := repo.SearchAttributesAndDocs(ctx, queryEmbedding, limit)
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, "failed to search: "+err.Error())
		return
	}

	// Format response
	type DocResult struct {
		Code         string `json:"code"`
		Title        string `json:"title"`
		Jurisdiction string `json:"jurisdiction"`
		DocType      string `json:"doc_type,omitempty"`
		Description  string `json:"description"`
	}

	type RegResult struct {
		Code     string `json:"code"`
		Title    string `json:"title"`
		Citation string `json:"citation,omitempty"`
		Summary  string `json:"summary"`
		Region   string `json:"region,omitempty"`
	}

	type EnrichedResult struct {
		Attribute   AttributeResult `json:"attribute"`
		Documents   []DocResult     `json:"documents"`
		Regulations []RegResult     `json:"regulations"`
	}

	enrichedResults := make([]EnrichedResult, 0, len(results))

	for _, r := range results {
		// Format attribute
		attr := AttributeResult{
			Code:                r.Attribute.AttributeCode,
			RiskLevel:           r.Attribute.RiskLevel,
			DataType:            r.Attribute.DataType,
			Description:         strings.TrimSpace(r.Attribute.BusinessContext),
			Synonyms:            r.Attribute.Synonyms,
			RegulatoryCitations: r.Attribute.RegulatoryCitations,
			ExampleValues:       r.Attribute.ExampleValues,
		}

		// Format documents
		docs := make([]DocResult, 0, len(r.Documents))
		for _, d := range r.Documents {
			docs = append(docs, DocResult{
				Code:         d.Code,
				Title:        d.Title,
				Jurisdiction: d.Jurisdiction,
				DocType:      d.DocType,
				Description:  strings.TrimSpace(d.Description),
			})
		}

		// Format regulations
		regs := make([]RegResult, 0, len(r.Regulations))
		for _, reg := range r.Regulations {
			regs = append(regs, RegResult{
				Code:     reg.Code,
				Title:    reg.Title,
				Citation: reg.Citation,
				Summary:  strings.TrimSpace(reg.Summary),
				Region:   reg.Region,
			})
		}

		enrichedResults = append(enrichedResults, EnrichedResult{
			Attribute:   attr,
			Documents:   docs,
			Regulations: regs,
		})
	}

	response := map[string]interface{}{
		"query":   query,
		"limit":   limit,
		"count":   len(enrichedResults),
		"results": enrichedResults,
	}

	h.sendJSON(w, http.StatusOK, response)
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

// HandleMultiModalSearch performs enriched semantic search with documents and regulations
// GET /rag/multimodal_search?q=<query>&limit=<limit>
func (h *RagHandler) HandleMultiModalSearch(w http.ResponseWriter, r *http.Request) {
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

	// Perform multi-modal search
	repo := ontology.NewMultiModalRepo(h.DB)
	results, err := repo.SearchAttributesAndDocs(ctx, queryEmbedding, limit)
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, "failed to search: "+err.Error())
		return
	}

	// Format response
	response := MultiModalResponse{
		Query:   query,
		Limit:   limit,
		Count:   len(results),
		Results: make([]MultiModalAttributeResult, 0, len(results)),
	}

	for _, result := range results {
		// Format attribute
		attrResult := AttributeResultSimple{
			Code:        result.Attribute.AttributeCode,
			RiskLevel:   result.Attribute.RiskLevel,
			Description: strings.TrimSpace(result.Attribute.BusinessContext),
		}

		// Format documents
		docs := make([]DocumentResult, 0, len(result.Documents))
		for _, doc := range result.Documents {
			docs = append(docs, DocumentResult{
				Code:         doc.Code,
				Title:        doc.Title,
				Jurisdiction: doc.Jurisdiction,
				Description:  strings.TrimSpace(doc.Description),
				DocType:      doc.DocType,
			})
		}

		// Format regulations
		regs := make([]RegulationResult, 0, len(result.Regulations))
		for _, reg := range result.Regulations {
			regs = append(regs, RegulationResult{
				Code:     reg.Code,
				Title:    reg.Title,
				Citation: reg.Citation,
				Summary:  strings.TrimSpace(reg.Summary),
				Region:   reg.Region,
			})
		}

		response.Results = append(response.Results, MultiModalAttributeResult{
			Attribute:   attrResult,
			Documents:   docs,
			Regulations: regs,
		})
	}

	h.sendJSON(w, http.StatusOK, response)
}

// HandleGetDocuments returns all documents with optional filtering
// GET /rag/documents?attribute=<code>
func (h *RagHandler) HandleGetDocuments(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	repo := ontology.NewMultiModalRepo(h.DB)

	// Check if filtering by attribute
	attributeCode := r.URL.Query().Get("attribute")

	var docs []model.Document
	var err error

	if attributeCode != "" {
		docs, err = repo.GetDocumentsByAttribute(ctx, attributeCode)
	} else {
		// For now, return error - full list could be large
		h.sendError(w, http.StatusBadRequest, "attribute parameter required")
		return
	}

	if err != nil {
		h.sendError(w, http.StatusInternalServerError, "failed to fetch documents: "+err.Error())
		return
	}

	// Format response
	results := make([]DocumentResult, 0, len(docs))
	for _, doc := range docs {
		results = append(results, DocumentResult{
			Code:         doc.Code,
			Title:        doc.Title,
			Jurisdiction: doc.Jurisdiction,
			Description:  strings.TrimSpace(doc.Description),
			DocType:      doc.DocType,
		})
	}

	h.sendJSON(w, http.StatusOK, map[string]interface{}{
		"attribute": attributeCode,
		"count":     len(results),
		"documents": results,
	})
}

// HandleGetRegulations returns all regulations with optional filtering
// GET /rag/regulations?attribute=<code>
func (h *RagHandler) HandleGetRegulations(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	repo := ontology.NewMultiModalRepo(h.DB)

	// Check if filtering by attribute
	attributeCode := r.URL.Query().Get("attribute")

	var regs []model.Regulation
	var err error

	if attributeCode != "" {
		regs, err = repo.GetRegulationsByAttribute(ctx, attributeCode)
	} else {
		// For now, return error - full list could be large
		h.sendError(w, http.StatusBadRequest, "attribute parameter required")
		return
	}

	if err != nil {
		h.sendError(w, http.StatusInternalServerError, "failed to fetch regulations: "+err.Error())
		return
	}

	// Format response
	results := make([]RegulationResult, 0, len(regs))
	for _, reg := range regs {
		results = append(results, RegulationResult{
			Code:     reg.Code,
			Title:    reg.Title,
			Citation: reg.Citation,
			Summary:  strings.TrimSpace(reg.Summary),
			Region:   reg.Region,
		})
	}

	h.sendJSON(w, http.StatusOK, map[string]interface{}{
		"attribute":   attributeCode,
		"count":       len(results),
		"regulations": results,
	})
}

// HandleFeedback handles POST /rag/feedback - submit feedback on search results
func (h *RagHandler) HandleFeedback(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.sendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req model.FeedbackSubmitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid JSON: "+err.Error())
		return
	}

	// Validate required fields
	if req.QueryText == "" {
		h.sendError(w, http.StatusBadRequest, "query_text is required")
		return
	}

	// Validate at least one entity is provided
	if req.AttributeCode == nil && req.DocumentCode == nil && req.RegulationCode == nil {
		h.sendError(w, http.StatusBadRequest, "At least one of attribute_code, document_code, or regulation_code must be provided")
		return
	}

	// Set defaults
	if req.Feedback == "" {
		req.Feedback = model.FeedbackSentimentPositive
	}
	if req.AgentType == "" {
		req.AgentType = model.AgentTypeHuman
	}
	if req.Confidence == 0 {
		req.Confidence = 1.0
	}

	// Validate confidence range
	if req.Confidence < 0 || req.Confidence > 1 {
		h.sendError(w, http.StatusBadRequest, "confidence must be between 0 and 1")
		return
	}

	// Create feedback entry
	feedback := model.Feedback{
		QueryText:      req.QueryText,
		AttributeCode:  req.AttributeCode,
		DocumentCode:   req.DocumentCode,
		RegulationCode: req.RegulationCode,
		Feedback:       req.Feedback,
		Confidence:     req.Confidence,
		AgentName:      req.AgentName,
		AgentType:      req.AgentType,
	}

	// Insert feedback
	repo := ontology.NewFeedbackRepo(h.DB)
	id, err := repo.InsertFeedback(feedback)
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, "Failed to save feedback: "+err.Error())
		return
	}

	// Return response
	response := model.FeedbackResponse{
		Status:    "ok",
		ID:        id,
		Feedback:  req.Feedback,
		AgentName: req.AgentName,
		CreatedAt: feedback.CreatedAt,
	}

	h.sendJSON(w, http.StatusOK, response)
}

// HandleRecentFeedback handles GET /rag/feedback/recent - get recent feedback entries
func (h *RagHandler) HandleRecentFeedback(w http.ResponseWriter, r *http.Request) {
	// Parse limit parameter
	limit := 50
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	// Get recent feedback
	repo := ontology.NewFeedbackRepo(h.DB)
	feedbacks, err := repo.GetRecentFeedback(limit)
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, "Failed to get recent feedback: "+err.Error())
		return
	}

	response := model.RecentFeedbackResponse{
		Count:     len(feedbacks),
		Feedbacks: feedbacks,
	}

	h.sendJSON(w, http.StatusOK, response)
}

// HandleFeedbackAnalytics handles GET /rag/feedback/analytics - get feedback analytics
func (h *RagHandler) HandleFeedbackAnalytics(w http.ResponseWriter, r *http.Request) {
	// Parse topN parameter
	topN := 10
	if topNStr := r.URL.Query().Get("top"); topNStr != "" {
		if n, err := strconv.Atoi(topNStr); err == nil && n > 0 {
			topN = n
		}
	}

	// Get analytics
	repo := ontology.NewFeedbackRepo(h.DB)
	analytics, err := repo.GetFeedbackAnalytics(topN)
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, "Failed to get feedback analytics: "+err.Error())
		return
	}

	h.sendJSON(w, http.StatusOK, analytics)
}

// HandleFeedbackByAttribute handles GET /rag/feedback/attribute/{code} - get feedback for a specific attribute
func (h *RagHandler) HandleFeedbackByAttribute(w http.ResponseWriter, r *http.Request) {
	// Extract attribute code from path
	path := strings.TrimPrefix(r.URL.Path, "/rag/feedback/attribute/")
	attributeCode := strings.TrimSpace(path)

	if attributeCode == "" {
		h.sendError(w, http.StatusBadRequest, "attribute_code is required")
		return
	}

	// Get feedback for attribute
	repo := ontology.NewFeedbackRepo(h.DB)
	feedbacks, err := repo.GetFeedbackByAttribute(attributeCode)
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, "Failed to get feedback for attribute: "+err.Error())
		return
	}

	response := model.RecentFeedbackResponse{
		Count:     len(feedbacks),
		Feedbacks: feedbacks,
	}

	h.sendJSON(w, http.StatusOK, response)
}

// HandleFeedbackSummary handles GET /rag/feedback/summary - get feedback summary
func (h *RagHandler) HandleFeedbackSummary(w http.ResponseWriter, r *http.Request) {
	repo := ontology.NewFeedbackRepo(h.DB)

	// Get overall summary
	summary, err := repo.GetFeedbackSummary()
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, "Failed to get feedback summary: "+err.Error())
		return
	}

	// Get attribute summary
	limit := 20
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	attrSummary, err := repo.GetAttributeFeedbackSummary(limit)
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, "Failed to get attribute feedback summary: "+err.Error())
		return
	}

	h.sendJSON(w, http.StatusOK, map[string]interface{}{
		"overall_summary":   summary,
		"attribute_summary": attrSummary,
	})
}
