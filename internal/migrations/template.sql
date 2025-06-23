-- Migration Template
-- This file provides a template for creating new migrations

-- Migration naming convention: {timestamp}_{descriptive_name}.sql
-- Example: 20240623140000_add_tenant_support.sql

-- Migration structure:
-- 1. Metadata comments describing the migration
-- 2. Up migration (changes to apply)
-- 3. Down migration (changes to rollback)

-- Example migration file:

-- Migration: Add tenant support
-- Version: 20240623140000
-- Description: Add tenant_id columns to existing tables for multi-tenant support

-- +migrate Up
-- This section contains the SQL statements to apply the migration

-- Create tenants table
CREATE TABLE IF NOT EXISTS tenants (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    settings JSONB DEFAULT '{}'::jsonb,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT check_tenant_status CHECK (status IN ('active', 'inactive', 'suspended'))
);

-- Add tenant_id column to projects table
ALTER TABLE projects ADD COLUMN IF NOT EXISTS tenant_id VARCHAR(36);

-- Add tenant_id column to tasks table  
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS tenant_id VARCHAR(36);

-- Add indexes for performance
CREATE INDEX IF NOT EXISTS idx_projects_tenant_id ON projects(tenant_id);
CREATE INDEX IF NOT EXISTS idx_tasks_tenant_id ON tasks(tenant_id);

-- +migrate Down
-- This section contains the SQL statements to rollback the migration

-- Remove indexes
DROP INDEX IF EXISTS idx_tasks_tenant_id;
DROP INDEX IF EXISTS idx_projects_tenant_id;

-- Remove tenant_id columns
ALTER TABLE tasks DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE projects DROP COLUMN IF EXISTS tenant_id;

-- Drop tenants table
DROP TABLE IF EXISTS tenants;

-- Notes:
-- - Always use IF EXISTS/IF NOT EXISTS for safe migrations
-- - Use transactions for complex migrations
-- - Test migrations on a copy of production data
-- - Consider data migration separately from schema changes
-- - Document any manual steps required
