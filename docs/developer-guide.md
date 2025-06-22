# 🔧 Developer Guide

This guide is for developers who want to extend, customize, or contribute to ProjectFlow's natural language chat interface.

## Architecture Overview

The chat interface is built with a modular architecture that separates concerns:

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Frontend      │    │   Backend       │    │   LLM Layer     │
│                 │    │                 │    │                 │
│ ┌─────────────┐ │    │ ┌─────────────┐ │    │ ┌─────────────┐ │
│ │ChatManager  │◄┼────┼►│Chat Handler │◄┼────┼►│LLM Provider │ │
│ │             │ │    │ │             │ │    │ │             │ │
│ └─────────────┘ │    │ └─────────────┘ │    │ └─────────────┘ │
│                 │    │        │        │    │                 │
│ ┌─────────────┐ │    │ ┌─────────────┐ │    │ ┌─────────────┐ │
│ │API Client   │◄┼────┼►│Translator   │◄┼────┼►│Translation  │ │
│ │             │ │    │ │             │ │    │ │Engine       │ │
│ └─────────────┘ │    │ └─────────────┘ │    │ └─────────────┘ │
│                 │    │        │        │    │                 │
│                 │    │ ┌─────────────┐ │    │                 │
│                 │    │ │MCP Commands │ │    │                 │
│                 │    │ │             │ │    │                 │
│                 │    │ └─────────────┘ │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

### Key Components

1. **Frontend (JavaScript)**
   - `ChatManager`: UI management and user interactions
   - `ApiClient`: HTTP communication with backend
   - CSS components for styling

2. **Backend (Go)**
   - `Chat Handlers`: REST API endpoints
   - `Translator`: Natural language to MCP command conversion
   - `LLM Service`: Provider abstraction layer

3. **LLM Layer**
   - Provider implementations (Groq, Ollama, OpenAI)
   - Translation engine with prompt templates
   - Response validation and formatting

## Frontend Development

### ChatManager Extension

The `ChatManager` class is designed for extensibility:

```javascript
// web/static/js/chat-manager.js
export class ChatManager {
    constructor(apiClient, notificationManager) {
        this.apiClient = apiClient;
        this.notificationManager = notificationManager;
        // Add your custom properties
        this.customHandlers = new Map();
    }

    // Add custom message handlers
    addCustomHandler(type, handler) {
        this.customHandlers.set(type, handler);
    }

    // Override or extend message processing
    async processMessage(message) {
        // Custom preprocessing
        message = this.preprocessMessage(message);
        
        // Call original logic
        return await this.sendMessage(message);
    }
}
```

### Adding Custom UI Components

```javascript
// Create custom chat components
class CustomChatComponent {
    constructor(chatManager) {
        this.chatManager = chatManager;
        this.init();
    }

    init() {
        // Add custom buttons or panels
        const customButton = document.createElement('button');
        customButton.textContent = 'Quick Actions';
        customButton.onclick = () => this.showQuickActions();
        
        // Insert into chat header
        const chatHeader = document.querySelector('.chat-header-actions');
        chatHeader.appendChild(customButton);
    }

    showQuickActions() {
        // Implement custom functionality
        const actions = [
            'Show my tasks',
            'Create daily standup',
            'List overdue items'
        ];
        
        actions.forEach(action => {
            this.chatManager.sendMessage(action);
        });
    }
}

// Initialize custom component
document.addEventListener('DOMContentLoaded', () => {
    const chatManager = window.projectFlowApp.chatManager;
    new CustomChatComponent(chatManager);
});
```

### Custom CSS Themes

```css
/* web/static/css/themes/custom-theme.css */
:root[data-theme="custom"] {
    /* Override chat colors */
    --chat-bg: #1a1a2e;
    --chat-panel-bg: #16213e;
    --chat-message-user-bg: #0f3460;
    --chat-message-assistant-bg: #533483;
    
    /* Custom animations */
    --chat-animation-duration: 0.3s;
    --chat-animation-easing: cubic-bezier(0.4, 0, 0.2, 1);
}

.chat-panel.custom-theme {
    border-radius: 20px;
    box-shadow: 0 10px 40px rgba(0, 0, 0, 0.3);
}

.chat-message.custom-style {
    border-radius: 18px;
    padding: 12px 16px;
    margin: 8px 0;
}
```

### WebSocket Support (Future Enhancement)

```javascript
// web/static/js/websocket-chat.js
class WebSocketChatManager extends ChatManager {
    constructor(apiClient, notificationManager) {
        super(apiClient, notificationManager);
        this.wsConnection = null;
        this.initWebSocket();
    }

    initWebSocket() {
        const wsUrl = `ws://${window.location.host}/ws/chat`;
        this.wsConnection = new WebSocket(wsUrl);
        
        this.wsConnection.onmessage = (event) => {
            const data = JSON.parse(event.data);
            this.handleRealtimeMessage(data);
        };
    }

    async sendMessage(message) {
        if (this.wsConnection.readyState === WebSocket.OPEN) {
            // Send via WebSocket for real-time response
            this.wsConnection.send(JSON.stringify({
                type: 'chat_message',
                message: message,
                conversation_id: this.currentConversationId
            }));
        } else {
            // Fallback to HTTP
            return super.sendMessage(message);
        }
    }
}
```

## Backend Development

### Adding New LLM Providers

Implement the `LLMProvider` interface:

```go
// internal/llm/custom_provider.go
package llm

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
)

type CustomProvider struct {
    apiKey    string
    baseURL   string
    model     string
    timeout   time.Duration
    client    *http.Client
}

func NewCustomProvider(config *config.Config) *CustomProvider {
    return &CustomProvider{
        apiKey:  config.LLMAPIKey,
        baseURL: config.LLMBaseURL,
        model:   config.LLMModel,
        timeout: time.Duration(config.LLMTimeout) * time.Second,
        client:  &http.Client{Timeout: time.Duration(config.LLMTimeout) * time.Second},
    }
}

func (p *CustomProvider) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
    // Implement your provider's API call
    payload := map[string]interface{}{
        "model":      p.model,
        "messages":   req.Messages,
        "max_tokens": req.MaxTokens,
    }

    jsonData, err := json.Marshal(payload)
    if err != nil {
        return nil, fmt.Errorf("failed to marshal request: %w", err)
    }

    httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/chat/completions", bytes.NewReader(jsonData))
    if err != nil {
        return nil, fmt.Errorf("failed to create request: %w", err)
    }

    httpReq.Header.Set("Content-Type", "application/json")
    httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)

    resp, err := p.client.Do(httpReq)
    if err != nil {
        return nil, fmt.Errorf("request failed: %w", err)
    }
    defer resp.Body.Close()

    // Parse response and return ChatResponse
    var result struct {
        Choices []struct {
            Message struct {
                Content string `json:"content"`
            } `json:"message"`
        } `json:"choices"`
    }

    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, fmt.Errorf("failed to decode response: %w", err)
    }

    if len(result.Choices) == 0 {
        return nil, fmt.Errorf("no response from provider")
    }

    return &ChatResponse{
        Content: result.Choices[0].Message.Content,
        Model:   p.model,
        Usage: Usage{
            TotalTokens: 0, // Implement token counting if available
        },
    }, nil
}

func (p *CustomProvider) HealthCheck(ctx context.Context) error {
    // Implement health check logic
    req, err := http.NewRequestWithContext(ctx, "GET", p.baseURL+"/health", nil)
    if err != nil {
        return err
    }

    resp, err := p.client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("health check failed with status: %d", resp.StatusCode)
    }

    return nil
}

func (p *CustomProvider) Name() string {
    return "custom"
}

func (p *CustomProvider) Model() string {
    return p.model
}
```

Register your provider in the factory:

```go
// internal/llm/factory.go
func CreateProvider(config *config.Config) (LLMProvider, error) {
    switch strings.ToLower(config.LLMProvider) {
    case "custom":
        return NewCustomProvider(config), nil
    case "groq":
        return NewGroqProvider(config), nil
    // ... other providers
    default:
        return nil, fmt.Errorf("unsupported LLM provider: %s", config.LLMProvider)
    }
}
```

### Extending the Translation Layer

Add new command intents:

```go
// internal/translator/custom_intents.go
package translator

// Add custom intent types
const (
    IntentCustomAction Intent = "custom_action"
    IntentAnalytics   Intent = "analytics"
    IntentReporting   Intent = "reporting"
)

// Add custom translation rules
func (t *Translator) translateCustomAction(response *LLMResponse) (*Translation, error) {
    translation := &Translation{
        Intent:     IntentCustomAction,
        Confidence: response.Confidence,
        Commands:   []MCPCommand{},
    }

    // Parse custom parameters
    if action, ok := response.Parameters["action"]; ok {
        switch action {
        case "generate_report":
            translation.Commands = append(translation.Commands, MCPCommand{
                Method: "generate_report",
                Params: map[string]interface{}{
                    "type":      response.Parameters["report_type"],
                    "timeframe": response.Parameters["timeframe"],
                },
            })
        case "analyze_performance":
            translation.Commands = append(translation.Commands, MCPCommand{
                Method: "analyze_performance",
                Params: map[string]interface{}{
                    "metric": response.Parameters["metric"],
                    "period": response.Parameters["period"],
                },
            })
        }
    }

    return translation, nil
}

// Update the main translate method
func (t *Translator) translate(response *LLMResponse) (*Translation, error) {
    switch response.Intent {
    case IntentCustomAction:
        return t.translateCustomAction(response)
    case IntentAnalytics:
        return t.translateAnalytics(response)
    // ... existing cases
    }
}
```

### Adding Custom API Endpoints

```go
// internal/handlers/custom_chat.go
package handlers

import (
    "encoding/json"
    "net/http"
)

type CustomChatHandler struct {
    storage     storage.Storage
    translator  *translator.Translator
    logger      *slog.Logger
}

func NewCustomChatHandler(storage storage.Storage, translator *translator.Translator, logger *slog.Logger) *CustomChatHandler {
    return &CustomChatHandler{
        storage:    storage,
        translator: translator,
        logger:     logger,
    }
}

// Add custom endpoint for advanced chat features
func (h *CustomChatHandler) HandleAdvancedChat(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    var req struct {
        Message        string            `json:"message"`
        ConversationID string            `json:"conversation_id"`
        Context        map[string]string `json:"context"`
        Mode           string            `json:"mode"` // "standard", "analysis", "bulk"
    }

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }

    // Implement custom logic based on mode
    switch req.Mode {
    case "analysis":
        h.handleAnalysisMode(w, r, &req)
    case "bulk":
        h.handleBulkMode(w, r, &req)
    default:
        h.handleStandardMode(w, r, &req)
    }
}

func (h *CustomChatHandler) handleAnalysisMode(w http.ResponseWriter, r *http.Request, req *ChatRequest) {
    // Implement analysis-specific logic
    // - Enhanced context processing
    // - Statistical analysis
    // - Performance metrics
}
```

### Custom MCP Commands

```go
// internal/mcp/custom_commands.go
package mcp

// Add custom MCP command handlers
func (s *MCPServer) handleCustomCommand(request JSONRPCRequest) JSONRPCResponse {
    switch request.Method {
    case "generate_report":
        return s.handleGenerateReport(request)
    case "analyze_performance":
        return s.handleAnalyzePerformance(request)
    case "bulk_operations":
        return s.handleBulkOperations(request)
    default:
        return JSONRPCResponse{
            JSONRPC: "2.0",
            ID:      request.ID,
            Error: &JSONRPCError{
                Code:    -32601,
                Message: "Method not found",
            },
        }
    }
}

func (s *MCPServer) handleGenerateReport(request JSONRPCRequest) JSONRPCResponse {
    var params struct {
        Type      string `json:"type"`
        Timeframe string `json:"timeframe"`
        Format    string `json:"format"`
    }

    if err := json.Unmarshal(request.Params, &params); err != nil {
        return s.errorResponse(request.ID, -32602, "Invalid params")
    }

    // Generate report based on parameters
    report, err := s.generateReport(params.Type, params.Timeframe, params.Format)
    if err != nil {
        return s.errorResponse(request.ID, -32603, err.Error())
    }

    return JSONRPCResponse{
        JSONRPC: "2.0",
        ID:      request.ID,
        Result:  report,
    }
}
```

## Testing Extensions

### Frontend Testing

```javascript
// test/frontend/chat-manager.test.js
describe('ChatManager Extensions', () => {
    let chatManager;
    let mockApiClient;

    beforeEach(() => {
        mockApiClient = {
            sendChatMessage: jest.fn(),
            getChatHistory: jest.fn()
        };
        chatManager = new ChatManager(mockApiClient);
    });

    test('should handle custom message types', async () => {
        // Test custom handler registration
        const customHandler = jest.fn();
        chatManager.addCustomHandler('custom_type', customHandler);

        // Simulate custom message
        await chatManager.processMessage('custom command');

        expect(customHandler).toHaveBeenCalled();
    });

    test('should preserve conversation context', () => {
        // Test conversation persistence
        chatManager.loadConversationFromStorage();
        expect(chatManager.messages).toBeDefined();
    });
});
```

### Backend Testing

```go
// internal/llm/custom_provider_test.go
func TestCustomProvider_Chat(t *testing.T) {
    // Mock HTTP server
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        response := map[string]interface{}{
            "choices": []map[string]interface{}{
                {
                    "message": map[string]string{
                        "content": "Test response",
                    },
                },
            },
        }
        json.NewEncoder(w).Encode(response)
    }))
    defer server.Close()

    config := &config.Config{
        LLMAPIKey:  "test-key",
        LLMBaseURL: server.URL,
        LLMModel:   "test-model",
        LLMTimeout: 30,
    }

    provider := NewCustomProvider(config)
    
    req := ChatRequest{
        Messages: []Message{
            {Role: "user", Content: "Hello"},
        },
        MaxTokens: 100,
    }

    ctx := context.Background()
    resp, err := provider.Chat(ctx, req)

    assert.NoError(t, err)
    assert.Equal(t, "Test response", resp.Content)
    assert.Equal(t, "test-model", resp.Model)
}
```

### Integration Testing

```go
// test/integration/chat_integration_test.go
func TestChatIntegration(t *testing.T) {
    // Setup test environment
    storage := setupTestStorage(t)
    translator := setupTestTranslator(t)
    handler := handlers.NewChatHandler(storage, translator, logger)

    // Test complete chat flow
    req := httptest.NewRequest("POST", "/api/chat", strings.NewReader(`{
        "message": "Create a test task",
        "conversation_id": "test-conv"
    }`))
    req.Header.Set("Content-Type", "application/json")

    w := httptest.NewRecorder()
    handler.HandleChat(w, req)

    assert.Equal(t, http.StatusOK, w.Code)

    var response map[string]interface{}
    err := json.Unmarshal(w.Body.Bytes(), &response)
    assert.NoError(t, err)
    assert.Contains(t, response, "response")
    assert.Contains(t, response, "actions_taken")
}
```

## Performance Optimization

### Frontend Optimization

```javascript
// Implement message virtualization for large conversations
class VirtualizedChatHistory {
    constructor(container, messages) {
        this.container = container;
        this.messages = messages;
        this.visibleRange = { start: 0, end: 50 };
        this.itemHeight = 60;
    }

    render() {
        // Only render visible messages
        const visible = this.messages.slice(
            this.visibleRange.start,
            this.visibleRange.end
        );

        this.container.innerHTML = visible
            .map(msg => this.renderMessage(msg))
            .join('');
    }

    onScroll() {
        // Update visible range based on scroll position
        const scrollTop = this.container.scrollTop;
        const start = Math.floor(scrollTop / this.itemHeight);
        const end = start + Math.ceil(this.container.clientHeight / this.itemHeight);

        this.visibleRange = { start, end };
        this.render();
    }
}
```

### Backend Optimization

```go
// Implement response caching
type CachedTranslator struct {
    *Translator
    cache map[string]*Translation
    mutex sync.RWMutex
}

func (ct *CachedTranslator) Translate(ctx context.Context, input string) (*Translation, error) {
    // Check cache first
    ct.mutex.RLock()
    if cached, exists := ct.cache[input]; exists {
        ct.mutex.RUnlock()
        return cached, nil
    }
    ct.mutex.RUnlock()

    // Translate and cache result
    translation, err := ct.Translator.Translate(ctx, input)
    if err != nil {
        return nil, err
    }

    ct.mutex.Lock()
    ct.cache[input] = translation
    ct.mutex.Unlock()

    return translation, nil
}
```

## Deployment Considerations

### Docker Extensions

```dockerfile
# Dockerfile.extended
FROM golang:1.24-alpine AS builder

# Install additional dependencies for extensions
RUN apk add --no-cache git make

WORKDIR /app
COPY . .

# Build with extension tags
RUN go build -tags "custom_providers,analytics" -o projectflow cmd/server/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/

COPY --from=builder /app/projectflow .
COPY --from=builder /app/web ./web
COPY --from=builder /app/docs ./docs

# Add custom configuration
COPY config/production.yaml ./config/

EXPOSE 16191
CMD ["./projectflow"]
```

### Kubernetes Configuration

```yaml
# k8s/configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: projectflow-config
data:
  LLM_PROVIDER: "ollama"
  LLM_BASE_URL: "http://ollama-service:11434"
  LLM_MODEL: "llama3.1"
  CUSTOM_FEATURE_FLAGS: "analytics,reporting,bulk_operations"
```

## Contributing Guidelines

### Code Style

```go
// Follow Go conventions
func (s *CustomService) ProcessRequest(ctx context.Context, req *Request) (*Response, error) {
    // Use structured logging
    s.logger.Info("Processing request",
        slog.String("request_id", req.ID),
        slog.String("type", req.Type),
    )

    // Validate input
    if err := req.Validate(); err != nil {
        return nil, fmt.Errorf("invalid request: %w", err)
    }

    // Process with proper error handling
    result, err := s.process(ctx, req)
    if err != nil {
        s.logger.Error("Request processing failed",
            slog.String("request_id", req.ID),
            slog.String("error", err.Error()),
        )
        return nil, err
    }

    return result, nil
}
```

### Documentation

```go
// Package custom provides extended functionality for ProjectFlow chat interface.
//
// This package implements custom LLM providers, enhanced translation capabilities,
// and additional MCP commands for advanced project management workflows.
//
// Example usage:
//
//     provider := custom.NewAdvancedProvider(config)
//     translator := custom.NewEnhancedTranslator(provider)
//     
//     translation, err := translator.Translate(ctx, "generate weekly report")
//     if err != nil {
//         log.Fatal(err)
//     }
//
package custom
```

### Testing Requirements

- Unit tests for all new functions
- Integration tests for API endpoints
- Frontend tests for UI components
- Performance benchmarks for critical paths
- Documentation for all public APIs

### Pull Request Process

1. Fork the repository
2. Create feature branch: `git checkout -b feature/custom-provider`
3. Implement changes with tests
4. Update documentation
5. Run full test suite: `make test`
6. Submit pull request with detailed description

For more information, see the [Contributing Guidelines](../CONTRIBUTING.md) and [Code of Conduct](../CODE_OF_CONDUCT.md).
