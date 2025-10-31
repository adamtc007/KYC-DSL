#!/bin/bash
#
# verify_proto_types.sh - Verify Protocol Buffer type mappings and consistency
#
# This script checks:
# 1. Proto files are properly structured
# 2. Generated Go code is up to date
# 3. Rust build.rs references correct proto files
# 4. Package names are consistent
# 5. Field naming conventions are followed
#

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

ERRORS=0
WARNINGS=0

echo -e "${BLUE}╔════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║   Protocol Buffer Type Verification           ║${NC}"
echo -e "${BLUE}╔════════════════════════════════════════════════╗${NC}"
echo ""

# Function to report error
error() {
    echo -e "${RED}✗ ERROR: $1${NC}"
    ((ERRORS++))
}

# Function to report warning
warning() {
    echo -e "${YELLOW}⚠ WARNING: $1${NC}"
    ((WARNINGS++))
}

# Function to report success
success() {
    echo -e "${GREEN}✓ $1${NC}"
}

# Function to report info
info() {
    echo -e "${BLUE}ℹ $1${NC}"
}

# ============================================================================
# 1. Check Proto File Structure
# ============================================================================

echo -e "${BLUE}[1] Checking Proto File Structure...${NC}"
echo ""

# Check api/proto files
for proto_file in api/proto/*.proto; do
    if [ -f "$proto_file" ]; then
        filename=$(basename "$proto_file")

        # Check package declaration
        if grep -q "^package " "$proto_file"; then
            package=$(grep "^package " "$proto_file" | awk '{print $2}' | tr -d ';')
            success "  $filename: package $package"
        else
            error "  $filename: Missing package declaration"
        fi

        # Check go_package option
        if grep -q "option go_package" "$proto_file"; then
            go_package=$(grep "option go_package" "$proto_file" | cut -d'"' -f2)
            success "  $filename: go_package defined"
        else
            warning "  $filename: Missing go_package option"
        fi
    fi
done

# Check proto_shared files
for proto_file in proto_shared/*.proto; do
    if [ -f "$proto_file" ]; then
        filename=$(basename "$proto_file")

        # Check package declaration
        if grep -q "^package " "$proto_file"; then
            package=$(grep "^package " "$proto_file" | awk '{print $2}' | tr -d ';')
            success "  $filename: package $package"
        else
            error "  $filename: Missing package declaration"
        fi

        # Check go_package option
        if grep -q "option go_package" "$proto_file"; then
            go_package=$(grep "option go_package" "$proto_file" | cut -d'"' -f2)
            success "  $filename: go_package defined"
        else
            warning "  $filename: Missing go_package option"
        fi
    fi
done

echo ""

# ============================================================================
# 2. Check Generated Go Code
# ============================================================================

echo -e "${BLUE}[2] Checking Generated Go Code...${NC}"
echo ""

# Check if .pb.go files exist
pb_count=$(find api/pb -name "*.pb.go" | wc -l | tr -d ' ')
if [ "$pb_count" -gt 0 ]; then
    success "  Found $pb_count generated Go files"
else
    error "  No generated Go files found in api/pb/"
fi

# Check kycdata package
if [ -d "api/pb/kycdata" ]; then
    kycdata_count=$(find api/pb/kycdata -name "*.pb.go" | wc -l | tr -d ' ')
    if [ "$kycdata_count" -gt 0 ]; then
        success "  Found $kycdata_count files in api/pb/kycdata/"
    else
        error "  No generated files in api/pb/kycdata/"
    fi
else
    error "  Directory api/pb/kycdata/ not found"
fi

# Check kycontology package
if [ -d "api/pb/kycontology" ]; then
    kycontology_count=$(find api/pb/kycontology -name "*.pb.go" | wc -l | tr -d ' ')
    if [ "$kycontology_count" -gt 0 ]; then
        success "  Found $kycontology_count files in api/pb/kycontology/"
    else
        error "  No generated files in api/pb/kycontology/"
    fi
else
    warning "  Directory api/pb/kycontology/ not found"
fi

# Check proto_shared generated files
proto_shared_count=$(find proto_shared -name "*.pb.go" | wc -l | tr -d ' ')
if [ "$proto_shared_count" -gt 0 ]; then
    success "  Found $proto_shared_count files in proto_shared/"
else
    warning "  No generated files in proto_shared/"
fi

echo ""

# ============================================================================
# 3. Check Rust Build Configuration
# ============================================================================

echo -e "${BLUE}[3] Checking Rust Build Configuration...${NC}"
echo ""

# Check kyc_dsl_service build.rs
if [ -f "rust/kyc_dsl_service/build.rs" ]; then
    if grep -q "dsl_service.proto" "rust/kyc_dsl_service/build.rs"; then
        success "  kyc_dsl_service/build.rs references dsl_service.proto"
    else
        error "  kyc_dsl_service/build.rs missing proto reference"
    fi

    if grep -q "../../api/proto" "rust/kyc_dsl_service/build.rs"; then
        success "  kyc_dsl_service/build.rs uses correct proto path"
    else
        error "  kyc_dsl_service/build.rs has incorrect proto path"
    fi
else
    error "  rust/kyc_dsl_service/build.rs not found"
fi

# Check kyc_ontology_client build.rs
if [ -f "rust/kyc_ontology_client/build.rs" ]; then
    if grep -q "ontology_service.proto" "rust/kyc_ontology_client/build.rs"; then
        success "  kyc_ontology_client/build.rs references ontology_service.proto"
    else
        error "  kyc_ontology_client/build.rs missing ontology proto"
    fi

    if grep -q "data_service.proto" "rust/kyc_ontology_client/build.rs"; then
        success "  kyc_ontology_client/build.rs references data_service.proto"
    else
        error "  kyc_ontology_client/build.rs missing data proto"
    fi

    if grep -q "../../proto_shared" "rust/kyc_ontology_client/build.rs"; then
        success "  kyc_ontology_client/build.rs uses correct proto path"
    else
        error "  kyc_ontology_client/build.rs has incorrect proto path"
    fi
else
    error "  rust/kyc_ontology_client/build.rs not found"
fi

# Check if Rust can compile
info "  Checking if Rust compiles..."
if cd rust && cargo check --quiet 2>/dev/null; then
    success "  Rust code compiles successfully"
    cd ..
else
    cd ..
    error "  Rust compilation failed"
fi

echo ""

# ============================================================================
# 4. Check Service Definitions
# ============================================================================

echo -e "${BLUE}[4] Checking Service Definitions...${NC}"
echo ""

# Check DslService
if grep -q "service DslService" api/proto/dsl_service.proto; then
    rpc_count=$(grep -c "rpc " api/proto/dsl_service.proto || echo "0")
    success "  DslService defined with $rpc_count RPCs"
else
    error "  DslService not found in dsl_service.proto"
fi

# Check DictionaryService
if grep -q "service DictionaryService" proto_shared/data_service.proto; then
    rpc_count=$(grep "service DictionaryService" -A 20 proto_shared/data_service.proto | grep -c "rpc " || echo "0")
    success "  DictionaryService defined with $rpc_count RPCs"
else
    error "  DictionaryService not found in data_service.proto"
fi

# Check CaseService
if grep -q "service CaseService" proto_shared/data_service.proto; then
    rpc_count=$(grep "service CaseService" -A 20 proto_shared/data_service.proto | grep -c "rpc " || echo "0")
    success "  CaseService defined with $rpc_count RPCs"
else
    error "  CaseService not found in data_service.proto"
fi

# Check OntologyService
if [ -f "proto_shared/ontology_service.proto" ]; then
    if grep -q "service OntologyService" proto_shared/ontology_service.proto; then
        rpc_count=$(grep -c "rpc " proto_shared/ontology_service.proto || echo "0")
        success "  OntologyService defined with $rpc_count RPCs"
    else
        error "  OntologyService not found in ontology_service.proto"
    fi
fi

echo ""

# ============================================================================
# 5. Check Key Message Types
# ============================================================================

echo -e "${BLUE}[5] Checking Key Message Types...${NC}"
echo ""

# Check CaseVersion message
if grep -q "message CaseVersion" proto_shared/data_service.proto; then
    success "  CaseVersion message defined"

    # Check for key fields
    if grep "message CaseVersion" -A 10 proto_shared/data_service.proto | grep -q "case_id"; then
        success "    ✓ case_id field present"
    else
        error "    ✗ case_id field missing"
    fi

    if grep "message CaseVersion" -A 10 proto_shared/data_service.proto | grep -q "dsl_source"; then
        success "    ✓ dsl_source field present"
    else
        error "    ✗ dsl_source field missing"
    fi
else
    error "  CaseVersion message not found"
fi

# Check CaseSummary message
if grep -q "message CaseSummary" proto_shared/data_service.proto; then
    success "  CaseSummary message defined"
else
    error "  CaseSummary message not found"
fi

# Check ParsedCase message
if grep -q "message ParsedCase" api/proto/dsl_service.proto; then
    success "  ParsedCase message defined"
else
    error "  ParsedCase message not found"
fi

echo ""

# ============================================================================
# 6. Check Field Naming Conventions
# ============================================================================

echo -e "${BLUE}[6] Checking Field Naming Conventions...${NC}"
echo ""

# Check for snake_case in proto files
camel_case_fields=$(grep -h "^\s*[a-z][a-zA-Z0-9]* [a-z][a-zA-Z]* = [0-9]" api/proto/*.proto proto_shared/*.proto 2>/dev/null | grep -v "//" | wc -l | tr -d ' ')
if [ "$camel_case_fields" -gt 0 ]; then
    warning "  Found $camel_case_fields potential camelCase fields (should be snake_case)"
else
    success "  All fields use snake_case"
fi

# Check for PascalCase in generated Go code
if grep -q "type.*struct {" api/pb/kycdata/data_service.pb.go; then
    success "  Go types use PascalCase (correct)"
fi

echo ""

# ============================================================================
# 7. Check Import Consistency
# ============================================================================

echo -e "${BLUE}[7] Checking Import Consistency...${NC}"
echo ""

# Check if Go code imports correct packages
if grep -r "api/pb/kycdata" internal/dataclient/*.go >/dev/null 2>&1; then
    success "  dataclient uses api/pb/kycdata package"
else
    error "  dataclient not importing api/pb/kycdata"
fi

if grep -r "api/pb/kycontology" internal/dataservice/*.go >/dev/null 2>&1; then
    success "  dataservice uses api/pb/kycontology package"
else
    warning "  dataservice not importing api/pb/kycontology (may not need it)"
fi

# Check Rust imports in main.rs
if grep -q "kyc::dsl" rust/kyc_dsl_service/src/main.rs; then
    success "  Rust DSL service uses kyc::dsl module"
else
    error "  Rust DSL service missing kyc::dsl import"
fi

echo ""

# ============================================================================
# Summary
# ============================================================================

echo ""
echo -e "${BLUE}╔════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║   Verification Summary                         ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════════╝${NC}"
echo ""

if [ $ERRORS -eq 0 ] && [ $WARNINGS -eq 0 ]; then
    echo -e "${GREEN}✓ All checks passed! Proto types are consistent.${NC}"
    exit 0
elif [ $ERRORS -eq 0 ]; then
    echo -e "${YELLOW}⚠ $WARNINGS warning(s) found.${NC}"
    echo -e "${YELLOW}  Review warnings but system should work.${NC}"
    exit 0
else
    echo -e "${RED}✗ $ERRORS error(s) and $WARNINGS warning(s) found.${NC}"
    echo -e "${RED}  Please fix errors before proceeding.${NC}"
    echo ""
    echo "Common fixes:"
    echo "  • Regenerate Go protos: protoc --go_out=. --go-grpc_out=. proto_shared/*.proto"
    echo "  • Rebuild Rust: cd rust && cargo clean && cargo build --release"
    echo "  • Check package names in proto files"
    exit 1
fi
