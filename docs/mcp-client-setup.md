# MCP Client Setup Guide

This guide explains how to configure various MCP clients to use the ProjectFlow MCP server.

## Prerequisites

1. Ensure the MCP server is built:
   ```bash
   cd /Users/vanilla/git/aykay76/projectflow
   go build -o bin/mcp-server cmd/mcp-server/main.go
   ```

2. Verify the server runs correctly:
   ```bash
   ./bin/mcp-server
   ```

## VSCode Configuration

### Option 1: Using Cline Extension (Recommended)

If you're using the Cline extension (formerly Claude Dev), add the following to your VSCode settings:

1. Open VSCode Settings (Cmd+,)
2. Search for "cline" or "mcp"
3. Add the MCP server configuration:

```json
{
  "cline.mcpServers": {
    "projectflow": {
      "command": "/Users/vanilla/git/aykay76/projectflow/bin/mcp-server",
      "args": [],
      "cwd": "/Users/vanilla/git/aykay76/projectflow",
      "env": {
        "STORAGE_DIR": "./data"
      }
    }
  }
}
```

### Option 2: Using Continue Extension

If you're using the Continue extension:

1. Open your Continue configuration file (`~/.continue/config.json`)
2. Add the MCP server configuration:

```json
{
  "mcpServers": [
    {
      "name": "projectflow",
      "command": "/Users/vanilla/git/aykay76/projectflow/bin/mcp-server",
      "args": [],
      "cwd": "/Users/vanilla/git/aykay76/projectflow",
      "env": {
        "STORAGE_DIR": "./data"
      }
    }
  ]
}
```

### Option 3: Direct MCP Extension

If you're using a dedicated MCP extension, create or update your MCP configuration file:

**Global Configuration** (`~/.mcp/config.json`):
```json
{
  "mcpServers": {
    "projectflow": {
      "command": "/Users/vanilla/git/aykay76/projectflow/bin/mcp-server",
      "args": [],
      "cwd": "/Users/vanilla/git/aykay76/projectflow",
      "env": {
        "STORAGE_DIR": "./data"
      }
    }
  }
}
```

**Project-specific Configuration** (`.vscode/mcp.json`):
```json
{
  "mcpServers": {
    "projectflow": {
      "command": "./bin/mcp-server",
      "args": [],
      "cwd": ".",
      "env": {
        "STORAGE_DIR": "./data"
      }
    }
  }
}
```

## Claude Desktop Configuration

To use the MCP server with Claude Desktop:

1. Locate your Claude Desktop configuration file:
   - macOS: `~/Library/Application Support/Claude/claude_desktop_config.json`

2. Add the server configuration:

```json
{
  "mcpServers": {
    "projectflow": {
      "command": "/Users/vanilla/git/aykay76/projectflow/bin/mcp-server",
      "args": [],
      "cwd": "/Users/vanilla/git/aykay76/projectflow",
      "env": {
        "STORAGE_DIR": "./data"
      }
    }
  }
}
```

3. Restart Claude Desktop

## Other MCP Clients

### Generic MCP Client Configuration

For any MCP client that supports configuration files, use this template:

```json
{
  "mcpServers": {
    "projectflow": {
      "command": "/Users/vanilla/git/aykay76/projectflow/bin/mcp-server",
      "args": [],
      "cwd": "/Users/vanilla/git/aykay76/projectflow",
      "env": {
        "STORAGE_DIR": "./data"
      }
    }
  }
}
```

### Environment Variables

You can customize the server behavior using these environment variables:

- `STORAGE_DIR`: Directory for task data storage (default: `./data`)
- `MCP_PORT`: Server port if running in server mode (default: 3001)

## Available Tools and Resources

Once configured, the MCP client will have access to these tools:

### Tools (Functions the AI can call):
- `create_task`: Create a new task
- `update_task`: Update an existing task
- `list_tasks`: List tasks with optional filtering
- `get_task`: Get details of a specific task
- `delete_task`: Delete a task

### Resources (Data the AI can query):
- `task://tasks`: List of all tasks
- `task://task/{id}`: Individual task details

## Testing the Configuration

1. Start your MCP client (VSCode with extension, Claude Desktop, etc.)
2. Ask the AI assistant to list tasks: "Can you show me all the tasks in ProjectFlow?"
3. Try creating a task: "Create a new task called 'Test MCP Integration' with description 'Testing the MCP server connection'"

## Troubleshooting

### Common Issues:

1. **Permission Denied**: Ensure the binary is executable:
   ```bash
   chmod +x /Users/vanilla/git/aykay76/projectflow/bin/mcp-server
   ```

2. **Path Issues**: Use absolute paths in configuration files

3. **Port Conflicts**: If running multiple MCP servers, ensure they use different ports

4. **Storage Directory**: Ensure the `STORAGE_DIR` exists and is writable:
   ```bash
   mkdir -p /Users/vanilla/git/aykay76/projectflow/data/tasks
   ```

### Debug Mode

To run the MCP server with debug logging:

```bash
cd /Users/vanilla/git/aykay76/projectflow
MCP_DEBUG=true ./bin/mcp-server
```

## Development Mode

For development, you can use the Go source directly:

```json
{
  "mcpServers": {
    "projectflow": {
      "command": "go",
      "args": ["run", "cmd/mcp-server/main.go"],
      "cwd": "/Users/vanilla/git/aykay76/projectflow",
      "env": {
        "STORAGE_DIR": "./data"
      }
    }
  }
}
```

This is useful when actively developing the MCP server as it will automatically use the latest code changes.
