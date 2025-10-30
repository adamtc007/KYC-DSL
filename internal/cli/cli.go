package cli

import (
	"fmt"
	"log"

	"github.com/adamtc007/KYC-DSL/internal/engine"
	"github.com/adamtc007/KYC-DSL/internal/model"
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
	err = storage.InsertGrammar(db, "KYC-DSL", "1.0", ebnf)
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
	fmt.Println("  kycctl grammar              - Store grammar definition in database")
	fmt.Println("  kycctl <dsl-file>           - Parse and process a DSL file")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  kycctl grammar")
	fmt.Println("  kycctl sample_case.dsl")
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

	case "help", "-h", "--help":
		ShowUsage()

	default:
		// Treat as DSL file path
		if err := RunProcessCommand(command); err != nil {
			log.Fatal(err)
		}
	}
}
