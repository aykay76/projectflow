-- Migration: Migrate existing data to default tenant
-- Version: 20250623214200
-- Description: Create default tenant and migrate all existing projects and tasks to it

-- +migrate Up
-- This migration creates a default tenant and assigns all existing data to it
-- This is a data migration that should be run after the schema migration

-- Start transaction for atomic operation
BEGIN;

-- Start transaction for atomic operation
BEGIN;

-- Define the default tenant ID as a constant UUID
-- Using a deterministic UUID based on "default" namespace
DO $$
DECLARE
    default_tenant_id VARCHAR(36) := '00000000-0000-0000-0000-000000000001';
    projects_updated INTEGER;
    tasks_updated INTEGER;
BEGIN
    -- Create default tenant if it doesn't exist
    INSERT INTO tenants (id, name, settings, status, created_at, updated_at)
    VALUES (
        default_tenant_id, 
        'Default Tenant', 
        '{"description": "Default tenant for existing data", "created_by": "migration"}',
        'active',
        NOW(),
        NOW()
    ) ON CONFLICT (id) DO UPDATE SET
        updated_at = NOW()
    WHERE tenants.id = default_tenant_id;
    
    -- Update all projects that don't have a tenant_id
    UPDATE projects 
    SET tenant_id = default_tenant_id, updated_at = NOW()
    WHERE tenant_id IS NULL;
    
    GET DIAGNOSTICS projects_updated = ROW_COUNT;
    
    -- Update all tasks that don't have a tenant_id
    UPDATE tasks 
    SET tenant_id = default_tenant_id, updated_at = NOW()
    WHERE tenant_id IS NULL;
    
    GET DIAGNOSTICS tasks_updated = ROW_COUNT;
    
    -- Log the migration results
    RAISE NOTICE 'Data migration completed: % projects and % tasks assigned to default tenant %', 
        projects_updated, tasks_updated, default_tenant_id;
        
    -- Validate that all projects and tasks now have tenant_id
    IF EXISTS (SELECT 1 FROM projects WHERE tenant_id IS NULL) THEN
        RAISE EXCEPTION 'Migration failed: some projects still have NULL tenant_id';
    END IF;
    
    IF EXISTS (SELECT 1 FROM tasks WHERE tenant_id IS NULL) THEN
        RAISE EXCEPTION 'Migration failed: some tasks still have NULL tenant_id';
    END IF;
    
END $$;

-- Commit the transaction
COMMIT;

-- +migrate Down
-- This section contains the SQL statements to rollback the migration

-- Start transaction for atomic rollback
BEGIN;

-- Start transaction for atomic rollback
BEGIN;

-- Define the default tenant ID (same as in the up migration)
DO $$
DECLARE
    default_tenant_id VARCHAR(36) := '00000000-0000-0000-0000-000000000001';
BEGIN
    -- Remove tenant assignments from tasks and projects
    UPDATE tasks SET tenant_id = NULL, updated_at = NOW() 
    WHERE tenant_id = default_tenant_id;

    UPDATE projects SET tenant_id = NULL, updated_at = NOW() 
    WHERE tenant_id = default_tenant_id;

    -- Remove the default tenant created by this migration
    DELETE FROM tenants 
    WHERE id = default_tenant_id 
    AND settings->>'created_by' = 'migration';

    RAISE NOTICE 'Data migration rollback completed for tenant %', default_tenant_id;
END $$;

-- Commit rollback transaction
COMMIT;
