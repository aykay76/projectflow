# 💬 Chat Interface Guide

The ProjectFlow Chat Interface allows you to manage projects and tasks using natural language commands. This guide covers everything you need to know to get the most out of this powerful feature.

## Getting Started

### Opening the Chat Interface

There are multiple ways to access the chat interface:

1. **Header Button**: Click the 💬 icon in the top-right header
2. **Keyboard Shortcut**: 
   - **Mac**: `⌘ + /`
   - **Windows/Linux**: `Ctrl + /`
3. **Floating Button**: Click the 💬 floating button in the bottom-right corner

### Basic Usage

1. **Type your request** in natural language
2. **Press Enter** or click the send button (➤)
3. **Review the response** and any actions taken
4. **Continue the conversation** with follow-up questions

## Natural Language Commands

### Task Management

#### Creating Tasks

```
Create a task to implement user authentication
Add a high priority bug fix for the login issue
Make a new story about user profile management
Create an epic for the mobile app development
```

**Supported Parameters:**
- **Priority**: low, medium, high, critical
- **Type**: task, story, epic, subtask
- **Project**: Specify project name or prefix
- **Due Date**: today, tomorrow, specific dates (2025-12-31)

#### Listing and Querying Tasks

```
Show me all tasks
List tasks in the PF project
What tasks are assigned to me?
Show overdue tasks
Display high priority tasks
List all stories
```

#### Updating Tasks

```
Mark task PF-123 as done
Set task PF-456 to in progress
Change the priority of PF-789 to high
Update task PF-321 status to blocked
Complete task PF-555
```

#### Task Information

```
Show me details for task PF-123
What's the status of PF-456?
Tell me about task PF-789
Get information on PF-321
```

### Project Management

#### Creating Projects

```
Create a new project called "Website Redesign"
Add a project for mobile app development
Make a project with prefix "WEB" called "Web Portal"
```

#### Project Queries

```
List all projects
Show me the PF project details
What projects do we have?
Display project information
```

### Advanced Queries

#### Filtered Searches

```
Show me all high priority tasks that are overdue
List completed tasks from last week
Find all blocked tasks in the PF project
Show me tasks created today
Display all epics that are in progress
```

#### Relative Dates

```
Tasks due today
Items overdue by more than 3 days
Tasks created this week
Work completed yesterday
Items due next week
```

## Chat Interface Features

### Conversation History

- **Automatic Persistence**: Conversations are automatically saved and restored
- **Cross-Session**: History persists across browser sessions
- **Search**: Find previous conversations and responses

### Smart Suggestions

The chat interface provides intelligent suggestions based on:
- Current project context
- Recent activity
- Common command patterns
- Available actions

### Error Handling

When the system can't understand your request:

1. **Clarification Requests**: The system will ask for more specific information
2. **Suggestions**: Alternative phrasings or commands will be provided
3. **Help**: Type "help" for command examples and guidance

### Keyboard Shortcuts

| Shortcut | Action |
|----------|---------|
| `⌘/Ctrl + /` | Toggle chat interface |
| `Enter` | Send message |
| `Shift + Enter` | New line in message |
| `Escape` | Close chat interface |
| `↑/↓` | Navigate message history |

## Visual Indicators and Emojis

The chat interface uses emojis and visual indicators to make information easier to scan and understand at a glance. Here's what each symbol means:

### Task Status Emojis

| Emoji | Status | Description |
|-------|---------|-------------|
| 📋 | **Todo** | Task is planned but not yet started |
| 🔄 | **In Progress** | Task is currently being worked on |
| ✅ | **Done** | Task has been completed |
| 🚫 | **Blocked** | Task cannot proceed due to dependencies or issues |

### Priority Indicators

| Emoji | Priority | Description |
|-------|----------|-------------|
| 🔴 | **Critical** | Urgent tasks that need immediate attention |
| 🟡 | **High** | Important tasks that should be prioritized |
| 🔵 | **Low** | Tasks that can be done when time permits |
| (none) | **Medium** | Standard priority tasks (default) |

### Example Task List Output

When you ask for tasks, you'll see output like this:

```
Here are your tasks:

• ✅ **PF-123** 🟡: Complete user authentication feature
• 🔄 **PF-124** 🔴 (due 2025-06-25): Fix critical security vulnerability  
• 📋 **PF-125**: Add project documentation
• 🚫 **PF-126** 🟡: Deploy to production (blocked by review)
```

**Reading the format:**
- **Status emoji** (✅🔄📋🚫) shows current state
- **Task ID** (PF-123) in bold for easy reference
- **Priority emoji** (🔴🟡🔵) when not medium priority
- **Due date** when applicable
- **Task title** describes the work

### Project List Format

Project listings use a clean, structured format:

```
Here are your projects:

• **ProjectFlow** (PF): Task management and project planning system
• **ABC System** (ABC): Customer journey management platform
```

**Format explanation:**
- **Project Name** in bold
- **(Prefix)** for task identification
- **Description** of the project's purpose

### Tips for Reading Responses

1. **Scan by emoji first** - Quickly identify status and priority
2. **Look for red circles** 🔴 - These need immediate attention  
3. **Check due dates** - Items with dates show urgency
4. **Use task IDs** - Reference specific tasks in follow-up messages

## Examples and Use Cases

### Daily Standups

```
What did I work on yesterday?
Show me my tasks for today
List any blockers in my current work
Create a task to review the deployment process
```

### Project Planning

```
Create an epic for the Q1 feature release
Add stories for user authentication, profile management, and settings
Set up tasks for database design and API development
Show me the project timeline
```

### Status Updates

```
Mark all my completed tasks as done
What's the progress on the mobile app epic?
Show overdue items that need attention
List high priority work for this week
```

### Team Coordination

```
What tasks are blocked and need review?
Show me all tasks assigned to the development team
Create tasks for code review and testing
List items waiting for approval
```

## Tips for Better Results

### Be Specific

✅ **Good**: "Create a high priority task to fix the login bug in the authentication module"
❌ **Avoid**: "Make something about login"

### Use Clear Actions

✅ **Good**: "Mark task PF-123 as completed"
❌ **Avoid**: "Do something with PF-123"

### Provide Context

✅ **Good**: "Show me overdue tasks in the mobile project"
❌ **Avoid**: "Show me stuff that's late"

### Use Natural Language

✅ **Good**: "What tasks are due tomorrow?"
✅ **Good**: "Tasks due tomorrow"
✅ **Good**: "Tomorrow's tasks"

## Troubleshooting

### Common Issues

**Chat doesn't respond:**
- Check your internet connection
- Verify LLM service is configured
- Try refreshing the page

**Commands not understood:**
- Try rephrasing with simpler language
- Be more specific about what you want
- Use task IDs when referencing specific items

**Missing tasks or projects:**
- Ensure you're in the correct project context
- Check spelling of project names
- Verify permissions for the requested data

### Error Messages

| Error | Solution |
|-------|----------|
| "LLM service unavailable" | Check LLM configuration and API keys |
| "Task not found" | Verify the task ID exists |
| "Insufficient permissions" | Check user access rights |
| "Invalid date format" | Use formats like "2025-12-31" or "tomorrow" |

## Configuration

### LLM Providers

The chat interface supports multiple LLM providers:

#### Groq (Recommended for Development)
```bash
export LLM_PROVIDER=groq
export LLM_API_KEY=your_groq_api_key
export LLM_MODEL=llama-3.1-8b-instant
```

#### Ollama (Recommended for Production)
```bash
export LLM_PROVIDER=ollama
export LLM_BASE_URL=http://localhost:11434
export LLM_MODEL=llama3.1
```

#### OpenAI
```bash
export LLM_PROVIDER=openai
export LLM_API_KEY=your_openai_api_key
export LLM_MODEL=gpt-4
```

#### Disabled (Testing)
```bash
export LLM_PROVIDER=disabled
```

### Performance Tuning

```bash
# Adjust response time and token limits
export LLM_TIMEOUT=30
export LLM_MAX_TOKENS=1000
```

For detailed setup instructions, see the [LLM Setup Guide](llm-setup-guide.md).

## Privacy and Security

### Data Handling

- **Local Storage**: Conversations are stored locally in your browser
- **Server Processing**: Messages are processed server-side for AI interpretation
- **No External Sharing**: Your data is not shared with external services beyond the configured LLM provider

### Best Practices

1. **Sensitive Information**: Avoid including passwords or API keys in chat messages
2. **Data Review**: Regularly review and clear conversation history if needed
3. **Access Control**: Ensure proper authentication is configured for your deployment

## Getting Help

### In-App Help

Type any of these commands for assistance:
```
help
what can you do?
show me examples
how do I create a task?
```

### Documentation Links

- [API Documentation](../README.md#api-documentation)
- [LLM Setup Guide](llm-setup-guide.md)
- [Troubleshooting Guide](troubleshooting.md)
- [Developer Guide](developer-guide.md)

### Community Support

- **GitHub Issues**: Report bugs and request features
- **Discussions**: Ask questions and share tips
- **Documentation**: Contribute to improving this guide

---

The chat interface is designed to make ProjectFlow more accessible and efficient. Start with simple commands and gradually explore more advanced features as you become comfortable with the system.
