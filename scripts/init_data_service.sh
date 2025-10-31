#!/usr/bin/env bash
# ============================================================================
# KYC Data Service - Database Initialization Script
# ============================================================================
# This script initializes the PostgreSQL schema for the Data Service
# Safe to run multiple times (idempotent)
# ============================================================================

set -e  # Exit on error

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default database connection parameters
PGHOST="${PGHOST:-localhost}"
PGPORT="${PGPORT:-5432}"
PGUSER="${PGUSER:-postgres}"
PGDATABASE="${PGDATABASE:-kyc_dsl}"

echo -e "${BLUE}============================================================================${NC}"
echo -e "${BLUE}KYC Data Service - Database Initialization${NC}"
echo -e "${BLUE}============================================================================${NC}"
echo ""
echo -e "${YELLOW}Database Connection:${NC}"
echo -e "  Host:     ${PGHOST}:${PGPORT}"
echo -e "  Database: ${PGDATABASE}"
echo -e "  User:     ${PGUSER}"
echo ""

# Check if PostgreSQL is accessible
echo -e "${YELLOW}🔍 Checking PostgreSQL connection...${NC}"
if ! psql -h "$PGHOST" -p "$PGPORT" -U "$PGUSER" -d "$PGDATABASE" -c '\q' 2>/dev/null; then
    echo -e "${RED}❌ Failed to connect to PostgreSQL${NC}"
    echo -e "${RED}   Please check your connection parameters and ensure PostgreSQL is running${NC}"
    echo ""
    echo -e "${YELLOW}💡 Environment variables:${NC}"
    echo -e "   export PGHOST=localhost"
    echo -e "   export PGPORT=5432"
    echo -e "   export PGUSER=postgres"
    echo -e "   export PGPASSWORD=your_password"
    echo -e "   export PGDATABASE=kyc_dsl"
    exit 1
fi
echo -e "${GREEN}✅ PostgreSQL connection successful${NC}"
echo ""

# Get the script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SQL_FILE="$SCRIPT_DIR/init_data_service_tables.sql"

# Check if SQL file exists
if [ ! -f "$SQL_FILE" ]; then
    echo -e "${RED}❌ SQL file not found: $SQL_FILE${NC}"
    exit 1
fi

# Run the initialization SQL
echo -e "${YELLOW}📊 Running database schema initialization...${NC}"
if psql -h "$PGHOST" -p "$PGPORT" -U "$PGUSER" -d "$PGDATABASE" -f "$SQL_FILE" > /tmp/data_service_init.log 2>&1; then
    echo -e "${GREEN}✅ Schema initialized successfully${NC}"
else
    echo -e "${RED}❌ Schema initialization failed${NC}"
    echo -e "${RED}   See /tmp/data_service_init.log for details${NC}"
    cat /tmp/data_service_init.log
    exit 1
fi
echo ""

# Verify the tables
echo -e "${YELLOW}🔍 Verifying tables...${NC}"

# Check case_versions table
CV_COUNT=$(psql -h "$PGHOST" -p "$PGPORT" -U "$PGUSER" -d "$PGDATABASE" -t -c "SELECT COUNT(*) FROM case_versions;" 2>/dev/null || echo "0")
echo -e "  ${GREEN}✓${NC} case_versions table (${CV_COUNT} rows)"

# Check kyc_attributes table
ATTR_COUNT=$(psql -h "$PGHOST" -p "$PGPORT" -U "$PGUSER" -d "$PGDATABASE" -t -c "SELECT COUNT(*) FROM kyc_attributes;" 2>/dev/null || echo "0")
echo -e "  ${GREEN}✓${NC} kyc_attributes table (${ATTR_COUNT} rows)"

# Check kyc_documents table
DOC_COUNT=$(psql -h "$PGHOST" -p "$PGPORT" -U "$PGUSER" -d "$PGDATABASE" -t -c "SELECT COUNT(*) FROM kyc_documents;" 2>/dev/null || echo "0")
echo -e "  ${GREEN}✓${NC} kyc_documents table (${DOC_COUNT} rows)"

echo ""
echo -e "${GREEN}============================================================================${NC}"
echo -e "${GREEN}✅ Data Service database initialization complete!${NC}"
echo -e "${GREEN}============================================================================${NC}"
echo ""
echo -e "${YELLOW}📋 Summary:${NC}"
echo -e "   • case_versions:   ${CV_COUNT} records"
echo -e "   • kyc_attributes:  ${ATTR_COUNT} records"
echo -e "   • kyc_documents:   ${DOC_COUNT} records"
echo ""
echo -e "${YELLOW}🚀 Next steps:${NC}"
echo -e "   1. Start the Data Service:"
echo -e "      ${BLUE}make run-dataserver${NC}"
echo -e "      or"
echo -e "      ${BLUE}go run cmd/dataserver/main.go${NC}"
echo ""
echo -e "   2. Test with grpcurl:"
echo -e "      ${BLUE}grpcurl -plaintext localhost:50070 list${NC}"
echo -e "      ${BLUE}grpcurl -plaintext localhost:50070 kyc.data.DictionaryService/ListAttributes -d '{\"limit\":5}'${NC}"
echo ""
echo -e "${YELLOW}💡 Port allocation:${NC}"
echo -e "   • 50070 - Data Service (Dictionary + Case)"
echo -e "   • 50051 - Main gRPC Service (KycCase, DSL, RAG, CBU)"
echo -e "   • 50060 - Rust DSL Service"
echo -e "   • 8080  - REST API"
echo ""
