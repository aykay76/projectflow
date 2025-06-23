-- Migration: 002_add_tenant_to_projects.sql
-- Description: Add tenant_id column to projects table for multi-tenant support
-- Author: ProjectFlow Team
-- Date: 2025-06-23

-- Add tenant_id column to projects table if it doesn't exist
DO $$ 
BEGIN 
    IF NOT EXISTS (
        SELECT column_name 
        FROM information_schema.columns 
        WHERE table_name='projects' AND column_name='tenant_id'
    ) THEN
        ALTER TABLE projects ADD COLUMN tenant_id VARCHAR(36);
    END IF;
END $$;

-- Add foreign key constraint referencing tenants table
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT constraint_name 
        FROM information_schema.table_constraints 
        WHERE table_name='projects' AND constraint_name='fk_projects_tenant_id'
    ) THEN
        ALTER TABLE projects 
        ADD CONSTRAINT fk_projects_tenant_id 
        FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE;
    END IF;
END $$;

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_projects_tenant_id ON projects(tenant_id);
CREATE INDEX IF NOT EXISTS idx_projects_tenant_name ON projects(tenant_id, name);

-- Add comment to the new column
COMMENT ON COLUMN projects.tenant_id IS 'Reference to tenant that owns this project';
