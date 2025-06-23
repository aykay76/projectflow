-- Migration: 003_add_tenant_to_tasks.sql
-- Description: Add tenant_id column to tasks table for multi-tenant support
-- Author: ProjectFlow Team
-- Date: 2025-06-23

-- Add tenant_id column to tasks table if it doesn't exist
DO $$ 
BEGIN 
    IF NOT EXISTS (
        SELECT column_name 
        FROM information_schema.columns 
        WHERE table_name='tasks' AND column_name='tenant_id'
    ) THEN
        ALTER TABLE tasks ADD COLUMN tenant_id VARCHAR(36);
    END IF;
END $$;

-- Add foreign key constraint referencing tenants table
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT constraint_name 
        FROM information_schema.table_constraints 
        WHERE table_name='tasks' AND constraint_name='fk_tasks_tenant_id'
    ) THEN
        ALTER TABLE tasks 
        ADD CONSTRAINT fk_tasks_tenant_id 
        FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE;
    END IF;
END $$;

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_tasks_tenant_id ON tasks(tenant_id);
CREATE INDEX IF NOT EXISTS idx_tasks_tenant_status ON tasks(tenant_id, status);
CREATE INDEX IF NOT EXISTS idx_tasks_tenant_priority ON tasks(tenant_id, priority);

-- Add comment to the new column
COMMENT ON COLUMN tasks.tenant_id IS 'Reference to tenant that owns this task';
