package dataservice

import (
	"context"
	"fmt"
	"log"
	"time"

	pb "github.com/adamtc007/KYC-DSL/api/pb/kycdata"
	"github.com/jackc/pgx/v5"
)

// DataService implements both DictionaryService and CaseService gRPC APIs
type DataService struct {
	pb.UnimplementedDictionaryServiceServer
	pb.UnimplementedCaseServiceServer
}

// NewDataService creates a new DataService instance
func NewDataService() *DataService {
	return &DataService{}
}

// ============================================================================
// Dictionary Service Implementation
// ============================================================================

// GetAttribute retrieves a single attribute by ID
func (s *DataService) GetAttribute(ctx context.Context, req *pb.GetAttributeRequest) (*pb.Attribute, error) {
	log.Printf("📖 GetAttribute: id=%s", req.Id)

	query := `
		SELECT
			attribute_code,
			attribute_name,
			description,
			attribute_type,
			jurisdiction,
			regulation_code
		FROM kyc_attributes
		WHERE attribute_code = $1
	`

	row := DB.QueryRow(ctx, query, req.Id)

	var attr pb.Attribute
	err := row.Scan(
		&attr.Id,
		&attr.Name,
		&attr.Description,
		&attr.AttrType,
		&attr.Jurisdiction,
		&attr.Regulation,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("attribute not found: %s", req.Id)
		}
		log.Printf("❌ GetAttribute error: %v", err)
		return nil, fmt.Errorf("database error: %w", err)
	}

	log.Printf("✅ Found attribute: %s", attr.Name)
	return &attr, nil
}

// ListAttributes retrieves a paginated list of attributes
func (s *DataService) ListAttributes(ctx context.Context, req *pb.ListAttributesRequest) (*pb.AttributeList, error) {
	log.Printf("📖 ListAttributes: limit=%d, offset=%d", req.Limit, req.Offset)

	// Default pagination
	limit := req.Limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	offset := req.Offset
	if offset < 0 {
		offset = 0
	}

	// Query for attributes
	query := `
		SELECT
			attribute_code,
			attribute_name,
			description,
			attribute_type,
			jurisdiction,
			regulation_code
		FROM kyc_attributes
		ORDER BY attribute_code
		LIMIT $1 OFFSET $2
	`

	rows, err := DB.Query(ctx, query, limit, offset)
	if err != nil {
		log.Printf("❌ ListAttributes query error: %v", err)
		return nil, fmt.Errorf("database error: %w", err)
	}
	defer rows.Close()

	var attributes []*pb.Attribute
	for rows.Next() {
		var attr pb.Attribute
		err := rows.Scan(
			&attr.Id,
			&attr.Name,
			&attr.Description,
			&attr.AttrType,
			&attr.Jurisdiction,
			&attr.Regulation,
		)
		if err != nil {
			log.Printf("❌ ListAttributes scan error: %v", err)
			return nil, fmt.Errorf("scan error: %w", err)
		}
		attributes = append(attributes, &attr)
	}

	if err := rows.Err(); err != nil {
		log.Printf("❌ ListAttributes rows error: %v", err)
		return nil, fmt.Errorf("rows error: %w", err)
	}

	// Get total count
	var totalCount int32
	countQuery := `SELECT COUNT(*) FROM kyc_attributes`
	err = DB.QueryRow(ctx, countQuery).Scan(&totalCount)
	if err != nil {
		log.Printf("⚠️ ListAttributes count error: %v", err)
		totalCount = int32(len(attributes)) //nolint:gosec
	}

	log.Printf("✅ Listed %d attributes (total: %d)", len(attributes), totalCount)

	return &pb.AttributeList{
		Attributes: attributes,
		TotalCount: totalCount,
	}, nil
}

// GetDocument retrieves a single document by ID
func (s *DataService) GetDocument(ctx context.Context, req *pb.GetDocumentRequest) (*pb.Document, error) {
	log.Printf("📄 GetDocument: id=%s", req.Id)

	query := `
		SELECT
			document_code,
			document_name,
			jurisdiction,
			category,
			description,
			COALESCE(reference_url, '')
		FROM kyc_documents
		WHERE document_code = $1
	`

	row := DB.QueryRow(ctx, query, req.Id)

	var doc pb.Document
	err := row.Scan(
		&doc.Id,
		&doc.Title,
		&doc.Jurisdiction,
		&doc.Category,
		&doc.Description,
		&doc.Url,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("document not found: %s", req.Id)
		}
		log.Printf("❌ GetDocument error: %v", err)
		return nil, fmt.Errorf("database error: %w", err)
	}

	log.Printf("✅ Found document: %s", doc.Title)
	return &doc, nil
}

// ListDocuments retrieves a paginated list of documents with optional jurisdiction filter
func (s *DataService) ListDocuments(ctx context.Context, req *pb.ListDocumentsRequest) (*pb.DocumentList, error) {
	log.Printf("📄 ListDocuments: limit=%d, offset=%d, jurisdiction=%s", req.Limit, req.Offset, req.Jurisdiction)

	// Default pagination
	limit := req.Limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	offset := req.Offset
	if offset < 0 {
		offset = 0
	}

	// Build query with optional jurisdiction filter
	query := `
		SELECT
			document_code,
			document_name,
			jurisdiction,
			category,
			description,
			COALESCE(reference_url, '')
		FROM kyc_documents
	`

	var args []interface{}
	argPosition := 1

	if req.Jurisdiction != "" {
		query += fmt.Sprintf(" WHERE jurisdiction = $%d", argPosition)
		args = append(args, req.Jurisdiction)
		argPosition++
	}

	query += fmt.Sprintf(" ORDER BY document_code LIMIT $%d OFFSET $%d", argPosition, argPosition+1)
	args = append(args, limit, offset)

	rows, err := DB.Query(ctx, query, args...)
	if err != nil {
		log.Printf("❌ ListDocuments query error: %v", err)
		return nil, fmt.Errorf("database error: %w", err)
	}
	defer rows.Close()

	var documents []*pb.Document
	for rows.Next() {
		var doc pb.Document
		err := rows.Scan(
			&doc.Id,
			&doc.Title,
			&doc.Jurisdiction,
			&doc.Category,
			&doc.Description,
			&doc.Url,
		)
		if err != nil {
			log.Printf("❌ ListDocuments scan error: %v", err)
			return nil, fmt.Errorf("scan error: %w", err)
		}
		documents = append(documents, &doc)
	}

	if err := rows.Err(); err != nil {
		log.Printf("❌ ListDocuments rows error: %v", err)
		return nil, fmt.Errorf("rows error: %w", err)
	}

	// Get total count
	var totalCount int32
	countQuery := `SELECT COUNT(*) FROM kyc_documents`
	if req.Jurisdiction != "" {
		countQuery += ` WHERE jurisdiction = $1`
		err = DB.QueryRow(ctx, countQuery, req.Jurisdiction).Scan(&totalCount)
	} else {
		err = DB.QueryRow(ctx, countQuery).Scan(&totalCount)
	}
	if err != nil {
		log.Printf("⚠️ ListDocuments count error: %v", err)
		totalCount = int32(len(documents)) //nolint:gosec
	}

	log.Printf("✅ Listed %d documents (total: %d)", len(documents), totalCount)

	return &pb.DocumentList{
		Documents:  documents,
		TotalCount: totalCount,
	}, nil
}

// ============================================================================
// Case Service Implementation
// ============================================================================

// SaveCaseVersion saves a new case version to the database
func (s *DataService) SaveCaseVersion(ctx context.Context, req *pb.CaseVersionRequest) (*pb.CaseVersionResponse, error) {
	log.Printf("💾 SaveCaseVersion: case_id=%s, status=%s", req.CaseId, req.Status)

	query := `
		INSERT INTO case_versions (case_id, dsl_source, compiled_json, status, created_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`

	var versionID string
	err := DB.QueryRow(
		ctx,
		query,
		req.CaseId,
		req.DslSource,
		req.CompiledJson,
		req.Status,
		time.Now(),
	).Scan(&versionID)

	if err != nil {
		log.Printf("❌ SaveCaseVersion error: %v", err)
		return &pb.CaseVersionResponse{
			Success:   false,
			Error:     err.Error(),
			VersionId: "",
		}, nil
	}

	log.Printf("✅ Saved case version: %s (version_id: %s)", req.CaseId, versionID)

	return &pb.CaseVersionResponse{
		Success:   true,
		Error:     "",
		VersionId: versionID,
	}, nil
}

// GetCaseVersion retrieves the latest version of a case
func (s *DataService) GetCaseVersion(ctx context.Context, req *pb.GetCaseRequest) (*pb.CaseVersion, error) {
	log.Printf("📦 GetCaseVersion: case_id=%s", req.CaseId)

	query := `
		SELECT
			id,
			case_id,
			dsl_source,
			compiled_json,
			status,
			created_at
		FROM case_versions
		WHERE case_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`

	row := DB.QueryRow(ctx, query, req.CaseId)

	var cv pb.CaseVersion
	var createdAt time.Time
	err := row.Scan(
		&cv.Id,
		&cv.CaseId,
		&cv.DslSource,
		&cv.CompiledJson,
		&cv.Status,
		&createdAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("case version not found: %s", req.CaseId)
		}
		log.Printf("❌ GetCaseVersion error: %v", err)
		return nil, fmt.Errorf("database error: %w", err)
	}

	cv.CreatedAt = createdAt.Format(time.RFC3339)

	log.Printf("✅ Found case version: %s (version_id: %s)", cv.CaseId, cv.Id)
	return &cv, nil
}

// ListCaseVersions retrieves all versions of a case with pagination
func (s *DataService) ListCaseVersions(ctx context.Context, req *pb.ListCaseVersionsRequest) (*pb.CaseVersionList, error) {
	log.Printf("📦 ListCaseVersions: case_id=%s, limit=%d, offset=%d", req.CaseId, req.Limit, req.Offset)

	// Default pagination
	limit := req.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	offset := req.Offset
	if offset < 0 {
		offset = 0
	}

	query := `
		SELECT
			id,
			case_id,
			dsl_source,
			compiled_json,
			status,
			created_at
		FROM case_versions
		WHERE case_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := DB.Query(ctx, query, req.CaseId, limit, offset)
	if err != nil {
		log.Printf("❌ ListCaseVersions query error: %v", err)
		return nil, fmt.Errorf("database error: %w", err)
	}
	defer rows.Close()

	var versions []*pb.CaseVersion
	for rows.Next() {
		var cv pb.CaseVersion
		var createdAt time.Time
		err := rows.Scan(
			&cv.Id,
			&cv.CaseId,
			&cv.DslSource,
			&cv.CompiledJson,
			&cv.Status,
			&createdAt,
		)
		if err != nil {
			log.Printf("❌ ListCaseVersions scan error: %v", err)
			return nil, fmt.Errorf("scan error: %w", err)
		}
		cv.CreatedAt = createdAt.Format(time.RFC3339)
		versions = append(versions, &cv)
	}

	if err := rows.Err(); err != nil {
		log.Printf("❌ ListCaseVersions rows error: %v", err)
		return nil, fmt.Errorf("rows error: %w", err)
	}

	// Get total count for this case
	var totalCount int32
	countQuery := `SELECT COUNT(*) FROM case_versions WHERE case_id = $1`
	err = DB.QueryRow(ctx, countQuery, req.CaseId).Scan(&totalCount)
	if err != nil {
		log.Printf("⚠️ ListCaseVersions count error: %v", err)
		totalCount = int32(len(versions)) //nolint:gosec
	}

	log.Printf("✅ Listed %d versions for case %s (total: %d)", len(versions), req.CaseId, totalCount)

	return &pb.CaseVersionList{
		Versions:   versions,
		TotalCount: totalCount,
	}, nil
}

// ListAllCases retrieves all cases with summary information
func (s *DataService) ListAllCases(ctx context.Context, req *pb.ListAllCasesRequest) (*pb.CaseList, error) {
	log.Printf("📦 ListAllCases: limit=%d, offset=%d, status_filter=%s", req.Limit, req.Offset, req.StatusFilter)

	// Default pagination
	limit := req.Limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	offset := req.Offset
	if offset < 0 {
		offset = 0
	}

	// Build query with optional status filter
	query := `
		SELECT
			case_id,
			COUNT(*) as version_count,
			MAX(status) as status,
			MAX(created_at) as last_updated
		FROM case_versions
	`

	var args []interface{}
	argPosition := 1

	if req.StatusFilter != "" {
		query += fmt.Sprintf(" WHERE status = $%d", argPosition)
		args = append(args, req.StatusFilter)
		argPosition++
	}

	query += fmt.Sprintf(" GROUP BY case_id ORDER BY MAX(created_at) DESC LIMIT $%d OFFSET $%d", argPosition, argPosition+1)
	args = append(args, limit, offset)

	rows, err := DB.Query(ctx, query, args...)
	if err != nil {
		log.Printf("❌ ListAllCases query error: %v", err)
		return nil, fmt.Errorf("database error: %w", err)
	}
	defer rows.Close()

	var cases []*pb.CaseSummary
	for rows.Next() {
		var cs pb.CaseSummary
		var lastUpdated time.Time
		err := rows.Scan(
			&cs.CaseId,
			&cs.VersionCount,
			&cs.Status,
			&lastUpdated,
		)
		if err != nil {
			log.Printf("❌ ListAllCases scan error: %v", err)
			return nil, fmt.Errorf("scan error: %w", err)
		}
		cs.LastUpdated = lastUpdated.Format(time.RFC3339)
		cases = append(cases, &cs)
	}

	if err := rows.Err(); err != nil {
		log.Printf("❌ ListAllCases rows error: %v", err)
		return nil, fmt.Errorf("rows error: %w", err)
	}

	// Get total count
	var totalCount int32
	countQuery := `SELECT COUNT(DISTINCT case_id) FROM case_versions`
	if req.StatusFilter != "" {
		countQuery += ` WHERE status = $1`
		err = DB.QueryRow(ctx, countQuery, req.StatusFilter).Scan(&totalCount)
	} else {
		err = DB.QueryRow(ctx, countQuery).Scan(&totalCount)
	}
	if err != nil {
		log.Printf("⚠️ ListAllCases count error: %v", err)
		totalCount = int32(len(cases))
	}

	log.Printf("✅ Listed %d cases (total: %d)", len(cases), totalCount)

	return &pb.CaseList{
		Cases:      cases,
		TotalCount: totalCount,
	}, nil
}
