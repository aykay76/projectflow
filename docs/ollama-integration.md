# Ollama Integration Guide

This guide covers how to set up, configure, and use Ollama with ProjectFlow for local LLM functionality.

## Table of Contents

- [Overview](#overview)
- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Configuration](#configuration)
- [Usage](#usage)
- [Troubleshooting](#troubleshooting)
- [Advanced Configuration](#advanced-configuration)
- [API Reference](#api-reference)

## Overview

Ollama integration allows ProjectFlow to use local Large Language Models (LLMs) for enhanced chat functionality. This provides:

- **Privacy**: All processing happens locally
- **Offline capability**: No internet required for LLM features
- **Cost-effective**: No API usage fees
- **Customization**: Use any compatible model
- **Performance**: Direct local access without network latency

### Supported Features

- **Smart Assistant Mode**: Translates natural language to ProjectFlow actions
- **Direct LLM Mode**: Chat directly with the language model
- **Health monitoring**: Real-time status and diagnostics
- **Model management**: Support for multiple Ollama models
- **Error handling**: Comprehensive troubleshooting information

## Prerequisites

- **Operating System**: macOS, Linux, or Windows
- **Memory**: At least 8GB RAM (16GB+ recommended for larger models)
- **Storage**: 4GB+ free space for models
- **ProjectFlow**: Version with Ollama support

### Hardware Requirements by Model Size

| Model Size | RAM Required | Storage | Performance |
|------------|--------------|---------|-------------|
| 7B (e.g., llama3.2) | 8GB | 4GB | Good |
| 13B | 16GB | 8GB | Better |
| 30B+ | 32GB+ | 20GB+ | Best |

## Installation

### Step 1: Install Ollama

#### macOS
```bash
# Using Homebrew (recommended)
brew install ollama

# Or download from https://ollama.com/download
```

#### Linux
```bash
# Using the install script
curl -fsSL https://ollama.com/install.sh | sh

# Or using package managers
# Ubuntu/Debian
sudo apt install ollama

# Arch Linux
yay -S ollama
```

#### Windows
```powershell
# Download and install from https://ollama.com/download
# Or using Chocolatey
choco install ollama
```

### Step 2: Start Ollama Service

```bash
# Start Ollama server
ollama serve
```

The service will start on `http://localhost:11434` by default.

### Step 3: Install a Model

```bash
# Install recommended model (llama3.2)
ollama pull llama3.2

# Or install other models
ollama pull codellama
ollama pull mistral
ollama pull phi3
```

### Step 4: Verify Installation

```bash
# List installed models
ollama list

# Test model
ollama run llama3.2 "Hello, how are you?"
```

## Configuration

### ProjectFlow Configuration

Update your ProjectFlow configuration to enable Ollama:

#### Environment Variables
```bash
# Enable LLM functionality
LLM_PROVIDER=ollama

# Ollama server configuration
LLM_OLLAMA_HOST=http://localhost:11434
LLM_OLLAMA_MODEL=llama3.2

# Optional: Timeout and token limits
LLM_TIMEOUT=60
LLM_MAX_TOKENS=1000
```

#### Configuration File (`config.yaml`)
```yaml
llm:
  provider: ollama
  ollama:
    host: http://localhost:11434
    model: llama3.2
  timeout_seconds: 60
  max_tokens: 1000
```

#### Command Line Arguments
```bash
./projectflow --llm-provider=ollama --llm-ollama-host=http://localhost:11434 --llm-ollama-model=llama3.2
```

### Configuration Options

| Option | Default | Description |
|--------|---------|-------------|
| `LLM_PROVIDER` | `disabled` | Set to `ollama` to enable |
| `LLM_OLLAMA_HOST` | `http://localhost:11434` | Ollama server URL |
| `LLM_OLLAMA_MODEL` | `llama3.2` | Model to use |
| `LLM_TIMEOUT` | `60` | Request timeout in seconds |
| `LLM_MAX_TOKENS` | `1000` | Maximum tokens per response |

## Usage

### Starting ProjectFlow with Ollama

1. **Start Ollama service**:
   ```bash
   ollama serve
   ```

2. **Start ProjectFlow**:
   ```bash
   LLM_PROVIDER=ollama ./projectflow
   ```

3. **Open the web interface** and look for the green status indicator in the chat panel.

### Chat Interface

The chat interface provides two modes:

#### Smart Assistant Mode (Default)
- Translates natural language to ProjectFlow actions
- Example: "Create a high priority task to fix the login bug"
- Actions are executed automatically
- Best for productivity and task management

#### Direct LLM Mode
- Chat directly with the language model
- General-purpose conversations
- No automatic actions
- Best for getting information or brainstorming

### Switching Chat Modes

1. Click the mode button (🔄 or 🤖) in the chat header
2. Select your preferred mode
3. Click "Apply"

### Status Monitoring

The chat interface shows LLM status:
- 🟢 **Green**: Healthy and ready
- 🟡 **Yellow**: Connected but issues detected
- ⚪ **White**: Disabled or not configured

Click the status indicator for detailed information.

## Troubleshooting

### Common Issues

#### Ollama Not Running
**Symptoms**: "Connection refused" errors, white status indicator

**Solutions**:
```bash
# Start Ollama service
ollama serve

# Check if running
curl http://localhost:11434/api/version

# Check process
ps aux | grep ollama
```

#### Model Not Found
**Symptoms**: "Model not found" errors

**Solutions**:
```bash
# List available models
ollama list

# Install the required model
ollama pull llama3.2

# Verify installation
ollama run llama3.2 "test"
```

#### Insufficient Memory
**Symptoms**: Model loading failures, system slowdown

**Solutions**:
- Use a smaller model (e.g., `phi3` instead of `llama3.2`)
- Close other memory-intensive applications
- Increase system RAM

#### Slow Responses
**Symptoms**: Long wait times for responses

**Solutions**:
- Use a smaller, faster model
- Reduce `LLM_MAX_TOKENS` setting
- Ensure SSD storage for models
- Check CPU usage and availability

#### Port Conflicts
**Symptoms**: "Port already in use" errors

**Solutions**:
```bash
# Check what's using port 11434
lsof -i :11434

# Kill conflicting process
sudo kill -9 <PID>

# Or use a different port
ollama serve --host 0.0.0.0:11435
```

### Health Check API

Use the health check API for diagnostics:

```bash
# Check LLM status
curl http://localhost:16191/api/llm/health

# Get LLM information
curl http://localhost:16191/api/llm/info

# Example response
{
  "enabled": true,
  "provider": "ollama",
  "model": "llama3.2",
  "status": "healthy",
  "timestamp": "2025-06-23T08:56:47.927Z",
  "metadata": {
    "host": "http://localhost:11434",
    "version": "0.1.17"
  }
}
```

### Debug Logging

Enable debug logging for detailed troubleshooting:

```bash
# Set log level to debug
LOG_LEVEL=DEBUG ./projectflow

# Or use environment variable
export LOG_LEVEL=DEBUG
```

### Performance Monitoring

Monitor system resources while using Ollama:

```bash
# Monitor CPU and memory usage
top -p $(pgrep ollama)

# Monitor disk I/O
iotop -p $(pgrep ollama)

# Check GPU usage (if applicable)
nvidia-smi  # NVIDIA GPUs
```

## Advanced Configuration

### Custom Models

You can use any Ollama-compatible model:

```bash
# Install from Ollama library
ollama pull mistral
ollama pull codellama
ollama pull dolphin-mixtral

# Use custom models
ollama create my-model -f ./Modelfile
```

Then update your configuration:
```bash
LLM_OLLAMA_MODEL=mistral
```

### Model Management

```bash
# List all models
ollama list

# Remove unused models
ollama rm unused-model

# Update existing model
ollama pull llama3.2

# Show model information
ollama show llama3.2
```

### Performance Tuning

#### Memory Management
```bash
# Limit concurrent requests
export OLLAMA_NUM_PARALLEL=1

# Adjust context window
export OLLAMA_CONTEXT_SIZE=2048
```

#### GPU Acceleration
```bash
# Enable GPU (if available)
export OLLAMA_GPU=1

# Specify GPU device
export OLLAMA_GPU_DEVICE=0
```

### Security Considerations

#### Network Security
```bash
# Bind to localhost only (default)
ollama serve --host 127.0.0.1:11434

# Enable authentication (if needed)
export OLLAMA_AUTH=true
```

#### Data Privacy
- All data stays local
- No external API calls
- Models stored locally in `~/.ollama/models`

### Docker Deployment

Use Ollama with Docker:

```dockerfile
# Dockerfile for Ollama
FROM ollama/ollama:latest

# Copy models (optional)
COPY ./models /root/.ollama/models

# Expose port
EXPOSE 11434

# Start service
CMD ["ollama", "serve"]
```

```yaml
# docker-compose.yml
version: '3.8'
services:
  ollama:
    image: ollama/ollama:latest
    ports:
      - "11434:11434"
    volumes:
      - ollama_data:/root/.ollama
    environment:
      - OLLAMA_MODELS=/root/.ollama/models

  projectflow:
    build: .
    ports:
      - "16191:16191"
    environment:
      - LLM_PROVIDER=ollama
      - LLM_OLLAMA_HOST=http://ollama:11434
    depends_on:
      - ollama

volumes:
  ollama_data:
```

## API Reference

### LLM Info Endpoint

**GET** `/api/llm/info`

Returns LLM provider information and status.

**Response**:
```json
{
  "enabled": true,
  "provider": "ollama",
  "model": "llama3.2",
  "status": "healthy",
  "timestamp": "2025-06-23T08:56:47.927Z",
  "metadata": {
    "host": "http://localhost:11434",
    "version": "0.1.17",
    "models": ["llama3.2", "codellama"]
  }
}
```

### LLM Health Check

**GET** `/api/llm/health`

Checks LLM provider health and connectivity.

**Response**:
```json
{
  "healthy": true,
  "status": "healthy",
  "provider": "ollama",
  "timestamp": "2025-06-23T08:56:47.927Z",
  "checks": {
    "connectivity": "ok",
    "model_availability": "ok",
    "response_time_ms": 45
  },
  "suggestions": []
}
```

### Direct LLM Chat

**POST** `/api/llm/chat`

Send messages directly to the LLM.

**Request**:
```json
{
  "messages": [
    {
      "role": "user",
      "content": "Hello, how are you?"
    }
  ],
  "max_tokens": 1000,
  "temperature": 0.7
}
```

**Response**:
```json
{
  "response": {
    "choices": [
      {
        "message": {
          "role": "assistant",
          "content": "Hello! I'm doing well, thank you for asking. How can I help you today?"
        },
        "finish_reason": "stop"
      }
    ],
    "usage": {
      "prompt_tokens": 12,
      "completion_tokens": 23,
      "total_tokens": 35
    }
  },
  "provider": "ollama",
  "model": "llama3.2",
  "timestamp": "2025-06-23T08:56:47.927Z"
}
```

## Best Practices

### Model Selection
- **llama3.2**: Balanced performance, good for general use
- **codellama**: Best for code-related tasks
- **mistral**: Fast and efficient
- **phi3**: Lightweight, good for resource-constrained environments

### Resource Management
- Monitor system resources during usage
- Use appropriate model sizes for your hardware
- Consider model quantization for better performance
- Regularly clean up unused models

### Security
- Keep Ollama updated to the latest version
- Limit network access if not needed
- Monitor for unusual resource usage
- Backup important models

### Performance Optimization
- Use SSD storage for better I/O performance
- Ensure adequate RAM for your chosen models
- Consider GPU acceleration if available
- Tune context window size based on use case

## Conclusion

Ollama integration provides powerful local LLM capabilities for ProjectFlow. With proper setup and configuration, you can enjoy private, fast, and cost-effective AI assistance for your project management tasks.

For additional help:
- Check the [troubleshooting section](#troubleshooting)
- Review [Ollama documentation](https://ollama.com/docs)
- Open an issue on the ProjectFlow repository
- Check server logs for detailed error information
