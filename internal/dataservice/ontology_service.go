package dataservice

import (
	"context"
	"fmt"
	"log"

	pb "github.com/adamtc007/KYC-DSL/api/pb/kycontology"
)

type OntologyService struct {
	pb.UnimplementedOntologyServiceServer
}

func NewOntologyService() *OntologyService {
	return &OntologyService{}
}

// ============================================================================
// Entities
// ============================================================================

func (s *OntologyService) GetEntity(ctx context.Context, req *pb.GetEntityRequest) (*pb.Entity, error) {
	log.Printf("📦 GetEntity: id=%s", req.Id)
	row := DB.QueryRow(ctx, `
	  SELECT id, name, entity_type, COALESCE(legal_form,''), jurisdiction,
	         COALESCE(registration_number,''), COALESCE(lei_code,''), status, COALESCE(description,'')
	    FROM entity WHERE id=$1`, req.Id)

	var e pb.Entity
	if err := row.Scan(&e.Id, &e.Name, &e.EntityType, &e.LegalForm, &e.Jurisdiction,
		&e.RegistrationNumber, &e.LeiCode, &e.Status, &e.Description); err != nil {
		return nil, fmt.Errorf("entity not found: %w", err)
	}
	log.Printf("✅ Found entity: %s", e.Name)
	return &e, nil
}

func (s *OntologyService) ListEntities(ctx context.Context, req *pb.ListEntitiesRequest) (*pb.EntityList, error) {
	log.Printf("📦 ListEntities: limit=%d, offset=%d", req.Limit, req.Offset)

	limit := req.Limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	rows, err := DB.Query(ctx, `
	  SELECT id, name, entity_type, COALESCE(legal_form,''), jurisdiction,
	         COALESCE(registration_number,''), COALESCE(lei_code,''), status, COALESCE(description,'')
	    FROM entity ORDER BY name LIMIT $1 OFFSET $2`, limit, req.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := &pb.EntityList{}
	for rows.Next() {
		var e pb.Entity
		if err := rows.Scan(&e.Id, &e.Name, &e.EntityType, &e.LegalForm, &e.Jurisdiction,
			&e.RegistrationNumber, &e.LeiCode, &e.Status, &e.Description); err != nil {
			return nil, err
		}
		list.Entities = append(list.Entities, &e)
	}

	var total int32
	if err := DB.QueryRow(ctx, "SELECT COUNT(*) FROM entity").Scan(&total); err != nil {
		total = int32(len(list.Entities)) //nolint:gosec
	}
	list.TotalCount = total

	log.Printf("✅ Listed %d entities (total: %d)", len(list.Entities), total)
	return list, nil
}

func (s *OntologyService) SearchEntities(ctx context.Context, req *pb.SearchRequest) (*pb.EntityList, error) {
	log.Printf("🔍 SearchEntities: query=%s", req.Query)

	limit := req.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	rows, err := DB.Query(ctx, `
	  SELECT id, name, entity_type, COALESCE(legal_form,''), jurisdiction,
	         COALESCE(registration_number,''), COALESCE(lei_code,''), status, COALESCE(description,'')
	    FROM entity 
	   WHERE name ILIKE '%' || $1 || '%' 
	      OR lei_code ILIKE '%' || $1 || '%'
	      OR description ILIKE '%' || $1 || '%'
	   ORDER BY name LIMIT $2`, req.Query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := &pb.EntityList{}
	for rows.Next() {
		var e pb.Entity
		if err := rows.Scan(&e.Id, &e.Name, &e.EntityType, &e.LegalForm, &e.Jurisdiction,
			&e.RegistrationNumber, &e.LeiCode, &e.Status, &e.Description); err != nil {
			continue
		}
		list.Entities = append(list.Entities, &e)
	}
	list.TotalCount = int32(len(list.Entities)) //nolint:gosec

	log.Printf("✅ Found %d entities matching '%s'", len(list.Entities), req.Query)
	return list, nil
}

// ============================================================================
// CBU and Roles
// ============================================================================

func (s *OntologyService) GetCbu(ctx context.Context, req *pb.GetCbuRequest) (*pb.Cbu, error) {
	log.Printf("🏢 GetCbu: id=%s", req.Id)

	row := DB.QueryRow(ctx, `
	  SELECT id, name, COALESCE(sponsor_entity_id::text,''), COALESCE(domicile,''), COALESCE(description,'')
	    FROM cbu WHERE id=$1`, req.Id)

	var c pb.Cbu
	if err := row.Scan(&c.Id, &c.Name, &c.SponsorEntityId, &c.Domicile, &c.Description); err != nil {
		return nil, fmt.Errorf("cbu not found: %w", err)
	}

	log.Printf("✅ Found CBU: %s", c.Name)
	return &c, nil
}

func (s *OntologyService) GetCbuRoles(ctx context.Context, req *pb.GetCbuRolesRequest) (*pb.CbuRoleList, error) {
	log.Printf("👥 GetCbuRoles: cbu_id=%s", req.CbuId)

	rows, err := DB.Query(ctx, `
	  SELECT cr.id, cr.cbu_id, cr.entity_id,
	         rt.id, rt.code, rt.name, COALESCE(rt.description,''), COALESCE(rt.category,''),
	         COALESCE(cr.jurisdiction,'')
	    FROM cbu_role cr
	    JOIN role_type rt ON cr.role_type_id = rt.id
	   WHERE cr.cbu_id=$1`, req.CbuId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := &pb.CbuRoleList{}
	for rows.Next() {
		var role pb.CbuRole
		role.RoleType = &pb.RoleType{}
		if err := rows.Scan(
			&role.Id, &role.CbuId, &role.EntityId,
			&role.RoleType.Id, &role.RoleType.Code, &role.RoleType.Name,
			&role.RoleType.Description, &role.RoleType.Category,
			&role.Jurisdiction); err != nil {
			continue
		}
		out.Roles = append(out.Roles, &role)
	}
	out.TotalCount = int32(len(out.Roles)) //nolint:gosec

	log.Printf("✅ Found %d roles for CBU", len(out.Roles))
	return out, nil
}

// ============================================================================
// Attributes / Dictionary
// ============================================================================

func (s *OntologyService) GetAttribute(ctx context.Context, req *pb.GetAttributeRequest) (*pb.Attribute, error) {
	log.Printf("📖 GetAttribute: id=%s", req.Id)

	row := DB.QueryRow(ctx, `
	  SELECT id, code, name, COALESCE(description,''), attr_type, 
	         COALESCE(jurisdiction,''), COALESCE(sink_table,''), 
	         COALESCE(sink_column,''), COALESCE(source_priority::text,'{}')
	    FROM dictionary_attribute WHERE id=$1`, req.Id)

	var a pb.Attribute
	if err := row.Scan(&a.Id, &a.Code, &a.Name, &a.Description, &a.AttrType,
		&a.Jurisdiction, &a.SinkTable, &a.SinkColumn, &a.SourcePriority); err != nil {
		return nil, fmt.Errorf("attribute not found: %w", err)
	}

	log.Printf("✅ Found attribute: %s", a.Name)
	return &a, nil
}

func (s *OntologyService) ListAttributes(ctx context.Context, req *pb.ListAttributesRequest) (*pb.AttributeList, error) {
	log.Printf("📖 ListAttributes: limit=%d, offset=%d", req.Limit, req.Offset)

	limit := req.Limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	rows, err := DB.Query(ctx, `
	  SELECT id, code, name, COALESCE(description,''), attr_type,
	         COALESCE(jurisdiction,''), COALESCE(sink_table,''),
	         COALESCE(sink_column,''), COALESCE(source_priority::text,'{}')
	    FROM dictionary_attribute
	    ORDER BY code LIMIT $1 OFFSET $2`, limit, req.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := &pb.AttributeList{}
	for rows.Next() {
		var a pb.Attribute
		if err := rows.Scan(&a.Id, &a.Code, &a.Name, &a.Description, &a.AttrType,
			&a.Jurisdiction, &a.SinkTable, &a.SinkColumn, &a.SourcePriority); err != nil {
			continue
		}
		list.Attributes = append(list.Attributes, &a)
	}

	var total int32
	if err := DB.QueryRow(ctx, "SELECT COUNT(*) FROM dictionary_attribute").Scan(&total); err != nil {
		total = int32(len(list.Attributes)) //nolint:gosec
	}
	list.TotalCount = total

	log.Printf("✅ Listed %d attributes (total: %d)", len(list.Attributes), total)
	return list, nil
}

func (s *OntologyService) SearchAttributes(ctx context.Context, req *pb.SearchRequest) (*pb.AttributeList, error) {
	log.Printf("🔍 SearchAttributes: query=%s", req.Query)

	limit := req.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	rows, err := DB.Query(ctx, `
	  SELECT id, code, name, COALESCE(description,''), attr_type,
	         COALESCE(jurisdiction,''), COALESCE(sink_table,''),
	         COALESCE(sink_column,''), COALESCE(source_priority::text,'{}')
	    FROM dictionary_attribute
	   WHERE name ILIKE '%' || $1 || '%' 
	      OR code ILIKE '%' || $1 || '%'
	      OR description ILIKE '%' || $1 || '%'
	   ORDER BY code LIMIT $2`, req.Query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := &pb.AttributeList{}
	for rows.Next() {
		var a pb.Attribute
		if err := rows.Scan(&a.Id, &a.Code, &a.Name, &a.Description, &a.AttrType,
			&a.Jurisdiction, &a.SinkTable, &a.SinkColumn, &a.SourcePriority); err != nil {
			continue
		}
		out.Attributes = append(out.Attributes, &a)
	}
	out.TotalCount = int32(len(out.Attributes)) //nolint:gosec

	log.Printf("✅ Found %d attributes matching '%s'", len(out.Attributes), req.Query)
	return out, nil
}

// ============================================================================
// Concepts
// ============================================================================

func (s *OntologyService) GetConcept(ctx context.Context, req *pb.GetConceptRequest) (*pb.Concept, error) {
	log.Printf("💡 GetConcept: id=%s", req.Id)

	row := DB.QueryRow(ctx, `
	  SELECT id, code, name, COALESCE(description,''), COALESCE(domain,''), synonyms
	    FROM dictionary_concept WHERE id=$1`, req.Id)

	var c pb.Concept
	if err := row.Scan(&c.Id, &c.Code, &c.Name, &c.Description, &c.Domain, &c.Synonyms); err != nil {
		return nil, fmt.Errorf("concept not found: %w", err)
	}

	log.Printf("✅ Found concept: %s", c.Name)
	return &c, nil
}

func (s *OntologyService) SearchConcepts(ctx context.Context, req *pb.SearchRequest) (*pb.ConceptList, error) {
	log.Printf("🔍 SearchConcepts: query=%s", req.Query)

	limit := req.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	rows, err := DB.Query(ctx, `
	  SELECT id, code, name, COALESCE(description,''), COALESCE(domain,''), synonyms
	    FROM dictionary_concept
	   WHERE name ILIKE '%' || $1 || '%' 
	      OR description ILIKE '%' || $1 || '%'
	   ORDER BY name LIMIT $2`, req.Query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := &pb.ConceptList{}
	for rows.Next() {
		var c pb.Concept
		if err := rows.Scan(&c.Id, &c.Code, &c.Name, &c.Description, &c.Domain, &c.Synonyms); err != nil {
			continue
		}
		out.Concepts = append(out.Concepts, &c)
	}
	out.TotalCount = int32(len(out.Concepts)) //nolint:gosec

	log.Printf("✅ Found %d concepts matching '%s'", len(out.Concepts), req.Query)
	return out, nil
}

// ============================================================================
// Regulations & Documents
// ============================================================================

func (s *OntologyService) GetRegulation(ctx context.Context, req *pb.GetRegulationRequest) (*pb.Regulation, error) {
	log.Printf("📜 GetRegulation: id=%s", req.Id)

	row := DB.QueryRow(ctx, `
	  SELECT id, code, name, jurisdiction, COALESCE(authority,''), COALESCE(description,'')
	    FROM dictionary_regulation WHERE id=$1`, req.Id)

	var r pb.Regulation
	if err := row.Scan(&r.Id, &r.Code, &r.Name, &r.Jurisdiction, &r.Authority, &r.Description); err != nil {
		return nil, fmt.Errorf("regulation not found: %w", err)
	}
	return &r, nil
}

func (s *OntologyService) ListRegulations(ctx context.Context, req *pb.ListRegulationsRequest) (*pb.RegulationList, error) {
	log.Printf("📜 ListRegulations: limit=%d", req.Limit)

	limit := req.Limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	rows, err := DB.Query(ctx, `
	  SELECT id, code, name, jurisdiction, COALESCE(authority,''), COALESCE(description,'')
	    FROM dictionary_regulation ORDER BY code LIMIT $1 OFFSET $2`, limit, req.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := &pb.RegulationList{}
	for rows.Next() {
		var r pb.Regulation
		if err := rows.Scan(&r.Id, &r.Code, &r.Name, &r.Jurisdiction, &r.Authority, &r.Description); err != nil {
			continue
		}
		list.Regulations = append(list.Regulations, &r)
	}

	var total int32
	if err := DB.QueryRow(ctx, "SELECT COUNT(*) FROM dictionary_regulation").Scan(&total); err != nil {
		total = int32(len(list.Regulations)) //nolint:gosec
	}
	list.TotalCount = total
	return list, nil
}

func (s *OntologyService) GetDocument(ctx context.Context, req *pb.GetDocumentRequest) (*pb.Document, error) {
	log.Printf("📄 GetDocument: id=%s", req.Id)

	row := DB.QueryRow(ctx, `
	  SELECT id, code, title, COALESCE(jurisdiction,''), COALESCE(category,''), COALESCE(description,'')
	    FROM dictionary_document WHERE id=$1`, req.Id)

	var d pb.Document
	if err := row.Scan(&d.Id, &d.Code, &d.Title, &d.Jurisdiction, &d.Category, &d.Description); err != nil {
		return nil, fmt.Errorf("document not found: %w", err)
	}
	return &d, nil
}

func (s *OntologyService) ListDocuments(ctx context.Context, req *pb.ListDocumentsRequest) (*pb.DocumentList, error) {
	log.Printf("📄 ListDocuments: limit=%d", req.Limit)

	limit := req.Limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	rows, err := DB.Query(ctx, `
	  SELECT id, code, title, COALESCE(jurisdiction,''), COALESCE(category,''), COALESCE(description,'')
	    FROM dictionary_document ORDER BY code LIMIT $1 OFFSET $2`, limit, req.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := &pb.DocumentList{}
	for rows.Next() {
		var d pb.Document
		if err := rows.Scan(&d.Id, &d.Code, &d.Title, &d.Jurisdiction, &d.Category, &d.Description); err != nil {
			continue
		}
		list.Documents = append(list.Documents, &d)
	}

	var total int32
	if err := DB.QueryRow(ctx, "SELECT COUNT(*) FROM dictionary_document").Scan(&total); err != nil {
		total = int32(len(list.Documents)) //nolint:gosec
	}
	list.TotalCount = total
	return list, nil
}

// ============================================================================
// Entity Control Graph
// ============================================================================

func (s *OntologyService) GetEntityControlGraph(ctx context.Context, req *pb.GetEntityControlRequest) (*pb.EntityControlGraph, error) {
	log.Printf("🕸️  GetEntityControlGraph: entity=%s", req.EntityId)

	rows, err := DB.Query(ctx, `
	  SELECT id, controller_entity_id, controlled_entity_id, control_type::text,
	         COALESCE(control_percentage, 0), COALESCE(control_basis,''), is_indirect, COALESCE(remarks,'')
	    FROM entity_control
	   WHERE controller_entity_id=$1 OR controlled_entity_id=$1`, req.EntityId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	graph := &pb.EntityControlGraph{}
	for rows.Next() {
		var edge pb.EntityControl
		if err := rows.Scan(&edge.Id, &edge.ControllerEntityId, &edge.ControlledEntityId,
			&edge.ControlType, &edge.ControlPercentage, &edge.ControlBasis, &edge.IsIndirect, &edge.Remarks); err != nil {
			continue
		}
		graph.Edges = append(graph.Edges, &edge)
	}
	graph.TotalEdges = int32(len(graph.Edges)) //nolint:gosec

	log.Printf("✅ Found %d control edges", len(graph.Edges))
	return graph, nil
}

// Stub implementations for unimplemented methods
func (s *OntologyService) CreateEntity(ctx context.Context, req *pb.CreateEntityRequest) (*pb.EntityResponse, error) {
	return &pb.EntityResponse{Success: false, Error: "not implemented"}, nil
}

func (s *OntologyService) UpdateEntity(ctx context.Context, req *pb.UpdateEntityRequest) (*pb.EntityResponse, error) {
	return &pb.EntityResponse{Success: false, Error: "not implemented"}, nil
}

func (s *OntologyService) ListCbus(ctx context.Context, req *pb.ListCbusRequest) (*pb.CbuList, error) {
	return &pb.CbuList{}, nil
}

func (s *OntologyService) CreateCbu(ctx context.Context, req *pb.CreateCbuRequest) (*pb.CbuResponse, error) {
	return &pb.CbuResponse{Success: false, Error: "not implemented"}, nil
}

func (s *OntologyService) AssignCbuRole(ctx context.Context, req *pb.AssignCbuRoleRequest) (*pb.CbuRoleResponse, error) {
	return &pb.CbuRoleResponse{Success: false, Error: "not implemented"}, nil
}

func (s *OntologyService) ListConcepts(ctx context.Context, req *pb.ListConceptsRequest) (*pb.ConceptList, error) {
	return &pb.ConceptList{}, nil
}

func (s *OntologyService) CreateControl(ctx context.Context, req *pb.CreateControlRequest) (*pb.ControlResponse, error) {
	return &pb.ControlResponse{Success: false, Error: "not implemented"}, nil
}

func (s *OntologyService) GetControlChain(ctx context.Context, req *pb.GetControlChainRequest) (*pb.ControlChain, error) {
	return &pb.ControlChain{}, nil
}

func (s *OntologyService) GetKycProfile(ctx context.Context, req *pb.GetKycProfileRequest) (*pb.KycProfile, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *OntologyService) UpdateKycProfile(ctx context.Context, req *pb.UpdateKycProfileRequest) (*pb.KycProfileResponse, error) {
	return &pb.KycProfileResponse{Success: false, Error: "not implemented"}, nil
}
