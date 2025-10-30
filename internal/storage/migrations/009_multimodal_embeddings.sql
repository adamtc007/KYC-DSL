-- ===========================================================
-- 009_multimodal_embeddings.sql
-- Add Vector Embeddings to Documents and Regulations
-- Enables multi-modal RAG across attributes, documents, and regulations
-- ===========================================================

-- Add embedding columns to existing tables
ALTER TABLE kyc_documents
ADD COLUMN IF NOT EXISTS embedding vector(1536),
ADD COLUMN IF NOT EXISTS doc_type TEXT,
ADD COLUMN IF NOT EXISTS title TEXT;

ALTER TABLE kyc_regulations
ADD COLUMN IF NOT EXISTS embedding vector(1536),
ADD COLUMN IF NOT EXISTS title TEXT,
ADD COLUMN IF NOT EXISTS region TEXT,
ADD COLUMN IF NOT EXISTS citation TEXT,
ADD COLUMN IF NOT EXISTS summary TEXT;

-- Update kyc_attr_doc_links to support relevance scoring
ALTER TABLE kyc_attr_doc_links
ADD COLUMN IF NOT EXISTS relevance_score FLOAT DEFAULT 1.0;

-- Create vector indexes for document embeddings
CREATE INDEX IF NOT EXISTS idx_documents_embedding
    ON kyc_documents
    USING ivfflat (embedding vector_cosine_ops)
    WITH (lists = 100);

-- Create vector indexes for regulation embeddings
CREATE INDEX IF NOT EXISTS idx_regulations_embedding
    ON kyc_regulations
    USING ivfflat (embedding vector_cosine_ops)
    WITH (lists = 100);

-- Add indexes for the link table
CREATE INDEX IF NOT EXISTS idx_attrdoc_relevance
    ON kyc_attr_doc_links(relevance_score DESC);

-- Comments
COMMENT ON COLUMN kyc_documents.embedding IS
    'Vector embedding of document description for semantic search (1536 dimensions)';

COMMENT ON COLUMN kyc_regulations.embedding IS
    'Vector embedding of regulation summary for semantic search (1536 dimensions)';

COMMENT ON COLUMN kyc_documents.doc_type IS
    'Document type classification (e.g., Form, Certificate, Declaration, Statement)';

COMMENT ON COLUMN kyc_documents.title IS
    'Human-readable title of the document';

COMMENT ON COLUMN kyc_regulations.title IS
    'Full title of the regulation or directive';

COMMENT ON COLUMN kyc_regulations.region IS
    'Geographic region or jurisdiction (e.g., EU, US, APAC, GLOBAL)';

COMMENT ON COLUMN kyc_regulations.citation IS
    'Official legal citation or reference number';

COMMENT ON COLUMN kyc_regulations.summary IS
    'Executive summary or key provisions of the regulation';

COMMENT ON COLUMN kyc_attr_doc_links.relevance_score IS
    'Relevance score between attribute and document (0.0-1.0, higher = more relevant)';

-- View: Multi-modal search results combining attributes, documents, and regulations
CREATE OR REPLACE VIEW multimodal_attribute_view AS
SELECT
    a.code as attribute_code,
    a.name as attribute_name,
    a.domain as attribute_domain,
    am.business_context,
    am.risk_level,
    am.embedding as attribute_embedding,
    d.code as document_code,
    d.title as document_title,
    d.doc_type,
    d.jurisdiction as document_jurisdiction,
    d.description as document_description,
    d.embedding as document_embedding,
    r.code as regulation_code,
    r.title as regulation_title,
    r.citation,
    r.summary as regulation_summary,
    r.region,
    r.embedding as regulation_embedding,
    adl.relevance_score
FROM kyc_attributes a
LEFT JOIN kyc_attribute_metadata am ON am.attribute_code = a.code
LEFT JOIN kyc_attr_doc_links adl ON adl.attribute_code = a.code
LEFT JOIN kyc_documents d ON d.code = adl.document_code
LEFT JOIN kyc_regulations r ON r.code = adl.regulation_code
ORDER BY a.code, adl.relevance_score DESC NULLS LAST;

COMMENT ON VIEW multimodal_attribute_view IS
    'Combined view of attributes with linked documents and regulations for multi-modal RAG queries';
