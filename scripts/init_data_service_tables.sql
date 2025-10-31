-- ============================================================================
-- KYC Data Service Database Schema
-- ============================================================================
-- This script initializes the tables required by the Data Service gRPC API
-- Safe to run multiple times (idempotent)
-- ============================================================================

BEGIN;

-- ============================================================================
-- Case Versions Table
-- ============================================================================
-- Stores all versions of KYC cases with DSL source and compiled JSON
CREATE TABLE IF NOT EXISTS case_versions (
    id SERIAL PRIMARY KEY,
    case_id VARCHAR(255) NOT NULL,
    dsl_source TEXT NOT NULL,
    compiled_json TEXT,
    status VARCHAR(50) NOT NULL DEFAULT 'draft',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for case_versions
CREATE INDEX IF NOT EXISTS idx_case_versions_case_id ON case_versions(case_id);
CREATE INDEX IF NOT EXISTS idx_case_versions_status ON case_versions(status);
CREATE INDEX IF NOT EXISTS idx_case_versions_created_at ON case_versions(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_case_versions_case_id_created ON case_versions(case_id, created_at DESC);

-- ============================================================================
-- Dictionary Tables (if not already created by ontology scripts)
-- ============================================================================

-- KYC Attributes Table
CREATE TABLE IF NOT EXISTS kyc_attributes (
    id SERIAL PRIMARY KEY,
    attribute_code VARCHAR(100) UNIQUE NOT NULL,
    attribute_name VARCHAR(255) NOT NULL,
    description TEXT,
    attribute_type VARCHAR(50),
    jurisdiction VARCHAR(10),
    regulation_code VARCHAR(50),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for kyc_attributes
CREATE INDEX IF NOT EXISTS idx_kyc_attributes_code ON kyc_attributes(attribute_code);
CREATE INDEX IF NOT EXISTS idx_kyc_attributes_jurisdiction ON kyc_attributes(jurisdiction);
CREATE INDEX IF NOT EXISTS idx_kyc_attributes_regulation ON kyc_attributes(regulation_code);

-- KYC Documents Table
CREATE TABLE IF NOT EXISTS kyc_documents (
    id SERIAL PRIMARY KEY,
    document_code VARCHAR(100) UNIQUE NOT NULL,
    document_name VARCHAR(255) NOT NULL,
    jurisdiction VARCHAR(10),
    category VARCHAR(100),
    description TEXT,
    reference_url TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for kyc_documents
CREATE INDEX IF NOT EXISTS idx_kyc_documents_code ON kyc_documents(document_code);
CREATE INDEX IF NOT EXISTS idx_kyc_documents_jurisdiction ON kyc_documents(jurisdiction);
CREATE INDEX IF NOT EXISTS idx_kyc_documents_category ON kyc_documents(category);

-- ============================================================================
-- Trigger Functions for Updated Timestamps
-- ============================================================================

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Apply trigger to case_versions
DROP TRIGGER IF EXISTS update_case_versions_updated_at ON case_versions;
CREATE TRIGGER update_case_versions_updated_at
    BEFORE UPDATE ON case_versions
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Apply trigger to kyc_attributes
DROP TRIGGER IF EXISTS update_kyc_attributes_updated_at ON kyc_attributes;
CREATE TRIGGER update_kyc_attributes_updated_at
    BEFORE UPDATE ON kyc_attributes
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Apply trigger to kyc_documents
DROP TRIGGER IF EXISTS update_kyc_documents_updated_at ON kyc_documents;
CREATE TRIGGER update_kyc_documents_updated_at
    BEFORE UPDATE ON kyc_documents
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- ============================================================================
-- Sample Data (if tables are empty)
-- ============================================================================

-- Sample attributes
INSERT INTO kyc_attributes (attribute_code, attribute_name, description, attribute_type, jurisdiction, regulation_code)
VALUES
    ('CLIENT_NAME', 'Client Name', 'Legal name of the client entity', 'text', 'GLOBAL', 'GENERAL'),
    ('CLIENT_LEI', 'Legal Entity Identifier', 'LEI code for the client', 'text', 'GLOBAL', 'FATCA'),
    ('UBO_NAME', 'Ultimate Beneficial Owner Name', 'Name of the UBO', 'text', 'GLOBAL', 'AMLD5'),
    ('UBO_PERCENT', 'UBO Ownership Percentage', 'Percentage ownership by UBO', 'number', 'GLOBAL', 'AMLD5'),
    ('TAX_ID', 'Tax Identification Number', 'Tax ID or TIN', 'text', 'US', 'FATCA')
ON CONFLICT (attribute_code) DO NOTHING;

-- Sample documents
INSERT INTO kyc_documents (document_code, document_name, jurisdiction, category, description, reference_url)
VALUES
    ('DOC_PASSPORT', 'Passport', 'GLOBAL', 'Identity', 'Government-issued passport', NULL),
    ('DOC_CERT_INC', 'Certificate of Incorporation', 'GLOBAL', 'Entity', 'Official certificate of company registration', NULL),
    ('DOC_W9', 'IRS Form W-9', 'US', 'Tax', 'Request for Taxpayer Identification Number', 'https://www.irs.gov/forms-pubs/about-form-w-9'),
    ('DOC_W8BEN', 'IRS Form W-8BEN', 'US', 'Tax', 'Certificate of Foreign Status for Individuals', 'https://www.irs.gov/forms-pubs/about-form-w-8-ben'),
    ('DOC_UTILITY_BILL', 'Utility Bill', 'GLOBAL', 'Address', 'Proof of address', NULL)
ON CONFLICT (document_code) DO NOTHING;

COMMIT;

-- ============================================================================
-- Verification Queries
-- ============================================================================
-- Run these to verify the setup:
--
-- SELECT COUNT(*) FROM case_versions;
-- SELECT COUNT(*) FROM kyc_attributes;
-- SELECT COUNT(*) FROM kyc_documents;
--
-- SELECT * FROM kyc_attributes ORDER BY attribute_code;
-- SELECT * FROM kyc_documents ORDER BY document_code;
-- ============================================================================
