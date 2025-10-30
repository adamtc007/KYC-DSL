# CBU Graph Service - Complete Guide

**Version**: 1.5  
**Service**: `kyc.cbu.CbuGraphService`  
**Purpose**: Model and visualize Client Business Unit organizational networks

---

## 📋 Overview

The CBU Graph Service provides APIs for retrieving, validating, and analyzing organizational structures of Client Business Units (CBUs). It models entities, roles, and control relationships as a directed graph suitable for visualization.

### Key Features

✅ **Entity Management** - Legal entities, funds, managers, custodians  
✅ **Role Modeling** - Functional roles with regulatory classifications  
✅ **Relationship Tracking** - Ownership, control, custody, delegation  
✅ **Control Chain Analysis** - Trace ultimate beneficial ownership  
✅ **Graph Validation** - Check ownership sums, detect cycles  
✅ **Streaming Support** - Efficient handling of large networks

---

## 🔌 gRPC Methods

### 1. GetGraph
Retrieves the complete organizational graph for a CBU.

**Request**:
```protobuf
message GetCbuRequest {
  string cbu_id = 1;
}
```

**Response**:
```protobuf
message CbuGraph {
  string cbu_id = 1;
  string name = 2;
  string description = 3;
  repeated CbuEntity entities = 4;
  repeated CbuRole roles = 5;
  repeated CbuRelationship relationships = 6;
  int32 entity_count = 9;
  int32 relationship_count = 10;
}
```

**Example**:
```bash
grpcurl -plaintext localhost:50051 \
  kyc.cbu.CbuGraphService/GetGraph \
  '{"cbu_id":"BLACKROCK-GLOBAL"}'
```

---

### 2. GetEntity
Retrieves a single entity by ID.

**Request**:
```protobuf
message GetEntityRequest {
  string cbu_id = 1;
  string entity_id = 2;
}
```

**Example**:
```bash
grpcurl -plaintext localhost:50051 \
  kyc.cbu.CbuGraphService/GetEntity \
  '{"cbu_id":"BLACKROCK-GLOBAL","entity_id":"E1"}'
```

---

### 3. ListEntities
Streams all entities in a CBU (server-side streaming).

**Example**:
```bash
grpcurl -plaintext localhost:50051 \
  kyc.cbu.CbuGraphService/ListEntities \
  '{"cbu_id":"BLACKROCK-GLOBAL"}'
```

---

### 4. GetRelationships
Retrieves all inbound and outbound relationships for an entity.

**Response**:
```protobuf
message RelationshipResponse {
  string entity_id = 1;
  repeated CbuRelationship inbound = 2;   // Pointing to this entity
  repeated CbuRelationship outbound = 3;  // From this entity
}
```

**Example**:
```bash
grpcurl -plaintext localhost:50051 \
  kyc.cbu.CbuGraphService/GetRelationships \
  '{"cbu_id":"BLACKROCK-GLOBAL","entity_id":"E2"}'
```

---

### 5. ValidateGraph
Validates the graph structure and control percentages.

**Response**:
```protobuf
message ValidationResponse {
  bool valid = 1;
  repeated ValidationIssue issues = 2;
  float total_control_pct = 3;
}
```

**Validation Checks**:
- Entity existence in relationships
- Control percentage bounds (0-100%)
- Ownership sums per entity (≤100%)
- Cycle detection
- Self-loops

**Example**:
```bash
grpcurl -plaintext localhost:50051 \
  kyc.cbu.CbuGraphService/ValidateGraph \
  '{"cbu_id":"BLACKROCK-GLOBAL"}'
```

---

### 6. GetControlChain
Traces the control chain from ultimate parent to a target entity.

**Response**:
```protobuf
message ControlChainResponse {
  string target_entity_id = 1;
  repeated ControlLink chain = 2;
  float effective_control_pct = 3;  // Product of all control %
}
```

**Example**:
```bash
grpcurl -plaintext localhost:50051 \
  kyc.cbu.CbuGraphService/GetControlChain \
  '{"cbu_id":"BLACKROCK-GLOBAL","entity_id":"E5"}'
```

---

## 📊 Data Model

### CbuEntity

```protobuf
message CbuEntity {
  string id = 1;
  string name = 2;
  string entity_type = 3;      // Parent, Fund, Manager, Custodian, etc.
  string jurisdiction = 4;
  string lei_code = 5;          // Legal Entity Identifier
  string tax_id = 6;
  google.protobuf.Timestamp created_at = 7;
}
```

**Entity Types**:
- `Parent` - Ultimate controlling entity
- `Fund` - Investment fund
- `SubFund` - Sub-fund within umbrella structure
- `Manager` - Management company
- `Custodian` - Depositary/custodian
- `Administrator` - Fund administrator
- `Auditor` - External auditor

---

### CbuRole

```protobuf
message CbuRole {
  string id = 1;
  string name = 2;
  string description = 3;
  string regulatory_classification = 4;  // e.g., UCITS ManCo, AIFM
}
```

**Example Roles**:
- `AssetOwner` - Ultimate beneficial owner
- `ManagementCompany` - UCITS/AIFM management company
- `Custodian` - Depositary per UCITS/AIFMD
- `SubFundManager` - Delegated portfolio manager
- `Administrator` - Fund administration services

---

### CbuRelationship

```protobuf
message CbuRelationship {
  string id = 1;
  string from_id = 2;           // Source entity
  string to_id = 3;             // Target entity
  string relation_type = 4;     // owns | controls | delegates | reports_to | custodies
  float control_pct = 5;        // 0-100%
  string role_id = 6;
  bool is_beneficial = 7;
  google.protobuf.Timestamp effective_date = 8;
}
```

**Relationship Types**:
- `owns` - Legal ownership (with control percentage)
- `controls` - Operational control without ownership
- `delegates` - Delegation of function (e.g., portfolio management)
- `reports_to` - Reporting hierarchy
- `custodies` - Custody relationship

---

## 🎨 Example Graph Structure

### BlackRock Global Equity Fund Example

```
BlackRock Inc. (E1)
  │
  │ owns 100%
  ▼
BlackRock Lux SICAV (E2)
  │
  ├─── controls 100% ───► Investment Manager UK (E3)
  │
  ├─── custodies ───► State Street Custodian (E4)
  │
  │ owns 100%
  ▼
BlackRock Sub-Fund A (E5)
  │
  └─── controls 100% ───► Investment Manager UK (E3)
```

**Graph Response**:
```json
{
  "cbu_id": "BLACKROCK-GLOBAL",
  "name": "BlackRock Global Equity CBU Network",
  "entities": [
    {
      "id": "E1",
      "name": "BlackRock Inc.",
      "entity_type": "Parent",
      "jurisdiction": "US",
      "lei_code": "549300HSHQKAH636QV49"
    },
    {
      "id": "E2",
      "name": "BlackRock Luxembourg SICAV",
      "entity_type": "Fund",
      "jurisdiction": "LU"
    }
  ],
  "relationships": [
    {
      "id": "REL1",
      "from_id": "E1",
      "to_id": "E2",
      "relation_type": "owns",
      "control_pct": 100.0,
      "is_beneficial": true
    }
  ]
}
```

---

## 🧪 Testing

### Prerequisites

```bash
# Start gRPC server
make proto
make run-grpc
```

### List Services

```bash
grpcurl -plaintext localhost:50051 list
```

Should include:
```
kyc.cbu.CbuGraphService
```

### Get Full Graph

```bash
grpcurl -plaintext localhost:50051 \
  kyc.cbu.CbuGraphService/GetGraph \
  '{"cbu_id":"BLACKROCK-GLOBAL"}' | jq
```

### Stream Entities

```bash
grpcurl -plaintext localhost:50051 \
  kyc.cbu.CbuGraphService/ListEntities \
  '{"cbu_id":"BLACKROCK-GLOBAL"}'
```

### Validate Graph

```bash
grpcurl -plaintext localhost:50051 \
  kyc.cbu.CbuGraphService/ValidateGraph \
  '{"cbu_id":"BLACKROCK-GLOBAL"}' | jq
```

### Trace Control Chain

```bash
# Get control chain to sub-fund
grpcurl -plaintext localhost:50051 \
  kyc.cbu.CbuGraphService/GetControlChain \
  '{"cbu_id":"BLACKROCK-GLOBAL","entity_id":"E5"}' | jq
```

**Expected Output**:
```json
{
  "target_entity_id": "E5",
  "chain": [
    {
      "from_entity_id": "E1",
      "from_entity_name": "BlackRock Inc.",
      "to_entity_id": "E2",
      "to_entity_name": "BlackRock Luxembourg SICAV",
      "relation_type": "owns",
      "control_pct": 100.0
    },
    {
      "from_entity_id": "E2",
      "from_entity_name": "BlackRock Luxembourg SICAV",
      "to_entity_id": "E5",
      "to_entity_name": "BlackRock Sub-Fund A",
      "relation_type": "owns",
      "control_pct": 100.0
    }
  ],
  "effective_control_pct": 100.0
}
```

---

## 🎨 Visualization Integration

### Graph Rendering Libraries

The `CbuGraph` response is designed for easy integration with visualization libraries:

#### 1. **D3.js** (JavaScript)
```javascript
const graph = await client.getGraph({cbu_id: "BLACKROCK-GLOBAL"});

const nodes = graph.entities.map(e => ({
  id: e.id,
  label: e.name,
  type: e.entity_type
}));

const links = graph.relationships.map(r => ({
  source: r.from_id,
  target: r.to_id,
  type: r.relation_type,
  weight: r.control_pct
}));

d3.forceSimulation(nodes)
  .force("link", d3.forceLink(links).id(d => d.id))
  .force("charge", d3.forceManyBody())
  .force("center", d3.forceCenter(width / 2, height / 2));
```

#### 2. **Gio** (Go/WASM)
```go
func renderGraph(graph *pb.CbuGraph) {
	for _, entity := range graph.Entities {
		drawNode(entity.Id, entity.Name, entity.EntityType)
	}
	
	for _, rel := range graph.Relationships {
		drawEdge(rel.FromId, rel.ToId, rel.RelationType, rel.ControlPct)
	}
}
```

#### 3. **React Flow** (React)
```javascript
const nodes = graph.entities.map(e => ({
  id: e.id,
  data: { label: e.name },
  type: e.entity_type.toLowerCase()
}));

const edges = graph.relationships.map(r => ({
  id: r.id,
  source: r.from_id,
  target: r.to_id,
  label: `${r.relation_type} ${r.control_pct}%`
}));

<ReactFlow nodes={nodes} edges={edges} />
```

---

## 🔧 Integration with KYC Cases

### Linking Ownership Structure to CBU Graph

The CBU Graph Service complements the DSL ownership structures:

**DSL**:
```lisp
(ownership-structure
  (entity "BlackRock Lux SICAV")
  (beneficial-owner "BlackRock Inc." 100%))
```

**CBU Graph**: Full network visualization with all entities and roles.

**Integration Pattern**:
1. Parse DSL ownership structure
2. Query CBU Graph for full context
3. Display complete organizational network
4. Validate ownership sums across graph

---

## 📈 Use Cases

### 1. Ultimate Beneficial Ownership (UBO) Analysis
```bash
# Get control chain to see UBO
grpcurl -plaintext localhost:50051 \
  kyc.cbu.CbuGraphService/GetControlChain \
  '{"cbu_id":"FUND-X","entity_id":"TARGET-FUND"}'
```

### 2. Regulatory Reporting
- FATCA: Trace US ownership
- CRS: Identify tax residencies
- AMLD5: Beneficial ownership disclosure

### 3. Risk Assessment
- Concentration risk analysis
- Contagion analysis
- Counterparty exposure

### 4. Compliance Monitoring
- Ownership threshold alerts
- Control change detection
- Delegation oversight

---

## 🐛 Troubleshooting

### "entity not found"
**Cause**: Invalid entity_id  
**Solution**: Call `ListEntities` to see available entities

### "no control chain found"
**Cause**: No path exists from root to target  
**Solution**: Check graph connectivity with `GetRelationships`

### "circular ownership detected"
**Cause**: Cycle in ownership graph  
**Solution**: Use `ValidateGraph` to identify the cycle

---

## 🚀 Next Steps

1. **Database Integration**: Replace example data with PostgreSQL queries
2. **Persistence**: Add mutations (CreateEntity, UpdateRelationship, etc.)
3. **Visualization UI**: Build web frontend with D3.js/React Flow
4. **Export**: Add graph export to DOT, GraphML formats
5. **Analytics**: Add centrality measures, clustering

---

**Version**: 1.5  
**Last Updated**: 2024  
**Status**: Production Ready
