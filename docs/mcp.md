# Model Context Protocol (MCP) Documentation

## Overview

ProjectFlow's MCP server implements the Model Context Protocol specification, enabling AI agents and assistants to interact with the task management system programmatically. This allows for seamless integration with AI-powered development workflows.

## Architecture

The MCP server is built on JSON-RPC 2.0 and provides:
- **Tools**: Functions that AI agents can call to perform operations
- **Resources**: Data endpoints that agents can query for information
- **Protocol**: Standard MCP communication protocol

## Server Configuration

### Starting the MCP Server

```bash
# Development mode
go run cmd/mcp-server/main.go

# Production build
go build -o bin/mcp-server cmd/mcp-server/main.go
./bin/mcp-server
```

### Environment Variables

- `MCP_PORT`: Server port (default: 3001)
- `STORAGE_DIR`: Data storage directory (default: ./data)

### Client Configuration

For VSCode Cline/Claude Desktop or other MCP clients:

```json
{
  "mcpServers": {
    "projectflow": {
      "command": "go",
      "args": ["run", "cmd/mcp-server/main.go"],
      "cwd": "/path/to/projectflow",
      "env": {
        "STORAGE_DIR": "./data"
      }
    }
  }
}
```

## Available Tools

### 1. list_tasks

List all tasks with optional filtering.

**Parameters:**
- `status` (optional): Filter by task status
- `priority` (optional): Filter by priority

**Example:**
```json
{
  "name": "list_tasks",
  "arguments": {
    "status": "in-progress",
    "priority": "high"
  }
}
```

### 2. create_task

Create a new task.

**Parameters:**
- `title` (required): Task title
- `description` (optional): Task description
- `status` (optional): Initial status (default: "todo")
- `priority` (optional): Task priority (default: "medium")
- `parent_id` (optional): Parent task ID for hierarchy

**Example:**
```json
{
  "name": "create_task",
  "arguments": {
    "title": "Implement user authentication",
    "description": "Add OAuth2 login functionality",
    "status": "todo",
    "priority": "high",
    "parent_id": "epic-123"
  }
}
```

### 3. get_task

Retrieve a specific task by ID.

**Parameters:**
- `id` (required): Task ID

**Example:**
```json
{
  "name": "get_task",
  "arguments": {
    "id": "task-456"
  }
}
```

### 4. update_task

Update an existing task.

**Parameters:**
- `id` (required): Task ID
- `title` (optional): New title
- `description` (optional): New description
- `status` (optional): New status
- `priority` (optional): New priority

**Example:**
```json
{
  "name": "update_task",
  "arguments": {
    "id": "task-456",
    "status": "completed",
    "description": "Updated implementation complete"
  }
}
```

### 5. delete_task

Delete a task.

**Parameters:**
- `id` (required): Task ID

**Example:**
```json
{
  "name": "delete_task",
  "arguments": {
    "id": "task-456"
  }
}
```

### 6. get_task_hierarchy

Get tasks organized in a hierarchical structure.

**Parameters:** None

**Example:**
```json
{
  "name": "get_task_hierarchy",
  "arguments": {}
}
```

## Available Resources

### 1. tasks://all

URI: `tasks://all`

Returns a list of all tasks in the system.

### 2. tasks://hierarchy

URI: `tasks://hierarchy`

Returns tasks organized in a hierarchical structure with parent-child relationships.

### 3. tasks://summary

URI: `tasks://summary`

Returns project statistics and summary information:
- Total task count
- Tasks by status
- Tasks by priority
- Recent activity

## Protocol Details

### JSON-RPC 2.0

All communication uses JSON-RPC 2.0 format:

**Request:**
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "tools/call",
  "params": {
    "name": "create_task",
    "arguments": {
      "title": "New Task",
      "description": "Task description"
    }
  }
}
```

**Response:**
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "content": [
      {
        "type": "text",
        "text": "Task created successfully with ID: abc-123"
      }
    ]
  }
}
```

### Error Handling

Errors follow JSON-RPC 2.0 error format:

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "error": {
    "code": -32602,
    "message": "Invalid params",
    "data": "Task ID is required"
  }
}
```

## Integration Examples

### VSCode with Cline Extension

1. Install the Cline extension in VSCode
2. Configure the MCP server in settings:
   ```json
   {
     "cline.mcpServers": {
       "projectflow": {
         "command": "go",
         "args": ["run", "cmd/mcp-server/main.go"],
         "cwd": "/path/to/projectflow"
       }
     }
   }
   ```
3. Cline can now create and manage tasks during development

### Claude Desktop

1. Edit Claude Desktop configuration:
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
2. Restart Claude Desktop
3. Claude can now interact with your task management system

### Custom AI Agent Integration

```python
import json
import subprocess
import asyncio

class ProjectFlowMCP:
    def __init__(self, projectflow_path):
        self.process = None
        self.projectflow_path = projectflow_path
    
    async def start(self):
        self.process = await asyncio.create_subprocess_exec(
            'go', 'run', 'cmd/mcp-server/main.go',
            cwd=self.projectflow_path,
            stdin=subprocess.PIPE,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE
        )
    
    async def create_task(self, title, description=None):
        request = {
            "jsonrpc": "2.0",
            "id": 1,
            "method": "tools/call",
            "params": {
                "name": "create_task",
                "arguments": {
                    "title": title,
                    "description": description
                }
            }
        }
        
        self.process.stdin.write(json.dumps(request).encode() + b'\n')
        await self.process.stdin.drain()
        
        response = await self.process.stdout.readline()
        return json.loads(response.decode())

# Usage
mcp = ProjectFlowMCP('/path/to/projectflow')
await mcp.start()
result = await mcp.create_task("Implement feature X", "Add new functionality")
```

## Testing

### Unit Tests

Run the MCP server tests:

```bash
go test ./internal/mcp/...
```

### Integration Testing

Test with actual MCP clients:

1. Start the MCP server
2. Use a tool like `mcpx` to test communication:
   ```bash
   echo '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' | nc localhost 3001
   ```

### Mock Client Testing

```bash
# Test tool listing
echo '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' | go run cmd/mcp-server/main.go

# Test task creation
echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"create_task","arguments":{"title":"Test Task"}}}' | go run cmd/mcp-server/main.go
```

## Best Practices

### For AI Agent Developers

1. **Always validate responses**: Check for errors before processing results
2. **Use meaningful task titles**: Help users understand what the AI is doing
3. **Leverage hierarchy**: Use parent-child relationships for complex workflows
4. **Monitor task status**: Update status as work progresses
5. **Clean up**: Delete temporary or test tasks when done

### For System Administrators

1. **Monitor logs**: The MCP server logs all operations
2. **Backup data**: Regularly backup the data directory
3. **Resource limits**: Consider implementing rate limiting for production
4. **Security**: Run in isolated environments for sensitive projects
5. **Updates**: Keep both servers (HTTP and MCP) updated together

## Troubleshooting

### Common Issues

1. **Connection refused**: Ensure MCP server is running on correct port
2. **Permission denied**: Check file system permissions for data directory
3. **Invalid JSON**: Validate request format against JSON-RPC 2.0 spec
4. **Tool not found**: Verify tool name matches available tools
5. **Resource unavailable**: Check if requested resource URI is correct

### Debug Mode

Enable debug logging:

```bash
MCP_DEBUG=true go run cmd/mcp-server/main.go
```

This will output detailed request/response information for troubleshooting.

## Performance Considerations

- **Concurrent requests**: The server handles multiple concurrent requests
- **Memory usage**: Task data is loaded on-demand for better memory efficiency
- **File I/O**: Operations are optimized for the file-based storage system
- **Response caching**: Consider implementing caching for frequently accessed data

## Future Enhancements

Planned improvements to the MCP interface:

1. **Real-time updates**: WebSocket support for live task updates
2. **Batch operations**: Support for bulk task operations
3. **Advanced filtering**: More sophisticated query capabilities
4. **Attachments**: Support for file attachments on tasks
5. **Notifications**: Event-driven notifications for task changes
6. **Authentication**: User-based access control for multi-user scenarios
