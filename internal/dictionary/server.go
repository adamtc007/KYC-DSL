package dictionary

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

// Server implements the DictionaryService gRPC service.
// It uses a simple in-memory map to mock the database.
type Server struct {
	pb.UnimplementedDictionaryServiceServer

	mu         sync.RWMutex
	attributes map[string]*pb.Attribute // Mock DB: map[attribute_id] -> Attribute
}

// NewServer creates and initializes the Dictionary server with mock data.
func NewServer() *Server {
	s := &Server{
		attributes: make(map[string]*pb.Attribute),
	}

	// Pre-populate with mock data
	log.Println("📖 Populating mock attribute dictionary...")
	mockAttrs := []*pb.Attribute{
		{
			Id:          "attr-name",
			Name:        "Full Name",
			Description: "Complete legal name of the individual",
			DataType:    "string",
			Required:    true,
			Category:    "identity",
			ValidationRules: []string{
				"min_length:2",
				"max_length:200",
				"regex:^[A-Za-z\\s'-]+$",
			},
		},
		{
			Id:          "attr-dob",
			Name:        "Date of Birth",
			Description: "Individual's date of birth",
			DataType:    "date",
			Required:    true,
			Category:    "identity",
			ValidationRules: []string{
				"format:YYYY-MM-DD",
				"min_age:18",
			},
		},
		{
			Id:          "attr-nationality",
			Name:        "Nationality",
			Description: "Individual's nationality or citizenship",
			DataType:    "string",
			Required:    true,
			Category:    "identity",
			ValidationRules: []string{
				"country_code",
				"iso_3166_alpha2",
			},
		},
		{
			Id:          "attr-doc-num",
			Name:        "Document Number",
			Description: "Unique identifier on the identity document",
			DataType:    "string",
			Required:    true,
			Category:    "document",
			ValidationRules: []string{
				"min_length:5",
				"max_length:50",
				"alphanumeric",
			},
		},
		{
			Id:          "attr-issue-date",
			Name:        "Issue Date",
			Description: "Date when the document was issued",
			DataType:    "date",
			Required:    true,
			Category:    "document",
			ValidationRules: []string{
				"format:YYYY-MM-DD",
				"not_future",
			},
		},
		{
			Id:          "attr-expiry-date",
			Name:        "Expiry Date",
			Description: "Date when the document expires",
			DataType:    "date",
			Required:    false,
			Category:    "document",
			ValidationRules: []string{
				"format:YYYY-MM-DD",
				"after_issue_date",
			},
		},
		{
			Id:          "attr-address",
			Name:        "Residential Address",
			Description: "Current residential address",
			DataType:    "string",
			Required:    false,
			Category:    "address",
			ValidationRules: []string{
				"min_length:10",
				"max_length:500",
			},
		},
		{
			Id:          "attr-email",
			Name:        "Email Address",
			Description: "Primary email address",
			DataType:    "string",
			Required:    false,
			Category:    "contact",
			ValidationRules: []string{
				"email_format",
				"max_length:255",
			},
		},
		{
			Id:          "attr-phone",
			Name:        "Phone Number",
			Description: "Primary phone number",
			DataType:    "string",
			Required:    false,
			Category:    "contact",
			ValidationRules: []string{
				"phone_format",
				"min_length:8",
				"max_length:20",
			},
		},
		{
			Id:          "attr-tax-id",
			Name:        "Tax Identification Number",
			Description: "National tax identification number",
			DataType:    "string",
			Required:    false,
			Category:    "financial",
			ValidationRules: []string{
				"alphanumeric",
				"country_specific_format",
			},
		},
		{
			Id:          "attr-occupation",
			Name:        "Occupation",
			Description: "Individual's current occupation or profession",
			DataType:    "string",
			Required:    false,
			Category:    "employment",
			ValidationRules: []string{
				"min_length:2",
				"max_length:100",
			},
		},
		{
			Id:          "attr-income",
			Name:        "Annual Income",
			Description: "Annual income in base currency",
			DataType:    "number",
			Required:    false,
			Category:    "financial",
			ValidationRules: []string{
				"positive",
				"decimal_places:2",
			},
		},
	}

	for _, attr := range mockAttrs {
		s.attributes[attr.Id] = attr
	}
	log.Printf("✅ Loaded %d mock attributes into dictionary", len(s.attributes))
	return s
}

// CreateAttribute creates a new attribute in the dictionary
func (s *Server) CreateAttribute(ctx context.Context, req *pb.CreateAttributeRequest) (*pb.Attribute, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Validate required fields
	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "attribute name is required")
	}

	// Check for duplicate names
	for _, attr := range s.attributes {
		if strings.EqualFold(attr.Name, req.Name) {
			return nil, status.Errorf(codes.AlreadyExists, "attribute with name '%s' already exists", req.Name)
		}
	}

	// Generate new ID
	newID := "attr-" + uuid.New().String()[:8]

	attr := &pb.Attribute{
		Id:              newID,
		Name:            req.Name,
		Description:     req.Description,
		DataType:        req.DataType,
		Required:        req.Required,
		Category:        req.Category,
		ValidationRules: req.ValidationRules,
	}
	s.attributes[newID] = attr

	log.Printf("✅ Created new attribute: %s (ID: %s, Type: %s)", attr.Name, attr.Id, attr.DataType)
	return attr, nil
}

// GetAttribute retrieves a specific attribute by ID
func (s *Server) GetAttribute(ctx context.Context, req *pb.DictGetAttributeRequest) (*pb.Attribute, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "attribute ID is required")
	}

	attr, ok := s.attributes[req.Id]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "attribute with id '%s' not found", req.Id)
	}

	log.Printf("📖 Retrieved attribute: %s (ID: %s)", attr.Name, attr.Id)
	return attr, nil
}

// SearchAttributes searches attributes by name or description
func (s *Server) SearchAttributes(ctx context.Context, req *pb.SearchAttributesRequest) (*pb.SearchAttributesResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query := strings.ToLower(req.Query)
	limit := req.Limit
	if limit <= 0 {
		limit = 100 // Default limit
	}

	var results []*pb.Attribute

	// If no query, return all (up to limit)
	if query == "" {
		count := int32(0)
		for _, attr := range s.attributes {
			results = append(results, attr)
			count++
			if count >= limit {
				break
			}
		}
		return &pb.SearchAttributesResponse{
			Attributes: results,
			TotalCount: int32(len(s.attributes)),
		}, nil
	}

	// Search in name and description
	for _, attr := range s.attributes {
		if strings.Contains(strings.ToLower(attr.Name), query) ||
			strings.Contains(strings.ToLower(attr.Description), query) ||
			strings.Contains(strings.ToLower(attr.Category), query) {
			results = append(results, attr)
			if int32(len(results)) >= limit {
				break
			}
		}
	}

	log.Printf("🔍 Search for '%s' found %d results (limit: %d)", req.Query, len(results), limit)
	return &pb.SearchAttributesResponse{
		Attributes: results,
		TotalCount: int32(len(results)),
	}, nil
}

// ListAttributes returns all attributes with optional filtering and pagination
func (s *Server) ListAttributes(ctx context.Context, req *pb.ListAttributesRequest) (*pb.ListAttributesResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var filtered []*pb.Attribute

	// Apply category filter if specified
	for _, attr := range s.attributes {
		if req.CategoryFilter != "" && !strings.EqualFold(attr.Category, req.CategoryFilter) {
			continue
		}
		filtered = append(filtered, attr)
	}

	// Apply pagination
	limit := req.Limit
	if limit <= 0 {
		limit = 50 // Default limit
	}
	offset := req.Offset
	if offset < 0 {
		offset = 0
	}

	var results []*pb.Attribute
	for i, attr := range filtered {
		if int32(i) < offset {
			continue
		}
		if int32(len(results)) >= limit {
			break
		}
		results = append(results, attr)
	}

	log.Printf("📋 ListAttributes (category:%s, limit:%d, offset:%d) returned %d of %d total attributes",
		req.CategoryFilter, limit, offset, len(results), len(filtered))

	return &pb.ListAttributesResponse{
		Attributes: results,
		TotalCount: int32(len(filtered)),
	}, nil
}
