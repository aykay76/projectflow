# Multi-Tenant Database Schema Evolution

This document describes how ProjectFlow handles database schema evolution for multi-tenant support.

## Approach: Go-Based Schema Initialization

ProjectFlow uses Go-based schema initialization rather than separate SQL migration files. This approach provides:

- **Simplicity**: All schema logic is in one place
- **Automatic**: Works on application startup without manual intervention
- **Backward Compatible**: Safely handles existing databases
- **Self-Contained**: No external migration tools required

## Schema Evolution Process

The `initializeSchema()` function in `internal/storage/postgres_storage.go` follows this careful sequence:

### Step 1: Create Core Tables
Creates the `tenants` table first, as it's required for foreign key relationships.

### Step 2: Create Base Tables (Without Tenant FKs)
Creates `tasks`, `projects`, and `users` tables. For new installations, the `users` table includes all constraints. For existing installations, `tasks` and `projects` are created without tenant foreign keys to avoid constraint failures.

### Step 3: Add Missing Columns
Uses conditional `ALTER TABLE` statements to add missing columns to existing databases:
- `task_counter` column to projects
- `tenant_id` column to projects and tasks
- `display_id` and `project_id` columns to tasks

### Step 4: Add Foreign Key Constraints
After all columns exist, adds foreign key constraints:
- `tenants.id` ← `projects.tenant_id`
- `tenants.id` ← `tasks.tenant_id`
- `tasks.id` ← `tasks.parent_id`
- `projects.id` ← `tasks.project_id`

### Step 5: Create Indexes
Creates all necessary indexes for performance.

## Database Compatibility

### New Installations
- All tables are created with complete schema including tenant support
- All constraints and indexes are applied immediately

### Existing Installations
- Existing tables are preserved
- Missing columns are added safely with conditional logic
- Foreign key constraints are added after columns exist
- Unique constraints are updated to support multi-tenancy

## Multi-Tenant Schema Design

### Tenants Table
```sql
CREATE TABLE tenants (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    settings JSONB DEFAULT '{}'::jsonb,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT check_tenant_status CHECK (status IN ('active', 'inactive', 'suspended'))
);
```

### Modified Tables
All major tables now include `tenant_id` for data isolation:

**Projects:**
- Added `tenant_id VARCHAR(36)` with foreign key to tenants
- Unique constraint changed from `(name)` to `(tenant_id, name)`

**Tasks:**
- Added `tenant_id VARCHAR(36)` with foreign key to tenants
- Maintains existing parent-child relationships within tenants

**Users (New):**
- Includes `tenant_id` from creation
- Unique constraints on `(tenant_id, username)` and `(tenant_id, email)`

## Performance Considerations

### Indexes Added
- `idx_tenants_status` - For tenant status filtering
- `idx_tenants_created_at` - For tenant creation date queries
- `idx_tasks_tenant_id` - For tenant-specific task queries
- `idx_tasks_tenant_status` - For tenant + status filtering
- `idx_tasks_tenant_priority` - For tenant + priority filtering
- `idx_projects_tenant_id` - For tenant-specific project queries
- `idx_projects_tenant_name` - For tenant + name lookups
- `idx_users_tenant_*` - For user authentication and management

### Query Patterns
All queries should now include `tenant_id` in WHERE clauses for data isolation and performance.

## Testing Recommendations

### Fresh Database Test
1. Start with empty PostgreSQL database
2. Run application - should create all tables with full schema
3. Verify all constraints and indexes exist

### Existing Database Test
1. Start with existing ProjectFlow database (pre-multi-tenant)
2. Run application - should add missing columns and constraints
3. Verify data integrity and all new columns exist
4. Test that existing functionality still works

## Rollback Considerations

Since this is Go-based schema initialization:
- **Forward Migration**: Automatic on application startup
- **Rollback**: Would require manual SQL scripts to remove columns/constraints
- **Recommendation**: Always backup database before upgrading

## Future Schema Changes

For future schema changes, follow the same pattern:
1. Add new table creation to Step 2
2. Add column additions to Step 3  
3. Add constraint additions to Step 4
4. Add indexes to Step 5
5. Test with both fresh and existing databases
