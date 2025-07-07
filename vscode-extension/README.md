# ProjectFlow VS Code Extension

A VS Code extension that provides hierarchical task management directly within your IDE, connecting to your ProjectFlow server.

## Features

- **Hierarchical Task View**: Display your ProjectFlow tasks in a tree structure within the VS Code Explorer
- **Status Indicators**: Visual icons showing task status (todo, in-progress, done, blocked)
- **Task Management**: Create, edit, delete, and update task status directly from VS Code
- **Quick Actions**: Context menu actions for common task operations
- **Auto-refresh**: Automatically sync with your ProjectFlow server
- **Workspace Configuration**: Per-workspace ProjectFlow server settings

## Requirements

- A running ProjectFlow server
- Network access to the ProjectFlow API

## Installation

1. Install the extension from the VS Code marketplace (coming soon)
2. Configure your ProjectFlow server URL in settings
3. Start managing tasks directly from VS Code!

## Configuration

The extension can be configured through VS Code settings:

### Extension Settings

- `projectflow.enabled`: Enable/disable the ProjectFlow extension (default: true)
- `projectflow.serverUrl`: URL of your ProjectFlow server (default: http://localhost:16191)
- `projectflow.project`: Default project ID (default: PF)
- `projectflow.refreshInterval`: Auto-refresh interval in milliseconds, 0 to disable (default: 30000)

### Workspace Settings

Add these settings to your workspace settings (`.vscode/settings.json`) to configure per-project:

```json
{
  "projectflow.serverUrl": "http://your-projectflow-server:16191",
  "projectflow.project": "YOUR_PROJECT_PREFIX",
  "projectflow.refreshInterval": 60000
}
```

## Usage

### Viewing Tasks

1. Open the Explorer panel in VS Code
2. Look for the "ProjectFlow Tasks" section
3. Expand the tree to see your task hierarchy

### Creating Tasks

1. Click the "+" icon in the ProjectFlow Tasks view header
2. Enter the task title and optional description
3. The task will be created and appear in the tree

### Managing Tasks

Right-click on any task to access context menu options:
- **Edit Task**: Modify the task title
- **Mark Complete**: Set task status to "done"
- **Mark In Progress**: Set task status to "in_progress"
- **Delete Task**: Remove the task (with confirmation)

### Task Details

Click on any task to open a detailed view in a new panel.

## Commands

The extension provides the following commands (available via Command Palette):

- `ProjectFlow: Refresh Tasks` - Manually refresh the task tree
- `ProjectFlow: Create Task` - Create a new task

## Development

### Building from Source

1. Clone the repository
2. Navigate to the `vscode-extension` directory
3. Install dependencies: `npm install`
4. Compile: `npm run compile`
5. Press F5 to launch a new Extension Development Host

### Testing

Run tests with: `npm test`

## Known Issues

- Large task hierarchies may take time to load
- Network connectivity issues will prevent task synchronization

## Release Notes

### 0.1.0

Initial release of ProjectFlow VS Code extension:
- Hierarchical task tree view
- Basic task management operations
- Server configuration options
- Auto-refresh functionality

---

## Contributing

Contributions are welcome! Please see the main ProjectFlow repository for contribution guidelines.

## License

This project is licensed under the MIT License.
