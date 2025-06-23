-- Migration: Make tenant_id columns NOT NULL
-- Version: 20250623214300
-- Description: Add NOT NULL constraints to tenant_id columns after data migration

-- +migrate Up
-- This migration adds NOT NULL constraints to tenant_id columns
-- This should be run AFTER the data migration is complete and verified

-- Verify that all records have tenant_id before proceeding
DO $$
BEGIN
    -- Check projects table
    IF EXISTS (SELECT 1 FROM projects WHERE tenant_id IS NULL) THEN
        RAISE EXCEPTION 'Cannot make tenant_id NOT NULL: projects table contains NULL values';
    END IF;
    
    -- Check tasks table
    IF EXISTS (SELECT 1 FROM tasks WHERE tenant_id IS NULL) THEN
        RAISE EXCEPTION 'Cannot make tenant_id NOT NULL: tasks table contains NULL values';
    END IF;
    
    -- If we get here, all records have tenant_id values
    RAISE NOTICE 'All records have tenant_id values, proceeding with NOT NULL constraints';
END $$;

-- Add NOT NULL constraint to projects.tenant_id
ALTER TABLE projects ALTER COLUMN tenant_id SET NOT NULL;

-- Add NOT NULL constraint to tasks.tenant_id
ALTER TABLE tasks ALTER COLUMN tenant_id SET NOT NULL;

-- +migrate Down
-- This section removes the NOT NULL constraints

-- Remove NOT NULL constraints
ALTER TABLE tasks ALTER COLUMN tenant_id DROP NOT NULL;
ALTER TABLE projects ALTER COLUMN tenant_id DROP NOT NULL;
