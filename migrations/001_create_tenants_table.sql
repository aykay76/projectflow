-- Migration: 001_create_tenants_table.sql
-- Description: Create tenants table for multi-tenant support
-- Author: ProjectFlow Team
-- Date: 2025-06-23

-- Create tenants table
CREATE TABLE IF NOT EXISTS tenants (
    id VARCHAR(36) PRIMARY KEY COMMENT 'Unique tenant identifier (UUID)',
    name VARCHAR(255) NOT NULL UNIQUE COMMENT 'Tenant name (must be unique)',
    settings JSONB DEFAULT '{}'::jsonb COMMENT 'Tenant-specific configuration settings',
    status VARCHAR(20) NOT NULL DEFAULT 'active' COMMENT 'Tenant status (active, inactive, suspended)',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW() COMMENT 'Timestamp when tenant was created',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW() COMMENT 'Timestamp when tenant was last updated',
    
    -- Add check constraint for valid status values
    CONSTRAINT check_tenant_status CHECK (status IN ('active', 'inactive', 'suspended'))
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_tenants_status ON tenants(status);
CREATE INDEX IF NOT EXISTS idx_tenants_created_at ON tenants(created_at);

-- Add comments to the table
COMMENT ON TABLE tenants IS 'Multi-tenant support - stores tenant information and settings';
