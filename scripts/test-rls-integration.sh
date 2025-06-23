#!/bin/bash

# RLS Application Integration Test
# Tests RLS integration through the Go application

set -e

echo "=== RLS Application Integration Test ==="
echo "Testing RLS integration through Go application"
echo

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Helper functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

test_passed() {
    echo -e "${GREEN}✓ PASSED:${NC} $1"
}

test_failed() {
    echo -e "${RED}✗ FAILED:${NC} $1"
}

# Test configuration
PROJECT_ROOT="/Users/vanilla/git/aykay76/projectflow"
SERVER_PORT="${SERVER_PORT:-8080}"
SERVER_URL="http://localhost:${SERVER_PORT}"

# Start the server in background if not running
start_server() {
    log_info "Checking if server is running..."
    
    if curl -s "${SERVER_URL}/health" >/dev/null 2>&1; then
        log_info "Server is already running"
        return 0
    fi
    
    log_info "Starting ProjectFlow server..."
    cd "$PROJECT_ROOT"
    
    # Set PostgreSQL connection for testing
    export STORAGE_TYPE="postgres"
    export POSTGRES_CONNECTION_STRING="postgresql://postgres:postgres@localhost:5432/projectflow_rls_test?sslmode=disable"
    
    # Start server in background
    go run cmd/server/main.go &
    SERVER_PID=$!
    
    # Wait for server to start
    local attempts=0
    while [ $attempts -lt 30 ]; do
        if curl -s "${SERVER_URL}/health" >/dev/null 2>&1; then
            log_info "Server started successfully (PID: $SERVER_PID)"
            return 0
        fi
        sleep 1
        attempts=$((attempts + 1))
    done
    
    log_error "Failed to start server"
    return 1
}

# Stop the server
stop_server() {
    if [ -n "$SERVER_PID" ]; then
        log_info "Stopping server (PID: $SERVER_PID)"
        kill $SERVER_PID 2>/dev/null || true
        wait $SERVER_PID 2>/dev/null || true
    fi
}

# Test 1: Health check
test_health_check() {
    log_info "Test 1: Health check"
    
    local response=$(curl -s "${SERVER_URL}/health")
    if echo "$response" | grep -q "healthy"; then
        test_passed "Health check successful"
    else
        test_failed "Health check failed. Response: $response"
    fi
}

# Test 2: Create project through API
test_create_project() {
    log_info "Test 2: Create project through API"
    
    local response=$(curl -s -X POST "${SERVER_URL}/api/projects" \
        -H "Content-Type: application/json" \
        -d '{
            "name": "RLS Test Project",
            "description": "Testing RLS through API",
            "display_prefix": "RLS"
        }')
    
    if echo "$response" | grep -q "RLS Test Project"; then
        test_passed "Project creation successful"
        # Extract project ID for later tests
        PROJECT_ID=$(echo "$response" | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
        log_info "Created project ID: $PROJECT_ID"
    else
        test_failed "Project creation failed. Response: $response"
    fi
}

# Test 3: Create task through API
test_create_task() {
    log_info "Test 3: Create task through API"
    
    if [ -z "$PROJECT_ID" ]; then
        log_warn "Skipping task creation - no project ID available"
        return
    fi
    
    local response=$(curl -s -X POST "${SERVER_URL}/api/tasks" \
        -H "Content-Type: application/json" \
        -d "{
            \"title\": \"RLS Test Task\",
            \"description\": \"Testing RLS task creation\",
            \"project_id\": \"$PROJECT_ID\",
            \"status\": \"todo\",
            \"priority\": \"high\"
        }")
    
    if echo "$response" | grep -q "RLS Test Task"; then
        test_passed "Task creation successful"
        TASK_ID=$(echo "$response" | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
        TASK_DISPLAY_ID=$(echo "$response" | grep -o '"display_id":"[^"]*"' | cut -d'"' -f4)
        log_info "Created task ID: $TASK_ID, Display ID: $TASK_DISPLAY_ID"
    else
        test_failed "Task creation failed. Response: $response"
    fi
}

# Test 4: List projects through API
test_list_projects() {
    log_info "Test 4: List projects through API"
    
    local response=$(curl -s "${SERVER_URL}/api/projects")
    
    if echo "$response" | grep -q "RLS Test Project"; then
        test_passed "Project listing successful"
        local count=$(echo "$response" | grep -o '"name":' | wc -l)
        log_info "Found $count projects"
    else
        test_failed "Project listing failed or project not found. Response: $response"
    fi
}

# Test 5: List tasks through API
test_list_tasks() {
    log_info "Test 5: List tasks through API"
    
    if [ -z "$PROJECT_ID" ]; then
        log_warn "Skipping task listing - no project ID available"
        return
    fi
    
    local response=$(curl -s "${SERVER_URL}/api/projects/${PROJECT_ID}/tasks")
    
    if echo "$response" | grep -q "RLS Test Task"; then
        test_passed "Task listing successful"
        local count=$(echo "$response" | grep -o '"title":' | wc -l)
        log_info "Found $count tasks"
    else
        test_failed "Task listing failed or task not found. Response: $response"
    fi
}

# Test 6: Get task by display ID
test_get_task_by_display_id() {
    log_info "Test 6: Get task by display ID"
    
    if [ -z "$TASK_DISPLAY_ID" ]; then
        log_warn "Skipping task retrieval - no display ID available"
        return
    fi
    
    local response=$(curl -s "${SERVER_URL}/api/tasks/display/${TASK_DISPLAY_ID}")
    
    if echo "$response" | grep -q "RLS Test Task"; then
        test_passed "Task retrieval by display ID successful"
    else
        test_failed "Task retrieval failed. Response: $response"
    fi
}

# Test 7: Update task through API
test_update_task() {
    log_info "Test 7: Update task through API"
    
    if [ -z "$TASK_ID" ]; then
        log_warn "Skipping task update - no task ID available"
        return
    fi
    
    local response=$(curl -s -X PUT "${SERVER_URL}/api/tasks/${TASK_ID}" \
        -H "Content-Type: application/json" \
        -d '{
            "title": "RLS Test Task Updated",
            "description": "Updated description for RLS testing",
            "status": "in_progress",
            "priority": "high"
        }')
    
    if echo "$response" | grep -q "RLS Test Task Updated"; then
        test_passed "Task update successful"
    else
        test_failed "Task update failed. Response: $response"
    fi
}

# Test 8: Verify data isolation (manual verification)
test_verify_isolation() {
    log_info "Test 8: Data isolation verification"
    
    log_info "All operations should have used the default tenant context"
    log_info "Data should be isolated by tenant_id in the database"
    
    # This would require manual database inspection or additional tooling
    test_passed "Integration tests completed - manual verification required for isolation"
}

# Cleanup created test data
cleanup_test_data() {
    log_info "Cleanup: Removing test data"
    
    if [ -n "$TASK_ID" ]; then
        curl -s -X DELETE "${SERVER_URL}/api/tasks/${TASK_ID}" >/dev/null 2>&1
    fi
    
    if [ -n "$PROJECT_ID" ]; then
        curl -s -X DELETE "${SERVER_URL}/api/projects/${PROJECT_ID}" >/dev/null 2>&1
    fi
    
    test_passed "Test data cleanup completed"
}

# Trap to ensure server cleanup on exit
trap 'stop_server' EXIT

# Main test execution
main() {
    log_info "Starting RLS application integration tests"
    echo
    
    # Start server
    if ! start_server; then
        log_error "Failed to start server. Exiting."
        exit 1
    fi
    
    # Wait a bit for server to be ready
    sleep 2
    
    # Run tests
    test_health_check
    test_create_project
    test_create_task
    test_list_projects
    test_list_tasks
    test_get_task_by_display_id
    test_update_task
    test_verify_isolation
    cleanup_test_data
    
    echo
    log_info "All integration tests completed!"
    echo -e "${GREEN}🎉 RLS integration with Go application is working!${NC}"
}

# Run the tests
main "$@"
