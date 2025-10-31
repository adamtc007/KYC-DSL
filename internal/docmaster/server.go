package docmaster

import (
	"context"
	"log"
	"strings"
	"sync"

	pb "github.com/adamtc007/KYC-DSL/api/pb"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Server implements the DocMasterService gRPC service.
// It uses a simple in-memory map to mock the database.
type Server struct {
	pb.UnimplementedDocMasterServiceServer

	mu        sync.RWMutex
	documents map[string]*pb.Document // Mock DB: map[document_id] -> Document
}

// NewServer creates and initializes the DocMaster server with mock data.
func NewServer() *Server {
	s := &Server{
		documents: make(map[string]*pb.Document),
	}

	// Pre-populate with mock data
	log.Println("📄 Populating mock document master catalogue...")
	mockDocs := []*pb.Document{
		{
			Id:                    "doc-ukpassport",
			Name:                  "UK Passport",
			CountryCode:           "GB",
			Region:                "EU",
			Category:              "passport",
			ContainedAttributeIds: []string{"attr-name", "attr-dob", "attr-nationality", "attr-doc-num", "attr-issue-date", "attr-expiry-date"},
		},
		{
			Id:                    "doc-ukdl",
			Name:                  "UK Driving Licence",
			CountryCode:           "GB",
			Region:                "EU",
			Category:              "drivers_license",
			ContainedAttributeIds: []string{"attr-name", "attr-dob", "attr-doc-num", "attr-issue-date", "attr-expiry-date", "attr-address"},
		},
		{
			Id:                    "doc-uspassport",
			Name:                  "US Passport",
			CountryCode:           "US",
			Region:                "AMERICAS",
			Category:              "passport",
			ContainedAttributeIds: []string{"attr-name", "attr-dob", "attr-nationality", "attr-doc-num", "attr-issue-date", "attr-expiry-date"},
		},
		{
			Id:                    "doc-frid",
			Name:                  "French National ID Card",
			CountryCode:           "FR",
			Region:                "EU",
			Category:              "id_card",
			ContainedAttributeIds: []string{"attr-name", "attr-dob", "attr-nationality", "attr-doc-num", "attr-issue-date", "attr-address"},
		},
		{
			Id:                    "doc-depassport",
			Name:                  "German Passport",
			CountryCode:           "DE",
			Region:                "EU",
			Category:              "passport",
			ContainedAttributeIds: []string{"attr-name", "attr-dob", "attr-nationality", "attr-doc-num", "attr-issue-date", "attr-expiry-date"},
		},
		{
			Id:                    "doc-jppassport",
			Name:                  "Japanese Passport",
			CountryCode:           "JP",
			Region:                "APAC",
			Category:              "passport",
			ContainedAttributeIds: []string{"attr-name", "attr-dob", "attr-nationality", "attr-doc-num", "attr-issue-date", "attr-expiry-date"},
		},
		{
			Id:                    "doc-sgic",
			Name:                  "Singapore National ID Card",
			CountryCode:           "SG",
			Region:                "APAC",
			Category:              "id_card",
			ContainedAttributeIds: []string{"attr-name", "attr-dob", "attr-nationality", "attr-doc-num", "attr-address"},
		},
		{
			Id:                    "doc-capassport",
			Name:                  "Canadian Passport",
			CountryCode:           "CA",
			Region:                "AMERICAS",
			Category:              "passport",
			ContainedAttributeIds: []string{"attr-name", "attr-dob", "attr-nationality", "attr-doc-num", "attr-issue-date", "attr-expiry-date"},
		},
	}

	for _, doc := range mockDocs {
		s.documents[doc.Id] = doc
	}
	log.Printf("✅ Loaded %d mock documents into DocMaster catalogue", len(s.documents))
	return s
}

// AddDocument creates a new document type in the catalog
func (s *Server) AddDocument(ctx context.Context, req *pb.AddDocumentRequest) (*pb.Document, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Validate required fields
	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "document name is required")
	}
	if req.CountryCode == "" {
		return nil, status.Error(codes.InvalidArgument, "country code is required")
	}

	// Generate new ID
	newID := "doc-" + uuid.New().String()[:8]

	doc := &pb.Document{
		Id:                    newID,
		Name:                  req.Name,
		CountryCode:           strings.ToUpper(req.CountryCode),
		Region:                req.Region,
		Category:              req.Category,
		ContainedAttributeIds: req.ContainedAttributeIds,
	}
	s.documents[newID] = doc

	log.Printf("✅ Added new document: %s (ID: %s, Country: %s)", doc.Name, doc.Id, doc.CountryCode)
	return doc, nil
}

// GetDocument retrieves a specific document by ID
func (s *Server) GetDocument(ctx context.Context, req *pb.GetDocumentRequest) (*pb.Document, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "document ID is required")
	}

	doc, ok := s.documents[req.Id]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "document with id '%s' not found", req.Id)
	}

	log.Printf("📄 Retrieved document: %s (ID: %s)", doc.Name, doc.Id)
	return doc, nil
}

// ListDocuments returns documents matching optional filters
func (s *Server) ListDocuments(ctx context.Context, req *pb.ListDocumentsRequest) (*pb.ListDocumentsResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var results []*pb.Document

	for _, doc := range s.documents {
		// Apply filters
		if req.CountryCodeFilter != "" && !strings.EqualFold(doc.CountryCode, req.CountryCodeFilter) {
			continue
		}
		if req.RegionFilter != "" && !strings.EqualFold(doc.Region, req.RegionFilter) {
			continue
		}
		if req.CategoryFilter != "" && !strings.EqualFold(doc.Category, req.CategoryFilter) {
			continue
		}
		results = append(results, doc)
	}

	log.Printf("📋 ListDocuments with filters (country:%s, region:%s, category:%s) returned %d documents",
		req.CountryCodeFilter, req.RegionFilter, req.CategoryFilter, len(results))

	return &pb.ListDocumentsResponse{
		Documents:  results,
		TotalCount: int32(len(results)),
	}, nil
}

// FindDocumentsByAttribute finds all documents that contain a specific attribute
func (s *Server) FindDocumentsByAttribute(ctx context.Context, req *pb.FindDocumentsByAttributeRequest) (*pb.FindDocumentsByAttributeResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if req.AttributeId == "" {
		return nil, status.Error(codes.InvalidArgument, "attribute_id is required")
	}

	var results []*pb.Document

	for _, doc := range s.documents {
		for _, attrID := range doc.ContainedAttributeIds {
			if attrID == req.AttributeId {
				results = append(results, doc)
				break // Move to the next document
			}
		}
	}

	log.Printf("🔍 FindDocumentsByAttribute for ID '%s' found %d documents", req.AttributeId, len(results))

	return &pb.FindDocumentsByAttributeResponse{
		Documents:  results,
		TotalCount: int32(len(results)),
	}, nil
}
