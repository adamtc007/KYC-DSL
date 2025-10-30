-- ===========================================================
-- 010_rag_enhancements.sql
-- RAG System Enhancements: Feedback Loop, Sections, Clusters, Audit
-- ===========================================================

-- ==================== Enhancement A: Feedback Loop ====================
-- Self-tuning RAG system with structured agent feedback

CREATE TYPE feedback_type AS ENUM ('positive', 'negative');

CREATE TABLE IF NOT EXISTS rag_feedback (
    id SERIAL PRIMARY KEY,
    query_text TEXT NOT NULL,
    attribute_code TEXT REFERENCES kyc_attributes(code),
    document_code TEXT REFERENCES kyc_documents(code),
    regulation_code TEXT REFERENCES kyc_regulations(code),
    feedback feedback_type DEFAULT 'positive',
    agent_name TEXT,
    session_id TEXT,
    relevance_score FLOAT,
    notes TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Indexes for feedback queries
CREATE INDEX IF NOT EXISTS idx_feedback_query ON rag_feedback(query_text);
CREATE INDEX IF NOT EXISTS idx_feedback_attribute ON rag_feedback(attribute_code);
CREATE INDEX IF NOT EXISTS idx_feedback_agent ON rag_feedback(agent_name);
CREATE INDEX IF NOT EXISTS idx_feedback_created ON rag_feedback(created_at DESC);

-- Trigger function to automatically update relevance scores based on feedback
CREATE OR REPLACE FUNCTION update_relevance_from_feedback()
RETURNS trigger AS $$
BEGIN
    -- Update attribute-document links based on feedback
    UPDATE kyc_attr_doc_links
    SET relevance_score = CASE
        WHEN NEW.feedback = 'positive' THEN
            LEAST(relevance_score + 0.1, 1.0)  -- Cap at 1.0
        ELSE
            GREATEST(relevance_score - 0.1, 0.0)  -- Floor at 0.0
    END
    WHERE (NEW.attribute_code IS NOT NULL AND attribute_code = NEW.attribute_code)
       OR (NEW.document_code IS NOT NULL AND document_code = NEW.document_code)
       OR (NEW.regulation_code IS NOT NULL AND regulation_code = NEW.regulation_code);

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger to fire after feedback insertion
CREATE TRIGGER trig_feedback_relevance
AFTER INSERT ON rag_feedback
FOR EACH ROW
EXECUTE FUNCTION update_relevance_from_feedback();

COMMENT ON TABLE rag_feedback IS
    'Agent feedback on RAG retrieval quality for self-tuning system';

COMMENT ON COLUMN rag_feedback.feedback IS
    'Positive feedback increases relevance scores, negative decreases them';

COMMENT ON FUNCTION update_relevance_from_feedback IS
    'Automatically adjusts relevance_score in kyc_attr_doc_links based on agent feedback';

-- ==================== Enhancement C: Snippet-Level Retrieval ====================
-- Fine-grained document section embeddings for precise retrieval

CREATE TABLE IF NOT EXISTS kyc_document_sections (
    id SERIAL PRIMARY KEY,
    document_code TEXT NOT NULL REFERENCES kyc_documents(code) ON DELETE CASCADE,
    section_number TEXT,
    section_title TEXT,
    text_excerpt TEXT NOT NULL,
    page_number INT,
    embedding vector(1536),
    created_at TIMESTAMP DEFAULT NOW()
);

-- Vector index for section-level semantic search
CREATE INDEX IF NOT EXISTS idx_doc_sections_embedding
    ON kyc_document_sections
    USING ivfflat (embedding vector_cosine_ops)
    WITH (lists = 100);

-- Index for document lookups
CREATE INDEX IF NOT EXISTS idx_doc_sections_document
    ON kyc_document_sections(document_code);

-- Full-text search index for text excerpts
CREATE INDEX IF NOT EXISTS idx_doc_sections_text
    ON kyc_document_sections USING GIN(to_tsvector('english', text_excerpt));

COMMENT ON TABLE kyc_document_sections IS
    'Document sections with embeddings for fine-grained retrieval (e.g., specific FATCA clauses)';

COMMENT ON COLUMN kyc_document_sections.text_excerpt IS
    'Actual text content of the section (typically 1-3 paragraphs)';

-- ==================== Enhancement D: Semantic Clusters ====================
-- Pre-computed attribute clusters for fast and precise retrieval

CREATE TABLE IF NOT EXISTS rag_clusters (
    id SERIAL PRIMARY KEY,
    cluster_code TEXT UNIQUE NOT NULL,
    cluster_name TEXT NOT NULL,
    description TEXT,
    centroid vector(1536),
    member_attribute_codes TEXT[] NOT NULL,
    member_count INT GENERATED ALWAYS AS (array_length(member_attribute_codes, 1)) STORED,
    quality_score FLOAT DEFAULT 0.0,
    last_computed TIMESTAMP DEFAULT NOW(),
    created_at TIMESTAMP DEFAULT NOW()
);

-- Vector index for cluster centroids
CREATE INDEX IF NOT EXISTS idx_clusters_centroid
    ON rag_clusters
    USING ivfflat (centroid vector_cosine_ops)
    WITH (lists = 50);

-- GIN index for member lookups
CREATE INDEX IF NOT EXISTS idx_clusters_members
    ON rag_clusters USING GIN(member_attribute_codes);

-- Index for cluster lookups
CREATE INDEX IF NOT EXISTS idx_clusters_code
    ON rag_clusters(cluster_code);

COMMENT ON TABLE rag_clusters IS
    'Semantic clusters of related attributes with centroid embeddings for fast search';

COMMENT ON COLUMN rag_clusters.centroid IS
    'Average embedding vector of all member attributes (computed nightly)';

COMMENT ON COLUMN rag_clusters.member_attribute_codes IS
    'Array of attribute codes belonging to this cluster';

COMMENT ON COLUMN rag_clusters.quality_score IS
    'Cluster cohesion score (0.0-1.0, higher = more cohesive)';

-- ==================== Enhancement E: RAG Audit Trail ====================
-- Complete audit log for regulatory compliance and explainability

CREATE TABLE IF NOT EXISTS rag_audit_log (
    id SERIAL PRIMARY KEY,
    query_text TEXT NOT NULL,
    query_embedding vector(1536),
    response JSONB NOT NULL,
    result_count INT,
    agent_name TEXT,
    session_id TEXT,
    endpoint TEXT,
    latency_ms INT,
    error_message TEXT,
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Indexes for audit queries
CREATE INDEX IF NOT EXISTS idx_audit_query ON rag_audit_log(query_text);
CREATE INDEX IF NOT EXISTS idx_audit_agent ON rag_audit_log(agent_name);
CREATE INDEX IF NOT EXISTS idx_audit_session ON rag_audit_log(session_id);
CREATE INDEX IF NOT EXISTS idx_audit_created ON rag_audit_log(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_endpoint ON rag_audit_log(endpoint);

-- GIN index for JSONB response analysis
CREATE INDEX IF NOT EXISTS idx_audit_response
    ON rag_audit_log USING GIN(response);

-- Partial index for errors only
CREATE INDEX IF NOT EXISTS idx_audit_errors
    ON rag_audit_log(created_at DESC)
    WHERE error_message IS NOT NULL;

COMMENT ON TABLE rag_audit_log IS
    'Complete audit trail of all RAG queries for compliance, debugging, and model improvement';

COMMENT ON COLUMN rag_audit_log.response IS
    'Full JSON response returned to caller (for explainability)';

COMMENT ON COLUMN rag_audit_log.latency_ms IS
    'Query execution time in milliseconds (for performance monitoring)';

COMMENT ON COLUMN rag_audit_log.query_embedding IS
    'Stored query embedding for similarity analysis of common queries';

-- ==================== Views ====================

-- View: Feedback statistics by attribute
CREATE OR REPLACE VIEW feedback_stats_by_attribute AS
SELECT
    attribute_code,
    COUNT(*) as total_feedback,
    COUNT(*) FILTER (WHERE feedback = 'positive') as positive_count,
    COUNT(*) FILTER (WHERE feedback = 'negative') as negative_count,
    ROUND(100.0 * COUNT(*) FILTER (WHERE feedback = 'positive') / COUNT(*), 2) as positive_pct,
    AVG(relevance_score) as avg_relevance,
    MAX(created_at) as last_feedback
FROM rag_feedback
WHERE attribute_code IS NOT NULL
GROUP BY attribute_code
ORDER BY total_feedback DESC;

-- View: Popular queries from audit log
CREATE OR REPLACE VIEW popular_queries AS
SELECT
    query_text,
    COUNT(*) as query_count,
    AVG(latency_ms) as avg_latency_ms,
    AVG(result_count) as avg_results,
    MAX(created_at) as last_queried
FROM rag_audit_log
WHERE error_message IS NULL
GROUP BY query_text
HAVING COUNT(*) > 1
ORDER BY query_count DESC;

-- View: Agent performance metrics
CREATE OR REPLACE VIEW agent_performance AS
SELECT
    agent_name,
    COUNT(DISTINCT session_id) as sessions,
    COUNT(*) as total_queries,
    AVG(latency_ms) as avg_latency_ms,
    COUNT(*) FILTER (WHERE error_message IS NOT NULL) as error_count,
    ROUND(100.0 * COUNT(*) FILTER (WHERE error_message IS NULL) / COUNT(*), 2) as success_rate
FROM rag_audit_log
WHERE agent_name IS NOT NULL
GROUP BY agent_name
ORDER BY total_queries DESC;

-- View: Cluster membership details
CREATE OR REPLACE VIEW cluster_details AS
SELECT
    c.cluster_code,
    c.cluster_name,
    c.description,
    c.member_count,
    c.quality_score,
    c.last_computed,
    a.code as attribute_code,
    a.name as attribute_name,
    am.risk_level
FROM rag_clusters c
CROSS JOIN LATERAL unnest(c.member_attribute_codes) as a_code
JOIN kyc_attributes a ON a.code = a_code
LEFT JOIN kyc_attribute_metadata am ON am.attribute_code = a.code
ORDER BY c.cluster_code, a.code;

-- View: Document sections with parent context
CREATE OR REPLACE VIEW document_section_context AS
SELECT
    s.id as section_id,
    s.section_number,
    s.section_title,
    s.text_excerpt,
    s.page_number,
    d.code as document_code,
    d.title as document_title,
    d.jurisdiction,
    d.doc_type,
    r.code as regulation_code,
    r.title as regulation_title
FROM kyc_document_sections s
JOIN kyc_documents d ON d.code = s.document_code
LEFT JOIN kyc_regulations r ON r.code = d.regulation_code
ORDER BY d.code, s.section_number;

-- ==================== Functions ====================

-- Function: Record audit log entry
CREATE OR REPLACE FUNCTION log_rag_query(
    p_query TEXT,
    p_response JSONB,
    p_agent TEXT DEFAULT 'system',
    p_latency_ms INT DEFAULT NULL,
    p_endpoint TEXT DEFAULT NULL
) RETURNS INT AS $$
DECLARE
    v_log_id INT;
BEGIN
    INSERT INTO rag_audit_log (query_text, response, agent_name, latency_ms, endpoint, result_count)
    VALUES (
        p_query,
        p_response,
        p_agent,
        p_latency_ms,
        p_endpoint,
        COALESCE((p_response->>'count')::INT, 0)
    )
    RETURNING id INTO v_log_id;

    RETURN v_log_id;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION log_rag_query IS
    'Convenience function to log RAG queries from application code';

-- Function: Get cluster recommendations for a query embedding
CREATE OR REPLACE FUNCTION recommend_clusters(
    p_embedding vector(1536),
    p_limit INT DEFAULT 3
) RETURNS TABLE (
    cluster_code TEXT,
    cluster_name TEXT,
    similarity FLOAT,
    member_count INT
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        c.cluster_code,
        c.cluster_name,
        1 - (c.centroid <=> p_embedding) as similarity,
        c.member_count
    FROM rag_clusters c
    WHERE c.centroid IS NOT NULL
    ORDER BY c.centroid <=> p_embedding
    LIMIT p_limit;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION recommend_clusters IS
    'Find the most relevant clusters for a given query embedding';

-- Function: Search document sections by vector
CREATE OR REPLACE FUNCTION search_document_sections(
    p_embedding vector(1536),
    p_limit INT DEFAULT 10
) RETURNS TABLE (
    section_id INT,
    document_code TEXT,
    section_title TEXT,
    text_excerpt TEXT,
    similarity FLOAT
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        s.id,
        s.document_code,
        s.section_title,
        s.text_excerpt,
        1 - (s.embedding <=> p_embedding) as similarity
    FROM kyc_document_sections s
    WHERE s.embedding IS NOT NULL
    ORDER BY s.embedding <=> p_embedding
    LIMIT p_limit;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION search_document_sections IS
    'Semantic search on document sections for fine-grained retrieval';

-- ==================== Indexes for Performance ====================

-- Composite index for feedback analysis
CREATE INDEX IF NOT EXISTS idx_feedback_composite
    ON rag_feedback(attribute_code, feedback, created_at DESC);

-- Composite index for audit analysis
CREATE INDEX IF NOT EXISTS idx_audit_composite
    ON rag_audit_log(agent_name, created_at DESC)
    WHERE error_message IS NULL;

-- ==================== Sample Data for Testing ====================

-- Insert sample clusters (will be recomputed by nightly job)
INSERT INTO rag_clusters (cluster_code, cluster_name, description, member_attribute_codes) VALUES
('TAX_COMPLIANCE', 'Tax & Reporting Compliance', 'Attributes related to tax residency, FATCA, CRS', ARRAY['TAX_RESIDENCY_COUNTRY', 'FATCA_STATUS', 'CRS_CLASSIFICATION']),
('OWNERSHIP_CONTROL', 'Ownership & Control', 'UBO identification and ownership structure', ARRAY['UBO_NAME', 'UBO_OWNERSHIP_PERCENT', 'DIRECTOR_NAME']),
('RISK_SCREENING', 'Risk Assessment & Screening', 'PEP, sanctions, adverse media screening', ARRAY['PEP_STATUS', 'SANCTIONS_SCREENING_STATUS', 'ADVERSE_MEDIA_FLAG', 'CUSTOMER_RISK_RATING']),
('ENTITY_ID', 'Entity Identification', 'Legal entity identification and registration', ARRAY['REGISTERED_NAME', 'INCORPORATION_COUNTRY', 'REGISTERED_ADDRESS', 'BUSINESS_ACTIVITY'])
ON CONFLICT (cluster_code) DO NOTHING;

-- ==================== Maintenance Functions ====================

-- Function: Clean old audit logs (retention policy)
CREATE OR REPLACE FUNCTION cleanup_old_audit_logs(
    p_retention_days INT DEFAULT 90
) RETURNS INT AS $$
DECLARE
    v_deleted_count INT;
BEGIN
    DELETE FROM rag_audit_log
    WHERE created_at < NOW() - (p_retention_days || ' days')::INTERVAL
      AND error_message IS NULL;  -- Keep errors longer

    GET DIAGNOSTICS v_deleted_count = ROW_COUNT;
    RETURN v_deleted_count;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION cleanup_old_audit_logs IS
    'Delete audit logs older than specified days (default 90, keeps errors)';

-- ==================== Grants & Permissions ====================
-- (Adjust based on your application user)

-- GRANT SELECT ON feedback_stats_by_attribute TO app_user;
-- GRANT SELECT ON popular_queries TO app_user;
-- GRANT SELECT ON agent_performance TO app_user;
-- GRANT INSERT ON rag_feedback TO app_user;
-- GRANT INSERT ON rag_audit_log TO app_user;
