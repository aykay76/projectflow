# Ollama Quick Start Guide

A quick guide to get up and running with Ollama and ProjectFlow in minutes.

## Quick Setup (5 minutes)

### 1. Install Ollama

**macOS**:
```bash
brew install ollama
```

**Linux**:
```bash
curl -fsSL https://ollama.com/install.sh | sh
```

**Windows**: Download from [ollama.com](https://ollama.com/download)

### 2. Start Ollama & Install Model

```bash
# Start the service
ollama serve &

# Install the recommended model (this may take a few minutes)
ollama pull llama3.2
```

### 3. Configure ProjectFlow

```bash
# Set environment variables
export LLM_PROVIDER=ollama
export LLM_OLLAMA_HOST=http://localhost:11434
export LLM_OLLAMA_MODEL=llama3.2

# Start ProjectFlow
./projectflow
```

### 4. Test the Integration

1. Open http://localhost:16191 in your browser
2. Click the chat button (💬)
3. Look for the green status indicator (🟢)
4. Try asking: "Create a task to test Ollama integration"

## Troubleshooting

### ❌ Status shows white/red circle
```bash
# Check if Ollama is running
curl http://localhost:11434/api/version

# If not running, start it
ollama serve
```

### ❌ "Model not found" error
```bash
# Install the model
ollama pull llama3.2

# Verify it's installed
ollama list
```

### ❌ Slow responses
Try a smaller, faster model:
```bash
ollama pull phi3
export LLM_OLLAMA_MODEL=phi3
```

## Chat Modes

- **🔄 Smart Assistant**: Converts your requests to ProjectFlow actions
- **🤖 Direct LLM**: Chat directly with the AI model

Click the mode button in the chat header to switch.

## Next Steps

- Read the [full integration guide](ollama-integration.md)
- Try different models: `ollama pull mistral`, `ollama pull codellama`
- Explore the chat interface features
- Set up your preferred configuration

Need help? Check the [troubleshooting guide](ollama-integration.md#troubleshooting).
