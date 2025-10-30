-- ===========================================================
-- 006_attribute_metadata.sql
-- Attribute Metadata & Clustering for Rich Semantics
-- Enables RAG, vector search, and agent optimization
-- ===========================================================

-- Enable pgvector extension for semantic search
CREATE EXTENSION IF NOT EXISTS vector;

-- ==================== Attribute Metadata ====================

CREATE TABLE IF NOT EXISTS kyc_attribute_metadata (
    id SERIAL PRIMARY KEY,
    attribute_code TEXT UNIQUE NOT NULL REFERENCES kyc_attributes(code) ON DELETE CASCADE,
    synonyms TEXT[],                    -- Alternative names for this attribute
    data_type TEXT,                     -- string, integer, float, boolean, date, array
    domain_values TEXT[],               -- Enumerated valid values (if applicable)
    validation_pattern TEXT,            -- Regex pattern for validation
    risk_level TEXT CHECK (risk_level IN ('LOW', 'MEDIUM', 'HIGH', 'CRITICAL')),
    example_values TEXT[],              -- Example valid values
    regulatory_citations TEXT[],        -- References to regulations (e.g., "AMLD5 Article 13")
    business_glossary_url TEXT,         -- Link to external documentation
    business_context TEXT,              -- Rich business definition for embedding generation
    data_sensitivity TEXT CHECK (data_sensitivity IN ('PUBLIC', 'INTERNAL', 'CONFIDENTIAL', 'RESTRICTED')),
    retention_period_days INT,          -- How long to retain data
    embedding vector(1536),             -- OpenAI text-embedding-3-large vector
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Create indexes for common queries
CREATE INDEX IF NOT EXISTS idx_attribute_metadata_code
    ON kyc_attribute_metadata(attribute_code);

CREATE INDEX IF NOT EXISTS idx_attribute_metadata_risk
    ON kyc_attribute_metadata(risk_level);

CREATE INDEX IF NOT EXISTS idx_attribute_metadata_sensitivity
    ON kyc_attribute_metadata(data_sensitivity);

-- GIN index for array searches (synonyms, domain values)
CREATE INDEX IF NOT EXISTS idx_attribute_metadata_synonyms
    ON kyc_attribute_metadata USING GIN(synonyms);

CREATE INDEX IF NOT EXISTS idx_attribute_metadata_domain
    ON kyc_attribute_metadata USING GIN(domain_values);

-- Vector index for semantic search (IVFFlat for cosine similarity)
CREATE INDEX IF NOT EXISTS idx_attribute_metadata_embedding
    ON kyc_attribute_metadata USING ivfflat (embedding vector_cosine_ops)
    WITH (lists = 100);

-- ==================== Attribute Clusters ====================
-- Logical groupings of attributes for agent optimization

CREATE TABLE IF NOT EXISTS kyc_attribute_clusters (
    id SERIAL PRIMARY KEY,
    cluster_code TEXT UNIQUE NOT NULL,
    cluster_name TEXT NOT NULL,
    attribute_codes TEXT[] NOT NULL,
    description TEXT,
    use_case TEXT,                      -- e.g., "Entity Identification", "Risk Assessment"
    priority INT DEFAULT 5,             -- 1=critical, 5=standard, 10=optional
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_attribute_clusters_code
    ON kyc_attribute_clusters(cluster_code);

CREATE INDEX IF NOT EXISTS idx_attribute_clusters_priority
    ON kyc_attribute_clusters(priority);

-- GIN index for array searches
CREATE INDEX IF NOT EXISTS idx_attribute_clusters_attrs
    ON kyc_attribute_clusters USING GIN(attribute_codes);

-- ==================== Attribute Relationships ====================
-- Semantic relationships between attributes (for knowledge graph)

CREATE TABLE IF NOT EXISTS kyc_attribute_relationships (
    id SERIAL PRIMARY KEY,
    source_attribute_code TEXT NOT NULL REFERENCES kyc_attributes(code),
    target_attribute_code TEXT NOT NULL REFERENCES kyc_attributes(code),
    relationship_type TEXT NOT NULL,    -- 'derives_from', 'validates', 'requires', 'conflicts_with', 'related_to'
    strength FLOAT DEFAULT 1.0,         -- 0.0 to 1.0, semantic similarity score
    description TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(source_attribute_code, target_attribute_code, relationship_type)
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_attribute_relationships_source
    ON kyc_attribute_relationships(source_attribute_code);

CREATE INDEX IF NOT EXISTS idx_attribute_relationships_target
    ON kyc_attribute_relationships(target_attribute_code);

CREATE INDEX IF NOT EXISTS idx_attribute_relationships_type
    ON kyc_attribute_relationships(relationship_type);

-- ==================== Views ====================

-- View: Complete attribute profile with metadata
CREATE OR REPLACE VIEW attribute_profile AS
SELECT
    a.code,
    a.name,
    a.domain,
    a.description,
    a.risk_category,
    a.is_personal_data,
    a.attribute_class,
    m.synonyms,
    m.data_type,
    m.domain_values,
    m.risk_level,
    m.example_values,
    m.regulatory_citations,
    m.data_sensitivity,
    m.retention_period_days
FROM kyc_attributes a
LEFT JOIN kyc_attribute_metadata m ON m.attribute_code = a.code
ORDER BY a.code;

-- View: Attribute clusters with full attribute details
CREATE OR REPLACE VIEW cluster_details AS
SELECT
    c.cluster_code,
    c.cluster_name,
    c.description as cluster_description,
    c.use_case,
    c.priority,
    c.attribute_codes,
    (SELECT COUNT(*) FROM unnest(c.attribute_codes)) as attribute_count
FROM kyc_attribute_clusters c
ORDER BY c.priority, c.cluster_name;

-- View: Attribute knowledge graph (relationships)
CREATE OR REPLACE VIEW attribute_knowledge_graph AS
SELECT
    r.source_attribute_code,
    sa.name as source_name,
    r.relationship_type,
    r.target_attribute_code,
    ta.name as target_name,
    r.strength,
    r.description
FROM kyc_attribute_relationships r
JOIN kyc_attributes sa ON sa.code = r.source_attribute_code
JOIN kyc_attributes ta ON ta.code = r.target_attribute_code
ORDER BY r.strength DESC;

-- ==================== Functions ====================

-- Function: Find attributes by synonym
CREATE OR REPLACE FUNCTION find_attribute_by_synonym(search_term TEXT)
RETURNS TABLE(attribute_code TEXT, attribute_name TEXT, synonyms TEXT[]) AS $$
BEGIN
    RETURN QUERY
    SELECT a.code, a.name, m.synonyms
    FROM kyc_attributes a
    JOIN kyc_attribute_metadata m ON m.attribute_code = a.code
    WHERE search_term = ANY(m.synonyms)
    ORDER BY a.code;
END;
$$ LANGUAGE plpgsql;

-- Function: Get all attributes in a cluster
CREATE OR REPLACE FUNCTION get_cluster_attributes(cluster_name TEXT)
RETURNS TABLE(attribute_code TEXT, attribute_name TEXT, domain TEXT) AS $$
BEGIN
    RETURN QUERY
    SELECT a.code, a.name, a.domain
    FROM kyc_attributes a
    WHERE a.code = ANY(
        SELECT unnest(attribute_codes)
        FROM kyc_attribute_clusters
        WHERE cluster_code = cluster_name OR kyc_attribute_clusters.cluster_name = cluster_name
    )
    ORDER BY a.code;
END;
$$ LANGUAGE plpgsql;

-- Function: Find related attributes (traverses relationship graph)
CREATE OR REPLACE FUNCTION find_related_attributes(attr_code TEXT, max_depth INT DEFAULT 2)
RETURNS TABLE(
    related_code TEXT,
    related_name TEXT,
    relationship_path TEXT[],
    depth INT
) AS $$
WITH RECURSIVE attr_graph AS (
    -- Base case: direct relationships
    SELECT
        target_attribute_code as related_code,
        ARRAY[source_attribute_code, target_attribute_code] as path,
        1 as depth
    FROM kyc_attribute_relationships
    WHERE source_attribute_code = attr_code

    UNION

    -- Recursive case: follow relationships
    SELECT
        r.target_attribute_code,
        g.path || r.target_attribute_code,
        g.depth + 1
    FROM attr_graph g
    JOIN kyc_attribute_relationships r ON r.source_attribute_code = g.related_code
    WHERE g.depth < max_depth
      AND NOT (r.target_attribute_code = ANY(g.path))  -- Prevent cycles
)
SELECT DISTINCT
    g.related_code,
    a.name as related_name,
    g.path,
    g.depth
FROM attr_graph g
JOIN kyc_attributes a ON a.code = g.related_code
ORDER BY g.depth, g.related_code;
$$ LANGUAGE sql;

-- ==================== COMMENT Statements ====================

COMMENT ON TABLE kyc_attribute_metadata IS
    'Rich metadata for attributes including synonyms, examples, validation rules, and regulatory citations';

COMMENT ON COLUMN kyc_attribute_metadata.synonyms IS
    'Alternative names or terms used for this attribute across different systems or jurisdictions';

COMMENT ON COLUMN kyc_attribute_metadata.domain_values IS
    'Enumerated list of valid values (for categorical attributes)';

COMMENT ON COLUMN kyc_attribute_metadata.regulatory_citations IS
    'References to specific regulations, articles, or guidelines that define this attribute';

COMMENT ON TABLE kyc_attribute_clusters IS
    'Logical groupings of related attributes for agent optimization and focus sets';

COMMENT ON TABLE kyc_attribute_relationships IS
    'Semantic relationships between attributes forming a knowledge graph';

COMMENT ON COLUMN kyc_attribute_relationships.relationship_type IS
    'Type of relationship: derives_from (computed), validates (checks), requires (dependency), conflicts_with (mutually exclusive), related_to (semantic similarity)';

COMMENT ON VIEW attribute_profile IS
    'Complete profile view combining base attribute data with rich metadata';

COMMENT ON VIEW attribute_knowledge_graph IS
    'Graph view of all attribute relationships for knowledge graph queries';

COMMENT ON FUNCTION find_attribute_by_synonym IS
    'Search for attributes by synonym - useful for natural language queries';

COMMENT ON FUNCTION get_cluster_attributes IS
    'Retrieve all attributes belonging to a named cluster';

COMMENT ON FUNCTION find_related_attributes IS
    'Traverse relationship graph to find semantically related attributes (depth-limited)';
