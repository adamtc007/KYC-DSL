(kyc-case BLACKROCK-GLOBAL-EQUITY-FUND
  (nature-purpose
    (nature "Institutional Fund Management")
    (purpose "EU Equity Fund KYC with Risk Assessment"))

  (client-business-unit FUND-SERVICES-EU)

  (policy KYCPOL-EU-2025)

  ; Public attributes with their document sources
  (data-dictionary
    (attribute REGISTERED_NAME
      (primary-source (document CERT-INC))
      (tertiary-source "Ops Validation"))
    (attribute TAX_RESIDENCY_COUNTRY
      (primary-source (document W8BENE))
      (secondary-source (document CRS-SELF-CERT)))
    (attribute INCORPORATION_JURISDICTION
      (primary-source (document CERT-INC)))
    (attribute INCORPORATION_DATE
      (primary-source (document CERT-INC)))
    (attribute UBO_NAME
      (primary-source (document UBO-DECL))
      (secondary-source (document SHARE-REGISTER)))
    (attribute UBO_PERCENT
      (primary-source (document UBO-DECL))
      (primary-source (document SHARE-REGISTER)))
    (attribute PEP_STATUS
      (primary-source (document UBO-DECL)))
  )

  ; Private (derived) attributes with lineage and rules
  (derived-attributes
    ; Risk flag: High-risk jurisdiction
    (attribute HIGH_RISK_JURISDICTION_FLAG
      (sources (TAX_RESIDENCY_COUNTRY))
      (rule "(if (in TAX_RESIDENCY_COUNTRY ['IR' 'KP' 'SY' 'YE' 'AF' 'MM']) true false)")
      (jurisdiction GLOBAL)
      (regulation AMLD5)
    )

    ; Risk flag: Sanctioned country exposure
    (attribute SANCTIONED_COUNTRY_FLAG
      (sources (TAX_RESIDENCY_COUNTRY INCORPORATION_JURISDICTION))
      (rule "(if (or (in TAX_RESIDENCY_COUNTRY ['IR' 'KP' 'SY' 'CU' 'RU']) (in INCORPORATION_JURISDICTION ['IR' 'KP' 'SY' 'CU' 'RU'])) true false)")
      (jurisdiction GLOBAL)
      (regulation BSAAML)
    )

    ; Risk flag: PEP exposure
    (attribute PEP_EXPOSURE_FLAG
      (sources (PEP_STATUS))
      (rule "(if (= PEP_STATUS true) true false)")
      (jurisdiction GLOBAL)
      (regulation AMLD5)
    )

    ; Ownership concentration metric
    (attribute UBO_CONCENTRATION_SCORE
      (sources (UBO_PERCENT))
      (rule "(max UBO_PERCENT)")
      (jurisdiction GLOBAL)
      (regulation AMLD5)
    )

    ; Entity age calculation
    (attribute ENTITY_AGE_YEARS
      (sources (INCORPORATION_DATE))
      (rule "(- (year (now)) (year INCORPORATION_DATE))")
      (jurisdiction GLOBAL)
    )

    ; Numeric risk score based on jurisdiction
    (attribute JURISDICTION_RISK_SCORE
      (sources (TAX_RESIDENCY_COUNTRY))
      (rule "(case TAX_RESIDENCY_COUNTRY (['IR' 'KP' 'SY'] 100) (['AF' 'YE' 'MM'] 90) (['RU' 'BY'] 80) (['CN' 'HK'] 60) (['US' 'GB' 'SG'] 20) (['CH' 'DE' 'FR'] 10) (else 50))")
      (jurisdiction GLOBAL)
      (regulation AMLD5)
    )

    ; Entity active status check
    (attribute ENTITY_ACTIVE_STATUS
      (sources (REGISTERED_NAME INCORPORATION_JURISDICTION))
      (rule "(registry-active? REGISTERED_NAME INCORPORATION_JURISDICTION)")
      (jurisdiction GLOBAL)
    )
  )

  ; Document requirements by jurisdiction
  (document-requirements
    (jurisdiction EU)
    (required
      (document CERT-INC "Certificate of Incorporation")
      (document UBO-DECL "Ultimate Beneficial Owner Declaration")
      (document W8BENE "IRS Form W-8BEN-E")
      (document SHARE-REGISTER "Share Register")
      (document AUDITED-FINANCIALS "Audited Financial Statements")
    ))

  (document-requirements
    (jurisdiction GLOBAL)
    (required
      (document CRS-SELF-CERT "CRS Self-Certification")
    ))

  ; Compliance functions
  (function DISCOVER-POLICIES)
  (function SOLICIT-DOCUMENTS)
  (function BUILD-OWNERSHIP-TREE)
  (function VERIFY-OWNERSHIP)
  (function ASSESS-RISK)

  ; Ownership structure with public data
  (ownership-structure
    (entity BLACKROCK-GLOBAL-FUNDS)
    (owner BLACKROCK-PLC 100%)
    (beneficial-owner LARRY-FINK 35%)
    (beneficial-owner INSTITUTIONAL-INVESTORS 45%)
    (beneficial-owner VANGUARD-GROUP 20%)
    (controller JANE-DOE "Senior Managing Official")
    (controller JOHN-SMITH "Director")
  )

  ; Obligations
  (obligation OBL-W8BEN)
  (obligation OBL-UBO-DECLARATION)
  (obligation OBL-PEP-001)

  (kyc-token "pending")
)
