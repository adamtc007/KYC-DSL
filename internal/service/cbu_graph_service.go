package service

import (
	"context"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/adamtc007/KYC-DSL/api/pb"
	"github.com/jmoiron/sqlx"
)

// CbuGraphService implements the CbuGraphService gRPC interface
type CbuGraphService struct {
	pb.UnimplementedCbuGraphServiceServer
	db *sqlx.DB
}

// NewCbuGraphService creates a new CBU graph service
func NewCbuGraphService(db *sqlx.DB) *CbuGraphService {
	return &CbuGraphService{
		db: db,
	}
}

// GetGraph retrieves the complete organizational graph for a CBU
func (s *CbuGraphService) GetGraph(ctx context.Context, req *pb.GetCbuRequest) (*pb.CbuGraph, error) {
	if req.CbuId == "" {
		return nil, status.Error(codes.InvalidArgument, "cbu_id is required")
	}

	// TODO: Replace with database query
	// For now, return example data based on CBU ID
	graph := s.getExampleGraph(req.CbuId)

	// Calculate counts
	graph.EntityCount = int32(len(graph.Entities))
	graph.RelationshipCount = int32(len(graph.Relationships))
	graph.CreatedAt = timestamppb.Now()
	graph.UpdatedAt = timestamppb.Now()

	return graph, nil
}

// GetEntity retrieves a single entity by ID
func (s *CbuGraphService) GetEntity(ctx context.Context, req *pb.GetEntityRequest) (*pb.CbuEntity, error) {
	if req.CbuId == "" {
		return nil, status.Error(codes.InvalidArgument, "cbu_id is required")
	}
	if req.EntityId == "" {
		return nil, status.Error(codes.InvalidArgument, "entity_id is required")
	}

	// Get the full graph
	graph := s.getExampleGraph(req.CbuId)

	// Find the entity
	for _, entity := range graph.Entities {
		if entity.Id == req.EntityId {
			return entity, nil
		}
	}

	return nil, status.Errorf(codes.NotFound, "entity not found: %s", req.EntityId)
}

// ListEntities streams all entities in a CBU
func (s *CbuGraphService) ListEntities(req *pb.GetCbuRequest, stream pb.CbuGraphService_ListEntitiesServer) error {
	if req.CbuId == "" {
		return status.Error(codes.InvalidArgument, "cbu_id is required")
	}

	// Get the full graph
	graph := s.getExampleGraph(req.CbuId)

	// Stream each entity
	for _, entity := range graph.Entities {
		if err := stream.Send(entity); err != nil {
			return status.Errorf(codes.Internal, "failed to send entity: %v", err)
		}
	}

	return nil
}

// GetRelationships retrieves relationships for a specific entity
func (s *CbuGraphService) GetRelationships(ctx context.Context, req *pb.GetEntityRequest) (*pb.RelationshipResponse, error) {
	if req.CbuId == "" {
		return nil, status.Error(codes.InvalidArgument, "cbu_id is required")
	}
	if req.EntityId == "" {
		return nil, status.Error(codes.InvalidArgument, "entity_id is required")
	}

	// Get the full graph
	graph := s.getExampleGraph(req.CbuId)

	// Separate inbound and outbound relationships
	var inbound []*pb.CbuRelationship
	var outbound []*pb.CbuRelationship

	for _, rel := range graph.Relationships {
		if rel.ToId == req.EntityId {
			inbound = append(inbound, rel)
		}
		if rel.FromId == req.EntityId {
			outbound = append(outbound, rel)
		}
	}

	return &pb.RelationshipResponse{
		EntityId: req.EntityId,
		Inbound:  inbound,
		Outbound: outbound,
	}, nil
}

// ValidateGraph validates the graph structure and control percentages
func (s *CbuGraphService) ValidateGraph(ctx context.Context, req *pb.GetCbuRequest) (*pb.ValidationResponse, error) {
	if req.CbuId == "" {
		return nil, status.Error(codes.InvalidArgument, "cbu_id is required")
	}

	// Get the full graph
	graph := s.getExampleGraph(req.CbuId)

	var issues []*pb.ValidationIssue
	valid := true
	totalControl := float32(0)

	// Build entity map for lookups
	entityMap := make(map[string]*pb.CbuEntity)
	for _, entity := range graph.Entities {
		entityMap[entity.Id] = entity
	}

	// Check each relationship
	for _, rel := range graph.Relationships {
		// Validate from entity exists
		if _, ok := entityMap[rel.FromId]; !ok {
			issues = append(issues, &pb.ValidationIssue{
				Severity:       "error",
				Message:        fmt.Sprintf("From entity not found: %s", rel.FromId),
				EntityId:       rel.FromId,
				RelationshipId: rel.Id,
			})
			valid = false
		}

		// Validate to entity exists
		if _, ok := entityMap[rel.ToId]; !ok {
			issues = append(issues, &pb.ValidationIssue{
				Severity:       "error",
				Message:        fmt.Sprintf("To entity not found: %s", rel.ToId),
				EntityId:       rel.ToId,
				RelationshipId: rel.Id,
			})
			valid = false
		}

		// Validate control percentage
		if rel.ControlPct < 0 || rel.ControlPct > 100 {
			issues = append(issues, &pb.ValidationIssue{
				Severity:       "error",
				Message:        fmt.Sprintf("Invalid control percentage: %.2f%%", rel.ControlPct),
				RelationshipId: rel.Id,
			})
			valid = false
		}

		// Accumulate total control for ownership relationships
		if rel.RelationType == "owns" {
			totalControl += rel.ControlPct
		}

		// Check for self-loops
		if rel.FromId == rel.ToId {
			issues = append(issues, &pb.ValidationIssue{
				Severity:       "warning",
				Message:        "Self-referential relationship detected",
				EntityId:       rel.FromId,
				RelationshipId: rel.Id,
			})
		}
	}

	// Check for control percentage sum per entity
	controlSums := make(map[string]float32)
	for _, rel := range graph.Relationships {
		if rel.RelationType == "owns" {
			controlSums[rel.ToId] += rel.ControlPct
		}
	}

	for entityId, sum := range controlSums {
		if sum > 100.01 { // Allow small floating point error
			issues = append(issues, &pb.ValidationIssue{
				Severity: "error",
				Message:  fmt.Sprintf("Total ownership exceeds 100%% (%.2f%%)", sum),
				EntityId: entityId,
			})
			valid = false
		} else if sum < 99.99 && sum > 0 {
			issues = append(issues, &pb.ValidationIssue{
				Severity: "warning",
				Message:  fmt.Sprintf("Ownership sum is %.2f%%, not 100%%", sum),
				EntityId: entityId,
			})
		}
	}

	// Check for cycles (simple DFS-based cycle detection)
	if s.hasCycles(graph) {
		issues = append(issues, &pb.ValidationIssue{
			Severity: "warning",
			Message:  "Circular ownership detected in graph",
		})
	}

	return &pb.ValidationResponse{
		Valid:           valid,
		Issues:          issues,
		TotalControlPct: totalControl,
	}, nil
}

// GetControlChain traces the control chain from root to a specific entity
func (s *CbuGraphService) GetControlChain(ctx context.Context, req *pb.GetEntityRequest) (*pb.ControlChainResponse, error) {
	if req.CbuId == "" {
		return nil, status.Error(codes.InvalidArgument, "cbu_id is required")
	}
	if req.EntityId == "" {
		return nil, status.Error(codes.InvalidArgument, "entity_id is required")
	}

	// Get the full graph
	graph := s.getExampleGraph(req.CbuId)

	// Build entity map
	entityMap := make(map[string]*pb.CbuEntity)
	for _, entity := range graph.Entities {
		entityMap[entity.Id] = entity
	}

	// Find the ultimate parent (entity with no inbound ownership relationships)
	rootEntity := s.findRootEntity(graph)
	if rootEntity == nil {
		return nil, status.Error(codes.Internal, "no root entity found in graph")
	}

	// Build adjacency list
	adjList := make(map[string][]*pb.CbuRelationship)
	for _, rel := range graph.Relationships {
		if rel.RelationType == "owns" || rel.RelationType == "controls" {
			adjList[rel.FromId] = append(adjList[rel.FromId], rel)
		}
	}

	// Perform BFS to find path from root to target
	chain, effectiveControl := s.tracePath(rootEntity.Id, req.EntityId, adjList, entityMap)

	if len(chain) == 0 {
		return nil, status.Errorf(codes.NotFound, "no control chain found from root to entity: %s", req.EntityId)
	}

	return &pb.ControlChainResponse{
		TargetEntityId:      req.EntityId,
		Chain:               chain,
		EffectiveControlPct: effectiveControl,
	}, nil
}

// Helper functions

// getExampleGraph returns example graph data for testing
func (s *CbuGraphService) getExampleGraph(cbuId string) *pb.CbuGraph {
	// Example entities
	entities := []*pb.CbuEntity{
		{
			Id:           "E1",
			Name:         "BlackRock Inc.",
			EntityType:   "Parent",
			Jurisdiction: "US",
			LeiCode:      "549300HSHQKAH636QV49",
			CreatedAt:    timestamppb.Now(),
		},
		{
			Id:           "E2",
			Name:         "BlackRock Luxembourg SICAV",
			EntityType:   "Fund",
			Jurisdiction: "LU",
			LeiCode:      "549300ABCD1234567890",
			CreatedAt:    timestamppb.Now(),
		},
		{
			Id:           "E3",
			Name:         "BlackRock Investment Management UK",
			EntityType:   "Manager",
			Jurisdiction: "UK",
			LeiCode:      "5493001234567890ABCD",
			CreatedAt:    timestamppb.Now(),
		},
		{
			Id:           "E4",
			Name:         "State Street Custodian Services",
			EntityType:   "Custodian",
			Jurisdiction: "IE",
			LeiCode:      "549300EFGH1234567890",
			CreatedAt:    timestamppb.Now(),
		},
		{
			Id:           "E5",
			Name:         "BlackRock Sub-Fund A",
			EntityType:   "SubFund",
			Jurisdiction: "LU",
			CreatedAt:    timestamppb.Now(),
		},
	}

	// Example roles
	roles := []*pb.CbuRole{
		{
			Id:                       "R1",
			Name:                     "AssetOwner",
			Description:              "Ultimate controlling parent entity",
			RegulatoryClassification: "Parent Company",
		},
		{
			Id:                       "R2",
			Name:                     "ManagementCompany",
			Description:              "UCITS ManCo overseeing fund operations",
			RegulatoryClassification: "UCITS ManCo",
		},
		{
			Id:                       "R3",
			Name:                     "Custodian",
			Description:              "Holds fund assets in safekeeping",
			RegulatoryClassification: "Depositary",
		},
		{
			Id:                       "R4",
			Name:                     "SubFundManager",
			Description:              "Manages specific sub-fund assets",
			RegulatoryClassification: "Delegated Portfolio Manager",
		},
	}

	// Example relationships
	relationships := []*pb.CbuRelationship{
		{
			Id:            "REL1",
			FromId:        "E1",
			ToId:          "E2",
			RelationType:  "owns",
			ControlPct:    100.0,
			RoleId:        "R1",
			IsBeneficial:  true,
			EffectiveDate: timestamppb.Now(),
		},
		{
			Id:            "REL2",
			FromId:        "E3",
			ToId:          "E2",
			RelationType:  "controls",
			ControlPct:    100.0,
			RoleId:        "R2",
			IsBeneficial:  false,
			EffectiveDate: timestamppb.Now(),
		},
		{
			Id:            "REL3",
			FromId:        "E4",
			ToId:          "E2",
			RelationType:  "custodies",
			ControlPct:    0.0,
			RoleId:        "R3",
			IsBeneficial:  false,
			EffectiveDate: timestamppb.Now(),
		},
		{
			Id:            "REL4",
			FromId:        "E2",
			ToId:          "E5",
			RelationType:  "owns",
			ControlPct:    100.0,
			RoleId:        "R1",
			IsBeneficial:  true,
			EffectiveDate: timestamppb.Now(),
		},
		{
			Id:            "REL5",
			FromId:        "E3",
			ToId:          "E5",
			RelationType:  "controls",
			ControlPct:    100.0,
			RoleId:        "R4",
			IsBeneficial:  false,
			EffectiveDate: timestamppb.Now(),
		},
	}

	return &pb.CbuGraph{
		CbuId:         cbuId,
		Name:          "BlackRock Global Equity CBU Network",
		Description:   "Complete organizational structure for BlackRock Global Equity fund family",
		Entities:      entities,
		Roles:         roles,
		Relationships: relationships,
	}
}

// findRootEntity finds the ultimate parent entity (no inbound ownership)
func (s *CbuGraphService) findRootEntity(graph *pb.CbuGraph) *pb.CbuEntity {
	// Build set of entities that have inbound ownership
	hasInbound := make(map[string]bool)
	for _, rel := range graph.Relationships {
		if rel.RelationType == "owns" {
			hasInbound[rel.ToId] = true
		}
	}

	// Find entity with no inbound ownership
	for _, entity := range graph.Entities {
		if !hasInbound[entity.Id] {
			return entity
		}
	}

	// If all have inbound (circular), return first entity
	if len(graph.Entities) > 0 {
		return graph.Entities[0]
	}

	return nil
}

// hasCycles detects cycles in the ownership graph using DFS
func (s *CbuGraphService) hasCycles(graph *pb.CbuGraph) bool {
	// Build adjacency list
	adjList := make(map[string][]string)
	for _, rel := range graph.Relationships {
		if rel.RelationType == "owns" || rel.RelationType == "controls" {
			adjList[rel.FromId] = append(adjList[rel.FromId], rel.ToId)
		}
	}

	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	var hasCycle func(string) bool
	hasCycle = func(node string) bool {
		visited[node] = true
		recStack[node] = true

		for _, neighbor := range adjList[node] {
			if !visited[neighbor] {
				if hasCycle(neighbor) {
					return true
				}
			} else if recStack[neighbor] {
				return true
			}
		}

		recStack[node] = false
		return false
	}

	for _, entity := range graph.Entities {
		if !visited[entity.Id] {
			if hasCycle(entity.Id) {
				return true
			}
		}
	}

	return false
}

// tracePath finds the control chain from source to target using BFS
func (s *CbuGraphService) tracePath(
	source, target string,
	adjList map[string][]*pb.CbuRelationship,
	entityMap map[string]*pb.CbuEntity,
) ([]*pb.ControlLink, float32) {
	if source == target {
		return []*pb.ControlLink{}, 100.0
	}

	// BFS to find path
	type queueItem struct {
		node    string
		path    []*pb.ControlLink
		control float32
	}

	queue := []queueItem{{node: source, path: []*pb.ControlLink{}, control: 100.0}}
	visited := make(map[string]bool)
	visited[source] = true

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		for _, rel := range adjList[current.node] {
			if visited[rel.ToId] {
				continue
			}

			newControl := current.control * rel.ControlPct / 100.0
			newPath := make([]*pb.ControlLink, len(current.path))
			copy(newPath, current.path)

			fromEntity := entityMap[rel.FromId]
			toEntity := entityMap[rel.ToId]

			link := &pb.ControlLink{
				FromEntityId:   rel.FromId,
				FromEntityName: fromEntity.Name,
				ToEntityId:     rel.ToId,
				ToEntityName:   toEntity.Name,
				RelationType:   rel.RelationType,
				ControlPct:     rel.ControlPct,
				RoleId:         rel.RoleId,
			}
			newPath = append(newPath, link)

			if rel.ToId == target {
				return newPath, newControl
			}

			queue = append(queue, queueItem{
				node:    rel.ToId,
				path:    newPath,
				control: newControl,
			})
			visited[rel.ToId] = true
		}
	}

	return nil, 0
}
