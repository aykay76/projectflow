#!/bin/bash

# MCP Server Test Script
# This script tests the basic functionality of the ProjectFlow MCP server

set -e

echo "🧪 Testing ProjectFlow MCP Server..."

# Check if binary exists
if [ ! -f "./bin/mcp-server" ]; then
    echo "❌ MCP server binary not found. Building..."
    go build -o bin/mcp-server cmd/mcp-server/main.go
    echo "✅ MCP server binary built successfully"
fi

# Make sure binary is executable
chmod +x ./bin/mcp-server

# Check if data directory exists
if [ ! -d "./data" ]; then
    echo "📁 Creating data directory..."
    mkdir -p ./data/projects
    echo "✅ Data directory created"
fi

echo "🚀 Starting MCP server for testing..."

# Test if the server starts without errors
timeout 5s ./bin/mcp-server > /dev/null 2>&1 || {
    if [ $? -eq 124 ]; then
        echo "✅ MCP server started successfully (timed out as expected)"
    else
        echo "❌ MCP server failed to start"
        exit 1
    fi
}

echo "📋 MCP Server Configuration Summary:"
echo "   Binary Path: $(pwd)/bin/mcp-server"
echo "   Working Directory: $(pwd)"
echo "   Storage Directory: ./data"
echo "   Projects Directory: ./data/projects"

# Count existing tasks across all projects
TASK_COUNT=$(find ./data/projects -name "*.json" -path "*/tasks/*" 2>/dev/null | wc -l | tr -d ' ')
echo "   Existing Tasks: $TASK_COUNT"

echo ""
echo "🎉 MCP Server is ready for use!"
echo ""
echo "📖 Next Steps:"
echo "   1. Configure your MCP client using the settings in docs/mcp-client-setup.md"
echo "   2. Use this configuration in your MCP client:"
echo ""
echo "   {\"mcpServers\": {"
echo "     \"projectflow\": {"
echo "       \"command\": \"$(pwd)/bin/mcp-server\","
echo "       \"args\": [],"
echo "       \"cwd\": \"$(pwd)\","
echo "       \"env\": {\"STORAGE_DIR\": \"./data\"}"
echo "     }"
echo "   }}"
echo ""
echo "   3. Test by asking your AI assistant to list tasks: 'Show me all ProjectFlow tasks'"
