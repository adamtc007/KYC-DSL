-- ===========================================================
-- ontology_seed.sql
-- Seed Data for Regulatory Data Ontology
-- Covers: US (FATCA), EU (AMLD5), APAC (Singapore MAS, Hong Kong HKMA)
-- ===========================================================

-- ==================== Regulations ====================

INSERT INTO kyc_regulations (code, name, jurisdiction, authority, description, effective_from)
VALUES
('FATCA', 'Foreign Account Tax Compliance Act', 'US', 'IRS', 'US tax residency and withholding compliance', '2014-07-01'),
('CRS', 'Common Reporting Standard', 'GLOBAL', 'OECD', 'Automatic exchange of financial account information', '2017-01-01'),
('AMLD5', '5th EU Anti-Money Laundering Directive', 'EU', 'European Commission', 'Customer Due Diligence and Beneficial Ownership', '2020-01-10'),
('AMLD6', '6th EU Anti-Money Laundering Directive', 'EU', 'European Commission', 'Extended criminal liability for AML offenses', '2021-12-03'),
('MAS626', 'MAS Notice 626', 'SG', 'Monetary Authority of Singapore', 'AML/CFT requirements for financial institutions', '2015-04-01'),
('HKMAAML', 'HKMA AML Guideline', 'HK', 'Hong Kong Monetary Authority', 'AML and CDD guidelines', '2018-06-01'),
('UKMLR2017', 'UK Money Laundering Regulations 2017', 'UK', 'HM Treasury', 'UK AML and CTF requirements', '2017-06-26'),
('BSAAML', 'Bank Secrecy Act / AML', 'US', 'FinCEN', 'US AML compliance requirements', '1970-10-26');

-- ==================== Documents ====================

INSERT INTO kyc_documents (code, name, domain, jurisdiction, regulation_code, source_type, validity_years, description)
VALUES
-- Tax Documents
('W8BEN', 'IRS Form W-8BEN', 'Tax', 'US', 'FATCA', 'Client', 3, 'Self-certification of tax residency for non-US persons'),
('W8BENE', 'IRS Form W-8BEN-E', 'Tax', 'US', 'FATCA', 'Client', 3, 'Self-certification of tax residency for entities'),
('W9', 'IRS Form W-9', 'Tax', 'US', 'FATCA', 'Client', 3, 'US taxpayer identification form'),
('CRS-SELF-CERT', 'CRS Self-Certification', 'Tax', 'GLOBAL', 'CRS', 'Client', 3, 'OECD CRS tax residency self-certification'),

-- Entity Documents
('CERT-INC', 'Certificate of Incorporation', 'Entity', 'GLOBAL', 'AMLD5', 'Official', 0, 'Proof of entity legal formation'),
('CERT-GOOD-STANDING', 'Certificate of Good Standing', 'Entity', 'GLOBAL', 'AMLD5', 'Official', 1, 'Proof of active legal status'),
('BR-CERT', 'Business Registration Certificate', 'Entity', 'HK', 'HKMAAML', 'Official', 1, 'Proof of active business registration'),
('ACRA-PROFILE', 'ACRA Business Profile', 'Entity', 'SG', 'MAS626', 'Official', 1, 'Singapore company registry profile'),
('COMPANIES-HOUSE-CERT', 'Companies House Certificate', 'Entity', 'UK', 'UKMLR2017', 'Official', 1, 'UK company registration certificate'),
('ARTICLES-ASSOC', 'Articles of Association', 'Entity', 'GLOBAL', 'AMLD5', 'Official', 0, 'Entity governing documents'),
('MEMORANDUM-ASSOC', 'Memorandum of Association', 'Entity', 'GLOBAL', 'AMLD5', 'Official', 0, 'Entity formation documents'),

-- Ownership Documents
('UBO-DECL', 'Ultimate Beneficial Owner Declaration', 'Ownership', 'GLOBAL', 'AMLD5', 'Client', 1, 'Declaration of beneficial owners >25%'),
('SHARE-REGISTER', 'Share Register', 'Ownership', 'GLOBAL', 'AMLD5', 'Official', 1, 'Official shareholder registry'),
('OWNERSHIP-CHART', 'Ownership Structure Chart', 'Ownership', 'GLOBAL', 'AMLD5', 'Client', 1, 'Visual representation of ownership hierarchy'),
('PSC-REGISTER', 'Persons with Significant Control Register', 'Ownership', 'UK', 'UKMLR2017', 'Official', 1, 'UK PSC filing'),

-- Individual Identity Documents
('PASSPORT', 'Passport', 'Identity', 'GLOBAL', 'AMLD5', 'Official', 5, 'Government-issued passport'),
('NATIONAL-ID', 'National Identity Card', 'Identity', 'GLOBAL', 'AMLD5', 'Official', 5, 'Government-issued national ID'),
('DRIVERS-LICENSE', 'Drivers License', 'Identity', 'GLOBAL', 'AMLD5', 'Official', 5, 'Government-issued drivers license'),

-- Address Verification Documents
('UTILITY-BILL', 'Utility Bill', 'Address', 'GLOBAL', 'AMLD5', 'Client', 0, 'Recent utility bill as proof of address'),
('BANK-STATEMENT', 'Bank Statement', 'Address', 'GLOBAL', 'AMLD5', 'Client', 0, 'Recent bank statement as proof of address'),
('COUNCIL-TAX-BILL', 'Council Tax Bill', 'Address', 'UK', 'UKMLR2017', 'Official', 0, 'UK council tax bill'),

-- Financial Documents
('AUDITED-FINANCIALS', 'Audited Financial Statements', 'Financial', 'GLOBAL', 'AMLD5', 'Official', 1, 'Audited annual financial statements'),
('TAX-RETURN', 'Tax Return', 'Financial', 'GLOBAL', 'AMLD5', 'Official', 1, 'Filed tax return'),
('SOURCE-WEALTH-LETTER', 'Source of Wealth Letter', 'Financial', 'GLOBAL', 'AMLD5', 'Client', 1, 'Declaration of source of wealth'),

-- Control and Management
('BOARD-RESOLUTION', 'Board Resolution', 'Control', 'GLOBAL', 'AMLD5', 'Client', 1, 'Board authorization for account opening'),
('POA', 'Power of Attorney', 'Control', 'GLOBAL', 'AMLD5', 'Official', 3, 'Legal power of attorney document'),
('DIRECTOR-LIST', 'List of Directors', 'Control', 'GLOBAL', 'AMLD5', 'Official', 1, 'Official list of company directors');

-- ==================== Attributes ====================

INSERT INTO kyc_attributes (code, name, domain, description, risk_category, is_personal_data)
VALUES
-- Entity Attributes
('REGISTERED_NAME', 'Registered Legal Name', 'Entity', 'Official registered entity name', 'LOW', FALSE),
('INCORPORATION_DATE', 'Date of Incorporation', 'Entity', 'Date entity was legally formed', 'LOW', FALSE),
('INCORPORATION_JURISDICTION', 'Jurisdiction of Incorporation', 'Entity', 'Legal jurisdiction where entity was formed', 'LOW', FALSE),
('REGISTRATION_NUMBER', 'Company Registration Number', 'Entity', 'Official company registration number', 'LOW', FALSE),
('LEI', 'Legal Entity Identifier', 'Entity', 'LEI code assigned by global LEI foundation', 'LOW', FALSE),
('ENTITY_TYPE', 'Entity Type', 'Entity', 'Legal form (Corporation, LLC, Trust, etc.)', 'LOW', FALSE),
('REGISTERED_ADDRESS', 'Registered Office Address', 'Entity', 'Official registered address', 'LOW', FALSE),
('BUSINESS_ADDRESS', 'Principal Business Address', 'Entity', 'Main place of business operations', 'LOW', FALSE),

-- Tax Attributes
('TAX_RESIDENCY_COUNTRY', 'Tax Residency Country', 'Tax', 'Jurisdiction of tax obligation', 'MEDIUM', FALSE),
('TAX_ID', 'Tax Identification Number', 'Tax', 'Official tax ID (TIN, SSN, etc.)', 'HIGH', TRUE),
('US_TAX_STATUS', 'US Tax Person Status', 'Tax', 'Whether entity/person is US tax person', 'HIGH', FALSE),
('FATCA_STATUS', 'FATCA Classification', 'Tax', 'FATCA entity classification', 'MEDIUM', FALSE),
('CRS_CLASSIFICATION', 'CRS Entity Classification', 'Tax', 'CRS reporting classification', 'MEDIUM', FALSE),

-- Ownership Attributes
('UBO_NAME', 'Ultimate Beneficial Owner Name', 'Ownership', 'Name of beneficial owner', 'HIGH', TRUE),
('UBO_PERCENT', 'Ownership Percentage', 'Ownership', 'Shareholding percentage of beneficial owner', 'HIGH', FALSE),
('UBO_DOB', 'UBO Date of Birth', 'Ownership', 'Date of birth of beneficial owner', 'HIGH', TRUE),
('UBO_NATIONALITY', 'UBO Nationality', 'Ownership', 'Nationality of beneficial owner', 'HIGH', TRUE),
('UBO_ADDRESS', 'UBO Residential Address', 'Ownership', 'Residential address of beneficial owner', 'HIGH', TRUE),
('SHAREHOLDER_NAME', 'Shareholder Name', 'Ownership', 'Name of direct shareholder', 'HIGH', TRUE),
('SHAREHOLDER_PERCENT', 'Shareholder Percentage', 'Ownership', 'Direct shareholding percentage', 'MEDIUM', FALSE),

-- Control Attributes
('DIRECTOR_NAME', 'Director Name', 'Control', 'Name of company director', 'HIGH', TRUE),
('AUTHORIZED_SIGNATORY', 'Authorized Signatory Name', 'Control', 'Person authorized to sign on behalf of entity', 'HIGH', TRUE),
('CONTROL_PERSON', 'Person with Significant Control', 'Control', 'UK PSC definition (>25% or control)', 'HIGH', TRUE),

-- Identity Attributes (Individuals)
('FULL_NAME', 'Full Legal Name', 'Identity', 'Complete legal name of individual', 'HIGH', TRUE),
('DATE_OF_BIRTH', 'Date of Birth', 'Identity', 'Individual date of birth', 'HIGH', TRUE),
('NATIONALITY', 'Nationality', 'Identity', 'Individual nationality/citizenship', 'MEDIUM', TRUE),
('PASSPORT_NUMBER', 'Passport Number', 'Identity', 'Passport identification number', 'HIGH', TRUE),
('ID_NUMBER', 'Identity Document Number', 'Identity', 'Government ID number', 'HIGH', TRUE),

-- Financial Risk Attributes
('SOURCE_OF_FUNDS', 'Source of Funds', 'Financial', 'Origin of funds for transactions', 'HIGH', FALSE),
('SOURCE_OF_WEALTH', 'Source of Wealth', 'Financial', 'Origin of overall wealth', 'HIGH', FALSE),
('EXPECTED_TRANSACTION_VOLUME', 'Expected Transaction Volume', 'Financial', 'Anticipated transaction volume', 'MEDIUM', FALSE),
('BUSINESS_ACTIVITY', 'Nature of Business Activity', 'Financial', 'Description of business operations', 'MEDIUM', FALSE),

-- Risk Scoring
('PEP_STATUS', 'Politically Exposed Person Status', 'Risk', 'Whether individual is PEP', 'HIGH', TRUE),
('SANCTIONS_STATUS', 'Sanctions Screening Status', 'Risk', 'Result of sanctions list screening', 'HIGH', FALSE),
('ADVERSE_MEDIA', 'Adverse Media Check', 'Risk', 'Negative news screening result', 'MEDIUM', FALSE),
('RISK_RATING', 'Overall Risk Rating', 'Risk', 'Calculated risk rating (Low/Medium/High)', 'HIGH', FALSE);

-- ==================== Attribute ↔ Document Mappings ====================

INSERT INTO kyc_attr_doc_links (attribute_code, document_code, source_tier, is_mandatory, regulation_code, notes)
VALUES
-- Entity Attributes
('REGISTERED_NAME', 'CERT-INC', 'Primary', TRUE, 'AMLD5', 'Primary legal name source'),
('REGISTERED_NAME', 'BR-CERT', 'Primary', TRUE, 'HKMAAML', 'HK business registration'),
('REGISTERED_NAME', 'ACRA-PROFILE', 'Primary', TRUE, 'MAS626', 'Singapore registry'),
('INCORPORATION_DATE', 'CERT-INC', 'Primary', TRUE, 'AMLD5', NULL),
('INCORPORATION_JURISDICTION', 'CERT-INC', 'Primary', TRUE, 'AMLD5', NULL),
('REGISTRATION_NUMBER', 'CERT-INC', 'Primary', TRUE, 'AMLD5', NULL),
('REGISTRATION_NUMBER', 'BR-CERT', 'Primary', TRUE, 'HKMAAML', NULL),
('REGISTRATION_NUMBER', 'ACRA-PROFILE', 'Primary', TRUE, 'MAS626', NULL),
('ENTITY_TYPE', 'CERT-INC', 'Primary', TRUE, 'AMLD5', NULL),
('ENTITY_TYPE', 'ARTICLES-ASSOC', 'Secondary', FALSE, 'AMLD5', NULL),
('REGISTERED_ADDRESS', 'CERT-INC', 'Primary', TRUE, 'AMLD5', NULL),
('REGISTERED_ADDRESS', 'CERT-GOOD-STANDING', 'Secondary', FALSE, 'AMLD5', NULL),

-- Tax Attributes
('TAX_RESIDENCY_COUNTRY', 'W8BEN', 'Primary', TRUE, 'FATCA', 'Non-US persons'),
('TAX_RESIDENCY_COUNTRY', 'W8BENE', 'Primary', TRUE, 'FATCA', 'Non-US entities'),
('TAX_RESIDENCY_COUNTRY', 'W9', 'Primary', TRUE, 'FATCA', 'US persons'),
('TAX_RESIDENCY_COUNTRY', 'CRS-SELF-CERT', 'Primary', TRUE, 'CRS', NULL),
('TAX_ID', 'W8BEN', 'Primary', TRUE, 'FATCA', NULL),
('TAX_ID', 'W9', 'Primary', TRUE, 'FATCA', NULL),
('TAX_ID', 'CRS-SELF-CERT', 'Secondary', TRUE, 'CRS', NULL),
('US_TAX_STATUS', 'W8BEN', 'Primary', TRUE, 'FATCA', NULL),
('US_TAX_STATUS', 'W9', 'Primary', TRUE, 'FATCA', NULL),
('FATCA_STATUS', 'W8BENE', 'Primary', TRUE, 'FATCA', NULL),

-- Ownership Attributes
('UBO_NAME', 'UBO-DECL', 'Primary', TRUE, 'AMLD5', NULL),
('UBO_NAME', 'SHARE-REGISTER', 'Secondary', FALSE, 'AMLD5', 'Official corroboration'),
('UBO_PERCENT', 'UBO-DECL', 'Primary', TRUE, 'AMLD5', NULL),
('UBO_PERCENT', 'SHARE-REGISTER', 'Primary', TRUE, 'AMLD5', NULL),
('UBO_DOB', 'UBO-DECL', 'Primary', TRUE, 'AMLD5', NULL),
('UBO_NATIONALITY', 'UBO-DECL', 'Primary', TRUE, 'AMLD5', NULL),
('UBO_ADDRESS', 'UBO-DECL', 'Primary', TRUE, 'AMLD5', NULL),
('UBO_ADDRESS', 'UTILITY-BILL', 'Secondary', FALSE, 'AMLD5', 'Address verification'),
('SHAREHOLDER_NAME', 'SHARE-REGISTER', 'Primary', TRUE, 'AMLD5', NULL),
('SHAREHOLDER_PERCENT', 'SHARE-REGISTER', 'Primary', TRUE, 'AMLD5', NULL),
('CONTROL_PERSON', 'PSC-REGISTER', 'Primary', TRUE, 'UKMLR2017', 'UK PSC requirement'),

-- Control Attributes
('DIRECTOR_NAME', 'DIRECTOR-LIST', 'Primary', TRUE, 'AMLD5', NULL),
('DIRECTOR_NAME', 'CERT-GOOD-STANDING', 'Secondary', FALSE, 'AMLD5', NULL),
('AUTHORIZED_SIGNATORY', 'BOARD-RESOLUTION', 'Primary', TRUE, 'AMLD5', NULL),
('AUTHORIZED_SIGNATORY', 'POA', 'Primary', TRUE, 'AMLD5', 'If using POA'),

-- Identity Attributes
('FULL_NAME', 'PASSPORT', 'Primary', TRUE, 'AMLD5', NULL),
('FULL_NAME', 'NATIONAL-ID', 'Primary', TRUE, 'AMLD5', NULL),
('DATE_OF_BIRTH', 'PASSPORT', 'Primary', TRUE, 'AMLD5', NULL),
('DATE_OF_BIRTH', 'NATIONAL-ID', 'Primary', TRUE, 'AMLD5', NULL),
('NATIONALITY', 'PASSPORT', 'Primary', TRUE, 'AMLD5', NULL),
('PASSPORT_NUMBER', 'PASSPORT', 'Primary', TRUE, 'AMLD5', NULL),
('ID_NUMBER', 'NATIONAL-ID', 'Primary', TRUE, 'AMLD5', NULL),
('ID_NUMBER', 'DRIVERS-LICENSE', 'Secondary', FALSE, 'AMLD5', NULL),

-- Financial Attributes
('SOURCE_OF_WEALTH', 'SOURCE-WEALTH-LETTER', 'Primary', TRUE, 'AMLD5', NULL),
('SOURCE_OF_WEALTH', 'AUDITED-FINANCIALS', 'Secondary', FALSE, 'AMLD5', 'Corroborating evidence'),
('SOURCE_OF_FUNDS', 'SOURCE-WEALTH-LETTER', 'Primary', TRUE, 'AMLD5', NULL),
('SOURCE_OF_FUNDS', 'BANK-STATEMENT', 'Secondary', FALSE, 'AMLD5', NULL),
('BUSINESS_ACTIVITY', 'CERT-INC', 'Secondary', FALSE, 'AMLD5', NULL),
('BUSINESS_ACTIVITY', 'ARTICLES-ASSOC', 'Primary', TRUE, 'AMLD5', NULL);

-- ==================== Document ↔ Regulation Links ====================

INSERT INTO kyc_doc_reg_links (document_code, regulation_code, applicability, jurisdiction)
VALUES
-- FATCA Requirements
('W8BEN', 'FATCA', 'Non-US Individual', 'US'),
('W8BENE', 'FATCA', 'Non-US Entity', 'US'),
('W9', 'FATCA', 'US Person/Entity', 'US'),

-- CRS Requirements
('CRS-SELF-CERT', 'CRS', 'All Reportable Persons', 'GLOBAL'),

-- AMLD5 Requirements (EU)
('CERT-INC', 'AMLD5', 'All Legal Entities', 'EU'),
('UBO-DECL', 'AMLD5', 'All Legal Entities', 'EU'),
('SHARE-REGISTER', 'AMLD5', 'Corporations', 'EU'),
('PASSPORT', 'AMLD5', 'All Individuals', 'EU'),
('UTILITY-BILL', 'AMLD5', 'All Individuals', 'EU'),
('SOURCE-WEALTH-LETTER', 'AMLD5', 'High Risk Clients', 'EU'),

-- UK Requirements
('COMPANIES-HOUSE-CERT', 'UKMLR2017', 'UK Entities', 'UK'),
('PSC-REGISTER', 'UKMLR2017', 'UK Entities', 'UK'),

-- Singapore Requirements
('ACRA-PROFILE', 'MAS626', 'Singapore Entities', 'SG'),
('CERT-INC', 'MAS626', 'Foreign Entities', 'SG'),
('UBO-DECL', 'MAS626', 'All Entities', 'SG'),

-- Hong Kong Requirements
('BR-CERT', 'HKMAAML', 'HK Entities', 'HK'),
('UBO-DECL', 'HKMAAML', 'All Entities', 'HK'),
('PASSPORT', 'HKMAAML', 'All Individuals', 'HK');
