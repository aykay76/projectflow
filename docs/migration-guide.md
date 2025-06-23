# Multi-Tenant Migration Guide

This guide provides step-by-step instructions for migrating existing ProjectFlow single-tenant data to the new multi-tenant schema.

## Overview

The migration process transforms a single-tenant ProjectFlow installation into a multi-tenant system by:

1. Adding tenant_id columns to existing tables
2. Creating a default tenant for existing data
3. Migrating all existing projects and tasks to the default tenant
4. Adding proper foreign key constraints and indexes
5. Enforcing data integrity with NOT NULL constraints

## Pre-Migration Checklist

### 1. System Requirements

- [ ] PostgreSQL 12+ database server
- [ ] ProjectFlow application stopped
- [ ] Database superuser access
- [ ] Sufficient disk space (estimate 20% additional for migration)
- [ ] Network connectivity to database server

### 2. Backup Requirements

- [ ] **CRITICAL**: Full database backup completed
- [ ] Backup verified and tested for restoration
- [ ] Backup stored in secure, accessible location
- [ ] Document backup location and restoration procedure

```bash
# Create full database backup
pg_dump -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME > projectflow_pre_migration_backup.sql

# Verify backup
psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d projectflow_test < projectflow_pre_migration_backup.sql
```

### 3. Environment Setup

- [ ] Migration tool compiled and tested
- [ ] Database connection parameters configured
- [ ] Test environment validated with production copy
- [ ] Rollback procedures tested and documented

### 4. Data Assessment

- [ ] Current data volume assessed
- [ ] Migration time estimated based on test runs
- [ ] Maintenance window scheduled
- [ ] Stakeholders notified of downtime

```bash
# Assess current data volume
psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "
SELECT 
    'projects' as table_name, COUNT(*) as record_count 
    FROM projects
UNION ALL
SELECT 
    'tasks' as table_name, COUNT(*) as record_count 
    FROM tasks;
"
```

## Migration Execution Process

### Phase 1: Schema Migration

**Estimated Time**: 5-10 seconds per table
**Downtime Required**: Yes

1. **Stop the ProjectFlow application**
   ```bash
   # Stop application (method depends on your deployment)
   systemctl stop projectflow
   # OR
   docker stop projectflow
   # OR
   pkill -f projectflow
   ```

2. **Initialize migration tracking**
   ```bash
   cd /path/to/projectflow
   ./build/migrate init
   ```

3. **Check migration status**
   ```bash
   ./build/migrate status
   ```

4. **Apply schema migrations**
   ```bash
   ./build/migrate up
   ```

   Expected output:
   ```
   🚀 Applying pending migrations...
   📋 Found 4 pending migrations
   ✅ Applied migration 20250623213900: add_tenant_id_columns
   ✅ Applied migration 20250623214100: add_tenant_foreign_keys
   ✅ Applied migration 20250623214200: migrate_existing_data_to_default_tenant
   ✅ Applied migration 20250623214300: make_tenant_id_not_null
   ✅ Successfully applied 4 migrations
   ```

### Phase 2: Data Validation

**Estimated Time**: 30 seconds to 2 minutes
**Downtime Required**: Yes (still in maintenance window)

1. **Validate data integrity**
   ```bash
   # Check that all records have tenant_id
   psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "
   SELECT 
       'projects_without_tenant' as check_name,
       COUNT(*) as count
       FROM projects WHERE tenant_id IS NULL
   UNION ALL
   SELECT 
       'tasks_without_tenant' as check_name,
       COUNT(*) as count
       FROM tasks WHERE tenant_id IS NULL
   UNION ALL
   SELECT 
       'default_tenant_exists' as check_name,
       COUNT(*) as count
       FROM tenants WHERE id = '00000000-0000-0000-0000-000000000001';
   "
   ```

   Expected results:
   ```
   check_name                | count
   -------------------------+-------
   projects_without_tenant  |     0
   tasks_without_tenant     |     0
   default_tenant_exists    |     1
   ```

2. **Validate foreign key constraints**
   ```bash
   # Test foreign key constraint enforcement
   psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "
   INSERT INTO projects (id, name, display_prefix, tenant_id) 
   VALUES ('test-proj', 'Test Project', 'TEST', 'invalid-tenant-id');
   "
   ```
   
   Expected: Should fail with foreign key constraint error

3. **Validate indexes**
   ```bash
   # Check that performance indexes exist
   psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "
   SELECT indexname, tablename 
   FROM pg_indexes 
   WHERE indexname IN ('idx_projects_tenant_id', 'idx_tasks_tenant_id');
   "
   ```

### Phase 3: Application Restart

**Estimated Time**: 30 seconds to 2 minutes
**Downtime Required**: Until application starts

1. **Update application configuration** (if needed)
   - Ensure multi-tenant features are enabled
   - Update any tenant-specific configurations

2. **Start the ProjectFlow application**
   ```bash
   # Start application (method depends on your deployment)
   systemctl start projectflow
   # OR
   docker start projectflow
   # OR
   ./projectflow-server
   ```

3. **Verify application functionality**
   - Test login and basic operations
   - Verify existing projects and tasks are accessible
   - Check that new data includes tenant_id

## Rollback Procedures

**⚠️ WARNING**: Rollback will lose any data created after migration

### Automatic Rollback

Use the migration tool to rollback to pre-migration state:

```bash
# Rollback to specific version (before tenant migration)
./build/migrate down -version 20250623213800

# Or rollback all migrations
./build/migrate down -version 0
```

### Manual Rollback

If automatic rollback fails:

1. **Stop the application**
   ```bash
   systemctl stop projectflow
   ```

2. **Restore from backup**
   ```bash
   # Drop current database
   psql -h $DB_HOST -p $DB_PORT -U postgres -c "DROP DATABASE $DB_NAME;"
   
   # Recreate database
   psql -h $DB_HOST -p $DB_PORT -U postgres -c "CREATE DATABASE $DB_NAME OWNER $DB_USER;"
   
   # Restore from backup
   psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME < projectflow_pre_migration_backup.sql
   ```

3. **Restart application**
   ```bash
   systemctl start projectflow
   ```

## Performance Estimates

Based on test data with various dataset sizes:

| Dataset Size | Migration Time | Notes |
|-------------|----------------|--------|
| Small (<1K records) | < 5 seconds | Typical small installation |
| Medium (1K-10K records) | 10-30 seconds | Typical medium installation |
| Large (10K-100K records) | 1-5 minutes | Large installation |
| Very Large (>100K records) | 5+ minutes | Enterprise installation |

**Note**: Times may vary based on:
- Database server performance
- Network latency
- Concurrent database load
- Storage type (SSD vs HDD)

## Troubleshooting Guide

### Common Issues

#### 1. Migration Tool Cannot Connect to Database

**Symptoms**: Connection timeout or authentication errors

**Solutions**:
```bash
# Check database connectivity
pg_isready -h $DB_HOST -p $DB_PORT

# Test authentication
psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "SELECT current_user;"

# Check environment variables
echo $DB_HOST $DB_PORT $DB_NAME $DB_USER
```

#### 2. Foreign Key Constraint Violations

**Symptoms**: Migration fails with "violates foreign key constraint"

**Possible Causes**:
- Orphaned tasks referencing non-existent projects
- Data inconsistencies in existing database

**Solutions**:
```bash
# Find orphaned tasks
psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "
SELECT t.id, t.display_id, t.project_id 
FROM tasks t 
LEFT JOIN projects p ON t.project_id = p.id 
WHERE t.project_id IS NOT NULL AND p.id IS NULL;
"

# Clean up orphaned data (CAREFUL!)
psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "
UPDATE tasks SET project_id = NULL 
WHERE project_id IS NOT NULL 
AND project_id NOT IN (SELECT id FROM projects);
"
```

#### 3. Insufficient Disk Space

**Symptoms**: Migration fails with disk space errors

**Solutions**:
```bash
# Check available disk space
df -h

# Clean up unnecessary files
# Free up space or migrate to larger storage
```

#### 4. Migration Appears Stuck

**Symptoms**: Migration command doesn't return after expected time

**Solutions**:
```bash
# Check database locks
psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "
SELECT pid, state, query_start, query 
FROM pg_stat_activity 
WHERE state <> 'idle' AND query NOT ILIKE '%pg_stat_activity%';
"

# Kill long-running queries if safe
# pg_cancel_backend(pid) or pg_terminate_backend(pid)
```

#### 5. Application Won't Start After Migration

**Symptoms**: Application crashes or fails to connect to database

**Solutions**:
1. Check application logs for specific errors
2. Verify database schema matches expectations
3. Check application configuration for multi-tenant settings
4. Test database connectivity manually

### Data Validation Errors

#### Unexpected NULL tenant_id Values

```sql
-- Find records without tenant_id after migration
SELECT 'projects' as table_name, COUNT(*) as null_count 
FROM projects WHERE tenant_id IS NULL
UNION ALL
SELECT 'tasks' as table_name, COUNT(*) as null_count 
FROM tasks WHERE tenant_id IS NULL;
```

**Solution**: Re-run the data migration:
```bash
./build/migrate down -version 20250623214100
./build/migrate up
```

#### Foreign Key Constraint Issues

```sql
-- Check for constraint violations
SELECT conname, conrelid::regclass as table_name
FROM pg_constraint 
WHERE contype = 'f' AND NOT convalidated;
```

## Post-Migration Checklist

### Immediate Verification (First 24 Hours)

- [ ] Application starts successfully
- [ ] All existing projects accessible
- [ ] All existing tasks accessible  
- [ ] New projects/tasks created with tenant_id
- [ ] Database performance within normal ranges
- [ ] Application logs show no migration-related errors

### Extended Monitoring (First Week)

- [ ] Monitor database performance metrics
- [ ] Check for any data consistency issues
- [ ] Verify backup procedures work with new schema
- [ ] Monitor application error rates
- [ ] User acceptance testing completed

### Long-term Maintenance

- [ ] Update backup procedures for multi-tenant schema
- [ ] Update monitoring dashboards for tenant-specific metrics
- [ ] Plan for tenant management features
- [ ] Document lessons learned from migration

## Emergency Contacts

Maintain contact information for:

- Database Administrator
- System Administrator  
- Application Developer
- Business Stakeholder
- Backup System Manager

## Recovery Time Objectives

| Scenario | Target Recovery Time | Recovery Method |
|----------|---------------------|-----------------|
| Migration failure (early detection) | 15 minutes | Automated rollback |
| Migration failure (late detection) | 1 hour | Manual rollback from backup |
| Partial data corruption | 2 hours | Selective data restoration |
| Complete system failure | 4 hours | Full system restoration |

## Documentation Updates

After successful migration:

1. Update system architecture documentation
2. Update database schema documentation  
3. Update backup and recovery procedures
4. Update monitoring and alerting configurations
5. Update user guides for multi-tenant features

---

**⚠️ IMPORTANT REMINDERS**

1. **ALWAYS backup before migration**
2. **Test migration on production copy first**
3. **Have rollback plan ready**
4. **Monitor system closely after migration**
5. **Document any deviations from this guide**

For additional support or questions, consult the migration testing documentation in `internal/migrations/tests/README.md`.
