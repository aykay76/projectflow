# PostgreSQL Storage Configuration

ProjectFlow now supports PostgreSQL as an alternative to file-based storage. This guide explains how to configure and use the PostgreSQL storage backend.

## Environment Variables

Configure PostgreSQL storage using these environment variables:

### Storage Type Selection
```bash
# Use PostgreSQL storage (default: "file")
STORAGE_TYPE=postgres
# Alternative format also supported
STORAGE_TYPE=postgresql
```

### Database Connection
```bash
# Database connection parameters
DB_HOST=localhost           # Database host (default: "localhost")
DB_PORT=5432               # Database port (default: "5432")
DB_NAME=projectflow        # Database name (default: "projectflow")
DB_USER=projectflow        # Database user (default: "projectflow")
DB_PASSWORD=your_password  # Database password (no default)
DB_SSL_MODE=prefer         # SSL mode (default: "prefer")
```

### SSL Mode Options
- `disable` - No SSL
- `require` - Require SSL (no verification)
- `verify-ca` - Require SSL and verify CA
- `verify-full` - Require SSL and verify hostname
- `prefer` - Try SSL first, fallback to non-SSL
- `allow` - Try non-SSL first, fallback to SSL

## Example Configurations

### Development (Local PostgreSQL)
```bash
export STORAGE_TYPE=postgres
export DB_HOST=localhost
export DB_PORT=5432
export DB_NAME=projectflow_dev
export DB_USER=developer
export DB_PASSWORD=devpassword
export DB_SSL_MODE=disable
```

### Production (Secure PostgreSQL)
```bash
export STORAGE_TYPE=postgres
export DB_HOST=prod-db.example.com
export DB_PORT=5432
export DB_NAME=projectflow_prod
export DB_USER=projectflow_app
export DB_PASSWORD=secure_production_password
export DB_SSL_MODE=require
```

### Docker Compose Example
```yaml
version: '3.8'
services:
  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_DB: projectflow
      POSTGRES_USER: projectflow
      POSTGRES_PASSWORD: projectflow123
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data

  projectflow:
    build: .
    environment:
      STORAGE_TYPE: postgres
      DB_HOST: postgres
      DB_PORT: 5432
      DB_NAME: projectflow
      DB_USER: projectflow
      DB_PASSWORD: projectflow123
      DB_SSL_MODE: disable
    ports:
      - "8080:8080"
    depends_on:
      - postgres

volumes:
  postgres_data:
```

## Database Schema

The PostgreSQL storage automatically creates the required schema on startup:

### Tasks Table
```sql
CREATE TABLE tasks (
    id VARCHAR(36) PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'todo',
    priority VARCHAR(20) NOT NULL DEFAULT 'medium',
    type VARCHAR(20) NOT NULL DEFAULT 'task',
    parent_id VARCHAR(36),
    children JSONB DEFAULT '[]'::jsonb,
    started_at TIMESTAMPTZ,
    due_date TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    FOREIGN KEY (parent_id) REFERENCES tasks(id) ON DELETE SET NULL
);
```

### Indexes
The following indexes are automatically created for optimal performance:
- `idx_tasks_status` - Status queries
- `idx_tasks_priority` - Priority filtering
- `idx_tasks_type` - Task type filtering
- `idx_tasks_parent_id` - Hierarchical queries
- `idx_tasks_created_at` - Creation date sorting
- `idx_tasks_due_date` - Due date queries

## Migration from File Storage

To migrate from file storage to PostgreSQL:

1. **Backup your data**: Copy the `./data` directory
2. **Set up PostgreSQL**: Create database and user
3. **Configure environment**: Set the PostgreSQL environment variables
4. **Start the application**: The schema will be created automatically
5. **Import data** (if needed): Use the migration tools or recreate tasks via API

## Features

### Automatic Schema Management
- Database tables and indexes are created automatically
- No manual schema setup required
- Safe to run multiple times

### Transaction Support
- All operations use database transactions
- Automatic rollback on failures
- Consistent parent-child relationships

### Hierarchical Relationships
- Foreign key constraints ensure data integrity
- Cascade deletes for parent-child relationships
- Efficient queries for hierarchical data

### Concurrent Access
- Thread-safe operations with proper locking
- Supports multiple application instances
- Database-level concurrency control

## Performance Considerations

### Connection Pooling
The PostgreSQL driver automatically manages connection pooling. For high-load scenarios, consider:
- Increasing `max_connections` in PostgreSQL
- Tuning connection pool settings
- Using a connection pooler like PgBouncer

### Indexing
All critical queries are indexed for optimal performance:
- Fast lookups by ID (primary key)
- Efficient filtering by status, priority, type
- Quick hierarchical queries
- Optimized date-based queries

### JSON Storage
Child task IDs are stored as JSONB for:
- Efficient array operations
- Native PostgreSQL JSON support
- Atomic updates of child relationships

## Troubleshooting

### Connection Issues
```bash
# Test connection manually
psql "host=localhost port=5432 dbname=projectflow user=projectflow password=yourpassword sslmode=disable"
```

### Common Errors

**"failed to ping database"**
- Check database server is running
- Verify connection parameters
- Check network connectivity
- Validate SSL configuration

**"permission denied"**
- Verify database user has necessary permissions
- Ensure user can create tables and indexes
- Check database ownership

**"database does not exist"**
- Create the database: `CREATE DATABASE projectflow;`
- Verify `DB_NAME` environment variable

### Logging
Enable debug logging to troubleshoot issues:
```bash
export LOG_LEVEL=DEBUG
```

## Backup and Recovery

### Backup
```bash
# Full database backup
pg_dump "host=localhost port=5432 dbname=projectflow user=projectflow" > projectflow_backup.sql

# Schema only
pg_dump --schema-only "host=localhost port=5432 dbname=projectflow user=projectflow" > schema.sql

# Data only
pg_dump --data-only "host=localhost port=5432 dbname=projectflow user=projectflow" > data.sql
```

### Recovery
```bash
# Restore full backup
psql "host=localhost port=5432 dbname=projectflow user=projectflow" < projectflow_backup.sql
```

## Security Best Practices

1. **Use strong passwords**: Generate secure database passwords
2. **Enable SSL**: Use `require` or higher SSL modes in production
3. **Network security**: Restrict database access to application servers
4. **User privileges**: Create dedicated database user with minimal permissions
5. **Regular backups**: Implement automated backup procedures
6. **Monitor access**: Enable PostgreSQL logging for security audits
