-- ===========================================================
-- attribute_metadata_seed.sql
-- Seed Data for Attribute Metadata & Clustering
-- ===========================================================

-- ==================== Attribute Metadata ====================
-- Enrich existing attributes with metadata

-- Entity Attributes
INSERT INTO kyc_attribute_metadata (attribute_code, synonyms, data_type, domain_values, risk_level, example_values, regulatory_citations, data_sensitivity, retention_period_days)
VALUES
('REGISTERED_NAME', ARRAY['Legal Name', 'Company Name', 'Entity Name', 'Corporate Name'], 'string', NULL, 'LOW', ARRAY['BlackRock Global Fund', 'HSBC Holdings PLC', 'Deutsche Bank AG'], ARRAY['AMLD5 Article 13', 'MAS 626 Annex A'], 'INTERNAL', 2555),
('INCORPORATION_DATE', ARRAY['Formation Date', 'Registration Date', 'Establishment Date'], 'date', NULL, 'LOW', ARRAY['2010-01-15', '2005-06-30'], ARRAY['Companies Act'], 'INTERNAL', 2555),
('INCORPORATION_JURISDICTION', ARRAY['Country of Incorporation', 'Jurisdiction of Formation', 'Place of Incorporation'], 'string', ARRAY['US', 'UK', 'DE', 'FR', 'SG', 'HK', 'CH', 'LU', 'IE', 'KY', 'BM'], 'MEDIUM', ARRAY['US', 'UK', 'SG'], ARRAY['AMLD5'], 'INTERNAL', 2555),
('REGISTRATION_NUMBER', ARRAY['Company Number', 'Registration ID', 'Corp ID', 'Entity Number'], 'string', NULL, 'LOW', ARRAY['12345678', 'C-123456', 'HE-987654'], ARRAY['Companies House', 'ACRA', 'MAS 626'], 'INTERNAL', 2555),
('LEI', ARRAY['Legal Entity Identifier', 'LEI Code'], 'string', NULL, 'LOW', ARRAY['549300ASDFJK38492DSF'], ARRAY['GLEIF Standards', 'EMIR', 'MiFID II'], 'PUBLIC', 2555),
('ENTITY_TYPE', ARRAY['Legal Form', 'Entity Structure', 'Corporate Type'], 'string', ARRAY['Corporation', 'LLC', 'Partnership', 'Trust', 'Fund', 'Branch'], 'LOW', ARRAY['Corporation', 'LLC', 'Fund'], ARRAY['AMLD5', 'MAS 626'], 'INTERNAL', 2555),
('REGISTERED_ADDRESS', ARRAY['Registered Office', 'Legal Address', 'Corporate Address'], 'string', NULL, 'LOW', ARRAY['1 Main Street, London, UK', '123 Wall St, New York, NY 10005'], ARRAY['Companies Act', 'AMLD5'], 'CONFIDENTIAL', 2555),
('BUSINESS_ADDRESS', ARRAY['Operating Address', 'Principal Place of Business', 'Trading Address'], 'string', NULL, 'LOW', ARRAY['456 Business Blvd, Singapore'], ARRAY['MAS 626'], 'CONFIDENTIAL', 2555);

-- Tax Attributes
INSERT INTO kyc_attribute_metadata (attribute_code, synonyms, data_type, domain_values, risk_level, example_values, regulatory_citations, data_sensitivity, retention_period_days)
VALUES
('TAX_RESIDENCY_COUNTRY', ARRAY['Tax Residence', 'Country of Tax Residence', 'Tax Jurisdiction'], 'string', ARRAY['US', 'UK', 'DE', 'FR', 'SG', 'HK', 'CH', 'CN', 'IR', 'KP'], 'HIGH', ARRAY['US', 'UK', 'SG'], ARRAY['FATCA', 'CRS', 'OECD Model Tax Convention'], 'CONFIDENTIAL', 2555),
('TAX_ID', ARRAY['TIN', 'Tax Identification Number', 'EIN', 'SSN', 'ITIN'], 'string', NULL, 'CRITICAL', ARRAY['12-3456789', '987-65-4321'], ARRAY['FATCA', 'CRS', 'IRS Code'], 'RESTRICTED', 2555),
('US_TAX_STATUS', ARRAY['US Person Status', 'FATCA Status'], 'boolean', NULL, 'HIGH', ARRAY['true', 'false'], ARRAY['FATCA', 'IRS Regulations'], 'CONFIDENTIAL', 2555),
('FATCA_STATUS', ARRAY['FATCA Classification', 'Chapter 4 Status'], 'string', ARRAY['Active NFFE', 'Passive NFFE', 'FFI', 'Excepted FFI', 'Participating FFI'], 'HIGH', ARRAY['Active NFFE', 'Passive NFFE'], ARRAY['FATCA Regulations', 'IRS Code Chapter 4'], 'CONFIDENTIAL', 2555),
('CRS_CLASSIFICATION', ARRAY['CRS Entity Type', 'CRS Status'], 'string', ARRAY['Financial Institution', 'Active NFE', 'Passive NFE', 'Investment Entity'], 'HIGH', ARRAY['Active NFE', 'Financial Institution'], ARRAY['OECD CRS', 'Multilateral Competent Authority Agreement'], 'CONFIDENTIAL', 2555);

-- Ownership Attributes
INSERT INTO kyc_attribute_metadata (attribute_code, synonyms, data_type, domain_values, risk_level, example_values, regulatory_citations, data_sensitivity, retention_period_days)
VALUES
('UBO_NAME', ARRAY['Ultimate Beneficial Owner Name', 'Beneficial Owner', 'UBO', 'Controller Name'], 'string', NULL, 'CRITICAL', ARRAY['John Smith', 'Jane Doe'], ARRAY['AMLD5 Article 3', 'FATF Recommendation 24', 'MAS 626'], 'RESTRICTED', 2555),
('UBO_PERCENT', ARRAY['Ownership Percentage', 'Beneficial Ownership %', 'UBO Share'], 'float', NULL, 'CRITICAL', ARRAY['25.0', '50.0', '100.0'], ARRAY['AMLD5', 'FATF 25% threshold'], 'CONFIDENTIAL', 2555),
('UBO_DOB', ARRAY['UBO Date of Birth', 'Beneficial Owner DOB'], 'date', NULL, 'CRITICAL', ARRAY['1970-01-15', '1985-06-30'], ARRAY['AMLD5', 'FATF'], 'RESTRICTED', 2555),
('UBO_NATIONALITY', ARRAY['UBO Citizenship', 'Beneficial Owner Nationality'], 'string', NULL, 'HIGH', ARRAY['US', 'UK', 'FR'], ARRAY['AMLD5', 'FATF'], 'RESTRICTED', 2555),
('UBO_ADDRESS', ARRAY['UBO Residential Address', 'Beneficial Owner Address'], 'string', NULL, 'CRITICAL', ARRAY['123 Main St, London'], ARRAY['AMLD5'], 'RESTRICTED', 2555),
('SHAREHOLDER_NAME', ARRAY['Stockholder Name', 'Direct Owner'], 'string', NULL, 'HIGH', ARRAY['ABC Corporation', 'XYZ Holdings'], ARRAY['AMLD5'], 'CONFIDENTIAL', 2555),
('SHAREHOLDER_PERCENT', ARRAY['Direct Ownership %', 'Share Percentage'], 'float', NULL, 'HIGH', ARRAY['10.0', '51.0', '100.0'], ARRAY['AMLD5'], 'CONFIDENTIAL', 2555);

-- Control Attributes
INSERT INTO kyc_attribute_metadata (attribute_code, synonyms, data_type, domain_values, risk_level, example_values, regulatory_citations, data_sensitivity, retention_period_days)
VALUES
('DIRECTOR_NAME', ARRAY['Board Member', 'Director', 'Board Director'], 'string', NULL, 'HIGH', ARRAY['Michael Johnson', 'Sarah Williams'], ARRAY['AMLD5', 'Companies Act'], 'CONFIDENTIAL', 2555),
('AUTHORIZED_SIGNATORY', ARRAY['Signatory', 'Authorized Person', 'Authorized Representative'], 'string', NULL, 'HIGH', ARRAY['Chief Executive Officer', 'Managing Director'], ARRAY['Corporate Governance Codes'], 'CONFIDENTIAL', 2555),
('CONTROL_PERSON', ARRAY['PSC', 'Person with Significant Control', 'Controller'], 'string', NULL, 'CRITICAL', ARRAY['John Controller'], ARRAY['UK PSC Register', 'Companies Act 2006'], 'CONFIDENTIAL', 2555);

-- Identity Attributes
INSERT INTO kyc_attribute_metadata (attribute_code, synonyms, data_type, domain_values, risk_level, example_values, regulatory_citations, data_sensitivity, retention_period_days)
VALUES
('FULL_NAME', ARRAY['Complete Name', 'Legal Name', 'Individual Name'], 'string', NULL, 'CRITICAL', ARRAY['John Michael Smith'], ARRAY['AMLD5', 'FATF'], 'RESTRICTED', 2555),
('DATE_OF_BIRTH', ARRAY['DOB', 'Birth Date', 'Date of Birth'], 'date', NULL, 'CRITICAL', ARRAY['1980-05-15'], ARRAY['AMLD5', 'FATF'], 'RESTRICTED', 2555),
('NATIONALITY', ARRAY['Citizenship', 'Country of Citizenship'], 'string', NULL, 'HIGH', ARRAY['US', 'UK', 'DE'], ARRAY['AMLD5'], 'RESTRICTED', 2555),
('PASSPORT_NUMBER', ARRAY['Passport ID', 'Passport No'], 'string', NULL, 'CRITICAL', ARRAY['P12345678'], ARRAY['AMLD5', 'FATF'], 'RESTRICTED', 2555),
('ID_NUMBER', ARRAY['Identity Document Number', 'National ID'], 'string', NULL, 'CRITICAL', ARRAY['123456789'], ARRAY['AMLD5'], 'RESTRICTED', 2555);

-- Financial Attributes
INSERT INTO kyc_attribute_metadata (attribute_code, synonyms, data_type, domain_values, risk_level, example_values, regulatory_citations, data_sensitivity, retention_period_days)
VALUES
('SOURCE_OF_FUNDS', ARRAY['Funds Source', 'Origin of Funds'], 'string', ARRAY['Salary', 'Business Income', 'Investment Returns', 'Inheritance', 'Sale of Assets'], 'CRITICAL', ARRAY['Salary', 'Business Income'], ARRAY['AMLD5', 'FATF Recommendation 10'], 'CONFIDENTIAL', 2555),
('SOURCE_OF_WEALTH', ARRAY['Wealth Source', 'Origin of Wealth'], 'string', ARRAY['Employment', 'Business Ownership', 'Investments', 'Inheritance', 'Family Wealth'], 'CRITICAL', ARRAY['Employment', 'Business Ownership'], ARRAY['AMLD5', 'FATF'], 'CONFIDENTIAL', 2555),
('EXPECTED_TRANSACTION_VOLUME', ARRAY['Transaction Volume', 'Expected Activity'], 'string', ARRAY['<$10K', '$10K-$100K', '$100K-$1M', '>$1M'], 'MEDIUM', ARRAY['$100K-$1M'], ARRAY['AML Regulations'], 'CONFIDENTIAL', 2555),
('BUSINESS_ACTIVITY', ARRAY['Nature of Business', 'Business Type', 'Industry'], 'string', ARRAY['Banking', 'Investment Management', 'Real Estate', 'Technology', 'Manufacturing', 'Crypto'], 'MEDIUM', ARRAY['Investment Management', 'Technology'], ARRAY['MAS 626'], 'INTERNAL', 2555);

-- Risk Attributes
INSERT INTO kyc_attribute_metadata (attribute_code, synonyms, data_type, domain_values, risk_level, example_values, regulatory_citations, data_sensitivity, retention_period_days)
VALUES
('PEP_STATUS', ARRAY['Politically Exposed Person', 'PEP Flag', 'PEP Indicator'], 'boolean', NULL, 'CRITICAL', ARRAY['true', 'false'], ARRAY['AMLD5 Article 20', 'FATF Recommendation 12'], 'RESTRICTED', 2555),
('SANCTIONS_STATUS', ARRAY['Sanctions Hit', 'Sanctions Flag', 'Sanctioned Entity'], 'boolean', NULL, 'CRITICAL', ARRAY['true', 'false'], ARRAY['OFAC Regulations', 'EU Sanctions', 'UN Sanctions'], 'RESTRICTED', 2555),
('ADVERSE_MEDIA', ARRAY['Negative News', 'Adverse Media Hit'], 'boolean', NULL, 'HIGH', ARRAY['true', 'false'], ARRAY['FATF Guidance'], 'CONFIDENTIAL', 2555),
('RISK_RATING', ARRAY['Risk Score', 'Risk Level', 'Risk Classification'], 'string', ARRAY['LOW', 'MEDIUM', 'HIGH', 'CRITICAL'], 'HIGH', ARRAY['LOW', 'MEDIUM', 'HIGH'], ARRAY['Risk-Based Approach Guidelines'], 'CONFIDENTIAL', 2555);

-- Private (Derived) Attributes Metadata
INSERT INTO kyc_attribute_metadata (attribute_code, synonyms, data_type, domain_values, risk_level, example_values, regulatory_citations, data_sensitivity, retention_period_days)
VALUES
('HIGH_RISK_JURISDICTION_FLAG', ARRAY['High Risk Country Flag', 'Risky Jurisdiction'], 'boolean', NULL, 'CRITICAL', ARRAY['true', 'false'], ARRAY['FATF High-Risk Jurisdictions'], 'CONFIDENTIAL', 2555),
('SANCTIONED_COUNTRY_FLAG', ARRAY['Sanctioned Jurisdiction', 'Embargo Flag'], 'boolean', NULL, 'CRITICAL', ARRAY['true', 'false'], ARRAY['OFAC SDN List', 'EU Sanctions List'], 'CONFIDENTIAL', 2555),
('PEP_EXPOSURE_FLAG', ARRAY['PEP Risk', 'PEP Association'], 'boolean', NULL, 'CRITICAL', ARRAY['true', 'false'], ARRAY['AMLD5', 'FATF R12'], 'CONFIDENTIAL', 2555),
('COMPLEX_STRUCTURE_FLAG', ARRAY['Complex Ownership', 'Layered Structure'], 'boolean', NULL, 'HIGH', ARRAY['true', 'false'], ARRAY['FATF Best Practices'], 'CONFIDENTIAL', 2555),
('UBO_CONCENTRATION_SCORE', ARRAY['Ownership Concentration', 'Control Concentration'], 'float', NULL, 'MEDIUM', ARRAY['25.0', '50.0', '100.0'], ARRAY['AMLD5'], 'CONFIDENTIAL', 2555),
('ENTITY_AGE_YEARS', ARRAY['Company Age', 'Years in Operation'], 'integer', NULL, 'LOW', ARRAY['1', '5', '10', '20'], ARRAY['None'], 'INTERNAL', 2555),
('JURISDICTION_RISK_SCORE', ARRAY['Country Risk Score', 'Geographic Risk'], 'integer', NULL, 'HIGH', ARRAY['10', '50', '100'], ARRAY['FATF', 'Basel AML Index'], 'CONFIDENTIAL', 2555),
('OVERALL_RISK_RATING', ARRAY['Composite Risk', 'Total Risk Score'], 'string', ARRAY['LOW', 'MEDIUM', 'HIGH', 'CRITICAL'], 'CRITICAL', ARRAY['MEDIUM', 'HIGH'], ARRAY['Risk-Based Approach'], 'CONFIDENTIAL', 2555),
('ENTITY_ACTIVE_STATUS', ARRAY['Business Active', 'Operational Status'], 'boolean', NULL, 'MEDIUM', ARRAY['true', 'false'], ARRAY['Companies House'], 'INTERNAL', 1825),
('DOCUMENT_COMPLETENESS_FLAG', ARRAY['Documentation Complete', 'All Docs Present'], 'boolean', NULL, 'MEDIUM', ARRAY['true', 'false'], ARRAY['AML Regulations'], 'INTERNAL', 1825),
('DATA_QUALITY_SCORE', ARRAY['Data Completeness', 'Quality Score'], 'float', NULL, 'LOW', ARRAY['75.0', '90.0', '100.0'], ARRAY['None'], 'INTERNAL', 1825);

-- ==================== Attribute Clusters ====================
-- Logical groupings for agent optimization

INSERT INTO kyc_attribute_clusters (cluster_code, cluster_name, attribute_codes, description, use_case, priority)
VALUES
('ENTITY_IDENTITY', 'Entity Identity',
 ARRAY['REGISTERED_NAME', 'INCORPORATION_DATE', 'INCORPORATION_JURISDICTION', 'REGISTRATION_NUMBER', 'LEI', 'ENTITY_TYPE'],
 'Core attributes that uniquely identify and describe a legal entity',
 'Entity onboarding, entity resolution, due diligence',
 1),

('TAX_PROFILE', 'Tax Profile',
 ARRAY['TAX_RESIDENCY_COUNTRY', 'TAX_ID', 'US_TAX_STATUS', 'FATCA_STATUS', 'CRS_CLASSIFICATION'],
 'Tax-related attributes for FATCA/CRS compliance',
 'Tax reporting, FATCA classification, CRS due diligence',
 2),

('BENEFICIAL_OWNERSHIP', 'Beneficial Ownership',
 ARRAY['UBO_NAME', 'UBO_PERCENT', 'UBO_DOB', 'UBO_NATIONALITY', 'UBO_ADDRESS', 'SHAREHOLDER_NAME', 'SHAREHOLDER_PERCENT'],
 'Attributes tracking beneficial ownership and control persons',
 'UBO identification, ownership transparency, AMLD5 compliance',
 1),

('OPERATIONAL_CONTROL', 'Operational Control',
 ARRAY['DIRECTOR_NAME', 'AUTHORIZED_SIGNATORY', 'CONTROL_PERSON'],
 'Attributes identifying who manages and controls the entity',
 'Management oversight, authorized persons, control verification',
 2),

('INDIVIDUAL_IDENTITY', 'Individual Identity',
 ARRAY['FULL_NAME', 'DATE_OF_BIRTH', 'NATIONALITY', 'PASSPORT_NUMBER', 'ID_NUMBER'],
 'Core attributes for individual identity verification',
 'Individual KYC, identity verification, CDD',
 1),

('FINANCIAL_PROFILE', 'Financial Profile',
 ARRAY['SOURCE_OF_FUNDS', 'SOURCE_OF_WEALTH', 'EXPECTED_TRANSACTION_VOLUME', 'BUSINESS_ACTIVITY'],
 'Financial background and transaction expectations',
 'Source of wealth verification, transaction monitoring setup',
 2),

('RISK_INDICATORS', 'Risk Indicators',
 ARRAY['PEP_STATUS', 'SANCTIONS_STATUS', 'ADVERSE_MEDIA', 'RISK_RATING'],
 'Core risk flags and ratings',
 'Risk assessment, screening, enhanced due diligence triggers',
 1),

('DERIVED_RISK_FLAGS', 'Derived Risk Flags',
 ARRAY['HIGH_RISK_JURISDICTION_FLAG', 'SANCTIONED_COUNTRY_FLAG', 'PEP_EXPOSURE_FLAG', 'COMPLEX_STRUCTURE_FLAG'],
 'Computed risk flags from public attributes',
 'Automated risk detection, rule-based screening',
 2),

('COMPUTED_METRICS', 'Computed Metrics',
 ARRAY['UBO_CONCENTRATION_SCORE', 'ENTITY_AGE_YEARS', 'JURISDICTION_RISK_SCORE', 'OVERALL_RISK_RATING', 'DATA_QUALITY_SCORE'],
 'Derived numeric scores and ratings',
 'Risk scoring, quality assessment, analytics',
 3),

('LOCATION_DATA', 'Location Data',
 ARRAY['INCORPORATION_JURISDICTION', 'TAX_RESIDENCY_COUNTRY', 'REGISTERED_ADDRESS', 'BUSINESS_ADDRESS'],
 'Geographic and jurisdictional information',
 'Geographic risk assessment, sanctions screening',
 2),

('COMPLIANCE_STATUS', 'Compliance Status',
 ARRAY['ENTITY_ACTIVE_STATUS', 'DOCUMENT_COMPLETENESS_FLAG', 'RISK_RATING', 'OVERALL_RISK_RATING'],
 'Current compliance and documentation status',
 'Compliance monitoring, remediation tracking',
 2);

-- ==================== Attribute Relationships ====================
-- Semantic relationships for knowledge graph

-- Derivation relationships
INSERT INTO kyc_attribute_relationships (source_attribute_code, target_attribute_code, relationship_type, strength, description)
VALUES
('TAX_RESIDENCY_COUNTRY', 'HIGH_RISK_JURISDICTION_FLAG', 'derives_from', 1.0, 'Risk flag computed from tax residency'),
('TAX_RESIDENCY_COUNTRY', 'JURISDICTION_RISK_SCORE', 'derives_from', 1.0, 'Risk score based on jurisdiction'),
('UBO_PERCENT', 'UBO_CONCENTRATION_SCORE', 'derives_from', 1.0, 'Concentration computed from ownership percentages'),
('INCORPORATION_DATE', 'ENTITY_AGE_YEARS', 'derives_from', 1.0, 'Age calculated from incorporation date'),
('PEP_STATUS', 'PEP_EXPOSURE_FLAG', 'derives_from', 1.0, 'Flag derived from PEP status'),
('REGISTERED_NAME', 'ENTITY_ACTIVE_STATUS', 'derives_from', 0.8, 'Active status checked via registry lookup');

-- Validation relationships
INSERT INTO kyc_attribute_relationships (source_attribute_code, target_attribute_code, relationship_type, strength, description)
VALUES
('PASSPORT_NUMBER', 'DATE_OF_BIRTH', 'validates', 0.9, 'Passport contains DOB for validation'),
('TAX_ID', 'TAX_RESIDENCY_COUNTRY', 'validates', 0.95, 'TIN format validates country'),
('REGISTRATION_NUMBER', 'INCORPORATION_JURISDICTION', 'validates', 0.95, 'Registration number format indicates jurisdiction'),
('LEI', 'REGISTERED_NAME', 'validates', 1.0, 'LEI record validates legal name');

-- Requirement relationships
INSERT INTO kyc_attribute_relationships (source_attribute_code, target_attribute_code, relationship_type, strength, description)
VALUES
('UBO_PERCENT', 'UBO_NAME', 'requires', 1.0, 'Ownership percentage requires owner name'),
('UBO_NAME', 'UBO_DOB', 'requires', 0.9, 'UBO identification requires date of birth'),
('UBO_NAME', 'UBO_NATIONALITY', 'requires', 0.9, 'UBO identification requires nationality'),
('FATCA_STATUS', 'TAX_RESIDENCY_COUNTRY', 'requires', 1.0, 'FATCA classification requires tax residency'),
('CRS_CLASSIFICATION', 'TAX_RESIDENCY_COUNTRY', 'requires', 1.0, 'CRS classification requires tax residency');

-- Conflict relationships
INSERT INTO kyc_attribute_relationships (source_attribute_code, target_attribute_code, relationship_type, strength, description)
VALUES
('SANCTIONED_COUNTRY_FLAG', 'ENTITY_ACTIVE_STATUS', 'conflicts_with', 0.7, 'Sanctioned entities may have restricted operations'),
('HIGH_RISK_JURISDICTION_FLAG', 'PEP_EXPOSURE_FLAG', 'related_to', 0.6, 'Both increase overall risk');

-- Semantic similarity relationships
INSERT INTO kyc_attribute_relationships (source_attribute_code, target_attribute_code, relationship_type, strength, description)
VALUES
('REGISTERED_NAME', 'FULL_NAME', 'related_to', 0.5, 'Both are primary identifiers'),
('UBO_NAME', 'SHAREHOLDER_NAME', 'related_to', 0.8, 'Both represent ownership'),
('UBO_PERCENT', 'SHAREHOLDER_PERCENT', 'related_to', 0.9, 'Different levels of ownership'),
('DIRECTOR_NAME', 'CONTROL_PERSON', 'related_to', 0.7, 'Both represent control'),
('SOURCE_OF_FUNDS', 'SOURCE_OF_WEALTH', 'related_to', 0.8, 'Both relate to financial origin'),
('SANCTIONS_STATUS', 'PEP_STATUS', 'related_to', 0.6, 'Both are critical risk indicators'),
('TAX_RESIDENCY_COUNTRY', 'INCORPORATION_JURISDICTION', 'related_to', 0.7, 'Both indicate geographic presence');

-- ==================== Verification Queries ====================

-- Count attributes with metadata
-- SELECT COUNT(*) FROM kyc_attribute_metadata;

-- Show all clusters with attribute counts
-- SELECT * FROM cluster_details;

-- Find attributes by synonym
-- SELECT * FROM find_attribute_by_synonym('Legal Name');

-- Get all attributes in a cluster
-- SELECT * FROM get_cluster_attributes('ENTITY_IDENTITY');

-- Find related attributes
-- SELECT * FROM find_related_attributes('TAX_RESIDENCY_COUNTRY', 2);

-- Show knowledge graph
-- SELECT * FROM attribute_knowledge_graph LIMIT 20;
