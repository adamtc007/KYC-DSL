-- ===========================================================
-- 005_lineage_evaluations.sql
-- Lineage Evaluation Audit Trail
-- Records the execution and results of derived attribute rules
-- ===========================================================

-- Lineage Evaluations: Audit trail of rule executions
CREATE TABLE IF NOT EXISTS kyc_lineage_evaluations (
    id SERIAL PRIMARY KEY,
    case_name TEXT NOT NULL,
    case_version INT,
    derived_code TEXT NOT NULL,
    value TEXT,
    value_type TEXT,                -- boolean, numeric, string
    success BOOLEAN NOT NULL,
    error TEXT,
    inputs JSONB,                   -- Source attribute values used
    rule TEXT NOT NULL,
    jurisdiction TEXT,
    regulation_code TEXT,
    evaluated_at TIMESTAMP DEFAULT NOW(),
    created_at TIMESTAMP DEFAULT NOW()
);

-- Indexes for efficient querying
CREATE INDEX IF NOT EXISTS idx_lineage_evaluations_case
    ON kyc_lineage_evaluations(case_name);

CREATE INDEX IF NOT EXISTS idx_lineage_evaluations_derived
    ON kyc_lineage_evaluations(derived_code);

CREATE INDEX IF NOT EXISTS idx_lineage_evaluations_success
    ON kyc_lineage_evaluations(success);

CREATE INDEX IF NOT EXISTS idx_lineage_evaluations_time
    ON kyc_lineage_evaluations(evaluated_at DESC);

-- View: Recent evaluation results with success/failure stats
CREATE OR REPLACE VIEW lineage_evaluation_summary AS
SELECT
    case_name,
    derived_code,
    COUNT(*) as total_evaluations,
    SUM(CASE WHEN success THEN 1 ELSE 0 END) as successful,
    SUM(CASE WHEN NOT success THEN 1 ELSE 0 END) as failed,
    MAX(evaluated_at) as last_evaluation,
    CASE
        WHEN SUM(CASE WHEN success THEN 1 ELSE 0 END)::float / COUNT(*) > 0.95 THEN 'HEALTHY'
        WHEN SUM(CASE WHEN success THEN 1 ELSE 0 END)::float / COUNT(*) > 0.80 THEN 'WARNING'
        ELSE 'CRITICAL'
    END as health_status
FROM kyc_lineage_evaluations
GROUP BY case_name, derived_code;

-- View: Failed evaluations for debugging
CREATE OR REPLACE VIEW failed_lineage_evaluations AS
SELECT
    case_name,
    derived_code,
    error,
    inputs,
    rule,
    evaluated_at
FROM kyc_lineage_evaluations
WHERE success = FALSE
ORDER BY evaluated_at DESC;

-- COMMENT statements for documentation
COMMENT ON TABLE kyc_lineage_evaluations IS
    'Audit trail of derived attribute rule executions with inputs, outputs, and success status';

COMMENT ON COLUMN kyc_lineage_evaluations.inputs IS
    'JSONB object containing source attribute codes and their values at evaluation time';

COMMENT ON COLUMN kyc_lineage_evaluations.value IS
    'Stringified result of rule evaluation (true/false for boolean, numeric value, or string)';

COMMENT ON COLUMN kyc_lineage_evaluations.rule IS
    'The rule expression that was evaluated, for audit and debugging purposes';
