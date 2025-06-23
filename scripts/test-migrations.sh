#!/bin/bash

# Migration Test Script
# This script sets up a test database and runs the migration test suite

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
TEST_DB_NAME="projectflow_migration_test"
TEST_DB_USER="projectflow_test"
TEST_DB_PASSWORD="test_password"
TEST_DB_HOST="localhost"
TEST_DB_PORT="5432"

echo -e "${YELLOW}ProjectFlow Migration Test Suite${NC}"
echo "=================================="

# Function to check if PostgreSQL is running
check_postgres() {
    if ! command -v psql &> /dev/null; then
        echo -e "${RED}❌ PostgreSQL client (psql) not found${NC}"
        echo "Please install PostgreSQL and ensure psql is in your PATH"
        exit 1
    fi

    if ! pg_isready -h $TEST_DB_HOST -p $TEST_DB_PORT &> /dev/null; then
        echo -e "${RED}❌ PostgreSQL is not running on ${TEST_DB_HOST}:${TEST_DB_PORT}${NC}"
        echo "Please start PostgreSQL server before running tests"
        exit 1
    fi

    echo -e "${GREEN}✅ PostgreSQL is running${NC}"
}

# Function to create test database
create_test_database() {
    echo -e "${YELLOW}Setting up test database...${NC}"

    # Drop existing test database if it exists
    psql -h $TEST_DB_HOST -p $TEST_DB_PORT -U postgres -c "DROP DATABASE IF EXISTS $TEST_DB_NAME;" 2>/dev/null || true

    # Drop existing test user if it exists
    psql -h $TEST_DB_HOST -p $TEST_DB_PORT -U postgres -c "DROP USER IF EXISTS $TEST_DB_USER;" 2>/dev/null || true

    # Create test user
    psql -h $TEST_DB_HOST -p $TEST_DB_PORT -U postgres -c "CREATE USER $TEST_DB_USER WITH PASSWORD '$TEST_DB_PASSWORD';" || {
        echo -e "${RED}❌ Failed to create test user${NC}"
        echo "Make sure you have PostgreSQL superuser access"
        exit 1
    }

    # Create test database
    psql -h $TEST_DB_HOST -p $TEST_DB_PORT -U postgres -c "CREATE DATABASE $TEST_DB_NAME OWNER $TEST_DB_USER;" || {
        echo -e "${RED}❌ Failed to create test database${NC}"
        exit 1
    }

    # Grant necessary permissions
    psql -h $TEST_DB_HOST -p $TEST_DB_PORT -U postgres -c "GRANT ALL PRIVILEGES ON DATABASE $TEST_DB_NAME TO $TEST_DB_USER;"

    echo -e "${GREEN}✅ Test database created: $TEST_DB_NAME${NC}"
}

# Function to run migration tests
run_migration_tests() {
    echo -e "${YELLOW}Running migration tests...${NC}"

    # Set test database URL
    export TEST_DATABASE_URL="postgres://$TEST_DB_USER:$TEST_DB_PASSWORD@$TEST_DB_HOST:$TEST_DB_PORT/$TEST_DB_NAME?sslmode=disable"

    # Build the project first
    echo "Building project..."
    go build -o build/migrate ./cmd/migrate || {
        echo -e "${RED}❌ Failed to build migration tool${NC}"
        exit 1
    }

    # Run the migration tests
    echo "Running Go tests..."
    go test -v ./internal/migrations/tests/ || {
        echo -e "${RED}❌ Migration tests failed${NC}"
        exit 1
    }

    echo -e "${GREEN}✅ All migration tests passed${NC}"
}

# Function to test migration tool directly
test_migration_tool() {
    echo -e "${YELLOW}Testing migration tool directly...${NC}"

    # Set database environment variables
    export DB_HOST=$TEST_DB_HOST
    export DB_PORT=$TEST_DB_PORT
    export DB_NAME=$TEST_DB_NAME
    export DB_USER=$TEST_DB_USER
    export DB_PASSWORD=$TEST_DB_PASSWORD
    export DB_SSL_MODE="disable"

    # Test migration tool commands
    echo "Testing migration tool initialization..."
    ./build/migrate init || {
        echo -e "${RED}❌ Migration initialization failed${NC}"
        exit 1
    }

    echo "Testing migration status..."
    ./build/migrate status || {
        echo -e "${RED}❌ Migration status failed${NC}"
        exit 1
    }

    echo "Testing migration up..."
    ./build/migrate up || {
        echo -e "${RED}❌ Migration up failed${NC}"
        exit 1
    }

    echo "Testing migration status after up..."
    ./build/migrate status || {
        echo -e "${RED}❌ Migration status after up failed${NC}"
        exit 1
    }

    echo -e "${GREEN}✅ Migration tool tests passed${NC}"
}

# Function to cleanup
cleanup() {
    echo -e "${YELLOW}Cleaning up...${NC}"

    # Drop test database and user
    psql -h $TEST_DB_HOST -p $TEST_DB_PORT -U postgres -c "DROP DATABASE IF EXISTS $TEST_DB_NAME;" 2>/dev/null || true
    psql -h $TEST_DB_HOST -p $TEST_DB_PORT -U postgres -c "DROP USER IF EXISTS $TEST_DB_USER;" 2>/dev/null || true

    echo -e "${GREEN}✅ Cleanup completed${NC}"
}

# Function to show usage
usage() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  --setup-only    Only set up the test database, don't run tests"
    echo "  --test-only     Only run tests (assume database is set up)"
    echo "  --cleanup-only  Only cleanup test database"
    echo "  --help          Show this help message"
    echo ""
    echo "Environment variables:"
    echo "  TEST_DB_HOST     PostgreSQL host (default: localhost)"
    echo "  TEST_DB_PORT     PostgreSQL port (default: 5432)"
    echo "  TEST_DB_NAME     Test database name (default: projectflow_migration_test)"
    echo "  TEST_DB_USER     Test database user (default: projectflow_test)"
    echo "  TEST_DB_PASSWORD Test database password (default: test_password)"
}

# Parse command line arguments
SETUP_ONLY=false
TEST_ONLY=false
CLEANUP_ONLY=false

while [[ $# -gt 0 ]]; do
    case $1 in
        --setup-only)
            SETUP_ONLY=true
            shift
            ;;
        --test-only)
            TEST_ONLY=true
            shift
            ;;
        --cleanup-only)
            CLEANUP_ONLY=true
            shift
            ;;
        --help)
            usage
            exit 0
            ;;
        *)
            echo -e "${RED}❌ Unknown option: $1${NC}"
            usage
            exit 1
            ;;
    esac
done

# Override defaults with environment variables if set
TEST_DB_HOST=${TEST_DB_HOST:-$TEST_DB_HOST}
TEST_DB_PORT=${TEST_DB_PORT:-$TEST_DB_PORT}
TEST_DB_NAME=${TEST_DB_NAME:-$TEST_DB_NAME}
TEST_DB_USER=${TEST_DB_USER:-$TEST_DB_USER}
TEST_DB_PASSWORD=${TEST_DB_PASSWORD:-$TEST_DB_PASSWORD}

# Main execution flow
main() {
    if [[ "$CLEANUP_ONLY" == true ]]; then
        cleanup
        exit 0
    fi

    check_postgres

    if [[ "$TEST_ONLY" != true ]]; then
        create_test_database
    fi

    if [[ "$SETUP_ONLY" != true ]]; then
        run_migration_tests
        test_migration_tool
    fi

    if [[ "$SETUP_ONLY" != true && "$TEST_ONLY" != true ]]; then
        cleanup
    fi

    echo -e "${GREEN}🎉 Migration test suite completed successfully!${NC}"
}

# Set trap to cleanup on script exit (if not in setup-only mode)
if [[ "$SETUP_ONLY" != true ]]; then
    trap cleanup EXIT
fi

# Run main function
main
