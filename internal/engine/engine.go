package engine

import (
	"fmt"
	"log"

	"github.com/adamtc007/KYC-DSL/internal/storage"
	"github.com/jmoiron/sqlx"
)

const DEBUG = true

func debugLog(format string, args ...interface{}) {
	if DEBUG {
		log.Printf("[ENGINE DEBUG] "+format, args...)
	}
}

type Executor struct {
	DB *sqlx.DB
}

func NewExecutor(db *sqlx.DB) *Executor {
	debugLog("=== ENGINE BREAKPOINT 1: NewExecutor called ===")
	debugLog("Database connection: %+v", db.Stats())
	return &Executor{DB: db}
}

func (e *Executor) RunCase(caseName string, dslText string) error {
	fmt.Println("🧩 Executing KYC Case:", caseName)

	// Insert case metadata if new
	if err := storage.InsertCase(e.DB, caseName); err != nil {
		return fmt.Errorf("insert failed: %w", err)
	}

	// Persist full DSL snapshot as versioned record
	if err := storage.SaveCaseVersion(e.DB, caseName, dslText); err != nil {
		return fmt.Errorf("save version failed: %w", err)
	}

	fmt.Println("💾 Case successfully persisted and versioned.")
	return nil
}
