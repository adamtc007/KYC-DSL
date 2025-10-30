package main

import (
	"fmt"
	"log"
	"os"

	"github.com/adamtc007/KYC-DSL/internal/engine"
	"github.com/adamtc007/KYC-DSL/internal/model"
	"github.com/adamtc007/KYC-DSL/internal/parser"
	"github.com/adamtc007/KYC-DSL/internal/storage"
)

func main() {
	if len(os.Args) >= 2 && os.Args[1] == "grammar" {
		db, err := storage.ConnectPostgres()
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()

		ebnf := parser.CurrentGrammarEBNF()
		err = storage.InsertGrammar(db, "KYC-DSL", "1.0", ebnf)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("✅ Grammar inserted into Postgres.")
		return
	}

	if len(os.Args) < 2 {
		fmt.Println("Usage: kycctl <dsl-file>")
		os.Exit(1)
	}
	file := os.Args[1]

	dsl, err := parser.ParseFile(file)
	if err != nil {
		log.Fatalf("parse error: %v", err)
	}

	cases, err := parser.Bind(dsl)
	if err != nil {
		log.Fatalf("bind error: %v", err)
	}

	db, err := storage.ConnectPostgres()
	if err != nil {
		log.Fatalf("database error: %v", err)
	}
	defer db.Close()

	// Fetch current grammar from DB
	ebnf, _ := storage.GetGrammar(db, "KYC-DSL")

	// Validate syntax and semantics
	if err := parser.ValidateDSL(db, cases, ebnf); err != nil {
		log.Fatalf("❌ DSL validation failed: %v", err)
	}
	fmt.Println("✅ DSL validated successfully (grammar + semantics).")

	// Serialize the typed model back into DSL text
	serialized := parser.SerializeCases(cases)

	fmt.Println("✅ Parsed DSL case:", cases[0].Name)
	fmt.Printf("   Nature: %s\n", cases[0].Nature)
	fmt.Printf("   Purpose: %s\n", cases[0].Purpose)
	fmt.Printf("   CBU: %s\n", cases[0].CBU.Name)
	fmt.Printf("   Functions: %v\n", getFunctionNames(cases[0]))

	exec := engine.NewExecutor(db)
	if err := exec.RunCase(cases[0].Name, serialized); err != nil {
		log.Fatalf("execution failed: %v", err)
	}

	fmt.Println("\n🧾 DSL snapshot stored and versioned successfully.")
}

func getFunctionNames(c *model.KycCase) []string {
	names := []string{}
	for _, f := range c.Functions {
		names = append(names, f.Action)
	}
	return names
}
