-- Migration: Add foreign key constraints for tenant_id columns
-- Version: 20250623214100
-- Description: Add foreign key constraints after data migration is complete

-- +migrate Up
-- This migration adds foreign key constraints to enforce referential integrity
-- This should be run AFTER the data migration is complete

-- Add foreign key constraint for projects.tenant_id
DO $$ 
BEGIN 
    -- Only add the constraint if it doesn't exist
    IF NOT EXISTS (
        SELECT constraint_name 
        FROM information_schema.table_constraints 
        WHERE table_name = 'projects' 
        AND constraint_name = 'fk_projects_tenant_id'
    ) THEN
        ALTER TABLE projects 
        ADD CONSTRAINT fk_projects_tenant_id 
        FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE;
    END IF;
END $$;

-- Add foreign key constraint for tasks.tenant_id
DO $$ 
BEGIN 
    -- Only add the constraint if it doesn't exist
    IF NOT EXISTS (
        SELECT constraint_name 
        FROM information_schema.table_constraints 
        WHERE table_name = 'tasks' 
        AND constraint_name = 'fk_tasks_tenant_id'
    ) THEN
        ALTER TABLE tasks 
        ADD CONSTRAINT fk_tasks_tenant_id 
        FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE;
    END IF;
END $$;

-- Add NOT NULL constraints to tenant_id columns
-- (This should only be done after all existing data has been migrated)
-- 
-- Note: Commented out for safety - enable these manually after confirming
-- all data has been properly migrated to avoid constraint violations
--
-- ALTER TABLE projects ALTER COLUMN tenant_id SET NOT NULL;
-- ALTER TABLE tasks ALTER COLUMN tenant_id SET NOT NULL;

-- +migrate Down
-- This section contains the SQL statements to rollback the migration

-- Remove NOT NULL constraints first (if they were added)
-- ALTER TABLE tasks ALTER COLUMN tenant_id DROP NOT NULL;
-- ALTER TABLE projects ALTER COLUMN tenant_id DROP NOT NULL;

-- Remove foreign key constraints
ALTER TABLE tasks DROP CONSTRAINT IF EXISTS fk_tasks_tenant_id;
ALTER TABLE projects DROP CONSTRAINT IF EXISTS fk_projects_tenant_id;
