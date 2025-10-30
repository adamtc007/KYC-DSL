# Service Dependencies & Preflight Checklist

This document lists all external dependencies and services required to run the KYC-DSL Rust components.

## 📋 Quick Preflight Checklist

Before running any services, verify all dependencies:

```bash
# Run this script to check all dependencies
./rust/preflight.sh
```

Or manually check each item below.

---

## 🔧 System Prerequisites

### 1. Rust Toolchain

**Required Version**: 1.70+

**Check Installation:**
```bash
rustc --version    # Should show 1.70 or higher
cargo --version
```

**Install:**
```bash
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
source $HOME/.cargo/env
```

**Install Components:**
```bash
rustup component add rustfmt clippy
```

---

### 2. Go Toolchain

**Required Version**: 1.21+

**Check Installation:**
```bash
go version    # Should show 1.21 or higher
```

**Install (macOS):**
```bash
brew install go
```

**Install (Linux):**
```bash
wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
```

---

### 3. Protocol Buffers Compiler (protoc)

**Required Version**: 3.0+

**Check Installation:**
```bash
protoc --version    # Should show libprotoc 3.x or higher
```

**Install (macOS):**
```bash
brew install protobuf
```

**Install (Ubuntu/Debian):**
```bash
sudo apt update
sudo apt install -y protobuf-compiler
```

**Install (From Source):**
```bash
# Download from https://github.com/protocolbuffers/protobuf/releases
curl -LO https://github.com/protocolbuffers/protobuf/releases/download/v24.4/protoc-24.4-linux-x86_64.zip
unzip protoc-24.4-linux-x86_64.zip -d $HOME/.local
export PATH="$PATH:$HOME/.local/bin"
```

---

### 4. PostgreSQL Database

**Required Version**: 12+

**Check Installation:**
```bash
psql --version     # Should show PostgreSQL 12 or higher
```

**Check if Running:**
```bash
# macOS
brew services list | grep postgresql

# Linux (systemd)
systemctl status postgresql

# Check connection
psql -U $USER -d postgres -c "SELECT version();"
```

**Install (macOS):**
```bash
brew install postgresql@15
brew services start postgresql@15
```

**Install (Ubuntu/Debian):**
```bash
sudo apt update
sudo apt install postgresql postgresql-contrib
sudo systemctl start postgresql
sudo systemctl enable postgresql
```

**Create Database:**
```bash
# Create database if it doesn't exist
createdb kyc_dsl

# Or with psql
psql -U postgres -c "CREATE DATABASE kyc_dsl;"
```

**Install pgvector Extension (for RAG features):**
```bash
# macOS
brew install pgvector

# Ubuntu
sudo apt install postgresql-15-pgvector

# Enable in database
psql -d kyc_dsl -c "CREATE EXTENSION IF NOT EXISTS vector;"
```

---

### 5. grpcurl (Testing Tool)

**Optional but Recommended**

**Check Installation:**
```bash
grpcurl --version
```

**Install (macOS):**
```bash
brew install grpcurl
```

**Install (Linux):**
```bash
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest
```

---

## 🌍 Environment Variables

### Required

```bash
# Database connection (if not using defaults)
export PGHOST=localhost
export PGPORT=5432
export PGUSER=$USER
export PGDATABASE=kyc_dsl
export PGPASSWORD=your_password  # Optional if using peer auth

# OpenAI (for RAG/Vector features only)
export OPENAI_API_KEY=sk-...  # Only needed for RAG features
```

### Optional

```bash
# Rust logging
export RUST_LOG=info          # debug, info, warn, error
export RUST_BACKTRACE=1       # Enable backtraces on panic

# Go experiments
export GOEXPERIMENT=greenteagc
```

**Set Permanently (add to ~/.bashrc or ~/.zshrc):**
```bash
echo 'export PGDATABASE=kyc_dsl' >> ~/.zshrc
echo 'export RUST_LOG=info' >> ~/.zshrc
source ~/.zshrc
```

---

## 🚀 Service Startup Order

### 1. PostgreSQL (Must be running first)

```bash
# macOS
brew services start postgresql@15

# Linux
sudo systemctl start postgresql

# Verify
psql -d kyc_dsl -c "SELECT 1;"
```

**Expected Output:**
```
 ?column?
----------
        1
(1 row)
```

---

### 2. Rust gRPC Service (Port 50060)

```bash
# Terminal 1
make run-rust

# Or directly
cd rust/kyc_dsl_service && cargo run
```

**Expected Output:**
```
🦀 Rust DSL gRPC Service
========================
Listening on: [::1]:50060
Ready to accept connections...
```

**Verify:**
```bash
# Check port is listening
lsof -i :50060

# Test with grpcurl
grpcurl -plaintext localhost:50060 list
```

---

### 3. Go gRPC Server (Port 50051) - Optional

```bash
# Terminal 2
make run-grpc

# Or directly
./bin/grpcserver
```

**Expected Output:**
```
Starting gRPC server on port 50051...
```

**Verify:**
```bash
lsof -i :50051
grpcurl -plaintext localhost:50051 list
```

---

### 4. Go REST API Server (Port 8080) - Optional

```bash
# Terminal 3
make run-server

# Or directly
./bin/kycserver
```

**Expected Output:**
```
Starting RAG API server on port 8080...
```

**Verify:**
```bash
curl http://localhost:8080/health
```

---

## ✅ Preflight Script

Create `rust/preflight.sh`:

```bash
#!/bin/bash
# Preflight checks for KYC-DSL Rust services

set -e

echo "🔍 KYC-DSL Preflight Checks"
echo "============================"
echo

FAIL=0

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

check() {
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓${NC} $1"
    else
        echo -e "${RED}✗${NC} $1"
        FAIL=$((FAIL + 1))
    fi
}

# 1. Check Rust
echo "Checking Rust..."
rustc --version > /dev/null 2>&1
check "Rust installed: $(rustc --version 2>/dev/null || echo 'NOT FOUND')"

# 2. Check Go
echo "Checking Go..."
go version > /dev/null 2>&1
check "Go installed: $(go version 2>/dev/null || echo 'NOT FOUND')"

# 3. Check protoc
echo "Checking Protocol Buffers..."
protoc --version > /dev/null 2>&1
check "protoc installed: $(protoc --version 2>/dev/null || echo 'NOT FOUND')"

# 4. Check PostgreSQL
echo "Checking PostgreSQL..."
psql --version > /dev/null 2>&1
check "PostgreSQL installed: $(psql --version 2>/dev/null || echo 'NOT FOUND')"

# Check if PostgreSQL is running
if psql -d postgres -c "SELECT 1;" > /dev/null 2>&1; then
    check "PostgreSQL is running"
else
    echo -e "${RED}✗${NC} PostgreSQL is NOT running"
    echo "  Start with: brew services start postgresql@15 (macOS)"
    echo "           or: sudo systemctl start postgresql (Linux)"
    FAIL=$((FAIL + 1))
fi

# Check if kyc_dsl database exists
if psql -lqt | cut -d \| -f 1 | grep -qw kyc_dsl; then
    check "Database 'kyc_dsl' exists"
else
    echo -e "${YELLOW}⚠${NC} Database 'kyc_dsl' does not exist"
    echo "  Create with: createdb kyc_dsl"
fi

# 5. Check grpcurl (optional)
if command -v grpcurl > /dev/null 2>&1; then
    check "grpcurl installed (optional)"
else
    echo -e "${YELLOW}⚠${NC} grpcurl not installed (optional, but recommended for testing)"
fi

# 6. Check ports
echo
echo "Checking ports..."
if lsof -i :50060 > /dev/null 2>&1; then
    echo -e "${YELLOW}⚠${NC} Port 50060 is in use (Rust service may already be running)"
else
    check "Port 50060 is available"
fi

if lsof -i :50051 > /dev/null 2>&1; then
    echo -e "${YELLOW}⚠${NC} Port 50051 is in use (Go gRPC service may already be running)"
else
    check "Port 50051 is available"
fi

if lsof -i :8080 > /dev/null 2>&1; then
    echo -e "${YELLOW}⚠${NC} Port 8080 is in use (Go API may already be running)"
else
    check "Port 8080 is available"
fi

# 7. Check environment variables
echo
echo "Checking environment..."
if [ -z "$PGDATABASE" ]; then
    echo -e "${YELLOW}⚠${NC} PGDATABASE not set (will use default)"
else
    check "PGDATABASE=$PGDATABASE"
fi

if [ -z "$OPENAI_API_KEY" ]; then
    echo -e "${YELLOW}⚠${NC} OPENAI_API_KEY not set (RAG features will not work)"
else
    check "OPENAI_API_KEY is set"
fi

# 8. Check Rust components
echo
echo "Checking Rust components..."
if rustup component list | grep -q "rustfmt-.*installed"; then
    check "rustfmt installed"
else
    echo -e "${YELLOW}⚠${NC} rustfmt not installed"
    echo "  Install with: rustup component add rustfmt"
fi

if rustup component list | grep -q "clippy-.*installed"; then
    check "clippy installed"
else
    echo -e "${YELLOW}⚠${NC} clippy not installed"
    echo "  Install with: rustup component add clippy"
fi

# Summary
echo
echo "============================"
if [ $FAIL -eq 0 ]; then
    echo -e "${GREEN}✓ All critical checks passed!${NC}"
    echo
    echo "You can now run:"
    echo "  make rust-build    # Build Rust components"
    echo "  make rust-test     # Run tests"
    echo "  make run-rust      # Start Rust gRPC service"
    exit 0
else
    echo -e "${RED}✗ $FAIL check(s) failed${NC}"
    echo
    echo "Please fix the issues above and run again."
    exit 1
fi
```

**Make it executable:**
```bash
chmod +x rust/preflight.sh
```

**Run it:**
```bash
./rust/preflight.sh
```

---

## 🐛 Troubleshooting

### PostgreSQL Not Running

**macOS:**
```bash
brew services start postgresql@15
brew services list
```

**Linux:**
```bash
sudo systemctl start postgresql
sudo systemctl status postgresql
```

**Docker Alternative:**
```bash
docker run --name kyc-postgres \
  -e POSTGRES_DB=kyc_dsl \
  -e POSTGRES_PASSWORD=postgres \
  -p 5432:5432 \
  -d postgres:15
```

---

### Port Already in Use

**Find and kill process:**
```bash
# Find process using port 50060
lsof -ti:50060

# Kill it
lsof -ti:50060 | xargs kill -9

# Or for other ports
lsof -ti:50051 | xargs kill -9
lsof -ti:8080 | xargs kill -9
```

---

### Cannot Connect to PostgreSQL

**Check connection:**
```bash
# Test connection
psql -h localhost -U $USER -d kyc_dsl

# Check pg_hba.conf allows connections
# macOS: /opt/homebrew/var/postgresql@15/pg_hba.conf
# Linux: /etc/postgresql/15/main/pg_hba.conf
```

**Common fix - Allow local connections:**
```bash
# Add to pg_hba.conf
local   all   all   trust
host    all   all   127.0.0.1/32   trust
host    all   all   ::1/128        trust

# Reload PostgreSQL
# macOS:
brew services restart postgresql@15

# Linux:
sudo systemctl restart postgresql
```

---

### Proto Generation Fails

```bash
# Ensure protoc is in PATH
which protoc

# Regenerate proto files
make proto

# Rebuild Rust (will regenerate protos)
cd rust && cargo clean && cargo build
```

---

## 📊 Service Health Checks

### Quick Health Check Script

```bash
#!/bin/bash
# Check all services are healthy

echo "🏥 Service Health Check"
echo "======================="

# Rust gRPC (port 50060)
if grpcurl -plaintext localhost:50060 list > /dev/null 2>&1; then
    echo "✓ Rust gRPC Service (50060) - HEALTHY"
else
    echo "✗ Rust gRPC Service (50060) - DOWN"
fi

# Go gRPC (port 50051)
if grpcurl -plaintext localhost:50051 list > /dev/null 2>&1; then
    echo "✓ Go gRPC Service (50051) - HEALTHY"
else
    echo "✗ Go gRPC Service (50051) - DOWN"
fi

# Go REST API (port 8080)
if curl -s http://localhost:8080/health > /dev/null 2>&1; then
    echo "✓ Go REST API (8080) - HEALTHY"
else
    echo "✗ Go REST API (8080) - DOWN"
fi

# PostgreSQL
if psql -d kyc_dsl -c "SELECT 1;" > /dev/null 2>&1; then
    echo "✓ PostgreSQL - HEALTHY"
else
    echo "✗ PostgreSQL - DOWN"
fi
```

---

## 🎯 Minimal Setup (Development)

For basic Rust development and testing:

```bash
# 1. Install Rust
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
rustup component add rustfmt clippy

# 2. Install protoc
brew install protobuf  # or apt install protobuf-compiler

# 3. Build and test
cd rust
cargo build
cargo test

# 4. Run service (no database needed for basic parsing)
cargo run --bin kyc_dsl_service
```

---

## 📦 Full Stack Setup (Production)

For complete system with database integration:

```bash
# 1. All prerequisites
./rust/preflight.sh

# 2. Initialize database
createdb kyc_dsl
psql -d kyc_dsl -c "CREATE EXTENSION vector;"
./scripts/init_ontology.sh

# 3. Start services in order
# Terminal 1: PostgreSQL (already running)
# Terminal 2: Rust gRPC
make run-rust

# Terminal 3: Go gRPC
make run-grpc

# Terminal 4: Go REST API
make run-server

# 4. Verify all services
./rust/health-check.sh
```

---

**Last Updated**: 2024-10-30  
**Version**: 1.0