# 👤 User Guide

Welcome to ProjectFlow! This comprehensive guide will help you get the most out of your workflow management experience, from basic task management to advanced natural language commands.

## Table of Contents

1. [Getting Started](#getting-started)
2. [Web Interface](#web-interface)
3. [Natural Language Chat](#natural-language-chat)
4. [Task Management](#task-management)
5. [Project Organization](#project-organization)
6. [Advanced Features](#advanced-features)
7. [Tips and Tricks](#tips-and-tricks)
8. [Troubleshooting](#troubleshooting)

## Getting Started

### First Steps

1. **Access ProjectFlow**: Open your browser and navigate to your ProjectFlow instance (typically `http://localhost:16191`)

2. **Explore the Interface**: Familiarize yourself with the main areas:
   - **Header**: Navigation and chat access
   - **Sidebar**: Project selection and filters
   - **Main Content**: Task lists and details
   - **Chat Interface**: AI-powered natural language commands

3. **Create Your First Project**: Start by organizing your work into projects

### Basic Concepts

**Projects**: Top-level containers for organizing related work
- Each project has a unique prefix (e.g., "PF", "WEB", "API")
- Projects contain tasks and help organize your workflow

**Tasks**: Individual work items with these types:
- **Epic**: Large initiatives broken down into stories
- **Story**: User-facing features or capabilities  
- **Task**: Specific work items
- **Subtask**: Smaller parts of larger tasks

**Task States**:
- **Todo**: Not yet started
- **In Progress**: Currently being worked on
- **Done**: Completed
- **Blocked**: Cannot proceed due to dependencies

**Priority Levels**:
- **Critical**: Urgent, blocking other work
- **High**: Important, should be done soon
- **Medium**: Normal priority (default)
- **Low**: Nice to have, can be delayed

### Visual Indicators

ProjectFlow uses emojis and visual indicators to make information easier to scan:

**Status Emojis**: 📋 Todo, 🔄 In Progress, ✅ Done, 🚫 Blocked  
**Priority Colors**: 🔴 Critical, 🟡 High, 🔵 Low

> 💡 **Tip**: For complete details on all visual indicators, see the [Chat Interface Guide](chat-interface-guide.md#visual-indicators-and-emojis)

## Web Interface

### Navigation

**Header Bar**:
- **ProjectFlow Logo**: Click to return to dashboard
- **Project Selector**: Switch between projects
- **Search**: Find tasks quickly
- **Chat Button (💬)**: Open natural language interface
- **Settings**: Configuration options

**Sidebar**:
- **Project List**: All available projects
- **Filters**: Filter tasks by status, priority, type
- **Quick Actions**: Common task operations

### Task Views

**List View**:
- See all tasks in a table format
- Sort by any column (title, status, priority, date)
- Quick status updates with click-to-edit
- Bulk operations on selected tasks

**Kanban Board**:
- Visual workflow with columns for each status
- Drag and drop to change task status
- Cards show key task information
- Filters apply to board view

**Hierarchy View**:
- Tree structure showing task relationships
- Expand/collapse task groups
- Clear parent-child relationships
- Helpful for epics with multiple stories

### Creating Tasks

**Quick Create**:
1. Click the **+ New Task** button
2. Enter a title (required)
3. Select project, type, and priority
4. Click **Create** or press Enter

**Detailed Create**:
1. Use the **Create Task** form
2. Fill in all relevant fields:
   - **Title**: Clear, descriptive name
   - **Description**: Detailed information
   - **Type**: Epic, Story, Task, or Subtask
   - **Priority**: Critical, High, Medium, Low
   - **Status**: Starting status
   - **Due Date**: Optional deadline
   - **Parent Task**: For subtasks and story breakdown

### Editing Tasks

**Inline Editing**:
- Click any editable field to modify
- Changes save automatically
- Supports title, status, priority

**Detailed Editing**:
- Click task title to open detail view
- Edit all fields including description
- View task history and relationships
- Add comments and attachments

## Natural Language Chat

### Getting Started with Chat

**Opening Chat**:
- Click the 💬 button in the header
- Use keyboard shortcut: `⌘+/` (Mac) or `Ctrl+/` (Windows/Linux)
- Click the floating chat button (bottom-right)

**First Conversation**:
```
You: Hi, how can I create a new task?
AI: I can help you create tasks! Just tell me what you need to do. For example, you can say:
- "Create a task to fix the login bug"
- "Add a high priority story for user registration"
- "Make a new epic for the mobile app project"

What task would you like to create?
```

### Basic Commands

**Creating Tasks**:
```
Create a task to implement user authentication
Add a high priority bug fix for the payment system
Make a new story about user profile management
Create an epic for mobile app development
```

**Listing Tasks**:
```
Show me all tasks
List tasks in the PF project
What tasks are high priority?
Show overdue tasks
Display completed tasks from last week
```

**Updating Tasks**:
```
Mark task PF-123 as done
Set the priority of PF-456 to high
Move task PF-789 to in progress
Update task PF-234 description to include API requirements
```

**Getting Information**:
```
What's the status of project ABC?
How many tasks are overdue?
Show me tasks assigned to John
What's in my todo list?
Give me a summary of completed work this week
```

### Advanced Chat Commands

**Project Management**:
```
Create a new project called "Website Redesign" with prefix WEB
Show project statistics for PF
List all projects and their task counts
Archive completed tasks in project ABC
```

**Task Relationships**:
```
Create a subtask under PF-123 to write unit tests
Make PF-456 a child of epic PF-100
Show all subtasks of PF-789
List the hierarchy for project PF
```

**Filtering and Search**:
```
Find tasks with 'authentication' in the title
Show critical tasks created this week
List blocked tasks with their reasons
Find tasks due before Friday
```

**Bulk Operations**:
```
Mark all todo tasks in project ABC as in progress
Set priority to high for all overdue tasks
Close all completed tasks older than 30 days
Update all bug tasks to high priority
```

### Chat Best Practices

**Be Specific**:
- ✅ "Create a high priority task to fix the login redirect bug"
- ❌ "Make a task for the bug"

**Use Clear Task Identifiers**:
- ✅ "Update task PF-123 status to done"  
- ❌ "Mark the login task as complete"

**Include Context**:
- ✅ "Show overdue tasks in the frontend project"
- ❌ "Show overdue tasks" (if you have multiple projects)

**Ask Follow-up Questions**:
```
You: Create a task for the API work
AI: I'd be happy to create an API task! Could you provide more details? For example:
- What specific API work needs to be done?
- Which project should this go in?
- What priority should it have?
- Are there any deadlines?
```

## Task Management

### Effective Task Creation

**Writing Good Titles**:
- ✅ "Implement user authentication with OAuth2"
- ✅ "Fix payment processing timeout on checkout"
- ✅ "Design mobile app onboarding flow"
- ❌ "Fix bug"
- ❌ "Update stuff"
- ❌ "Meeting"

**Useful Descriptions**:
```markdown
## Problem
Users are experiencing timeouts during payment processing

## Acceptance Criteria
- [ ] Payment requests complete within 10 seconds
- [ ] Error handling for timeout scenarios
- [ ] User feedback during processing
- [ ] Retry mechanism for failed payments

## Technical Notes
- Current timeout is set to 30s in payment-service.yaml
- Consider caching payment method data
- May need to optimize database queries
```

**Setting Priorities**:
- **Critical**: Production down, security issues, blocking launches
- **High**: Important features, bugs affecting many users
- **Medium**: Standard features, minor bugs, improvements
- **Low**: Nice-to-have features, technical debt, cleanup

### Task Workflow

**Typical Task Lifecycle**:
1. **Todo**: Task created, not yet started
2. **In Progress**: Actively being worked on
3. **Done**: Work completed and ready

**Advanced Workflow**:
1. **Todo**: Backlog item
2. **In Progress**: Active development
3. **Review**: Code review or testing
4. **Testing**: QA validation
5. **Done**: Completed and deployed
6. **Blocked**: Cannot proceed (dependency issues)

### Task Organization

**Using Task Types Effectively**:

**Epic** → Large initiative (3-6 months):
```
Epic: Mobile App Development
├── Story: User Registration Flow
│   ├── Task: Design registration screens
│   ├── Task: Implement API endpoints
│   └── Task: Add form validation
├── Story: Push Notifications
│   ├── Task: Set up push notification service
│   └── Task: Implement notification UI
└── Story: Offline Data Sync
    ├── Task: Design offline storage
    └── Task: Implement sync mechanism
```

**Story** → User-facing feature (1-4 weeks):
```
Story: User Profile Management
├── Task: Create profile edit form
├── Task: Add profile picture upload
├── Task: Implement privacy settings
└── Subtask: Write user acceptance tests
```

### Due Dates and Scheduling

**Setting Realistic Due Dates**:
- Consider dependencies and blockers
- Add buffer time for testing and review
- Account for other commitments
- Break large tasks into smaller ones with interim dates

**Due Date Commands**:
```
Create a task due tomorrow to review the deployment
Set due date for PF-123 to Friday
Show tasks due this week
List overdue tasks with their original due dates
```

## Project Organization

### Creating Projects

**Project Planning**:
1. **Define Scope**: What work will this project include?
2. **Choose Prefix**: Short, memorable identifier (2-4 characters)
3. **Set Description**: Clear project purpose and goals
4. **Plan Structure**: How will you organize epics and stories?

**Example Projects**:
- **WEB** - Website Redesign Project
- **API** - Backend API Development  
- **MOB** - Mobile App Project
- **SEC** - Security Improvements
- **DOC** - Documentation Updates

### Project Structure Best Practices

**Small Project** (1-3 months):
```
Project: API (Backend API Development)
├── Epic: Authentication System
│   ├── Story: User Login/Logout
│   ├── Story: Password Reset
│   └── Story: OAuth Integration
├── Epic: Data Management
│   ├── Story: User CRUD Operations
│   ├── Story: Data Validation
│   └── Story: Database Optimization
└── Epic: API Documentation
    ├── Story: OpenAPI Specification
    └── Story: Developer Examples
```

**Large Project** (6+ months):
```
Project: WEB (Website Redesign)
├── Epic: User Experience Research
├── Epic: Design System
├── Epic: Frontend Implementation
├── Epic: Backend Integration
├── Epic: Performance Optimization
├── Epic: Testing and QA
└── Epic: Deployment and Launch
```

### Multi-Project Management

**Cross-Project Dependencies**:
```
You: Create a task in project API that depends on task WEB-45
AI: I've created task API-123 "Implement user endpoint" with a dependency note on WEB-45. I'll include this in the task description so you can track the dependency.
```

**Project Status Tracking**:
```
You: Give me a status report for all projects
AI: Here's your project status:

**WEB (Website Redesign)**: 12 tasks, 60% complete
- 2 blocked tasks (waiting for design approval)
- 3 in progress, 7 done, 2 todo

**API (Backend Development)**: 8 tasks, 25% complete  
- 1 critical task overdue
- 2 in progress, 2 done, 4 todo

**MOB (Mobile App)**: 15 tasks, 80% complete
- On track for release next week
- 1 in progress, 12 done, 2 todo
```

## Advanced Features

### Keyboard Shortcuts

**Global Shortcuts**:
- `⌘+/` or `Ctrl+/` - Toggle chat interface
- `⌘+K` or `Ctrl+K` - Quick search
- `⌘+N` or `Ctrl+N` - New task
- `Escape` - Close modals/panels

**Task List Shortcuts**:
- `↑/↓` - Navigate tasks
- `Enter` - Open task details  
- `Space` - Select/deselect task
- `Delete` - Delete selected tasks

**Chat Shortcuts**:
- `↑` - Previous message in history
- `↓` - Next message in history
- `⌘+Enter` or `Ctrl+Enter` - Send message
- `Escape` - Close chat

### Filtering and Search

**Basic Filters**:
- **Status**: Show only tasks with specific status
- **Priority**: Filter by priority level
- **Type**: Show only epics, stories, tasks, or subtasks
- **Project**: Limit to specific project

**Advanced Search**:
```
Search for: status:todo priority:high
Results: All high-priority todo items

Search for: created:2025-01-01..2025-01-31
Results: Tasks created in January 2025

Search for: assignee:john overdue:true
Results: John's overdue tasks
```

**Chat-Based Filtering**:
```
Show me high priority tasks created this week
Find blocked tasks with 'API' in the title
List overdue tasks in the frontend project
Display completed stories from last month
```

### Automation and Integrations

**Automated Actions**:
- **Auto-status Updates**: Moving to "In Progress" when work starts
- **Due Date Notifications**: Reminders for upcoming deadlines
- **Completion Cascading**: Mark parent tasks done when all children complete

**Integration Points**:
- **Git Hooks**: Create tasks from commit messages
- **CI/CD**: Update task status based on deployment success
- **Calendar**: Sync due dates with calendar applications
- **Slack/Teams**: Task notifications in team channels

### Reporting and Analytics

**Chat-Based Reports**:
```
You: Give me a productivity report for last week
AI: **Productivity Report (Week of Jan 15-21, 2025)**

**Tasks Completed**: 23 tasks across 3 projects
- 8 stories, 12 tasks, 3 subtasks
- Average completion time: 2.3 days

**Top Performing Project**: API (8 tasks completed)
**Longest Running Task**: WEB-45 (14 days in progress)

**Trends**:
- 15% increase from previous week
- Bug resolution time improved by 2 days
- 3 tasks carried over from previous sprint
```

**Visual Analytics**:
- Task completion trends over time
- Priority distribution across projects
- Average task lifecycle duration
- Bottleneck identification

## Tips and Tricks

### Productivity Tips

**Daily Workflow**:
1. **Morning Review**: Check overdue and high-priority tasks
2. **Chat Check-in**: "What should I work on today?"
3. **Regular Updates**: Move tasks to "In Progress" when starting
4. **End-of-Day**: Update progress and plan tomorrow

**Task Creation Efficiency**:
```
# Quick batch creation via chat
Create these tasks for project API:
1. Set up authentication middleware
2. Implement user CRUD operations  
3. Add input validation
4. Write API documentation
5. Set up error handling
```

**Smart Organization**:
- Use consistent naming conventions
- Create templates for common task types
- Break large tasks into 1-2 day chunks
- Group related tasks under stories/epics

### Chat Power User Tips

**Context Management**:
```
# Set project context
I'm working on project WEB today

# Later in conversation...
Create a high priority task to fix the header bug
# AI understands this goes in project WEB
```

**Batch Operations**:
```
# Update multiple tasks at once
Set these tasks to high priority: PF-123, PF-124, PF-125

# Create related tasks
Create subtasks for PF-100:
- Design user interface
- Implement backend logic
- Write unit tests
- Update documentation
```

**Smart Queries**:
```
# Relative date queries
Show tasks completed in the last 2 weeks
Find tasks due in the next 3 days
List tasks created yesterday

# Comparison queries  
Which project has the most overdue tasks?
Compare completion rates between projects
Show trends for this month vs last month
```

### Collaboration Tips

**Task Communication**:
- Use clear, descriptive titles
- Include context in descriptions
- Update status regularly
- Add comments for important changes

**Team Coordination**:
```
# Check team status
Show all tasks assigned to the frontend team
What's blocking the backend team?
Which tasks need review this week?

# Plan team work
Create tasks for next sprint:
- User story estimation
- Sprint planning meeting
- Demo preparation
```

**Knowledge Sharing**:
- Link to relevant documentation
- Include acceptance criteria
- Add technical notes for complex tasks
- Reference related tasks and dependencies

## Troubleshooting

### Common Issues

**Chat Not Responding**:
1. Check internet connection
2. Refresh the page
3. Clear browser cache
4. Check for JavaScript errors in console

**Tasks Not Saving**:
1. Check network connectivity
2. Verify server is running
3. Look for error messages
4. Try refreshing and re-entering data

**Search Not Working**:
1. Try simpler search terms
2. Check spelling and filters
3. Clear search and try again
4. Use chat interface as alternative

### Getting Help

**Built-in Help**:
- Hover tooltips on interface elements
- Help links in settings menu
- Chat interface can answer questions

**Self-Service Resources**:
- [Troubleshooting Guide](troubleshooting.md)
- [Chat Interface Guide](chat-interface-guide.md)
- [Developer Guide](developer-guide.md)

**Community Support**:
- GitHub Issues for bug reports
- Discussion forums for questions
- Community wiki for tips and tricks

**Professional Support**:
- Available for enterprise deployments
- Custom training and setup assistance
- Priority bug fixes and feature requests

---

## Quick Reference

### Essential Chat Commands

```bash
# Task Creation
"Create a [priority] [type] to [description]"
"Add task: [title] in project [prefix]"

# Task Updates  
"Mark task [ID] as [status]"
"Set priority of [ID] to [priority]"
"Update [ID] description to [text]"

# Queries
"Show [filter] tasks [in project]"
"What's the status of [project/task]?"
"List [type] tasks due [timeframe]"

# Project Management
"Create project [name] with prefix [code]"  
"Show project stats for [prefix]"
"List all projects"
```

### Status Quick Actions

| Status | Meaning | Next Steps |
|--------|---------|------------|
| **Todo** | Not started | Move to "In Progress" when beginning |
| **In Progress** | Active work | Update regularly, move to "Done" when complete |
| **Done** | Completed | Archive periodically, celebrate! |
| **Blocked** | Cannot proceed | Identify and resolve blockers |

### Priority Guidelines

| Priority | Use When | Examples |
|----------|----------|----------|
| **Critical** | Production issues, security | "Site down", "Data breach" |
| **High** | Important features, major bugs | "Payment not working", "Key feature" |
| **Medium** | Normal work items | "New feature", "Minor bug" |
| **Low** | Nice-to-have, cleanup | "Code refactor", "Documentation" |

---

**Ready to get started?** Open the chat interface and type "Hi, what can I help you with today?" to begin your ProjectFlow journey!
