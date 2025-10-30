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
