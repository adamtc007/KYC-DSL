package cli

import (
	"fmt"
	"log"
	"strings"

	"github.com/adamtc007/KYC-DSL/internal/amend"
	"github.com/adamtc007/KYC-DSL/internal/engine"
	"github.com/adamtc007/KYC-DSL/internal/model"
	"github.com/adamtc007/KYC-DSL/internal/ontology"
	"github.com/adamtc007/KYC-DSL/internal/parser"
	"github.com/adamtc007/KYC-DSL/internal/storage"
)

// RunGrammarCommand stores the current grammar definition in the database.
func RunGrammarCommand() error {
	db, err := storage.ConnectPostgres()
	if err != nil {
		return fmt.Errorf("database connection failed: %w", err)
	}
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			log.Printf("failed to close database: %v", closeErr)
		}
	}()

	ebnf := parser.CurrentGrammarEBNF()
	err = storage.InsertGrammar(db, "KYC-DSL", "1.1", ebnf)
	if err != nil {
		return fmt.Errorf("insert grammar failed: %w", err)
	}

	fmt.Println("✅ Grammar inserted into Postgres.")
	return nil
}

// RunProcessCommand parses, validates, and persists a DSL file.
func RunProcessCommand(filePath string) error {
	// Parse DSL file
	dsl, err := parser.ParseFile(filePath)
	if err != nil {
		return fmt.Errorf("parse error: %w", err)
	}

	// Bind to typed models
	cases, err := parser.Bind(dsl)
	if err != nil {
		return fmt.Errorf("bind error: %w", err)
	}

	// Connect to database
	db, err := storage.ConnectPostgres()
	if err != nil {
		return fmt.Errorf("database error: %w", err)
	}
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			log.Printf("failed to close database: %v", closeErr)
		}
	}()

	// Validate against grammar and semantics
	ebnf, _ := storage.GetGrammar(db, "KYC-DSL")
	if err := parser.ValidateDSL(db, cases, ebnf); err != nil {
		return fmt.Errorf("❌ DSL validation failed: %w", err)
	}
	fmt.Println("✅ DSL validated successfully (grammar + semantics).")

	// Serialize the typed model back into DSL text
	serialized := parser.SerializeCases(cases)

	// Display parsed case information
	displayCaseInfo(cases[0])

	// Execute and persist
	exec := engine.NewExecutor(db)
	if err := exec.RunCase(cases[0].Name, serialized); err != nil {
		return fmt.Errorf("execution failed: %w", err)
	}

	fmt.Println("\n🧾 DSL snapshot stored and versioned successfully.")
	return nil
}

// RunAmendCommand applies an incremental amendment to an existing case.
func RunAmendCommand(caseName, step string) error {
	db, err := storage.ConnectPostgres()
	if err != nil {
		return fmt.Errorf("database connection failed: %w", err)
	}
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			log.Printf("failed to close database: %v", closeErr)
		}
	}()

	// Map step names to mutation functions
	var mutation func(*model.KycCase)

	// Special handling for ontology-aware amendments
	if step == "document-discovery" {
		repo := ontology.NewRepository(db)
		mutation = func(c *model.KycCase) {
			if err := amend.AddDocumentDiscovery(c, repo); err != nil {
				log.Printf("Error in document discovery: %v", err)
			}
		}
	} else {
		switch step {
		case "policy-discovery":
			mutation = amend.AddPolicyDiscovery
		case "document-solicitation":
			mutation = amend.AddDocumentSolicitation
		case "ownership-discovery":
			mutation = amend.AddOwnershipStructure
		case "risk-assessment":
			mutation = amend.AddRiskAssessment
		case "regulator-notify":
			mutation = amend.AddRegulatorNotification
		case "approve":
			mutation = amend.ApproveCase
		case "decline":
			mutation = amend.DeclineCase
		case "review":
			mutation = amend.RequestReviewCase
		default:
			return fmt.Errorf("unknown amendment step: %s", step)
		}
	}

	// Apply the amendment
	if err := amend.ApplyAmendment(db, caseName, step, mutation); err != nil {
		return fmt.Errorf("amendment failed: %w", err)
	}

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
			log.Printf("failed to close database: %v", closeErr)
		}
	}()

	repo := ontology.NewRepository(db)
	if err := repo.DebugPrintOntologySummary(); err != nil {
		return fmt.Errorf("ontology query failed: %w", err)
	}

	return nil
}

// displayCaseInfo prints a summary of the parsed case.
func displayCaseInfo(c *model.KycCase) {
	fmt.Println("✅ Parsed DSL case:", c.Name)
	fmt.Printf("   Nature: %s\n", c.Nature)
	fmt.Printf("   Purpose: %s\n", c.Purpose)
	fmt.Printf("   CBU: %s\n", c.CBU.Name)
	fmt.Printf("   Functions: %v\n", getFunctionNames(c))
}

// getFunctionNames extracts function action names from a case.
func getFunctionNames(c *model.KycCase) []string {
	names := []string{}
	for _, f := range c.Functions {
		names = append(names, f.Action)
	}
	return names
}

// ShowUsage displays usage information.
func ShowUsage() {
	fmt.Println("Usage:")
	fmt.Println("  kycctl grammar                      - Store grammar definition in database")
	fmt.Println("  kycctl ontology                     - Display regulatory data ontology")
	fmt.Println("  kycctl <dsl-file>                   - Parse and process a DSL file")
	fmt.Println("  kycctl amend <case> --step=<phase>  - Apply incremental amendment to case")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  kycctl grammar")
	fmt.Println("  kycctl ontology")
	fmt.Println("  kycctl sample_case.dsl")
	fmt.Println("  kycctl amend AVIVA-EU-EQUITY-FUND --step=policy-discovery")
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

	case "help", "-h", "--help":
		ShowUsage()

	default:
		// Treat as DSL file path
		if err := RunProcessCommand(command); err != nil {
			log.Fatal(err)
		}
	}
}
