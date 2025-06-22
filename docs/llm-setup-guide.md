# 🤖 LLM Setup Guide

This guide covers how to configure and set up different Large Language Model (LLM) providers for ProjectFlow's natural language chat interface.

## Provider Overview

ProjectFlow supports multiple LLM providers, each with different strengths:

| Provider | Best For | Cost | Setup Complexity | Performance |
|----------|----------|------|------------------|-------------|
| **Groq** | Development, Testing | Free tier available | Low | Very Fast |
| **Ollama** | Production, Privacy | Free (local) | Medium | Good |
| **OpenAI** | Enterprise, Accuracy | Pay-per-use | Low | Excellent |
| **Disabled** | Testing UI only | Free | None | N/A |

## Groq Setup (Recommended for Development)

Groq provides fast, free LLM inference perfect for development and testing.

### 1. Get API Key

1. Visit [Groq Console](https://console.groq.com/)
2. Sign up for a free account
3. Navigate to API Keys section
4. Create a new API key
5. Copy the key (starts with `gsk_...`)

### 2. Configure ProjectFlow

```bash
# Set environment variables
export LLM_PROVIDER=groq
export LLM_API_KEY=gsk_your_api_key_here
export LLM_MODEL=llama-3.1-8b-instant

# Optional: Adjust settings
export LLM_TIMEOUT=30
export LLM_MAX_TOKENS=1000
```

### 3. Available Models

| Model | Speed | Accuracy | Use Case |
|-------|-------|----------|----------|
| `llama-3.1-8b-instant` | Fastest | Good | Development, Quick responses |
| `llama-3.1-70b-versatile` | Fast | Better | Production, Complex queries |
| `mixtral-8x7b-32768` | Medium | Excellent | Advanced reasoning |

### 4. Rate Limits

- **Free Tier**: 30 requests/minute
- **Paid Tier**: Higher limits available
- **Token Limits**: Varies by model

## Ollama Setup (Recommended for Production)

Ollama allows you to run LLMs locally, providing privacy and control.

### 1. Install Ollama

#### macOS
```bash
# Using Homebrew
brew install ollama

# Or download from https://ollama.ai
```

#### Linux
```bash
# Install script
curl -fsSL https://ollama.ai/install.sh | sh
```

#### Windows
Download the installer from [ollama.ai](https://ollama.ai)

### 2. Start Ollama Service

```bash
# Start the Ollama service
ollama serve

# The service will run on http://localhost:11434
```

### 3. Download Models

```bash
# Recommended model for ProjectFlow
ollama pull llama3.1

# Alternative models
ollama pull mistral
ollama pull codellama
ollama pull phi3
```

### 4. Configure ProjectFlow

```bash
# Set environment variables
export LLM_PROVIDER=ollama
export LLM_BASE_URL=http://localhost:11434
export LLM_MODEL=llama3.1

# Optional: Adjust settings
export LLM_TIMEOUT=60  # Local models may be slower
export LLM_MAX_TOKENS=1000
```

### 5. Model Recommendations

| Model | Size | Memory Req | Best For |
|-------|------|------------|----------|
| `llama3.1` | 4.7GB | 8GB RAM | General purpose |
| `mistral` | 4.1GB | 6GB RAM | Fast responses |
| `codellama` | 3.8GB | 6GB RAM | Code-related tasks |
| `phi3` | 2.3GB | 4GB RAM | Resource-constrained environments |

### 6. Docker Deployment

```dockerfile
# Dockerfile.ollama
FROM ollama/ollama:latest

# Install required model
RUN ollama serve & sleep 5 && ollama pull llama3.1

EXPOSE 11434
```

```bash
# Run Ollama in Docker
docker run -d \
  --name ollama \
  -p 11434:11434 \
  -v ollama_data:/root/.ollama \
  ollama/ollama

# Pull model
docker exec ollama ollama pull llama3.1
```

## OpenAI Setup (Enterprise)

For production environments requiring the highest accuracy.

### 1. Get API Key

1. Visit [OpenAI Platform](https://platform.openai.com/)
2. Sign up or log in to your account
3. Navigate to API Keys section
4. Create a new secret key
5. Copy the key (starts with `sk-...`)

### 2. Configure ProjectFlow

```bash
# Set environment variables
export LLM_PROVIDER=openai
export LLM_API_KEY=sk_your_api_key_here
export LLM_MODEL=gpt-4o-mini

# Optional: Custom base URL for Azure OpenAI
export LLM_BASE_URL=https://your-resource.openai.azure.com/
```

### 3. Available Models

| Model | Cost | Speed | Accuracy | Best For |
|-------|------|-------|----------|----------|
| `gpt-4o-mini` | Low | Fast | Very Good | Cost-effective production |
| `gpt-4` | High | Medium | Excellent | Complex reasoning |
| `gpt-3.5-turbo` | Very Low | Very Fast | Good | High-volume applications |

### 4. Cost Management

```bash
# Limit token usage to control costs
export LLM_MAX_TOKENS=500  # Reduce for cost savings
export LLM_TIMEOUT=20      # Prevent long-running requests
```

## Custom Provider Setup

You can extend ProjectFlow to support additional LLM providers.

### 1. Implement Provider Interface

```go
// internal/llm/custom.go
type CustomProvider struct {
    apiKey  string
    baseURL string
    model   string
}

func (p *CustomProvider) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
    // Implement your provider's API call
}

func (p *CustomProvider) HealthCheck(ctx context.Context) error {
    // Implement health check
}

func (p *CustomProvider) Name() string {
    return "custom"
}

func (p *CustomProvider) Model() string {
    return p.model
}
```

### 2. Register Provider

```go
// internal/llm/factory.go
func CreateProvider(config *config.Config) (LLMProvider, error) {
    switch strings.ToLower(config.LLMProvider) {
    case "custom":
        return NewCustomProvider(config)
    // ... existing providers
    }
}
```

## Testing Configuration

Verify your LLM setup is working correctly.

### 1. Health Check

```bash
# Check service health
curl http://localhost:16191/health
```

Look for LLM service status in the response.

### 2. Chat Test

```bash
# Test chat endpoint
curl -X POST http://localhost:16191/api/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "Hello, can you help me?", "conversation_id": "test"}'
```

### 3. Debug Logging

```bash
# Enable debug logging
export LOG_LEVEL=DEBUG

# Run ProjectFlow and check logs
go run cmd/server/main.go
```

## Performance Optimization

### Response Time Optimization

```bash
# Reduce timeout for faster failures
export LLM_TIMEOUT=15

# Limit response length
export LLM_MAX_TOKENS=500

# Use faster models
export LLM_MODEL=llama-3.1-8b-instant  # For Groq
export LLM_MODEL=llama3.1               # For Ollama
```

### Memory Optimization (Ollama)

```bash
# For systems with limited memory
ollama pull phi3        # Smaller model
export LLM_MODEL=phi3

# Or use quantized models
ollama pull llama3.1:q4_0  # 4-bit quantization
```

### Concurrent Requests

```bash
# Configure for high concurrency
export LLM_TIMEOUT=30
export SHUTDOWN_TIMEOUT=60

# Monitor resource usage
htop  # or Activity Monitor on macOS
```

## Troubleshooting

### Common Issues

#### "LLM service unavailable"
- **Groq**: Check API key validity and rate limits
- **Ollama**: Ensure `ollama serve` is running
- **OpenAI**: Verify API key and billing status

#### Slow Responses
- **Ollama**: Use smaller models or increase `LLM_TIMEOUT`
- **Network**: Check internet connection for cloud providers
- **Resources**: Ensure adequate RAM/CPU for local models

#### High Error Rates
- **Rate Limiting**: Implement exponential backoff
- **Model Issues**: Try different models
- **Configuration**: Verify all environment variables

### Debug Commands

```bash
# Test Ollama directly
curl http://localhost:11434/api/generate \
  -d '{"model": "llama3.1", "prompt": "Hello"}'

# Test Groq directly
curl https://api.groq.com/openai/v1/chat/completions \
  -H "Authorization: Bearer $LLM_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"messages": [{"role": "user", "content": "Hello"}], "model": "llama-3.1-8b-instant"}'

# Check ProjectFlow logs
tail -f server.log | grep LLM
```

## Security Considerations

### API Key Management

```bash
# Use environment files
echo "LLM_API_KEY=your_key_here" > .env
source .env

# Or use secret management
# - AWS Secrets Manager
# - Azure Key Vault  
# - HashiCorp Vault
```

### Network Security

```bash
# For Ollama, restrict access
iptables -A INPUT -p tcp --dport 11434 -s 127.0.0.1 -j ACCEPT
iptables -A INPUT -p tcp --dport 11434 -j DROP
```

### Data Privacy

- **Groq/OpenAI**: Data sent to external services
- **Ollama**: Data processed locally
- **Logs**: Avoid logging sensitive information

## Production Deployment

### Docker Compose Example

```yaml
# docker-compose.yml
version: '3.8'
services:
  projectflow:
    build: .
    ports:
      - "16191:16191"
    environment:
      - LLM_PROVIDER=ollama
      - LLM_BASE_URL=http://ollama:11434
      - LLM_MODEL=llama3.1
    depends_on:
      - ollama

  ollama:
    image: ollama/ollama:latest
    ports:
      - "11434:11434"
    volumes:
      - ollama_data:/root/.ollama
    command: >
      sh -c "ollama serve & sleep 10 && ollama pull llama3.1 && wait"

volumes:
  ollama_data:
```

### Kubernetes Deployment

```yaml
# ollama-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ollama
spec:
  replicas: 1
  selector:
    matchLabels:
      app: ollama
  template:
    metadata:
      labels:
        app: ollama
    spec:
      containers:
      - name: ollama
        image: ollama/ollama:latest
        ports:
        - containerPort: 11434
        volumeMounts:
        - name: ollama-data
          mountPath: /root/.ollama
      volumes:
      - name: ollama-data
        persistentVolumeClaim:
          claimName: ollama-pvc
---
apiVersion: v1
kind: Service
metadata:
  name: ollama-service
spec:
  selector:
    app: ollama
  ports:
  - port: 11434
    targetPort: 11434
```

## Monitoring and Maintenance

### Metrics to Monitor

- Response times
- Error rates
- Token usage (for paid providers)
- Memory usage (for local models)
- Request volume

### Maintenance Tasks

- **Model Updates**: Regularly update to newer model versions
- **Log Rotation**: Implement log rotation for debug logs
- **Health Checks**: Monitor service availability
- **Backup**: Backup conversation history if needed

For more information, see the [Troubleshooting Guide](troubleshooting.md) and [API Documentation](../README.md).
