#!/bin/bash

cd /Users/vanilla/git/aykay76/projectflow

# Start server in background
PORT=8082 go run cmd/server/main.go &
SERVER_PID=$!

echo "Server started with PID: $SERVER_PID"

# Wait for server to start
sleep 2

# Test that server is running
echo "Testing server response..."
curl -s http://localhost:8082/health > /dev/null
if [ $? -eq 0 ]; then
    echo "Server is responding"
else
    echo "Server is not responding"
    kill $SERVER_PID 2>/dev/null
    exit 1
fi

# Send SIGTERM for graceful shutdown
echo "Sending SIGTERM for graceful shutdown..."
kill -TERM $SERVER_PID

# Wait for server to shutdown
wait $SERVER_PID
echo "Server shutdown complete with exit code: $?"
