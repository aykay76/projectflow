# 🔧 Troubleshooting Guide

This guide helps you diagnose and resolve common issues with ProjectFlow's natural language chat interface.

## Quick Diagnostic Checklist

Before diving into specific issues, run through this quick checklist:

- [ ] Is ProjectFlow server running?
- [ ] Is the chat interface visible in the UI?
- [ ] Are there any error messages in the browser console?
- [ ] Is the LLM service configured and accessible?
- [ ] Are environment variables set correctly?

## Common Issues and Solutions

### 🚫 Chat Interface Issues

#### Chat button not visible
**Symptoms**: No 💬 button in header or floating area

**Solutions**:
1. **Check browser compatibility**:
   ```bash
   # Ensure modern browser (Chrome 90+, Firefox 88+, Safari 14+)
   ```

2. **Clear browser cache**:
   ```bash
   # Hard refresh: Ctrl+F5 (Windows) or Cmd+Shift+R (Mac)
   ```

3. **Check JavaScript errors**:
   ```javascript
   // Open browser console (F12) and look for errors
   console.error // Look for any red error messages
   ```

4. **Verify file loading**:
   ```bash
   # Check if chat-manager.js is loading
   curl http://localhost:16191/static/js/chat-manager.js
   ```

#### Chat panel won't open
**Symptoms**: Button is visible but clicking doesn't open chat panel

**Solutions**:
1. **Check for JavaScript errors**:
   ```javascript
   // Browser console should show ChatManager initialization
   // Look for: "ChatManager initialized"
   ```

2. **Verify CSS loading**:
   ```bash
   # Check if chat.css is loading
   curl http://localhost:16191/static/css/components/chat.css
   ```

3. **Test keyboard shortcut**:
   ```
   Try Cmd+/ (Mac) or Ctrl+/ (Windows) to toggle chat
   ```

### 🔄 API Connection Issues

#### "Failed to send message" error
**Symptoms**: Error message when trying to send chat messages

**Solutions**:
1. **Check server status**:
   ```bash
   curl http://localhost:16191/health
   # Should return 200 OK with service status
   ```

2. **Verify chat endpoint**:
   ```bash
   curl -X POST http://localhost:16191/api/chat \
     -H "Content-Type: application/json" \
     -d '{"message": "test", "conversation_id": "debug"}'
   ```

3. **Check network connectivity**:
   ```bash
   # Test basic connectivity
   ping localhost
   netstat -an | grep 16191
   ```

4. **Review server logs**:
   ```bash
   # Look for error messages in server output
   tail -f server.log | grep ERROR
   ```

#### Connection timeout errors
**Symptoms**: Requests take too long or timeout

**Solutions**:
1. **Check LLM provider status**:
   ```bash
   # For Groq
   curl https://api.groq.com/openai/v1/models \
     -H "Authorization: Bearer $LLM_API_KEY"
   
   # For Ollama
   curl http://localhost:11434/api/tags
   ```

2. **Increase timeout settings**:
   ```bash
   export LLM_TIMEOUT=60
   export SHUTDOWN_TIMEOUT=90
   ```

3. **Monitor resource usage**:
   ```bash
   # Check CPU and memory usage
   top
   free -h
   ```

### 🤖 LLM Service Issues

#### "LLM service is not enabled" error
**Symptoms**: Chat responds with "LLM service is not enabled" message

**Solutions**:
1. **Check LLM provider configuration**:
   ```bash
   echo $LLM_PROVIDER
   echo $LLM_API_KEY
   echo $LLM_MODEL
   ```

2. **Set correct provider**:
   ```bash
   # For development with Groq
   export LLM_PROVIDER=groq
   export LLM_API_KEY=gsk_your_key_here
   
   # For local with Ollama
   export LLM_PROVIDER=ollama
   export LLM_BASE_URL=http://localhost:11434
   ```

3. **Restart server**:
   ```bash
   # Stop current server (Ctrl+C)
   # Restart with new environment
   go run cmd/server/main.go
   ```

#### "Invalid API key" or authentication errors
**Symptoms**: LLM provider returns authentication errors

**Solutions**:
1. **Verify API key format**:
   ```bash
   # Groq keys start with 'gsk_'
   # OpenAI keys start with 'sk-'
   echo $LLM_API_KEY | cut -c1-4
   ```

2. **Test API key directly**:
   ```bash
   # Test Groq API key
   curl https://api.groq.com/openai/v1/models \
     -H "Authorization: Bearer $LLM_API_KEY"
   
   # Test OpenAI API key
   curl https://api.openai.com/v1/models \
     -H "Authorization: Bearer $LLM_API_KEY"
   ```

3. **Check billing/quotas**:
   - Visit your provider's dashboard
   - Verify account is active and has credits
   - Check usage quotas and rate limits

#### Ollama service not responding
**Symptoms**: "Connection refused" when using Ollama provider

**Solutions**:
1. **Check if Ollama is running**:
   ```bash
   ps aux | grep ollama
   # Should show ollama serve process
   ```

2. **Start Ollama service**:
   ```bash
   ollama serve
   # Should start on http://localhost:11434
   ```

3. **Verify model availability**:
   ```bash
   ollama list
   # Should show installed models
   
   # Pull required model if missing
   ollama pull llama3.1
   ```

4. **Test Ollama directly**:
   ```bash
   curl http://localhost:11434/api/generate \
     -d '{"model": "llama3.1", "prompt": "Hello", "stream": false}'
   ```

### 📱 UI/UX Issues

#### Chat interface appears broken or unstyled
**Symptoms**: Chat panel has no styling or appears broken

**Solutions**:
1. **Check CSS loading**:
   ```bash
   # Verify CSS files are accessible
   curl -I http://localhost:16191/static/css/main.css
   curl -I http://localhost:16191/static/css/components/chat.css
   ```

2. **Clear browser cache**:
   ```bash
   # Force reload all assets
   Ctrl+Shift+R (Windows/Linux) or Cmd+Shift+R (Mac)
   ```

3. **Check for CSS conflicts**:
   ```javascript
   // In browser console, check for CSS errors
   console.log(document.querySelectorAll('link[rel="stylesheet"]'))
   ```

#### Mobile responsiveness issues
**Symptoms**: Chat interface doesn't work properly on mobile devices

**Solutions**:
1. **Test mobile viewport**:
   ```javascript
   // In browser dev tools
   // Toggle device emulation (mobile view)
   ```

2. **Check touch events**:
   ```javascript
   // Ensure touch events are working
   document.addEventListener('touchstart', (e) => console.log('touch', e))
   ```

3. **Verify viewport meta tag**:
   ```html
   <!-- Should be present in HTML head -->
   <meta name="viewport" content="width=device-width, initial-scale=1.0">
   ```

#### Keyboard shortcuts not working
**Symptoms**: Cmd+/ or Ctrl+/ doesn't toggle chat

**Solutions**:
1. **Check for conflicting shortcuts**:
   ```
   Ensure no browser extensions are capturing the shortcut
   Try in incognito/private mode
   ```

2. **Verify keyboard event handling**:
   ```javascript
   // In browser console
   document.addEventListener('keydown', (e) => {
     if ((e.metaKey || e.ctrlKey) && e.key === '/') {
       console.log('Chat shortcut triggered');
     }
   });
   ```

3. **Test alternative methods**:
   ```
   Use header button or floating button as alternative
   ```

### 🗄️ Data and Storage Issues

#### Conversation history not persisting
**Symptoms**: Chat history disappears after page refresh

**Solutions**:
1. **Check localStorage availability**:
   ```javascript
   // In browser console
   console.log(localStorage.getItem('projectflow_chat_history'))
   ```

2. **Verify storage quota**:
   ```javascript
   // Check if localStorage is full
   try {
     localStorage.setItem('test', 'test')
     localStorage.removeItem('test')
     console.log('localStorage available')
   } catch (e) {
     console.error('localStorage issue:', e)
   }
   ```

3. **Clear corrupted data**:
   ```javascript
   // Reset chat storage
   localStorage.removeItem('projectflow_chat_history')
   localStorage.removeItem('projectflow_conversations')
   ```

#### Server-side conversation storage issues
**Symptoms**: Chat history endpoint returns errors

**Solutions**:
1. **Check server logs**:
   ```bash
   grep "conversation" server.log
   grep "ERROR" server.log
   ```

2. **Verify data directory permissions**:
   ```bash
   ls -la data/
   # Ensure ProjectFlow has write permissions
   ```

3. **Test conversation endpoint**:
   ```bash
   curl -X GET "http://localhost:16191/api/chat/history?conversation_id=test"
   ```

### ⚡ Performance Issues

#### Slow response times
**Symptoms**: Chat takes longer than 5 seconds to respond

**Solutions**:
1. **Check LLM provider performance**:
   ```bash
   # Time the LLM API directly
   time curl -X POST https://api.groq.com/openai/v1/chat/completions \
     -H "Authorization: Bearer $LLM_API_KEY" \
     -H "Content-Type: application/json" \
     -d '{"messages": [{"role": "user", "content": "Hello"}], "model": "llama-3.1-8b-instant"}'
   ```

2. **Monitor resource usage**:
   ```bash
   # Check system resources
   htop
   iostat 1
   ```

3. **Optimize configuration**:
   ```bash
   # Reduce token limits for faster responses
   export LLM_MAX_TOKENS=500
   export LLM_TIMEOUT=15
   ```

4. **Use faster models**:
   ```bash
   # Switch to faster models
   export LLM_MODEL=llama-3.1-8b-instant  # For Groq
   export LLM_MODEL=phi3                   # For Ollama
   ```

#### High memory usage
**Symptoms**: System becomes slow or runs out of memory

**Solutions**:
1. **Monitor memory usage**:
   ```bash
   # Check memory consumption
   ps aux | grep projectflow
   free -h
   ```

2. **Optimize Ollama models**:
   ```bash
   # Use smaller or quantized models
   ollama pull llama3.1:q4_0  # 4-bit quantization
   export LLM_MODEL=llama3.1:q4_0
   ```

3. **Limit concurrent requests**:
   ```bash
   # Implement rate limiting in reverse proxy
   # nginx example:
   limit_req_zone $binary_remote_addr zone=chat:10m rate=10r/m;
   ```

## Advanced Debugging

### Enable Debug Logging

```bash
# Set debug level
export LOG_LEVEL=DEBUG

# Run server with verbose output
go run cmd/server/main.go 2>&1 | tee debug.log
```

### Network Debugging

```bash
# Monitor network traffic
tcpdump -i lo port 16191

# Check connection states
netstat -tulpn | grep :16191

# Test with curl verbose
curl -v -X POST http://localhost:16191/api/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "debug test"}'
```

### Database/Storage Debugging

```bash
# Check file storage structure
find data/ -type f -name "*.json" | head -10

# Verify JSON validity
jq . data/projects/PF.json
jq . data/PF/tasks/PF-1.json

# Check permissions
ls -la data/
```

### Browser Developer Tools

```javascript
// Monitor chat manager state
window.projectFlowApp.chatManager.isOpen
window.projectFlowApp.chatManager.messages

// Test API client directly
window.projectFlowApp.apiClient.sendChatMessage('test message', 'debug-id')

// Check for memory leaks
console.log(performance.memory)
```

## Error Code Reference

| Error Code | Description | Solution |
|------------|-------------|----------|
| `CHAT_001` | LLM service unavailable | Check LLM configuration |
| `CHAT_002` | Invalid conversation ID | Use valid UUID or omit for new conversation |
| `CHAT_003` | Message too long | Reduce message length or increase MAX_TOKENS |
| `CHAT_004` | Rate limit exceeded | Wait and retry, or check provider limits |
| `CHAT_005` | Translation failed | Check LLM response format |
| `API_001` | Invalid JSON request | Verify request body format |
| `API_002` | Missing required fields | Include all required fields |
| `API_003` | Authentication required | Set up authentication middleware |
| `STORAGE_001` | File system error | Check permissions and disk space |
| `STORAGE_002` | Conversation not found | Verify conversation ID exists |

## Getting Additional Help

### Collect Diagnostic Information

Before seeking help, collect this information:

```bash
# System information
uname -a
go version
echo "LLM_PROVIDER: $LLM_PROVIDER"
echo "LLM_MODEL: $LLM_MODEL"

# Service status
curl http://localhost:16191/health

# Recent logs
tail -100 server.log

# Browser information
# - Browser version
# - Console errors (F12 -> Console)
# - Network tab errors (F12 -> Network)
```

### Community Resources

- **GitHub Issues**: [Create an issue](https://github.com/aykay76/projectflow/issues)
- **Discussions**: [Community discussions](https://github.com/aykay76/projectflow/discussions)
- **Documentation**: [Complete docs](../README.md)

### Professional Support

For production deployments requiring professional support:
- Include diagnostic information
- Describe the business impact
- Provide reproduction steps
- Include configuration (excluding sensitive data)

---

If you've followed this troubleshooting guide and still have issues, please create a GitHub issue with the diagnostic information and steps you've already tried.
