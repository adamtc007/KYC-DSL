(kyc-case BLACKROCK-GLOBAL-EQUITY-FUND
  (nature-purpose
    (nature "Institutional Fund Management")
    (purpose "EU Equity Fund KYC for Institutional Client"))

  (client-business-unit FUND-SERVICES-EU)

  (policy KYCPOL-EU-2025)
  (policy AML-GLOBAL-BASE)

  (data-dictionary
    (attribute REGISTERED_NAME
      (primary-source (document CERT-INC))
      (tertiary-source "Ops Validation"))
    (attribute UBO_NAME
      (primary-source (document UBO-DECL))
      (secondary-source (document SHARE-REGISTER)))
    (attribute TAX_RESIDENCY_COUNTRY
      (primary-source (document W8BENE))
      (secondary-source (document CRS-SELF-CERT)))
    (attribute UBO_PERCENT
      (primary-source (document UBO-DECL))
      (primary-source (document SHARE-REGISTER)))
  )

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

  (function DISCOVER-POLICIES)
  (function SOLICIT-DOCUMENTS)
  (function BUILD-OWNERSHIP-TREE)
  (function VERIFY-OWNERSHIP)
  (function ASSESS-RISK)

  (ownership-structure
    (entity BLACKROCK-GLOBAL-FUNDS)
    (owner BLACKROCK-PLC 100%)
    (beneficial-owner LARRY-FINK 35%)
    (beneficial-owner INSTITUTIONAL-INVESTORS 45%)
    (controller JANE-DOE "Senior Managing Official")
    (controller JOHN-SMITH "Director")
  )

  (obligation OBL-W8BEN)
  (obligation OBL-UBO-DECLARATION)
  (obligation OBL-PEP-001)

  (kyc-token "pending")
)
