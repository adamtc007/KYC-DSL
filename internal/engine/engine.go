package engine

import (
	"fmt"

	"github.com/adamtc007/KYC-DSL/internal/storage"
	"github.com/jmoiron/sqlx"
)

type Executor struct {
	DB *sqlx.DB
}

func NewExecutor(db *sqlx.DB) *Executor {
	return &Executor{DB: db}
}

func (e *Executor) RunCase(caseName string) error {
	fmt.Println("🧩 Executing KYC Case:", caseName)
	if err := storage.InsertCase(e.DB, caseName); err != nil {
		return fmt.Errorf("insert failed: %w", err)
	}
	fmt.Println("🔗 Case recorded in database.")
	return nil
}
