# Changelog

All notable changes to the ProjectFlow VS Code extension will be documented in this file.

## [0.1.0] - 2025-07-07

### Added
- Initial release of ProjectFlow VS Code extension
- Hierarchical task tree view in VS Code Explorer
- Task status indicators with colored icons
- Priority indicators with emoji
- Task management commands:
  - Create tasks with title, description, type, and priority selection
  - Edit task titles
  - Mark tasks as complete or in progress
  - Delete tasks with confirmation
- Context menu actions for tasks
- Task detail view in webview panel
- Connection status indicator in status bar
- Project switching functionality
- Auto-refresh capability
- Workspace-specific configuration support
- Settings for server URL, project, API key, and refresh interval

### Features
- **Tree View**: Displays tasks in hierarchical structure
- **Status Icons**: Visual indicators for todo, in-progress, done, and blocked tasks
- **Priority Colors**: Emoji indicators for critical, high, medium, and low priority
- **Rich Tooltips**: Detailed task information on hover
- **Connection Management**: Real-time connection status with error reporting
- **Configuration**: Per-workspace settings for different ProjectFlow servers

### Commands
- `ProjectFlow: Refresh Tasks` - Manually refresh the task tree
- `ProjectFlow: Create Task` - Create a new task with guided prompts
- `ProjectFlow: Check Connection` - Test connection to ProjectFlow server
- `ProjectFlow: Switch Project` - Change active project

### Configuration Options
- `projectflow.enabled` - Enable/disable extension (default: true)
- `projectflow.serverUrl` - ProjectFlow server URL (default: http://localhost:16191)
- `projectflow.project` - Default project prefix (default: PF)
- `projectflow.apiKey` - Optional API key for authentication
- `projectflow.refreshInterval` - Auto-refresh interval in milliseconds (default: 30000)

### Technical Details
- Built with TypeScript and VS Code Extension API
- Uses Axios for HTTP communication with ProjectFlow REST API
- Implements VS Code TreeDataProvider for hierarchical display
- Supports workspace-specific configuration
- Includes comprehensive error handling and user feedback
