package main

import (
	"fmt"
	"log"
	"os"

	"github.com/adamtc007/KYC-DSL/internal/engine"
	"github.com/adamtc007/KYC-DSL/internal/parser"
	"github.com/adamtc007/KYC-DSL/internal/storage"
)

const DEBUG = true

func debugLog(format string, args ...interface{}) {
	if DEBUG {
		log.Printf("[DEBUG] "+format, args...)
	}
}

func main() {
	debugLog("=== BREAKPOINT 1: Starting main() ===")
	debugLog("Command line args: %v", os.Args)

	if len(os.Args) < 2 {
		fmt.Println("Usage: kycctl <dsl-file>")
		os.Exit(1)
	}

	file := os.Args[1]
	debugLog("=== BREAKPOINT 2: About to parse file: %s ===", file)

	ast, err := parser.ParseFile(file)
	if err != nil {
		log.Fatalf("parse error: %v", err)
	}

	debugLog("=== BREAKPOINT 3: Parse complete ===")
	debugLog("Number of cases parsed: %d", len(ast.Cases))
	debugLog("First case name: %s", ast.Cases[0].Name)
	debugLog("First case body: %+v", ast.Cases[0].Body)
	fmt.Println("✅ Parsed DSL case:", ast.Cases[0].Name)

	debugLog("=== BREAKPOINT 4: Connecting to database ===")
	db, err := storage.ConnectPostgres()
	if err != nil {
		log.Fatalf("database error: %v", err)
	}
	defer func() {
		debugLog("=== BREAKPOINT 7: Closing database connection ===")
		if err := db.Close(); err != nil {
			log.Printf("failed to close database: %v", err)
		}
	}()
	debugLog("Database connected successfully")

	debugLog("=== BREAKPOINT 5: Creating executor ===")
	exec := engine.NewExecutor(db)
	debugLog("Executor created: %+v", exec)

	debugLog("=== BREAKPOINT 6: Running case: %s ===", ast.Cases[0].Name)
	if err := exec.RunCase(ast.Cases[0].Name); err != nil {
		log.Fatalf("execution failed: %v", err)
	}

	fmt.Println("💾 Case successfully persisted to Postgres.")
	debugLog("=== EXECUTION COMPLETE ===")
}
