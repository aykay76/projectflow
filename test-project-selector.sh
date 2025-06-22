#!/bin/bash

echo "Testing Project Selector Functionality"
echo "====================================="

# Start the server in the background
echo "Starting server..."
cd /Users/vanilla/git/aykay76/projectflow
go run cmd/server/main.go &
SERVER_PID=$!

# Wait for server to start
sleep 3

echo "Testing API endpoints..."

# Test projects API
echo "1. Testing /api/projects"
curl -s http://localhost:16191/api/projects | jq -r '.[] | "Project: \(.name) (ID: \(.id))"'

echo ""
echo "2. Testing individual project endpoints"
# Get project IDs
PROJECT1=$(curl -s http://localhost:16191/api/projects | jq -r '.[0].id')
PROJECT2=$(curl -s http://localhost:16191/api/projects | jq -r '.[1].id')

echo "Testing project 1: $PROJECT1"
curl -s http://localhost:16191/api/projects/$PROJECT1 | jq -r '.name'

echo "Testing project 2: $PROJECT2"
curl -s http://localhost:16191/api/projects/$PROJECT2 | jq -r '.name'

echo ""
echo "3. Testing tasks for each project"
echo "Tasks for project 1:"
curl -s "http://localhost:16191/api/tasks?project_id=$PROJECT1" | jq -r 'length'

echo "Tasks for project 2:"
curl -s "http://localhost:16191/api/tasks?project_id=$PROJECT2" | jq -r 'length'

echo ""
echo "4. Opening browser to test UI..."
open http://localhost:16191

echo ""
echo "Manual Test Instructions:"
echo "1. Look at the header - should show 'Current: [ProjectName]'"
echo "2. Click the gear icon (⚙️) in the top right"
echo "3. Click the 'Select Project' button"
echo "4. Verify both projects appear in the dropdown"
echo "5. Select a different project"
echo "6. Verify the header updates to show the new project"
echo ""
echo "Press Enter to stop the server..."
read

# Kill the server
kill $SERVER_PID
echo "Server stopped."
