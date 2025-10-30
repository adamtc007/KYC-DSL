(kyc-case TEST-INVALID-NO-CONTROLLER
  (nature-purpose
    (nature "Test case for validation")
    (purpose "Should fail - multiple owners without controller"))
  (client-business-unit TEST-CBU)
  (function BUILD-OWNERSHIP-TREE)
  (ownership-structure
    (owner COMPANY-A 60)
    (owner COMPANY-B 40))
  (kyc-token "pending"))
