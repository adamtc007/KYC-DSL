(kyc-case TEST-VALID-MULTI-OWNER
  (nature-purpose
    (nature "Test case for validation")
    (purpose "Valid case with multiple owners and controller"))
  (client-business-unit TEST-CBU)
  (policy KYCPOL-UK-2025)
  (function BUILD-OWNERSHIP-TREE)
  (ownership-structure
    (owner COMPANY-A 60)
    (owner COMPANY-B 40)
    (controller JANE-DOE "Senior Managing Official")
    (controller JOHN-SMITH "Director"))
  (kyc-token "pending"))
