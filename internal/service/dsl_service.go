package service

import (
	"context"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/adamtc007/KYC-DSL/api/pb"
	"github.com/adamtc007/KYC-DSL/internal/amend"
	"github.com/adamtc007/KYC-DSL/internal/engine"
	"github.com/adamtc007/KYC-DSL/internal/ontology"
	"github.com/adamtc007/KYC-DSL/internal/parser"
	"github.com/adamtc007/KYC-DSL/internal/storage"
	"github.com/jmoiron/sqlx"
)

// DslService implements the gRPC DslService interface
type DslService struct {
	pb.UnimplementedDslServiceServer
	db     *sqlx.DB
	engine *engine.Executor
}

// NewDslService creates a new DSL service
func NewDslService(db *sqlx.DB) *DslService {
	return &DslService{
		db:     db,
		engine: engine.NewExecutor(db),
	}
}

// Execute runs a specific function/amendment on a KYC case
func (s *DslService) Execute(ctx context.Context, req *pb.ExecuteRequest) (*pb.ExecuteResponse, error) {
	if req.CaseId == "" {
		return nil, status.Error(codes.InvalidArgument, "case_id is required")
	}
	if req.FunctionName == "" {
		return nil, status.Error(codes.InvalidArgument, "function_name is required")
	}

	// Get the latest case
	store := storage.NewStorage(s.db)
	caseData, err := store.GetLatestCase(req.CaseId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "case not found: %v", err)
	}

	// Parse the case
	cases, err := parser.ParseFile([]byte(caseData.DSL))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to parse DSL: %v", err)
	}

	if len(cases) == 0 {
		return nil, status.Error(codes.Internal, "no cases found in DSL")
	}

	kycCase, err := parser.Bind(cases[0])
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to bind case: %v", err)
	}

	// Execute the function
	switch req.FunctionName {
	case "policy-discovery":
		kycCase.Function = "policy-discovery"
	case "document-solicitation":
		kycCase.Function = "document-solicitation"
	case "document-discovery":
		kycCase.Function = "document-discovery"
	case "ownership-discovery":
		kycCase.Function = "ownership-discovery"
	case "risk-assessment":
		kycCase.Function = "risk-assessment"
	default:
		return nil, status.Errorf(codes.InvalidArgument, "unknown function: %s", req.FunctionName)
	}

	// Serialize the updated case
	updatedDSL, err := parser.SerializeCases([]interface{}{kycCase})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to serialize case: %v", err)
	}

	// Save the updated case
	if err := store.InsertCase(req.CaseId, updatedDSL); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to save case: %v", err)
	}

	// Get the new version
	updatedCase, _ := store.GetLatestCase(req.CaseId)
	newVersion := int32(1)
	if updatedCase != nil {
		newVersion = int32(updatedCase.Version)
	}

	return &pb.ExecuteResponse{
		UpdatedDsl: updatedDSL,
		Message:    fmt.Sprintf("Executed function '%s' successfully", req.FunctionName),
		Success:    true,
		CaseId:     req.CaseId,
		NewVersion: newVersion,
	}, nil
}

// Validate checks if a DSL case is valid
func (s *DslService) Validate(ctx context.Context, req *pb.ValidateRequest) (*pb.ValidationResult, error) {
	var dsl string
	var err error

	// Use provided DSL or fetch from database
	if req.Dsl != "" {
		dsl = req.Dsl
	} else if req.CaseId != "" {
		store := storage.NewStorage(s.db)
		caseData, err := store.GetLatestCase(req.CaseId)
		if err != nil {
			return nil, status.Errorf(codes.NotFound, "case not found: %v", err)
		}
		dsl = caseData.DSL
	} else {
		return nil, status.Error(codes.InvalidArgument, "either case_id or dsl must be provided")
	}

	// Parse the DSL
	cases, err := parser.ParseFile([]byte(dsl))
	if err != nil {
		return &pb.ValidationResult{
			Valid:  false,
			Errors: []string{fmt.Sprintf("Parse error: %v", err)},
			Issues: []*pb.ValidationIssue{
				{
					Severity: "error",
					Message:  fmt.Sprintf("Failed to parse DSL: %v", err),
					Code:     "PARSE_ERROR",
				},
			},
		}, nil
	}

	if len(cases) == 0 {
		return &pb.ValidationResult{
			Valid:  false,
			Errors: []string{"No cases found in DSL"},
			Issues: []*pb.ValidationIssue{
				{
					Severity: "error",
					Message:  "No cases found in DSL",
					Code:     "NO_CASES",
				},
			},
		}, nil
	}

	// Validate the case
	err = parser.ValidateDSL(cases[0], s.db)
	if err != nil {
		return &pb.ValidationResult{
			Valid:  false,
			Errors: []string{err.Error()},
			Issues: []*pb.ValidationIssue{
				{
					Severity: "error",
					Message:  err.Error(),
					Code:     "VALIDATION_ERROR",
				},
			},
		}, nil
	}

	// Check for warnings (optional completeness checks)
	warnings := []string{}
	kycCase, _ := parser.Bind(cases[0])
	if kycCase != nil {
		if kycCase.OwnershipStructure == nil {
			warnings = append(warnings, "No ownership structure defined")
		}
		if kycCase.DataDictionary == nil || len(kycCase.DataDictionary) == 0 {
			warnings = append(warnings, "No data dictionary defined")
		}
		if kycCase.DocumentRequirements == nil {
			warnings = append(warnings, "No document requirements defined")
		}
	}

	return &pb.ValidationResult{
		Valid:    true,
		Warnings: warnings,
		Issues:   []*pb.ValidationIssue{},
	}, nil
}

// Parse converts DSL text into structured format
func (s *DslService) Parse(ctx context.Context, req *pb.ParseRequest) (*pb.ParseResponse, error) {
	if req.Dsl == "" {
		return nil, status.Error(codes.InvalidArgument, "dsl is required")
	}

	// Parse the DSL
	cases, err := parser.ParseFile([]byte(req.Dsl))
	if err != nil {
		return &pb.ParseResponse{
			Success: false,
			Message: fmt.Sprintf("Parse error: %v", err),
			Errors:  []string{err.Error()},
		}, nil
	}

	if len(cases) == 0 {
		return &pb.ParseResponse{
			Success: false,
			Message: "No cases found in DSL",
			Errors:  []string{"No cases found"},
		}, nil
	}

	// Convert to proto messages
	protoCases := make([]*pb.ParsedCase, 0, len(cases))
	for _, c := range cases {
		kycCase, err := parser.Bind(c)
		if err != nil {
			return &pb.ParseResponse{
				Success: false,
				Message: fmt.Sprintf("Bind error: %v", err),
				Errors:  []string{err.Error()},
			}, nil
		}

		protoCase := convertToProtoParsedCase(kycCase)
		protoCases = append(protoCases, protoCase)
	}

	return &pb.ParseResponse{
		Success: true,
		Message: fmt.Sprintf("Successfully parsed %d case(s)", len(cases)),
		Cases:   protoCases,
	}, nil
}

// Serialize converts structured case back to DSL
func (s *DslService) Serialize(ctx context.Context, req *pb.SerializeRequest) (*pb.SerializeResponse, error) {
	if req.Case == nil {
		return nil, status.Error(codes.InvalidArgument, "case is required")
	}

	// Convert proto case to internal model
	kycCase := convertFromProtoParsedCase(req.Case)

	// Serialize to DSL
	dsl, err := parser.SerializeCases([]interface{}{kycCase})
	if err != nil {
		return &pb.SerializeResponse{
			Success: false,
			Message: fmt.Sprintf("Serialization error: %v", err),
		}, nil
	}

	return &pb.SerializeResponse{
		Success: true,
		Dsl:     dsl,
		Message: "Successfully serialized case to DSL",
	}, nil
}

// Amend applies a predefined amendment to a case
func (s *DslService) Amend(ctx context.Context, req *pb.AmendRequest) (*pb.AmendResponse, error) {
	if req.CaseName == "" {
		return nil, status.Error(codes.InvalidArgument, "case_name is required")
	}
	if req.AmendmentType == "" {
		return nil, status.Error(codes.InvalidArgument, "amendment_type is required")
	}

	// Get the latest case
	store := storage.NewStorage(s.db)
	caseData, err := store.GetLatestCase(req.CaseName)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "case not found: %v", err)
	}

	// Parse the case
	cases, err := parser.ParseFile([]byte(caseData.DSL))
	if err != nil {
		return &pb.AmendResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to parse case: %v", err),
		}, nil
	}

	if len(cases) == 0 {
		return &pb.AmendResponse{
			Success: false,
			Message: "No cases found in DSL",
		}, nil
	}

	kycCase, err := parser.Bind(cases[0])
	if err != nil {
		return &pb.AmendResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to bind case: %v", err),
		}, nil
	}

	// Apply the amendment
	var mutation func(interface{}) error
	repo := ontology.NewRepository(s.db)

	switch req.AmendmentType {
	case "policy-discovery":
		mutation = amend.AddPolicyDiscovery
	case "document-solicitation":
		mutation = amend.AddDocumentSolicitation
	case "document-discovery":
		mutation = amend.AddDocumentDiscovery(repo)
	case "ownership-discovery":
		mutation = amend.AddOwnershipDiscovery
	case "risk-assessment":
		mutation = amend.AddRiskAssessment
	case "regulator-notify":
		mutation = amend.AddRegulatorNotify
	case "approve":
		mutation = amend.ApproveCase
	case "decline":
		mutation = amend.DeclineCase
	case "review":
		mutation = amend.SetCaseToReview
	default:
		return nil, status.Errorf(codes.InvalidArgument, "unknown amendment type: %s", req.AmendmentType)
	}

	// Apply the mutation
	if err := amend.ApplyAmendment(s.db, req.CaseName, mutation); err != nil {
		return &pb.AmendResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to apply amendment: %v", err),
		}, nil
	}

	// Get the updated case
	updatedCase, err := store.GetLatestCase(req.CaseName)
	if err != nil {
		return &pb.AmendResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to retrieve updated case: %v", err),
		}, nil
	}

	return &pb.AmendResponse{
		Success:    true,
		Message:    fmt.Sprintf("Successfully applied amendment '%s'", req.AmendmentType),
		UpdatedDsl: updatedCase.DSL,
		NewVersion: int32(updatedCase.Version),
		Sha256Hash: updatedCase.SHA256Hash,
	}, nil
}

// ListAmendments returns available amendment types
func (s *DslService) ListAmendments(ctx context.Context, req *pb.ListAmendmentsRequest) (*pb.ListAmendmentsResponse, error) {
	amendments := []*pb.AmendmentType{
		{
			Name:        "policy-discovery",
			Description: "Add policy discovery function and policies",
			Parameters:  []string{},
		},
		{
			Name:        "document-solicitation",
			Description: "Add document solicitation and obligations",
			Parameters:  []string{},
		},
		{
			Name:        "document-discovery",
			Description: "Auto-populate documents from ontology (ontology-aware)",
			Parameters:  []string{},
		},
		{
			Name:        "ownership-discovery",
			Description: "Add ownership structure and control hierarchy",
			Parameters:  []string{},
		},
		{
			Name:        "risk-assessment",
			Description: "Add risk assessment function",
			Parameters:  []string{},
		},
		{
			Name:        "regulator-notify",
			Description: "Add regulator notification",
			Parameters:  []string{},
		},
		{
			Name:        "approve",
			Description: "Finalize case as approved",
			Parameters:  []string{},
		},
		{
			Name:        "decline",
			Description: "Finalize case as declined",
			Parameters:  []string{},
		},
		{
			Name:        "review",
			Description: "Set case to review status",
			Parameters:  []string{},
		},
	}

	return &pb.ListAmendmentsResponse{
		Amendments: amendments,
	}, nil
}

// GetGrammar returns the current DSL grammar definition
func (s *DslService) GetGrammar(ctx context.Context, req *pb.GetGrammarRequest) (*pb.GrammarResponse, error) {
	// Get the latest grammar from database
	var version string
	var createdAt string
	err := s.db.QueryRow("SELECT version, created_at FROM grammar_versions ORDER BY created_at DESC LIMIT 1").
		Scan(&version, &createdAt)

	if err != nil {
		// If no grammar in DB, return the current one
		ebnf := parser.CurrentGrammarEBNF()
		return &pb.GrammarResponse{
			Ebnf:    ebnf,
			Version: "1.2",
		}, nil
	}

	// Get the EBNF content
	var ebnf string
	err = s.db.QueryRow("SELECT ebnf FROM grammar_versions WHERE version = $1", version).Scan(&ebnf)
	if err != nil {
		ebnf = parser.CurrentGrammarEBNF()
	}

	// Parse timestamp
	timestamp := timestamppb.Now()
	// Could parse createdAt string to proper timestamp if needed

	return &pb.GrammarResponse{
		Ebnf:      ebnf,
		Version:   version,
		CreatedAt: timestamp,
	}, nil
}

// Helper functions

func convertToProtoParsedCase(kycCase interface{}) *pb.ParsedCase {
	// Type assertion to access fields
	// This is a simplified version - you may need to adjust based on your actual model
	return &pb.ParsedCase{
		Name: "TODO", // Extract from kycCase
		// Fill in other fields...
	}
}

func convertFromProtoParsedCase(protoCase *pb.ParsedCase) interface{} {
	// Convert proto case back to internal model
	// This is a placeholder - implement based on your actual model
	return struct {
		Name string
	}{
		Name: protoCase.Name,
	}
}
