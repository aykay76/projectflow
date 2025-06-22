# ❓ Frequently Asked Questions (FAQ)

Common questions and answers about ProjectFlow's features, setup, and usage.

## Table of Contents

1. [General Questions](#general-questions)
2. [Getting Started](#getting-started)
3. [Natural Language Chat](#natural-language-chat)
4. [Task Management](#task-management)
5. [Project Organization](#project-organization)
6. [Technical Questions](#technical-questions)
7. [Troubleshooting](#troubleshooting)
8. [Advanced Usage](#advanced-usage)

## General Questions

### What is ProjectFlow?

ProjectFlow is a modern workflow management system designed for AI-assisted development. It combines traditional project management features with a natural language chat interface, allowing you to manage tasks and projects using conversational commands.

### How is ProjectFlow different from Jira or Trello?

ProjectFlow offers several unique features:
- **Natural Language Interface**: Manage tasks using conversational AI
- **AI-First Design**: Built for seamless AI agent integration
- **Lightweight**: Simple setup without complex configuration
- **Developer-Focused**: Git integration and MCP support
- **Modern UI**: Clean, accessible interface with keyboard shortcuts

### Is ProjectFlow free to use?

Yes, ProjectFlow is open source and free to use. You can self-host it with no licensing fees. The only costs are for infrastructure (if using cloud hosting) and optional LLM API usage (if using external AI providers like OpenAI or Groq).

### Can I use ProjectFlow for personal projects?

Absolutely! ProjectFlow works great for personal task management, side projects, and individual development work. The natural language interface makes it easy to quickly capture and organize your thoughts and tasks.

### Does ProjectFlow work offline?

The web interface requires an internet connection to your ProjectFlow server. However, if you're running ProjectFlow locally and using the "disabled" LLM provider, you can use most features offline. The chat interface requires an internet connection to LLM providers.

## Getting Started

### How do I install ProjectFlow?

The quickest way to get started:

```bash
# Clone the repository
git clone https://github.com/aykay76/projectflow.git
cd projectflow

# Run locally
go run cmd/server/main.go

# Access at http://localhost:16191
```

For production deployments, see our [Deployment Guide](deployment-guide.md).

### Do I need to set up a database?

No, ProjectFlow works with file-based storage by default. For production use or team collaboration, you can optionally configure PostgreSQL. See [PostgreSQL Storage Documentation](postgresql-storage.md) for setup instructions.

### How do I set up the chat interface?

The chat interface requires an LLM provider. For development, we recommend Groq (free tier available):

```bash
export LLM_PROVIDER=groq
export LLM_API_KEY=your_groq_api_key
```

See our [LLM Setup Guide](llm-setup-guide.md) for detailed configuration instructions.

### Can I import my existing tasks from Jira/Trello?

Currently, ProjectFlow doesn't have built-in import tools. However, you can:
1. Use the REST API to programmatically create tasks
2. Use the chat interface for quick task recreation: "Create these tasks: [list]"
3. Contribute an import tool to the project

### How do I back up my data?

**File Storage**: Copy the `data/` directory
```bash
cp -r data/ backup-$(date +%Y%m%d)/
```

**PostgreSQL**: Use standard PostgreSQL backup tools
```bash
pg_dump -h localhost -U projectflow projectflow > backup.sql
```

## Natural Language Chat

### How accurate is the natural language understanding?

ProjectFlow uses advanced LLMs that understand context and intent very well. Accuracy depends on:
- **Clarity of your request**: Be specific about what you want
- **LLM provider**: Different providers have different capabilities
- **Context**: The AI learns from your conversation history

### What commands can I use in the chat?

The chat interface supports a wide range of commands:

**Task Management**:
- "Create a task to fix the login bug"
- "Mark task PF-123 as done"
- "Show me all high priority tasks"

**Project Management**:
- "Create a new project called Website Redesign"
- "Show project statistics for PF"

**Queries**:
- "What tasks are overdue?"
- "Show me completed work from last week"

See the [Chat Interface Guide](chat-interface-guide.md) for comprehensive examples.

### Can I use the chat in different languages?

Currently, the chat interface is optimized for English. While many LLMs can understand other languages, the command parsing and responses are primarily designed for English. We're open to contributions for multi-language support.

### Why isn't the chat understanding my request?

Try these troubleshooting steps:

1. **Be more specific**: Instead of "fix the bug", try "create a high priority task to fix the login bug"
2. **Use clear task IDs**: "Update task PF-123" instead of "update the login task"
3. **Check your LLM provider**: Ensure your API key is valid and you have quota remaining
4. **Simplify complex requests**: Break down multi-step requests into individual commands

### Can I train the chat with my own data?

The chat interface uses external LLM providers and doesn't store or train on your conversations. However, conversation history is maintained within each session to provide context for follow-up questions.

## Task Management

### What's the difference between task types?

- **Epic**: Large initiatives (3-6 months), broken down into stories
- **Story**: User-facing features (1-4 weeks), contains related tasks  
- **Task**: Specific work items (1-5 days)
- **Subtask**: Small parts of larger tasks (hours to 1 day)

### How do I create task hierarchies?

You can create parent-child relationships:

**Via Chat**:
```
Create an epic called "User Authentication System"
Create a story under that epic for "Login Flow"
Add a task under the login story to "Design login form"
```

**Via Web Interface**:
1. Create parent task first
2. When creating child task, select parent in "Parent Task" field
3. Or drag-and-drop in hierarchy view

### Can I assign tasks to team members?

Currently, ProjectFlow focuses on task organization rather than user management. You can:
- Include assignee information in task titles or descriptions
- Use tags or labels in descriptions
- Filter tasks by searching for team member names

Future versions may include formal user assignment features.

### How do I set up recurring tasks?

ProjectFlow doesn't have built-in recurring task functionality. You can:
1. Create task templates and recreate them periodically
2. Use the chat interface: "Create weekly status report tasks for the next month"
3. Set up external automation tools to create tasks via the API

### Can I track time spent on tasks?

ProjectFlow doesn't include built-in time tracking. You can:
- Add time estimates and actuals in task descriptions
- Integrate with external time tracking tools via API
- Use comments to log time spent

## Project Organization

### How many projects can I create?

There's no built-in limit on the number of projects. Performance depends on your storage backend and system resources. File storage can handle hundreds of projects efficiently.

### What makes a good project prefix?

Good project prefixes are:
- **Short**: 2-4 characters
- **Memorable**: Easy to remember and type
- **Descriptive**: Related to the project name
- **Unique**: Don't reuse prefixes

Examples:
- WEB (Website Redesign)
- API (Backend API)  
- MOB (Mobile App)
- DOC (Documentation)

### Can I change a project prefix after creation?

Currently, you cannot change project prefixes through the UI. This is because task IDs (like PF-123) are based on the prefix. To change a prefix, you would need to:
1. Create a new project with the desired prefix
2. Migrate tasks manually or via API
3. Delete the old project

### How do I organize large projects?

For large projects (6+ months):

1. **Create multiple epics** for different work streams
2. **Use story mapping** to organize user journeys
3. **Break down epics** into 2-4 week stories
4. **Keep tasks small** (1-5 days each)
5. **Use descriptive naming** conventions

Example structure:
```
Project: WEB (Website Redesign)
├── Epic: User Research & Strategy
├── Epic: Design System
├── Epic: Frontend Development  
├── Epic: Backend Integration
├── Epic: Testing & QA
└── Epic: Launch & Optimization
```

### Can I archive or delete completed projects?

Currently, there's no archive function in the UI. You can:
- Keep completed projects for reference
- Manually delete project data from the file system
- Use API calls to clean up old projects

Future versions may include archive functionality.

## Technical Questions

### What technology stack does ProjectFlow use?

- **Backend**: Go 1.24
- **Frontend**: HTML templates, CSS, JavaScript (no frameworks)
- **Storage**: File system (JSON) or PostgreSQL
- **Containerization**: Docker/Podman support
- **Protocols**: HTTP REST API + Model Context Protocol (MCP)

### Can I run ProjectFlow in Docker?

Yes! ProjectFlow includes a Dockerfile:

```bash
# Build image
podman build -t projectflow .

# Run container
podman run -p 16191:16191 -v $(pwd)/data:/app/data projectflow
```

See the [Deployment Guide](deployment-guide.md) for more container options.

### Is there an API I can use?

Yes, ProjectFlow provides a comprehensive REST API:

**Task Management**:
- `GET /api/tasks` - List tasks
- `POST /api/tasks` - Create task
- `PUT /api/tasks/{id}` - Update task
- `DELETE /api/tasks/{id}` - Delete task

**Chat Interface**:
- `POST /api/chat` - Send message
- `GET /api/chat/history` - Get conversation history

**Project Management**:
- `GET /api/projects` - List projects
- `POST /api/projects` - Create project

### Does ProjectFlow integrate with VS Code?

Yes! ProjectFlow includes Model Context Protocol (MCP) support for seamless VS Code integration:

1. Start the MCP server: `go run cmd/mcp-server/main.go`
2. Configure your MCP client to connect
3. AI agents can now create and manage tasks directly

See [MCP Documentation](mcp.md) for setup details.

### Can I customize the UI?

The UI is built with standard HTML, CSS, and JavaScript. You can:
- Modify CSS files in `web/static/css/`
- Customize HTML templates in `web/templates/`
- Extend JavaScript functionality in `web/static/js/`

### How do I contribute to the project?

We welcome contributions! Here's how to get started:

1. **Fork the repository** on GitHub
2. **Create a feature branch** for your changes
3. **Write tests** for new functionality
4. **Submit a pull request** with a clear description

See our contributing guidelines in the repository for more details.

## Troubleshooting

### The chat interface isn't working

Check these common issues:

1. **LLM provider configuration**:
   ```bash
   # Verify environment variables
   echo $LLM_PROVIDER
   echo $LLM_API_KEY
   ```

2. **API connectivity**:
   ```bash
   # Test chat endpoint
   curl -X POST http://localhost:16191/api/chat \
     -H "Content-Type: application/json" \
     -d '{"message": "Hello"}'
   ```

3. **Browser console errors**: Open F12 and check for JavaScript errors

### Tasks aren't saving

1. **Check server logs** for error messages
2. **Verify storage configuration**:
   ```bash
   # File storage
   ls -la data/
   
   # PostgreSQL
   psql -h localhost -U projectflow projectflow -c "SELECT COUNT(*) FROM tasks;"
   ```

3. **Test API endpoints**:
   ```bash
   curl -X GET http://localhost:16191/api/tasks
   ```

### The server won't start

1. **Check port availability**:
   ```bash
   lsof -i :16191
   ```

2. **Verify dependencies**:
   ```bash
   go mod tidy
   go run cmd/server/main.go
   ```

3. **Check environment variables**: Ensure required variables are set

### Performance is slow

1. **Check resource usage**:
   ```bash
   # CPU and memory
   top
   
   # Disk space
   df -h
   ```

2. **Optimize database** (PostgreSQL):
   ```sql
   VACUUM ANALYZE;
   ```

3. **Review logs** for error patterns

### I can't access the web interface

1. **Verify server is running**:
   ```bash
   curl http://localhost:16191/health
   ```

2. **Check firewall settings**: Ensure port 16191 is accessible

3. **Try different browser**: Rule out browser-specific issues

## Advanced Usage

### Can I run multiple ProjectFlow instances?

Yes, you can run multiple instances:

1. **Different ports**: Use different `PORT` environment variables
2. **Different data directories**: Use different `DATA_DIR` paths
3. **Different databases**: Use separate PostgreSQL databases

### How do I integrate with CI/CD pipelines?

You can integrate ProjectFlow with CI/CD using the API:

**GitHub Actions Example**:
```yaml
- name: Create deployment task
  run: |
    curl -X POST ${{ secrets.PROJECTFLOW_URL }}/api/tasks \
      -H "Content-Type: application/json" \
      -d '{"title": "Deploy build ${{ github.sha }}", "status": "done"}'
```

### Can I create custom task types?

Currently, task types are fixed (Epic, Story, Task, Subtask). You can:
- Use tags in descriptions for custom categorization
- Extend the codebase to add new types
- Use projects to separate different types of work

### How do I set up team collaboration?

For team collaboration:

1. **Deploy to shared server**: Use cloud hosting or shared infrastructure
2. **Use PostgreSQL**: For concurrent access and data integrity
3. **Establish conventions**: Agree on naming, project structure, and workflows
4. **Use chat for coordination**: "What's blocking the frontend team?"

### Can I backup to cloud storage?

For automated cloud backups:

**File Storage**:
```bash
# AWS S3
aws s3 sync data/ s3://your-bucket/projectflow-backup/

# Google Cloud Storage  
gsutil -m rsync -r data/ gs://your-bucket/projectflow-backup/
```

**PostgreSQL**:
```bash
# Backup to cloud
pg_dump -h localhost -U projectflow projectflow | \
  aws s3 cp - s3://your-bucket/projectflow-db-backup.sql
```

### How do I scale ProjectFlow for large teams?

For large team deployments:

1. **Use PostgreSQL**: Better concurrent access than file storage
2. **Deploy on Kubernetes**: For high availability and scaling
3. **Set up monitoring**: Track performance and usage
4. **Consider load balancing**: Multiple backend instances
5. **Implement caching**: Redis for session management

See the [Deployment Guide](deployment-guide.md) for production scaling patterns.

---

## Still Have Questions?

If you can't find the answer to your question here:

1. **Check our documentation**: [User Guide](user-guide.md), [Chat Interface Guide](chat-interface-guide.md)
2. **Search GitHub issues**: Someone may have asked the same question
3. **Create a new issue**: We're happy to help with specific problems
4. **Join the community**: Discussion forums and community chat

**For urgent production issues**, consider our professional support options for enterprise users.

---

**Pro Tip**: You can ask the chat interface many of these questions directly! Try "How do I create a new project?" or "What's the difference between a task and a story?" to get instant answers.
