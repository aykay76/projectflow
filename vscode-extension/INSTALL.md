# ProjectFlow VS Code Extension - Installation & Usage Guide

## Installation

### From VSIX File (Current)

1. Download the latest `projectflow-0.1.0.vsix` file
2. Open VS Code
3. Press `Ctrl+Shift+P` (Cmd+Shift+P on Mac) to open Command Palette
4. Type "Extensions: Install from VSIX"
5. Select the downloaded VSIX file
6. Reload VS Code when prompted

### From VS Code Marketplace (Coming Soon)

The extension will be available on the VS Code Marketplace once published.

## Quick Setup

### 1. Start ProjectFlow Server

Ensure your ProjectFlow server is running:

```bash
# In your ProjectFlow directory
go run cmd/server/main.go
```

The server should be accessible at `http://localhost:16191` by default.

### 2. Configure Extension

Open VS Code settings (`Ctrl+,` or `Cmd+,`) and search for "projectflow":

- **Server URL**: Set to your ProjectFlow server URL (default: `http://localhost:16191`)
- **Project**: Set to your project prefix (default: `PF`)
- **Refresh Interval**: Auto-refresh interval in milliseconds (default: 30000)

### 3. View Tasks

1. Open the Explorer panel in VS Code
2. Look for the "ProjectFlow Tasks" section
3. Click the refresh button if tasks don't appear automatically

## Usage Guide

### Viewing Tasks

- **Tree Structure**: Tasks are displayed in a hierarchical tree with parent-child relationships
- **Status Icons**: 
  - ⭕ Todo (gray circle)
  - 🔄 In Progress (blue sync icon)
  - ✅ Done (green checkmark)
  - ❌ Blocked (red error icon)
- **Priority Indicators**: 
  - 🔴 Critical
  - 🟠 High
  - 🟡 Medium
  - 🟢 Low

### Creating Tasks

1. Click the "+" button in the ProjectFlow Tasks view header
2. Enter task title (required)
3. Enter description (optional)
4. Select task type (Task, Story, Epic, Subtask)
5. Select priority (Low, Medium, High, Critical)
6. Task will be created and appear in the tree

### Managing Tasks

Right-click on any task to access context menu options:

- **Edit Task**: Modify the task title
- **Mark Complete**: Change status to "done"
- **Mark In Progress**: Change status to "in_progress"
- **Delete Task**: Remove the task (requires confirmation)

### Task Details

Click on any task to open a detailed view in a new panel showing:
- Task title and ID
- Status, priority, and type
- Full description
- Creation and update timestamps

### Project Management

- **Switch Projects**: Click the folder icon in the view header to switch between projects
- **Project Status**: Current project is shown in the view title
- **Workspace Configuration**: Each workspace can have different ProjectFlow server settings

### Connection Status

- **Status Bar**: Shows connection status in the bottom status bar
- **Green Check**: Connected to ProjectFlow server
- **Red Error**: Connection failed (click for details)
- **Manual Check**: Use "ProjectFlow: Check Connection" command

## Workspace Configuration

For project-specific settings, create `.vscode/settings.json` in your workspace:

```json
{
  "projectflow.serverUrl": "http://your-server:16191",
  "projectflow.project": "YOUR_PREFIX",
  "projectflow.refreshInterval": 60000,
  "projectflow.apiKey": "your-api-key-if-needed"
}
```

## Troubleshooting

### Extension Not Showing Tasks

1. Check that ProjectFlow server is running
2. Verify server URL in settings
3. Check connection status in status bar
4. Try refreshing tasks manually
5. Check VS Code Developer Console for errors

### Cannot Connect to Server

1. Verify ProjectFlow server is accessible at the configured URL
2. Check firewall settings
3. Ensure server is running on correct port
4. Test connection using "ProjectFlow: Check Connection" command

### Tasks Not Updating

1. Check auto-refresh interval setting
2. Manually refresh using the refresh button
3. Verify network connectivity
4. Check server logs for errors

### Performance Issues

1. Reduce refresh interval if set too low
2. Limit task hierarchy depth if very large
3. Check network latency to server

## Development & Debugging

To debug the extension:

1. Clone the ProjectFlow repository
2. Open the `vscode-extension` folder in VS Code
3. Press F5 to launch Extension Development Host
4. Test functionality in the new VS Code window

## Support

For issues, bugs, or feature requests:
- Check the ProjectFlow GitHub repository
- Review the troubleshooting section above
- Check VS Code Developer Console for error messages

## Version History

See [CHANGELOG.md](CHANGELOG.md) for detailed version history.
