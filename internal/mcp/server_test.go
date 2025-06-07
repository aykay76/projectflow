package mcp

import (
	"errors"
	"strings"
	"testing"

	"github.com/aykay76/projectflow/internal/models"
)

var ErrTaskNotFound = errors.New("task not found")

// mockStorage implements a simple in-memory storage for testing
type mockStorage struct {
	tasks map[string]*models.Task
}

func newMockStorage() *mockStorage {
	return &mockStorage{
		tasks: make(map[string]*models.Task),
	}
}

func (m *mockStorage) CreateTask(task *models.Task) error {
	m.tasks[task.ID] = task
	return nil
}

func (m *mockStorage) GetTask(id string) (*models.Task, error) {
	task, exists := m.tasks[id]
	if !exists {
		return nil, ErrTaskNotFound
	}
	return task, nil
}

func (m *mockStorage) UpdateTask(task *models.Task) error {
	if _, exists := m.tasks[task.ID]; !exists {
		return ErrTaskNotFound
	}
	m.tasks[task.ID] = task
	return nil
}

func (m *mockStorage) DeleteTask(id string) error {
	if _, exists := m.tasks[id]; !exists {
		return ErrTaskNotFound
	}
	delete(m.tasks, id)
	return nil
}

func (m *mockStorage) ListTasks() ([]*models.Task, error) {
	tasks := make([]*models.Task, 0, len(m.tasks))
	for _, task := range m.tasks {
		tasks = append(tasks, task)
	}
	return tasks, nil
}

func (m *mockStorage) GetTaskChildren(parentID string) ([]*models.Task, error) {
	var children []*models.Task
	for _, task := range m.tasks {
		if task.ParentID == parentID {
			children = append(children, task)
		}
	}
	return children, nil
}

func (m *mockStorage) GetTaskParent(childID string) (*models.Task, error) {
	child, exists := m.tasks[childID]
	if !exists {
		return nil, ErrTaskNotFound
	}
	if child.ParentID == "" {
		return nil, ErrTaskNotFound
	}
	return m.GetTask(child.ParentID)
}

func (m *mockStorage) GetTaskHierarchy() ([]*models.HierarchyTask, error) {
	var rootTasks []*models.HierarchyTask
	for _, task := range m.tasks {
		if task.ParentID == "" {
			hierarchyTask := &models.HierarchyTask{
				Task:       task,
				ChildTasks: []*models.HierarchyTask{},
			}
			rootTasks = append(rootTasks, hierarchyTask)
		}
	}
	return rootTasks, nil
}

func (m *mockStorage) TaskExists(id string) bool {
	_, exists := m.tasks[id]
	return exists
}

func (m *mockStorage) Close() error {
	return nil
}

func TestMCPServer_Initialize(t *testing.T) {
	storage := newMockStorage()
	server := NewMCPServer(storage)
	
	request := JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  "initialize",
		ID:      1,
	}
	
	response := server.handleRequest(request)
	
	if response.Error != nil {
		t.Errorf("Expected no error, got: %v", response.Error)
	}
	
	result, ok := response.Result.(InitializeResult)
	if !ok {
		t.Errorf("Expected InitializeResult, got: %T", response.Result)
	}
	
	if result.ServerInfo.Name != "projectflow-mcp" {
		t.Errorf("Expected server name 'projectflow-mcp', got: %s", result.ServerInfo.Name)
	}
}

func TestMCPServer_ToolsList(t *testing.T) {
	storage := newMockStorage()
	server := NewMCPServer(storage)
	
	request := JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  "tools/list",
		ID:      1,
	}
	
	response := server.handleRequest(request)
	
	if response.Error != nil {
		t.Errorf("Expected no error, got: %v", response.Error)
	}
	
	result, ok := response.Result.(ToolsListResult)
	if !ok {
		t.Errorf("Expected ToolsListResult, got: %T", response.Result)
	}
	
	expectedTools := []string{"list_tasks", "create_task", "get_task", "update_task", "delete_task", "get_task_hierarchy"}
	if len(result.Tools) != len(expectedTools) {
		t.Errorf("Expected %d tools, got %d", len(expectedTools), len(result.Tools))
	}
	
	for i, tool := range result.Tools {
		if tool.Name != expectedTools[i] {
			t.Errorf("Expected tool %s, got %s", expectedTools[i], tool.Name)
		}
	}
}

func TestMCPServer_CreateTask(t *testing.T) {
	storage := newMockStorage()
	server := NewMCPServer(storage)
	
	args := map[string]interface{}{
		"title":       "Test Task",
		"description": "Test Description",
		"status":      "todo",
		"priority":    "medium",
		"type":        "task",
	}
	
	result, err := server.handleCreateTask(args)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	
	if result.IsError {
		t.Errorf("Expected successful result, got error")
	}
	
	if len(result.Content) != 1 {
		t.Errorf("Expected 1 content item, got %d", len(result.Content))
	}
	
	if !strings.Contains(result.Content[0].Text, "Test Task") {
		t.Errorf("Expected content to contain task title")
	}
}

func TestMCPServer_ListTasks(t *testing.T) {
	storage := newMockStorage()
	server := NewMCPServer(storage)
	
	// Create a test task
	task := models.NewTask("Test Task", "Test Description")
	storage.CreateTask(task)
	
	args := map[string]interface{}{}
	result, err := server.handleListTasks(args)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	
	if result.IsError {
		t.Errorf("Expected successful result, got error")
	}
	
	if len(result.Content) != 1 {
		t.Errorf("Expected 1 content item, got %d", len(result.Content))
	}
	
	if !strings.Contains(result.Content[0].Text, "Found 1 tasks") {
		t.Errorf("Expected content to show 1 task")
	}
}

func TestMCPServer_InvalidToolCall(t *testing.T) {
	storage := newMockStorage()
	server := NewMCPServer(storage)
	
	request := JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  "tools/call",
		Params: map[string]interface{}{
			"name":      "invalid_tool",
			"arguments": map[string]interface{}{},
		},
		ID: 1,
	}
	
	response := server.handleRequest(request)
	
	if response.Error == nil {
		t.Errorf("Expected error for invalid tool call")
	}
	
	if response.Error.Code != -32601 {
		t.Errorf("Expected error code -32601, got %d", response.Error.Code)
	}
}
