package model

import "time"

// FeedbackSentiment represents the type of feedback (positive, negative, neutral)
// Extends the existing FeedbackType with neutral option
type FeedbackSentiment string

const (
	// Reuse existing positive/negative, add neutral
	FeedbackSentimentPositive FeedbackSentiment = "positive"
	FeedbackSentimentNegative FeedbackSentiment = "negative"
	FeedbackSentimentNeutral  FeedbackSentiment = "neutral"
)

// AgentType represents the type of agent providing feedback
type AgentType string

const (
	AgentTypeHuman     AgentType = "human"
	AgentTypeAI        AgentType = "ai"
	AgentTypeAutomated AgentType = "automated"
)

// Feedback represents user or AI agent feedback on RAG search results
// This is the enhanced version with confidence weighting and agent types
type Feedback struct {
	ID             int               `db:"id" json:"id"`
	QueryText      string            `db:"query_text" json:"query_text"`
	AttributeCode  *string           `db:"attribute_code" json:"attribute_code,omitempty"`
	DocumentCode   *string           `db:"document_code" json:"document_code,omitempty"`
	RegulationCode *string           `db:"regulation_code" json:"regulation_code,omitempty"`
	Feedback       FeedbackSentiment `db:"feedback" json:"feedback"`
	Confidence     float64           `db:"confidence" json:"confidence"`
	AgentName      *string           `db:"agent_name" json:"agent_name,omitempty"`
	AgentType      AgentType         `db:"agent_type" json:"agent_type"`
	CreatedAt      time.Time         `db:"created_at" json:"created_at"`
}

// FeedbackSummary represents aggregated feedback statistics
type FeedbackSummary struct {
	Feedback      FeedbackSentiment `db:"feedback" json:"feedback"`
	AgentType     AgentType         `db:"agent_type" json:"agent_type"`
	Count         int               `db:"count" json:"count"`
	AvgConfidence float64           `db:"avg_confidence" json:"avg_confidence"`
	FirstFeedback time.Time         `db:"first_feedback" json:"first_feedback"`
	LastFeedback  time.Time         `db:"last_feedback" json:"last_feedback"`
}

// AttributeFeedbackSummary represents feedback statistics per attribute
type AttributeFeedbackSummary struct {
	AttributeCode string            `db:"attribute_code" json:"attribute_code"`
	Feedback      FeedbackSentiment `db:"feedback" json:"feedback"`
	FeedbackCount int               `db:"feedback_count" json:"feedback_count"`
	AvgConfidence float64           `db:"avg_confidence" json:"avg_confidence"`
	AgentTypes    string            `db:"agent_types" json:"agent_types"`
}

// FeedbackSubmitRequest represents an incoming feedback submission
// This replaces the simple FeedbackRequest with enhanced version
type FeedbackSubmitRequest struct {
	QueryText      string            `json:"query_text" binding:"required"`
	AttributeCode  *string           `json:"attribute_code,omitempty"`
	DocumentCode   *string           `json:"document_code,omitempty"`
	RegulationCode *string           `json:"regulation_code,omitempty"`
	Feedback       FeedbackSentiment `json:"feedback"`
	Confidence     float64           `json:"confidence"`
	AgentName      *string           `json:"agent_name,omitempty"`
	AgentType      AgentType         `json:"agent_type"`
}

// FeedbackResponse represents the response after submitting feedback
type FeedbackResponse struct {
	Status    string            `json:"status"`
	ID        int               `json:"id"`
	Feedback  FeedbackSentiment `json:"feedback"`
	AgentName *string           `json:"agent_name,omitempty"`
	CreatedAt time.Time         `json:"created_at"`
}

// RecentFeedbackResponse represents a list of recent feedback entries
type RecentFeedbackResponse struct {
	Count     int        `json:"count"`
	Feedbacks []Feedback `json:"feedbacks"`
}

// FeedbackAnalytics represents detailed analytics on feedback patterns
type FeedbackAnalytics struct {
	TotalFeedback  int                        `json:"total_feedback"`
	PositiveCount  int                        `json:"positive_count"`
	NegativeCount  int                        `json:"negative_count"`
	NeutralCount   int                        `json:"neutral_count"`
	ByAgentType    map[AgentType]int          `json:"by_agent_type"`
	AvgConfidence  float64                    `json:"avg_confidence"`
	TopAttributes  []AttributeFeedbackSummary `json:"top_attributes"`
	RecentFeedback []Feedback                 `json:"recent_feedback"`
	SentimentTrend []FeedbackSummary          `json:"sentiment_trend"`
}
