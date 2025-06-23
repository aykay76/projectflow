-- Row-Level Security (RLS) Implementation for ProjectFlow
-- This script enables RLS on tenant-aware tables and creates policies for tenant isolation

-- Enable RLS on tenant-aware tables
ALTER TABLE tenants ENABLE ROW LEVEL SECURITY;
ALTER TABLE projects ENABLE ROW LEVEL SECURITY;
ALTER TABLE tasks ENABLE ROW LEVEL SECURITY;
ALTER TABLE users ENABLE ROW LEVEL SECURITY;

-- Create function to get current tenant ID from session context
CREATE OR REPLACE FUNCTION get_current_tenant_id() RETURNS VARCHAR(36) AS $$
BEGIN
    RETURN current_setting('app.current_tenant_id', true);
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Create function to set current tenant ID in session context
CREATE OR REPLACE FUNCTION set_current_tenant_id(tenant_id VARCHAR(36)) RETURNS VOID AS $$
BEGIN
    PERFORM set_config('app.current_tenant_id', tenant_id, true);
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Create function to check if current user is admin (bypasses RLS)
CREATE OR REPLACE FUNCTION is_admin_user() RETURNS BOOLEAN AS $$
BEGIN
    RETURN current_setting('app.is_admin', true)::BOOLEAN;
EXCEPTION
    WHEN OTHERS THEN
        RETURN FALSE;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Create function to set admin context
CREATE OR REPLACE FUNCTION set_admin_context(is_admin BOOLEAN) RETURNS VOID AS $$
BEGIN
    PERFORM set_config('app.is_admin', is_admin::TEXT, true);
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- RLS Policies for tenants table
-- Admin users can see all tenants, regular users can only see their own tenant
CREATE POLICY tenant_isolation_policy ON tenants
    FOR ALL
    TO PUBLIC
    USING (
        is_admin_user() OR 
        id = get_current_tenant_id()
    );

-- RLS Policies for projects table
-- Users can only access projects belonging to their tenant
CREATE POLICY project_tenant_isolation_policy ON projects
    FOR ALL
    TO PUBLIC
    USING (
        is_admin_user() OR 
        tenant_id = get_current_tenant_id()
    );

-- RLS Policies for tasks table
-- Users can only access tasks belonging to their tenant
CREATE POLICY task_tenant_isolation_policy ON tasks
    FOR ALL
    TO PUBLIC
    USING (
        is_admin_user() OR 
        tenant_id = get_current_tenant_id()
    );

-- RLS Policies for users table
-- Users can only access users belonging to their tenant
CREATE POLICY user_tenant_isolation_policy ON users
    FOR ALL
    TO PUBLIC
    USING (
        is_admin_user() OR 
        tenant_id = get_current_tenant_id()
    );

-- Create helper function to initialize tenant context for a transaction
CREATE OR REPLACE FUNCTION init_tenant_context(tenant_id VARCHAR(36), is_admin BOOLEAN DEFAULT FALSE) RETURNS VOID AS $$
BEGIN
    -- Set tenant context
    PERFORM set_current_tenant_id(tenant_id);
    
    -- Set admin context if specified
    IF is_admin THEN
        PERFORM set_admin_context(TRUE);
    ELSE
        PERFORM set_admin_context(FALSE);
    END IF;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Create function to clear tenant context
CREATE OR REPLACE FUNCTION clear_tenant_context() RETURNS VOID AS $$
BEGIN
    PERFORM set_config('app.current_tenant_id', '', true);
    PERFORM set_config('app.is_admin', 'false', true);
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;
