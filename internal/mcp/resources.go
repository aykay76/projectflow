package mcp

import (
	"encoding/json"
	"fmt"
)

// handleResourcesList handles the resources/list request
func (s *MCPServer) handleResourcesList(request JSONRPCRequest) JSONRPCResponse {
	resources := []Resource{
		{
			URI:         "projectflow://tasks",
			Name:        "All Tasks",
			Description: "Complete list of all tasks in the project",
			MimeType:    "application/json",
		},
		{
			URI:         "projectflow://hierarchy",
			Name:        "Task Hierarchy",
			Description: "Hierarchical view of all tasks organized by parent-child relationships",
			MimeType:    "application/json",
		},
		{
			URI:         "projectflow://summary",
			Name:        "Project Summary",
			Description: "Summary statistics and overview of the project status",
			MimeType:    "text/plain",
		},
	}

	result := ResourcesListResult{
		Resources: resources,
	}

	return JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      request.ID,
		Result:  result,
	}
}

// handleResourcesRead handles the resources/read request
func (s *MCPServer) handleResourcesRead(request JSONRPCRequest) JSONRPCResponse {
	var readReq ResourceReadRequest

	// Parse the params
	paramsBytes, err := json.Marshal(request.Params)
	if err != nil {
		return s.createErrorResponse(request.ID, -32602, "Invalid params", nil)
	}

	if err := json.Unmarshal(paramsBytes, &readReq); err != nil {
		return s.createErrorResponse(request.ID, -32602, "Invalid params", nil)
	}

	var contents []Content
	var readErr error

	switch readReq.URI {
	case "projectflow://tasks":
		contents, readErr = s.readTasksResource()
	case "projectflow://hierarchy":
		contents, readErr = s.readHierarchyResource()
	case "projectflow://summary":
		contents, readErr = s.readSummaryResource()
	default:
		return s.createErrorResponse(request.ID, -32602, "Unknown resource URI", nil)
	}

	if readErr != nil {
		return s.createErrorResponse(request.ID, -32603, readErr.Error(), nil)
	}

	result := ResourceReadResult{
		Contents: contents,
	}

	return JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      request.ID,
		Result:  result,
	}
}

// readTasksResource reads the tasks resource
func (s *MCPServer) readTasksResource() ([]Content, error) {
	tasks, err := s.storage.ListTasks()
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}

	tasksJSON, err := json.MarshalIndent(tasks, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal tasks: %w", err)
	}

	return []Content{{
		Type: "text",
		Text: string(tasksJSON),
	}}, nil
}

// readHierarchyResource reads the hierarchy resource
func (s *MCPServer) readHierarchyResource() ([]Content, error) {
	hierarchy, err := s.storage.GetTaskHierarchy()
	if err != nil {
		return nil, fmt.Errorf("failed to get hierarchy: %w", err)
	}

	hierarchyJSON, err := json.MarshalIndent(hierarchy, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal hierarchy: %w", err)
	}

	return []Content{{
		Type: "text",
		Text: string(hierarchyJSON),
	}}, nil
}

// readSummaryResource reads the summary resource
func (s *MCPServer) readSummaryResource() ([]Content, error) {
	tasks, err := s.storage.ListTasks()
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}

	// Calculate statistics
	stats := map[string]int{
		"total":       len(tasks),
		"todo":        0,
		"in_progress": 0,
		"done":        0,
		"blocked":     0,
		"epic":        0,
		"story":       0,
		"task":        0,
		"subtask":     0,
		"low":         0,
		"medium":      0,
		"high":        0,
		"critical":    0,
		"overdue":     0,
	}

	for _, task := range tasks {
		// Count by status
		switch task.Status {
		case "todo":
			stats["todo"]++
		case "in_progress":
			stats["in_progress"]++
		case "done":
			stats["done"]++
		case "blocked":
			stats["blocked"]++
		}

		// Count by type
		switch task.Type {
		case "epic":
			stats["epic"]++
		case "story":
			stats["story"]++
		case "task":
			stats["task"]++
		case "subtask":
			stats["subtask"]++
		}

		// Count by priority
		switch task.Priority {
		case "low":
			stats["low"]++
		case "medium":
			stats["medium"]++
		case "high":
			stats["high"]++
		case "critical":
			stats["critical"]++
		}

		// Count overdue
		if task.IsOverdue() {
			stats["overdue"]++
		}
	}

	summary := fmt.Sprintf(`ProjectFlow Summary
==================

Total Tasks: %d

Status Breakdown:
- To Do: %d
- In Progress: %d
- Done: %d
- Blocked: %d

Type Breakdown:
- Epics: %d
- Stories: %d
- Tasks: %d
- Subtasks: %d

Priority Breakdown:
- Low: %d
- Medium: %d
- High: %d
- Critical: %d

Other:
- Overdue: %d

Progress: %.1f%% complete
`,
		stats["total"],
		stats["todo"],
		stats["in_progress"],
		stats["done"],
		stats["blocked"],
		stats["epic"],
		stats["story"],
		stats["task"],
		stats["subtask"],
		stats["low"],
		stats["medium"],
		stats["high"],
		stats["critical"],
		stats["overdue"],
		float64(stats["done"])/float64(stats["total"])*100,
	)

	return []Content{{
		Type: "text",
		Text: summary,
	}}, nil
}
