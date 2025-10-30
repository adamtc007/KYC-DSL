# Rust Quick Reference Card

## 🚀 Essential Commands

### Build
```bash
make rust-build              # Release build
cd rust && cargo build       # Debug build
cd rust && cargo check       # Fast syntax check
```

### Test
```bash
make rust-test               # All tests
cd rust && cargo test        # Same
cd rust && cargo test -- --nocapture  # With output
```

### Run
```bash
make run-rust                # Start gRPC service (port 50060)
cd rust/kyc_dsl_service && cargo run  # Direct run
```

### Code Quality
```bash
make rust-fmt                # Format code
make rust-lint               # Run clippy
make fmt-all                 # Format Go + Rust
make lint-all                # Lint Go + Rust
```

---

## 📁 File Locations

```
rust/
├── Cargo.toml                          # Workspace root
├── kyc_dsl_core/                       # Core library
│   ├── Cargo.toml
│   └── src/
│       ├── lib.rs                      # Public API
│       ├── parser.rs                   # nom parser
│       ├── compiler.rs                 # AST → Instructions
│       └── executor.rs                 # Execution engine
│
└── kyc_dsl_service/                    # gRPC service
    ├── Cargo.toml
    ├── build.rs                        # Proto generation
    └── src/
        └── main.rs                     # gRPC server

Build Output:
└── rust/target/release/kyc_dsl_service  # Binary
```

---

## 🔌 gRPC Service

### Port
- **50060** (IPv6: `[::1]:50060`)

### Available RPCs
```
kyc.dsl.DslService/Execute
kyc.dsl.DslService/Validate
kyc.dsl.DslService/Parse
kyc.dsl.DslService/Serialize
kyc.dsl.DslService/Amend
kyc.dsl.DslService/ListAmendments
kyc.dsl.DslService/GetGrammar
```

### Test with grpcurl
```bash
# List services
grpcurl -plaintext localhost:50060 list

# Validate DSL
grpcurl -plaintext localhost:50060 \
  kyc.dsl.DslService/Validate \
  -d '{"dsl":"(kyc-case TEST)"}'

# Get grammar
grpcurl -plaintext localhost:50060 \
  kyc.dsl.DslService/GetGrammar -d '{}'

# List amendments
grpcurl -plaintext localhost:50060 \
  kyc.dsl.DslService/ListAmendments -d '{}'
```

---

## 🧪 Testing

### Run Specific Test
```bash
cd rust
cargo test test_parse_atom
cargo test test_compile_simple_case
```

### Test with Output
```bash
cd rust
cargo test -- --nocapture
```

### Test One Crate
```bash
cd rust
cargo test -p kyc_dsl_core
cargo test -p kyc_dsl_service
```

---

## 🐛 Debugging

### With Logs
```bash
RUST_LOG=debug cargo run
RUST_LOG=info cargo run
```

### With Backtrace
```bash
RUST_BACKTRACE=1 cargo run
RUST_BACKTRACE=full cargo run
```

### Check Without Building
```bash
cd rust
cargo check       # Fast
cargo clippy      # With lints
```

---

## 📦 Core API

### Compile DSL
```rust
use kyc_dsl_core::compile_dsl;

let dsl = "(kyc-case TEST (nature \"Corporate\"))";
let plan = compile_dsl(dsl)?;  // Returns JSON
```

### Execute Plan
```rust
use kyc_dsl_core::execute_plan;

let result = execute_plan(&plan_json)?;
```

### Parse DSL
```rust
use kyc_dsl_core::parser;

let ast = parser::parse(dsl)?;
// Returns Expr::Call or Expr::Atom
```

---

## 🔧 Common Workflows

### Add New Instruction
1. Add handler in `executor.rs`:
   ```rust
   fn execute_my_new_instruction(args: &[String], ctx: &mut ExecutionContext) -> Result<String, String> {
       // Implementation
   }
   ```

2. Register in `execute_instruction()`:
   ```rust
   "my-new-instruction" => execute_my_new_instruction(&instruction.args, ctx)?,
   ```

3. Add test:
   ```rust
   #[test]
   fn test_my_new_instruction() {
       // Test
   }
   ```

### Add New RPC
1. Update `api/proto/dsl_service.proto`
2. Regenerate proto:
   ```bash
   make proto
   ```
3. Rebuild Rust:
   ```bash
   make rust-build
   ```
4. Implement in `kyc_dsl_service/src/main.rs`

---

## ⚡ Performance Tips

### Build Optimizations
```bash
# Release build (optimized)
cargo build --release

# Debug build (faster compile)
cargo build
```

### Profile
```bash
# Install cargo-flamegraph
cargo install flamegraph

# Generate flamegraph
cargo flamegraph
```

---

## 🚨 Troubleshooting

### Port Already in Use
```bash
lsof -ti:50060 | xargs kill -9
```

### Cannot Connect
```bash
# Try IPv4 instead of IPv6
# In main.rs: "127.0.0.1:50060" instead of "[::1]:50060"
```

### Build Fails
```bash
# Update Rust
rustup update

# Clean and rebuild
cd rust && cargo clean && cargo build
```

### Proto Generation Fails
```bash
# Check protoc installed
protoc --version

# Reinstall
brew install protobuf  # macOS
```

### Missing Dependencies
```bash
cd rust
cargo update
cargo build
```

---

## 📊 Key Metrics

| Metric | Value |
|--------|-------|
| Lines of Rust Code | 1,037 |
| Unit Tests | 12 |
| Dependencies | 4 (core) + 5 (service) |
| Build Time (release) | ~18s |
| Binary Size | ~5-10 MB |
| gRPC Port | 50060 |

---

## 🔗 Integration Points

### With Go gRPC Server (port 50051)
```go
// internal/service/dsl_client.go
conn, err := grpc.Dial("localhost:50060", ...)
```

### With REST API (port 8080)
```bash
# Go API can proxy to Rust service
curl http://localhost:8080/api/validate
```

---

## 📚 Documentation

- **Architecture**: `rust/README.md`
- **Quick Start**: `RUST_QUICKSTART.md`
- **Verification**: `RUST_MIGRATION_REPORT.md`
- **This Card**: `rust/QUICK_REFERENCE.md`

---

## 🎯 One-Liners

```bash
# Build, test, and run
make rust-build && make rust-test && make run-rust

# Full quality check
make rust-fmt && make rust-lint && make rust-test

# Development cycle
cd rust && cargo watch -x check -x test -x run

# Production build
cd rust && cargo build --release --locked

# Update dependencies
cd rust && cargo update && cargo build
```

---

**Last Updated**: 2024-10-30  
**Version**: 0.1.0