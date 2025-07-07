import * as vscode from 'vscode';
import { ProjectFlowApiClient, Task, HierarchyTask } from './apiClient';

export class ProjectFlowTaskProvider implements vscode.TreeDataProvider<TaskItem> {
	private _onDidChangeTreeData: vscode.EventEmitter<TaskItem | undefined | null | void> = new vscode.EventEmitter<TaskItem | undefined | null | void>();
	readonly onDidChangeTreeData: vscode.Event<TaskItem | undefined | null | void> = this._onDidChangeTreeData.event;

	constructor(private apiClient: ProjectFlowApiClient) {}

	refresh(): void {
		this._onDidChangeTreeData.fire();
	}

	getTreeItem(element: TaskItem): vscode.TreeItem {
		return element;
	}

	async getChildren(element?: TaskItem): Promise<TaskItem[]> {
		if (!element) {
			// Root level - get hierarchy
			try {
				const hierarchy = await this.apiClient.getTaskHierarchy();
				return hierarchy.map(item => {
					try {
						return this.createTaskItem(item.task, item.child_tasks);
					} catch (error) {
						console.error('Error creating task item:', error);
						// Skip malformed tasks but log the issue
						return null;
					}
				}).filter(item => item !== null) as TaskItem[];
			} catch (error) {
				const errorMessage = error instanceof Error ? error.message : String(error);
				console.error('Failed to load task hierarchy:', error);
				vscode.window.showErrorMessage(`Failed to load tasks: ${errorMessage}`);
				return [];
			}
		} else {
			// Return children of the given element
			return element.children || [];
		}
	}

	private createTaskItem(task: Task, childTasks?: HierarchyTask[]): TaskItem {
		// Add defensive check for task and required properties
		if (!task || !task.title) {
			console.error('Invalid task object:', task);
			throw new Error(`Invalid task: missing required properties. Task: ${JSON.stringify(task)}`);
		}

		const hasChildren = childTasks && childTasks.length > 0;
		const item = new TaskItem(
			task.title,
			hasChildren ? vscode.TreeItemCollapsibleState.Collapsed : vscode.TreeItemCollapsibleState.None,
			task
		);

		// Set icon based on task status
		item.iconPath = this.getStatusIcon(task.status);

		// Set description to show status, priority, and ID
		const displayId = task.display_id || task.id.substring(0, 8);
		item.description = `${displayId} | ${task.status} | ${this.getPriorityIcon(task.priority)}${task.priority}`;

		// Set tooltip with comprehensive task information
		item.tooltip = new vscode.MarkdownString(
			`**${task.title}**\n\n` +
			`**ID:** ${task.display_id || task.id}\n` +
			`**Status:** ${task.status}\n` +
			`**Priority:** ${task.priority}\n` +
			`**Type:** ${task.type}\n` +
			`**Project:** ${task.project_id}\n` +
			`**Created:** ${new Date(task.created_at).toLocaleDateString()}\n` +
			`**Updated:** ${new Date(task.updated_at).toLocaleDateString()}\n` +
			(task.description ? `\n**Description:**\n${task.description}` : '')
		);

		// Set context value for context menu
		item.contextValue = 'task';

		// Add children if they exist
		if (hasChildren) {
			item.children = childTasks!.map(child => {
				try {
					return this.createTaskItem(child.task, child.child_tasks);
				} catch (error) {
					console.error('Error creating child task item:', error);
					// Skip malformed child tasks
					return null;
				}
			}).filter(item => item !== null) as TaskItem[];
		}

		return item;
	}

	private getStatusIcon(status: string): vscode.ThemeIcon {
		switch (status) {
			case 'todo':
				return new vscode.ThemeIcon('circle-outline', new vscode.ThemeColor('list.inactiveSelectionForeground'));
			case 'in_progress':
				return new vscode.ThemeIcon('sync', new vscode.ThemeColor('charts.blue'));
			case 'done':
				return new vscode.ThemeIcon('check', new vscode.ThemeColor('charts.green'));
			case 'blocked':
				return new vscode.ThemeIcon('error', new vscode.ThemeColor('charts.red'));
			default:
				return new vscode.ThemeIcon('circle-outline');
		}
	}

	private getPriorityIcon(priority: string): string {
		switch (priority) {
			case 'critical':
				return '🔴 ';
			case 'high':
				return '🟠 ';
			case 'medium':
				return '🟡 ';
			case 'low':
				return '🟢 ';
			default:
				return '';
		}
	}
}

export class TaskItem extends vscode.TreeItem {
	public children?: TaskItem[];

	constructor(
		public readonly label: string,
		public readonly collapsibleState: vscode.TreeItemCollapsibleState,
		public readonly task: Task
	) {
		super(label, collapsibleState);
		this.id = this.task.id;
		this.command = {
			command: 'projectflow.openTask',
			title: 'Open Task',
			arguments: [this.task]
		};
	}
}
