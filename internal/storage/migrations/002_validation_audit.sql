-- ===========================================================
-- 002_validation_audit.sql
-- Validation and Ontology Audit Trail
-- Provides audit logging for compliance with FCA SYSC, MAS 626 §4.2,
-- HKMA AML §3.6, EU AMLD6 Article 30 (record-keeping)
-- ===========================================================

-- Validation audit table: Records every validation attempt
CREATE TABLE IF NOT EXISTS kyc_case_validations (
    id SERIAL PRIMARY KEY,
    case_name TEXT NOT NULL,
    version INT NOT NULL,
    validation_time TIMESTAMP DEFAULT NOW(),
    grammar_version TEXT,
    ontology_version TEXT,
    validator_actor TEXT,          -- e.g., "System", "Agent:Claude", "User:adam"
    validation_status TEXT CHECK (validation_status IN ('PASS','FAIL')) NOT NULL,
    error_message TEXT,
    total_checks INT DEFAULT 0,
    passed_checks INT DEFAULT 0,
    failed_checks INT DEFAULT 0,
    metadata JSONB,                -- Additional validation context
    created_at TIMESTAMP DEFAULT NOW()
);

-- Detailed validation findings: Individual check results
CREATE TABLE IF NOT EXISTS kyc_validation_findings (
    id SERIAL PRIMARY KEY,
    validation_id INT REFERENCES kyc_case_validations(id) ON DELETE CASCADE,
    check_type TEXT NOT NULL,      -- e.g., "ontology_document", "ownership_sum", "semantic"
    check_name TEXT NOT NULL,
    check_status TEXT CHECK (check_status IN ('PASS','WARN','FAIL')) NOT NULL,
    check_message TEXT,
    entity_ref TEXT,               -- Reference to entity being checked (e.g., "UBO_NAME")
    severity TEXT CHECK (severity IN ('INFO','WARNING','ERROR','CRITICAL')),
    created_at TIMESTAMP DEFAULT NOW()
);

-- Indexes for efficient querying
CREATE INDEX IF NOT EXISTS idx_case_validations_case
    ON kyc_case_validations(case_name);

CREATE INDEX IF NOT EXISTS idx_case_validations_status
    ON kyc_case_validations(validation_status);

CREATE INDEX IF NOT EXISTS idx_case_validations_time
    ON kyc_case_validations(validation_time DESC);

CREATE INDEX IF NOT EXISTS idx_validation_findings_validation
    ON kyc_validation_findings(validation_id);

CREATE INDEX IF NOT EXISTS idx_validation_findings_type
    ON kyc_validation_findings(check_type);

CREATE INDEX IF NOT EXISTS idx_validation_findings_status
    ON kyc_validation_findings(check_status);

-- View for validation summary statistics
CREATE OR REPLACE VIEW validation_summary AS
SELECT
    case_name,
    COUNT(*) as total_validations,
    SUM(CASE WHEN validation_status = 'PASS' THEN 1 ELSE 0 END) as passed,
    SUM(CASE WHEN validation_status = 'FAIL' THEN 1 ELSE 0 END) as failed,
    MAX(validation_time) as last_validation,
    AVG(total_checks) as avg_checks_per_validation
FROM kyc_case_validations
GROUP BY case_name;

-- View for compliance audit report
CREATE OR REPLACE VIEW compliance_audit_trail AS
SELECT
    v.case_name,
    v.version,
    v.validation_time,
    v.validator_actor,
    v.validation_status,
    v.grammar_version,
    v.ontology_version,
    COUNT(f.id) as finding_count,
    SUM(CASE WHEN f.severity = 'CRITICAL' THEN 1 ELSE 0 END) as critical_findings,
    SUM(CASE WHEN f.severity = 'ERROR' THEN 1 ELSE 0 END) as error_findings,
    SUM(CASE WHEN f.severity = 'WARNING' THEN 1 ELSE 0 END) as warning_findings
FROM kyc_case_validations v
LEFT JOIN kyc_validation_findings f ON f.validation_id = v.id
GROUP BY v.id, v.case_name, v.version, v.validation_time,
         v.validator_actor, v.validation_status,
         v.grammar_version, v.ontology_version
ORDER BY v.validation_time DESC;
