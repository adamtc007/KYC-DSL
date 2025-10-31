package cli

import (
	"fmt"
	"log"
	"os"
	"strings"

	pb "github.com/adamtc007/KYC-DSL/api/pb"
	"github.com/adamtc007/KYC-DSL/internal/amend"
	"github.com/adamtc007/KYC-DSL/internal/model"
	"github.com/adamtc007/KYC-DSL/internal/ontology"
	"github.com/adamtc007/KYC-DSL/internal/rustclient"
	"github.com/adamtc007/KYC-DSL/internal/storage"
)

// RunGrammarCommand stores the current grammar definition in the database.
func RunGrammarCommand() error {
	// Connect to Rust DSL service to get grammar
	rustClient, err := rustclient.NewDslClient("")
	if err != nil {
		return fmt.Errorf("failed to connect to Rust DSL service: %w", err)
	}
	defer rustClient.Close()

	grammarResp, err := rustClient.GetGrammar()
	if err != nil {
		return fmt.Errorf("failed to get grammar from Rust service: %w", err)
	}

	// Connect to database
	db, err := storage.ConnectPostgres()
	if err != nil {
		return fmt.Errorf("database connection failed: %w", err)
	}
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			log.Printf("failed to close database: %v", closeErr)
		}
	}()

	// Store grammar in database
	err = storage.InsertGrammar(db, "KYC-DSL", grammarResp.Version, grammarResp.Ebnf)
	if err != nil {
		return fmt.Errorf("insert grammar failed: %w", err)
	}

	fmt.Printf("✅ Grammar (v%s) inserted into Postgres via Rust service.\n", grammarResp.Version)
	return nil
}

// RunProcessCommand parses, validates, and persists a DSL file via Rust service.
func RunProcessCommand(filePath string) error {
	// Read DSL file
	dslContent, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Connect to Rust DSL service
	rustClient, err := rustclient.NewDslClient("")
	if err != nil {
		return fmt.Errorf("failed to connect to Rust DSL service: %w", err)
	}
	defer rustClient.Close()

	dslText := string(dslContent)

	// Parse via Rust
	parseResp, err := rustClient.ParseDSL(dslText)
	if err != nil {
		return fmt.Errorf("parse error: %w", err)
	}
	if !parseResp.Success {
		return fmt.Errorf("parse failed: %s (errors: %v)", parseResp.Message, parseResp.Errors)
	}

	if len(parseResp.Cases) == 0 {
		return fmt.Errorf("no cases found in DSL")
	}

	// Validate via Rust
	valResult, err := rustClient.ValidateDSL(dslText)
	if err != nil {
		return fmt.Errorf("validation error: %w", err)
	}
	if !valResult.Valid {
		return fmt.Errorf("❌ DSL validation failed: %v", valResult.Errors)
	}
	fmt.Println("✅ DSL validated successfully (grammar + semantics) via Rust service.")

	// Connect to database for persistence
	db, err := storage.ConnectPostgres()
	if err != nil {
		return fmt.Errorf("database error: %w", err)
	}
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			log.Printf("WARNING: failed to close database: %v", closeErr)
		}
	}()

	// Extract case information
	caseName := parseResp.Cases[0].Name
	displayParsedCaseInfo(parseResp.Cases[0])

	// Save to database
	if err := storage.SaveCaseVersion(db, caseName, dslText); err != nil {
		return fmt.Errorf("failed to save case: %w", err)
	}

	fmt.Printf("\n🧾 DSL snapshot stored and versioned successfully (case: %s)\n", caseName)
	return nil
}

// RunValidateCommand validates an existing case and records audit trail.
func RunValidateCommand(caseName, actor string) error {
	db, err := storage.ConnectPostgres()
	if err != nil {
		return fmt.Errorf("database connection failed: %w", err)
	}
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			log.Printf("WARNING: failed to close database: %v", closeErr)
		}
	}()

	// Load most recent version
	dsl, err := storage.GetLatestDSL(db, caseName)
	if err != nil {
		return fmt.Errorf("failed to load case: %w", err)
	}

	// Connect to Rust DSL service
	rustClient, err := rustclient.NewDslClient("")
	if err != nil {
		return fmt.Errorf("failed to connect to Rust DSL service: %w", err)
	}
	defer rustClient.Close()

	// Validate via Rust
	valResult, err := rustClient.ValidateDSL(dsl)
	if err != nil {
		return fmt.Errorf("validation error: %w", err)
	}
	if !valResult.Valid {
		return fmt.Errorf("validation failed: %v", valResult.Errors)
	}

	fmt.Printf("✅ Case %s validated via Rust service.\n", caseName)
	return nil
}

// RunAmendCommand applies an incremental amendment to an existing case via Rust service.
func RunAmendCommand(caseName, step string) error {
	db, err := storage.ConnectPostgres()
	if err != nil {
		return fmt.Errorf("database connection failed: %w", err)
	}
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			log.Printf("WARNING: failed to close database: %v", closeErr)
		}
	}()

	// Special handling for ontology-aware amendments that need DB access
	if step == "document-discovery" {
		// This one needs the ontology repo, so use the existing amend package
		repo := ontology.NewRepository(db)
		mutation := func(c *model.KycCase) {
			if err := amend.AddDocumentDiscovery(c, repo); err != nil {
				log.Printf("Error in document discovery: %v", err)
			}
		}
		if err := amend.ApplyAmendment(db, caseName, step, mutation); err != nil {
			return fmt.Errorf("amendment failed: %w", err)
		}
		fmt.Printf("✅ Amendment '%s' applied successfully to case %s\n", step, caseName)
		return nil
	}

	// For all other amendments, use Rust service
	rustClient, err := rustclient.NewDslClient("")
	if err != nil {
		return fmt.Errorf("failed to connect to Rust DSL service: %w", err)
	}
	defer rustClient.Close()

	// Apply amendment via Rust
	amendResp, err := rustClient.AmendCase(caseName, step)
	if err != nil {
		return fmt.Errorf("amendment RPC failed: %w", err)
	}
	if !amendResp.Success {
		return fmt.Errorf("amendment failed: %s", amendResp.Message)
	}

	// Save new version to database
	if err := storage.SaveCaseVersion(db, caseName, amendResp.UpdatedDsl); err != nil {
		return fmt.Errorf("failed to save amended version: %w", err)
	}

	// Log amendment
	if err := storage.InsertAmendment(db, caseName, step, "rust-applied", amendResp.Message); err != nil {
		log.Printf("Warning: failed to log amendment: %v", err)
	}

	fmt.Printf("✅ Amendment '%s' applied successfully to case %s (via Rust service)\n", step, caseName)
	return nil
}

// RunOntologyCommand displays the regulatory data ontology summary.
func RunOntologyCommand() error {
	db, err := storage.ConnectPostgres()
	if err != nil {
		return fmt.Errorf("database connection failed: %w", err)
	}
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			log.Printf("WARNING: failed to close database: %v", closeErr)
		}
	}()

	repo := ontology.NewRepository(db)
	if err := repo.DebugPrintOntologySummary(); err != nil {
		return fmt.Errorf("ontology query failed: %w", err)
	}

	return nil
}

// displayParsedCaseInfo prints a summary of the parsed case from Rust service.
func displayParsedCaseInfo(c *pb.ParsedCase) {
	fmt.Println("✅ Parsed DSL case:", c.Name)
	if c.Nature != "" {
		fmt.Printf("   Nature: %s\n", c.Nature)
	}
	if c.Purpose != "" {
		fmt.Printf("   Purpose: %s\n", c.Purpose)
	}
	if c.ClientBusinessUnit != "" {
		fmt.Printf("   CBU: %s\n", c.ClientBusinessUnit)
	}
	if c.Function != "" {
		fmt.Printf("   Function: %s\n", c.Function)
	}
	if c.Policy != "" {
		fmt.Printf("   Policy: %s\n", c.Policy)
	}
}

// ShowUsage displays usage information.
func ShowUsage() {
	fmt.Println("KYC-DSL Command Line Tool (Rust-powered)")
	fmt.Println("========================================")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  kycctl grammar                          - Store grammar definition in database")
	fmt.Println("  kycctl ontology                         - Display regulatory data ontology")
	fmt.Println("  kycctl validate <case>                  - Validate case and record audit trail")
	fmt.Println("  kycctl <dsl-file>                       - Parse and process a DSL file")
	fmt.Println("  kycctl amend <case> --step=<phase>      - Apply incremental amendment to case")
	fmt.Println()
	fmt.Println("RAG & Vector Search Commands:")
	fmt.Println("  kycctl seed-metadata                    - Seed attribute metadata with embeddings")
	fmt.Println("  kycctl search-metadata <query>          - Semantic search for attributes")
	fmt.Println("  kycctl similar-attributes <code>        - Find similar attributes")
	fmt.Println("  kycctl text-search <term>               - Text-based attribute search")
	fmt.Println("  kycctl metadata-stats                   - Display metadata statistics")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  kycctl grammar")
	fmt.Println("  kycctl ontology")
	fmt.Println("  kycctl validate BLACKROCK-GLOBAL-EQUITY-FUND")
	fmt.Println("  kycctl sample_case.dsl")
	fmt.Println("  kycctl amend AVIVA-EU-EQUITY-FUND --step=policy-discovery")
	fmt.Println("  kycctl seed-metadata")
	fmt.Println("  kycctl search-metadata \"tax residency\"")
	fmt.Println("  kycctl similar-attributes UBO_NAME")
	fmt.Println()
	fmt.Println("Amendment steps:")
	fmt.Println("  policy-discovery        - Add policy discovery function and policies")
	fmt.Println("  document-solicitation   - Add document solicitation and obligations")
	fmt.Println("  document-discovery      - Auto-populate documents from regulatory ontology")
	fmt.Println("  ownership-discovery     - Add ownership structure and control hierarchy")
	fmt.Println("  risk-assessment         - Add risk assessment function")
	fmt.Println("  regulator-notify        - Add regulator notification")
	fmt.Println("  approve                 - Finalize case as approved")
	fmt.Println("  decline                 - Finalize case as declined")
	fmt.Println("  review                  - Set case to review status")
	fmt.Println()
	fmt.Println("Environment Variables:")
	fmt.Println("  RUST_DSL_SERVICE_ADDR   - Rust DSL service address (default: localhost:50060)")
	fmt.Println("  PGHOST                  - PostgreSQL host (default: localhost)")
	fmt.Println("  PGPORT                  - PostgreSQL port (default: 5432)")
	fmt.Println("  PGUSER                  - PostgreSQL user")
	fmt.Println("  PGDATABASE              - PostgreSQL database (default: kyc_dsl)")
	fmt.Println("  OPENAI_API_KEY          - OpenAI API key (required for RAG features)")
}

// Run is the main CLI entry point that routes commands.
func Run(args []string) {
	if len(args) < 1 {
		ShowUsage()
		log.Fatal("Error: no command or file specified")
	}

	command := args[0]

	switch command {
	case "grammar":
		if err := RunGrammarCommand(); err != nil {
			log.Fatal(err)
		}

	case "ontology":
		if err := RunOntologyCommand(); err != nil {
			log.Fatal(err)
		}

	case "validate":
		if len(args) < 2 {
			fmt.Println("Error: validate command requires case name")
			ShowUsage()
			log.Fatal("missing case name")
		}
		caseName := args[1]
		actor := "System" // Default actor
		if len(args) >= 3 && strings.HasPrefix(args[2], "--actor=") {
			actor = strings.TrimPrefix(args[2], "--actor=")
		}
		if err := RunValidateCommand(caseName, actor); err != nil {
			log.Fatal(err)
		}

	case "amend":
		if len(args) < 2 {
			fmt.Println("Error: amend command requires case name and --step flag")
			ShowUsage()
			log.Fatal("missing arguments")
		}
		caseName := args[1]
		if len(args) < 3 || !strings.HasPrefix(args[2], "--step=") {
			fmt.Println("Error: --step flag required")
			ShowUsage()
			log.Fatal("missing --step flag")
		}
		step := strings.TrimPrefix(args[2], "--step=")
		if err := RunAmendCommand(caseName, step); err != nil {
			log.Fatal(err)
		}

	case "seed-metadata":
		if err := RunSeedMetadataCommand(); err != nil {
			log.Fatal(err)
		}

	case "search-metadata":
		if len(args) < 2 {
			fmt.Println("Error: search-metadata command requires a query")
			ShowUsage()
			log.Fatal("missing search query")
		}
		query := args[1]
		limit := 10
		if len(args) >= 3 && strings.HasPrefix(args[2], "--limit=") {
			fmt.Sscanf(strings.TrimPrefix(args[2], "--limit="), "%d", &limit)
		}
		if err := RunSearchMetadataCommand(query, limit); err != nil {
			log.Fatal(err)
		}

	case "similar-attributes":
		if len(args) < 2 {
			fmt.Println("Error: similar-attributes command requires an attribute code")
			ShowUsage()
			log.Fatal("missing attribute code")
		}
		attributeCode := args[1]
		limit := 10
		if len(args) >= 3 && strings.HasPrefix(args[2], "--limit=") {
			fmt.Sscanf(strings.TrimPrefix(args[2], "--limit="), "%d", &limit)
		}
		if err := RunSimilarAttributesCommand(attributeCode, limit); err != nil {
			log.Fatal(err)
		}

	case "text-search":
		if len(args) < 2 {
			fmt.Println("Error: text-search command requires a search term")
			ShowUsage()
			log.Fatal("missing search term")
		}
		searchTerm := args[1]
		if err := RunTextSearchCommand(searchTerm); err != nil {
			log.Fatal(err)
		}

	case "metadata-stats":
		if err := RunMetadataStatsCommand(); err != nil {
			log.Fatal(err)
		}

	case "help", "-h", "--help":
		ShowUsage()

	default:
		// Treat as DSL file path
		if err := RunProcessCommand(command); err != nil {
			log.Fatal(err)
		}
	}
}
