-- Migration: 004_create_users_table.sql
-- Description: Create users table with tenant support
-- Author: ProjectFlow Team
-- Date: 2025-06-23

-- Create users table
CREATE TABLE IF NOT EXISTS users (
    id VARCHAR(36) PRIMARY KEY COMMENT 'Unique user identifier (UUID)',
    tenant_id VARCHAR(36) NOT NULL COMMENT 'Reference to tenant that owns this user',
    username VARCHAR(255) NOT NULL COMMENT 'Username for authentication',
    email VARCHAR(255) NOT NULL COMMENT 'User email address',
    password_hash VARCHAR(255) NOT NULL COMMENT 'Hashed password for authentication',
    role VARCHAR(50) NOT NULL DEFAULT 'user' COMMENT 'User role (admin, user, viewer)',
    is_active BOOLEAN NOT NULL DEFAULT true COMMENT 'Whether the user account is active',
    last_login TIMESTAMPTZ COMMENT 'Timestamp of last successful login',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW() COMMENT 'Timestamp when user was created',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW() COMMENT 'Timestamp when user was last updated',
    
    -- Foreign key constraint
    FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Unique constraints for multi-tenant uniqueness
    UNIQUE(tenant_id, username),
    UNIQUE(tenant_id, email),
    
    -- Check constraint for valid role values
    CONSTRAINT check_user_role CHECK (role IN ('admin', 'user', 'viewer'))
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_users_tenant_id ON users(tenant_id);
CREATE INDEX IF NOT EXISTS idx_users_tenant_email ON users(tenant_id, email);
CREATE INDEX IF NOT EXISTS idx_users_tenant_username ON users(tenant_id, username);
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);
CREATE INDEX IF NOT EXISTS idx_users_is_active ON users(is_active);
CREATE INDEX IF NOT EXISTS idx_users_last_login ON users(last_login);

-- Add comments to the table
COMMENT ON TABLE users IS 'Multi-tenant user accounts with authentication and authorization';
