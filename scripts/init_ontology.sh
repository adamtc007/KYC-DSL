#!/bin/bash
# ===========================================================
# init_ontology.sh
# Initialize the Regulatory Data Ontology in PostgreSQL
# ===========================================================

set -e

# Database connection parameters (override with environment variables)
PGHOST=${PGHOST:-localhost}
PGPORT=${PGPORT:-5432}
PGUSER=${PGUSER:-$(whoami)}
PGDATABASE=${PGDATABASE:-kyc_dsl}

echo "🚀 Initializing Regulatory Data Ontology..."
echo "   Host: $PGHOST:$PGPORT"
echo "   Database: $PGDATABASE"
echo "   User: $PGUSER"
echo ""

# Check if psql is available
if ! command -v psql &> /dev/null; then
    echo "❌ Error: psql command not found. Please install PostgreSQL client tools."
    exit 1
fi

# Test database connection
echo "🔌 Testing database connection..."
if ! psql -h "$PGHOST" -p "$PGPORT" -U "$PGUSER" -d "$PGDATABASE" -c '\q' 2>/dev/null; then
    echo "❌ Error: Cannot connect to database."
    echo "   Please ensure PostgreSQL is running and credentials are correct."
    exit 1
fi
echo "✅ Database connection successful"
echo ""

# Run schema migration
echo "📋 Running schema migration (001_regulatory_ontology.sql)..."
psql -h "$PGHOST" -p "$PGPORT" -U "$PGUSER" -d "$PGDATABASE" \
     -f internal/storage/migrations/001_regulatory_ontology.sql

if [ $? -eq 0 ]; then
    echo "✅ Schema migration completed"
else
    echo "❌ Schema migration failed"
    exit 1
fi
echo ""

# Load seed data
echo "📦 Loading seed data (ontology_seed.sql)..."
psql -h "$PGHOST" -p "$PGPORT" -U "$PGUSER" -d "$PGDATABASE" \
     -f internal/ontology/seeds/ontology_seed.sql

if [ $? -eq 0 ]; then
    echo "✅ Seed data loaded successfully"
else
    echo "❌ Seed data loading failed"
    exit 1
fi
echo ""

# Display summary
echo "📊 Ontology Summary:"
psql -h "$PGHOST" -p "$PGPORT" -U "$PGUSER" -d "$PGDATABASE" -c "
SELECT 'Regulations' as entity, COUNT(*) as count FROM kyc_regulations
UNION ALL
SELECT 'Documents', COUNT(*) FROM kyc_documents
UNION ALL
SELECT 'Attributes', COUNT(*) FROM kyc_attributes
UNION ALL
SELECT 'Attribute-Document Links', COUNT(*) FROM kyc_attr_doc_links
UNION ALL
SELECT 'Document-Regulation Links', COUNT(*) FROM kyc_doc_reg_links
ORDER BY entity;
"

echo ""
echo "✅ Regulatory Data Ontology initialized successfully!"
echo ""
echo "Next steps:"
echo "  1. Run: ./kycctl ontology               # View ontology structure"
echo "  2. Run: ./kycctl ontology_example.dsl   # Process ontology-aware DSL"
echo "  3. Run: ./kycctl amend BLACKROCK-GLOBAL-EQUITY-FUND --step=document-discovery"
echo ""
