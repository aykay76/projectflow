# ProjectFlow Freemium SaaS Architecture Plan

## Overview

Transform ProjectFlow from a single-user tool into a multi-tier freemium SaaS product while maintaining a single codebase. The solution will support multiple SKUs (Stock Keeping Units) through configuration-driven feature enablement, licensing, and tenant isolation.

## Architecture Strategy: Single Codebase, Multiple SKUs

### Core Principle
Use **feature flags**, **licensing**, and **configuration-driven** approaches to enable/disable features based on subscription tiers, rather than maintaining separate codebases.

## Subscription Tiers

### 1. Free Tier ("Community")
- **Target**: Individual developers, open source projects
- **Features**:
  - Up to 3 projects
  - Up to 100 tasks per project
  - File-based storage only
  - Basic web interface
  - MCP integration
  - Single user only
- **Limitations**:
  - No authentication required
  - No cloud sync
  - No collaboration features
  - No advanced analytics

### 2. Pro Tier ("Professional")
- **Target**: Individual professionals, small teams (1-5 users)
- **Features**:
  - Everything in Free +
  - Unlimited projects and tasks
  - PostgreSQL database support
  - User authentication (OAuth2)
  - Basic team collaboration
  - Cloud backup/sync
  - API rate limiting removal
  - Priority support
- **Price**: $9/month per user

### 3. Enterprise Tier ("Enterprise")
- **Target**: Large teams, organizations (5+ users)
- **Features**:
  - Everything in Pro +
  - Role-Based Access Control (RBAC)
  - Advanced project templates
  - Custom integrations
  - SSO/SAML authentication
  - Advanced analytics and reporting
  - Audit logging
  - SLA guarantees
  - On-premise deployment option
- **Price**: $19/month per user (minimum 5 users)

## Technical Implementation Strategy

### 1. License Management System
```go
type License struct {
    ID          string    `json:"id"`
    TierType    TierType  `json:"tier"`
    UserLimit   int       `json:"user_limit"`
    ProjectLimit int      `json:"project_limit"`
    TaskLimit   int       `json:"task_limit"`
    Features    []string  `json:"features"`
    ExpiresAt   time.Time `json:"expires_at"`
    IsActive    bool      `json:"is_active"`
}

type TierType string
const (
    TierFree       TierType = "free"
    TierPro        TierType = "pro"
    TierEnterprise TierType = "enterprise"
)
```

### 2. Feature Flag System
```go
type FeatureFlag struct {
    Name        string   `json:"name"`
    EnabledTiers []TierType `json:"enabled_tiers"`
    Description string   `json:"description"`
}

// Feature constants
const (
    FeatureUnlimitedTasks    = "unlimited_tasks"
    FeatureTeamCollaboration = "team_collaboration"
    FeatureAdvancedAuth      = "advanced_auth"
    FeatureRBAC             = "rbac"
    FeatureAnalytics        = "analytics"
    FeatureCloudSync        = "cloud_sync"
    FeatureCustomIntegration = "custom_integration"
)
```

### 3. Multi-Tenant Architecture
```go
type Tenant struct {
    ID          string    `json:"id"`
    Name        string    `json:"name"`
    License     License   `json:"license"`
    Settings    TenantSettings `json:"settings"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}

type TenantSettings struct {
    StorageType    string            `json:"storage_type"`
    DatabaseConfig map[string]string `json:"database_config"`
    AuthProvider   string            `json:"auth_provider"`
    CustomDomain   string            `json:"custom_domain,omitempty"`
}
```

### 4. Authentication & Authorization Layer
- JWT-based authentication for Pro/Enterprise tiers
- OAuth2 providers (Google, GitHub, Microsoft)
- RBAC system for Enterprise tier
- Session management
- API key management for integrations

### 5. Storage Abstraction Enhancement
```go
type TenantAwareStorage interface {
    storage.Storage
    SetTenant(tenantID string)
    GetTenant() string
    IsFeatureEnabled(feature string) bool
    CheckLimits(operation string, count int) error
}
```

## Database Schema Extensions

### New Tables
```sql
-- Tenants table
CREATE TABLE tenants (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    license_data JSONB NOT NULL,
    settings JSONB NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Users table
CREATE TABLE users (
    id VARCHAR(36) PRIMARY KEY,
    tenant_id VARCHAR(36) REFERENCES tenants(id),
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255),
    role VARCHAR(50) DEFAULT 'member',
    auth_provider VARCHAR(50),
    auth_provider_id VARCHAR(255),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Projects table (new concept for organizing tasks)
CREATE TABLE projects (
    id VARCHAR(36) PRIMARY KEY,
    tenant_id VARCHAR(36) REFERENCES tenants(id),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    owner_id VARCHAR(36) REFERENCES users(id),
    settings JSONB,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Update tasks table to include tenant_id and project_id
ALTER TABLE tasks ADD COLUMN tenant_id VARCHAR(36) REFERENCES tenants(id);
ALTER TABLE tasks ADD COLUMN project_id VARCHAR(36) REFERENCES projects(id);
ALTER TABLE tasks ADD COLUMN created_by VARCHAR(36) REFERENCES users(id);
ALTER TABLE tasks ADD COLUMN assigned_to VARCHAR(36) REFERENCES users(id);
```

## Configuration-Driven Feature Management

### Environment Variables
```bash
# Deployment mode
PROJECTFLOW_MODE=saas|onpremise|opensource

# Default tier for new installs
DEFAULT_TIER=free

# License validation
LICENSE_VALIDATION_ENDPOINT=https://api.projectflow.com/validate
LICENSE_KEY=your-license-key

# Feature flags
FEATURES_ENABLED=team_collaboration,analytics,cloud_sync

# Multi-tenancy
ENABLE_MULTI_TENANCY=true
DEFAULT_STORAGE_TYPE=postgres
```

### Configuration File (projectflow.yaml)
```yaml
mode: saas
default_tier: free
features:
  free:
    - basic_tasks
    - file_storage
    - mcp_integration
    - single_user
  pro:
    - unlimited_tasks
    - postgres_storage
    - team_collaboration
    - oauth_auth
    - cloud_sync
  enterprise:
    - rbac
    - advanced_analytics
    - custom_integrations
    - sso_saml
    - audit_logging

limits:
  free:
    projects: 3
    tasks_per_project: 100
    users: 1
  pro:
    projects: -1  # unlimited
    tasks_per_project: -1
    users: 5
  enterprise:
    projects: -1
    tasks_per_project: -1
    users: -1
```

## API Changes

### Authentication Middleware
```go
func AuthenticationMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Extract tenant from subdomain or header
        tenant := extractTenant(r)
        
        // Validate license and features
        if !tenant.License.IsActive {
            http.Error(w, "License expired", http.StatusForbidden)
            return
        }
        
        // Add tenant context
        ctx := context.WithValue(r.Context(), "tenant", tenant)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

### Feature-Gated Endpoints
```go
func (h *Handler) createTask(w http.ResponseWriter, r *http.Request) {
    tenant := getTenantFromContext(r.Context())
    
    // Check feature access
    if !tenant.HasFeature(FeatureUnlimitedTasks) {
        count, _ := h.storage.CountTasks(tenant.ID)
        if count >= tenant.License.TaskLimit {
            http.Error(w, "Task limit reached. Upgrade to Pro.", http.StatusForbidden)
            return
        }
    }
    
    // Existing logic...
}
```

## MCP Server Enhancements

### Tenant-Aware MCP Tools
```go
func (s *MCPServer) handleCreateTask(args map[string]interface{}) (ToolCallResult, error) {
    // Extract tenant from MCP client context
    tenant := s.getCurrentTenant()
    
    // Check feature access
    if !tenant.HasFeature(FeatureUnlimitedTasks) {
        // Apply limits...
    }
    
    // Existing logic with tenant isolation...
}
```

## Deployment Architecture

### SaaS Deployment
```yaml
# docker-compose.saas.yml
version: '3.8'
services:
  projectflow-app:
    build: .
    environment:
      - PROJECTFLOW_MODE=saas
      - ENABLE_MULTI_TENANCY=true
      - DATABASE_URL=postgresql://...
    depends_on:
      - postgres
      - redis
  
  postgres:
    image: postgres:15
    environment:
      POSTGRES_DB: projectflow_saas
  
  redis:
    image: redis:7-alpine
    # For session management and caching
  
  nginx:
    image: nginx:alpine
    # For SSL termination and tenant routing
```

### On-Premise Enterprise
```yaml
# docker-compose.enterprise.yml
version: '3.8'
services:
  projectflow-app:
    build: .
    environment:
      - PROJECTFLOW_MODE=onpremise
      - DEFAULT_TIER=enterprise
      - ENABLE_MULTI_TENANCY=false
```

## Migration Strategy

### Phase 1: Foundation (4-6 weeks)
1. Add tenant and user models
2. Implement license management
3. Create feature flag system
4. Add authentication middleware
5. Database schema migrations

### Phase 2: Multi-Tenancy (3-4 weeks)
1. Tenant-aware storage layer
2. Project management system
3. User management and invitation
4. Basic RBAC implementation

### Phase 3: SaaS Features (4-5 weeks)
1. Subscription management
2. Billing integration (Stripe)
3. OAuth2 providers
4. Admin dashboard
5. Usage analytics

### Phase 4: Enterprise Features (3-4 weeks)
1. Advanced RBAC
2. SSO/SAML integration
3. Audit logging
4. Advanced reporting
5. Custom integrations framework

## Revenue Model

### Free Tier
- Loss leader to attract users
- Limited features drive upgrades
- Community building and feedback

### Pro Tier ($9/user/month)
- Target: 70% of revenue
- Sweet spot for small teams
- Covers infrastructure costs + profit

### Enterprise Tier ($19/user/month)
- Target: 25% of revenue
- High-margin customers
- Custom features and support

### Additional Revenue Streams
- Professional services (implementation, training)
- Custom integrations development
- Priority support subscriptions
- White-label licensing

## Success Metrics

### Product Metrics
- Monthly Active Users (MAU)
- Trial to paid conversion rate
- Customer lifetime value (CLV)
- Churn rate by tier
- Feature adoption rates

### Technical Metrics
- API response times by tier
- System uptime (99.9% SLA for Enterprise)
- Database performance
- Storage costs per tenant

## Risk Mitigation

### Technical Risks
- **Performance**: Implement caching, database optimization
- **Security**: Regular audits, penetration testing
- **Scalability**: Microservices migration plan if needed

### Business Risks
- **Competition**: Focus on AI integration as differentiator
- **Pricing**: A/B test pricing strategies
- **Customer Support**: Automated support for lower tiers

## Next Steps

1. Create detailed technical specifications
2. Set up development environment for multi-tenancy
3. Implement license management system
4. Build authentication and authorization layer
5. Create migration scripts for existing data
6. Set up CI/CD for multiple deployment targets
7. Develop admin dashboard for tenant management
8. Implement billing and subscription management
9. Create customer onboarding flow
10. Set up monitoring and analytics

This architecture maintains the simplicity and power of the current ProjectFlow while enabling a sustainable SaaS business model through strategic feature gating and tenant isolation.
