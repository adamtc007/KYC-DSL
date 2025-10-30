#!/bin/bash

# ===========================================================
# migrate_feedback.sh
# Apply RAG Feedback Loop Migration (007_rag_feedback.sql)
# ===========================================================

set -e  # Exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
DB_NAME="${PGDATABASE:-kyc_dsl}"
DB_HOST="${PGHOST:-localhost}"
DB_PORT="${PGPORT:-5432}"
DB_USER="${PGUSER:-$USER}"
MIGRATION_FILE="internal/storage/migrations/007_rag_feedback.sql"

# Helper functions
print_header() {
    echo -e "\n${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}\n"
}

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

print_info() {
    echo -e "${YELLOW}ℹ $1${NC}"
}

# Check prerequisites
check_prerequisites() {
    print_header "🔍 Checking Prerequisites"

    # Check if psql is installed
    if ! command -v psql &> /dev/null; then
        print_error "psql is not installed"
        echo "Please install PostgreSQL client tools:"
        echo "  macOS:   brew install postgresql"
        echo "  Ubuntu:  sudo apt install postgresql-client"
        exit 1
    fi
    print_success "psql is installed"

    # Check if migration file exists
    if [ ! -f "$MIGRATION_FILE" ]; then
        print_error "Migration file not found: $MIGRATION_FILE"
        echo "Please ensure you're running this script from the KYC-DSL project root"
        exit 1
    fi
    print_success "Migration file found: $MIGRATION_FILE"

    # Check database connection
    if ! psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c '\q' 2>/dev/null; then
        print_error "Cannot connect to database: $DB_NAME"
        echo ""
        echo "Connection details:"
        echo "  Host:     $DB_HOST"
        echo "  Port:     $DB_PORT"
        echo "  Database: $DB_NAME"
        echo "  User:     $DB_USER"
        echo ""
        echo "Please check your database configuration and ensure:"
        echo "  1. PostgreSQL is running"
        echo "  2. Database '$DB_NAME' exists"
        echo "  3. Connection credentials are correct"
        exit 1
    fi
    print_success "Database connection successful: $DB_NAME"
}

# Check if migration already applied
check_if_applied() {
    print_header "📋 Checking Migration Status"

    # Check if rag_feedback table exists
    if psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" \
        -t -c "SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name='rag_feedback');" \
        2>/dev/null | grep -q 't'; then

        print_info "Migration appears to be already applied"
        echo ""
        read -p "Do you want to re-apply the migration? This will DROP and recreate tables. (y/N): " -n 1 -r
        echo ""
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            print_info "Migration cancelled by user"
            exit 0
        fi
    else
        print_success "Migration not yet applied"
    fi
}

# Apply migration
apply_migration() {
    print_header "🚀 Applying Migration"

    print_info "Applying: $MIGRATION_FILE"

    if psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -f "$MIGRATION_FILE" 2>&1; then
        print_success "Migration applied successfully"
    else
        print_error "Migration failed"
        exit 1
    fi
}

# Verify migration
verify_migration() {
    print_header "✅ Verifying Migration"

    local errors=0

    # Check table exists
    if psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" \
        -t -c "SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name='rag_feedback');" \
        2>/dev/null | grep -q 't'; then
        print_success "Table 'rag_feedback' exists"
    else
        print_error "Table 'rag_feedback' not found"
        errors=$((errors + 1))
    fi

    # Check enum type exists
    if psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" \
        -t -c "SELECT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'feedback_sentiment');" \
        2>/dev/null | grep -q 't'; then
        print_success "Enum type 'feedback_sentiment' exists"
    else
        print_error "Enum type 'feedback_sentiment' not found"
        errors=$((errors + 1))
    fi

    # Check trigger exists
    if psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" \
        -t -c "SELECT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'trig_feedback_relevance');" \
        2>/dev/null | grep -q 't'; then
        print_success "Trigger 'trig_feedback_relevance' exists"
    else
        print_error "Trigger 'trig_feedback_relevance' not found"
        errors=$((errors + 1))
    fi

    # Check function exists
    if psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" \
        -t -c "SELECT EXISTS (SELECT 1 FROM pg_proc WHERE proname = 'update_relevance');" \
        2>/dev/null | grep -q 't'; then
        print_success "Function 'update_relevance()' exists"
    else
        print_error "Function 'update_relevance()' not found"
        errors=$((errors + 1))
    fi

    # Check views exist
    if psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" \
        -t -c "SELECT EXISTS (SELECT FROM information_schema.views WHERE table_name='rag_feedback_summary');" \
        2>/dev/null | grep -q 't'; then
        print_success "View 'rag_feedback_summary' exists"
    else
        print_error "View 'rag_feedback_summary' not found"
        errors=$((errors + 1))
    fi

    if psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" \
        -t -c "SELECT EXISTS (SELECT FROM information_schema.views WHERE table_name='attribute_feedback_summary');" \
        2>/dev/null | grep -q 't'; then
        print_success "View 'attribute_feedback_summary' exists"
    else
        print_error "View 'attribute_feedback_summary' not found"
        errors=$((errors + 1))
    fi

    # Count indexes
    local index_count=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" \
        -t -c "SELECT COUNT(*) FROM pg_indexes WHERE tablename = 'rag_feedback';" 2>/dev/null | tr -d ' ')

    if [ "$index_count" -ge 7 ]; then
        print_success "Indexes created: $index_count"
    else
        print_error "Expected at least 7 indexes, found: $index_count"
        errors=$((errors + 1))
    fi

    echo ""
    if [ $errors -eq 0 ]; then
        print_success "All verification checks passed!"
        return 0
    else
        print_error "Verification failed with $errors error(s)"
        return 1
    fi
}

# Show schema info
show_schema_info() {
    print_header "📊 Schema Information"

    echo "Table Structure:"
    psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" \
        -c "\d rag_feedback" 2>/dev/null || true

    echo ""
    echo "Indexes:"
    psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" \
        -c "\di+ *rag_feedback*" 2>/dev/null || true

    echo ""
    echo "Trigger:"
    psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" \
        -c "SELECT tgname, tgenabled, tgtype FROM pg_trigger WHERE tgrelid = 'rag_feedback'::regclass;" \
        2>/dev/null || true
}

# Main execution
main() {
    print_header "🚀 RAG Feedback Loop Migration"
    echo "Database: $DB_NAME @ $DB_HOST:$DB_PORT"
    echo "User: $DB_USER"
    echo ""

    check_prerequisites
    check_if_applied
    apply_migration

    if verify_migration; then
        show_schema_info

        print_header "✅ Migration Complete"
        echo "The RAG Feedback Loop is now ready to use!"
        echo ""
        echo "Next steps:"
        echo "  1. Start the server: go run cmd/kycserver/main.go"
        echo "  2. Run tests: ./scripts/test_feedback.sh"
        echo "  3. Submit feedback: curl -X POST http://localhost:8080/rag/feedback ..."
        echo ""
        echo "Documentation:"
        echo "  • RAG_FEEDBACK.md - Complete guide"
        echo "  • API endpoints available at http://localhost:8080/"
        echo ""
    else
        print_error "Migration verification failed"
        echo "Please check the error messages above and try again"
        exit 1
    fi
}

# Run main function
main
