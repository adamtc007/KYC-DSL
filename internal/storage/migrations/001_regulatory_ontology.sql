-- ===========================================================
-- 001_regulatory_ontology.sql
-- Regulatory Data Ontology Core Schema
-- ===========================================================

-- Regulations: Laws, directives, and regulatory frameworks
CREATE TABLE IF NOT EXISTS kyc_regulations (
    id SERIAL PRIMARY KEY,
    code TEXT UNIQUE NOT NULL,
    name TEXT NOT NULL,
    jurisdiction TEXT,
    authority TEXT,
    description TEXT,
    effective_from DATE,
    effective_to DATE,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Documents: Evidence types that prove compliance attributes
CREATE TABLE IF NOT EXISTS kyc_documents (
    id SERIAL PRIMARY KEY,
    code TEXT UNIQUE NOT NULL,
    name TEXT NOT NULL,
    domain TEXT,                -- e.g. "Entity", "Tax", "Ownership", "Control"
    jurisdiction TEXT,
    regulation_code TEXT REFERENCES kyc_regulations(code),
    source_type TEXT CHECK (source_type IN ('Official', 'Client', 'Operational')),
    validity_years INT,
    description TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Attributes: Data points required for compliance
CREATE TABLE IF NOT EXISTS kyc_attributes (
    id SERIAL PRIMARY KEY,
    code TEXT UNIQUE NOT NULL,
    name TEXT NOT NULL,
    domain TEXT,
    description TEXT,
    risk_category TEXT,
    is_personal_data BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Links attributes to documents that can evidence them
CREATE TABLE IF NOT EXISTS kyc_attr_doc_links (
    id SERIAL PRIMARY KEY,
    attribute_code TEXT REFERENCES kyc_attributes(code),
    document_code TEXT REFERENCES kyc_documents(code),
    source_tier TEXT CHECK (source_tier IN ('Primary','Secondary','Tertiary')),
    is_mandatory BOOLEAN DEFAULT TRUE,
    jurisdiction TEXT,
    regulation_code TEXT,
    notes TEXT
);

-- Links documents to regulations that require them
CREATE TABLE IF NOT EXISTS kyc_doc_reg_links (
    id SERIAL PRIMARY KEY,
    document_code TEXT REFERENCES kyc_documents(code),
    regulation_code TEXT REFERENCES kyc_regulations(code),
    applicability TEXT,          -- EntityType / Product / RiskLevel etc.
    jurisdiction TEXT
);

-- Create indexes for common queries
CREATE INDEX IF NOT EXISTS idx_kyc_documents_regulation ON kyc_documents(regulation_code);
CREATE INDEX IF NOT EXISTS idx_kyc_documents_domain ON kyc_documents(domain);
CREATE INDEX IF NOT EXISTS idx_kyc_documents_jurisdiction ON kyc_documents(jurisdiction);
CREATE INDEX IF NOT EXISTS idx_kyc_attributes_domain ON kyc_attributes(domain);
CREATE INDEX IF NOT EXISTS idx_kyc_attr_doc_links_attr ON kyc_attr_doc_links(attribute_code);
CREATE INDEX IF NOT EXISTS idx_kyc_attr_doc_links_doc ON kyc_attr_doc_links(document_code);
CREATE INDEX IF NOT EXISTS idx_kyc_doc_reg_links_doc ON kyc_doc_reg_links(document_code);
CREATE INDEX IF NOT EXISTS idx_kyc_doc_reg_links_reg ON kyc_doc_reg_links(regulation_code);
