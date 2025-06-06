# ProjectFlow

A workflow management system for AI-assisted development, similar to Jira or Azure DevOps. Supports both API-driven interactions and Model Context Protocol for seamless AI agent integration.

## Features

- Hierarchical task management (Epics, Stories, Subtasks)
- REST API for programmatic access
- Model Context Protocol (MCP) support for AI agents
- Web interface for human users
- File system-based storage (adaptable architecture)
- Clean, modern UI
- Containerized deployment

## Tech Stack

- **Backend**: Go 1.24
- **Storage**: File system (JSON) with adaptable interface
- **Frontend**: HTML templates, CSS, JavaScript
- **Containerization**: Docker/Podman
- **Protocols**: HTTP REST API + Model Context Protocol

## Quick Start

### Prerequisites

- Go 1.24 or later
- Docker/Podman (for containerized deployment)

### Running Locally

1. Clone the repository:
   ```bash
   git clone https://github.com/aykay76/projectflow.git
   cd projectflow
   ```

2. Run the application:
   ```bash
   go run cmd/server/main.go
   ```

3. Open your browser and navigate to `http://localhost:8080`

### Environment Variables

- `PORT`: Server port (default: 8080)
- `STORAGE_DIR`: Directory for data storage (default: ./data)

### Using Docker

1. Build the image:
   ```bash
   docker build -t projectflow .
   ```

2. Run the container:
   ```bash
   docker run -p 8080:8080 -v $(pwd)/data:/app/data projectflow
   ```

## API Documentation

### Tasks API

- `GET /api/tasks` - List all tasks
- `POST /api/tasks` - Create a new task
- `GET /api/tasks/{id}` - Get task by ID
- `PUT /api/tasks/{id}` - Update task
- `DELETE /api/tasks/{id}` - Delete task
- `GET /api/hierarchy` - Get tasks in hierarchical structure

### Task Structure

```json
{
  "id": "string",
  "title": "string",
  "description": "string",
  "status": "string",
  "priority": "string",
  "parent_id": "string",
  "children": ["string"],
  "created_at": "timestamp",
  "updated_at": "timestamp"
}
```

### Hierarchy Structure

The `/api/hierarchy` endpoint returns tasks in a nested structure:

```json
[
  {
    "task": {
      "id": "string",
      "title": "string",
      "description": "string",
      "status": "string",
      "priority": "string",
      "type": "string",
      "parent_id": "string",
      "children": ["string"],
      "created_at": "timestamp",
      "updated_at": "timestamp"
    },
    "child_tasks": [
      {
        "task": { /* nested task */ },
        "child_tasks": [ /* recursively nested */ ]
      }
    ]
  }
]
```

## Development

### Project Structure

```
├── cmd/server/          # Application entry point
├── internal/
│   ├── handlers/        # HTTP handlers
│   ├── models/          # Data models
│   └── storage/         # Storage implementations
├── pkg/api/            # Public API definitions
├── web/
│   ├── templates/      # HTML templates
│   └── static/         # CSS, JS, images
├── data/               # Local data storage
└── Dockerfile          # Container definition
```

### Running Tests

```bash
go test ./...
```

### Building

```bash
go build -o bin/projectflow cmd/server/main.go
```

## Model Context Protocol (MCP)

ProjectFlow supports MCP for AI agent integration. See the [MCP documentation](docs/mcp.md) for detailed information on:

- Setting up MCP server
- Available MCP tools and resources
- AI agent integration examples

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes with proper tests
4. Submit a pull request

## License

MIT License - see [LICENSE](LICENSE) file for details.
