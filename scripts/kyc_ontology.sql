-- ============================================================================
-- KYC-DSL Ontology Schema
-- ============================================================================
-- Comprehensive schema for entities, CBUs, ownership control, and regulatory
-- dictionary with semantic search capabilities
-- ============================================================================

BEGIN;

-- ============================================================================
-- Extensions
-- ============================================================================

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgvector";

-- ============================================================================
-- 1. Core Entities (Legal Entities, CBU, Roles)
-- ============================================================================

-- Entity: Represents any legal entity (company, fund, person, trust, etc.)
CREATE TABLE IF NOT EXISTS entity (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name TEXT NOT NULL,
    entity_type TEXT NOT NULL,               -- COMPANY, FUND, PERSON, PARTNERSHIP, TRUST
    legal_form TEXT,                         -- LLC, PLC, LP, GP, etc.
    jurisdiction TEXT,                       -- US, UK, LU, IE, etc.
    registration_number TEXT,                -- Company registration number
    lei_code TEXT UNIQUE,                    -- Legal Entity Identifier
    incorporation_date DATE,
    dissolution_date DATE,
    status TEXT DEFAULT 'ACTIVE',            -- ACTIVE, INACTIVE, DISSOLVED, SUSPENDED
    description TEXT,
    metadata JSONB,                          -- Additional flexible data
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now(),
    CONSTRAINT entity_type_check CHECK (entity_type IN ('COMPANY', 'FUND', 'PERSON', 'PARTNERSHIP', 'TRUST', 'OTHER'))
);

-- Client Business Unit: Organizational structure grouping entities under common management
CREATE TABLE IF NOT EXISTS cbu (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name TEXT NOT NULL,
    code TEXT UNIQUE,                        -- Unique identifier code (e.g., BLACKROCK-GLOBAL)
    sponsor_entity_id UUID REFERENCES entity(id),
    domicile TEXT,                           -- Primary jurisdiction
    description TEXT,
    status TEXT DEFAULT 'ACTIVE',
    metadata JSONB,
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now()
);

-- Role Type: Defines types of roles entities can play in CBU relationships
CREATE TABLE IF NOT EXISTS role_type (
    id SERIAL PRIMARY KEY,
    code TEXT UNIQUE NOT NULL,               -- MANCO, FUND, DEPOSITARY, INVESTMENT_MANAGER, TRUSTEE
    name TEXT NOT NULL,
    description TEXT,
    category TEXT,                           -- SERVICE_PROVIDER, REGULATORY, OPERATIONAL
    created_at TIMESTAMP DEFAULT now()
);

-- CBU Role: Links entities to CBUs with specific roles
CREATE TABLE IF NOT EXISTS cbu_role (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    cbu_id UUID NOT NULL REFERENCES cbu(id) ON DELETE CASCADE,
    entity_id UUID NOT NULL REFERENCES entity(id) ON DELETE CASCADE,
    role_type_id INT NOT NULL REFERENCES role_type(id),
    start_date DATE NOT NULL DEFAULT CURRENT_DATE,
    end_date DATE,
    jurisdiction TEXT,
    is_primary BOOLEAN DEFAULT false,        -- Is this the primary role for this entity in this CBU?
    status TEXT DEFAULT 'ACTIVE',
    metadata JSONB,
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now(),
    CONSTRAINT cbu_role_dates_check CHECK (end_date IS NULL OR end_date >= start_date)
);

-- ============================================================================
-- 2. Ownership & Control
-- ============================================================================

-- Control Type: Enum for different types of control relationships
CREATE TYPE control_type AS ENUM (
    'LEGAL_OWNERSHIP',
    'BENEFICIAL_OWNERSHIP',
    'OPERATIONAL_CONTROL',
    'VOTING_CONTROL',
    'MANAGEMENT_CONTROL',
    'ECONOMIC_INTEREST'
);

-- Entity Control: Represents ownership and control relationships between entities
CREATE TABLE IF NOT EXISTS entity_control (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    controller_entity_id UUID NOT NULL REFERENCES entity(id) ON DELETE CASCADE,
    controlled_entity_id UUID NOT NULL REFERENCES entity(id) ON DELETE CASCADE,
    control_type control_type NOT NULL,
    control_basis TEXT,                      -- e.g., "Direct shareholding", "Voting agreement"
    control_percentage NUMERIC(5,2),         -- 0.00 to 100.00
    effective_percentage NUMERIC(5,2),       -- Calculated effective control (after indirect)
    start_date DATE NOT NULL DEFAULT CURRENT_DATE,
    end_date DATE,
    is_indirect BOOLEAN DEFAULT false,
    indirect_via_entity_id UUID REFERENCES entity(id),
    remarks TEXT,
    source_document TEXT,                    -- Reference to source document
    verified_at TIMESTAMP,
    verified_by TEXT,
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now(),
    CONSTRAINT entity_control_not_self CHECK (controller_entity_id != controlled_entity_id),
    CONSTRAINT entity_control_dates_check CHECK (end_date IS NULL OR end_date >= start_date),
    CONSTRAINT control_percentage_check CHECK (control_percentage IS NULL OR (control_percentage >= 0 AND control_percentage <= 100)),
    CONSTRAINT effective_percentage_check CHECK (effective_percentage IS NULL OR (effective_percentage >= 0 AND effective_percentage <= 100))
);

-- Entity KYC Profile: KYC-specific information for each entity
CREATE TABLE IF NOT EXISTS entity_kyc_profile (
    entity_id UUID PRIMARY KEY REFERENCES entity(id) ON DELETE CASCADE,
    risk_rating TEXT,                        -- LOW, MEDIUM, HIGH, CRITICAL
    kyc_status TEXT DEFAULT 'PENDING',       -- PENDING, IN_PROGRESS, APPROVED, REJECTED, EXPIRED
    last_review_date DATE,
    next_review_date DATE,
    policy_id UUID,                          -- Reference to policy governing this entity
    kyc_token TEXT,                          -- Internal KYC token/identifier
    sanctions_check_status TEXT,
    pep_status BOOLEAN DEFAULT false,        -- Politically Exposed Person
    adverse_media_status TEXT,
    remarks TEXT,
    metadata JSONB,
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now()
);

-- ============================================================================
-- 3. Regulatory Dictionary & Ontology
-- ============================================================================

-- Regulation: Master list of regulations
CREATE TABLE IF NOT EXISTS dictionary_regulation (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    code TEXT UNIQUE NOT NULL,               -- FATCA, CRS, AMLD5, MAS626, etc.
    name TEXT NOT NULL,
    jurisdiction TEXT NOT NULL,              -- US, EU, SG, etc.
    authority TEXT,                          -- IRS, ESMA, MAS, etc.
    effective_date DATE,
    description TEXT,
    url TEXT,                                -- Link to official text
    status TEXT DEFAULT 'ACTIVE',
    metadata JSONB,
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now()
);

-- Document: Types of documents required by regulations
CREATE TABLE IF NOT EXISTS dictionary_document (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    code TEXT UNIQUE NOT NULL,               -- DOC_PASSPORT, DOC_W9, etc.
    title TEXT NOT NULL,
    jurisdiction TEXT,
    category TEXT,                           -- IDENTITY, TAX, ADDRESS, ENTITY, FINANCIAL
    description TEXT,
    url TEXT,                                -- Reference URL or template
    regulation_id UUID REFERENCES dictionary_regulation(id),
    validity_period_days INT,                -- How long document remains valid
    is_mandatory BOOLEAN DEFAULT true,
    metadata JSONB,
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now()
);

-- Concept: High-level ontological concepts for semantic organization
CREATE TABLE IF NOT EXISTS dictionary_concept (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    code TEXT UNIQUE NOT NULL,               -- BENEFICIAL_OWNER, TAX_RESIDENCE, etc.
    name TEXT NOT NULL,
    description TEXT,
    domain TEXT,                             -- ENTITY, CONTROL, POLICY, DOCUMENT, RISK, COMPLIANCE
    parent_concept_id UUID REFERENCES dictionary_concept(id),
    synonyms TEXT[],                         -- Array of synonym terms
    embedding VECTOR(1536),                  -- OpenAI embedding for semantic search
    metadata JSONB,
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now()
);

-- Attribute: Detailed data attributes mapped to regulations and concepts
CREATE TABLE IF NOT EXISTS dictionary_attribute (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    code TEXT UNIQUE NOT NULL,               -- UBO_NAME, CLIENT_LEI, TAX_ID, etc.
    name TEXT NOT NULL,
    description TEXT,
    attr_type TEXT NOT NULL,                 -- string, number, date, boolean, json, array
    jurisdiction TEXT,                       -- Applicable jurisdiction
    regulation_id UUID REFERENCES dictionary_regulation(id),
    concept_id UUID REFERENCES dictionary_concept(id),
    sink_table TEXT,                         -- Target database table (e.g., "entity")
    sink_column TEXT,                        -- Target database column (e.g., "lei_code")
    source_priority JSONB,                   -- {"primary": "registry_api", "secondary": "manual_entry"}
    validation_rules JSONB,                  -- JSON schema or validation rules
    is_pii BOOLEAN DEFAULT false,            -- Is personally identifiable information
    is_required BOOLEAN DEFAULT false,
    vector VECTOR(1536),                     -- Embedding for semantic search
    metadata JSONB,
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now()
);

-- Document-Attribute Link: Maps which attributes are required for each document
CREATE TABLE IF NOT EXISTS dictionary_doc_attr_link (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    document_id UUID NOT NULL REFERENCES dictionary_document(id) ON DELETE CASCADE,
    attribute_id UUID NOT NULL REFERENCES dictionary_attribute(id) ON DELETE CASCADE,
    is_required BOOLEAN DEFAULT true,
    order_index INT,                         -- Display order
    created_at TIMESTAMP DEFAULT now(),
    UNIQUE(document_id, attribute_id)
);

-- Document-Regulation Link: Maps which regulations require which documents
CREATE TABLE IF NOT EXISTS dictionary_doc_reg_link (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    document_id UUID NOT NULL REFERENCES dictionary_document(id) ON DELETE CASCADE,
    regulation_id UUID NOT NULL REFERENCES dictionary_regulation(id) ON DELETE CASCADE,
    is_mandatory BOOLEAN DEFAULT true,
    jurisdiction TEXT,
    created_at TIMESTAMP DEFAULT now(),
    UNIQUE(document_id, regulation_id, jurisdiction)
);

-- Source Feedback: Tracks success/failure rates for data sources per attribute
CREATE TABLE IF NOT EXISTS dictionary_source_feedback (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    attribute_id UUID NOT NULL REFERENCES dictionary_attribute(id) ON DELETE CASCADE,
    source_name TEXT NOT NULL,               -- "registry_api", "manual_entry", "vendor_api"
    success_count INT DEFAULT 0,
    failure_count INT DEFAULT 0,
    last_success TIMESTAMP,
    last_failure TIMESTAMP,
    last_error TEXT,
    metadata JSONB,
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now(),
    UNIQUE(attribute_id, source_name)
);

-- Attribute Relationship: Tracks dependencies and derivations between attributes
CREATE TABLE IF NOT EXISTS dictionary_relationship (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    parent_attribute_id UUID NOT NULL REFERENCES dictionary_attribute(id) ON DELETE CASCADE,
    child_attribute_id UUID NOT NULL REFERENCES dictionary_attribute(id) ON DELETE CASCADE,
    relationship_type TEXT NOT NULL,         -- derived_from, depends_on, validates, contradicts, supersedes
    description TEXT,
    derivation_rule TEXT,                    -- Formula or rule for derivation
    metadata JSONB,
    created_at TIMESTAMP DEFAULT now(),
    CONSTRAINT relationship_not_self CHECK (parent_attribute_id != child_attribute_id)
);

-- ============================================================================
-- 4. Indexes for Performance
-- ============================================================================

-- Entity indexes
CREATE INDEX IF NOT EXISTS idx_entity_type ON entity(entity_type);
CREATE INDEX IF NOT EXISTS idx_entity_jurisdiction ON entity(jurisdiction);
CREATE INDEX IF NOT EXISTS idx_entity_status ON entity(status);
CREATE INDEX IF NOT EXISTS idx_entity_lei_code ON entity(lei_code);
CREATE INDEX IF NOT EXISTS idx_entity_name_trgm ON entity USING gin (name gin_trgm_ops);

-- CBU indexes
CREATE INDEX IF NOT EXISTS idx_cbu_code ON cbu(code);
CREATE INDEX IF NOT EXISTS idx_cbu_sponsor ON cbu(sponsor_entity_id);
CREATE INDEX IF NOT EXISTS idx_cbu_domicile ON cbu(domicile);

-- CBU Role indexes
CREATE INDEX IF NOT EXISTS idx_cbu_role_cbu ON cbu_role(cbu_id);
CREATE INDEX IF NOT EXISTS idx_cbu_role_entity ON cbu_role(entity_id);
CREATE INDEX IF NOT EXISTS idx_cbu_role_type ON cbu_role(role_type_id);
CREATE INDEX IF NOT EXISTS idx_cbu_role_status ON cbu_role(status);

-- Entity Control indexes
CREATE INDEX IF NOT EXISTS idx_entity_control_controller ON entity_control(controller_entity_id);
CREATE INDEX IF NOT EXISTS idx_entity_control_controlled ON entity_control(controlled_entity_id);
CREATE INDEX IF NOT EXISTS idx_entity_control_type ON entity_control(control_type);
CREATE INDEX IF NOT EXISTS idx_entity_control_indirect ON entity_control(is_indirect);

-- Dictionary indexes
CREATE INDEX IF NOT EXISTS idx_regulation_code ON dictionary_regulation(code);
CREATE INDEX IF NOT EXISTS idx_regulation_jurisdiction ON dictionary_regulation(jurisdiction);
CREATE INDEX IF NOT EXISTS idx_document_code ON dictionary_document(code);
CREATE INDEX IF NOT EXISTS idx_document_category ON dictionary_document(category);
CREATE INDEX IF NOT EXISTS idx_concept_code ON dictionary_concept(code);
CREATE INDEX IF NOT EXISTS idx_concept_domain ON dictionary_concept(domain);
CREATE INDEX IF NOT EXISTS idx_attribute_code ON dictionary_attribute(code);
CREATE INDEX IF NOT EXISTS idx_attribute_jurisdiction ON dictionary_attribute(jurisdiction);
CREATE INDEX IF NOT EXISTS idx_attribute_type ON dictionary_attribute(attr_type);

-- Vector indexes for semantic search (using IVFFlat algorithm)
CREATE INDEX IF NOT EXISTS idx_attribute_vector ON dictionary_attribute
    USING ivfflat (vector vector_l2_ops) WITH (lists = 100);
CREATE INDEX IF NOT EXISTS idx_concept_embedding ON dictionary_concept
    USING ivfflat (embedding vector_l2_ops) WITH (lists = 100);

-- Source feedback indexes
CREATE INDEX IF NOT EXISTS idx_source_feedback_attr ON dictionary_source_feedback(attribute_id);
CREATE INDEX IF NOT EXISTS idx_source_feedback_source ON dictionary_source_feedback(source_name);

-- Relationship indexes
CREATE INDEX IF NOT EXISTS idx_relationship_parent ON dictionary_relationship(parent_attribute_id);
CREATE INDEX IF NOT EXISTS idx_relationship_child ON dictionary_relationship(child_attribute_id);
CREATE INDEX IF NOT EXISTS idx_relationship_type ON dictionary_relationship(relationship_type);

-- ============================================================================
-- 5. Triggers for Updated Timestamps
-- ============================================================================

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Apply to all tables with updated_at
CREATE TRIGGER update_entity_updated_at BEFORE UPDATE ON entity
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_cbu_updated_at BEFORE UPDATE ON cbu
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_cbu_role_updated_at BEFORE UPDATE ON cbu_role
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_entity_control_updated_at BEFORE UPDATE ON entity_control
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_entity_kyc_profile_updated_at BEFORE UPDATE ON entity_kyc_profile
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_dictionary_regulation_updated_at BEFORE UPDATE ON dictionary_regulation
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_dictionary_document_updated_at BEFORE UPDATE ON dictionary_document
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_dictionary_concept_updated_at BEFORE UPDATE ON dictionary_concept
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_dictionary_attribute_updated_at BEFORE UPDATE ON dictionary_attribute
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_dictionary_source_feedback_updated_at BEFORE UPDATE ON dictionary_source_feedback
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================================================
-- 6. Seed Data
-- ============================================================================

-- Role Types
INSERT INTO role_type (code, name, description, category) VALUES
    ('MANCO', 'Management Company', 'Entity providing management services', 'SERVICE_PROVIDER'),
    ('FUND', 'Fund', 'Investment fund entity', 'OPERATIONAL'),
    ('DEPOSITARY', 'Depositary', 'Entity holding fund assets', 'SERVICE_PROVIDER'),
    ('INVESTMENT_MANAGER', 'Investment Manager', 'Entity managing investments', 'SERVICE_PROVIDER'),
    ('TRUSTEE', 'Trustee', 'Trust management entity', 'SERVICE_PROVIDER'),
    ('CUSTODIAN', 'Custodian', 'Asset custody provider', 'SERVICE_PROVIDER'),
    ('ADMINISTRATOR', 'Administrator', 'Fund administration provider', 'SERVICE_PROVIDER'),
    ('AUDITOR', 'Auditor', 'External auditor', 'REGULATORY'),
    ('LEGAL_COUNSEL', 'Legal Counsel', 'Legal advisory services', 'REGULATORY')
ON CONFLICT (code) DO NOTHING;

-- Regulations
INSERT INTO dictionary_regulation (code, name, jurisdiction, authority, description) VALUES
    ('FATCA', 'Foreign Account Tax Compliance Act', 'US', 'IRS', 'US tax compliance for foreign financial accounts'),
    ('CRS', 'Common Reporting Standard', 'GLOBAL', 'OECD', 'International automatic exchange of financial account information'),
    ('AMLD5', 'Fifth Anti-Money Laundering Directive', 'EU', 'European Commission', 'EU anti-money laundering framework'),
    ('AMLD6', 'Sixth Anti-Money Laundering Directive', 'EU', 'European Commission', 'Enhanced EU AML provisions'),
    ('MAS626', 'MAS Notice 626', 'SG', 'MAS', 'Singapore AML/CFT requirements for financial institutions'),
    ('GDPR', 'General Data Protection Regulation', 'EU', 'European Commission', 'Data protection and privacy regulation'),
    ('KYCA', 'Know Your Customer Act', 'GLOBAL', 'Various', 'General KYC requirements for financial institutions'),
    ('AIFMD', 'Alternative Investment Fund Managers Directive', 'EU', 'ESMA', 'EU regulation for alternative investment fund managers')
ON CONFLICT (code) DO NOTHING;

-- Documents
INSERT INTO dictionary_document (code, title, jurisdiction, category, description) VALUES
    ('DOC_PASSPORT', 'Passport', 'GLOBAL', 'IDENTITY', 'Government-issued passport for individual identification'),
    ('DOC_NATIONAL_ID', 'National Identity Card', 'GLOBAL', 'IDENTITY', 'Government-issued national ID card'),
    ('DOC_DRIVERS_LICENSE', 'Drivers License', 'GLOBAL', 'IDENTITY', 'Government-issued drivers license'),
    ('DOC_CERT_INC', 'Certificate of Incorporation', 'GLOBAL', 'ENTITY', 'Official certificate of company registration'),
    ('DOC_ARTICLES', 'Articles of Association', 'GLOBAL', 'ENTITY', 'Company constitutional document'),
    ('DOC_SHAREHOLDERS', 'Shareholders Register', 'GLOBAL', 'ENTITY', 'Register of company shareholders'),
    ('DOC_W9', 'IRS Form W-9', 'US', 'TAX', 'Request for Taxpayer Identification Number'),
    ('DOC_W8BEN', 'IRS Form W-8BEN', 'US', 'TAX', 'Certificate of Foreign Status for Individuals'),
    ('DOC_W8BENE', 'IRS Form W-8BEN-E', 'US', 'TAX', 'Certificate of Foreign Status for Entities'),
    ('DOC_UTILITY_BILL', 'Utility Bill', 'GLOBAL', 'ADDRESS', 'Proof of residential address'),
    ('DOC_BANK_STATEMENT', 'Bank Statement', 'GLOBAL', 'FINANCIAL', 'Bank account statement for verification'),
    ('DOC_TAX_CERT', 'Tax Residency Certificate', 'GLOBAL', 'TAX', 'Official tax residency certification')
ON CONFLICT (code) DO NOTHING;

-- Concepts
INSERT INTO dictionary_concept (code, name, description, domain, synonyms) VALUES
    ('BENEFICIAL_OWNER', 'Beneficial Owner', 'Individual who ultimately owns or controls an entity', 'CONTROL', ARRAY['UBO', 'Ultimate Beneficial Owner', 'True Owner']),
    ('TAX_RESIDENCE', 'Tax Residence', 'Country where entity is liable for taxation', 'COMPLIANCE', ARRAY['Tax Domicile', 'Fiscal Residence']),
    ('LEGAL_ENTITY', 'Legal Entity', 'Organization with legal rights and obligations', 'ENTITY', ARRAY['Corporation', 'Company', 'Organization']),
    ('CONTROL_PERSON', 'Control Person', 'Individual with significant control or influence', 'CONTROL', ARRAY['Controlling Person', 'Key Person']),
    ('REPORTING_OBLIGATION', 'Reporting Obligation', 'Regulatory requirement to report information', 'COMPLIANCE', ARRAY['Reporting Requirement', 'Disclosure Obligation']),
    ('RISK_ASSESSMENT', 'Risk Assessment', 'Evaluation of entity risk profile', 'RISK', ARRAY['Risk Rating', 'Risk Evaluation']),
    ('DUE_DILIGENCE', 'Due Diligence', 'Investigation and verification process', 'COMPLIANCE', ARRAY['DD', 'KYC Process', 'Verification'])
ON CONFLICT (code) DO NOTHING;

-- Attributes
INSERT INTO dictionary_attribute (code, name, description, attr_type, jurisdiction, sink_table, sink_column, is_required) VALUES
    ('CLIENT_NAME', 'Client Legal Name', 'Official registered name of the entity', 'string', 'GLOBAL', 'entity', 'name', true),
    ('CLIENT_LEI', 'Legal Entity Identifier', 'Global LEI code for entity', 'string', 'GLOBAL', 'entity', 'lei_code', false),
    ('CLIENT_JURISDICTION', 'Jurisdiction of Incorporation', 'Country/state of entity incorporation', 'string', 'GLOBAL', 'entity', 'jurisdiction', true),
    ('CLIENT_REG_NUMBER', 'Registration Number', 'Official company registration number', 'string', 'GLOBAL', 'entity', 'registration_number', true),
    ('UBO_NAME', 'Ultimate Beneficial Owner Name', 'Name of individual UBO', 'string', 'GLOBAL', NULL, NULL, true),
    ('UBO_PERCENT', 'UBO Ownership Percentage', 'Percentage of ownership by UBO', 'number', 'GLOBAL', NULL, NULL, true),
    ('UBO_ADDRESS', 'UBO Residential Address', 'Residential address of UBO', 'string', 'GLOBAL', NULL, NULL, true),
    ('UBO_DOB', 'UBO Date of Birth', 'Date of birth of UBO', 'date', 'GLOBAL', NULL, NULL, true),
    ('TAX_ID', 'Tax Identification Number', 'Tax ID or TIN', 'string', 'GLOBAL', NULL, NULL, false),
    ('TAX_COUNTRY', 'Tax Residence Country', 'Primary country of tax residence', 'string', 'GLOBAL', NULL, NULL, true),
    ('ENTITY_TYPE', 'Entity Type', 'Type of legal entity', 'string', 'GLOBAL', 'entity', 'entity_type', true),
    ('LEGAL_FORM', 'Legal Form', 'Specific legal structure', 'string', 'GLOBAL', 'entity', 'legal_form', false),
    ('RISK_RATING', 'Risk Rating', 'Assessed risk level', 'string', 'GLOBAL', 'entity_kyc_profile', 'risk_rating', false),
    ('KYC_STATUS', 'KYC Status', 'Current KYC approval status', 'string', 'GLOBAL', 'entity_kyc_profile', 'kyc_status', true)
ON CONFLICT (code) DO NOTHING;

COMMIT;

-- ============================================================================
-- End of Schema
-- ============================================================================
