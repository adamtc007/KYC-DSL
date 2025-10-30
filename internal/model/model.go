package model

import "time"

type CaseStatus string

const (
	Pending  CaseStatus = "pending"
	Complete CaseStatus = "complete"
	Failed   CaseStatus = "failed"
)

type KycCase struct {
	ID          int        `db:"id"`
	Name        string     `db:"name"`
	Version     int        `db:"version"`
	Status      CaseStatus `db:"status"`
	LastUpdated time.Time  `db:"last_updated"`

	// DSL-derived fields
	Nature      string
	Purpose     string
	CBU         ClientBusinessUnit
	Policies    []KycPolicy
	Obligations []KycObligation
	Functions   []Function
	Token       *KycToken
	Ownership   []OwnershipNode
}

type ClientBusinessUnit struct {
	Name string
}

type KycPolicy struct {
	Code string
}

type KycObligation struct {
	PolicyCode string
}

type Function struct {
	Action string
	Status CaseStatus
}

type KycToken struct {
	Status string
}

// OwnershipNode represents a single ownership or control relationship
type OwnershipNode struct {
	Entity           string  `db:"entity"`
	Owner            string  `db:"owner"`
	BeneficialOwner  string  `db:"beneficial_owner"`
	Controller       string  `db:"controller"`
	Role             string  `db:"role"`
	OwnershipPercent float64 `db:"ownership_percent"`
}
