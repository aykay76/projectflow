-- Migration: Add tenant_id columns to existing tables
-- Version: 20250623213900
-- Description: Add tenant_id columns to projects and tasks tables for multi-tenant support

-- +migrate Up
-- This migration adds tenant_id columns to existing tables to support multi-tenancy
-- The columns are initially nullable to allow for gradual migration of existing data

-- First, ensure the tenants table exists with proper structure
CREATE TABLE IF NOT EXISTS tenants (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    settings JSONB DEFAULT '{}'::jsonb,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT check_tenant_status CHECK (status IN ('active', 'inactive', 'suspended'))
);

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

-- Create indexes on tenant_id columns for performance
CREATE INDEX IF NOT EXISTS idx_projects_tenant_id ON projects(tenant_id);
CREATE INDEX IF NOT EXISTS idx_tasks_tenant_id ON tasks(tenant_id);

-- Note: Foreign key constraints will be added in a later migration
-- after data migration is complete, to avoid constraint violations
-- during the transition period

-- +migrate Down
-- This section contains the SQL statements to rollback the migration

-- Remove indexes first
DROP INDEX IF EXISTS idx_tasks_tenant_id;
DROP INDEX IF EXISTS idx_projects_tenant_id;

-- Remove tenant_id columns
ALTER TABLE tasks DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE projects DROP COLUMN IF EXISTS tenant_id;

-- Note: We don't drop the tenants table here as it might be referenced elsewhere
-- The tenants table will be managed by its own migration
