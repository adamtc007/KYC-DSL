package service

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/adamtc007/KYC-DSL/api/pb"
	"github.com/adamtc007/KYC-DSL/internal/engine"
	"github.com/adamtc007/KYC-DSL/internal/parser"
	"github.com/adamtc007/KYC-DSL/internal/storage"
	"github.com/jmoiron/sqlx"
)

// KycCaseService implements the gRPC KycCaseService interface
type KycCaseService struct {
	pb.UnimplementedKycCaseServiceServer
	db     *sqlx.DB
	engine *engine.Executor
}

// NewKycCaseService creates a new KYC case service
func NewKycCaseService(db *sqlx.DB) *KycCaseService {
	return &KycCaseService{
		db:     db,
		engine: engine.NewExecutor(db),
	}
}

// GetCase retrieves a single KYC case by ID
func (s *KycCaseService) GetCase(ctx context.Context, req *pb.GetCaseRequest) (*pb.KycCase, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "case ID is required")
	}

	// Get the latest case from storage
	store := storage.NewStorage(s.db)
	caseData, err := store.GetLatestCase(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "case not found: %v", err)
	}

	// Convert to protobuf message
	return &pb.KycCase{
		Id:                 caseData.ID,
		Name:               caseData.Name,
		Jurisdiction:       extractJurisdiction(caseData.DSL),
		Dsl:                caseData.DSL,
		Policy:             extractPolicy(caseData.DSL),
		CreatedAt:          timestamppb.New(caseData.CreatedAt),
		UpdatedAt:          timestamppb.New(caseData.UpdatedAt),
		ClientBusinessUnit: extractCBU(caseData.DSL),
		Function:           extractFunction(caseData.DSL),
		Obligation:         extractObligation(caseData.DSL),
		KycToken:           extractToken(caseData.DSL),
		Version:            int32(caseData.Version),
		Sha256Hash:         caseData.SHA256Hash,
	}, nil
}

// UpdateCase applies updates to an existing case
func (s *KycCaseService) UpdateCase(ctx context.Context, req *pb.UpdateCaseRequest) (*pb.KycCase, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "case ID is required")
	}

	// Get current case
	store := storage.NewStorage(s.db)
	caseData, err := store.GetLatestCase(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "case not found: %v", err)
	}

	// Parse current DSL
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

	// Apply updates
	for key, value := range req.Updates {
		switch key {
		case "jurisdiction":
			// Update jurisdiction in case structure
			if kycCase.DocumentRequirements != nil {
				kycCase.DocumentRequirements.Jurisdiction = value
			}
		case "policy":
			kycCase.Policy = value
		case "function":
			kycCase.Function = value
		case "obligation":
			kycCase.Obligation = value
		case "kyc_token":
			kycCase.KYCToken = value
		}
	}

	// Serialize back to DSL
	updatedDSL, err := parser.SerializeCases([]interface{}{kycCase})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to serialize case: %v", err)
	}

	// Save updated case
	if err := store.InsertCase(req.Id, updatedDSL); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to save case: %v", err)
	}

	// Return updated case
	return s.GetCase(ctx, &pb.GetCaseRequest{Id: req.Id})
}

// ListCases streams all available KYC cases
func (s *KycCaseService) ListCases(req *pb.ListCasesRequest, stream pb.KycCaseService_ListCasesServer) error {
	store := storage.NewStorage(s.db)

	// Build query with filters
	query := `
		SELECT DISTINCT ON (name)
			id, name, dsl, version, sha256_hash, created_at, updated_at
		FROM kyc_cases
		WHERE 1=1
	`
	args := []interface{}{}

	if req.Jurisdiction != "" {
		query += " AND dsl ILIKE $" + fmt.Sprintf("%d", len(args)+1)
		args = append(args, "%"+req.Jurisdiction+"%")
	}

	if req.Policy != "" {
		query += " AND dsl ILIKE $" + fmt.Sprintf("%d", len(args)+1)
		args = append(args, "%(policy "+req.Policy+")%")
	}

	query += " ORDER BY name, version DESC"

	if req.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", req.Limit)
	}

	if req.Offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", req.Offset)
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return status.Errorf(codes.Internal, "failed to query cases: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id, name, dsl, sha256Hash string
		var version int
		var createdAt, updatedAt time.Time

		if err := rows.Scan(&id, &name, &dsl, &version, &sha256Hash, &createdAt, &updatedAt); err != nil {
			return status.Errorf(codes.Internal, "failed to scan row: %v", err)
		}

		kycCase := &pb.KycCase{
			Id:                 id,
			Name:               name,
			Jurisdiction:       extractJurisdiction(dsl),
			Dsl:                dsl,
			Policy:             extractPolicy(dsl),
			CreatedAt:          timestamppb.New(createdAt),
			UpdatedAt:          timestamppb.New(updatedAt),
			ClientBusinessUnit: extractCBU(dsl),
			Function:           extractFunction(dsl),
			Obligation:         extractObligation(dsl),
			KycToken:           extractToken(dsl),
			Version:            int32(version),
			Sha256Hash:         sha256Hash,
		}

		if err := stream.Send(kycCase); err != nil {
			return status.Errorf(codes.Internal, "failed to send case: %v", err)
		}
	}

	return nil
}

// CreateCase creates a new KYC case from DSL
func (s *KycCaseService) CreateCase(ctx context.Context, req *pb.CreateCaseRequest) (*pb.KycCase, error) {
	if req.Dsl == "" {
		return nil, status.Error(codes.InvalidArgument, "DSL is required")
	}

	// Parse DSL
	cases, err := parser.ParseFile([]byte(req.Dsl))
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid DSL: %v", err)
	}

	if len(cases) == 0 {
		return nil, status.Error(codes.InvalidArgument, "no cases found in DSL")
	}

	// Bind and validate
	kycCase, err := parser.Bind(cases[0])
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to bind case: %v", err)
	}

	// Validate
	if err := parser.ValidateDSL(cases[0], s.db); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "validation failed: %v", err)
	}

	// Save to database
	store := storage.NewStorage(s.db)
	if err := store.InsertCase(kycCase.Name, req.Dsl); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to save case: %v", err)
	}

	// Return created case
	return s.GetCase(ctx, &pb.GetCaseRequest{Id: kycCase.Name})
}

// DeleteCase removes a KYC case
func (s *KycCaseService) DeleteCase(ctx context.Context, req *pb.DeleteCaseRequest) (*pb.DeleteCaseResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "case ID is required")
	}

	// Delete all versions of the case
	result, err := s.db.Exec("DELETE FROM kyc_cases WHERE name = $1", req.Id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete case: %v", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "case not found: %s", req.Id)
	}

	return &pb.DeleteCaseResponse{
		Success: true,
		Message: fmt.Sprintf("Deleted %d version(s) of case %s", rowsAffected, req.Id),
	}, nil
}

// GetCaseVersions retrieves all versions of a case
func (s *KycCaseService) GetCaseVersions(req *pb.GetCaseVersionsRequest, stream pb.KycCaseService_GetCaseVersionsServer) error {
	if req.CaseName == "" {
		return status.Error(codes.InvalidArgument, "case name is required")
	}

	query := `
		SELECT version, dsl, sha256_hash, created_at
		FROM kyc_cases
		WHERE name = $1
		ORDER BY version DESC
	`

	rows, err := s.db.Query(query, req.CaseName)
	if err != nil {
		return status.Errorf(codes.Internal, "failed to query versions: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var version int
		var dsl, sha256Hash string
		var createdAt time.Time

		if err := rows.Scan(&version, &dsl, &sha256Hash, &createdAt); err != nil {
			return status.Errorf(codes.Internal, "failed to scan row: %v", err)
		}

		// Try to get amendment info
		var amendmentType string
		s.db.QueryRow(`
			SELECT amendment_type
			FROM case_amendments
			WHERE case_name = $1 AND new_version = $2
		`, req.CaseName, version).Scan(&amendmentType)

		version := &pb.KycCaseVersion{
			CaseName:      req.CaseName,
			Version:       int32(version),
			DslSnapshot:   dsl,
			Sha256Hash:    sha256Hash,
			CreatedAt:     timestamppb.New(createdAt),
			AmendmentType: amendmentType,
		}

		if err := stream.Send(version); err != nil {
			return status.Errorf(codes.Internal, "failed to send version: %v", err)
		}
	}

	return nil
}

// Helper functions to extract fields from DSL
func extractJurisdiction(dsl string) string {
	// Simple extraction - could use parser for more robust solution
	if idx := findPattern(dsl, "(jurisdiction "); idx >= 0 {
		return extractValue(dsl, idx)
	}
	return ""
}

func extractPolicy(dsl string) string {
	if idx := findPattern(dsl, "(policy "); idx >= 0 {
		return extractValue(dsl, idx)
	}
	return ""
}

func extractCBU(dsl string) string {
	if idx := findPattern(dsl, "(client-business-unit "); idx >= 0 {
		return extractValue(dsl, idx)
	}
	return ""
}

func extractFunction(dsl string) string {
	if idx := findPattern(dsl, "(function "); idx >= 0 {
		return extractValue(dsl, idx)
	}
	return ""
}

func extractObligation(dsl string) string {
	if idx := findPattern(dsl, "(obligation "); idx >= 0 {
		return extractValue(dsl, idx)
	}
	return ""
}

func extractToken(dsl string) string {
	if idx := findPattern(dsl, "(kyc-token "); idx >= 0 {
		return extractValue(dsl, idx)
	}
	return ""
}

func findPattern(s, pattern string) int {
	for i := 0; i < len(s)-len(pattern); i++ {
		if s[i:i+len(pattern)] == pattern {
			return i + len(pattern)
		}
	}
	return -1
}

func extractValue(s string, start int) string {
	end := start
	for end < len(s) && s[end] != ')' && s[end] != ' ' && s[end] != '\n' {
		end++
	}
	return s[start:end]
}
