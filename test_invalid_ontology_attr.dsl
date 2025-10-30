(kyc-case TEST-INVALID-ATTR
  (nature-purpose
    (nature "Test Case")
    (purpose "Test invalid attribute reference"))

  (client-business-unit OPERATIONS)

  (data-dictionary
    (attribute REGISTERED_NAME
      (primary-source (document CERT-INC)))
    (attribute FAKE_ATTRIBUTE_XYZ
      (primary-source (document UBO-DECL)))
  )

  (document-requirements
    (jurisdiction EU)
    (required
      (document CERT-INC "Certificate of Incorporation")
      (document UBO-DECL "Ultimate Beneficial Owner Declaration")
    ))

  (kyc-token "pending")
)
