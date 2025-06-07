package mcp

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"os"

	"github.com/aykay76/projectflow/internal/storage"
)

// MCPServer represents the Model Context Protocol server
type MCPServer struct {
	storage storage.Storage
	stdin   io.Reader
	stdout  io.Writer
	stderr  io.Writer
}

// NewMCPServer creates a new MCP server instance
func NewMCPServer(storage storage.Storage) *MCPServer {
	return &MCPServer{
		storage: storage,
		stdin:   os.Stdin,
		stdout:  os.Stdout,
		stderr:  os.Stderr,
	}
}

// Start starts the MCP server and handles incoming requests
func (s *MCPServer) Start(ctx context.Context) error {
	log.Printf("Starting MCP server for ProjectFlow")
	
	decoder := json.NewDecoder(s.stdin)
	encoder := json.NewEncoder(s.stdout)
	
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			var request JSONRPCRequest
			if err := decoder.Decode(&request); err != nil {
				if err == io.EOF {
					return nil
				}
				s.sendError(encoder, "", -32700, "Parse error", nil)
				continue
			}
			
			response := s.handleRequest(request)
			if err := encoder.Encode(response); err != nil {
				log.Printf("Error encoding response: %v", err)
				return err
			}
		}
	}
}

// handleRequest processes incoming JSON-RPC requests
func (s *MCPServer) handleRequest(request JSONRPCRequest) JSONRPCResponse {
	switch request.Method {
	case "initialize":
		return s.handleInitialize(request)
	case "tools/list":
		return s.handleToolsList(request)
	case "tools/call":
		return s.handleToolsCall(request)
	case "resources/list":
		return s.handleResourcesList(request)
	case "resources/read":
		return s.handleResourcesRead(request)
	default:
		return s.createErrorResponse(request.ID, -32601, "Method not found", nil)
	}
}

// handleInitialize handles the initialize request
func (s *MCPServer) handleInitialize(request JSONRPCRequest) JSONRPCResponse {
	capabilities := ServerCapabilities{
		Tools: &ToolsCapability{
			ListChanged: false,
		},
		Resources: &ResourcesCapability{
			Subscribe:   false,
			ListChanged: false,
		},
	}
	
	result := InitializeResult{
		ProtocolVersion: "2024-11-05",
		Capabilities:    capabilities,
		ServerInfo: ServerInfo{
			Name:    "projectflow-mcp",
			Version: "1.0.0",
		},
	}
	
	return JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      request.ID,
		Result:  result,
	}
}

// handleToolsList returns the list of available tools
func (s *MCPServer) handleToolsList(request JSONRPCRequest) JSONRPCResponse {
	tools := []Tool{
		{
			Name:        "list_tasks",
			Description: "List all tasks in the project",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
				"required":   []string{},
			},
		},
		{
			Name:        "create_task",
			Description: "Create a new task",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"title": map[string]interface{}{
						"type":        "string",
						"description": "The title of the task",
					},
					"description": map[string]interface{}{
						"type":        "string",
						"description": "The description of the task",
					},
					"status": map[string]interface{}{
						"type":        "string",
						"description": "The status of the task",
						"enum":        []string{"todo", "in_progress", "done", "blocked"},
					},
					"priority": map[string]interface{}{
						"type":        "string",
						"description": "The priority of the task",
						"enum":        []string{"low", "medium", "high", "critical"},
					},
					"type": map[string]interface{}{
						"type":        "string",
						"description": "The type of the task",
						"enum":        []string{"epic", "story", "task", "subtask"},
					},
					"parent_id": map[string]interface{}{
						"type":        "string",
						"description": "The ID of the parent task (for subtasks)",
					},
					"due_date": map[string]interface{}{
						"type":        "string",
						"description": "The due date in YYYY-MM-DD format",
					},
				},
				"required": []string{"title"},
			},
		},
		{
			Name:        "get_task",
			Description: "Get a specific task by ID",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"id": map[string]interface{}{
						"type":        "string",
						"description": "The ID of the task to retrieve",
					},
				},
				"required": []string{"id"},
			},
		},
		{
			Name:        "update_task",
			Description: "Update an existing task",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"id": map[string]interface{}{
						"type":        "string",
						"description": "The ID of the task to update",
					},
					"title": map[string]interface{}{
						"type":        "string",
						"description": "The title of the task",
					},
					"description": map[string]interface{}{
						"type":        "string",
						"description": "The description of the task",
					},
					"status": map[string]interface{}{
						"type":        "string",
						"description": "The status of the task",
						"enum":        []string{"todo", "in_progress", "done", "blocked"},
					},
					"priority": map[string]interface{}{
						"type":        "string",
						"description": "The priority of the task",
						"enum":        []string{"low", "medium", "high", "critical"},
					},
					"type": map[string]interface{}{
						"type":        "string",
						"description": "The type of the task",
						"enum":        []string{"epic", "story", "task", "subtask"},
					},
					"parent_id": map[string]interface{}{
						"type":        "string",
						"description": "The ID of the parent task (for subtasks)",
					},
					"due_date": map[string]interface{}{
						"type":        "string",
						"description": "The due date in YYYY-MM-DD format",
					},
				},
				"required": []string{"id"},
			},
		},
		{
			Name:        "delete_task",
			Description: "Delete a task by ID",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"id": map[string]interface{}{
						"type":        "string",
						"description": "The ID of the task to delete",
					},
				},
				"required": []string{"id"},
			},
		},
		{
			Name:        "get_task_hierarchy",
			Description: "Get the hierarchical structure of all tasks",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
				"required":   []string{},
			},
		},
	}
	
	result := ToolsListResult{
		Tools: tools,
	}
	
	return JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      request.ID,
		Result:  result,
	}
}

// sendError sends an error response
func (s *MCPServer) sendError(encoder *json.Encoder, id interface{}, code int, message string, data interface{}) {
	response := s.createErrorResponse(id, code, message, data)
	if err := encoder.Encode(response); err != nil {
		log.Printf("Error encoding error response: %v", err)
	}
}

// createErrorResponse creates an error response
func (s *MCPServer) createErrorResponse(id interface{}, code int, message string, data interface{}) JSONRPCResponse {
	return JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &JSONRPCError{
			Code:    code,
			Message: message,
			Data:    data,
		},
	}
}
