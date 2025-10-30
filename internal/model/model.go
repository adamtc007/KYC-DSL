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
	Ownership   *OwnershipStructure
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

// OwnershipStructure represents the ownership and control hierarchy of an entity
type OwnershipStructure struct {
	Entity           string
	LegalOwners      []Owner
	BeneficialOwners []BeneficialOwner
	Controllers      []Controller
	OperationalRoles []OperationalRole
}

// Owner represents a legal owner with ownership percentage
type Owner struct {
	Name       string
	Percentage float64
}

// BeneficialOwner represents a beneficial owner with economic interest
type BeneficialOwner struct {
	Name       string
	Percentage float64
	Interest   string // e.g., "voting rights", "economic interest"
}

// Controller represents a person with significant control or influence
type Controller struct {
	Name string
	Role string // e.g., "Senior Managing Official", "Director", "Trustee"
}

// OperationalRole represents key operational management personnel
type OperationalRole struct {
	Name     string
	Title    string
	Function string // e.g., "day-to-day management", "compliance officer"
}
