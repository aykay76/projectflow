import * as vscode from 'vscode';
import { ProjectFlowApiClient, Task, HierarchyTask } from './apiClient';

export class ProjectFlowTaskProvider implements vscode.TreeDataProvider<TaskItem> {
	private _onDidChangeTreeData: vscode.EventEmitter<TaskItem | undefined | null | void> = new vscode.EventEmitter<TaskItem | undefined | null | void>();
	readonly onDidChangeTreeData: vscode.Event<TaskItem | undefined | null | void> = this._onDidChangeTreeData.event;

	constructor(private apiClient: ProjectFlowApiClient) {
		console.log('[ProjectFlowTaskProvider] Constructor called');
		console.log('[ProjectFlowTaskProvider] API Client server URL:', this.apiClient.getServerUrl());
		console.log('[ProjectFlowTaskProvider] API Client current project:', this.apiClient.getCurrentProject());
	}

	refresh(): void {
		console.log('[ProjectFlowTaskProvider] Refresh called - firing tree data change event');
		this._onDidChangeTreeData.fire();
	}

	getTreeItem(element: TaskItem): vscode.TreeItem {
		return element;
	}

	async getChildren(element?: TaskItem): Promise<TaskItem[]> {
		console.log('[ProjectFlowTaskProvider] getChildren called with element:', element ? `"${element.label}" (${element.task.id})` : 'ROOT');
		
		if (!element) {
			// Root level - get hierarchy
			console.log('[ProjectFlowTaskProvider] Fetching root level task hierarchy...');
			try {
				// First check connection status
				const connectionStatus = await this.apiClient.getConnectionStatus();
				console.log('[ProjectFlowTaskProvider] Connection status:', connectionStatus);
				
				if (!connectionStatus.connected) {
					const errorMsg = `Cannot connect to ProjectFlow server: ${connectionStatus.error}`;
					console.error('[ProjectFlowTaskProvider]', errorMsg);
					vscode.window.showErrorMessage(errorMsg);
					return [];
				}

				console.log('[ProjectFlowTaskProvider] Connection successful, requesting task hierarchy...');
				const hierarchy = await this.apiClient.getTaskHierarchy();
				console.log('[ProjectFlowTaskProvider] Raw hierarchy response:', JSON.stringify(hierarchy, null, 2));
				console.log('[ProjectFlowTaskProvider] Hierarchy length:', hierarchy.length);

				if (hierarchy.length === 0) {
					console.log('[ProjectFlowTaskProvider] No tasks found in hierarchy - returning empty array');
					vscode.window.showInformationMessage(`No tasks found for project: ${this.apiClient.getCurrentProject()}`);
					return [];
				}

				const taskItems = hierarchy.map((item, index) => {
					console.log(`[ProjectFlowTaskProvider] Processing hierarchy item ${index + 1}/${hierarchy.length}:`, {
						item: item,
						hasTask: !!item,
						taskId: item?.id,
						taskTitle: item?.title,
						childCount: item.child_tasks?.length || 0,
						itemKeys: Object.keys(item || {}),
						taskFields: item ? {
							id: item.id,
							title: item.title,
							status: item.status,
							priority: item.priority
						} : null
					});
					
					// Check if the hierarchy item itself is valid (since it extends Task)
					if (!item || !item.id || !item.title) {
						console.warn(`[ProjectFlowTaskProvider] Skipping hierarchy item ${index + 1} - invalid task data`);
						console.warn('[ProjectFlowTaskProvider] Full item data:', JSON.stringify(item, null, 2));
						return null;
					}
					
					try {
						const taskItem = this.createTaskItem(item, item.child_tasks);
						console.log(`[ProjectFlowTaskProvider] Successfully created task item for "${taskItem.label}"`);
						return taskItem;
					} catch (error) {
						console.error('[ProjectFlowTaskProvider] Error creating task item:', error);
						console.error('[ProjectFlowTaskProvider] Problematic task data:', JSON.stringify(item, null, 2));
						console.error('[ProjectFlowTaskProvider] Full hierarchy item:', JSON.stringify(item, null, 2));
						// Skip malformed tasks but log the issue
						return null;
					}
				}).filter(item => item !== null) as TaskItem[];

				console.log(`[ProjectFlowTaskProvider] Successfully created ${taskItems.length} task items from ${hierarchy.length} hierarchy items`);
				console.log('[ProjectFlowTaskProvider] Task items summary:', taskItems.map(item => ({
					label: item.label,
					id: item.task.id,
					hasChildren: (item.children?.length || 0) > 0,
					childCount: item.children?.length || 0
				})));

				return taskItems;
			} catch (error) {
				const errorMessage = error instanceof Error ? error.message : String(error);
				console.error('[ProjectFlowTaskProvider] Failed to load task hierarchy:', error);
				console.error('[ProjectFlowTaskProvider] Error details:', {
					name: error instanceof Error ? error.name : 'Unknown',
					message: errorMessage,
					stack: error instanceof Error ? error.stack : 'No stack trace'
				});
				vscode.window.showErrorMessage(`Failed to load tasks: ${errorMessage}`);
				return [];
			}
		} else {
			// Return children of the given element
			console.log(`[ProjectFlowTaskProvider] Returning children for task "${element.label}":`, {
				childCount: element.children?.length || 0,
				children: element.children?.map(child => ({ label: child.label, id: child.task.id })) || []
			});
			return element.children || [];
		}
	}

	private createTaskItem(task: Task, childTasks?: HierarchyTask[]): TaskItem {
		console.log('[ProjectFlowTaskProvider] Creating task item for:', {
			id: task?.id,
			title: task?.title,
			status: task?.status,
			priority: task?.priority,
			type: task?.type,
			childTaskCount: childTasks?.length || 0
		});

		// Add defensive check for task and required properties
		if (!task || !task.title) {
			const errorDetails = {
				hasTask: !!task,
				taskKeys: task ? Object.keys(task) : [],
				taskId: task?.id,
				taskTitle: task?.title
			};
			console.error('[ProjectFlowTaskProvider] Invalid task object:', errorDetails);
			console.error('[ProjectFlowTaskProvider] Full task data:', JSON.stringify(task, null, 2));
			throw new Error(`Invalid task: missing required properties. Task details: ${JSON.stringify(errorDetails)}`);
		}

		const hasChildren = childTasks && childTasks.length > 0;
		console.log(`[ProjectFlowTaskProvider] Task "${task.title}" has ${hasChildren ? childTasks!.length : 0} children`);
		
		const item = new TaskItem(
			task.title,
			hasChildren ? vscode.TreeItemCollapsibleState.Collapsed : vscode.TreeItemCollapsibleState.None,
			task
		);

		// Set icon based on task status
		item.iconPath = this.getStatusIcon(task.status);
		console.log(`[ProjectFlowTaskProvider] Set icon for status "${task.status}"`);

		// Set description to show status, priority, and ID
		const displayId = task.display_id || task.id.substring(0, 8);
		item.description = `${displayId} | ${task.status} | ${this.getPriorityIcon(task.priority)}${task.priority}`;
		console.log(`[ProjectFlowTaskProvider] Set description: "${item.description}"`);

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
			console.log(`[ProjectFlowTaskProvider] Processing ${childTasks!.length} child tasks for "${task.title}"`);
			item.children = childTasks!.map((child, index) => {
				console.log(`[ProjectFlowTaskProvider] Processing child ${index + 1}/${childTasks!.length}:`, {
					child: child,
					hasTask: !!child,
					childId: child?.id,
					childTitle: child?.title,
					childKeys: child ? Object.keys(child) : []
				});
				
				// Check if the child hierarchy item itself is valid (since it extends Task)
				if (!child || !child.id || !child.title) {
					console.warn(`[ProjectFlowTaskProvider] Skipping child ${index + 1} of "${task.title}" - invalid task data`);
					console.warn('[ProjectFlowTaskProvider] Full child data:', JSON.stringify(child, null, 2));
					return null;
				}
				
				try {
					const childItem = this.createTaskItem(child, child.child_tasks);
					console.log(`[ProjectFlowTaskProvider] Successfully created child task item: "${childItem.label}"`);
					return childItem;
				} catch (error) {
					console.error('[ProjectFlowTaskProvider] Error creating child task item:', error);
					console.error('[ProjectFlowTaskProvider] Problematic child task data:', JSON.stringify(child, null, 2));
					console.error('[ProjectFlowTaskProvider] Full child hierarchy item:', JSON.stringify(child, null, 2));
					// Skip malformed child tasks
					return null;
				}
			}).filter(item => item !== null) as TaskItem[];
			console.log(`[ProjectFlowTaskProvider] Successfully created ${item.children.length} child items for "${task.title}"`);
		}

		console.log(`[ProjectFlowTaskProvider] Successfully created task item: "${item.label}" with ${item.children?.length || 0} children`);
		return item;
	}

	private getStatusIcon(status: string): vscode.ThemeIcon {
		console.log(`[ProjectFlowTaskProvider] Getting status icon for: "${status}"`);
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
				console.warn(`[ProjectFlowTaskProvider] Unknown status "${status}", using default icon`);
				return new vscode.ThemeIcon('circle-outline');
		}
	}

	private getPriorityIcon(priority: string): string {
		console.log(`[ProjectFlowTaskProvider] Getting priority icon for: "${priority}"`);
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
				console.warn(`[ProjectFlowTaskProvider] Unknown priority "${priority}", using no icon`);
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
		console.log(`[TaskItem] Constructor called for: "${label}" (${task.id})`);
		console.log(`[TaskItem] Collapsible state:`, collapsibleState === vscode.TreeItemCollapsibleState.None ? 'None' : 
			collapsibleState === vscode.TreeItemCollapsibleState.Collapsed ? 'Collapsed' : 'Expanded');
		
		this.id = this.task.id;
		this.command = {
			command: 'projectflow.openTask',
			title: 'Open Task',
			arguments: [this.task]
		};
		
		console.log(`[TaskItem] Set command: projectflow.openTask for task ${this.task.id}`);
	}
}
