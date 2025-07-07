import * as vscode from 'vscode';
import { ProjectFlowTaskProvider } from './taskProvider';
import { ProjectFlowApiClient } from './apiClient';

let taskProvider: ProjectFlowTaskProvider;
let statusBarItem: vscode.StatusBarItem;

export function activate(context: vscode.ExtensionContext) {
	console.log('ProjectFlow extension is now active!');

	// Initialize the API client
	const apiClient = new ProjectFlowApiClient();

	// Create status bar item
	statusBarItem = vscode.window.createStatusBarItem(vscode.StatusBarAlignment.Left, 100);
	statusBarItem.command = 'projectflow.checkConnection';
	context.subscriptions.push(statusBarItem);

	// Check initial connection
	updateConnectionStatus(apiClient);

	// Create the task provider
	taskProvider = new ProjectFlowTaskProvider(apiClient);

	// Register the tree data provider
	const treeView = vscode.window.createTreeView('projectflow-tasks', {
		treeDataProvider: taskProvider,
		showCollapseAll: true
	});

	// Show welcome message when view is first opened
	treeView.onDidChangeVisibility((e) => {
		if (e.visible) {
			updateConnectionStatus(apiClient);
		}
	});

	// Register commands
	const refreshCommand = vscode.commands.registerCommand('projectflow.refreshTasks', () => {
		taskProvider.refresh();
	});

	const createTaskCommand = vscode.commands.registerCommand('projectflow.createTask', async () => {
		const title = await vscode.window.showInputBox({
			prompt: 'Enter task title',
			placeHolder: 'Task title',
			validateInput: (value) => {
				return value.trim().length === 0 ? 'Title cannot be empty' : undefined;
			}
		});

		if (!title) {
			return;
		}

		const description = await vscode.window.showInputBox({
			prompt: 'Enter task description (optional)',
			placeHolder: 'Task description'
		});

		const typeItems = [
			{ label: 'Task', description: 'A standard task' },
			{ label: 'Story', description: 'A user story' },
			{ label: 'Epic', description: 'A large feature or initiative' },
			{ label: 'Subtask', description: 'A subtask of another task' }
		];

		const selectedType = await vscode.window.showQuickPick(typeItems, {
			placeHolder: 'Select task type'
		});

		const priorityItems = [
			{ label: 'Low', description: '🟢 Low priority' },
			{ label: 'Medium', description: '🟡 Medium priority' },
			{ label: 'High', description: '🟠 High priority' },
			{ label: 'Critical', description: '🔴 Critical priority' }
		];

		const selectedPriority = await vscode.window.showQuickPick(priorityItems, {
			placeHolder: 'Select priority'
		});

		if (!selectedType || !selectedPriority) {
			return;
		}

		try {
			await apiClient.createTask({
				title,
				description: description || '',
				status: 'todo',
				priority: selectedPriority.label.toLowerCase(),
				type: selectedType.label.toLowerCase()
			});
			taskProvider.refresh();
			vscode.window.showInformationMessage(`Task "${title}" created successfully!`);
		} catch (error) {
			vscode.window.showErrorMessage(`Failed to create task: ${error}`);
		}
	});

	const editTaskCommand = vscode.commands.registerCommand('projectflow.editTask', async (task) => {
		if (!task) {
			return;
		}

		const newTitle = await vscode.window.showInputBox({
			prompt: 'Edit task title',
			value: task.title
		});

		if (newTitle && newTitle !== task.title) {
			try {
				await apiClient.updateTask(task.id, { title: newTitle });
				taskProvider.refresh();
				vscode.window.showInformationMessage('Task updated successfully!');
			} catch (error) {
				vscode.window.showErrorMessage(`Failed to update task: ${error}`);
			}
		}
	});

	const deleteTaskCommand = vscode.commands.registerCommand('projectflow.deleteTask', async (task) => {
		if (!task) {
			return;
		}

		const confirmation = await vscode.window.showWarningMessage(
			`Are you sure you want to delete "${task.title}"?`,
			'Yes',
			'No'
		);

		if (confirmation === 'Yes') {
			try {
				await apiClient.deleteTask(task.id);
				taskProvider.refresh();
				vscode.window.showInformationMessage('Task deleted successfully!');
			} catch (error) {
				vscode.window.showErrorMessage(`Failed to delete task: ${error}`);
			}
		}
	});

	const markCompleteCommand = vscode.commands.registerCommand('projectflow.markTaskComplete', async (task) => {
		if (!task) {
			return;
		}

		try {
			await apiClient.updateTask(task.id, { status: 'done' });
			taskProvider.refresh();
			vscode.window.showInformationMessage('Task marked as complete!');
		} catch (error) {
			vscode.window.showErrorMessage(`Failed to update task: ${error}`);
		}
	});

	const markInProgressCommand = vscode.commands.registerCommand('projectflow.markTaskInProgress', async (task) => {
		if (!task) {
			return;
		}

		try {
			await apiClient.updateTask(task.id, { status: 'in_progress' });
			taskProvider.refresh();
			vscode.window.showInformationMessage('Task marked as in progress!');
		} catch (error) {
			vscode.window.showErrorMessage(`Failed to update task: ${error}`);
		}
	});

	const openTaskCommand = vscode.commands.registerCommand('projectflow.openTask', async (task) => {
		if (!task) {
			return;
		}

		// Create a webview to show task details
		const panel = vscode.window.createWebviewPanel(
			'projectflowTask',
			`Task: ${task.title}`,
			vscode.ViewColumn.Two,
			{
				enableScripts: true
			}
		);

		panel.webview.html = getTaskDetailHtml(task);
	});

	const checkConnectionCommand = vscode.commands.registerCommand('projectflow.checkConnection', async () => {
		await updateConnectionStatus(apiClient);
		const status = await apiClient.getConnectionStatus();
		if (status.connected) {
			vscode.window.showInformationMessage(`Connected to ProjectFlow at ${apiClient.getServerUrl()}`);
		} else {
			vscode.window.showErrorMessage(`Failed to connect to ProjectFlow at ${apiClient.getServerUrl()}: ${status.error}`);
		}
	});

	const switchProjectCommand = vscode.commands.registerCommand('projectflow.switchProject', async () => {
		try {
			const projects = await apiClient.getProjects();
			interface ProjectPickItem extends vscode.QuickPickItem {
				project: any;
			}
			
			const items: ProjectPickItem[] = projects.map(p => ({
				label: p.name,
				description: p.display_prefix,
				detail: p.description,
				project: p
			}));

			const selected = await vscode.window.showQuickPick(items, {
				placeHolder: 'Select a project'
			});

			if (selected) {
				const config = vscode.workspace.getConfiguration('projectflow');
				await config.update('project', selected.project.display_prefix, vscode.ConfigurationTarget.Workspace);
				taskProvider.refresh();
				vscode.window.showInformationMessage(`Switched to project: ${selected.project.name}`);
			}
		} catch (error) {
			vscode.window.showErrorMessage(`Failed to load projects: ${error}`);
		}
	});

	// Register all commands and providers
	context.subscriptions.push(
		treeView,
		statusBarItem,
		refreshCommand,
		createTaskCommand,
		editTaskCommand,
		deleteTaskCommand,
		markCompleteCommand,
		markInProgressCommand,
		openTaskCommand,
		checkConnectionCommand,
		switchProjectCommand
	);

	// Auto-refresh based on configuration
	const config = vscode.workspace.getConfiguration('projectflow');
	const refreshInterval = config.get<number>('refreshInterval', 30000);
	
	if (refreshInterval > 0) {
		setInterval(() => {
			taskProvider.refresh();
		}, refreshInterval);
	}

	// Watch for configuration changes
	vscode.workspace.onDidChangeConfiguration(event => {
		if (event.affectsConfiguration('projectflow')) {
			// Update API client configuration when ProjectFlow settings change
			apiClient.updateConfiguration();
			taskProvider.refresh();
		}
	});
}

export function deactivate() {
	// Clean up resources
	if (statusBarItem) {
		statusBarItem.dispose();
	}
}

async function updateConnectionStatus(apiClient: ProjectFlowApiClient) {
	const status = await apiClient.getConnectionStatus();
	if (status.connected) {
		statusBarItem.text = `$(check) ProjectFlow`;
		statusBarItem.tooltip = `Connected to ${apiClient.getServerUrl()}`;
		statusBarItem.backgroundColor = undefined;
	} else {
		statusBarItem.text = `$(error) ProjectFlow`;
		statusBarItem.tooltip = `Disconnected: ${status.error}`;
		statusBarItem.backgroundColor = new vscode.ThemeColor('statusBarItem.errorBackground');
	}
	statusBarItem.show();
}

function getTaskDetailHtml(task: any): string {
	return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Task Details</title>
    <style>
        body {
            font-family: var(--vscode-font-family);
            padding: 20px;
            color: var(--vscode-foreground);
            background-color: var(--vscode-editor-background);
        }
        .task-header {
            border-bottom: 1px solid var(--vscode-panel-border);
            padding-bottom: 10px;
            margin-bottom: 20px;
        }
        .task-title {
            font-size: 1.5em;
            font-weight: bold;
            margin-bottom: 5px;
        }
        .task-meta {
            display: flex;
            gap: 15px;
            margin-bottom: 10px;
        }
        .meta-item {
            background: var(--vscode-badge-background);
            color: var(--vscode-badge-foreground);
            padding: 2px 8px;
            border-radius: 3px;
            font-size: 0.9em;
        }
        .task-description {
            white-space: pre-wrap;
            line-height: 1.6;
        }
    </style>
</head>
<body>
    <div class="task-header">
        <div class="task-title">${task.title}</div>
        <div class="task-meta">
            <span class="meta-item">ID: ${task.display_id || task.id}</span>
            <span class="meta-item">Status: ${task.status}</span>
            <span class="meta-item">Priority: ${task.priority}</span>
            <span class="meta-item">Type: ${task.type}</span>
        </div>
    </div>
    <div class="task-description">
        ${task.description || 'No description provided.'}
    </div>
</body>
</html>`;
}
