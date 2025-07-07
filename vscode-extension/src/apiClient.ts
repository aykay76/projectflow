import axios, { AxiosInstance } from 'axios';
import * as vscode from 'vscode';

export interface Task {
	id: string;
	display_id?: string;
	project_id: string;
	title: string;
	description: string;
	status: 'todo' | 'in_progress' | 'done' | 'blocked';
	priority: 'low' | 'medium' | 'high' | 'critical';
	type: 'epic' | 'story' | 'task' | 'subtask';
	parent_id?: string;
	children: string[];
	created_at: string;
	updated_at: string;
	started_at?: string;
}

export interface HierarchyTask {
	task: Task;
	child_tasks: HierarchyTask[];
}

export interface Project {
	id: string;
	name: string;
	description: string;
	display_prefix: string;
	settings: any;
	created_at: string;
	updated_at: string;
}

export interface CreateTaskRequest {
	title: string;
	description?: string;
	status?: string;
	priority?: string;
	type?: string;
	parent_id?: string;
	project_id?: string;
}

export interface UpdateTaskRequest {
	title?: string;
	description?: string;
	status?: string;
	priority?: string;
	type?: string;
	parent_id?: string;
}

export class ProjectFlowApiClient {
	private client: AxiosInstance;
	private baseUrl!: string;
	private project!: string;
	private apiKey?: string;

	constructor() {
		this.updateConfiguration();
		this.client = axios.create({
			timeout: 10000,
			headers: {
				'Content-Type': 'application/json',
			},
		});

		// Add request interceptor for authentication
		this.client.interceptors.request.use((config) => {
			if (this.apiKey) {
				config.headers['Authorization'] = `Bearer ${this.apiKey}`;
			}
			return config;
		});

		// Add response interceptor for error handling
		this.client.interceptors.response.use(
			(response) => response,
			(error) => {
				console.error('ProjectFlow API Error:', error);
				return Promise.reject(error);
			}
		);
	}

	public updateConfiguration(): void {
		const config = vscode.workspace.getConfiguration('projectflow');
		this.baseUrl = config.get<string>('serverUrl', 'http://localhost:16191');
		this.project = config.get<string>('project', 'PF');
		this.apiKey = config.get<string>('apiKey');

		if (this.client) {
			this.client.defaults.baseURL = this.baseUrl;
		}
	}

	async getProjects(): Promise<Project[]> {
		try {
			const response = await this.client.get<Project[]>('/api/projects');
			return response.data;
		} catch (error) {
			throw this.handleError(error, 'Failed to fetch projects');
		}
	}

	async getTasks(): Promise<Task[]> {
		try {
			const response = await this.client.get<Task[]>('/api/tasks');
			return response.data;
		} catch (error) {
			throw this.handleError(error, 'Failed to fetch tasks');
		}
	}

	async getTaskHierarchy(): Promise<HierarchyTask[]> {
		try {
			const response = await this.client.get<HierarchyTask[]>('/api/hierarchy');
			return response.data;
		} catch (error) {
			throw this.handleError(error, 'Failed to fetch task hierarchy');
		}
	}

	async getTask(id: string): Promise<Task> {
		try {
			const response = await this.client.get<Task>(`/api/tasks/${id}`);
			return response.data;
		} catch (error) {
			throw this.handleError(error, `Failed to fetch task ${id}`);
		}
	}

	async createTask(task: CreateTaskRequest): Promise<Task> {
		try {
			const taskData = {
				...task,
				project_id: task.project_id || this.project,
			};
			const response = await this.client.post<Task>('/api/tasks', taskData);
			return response.data;
		} catch (error) {
			throw this.handleError(error, 'Failed to create task');
		}
	}

	async updateTask(id: string, updates: UpdateTaskRequest): Promise<Task> {
		try {
			const response = await this.client.put<Task>(`/api/tasks/${id}`, updates);
			return response.data;
		} catch (error) {
			throw this.handleError(error, `Failed to update task ${id}`);
		}
	}

	async deleteTask(id: string): Promise<void> {
		try {
			await this.client.delete(`/api/tasks/${id}`);
		} catch (error) {
			throw this.handleError(error, `Failed to delete task ${id}`);
		}
	}

	async checkConnection(): Promise<boolean> {
		try {
			const response = await this.client.get('/health');
			return response.status === 200;
		} catch (error) {
			return false;
		}
	}

	async getConnectionStatus(): Promise<{ connected: boolean; error?: string }> {
		try {
			const response = await this.client.get('/health');
			return { connected: response.status === 200 };
		} catch (error) {
			return { 
				connected: false, 
				error: this.handleError(error, 'Connection failed').message 
			};
		}
	}

	getCurrentProject(): string {
		return this.project;
	}

	getServerUrl(): string {
		return this.baseUrl;
	}

	private handleError(error: any, message: string): Error {
		if (axios.isAxiosError(error)) {
			if (error.response) {
				return new Error(`${message}: ${error.response.status} ${error.response.statusText}`);
			} else if (error.request) {
				return new Error(`${message}: No response from server. Check if ProjectFlow server is running at ${this.baseUrl}`);
			}
		}
		return new Error(`${message}: ${error.message || error}`);
	}
}
