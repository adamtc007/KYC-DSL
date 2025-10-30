-- ===========================================================
-- 004_attribute_derivations.sql
-- Public vs Private Attribute Classification + Lineage Rules
-- ===========================================================

-- Add attribute_class column to existing kyc_attributes table
ALTER TABLE kyc_attributes
ADD COLUMN IF NOT EXISTS attribute_class TEXT
    CHECK (attribute_class IN ('Public', 'Private'))
    DEFAULT 'Public';

-- Create index on attribute_class for filtering
CREATE INDEX IF NOT EXISTS idx_kyc_attributes_class
    ON kyc_attributes(attribute_class);

-- Attribute Derivations: Tracks how private attributes are computed from public ones
CREATE TABLE IF NOT EXISTS kyc_attribute_derivations (
    id SERIAL PRIMARY KEY,
    derived_attribute_code TEXT NOT NULL REFERENCES kyc_attributes(code) ON DELETE CASCADE,
    source_attribute_code TEXT NOT NULL REFERENCES kyc_attributes(code),
    rule_expression TEXT NOT NULL,
    jurisdiction TEXT,
    regulation_code TEXT REFERENCES kyc_regulations(code),
    created_at TIMESTAMP DEFAULT NOW()
);

-- Allow multiple source attributes for a single derived attribute
CREATE INDEX IF NOT EXISTS idx_derivations_derived
    ON kyc_attribute_derivations(derived_attribute_code);

CREATE INDEX IF NOT EXISTS idx_derivations_source
    ON kyc_attribute_derivations(source_attribute_code);

-- View: Show all private attributes with their derivation sources
CREATE OR REPLACE VIEW attribute_lineage AS
SELECT
    a.code as derived_attribute,
    a.name as derived_attribute_name,
    a.domain as derived_domain,
    d.source_attribute_code as source_attribute,
    sa.name as source_attribute_name,
    d.rule_expression,
    d.jurisdiction,
    d.regulation_code
FROM kyc_attributes a
JOIN kyc_attribute_derivations d ON d.derived_attribute_code = a.code
JOIN kyc_attributes sa ON sa.code = d.source_attribute_code
WHERE a.attribute_class = 'Private'
ORDER BY a.code, d.source_attribute_code;

-- COMMENT statements for documentation
COMMENT ON COLUMN kyc_attributes.attribute_class IS
    'Public: Observable data from documents or client facts. Private: Derived/computed values.';

COMMENT ON TABLE kyc_attribute_derivations IS
    'Tracks derivation lineage: which public attributes and rules produce private attributes.';

COMMENT ON COLUMN kyc_attribute_derivations.rule_expression IS
    'DSL or formula describing transformation, e.g., "(if (in TaxResidencyCountry [''IR'' ''KP'']) true false)"';
