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

ProjectFlow includes a Model Context Protocol (MCP) server that enables AI agents to interact with tasks programmatically. This allows AI assistants to create, read, update, and delete tasks as part of their workflow.

### MCP Server Setup

1. **Start the MCP server:**
   ```bash
   go run cmd/mcp-server/main.go
   ```
   The MCP server runs on port 3001 by default.

2. **Configure your MCP client:**
   Use the provided `mcp-config.json` file or configure manually:
   ```json
   {
     "mcpServers": {
       "projectflow": {
         "command": "go",
         "args": ["run", "cmd/mcp-server/main.go"],
         "cwd": "/path/to/projectflow"
       }
     }
   }
   ```

### Available MCP Tools

The MCP server provides these tools for task management:

- **`list_tasks`** - List all tasks with optional filtering
- **`create_task`** - Create a new task
- **`get_task`** - Get a specific task by ID
- **`update_task`** - Update an existing task
- **`delete_task`** - Delete a task
- **`get_task_hierarchy`** - Get tasks in hierarchical structure

### Available MCP Resources

The MCP server exposes these resources:

- **`tasks://all`** - List of all tasks
- **`tasks://hierarchy`** - Hierarchical task structure
- **`tasks://summary`** - Project summary with statistics

### Example Usage

```bash
# Start both servers
go run cmd/server/main.go &          # HTTP server on :8080
go run cmd/mcp-server/main.go &      # MCP server on :3001

# Use with MCP-compatible AI clients
# The AI can now create, manage, and query tasks programmatically
```

### Integration with AI Agents

AI agents can use the MCP interface to:
- Create and manage development tasks
- Track project progress
- Generate reports and summaries
- Automate workflow processes
- Integrate with other development tools

For detailed MCP documentation, see [docs/mcp.md](docs/mcp.md).

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes with proper tests
4. Submit a pull request

## License

MIT License - see [LICENSE](LICENSE) file for details.
