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

func (e *Executor) RunCase(caseName string) error {
	debugLog("=== ENGINE BREAKPOINT 2: RunCase called with caseName='%s' ===", caseName)
	fmt.Println("🧩 Executing KYC Case:", caseName)

	debugLog("=== ENGINE BREAKPOINT 3: About to insert case into database ===")
	if err := storage.InsertCase(e.DB, caseName); err != nil {
		debugLog("Insert failed with error: %v", err)
		return fmt.Errorf("insert failed: %w", err)
	}
	debugLog("=== ENGINE BREAKPOINT 4: Case inserted successfully ===")
	fmt.Println("🔗 Case recorded in database.")
	return nil
}
