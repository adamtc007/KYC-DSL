-- ===========================================================
-- 007_rag_feedback.sql
-- Feedback Loop for RAG Attribute/Document Relevance
-- ===========================================================

-- Create feedback sentiment enum type
CREATE TYPE feedback_sentiment AS ENUM ('positive', 'negative', 'neutral');

-- Create feedback table for tracking user and AI agent feedback
CREATE TABLE IF NOT EXISTS rag_feedback (
    id SERIAL PRIMARY KEY,
    query_text TEXT NOT NULL,
    attribute_code TEXT,
    document_code TEXT,
    regulation_code TEXT,
    feedback feedback_sentiment DEFAULT 'positive',
    confidence FLOAT DEFAULT 1.0,
    agent_name TEXT,
    agent_type TEXT, -- human, ai, automated
    created_at TIMESTAMP DEFAULT NOW(),

    -- Add foreign key constraints for data integrity
    CONSTRAINT fk_attribute_code
        FOREIGN KEY (attribute_code)
        REFERENCES kyc_attributes(code)
        ON DELETE CASCADE,
    CONSTRAINT fk_document_code
        FOREIGN KEY (document_code)
        REFERENCES kyc_documents(code)
        ON DELETE CASCADE,
    CONSTRAINT fk_regulation_code
        FOREIGN KEY (regulation_code)
        REFERENCES kyc_regulations(code)
        ON DELETE CASCADE,

    -- Ensure at least one entity is provided
    CONSTRAINT check_entity_provided
        CHECK (
            attribute_code IS NOT NULL OR
            document_code IS NOT NULL OR
            regulation_code IS NOT NULL
        )
);

-- Create indexes for performance
CREATE INDEX idx_rag_feedback_query ON rag_feedback(query_text);
CREATE INDEX idx_rag_feedback_attribute ON rag_feedback(attribute_code);
CREATE INDEX idx_rag_feedback_document ON rag_feedback(document_code);
CREATE INDEX idx_rag_feedback_regulation ON rag_feedback(regulation_code);
CREATE INDEX idx_rag_feedback_created_at ON rag_feedback(created_at DESC);
CREATE INDEX idx_rag_feedback_agent_type ON rag_feedback(agent_type);
CREATE INDEX idx_rag_feedback_sentiment ON rag_feedback(feedback);

-- Function to adjust relevance_score in kyc_attr_doc_links
CREATE OR REPLACE FUNCTION update_relevance()
RETURNS trigger AS $$
BEGIN
    -- Update attribute-document link relevance scores
    IF NEW.attribute_code IS NOT NULL OR NEW.document_code IS NOT NULL THEN
        UPDATE kyc_attr_doc_links
        SET relevance_score = GREATEST(0.0, LEAST(1.0,
            CASE
                WHEN NEW.feedback = 'positive' THEN relevance_score + (0.05 * NEW.confidence)
                WHEN NEW.feedback = 'negative' THEN relevance_score - (0.05 * NEW.confidence)
                ELSE relevance_score
            END
        ))
        WHERE (NEW.attribute_code IS NULL OR attribute_code = NEW.attribute_code)
          AND (NEW.document_code IS NULL OR document_code = NEW.document_code);
    END IF;

    -- Future: Could also update document-regulation links if needed
    -- For now, we focus on attribute-document relationships

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger to automatically adjust relevance scores on feedback insert
CREATE TRIGGER trig_feedback_relevance
AFTER INSERT ON rag_feedback
FOR EACH ROW
EXECUTE FUNCTION update_relevance();

-- Create view for feedback analytics
CREATE OR REPLACE VIEW rag_feedback_summary AS
SELECT
    feedback,
    agent_type,
    COUNT(*) as count,
    AVG(confidence) as avg_confidence,
    MIN(created_at) as first_feedback,
    MAX(created_at) as last_feedback
FROM rag_feedback
GROUP BY feedback, agent_type
ORDER BY feedback, agent_type;

-- Create view for attribute feedback summary
CREATE OR REPLACE VIEW attribute_feedback_summary AS
SELECT
    attribute_code,
    feedback,
    COUNT(*) as feedback_count,
    AVG(confidence) as avg_confidence,
    STRING_AGG(DISTINCT agent_type, ', ') as agent_types
FROM rag_feedback
WHERE attribute_code IS NOT NULL
GROUP BY attribute_code, feedback
ORDER BY attribute_code, feedback;

COMMENT ON TABLE rag_feedback IS 'Stores user and AI agent feedback on RAG search results to enable continuous improvement of relevance scores';
COMMENT ON COLUMN rag_feedback.query_text IS 'The original search query that produced the result';
COMMENT ON COLUMN rag_feedback.confidence IS 'Weight factor (0.0-1.0) for how much this feedback should impact relevance scores';
COMMENT ON COLUMN rag_feedback.agent_type IS 'Type of agent providing feedback: human, ai, or automated';
COMMENT ON FUNCTION update_relevance() IS 'Automatically adjusts relevance scores in kyc_attr_doc_links based on feedback';
