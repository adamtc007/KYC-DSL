(kyc-case TEST-INVALID-SUM
  (nature-purpose
    (nature "Test case for validation")
    (purpose "Should fail - ownership sum exceeds 100%"))
  (client-business-unit TEST-CBU)
  (function BUILD-OWNERSHIP-TREE)
  (ownership-structure
    (owner COMPANY-A 70)
    (owner COMPANY-B 50)
    (controller JANE-DOE "Director"))
  (kyc-token "pending"))
