# KYC-DSL Improvement Recommendations

**Status**: Non-urgent refinements  
**Priority**: Low to Medium  
**Timeline**: Post-production optimization

This document tracks architectural and operational improvements for the KYC-DSL system.

---

## 🎯 Recommended Micro-Refinements

### 1. Unify Proto Sources

**Current State:**
- Proto files scattered across `api/proto/` and referenced from both Go and Rust
- Build scripts use relative paths (`../../api/proto/`)

**Improvement:**
Move all `.proto` files into a single `proto_shared/` directory and regenerate stubs.

**Rationale:**
- Single source of truth for API definitions
- Cleaner build scripts
- Easier to version and maintain
- Better for mono-repo structure

**Implementation:**
```bash
# 1. Create unified proto directory
mkdir -p proto_shared
mv api/proto/*.proto proto_shared/

# 2. Update Rust build.rs
# Change: ../../api/proto/dsl_service.proto
# To:     ../../proto_shared/dsl_service.proto

# 3. Update Go protoc commands in Makefile
# Change: api/proto/
# To:     proto_shared/

# 4. Regenerate all stubs
make proto          # Go stubs
cd rust && cargo build -p kyc_dsl_service  # Rust stubs
```

**Files to Update:**
- `rust/kyc_dsl_service/build.rs`
- `Makefile` (proto target)
- Any CI/CD scripts

---

### 2. Add Service Registry in Go

**Current State:**
- Services expose gRPC endpoints but no runtime reflection registry
- Client discovery requires hardcoded endpoints

**Improvement:**
Add `api/registry.go` so each service (DSL, Ontology, RAG, Data) registers its proto descriptor for runtime reflection.

**Rationale:**
- Dynamic service discovery
- Better tooling support (grpcurl, grpc-ui)
- Easier testing and debugging
- Supports service mesh integration

**Implementation:**

```go
// api/registry.go
package api

import (
    "google.golang.org/grpc"
    "google.golang.org/grpc/reflection"
)

type ServiceRegistry struct {
    services map[string]grpc.ServiceDesc
}

func NewRegistry() *ServiceRegistry {
    return &ServiceRegistry{
        services: make(map[string]grpc.ServiceDesc),
    }
}

func (r *ServiceRegistry) Register(name string, desc grpc.ServiceDesc) {
    r.services[name] = desc
}

func (r *ServiceRegistry) RegisterWithServer(s *grpc.Server) {
    reflection.Register(s)
    for _, desc := range r.services {
        s.RegisterService(&desc, nil)
    }
}
```

**Usage in cmd/server/main.go:**
```go
registry := api.NewRegistry()
registry.Register("DslService", pb.DslService_ServiceDesc)
registry.Register("RagService", pb.RagService_ServiceDesc)
registry.RegisterWithServer(grpcServer)
```

**Benefits:**
- ✅ Runtime introspection
- ✅ Auto-generated API docs
- ✅ Service mesh compatibility
- ✅ Better monitoring

---

### 3. Enable SQL Linting

**Current State:**
- SQL queries use string concatenation in some places
- No automated SQL validation
- Potential for SQL injection or syntax errors

**Improvement:**
Implement SQL linting (sqlc or go-vet rules) to catch string concatenation and validate queries at build time.

**Rationale:**
- Prevent SQL injection vulnerabilities
- Catch syntax errors before runtime
- Type-safe query generation
- Better IDE support

**Option A: sqlc (Recommended)**

```yaml
# sqlc.yaml
version: "2"
sql:
  - schema: "internal/storage/migrations/"
    queries: "internal/storage/queries/"
    engine: "postgresql"
    gen:
      go:
        package: "storage"
        out: "internal/storage"
        emit_json_tags: true
        emit_prepared_queries: true
```

**Install:**
```bash
brew install sqlc
# or
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
```

**Usage:**
```bash
sqlc generate
```

**Option B: Custom go-vet Rules**

```go
// tools/sqlvet/main.go
// Custom vet rule to detect unsafe SQL patterns
func checkSQLConcatenation(pass *analysis.Pass) {
    // Flag: fmt.Sprintf("SELECT * FROM ... %s", userInput)
    // Flag: "SELECT * FROM " + variable
}
```

**Integration:**
```makefile
# Makefile
lint-sql:
	sqlc vet
	go vet -vettool=$(which sqlvet) ./...
```

**Benefits:**
- ✅ Compile-time SQL validation
- ✅ Type-safe queries
- ✅ Prevent injection attacks
- ✅ Auto-complete in IDE

---

### 4. Implement Connection Pool Metrics

**Current State:**
- Database connections managed by sqlx
- Rust→gRPC→Go calls have no visibility into queuing
- No metrics on connection pool exhaustion

**Improvement:**
Add Prometheus or OpenTelemetry metrics to track connection pool usage and gRPC latency.

**Rationale:**
- Identify bottlenecks under load
- Prevent connection pool exhaustion
- Monitor Rust→Go→DB latency
- Capacity planning data

**Implementation:**

**Option A: Prometheus (Simpler)**

```go
// internal/storage/metrics.go
package storage

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    dbConnectionsOpen = promauto.NewGauge(prometheus.GaugeOpts{
        Name: "kyc_db_connections_open",
        Help: "Number of open database connections",
    })
    
    dbConnectionsInUse = promauto.NewGauge(prometheus.GaugeOpts{
        Name: "kyc_db_connections_in_use",
        Help: "Number of connections currently in use",
    })
    
    dbConnectionWaitDuration = promauto.NewHistogram(prometheus.HistogramOpts{
        Name:    "kyc_db_connection_wait_duration_seconds",
        Help:    "Time spent waiting for a connection",
        Buckets: prometheus.DefBuckets,
    })
    
    grpcRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
        Name:    "kyc_grpc_request_duration_seconds",
        Help:    "gRPC request latency",
        Buckets: prometheus.DefBuckets,
    }, []string{"service", "method"})
)

func RecordPoolMetrics(db *sql.DB) {
    stats := db.Stats()
    dbConnectionsOpen.Set(float64(stats.OpenConnections))
    dbConnectionsInUse.Set(float64(stats.InUse))
    dbConnectionWaitDuration.Observe(stats.WaitDuration.Seconds())
}
```

**Usage in main.go:**
```go
// Start metrics endpoint
go func() {
    http.Handle("/metrics", promhttp.Handler())
    http.ListenAndServe(":9090", nil)
}()

// Record metrics periodically
go func() {
    ticker := time.NewTicker(10 * time.Second)
    for range ticker.C {
        storage.RecordPoolMetrics(db)
    }
}()
```

**Option B: OpenTelemetry (More Features)**

```go
// internal/telemetry/otel.go
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/metric"
)

func InitMetrics() {
    meter := otel.Meter("kyc-dsl")
    
    connectionGauge, _ := meter.Int64ObservableGauge(
        "db.connections.open",
        metric.WithDescription("Open DB connections"),
    )
    
    // Auto-observe from db.Stats()
    meter.RegisterCallback(func(ctx context.Context, o metric.Observer) error {
        stats := db.Stats()
        o.ObserveInt64(connectionGauge, int64(stats.OpenConnections))
        return nil
    }, connectionGauge)
}
```

**Grafana Dashboard Queries:**
```promql
# Connection pool usage
rate(kyc_db_connections_in_use[5m])

# Wait time P95
histogram_quantile(0.95, kyc_db_connection_wait_duration_seconds)

# gRPC latency by method
histogram_quantile(0.99, kyc_grpc_request_duration_seconds)
```

**Benefits:**
- ✅ Real-time pool monitoring
- ✅ Identify slow queries
- ✅ Prevent connection exhaustion
- ✅ Capacity planning
- ✅ Alert on anomalies

---

## 📊 Priority Matrix

| Improvement | Effort | Impact | Priority |
|-------------|--------|--------|----------|
| Unify Proto Sources | Low | Medium | **P2** |
| Service Registry | Low | Medium | **P2** |
| SQL Linting | Medium | High | **P1** |
| Connection Pool Metrics | Medium | High | **P1** |

---

## 🚀 Implementation Plan

### Phase 1: Security & Quality (P1)
1. **Week 1**: Implement SQL linting with sqlc
   - Set up sqlc.yaml
   - Migrate queries to .sql files
   - Integrate with CI/CD

2. **Week 2**: Add connection pool metrics
   - Set up Prometheus endpoint
   - Add pool metrics
   - Create Grafana dashboard
   - Set up alerts

### Phase 2: Architecture Refinement (P2)
3. **Week 3**: Unify proto sources
   - Create proto_shared/
   - Update build scripts
   - Regenerate all stubs
   - Update documentation

4. **Week 4**: Implement service registry
   - Create api/registry.go
   - Update service initialization
   - Add reflection support
   - Test with grpcurl

---

## 📝 Additional Improvements (Future)

### Performance
- [ ] Add Redis cache for ontology lookups
- [ ] Implement gRPC connection pooling
- [ ] Enable HTTP/2 keepalive
- [ ] Add request batching for RAG queries

### Observability
- [ ] Distributed tracing (Jaeger/Tempo)
- [ ] Structured logging (zap/zerolog)
- [ ] Error tracking (Sentry)
- [ ] APM integration

### Testing
- [ ] Load testing with k6
- [ ] Chaos engineering tests
- [ ] Fuzz testing for parser
- [ ] Contract testing for gRPC

### DevOps
- [ ] Kubernetes deployment manifests
- [ ] Helm charts
- [ ] CI/CD pipelines (GitHub Actions)
- [ ] Multi-stage Docker builds

---

## 🔍 Monitoring Dashboard Goals

**Key Metrics to Track:**
1. **Database Pool**
   - Open connections
   - In-use connections
   - Wait duration (P95, P99)
   - Max lifetime reached

2. **gRPC Performance**
   - Request rate (by service/method)
   - Latency (P50, P95, P99)
   - Error rate
   - Queue depth

3. **Application**
   - Case processing time
   - Amendment throughput
   - RAG query latency
   - Cache hit ratio

4. **System**
   - CPU/Memory usage
   - Goroutine count
   - GC pause time
   - Network I/O

---

## 📚 References

**SQL Linting:**
- sqlc: https://sqlc.dev/
- go-vet: https://pkg.go.dev/cmd/vet

**Metrics:**
- Prometheus: https://prometheus.io/
- OpenTelemetry: https://opentelemetry.io/
- Grafana: https://grafana.com/

**gRPC:**
- gRPC Reflection: https://github.com/grpc/grpc-go/tree/master/reflection
- Service Discovery: https://grpc.io/docs/guides/service-discovery/

---

**Last Updated**: 2024  
**Version**: 1.5  
**Status**: Planning Phase