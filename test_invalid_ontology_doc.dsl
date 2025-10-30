(kyc-case TEST-INVALID-DOC
  (nature-purpose
    (nature "Test Case")
    (purpose "Test invalid document reference"))

  (client-business-unit OPERATIONS)

  (data-dictionary
    (attribute REGISTERED_NAME
      (primary-source (document CERT-INC)))
    (attribute UBO_NAME
      (primary-source (document W8BENZ)))
  )

  (document-requirements
    (jurisdiction EU)
    (required
      (document CERT-INC "Certificate of Incorporation")
      (document FAKE-DOC-123 "Non-existent Document")
    ))

  (kyc-token "pending")
)
