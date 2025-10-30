package main

import (
	"fmt"
	"log"
	"os"

	"github.com/adamtc007/KYC-DSL/internal/engine"
	"github.com/adamtc007/KYC-DSL/internal/parser"
	"github.com/adamtc007/KYC-DSL/internal/storage"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: kycctl <dsl-file>")
		os.Exit(1)
	}

	file := os.Args[1]
	ast, err := parser.ParseFile(file)
	if err != nil {
		log.Fatalf("parse error: %v", err)
	}

	fmt.Println("✅ Parsed DSL case:", ast.Cases[0].Name)

	db, err := storage.ConnectPostgres()
	if err != nil {
		log.Fatalf("database error: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("failed to close database: %v", err)
		}
	}()

	exec := engine.NewExecutor(db)
	if err := exec.RunCase(ast.Cases[0].Name); err != nil {
		log.Fatalf("execution failed: %v", err)
	}

	fmt.Println("💾 Case successfully persisted to Postgres.")
}
