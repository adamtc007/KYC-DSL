package ontology

import (
	"database/sql"
	"fmt"

	"github.com/adamtc007/KYC-DSL/internal/model"
	"github.com/jmoiron/sqlx"
)

// FeedbackRepo manages RAG feedback operations
type FeedbackRepo struct {
	db *sqlx.DB
}

// NewFeedbackRepo creates a new feedback repository
func NewFeedbackRepo(db *sqlx.DB) *FeedbackRepo {
	return &FeedbackRepo{db: db}
}

// InsertFeedback inserts a new feedback entry into the database
func (r *FeedbackRepo) InsertFeedback(f model.Feedback) (int, error) {
	query := `
		INSERT INTO rag_feedback
			(query_text, attribute_code, document_code, regulation_code,
			 feedback, confidence, agent_name, agent_type)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id`

	var id int
	err := r.db.QueryRow(query,
		f.QueryText,
		f.AttributeCode,
		f.DocumentCode,
		f.RegulationCode,
		f.Feedback,
		f.Confidence,
		f.AgentName,
		f.AgentType,
	).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("failed to insert feedback: %w", err)
	}

	return id, nil
}

// GetRecentFeedback retrieves the most recent feedback entries
func (r *FeedbackRepo) GetRecentFeedback(limit int) ([]model.Feedback, error) {
	if limit <= 0 {
		limit = 50
	}

	var feedbacks []model.Feedback
	query := `
		SELECT id, query_text, attribute_code, document_code, regulation_code,
		       feedback, confidence, agent_name, agent_type, created_at
		FROM rag_feedback
		ORDER BY created_at DESC
		LIMIT $1`

	err := r.db.Select(&feedbacks, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent feedback: %w", err)
	}

	return feedbacks, nil
}

// GetFeedbackSummary retrieves aggregated feedback statistics
func (r *FeedbackRepo) GetFeedbackSummary() ([]model.FeedbackSummary, error) {
	var summaries []model.FeedbackSummary
	query := `
		SELECT feedback, agent_type, count, avg_confidence,
		       first_feedback, last_feedback
		FROM rag_feedback_summary
		ORDER BY feedback, agent_type`

	err := r.db.Select(&summaries, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get feedback summary: %w", err)
	}

	return summaries, nil
}

// GetAttributeFeedbackSummary retrieves feedback statistics per attribute
func (r *FeedbackRepo) GetAttributeFeedbackSummary(limit int) ([]model.AttributeFeedbackSummary, error) {
	if limit <= 0 {
		limit = 20
	}

	var summaries []model.AttributeFeedbackSummary
	query := `
		SELECT attribute_code, feedback, feedback_count,
		       avg_confidence, agent_types
		FROM attribute_feedback_summary
		ORDER BY feedback_count DESC
		LIMIT $1`

	err := r.db.Select(&summaries, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get attribute feedback summary: %w", err)
	}

	return summaries, nil
}

// GetFeedbackByAttribute retrieves all feedback for a specific attribute
func (r *FeedbackRepo) GetFeedbackByAttribute(attributeCode string) ([]model.Feedback, error) {
	var feedbacks []model.Feedback
	query := `
		SELECT id, query_text, attribute_code, document_code, regulation_code,
		       feedback, confidence, agent_name, agent_type, created_at
		FROM rag_feedback
		WHERE attribute_code = $1
		ORDER BY created_at DESC`

	err := r.db.Select(&feedbacks, query, attributeCode)
	if err != nil {
		return nil, fmt.Errorf("failed to get feedback by attribute: %w", err)
	}

	return feedbacks, nil
}

// GetFeedbackByQuery retrieves all feedback for a specific query
func (r *FeedbackRepo) GetFeedbackByQuery(queryText string) ([]model.Feedback, error) {
	var feedbacks []model.Feedback
	query := `
		SELECT id, query_text, attribute_code, document_code, regulation_code,
		       feedback, confidence, agent_name, agent_type, created_at
		FROM rag_feedback
		WHERE query_text ILIKE $1
		ORDER BY created_at DESC`

	err := r.db.Select(&feedbacks, query, "%"+queryText+"%")
	if err != nil {
		return nil, fmt.Errorf("failed to get feedback by query: %w", err)
	}

	return feedbacks, nil
}

// GetFeedbackAnalytics retrieves comprehensive feedback analytics
func (r *FeedbackRepo) GetFeedbackAnalytics(topN int) (*model.FeedbackAnalytics, error) {
	if topN <= 0 {
		topN = 10
	}

	analytics := &model.FeedbackAnalytics{
		ByAgentType: make(map[model.AgentType]int),
	}

	// Get total counts by sentiment
	sentimentQuery := `
		SELECT
			feedback,
			COUNT(*) as count,
			AVG(confidence) as avg_conf
		FROM rag_feedback
		GROUP BY feedback`

	type SentimentCount struct {
		Feedback      model.FeedbackSentiment `db:"feedback"`
		Count         int                     `db:"count"`
		AvgConfidence float64                 `db:"avg_conf"`
	}

	var sentimentCounts []SentimentCount
	err := r.db.Select(&sentimentCounts, sentimentQuery)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to get sentiment counts: %w", err)
	}

	for _, sc := range sentimentCounts {
		analytics.TotalFeedback += sc.Count
		switch sc.Feedback {
		case model.FeedbackSentimentPositive:
			analytics.PositiveCount = sc.Count
		case model.FeedbackSentimentNegative:
			analytics.NegativeCount = sc.Count
		case model.FeedbackSentimentNeutral:
			analytics.NeutralCount = sc.Count
		}
	}

	// Calculate average confidence across all feedback
	if analytics.TotalFeedback > 0 {
		var avgConf sql.NullFloat64
		err = r.db.Get(&avgConf, "SELECT AVG(confidence) FROM rag_feedback")
		if err == nil && avgConf.Valid {
			analytics.AvgConfidence = avgConf.Float64
		}
	}

	// Get counts by agent type
	agentQuery := `
		SELECT agent_type, COUNT(*) as count
		FROM rag_feedback
		GROUP BY agent_type`

	type AgentCount struct {
		AgentType model.AgentType `db:"agent_type"`
		Count     int             `db:"count"`
	}

	var agentCounts []AgentCount
	err = r.db.Select(&agentCounts, agentQuery)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to get agent counts: %w", err)
	}

	for _, ac := range agentCounts {
		analytics.ByAgentType[ac.AgentType] = ac.Count
	}

	// Get top attributes with most feedback
	topAttrs, err := r.GetAttributeFeedbackSummary(topN)
	if err != nil {
		return nil, fmt.Errorf("failed to get top attributes: %w", err)
	}
	analytics.TopAttributes = topAttrs

	// Get recent feedback
	recentFeedback, err := r.GetRecentFeedback(topN)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent feedback: %w", err)
	}
	analytics.RecentFeedback = recentFeedback

	// Get sentiment trend
	summaries, err := r.GetFeedbackSummary()
	if err != nil {
		return nil, fmt.Errorf("failed to get feedback summary: %w", err)
	}
	analytics.SentimentTrend = summaries

	return analytics, nil
}

// GetFeedbackCount returns the total number of feedback entries
func (r *FeedbackRepo) GetFeedbackCount() (int, error) {
	var count int
	err := r.db.Get(&count, "SELECT COUNT(*) FROM rag_feedback")
	if err != nil {
		return 0, fmt.Errorf("failed to get feedback count: %w", err)
	}
	return count, nil
}

// GetFeedbackCountBySentiment returns counts grouped by sentiment
func (r *FeedbackRepo) GetFeedbackCountBySentiment() (map[model.FeedbackSentiment]int, error) {
	query := `
		SELECT feedback, COUNT(*) as count
		FROM rag_feedback
		GROUP BY feedback`

	type SentimentCount struct {
		Feedback model.FeedbackSentiment `db:"feedback"`
		Count    int                     `db:"count"`
	}

	var counts []SentimentCount
	err := r.db.Select(&counts, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get sentiment counts: %w", err)
	}

	result := make(map[model.FeedbackSentiment]int)
	for _, c := range counts {
		result[c.Feedback] = c.Count
	}

	return result, nil
}

// DeleteFeedback deletes a feedback entry by ID
func (r *FeedbackRepo) DeleteFeedback(id int) error {
	query := "DELETE FROM rag_feedback WHERE id = $1"
	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete feedback: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("feedback with id %d not found", id)
	}

	return nil
}

// DeleteOldFeedback deletes feedback entries older than the specified number of days
func (r *FeedbackRepo) DeleteOldFeedback(daysOld int) (int64, error) {
	query := `
		DELETE FROM rag_feedback
		WHERE created_at < NOW() - INTERVAL '1 day' * $1`

	result, err := r.db.Exec(query, daysOld)
	if err != nil {
		return 0, fmt.Errorf("failed to delete old feedback: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rowsAffected, nil
}
