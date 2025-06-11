# Configuration Documentation

ProjectFlow supports configuration through environment variables. This document describes all available configuration options.

## Environment Variables

### Server Configuration

#### PORT
- **Description**: The HTTP port the server will listen on
- **Type**: String (representing a valid port number)
- **Default**: `8081`
- **Validation**: Must be a valid port number (1-65535)
- **Example**: `PORT=9000`

#### SHUTDOWN_TIMEOUT
- **Description**: Maximum time in seconds to wait for graceful shutdown
- **Type**: Integer
- **Default**: `30`
- **Validation**: Must be non-negative, recommended maximum is 300 seconds
- **Example**: `SHUTDOWN_TIMEOUT=60`

### Storage Configuration

#### DATA_DIR
- **Description**: Directory path where task data will be stored
- **Type**: String (file system path)
- **Default**: `./data`
- **Validation**: Cannot be empty, directory will be created if it doesn't exist
- **Example**: `DATA_DIR=/app/data` or `DATA_DIR=./projectflow-data`

### Logging Configuration

#### LOG_LEVEL
- **Description**: Minimum log level for application logging
- **Type**: String (case-insensitive)
- **Default**: `INFO`
- **Valid Values**: `DEBUG`, `INFO`, `WARN`, `ERROR`
- **Example**: `LOG_LEVEL=DEBUG`

#### LOG_FORMAT
- **Description**: Format for log output
- **Type**: String (case-insensitive)
- **Default**: `text`
- **Valid Values**: `json`, `text`
- **Example**: `LOG_FORMAT=json`

## Configuration Examples

### Development Environment
```bash
export PORT=8081
export DATA_DIR=./data
export LOG_LEVEL=DEBUG
export LOG_FORMAT=text
export SHUTDOWN_TIMEOUT=30
```

### Production Environment
```bash
export PORT=80
export DATA_DIR=/app/data
export LOG_LEVEL=INFO
export LOG_FORMAT=json
export SHUTDOWN_TIMEOUT=60
```

### Docker Environment
```dockerfile
ENV PORT=8081
ENV DATA_DIR=/app/data
ENV LOG_LEVEL=INFO
ENV LOG_FORMAT=json
ENV SHUTDOWN_TIMEOUT=30
```

## Configuration Validation

The application validates all configuration values on startup:

- **Port validation**: Ensures the port is a valid integer between 1 and 65535
- **Timeout validation**: Ensures shutdown timeout is non-negative and reasonable (≤300s)
- **Directory validation**: Ensures the data directory path is not empty
- **Log level validation**: Ensures log level is one of the supported values
- **Log format validation**: Ensures log format is either "json" or "text"

If any configuration value is invalid, the application will fail to start with a descriptive error message.

## Configuration Logging

On startup, the application logs all configuration values (excluding sensitive data) for operational visibility:

```json
{
  "time": "2025-01-01T12:00:00Z",
  "level": "INFO",
  "msg": "Application configuration loaded",
  "port": "8081",
  "shutdown_timeout_seconds": 30,
  "data_dir": "./data",
  "log_level": "INFO",
  "log_format": "text"
}
```

This helps with troubleshooting configuration issues in deployment environments.
