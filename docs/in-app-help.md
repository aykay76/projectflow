# 💡 In-App Help System

This document describes the help system integrated into ProjectFlow's web interface, including tooltips, guided tours, and contextual help.

## Help System Components

### 1. Tooltips and Hints

**Chat Interface**:
- Chat button: "Open natural language interface (⌘+/ or Ctrl+/)"
- Chat input: "Type commands like 'Create a task to fix the login bug' or 'Show high priority tasks'"
- Send button: "Send message (Enter or click to send)"
- Chat toggle: "Minimize/maximize chat panel"

**Task Management**:
- New Task button: "Create a new task (⌘+N or Ctrl+N)"
- Task status: "Click to change status: Todo → In Progress → Done"
- Task priority: "Set priority: Low | Medium | High | Critical"
- Task type: "Epic (large initiative) | Story (feature) | Task (work item) | Subtask (small part)"

**Project Management**:
- Project selector: "Switch between projects or create new ones"
- Project prefix: "Short identifier used in task IDs (e.g., PF-123)"
- Filter panel: "Filter tasks by status, priority, type, or custom criteria"

### 2. Contextual Help Messages

**Empty States**:
```html
<!-- No tasks in project -->
<div class="empty-state">
  <h3>No tasks yet</h3>
  <p>Get started by creating your first task:</p>
  <ul>
    <li>Click the "New Task" button above</li>
    <li>Use the chat: "Create a task to implement user login"</li>
    <li>Import tasks from your backlog</li>
  </ul>
</div>

<!-- No projects -->
<div class="empty-state">
  <h3>Welcome to ProjectFlow!</h3>
  <p>Create your first project to get started:</p>
  <ol>
    <li>Click "New Project" or use chat: "Create a project called Website"</li>
    <li>Choose a short prefix (e.g., WEB, API, DOC)</li>
    <li>Start adding tasks and organizing your work</li>
  </ol>
</div>

<!-- Chat first time -->
<div class="chat-welcome">
  <h4>👋 Welcome to ProjectFlow Chat!</h4>
  <p>I can help you manage tasks and projects using natural language. Try:</p>
  <div class="example-commands">
    <button class="example-cmd">"Create a task to fix the login bug"</button>
    <button class="example-cmd">"Show me all high priority tasks"</button>
    <button class="example-cmd">"What should I work on next?"</button>
  </div>
</div>
```

### 3. Quick Start Guide

**First Visit Flow**:
1. **Welcome Modal**: Introduction to ProjectFlow features
2. **Project Creation**: Guided project setup
3. **First Task**: Help creating initial task
4. **Chat Introduction**: Brief chat interface tour
5. **Next Steps**: Links to documentation and advanced features

### 4. Keyboard Shortcuts Help

**Shortcut Panel** (accessible via `?` key):
```
Global Shortcuts:
⌘+/ or Ctrl+/    Toggle chat interface
⌘+K or Ctrl+K    Quick search tasks
⌘+N or Ctrl+N    Create new task
Escape           Close panels/modals
?                Show this help panel

Task List:
↑/↓              Navigate tasks
Enter            Open task details
Space            Select/deselect task
Delete           Delete selected tasks

Chat Interface:
↑/↓              Message history
⌘+Enter          Send message
Escape           Close chat panel
```

### 5. Interactive Examples

**Chat Command Examples**:
```javascript
// Interactive examples that users can click to try
const chatExamples = [
  {
    category: "Creating Tasks",
    examples: [
      "Create a high priority task to fix the login bug",
      "Add a story for user profile management", 
      "Make a new epic for mobile app development"
    ]
  },
  {
    category: "Managing Tasks", 
    examples: [
      "Mark task PF-123 as done",
      "Set priority of PF-456 to high",
      "Show me all tasks in the API project"
    ]
  },
  {
    category: "Getting Information",
    examples: [
      "What tasks are overdue?",
      "Show me completed work from last week",
      "Give me a project status summary"
    ]
  }
];
```

## Implementation Guidelines

### 1. Help Content Integration

**HTML Data Attributes**:
```html
<!-- Add help attributes to elements -->
<button 
  class="chat-toggle-btn"
  data-help-text="Open natural language interface (⌘+/ or Ctrl+/)"
  data-help-category="chat"
  aria-label="Open chat interface">
  💬
</button>

<select 
  class="task-status"
  data-help-text="Task status: Todo (not started) → In Progress (active) → Done (complete)"
  data-help-category="tasks">
  <option value="todo">Todo</option>
  <option value="in_progress">In Progress</option>
  <option value="done">Done</option>
</select>
```

### 2. Help System JavaScript

**Help Manager**:
```javascript
class HelpManager {
  constructor() {
    this.isEnabled = localStorage.getItem('help-enabled') !== 'false';
    this.initializeTooltips();
    this.setupKeyboardShortcuts();
  }

  initializeTooltips() {
    if (!this.isEnabled) return;
    
    document.querySelectorAll('[data-help-text]').forEach(element => {
      this.attachTooltip(element);
    });
  }

  attachTooltip(element) {
    const helpText = element.dataset.helpText;
    const category = element.dataset.helpCategory;
    
    // Create tooltip element
    const tooltip = document.createElement('div');
    tooltip.className = 'help-tooltip';
    tooltip.textContent = helpText;
    
    // Position and show/hide logic
    element.addEventListener('mouseenter', () => this.showTooltip(element, tooltip));
    element.addEventListener('mouseleave', () => this.hideTooltip(tooltip));
  }

  showTooltip(element, tooltip) {
    document.body.appendChild(tooltip);
    const rect = element.getBoundingClientRect();
    tooltip.style.left = rect.left + 'px';
    tooltip.style.top = (rect.bottom + 5) + 'px';
    tooltip.classList.add('visible');
  }

  hideTooltip(tooltip) {
    tooltip.classList.remove('visible');
    setTimeout(() => {
      if (tooltip.parentNode) {
        tooltip.parentNode.removeChild(tooltip);
      }
    }, 200);
  }

  setupKeyboardShortcuts() {
    document.addEventListener('keydown', (e) => {
      // Show shortcuts help with '?'
      if (e.key === '?' && !this.isInputFocused()) {
        this.showShortcutsPanel();
      }
      
      // Toggle help system with F1
      if (e.key === 'F1') {
        e.preventDefault();
        this.toggleHelpSystem();
      }
    });
  }

  showShortcutsPanel() {
    const panel = document.createElement('div');
    panel.className = 'shortcuts-panel modal';
    panel.innerHTML = `
      <div class="modal-content">
        <div class="modal-header">
          <h3>Keyboard Shortcuts</h3>
          <button class="close-btn" onclick="this.closest('.modal').remove()">×</button>
        </div>
        <div class="modal-body">
          <div class="shortcut-section">
            <h4>Global</h4>
            <div class="shortcut-item">
              <kbd>⌘</kbd> + <kbd>/</kbd> or <kbd>Ctrl</kbd> + <kbd>/</kbd>
              <span>Toggle chat interface</span>
            </div>
            <!-- More shortcuts... -->
          </div>
        </div>
      </div>
    `;
    document.body.appendChild(panel);
  }

  toggleHelpSystem() {
    this.isEnabled = !this.isEnabled;
    localStorage.setItem('help-enabled', this.isEnabled.toString());
    
    if (this.isEnabled) {
      this.initializeTooltips();
      this.showNotification('Help system enabled');
    } else {
      this.hideAllTooltips();
      this.showNotification('Help system disabled');
    }
  }
}

// Initialize help system
document.addEventListener('DOMContentLoaded', () => {
  window.helpManager = new HelpManager();
});
```

### 3. CSS Styling

**Help System Styles**:
```css
/* Tooltip styling */
.help-tooltip {
  position: absolute;
  background: var(--surface-elevated);
  color: var(--text-primary);
  padding: 8px 12px;
  border-radius: 6px;
  font-size: 14px;
  max-width: 300px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  border: 1px solid var(--border-light);
  z-index: 1000;
  opacity: 0;
  transform: translateY(-5px);
  transition: opacity 0.2s ease, transform 0.2s ease;
  pointer-events: none;
}

.help-tooltip.visible {
  opacity: 1;
  transform: translateY(0);
}

.help-tooltip::before {
  content: '';
  position: absolute;
  top: -5px;
  left: 20px;
  width: 0;
  height: 0;
  border-left: 5px solid transparent;
  border-right: 5px solid transparent;
  border-bottom: 5px solid var(--surface-elevated);
}

/* Help-enabled elements highlighting */
body.help-mode [data-help-text] {
  position: relative;
  outline: 2px dashed var(--accent-color);
  outline-offset: 2px;
}

/* Shortcuts panel */
.shortcuts-panel {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 2000;
}

.shortcuts-panel .modal-content {
  background: var(--surface-primary);
  border-radius: 8px;
  max-width: 600px;
  max-height: 80vh;
  overflow-y: auto;
}

.shortcut-section {
  margin-bottom: 24px;
}

.shortcut-section h4 {
  margin-bottom: 12px;
  color: var(--text-primary);
  font-weight: 600;
}

.shortcut-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 0;
  border-bottom: 1px solid var(--border-light);
}

.shortcut-item:last-child {
  border-bottom: none;
}

.shortcut-item kbd {
  background: var(--surface-elevated);
  border: 1px solid var(--border-medium);
  border-radius: 4px;
  padding: 2px 6px;
  font-size: 12px;
  font-family: monospace;
}

/* Welcome and empty states */
.empty-state {
  text-align: center;
  padding: 48px 24px;
  color: var(--text-secondary);
}

.empty-state h3 {
  margin-bottom: 16px;
  color: var(--text-primary);
}

.empty-state ul, .empty-state ol {
  text-align: left;
  display: inline-block;
  margin: 16px 0;
}

.chat-welcome {
  background: var(--surface-elevated);
  border: 1px solid var(--border-light);
  border-radius: 8px;
  padding: 16px;
  margin-bottom: 16px;
}

.example-commands {
  margin-top: 12px;
}

.example-cmd {
  display: block;
  width: 100%;
  text-align: left;
  padding: 8px 12px;
  margin: 4px 0;
  background: var(--surface-primary);
  border: 1px solid var(--border-light);
  border-radius: 4px;
  cursor: pointer;
  transition: background-color 0.2s ease;
}

.example-cmd:hover {
  background: var(--surface-hover);
}

/* Help toggle button */
.help-toggle {
  position: fixed;
  bottom: 20px;
  left: 20px;
  background: var(--accent-color);
  color: white;
  border: none;
  border-radius: 50%;
  width: 48px;
  height: 48px;
  font-size: 18px;
  cursor: pointer;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  z-index: 999;
  transition: transform 0.2s ease;
}

.help-toggle:hover {
  transform: scale(1.1);
}

/* Guided tour highlight */
.tour-highlight {
  position: relative;
  z-index: 1001;
}

.tour-highlight::after {
  content: '';
  position: absolute;
  top: -4px;
  left: -4px;
  right: -4px;
  bottom: -4px;
  background: var(--accent-color);
  opacity: 0.2;
  border-radius: 6px;
  animation: pulse 2s infinite;
}

@keyframes pulse {
  0% { opacity: 0.2; }
  50% { opacity: 0.4; }
  100% { opacity: 0.2; }
}
```

### 4. Context-Aware Help

**Dynamic Help Content**:
```javascript
class ContextualHelp {
  getHelpForCurrentView() {
    const currentPath = window.location.pathname;
    const activeProject = this.getCurrentProject();
    const taskCount = this.getTaskCount();
    
    if (taskCount === 0) {
      return this.getEmptyStateHelp();
    }
    
    if (this.isFirstTimeUser()) {
      return this.getOnboardingHelp();
    }
    
    return this.getStandardHelp();
  }

  getEmptyStateHelp() {
    return {
      title: "Getting Started",
      content: [
        "Create your first task to begin organizing your work",
        "Use the chat interface for quick task creation", 
        "Try: 'Create a task to set up the project structure'"
      ],
      actions: [
        { label: "Create Task", action: "openTaskForm" },
        { label: "Try Chat", action: "openChat" }
      ]
    };
  }

  getOnboardingHelp() {
    return {
      title: "Welcome to ProjectFlow!",
      content: [
        "ProjectFlow helps you manage tasks and projects efficiently",
        "Use the natural language chat for quick commands",
        "Organize work into Projects → Epics → Stories → Tasks"
      ],
      actions: [
        { label: "Take Tour", action: "startGuidedTour" },
        { label: "View Examples", action: "showExamples" }
      ]
    };
  }
}
```

## Help Content Updates

### 1. Version-Specific Help

When new features are added, update help content:

```javascript
const helpContent = {
  version: "1.2.0",
  features: {
    "natural-language-chat": {
      introduced: "1.2.0",
      helpText: "Use conversational commands to manage tasks",
      examples: ["Create a task", "Show overdue items", "Project status"]
    }
  }
};
```

### 2. Progressive Disclosure

Show help based on user experience level:

```javascript
class ProgressiveHelp {
  getHelpLevel() {
    const userStats = this.getUserStats();
    
    if (userStats.tasksCreated < 5) return 'beginner';
    if (userStats.tasksCreated < 50) return 'intermediate';
    return 'advanced';
  }

  getRelevantTips() {
    const level = this.getHelpLevel();
    
    switch (level) {
      case 'beginner':
        return this.getBeginnerTips();
      case 'intermediate':
        return this.getIntermediateTips();
      case 'advanced':
        return this.getAdvancedTips();
    }
  }
}
```

## Accessibility Considerations

### 1. Screen Reader Support

```html
<!-- ARIA labels for help elements -->
<button 
  class="help-btn"
  aria-label="Get help with this feature"
  aria-describedby="help-tooltip-123">
  ?
</button>

<div id="help-tooltip-123" role="tooltip" aria-live="polite">
  This feature helps you create tasks quickly using natural language
</div>
```

### 2. Keyboard Navigation

- All help elements are keyboard accessible
- Tab order includes help buttons and tooltips
- Escape key closes help panels
- Arrow keys navigate help content

### 3. High Contrast Support

```css
@media (prefers-contrast: high) {
  .help-tooltip {
    background: #000;
    color: #fff;
    border: 2px solid #fff;
  }
  
  .help-tooltip::before {
    border-bottom-color: #000;
  }
}
```

## Testing Help System

### 1. User Testing Scenarios

Test help system with different user types:
- First-time users (onboarding flow)
- Returning users (contextual help)
- Power users (advanced shortcuts)

### 2. Automated Testing

```javascript
// Test help tooltip display
test('shows help tooltip on hover', async () => {
  const element = screen.getByTestId('chat-button');
  userEvent.hover(element);
  
  await waitFor(() => {
    expect(screen.getByRole('tooltip')).toBeInTheDocument();
  });
});

// Test keyboard shortcuts
test('opens shortcuts panel with ? key', async () => {
  userEvent.keyboard('?');
  
  await waitFor(() => {
    expect(screen.getByText('Keyboard Shortcuts')).toBeInTheDocument();
  });
});
```

---

This in-app help system provides comprehensive guidance while maintaining a clean, unobtrusive interface. Users can access help when needed without cluttering the main interface.
