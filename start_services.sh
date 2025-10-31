#!/bin/bash
#
# start_services.sh - Start KYC-DSL services for development
#
# Usage: ./start_services.sh [--stop]
#

set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Database configuration
export DATABASE_URL="postgres://adamtc007@localhost:5432/kyc_dsl?sslmode=disable"

# Stop services
if [ "$1" == "--stop" ]; then
    echo -e "${YELLOW}🛑 Stopping services...${NC}"
    pkill -f "dataserver" 2>/dev/null && echo -e "${GREEN}✅ Data Service stopped${NC}" || echo "Data Service not running"
    pkill -f "kyc_dsl_service" 2>/dev/null && echo -e "${GREEN}✅ Rust Service stopped${NC}" || echo "Rust Service not running"
    exit 0
fi

echo -e "${GREEN}🚀 Starting KYC-DSL Services...${NC}"
echo ""

# Check if Data Service is already running
if lsof -i :50070 > /dev/null 2>&1; then
    echo -e "${YELLOW}⚠️  Data Service already running on port 50070${NC}"
else
    echo -e "${GREEN}📊 Starting Data Service (port 50070)...${NC}"
    DATABASE_URL="$DATABASE_URL" ./bin/dataserver > /tmp/dataserver.log 2>&1 &
    sleep 2
    if lsof -i :50070 > /dev/null 2>&1; then
        echo -e "${GREEN}✅ Data Service started${NC}"
    else
        echo -e "${RED}❌ Data Service failed to start. Check /tmp/dataserver.log${NC}"
        exit 1
    fi
fi

# Check if Rust Service is already running
if lsof -i :50060 > /dev/null 2>&1; then
    echo -e "${YELLOW}⚠️  Rust DSL Service already running on port 50060${NC}"
else
    echo -e "${GREEN}🦀 Starting Rust DSL Service (port 50060)...${NC}"
    cd rust && ./target/release/kyc_dsl_service > /tmp/rust_service.log 2>&1 &
    cd ..
    sleep 2
    if lsof -i :50060 > /dev/null 2>&1; then
        echo -e "${GREEN}✅ Rust DSL Service started${NC}"
    else
        echo -e "${RED}❌ Rust DSL Service failed to start. Check /tmp/rust_service.log${NC}"
        exit 1
    fi
fi

echo ""
echo -e "${GREEN}🎉 All services running!${NC}"
echo ""
echo "Services:"
echo "  • Data Service:     http://localhost:50070"
echo "  • Rust DSL Service: http://localhost:50060"
echo ""
echo "Logs:"
echo "  • Data Service:     /tmp/dataserver.log"
echo "  • Rust Service:     /tmp/rust_service.log"
echo ""
echo "Commands:"
echo "  • Stop services:    ./start_services.sh --stop"
echo "  • Test CLI:         ./kycctl get AVIVA-EU-EQUITY-FUND"
echo "  • Check status:     lsof -i :50070 && lsof -i :50060"
echo ""
