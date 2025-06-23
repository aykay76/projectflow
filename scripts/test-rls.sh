#!/bin/bash

# RLS Test Script for ProjectFlow
# This script tests Row-Level Security implementation with multiple tenant scenarios

set -e

echo "=== ProjectFlow RLS Comprehensive Test ==="
echo "Testing Row-Level Security policies with multiple tenant scenarios"
echo

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test configuration
TEST_DB_HOST="${TEST_DB_HOST:-localhost}"
TEST_DB_PORT="${TEST_DB_PORT:-5432}"
TEST_DB_NAME="${TEST_DB_NAME:-projectflow_test}"
TEST_DB_USER="${TEST_DB_USER:-projectflow_user}"
TEST_DB_PASS="${TEST_DB_PASS:-projectflow_pass}"

# Database connection string
DB_CONN="postgresql://${TEST_DB_USER}:${TEST_DB_PASS}@${TEST_DB_HOST}:${TEST_DB_PORT}/${TEST_DB_NAME}?sslmode=disable"

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
    exit 1
}

# Test 1: Verify RLS is enabled on all tables
test_rls_enabled() {
    log_info "Test 1: Verifying RLS is enabled on all tenant-aware tables"
    
    local result=$(psql "$DB_CONN" -t -c "
        SELECT 
            schemaname,
            tablename,
            rowsecurity
        FROM pg_tables 
        WHERE schemaname = 'public' 
        AND tablename IN ('tenants', 'projects', 'tasks', 'users')
        AND rowsecurity = true;
    " | wc -l | tr -d ' ')
    
    if [ "$result" -eq 4 ]; then
        test_passed "RLS is enabled on all 4 tenant-aware tables"
    else
        test_failed "RLS is not enabled on all required tables. Found $result/4 tables with RLS enabled"
    fi
}

# Test 2: Verify RLS policies exist
test_rls_policies_exist() {
    log_info "Test 2: Verifying RLS policies exist for all tables"
    
    local result=$(psql "$DB_CONN" -t -c "
        SELECT COUNT(*) 
        FROM pg_policies 
        WHERE schemaname = 'public' 
        AND tablename IN ('tenants', 'projects', 'tasks', 'users');
    " | tr -d ' ')
    
    if [ "$result" -eq 4 ]; then
        test_passed "All 4 RLS policies exist"
    else
        test_failed "Not all RLS policies exist. Found $result/4 policies"
    fi
}

# Test 3: Create test tenants and data
setup_test_data() {
    log_info "Test 3: Setting up test data with multiple tenants"
    
    psql "$DB_CONN" -c "
        -- Create test tenants
        INSERT INTO tenants (id, name, status) VALUES 
            ('tenant-alpha', 'Alpha Corp', 'active'),
            ('tenant-beta', 'Beta Inc', 'active'),
            ('tenant-gamma', 'Gamma LLC', 'active')
        ON CONFLICT (id) DO NOTHING;
        
        -- Create test projects for each tenant
        INSERT INTO projects (id, name, description, display_prefix, tenant_id, created_at, updated_at) VALUES
            ('proj-alpha-1', 'Alpha Project 1', 'First project for Alpha', 'ALPHA', 'tenant-alpha', NOW(), NOW()),
            ('proj-alpha-2', 'Alpha Project 2', 'Second project for Alpha', 'ALPHA2', 'tenant-alpha', NOW(), NOW()),
            ('proj-beta-1', 'Beta Project 1', 'First project for Beta', 'BETA', 'tenant-beta', NOW(), NOW()),
            ('proj-gamma-1', 'Gamma Project 1', 'First project for Gamma', 'GAMMA', 'tenant-gamma', NOW(), NOW())
        ON CONFLICT (id) DO NOTHING;
        
        -- Create test tasks for each tenant
        INSERT INTO tasks (id, display_id, project_id, title, status, priority, type, tenant_id, created_at, updated_at) VALUES
            ('task-alpha-1', 'ALPHA-1', 'proj-alpha-1', 'Alpha Task 1', 'todo', 'high', 'task', 'tenant-alpha', NOW(), NOW()),
            ('task-alpha-2', 'ALPHA-2', 'proj-alpha-1', 'Alpha Task 2', 'in_progress', 'medium', 'task', 'tenant-alpha', NOW(), NOW()),
            ('task-beta-1', 'BETA-1', 'proj-beta-1', 'Beta Task 1', 'todo', 'low', 'task', 'tenant-beta', NOW(), NOW()),
            ('task-gamma-1', 'GAMMA-1', 'proj-gamma-1', 'Gamma Task 1', 'done', 'critical', 'task', 'tenant-gamma', NOW(), NOW())
        ON CONFLICT (id) DO NOTHING;
        
        -- Create test users for each tenant
        INSERT INTO users (id, tenant_id, username, email, password_hash, role, created_at, updated_at) VALUES
            ('user-alpha-1', 'tenant-alpha', 'alice', 'alice@alpha.com', 'hashed_password', 'admin', NOW(), NOW()),
            ('user-beta-1', 'tenant-beta', 'bob', 'bob@beta.com', 'hashed_password', 'user', NOW(), NOW()),
            ('user-gamma-1', 'tenant-gamma', 'charlie', 'charlie@gamma.com', 'hashed_password', 'user', NOW(), NOW())
        ON CONFLICT (id) DO NOTHING;
    " >/dev/null 2>&1
    
    test_passed "Test data created successfully"
}

# Test 4: Test tenant isolation - users should only see their own tenant's data
test_tenant_isolation() {
    log_info "Test 4: Testing tenant isolation"
    
    # Test Alpha tenant can only see Alpha data
    local alpha_projects=$(psql "$DB_CONN" -t -c "
        SELECT init_tenant_context('tenant-alpha', false);
        SELECT COUNT(*) FROM projects;
    " | tail -1 | tr -d ' ')
    
    local alpha_tasks=$(psql "$DB_CONN" -t -c "
        SELECT init_tenant_context('tenant-alpha', false);
        SELECT COUNT(*) FROM tasks;
    " | tail -1 | tr -d ' ')
    
    if [ "$alpha_projects" -eq 2 ] && [ "$alpha_tasks" -eq 2 ]; then
        test_passed "Alpha tenant isolation works correctly (2 projects, 2 tasks)"
    else
        test_failed "Alpha tenant isolation failed. Found $alpha_projects projects, $alpha_tasks tasks (expected 2, 2)"
    fi
    
    # Test Beta tenant can only see Beta data
    local beta_projects=$(psql "$DB_CONN" -t -c "
        SELECT init_tenant_context('tenant-beta', false);
        SELECT COUNT(*) FROM projects;
    " | tail -1 | tr -d ' ')
    
    local beta_tasks=$(psql "$DB_CONN" -t -c "
        SELECT init_tenant_context('tenant-beta', false);
        SELECT COUNT(*) FROM tasks;
    " | tail -1 | tr -d ' ')
    
    if [ "$beta_projects" -eq 1 ] && [ "$beta_tasks" -eq 1 ]; then
        test_passed "Beta tenant isolation works correctly (1 project, 1 task)"
    else
        test_failed "Beta tenant isolation failed. Found $beta_projects projects, $beta_tasks tasks (expected 1, 1)"
    fi
}

# Test 5: Test admin bypass - admin should see all data
test_admin_bypass() {
    log_info "Test 5: Testing admin bypass functionality"
    
    local admin_projects=$(psql "$DB_CONN" -t -c "
        SELECT init_tenant_context('tenant-alpha', true);
        SELECT COUNT(*) FROM projects;
    " | tail -1 | tr -d ' ')
    
    local admin_tasks=$(psql "$DB_CONN" -t -c "
        SELECT init_tenant_context('tenant-alpha', true);
        SELECT COUNT(*) FROM tasks;
    " | tail -1 | tr -d ' ')
    
    # Admin should see all projects and tasks from all tenants (plus default tenant if it exists)
    if [ "$admin_projects" -ge 4 ] && [ "$admin_tasks" -ge 4 ]; then
        test_passed "Admin bypass works correctly ($admin_projects projects, $admin_tasks tasks visible)"
    else
        test_failed "Admin bypass failed. Found $admin_projects projects, $admin_tasks tasks (expected >= 4 each)"
    fi
}

# Test 6: Test INSERT operations respect tenant isolation
test_insert_isolation() {
    log_info "Test 6: Testing INSERT operations respect tenant isolation"
    
    # Try to insert as Alpha tenant
    psql "$DB_CONN" -c "
        SELECT init_tenant_context('tenant-alpha', false);
        INSERT INTO projects (id, name, description, display_prefix, tenant_id, created_at, updated_at) 
        VALUES ('test-proj-alpha', 'Test Project Alpha', 'Test project', 'TEST', 'tenant-alpha', NOW(), NOW());
    " >/dev/null 2>&1
    
    # Verify Alpha can see their new project
    local alpha_can_see=$(psql "$DB_CONN" -t -c "
        SELECT init_tenant_context('tenant-alpha', false);
        SELECT COUNT(*) FROM projects WHERE id = 'test-proj-alpha';
    " | tail -1 | tr -d ' ')
    
    # Verify Beta cannot see Alpha's new project
    local beta_cannot_see=$(psql "$DB_CONN" -t -c "
        SELECT init_tenant_context('tenant-beta', false);
        SELECT COUNT(*) FROM projects WHERE id = 'test-proj-alpha';
    " | tail -1 | tr -d ' ')
    
    if [ "$alpha_can_see" -eq 1 ] && [ "$beta_cannot_see" -eq 0 ]; then
        test_passed "INSERT isolation works correctly"
    else
        test_failed "INSERT isolation failed. Alpha sees: $alpha_can_see, Beta sees: $beta_cannot_see"
    fi
}

# Test 7: Test UPDATE operations respect tenant isolation
test_update_isolation() {
    log_info "Test 7: Testing UPDATE operations respect tenant isolation"
    
    # Try to update Alpha's project as Beta tenant (should fail/not affect anything)
    psql "$DB_CONN" -c "
        SELECT init_tenant_context('tenant-beta', false);
        UPDATE projects SET description = 'Updated by Beta' WHERE id = 'test-proj-alpha';
    " >/dev/null 2>&1
    
    # Verify the project was not updated
    local description=$(psql "$DB_CONN" -t -c "
        SELECT init_tenant_context('tenant-alpha', false);
        SELECT description FROM projects WHERE id = 'test-proj-alpha';
    " | xargs)
    
    if [ "$description" = "Test project" ]; then
        test_passed "UPDATE isolation works correctly"
    else
        test_failed "UPDATE isolation failed. Description was changed to: '$description'"
    fi
}

# Test 8: Test DELETE operations respect tenant isolation
test_delete_isolation() {
    log_info "Test 8: Testing DELETE operations respect tenant isolation"
    
    # Count Alpha's projects before deletion attempt
    local before_count=$(psql "$DB_CONN" -t -c "
        SELECT init_tenant_context('tenant-alpha', false);
        SELECT COUNT(*) FROM projects WHERE tenant_id = 'tenant-alpha';
    " | tail -1 | tr -d ' ')
    
    # Try to delete Alpha's project as Beta tenant (should not delete anything)
    psql "$DB_CONN" -c "
        SELECT init_tenant_context('tenant-beta', false);
        DELETE FROM projects WHERE id = 'test-proj-alpha';
    " >/dev/null 2>&1
    
    # Count Alpha's projects after deletion attempt
    local after_count=$(psql "$DB_CONN" -t -c "
        SELECT init_tenant_context('tenant-alpha', false);
        SELECT COUNT(*) FROM projects WHERE tenant_id = 'tenant-alpha';
    " | tail -1 | tr -d ' ')
    
    if [ "$before_count" -eq "$after_count" ]; then
        test_passed "DELETE isolation works correctly"
    else
        test_failed "DELETE isolation failed. Count before: $before_count, after: $after_count"
    fi
}

# Test 9: Performance test - measure query time with RLS
test_performance() {
    log_info "Test 9: Performance testing with RLS policies"
    
    # Create some additional test data for performance testing
    psql "$DB_CONN" -c "
        SELECT init_tenant_context('tenant-alpha', false);
        INSERT INTO tasks (id, display_id, project_id, title, status, priority, type, tenant_id, created_at, updated_at)
        SELECT 
            'perf-task-' || generate_series(1, 1000),
            'PERF-' || generate_series(1, 1000),
            'proj-alpha-1',
            'Performance Task ' || generate_series(1, 1000),
            'todo',
            'medium',
            'task',
            'tenant-alpha',
            NOW(),
            NOW();
    " >/dev/null 2>&1
    
    # Measure query time
    local start_time=$(date +%s%N)
    psql "$DB_CONN" -c "
        SELECT init_tenant_context('tenant-alpha', false);
        SELECT COUNT(*) FROM tasks WHERE status = 'todo';
    " >/dev/null 2>&1
    local end_time=$(date +%s%N)
    
    local duration_ms=$(( (end_time - start_time) / 1000000 ))
    
    test_passed "Performance test completed in ${duration_ms}ms (queried ~1000 tasks)"
    
    if [ "$duration_ms" -lt 1000 ]; then
        log_info "Performance is acceptable (< 1000ms for 1000 tasks)"
    else
        log_warn "Performance may need optimization (${duration_ms}ms for 1000 tasks)"
    fi
}

# Test 10: Cleanup test data
cleanup_test_data() {
    log_info "Test 10: Cleaning up test data"
    
    psql "$DB_CONN" -c "
        -- Set admin context to clean up all test data
        SELECT init_tenant_context('tenant-alpha', true);
        
        -- Delete test data
        DELETE FROM tasks WHERE id LIKE 'task-%' OR id LIKE 'perf-task-%' OR id = 'test-task-alpha';
        DELETE FROM projects WHERE id LIKE 'proj-%' OR id = 'test-proj-alpha';
        DELETE FROM users WHERE id LIKE 'user-%';
        DELETE FROM tenants WHERE id LIKE 'tenant-%';
    " >/dev/null 2>&1
    
    test_passed "Test data cleaned up successfully"
}

# Main test execution
main() {
    log_info "Starting RLS comprehensive test suite"
    echo
    
    # Check if database is accessible
    if ! psql "$DB_CONN" -c "SELECT 1;" >/dev/null 2>&1; then
        log_error "Cannot connect to test database. Please ensure PostgreSQL is running and database exists."
        log_error "Connection string: $DB_CONN"
        exit 1
    fi
    
    # Run tests
    test_rls_enabled
    test_rls_policies_exist
    setup_test_data
    test_tenant_isolation
    test_admin_bypass
    test_insert_isolation
    test_update_isolation
    test_delete_isolation
    test_performance
    cleanup_test_data
    
    echo
    log_info "All RLS tests completed successfully!"
    echo -e "${GREEN}🎉 Row-Level Security is working correctly!${NC}"
}

# Run the tests
main "$@"
