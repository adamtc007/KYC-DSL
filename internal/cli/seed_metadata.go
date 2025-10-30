package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/adamtc007/KYC-DSL/internal/model"
	"github.com/adamtc007/KYC-DSL/internal/ontology"
	"github.com/adamtc007/KYC-DSL/internal/rag"
	"github.com/adamtc007/KYC-DSL/internal/storage"
)

// RunSeedMetadataCommand seeds attribute metadata with embeddings
func RunSeedMetadataCommand() error {
	fmt.Println("🌱 Seeding Attribute Metadata with Embeddings...")
	fmt.Println("================================================")

	// Connect to database
	db, err := storage.ConnectPostgres()
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	// Initialize repositories and embedder
	repo := ontology.NewMetadataRepo(db)
	embedder := rag.NewEmbedder()
	ctx := context.Background()

	// Sample metadata to seed
	sampleMetadata := []model.AttributeMetadata{
		{
			AttributeCode:       "REGISTERED_NAME",
			Synonyms:            []string{"Legal Name", "Company Name", "Entity Name", "Corporate Name"},
			DataType:            "string",
			RiskLevel:           "LOW",
			RegulatoryCitations: []string{"AMLD5 Article 13", "MAS 626 Annex A", "Companies Act 2006 s.86"},
			ExampleValues:       []string{"BlackRock Global Funds", "HSBC Holdings PLC", "Deutsche Bank AG"},
			BusinessContext:     "Official legal name of the entity as registered with the competent authority. This is the primary identifier for legal entity identification across all jurisdictions.",
		},
		{
			AttributeCode:       "TAX_RESIDENCY_COUNTRY",
			Synonyms:            []string{"Tax Country", "Country of Tax Residence", "Tax Jurisdiction", "Fiscal Residence"},
			DataType:            "enum(ISO 3166-1 Alpha-2)",
			RiskLevel:           "HIGH",
			RegulatoryCitations: []string{"FATCA §1471(b)(1)(D)", "CRS Common Reporting Standard", "OECD Model Tax Convention"},
			ExampleValues:       []string{"US", "GB", "HK", "SG", "DE"},
			BusinessContext:     "Jurisdiction where the entity is considered tax resident under FATCA/CRS regulations. Critical for automatic exchange of information and tax reporting obligations.",
		},
		{
			AttributeCode:       "UBO_NAME",
			Synonyms:            []string{"Ultimate Beneficial Owner Name", "Beneficial Owner", "UBO", "Controller Name"},
			DataType:            "string",
			RiskLevel:           "CRITICAL",
			RegulatoryCitations: []string{"AMLD5 Article 3", "FATF Recommendation 24", "MAS 626"},
			ExampleValues:       []string{"John Smith", "Jane Doe", "Michael Chen"},
			BusinessContext:     "Full legal name of the ultimate beneficial owner who directly or indirectly owns or controls more than 25% of the entity or exercises control through other means.",
		},
		{
			AttributeCode:       "UBO_OWNERSHIP_PERCENT",
			Synonyms:            []string{"Ownership Percentage", "Beneficial Ownership %", "Control Percentage"},
			DataType:            "float",
			RiskLevel:           "CRITICAL",
			RegulatoryCitations: []string{"AMLD5 Article 3(6)", "FATF Recommendation 10"},
			ExampleValues:       []string{"26.5", "50.0", "100.0"},
			BusinessContext:     "Percentage of ownership or voting rights held by the ultimate beneficial owner. Threshold of 25% triggers reporting requirements under most AML regulations.",
		},
		{
			AttributeCode:       "INCORPORATION_COUNTRY",
			Synonyms:            []string{"Country of Incorporation", "Registration Country", "Legal Jurisdiction"},
			DataType:            "enum(ISO 3166-1 Alpha-2)",
			RiskLevel:           "MEDIUM",
			RegulatoryCitations: []string{"AMLD5", "MAS 626", "Companies Act"},
			ExampleValues:       []string{"US", "GB", "KY", "BM"},
			BusinessContext:     "Country where the legal entity was incorporated or registered. Determines applicable corporate law and regulatory oversight jurisdiction.",
		},
		{
			AttributeCode:       "PEP_STATUS",
			Synonyms:            []string{"Politically Exposed Person", "PEP Flag", "PEP Indicator", "Political Exposure"},
			DataType:            "boolean",
			RiskLevel:           "CRITICAL",
			RegulatoryCitations: []string{"AMLD5 Article 20", "FATF Recommendation 12", "MAS 626 Part VII"},
			ExampleValues:       []string{"true", "false"},
			BusinessContext:     "Indicator of whether the individual holds or has held a prominent public function. PEPs present higher money laundering and corruption risks requiring enhanced due diligence.",
		},
		{
			AttributeCode:       "SOURCE_OF_FUNDS",
			Synonyms:            []string{"Funds Source", "Origin of Funds", "Wealth Source"},
			DataType:            "enum",
			RiskLevel:           "CRITICAL",
			RegulatoryCitations: []string{"AMLD5", "FATF Recommendation 10", "MAS 626"},
			ExampleValues:       []string{"Salary", "Business Income", "Investment Returns", "Inheritance", "Sale of Assets"},
			BusinessContext:     "Origin and source of the funds being used in the business relationship. Critical for understanding the economic profile and detecting potential money laundering.",
		},
		{
			AttributeCode:       "SOURCE_OF_WEALTH",
			Synonyms:            []string{"Wealth Source", "Origin of Wealth", "Accumulated Wealth"},
			DataType:            "string",
			RiskLevel:           "HIGH",
			RegulatoryCitations: []string{"FATF Recommendation 10", "MAS 626"},
			ExampleValues:       []string{"Tech Industry Career", "Real Estate Development", "Family Business"},
			BusinessContext:     "Description of how the customer or UBO accumulated their total net worth. Broader than source of funds, encompasses lifetime wealth accumulation.",
		},
		{
			AttributeCode:       "BUSINESS_ACTIVITY",
			Synonyms:            []string{"Nature of Business", "Business Type", "Industry Sector", "Economic Activity"},
			DataType:            "string",
			RiskLevel:           "MEDIUM",
			RegulatoryCitations: []string{"AMLD5", "MAS 626 Annex A"},
			ExampleValues:       []string{"Asset Management", "Investment Banking", "Real Estate", "Technology Services"},
			BusinessContext:     "Primary business activity or industry sector of the entity. Used for risk assessment and determining appropriate regulatory treatment.",
		},
		{
			AttributeCode:       "REGISTERED_ADDRESS",
			Synonyms:            []string{"Legal Address", "Official Address", "Registered Office"},
			DataType:            "string",
			RiskLevel:           "LOW",
			RegulatoryCitations: []string{"Companies Act 2006", "AMLD5"},
			ExampleValues:       []string{"123 Main Street, London, UK", "456 Wall Street, New York, USA"},
			BusinessContext:     "Official registered address of the entity as recorded with the corporate registry. Required for legal correspondence and regulatory filings.",
		},
		{
			AttributeCode:       "DIRECTOR_NAME",
			Synonyms:            []string{"Board Member", "Director", "Board Director", "Company Director"},
			DataType:            "string",
			RiskLevel:           "HIGH",
			RegulatoryCitations: []string{"AMLD5", "Companies Act 2006", "MAS 626"},
			ExampleValues:       []string{"Michael Johnson", "Sarah Williams", "Robert Chen"},
			BusinessContext:     "Full legal name of a director on the board of the entity. Directors exercise significant control and are subject to background checks and sanctions screening.",
		},
		{
			AttributeCode:       "FATCA_STATUS",
			Synonyms:            []string{"FATCA Classification", "Chapter 4 Status", "US Tax Status"},
			DataType:            "enum",
			RiskLevel:           "HIGH",
			RegulatoryCitations: []string{"FATCA §1471-1474", "IRS Publication 5190"},
			ExampleValues:       []string{"Participating FFI", "Certified Deemed Compliant FFI", "Active NFFE"},
			BusinessContext:     "Entity classification under the Foreign Account Tax Compliance Act. Determines US withholding and reporting obligations for financial institutions.",
		},
		{
			AttributeCode:       "CRS_CLASSIFICATION",
			Synonyms:            []string{"CRS Status", "AEOI Classification", "Common Reporting Standard Type"},
			DataType:            "enum",
			RiskLevel:           "HIGH",
			RegulatoryCitations: []string{"CRS OECD Standard", "AEOI XML Schema v2.0"},
			ExampleValues:       []string{"Financial Institution", "Active NFE", "Passive NFE"},
			BusinessContext:     "Entity classification under the Common Reporting Standard for automatic exchange of financial account information. Determines international tax reporting requirements.",
		},
		{
			AttributeCode:       "SANCTIONS_SCREENING_STATUS",
			Synonyms:            []string{"Sanctions Check", "Watchlist Status", "SDN Status"},
			DataType:            "enum",
			RiskLevel:           "CRITICAL",
			RegulatoryCitations: []string{"OFAC Regulations", "EU Sanctions", "UN Security Council"},
			ExampleValues:       []string{"Clear", "Match", "Potential Match", "Under Review"},
			BusinessContext:     "Result of screening against global sanctions lists including OFAC SDN, EU sanctions, and UN lists. Critical for compliance with economic sanctions programs.",
		},
		{
			AttributeCode:       "ADVERSE_MEDIA_FLAG",
			Synonyms:            []string{"Negative News", "Adverse News", "Reputational Risk Flag"},
			DataType:            "boolean",
			RiskLevel:           "HIGH",
			RegulatoryCitations: []string{"FATF Recommendation 10", "MAS 626"},
			ExampleValues:       []string{"true", "false"},
			BusinessContext:     "Indicator of negative news coverage related to financial crimes, corruption, or regulatory violations. Used in enhanced due diligence and ongoing monitoring.",
		},
		{
			AttributeCode:       "EXPECTED_TRANSACTION_VOLUME",
			Synonyms:            []string{"Anticipated Volume", "Transaction Volume", "Expected Activity"},
			DataType:            "string",
			RiskLevel:           "MEDIUM",
			RegulatoryCitations: []string{"FATF Recommendation 10"},
			ExampleValues:       []string{"$1M-$10M annually", "$10M-$50M annually", ">$100M annually"},
			BusinessContext:     "Expected annual transaction volume or value for the business relationship. Used to establish baseline for transaction monitoring and anomaly detection.",
		},
		{
			AttributeCode:       "INDUSTRY_SECTOR",
			Synonyms:            []string{"Business Sector", "Economic Sector", "NAICS Code", "SIC Code"},
			DataType:            "enum",
			RiskLevel:           "MEDIUM",
			RegulatoryCitations: []string{"FATF Guidance"},
			ExampleValues:       []string{"Financial Services", "Healthcare", "Technology", "Energy"},
			BusinessContext:     "Standardized industry classification of the entity. Used for risk rating as certain sectors present higher ML/TF risks (e.g., MSBs, casinos, precious metals).",
		},
		{
			AttributeCode:       "CUSTOMER_RISK_RATING",
			Synonyms:            []string{"Risk Score", "Risk Level", "Client Risk Classification"},
			DataType:            "enum",
			RiskLevel:           "CRITICAL",
			RegulatoryCitations: []string{"FATF Recommendation 10", "MAS 626 Part VI"},
			ExampleValues:       []string{"Low", "Medium", "High", "Prohibited"},
			BusinessContext:     "Overall risk rating assigned to the customer based on risk assessment methodology. Determines level of due diligence, monitoring frequency, and approval authority required.",
		},
		{
			AttributeCode:       "RELATIONSHIP_START_DATE",
			Synonyms:            []string{"Onboarding Date", "Account Opening Date", "Relationship Date"},
			DataType:            "date",
			RiskLevel:           "LOW",
			RegulatoryCitations: []string{"Record Keeping Requirements"},
			ExampleValues:       []string{"2023-01-15", "2022-06-30"},
			BusinessContext:     "Date when the business relationship commenced. Used for calculating review cycles and data retention periods.",
		},
		{
			AttributeCode:       "LAST_REVIEW_DATE",
			Synonyms:            []string{"Last CDD Review", "Last Refresh Date", "Periodic Review Date"},
			DataType:            "date",
			RiskLevel:           "MEDIUM",
			RegulatoryCitations: []string{"FATF Recommendation 10", "MAS 626"},
			ExampleValues:       []string{"2024-01-15", "2023-12-01"},
			BusinessContext:     "Date of most recent periodic customer due diligence review. Risk-based approach determines review frequency (e.g., annually for high risk, every 3 years for low risk).",
		},
	}

	fmt.Printf("\n📊 Processing %d attributes...\n\n", len(sampleMetadata))

	successCount := 0
	errorCount := 0
	startTime := time.Now()

	for i, metadata := range sampleMetadata {
		fmt.Printf("[%d/%d] Processing: %s\n", i+1, len(sampleMetadata), metadata.AttributeCode)

		// Generate embedding
		embedding, err := embedder.GenerateEmbedding(ctx, metadata)
		if err != nil {
			fmt.Printf("  ❌ Failed to generate embedding: %v\n", err)
			errorCount++
			continue
		}

		metadata.Embedding = embedding

		// Upsert to database
		err = repo.UpsertMetadata(ctx, metadata)
		if err != nil {
			fmt.Printf("  ❌ Failed to save metadata: %v\n", err)
			errorCount++
			continue
		}

		fmt.Printf("  ✅ Seeded with %d-dimensional embedding\n", len(embedding))
		successCount++

		// Rate limiting
		if i < len(sampleMetadata)-1 {
			time.Sleep(200 * time.Millisecond)
		}
	}

	elapsed := time.Since(startTime)

	// Print summary
	fmt.Println("\n================================================")
	fmt.Println("📈 Seeding Summary")
	fmt.Println("================================================")
	fmt.Printf("✅ Successfully seeded: %d attributes\n", successCount)
	fmt.Printf("❌ Failed: %d attributes\n", errorCount)
	fmt.Printf("⏱️  Total time: %s\n", elapsed.Round(time.Millisecond))
	fmt.Printf("🚀 Average time per attribute: %s\n", (elapsed / time.Duration(len(sampleMetadata))).Round(time.Millisecond))

	// Get stats
	stats, err := repo.GetMetadataStats(ctx)
	if err != nil {
		return fmt.Errorf("failed to get stats: %w", err)
	}

	fmt.Println("\n================================================")
	fmt.Println("📊 Repository Statistics")
	fmt.Println("================================================")
	fmt.Printf("Total attributes with metadata: %v\n", stats["total_attributes"])
	fmt.Printf("Attributes with embeddings: %v\n", stats["attributes_with_embeddings"])
	fmt.Printf("Embedding coverage: %.1f%%\n", stats["embedding_coverage_percent"])

	fmt.Println("\n✅ Seeding complete! You can now run semantic searches.")
	fmt.Println("\nExample queries:")
	fmt.Println("  ./kycctl search-metadata \"tax residency\"")
	fmt.Println("  ./kycctl similar-attributes UBO_NAME")

	return nil
}
