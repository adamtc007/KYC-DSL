package ontology

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"github.com/adamtc007/KYC-DSL/internal/model"
)

// EnhancementsRepo handles RAG enhancement operations
type EnhancementsRepo struct {
	db *sqlx.DB
}

// NewEnhancementsRepo creates a new enhancements repository
func NewEnhancementsRepo(db *sqlx.DB) *EnhancementsRepo {
	return &EnhancementsRepo{db: db}
}

// ==================== Enhancement A: Feedback Loop ====================

// InsertFeedback records agent feedback on retrieval quality
func (r *EnhancementsRepo) InsertFeedback(ctx context.Context, feedback model.RAGFeedback) (int, error) {
	query := `
		INSERT INTO rag_feedback
			(query_text, attribute_code, document_code, regulation_code,
			 feedback, agent_name, session_id, relevance_score, notes)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id
	`

	var id int
	err := r.db.QueryRowContext(ctx, query,
		feedback.QueryText,
		nullString(feedback.AttributeCode),
		nullString(feedback.DocumentCode),
		nullString(feedback.RegulationCode),
		feedback.Feedback,
		nullString(feedback.AgentName),
		nullString(feedback.SessionID),
		nullFloat64(feedback.RelevanceScore),
		nullString(feedback.Notes),
	).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("failed to insert feedback: %w", err)
	}

	return id, nil
}

// GetFeedbackStats retrieves aggregated feedback statistics by attribute
func (r *EnhancementsRepo) GetFeedbackStats(ctx context.Context) ([]model.FeedbackStats, error) {
	query := `
		SELECT
			attribute_code,
			total_feedback,
			positive_count,
			negative_count,
			positive_pct,
			avg_relevance,
			last_feedback
		FROM feedback_stats_by_attribute
		ORDER BY total_feedback DESC
	`

	var stats []model.FeedbackStats
	err := r.db.SelectContext(ctx, &stats, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get feedback stats: %w", err)
	}

	return stats, nil
}

// GetFeedbackByAttribute retrieves all feedback for a specific attribute
func (r *EnhancementsRepo) GetFeedbackByAttribute(ctx context.Context, attributeCode string) ([]model.RAGFeedback, error) {
	query := `
		SELECT
			id, query_text, attribute_code, document_code, regulation_code,
			feedback, agent_name, session_id, relevance_score, notes, created_at
		FROM rag_feedback
		WHERE attribute_code = $1
		ORDER BY created_at DESC
	`

	var feedback []model.RAGFeedback
	err := r.db.SelectContext(ctx, &feedback, query, attributeCode)
	if err != nil {
		return nil, fmt.Errorf("failed to get feedback: %w", err)
	}

	return feedback, nil
}

// GetRecentFeedback retrieves the most recent feedback entries
func (r *EnhancementsRepo) GetRecentFeedback(ctx context.Context, limit int) ([]model.RAGFeedback, error) {
	query := `
		SELECT
			id, query_text, attribute_code, document_code, regulation_code,
			feedback, agent_name, session_id, relevance_score, notes, created_at
		FROM rag_feedback
		ORDER BY created_at DESC
		LIMIT $1
	`

	var feedback []model.RAGFeedback
	err := r.db.SelectContext(ctx, &feedback, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent feedback: %w", err)
	}

	return feedback, nil
}

// ==================== Enhancement C: Snippet-Level Retrieval ====================

// InsertDocumentSection inserts a document section with embedding
func (r *EnhancementsRepo) InsertDocumentSection(ctx context.Context, section model.DocumentSection) (int, error) {
	query := `
		INSERT INTO kyc_document_sections
			(document_code, section_number, section_title, text_excerpt, page_number, embedding)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

	var id int
	err := r.db.QueryRowContext(ctx, query,
		section.DocumentCode,
		nullString(section.SectionNumber),
		nullString(section.SectionTitle),
		section.TextExcerpt,
		nullInt(section.PageNumber),
		pq.Array(section.Embedding),
	).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("failed to insert document section: %w", err)
	}

	return id, nil
}

// SearchDocumentSections performs semantic search on document sections
func (r *EnhancementsRepo) SearchDocumentSections(ctx context.Context, vec []float32, limit int) ([]model.DocumentSectionSearchResult, error) {
	query := `
		SELECT
			id, document_code, section_number, section_title, text_excerpt, page_number,
			embedding, created_at,
			1 - (embedding <=> $1::vector) as similarity_score,
			embedding <=> $1::vector as distance
		FROM kyc_document_sections
		WHERE embedding IS NOT NULL
		ORDER BY embedding <=> $1::vector
		LIMIT $2
	`

	var results []model.DocumentSectionSearchResult
	err := r.db.SelectContext(ctx, &results, query, pq.Array(vec), limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search document sections: %w", err)
	}

	return results, nil
}

// GetSectionsByDocument retrieves all sections for a document
func (r *EnhancementsRepo) GetSectionsByDocument(ctx context.Context, documentCode string) ([]model.DocumentSection, error) {
	query := `
		SELECT
			id, document_code, section_number, section_title, text_excerpt,
			page_number, embedding, created_at
		FROM kyc_document_sections
		WHERE document_code = $1
		ORDER BY section_number, page_number
	`

	var sections []model.DocumentSection
	err := r.db.SelectContext(ctx, &sections, query, documentCode)
	if err != nil {
		return nil, fmt.Errorf("failed to get sections for document %s: %w", documentCode, err)
	}

	return sections, nil
}

// GetSectionContext retrieves section with full document and regulation context
func (r *EnhancementsRepo) GetSectionContext(ctx context.Context, sectionID int) (*model.DocumentSectionContext, error) {
	query := `
		SELECT
			section_id, section_number, section_title, text_excerpt, page_number,
			document_code, document_title, jurisdiction, doc_type,
			regulation_code, regulation_title
		FROM document_section_context
		WHERE section_id = $1
	`

	var context model.DocumentSectionContext
	err := r.db.GetContext(ctx, &context, query, sectionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get section context: %w", err)
	}

	return &context, nil
}

// CountSections returns total count of document sections
func (r *EnhancementsRepo) CountSections(ctx context.Context) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count, "SELECT COUNT(*) FROM kyc_document_sections")
	if err != nil {
		return 0, fmt.Errorf("failed to count sections: %w", err)
	}
	return count, nil
}

// CountSectionEmbeddings returns count of sections with embeddings
func (r *EnhancementsRepo) CountSectionEmbeddings(ctx context.Context) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count, "SELECT COUNT(*) FROM kyc_document_sections WHERE embedding IS NOT NULL")
	if err != nil {
		return 0, fmt.Errorf("failed to count section embeddings: %w", err)
	}
	return count, nil
}

// ==================== Enhancement D: Semantic Clusters ====================

// GetCluster retrieves a cluster by code
func (r *EnhancementsRepo) GetCluster(ctx context.Context, clusterCode string) (*model.RAGCluster, error) {
	query := `
		SELECT
			id, cluster_code, cluster_name, description, centroid,
			member_attribute_codes, member_count, quality_score,
			last_computed, created_at
		FROM rag_clusters
		WHERE cluster_code = $1
	`

	var cluster model.RAGCluster
	err := r.db.GetContext(ctx, &cluster, query, clusterCode)
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster %s: %w", clusterCode, err)
	}

	return &cluster, nil
}

// GetAllClusters retrieves all clusters
func (r *EnhancementsRepo) GetAllClusters(ctx context.Context) ([]model.RAGCluster, error) {
	query := `
		SELECT
			id, cluster_code, cluster_name, description, centroid,
			member_attribute_codes, member_count, quality_score,
			last_computed, created_at
		FROM rag_clusters
		ORDER BY cluster_code
	`

	var clusters []model.RAGCluster
	err := r.db.SelectContext(ctx, &clusters, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get clusters: %w", err)
	}

	return clusters, nil
}

// RecommendClusters finds the most relevant clusters for a query embedding
func (r *EnhancementsRepo) RecommendClusters(ctx context.Context, vec []float32, limit int) ([]model.ClusterRecommendation, error) {
	query := `SELECT * FROM recommend_clusters($1, $2)`

	var recommendations []model.ClusterRecommendation
	err := r.db.SelectContext(ctx, &recommendations, query, pq.Array(vec), limit)
	if err != nil {
		return nil, fmt.Errorf("failed to recommend clusters: %w", err)
	}

	return recommendations, nil
}

// SearchWithinCluster searches attributes within a specific cluster
func (r *EnhancementsRepo) SearchWithinCluster(ctx context.Context, clusterCode string, vec []float32, limit int) ([]model.AttributeSearchResult, error) {
	// Get cluster members
	cluster, err := r.GetCluster(ctx, clusterCode)
	if err != nil {
		return nil, err
	}

	// Search only within cluster members
	query := `
		SELECT
			id, attribute_code, synonyms, data_type, domain_values, risk_level,
			example_values, regulatory_citations, business_context, embedding, created_at,
			1 - (embedding <=> $1::vector) as similarity_score,
			embedding <=> $1::vector as distance
		FROM kyc_attribute_metadata
		WHERE attribute_code = ANY($2)
		  AND embedding IS NOT NULL
		ORDER BY embedding <=> $1::vector
		LIMIT $3
	`

	var results []model.AttributeSearchResult
	err = r.db.SelectContext(ctx, &results, query, pq.Array(vec), pq.Array(cluster.MemberAttributeCodes), limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search within cluster: %w", err)
	}

	return results, nil
}

// GetClusterDetails retrieves detailed cluster membership information
func (r *EnhancementsRepo) GetClusterDetails(ctx context.Context, clusterCode string) ([]model.ClusterDetails, error) {
	query := `
		SELECT
			cluster_code, cluster_name, description, member_count, quality_score,
			last_computed, attribute_code, attribute_name, risk_level
		FROM cluster_details
		WHERE cluster_code = $1
		ORDER BY attribute_code
	`

	var details []model.ClusterDetails
	err := r.db.SelectContext(ctx, &details, query, clusterCode)
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster details: %w", err)
	}

	return details, nil
}

// UpsertCluster inserts or updates a cluster
func (r *EnhancementsRepo) UpsertCluster(ctx context.Context, cluster model.RAGCluster) (int, error) {
	query := `
		INSERT INTO rag_clusters
			(cluster_code, cluster_name, description, centroid, member_attribute_codes, quality_score, last_computed)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (cluster_code)
		DO UPDATE SET
			cluster_name = EXCLUDED.cluster_name,
			description = EXCLUDED.description,
			centroid = EXCLUDED.centroid,
			member_attribute_codes = EXCLUDED.member_attribute_codes,
			quality_score = EXCLUDED.quality_score,
			last_computed = EXCLUDED.last_computed
		RETURNING id
	`

	var id int
	err := r.db.QueryRowContext(ctx, query,
		cluster.ClusterCode,
		cluster.ClusterName,
		cluster.Description,
		pq.Array(cluster.Centroid),
		pq.Array(cluster.MemberAttributeCodes),
		cluster.QualityScore,
		cluster.LastComputed,
	).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("failed to upsert cluster: %w", err)
	}

	return id, nil
}

// ComputeClusterCentroid computes the centroid embedding for a cluster
func (r *EnhancementsRepo) ComputeClusterCentroid(ctx context.Context, clusterCode string) error {
	query := `
		UPDATE rag_clusters
		SET centroid = (
			SELECT AVG(embedding)::vector(1536)
			FROM kyc_attribute_metadata
			WHERE attribute_code = ANY(rag_clusters.member_attribute_codes)
			  AND embedding IS NOT NULL
		),
		last_computed = NOW()
		WHERE cluster_code = $1
	`

	_, err := r.db.ExecContext(ctx, query, clusterCode)
	if err != nil {
		return fmt.Errorf("failed to compute centroid for cluster %s: %w", clusterCode, err)
	}

	return nil
}

// ComputeAllClusterCentroids recomputes centroids for all clusters (nightly job)
func (r *EnhancementsRepo) ComputeAllClusterCentroids(ctx context.Context) (int, error) {
	query := `
		UPDATE rag_clusters
		SET centroid = (
			SELECT AVG(embedding)::vector(1536)
			FROM kyc_attribute_metadata
			WHERE attribute_code = ANY(rag_clusters.member_attribute_codes)
			  AND embedding IS NOT NULL
		),
		last_computed = NOW()
		WHERE member_attribute_codes IS NOT NULL
	`

	result, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("failed to compute all centroids: %w", err)
	}

	count, _ := result.RowsAffected()
	return int(count), nil
}

// ==================== Enhancement E: RAG Audit Trail ====================

// LogQuery records a RAG query in the audit log
func (r *EnhancementsRepo) LogQuery(ctx context.Context, log model.RAGAuditLog) (int, error) {
	query := `
		INSERT INTO rag_audit_log
			(query_text, query_embedding, response, result_count, agent_name,
			 session_id, endpoint, latency_ms, error_message, ip_address, user_agent)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id
	`

	var id int
	err := r.db.QueryRowContext(ctx, query,
		log.QueryText,
		pq.Array(log.QueryEmbedding),
		log.Response,
		log.ResultCount,
		nullString(log.AgentName),
		nullString(log.SessionID),
		nullString(log.Endpoint),
		nullInt(log.LatencyMs),
		nullString(log.ErrorMessage),
		nullString(log.IPAddress),
		nullString(log.UserAgent),
	).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("failed to log query: %w", err)
	}

	return id, nil
}

// LogQueryWithJSON is a convenience method that accepts an interface{} for response
func (r *EnhancementsRepo) LogQueryWithJSON(ctx context.Context, queryText string, response interface{}, agentName string, latencyMs int, endpoint string) (int, error) {
	// Marshal response to JSON string
	responseJSON, err := json.Marshal(response)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal response: %w", err)
	}

	// Count results if response has a "count" field
	resultCount := 0
	if responseMap, ok := response.(map[string]interface{}); ok {
		if count, ok := responseMap["count"].(int); ok {
			resultCount = count
		}
	}

	log := model.RAGAuditLog{
		QueryText:   queryText,
		Response:    string(responseJSON),
		ResultCount: resultCount,
		AgentName:   agentName,
		Endpoint:    endpoint,
		LatencyMs:   latencyMs,
	}

	return r.LogQuery(ctx, log)
}

// GetPopularQueries retrieves the most frequently asked queries
func (r *EnhancementsRepo) GetPopularQueries(ctx context.Context, limit int) ([]model.PopularQuery, error) {
	query := `
		SELECT
			query_text, query_count, avg_latency_ms, avg_results, last_queried
		FROM popular_queries
		ORDER BY query_count DESC
		LIMIT $1
	`

	var queries []model.PopularQuery
	err := r.db.SelectContext(ctx, &queries, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get popular queries: %w", err)
	}

	return queries, nil
}

// GetAgentPerformance retrieves performance metrics for all agents
func (r *EnhancementsRepo) GetAgentPerformance(ctx context.Context) ([]model.AgentPerformance, error) {
	query := `
		SELECT
			agent_name, sessions, total_queries, avg_latency_ms,
			error_count, success_rate
		FROM agent_performance
		ORDER BY total_queries DESC
	`

	var performance []model.AgentPerformance
	err := r.db.SelectContext(ctx, &performance, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get agent performance: %w", err)
	}

	return performance, nil
}

// GetRecentQueries retrieves the most recent queries from audit log
func (r *EnhancementsRepo) GetRecentQueries(ctx context.Context, limit int) ([]model.RAGAuditLog, error) {
	query := `
		SELECT
			id, query_text, response, result_count, agent_name,
			session_id, endpoint, latency_ms, error_message, created_at
		FROM rag_audit_log
		ORDER BY created_at DESC
		LIMIT $1
	`

	var logs []model.RAGAuditLog
	err := r.db.SelectContext(ctx, &logs, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent queries: %w", err)
	}

	return logs, nil
}

// GetQueriesByAgent retrieves all queries from a specific agent
func (r *EnhancementsRepo) GetQueriesByAgent(ctx context.Context, agentName string, limit int) ([]model.RAGAuditLog, error) {
	query := `
		SELECT
			id, query_text, response, result_count, agent_name,
			session_id, endpoint, latency_ms, error_message, created_at
		FROM rag_audit_log
		WHERE agent_name = $1
		ORDER BY created_at DESC
		LIMIT $2
	`

	var logs []model.RAGAuditLog
	err := r.db.SelectContext(ctx, &logs, query, agentName, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get queries for agent: %w", err)
	}

	return logs, nil
}

// GetErrorQueries retrieves queries that resulted in errors
func (r *EnhancementsRepo) GetErrorQueries(ctx context.Context, limit int) ([]model.RAGAuditLog, error) {
	query := `
		SELECT
			id, query_text, response, result_count, agent_name,
			session_id, endpoint, latency_ms, error_message, created_at
		FROM rag_audit_log
		WHERE error_message IS NOT NULL
		ORDER BY created_at DESC
		LIMIT $1
	`

	var logs []model.RAGAuditLog
	err := r.db.SelectContext(ctx, &logs, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get error queries: %w", err)
	}

	return logs, nil
}

// CleanupOldAuditLogs removes audit logs older than specified days (except errors)
func (r *EnhancementsRepo) CleanupOldAuditLogs(ctx context.Context, retentionDays int) (int, error) {
	query := `SELECT cleanup_old_audit_logs($1)`

	var deletedCount int
	err := r.db.GetContext(ctx, &deletedCount, query, retentionDays)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup old audit logs: %w", err)
	}

	return deletedCount, nil
}

// GetAuditStats returns statistics about the audit log
func (r *EnhancementsRepo) GetAuditStats(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Total queries
	var totalQueries int
	err := r.db.GetContext(ctx, &totalQueries, "SELECT COUNT(*) FROM rag_audit_log")
	if err != nil {
		return nil, err
	}
	stats["total_queries"] = totalQueries

	// Queries today
	var queriesToday int
	err = r.db.GetContext(ctx, &queriesToday, "SELECT COUNT(*) FROM rag_audit_log WHERE created_at >= CURRENT_DATE")
	if err != nil {
		return nil, err
	}
	stats["queries_today"] = queriesToday

	// Error rate
	var errorCount int
	err = r.db.GetContext(ctx, &errorCount, "SELECT COUNT(*) FROM rag_audit_log WHERE error_message IS NOT NULL")
	if err != nil {
		return nil, err
	}
	stats["error_count"] = errorCount
	if totalQueries > 0 {
		stats["error_rate"] = float64(errorCount) / float64(totalQueries) * 100
	} else {
		stats["error_rate"] = 0.0
	}

	// Average latency
	var avgLatency float64
	err = r.db.GetContext(ctx, &avgLatency, "SELECT COALESCE(AVG(latency_ms), 0) FROM rag_audit_log WHERE latency_ms IS NOT NULL")
	if err != nil {
		return nil, err
	}
	stats["avg_latency_ms"] = avgLatency

	// Unique agents
	var uniqueAgents int
	err = r.db.GetContext(ctx, &uniqueAgents, "SELECT COUNT(DISTINCT agent_name) FROM rag_audit_log WHERE agent_name IS NOT NULL")
	if err != nil {
		return nil, err
	}
	stats["unique_agents"] = uniqueAgents

	return stats, nil
}

// ==================== Helper Functions ====================

func nullString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

func nullInt(i int) interface{} {
	if i == 0 {
		return nil
	}
	return i
}

func nullFloat64(f float64) interface{} {
	if f == 0.0 {
		return nil
	}
	return f
}
