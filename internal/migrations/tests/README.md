# Migration Testing Suite

This directory contains comprehensive tests for database migration scripts to ensure safe migration of existing single-tenant data to the new multi-tenant schema.

## Overview

The migration testing suite provides:

- **Automated migration testing** with different data volumes
- **Data integrity validation** after migrations
- **Performance benchmarking** with large datasets
- **Rollback testing** to ensure safe reversal
- **Constraint and index validation** 

## Test Structure

```
internal/migrations/tests/
├── migration_test.go          # Main test suite
├── README.md                  # This file
└── scripts/
    ├── test-migrations.sh     # Main test runner
    └── test-performance.sh    # Performance testing
```

## Running Tests

### Prerequisites

1. **PostgreSQL**: Ensure PostgreSQL is installed and running
2. **Go**: Go 1.24+ for running tests
3. **Database Access**: Superuser access to create test databases

### Quick Start

Run the complete test suite:

```bash
# Run all migration tests
./scripts/test-migrations.sh

# Run performance tests
./scripts/test-migration-performance.sh
```

### Test Options

#### Basic Migration Testing

```bash
# Setup test database only
./scripts/test-migrations.sh --setup-only

# Run tests on existing database
./scripts/test-migrations.sh --test-only

# Cleanup test database
./scripts/test-migrations.sh --cleanup-only
```

#### Performance Testing

The performance test script tests migration performance with different dataset sizes:

- **Small**: 1,000 records
- **Medium**: 10,000 records  
- **Large**: 100,000 records

## Test Categories

### 1. Schema Migration Tests

Tests that verify:
- Migration tracking table creation
- Tenant ID column addition
- Index creation
- Foreign key constraint setup

### 2. Data Migration Tests

Tests that verify:
- Default tenant creation
- Existing data migration to default tenant
- Data integrity after migration
- No NULL tenant_id values remain

### 3. Constraint Tests

Tests that verify:
- Foreign key constraints work correctly
- Unique constraints are enforced
- Check constraints function properly

### 4. Index Tests

Tests that verify:
- Performance indexes are created
- Indexes improve query performance
- Index naming conventions

### 5. Rollback Tests

Tests that verify:
- Migrations can be safely rolled back
- Data is restored to previous state
- No orphaned data remains

### 6. Performance Tests

Tests that verify:
- Migration performance with large datasets
- Memory usage during migration
- Transaction handling under load

## Configuration

### Environment Variables

```bash
# Test database configuration
export TEST_DB_HOST="localhost"
export TEST_DB_PORT="5432"
export TEST_DB_NAME="projectflow_migration_test"
export TEST_DB_USER="projectflow_test"
export TEST_DB_PASSWORD="test_password"

# Performance test configuration
export PERF_DB_NAME="projectflow_perf_test"
export PERF_DB_USER="projectflow_perf"
export PERF_DB_PASSWORD="perf_password"
```

### Database Requirements

The test scripts require:
- PostgreSQL 12+ running
- Superuser access (postgres user)
- Ability to create/drop databases
- Network access to database server

## Test Data

### Sample Data Structure

The tests create realistic sample data:

```sql
-- Projects
proj-1: Test Project 1 (TP1)
proj-2: Test Project 2 (TP2)  
proj-3: Test Project 3 (TP3)

-- Tasks
TP1-1: Test Task 1 (todo, high priority)
TP1-2: Test Task 2 (in_progress, medium priority)
TP2-1: Test Task 3 (done, low priority)
TP3-1: Test Task 4 (todo, high priority, story type)
TP3-2: Test Task 5 (blocked, medium priority, epic type)
```

### Performance Data

Performance tests generate:
- Projects: 10% of total records
- Tasks: 90% of total records
- Realistic foreign key relationships
- Varied status, priority, and type distributions

## Expected Results

### Migration Success Criteria

✅ **Schema Migration**
- All tenant_id columns added successfully
- Indexes created on tenant_id columns
- Foreign key constraints optional until data migration

✅ **Data Migration** 
- Default tenant created with consistent UUID
- All existing projects assigned to default tenant
- All existing tasks assigned to default tenant
- No NULL tenant_id values remain

✅ **Performance Benchmarks**
- Small dataset (1K): < 5 seconds
- Medium dataset (10K): < 30 seconds
- Large dataset (100K): < 5 minutes

✅ **Rollback Capability**
- All migrations can be reversed
- Original data structure restored
- No data loss during rollback

## Troubleshooting

### Common Issues

**PostgreSQL Connection Errors**
```bash
# Check if PostgreSQL is running
pg_isready -h localhost -p 5432

# Check if you have superuser access
psql -U postgres -c "SELECT current_user;"
```

**Permission Errors**
```bash
# Grant necessary permissions
psql -U postgres -c "ALTER USER $TEST_DB_USER CREATEDB;"
psql -U postgres -c "GRANT ALL PRIVILEGES ON DATABASE $TEST_DB_NAME TO $TEST_DB_USER;"
```

**Migration Failures**
```bash
# Check migration status
./build/migrate status

# View detailed migration logs
./build/migrate up 2>&1 | tee migration.log
```

### Debug Mode

Enable verbose logging:
```bash
export DEBUG=1
./scripts/test-migrations.sh
```

## Integration with CI/CD

### GitHub Actions Example

```yaml
name: Migration Tests
on: [push, pull_request]

jobs:
  migration-tests:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_PASSWORD: postgres
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5

    steps:
    - uses: actions/checkout@v4
    - uses: actions/setup-go@v4
      with:
        go-version: '1.24'
    
    - name: Run Migration Tests
      run: ./scripts/test-migrations.sh
      env:
        TEST_DB_HOST: localhost
        TEST_DB_PORT: 5432
```

## Best Practices

### Before Running in Production

1. **Backup Database**: Always backup before migration
2. **Test on Copy**: Run tests on production copy first
3. **Monitor Performance**: Check migration performance estimates
4. **Plan Downtime**: Estimate required maintenance window
5. **Rollback Plan**: Ensure rollback procedures are tested

### Continuous Testing

1. **Run tests on every migration change**
2. **Include performance regression tests**
3. **Validate against production-like data volumes**
4. **Test rollback procedures regularly**

## Contributing

When adding new migrations:

1. **Add corresponding tests** in `migration_test.go`
2. **Update performance benchmarks** if needed
3. **Document breaking changes** and rollback procedures
4. **Test with various data scenarios**

## Support

For issues with migration tests:

1. Check the troubleshooting section above
2. Review migration logs for detailed errors
3. Ensure all prerequisites are met
4. Test with smaller datasets first
