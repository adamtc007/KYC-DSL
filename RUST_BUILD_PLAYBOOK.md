# 🧭 Rust Build & Sanity Playbook

**Purpose**: Quickly verify and fix compile or link errors across the Rust half of the KYC-DSL repo after changes or refactors.

---

## ⚙️ 1️⃣ — Validate the Workspace

Run this **before anything else** when files or modules move.

```bash
cd rust
cargo metadata --format-version=1 | jq '.workspace_members'
```

✅ **Expected output**:
- `"kyc_dsl_core"` and `"kyc_dsl_service"` appear

💡 **If missing**:
1. Ensure `rust/Cargo.toml` has:
   ```toml
   [workspace]
   members = ["kyc_dsl_core", "kyc_dsl_service"]
   resolver = "2"
   ```
2. Run `cargo clean && cargo check` again

---

## ⚙️ 2️⃣ — Check the Core Crate

```bash
cargo check -p kyc_dsl_core
```

### Common Errors

| Error | Root Cause | One-line Fix |
|-------|-----------|--------------|
| `use of undeclared crate or module 'parser'` | Missing `pub mod parser;` in `lib.rs` | Add `pub mod parser;` to top of `lib.rs` |
| `no function or associated item named 'parse'` | Parser function not `pub` | Ensure `pub fn parse(...)` in `parser.rs` |
| `cannot find type 'Expr' in this scope` | Not re-exported from lib | Add `pub use parser::Expr;` in `lib.rs` |
| `expected struct, found enum` | AST mismatch between parser/compiler | Update compiler match arms to handle `Expr::Call` and `Expr::Atom` |
| `serde_json` or `nom` not found | Missing crate features | Run `cargo add serde --features derive` or check `Cargo.toml` syntax |

---

## ⚙️ 3️⃣ — Check the gRPC Service

```bash
cargo check -p kyc_dsl_service
```

### Common Errors

| Error | Root Cause | One-line Fix |
|-------|-----------|--------------|
| `could not find file 'dsl_service.proto'` | `build.rs` path wrong | Change `tonic_build::compile_protos("../../api/proto/dsl_service.proto")?;` |
| `no method named 'execute' found for trait object` | Trait mismatch | Ensure `use dsl::dsl_service_server::{DslService, DslServiceServer};` |
| `field 'dsl_source' does not exist` | Proto regenerated with diff name | Re-run `cargo clean && cargo build` to regenerate Rust proto files |
| `tokio::main` not found | Missing feature | Ensure in `Cargo.toml`: `tokio = { version = "1", features = ["macros","rt-multi-thread"] }` |

---

## ⚙️ 4️⃣ — End-to-End Runtime Validation

Start the Rust gRPC service:

```bash
cargo run -p kyc_dsl_service
```

✅ **Expected output**:
```
🦀 Rust DSL gRPC Service
========================
Listening on: [::1]:50060
Protocol: gRPC (HTTP/2)
Service: kyc.dsl.DslService
...
Ready to accept connections...
```

Then, in a **new terminal**:

```bash
grpcurl -plaintext localhost:50060 list
```

**Should list**:
```
grpc.reflection.v1.ServerReflection
kyc.dsl.DslService
```

---

## ⚙️ 5️⃣ — Test Go → Rust Round-Trip

Test individual RPC calls:

```bash
# Execute function
grpcurl -plaintext -d '{
  "case_id": "TEST-001",
  "function_name": "DISCOVER-POLICIES"
}' localhost:50060 kyc.dsl.DslService/Execute
```

✅ **Expect**:
```json
{
  "updatedDsl": "(kyc-case TEST-001 (function DISCOVER-POLICIES))",
  "message": "Executed function 'DISCOVER-POLICIES' on case 'TEST-001'",
  "success": true,
  "caseId": "TEST-001",
  "newVersion": 1
}
```

**If you get** `"rpc error: code = Unavailable desc = connection refused"`:
→ Rust service not running

---

## ⚙️ 6️⃣ — Clean Rebuild Recipe

Use when **switching branches** or **modifying Cargo files**.

```bash
cd rust
cargo clean
rm -rf kyc_dsl_service/target kyc_dsl_core/target
cargo build --release
```

---

## ⚙️ 7️⃣ — IDE & Zed Integration

In **Zed's terminal**, run:

```bash
cargo fmt
cargo clippy --fix
```

This ensures **agents** (Claude, ChatGPT) and **Zed's LSP** see the same canonical AST and style.

---

## ⚙️ 8️⃣ — CI Preflight (Automated Script)

You can automate all the above checks via `/rust/preflight.sh`:

```bash
#!/usr/bin/env bash
set -e
cd "$(dirname "$0")"
echo "==> Checking workspace..."
cargo metadata --quiet
echo "==> Checking core..."
cargo check -p kyc_dsl_core
echo "==> Checking gRPC service..."
cargo check -p kyc_dsl_service
echo "==> Running tests..."
cargo test
echo "==> All good ✅"
```

**Run from repo root**:

```bash
bash rust/preflight.sh
```

Or use the **enhanced version** (already in repo):

```bash
cd rust
./preflight.sh
```

The enhanced version includes:
- ✅ Toolchain validation (Rust, Go, protoc)
- ✅ Database connectivity checks (PostgreSQL, pgvector)
- ✅ Port availability checks (50060, 50051, 8080, 5432)
- ✅ Workspace metadata validation
- ✅ Cargo check for both crates
- ✅ Test execution
- ✅ Clippy linting
- ✅ Format checking
- ✅ Colorized output

---

## ⚙️ 9️⃣ — Make Targets

From **project root**, use these convenient targets:

```bash
make rust-build      # Build Rust workspace in release mode
make rust-test       # Run all Rust tests
make rust-service    # Start Rust gRPC service on port 50060
make rust-verify     # Full verification suite (build + test + lint)
```

---

## ✅ Final Success Criteria

| Stage | Confirmation |
|-------|--------------|
| Workspace loads in `cargo metadata` | ✅ |
| `cargo check` passes for both crates | ✅ |
| gRPC service starts on port 50060 | ✅ |
| Go service can call it | ✅ |
| No compiler warnings after `cargo clippy` | ✅ |
| Tests pass with `cargo test` | ✅ |

---

## 🔧 Quick Troubleshooting

### Build fails after Git pull

```bash
cd rust
cargo clean
cargo update
cargo build --release
```

### Proto changes not reflected

```bash
cd rust
rm -rf target
cargo build
```

### Service won't start

```bash
# Check what's using port 50060
lsof -i :50060
kill -9 <PID>

# Start with debug logging
RUST_LOG=debug cargo run -p kyc_dsl_service
```

### Tests failing

```bash
# Run specific test
cargo test test_name -- --nocapture

# Run tests in single-threaded mode
cargo test -- --test-threads=1

# Run with full output
cargo test -- --nocapture --test-threads=1
```

---

## 📦 Dependencies

All dependencies are managed in:
- `rust/Cargo.toml` - Workspace config
- `rust/kyc_dsl_core/Cargo.toml` - Core library deps
- `rust/kyc_dsl_service/Cargo.toml` - Service deps

**Key dependencies**:
- `nom` - Parser combinators
- `serde`, `serde_json` - Serialization
- `tonic`, `prost` - gRPC framework & Protocol Buffers
- `tokio` - Async runtime
- `tonic-reflection` - Service discovery

---

## 🚀 Next Steps After Successful Build

1. ✅ Start Rust service: `cargo run -p kyc_dsl_service`
2. ✅ Test with grpcurl: `grpcurl -plaintext localhost:50060 list`
3. ✅ Test RPC calls: See section 5️⃣ above
4. ✅ Begin validator + audit chain phase
5. ✅ Add integration tests
6. ✅ Performance benchmarking

---

## 📚 Related Documentation

- `rust/README.md` - Rust architecture overview
- `RUST_QUICKSTART.md` - 5-minute getting started guide
- `RUST_MIGRATION_REPORT.md` - Dual Go/Rust architecture details
- `RUST_SERVICE_TEST.md` - Manual testing instructions
- `CLAUDE.md` - Main project documentation

---

**Last Updated**: 2024  
**Version**: 1.5  
**Rust Workspace**: kyc_dsl_core + kyc_dsl_service