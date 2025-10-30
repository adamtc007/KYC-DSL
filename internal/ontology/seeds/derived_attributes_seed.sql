-- ===========================================================
-- derived_attributes_seed.sql
-- Seed Data for Public/Private Attributes and Derivation Rules
-- ===========================================================

-- ==================== Update Existing Attributes to Public ====================
-- Mark all existing attributes as Public (they are observable from documents)
UPDATE kyc_attributes SET attribute_class = 'Public' WHERE attribute_class IS NULL;

-- ==================== Private (Derived) Attributes ====================

-- Risk Flags
INSERT INTO kyc_attributes (code, name, domain, description, risk_category, is_personal_data, attribute_class)
VALUES
('HIGH_RISK_JURISDICTION_FLAG', 'High Risk Jurisdiction Flag', 'Risk', 'Flag indicating client operates in high-risk jurisdiction', 'HIGH', FALSE, 'Private'),
('SANCTIONED_COUNTRY_FLAG', 'Sanctioned Country Flag', 'Risk', 'Flag indicating exposure to sanctioned countries', 'CRITICAL', FALSE, 'Private'),
('PEP_EXPOSURE_FLAG', 'PEP Exposure Flag', 'Risk', 'Flag indicating beneficial owner is PEP', 'HIGH', FALSE, 'Private'),
('COMPLEX_STRUCTURE_FLAG', 'Complex Ownership Structure Flag', 'Risk', 'Flag for ownership structures with >5 layers', 'MEDIUM', FALSE, 'Private');

-- Computed Scores
INSERT INTO kyc_attributes (code, name, domain, description, risk_category, is_personal_data, attribute_class)
VALUES
('UBO_CONCENTRATION_SCORE', 'UBO Concentration Score', 'Ownership', 'Percentage of ownership held by largest UBO', 'MEDIUM', FALSE, 'Private'),
('ENTITY_AGE_YEARS', 'Entity Age in Years', 'Entity', 'Years since incorporation', 'LOW', FALSE, 'Private'),
('JURISDICTION_RISK_SCORE', 'Jurisdiction Risk Score', 'Risk', 'Numeric risk score for tax residency country (0-100)', 'HIGH', FALSE, 'Private'),
('OVERALL_RISK_RATING', 'Overall Risk Rating', 'Risk', 'Composite risk rating from all factors', 'HIGH', FALSE, 'Private');

-- Status Indicators
INSERT INTO kyc_attributes (code, name, domain, description, risk_category, is_personal_data, attribute_class)
VALUES
('ENTITY_ACTIVE_STATUS', 'Entity Active Status', 'Entity', 'Whether entity is currently active in business registry', 'MEDIUM', FALSE, 'Private'),
('DOCUMENT_COMPLETENESS_FLAG', 'Document Completeness Flag', 'Compliance', 'Whether all required documents are present', 'MEDIUM', FALSE, 'Private'),
('DATA_QUALITY_SCORE', 'Data Quality Score', 'Compliance', 'Percentage of required attributes populated', 'LOW', FALSE, 'Private');

-- ==================== Derivation Rules ====================

-- HIGH_RISK_JURISDICTION_FLAG
-- Rule: TRUE if TaxResidencyCountry is in high-risk list
INSERT INTO kyc_attribute_derivations (derived_attribute_code, source_attribute_code, rule_expression, jurisdiction, regulation_code)
VALUES
('HIGH_RISK_JURISDICTION_FLAG', 'TAX_RESIDENCY_COUNTRY',
 '(if (in TaxResidencyCountry [''IR'' ''KP'' ''SY'' ''YE'' ''AF'' ''MM'']) true false)',
 'GLOBAL', 'AMLD5');

-- SANCTIONED_COUNTRY_FLAG
-- Rule: TRUE if any jurisdiction field matches sanctioned countries
INSERT INTO kyc_attribute_derivations (derived_attribute_code, source_attribute_code, rule_expression, jurisdiction, regulation_code)
VALUES
('SANCTIONED_COUNTRY_FLAG', 'TAX_RESIDENCY_COUNTRY',
 '(if (in TaxResidencyCountry [''IR'' ''KP'' ''SY'' ''CU'' ''RU'']) true false)',
 'GLOBAL', 'BSAAML'),
('SANCTIONED_COUNTRY_FLAG', 'INCORPORATION_JURISDICTION',
 '(if (in IncorporationJurisdiction [''IR'' ''KP'' ''SY'' ''CU'' ''RU'']) true false)',
 'GLOBAL', 'BSAAML');

-- PEP_EXPOSURE_FLAG
-- Rule: TRUE if any UBO or Director has PEP status
INSERT INTO kyc_attribute_derivations (derived_attribute_code, source_attribute_code, rule_expression, jurisdiction, regulation_code)
VALUES
('PEP_EXPOSURE_FLAG', 'PEP_STATUS',
 '(if (= PEP_STATUS true) true false)',
 'GLOBAL', 'AMLD5');

-- COMPLEX_STRUCTURE_FLAG
-- Rule: TRUE if more than 3 beneficial owners (indicates complexity)
INSERT INTO kyc_attribute_derivations (derived_attribute_code, source_attribute_code, rule_expression, jurisdiction, regulation_code)
VALUES
('COMPLEX_STRUCTURE_FLAG', 'UBO_NAME',
 '(if (> (count UBO_NAME) 3) true false)',
 'GLOBAL', 'AMLD5');

-- UBO_CONCENTRATION_SCORE
-- Rule: Maximum ownership percentage held by single UBO
INSERT INTO kyc_attribute_derivations (derived_attribute_code, source_attribute_code, rule_expression, jurisdiction, regulation_code)
VALUES
('UBO_CONCENTRATION_SCORE', 'UBO_PERCENT',
 '(max UBO_PERCENT)',
 'GLOBAL', 'AMLD5');

-- ENTITY_AGE_YEARS
-- Rule: Current year minus incorporation year
INSERT INTO kyc_attribute_derivations (derived_attribute_code, source_attribute_code, rule_expression, jurisdiction, regulation_code)
VALUES
('ENTITY_AGE_YEARS', 'INCORPORATION_DATE',
 '(- (year (now)) (year IncorporationDate))',
 'GLOBAL', NULL);

-- JURISDICTION_RISK_SCORE
-- Rule: Map jurisdiction to risk score
INSERT INTO kyc_attribute_derivations (derived_attribute_code, source_attribute_code, rule_expression, jurisdiction, regulation_code)
VALUES
('JURISDICTION_RISK_SCORE', 'TAX_RESIDENCY_COUNTRY',
 '(case TaxResidencyCountry
    ([''IR'' ''KP'' ''SY''] 100)
    ([''AF'' ''YE'' ''MM''] 90)
    ([''RU'' ''BY''] 80)
    ([''CN'' ''HK''] 60)
    ([''US'' ''GB'' ''SG''] 20)
    ([''CH'' ''DE'' ''FR''] 10)
    (else 50))',
 'GLOBAL', 'AMLD5');

-- ENTITY_ACTIVE_STATUS
-- Rule: Check if entity appears in business registry (pseudo-code)
INSERT INTO kyc_attribute_derivations (derived_attribute_code, source_attribute_code, rule_expression, jurisdiction, regulation_code)
VALUES
('ENTITY_ACTIVE_STATUS', 'REGISTERED_NAME',
 '(registry-active? RegisteredName IncorporationJurisdiction)',
 'GLOBAL', NULL),
('ENTITY_ACTIVE_STATUS', 'INCORPORATION_JURISDICTION',
 '(registry-active? RegisteredName IncorporationJurisdiction)',
 'GLOBAL', NULL);

-- DOCUMENT_COMPLETENESS_FLAG
-- Rule: Check if all required documents present (simplified)
INSERT INTO kyc_attribute_derivations (derived_attribute_code, source_attribute_code, rule_expression, jurisdiction, regulation_code)
VALUES
('DOCUMENT_COMPLETENESS_FLAG', 'REGISTERED_NAME',
 '(and (exists? CERT-INC) (exists? UBO-DECL) (exists? W8BENE))',
 'EU', 'AMLD5');

-- DATA_QUALITY_SCORE
-- Rule: Percentage of required fields populated
INSERT INTO kyc_attribute_derivations (derived_attribute_code, source_attribute_code, rule_expression, jurisdiction, regulation_code)
VALUES
('DATA_QUALITY_SCORE', 'REGISTERED_NAME',
 '(* 100 (/ (count-populated [RegisteredName TaxResidencyCountry UBO_NAME]) (count-required)))',
 'GLOBAL', NULL);

-- OVERALL_RISK_RATING
-- Rule: Composite score from multiple risk factors
INSERT INTO kyc_attribute_derivations (derived_attribute_code, source_attribute_code, rule_expression, jurisdiction, regulation_code)
VALUES
('OVERALL_RISK_RATING', 'HIGH_RISK_JURISDICTION_FLAG',
 '(if (= HighRiskJurisdictionFlag true) (+ risk-score 30) risk-score)',
 'GLOBAL', 'AMLD5'),
('OVERALL_RISK_RATING', 'SANCTIONED_COUNTRY_FLAG',
 '(if (= SanctionedCountryFlag true) (+ risk-score 50) risk-score)',
 'GLOBAL', 'BSAAML'),
('OVERALL_RISK_RATING', 'PEP_EXPOSURE_FLAG',
 '(if (= PEPExposureFlag true) (+ risk-score 40) risk-score)',
 'GLOBAL', 'AMLD5');

-- ==================== Verification Queries ====================

-- Count public vs private attributes
-- SELECT attribute_class, COUNT(*) FROM kyc_attributes GROUP BY attribute_class;

-- Show all derivation lineage
-- SELECT * FROM attribute_lineage ORDER BY derived_attribute;

-- Find attributes with multiple derivation sources
-- SELECT derived_attribute_code, COUNT(*) as source_count
-- FROM kyc_attribute_derivations
-- GROUP BY derived_attribute_code
-- HAVING COUNT(*) > 1;
