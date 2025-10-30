package service

import (
	"context"

	pb "github.com/adamtc007/KYC-DSL/api/pb"
	"github.com/adamtc007/KYC-DSL/internal/model"
	"github.com/adamtc007/KYC-DSL/internal/ontology"
	"github.com/adamtc007/KYC-DSL/internal/rag"
	"github.com/jmoiron/sqlx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type RagService struct {
	pb.UnimplementedRagServiceServer
	db       *sqlx.DB
	embedder *rag.Embedder
	repo     *ontology.MetadataRepository
	feedback *ontology.FeedbackRepo
}

func NewRagService(db *sqlx.DB) *RagService {
	return &RagService{
		db:       db,
		embedder: rag.NewEmbedder(),
		repo:     ontology.NewMetadataRepo(db),
		feedback: ontology.NewFeedbackRepo(db),
	}
}

func (s *RagService) AttributeSearch(ctx context.Context, req *pb.RagSearchRequest) (*pb.RagSearchResponse, error) {
	if req.Query == "" {
		return nil, status.Error(codes.InvalidArgument, "query is required")
	}
	
	limit := int(req.Limit)
	if limit <= 0 {
		limit = 10
	}

	embedding, err := s.embedder.GenerateEmbeddingFromText(req.Query)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate embedding: %v", err)
	}

	results, err := s.repo.SearchByVector(embedding, limit)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to search: %v", err)
	}

	protoResults := make([]*pb.RagResult, len(results))
	for i, r := range results {
		protoResults[i] = &pb.RagResult{
			AttributeCode:       r.Code,
			RiskLevel:           r.RiskLevel,
			DataType:            r.DataType,
			Description:         r.Description,
			Synonyms:            r.Synonyms,
			RegulatoryCitations: r.RegulatoryCitations,
			ExampleValues:       r.ExampleValues,
			Relevance:           float32(r.SimilarityScore),
		}
	}

	return &pb.RagSearchResponse{
		Results: protoResults,
	}, nil
}

func (s *RagService) SubmitFeedback(ctx context.Context, req *pb.RagFeedbackRequest) (*pb.RagFeedbackResponse, error) {
	if req.QueryText == "" {
		return nil, status.Error(codes.InvalidArgument, "query_text is required")
	}

	var attrCode, docCode, regCode *string
	if req.AttributeCode != "" {
		attrCode = &req.AttributeCode
	}
	if req.DocumentCode != "" {
		docCode = &req.DocumentCode
	}
	if req.RegulationCode != "" {
		regCode = &req.RegulationCode
	}

	var agentName *string
	if req.AgentName != "" {
		agentName = &req.AgentName
	}

	feedback := model.Feedback{
		QueryText:      req.QueryText,
		AttributeCode:  attrCode,
		DocumentCode:   docCode,
		RegulationCode: regCode,
		Feedback:       model.FeedbackSentiment(req.Feedback),
		Confidence:     req.Confidence,
		AgentName:      agentName,
		AgentType:      model.AgentType(req.AgentType),
	}

	id, err := s.feedback.InsertFeedback(feedback)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to save feedback: %v", err)
	}

	return &pb.RagFeedbackResponse{
		Status:    "ok",
		Id:        int32(id),
		Feedback:  req.Feedback,
		AgentName: req.AgentName,
		CreatedAt: timestamppb.Now(),
	}, nil
}

func (s *RagService) HealthCheck(ctx context.Context, req *pb.HealthCheckRequest) (*pb.HealthCheckResponse, error) {
	return &pb.HealthCheckResponse{
		Status:         "healthy",
		Model:          s.embedder.GetModel(),
		Dimensions:     int32(s.embedder.GetDimensions()),
		Timestamp:      timestamppb.Now(),
		DatabaseStatus: "connected",
		EmbedderStatus: "ready",
	}, nil
}

// Stubs for other methods
func (s *RagService) SimilarAttributes(ctx context.Context, req *pb.SimilarAttributesRequest) (*pb.RagSearchResponse, error) {
	return &pb.RagSearchResponse{}, nil
}

func (s *RagService) TextSearch(ctx context.Context, req *pb.TextSearchRequest) (*pb.RagSearchResponse, error) {
	return &pb.RagSearchResponse{}, nil
}

func (s *RagService) GetAttribute(ctx context.Context, req *pb.GetAttributeRequest) (*pb.AttributeMetadata, error) {
	return &pb.AttributeMetadata{}, nil
}

func (s *RagService) GetRecentFeedback(req *pb.GetRecentFeedbackRequest, stream pb.RagService_GetRecentFeedbackServer) error {
	return nil
}

func (s *RagService) GetFeedbackAnalytics(ctx context.Context, req *pb.GetFeedbackAnalyticsRequest) (*pb.FeedbackAnalytics, error) {
	return &pb.FeedbackAnalytics{}, nil
}

func (s *RagService) GetMetadataStats(ctx context.Context, req *pb.GetMetadataStatsRequest) (*pb.MetadataStats, error) {
	return &pb.MetadataStats{}, nil
}

func (s *RagService) EnrichedAttributeSearch(ctx context.Context, req *pb.RagSearchRequest) (*pb.EnrichedSearchResponse, error) {
	return &pb.EnrichedSearchResponse{}, nil
}
